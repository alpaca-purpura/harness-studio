// Package loader implementa el reconocedor de arneses en disco (Puente 1, HS-11): convierte
// un directorio con forma de plugin CC o de arnés instalado en el domain.Graph L0, según el
// contrato FIRMADO docs/architecture/contracts/nomenclatura-arnes.md (v1, HS-10). Reemplaza los grafos
// dogfood armados a mano: los nodos salen de los archivos reales y el loader ESTAMPA
// fuente_path/clase/procedencia (§4.1 — «fuente_path deja de ser manual»).
//
// Reglas estrella que este package cumple al pie de la letra:
//   - Detección §1: `.claude-plugin/plugin.json` → plugin; `.claude/` → instalado; si hay
//     ambos, plugin manda; ninguno → ErrNoEsArnes (jamás un grafo vacío silencioso).
//   - Manifiesto §2: `arnes.l0.json` en la raíz. Ausente → modo degradado honesto: el Graph
//     sale SIN Arnes (nil) y sin error — el caller decide el check `manifiesto-ausente` rojo.
//   - Reconciliación §4.5 (D-c firmada): lo que ningún reconocedor entiende se emite como
//     nodo VISIBLE con domain.ClaseNoReconocido — nunca descarte silencioso, nunca crash.
//
// Reconocedores v1 (§3): skills (skills/<id>/SKILL.md) · rules (CLAUDE.md de la raíz) ·
// hooks (hooks/hooks.json forma-plugin + settings.json#hooks forma-instalada, una entrada =
// un nodo de banda Guardia) · command (commands/<id>.md) · output-style
// (output-styles/<id>.md) · mcp (.mcp.json, un server = un nodo) · settings (la presencia de
// settings.json) · statusline (la clave `statusLine` dentro de settings.json) — auditoría
// colateral HS-16, soporte.go. TODO honesto — reconocedores pendientes de §3, se añaden
// cuando la relación caja↔elemento tenga una decisión de diseño (no un stub que la
// invente): subagent (agents/<id>.md — falta decidir cómo se vincula a la caja que lo
// invoca) · plugin (el contenedor mismo como nodo raíz — posible colisión con el manifiesto
// arnes.l0.json, D-a). Los edges de Guardia (§4.3 «hooks matchers → edges») quedan como
// deuda declarada: el matcher nombra HERRAMIENTAS (Write|Edit), no cajas — no hay derivación
// determinista matcher→caja que no fabrique relaciones. Tampoco derivan edge los orígenes
// `libreria:`/`maquinaria:`/`terceros:`/`marcas-dormidas:` (deuda declarada, edges.go) — el
// `EdgeTipo` semántico de cada uno es una decisión de diseño, no mecánica.
package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// ErrNoEsArnes marca un directorio que no tiene ninguna de las dos formas físicas de arnés
// (§1): ni `.claude-plugin/plugin.json` (plugin CC) ni `.claude/` (arnés instalado). Error
// explícito por contrato — un no-arnés jamás produce un grafo vacío.
var ErrNoEsArnes = errors.New("no es un arnés: sin .claude-plugin/plugin.json ni .claude/")

// LoadArnes reconoce el arnés que vive en dir y lo convierte en su grafo L0.
//
// Sobre fuente_path (decisión documentada, §4.1): cada path estampado hereda la BASE que el
// caller pasó en dir — `filepath.Join(dir, <ruta interna>)` normalizado a '/'. Si el caller
// pasa un dir repo-relativo (la convención del repo: conformance resuelve fuente_path
// relativo contra la raíz del repo), los paths salen repo-relativos, idénticos a los de los
// fixtures; si pasa un dir absoluto, salen absolutos — en ambos casos el path resultante es
// resoluble tal cual, sin estado oculto ni detección mágica de raíz.
func LoadArnes(dir string) (domain.Graph, error) {
	g, _, err := LoadArnesInfo(dir)
	return g, err
}

// Info trae metadata de la carga que no cabe en domain.Graph — hoy solo el aviso de
// reconciliación entre manifiestos (RN-IDENT-3: arnes.l0.id ≠ plugin.json.name, u otro
// aviso degradado honesto); crece si hace falta. Aviso=="" ⇒ nada que reportar.
type Info struct {
	Aviso string
}

// LoadArnesInfo es LoadArnes con el aviso de reconciliación de manifiestos visible al
// caller (el Portafolio lo necesita para anotar discrepancias, S0-D del paquete
// Slice 0); LoadArnes lo descarta por compatibilidad con los callers existentes.
func LoadArnesInfo(dir string) (domain.Graph, Info, error) {
	elementos, err := detectarElementos(dir)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}

	// Nodes arranca NO-nil: un arnés reconocido sin componentes emite `nodos: []`
	// (schema-válido), nunca `null`.
	g := domain.Graph{Nodes: []domain.Box{}}

	var info Info
	g.Arnes, info.Aviso, err = leerManifiesto(dir)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}

	skills, err := reconocerSkills(elementos)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, skills...)

	// La celda de `hook` en forma-plugin es hooks/hooks.json (§3): la Guardia del arnés.
	hooks, err := reconocerHooks(elementos)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, hooks...)

	// La celda de `rule` es el CLAUDE.md de la raíz del arnés en AMBAS formas (§3).
	if regla, ok, rerr := reconocerRegla(dir); rerr != nil {
		return domain.Graph{}, Info{}, rerr
	} else if ok {
		g.Nodes = append(g.Nodes, regla)
	}

	// `command`/`output-style`: archivos sueltos <id>.md bajo su subcarpeta (§3) — mismo
	// layout en ambas formas físicas (elementos ya resuelve cuál base usar).
	comandos, err := reconocerArchivosSoporte(elementos, "commands", domain.ClaseCommand)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, comandos...)

	outputStyles, err := reconocerArchivosSoporte(elementos, "output-styles", domain.ClaseOutputStyle)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, outputStyles...)

	// `mcp`: `.mcp.json` vive en la RAÍZ del arnés en ambas formas (§3) — no bajo `.claude/`
	// como el resto de la forma instalada, así que se resuelve contra dir, no elementos.
	servidoresMCP, err := reconocerMCP(dir)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, servidoresMCP...)

	// `settings`/`statusline`/`hook` forma-instalada: las tres celdas viven en el MISMO
	// settings.json (§3) — un solo reconocedor, un solo archivo leído.
	settings, err := reconocerSettings(elementos)
	if err != nil {
		return domain.Graph{}, Info{}, err
	}
	g.Nodes = append(g.Nodes, settings...)

	g.Edges = derivarEdges(g.Nodes)

	// Modo degradado (S1-D27, contrato §2): el arnés se reconoció por sus archivos pero NO
	// tiene manifiesto que lo selle (`g.Arnes==nil`). Se marca visible; el aviso
	// `manifiesto-ausente` es la señal única que consumen el Mapa (marca roja) y el drawer.
	if g.Arnes == nil {
		g.Degradado = true
		if info.Aviso == "" {
			info.Aviso = "manifiesto-ausente: sin arnes.l0.json ni .claude-plugin/plugin.json (sin sellar)"
		}
	}
	return g, info, nil
}

// detectarElementos aplica el detector de §1 y devuelve el directorio bajo el que viven los
// reconocedores de elementos: la raíz misma en forma plugin, `.claude/` en forma instalada
// (si hay ambos, plugin manda — §1). Ninguna forma → ErrNoEsArnes.
func detectarElementos(dir string) (string, error) {
	if fi, err := os.Stat(filepath.Join(dir, ".claude-plugin", "plugin.json")); err == nil && !fi.IsDir() {
		return dir, nil
	}
	if fi, err := os.Stat(filepath.Join(dir, ".claude")); err == nil && fi.IsDir() {
		return filepath.Join(dir, ".claude"), nil
	}
	return "", fmt.Errorf("%w: %s", ErrNoEsArnes, dir)
}

// pluginJSON es el subset de `.claude-plugin/plugin.json` que leerManifiesto necesita —
// la forma oficial de plugin CC (nomenclatura-arnes.md §2, HS-12).
type pluginJSON struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// leerPluginJSON lee `.claude-plugin/plugin.json`. Ausente → (nil, nil); corrupto → error
// (el caller decide si degrada o si aborta según qué otra fuente tenga).
func leerPluginJSON(dir string) (*pluginJSON, error) {
	ruta := filepath.Join(dir, ".claude-plugin", "plugin.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta bajo el dir del arnés que el caller eligió cargar.
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", ruta, err)
	}
	var p pluginJSON
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("%s inválido: %w", ruta, err)
	}
	return &p, nil
}

// firstNonEmpty devuelve el primer string no vacío de vals, o "" si ninguno lo es.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// leerManifiesto lee el manifiesto del arnés (§2, D-a firmada + C-N-14). `arnes.l0.json`
// en la raíz MANDA cuando existe: si además hay `.claude-plugin/plugin.json`, lo usa para
// completar `Version` (y `Nombre`/`Descripcion` si vinieran vacíos — cadena bendecida §2)
// y anota una discrepancia visible si `arnes.l0.id` ≠ `plugin.json.name` (RN-IDENT-3:
// arnes.l0 gana, nunca elección silenciosa). Sin `arnes.l0.json` pero CON `plugin.json`
// (C-N-14, el caso más común — un plugin de marketplace normal) → fallback:
// `Arnes{ID:name, Nombre:displayName‖name, Descripcion:description, Version:version,
// FuenteManifiesto:"plugin.json"}`. Un `plugin.json` presente pero corrupto degrada a
// (nil, aviso, nil) — visible, jamás error fatal; a diferencia de un `arnes.l0.json`
// corrupto, que SIGUE siendo error real (es la fuente que el autor escribió a mano).
// Ninguno de los dos → (nil, "", nil), como antes: forma-instalada sin manifiesto sigue
// siendo legal.
func leerManifiesto(dir string) (*domain.Arnes, string, error) {
	rutaL0 := filepath.Join(dir, "arnes.l0.json")
	b, err := os.ReadFile(rutaL0) //nolint:gosec // G304: ruta construida sobre el dir del arnés que el caller eligió cargar (local-first).
	switch {
	case err == nil:
		var a domain.Arnes
		if uerr := json.Unmarshal(b, &a); uerr != nil {
			return nil, "", fmt.Errorf("manifiesto %s inválido: %w", rutaL0, uerr)
		}
		a.FuenteManifiesto = "arnes.l0.json"

		pj, perr := leerPluginJSON(dir)
		if perr != nil {
			// plugin.json roto no invalida un arnes.l0.json bueno: se ignora el
			// complemento, visible solo como aviso — no bloquea la fuente real.
			return &a, fmt.Sprintf("plugin.json ilegible, se ignora el complemento: %v", perr), nil
		}
		var aviso string
		if pj != nil {
			if a.Version == "" {
				a.Version = pj.Version
			}
			if a.Nombre == "" {
				a.Nombre = firstNonEmpty(pj.DisplayName, pj.Name)
			}
			if a.Descripcion == "" {
				a.Descripcion = pj.Description
			}
			if pj.Name != "" && a.ID != "" && pj.Name != a.ID {
				aviso = fmt.Sprintf("id discrepante: arnes.l0.id=%q ≠ plugin.json.name=%q (gana arnes.l0)", a.ID, pj.Name)
			}
		}
		return &a, aviso, nil

	case errors.Is(err, os.ErrNotExist):
		pj, perr := leerPluginJSON(dir)
		if perr != nil {
			return nil, fmt.Sprintf("plugin.json ilegible: %v", perr), nil // degradado visible (C-N-14), jamás error fatal.
		}
		if pj == nil {
			return nil, "", nil // ninguno de los dos manifiestos: forma-instalada sin manifiesto, legal.
		}
		return &domain.Arnes{
			ID:               pj.Name,
			Nombre:           firstNonEmpty(pj.DisplayName, pj.Name),
			Descripcion:      pj.Description,
			Version:          pj.Version,
			FuenteManifiesto: "plugin.json",
		}, "", nil

	default:
		return nil, "", fmt.Errorf("leer manifiesto %s: %w", rutaL0, err)
	}
}

// reconocerSkills escanea `skills/<id>/SKILL.md` (§3, fila skill) bajo el dir de elementos.
// El frontmatter fusionado del SKILL.md es la fuente del contract (§4.2, METODOLOGIA §3).
// Reconciliación §4.5: un dir de skill sin SKILL.md, un SKILL.md sin frontmatter parseable
// o un bloque contract que no mapea al contrato fusionado → nodo no-reconocido VISIBLE con
// su fuente_path estampado. Entradas ocultas (dotfiles) no son componentes y se ignoran.
func reconocerSkills(elementos string) ([]domain.Box, error) {
	dirSkills := filepath.Join(elementos, "skills")
	entradas, err := os.ReadDir(dirSkills)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // un arnés sin skills es legal — cero nodos, cero drama.
	}
	if err != nil {
		return nil, fmt.Errorf("escanear %s: %w", dirSkills, err)
	}

	var nodos []domain.Box
	for _, e := range entradas {
		id := e.Name()
		if strings.HasPrefix(id, ".") {
			continue
		}
		if !e.IsDir() {
			// Archivo suelto bajo skills/ (p.ej. el layout plano legacy) — visible como
			// no-reconocido, jamás descartado en silencio (§4.5).
			nodos = append(nodos, nodoNoReconocido(id, rutaEstampada(dirSkills, id)))
			continue
		}
		rutaSkill := rutaEstampada(dirSkills, id, "SKILL.md")
		b, rerr := os.ReadFile(filepath.Join(dirSkills, id, "SKILL.md")) //nolint:gosec // G304: paths bajo el dir del arnés que el caller eligió cargar.
		if rerr != nil {
			// Dir de skill sin SKILL.md legible: el fuente_path apunta al dir — es el
			// artefacto que existe.
			nodos = append(nodos, nodoNoReconocido(id, rutaEstampada(dirSkills, id)))
			continue
		}
		fm, ferr := frontmatter(b)
		if ferr != nil {
			nodos = append(nodos, nodoNoReconocido(id, rutaSkill))
			continue
		}
		contrato, cerr := contratoDe(fm)
		if cerr != nil {
			nodos = append(nodos, nodoNoReconocido(id, rutaSkill))
			continue
		}
		nodos = append(nodos, nodoSkill(id, rutaSkill, fm, contrato))
	}
	return nodos, nil
}

// nodoSkill construye el nodo de una skill reconocida: id = nombre del dir (§3), clase
// estampada por el reconocedor, nombre del frontmatter, procedencia `declarado` (sale de un
// archivo que el autor escribió). Si el contract declara caja=true, el nodo entra a la banda
// de fase con la fase y la transición de estado que la caja posee (§4.4).
//
// TODO honesto: una skill de apoyo (sin contract o caja=false) queda sin banda en v1 — el
// Mapa de HS-09 solo pinta cajas; la banda de soporte se cementa cuando el primer arnés real
// la necesite (no antes: nada se inventa por si acaso).
func nodoSkill(id, fuente string, fm map[string]any, contrato *domain.Contract) domain.Box {
	n := domain.Box{
		ID:          id,
		Clase:       domain.ClaseSkill,
		Nombre:      nombreDe(fm, id),
		FuentePath:  fuente,
		Procedencia: domain.ProcDeclarado,
		Contract:    contrato,
	}
	if contrato != nil && contrato.Caja {
		n.Banda = domain.BandaFase
		n.Fase = domain.Fase(contrato.Fase)
		n.Estado = domain.Estado(contrato.Estado)
	}
	return n
}

// nombreDe resuelve el nombre display del nodo desde el frontmatter: la clave `nombre:`
// (display en el Mapa) manda; `name:` (el id oficial CC de la skill) es el fallback; en
// última instancia el id del dir — un nodo jamás queda sin nombre (el schema lo exige).
func nombreDe(fm map[string]any, id string) string {
	if v, ok := fm["nombre"].(string); ok && v != "" {
		return v
	}
	if v, ok := fm["name"].(string); ok && v != "" {
		return v
	}
	return id
}

// contratoDe extrae el bloque `contract:` del frontmatter fusionado y lo mapea al contrato
// de dominio vía JSON — los tags json de domain.Contract (espejo de box.contract.schema.json)
// hacen TODO el trabajo de forma; aquí no se duplica ningún struct. Sin bloque contract →
// (nil, nil): skill de apoyo legal. Bloque presente que no mapea → error (el caller lo vuelve
// nodo no-reconocido §4.5; la validación semántica fina es del gate G1, no del loader).
func contratoDe(fm map[string]any) (*domain.Contract, error) {
	raw, ok := fm["contract"]
	if !ok {
		return nil, nil
	}
	jb, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("contract no serializable: %w", err)
	}
	var c domain.Contract
	if err := json.Unmarshal(jb, &c); err != nil {
		return nil, fmt.Errorf("contract no mapea al contrato fusionado: %w", err)
	}
	return &c, nil
}

// hooksJSON es la forma del archivo hooks/hooks.json de un plugin CC (subset que el
// reconocedor necesita): evento → entradas, cada entrada con su matcher opcional.
type hooksJSON struct {
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
	} `json:"hooks"`
}

// reconocerHooks escanea hooks/hooks.json (§3, fila hook — forma plugin): UNA entrada =
// UN nodo de banda Guardia, con el evento y su matcher como nombre display. Sin archivo →
// cero nodos (un arnés sin Guardia es legal). Archivo presente pero no parseable → nodo
// no-reconocido VISIBLE (§4.5) — jamás descarte silencioso.
func reconocerHooks(elementos string) ([]domain.Box, error) {
	ruta := filepath.Join(elementos, "hooks", "hooks.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: path bajo el dir del arnés que el caller eligió cargar.
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", ruta, err)
	}
	fuente := rutaEstampada(elementos, "hooks", "hooks.json")
	var h hooksJSON
	if err := json.Unmarshal(b, &h); err != nil || len(h.Hooks) == 0 {
		// §4.5 (D-c firmada): un hooks.json roto NO es error del load — es un nodo
		// no-reconocido VISIBLE; el warn queda a la vista, el arnés sigue cargando.
		return []domain.Box{nodoNoReconocido("hooks", fuente)}, nil //nolint:nilerr // reconciliación honesta: roto = visible, jamás abortar el grafo.
	}
	return nodosDeHooks(h, fuente), nil
}

// nodosDeHooks convierte una hooksJSON ya parseada en nodos de banda Guardia — compartido
// entre la forma-plugin (hooks/hooks.json) y la forma-instalada (settings.json#hooks,
// reconocerSettings en soporte.go): misma forma de datos, misma derivación, una sola fuente
// que puede diferir (el fuente_path que cada caller estampa).
func nodosDeHooks(h hooksJSON, fuente string) []domain.Box {
	// Orden determinista: eventos alfabéticos (el JSON map no tiene orden).
	eventos := make([]string, 0, len(h.Hooks))
	for ev := range h.Hooks {
		eventos = append(eventos, ev)
	}
	sort.Strings(eventos)

	var nodos []domain.Box
	for _, ev := range eventos {
		for i, entrada := range h.Hooks[ev] {
			id := "hook-" + strings.ToLower(ev)
			if i > 0 {
				id += "-" + strconv.Itoa(i+1)
			}
			nombre := ev
			if entrada.Matcher != "" {
				nombre += " · " + entrada.Matcher
			}
			nodos = append(nodos, domain.Box{
				ID:          id,
				Clase:       domain.ClaseHook,
				Nombre:      nombre,
				Banda:       domain.BandaGuardia,
				FuentePath:  fuente,
				Procedencia: domain.ProcDeclarado,
			})
		}
	}
	return nodos
}

// nodoNoReconocido emite el marcador de reconciliación honesta (§4.5, D-c firmada): visible,
// con clase domain.ClaseNoReconocido y su fuente_path estampado — obliga el fallback en FE y
// deja el warn a la vista; jamás invisible, jamás crash.
func nodoNoReconocido(id, fuente string) domain.Box {
	return domain.Box{
		ID:         id,
		Clase:      domain.ClaseNoReconocido,
		Nombre:     id + " (no reconocido)",
		FuentePath: fuente,
	}
}

// rutaEstampada construye el fuente_path de un nodo (§4.1: lo estampa el loader, deja de ser
// manual). Hereda la base tal cual la pasó el caller (ver godoc de LoadArnes) y normaliza los
// separadores a '/' — los fixtures y el FE hablan rutas con slash.
func rutaEstampada(elems ...string) string {
	return filepath.ToSlash(filepath.Join(elems...))
}
