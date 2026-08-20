// handlers_create.go — POST /api/extend-cap · /api/from-done · /api/refs/upload.
package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	moduleSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	storySlugRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
)

// newOperatorInputContent: operator-input.md con las 3 secciones vacías.
func newOperatorInputContent(storyID string) (string, error) {
	oi := &operatorInput{
		Frontmatter: map[string]any{
			"story_id":      storyID,
			"created_at":    localISOWithOffset(),
			"last_modified": localISOWithOffset(),
		},
		Notes: []opNote{}, Refs: []opRef{}, Conversation: []opConvEntry{},
	}
	return serializeOperatorInput(oi)
}

var checkpointKeyOrder = []string{
	"story_id", "title", "state", "goal", "release", "cap_target",
	"cap_change_type", "parent_story", "agent_owner", "module",
}

func createStoryFiles(sistema, slug string, frontmatter map[string]any) (string, string, error) {
	storyDir := filepath.Join(storiesPath(sistema), slug)
	ckptPath := filepath.Join(storyDir, "checkpoint.md")
	oiPath := filepath.Join(storyDir, "operator-input.md")

	body := "\n## Overview\n\n## Acceptance Criteria\n- [ ] \n\n## Technical Notes\n"
	if err := writeMarkdownWithFrontmatter(ckptPath, frontmatter, body, checkpointKeyOrder); err != nil {
		return "", "", err
	}
	oiContent, err := newOperatorInputContent(slug)
	if err != nil {
		return "", "", err
	}
	if err := writeFileAtomic(oiPath, oiContent); err != nil {
		return "", "", err
	}
	return ckptPath, oiPath, nil
}

// handleStoryNew — "+ Nueva story": scaffold de una idea DESDE CERO (sin parent
// cap ni parent story). Crea checkpoint.md + operator-input.md juntos (R4) en
// state=idea. La cap se declara recién al refinar (cap_target/cap_change_type
// quedan null — el cap-gate del pre-commit lo exige así en idea-stage).
func handleStoryNew(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema   string `json:"sistema"`
		Slug    string `json:"slug"`
		Goal    string `json:"goal"`
		Release string `json:"release"`
		Type    string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if !storySlugRe.MatchString(body.Slug) || len(body.Slug) < 3 {
		writeError(w, 400, "slug inválido (kebab-case, ≥3 chars)", nil)
		return
	}
	if len(strings.TrimSpace(body.Goal)) < 10 {
		writeError(w, 400, "goal requerido (≥10 chars): qué busca lograr la idea", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}
	if body.Type == "" {
		body.Type = "feature"
	}
	if body.Type != "feature" && body.Type != "bugfix" {
		writeError(w, 400, "type inválido (feature | bugfix)", nil)
		return
	}
	if dirExists(filepath.Join(storiesPath(body.Sistema), body.Slug)) {
		writeError(w, 409, "story ya existe", map[string]any{"story_id": body.Slug})
		return
	}

	fm := map[string]any{
		"story_id":        body.Slug,
		"state":           "idea",
		"goal":            strings.TrimSpace(body.Goal),
		"cap_target":      nil,
		"cap_change_type": nil,
		"parent_story":    nil,
		"type":            body.Type,
		"spawned_by":      "cockpit",
		"spawned_at":      localISOWithOffset(),
	}
	if body.Release != "" {
		fm["release"] = body.Release
	}
	ckptPath, oiPath, err := createStoryFiles(body.Sistema, body.Slug, fm)
	if err != nil {
		writeError(w, 500, "error creando story", map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{
		"storyId": body.Slug,
		"checkpointPath": ckptPath, "operatorInputPath": oiPath,
	})
}

func handleExtendCap(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema     string `json:"sistema"`
		ParentCap struct {
			Module string `json:"module"`
			Slug   string `json:"slug"`
		} `json:"parentCap"`
		CapChangeType string `json:"capChangeType"`
		NewStorySlug  string `json:"newStorySlug"`
		Goal          string `json:"goal"`
		Release       string `json:"release"`
		DerivedName   string `json:"derivedName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	validType := map[string]bool{"fix": true, "extend": true, "derive": true}
	if !moduleSlugRe.MatchString(body.ParentCap.Module) || !moduleSlugRe.MatchString(body.ParentCap.Slug) ||
		!validType[body.CapChangeType] || !storySlugRe.MatchString(body.NewStorySlug) ||
		len(body.Goal) < 10 || body.Release == "" {
		writeError(w, 400, "body inválido", nil)
		return
	}
	if !containsStr(writableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}

	parentPath := capYamlPath(body.Sistema, body.ParentCap.Module, body.ParentCap.Slug)
	parent, err := readCapability(parentPath)
	if err != nil {
		writeError(w, 404, "parentCap no encontrada", map[string]any{
			"parent": body.ParentCap.Module + "/" + body.ParentCap.Slug,
		})
		return
	}

	if body.CapChangeType == "derive" && !storySlugRe.MatchString(body.DerivedName) {
		writeError(w, 400, "capChangeType=derive requiere \"derivedName\" (slug de la cap hija)", nil)
		return
	}

	if dirExists(filepath.Join(storiesPath(body.Sistema), body.NewStorySlug)) {
		writeError(w, 409, "story ya existe", map[string]any{"story_id": body.NewStorySlug})
		return
	}

	capTarget := body.ParentCap.Slug
	if body.CapChangeType == "derive" {
		capTarget = body.DerivedName
	}

	fm := map[string]any{
		"story_id":        body.NewStorySlug,
		"state":           "idea",
		"goal":            body.Goal,
		"release":         body.Release,
		"cap_target":      capTarget,
		"cap_change_type": body.CapChangeType,
		"parent_story":    nil,
		"agent_owner":     parent["agent_owner"],
		"module":          parent["module"],
	}
	ckptPath, oiPath, err := createStoryFiles(body.Sistema, body.NewStorySlug, fm)
	if err != nil {
		writeError(w, 500, "error creando story", map[string]any{"detail": err.Error()})
		return
	}
	writeJSON(w, 200, map[string]any{
		"storyId": body.NewStorySlug, "slug": body.NewStorySlug,
		"checkpointPath": ckptPath, "operatorInputPath": oiPath,
	})
}

func handleFromDone(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema         string `json:"sistema"`
		ParentStoryID string `json:"parentStoryId"`
		NewStorySlug  string `json:"newStorySlug"`
		Goal          string `json:"goal"`
		Release       string `json:"release"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if body.ParentStoryID == "" || !storySlugRe.MatchString(body.NewStorySlug) ||
		len(body.Goal) < 10 || body.Release == "" {
		writeError(w, 400, "body inválido", nil)
		return
	}
	if !containsStr(writableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}

	parentDir := findStoryDir(body.Sistema, body.ParentStoryID)
	if parentDir == "" {
		writeError(w, 404, "parent story no encontrada", map[string]any{"story_id": body.ParentStoryID})
		return
	}
	parentDoc, err := readMarkdownWithFrontmatter(filepath.Join(parentDir, "checkpoint.md"))
	if err != nil {
		writeError(w, 500, "error leyendo story", map[string]any{"detail": err.Error()})
		return
	}
	if state, _ := parentDoc.Frontmatter["state"].(string); state != "done" {
		writeError(w, 409, "parent story no está en state=done", map[string]any{
			"story_id": body.ParentStoryID, "state": parentDoc.Frontmatter["state"],
		})
		return
	}

	if dirExists(filepath.Join(storiesPath(body.Sistema), body.NewStorySlug)) {
		writeError(w, 409, "story ya existe", map[string]any{"story_id": body.NewStorySlug})
		return
	}

	fm := map[string]any{
		"story_id":     body.NewStorySlug,
		"state":        "idea",
		"goal":         body.Goal,
		"release":      body.Release,
		"parent_story": body.ParentStoryID,
		"cap_target":   parentDoc.Frontmatter["cap_target"],
		"agent_owner":  parentDoc.Frontmatter["agent_owner"],
		"module":       parentDoc.Frontmatter["module"],
	}
	if fm["cap_target"] != nil {
		fm["cap_change_type"] = "extend"
	}
	ckptPath, oiPath, err := createStoryFiles(body.Sistema, body.NewStorySlug, fm)
	if err != nil {
		writeError(w, 500, "error creando story", map[string]any{"detail": err.Error()})
		return
	}

	// story-ref al operator-input nuevo
	if raw, err := os.ReadFile(oiPath); err == nil {
		if oi, err := parseOperatorInput(string(raw)); err == nil {
			oi.Refs = append(oi.Refs, opRef{
				Type: "story-ref", Value: body.ParentStoryID,
				Comment: "Spawned from parent: " + body.ParentStoryID + " (done)",
			})
			if serialized, err := serializeOperatorInput(oi); err == nil {
				_ = writeFileAtomic(oiPath, serialized)
			}
		}
	}

	writeJSON(w, 200, map[string]any{
		"storyId": body.NewStorySlug, "checkpointPath": ckptPath, "operatorInputPath": oiPath,
	})
}

// ── POST /api/refs/upload (multipart) ───────────────────────────────────────

const maxUploadBytes = 10 * 1024 * 1024

var allowedExts = map[string]bool{
	"png": true, "jpg": true, "jpeg": true, "gif": true,
	"webp": true, "svg": true, "pdf": true, "md": true, "txt": true,
}
var unsafeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func refTypeForExt(ext string) string {
	switch ext {
	case "png", "jpg", "jpeg", "gif", "webp", "svg":
		return "img"
	case "pdf", "md", "txt":
		return "doc"
	}
	return "doc"
}

func handleRefsUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadBytes + 1024); err != nil {
		writeError(w, 400, "body multipart inválido", nil)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, 400, "campo \"file\" requerido (Blob)", nil)
		return
	}
	defer file.Close()

	sistema := r.FormValue("sistema")
	storyID := r.FormValue("storyId")
	comment := r.FormValue("comment")
	if sistema == "" {
		writeError(w, 400, "campo \"sistema\" requerido", nil)
		return
	}
	if storyID == "" {
		writeError(w, 400, "campo \"storyId\" requerido", nil)
		return
	}
	if !containsStr(writableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}
	if header.Size > maxUploadBytes {
		writeError(w, 413, "archivo excede 10MB", map[string]any{"size": header.Size, "max": maxUploadBytes})
		return
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))
	if !allowedExts[ext] {
		allowed := []string{}
		for e := range allowedExts {
			allowed = append(allowed, e)
		}
		writeError(w, 400, "extensión no permitida", map[string]any{"extension": ext, "allowed": allowed})
		return
	}

	storyDir := findStoryDir(sistema, storyID)
	if storyDir == "" {
		writeError(w, 404, "story no encontrada", map[string]any{"story_id": storyID, "sistema": sistema})
		return
	}

	name := unsafeFilenameRe.ReplaceAllString(header.Filename, "_")
	name = strings.TrimLeft(name, ".")
	if len(name) > 80 {
		name = name[len(name)-80:]
	}
	if name == "" || strings.Contains(name, "..") {
		writeError(w, 400, "extensión no permitida", map[string]any{"extension": ext})
		return
	}

	refsDir := filepath.Join(storyDir, "refs")
	if err := os.MkdirAll(refsDir, 0o755); err != nil {
		writeError(w, 500, "error preparando ruta destino", map[string]any{"detail": err.Error()})
		return
	}

	// Colisiones: foo.png → foo-2.png
	target := filepath.Join(refsDir, name)
	baseName := strings.TrimSuffix(name, "."+ext)
	for i := 2; fileExists(target); i++ {
		name = baseName + "-" + strconv.Itoa(i) + "." + ext
		target = filepath.Join(refsDir, name)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
	if err != nil || len(data) > maxUploadBytes {
		writeError(w, 413, "archivo excede 10MB", map[string]any{"size": len(data), "max": maxUploadBytes})
		return
	}
	if err := writeFileAtomic(target, string(data)); err != nil {
		writeError(w, 500, "error escribiendo archivo", map[string]any{"detail": err.Error()})
		return
	}

	// Side-effect: appendear ref al operator-input si existe
	if oiPath := findOperatorInputPath(sistema, storyID); oiPath != "" {
		if raw, err := os.ReadFile(oiPath); err == nil {
			if oi, err := parseOperatorInput(string(raw)); err == nil {
				ref := opRef{Type: refTypeForExt(ext), Value: "refs/" + name}
				if comment != "" {
					ref.Comment = comment
				}
				oi.Refs = append(oi.Refs, ref)
				oi.Frontmatter["last_modified"] = localISOWithOffset()
				if serialized, err := serializeOperatorInput(oi); err == nil {
					_ = writeFileAtomic(oiPath, serialized)
				}
			}
		}
	}

	writeJSON(w, 200, map[string]any{
		"relativePath": "refs/" + name, "displayName": name, "absolutePath": target,
	})
}
