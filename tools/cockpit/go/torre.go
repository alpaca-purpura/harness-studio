// torre.go — GET /api/torre: la Torre de Control read-only (F2 · DH-04; SPEC congelada
// products/devhub/specs/torre-read-only.md, RN-01..RN-17).
//
// Dos ejes por sistema del registro (~/.cockpit/cockpit.yaml, emitido por gen_registro.py):
//   REPO      rama · en_vuelo · ultimo_tag · gate_fabrica — hechos git computados EN el path
//             del workspace (`git -C`, worktree-aware — RN-01); cache corto ≤5s (RN-16).
//   PROYECTO  ledger de célula · board SDD — descubiertos por convención (RN-03), 100%
//             nullable con empty-state honesto (RN-02).
//
// Veredictos = categorías semánticas FIJAS + dato nativo al lado (D4, patrón DevLake);
// `no-medido` ≠ `rojo` (RN-13); todo veredicto viaja con `medido_en` (RN-17).
// Cero escritura sobre los repos monitoreados; sin `git fetch` (fuera de alcance §7).
package main

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	torreGitTTL      = 5 * time.Second  // hechos baratos por request, cacheados (RN-16)
	torreGitTimeout  = 5 * time.Second  // por comando git
	torreGateTTL     = 10 * time.Minute // gate_fabrica: boot + ?medir explícito, jamás loop (RN-16)
	torreGateTimeout = 3 * time.Minute  // el gate corre suites/generadores — es CARO por diseño
)

// ── infraestructura: git -C + caches ────────────────────────────────────────

func torreGit(ws string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), torreGitTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", ws}, args...)...).Output()
	return strings.TrimSpace(string(out)), err
}

type torreFactsEntry struct {
	at        time.Time
	repo      map[string]any
	proyectos []map[string]any
}

var (
	torreFactsMu    sync.Mutex
	torreFactsCache = map[string]torreFactsEntry{}

	torreGateMu      sync.Mutex
	torreGateCache   = map[string]map[string]any{} // workspace → veredicto (con medido_en)
	torreGateRunning = map[string]bool{}
)

func torreNoMedido(motivo string) map[string]any {
	return map[string]any{"categoria": "no-medido", "motivo": motivo, "medido_en": nowISO()}
}

// ── eje REPO (RN-01, categorías §4) ─────────────────────────────────────────

func torreRama(ws string) map[string]any {
	medido := nowISO()
	rama, err := torreGit(ws, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return torreNoMedido("git rev-parse falló: " + err.Error())
	}
	v := map[string]any{"rama": rama, "medido_en": medido}
	upstream, err := torreGit(ws, "rev-parse", "--abbrev-ref", "@{u}")
	if err != nil {
		v["categoria"] = "sin-upstream"
		return v
	}
	v["upstream"] = upstream
	counts, err := torreGit(ws, "rev-list", "--left-right", "--count", "@{u}...HEAD")
	fields := strings.Fields(counts)
	if err != nil || len(fields) != 2 {
		v["categoria"] = "no-medido"
		v["motivo"] = "git rev-list falló"
		return v
	}
	behind, _ := strconv.Atoi(fields[0]) // izquierda = solo en @{u}
	ahead, _ := strconv.Atoi(fields[1])  // derecha = solo en HEAD
	v["ahead"], v["behind"] = ahead, behind
	switch {
	case ahead == 0 && behind == 0:
		v["categoria"] = "sincronizada"
	case ahead > 0 && behind == 0:
		v["categoria"] = "adelante"
	case ahead == 0 && behind > 0:
		v["categoria"] = "atras"
	default:
		v["categoria"] = "divergida"
	}
	return v
}

func torreEnVuelo(ws string) map[string]any {
	medido := nowISO()
	ctx, cancel := context.WithTimeout(context.Background(), torreGitTimeout)
	defer cancel()
	// sin TrimSpace global: el primer caracter XY del porcelain puede ser espacio.
	out, err := exec.CommandContext(ctx, "git", "-C", ws, "status", "--porcelain").Output()
	if err != nil {
		return torreNoMedido("git status falló: " + err.Error())
	}
	archivos := []map[string]any{}
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		archivos = append(archivos, map[string]any{"estado": line[:2], "path": line[3:]})
	}
	cat := "limpio"
	if len(archivos) > 0 {
		cat = "en-vuelo"
	}
	return map[string]any{"categoria": cat, "archivos": archivos, "total": len(archivos), "medido_en": medido}
}

func torreUltimoTag(ws string) map[string]any {
	medido := nowISO()
	tag, err := torreGit(ws, "describe", "--tags", "--abbrev=0")
	if err != nil {
		// el repo existe (torreRepo ya validó rev-parse) → sin tags alcanzables = sin-tag, no error
		return map[string]any{"categoria": "sin-tag", "medido_en": medido}
	}
	v := map[string]any{"categoria": "taggeado", "tag": tag, "medido_en": medido}
	if n, err := torreGit(ws, "rev-list", tag+"..HEAD", "--count"); err == nil {
		if i, convErr := strconv.Atoi(n); convErr == nil {
			v["commits_desde"] = i
		}
	}
	if fecha, err := torreGit(ws, "log", "-1", "--format=%cs", tag); err == nil {
		v["fecha"] = fecha
	}
	return v
}

// torreRepo: el eje REPO completo menos gate_fabrica (que tiene su propio ciclo — RN-16).
func torreRepo(ws string) map[string]any {
	root, err := torreGit(ws, "rev-parse", "--show-toplevel")
	if err != nil {
		motivo := "no es un repo git: " + ws
		return map[string]any{
			"rama": torreNoMedido(motivo), "en_vuelo": torreNoMedido(motivo),
			"ultimo_tag": torreNoMedido(motivo),
		}
	}
	return map[string]any{
		"repo_root":  root, // RN-01: el workspace puede ser worktree — se reporta el root real
		"rama":       torreRama(ws),
		"en_vuelo":   torreEnVuelo(ws),
		"ultimo_tag": torreUltimoTag(ws),
	}
}

// ── gate_fabrica (RN-05/RN-16: boot + ?medir explícito, TTL 10 min, jamás loop) ──

// torreShell: el intérprete con el que se corre un gate_check.
//
// Los gate_check del registry se escriben en sintaxis POSIX (`python3 x.py && make y`).
// En Windows `sh` existe si hay Git Bash — se prefiere, porque es el único que entiende
// esa sintaxis. Sin `sh` el gate NO se traduce a cmd.exe a la fuerza: se devuelve vacío y
// el caller lo reporta `no-medido` con motivo, en vez de inventar un rojo por un `&&` que
// cmd interpreta distinto. Gris honesto, no falso veredicto.
func torreShell() []string {
	if runtime.GOOS != "windows" {
		return []string{"sh", "-c"}
	}
	if _, err := exec.LookPath("sh"); err == nil {
		return []string{"sh", "-c"}
	}
	return nil
}

func torreRunGate(ws, comando string) map[string]any {
	shell := torreShell()
	if shell == nil {
		return map[string]any{
			"comando":   comando,
			"categoria": "no-medido",
			"motivo":    "sin shell POSIX en el PATH (instalá Git Bash) — el gate no se traduce a cmd.exe a la fuerza",
			"medido_en": nowISO(),
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), torreGateTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, shell[0], append(shell[1:], comando)...)
	cmd.Dir = ws
	out, err := cmd.CombinedOutput()
	medido := nowISO()

	resumen := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(resumen) > 30 {
		resumen = append(resumen[:30], "… (salida truncada)")
	}
	v := map[string]any{"comando": comando, "resumen": strings.Join(resumen, "\n"), "medido_en": medido}

	if ctx.Err() == context.DeadlineExceeded {
		v["categoria"] = "no-medido" // timeout ≠ rojo (RN-13): el check no llegó a veredicto
		v["motivo"] = "timeout tras " + torreGateTimeout.String()
		return v
	}
	exit := 0
	if err != nil {
		exit = -1
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			v["categoria"] = "no-medido" // no se pudo EJECUTAR (sh ausente, permiso) ≠ falló
			v["motivo"] = "no se pudo ejecutar: " + err.Error()
			v["exit_code"] = exit
			return v
		}
	}
	v["exit_code"] = exit
	if exit == 0 {
		v["categoria"] = "verde"
	} else {
		v["categoria"] = "rojo" // rojo = el check corrió y FALLÓ (RN-13)
	}
	return v
}

// torreMedirGate mide y cachea el gate de UN workspace (serializado por workspace).
func torreMedirGate(ws, comando string) map[string]any {
	torreGateMu.Lock()
	if torreGateRunning[ws] {
		cached := torreGateCache[ws]
		torreGateMu.Unlock()
		if cached != nil {
			return cached
		}
		return torreNoMedido("medición en curso")
	}
	torreGateRunning[ws] = true
	torreGateMu.Unlock()

	v := torreRunGate(ws, comando)

	torreGateMu.Lock()
	torreGateCache[ws] = v
	torreGateRunning[ws] = false
	torreGateMu.Unlock()
	return v
}

func torreGateEdad(v map[string]any) time.Duration {
	s, _ := v["medido_en"].(string)
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		return torreGateTTL + time.Hour
	}
	return time.Since(t)
}

// torreGateFabrica: veredicto cacheado; `medir` (query explícita) re-mide solo si el
// cache superó el TTL — el dato viejo no se descarta, viaja con su edad (RN-17).
func torreGateFabrica(p registryProject, medir bool) map[string]any {
	if p.GateCheck == "" {
		return map[string]any{"categoria": "sin-gate", "medido_en": nowISO()} // RN-05
	}
	torreGateMu.Lock()
	cached := torreGateCache[p.Path]
	torreGateMu.Unlock()
	if medir && (cached == nil || torreGateEdad(cached) > torreGateTTL) {
		return torreMedirGate(p.Path, p.GateCheck)
	}
	if cached != nil {
		return cached
	}
	return torreNoMedido("aún no medido — re-mide con ?medir=gate_fabrica")
}

// torreBootGates: medición inicial al boot del daemon (RN-16), secuencial para no
// apilar checks caros en paralelo.
func torreBootGates(reg *registry) {
	for _, p := range reg.Projects {
		if p.GateCheck != "" && p.Path != "" && dirExists(p.Path) {
			torreMedirGate(p.Path, p.GateCheck)
		}
	}
}

// ── eje PROYECTO (RN-03: descubrimiento por convención · RN-11: contrato ledger) ──

// torreLedger lee un espejo máquina de ledger (célula RN-10 o global) → contrato RN-11.
func torreLedger(path string) map[string]any {
	medido := nowISO()
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"categoria": "sin-ledger", "medido_en": medido}
	}
	data, err := parseYAMLMap(raw)
	if err != nil {
		return torreNoMedido("ledger.yaml ilegible: " + err.Error())
	}
	v := map[string]any{"categoria": "con-ledger", "path": path, "medido_en": medido}
	if fichas, ok := data["fichas"].([]any); ok && len(fichas) > 0 {
		v["total_fichas"] = len(fichas)
		if f, ok := fichas[len(fichas)-1].(map[string]any); ok {
			v["ultima_ficha"] = map[string]any{
				"id": f["id"], "titulo": f["titulo"], "estado": f["estado"], "vigencia": f["vigencia"],
			}
		}
	} else {
		v["total_fichas"] = 0
	}
	// la FECHA de actividad: max del log (células, RN-10) o de decisiones (global)
	entradas, _ := data["log"].([]any)
	if entradas == nil {
		entradas, _ = data["decisiones"].([]any)
	}
	ultima := ""
	for _, e := range entradas {
		if m, ok := e.(map[string]any); ok {
			if f, ok := m["fecha"].(string); ok && f > ultima {
				ultima = f
			}
		}
	}
	if ultima != "" {
		v["ultima_fecha"] = ultima
	}
	return v
}

// torreArquitectura lee el modelo curado de una célula (arquitectura-como-dato, I-73) →
// veredicto RN-25. El YAML gated ES el contrato cross-célula (RN-21) — lectura por
// convención, cero imports. meta.clase = discriminador de render; ausente → `modelo`
// (el CURADO es arquitectura-como-dato por naturaleza — RN-22).
func torreArquitectura(path string) map[string]any {
	medido := nowISO()
	raw, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"categoria": "sin-arquitectura", "medido_en": medido} // RN-02: jamás inventa
	}
	data, err := parseYAMLMap(raw)
	if err != nil {
		return torreNoMedido("arquitectura.yaml ilegible: " + err.Error()) // RN-13
	}
	comps, _ := data["componentes"].([]any)
	rels, _ := data["relaciones"].([]any)
	v := map[string]any{
		"categoria": "con-arquitectura", "clase": "modelo", "path": path, "medido_en": medido,
		"total_componentes": len(comps), "total_relaciones": len(rels),
		"planos": data["planos"], "tipos": data["tipos"],
		"componentes": comps, "relaciones": rels,
	}
	if meta, ok := data["meta"].(map[string]any); ok {
		if c, ok := meta["clase"].(string); ok && c != "" {
			v["clase"] = c // declaración explícita gana (RN-22)
		}
		for _, k := range []string{"id", "nombre", "version", "proposito"} {
			if val, ok := meta[k]; ok {
				v[k] = val
			}
		}
	}
	return v
}

// torreBoard: conteos del board SDD de (workspace, slug) — stories por estado + release en curso.
func torreBoard(ws, slug string) map[string]any {
	medido := nowISO()
	dir := boardDirFor(ws, slug)
	porEstado := map[string]int{}
	total := 0
	for _, d := range listStoryDirs(filepath.Join(dir, "stories")) {
		doc, err := readMarkdownWithFrontmatter(filepath.Join(d, "story.md"))
		if err != nil {
			continue
		}
		estado, _ := fmGetOr(doc.Frontmatter, "idea", "state").(string)
		porEstado[estado]++
		total++
	}

	var enCurso map[string]any
	relDir := filepath.Join(dir, "releases")
	if entries, err := os.ReadDir(relDir); err == nil {
		releases := []map[string]any{}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
				continue
			}
			if rel, err := readRelease(filepath.Join(relDir, e.Name())); err == nil {
				releases = append(releases, rel)
			}
		}
		sort.SliceStable(releases, func(i, j int) bool {
			return toFloat(releases[i]["order"]) < toFloat(releases[j]["order"])
		})
		for _, rel := range releases {
			status, _ := rel["status"].(string)
			if status != "shipped" && status != "archived" {
				enCurso = map[string]any{"release_id": rel["release_id"], "name": rel["name"], "status": status}
				break
			}
		}
	}

	if total == 0 && enCurso == nil {
		return map[string]any{"categoria": "sin-board", "medido_en": medido}
	}
	return map[string]any{
		"categoria": "con-board", "stories": porEstado, "total_stories": total,
		"release_en_curso": enCurso, "medido_en": medido,
	}
}

// torreProyectos: eje PROYECTO por convención (RN-03) — products/<dir>/LEDGER.md → célula;
// docs/product/ propio o <dir>/docs/product/ → sdd (la convención de sistemasIn).
func torreProyectos(ws string) []map[string]any {
	out := []map[string]any{}
	prodDir := filepath.Join(ws, "products")
	if entries, err := os.ReadDir(prodDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() || !fileExists(filepath.Join(prodDir, e.Name(), "LEDGER.md")) {
				continue
			}
			out = append(out, map[string]any{
				"slug": e.Name(), "tipo": "celula",
				"ledger":       torreLedger(filepath.Join(prodDir, e.Name(), "ledger.yaml")),
				"board":        map[string]any{"categoria": "sin-board", "medido_en": nowISO()},
				"arquitectura": torreArquitectura(filepath.Join(prodDir, e.Name(), "arquitectura.yaml")), // RN-20/RN-25
			})
		}
	}
	for _, slug := range selectableSistemasIn(ws) {
		out = append(out, map[string]any{
			"slug": slug, "tipo": "sdd",
			"ledger": torreLedger(filepath.Join(boardDirFor(ws, slug), "ledger.yaml")),
			"board":  torreBoard(ws, slug),
			// RN-20: los engagements aún no curan arquitectura — v2
			"arquitectura": map[string]any{"categoria": "sin-arquitectura", "medido_en": nowISO()},
		})
	}
	return out
}

// ── la fila y el handler (RN-15/RN-17) ──────────────────────────────────────

// torreFacts: eje REPO (sin gate) + eje PROYECTO, cacheados juntos ≤5s (RN-16).
func torreFacts(ws string) (map[string]any, []map[string]any) {
	torreFactsMu.Lock()
	if e, ok := torreFactsCache[ws]; ok && time.Since(e.at) < torreGitTTL {
		torreFactsMu.Unlock()
		return e.repo, e.proyectos
	}
	torreFactsMu.Unlock()

	repo := torreRepo(ws)
	proyectos := torreProyectos(ws)

	torreFactsMu.Lock()
	torreFactsCache[ws] = torreFactsEntry{at: time.Now(), repo: repo, proyectos: proyectos}
	torreFactsMu.Unlock()
	return repo, proyectos
}

func torreSistemas(reg *registry, medir bool) []map[string]any {
	sistemas := []map[string]any{}
	for _, p := range reg.Projects {
		nombre := p.Nombre
		if nombre == "" {
			nombre = p.Name
		}
		fila := map[string]any{"slug": p.Name, "nombre": nombre, "workspace": p.Path}
		if p.Path == "" || !dirExists(p.Path) {
			// RN-17: el registro dice que debería existir — eso ES señal; la fila no desaparece.
			motivo := "workspace no existe en disco: " + p.Path
			if p.Path == "" {
				motivo = "sistema sin workspace en la config del operador (~/.config/prenter/devhub.yaml)"
			}
			fila["repo"] = map[string]any{
				"rama": torreNoMedido(motivo), "en_vuelo": torreNoMedido(motivo),
				"ultimo_tag": torreNoMedido(motivo), "gate_fabrica": torreNoMedido(motivo),
			}
			fila["proyectos"] = []map[string]any{}
			sistemas = append(sistemas, fila)
			continue
		}
		repo, proyectos := torreFacts(p.Path)
		repoConGate := map[string]any{"gate_fabrica": torreGateFabrica(p, medir)}
		for k, v := range repo {
			repoConGate[k] = v
		}
		fila["repo"] = repoConGate
		fila["proyectos"] = proyectos
		sistemas = append(sistemas, fila)
	}
	return sistemas
}

func handleTorre(w http.ResponseWriter, r *http.Request) {
	medir := r.URL.Query().Get("medir") == "gate_fabrica"
	reg := loadRegistry()
	writeJSON(w, 200, map[string]any{
		"sistemas":    torreSistemas(reg, medir),
		"generado_en": nowISO(),
	})
}
