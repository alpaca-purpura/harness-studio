// Package loader — reconocedores mecánicos añadidos en la auditoría colateral HS-16 (§3 de
// nomenclatura-arnes.md): command, output-style, mcp, settings, statusline. Los cinco emiten
// nodos "soporte" (sin banda — §4 regla 4, «resto→soporte») salvo los hooks de forma
// instalada, que reusan nodosDeHooks (loader.go) hacia la banda Guardia. Ninguno tiene
// contract/caja: esa semántica es exclusiva de `skill` (§4 regla 2).
package loader

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// reconocerArchivosSoporte escanea un directorio plano de `<id>.md` (commands/, output-styles/
// — §3) y emite un nodo soporte por archivo. A diferencia de skills, el frontmatter es
// OPCIONAL (muchos comandos son prosa simple sin metadata) — su ausencia NUNCA vuelve el nodo
// no-reconocido, solo pierde el nombre display bonito (cae al id). Un dir suelto o un archivo
// que no sea `.md` bajo esta carpeta sí es no-reconocido visible (§4.5): no matchea ningún
// reconocedor de esta celda.
func reconocerArchivosSoporte(elementos, subdir string, clase domain.Clase) ([]domain.Box, error) {
	dir := filepath.Join(elementos, subdir)
	entradas, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("escanear %s: %w", dir, err)
	}

	var nodos []domain.Box
	for _, e := range entradas {
		nombreArchivo := e.Name()
		if strings.HasPrefix(nombreArchivo, ".") {
			continue
		}
		if e.IsDir() || !strings.HasSuffix(nombreArchivo, ".md") {
			nodos = append(nodos, nodoNoReconocido(nombreArchivo, rutaEstampada(dir, nombreArchivo)))
			continue
		}
		id := strings.TrimSuffix(nombreArchivo, ".md")
		fuente := rutaEstampada(dir, nombreArchivo)
		nombre := id
		if b, rerr := os.ReadFile(filepath.Join(dir, nombreArchivo)); rerr == nil { //nolint:gosec // G304: path bajo el dir del arnés que el caller eligió cargar.
			if fm, ferr := frontmatter(b); ferr == nil {
				nombre = nombreDe(fm, id)
			}
		}
		nodos = append(nodos, domain.Box{
			ID: id, Clase: clase, Nombre: nombre,
			FuentePath: fuente, Procedencia: domain.ProcDeclarado,
		})
	}
	return nodos, nil
}

// mcpJSON es la forma de `.mcp.json` (subset que el reconocedor necesita): un mapa de
// nombre-de-servidor → config (la config en sí no importa al grafo, solo su identidad).
type mcpJSON struct {
	MCPServers map[string]json.RawMessage `json:"mcpServers"`
}

// reconocerMCP escanea `.mcp.json` (§3, fila mcp — vive en la RAÍZ del arnés en AMBAS formas,
// no bajo `.claude/`): un server declarado = un nodo soporte. Archivo presente pero sin
// servers o no parseable → nodo no-reconocido visible (§4.5).
func reconocerMCP(dir string) ([]domain.Box, error) {
	ruta := filepath.Join(dir, ".mcp.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: path bajo el dir del arnés que el caller eligió cargar.
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", ruta, err)
	}
	fuente := rutaEstampada(dir, ".mcp.json")

	var m mcpJSON
	if err := json.Unmarshal(b, &m); err != nil || len(m.MCPServers) == 0 {
		return []domain.Box{nodoNoReconocido("mcp", fuente)}, nil //nolint:nilerr // reconciliación honesta: roto/vacío = visible, jamás abortar el grafo.
	}

	nombres := make([]string, 0, len(m.MCPServers))
	for nombre := range m.MCPServers {
		nombres = append(nombres, nombre)
	}
	sort.Strings(nombres) // orden determinista: el JSON map no tiene orden.

	nodos := make([]domain.Box, 0, len(nombres))
	for _, nombre := range nombres {
		nodos = append(nodos, domain.Box{
			ID: "mcp-" + nombre, Clase: domain.ClaseMCP, Nombre: nombre,
			FuentePath: fuente, Procedencia: domain.ProcDeclarado,
		})
	}
	return nodos, nil
}

// settingsJSON es el subset de settings.json (plugin `settings.json` / instalado
// `.claude/settings.json`, §3) que el loader lee: `hooks` con la MISMA forma que
// hooks/hooks.json (la Guardia de la forma instalada) y la presencia de `statusLine`
// (custom statusline, Claude Code nativo).
type settingsJSON struct {
	Hooks map[string][]struct {
		Matcher string `json:"matcher"`
	} `json:"hooks"`
	StatusLine json.RawMessage `json:"statusLine"`
}

// reconocerSettings escanea settings.json (§3, filas settings/statusline/hook forma-instalada
// — las tres celdas viven en el MISMO archivo): su sola presencia es el nodo `settings`
// (soporte); una clave `statusLine` presente suma el nodo `statusline` (soporte); su clave
// `hooks` reusa nodosDeHooks (loader.go) — la Guardia de la forma instalada. settings.json
// presente pero no parseable → nodo no-reconocido visible (§4.5), jamás descarte silencioso.
func reconocerSettings(elementos string) ([]domain.Box, error) {
	ruta := filepath.Join(elementos, "settings.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: path bajo el dir del arnés que el caller eligió cargar.
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", ruta, err)
	}
	fuente := rutaEstampada(elementos, "settings.json")

	var s settingsJSON
	if err := json.Unmarshal(b, &s); err != nil {
		return []domain.Box{nodoNoReconocido("settings", fuente)}, nil //nolint:nilerr // reconciliación honesta: roto = visible, jamás abortar el grafo.
	}

	nodos := []domain.Box{{
		ID: "settings", Clase: domain.ClaseSettings, Nombre: "settings",
		FuentePath: fuente, Procedencia: domain.ProcDeclarado,
	}}
	if len(s.StatusLine) > 0 {
		nodos = append(nodos, domain.Box{
			ID: "statusline", Clase: domain.ClaseStatusline, Nombre: "statusline",
			FuentePath: fuente, Procedencia: domain.ProcDeclarado,
		})
	}
	if len(s.Hooks) > 0 {
		nodos = append(nodos, nodosDeHooks(hooksJSON{Hooks: s.Hooks}, fuente)...)
	}
	return nodos, nil
}
