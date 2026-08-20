// handlers_releases.go — port de app/api/releases (GET/POST/PUT/DELETE) + merge-release.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var releaseIDRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// releaseKeyOrder fija la serialización (paridad con el objeto JS del route TS).
var releaseKeyOrder = []string{
	"release_id", "sistema", "name", "description", "status", "target_date",
	"shipped_date", "order", "created_at", "created_by", "stories",
	"verified_by", "verified_at", "verification_note",
	"production_status", "production_version", "production_scheduled_at",
	"deployed_at", "release_branch", "maps_legacy_outcome", "maps_legacy_phase",
}

// readRelease replica release-resolver.readReleaseFromPath (defaults).
func readRelease(absPath string) (map[string]any, error) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}
	doc, parseErr := parseFrontmatter(string(raw))
	if parseErr != nil {
		return nil, parseErr
	}
	fm := doc.Frontmatter
	base := strings.TrimSuffix(filepath.Base(absPath), ".yaml")

	rel := map[string]any{
		"release_id":              fmGetOr(fm, base, "release_id"),
		"sistema":                   fmGetOr(fm, "", "sistema"),
		"name":                    fmGetOr(fm, "", "name"),
		"description":             fmGetOr(fm, "", "description"),
		"status":                  fmGetOr(fm, "planning", "status"),
		"target_date":             fmGetOr(fm, nil, "target_date"),
		"shipped_date":            fmGetOr(fm, nil, "shipped_date"),
		"order":                   fmGetOr(fm, 0, "order"),
		"created_at":              fmGetOr(fm, "", "created_at"),
		"created_by":              fmGetOr(fm, "operador", "created_by"),
		"stories":                 fmGetOr(fm, []any{}, "stories"),
		"verified_by":             fmGetOr(fm, nil, "verified_by"),
		"verified_at":             fmGetOr(fm, nil, "verified_at"),
		"verification_note":       fmGetOr(fm, nil, "verification_note"),
		"production_status":       fmGetOr(fm, nil, "production_status"),
		"production_version":      fmGetOr(fm, nil, "production_version"),
		"production_scheduled_at": fmGetOr(fm, nil, "production_scheduled_at"),
		"deployed_at":             fmGetOr(fm, nil, "deployed_at"),
		"release_branch":          fmGetOr(fm, nil, "release_branch"),
		"maps_legacy_outcome":     fmGetOr(fm, nil, "maps_legacy_outcome"),
		"maps_legacy_phase":       fmGetOr(fm, nil, "maps_legacy_phase"),
		"body":                    doc.Content,
		"path":                    absPath,
	}
	return rel, nil
}

func writeRelease(rel map[string]any) error {
	absPath, _ := rel["path"].(string)
	body, _ := rel["body"].(string)
	fm := map[string]any{}
	for k, v := range rel {
		if k == "body" || k == "path" {
			continue
		}
		fm[k] = v
	}
	serialized, err := stringifyFrontmatter(fm, body, releaseKeyOrder)
	if err != nil {
		return err
	}
	return writeFileAtomic(absPath, serialized)
}

func listReleases(sistema string) []map[string]any {
	dir := releasesPath(sistema)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]any{}
	}
	releases := []map[string]any{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		if rel, err := readRelease(filepath.Join(dir, e.Name())); err == nil {
			releases = append(releases, rel)
		}
	}
	sort.SliceStable(releases, func(i, j int) bool {
		return toFloat(releases[i]["order"]) < toFloat(releases[j]["order"])
	})
	return releases
}

func toFloat(v any) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func handleReleases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		sistema := r.URL.Query().Get("sistema")
		if sistema == "" {
			writeError(w, 400, "query param \"sistema\" requerido", nil)
			return
		}
		if !containsStr(getSelectableSistemas(), sistema) {
			writeError(w, 400, "sistema desconocido: "+sistema, nil)
			return
		}
		writeJSON(w, 200, map[string]any{"releases": listReleases(sistema)})

	case http.MethodPost:
		var body struct {
			ReleaseID   string   `json:"release_id"`
			Sistema       string   `json:"sistema"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			TargetDate  *string  `json:"target_date"`
			Order       *float64 `json:"order"`
			Stories     []string `json:"stories"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "body JSON inválido", nil)
			return
		}
		if !releaseIDRe.MatchString(body.ReleaseID) || body.Name == "" || body.Sistema == "" {
			writeError(w, 400, "body inválido", nil)
			return
		}
		if !containsStr(writableSistemas(), body.Sistema) {
			writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
			return
		}

		absPath := filepath.Join(releasesPath(body.Sistema), body.ReleaseID+".yaml")
		if fileExists(absPath) {
			writeError(w, 409, "release ya existe", map[string]any{"release_id": body.ReleaseID})
			return
		}

		var targetDate any
		if body.TargetDate != nil {
			targetDate = *body.TargetDate
		}
		var order any = 0
		if body.Order != nil {
			order = *body.Order
		}
		stories := []any{}
		for _, s := range body.Stories {
			stories = append(stories, s)
		}

		rel := map[string]any{
			"release_id": body.ReleaseID, "sistema": body.Sistema,
			"name": body.Name, "description": body.Description,
			"status": "planning", "target_date": targetDate, "shipped_date": nil,
			"order": order, "created_at": nowISO(), "created_by": "operador",
			"stories": stories,
			"verified_by": nil, "verified_at": nil, "verification_note": nil,
			"production_status": "not_deployed", "production_version": nil,
			"production_scheduled_at": nil, "deployed_at": nil,
			"release_branch": nil, "maps_legacy_outcome": nil, "maps_legacy_phase": nil,
			"body": "\n## Stories\n\n## Notas\n", "path": absPath,
		}
		if err := writeRelease(rel); err != nil {
			writeError(w, 500, "error escribiendo release", map[string]any{"detail": err.Error()})
			return
		}
		created, _ := readRelease(absPath)
		writeJSON(w, 200, map[string]any{"release": created})

	case http.MethodPut:
		id := r.URL.Query().Get("id")
		sistema := r.URL.Query().Get("sistema")
		if id == "" || sistema == "" {
			writeError(w, 400, "query params \"id\" y \"sistema\" requeridos", nil)
			return
		}
		if !containsStr(writableSistemas(), sistema) {
			writeError(w, 400, "sistema desconocido: "+sistema, nil)
			return
		}
		absPath := filepath.Join(releasesPath(sistema), id+".yaml")
		rel, err := readRelease(absPath)
		if err != nil {
			writeError(w, 404, "release no encontrado", map[string]any{"release_id": id, "sistema": sistema})
			return
		}

		if status, _ := rel["status"].(string); status == "shipped" {
			writeError(w, 403, "no se puede editar un release shipped", map[string]any{"release_id": id})
			return
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "body JSON inválido", nil)
			return
		}
		readOnly := map[string]bool{"status": true, "shipped_date": true, "release_id": true, "sistema": true, "created_at": true, "created_by": true}
		var sentReadOnly []string
		for k := range body {
			if readOnly[k] {
				sentReadOnly = append(sentReadOnly, k)
			}
		}
		if len(sentReadOnly) > 0 {
			writeError(w, 403, "algunos campos no son editables desde el cockpit", map[string]any{"read_only_fields": sentReadOnly})
			return
		}
		editable := map[string]bool{"name": true, "description": true, "target_date": true, "order": true, "stories": true}
		for k, v := range body {
			if editable[k] {
				rel[k] = v
			}
		}
		if err := writeRelease(rel); err != nil {
			writeError(w, 500, "error escribiendo release", map[string]any{"detail": err.Error()})
			return
		}
		updated, _ := readRelease(absPath)
		writeJSON(w, 200, map[string]any{"release": updated})

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		sistema := r.URL.Query().Get("sistema")
		if id == "" || sistema == "" {
			writeError(w, 400, "query params \"id\" y \"sistema\" requeridos", nil)
			return
		}
		if !containsStr(writableSistemas(), sistema) {
			writeError(w, 400, "sistema desconocido: "+sistema, nil)
			return
		}
		absPath := filepath.Join(releasesPath(sistema), id+".yaml")
		if !fileExists(absPath) {
			writeError(w, 404, "release no encontrado", map[string]any{"release_id": id, "sistema": sistema})
			return
		}
		if rel, err := readRelease(absPath); err == nil {
			if status, _ := rel["status"].(string); status == "shipped" {
				writeError(w, 403, "no se puede archivar release shipped", map[string]any{"release_id": id})
				return
			}
		}
		archiveDir := filepath.Join(releasesPath(sistema), "_archived")
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			writeError(w, 500, "error archivando release", map[string]any{"detail": err.Error()})
			return
		}
		target := filepath.Join(archiveDir, id+".yaml")
		if err := os.Rename(absPath, target); err != nil {
			writeError(w, 500, "error archivando release", map[string]any{"detail": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "archived_path": target})

	default:
		writeError(w, 405, "method not allowed", nil)
	}
}

// ── POST /api/merge-release ─────────────────────────────────────────────────

// storyStateReader devuelve el estado de una story (o "" si no se encuentra).
// Seam testeable: la impl real (fsStoryState) lee checkpoint.md de disco;
// los tests inyectan un fake → el merge-gate se prueba sin filesystem.
type storyStateReader func(sistema, storyID string) string

// fsStoryState lee el estado desde checkpoint.md (live primero, luego archive del año).
func fsStoryState(sistema, storyID string) string {
	currentYear := strconv.Itoa(time.Now().Year())
	for _, base := range []string{storiesPath(sistema), archivePath(sistema, currentYear)} {
		if doc, err := readMarkdownWithFrontmatter(filepath.Join(base, storyID, "checkpoint.md")); err == nil {
			if s, ok := doc.Frontmatter["state"].(string); ok {
				return s
			}
			return ""
		}
	}
	return ""
}

// classifyReleaseStories aplica el merge-gate (G7/G8): separa estados conocidos y
// stories no-terminales (las que bloquean el merge). Puro respecto al I/O — el lector
// de estado se inyecta. Una story sin estado ("") cuenta como no-terminal (bloquea).
func classifyReleaseStories(sistema string, storyIDs []any, read storyStateReader) (states, nonTerminal []string) {
	for _, idAny := range storyIDs {
		id, _ := idAny.(string)
		state := read(sistema, id)
		if state != "" {
			states = append(states, state)
		}
		if !isTerminalState(state) {
			nonTerminal = append(nonTerminal, id)
		}
	}
	return states, nonTerminal
}

func loadStoryStatesForRelease(sistema string, storyIDs []any) ([]string, []string) {
	return classifyReleaseStories(sistema, storyIDs, fsStoryState)
}

func handleMergeRelease(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema            string `json:"sistema"`
		ReleaseID        string `json:"releaseId"`
		ConfirmFinal     bool   `json:"confirmFinal"`
		Verified         bool   `json:"verified"`
		VerificationNote string `json:"verificationNote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if body.Sistema == "" || body.ReleaseID == "" {
		writeError(w, 400, "body inválido", nil)
		return
	}
	if !containsStr(writableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}

	absPath := filepath.Join(releasesPath(body.Sistema), body.ReleaseID+".yaml")
	rel, err := readRelease(absPath)
	if err != nil {
		writeError(w, 404, "release no encontrado", map[string]any{"release_id": body.ReleaseID, "sistema": body.Sistema})
		return
	}
	if status, _ := rel["status"].(string); status == "shipped" {
		writeError(w, 409, "release ya está shipped", map[string]any{"release_id": body.ReleaseID})
		return
	}

	storyIDs, _ := rel["stories"].([]any)
	states, nonTerminal := loadStoryStatesForRelease(body.Sistema, storyIDs)
	if len(nonTerminal) > 0 {
		writeError(w, 409, "release no está listo para merge · hay stories no terminales", map[string]any{
			"release_id":           body.ReleaseID,
			"non_terminal_stories": nonTerminal,
			"states":               states,
		})
		return
	}

	year := strconv.Itoa(time.Now().Year())
	today := time.Now().UTC().Format("2006-01-02")
	plan := []map[string]any{}
	for _, idAny := range storyIDs {
		id, _ := idAny.(string)
		src := filepath.Join(storiesPath(body.Sistema), id)
		if dirExists(src) {
			plan = append(plan, map[string]any{
				"op":          "archive_story",
				"description": "git mv " + id + " → archive/" + year + "/stories/" + id,
				"source":      src,
				"target":      filepath.Join(archivePath(body.Sistema, year), id),
			})
		}
	}
	plan = append(plan, map[string]any{
		"op":          "update_release",
		"description": "update " + body.ReleaseID + ".yaml status=shipped + shipped_date=" + today,
		"target":      absPath,
	})
	notesPath := filepath.Join(sistemaPathMust(body.Sistema), "release-notes", body.ReleaseID+".md")
	plan = append(plan, map[string]any{
		"op":          "generate_release_notes",
		"description": "generar release-notes/" + body.ReleaseID + ".md con summary de stories shipped",
		"target":      notesPath,
	})

	if !body.ConfirmFinal {
		writeJSON(w, 200, map[string]any{"plan": plan, "executed": false, "preview": true})
		return
	}

	if !body.Verified {
		writeError(w, 409, "falta confirmar la prueba de comportamiento (verified=true)", map[string]any{
			"release_id": body.ReleaseID,
		})
		return
	}

	rel["status"] = "shipped"
	rel["shipped_date"] = nowISO()
	rel["verified_by"] = "operador"
	rel["verified_at"] = nowISO()
	if rel["production_status"] == nil {
		rel["production_status"] = "not_deployed"
	}
	if body.VerificationNote != "" {
		rel["verification_note"] = body.VerificationNote
	}
	if err := writeRelease(rel); err != nil {
		writeError(w, 500, "error escribiendo release", map[string]any{"detail": err.Error()})
		return
	}

	updated, _ := readRelease(absPath)
	writeJSON(w, 200, map[string]any{"plan": plan, "executed": true, "release": updated})
}

func sistemaPathMust(sistema string) string {
	p, _ := sistemaPath(sistema)
	return p
}
