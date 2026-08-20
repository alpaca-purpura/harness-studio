// handlers_verify.go — gate G del proceso v5: firma del operador (chris_verify.signoff).
//
// La fase G (phase: AWAIT_CHRIS_VERIFY) pide que el operador ejerza el kit live
// ANTES del auditor y firme. Este endpoint escribe SOLO el bloque `chris_verify:`
// del frontmatter con un edit quirúrgico (no re-serializa el checkpoint entero:
// los comentarios y el orden del resto del frontmatter se preservan — lección
// aprendida del writeCheckpoint lossy del cockpit upstream).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var validVerifyResults = map[string]bool{
	"SATISFIED":                 true,
	"SATISFIED_WITH_FOLLOWUPS":  true,
	"REJECTED":                  true,
}

type operatorVerifyBody struct {
	Sistema   string `json:"sistema"`
	StoryID string `json:"story_id"`
	Result  string `json:"result"`
	Notes   string `json:"notes"`
}

func handleOperatorVerify(w http.ResponseWriter, r *http.Request) {
	var body operatorVerifyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if body.Sistema == "" || body.StoryID == "" {
		writeError(w, 400, "sistema y story_id requeridos", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}
	if !validVerifyResults[body.Result] {
		writeError(w, 400, "result inválido (SATISFIED | SATISFIED_WITH_FOLLOWUPS | REJECTED)", nil)
		return
	}

	dir := findStoryDir(body.Sistema, body.StoryID)
	if dir == "" {
		writeError(w, 404, "story no encontrada", map[string]any{"story_id": body.StoryID, "sistema": body.Sistema})
		return
	}

	ckptPath := filepath.Join(dir, "checkpoint.md")
	rawBytes, err := os.ReadFile(ckptPath)
	if err != nil {
		writeError(w, 500, "error leyendo checkpoint", map[string]any{"detail": err.Error()})
		return
	}
	doc, parseErr := parseFrontmatter(string(rawBytes))
	if parseErr != nil {
		writeError(w, 500, "checkpoint con frontmatter inválido", map[string]any{"detail": parseErr.Error()})
		return
	}

	// El gate G solo aplica con la story en developed (la firma ANTES del auditor).
	if state, _ := doc.Frontmatter["state"].(string); state != "developed" {
		writeError(w, 409, "el signoff del gate G solo aplica con la story en developed (estado actual: "+fmt.Sprint(doc.Frontmatter["state"])+")", nil)
		return
	}

	// Preservar required/rounds existentes; rellenar signoff.
	cv, _ := doc.Frontmatter["chris_verify"].(map[string]any)
	if cv == nil {
		cv = map[string]any{"required": true, "rounds": []any{}}
	}
	signoff, _ := cv["signoff"].(map[string]any)
	if signoff == nil {
		signoff = map[string]any{}
	}
	openItems := signoff["open_items"]
	if openItems == nil {
		openItems = []any{}
	}
	cv["signoff"] = map[string]any{
		"by":         "operador",
		"date":       time.Now().Format("2006-01-02"),
		"result":     body.Result,
		"notes":      body.Notes,
		"open_items": openItems,
	}

	block, err := yamlBlockFor("chris_verify", cv)
	if err != nil {
		writeError(w, 500, "error serializando chris_verify", map[string]any{"detail": err.Error()})
		return
	}
	updated, err := replaceFrontmatterBlock(string(rawBytes), "chris_verify", block)
	if err != nil {
		writeError(w, 500, "error editando frontmatter", map[string]any{"detail": err.Error()})
		return
	}
	if err := writeFileAtomic(ckptPath, updated); err != nil {
		writeError(w, 500, "error escribiendo checkpoint", map[string]any{"detail": err.Error()})
		return
	}

	isArchived := strings.Contains(dir, string(filepath.Separator)+"archive"+string(filepath.Separator))
	patched := readCheckpoint(dir, body.Sistema, isArchived)
	delete(patched, "is_archived")
	writeJSON(w, 200, map[string]any{"story": patched})
}

// yamlBlockFor serializa {key: value} como bloque YAML indentado a 2 espacios,
// sin newline final (listo para insertar como línea(s) del frontmatter).
func yamlBlockFor(key string, value any) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(map[string]any{key: value}); err != nil {
		return "", err
	}
	_ = enc.Close()
	return strings.TrimRight(buf.String(), "\n"), nil
}

// replaceFrontmatterBlock reemplaza SOLO el bloque `key:` (línea + continuación
// indentada) dentro del frontmatter, preservando el resto byte a byte.
// Si el key no existe, inserta el bloque antes del `---` de cierre.
func replaceFrontmatterBlock(content, key, block string) (string, error) {
	lines := strings.Split(content, "\n")
	fmEnd := 0
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			fmEnd = i
			break
		}
	}
	if fmEnd == 0 || !strings.HasPrefix(lines[0], "---") {
		return "", fmt.Errorf("no se encontró frontmatter delimitado por ---")
	}

	blockStart := -1
	for i := 1; i < fmEnd; i++ {
		if strings.HasPrefix(lines[i], key+":") {
			blockStart = i
			break
		}
	}

	var out []string
	if blockStart < 0 {
		out = append(out, lines[:fmEnd]...)
		out = append(out, block)
		out = append(out, lines[fmEnd:]...)
		return strings.Join(out, "\n"), nil
	}

	blockEnd := blockStart + 1
	for blockEnd < fmEnd {
		line := lines[blockEnd]
		if line != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			break
		}
		blockEnd++
	}
	out = append(out, lines[:blockStart]...)
	out = append(out, block)
	out = append(out, lines[blockEnd:]...)
	return strings.Join(out, "\n"), nil
}
