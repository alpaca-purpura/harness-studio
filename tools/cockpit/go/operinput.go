// operinput.go — port de lib/operator-input-parser.ts (round-trip idempotente).
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	headerNotes = "## 💭 Notas"
	headerRefs  = "## 📎 Referencias"
	headerConv  = "## 💬 Conversación"
)

var (
	timestampRe    = regexp.MustCompile(`^### (\d{4}-\d{2}-\d{2} \d{2}:\d{2})$`)
	convOperatorRe = regexp.MustCompile(`^### (\d{4}-\d{2}-\d{2} \d{2}:\d{2}) · 🧑 (?:operador|chris)$`)
	convClaudeRe   = regexp.MustCompile("^### (\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}) · 🤖 claude · `/([^`]+)` · (\\S+) (.+)$")
	refRe          = regexp.MustCompile(`^- \*\*(\S+) ([a-z-]+)\*\* · (.+)$`)
	refCommentRe   = regexp.MustCompile(`^  > (.+)$`)
)

var refTypeToEmoji = map[string]string{
	"link": "🔗", "img": "🖼", "text": "💬",
	"story-ref": "📖", "learning-ref": "📚", "doc": "📄",
}
var emojiToRefType = map[string]string{
	"🔗": "link", "🖼": "img", "💬": "text",
	"📖": "story-ref", "📚": "learning-ref", "📄": "doc",
}
var verdictToLabel = map[string][2]string{
	"applied": {"✓", "APLICADO"}, "doubt": {"⚠️", "DUDA"},
	"refuted": {"❌", "REFUTADO"}, "proposed": {"💡", "PROPONE"},
}
var labelToVerdict = map[string]string{
	"APLICADO": "applied", "DUDA": "doubt", "REFUTADO": "refuted", "PROPONE": "proposed",
}

type opNote struct {
	Timestamp string `json:"timestamp"`
	Text      string `json:"text"`
}
type opRef struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}
type opConvEntry struct {
	Timestamp string `json:"timestamp"`
	Author    string `json:"author"`
	Skill     string `json:"skill,omitempty"`
	Verdict   string `json:"verdict,omitempty"`
	Text      string `json:"text"`
}
type operatorInput struct {
	Frontmatter map[string]any `json:"frontmatter"`
	Notes       []opNote       `json:"notes"`
	Refs        []opRef        `json:"refs"`
	Conversation []opConvEntry `json:"conversation"`
	Preamble    string         `json:"preamble,omitempty"`
}

func extractLeadingHTMLComments(content string) (string, string) {
	lines := strings.Split(content, "\n")
	var collected []string
	idx := 0
	for idx < len(lines) {
		line := lines[idx]
		if strings.HasPrefix(line, "<!--") {
			collected = append(collected, line)
			idx++
			continue
		}
		if strings.TrimSpace(line) == "" {
			idx++
			continue
		}
		break
	}
	return strings.Join(collected, "\n"), strings.Join(lines[idx:], "\n")
}

func parseOperatorInput(content string) (*operatorInput, error) {
	leadingComments, rest := extractLeadingHTMLComments(content)
	doc, err := parseFrontmatter(rest)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(doc.Content, "\n")
	idxNotes, idxRefs, idxConv := -1, -1, -1
	for i, l := range lines {
		switch strings.TrimSpace(l) {
		case headerNotes:
			if idxNotes == -1 {
				idxNotes = i
			}
		case headerRefs:
			if idxRefs == -1 {
				idxRefs = i
			}
		case headerConv:
			if idxConv == -1 {
				idxConv = i
			}
		}
	}
	if idxNotes == -1 || idxRefs == -1 || idxConv == -1 {
		return nil, fmt.Errorf("operator-input malformado: faltan una o más de las 3 secciones (💭 Notas / 📎 Referencias / 💬 Conversación)")
	}

	innerPreamble := strings.TrimSpace(strings.Join(lines[:idxNotes], "\n"))
	var preambleParts []string
	if leadingComments != "" {
		preambleParts = append(preambleParts, leadingComments)
	}
	if innerPreamble != "" {
		preambleParts = append(preambleParts, innerPreamble)
	}
	preamble := strings.TrimSpace(strings.Join(preambleParts, "\n\n"))

	oi := &operatorInput{
		Frontmatter:  doc.Frontmatter,
		Notes:        parseNotesSection(lines[idxNotes+1 : idxRefs]),
		Refs:         parseRefsSection(lines[idxRefs+1 : idxConv]),
		Conversation: parseConvSection(lines[idxConv+1:]),
		Preamble:     preamble,
	}
	return oi, nil
}

func isPlaceholderNote(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "sin notas todavía") ||
		strings.Contains(lower, "escribe aqui") ||
		strings.Contains(lower, "escribe aquí")
}

func parseNotesSection(lines []string) []opNote {
	notes := []opNote{}
	var current *opNote
	var acc []string
	flush := func() {
		if current != nil {
			current.Text = strings.TrimSpace(strings.Join(acc, "\n"))
			if current.Text != "" && !isPlaceholderNote(current.Text) {
				notes = append(notes, *current)
			}
			acc = nil
			current = nil
		}
	}
	for _, line := range lines {
		if m := timestampRe.FindStringSubmatch(line); m != nil {
			flush()
			current = &opNote{Timestamp: m[1]}
			continue
		}
		if current != nil {
			acc = append(acc, line)
		}
	}
	flush()
	return notes
}

func parseRefsSection(lines []string) []opRef {
	refs := []opRef{}
	var pending *opRef
	for _, line := range lines {
		if m := refRe.FindStringSubmatch(line); m != nil {
			if pending != nil {
				refs = append(refs, *pending)
			}
			emoji, typeStr, value := m[1], m[2], m[3]
			refType := typeStr
			if t, ok := emojiToRefType[emoji]; ok {
				refType = t
			}
			pending = &opRef{Type: refType, Value: strings.TrimSpace(value)}
			continue
		}
		if m := refCommentRe.FindStringSubmatch(line); m != nil && pending != nil {
			pending.Comment = strings.TrimSpace(m[1])
			continue
		}
	}
	if pending != nil {
		refs = append(refs, *pending)
	}
	return refs
}

func parseConvSection(lines []string) []opConvEntry {
	entries := []opConvEntry{}
	var current *opConvEntry
	var acc []string
	flush := func() {
		if current != nil {
			current.Text = strings.TrimSpace(strings.Join(acc, "\n"))
			entries = append(entries, *current)
			acc = nil
			current = nil
		}
	}
	for _, line := range lines {
		if m := convClaudeRe.FindStringSubmatch(line); m != nil {
			flush()
			label := strings.Fields(strings.TrimSpace(m[4]))
			verdict := ""
			if len(label) > 0 {
				verdict = labelToVerdict[label[0]]
			}
			if verdict == "" {
				for v, info := range verdictToLabel {
					if info[0] == m[3] {
						verdict = v
						break
					}
				}
			}
			current = &opConvEntry{Timestamp: m[1], Author: "claude", Skill: m[2], Verdict: verdict}
			continue
		}
		if m := convOperatorRe.FindStringSubmatch(line); m != nil {
			flush()
			current = &opConvEntry{Timestamp: m[1], Author: "operador"}
			continue
		}
		if current != nil {
			acc = append(acc, line)
		}
	}
	flush()
	return entries
}

func splitPreamble(preamble string) (string, string) {
	if preamble == "" {
		return "", ""
	}
	lines := strings.Split(preamble, "\n")
	var htmlLines []string
	i := 0
	for i < len(lines) {
		ln := lines[i]
		if strings.HasPrefix(ln, "<!--") {
			htmlLines = append(htmlLines, ln)
			i++
			continue
		}
		if strings.TrimSpace(ln) == "" && len(htmlLines) > 0 {
			i++
			continue
		}
		break
	}
	return strings.Join(htmlLines, "\n"), strings.TrimSpace(strings.Join(lines[i:], "\n"))
}

var operInputKeyOrder = []string{
	"story_id", "created_at", "last_modified", "notes_count", "refs_count", "conversation_count",
}

func serializeOperatorInput(data *operatorInput) (string, error) {
	leadingHTML, innerPreamble := splitPreamble(data.Preamble)

	var body []string
	if innerPreamble != "" {
		body = append(body, innerPreamble, "")
	}

	body = append(body, headerNotes, "")
	for _, note := range data.Notes {
		body = append(body, "### "+note.Timestamp, note.Text, "")
	}

	body = append(body, headerRefs, "")
	if len(data.Refs) == 0 {
		body = append(body, "(sin referencias todavía)", "")
	} else {
		for _, ref := range data.Refs {
			emoji := refTypeToEmoji[ref.Type]
			body = append(body, "- **"+emoji+" "+ref.Type+"** · "+ref.Value)
			if ref.Comment != "" {
				body = append(body, "  > "+ref.Comment)
			}
		}
		body = append(body, "")
	}

	body = append(body, headerConv, "")
	for _, entry := range data.Conversation {
		if entry.Author == "claude" {
			verdict := entry.Verdict
			if verdict == "" {
				verdict = "applied"
			}
			info := verdictToLabel[verdict]
			skill := entry.Skill
			if skill == "" {
				skill = "unknown"
			}
			body = append(body, "### "+entry.Timestamp+" · 🤖 claude · `/"+skill+"` · "+info[0]+" "+info[1])
		} else {
			body = append(body, "### "+entry.Timestamp+" · 🧑 operador")
		}
		body = append(body, entry.Text, "")
	}

	fm := map[string]any{}
	for k, v := range data.Frontmatter {
		fm[k] = v
	}
	fm["notes_count"] = len(data.Notes)
	fm["refs_count"] = len(data.Refs)
	fm["conversation_count"] = len(data.Conversation)

	bodyStr := strings.Join(body, "\n")
	bodyStr = strings.TrimRight(bodyStr, "\n") + "\n"
	out, err := stringifyFrontmatter(fm, bodyStr, operInputKeyOrder)
	if err != nil {
		return "", err
	}
	if leadingHTML != "" {
		return leadingHTML + "\n" + out, nil
	}
	return out, nil
}

// findOperatorInputPath: live operator-input.md → chris-input.md → archive variants.
func findOperatorInputPath(sistema, storyID string) string {
	dir := findStoryDir(sistema, storyID)
	if dir == "" {
		return ""
	}
	for _, name := range []string{"operator-input.md", "chris-input.md"} {
		p := dir + string(os.PathSeparator) + name
		if fileExists(p) {
			return p
		}
	}
	return ""
}

func localTimestamp() string {
	return time.Now().Format("2006-01-02 15:04")
}

func localISOWithOffset() string {
	return time.Now().Format("2006-01-02T15:04:05-07:00")
}
