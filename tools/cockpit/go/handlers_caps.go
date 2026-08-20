// handlers_caps.go — port de app/api/capabilities (list, single GET/PATCH).
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// readCapability replica lib/cap-ledger.ts readCapability (defaults defensivos).
func readCapability(absPath string) (map[string]any, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	doc, parseErr := parseFrontmatter(string(raw))
	if parseErr != nil {
		return nil, parseErr
	}
	fm := doc.Frontmatter
	// Un .yaml que abre `---` y NUNCA cierra es un DOCUMENTO YAML, no un frontmatter:
	// parseFrontmatter devuelve mapa vacío SIN error y los defensive defaults de abajo
	// terminan fabricando una capability (`capability_id: ""`, `status: "live"`) que no
	// existe en el archivo. Inventar es peor que fallar. El doctor del repo
	// (scripts/cap_doctor.py::_load) ya trata las dos formas como válidas; acá se hace
	// lo mismo para que cockpit y doctor coincidan en QUÉ leen.
	if len(fm) == 0 {
		if doc2, err2 := parseYAMLMap(raw); err2 == nil && len(doc2) > 0 {
			fm = doc2
		}
	}

	cap := map[string]any{
		"capability_id":            fmGetOr(fm, "", "capability_id"),
		"module":                   fmGetOr(fm, "", "module"),
		"slug":                     fmGetOr(fm, "", "slug"),
		"status":                   fmGetOr(fm, "live", "status"),
		"license":                  fmGetOr(fm, "sistema-local", "license"),
		"created_in_story":         fmGetOr(fm, "", "created_in_story", "story_introduced"),
		"created_date":             fmGetOr(fm, "", "created_date", "date_introduced"),
		"last_modified":            fmGetOr(fm, "", "last_modified", "date_updated", "created_date"),
		"package_version":          fmGetOr(fm, nil, "package_version"),
		"package_path":             fmGetOr(fm, nil, "package_path"),
		"architecture_pattern":     fmGetOr(fm, nil, "architecture_pattern"),
		"hipaa_lite_overlay":       fmGetOr(fm, false, "hipaa_lite_overlay"),
		"parent_cap":               fmGetOr(fm, nil, "parent_cap", "extends_capability"),
		"derives_capabilities":     fmGetOr(fm, []any{}, "derives_capabilities"),
		"change_log":               fmGetOr(fm, []any{}, "change_log"),
		"date_introduced":          fmGetOr(fm, nil, "date_introduced"),
		"story_introduced":         fmGetOr(fm, nil, "story_introduced"),
		"date_updated":             fmGetOr(fm, nil, "date_updated"),
		"extends_capability":       fmGetOr(fm, nil, "extends_capability"),
		"tech_module":              fmGetOr(fm, nil, "tech_module", "module"),
		"agent_owner":              fmGetOr(fm, nil, "agent_owner"),
		"functional_area":          fmGetOr(fm, nil, "functional_area"),
		"user_visible":             fmGetOr(fm, true, "user_visible"),
		"nature":                   fmGetOr(fm, nil, "nature"),
		// `name` es el fallback del vendored ArnesIA: sus capabilities son GENERADAS por
		// scripts/capabilities_to_yaml.py y llevan `name`, no `user_facing_name`. Sin este
		// coalesce la UI pinta el slug (`abrir-shell`) en vez del título legible
		// («Abrir shell (`open`)»). Coalesce, no rename: un workspace que sí declare
		// user_facing_name sigue ganando.
		"user_facing_name":         fmGetOr(fm, nil, "user_facing_name", "name"),
		"user_facing_description":  fmGetOr(fm, nil, "user_facing_description"),
		"access":                   fmGetOr(fm, nil, "access"),
		"scenarios":                fmGetOr(fm, nil, "scenarios"),
		"business_rules":           fmGetOr(fm, nil, "business_rules"),
		"related_capabilities":     fmGetOr(fm, nil, "related_capabilities"),
		"dev_preview":              fmGetOr(fm, nil, "dev_preview"),
		"superseded_by":            fmGetOr(fm, nil, "superseded_by"),
		"body":                     doc.Content,
		"path":                     absPath,
	}
	return cap, nil
}

// globCapFiles: capabilities/*/*.yaml (un nivel de módulo).
func globCapFiles(baseDir string) []string {
	var out []string
	modules, err := os.ReadDir(baseDir)
	if err != nil {
		return nil
	}
	for _, m := range modules {
		if !m.IsDir() {
			continue
		}
		files, err := os.ReadDir(filepath.Join(baseDir, m.Name()))
		if err != nil {
			continue
		}
		for _, f := range files {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".yaml" {
				out = append(out, filepath.Join(baseDir, m.Name(), f.Name()))
			}
		}
	}
	sort.Strings(out)
	return out
}

func handleCapabilitiesList(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	// Lectura: getSelectableSistemas (incluye platform), no getSistemas (I-40).
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}

	capabilities := []map[string]any{}
	for _, abs := range globCapFiles(capabilitiesPath(sistema)) {
		if c, err := readCapability(abs); err == nil {
			capabilities = append(capabilities, c)
		}
	}
	writeJSON(w, 200, map[string]any{"capabilities": capabilities})
}

func capYamlPath(sistema, module, slug string) string {
	return filepath.Join(capabilitiesPath(sistema), module, slug+".yaml")
}

// capsSonGenerados: ¿este workspace DERIVA sus capabilities de un generador?
//
// La señal es la presencia del generador mismo en el workspace — no una lista de
// nombres de proyecto hardcodeada. Si el archivo existe, los .yaml de capabilities son
// su salida y editarlos a mano es escribir sobre algo que la próxima corrida pisa.
// Un workspace sin generador (vitalia upstream) sigue con el PATCH habilitado.
func capsSonGenerados() bool {
	for _, root := range workspaceRoots() {
		if fileExists(filepath.Join(root, "scripts", "capabilities_to_yaml.py")) {
			return true
		}
	}
	return false
}

func handleCapabilitySingle(w http.ResponseWriter, r *http.Request) {
	module := r.PathValue("module")
	capSlug := r.PathValue("cap")
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	absPath := capYamlPath(sistema, module, capSlug)

	switch r.Method {
	case http.MethodGet:
		// Lectura: getSelectableSistemas (incluye platform), no getSistemas (I-40).
		if !containsStr(getSelectableSistemas(), sistema) {
			writeError(w, 400, "sistema desconocido: "+sistema, nil)
			return
		}
		capability, err := readCapability(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				writeError(w, 404, "capability no encontrada", map[string]any{
					"sistema": sistema, "module": module, "cap": capSlug,
				})
				return
			}
			writeError(w, 500, "error leyendo capability", map[string]any{"detail": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"capability": capability})

	case http.MethodPatch:
		// Vendored ArnesIA: las capabilities de este workspace son ARTEFACTOS GENERADOS
		// (`scripts/capabilities_to_yaml.py`, regla R4 «estado generado — no teclear»), y
		// su enum de status es propio (`vivo · vivo·nc · parcial · stub`) — incompatible
		// con validCapStatuses (`live/beta/deprecated/sunset`, gates.go). Un PATCH acá
		// escribiría a mano un archivo que el próximo `make cap-*` pisa sin avisar.
		//
		// Se bloquea EXPLÍCITO en vez de dejar que caiga en «sistema desconocido»: el
		// operador merece saber por qué no puede editar, no un 400 que miente sobre la causa.
		// Deliberadamente ANTES del gate de sistema — el motivo real es el artefacto, no el scope.
		if capsSonGenerados() {
			writeError(w, 409, "las capabilities de este workspace son artefactos GENERADOS — no se editan por PATCH", map[string]any{
				"reason": "regla R4 (estado-generado): el status sale del código y los checks, no del teclado. " +
					"Para cambiarlo, cambiá el código/los tests y regenerá con `python scripts/capabilities_to_yaml.py`.",
				"cap": module + "/" + capSlug,
			})
			return
		}
		// Escritura: rechaza el pseudo-sistema platform (getSistemas, no selectable).
		if !containsStr(getSistemas(), sistema) {
			writeError(w, 400, "sistema desconocido: "+sistema, nil)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "body JSON inválido", nil)
			return
		}
		forbidden := forbiddenCapFields(body)
		if len(forbidden) > 0 {
			writeError(w, 403, "edición directa de scenarios, change_log y campos estructurales prohibida", map[string]any{
				"forbidden_fields": forbidden,
				"reason":           "modificar scenarios o change_log requiere crear una story con cap_change_type. Usa /api/extend-cap o el botón \"Extender\" del cockpit.",
			})
			return
		}

		status, _ := body["status"].(string)
		reason, _ := body["reason"].(string)
		if !isValidCapStatus(status) || len(reason) < capStatusReasonMinLen {
			writeError(w, 400, "body inválido", nil)
			return
		}

		rawBytes, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				writeError(w, 404, "capability no encontrada", map[string]any{
					"sistema": sistema, "module": module, "cap": capSlug,
				})
				return
			}
			writeError(w, 500, "error leyendo capability", map[string]any{"detail": err.Error()})
			return
		}
		doc, parseErr := parseFrontmatter(string(rawBytes))
		if parseErr != nil {
			writeError(w, 500, "capability con frontmatter inválido", map[string]any{"detail": parseErr.Error()})
			return
		}

		today := time.Now().UTC().Format("2006-01-02")
		doc.Frontmatter["status"] = status
		doc.Frontmatter["last_modified"] = today

		entry := map[string]any{
			"date":    today,
			"type":    "fix",
			"summary": reason,
			"by":      "operador",
		}
		if cl, ok := doc.Frontmatter["change_log"].([]any); ok {
			doc.Frontmatter["change_log"] = append(cl, entry)
		} else {
			doc.Frontmatter["change_log"] = []any{entry}
		}

		keyOrder := frontmatterKeyOrder(string(rawBytes))
		if err := writeMarkdownWithFrontmatter(absPath, doc.Frontmatter, doc.Content, keyOrder); err != nil {
			writeError(w, 500, "error escribiendo capability", map[string]any{"detail": err.Error()})
			return
		}
		capability, _ := readCapability(absPath)
		writeJSON(w, 200, map[string]any{"capability": capability})

	default:
		writeError(w, 405, "method not allowed", nil)
	}
}
