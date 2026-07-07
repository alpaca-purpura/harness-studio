// Package loader implementa el reconocedor de arneses en disco (Puente 1, HS-11): convierte
// un directorio con forma de plugin CC o de arnés instalado en el domain.Graph L0, según el
// contrato FIRMADO arch/contracts/nomenclatura-arnes.md (v1, HS-10). Reemplaza los grafos
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
// Reconocedores v1 (§3): skills (skills/<id>/SKILL.md) y rules (CLAUDE.md de la raíz).
// TODO honesto — reconocedores pendientes de §3, se añaden con el primer arnés real que los
// use (el dogfood dev-full-cycle aún no los tiene; jamás un stub que fabrique nodos):
// hook (hooks/hooks.json → banda Guardia) · mcp (.mcp.json) · command (commands/<id>.md) ·
// subagent (agents/<id>.md) · settings · output-style (output-styles/<id>.md) · statusline ·
// plugin (el contenedor mismo como nodo raíz).
package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	elementos, err := detectarElementos(dir)
	if err != nil {
		return domain.Graph{}, err
	}

	// Nodes arranca NO-nil: un arnés reconocido sin componentes emite `nodos: []`
	// (schema-válido), nunca `null`.
	g := domain.Graph{Nodes: []domain.Box{}}

	g.Arnes, err = leerManifiesto(dir)
	if err != nil {
		return domain.Graph{}, err
	}

	skills, err := reconocerSkills(elementos)
	if err != nil {
		return domain.Graph{}, err
	}
	g.Nodes = append(g.Nodes, skills...)

	// La celda de `rule` es el CLAUDE.md de la raíz del arnés en AMBAS formas (§3).
	if regla, ok, rerr := reconocerRegla(dir); rerr != nil {
		return domain.Graph{}, rerr
	} else if ok {
		g.Nodes = append(g.Nodes, regla)
	}

	g.Edges = derivarEdges(g.Nodes)
	return g, nil
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

// leerManifiesto lee `arnes.l0.json` de la raíz (§2, D-a firmada). Ausente → (nil, nil):
// modo degradado honesto — el grafo sale sin carriles de fase ni spine y el caller emite el
// check `manifiesto-ausente` rojo. Un manifiesto presente pero corrupto sí es error (%w):
// eso no es degradación, es un arnés roto que no debe cargarse como si nada.
func leerManifiesto(dir string) (*domain.Arnes, error) {
	ruta := filepath.Join(dir, "arnes.l0.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta construida sobre el dir del arnés que el caller eligió cargar (local-first).
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer manifiesto %s: %w", ruta, err)
	}
	var a domain.Arnes
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("manifiesto %s inválido: %w", ruta, err)
	}
	return &a, nil
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
