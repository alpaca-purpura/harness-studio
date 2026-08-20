// handlers_learnings.go — Learnings v2 (2026-06-11, pedido operador):
// el carril L2 del CIL deja de ser un cementerio de archivos.
//
//   · GET /api/learnings    — las 3 fuentes (sistema + transversal + tooling) con
//     ESTADO de ciclo de vida derivado: pending (espera decisión) · applied ·
//     promoted · wont-apply · reference (promotable: no — solo consulta).
//   · PATCH /api/learnings  — escribe `applied:` en el frontmatter (edit
//     quirúrgico; si el archivo no tiene frontmatter, lo crea).
//
// Doctrina: un learning capturado NO es estado final — muere en una de 4
// salidas (gate/hook > rule/skill > proposal core > referencia consciente).
// El cockpit muestra el stock "esperando decisión" para que el loop se VEA.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var validApplied = map[string]bool{
	"pending": true, "applied": true, "promoted": true, "wont-apply": true,
}

// learningSources: dirs a escanear para un sistema + label de origen.
func learningSources(sistema string) []struct{ dir, source string } {
	root, err := getWorkspaceRoot()
	if err != nil {
		return nil
	}
	out := []struct{ dir, source string }{
		{learningsPath(sistema), "sistema"},
		{filepath.Join(root, "docs", "learnings"), "transversal"},
		{filepath.Join(root, "docs", "learnings", "tooling"), "tooling"},
	}
	return out
}

func handleLearningsV2(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listLearningsV2(w, r)
	case http.MethodPatch:
		patchLearning(w, r)
	default:
		writeError(w, 405, "method not allowed", nil)
	}
}

func listLearningsV2(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}
	root, err := getWorkspaceRoot()
	if err != nil {
		writeError(w, 500, "workspace root no resuelto", nil)
		return
	}

	seen := map[string]bool{}
	learnings := []map[string]any{}
	for _, src := range learningSources(sistema) {
		entries, readErr := os.ReadDir(src.dir)
		if readErr != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || strings.EqualFold(e.Name(), "README.md") {
				continue
			}
			abs := filepath.Join(src.dir, e.Name())
			if seen[abs] {
				continue
			}
			seen[abs] = true
			doc, parseErr := readMarkdownWithFrontmatter(abs)
			if parseErr != nil {
				continue
			}

			preview := headingRe.ReplaceAllString(doc.Content, "")
			preview = strings.TrimSpace(preview)
			if parts := strings.SplitN(preview, "\n\n", 2); len(parts) > 0 {
				preview = parts[0]
			}
			preview = strings.ReplaceAll(preview, "\n", " ")
			if len([]rune(preview)) > 240 {
				preview = string([]rune(preview)[:240])
			}

			rel, _ := filepath.Rel(root, abs)
			entry := map[string]any{
				"slug":     strings.TrimSuffix(e.Name(), ".md"),
				"path":     abs,
				"rel_path": rel,
				"source":   src.source,
				"preview":  preview,
			}
			for _, k := range []string{"title", "date", "type", "sistema", "promotable", "applied"} {
				if v, ok := doc.Frontmatter[k].(string); ok {
					entry[k] = v
				}
			}
			for _, k := range []string{"sistemas_affected", "tags"} {
				if v, ok := doc.Frontmatter[k].([]any); ok {
					entry[k] = v
				}
			}

			// Estado derivado: applied explícito manda; promotable:no = referencia
			// (no espera decisión); el resto está PENDIENTE de triage.
			status, _ := entry["applied"].(string)
			if status == "" {
				if p, _ := entry["promotable"].(string); p == "no" {
					status = "reference"
				} else {
					status = "pending"
				}
			}
			entry["status"] = status

			learnings = append(learnings, entry)
		}
	}

	sort.SliceStable(learnings, func(i, j int) bool {
		di, _ := learnings[i]["date"].(string)
		dj, _ := learnings[j]["date"].(string)
		return di > dj
	})

	writeJSON(w, 200, map[string]any{"learnings": learnings})
}

// patchLearning: {path: relPath, applied: pending|applied|promoted|wont-apply}.
// Solo archivos .md dentro de un dir de learnings (validación de containment).
func patchLearning(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path    string `json:"path"`
		Applied string `json:"applied"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if !validApplied[body.Applied] {
		writeError(w, 400, "applied inválido (pending | applied | promoted | wont-apply)", nil)
		return
	}
	abs := resolveSafePath(body.Path)
	if abs == "" || !strings.HasSuffix(abs, ".md") ||
		!strings.Contains(abs, string(filepath.Separator)+"learnings"+string(filepath.Separator)) {
		writeError(w, 403, "path fuera de un dir de learnings", map[string]any{"path": body.Path})
		return
	}

	raw, err := os.ReadFile(abs)
	if err != nil {
		writeError(w, 404, "learning no existe", map[string]any{"path": body.Path})
		return
	}
	content := string(raw)

	var updated string
	if strings.HasPrefix(content, "---\n") {
		updated, err = replaceFrontmatterBlock(content, "applied", "applied: "+body.Applied)
		if err != nil {
			writeError(w, 500, "error editando frontmatter", map[string]any{"detail": err.Error()})
			return
		}
	} else {
		// Learning sin frontmatter → crearlo mínimo (no tocar el body).
		updated = "---\napplied: " + body.Applied + "\n---\n" + content
	}

	if err := writeFileAtomic(abs, updated); err != nil {
		writeError(w, 500, "error escribiendo learning", map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true, "path": body.Path, "applied": body.Applied})
}
