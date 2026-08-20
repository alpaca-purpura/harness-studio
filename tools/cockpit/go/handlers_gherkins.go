// handlers_gherkins.go — gherkin discovery enriquecido (2026-06-11).
//
// Decisión ratificada por el operador (ROADMAP §3 del cockpit upstream): el
// cockpit NO ejecuta tests (doctrina "el cockpit LEE, no genera" — la ejecución
// real vive en el proceso: /auditor + live-verify). Lo que sí da: cada escenario
// del 01-spec.md con su estado de última verificación (join con el
// gherkin-matrix.md del auditor Phase D) + el comando listo para copiar.
//
// Fuentes:
//   · {story}/01-spec.md            — headings `### SC-N título` + bloques ```gherkin
//                                     + graders `path: "..."` (specs e2e/vitest)
//   · {story}/06-audit/gherkin-matrix.md — tabla | Scenario | Spec ref | Test file | Status | Notes |
package main

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type gherkinScenario struct {
	ID      string   `json:"id"`      // SC-1
	Title   string   `json:"title"`   // happy · click sub-tab navega
	Gherkin string   `json:"gherkin"` // bloque verbatim
	Graders []string `json:"graders"` // paths de test declarados en la spec
	Status  string   `json:"status"`  // celda Status del matrix (vacío si sin auditar)
	Notes   string   `json:"notes"`   // celda Notes del matrix
}

var (
	scHeadingRe = regexp.MustCompile(`(?m)^#{2,4}\s+(SC-\d+)\s*[·:—-]?\s*(.*)$`)
	scIDRe      = regexp.MustCompile(`SC-\d+`)
	graderRe    = regexp.MustCompile(`path:\s*"([^"]+)"`)
)

func handleGherkinStatus(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	storyID := r.URL.Query().Get("story")
	if sistema == "" || storyID == "" {
		writeError(w, 400, "query params sistema y story requeridos", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}
	dir := findStoryDir(sistema, storyID)
	if dir == "" {
		writeError(w, 404, "story no encontrada", map[string]any{"story_id": storyID})
		return
	}

	specRaw, specErr := os.ReadFile(filepath.Join(dir, "01-spec.md"))
	scenarios := []gherkinScenario{}
	if specErr == nil {
		scenarios = parseSpecScenarios(string(specRaw))
	}

	matrixPath := filepath.Join(dir, "06-audit", "gherkin-matrix.md")
	matrixRaw, matrixErr := os.ReadFile(matrixPath)
	if matrixErr == nil {
		joinMatrixStatus(scenarios, string(matrixRaw))
	}

	writeJSON(w, 200, map[string]any{
		"scenarios":     scenarios,
		"spec_exists":   specErr == nil,
		"matrix_exists": matrixErr == nil,
		"sistema":         sistema,
		"story_id":      storyID,
	})
}

// parseSpecScenarios: corta el spec por headings SC-N y extrae de cada sección
// el primer bloque ```gherkin``` + los grader paths.
func parseSpecScenarios(spec string) []gherkinScenario {
	matches := scHeadingRe.FindAllStringSubmatchIndex(spec, -1)
	out := make([]gherkinScenario, 0, len(matches))
	for i, m := range matches {
		id := spec[m[2]:m[3]]
		title := strings.TrimSpace(spec[m[4]:m[5]])
		end := len(spec)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		section := spec[m[1]:end]

		gherkin := ""
		if start := strings.Index(section, "```gherkin"); start >= 0 {
			rest := section[start+len("```gherkin"):]
			if stop := strings.Index(rest, "```"); stop >= 0 {
				gherkin = strings.TrimSpace(rest[:stop])
			}
		}

		var graders []string
		for _, g := range graderRe.FindAllStringSubmatch(section, -1) {
			graders = append(graders, g[1])
		}

		out = append(out, gherkinScenario{
			ID: id, Title: title, Gherkin: gherkin, Graders: graders,
		})
	}
	return out
}

// joinMatrixStatus: filas de la tabla del matrix → status/notes por SC-id.
// Formato: | SC-1 happy · … | `01-spec.md § 10 SC-1` | test file | ✅ COVERED | notes |
func joinMatrixStatus(scenarios []gherkinScenario, matrix string) {
	type row struct{ status, notes string }
	byID := map[string]row{}
	for _, line := range strings.Split(matrix, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 5 {
			continue
		}
		first := strings.TrimSpace(cells[1])
		id := scIDRe.FindString(first)
		if id == "" || strings.HasPrefix(first, "Scenario") || strings.HasPrefix(first, "---") {
			continue
		}
		status := ""
		notes := ""
		// celda Status = la que contiene ✅/⚠️/❌ (la posición exacta varía por matrix)
		for _, c := range cells[2:] {
			t := strings.TrimSpace(c)
			if status == "" && (strings.Contains(t, "✅") || strings.Contains(t, "⚠️") || strings.Contains(t, "❌")) {
				status = t
			}
		}
		if len(cells) >= 6 {
			notes = strings.TrimSpace(cells[len(cells)-2])
		}
		if _, seen := byID[id]; !seen {
			byID[id] = row{status: status, notes: notes}
		}
	}
	for i := range scenarios {
		if r, ok := byID[scenarios[i].ID]; ok {
			scenarios[i].Status = r.status
			scenarios[i].Notes = r.notes
		}
	}
}
