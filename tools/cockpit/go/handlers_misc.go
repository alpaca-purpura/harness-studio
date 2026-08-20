// handlers_misc.go — sistemas, learnings, system-map, value-stream, sessions,
// harness, cil, cap-status readers, operator-input.
// (El file-API + /api/open + la guarda de path-traversal viven en handlers_file.go.)
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
)

// ── GET /api/sistemas (endpoint nuevo · UI estática lo pide al montar) ────────

func handleSistemas(w http.ResponseWriter, r *http.Request) {
	sistemas := getSelectableSistemas()
	if sistemas == nil {
		sistemas = []string{}
	}
	// `escribibles` (vendored ArnesIA): el SERVIDOR es la autoridad sobre dónde se puede
	// escribir — es quien tiene el flag. Antes la UI lo deducía sola (`isPlatform(sistema)`
	// ⇒ solo lectura, hardcodeado), así que con el backend habilitado el drag del board
	// seguía bloqueado y el usuario veía un candado sin causa. Ahora la UI refleja este
	// dato en vez de adivinarlo, y las dos capas no pueden discrepar.
	escribibles := writableSistemas()
	if escribibles == nil {
		escribibles = []string{}
	}
	writeJSON(w, 200, map[string]any{
		"sistemas":        sistemas,
		"escribibles":     escribibles,
		"default_sistema": defaultSistemaFromEnv(),
	})
}

// ── GET /api/learnings ──────────────────────────────────────────────────────

var headingRe = regexp.MustCompile(`(?m)^#+\s+.*`)

func handleSystemMap(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	// Handler de LECTURA: valida con getSelectableSistemas (incluye el pseudo
	// `platform`), NO getSistemas (lo omite) → vitalia/nicolify/comunify resuelven
	// a platform; con getSistemas darían "sistema desconocido" (bug latente; I-40
	// fijó getSelectableSistemas = lo que ofrece el selector, lo que valida lectura).
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}

	yamlPath := filepath.Join(sistemaDocsRoot(sistema), "architecture", "SYSTEM-MAP.yaml")
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		writeError(w, 500, "error leyendo SYSTEM-MAP.yaml: "+errnoMessage(err, yamlPath), nil)
		return
	}
	data, err := parseYAMLMap(raw)
	if err != nil {
		writeError(w, 500, "error leyendo SYSTEM-MAP.yaml: "+err.Error(), nil)
		return
	}
	writeJSON(w, 200, map[string]any{"system_map": data, "path": yamlPath})
}

// ── GET /api/ledger ───────────────────────────────────────────────────────
// Lente "Evolución" (I-45): un sistema NO-SDD (la fábrica) trackea su evolución por
// el Ledger (ledger.yaml generado desde PRODUCT-VISION) en vez de stories/releases.
// Genérico: cualquier sistema con docs/product/ledger.yaml lo expone; sin él → ledger:null
// (empty-state honesto en la UI, no rompe los sistemas que sí corren SDD).
func handleLedger(w http.ResponseWriter, r *http.Request) {
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	// Valida contra getSelectableSistemas (incluye el pseudo `platform`), NO getSistemas
	// (que lo omite): vitalia/nicolify/comunify resuelven al platform-pseudo → con getSistemas
	// darían "sistema desconocido". Espeja a stories/learnings/releases (handlers de lectura OK).
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}

	yamlPath := filepath.Join(sistemaDocsRoot(sistema), "product", "ledger.yaml")
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, 200, map[string]any{"ledger": nil, "path": yamlPath})
			return
		}
		writeError(w, 500, "error leyendo ledger.yaml: "+errnoMessage(err, yamlPath), nil)
		return
	}
	data, err := parseYAMLMap(raw)
	if err != nil {
		writeError(w, 500, "error leyendo ledger.yaml: "+err.Error(), nil)
		return
	}
	writeJSON(w, 200, map[string]any{"ledger": data, "path": yamlPath})
}

// errnoMessage replica el formato de error de Node (`ENOENT: no such file or directory, open '...'`).
func errnoMessage(err error, p string) string {
	if os.IsNotExist(err) {
		return "ENOENT: no such file or directory, open '" + p + "'"
	}
	return err.Error()
}

// ── GET /api/value-stream ───────────────────────────────────────────────────

func handleValueStream(w http.ResponseWriter, r *http.Request) {
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
	stages := normalizeValueStream(getValueStreamStages(sistema))
	writeJSON(w, 200, map[string]any{"stages": stages})
}

// normalizeValueStream replica getValueStreamStages del TS:
// strings → {id,name,description,boxIds} · objetos → sort por order.
func normalizeValueStream(v any) any {
	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	if s, isStr := arr[0].(string); isStr {
		_ = s
		out := []map[string]any{}
		for _, item := range arr {
			name, _ := item.(string)
			if name == "" || name == fillSentinel {
				continue
			}
			id := strings.ToLower(name)
			id = regexp.MustCompile(`\s+`).ReplaceAllString(id, "-")
			out = append(out, map[string]any{"id": id, "name": name, "description": "", "boxIds": []any{}})
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}
	// objetos: sort por order ASC
	objs := []map[string]any{}
	for _, item := range arr {
		if m, isMap := item.(map[string]any); isMap {
			objs = append(objs, m)
		}
	}
	sort.SliceStable(objs, func(i, j int) bool {
		return toFloat(objs[i]["order"]) < toFloat(objs[j]["order"])
	})
	out := []map[string]any{}
	for _, m := range objs {
		boxIds := m["boxIds"]
		if boxIds == nil {
			boxIds = []any{}
		}
		out = append(out, map[string]any{
			"id": m["id"], "name": m["name"],
			"description": fmGetOr(m, "", "description"),
			"boxIds":      boxIds,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// ── GET /api/sessions ───────────────────────────────────────────────────────

func lockField(v string) any {
	if v == "" || v == "—" || v == "-" {
		return nil
	}
	return v
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []map[string]any{}
	for _, root := range workspaceRoots() {
		lockDir := filepath.Join(root, ".session-locks")
		entries, err := os.ReadDir(lockDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".lock") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(lockDir, e.Name()))
		if err != nil {
			continue
		}
		firstLine := strings.TrimSpace(strings.SplitN(string(raw), "\n", 2)[0])
		fields := strings.Fields(firstLine)
		if len(fields) < 3 {
			continue
		}
		pid, _ := strconv.Atoi(fields[0])
		if !pidAlive(pid) {
			continue
		}
		bucket := strings.ReplaceAll(strings.TrimSuffix(e.Name(), ".lock"), "__", ":")
		if len(fields) >= 6 {
			if b := lockField(fields[5]); b != nil {
				bucket = b.(string)
			}
		}
		s := map[string]any{
			"bucket":    bucket,
			"pid":       pid,
			"skill":     fields[1],
			"startedAt": lockField(fields[2]),
			"storyId":   nil,
			"lane":      nil,
		}
		if len(fields) >= 4 {
			s["storyId"] = lockField(fields[3])
		}
			if len(fields) >= 5 {
				s["lane"] = lockField(fields[4])
			}
			sessions = append(sessions, s)
		}
	}

	byStory := map[string]any{}
	for _, s := range sessions {
		storyID, _ := s["storyId"].(string)
		if storyID == "" {
			continue
		}
		prev, exists := byStory[storyID].(map[string]any)
		if !exists {
			byStory[storyID] = s
			continue
		}
		prevAt, _ := prev["startedAt"].(string)
		curAt, _ := s["startedAt"].(string)
		if curAt > prevAt {
			byStory[storyID] = s
		}
	}
	writeJSON(w, 200, map[string]any{"sessions": sessions, "by_story": byStory})
}

// ── Capability artifact readers (status / bidirectional / code-index / doctor) ──

func capArtifactHandler(filename, key, hintTemplate string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		absPath := filepath.Join(capabilitiesPath(sistema), filename)
		raw, err := os.ReadFile(absPath)
		if err != nil {
			writeJSON(w, 200, map[string]any{
				key: nil, "path": absPath, "sistema": sistema,
				"hint": strings.ReplaceAll(hintTemplate, "{sistema}", sistema),
			})
			return
		}
		var data any
		if err := json.Unmarshal(raw, &data); err != nil {
			writeError(w, 500, filename+" mal formado: "+err.Error(), map[string]any{"path": absPath})
			return
		}
		writeJSON(w, 200, map[string]any{key: data, "path": absPath, "sistema": sistema})
	}
}

var doctorGateLabels = [][2]string{
	{"G1", "headers huérfanos (→ cap inexistente)"},
	{"G2", "cajas live vacías (área sin cap)"},
	{"G3", "caps sin hogar (fa ∉ SYSTEM-MAP)"},
	{"G4", "paths rotos (declarados, no existen)"},
	{"G5", "supersesiones rotas"},
	{"G6", "caps live invisibles en el mapa"},
	{"G7", "caps ilegibles por el cockpit (YAML dup-key)"},
	{"G8", "caps live+visibles sin user_facing_description"},
	{"G9", "caps live+visibles sin scenarios"},
}

func handleCapDoctor(w http.ResponseWriter, r *http.Request) {
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
	absPath := filepath.Join(capabilitiesPath(sistema), "_bidirectional-validation.json")
	raw, err := os.ReadFile(absPath)
	if err != nil {
		writeJSON(w, 200, map[string]any{
			"doctor": nil, "path": absPath, "sistema": sistema,
			"hint": "Ejecuta: make cap-doctor SISTEMA=" + sistema + " (o python3 scripts/validate_code_cap_bidirectional.py --sistema " + sistema + ")",
		})
		return
	}
	var report map[string]any
	if err := json.Unmarshal(raw, &report); err != nil {
		writeError(w, 500, "_bidirectional-validation.json mal formado: "+err.Error(), map[string]any{"path": absPath})
		return
	}

	// Schema actual del validador = `cap_gates` (G1-G9); `gates` queda como
	// fallback de reports viejos pre-rename.
	gatesRaw, _ := report["cap_gates"].(map[string]any)
	if gatesRaw == nil {
		gatesRaw, _ = report["gates"].(map[string]any)
	}
	totalDrift := 0.0
	gates := []map[string]any{}
	for _, gl := range doctorGateLabels {
		id, label := gl[0], gl[1]
		g, _ := gatesRaw[id].(map[string]any)
		drift := toFloat(g["drift"])
		total := toFloat(g["total"])
		details, _ := g["details"].([]any)
		if details == nil {
			details = []any{}
		}
		totalDrift += drift
		gates = append(gates, map[string]any{
			"id": id, "label": label, "drift": drift, "total": total, "details": details,
		})
	}
	writeJSON(w, 200, map[string]any{
		"doctor": map[string]any{
			"sistema": sistema, "healthy": totalDrift == 0, "total_drift": totalDrift,
			"hard_enforced": fmGetOr(report, false, "cap_gates_hard"),
			"validated_at":  fmGetOr(report, nil, "validated_at"),
			"gates":         gates,
		},
		"path": absPath, "sistema": sistema,
	})
}

// ── Operator input GET / PATCH ──────────────────────────────────────────────

func handleOperatorInput(w http.ResponseWriter, r *http.Request) {
	storyID := r.PathValue("storyId")
	sistema := r.URL.Query().Get("sistema")
	if sistema == "" {
		writeError(w, 400, "query param \"sistema\" requerido", nil)
		return
	}
	if !containsStr(getSelectableSistemas(), sistema) {
		writeError(w, 400, "sistema desconocido: "+sistema, nil)
		return
	}

	absPath := findOperatorInputPath(sistema, storyID)
	if absPath == "" {
		writeError(w, 404, "operator-input no encontrado", map[string]any{"story_id": storyID, "sistema": sistema})
		return
	}

	raw, err := os.ReadFile(absPath)
	if err != nil {
		writeError(w, 500, "error parseando operator-input", map[string]any{"detail": err.Error()})
		return
	}
	oi, err := parseOperatorInput(string(raw))
	if err != nil {
		writeError(w, 500, "error parseando operator-input", map[string]any{"detail": err.Error()})
		return
	}

	if r.Method == http.MethodGet {
		writeJSON(w, 200, map[string]any{"operatorInput": oi})
		return
	}

	// PATCH: { section: notes|refs|conversation, entry: {...} }
	var body struct {
		Section string         `json:"section"`
		Entry   map[string]any `json:"entry"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	ts, _ := body.Entry["timestamp"].(string)
	if ts == "" {
		ts = localTimestamp()
	}

	switch body.Section {
	case "notes":
		text, _ := body.Entry["text"].(string)
		if text == "" {
			writeError(w, 400, "body inválido", nil)
			return
		}
		oi.Notes = append(oi.Notes, opNote{Timestamp: ts, Text: text})
	case "refs":
		refType, _ := body.Entry["type"].(string)
		value, _ := body.Entry["value"].(string)
		if refTypeToEmoji[refType] == "" || value == "" {
			writeError(w, 400, "body inválido", nil)
			return
		}
		ref := opRef{Type: refType, Value: value}
		if c, ok := body.Entry["comment"].(string); ok {
			ref.Comment = c
		}
		oi.Refs = append(oi.Refs, ref)
	case "conversation":
		text, _ := body.Entry["text"].(string)
		if text == "" {
			writeError(w, 400, "body inválido", nil)
			return
		}
		oi.Conversation = append(oi.Conversation, opConvEntry{Timestamp: ts, Author: "operador", Text: text})
	default:
		writeError(w, 400, "body inválido", nil)
		return
	}

	oi.Frontmatter["last_modified"] = localISOWithOffset()
	serialized, err := serializeOperatorInput(oi)
	if err != nil {
		writeError(w, 500, "error escribiendo operator-input", map[string]any{"detail": err.Error()})
		return
	}
	if err := writeFileAtomic(absPath, serialized); err != nil {
		writeError(w, 500, "error escribiendo operator-input", map[string]any{"detail": err.Error()})
		return
	}
	updated, _ := parseOperatorInput(serialized)
	writeJSON(w, 200, map[string]any{"operatorInput": updated})
}

