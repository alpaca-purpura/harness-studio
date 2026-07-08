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
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alpacapurpura/arnesia/internal/domain"
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

// SpawnArgs returns the argv (sans binary) a SpawnOpts materializes. Exported so the
// fitness tests (arch/fitness) can assert the permission/turn-cap flags without
// spawning a real process — the flags ARE the enforcement surface.
func SpawnArgs(opts ports.SpawnOpts) []string {
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
	// --max-turns caps the agent loop of every turn: a hijacked or looping conductor must
	// not run unbounded (boundary permisos-gui `max-turns-siempre`, headless-sdk
	// `headless-max-turns`; both error-severity). 0 means the caller left it unset.
	if opts.MaxTurns > 0 {
		args = append(args, "--max-turns", strconv.Itoa(opts.MaxTurns))
	}
	// Inyección de doctrina (HS-11 puente 2, research firmado HS-10): los cuerpos ①+②
	// entran por FLAGS, session-scoped, desde dirs de la app — jamás se escribe nada en
	// el árbol del arnés (② ↛ ③). Sin --bare: la suscripción del usuario queda intacta.
	for _, d := range opts.Injection.PluginDirs {
		args = append(args, "--plugin-dir", d)
	}
	if f := opts.Injection.SystemPromptFile; f != "" {
		args = append(args, "--append-system-prompt-file", f)
	}
	for _, d := range opts.Injection.AddDirs {
		args = append(args, "--add-dir", d)
	}
	return append(args, permissionArgs(opts.Permisos)...)
}

// escrituraDirecta are the file-mutating tools that NEVER ride --allowedTools: even when
// the role's set allows them, each write goes through the GUI diff-approval
// (control_request → Dock), per boundary permisos-gui `write-requiere-aprobacion` +
// `allowedtools-readonly`. The role's allow means "approvable by this role", not "auto".
var escrituraDirecta = map[string]bool{
	"Write": true, "Edit": true, "MultiEdit": true, "NotebookEdit": true,
}

// permissionArgs materializes a role-derived PermissionSet in CC-native flags — the
// sanctioned surface per knowledge/elements/settings-permissions.md L1:
//
//   - `--permission-mode default` — deny-by-default «Manual» (L1.4); never bypass.
//   - `--allowedTools` — ONLY the genuinely read-only part of the role's allow
//     (escrituraDirecta filtered out; auto-approved tools never reach the callback).
//   - `--disallowedTools` — the role's hard deny (deny > ask > allow, enforced by CC
//     outside the model's reasoning).
//   - `--permission-prompt-tool stdio` — routes every non-pre-approved tool to the
//     control channel (`control_request:can_use_tool`), which the adapter forwards to
//     the daemon (research fase3 §frente B command line).
//
// GAP honesto: `Ask` has no dedicated CC flag — it is realized by NOT pre-approving +
// prompt-tool stdio (deny-by-default posture: unlisted/ask tools hit the control
// channel). `TTL` maps to no flag either: the ephemeral grants live in the daemon
// (SessionService), not in the CLI. A zero-value set emits NO flags (Dock unchanged).
func permissionArgs(ps domain.PermissionSet) []string {
	if ps.Rol == "" && len(ps.Allow) == 0 && len(ps.Ask) == 0 && len(ps.Deny) == 0 && ps.TTL == 0 {
		return nil
	}
	args := []string{"--permission-mode", "default"}
	var readOnly []string
	for _, tool := range ps.Allow {
		if !escrituraDirecta[tool] {
			readOnly = append(readOnly, tool)
		}
	}
	if len(readOnly) > 0 {
		args = append(args, "--allowedTools", strings.Join(readOnly, ","))
	}
	if len(ps.Deny) > 0 {
		args = append(args, "--disallowedTools", strings.Join(ps.Deny, ","))
	}
	return append(args, "--permission-prompt-tool", "stdio")
}

// Spawn starts a persistent conductor in streaming stream-json mode and returns the
// live session. The subprocess stays alive across turns until Close.
func (c *Conductor) Spawn(ctx context.Context, opts ports.SpawnOpts) (ports.AgentSession, error) {
	// The binary is the operator-configured local `claude` (conductor pattern,
	// local-first) and the args are built right here — never remote input.
	cmd := exec.CommandContext(ctx, c.bin, SpawnArgs(opts)...) //nolint:gosec // G204: c.bin is local daemon configuration (the --claude flag), never external input.
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
	// Handshake `initialize` (VERIFICADO contra claude 2.1.204, paquete chat-cc-funcional):
	// sin él, el binario NO emite `can_use_tool` — auto-deniega en silencio y el
	// human-in-the-loop jamás dispara. El SDK oficial lo manda siempre; nosotros también.
	// La respuesta (control_response del CLI) se ignora hoy — honesto: subscriptionType/
	// comandos quedan sin consumir hasta que alguien los necesite.
	if err := s.sendInitialize(); err != nil {
		_ = s.Close()
		return nil, err
	}
	return s, nil
}

// sendInitialize writes the app→CLI initialize control_request that arms the control
// channel (can_use_tool asks only flow after it).
func (s *ccSession) sendInitialize() error {
	n := s.ctrlSeq.Add(1)
	line, err := json.Marshal(ctrlRequestEnvelope{
		Type:      "control_request",
		RequestID: fmt.Sprintf("req_%d_arnesia", n),
		Request:   map[string]any{"subtype": "initialize"},
	})
	if err != nil {
		return fmt.Errorf("claudecode: marshal initialize: %w", err)
	}
	line = append(line, '\n')
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if _, werr := s.stdin.Write(line); werr != nil {
		return fmt.Errorf("claudecode: write initialize: %w", werr)
	}
	return nil
}

// ccSession is a live conductor subprocess implementing ports.AgentSession.
type ccSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	events chan ports.AgentEvent

	sendMu    sync.Mutex   // serializes writes to stdin.
	ctrlSeq   atomic.Int64 // request_id counter for daemon→CLI control_requests.
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

// ctrlAllow / ctrlDeny are the inner permission results of a control_response. The wire
// (verified: SDK oficial + claude-code-chat + vibe-kanban, investigación 2026-07-08
// §Frente ①) REQUIRES updatedInput on allow and message on deny; toolUseID echoes the
// ask's tool_use id.
type ctrlAllow struct {
	Behavior     string          `json:"behavior"`
	UpdatedInput json.RawMessage `json:"updatedInput"`
	ToolUseID    string          `json:"toolUseID,omitempty"`
}

type ctrlDeny struct {
	Behavior  string `json:"behavior"`
	Message   string `json:"message"`
	ToolUseID string `json:"toolUseID,omitempty"`
}

// ctrlResponseEnvelope is the stdin frame answering a control_request.
type ctrlResponseEnvelope struct {
	Type     string           `json:"type"`
	Response ctrlResponseBody `json:"response"`
}

type ctrlResponseBody struct {
	Subtype   string `json:"subtype"`
	RequestID string `json:"request_id"`
	Response  any    `json:"response"`
}

// controlResponseLine renders one control_response NDJSON line.
//
// HONESTIDAD sobre el wire format: el shape del canal control del binario `claude` está
// semi-documentado (oficial solo para el Agent SDK; anthropics/claude-code#24594 sigue
// abierto). Este envelope — {"type":"control_response","response":{subtype:"success",
// request_id, response:{behavior:"allow",updatedInput}|{behavior:"deny",message}}} — es
// el que implementan los SDKs oficiales y el que el research del repo cementó
// (research/2026-07-05-arquitectura-fase3.md §frente B ·
// research/2026-07-06-deuda-backend-arch.md item 2: «{behavior, updatedInput?,
// message?} — verificar el exacto al implementar contra el binario»). Best-effort
// verificado contra los SDKs; el cableado se prueba con fakes (arch/fitness).
func controlResponseLine(requestID string, d ports.ControlDecision) ([]byte, error) {
	var inner any
	if d.Allow {
		upd := json.RawMessage(d.UpdatedInput)
		if len(upd) == 0 {
			upd = json.RawMessage("{}")
		}
		inner = ctrlAllow{Behavior: "allow", UpdatedInput: upd, ToolUseID: d.ToolUseID}
	} else {
		msg := d.Message
		if msg == "" {
			msg = "denegado por el operador" // message es requerido en el wire del deny.
		}
		inner = ctrlDeny{Behavior: "deny", Message: msg, ToolUseID: d.ToolUseID}
	}
	line, err := json.Marshal(ctrlResponseEnvelope{
		Type:     "control_response",
		Response: ctrlResponseBody{Subtype: "success", RequestID: requestID, Response: inner},
	})
	if err != nil {
		return nil, fmt.Errorf("claudecode: marshal control_response: %w", err)
	}
	return append(line, '\n'), nil
}

// RespondControl answers a forwarded control_request over stdin. The adapter never
// decides — the daemon (role permission-set + Dock approval) already did; this only
// speaks the wire format (see controlResponseLine for the honest-uncertainty note).
func (s *ccSession) RespondControl(_ context.Context, requestID string, d ports.ControlDecision) error {
	line, err := controlResponseLine(requestID, d)
	if err != nil {
		return err
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if _, err := s.stdin.Write(line); err != nil {
		return fmt.Errorf("claudecode: write control_response: %w", err)
	}
	return nil
}

// ctrlRequestEnvelope is a control_request the DAEMON sends to the CLI (app→CLI
// direction: interrupt, set_permission_mode…). Mirror of the CLI→app envelope.
type ctrlRequestEnvelope struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Request   map[string]any `json:"request"`
}

// Interrupt stops the in-flight turn in-band: control_request subtype=interrupt over
// stdin (requires --input-format stream-json, which every spawn sets). The subprocess
// answers with its result frame and stays alive for the next turn (RF-116). Wire
// verified against the official SDK + reference implementations (investigación
// 2026-07-08 §Frente ①).
func (s *ccSession) Interrupt(_ context.Context) error {
	n := s.ctrlSeq.Add(1)
	line, err := json.Marshal(ctrlRequestEnvelope{
		Type:      "control_request",
		RequestID: fmt.Sprintf("req_%d_arnesia", n),
		Request:   map[string]any{"subtype": "interrupt"},
	})
	if err != nil {
		return fmt.Errorf("claudecode: marshal interrupt: %w", err)
	}
	line = append(line, '\n')
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	if _, err := s.stdin.Write(line); err != nil {
		return fmt.Errorf("claudecode: write interrupt: %w", err)
	}
	return nil
}

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
	RequestID  string          `json:"request_id"`
	Request    *ctrlRequest    `json:"request"`
}

// ctrlRequest is the inner payload of a control_request frame (can_use_tool).
type ctrlRequest struct {
	Subtype   string          `json:"subtype"`
	ToolName  string          `json:"tool_name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
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

// emit sends ev on the events channel. The send is BLOCKING on purpose: dropping a frame
// here would strand the session (a lost `result` leaves the dock stuck "streaming"), which
// is exactly what boundary sesion-viva-consistente `sin-perdida-silenciosa` forbids. It is
// safe to block because pump is the sole caller and the SessionService consumer always
// drains Events() (its own downstream — the SSE broker — never blocks, it sheds slow
// subscribers instead). Back-pressure here correctly slows the reader rather than losing
// data. No post-close guard is needed: pump closes s.events only after its loop returns, so
// emit is never called on a closed channel.
func (s *ccSession) emit(ev ports.AgentEvent) {
	s.events <- ev
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
		return ports.AgentEvent{Kind: ports.EventResult, Text: f.Result, Subtype: f.Subtype, CtxPct: ctxPct(f), Raw: line}, true

	case "control_request":
		// Forward can_use_tool VERBATIM instead of discarding it (Fase E): the daemon —
		// role permission-set + Dock human-in-the-loop — resolves it and answers via
		// RespondControl. Other control subtypes (hook_callback, mcp_message…) stay
		// out of scope: not forwarded, honestly ignored.
		if f.Request != nil && f.Request.Subtype == "can_use_tool" {
			return ports.AgentEvent{
				Kind:      ports.EventControlRequest,
				RequestID: f.RequestID,
				Tool:      f.Request.ToolName,
				Input:     f.Request.Input,
				ToolUseID: f.Request.ToolUseID,
				Raw:       line,
			}, true
		}
		return ports.AgentEvent{}, false

	default:
		return ports.AgentEvent{}, false
	}
}

// assistantText concatenates the text blocks of a complete assistant message.
func assistantText(m *assistantMsg) string {
	if m == nil {
		return ""
	}
	var out strings.Builder
	for _, b := range m.Content {
		if b.Type == "text" {
			out.WriteString(b.Text)
		}
	}
	return out.String()
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
