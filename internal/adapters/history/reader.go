// Package history lee el corpus JSONL nativo de Claude Code (RF-201, paquete
// mejorar-arnes-conversando, decisión B2/MC-D7): la JSONL es LA fuente de verdad de una
// conversación — este adapter solo la LEE (boundary indice-desechable-jsonl-es-verdad),
// jamás la escribe ni la mueve. ArnesIA aporta el join (Session.CadenaCC + cwd) para
// coser una conversación lógica partida en N JSONLs por la rotación de contexto.
package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Reader resuelve y parsea las JSONL de ~/.claude/projects/<dir-del-cwd>/<ccid>.jsonl.
type Reader struct{ base string }

// New devuelve un Reader sobre base (vacío = ~/.claude/projects).
func New(base string) (*Reader, error) {
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("history: home dir: %w", err)
		}
		base = filepath.Join(home, ".claude", "projects")
	}
	return &Reader{base: base}, nil
}

// DirParaCwd aplica el encoding observado del CLI real (verificado 2026-07-22 contra
// ~/.claude/projects: `/home/chalreme/Proyectos/luana-vitalia/vitalia` →
// `-home-chalreme-Proyectos-luana-vitalia-vitalia`): todo carácter no alfanumérico → `-`.
func DirParaCwd(cwd string) string {
	var b strings.Builder
	for _, r := range cwd {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	return b.String()
}

// lineaJSONL es el subset de una línea del corpus que el historial necesita.
type lineaJSONL struct {
	Type    string `json:"type"`
	Message *struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// Turnos reconstruye los turnos user/assistant de UNA JSONL (un ClaudeSessionID). Los
// bloques sin texto (tool_use/tool_result) se saltan — el historial es la conversación,
// no el trace. JSONL ausente → error ErrNotExist envuelto: el caller lo reporta honesto
// («esa conversación ya no está en disco»), jamás inventa.
func (r *Reader) Turnos(cwd, claudeSessionID string) ([]domain.Turn, error) {
	ruta := filepath.Join(r.base, DirParaCwd(cwd), claudeSessionID+".jsonl")
	f, err := os.Open(ruta) //nolint:gosec // G304: base propia (~/.claude/projects) + id de sesión del registro.
	if err != nil {
		return nil, fmt.Errorf("history: %s: %w", claudeSessionID, err)
	}
	defer f.Close() //nolint:errcheck // lectura pura.

	var turnos []domain.Turn
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<22) // líneas grandes (mensajes largos + tools).
	for sc.Scan() {
		var l lineaJSONL
		if err := json.Unmarshal(sc.Bytes(), &l); err != nil {
			continue // línea rara: se salta, no rompe el historial entero.
		}
		if (l.Type != "user" && l.Type != "assistant") || l.Message == nil {
			continue
		}
		texto := textoDe(l.Message.Content)
		if strings.TrimSpace(texto) == "" {
			continue
		}
		rol := domain.RolUser
		if l.Type == "assistant" {
			rol = domain.RolAssistant
		}
		turnos = append(turnos, domain.Turn{Rol: rol, Text: texto})
	}
	if err := sc.Err(); err != nil {
		return turnos, fmt.Errorf("history: leer %s: %w", claudeSessionID, err)
	}
	return turnos, nil
}

// Existe reporta si la JSONL de una sesión CC sigue en disco.
func (r *Reader) Existe(cwd, claudeSessionID string) bool {
	_, err := os.Stat(filepath.Join(r.base, DirParaCwd(cwd), claudeSessionID+".jsonl"))
	return !errors.Is(err, fs.ErrNotExist)
}

// textoDe extrae el texto de un content: string directo o bloques [{type:text,text}].
func textoDe(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var bloques []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &bloques); err != nil {
		return ""
	}
	var partes []string
	for _, b := range bloques {
		if b.Type == "text" && b.Text != "" {
			partes = append(partes, b.Text)
		}
	}
	return strings.Join(partes, "\n")
}
