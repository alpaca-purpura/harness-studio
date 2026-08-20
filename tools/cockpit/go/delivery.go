// delivery.go — cockpit de delivery (F5 · DH-08 · SPEC delivery-cockpit.md RN-35..RN-43):
// la capa 1 (descriptor de proceso) PARAMETRIZA la capa 3 (sesiones de agente) con la
// capa 2 (contexto as-code) inyectada — el norte v3 conectado.
//
// Tres piezas:
//  1. Supervisor del sidecar (RN-35): el binario spawnea `node dist/index.js` al boot,
//     handshake por stdout (SIDECAR_PORT=n), muerte ligada por stdin-pipe. Sin sidecar
//     la consola vive y responde honesto (RN-40).
//  2. Plantilla parametrizada (RN-39): Go-nativo, deriva el tramo SOLO del descriptor
//     (cero literales de estado) + descubre el contexto as-code por convención (D14).
//  3. Proxy /api/delivery/* → sidecar (RN-40): salud y sesiones.
//
// La sesión NO transiciona stories (RN-31): el humano revisa salida/ y aplica.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── supervisor del sidecar (RN-35) ───────────────────────────────────────────

type sidecarSupervisor struct {
	mu     sync.Mutex
	port   int
	motivo string
}

func (s *sidecarSupervisor) set(port int, motivo string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.port, s.motivo = port, motivo
}

func (s *sidecarSupervisor) estado() (int, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port, s.motivo
}

var deliverySidecar = &sidecarSupervisor{motivo: "sidecar no arrancado"}

// flagSidecarDir: override por flag -sidecar (cli.go); "" = auto-detección.
var flagSidecarDir string

// sidecarStdin se retiene la vida del proceso: si el binario muere, el pipe se
// cierra y el sidecar se apaga solo (contrato RN-35, portable sin Pdeathsig).
var sidecarStdin io.WriteCloser

// encontrarSidecar: flag > env DEVHUB_SIDECAR > {exeDir}/sidecar > {exeDir}/../sidecar.
func encontrarSidecar() (string, string) {
	var candidatos []string
	if flagSidecarDir != "" {
		candidatos = append(candidatos, flagSidecarDir)
	}
	if env := os.Getenv("DEVHUB_SIDECAR"); env != "" {
		candidatos = append(candidatos, env)
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidatos = append(candidatos,
			filepath.Join(exeDir, "sidecar"),
			filepath.Join(exeDir, "..", "sidecar"))
	}
	for _, c := range candidatos {
		if abs, err := filepath.Abs(expandHome(c)); err == nil {
			if _, err := os.Stat(filepath.Join(abs, "dist", "index.js")); err == nil {
				return abs, ""
			}
		}
	}
	return "", "sidecar no encontrado (dist/index.js) — corre `npm install && npm run build` en products/devhub/sidecar"
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// deliveryBoot arranca y acompaña el sidecar (se llama en goroutine desde serve()).
// v1 NO reinicia un sidecar caído (RN-35): el proxy responde honesto y el resto vive.
func deliveryBoot() {
	dir, motivo := encontrarSidecar()
	if dir == "" {
		deliverySidecar.set(0, motivo)
		log.Printf("⚠ delivery: %s", motivo)
		return
	}
	node, err := exec.LookPath("node")
	if err != nil {
		deliverySidecar.set(0, "node no está en PATH — el sidecar de delivery necesita Node")
		log.Printf("⚠ delivery: node no está en PATH")
		return
	}
	cmd := exec.Command(node, filepath.Join("dist", "index.js"))
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		deliverySidecar.set(0, "no se pudo abrir stdin del sidecar: "+err.Error())
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		deliverySidecar.set(0, "no se pudo abrir stdout del sidecar: "+err.Error())
		return
	}
	if err := cmd.Start(); err != nil {
		deliverySidecar.set(0, "el sidecar no arrancó: "+err.Error())
		log.Printf("⚠ delivery: sidecar no arrancó: %v", err)
		return
	}
	sidecarStdin = stdin // retenido: su cierre (muerte del binario) apaga el sidecar
	deliverySidecar.set(0, "sidecar arrancando")

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		if p, ok := strings.CutPrefix(line, "SIDECAR_PORT="); ok {
			if port, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && port > 0 {
				deliverySidecar.set(port, "")
				log.Printf("🛰 devhub-sidecar arriba · 127.0.0.1:%d (%s)", port, dir)
			}
			continue
		}
		log.Printf("sidecar: %s", line)
	}
	_ = cmd.Wait()
	deliverySidecar.set(0, "el sidecar terminó — reinicia el binario para recuperar delivery")
	log.Printf("⚠ delivery: sidecar terminó")
}

// ── derivaciones del descriptor (RN-39 — cero literales de estado) ────────────

// siguienteTramoRol: primera transición ejecutor:rol alcanzable desde `estado`
// recorriendo transiciones[] en su orden, profundidad ≤ 2 (cubre el estado inicial,
// que solo tiene salidas de operador).
func siguienteTramoRol(p *procesoDescriptor, estado string) (string, string, bool) {
	for _, t := range p.Transiciones {
		if t.De == estado && t.Ejecutor == "rol" {
			return t.De, t.A, true
		}
	}
	for _, t1 := range p.Transiciones {
		if t1.De != estado {
			continue
		}
		for _, t2 := range p.Transiciones {
			if t2.De == t1.A && t2.Ejecutor == "rol" {
				return t2.De, t2.A, true
			}
		}
	}
	return "", "", false
}

// gatesDelTramo: gates cuyo momento toca el tramo (estado destino o la transición de→a).
func gatesDelTramo(p *procesoDescriptor, de, a string) []procesoGate {
	var out []procesoGate
	for _, g := range p.Gates {
		for _, m := range g.Momento {
			if m == a || m == de+"→"+a {
				out = append(out, g)
				break
			}
		}
	}
	return out
}

// ── contexto as-code por convención (D14/RN-39) ───────────────────────────────

type deliveryFuente struct {
	Nombre  string `json:"nombre"`
	RutaAbs string `json:"ruta_abs"`
}

type deliveryContexto struct {
	ID      string           `json:"id"`
	Fuentes []deliveryFuente `json:"fuentes"`
}

// contextosAsCode: candidatos = {sistemaRoot}/arquitectura.yaml y
// {sistemaRoot}/*/arquitectura.yaml (el MISMO glob que el gate D7), cada uno con su
// VISION.md hermano si existe. 0 hallazgos → lista vacía (contexto honesto vacío).
func contextosAsCode(sistemaRoot string) []deliveryContexto {
	var out []deliveryContexto
	agregar := func(dir string) {
		arq := filepath.Join(dir, "arquitectura.yaml")
		if _, err := os.Stat(arq); err != nil {
			return
		}
		ctx := deliveryContexto{
			ID:      filepath.Base(dir),
			Fuentes: []deliveryFuente{{Nombre: "arquitectura.yaml", RutaAbs: arq}},
		}
		if vis := filepath.Join(dir, "VISION.md"); fileExists(vis) {
			ctx.Fuentes = append(ctx.Fuentes, deliveryFuente{Nombre: "VISION.md", RutaAbs: vis})
		}
		out = append(out, ctx)
	}
	agregar(sistemaRoot)
	if entries, err := os.ReadDir(sistemaRoot); err == nil {
		var subdirs []string
		for _, e := range entries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				subdirs = append(subdirs, e.Name())
			}
		}
		sort.Strings(subdirs)
		for _, s := range subdirs {
			agregar(filepath.Join(sistemaRoot, s))
		}
	}
	return out
}

// ── plantilla parametrizada (RN-39) ───────────────────────────────────────────

type deliveryPlantilla struct {
	Prompt               string           `json:"prompt"`
	Tramo                map[string]any   `json:"tramo"`
	Gates                []procesoGate    `json:"gates"`
	Contexto             deliveryContexto `json:"contexto"`
	ContextosDisponibles []string         `json:"contextos_disponibles"`
	Fuentes              []deliveryFuente `json:"fuentes"`
}

// plantillaPara arma la plantilla completa. status != 200 → (nil, status, motivo).
func plantillaPara(sistema, storyID, contextoID string) (*deliveryPlantilla, int, string) {
	if !containsStr(getSelectableSistemas(), sistema) {
		return nil, 400, "sistema desconocido: " + sistema
	}
	storyDir := filepath.Join(storiesPath(sistema), storyID)
	ckpt := readCheckpoint(storyDir, sistema, false)
	if ckpt == nil {
		return nil, 404, "story desconocida: " + storyID
	}
	estadoID, _ := ckpt["state"].(string)
	var actual *procesoEstado
	for i := range proceso.Estados {
		if proceso.Estados[i].ID == estadoID {
			actual = &proceso.Estados[i]
			break
		}
	}
	if actual == nil {
		return nil, 422, fmt.Sprintf("estado %q fuera del descriptor", estadoID)
	}
	if proceso.esTerminal(estadoID) {
		return nil, 422, fmt.Sprintf("no-lanzable: %s está en categoría terminal (%s)", estadoID, actual.Categoria)
	}
	if actual.Categoria == "pausado" {
		return nil, 422, fmt.Sprintf("no-lanzable: %s está en categoría pausado", estadoID)
	}
	de, a, ok := siguienteTramoRol(proceso, estadoID)
	if !ok {
		return nil, 422, fmt.Sprintf("no-lanzable: sin tramo de rol alcanzable desde %s", estadoID)
	}
	var destino *procesoEstado
	for i := range proceso.Estados {
		if proceso.Estados[i].ID == a {
			destino = &proceso.Estados[i]
			break
		}
	}
	gates := gatesDelTramo(proceso, de, a)

	// contexto as-code
	sistemaRoot := filepath.Dir(sistemaDocsRoot(sistema))
	contextos := contextosAsCode(sistemaRoot)
	ids := make([]string, len(contextos))
	for i, c := range contextos {
		ids[i] = c.ID
	}
	elegido := deliveryContexto{ID: "sin-contexto"}
	if len(contextos) > 0 {
		elegido = contextos[0] // default: el primero (RN-39)
		for _, c := range contextos {
			if c.ID == contextoID {
				elegido = c
				break
			}
		}
	}
	// fuentes: el checkpoint de la story SIEMPRE primero
	fuentes := append([]deliveryFuente{{Nombre: "checkpoint.md", RutaAbs: filepath.Join(storyDir, "checkpoint.md")}},
		elegido.Fuentes...)

	goal, _ := ckpt["goal"].(string)
	prompt := construirPrompt(proceso, sistema, storyID, goal, actual, destino, gates, elegido)

	return &deliveryPlantilla{
		Prompt: prompt,
		Tramo: map[string]any{
			"de": de, "a": a,
			"verbo":            proceso.verboDe(de, a), // el evento que el tramo emite (RN-47)
			"categoria_actual": actual.Categoria,
			"duenos":           proceso.Duenos[a],
		},
		Gates:                gates,
		Contexto:             elegido,
		ContextosDisponibles: ids,
		Fuentes:              fuentes,
	}, 200, ""
}

// construirPrompt: el prompt precargado — el descriptor manda, el humano encausa.
func construirPrompt(p *procesoDescriptor, sistema, storyID, goal string,
	actual, destino *procesoEstado, gates []procesoGate, ctx deliveryContexto) string {

	var b strings.Builder
	duenos := p.Duenos[destino.ID]
	if len(duenos) > 0 {
		d := duenos[0]
		quien := d.Arnes
		if quien == "" {
			quien = d.Rol
		}
		fmt.Fprintf(&b, "Eres el arnés %s (rol %s", quien, d.Rol)
		if d.Nota != "" {
			fmt.Fprintf(&b, " — %s", d.Nota)
		}
		b.WriteString(").\n")
	} else {
		b.WriteString("Eres un agente de delivery.\n")
	}
	fmt.Fprintf(&b, "Sistema: %s · Proceso: %s v%d (%s).\n\n",
		sistema, p.Descriptor.Nombre, p.Descriptor.Version, p.Descriptor.ID)

	fmt.Fprintf(&b, "Story «%s» — estado actual: %s (categoría %s)", storyID, actual.ID, actual.Categoria)
	if actual.Descripcion != "" {
		fmt.Fprintf(&b, ": %s", actual.Descripcion)
	}
	b.WriteString(".\n")
	if goal != "" {
		fmt.Fprintf(&b, "Goal de la story: %s\n", goal)
	}
	fmt.Fprintf(&b, "Objetivo del tramo: dejar el trabajo listo para «%s»", destino.ID)
	if destino.Descripcion != "" {
		fmt.Fprintf(&b, " (%s)", destino.Descripcion)
	}
	b.WriteString(".\n")
	for _, g := range gates {
		fmt.Fprintf(&b, "\nGate «%s» (autoridad: %s) — checklist que gobierna el tramo:\n", g.Nombre, g.Autoridad.Rol)
		for _, item := range g.Checklist {
			fmt.Fprintf(&b, "- %s\n", item)
		}
	}

	b.WriteString("\nContexto as-code (COPIAS en ./contexto/ — tu única fuente de verdad):\n")
	b.WriteString("- checkpoint.md — la story completa (spec actual + historial)\n")
	for _, f := range ctx.Fuentes {
		fmt.Fprintf(&b, "- %s — de la célula «%s»\n", f.Nombre, ctx.ID)
	}
	if len(ctx.Fuentes) == 0 {
		b.WriteString("- (sin arquitectura/visión declaradas para este sistema — trabaja solo con la story)\n")
	}

	b.WriteString(`
Mandato:
1. Lee TODO ./contexto/ antes de escribir.
2. Produce tu entregable en ./salida/ (Markdown): la spec refinada que deja la story lista para el estado destino` + fmt.Sprintf(" «%s»", destino.ID) + `, cumpliendo el checklist del gate si lo hay.
3. Trabaja SOLO en este workspace: no existe el repo para ti; no edites ./contexto/.
4. NO transiciones estados ni simules hacerlo — el ciclo lo ejecutan los skills de la casa; tu salida la revisa y aplica un humano.
`)
	return b.String()
}

// ── handlers HTTP ─────────────────────────────────────────────────────────────

const sidecarTimeout = 5 * time.Second

func sidecarURL(path string) (string, string) {
	port, motivo := deliverySidecar.estado()
	if port == 0 {
		if motivo == "" {
			motivo = "sidecar no disponible"
		}
		return "", motivo
	}
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, path), ""
}

// noDisponible: la respuesta honesta cuando el sidecar no está (RN-40).
func noDisponible(w http.ResponseWriter, motivo string) {
	writeJSON(w, http.StatusServiceUnavailable, map[string]any{"disponible": false, "motivo": motivo})
}

func handleDeliverySalud(w http.ResponseWriter, r *http.Request) {
	url, motivo := sidecarURL("/salud")
	if url == "" {
		noDisponible(w, motivo)
		return
	}
	client := http.Client{Timeout: sidecarTimeout}
	resp, err := client.Get(url)
	if err != nil {
		noDisponible(w, "el sidecar no responde: "+err.Error())
		return
	}
	defer resp.Body.Close()
	var salud map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&salud); err != nil {
		noDisponible(w, "respuesta ilegible del sidecar")
		return
	}
	salud["disponible"] = true
	writeJSON(w, http.StatusOK, salud)
}

func handleDeliveryPlantilla(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	pl, status, motivo := plantillaPara(q.Get("sistema"), q.Get("story"), q.Get("contexto"))
	if status != 200 {
		writeError(w, status, motivo, nil)
		return
	}
	writeJSON(w, http.StatusOK, pl)
}

// reenviar: proxy chico al sidecar — copia status + body (JSON puro, RN-40).
func reenviar(w http.ResponseWriter, method, path string, body []byte) {
	url, motivo := sidecarURL(path)
	if url == "" {
		noDisponible(w, motivo)
		return
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		noDisponible(w, err.Error())
		return
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := http.Client{Timeout: sidecarTimeout}
	resp, err := client.Do(req)
	if err != nil {
		noDisponible(w, "el sidecar no responde: "+err.Error())
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func handleDeliverySesionesList(w http.ResponseWriter, r *http.Request) {
	reenviar(w, http.MethodGet, "/sesiones", nil)
}

func handleDeliverySesionUna(w http.ResponseWriter, r *http.Request) {
	reenviar(w, http.MethodGet, "/sesiones/"+r.PathValue("id"), nil)
}

// handleDeliverySesionCrear: valida en Go (story real, prompt no vacío), re-deriva las
// fuentes as-code (el cliente manda IDs, jamás rutas) y lanza en el sidecar.
func handleDeliverySesionCrear(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema  string `json:"sistema"`
		Story    string `json:"story"`
		Prompt   string `json:"prompt"`
		Contexto string `json:"contexto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if strings.TrimSpace(body.Prompt) == "" {
		writeError(w, 400, "prompt requerido", nil)
		return
	}
	pl, status, motivo := plantillaPara(body.Sistema, body.Story, body.Contexto)
	if status != 200 {
		writeError(w, status, motivo, nil)
		return
	}
	payload, _ := json.Marshal(map[string]any{
		"prompt":   body.Prompt, // el prompt ENCAUSADO por el humano, no la plantilla cruda
		"contexto": pl.Fuentes,
		"meta": map[string]any{
			"sistema": body.Sistema,
			"story":   body.Story,
			"tramo":   fmt.Sprintf("%s→%s", pl.Tramo["de"], pl.Tramo["a"]),
		},
	})
	reenviar(w, http.MethodPost, "/sesiones", payload)
}
