// handlers_stories.go — port de app/api/stories (list + single + PATCH) y transition.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// El ciclo de vida (storyStates, isStoryState) y las transiciones de operador
// (operatorAllowed, stateOwners, evaluateOperatorTransition) viven en gates.go.

var stateRescueRe = regexp.MustCompile(`(?m)^state:[ \t]*([a-z]+)`)

// readCheckpoint replica la normalización del route TS (spread fm + overrides),
// incluyendo el rescate por regex cuando el frontmatter no parsea.
func readCheckpoint(storyDir, sistema string, isArchived bool) map[string]any {
	ckptPath := filepath.Join(storyDir, "checkpoint.md")
	raw, readErr := os.ReadFile(ckptPath)
	if readErr != nil {
		return nil
	}

	doc, parseErr := parseFrontmatter(string(raw))
	if parseErr != nil {
		state := "idea"
		if m := stateRescueRe.FindStringSubmatch(string(raw)); m != nil {
			state = m[1]
		}
		return map[string]any{
			"story_id":        filepath.Base(storyDir),
			"path":            storyDir,
			"sistema":           sistema,
			"release":         nil,
			"cap_target":      nil,
			"cap_change_type": nil,
			"parent_story":    nil,
			"state":           state,
			"body":            string(raw),
			"is_archived":     isArchived,
			"parse_error":     strings.SplitN(parseErr.Error(), "\n", 2)[0],
		}
	}

	story := map[string]any{}
	for k, v := range doc.Frontmatter {
		story[k] = v
	}
	storyID, _ := story["story_id"].(string)
	if storyID == "" {
		storyID = filepath.Base(storyDir)
	}
	story["story_id"] = storyID
	story["path"] = storyDir
	story["sistema"] = sistema
	story["release"] = fmGetOr(doc.Frontmatter, nil, "release")
	story["cap_target"] = fmGetOr(doc.Frontmatter, nil, "cap_target")
	story["cap_change_type"] = fmGetOr(doc.Frontmatter, nil, "cap_change_type")
	story["parent_story"] = fmGetOr(doc.Frontmatter, nil, "parent_story")
	story["state"] = fmGetOr(doc.Frontmatter, "idea", "state")
	story["body"] = doc.Content
	story["is_archived"] = isArchived
	return story
}

func listStoryDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// archivedStoryDirs escanea {sistema}/docs/archive/{year}/stories/*.
func archivedStoryDirs(sistema string) []string {
	var out []string
	years, err := os.ReadDir(archiveRootPath(sistema))
	if err != nil {
		return nil
	}
	for _, y := range years {
		if y.IsDir() {
			out = append(out, listStoryDirs(filepath.Join(archiveRootPath(sistema), y.Name(), "stories"))...)
		}
	}
	return out
}

func handleStoriesList(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	valid := getSelectableSistemas()
	if !containsStr(valid, sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, map[string]any{"valid_sistemas": valid})
		return
	}

	var all []map[string]any
	for _, d := range listStoryDirs(storiesPath(sistema)) {
		if s := readCheckpoint(d, sistema, false); s != nil {
			all = append(all, s)
		}
	}
	for _, d := range archivedStoryDirs(sistema) {
		if s := readCheckpoint(d, sistema, true); s != nil {
			all = append(all, s)
		}
	}

	// Dedup por story_id: archivada gana; colisiones marcadas dup_collision.
	byID := map[string]map[string]any{}
	var order []string
	collisions := map[string]bool{}
	for _, s := range all {
		id, _ := s["story_id"].(string)
		prev, exists := byID[id]
		if !exists {
			byID[id] = s
			order = append(order, id)
			continue
		}
		collisions[id] = true
		prevArch, _ := prev["is_archived"].(bool)
		curArch, _ := s["is_archived"].(bool)
		if curArch && !prevArch {
			byID[id] = s
		}
	}
	stories := []map[string]any{}
	for _, id := range order {
		s := byID[id]
		if collisions[id] {
			s["dup_collision"] = true
		}
		stories = append(stories, s)
	}

	writeJSON(w, 200, map[string]any{"stories": stories})
}

// findStoryDir: live primero, luego archive/{year}/stories/{id}.
func findStoryDir(sistema, storyID string) string {
	live := filepath.Join(storiesPath(sistema), storyID)
	if fileExists(filepath.Join(live, "checkpoint.md")) {
		return live
	}
	years, err := os.ReadDir(archiveRootPath(sistema))
	if err != nil {
		return ""
	}
	for _, y := range years {
		if !y.IsDir() {
			continue
		}
		candidate := filepath.Join(archiveRootPath(sistema), y.Name(), "stories", storyID)
		if fileExists(filepath.Join(candidate, "checkpoint.md")) {
			return candidate
		}
	}
	return ""
}

func handleStorySingle(w http.ResponseWriter, r *http.Request) {
	storyID := r.PathValue("id")
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}

	dir := findStoryDir(sistema, storyID)
	if dir == "" {
		writeError(w, 404, "story no encontrada", map[string]any{"story_id": storyID, "sistema": sistema})
		return
	}

	switch r.Method {
	case http.MethodGet:
		isArchived := strings.Contains(dir, string(filepath.Separator)+"archive"+string(filepath.Separator))
		story := readCheckpoint(dir, sistema, isArchived)
		delete(story, "is_archived") // el single-story del TS no incluye este campo
		writeJSON(w, 200, map[string]any{"story": story})

	case http.MethodPatch:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "body JSON inválido", nil)
			return
		}
		// {field, value} o {fields:{...}}
		fields := map[string]any{}
		if f, ok := body["fields"].(map[string]any); ok {
			fields = f
		} else if fieldName, ok := body["field"].(string); ok {
			fields[fieldName] = body["value"]
		} else {
			writeError(w, 400, "body inválido", nil)
			return
		}

		forbidden := forbiddenStoryFields(fields)
		if len(forbidden) > 0 {
			writeError(w, 403, "campo no editable por el operador", map[string]any{"forbidden_fields": forbidden})
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
		for k, v := range fields {
			doc.Frontmatter[k] = v
		}
		keyOrder := frontmatterKeyOrder(string(rawBytes))
		if err := writeMarkdownWithFrontmatter(ckptPath, doc.Frontmatter, doc.Content, keyOrder); err != nil {
			writeError(w, 500, "error escribiendo checkpoint", map[string]any{"detail": err.Error()})
			return
		}
		isArchived := strings.Contains(dir, string(filepath.Separator)+"archive"+string(filepath.Separator))
		patched := readCheckpoint(dir, sistema, isArchived)
		delete(patched, "is_archived")
		writeJSON(w, 200, map[string]any{"story": patched})

	default:
		writeError(w, 405, "method not allowed", nil)
	}
}

// ── POST /api/transition ────────────────────────────────────────────────────
// La máquina de transiciones (operatorAllowed, stateOwners, evaluateOperatorTransition,
// transitionOwnerNote) es dominio puro → gates.go. Acá solo el transporte HTTP + I/O.

func handleTransition(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema       string `json:"sistema"`
		StoryID     string `json:"storyId"`
		TargetState string `json:"targetState"`
		Reason      string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if body.Sistema == "" || body.StoryID == "" || !isStoryState(body.TargetState) {
		writeError(w, 400, "body inválido", nil)
		return
	}
	if !containsStr(writableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
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
	fromState, _ := doc.Frontmatter["state"].(string)
	if fromState == "" {
		fromState = "idea"
	}

	switch evaluateOperatorTransition(fromState, body.TargetState, body.Reason) {
	case transitionForbidden:
		writeError(w, 403, "transition_forbidden", map[string]any{
			"from": fromState, "to": body.TargetState, "reason": transitionOwnerNote(body.TargetState),
		})
		return
	case transitionNeedsReason:
		writeError(w, 400, "reason requerido (≥10 chars) para esta transition", map[string]any{
			"from": fromState, "to": body.TargetState,
		})
		return
	}

	doc.Frontmatter["state"] = body.TargetState
	// Razón genérica (RN-49): la transición con requiere_razon escribe {target}_reason —
	// el motor ya no conoce parked/dropped por nombre (byte-igual para sdd-default).
	for _, t := range operatorAllowed[fromState] {
		if t.to == body.TargetState && t.requiresReason {
			doc.Frontmatter[body.TargetState+"_reason"] = body.Reason
		}
	}

	keyOrder := frontmatterKeyOrder(string(rawBytes))
	if err := writeMarkdownWithFrontmatter(ckptPath, doc.Frontmatter, doc.Content, keyOrder); err != nil {
		writeError(w, 500, "error escribiendo checkpoint", map[string]any{"detail": err.Error()})
		return
	}
	// El evento nombrado (RN-47): el verbo viaja como DATO en la respuesta — sin log.
	writeJSON(w, 200, map[string]any{
		"ok": true, "newState": body.TargetState,
		"verbo": proceso.verboDe(fromState, body.TargetState),
	})
}
