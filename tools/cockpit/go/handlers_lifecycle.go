// handlers_lifecycle.go — el CIL (Continuous Improvement Loop): harness-backlog (L1)
// + tech-debt (L3) + learnings (L2) + cap-doctor (L4). Incluye el parser de tablas
// markdown lifecycle (estados reported→…→verified) — código puro y no trivial,
// extraído del god-file handlers_misc.go. El parser es testeable → handlers_lifecycle_test.go.
package main

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var lifecycleOrder = []string{"reported", "triaged", "ratified", "applied", "verified", "deferred"}
var validEstados = map[string]bool{
	"reported": true, "triaged": true, "ratified": true,
	"applied": true, "verified": true, "deferred": true,
}
var openEstados = map[string]bool{
	"reported": true, "triaged": true, "ratified": true, "applied": true,
}
var emphasisRe = regexp.MustCompile("\\*\\*|__|\\*|`")
var estadoWordRe = regexp.MustCompile(`[a-záéíóúñ]+`)
var carrilRe = regexp.MustCompile(`(?i)\[(L[234])\]`)

func stripEmphasis(s string) string {
	return strings.TrimSpace(emphasisRe.ReplaceAllString(s, ""))
}

func normalizeEstado(raw string) string {
	first := estadoWordRe.FindString(strings.ToLower(stripEmphasis(raw)))
	if validEstados[first] {
		return first
	}
	return "otro"
}

func splitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return nil
	}
	cells := strings.Split(trimmed, "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}
	if len(cells) >= 2 && cells[0] == "" {
		cells = cells[1:]
	}
	if len(cells) >= 1 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

func tableRows(md string, idRe *regexp.Regexp) [][]string {
	var rows [][]string
	for _, line := range strings.Split(md, "\n") {
		cells := splitTableRow(line)
		if cells == nil || len(cells) < 6 || !idRe.MatchString(cells[0]) {
			continue
		}
		rows = append(rows, cells)
	}
	return rows
}

type lifecycleItem = map[string]any

func parseLifecycleTable(md string, idRe *regexp.Regexp, idPrefix string, sevMap map[string]string, withCarril bool) []lifecycleItem {
	items := []lifecycleItem{}
	for _, cells := range tableRows(md, idRe) {
		id := cells[0]
		fecha := cells[1]
		sevEmoji := cells[2]
		ref := cells[len(cells)-1]
		estadoRaw := cells[len(cells)-2]
		item := strings.TrimSpace(strings.Join(cells[3:len(cells)-2], " | "))

		num, _ := strconv.Atoi(strings.TrimPrefix(id, idPrefix))
		sevLabel := sevMap[sevEmoji]
		if sevLabel == "" {
			sevLabel = "otro"
		}
		entry := lifecycleItem{
			"id": id, "num": num, "fecha": fecha,
			"sevEmoji": sevEmoji, "sevLabel": sevLabel,
			"item": item, "title": stripEmphasis(item),
			"estado": normalizeEstado(estadoRaw), "estadoRaw": estadoRaw, "ref": ref,
		}
		if withCarril {
			carril := "L1"
			if m := carrilRe.FindStringSubmatch(item); m != nil {
				carril = strings.ToUpper(m[1])
			}
			entry["carril"] = carril
		}
		items = append(items, entry)
	}
	return items
}

var harnessSevMap = map[string]string{
	"🔴": "silent-killer", "🟡": "quick-win", "🔵": "decision", "🟣": "wave",
}
var techDebtSevMap = map[string]string{
	"🔴": "bloquea-pronto", "🟡": "friccion", "🔵": "mejora",
}
var hbIDRe = regexp.MustCompile(`^HB-\d+$`)
var tdIDRe = regexp.MustCompile(`^TD-\d+$`)

func countByEstado(items []lifecycleItem) map[string]int {
	counts := map[string]int{}
	for _, it := range items {
		estado, _ := it["estado"].(string)
		counts[estado]++
	}
	return counts
}

func countOpen(items []lifecycleItem) int {
	n := 0
	for _, it := range items {
		estado, _ := it["estado"].(string)
		if openEstados[estado] {
			n++
		}
	}
	return n
}

const harnessSource = "docs/process/harness-backlog.md"
const techDebtSource = "docs/process/tech-debt.md"

// harnessRoot: single → el root; multi → primer proyecto con harness-backlog.md
// (el CIL es factory-level; en multi se sirve desde el proyecto que lo tenga).
func harnessRoot() string {
	roots := workspaceRoots()
	if len(roots) == 0 {
		return ""
	}
	for _, r := range roots {
		if fileExists(filepath.Join(r, harnessSource)) {
			return r
		}
	}
	return roots[0]
}

func readHarnessItems() []lifecycleItem {
	raw, err := os.ReadFile(filepath.Join(harnessRoot(), harnessSource))
	if err != nil {
		return []lifecycleItem{}
	}
	return parseLifecycleTable(string(raw), hbIDRe, "HB-", harnessSevMap, true)
}

func handleHarness(w http.ResponseWriter, r *http.Request) {
	items := readHarnessItems()
	writeJSON(w, 200, map[string]any{
		"items": items, "counts": countByEstado(items), "source": harnessSource,
	})
}

func countMarkdownFiles(dir string) int {
	n := 0
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			n++
		}
		return nil
	})
	return n
}

func handleCIL(w http.ResponseWriter, r *http.Request) {
	root := harnessRoot()
	l1Items := readHarnessItems()

	l3Items := []lifecycleItem{}
	if raw, err := os.ReadFile(filepath.Join(root, techDebtSource)); err == nil {
		l3Items = parseLifecycleTable(string(raw), tdIDRe, "TD-", techDebtSevMap, false)
	}

	writeJSON(w, 200, map[string]any{
		"l1": map[string]any{
			"items": l1Items, "counts": countByEstado(l1Items),
			"open": countOpen(l1Items), "source": harnessSource,
		},
		"l3": map[string]any{
			"items": l3Items, "counts": countByEstado(l3Items),
			"open": countOpen(l3Items), "source": techDebtSource,
		},
		"l2": map[string]any{
			"count": countMarkdownFiles(filepath.Join(root, "docs", "learnings")),
			"source": "docs/learnings", "link": "/learnings",
		},
		"l4": map[string]any{
			"source": "cap-doctor (por sistema)", "link": "/drift",
			"note": "Capabilities desfasadas se monitorean por sistema en la vista Drift.",
		},
	})
}
