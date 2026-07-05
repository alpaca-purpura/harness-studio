// Package claudecode implements ports.AgentPort by spawning the local `claude` binary
// as a persistent subprocess and speaking stream-json over its stdin/stdout — the
// conductor pattern (I-76 / OBS-16 / OBS-18): Go owns the process and drives it, the
// agent is just an adapter.
//
// This package is the sole owner of the Claude Code protocol. It translates raw
// stream-json frames into normalized ports.AgentEvent values; nothing upstream
// re-decodes the wire format (arch/boundaries/conductor-no-parsea-jsonl.md).
//
// Wire contract (verified against claude 2.1.201):
//   - input  (stdin, NDJSON):  {"type":"user","message":{"role":"user","content":[{"type":"text","text":"…"}]}}
//   - output (stdout, NDJSON): system/init (session_id, model) · stream_event
//     (content_block_delta → text_delta) · assistant (full message) · result
//     (usage + modelUsage.contextWindow). hook_*/rate_limit frames are ignored.
//
// stdin is held open, so one process serves many turns; closing stdin (Close) ends it.
package claudecode

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os/exec"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// defaultContextWindow is used when a result frame omits modelUsage.contextWindow.
const defaultContextWindow = 200_000

// Conductor is the Claude Code adapter. bin is the path to the `claude` executable.
type Conductor struct {
	bin string
}

var _ ports.AgentPort = (*Conductor)(nil)

// New returns a Conductor that spawns the given `claude` binary (e.g. "claude").
func New(bin string) *Conductor {
	if bin == "" {
		bin = "claude"
	}
	return &Conductor{bin: bin}
}

// Spawn starts a persistent conductor in streaming stream-json mode and returns the
// live session. The subprocess stays alive across turns until Close.
func (c *Conductor) Spawn(ctx context.Context, opts ports.SpawnOpts) (ports.AgentSession, error) {
	args := []string{
		"-p",
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--include-partial-messages",
		"--verbose",
	}
	if opts.Resume != "" {
		args = append(args, "--resume", opts.Resume)
	}
	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	cmd := exec.CommandContext(ctx, c.bin, args...)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("claudecode: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("claudecode: stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("claudecode: stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("claudecode: start %s: %w", c.bin, err)
	}

	s := &ccSession{
		cmd:    cmd,
		stdin:  stdin,
		events: make(chan ports.AgentEvent, 64),
	}
	go s.logStderr(stderr)
	go s.pump(stdout)
	return s, nil
}

// ccSession is a live conductor subprocess implementing ports.AgentSession.
type ccSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan ports.AgentEvent

	sendMu    sync.Mutex // serializes writes to stdin.
	closeOnce sync.Once
	closeErr  error
}

var _ ports.AgentSession = (*ccSession)(nil)

// contentText is one content block of a user/assistant message.
type contentText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// userMsg is the stdin envelope for one user turn.
type userMsg struct {
	Type    string `json:"type"`
	Message struct {
		Role    string        `json:"role"`
		Content []contentText `json:"content"`
	} `json:"message"`
}

// Send streams one user turn to the subprocess. Writes are serialized so concurrent
// callers cannot interleave a JSON line.
func (s *ccSession) Send(_ context.Context, turn string) error {
	var m userMsg
	m.Type = "user"
	m.Message.Role = "user"
	m.Message.Content = []contentText{{Type: "text", Text: turn}}

	line, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("claudecode: marshal turn: %w", err)
	}
	line = append(line, '\n')

	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if _, err := s.stdin.Write(line); err != nil {
		return fmt.Errorf("claudecode: write turn: %w", err)
	}
	return nil
}

// Events returns the normalized event stream (closed when the subprocess exits).
func (s *ccSession) Events() <-chan ports.AgentEvent { return s.events }

// Close ends the subprocess: closing stdin signals EOF (a clean exit), then we wait.
func (s *ccSession) Close() error {
	s.closeOnce.Do(func() {
		_ = s.stdin.Close()
		s.closeErr = s.cmd.Wait()
	})
	return s.closeErr
}

// rawFrame is the minimal shape needed to route a stream-json line without decoding
// the whole protocol.
type rawFrame struct {
	Type       string          `json:"type"`
	Subtype    string          `json:"subtype"`
	SessionID  string          `json:"session_id"`
	Model      string          `json:"model"`
	Message    *assistantMsg   `json:"message"`
	Event      *innerEvent     `json:"event"`
	Result     string          `json:"result"`
	Usage      *usage          `json:"usage"`
	ModelUsage map[string]cwin `json:"modelUsage"`
}

type assistantMsg struct {
	Content []contentText `json:"content"`
}

type innerEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
}

type usage struct {
	InputTokens              int `json:"input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

type cwin struct {
	ContextWindow int `json:"contextWindow"`
}

// pump reads stdout NDJSON, translates each frame, and fans it out on events. It uses
// a bufio.Reader (not Scanner) because init frames exceed Scanner's default 64KB line
// cap. It closes events when stdout ends.
func (s *ccSession) pump(stdout io.Reader) {
	defer close(s.events)
	r := bufio.NewReaderSize(stdout, 1<<20)

	for {
		line, err := readLine(r)
		if len(line) > 0 {
			if ev, ok := translate(line); ok {
				s.emit(ev)
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.emit(ports.AgentEvent{Kind: ports.EventError, Text: err.Error()})
			}
			return
		}
	}
}

// emit sends ev, dropping it if the consumer is gone (channel buffer full and unread
// after the session was abandoned) rather than blocking the pump forever.
func (s *ccSession) emit(ev ports.AgentEvent) {
	select {
	case s.events <- ev:
	default:
		// Rare: consumer detached. Dropping keeps the subprocess drained; a
		// reconnecting client replays via SSE Last-Event-ID.
	}
}

// readLine returns one line without the trailing newline. A final unterminated line
// is returned with io.EOF.
func readLine(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if n := len(line); n > 0 && line[n-1] == '\n' {
		line = line[:n-1]
	}
	return line, err
}

// translate maps one raw stream-json line to a normalized event. ok is false for
// frames we deliberately ignore (hooks, rate limits, status, non-text deltas).
func translate(line []byte) (ports.AgentEvent, bool) {
	var f rawFrame
	if err := json.Unmarshal(line, &f); err != nil {
		// A malformed line is not fatal; skip it.
		return ports.AgentEvent{}, false
	}

	switch f.Type {
	case "system":
		if f.Subtype == "init" {
			return ports.AgentEvent{
				Kind:            ports.EventInit,
				ClaudeSessionID: f.SessionID,
				Model:           f.Model,
				Raw:             line,
			}, true
		}
		return ports.AgentEvent{}, false

	case "stream_event":
		if f.Event != nil && f.Event.Type == "content_block_delta" &&
			f.Event.Delta != nil && f.Event.Delta.Type == "text_delta" {
			return ports.AgentEvent{Kind: ports.EventDelta, Text: f.Event.Delta.Text, Raw: line}, true
		}
		return ports.AgentEvent{}, false

	case "assistant":
		return ports.AgentEvent{Kind: ports.EventMessage, Text: assistantText(f.Message), Raw: line}, true

	case "result":
		return ports.AgentEvent{Kind: ports.EventResult, Text: f.Result, CtxPct: ctxPct(f), Raw: line}, true

	default:
		return ports.AgentEvent{}, false
	}
}

// assistantText concatenates the text blocks of a complete assistant message.
func assistantText(m *assistantMsg) string {
	if m == nil {
		return ""
	}
	var out string
	for _, b := range m.Content {
		if b.Type == "text" {
			out += b.Text
		}
	}
	return out
}

// ctxPct estimates context-window usage (0–100) from a result frame: the prompt
// tokens that will occupy the window next turn over the model's context window.
func ctxPct(f rawFrame) int {
	if f.Usage == nil {
		return 0
	}
	used := f.Usage.InputTokens + f.Usage.CacheReadInputTokens + f.Usage.CacheCreationInputTokens
	window := defaultContextWindow
	if cw, ok := f.ModelUsage[f.Model]; ok && cw.ContextWindow > 0 {
		window = cw.ContextWindow
	} else {
		for _, cw := range f.ModelUsage {
			if cw.ContextWindow > 0 {
				window = cw.ContextWindow
				break
			}
		}
	}
	pct := int(math.Round(float64(used) / float64(window) * 100))
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return pct
}

// logStderr drains the subprocess stderr into the daemon log (diagnostics only).
func (s *ccSession) logStderr(stderr io.Reader) {
	r := bufio.NewReaderSize(stderr, 1<<16)
	for {
		line, err := readLine(r)
		if len(line) > 0 {
			slog.Debug("claudecode: stderr", "line", string(line))
		}
		if err != nil {
			return
		}
	}
}
