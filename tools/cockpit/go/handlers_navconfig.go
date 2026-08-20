// handlers_navconfig.go — nav configurable por workspace (pedido operador 2026-06-11):
// poder agregar/quitar/reordenar módulos (tabs del sidebar) sin tocar la UI.
//
// Config opcional en la raíz del workspace: `cockpit.config.yaml`
//
//	nav:
//	  - roadmap
//	  - board
//	  - map
//	  - drift
//	  - harness
//
// La lista define QUÉ tabs se muestran y EN QUÉ ORDEN. Sin archivo (o sin key
// `nav`) la UI muestra el set completo default. IDs desconocidos se devuelven
// igual (la UI ignora los que no conoce — forward-compat para tabs futuros).
//
// Per-board en multi (I-51): el allowlist es PER-BOARD, no global. El front pasa
// ?sistema=<key navegable>; resolveSistema la traduce a la raíz del repo de ESE
// board (incl. cross-repo vía systemWorkspaces) y se lee su cockpit.config.yaml.
// Así cada board cura su propio nav sistema-scoped. (Las pestañas transversales
// —Negocio/Harness— quedan always-on en la UI: no cuelgan de un board.) Sin
// ?sistema= o en single-mode → getWorkspaceRoot() como siempre (caso cliente,
// intacto). Antes el root salía del cwd de arranque del daemon (global, frágil).
package main

import (
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func handleNavConfig(w http.ResponseWriter, r *http.Request) {
	// Per-board (I-51): en multi, el board activo (?sistema=) decide de qué repo se
	// lee el nav. Key que no resuelve → default (nav:null), NUNCA error: el sidebar
	// debe renderizar igual. Sin param o single-mode → getWorkspaceRoot() (intacto).
	var root string
	if sistema := r.URL.Query().Get("sistema"); sistema != "" && isMultiMode() {
		rt, _, err := resolveSistema(sistema)
		if err != nil {
			writeJSON(w, 200, map[string]any{"nav": nil, "path": "", "exists": false})
			return
		}
		root = rt
	} else {
		rt, err := getWorkspaceRoot()
		if err != nil {
			writeError(w, 500, "workspace root no resuelto: "+err.Error(), nil)
			return
		}
		root = rt
	}
	cfgPath := filepath.Join(root, "cockpit.config.yaml")

	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		writeJSON(w, 200, map[string]any{"nav": nil, "path": cfgPath, "exists": false})
		return
	}

	var cfg struct {
		Nav []string `yaml:"nav"`
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		writeError(w, 500, "cockpit.config.yaml mal formado: "+err.Error(), map[string]any{"path": cfgPath})
		return
	}
	var nav any
	if len(cfg.Nav) > 0 {
		nav = cfg.Nav
	}
	writeJSON(w, 200, map[string]any{"nav": nav, "path": cfgPath, "exists": true})
}
