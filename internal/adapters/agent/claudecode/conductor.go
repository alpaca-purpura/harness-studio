// Package claudecode implements ports.AgentPort by spawning the local `claude` binary
// as a persistent subprocess and speaking stream-json over its stdin/stdout — the
// conductor pattern (I-76 / OBS-16 / OBS-18): Go owns the process and drives it, the
// agent is just an adapter.
//
// This package is the sole owner of the Claude Code protocol. It translates raw
// stream-json frames into normalized ports.AgentEvent values; nothing upstream
// re-decodes the wire format (docs/architecture/boundaries/conductor-no-parsea-jsonl.md).
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
	"os"
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
	// EnvExtra son las variables que se AGREGAN al entorno de cada subproceso. Hoy `cmd.Env`
	// es nil (el subproceso hereda el del daemon), así que esto es un cambio de CÓDIGO y no
	// de configuración: sin él no hay forma de encender la telemetría del spawn.
	EnvExtra []string
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
// fitness tests (docs/architecture/fitness) can assert the permission/turn-cap flags without
// spawning a real process — the flags ARE the enforcement surface.
func SpawnArgs(opts ports.SpawnOpts) []string {
	args := []string{
		"-p",
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--include-partial-messages",
		"--verbose",
		// Aislamiento de superficie de config (HS-17 D3): fijo, jamás configurable por el
		// caller — nadie debe poder reintroducir "user" (~/.claude/settings.json del
		// OPERADOR, con SUS enabledPlugins/hooks personales). El CLAUDE.md propio del cwd
		// del arnés sobrevive de todos modos (discovery propio, independiente de este flag).
		"--setting-sources", "project,local",
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
	// Inyección de doctrina (HS-11 puente 2, investigación firmada HS-10): los cuerpos ①+②
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
	// Aislamiento de superficie de config (HS-17 D2): todo spawn con Injection poblada
	// corta el MCP a SOLO lo que el kit propio declara — nunca el ~/.claude del
	// operador ni sus conectores de cuenta (claude.ai). Ortogonal a auth: MCP no toca
	// OAuth/keychain.
	if f := opts.Injection.MCPConfigFile; f != "" {
		args = append(args, "--mcp-config", f, "--strict-mcp-config")
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
// sanctioned surface per docs/architecture/knowledge/elements/settings-permissions.md L1:
//
//   - `--permission-mode default` — deny-by-default «Manual» (L1.4); never bypass.
//   - `--allowedTools` — ONLY the genuinely read-only part of the role's allow
//     (escrituraDirecta filtered out; auto-approved tools never reach the callback).
//   - `--disallowedTools` — the role's hard deny (deny > ask > allow, enforced by CC
//     outside the model's reasoning).
//   - `--permission-prompt-tool stdio` — routes every non-pre-approved tool to the
//     control channel (`control_request:can_use_tool`), which the adapter forwards to
//     the daemon (design record docs/product/research/2026-07-05-arquitectura-fase3.md §frente B command line).
//
// GAP honesto: `Ask` has no dedicated CC flag — it is realized by NOT pre-approving +
// prompt-tool stdio (deny-by-default posture: unlisted/ask tools hit the control
// channel). `TTL` maps to no flag either: the ephemeral grants live in the daemon
// (SessionService), not in the CLI.
//
// El canal va SIEMPRE cableado (DD-1, deuda D del dogfood): un set de valor cero emite
// igual `--permission-mode default --permission-prompt-tool stdio` — nada pre-aprobado,
// TODO pasa por la tarjeta del panel. Antes el set vacío emitía cero flags y una sesión
// sobre material sin sello (sin rol) corría headless SIN canal: CC auto-negaba Write y el
// permiso jamás llegaba al Dock — read-only de facto, forja conversacional imposible.
func permissionArgs(ps domain.PermissionSet) []string {
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

// SpawnEnv arma las variables de entorno que instrumentan un subproceso de Claude Code (S1).
//
// 🔴 **`OTEL_LOGS_EXPORTER=otlp` es OBLIGATORIA, y está MEDIDA.** Sin ella llegan **0** log
// events contra 2 del control positivo (ANEXO H10.4, misma corrida, mismo receptor,
// marcadores distintos). Como el canal primario es `/v1/logs` —por donde viaja
// `api_request`, o sea el dinero— omitirla **apaga la señal de dinero entera sin un solo
// error visible**, que es el peor modo de falla posible de este módulo.
//
// Lo mismo vale, por otra razón, para `OTEL_EXPORTER_OTLP_PROTOCOL=http/json`: el default de
// Claude Code es gRPC en el 4317, así que sin esa línea el exportador habla un protocolo que
// nuestro receptor no entiende y tampoco llega nada.
//
// El token que viaja es el de INGESTA, jamás el de la API: filtrar el primero concede
// «escribime telemetría»; filtrar el segundo concede «conducí un agente con acceso al
// filesystem».
func SpawnEnv(opts ports.SpawnOpts, tokenIngesta, endpoint string, atrib AtribucionSpawn) []string {
	if endpoint == "" {
		return nil // sin receptor no se instrumenta: apuntar a la nada solo agrega latencia.
	}
	var env []string
	for _, kv := range VariablesTelemetria(endpoint, atrib) {
		env = append(env, kv.K+"="+kv.V)
	}
	if tokenIngesta != "" {
		// Por header, nunca por query string: un query string va a los logs de cualquier
		// proxy que se interponga.
		env = append(env, "OTEL_EXPORTER_OTLP_HEADERS=x-arnesia-token="+tokenIngesta)
	}
	return env
}

// VarTelemetria es una variable del contrato de instrumentación, con su nombre y su valor.
type VarTelemetria struct{ K, V string }

// VariablesTelemetria es **la fuente única** del juego de variables que enciende la telemetría
// de Claude Code (D26.2 · A20 opción A).
//
// Existe porque hay DOS caminos para el mismo contrato y tienen que decir lo mismo:
//
//  1. **S1** — ArnesIA lanza el subproceso y le pasa el entorno (`SpawnEnv`);
//  2. **S2 instrumentado** — el arnés lleva su propio `.claude/settings.json` y quien corre
//     Claude Code a mano queda instrumentado igual (`BloqueEnvSettings`).
//
// Si los dos juegos divergen, uno de los dos escenarios manda una señal que el receptor no
// entiende **y no hay ningún error visible**: simplemente no llega nada. Por eso hay una sola
// lista y un test que compara el archivo shipeado contra ella.
//
// 🔴 **El token de ingesta NO está acá.** No es un olvido: Claude Code no expande `${VAR}`
// dentro del bloque `env` (H10.2), y un token literal en un archivo versionado es publicar un
// secreto. La salida es A22 — `/v1/*` acepta sin token bajo Host loopback. `SpawnEnv` sí lo
// agrega porque ahí el valor no se versiona: viaja en memoria al subproceso.
func VariablesTelemetria(endpoint string, atrib AtribucionSpawn) []VarTelemetria {
	vs := []VarTelemetria{
		{"CLAUDE_CODE_ENABLE_TELEMETRY", "1"},
		// OBLIGATORIA — verificada en vivo. Ver el comentario de `SpawnEnv`.
		{"OTEL_LOGS_EXPORTER", "otlp"},
		{"OTEL_METRICS_EXPORTER", "otlp"},
		// http/json y NO http/protobuf: es lo que hace barato al decodificador
		// (+0,49 MB contra +10,79 MB medidos) y lo que nuestro receptor habla.
		{"OTEL_EXPORTER_OTLP_PROTOCOL", "http/json"},
		// SIN `/v1/...`: el exportador concatena la ruta por spec.
		{"OTEL_EXPORTER_OTLP_ENDPOINT", endpoint},
		{"OTEL_METRIC_EXPORT_INTERVAL", "10000"},
		{"OTEL_LOGS_EXPORT_INTERVAL", "5000"},
	}
	// El vector de atribución: viaja COPIADO en cada log record y en cada punto (verificado),
	// que es lo que hace posible la atribución exacta sin heurísticas.
	if ra := atrib.recursoOTel(); ra != "" {
		vs = append(vs, VarTelemetria{"OTEL_RESOURCE_ATTRIBUTES", ra})
	}
	return vs
}

// BloqueEnvSettings arma el bloque `env` de un `.claude/settings.json` — la vía VERIFICADA
// (ANEXO H9) para instrumentar a quien corre Claude Code a mano sobre el árbol del arnés.
//
// **A20 = opción A** (D26.2): el archivo vive en el repo del propio arnés y viaja en el
// paquete. La opción B —escribirlo en el proyecto del usuario— no se construye: es escritura
// en árbol ajeno, y eso se pide, no se hace.
//
// ⚡ Un `plugin.json` **no** puede aportar este bloque: medido, 0 payloads contra 2 del control
// positivo (H10.1). No existe la auto-instrumentación al instalar.
func BloqueEnvSettings(endpoint string, atrib AtribucionSpawn) map[string]string {
	out := map[string]string{}
	for _, kv := range VariablesTelemetria(endpoint, atrib) {
		out[kv.K] = kv.V
	}
	return out
}

// AtribucionSpawn son las etiquetas que identifican QUÉ se está corriendo. `Caja` y `Corrida`
// solo se llenan en el spawn del conductor de una caja; en el chat quedan vacías y la
// atribución es a nivel de sesión — que es la verdad, no una degradación.
type AtribucionSpawn struct {
	ArnesID       string
	InstalacionID string
	CajaID        string
	CorridaID     string
}

func (a AtribucionSpawn) recursoOTel() string {
	var partes []string
	for _, p := range []struct{ k, v string }{
		{"arnesia.arnes", a.ArnesID},
		{"arnesia.instalacion", a.InstalacionID},
		{"arnesia.caja", a.CajaID},
		{"arnesia.corrida", a.CorridaID},
	} {
		if p.v == "" {
			continue // una etiqueta vacía no se manda: sería ruido con forma de dato.
		}
		partes = append(partes, p.k+"="+p.v)
	}
	return strings.Join(partes, ",")
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
	// El entorno del subproceso: el del daemon MÁS lo nuestro. Se agrega, no se reemplaza —
	// el subproceso necesita PATH, HOME y lo demás para funcionar.
	if len(c.EnvExtra) > 0 {
		cmd.Env = append(os.Environ(), c.EnvExtra...)
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
	// tool_use blocks (CH-D2): the tool name + raw input the activity target derives
	// from. omitempty es OBLIGATORIO: este struct también SE MANDA (userMsg por stdin) y
	// un campo extra en un bloque text rompe la API con 400 (regresión cazada en vivo).
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
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
// el que implementan los SDKs oficiales y el que la investigación del repo cementó
// (docs/product/research/2026-07-05-arquitectura-fase3.md §frente B ·
// docs/product/research/2026-07-06-deuda-backend-arch.md item 2: «{behavior, updatedInput?,
// message?} — verificar el exacto al implementar contra el binario»). Best-effort
// verificado contra los SDKs; el cableado se prueba con fakes (docs/architecture/fitness).
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
	TotalCost  float64         `json:"total_cost_usd"`
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
	// Usage del API call que produjo ESTE mensaje: la ocupación real de la ventana en ese
	// momento (a diferencia del usage acumulado del frame result — ver ctxPct).
	Usage *usage `json:"usage"`
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
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	// 🔴 CacheCreation es EL split por vencimiento, y el `result` lo trae desde siempre —
	// hasta hoy se tiraba. Es la única fuente del dato: no viaja por OTel, así que solo se ve
	// cuando ArnesIA es el proceso padre. Sin él, el costo del cache write no se puede cotizar
	// al tramo correcto (y asumir el barato subestima un 33 %, medido).
	CacheCreation *struct {
		Ephemeral5m int `json:"ephemeral_5m_input_tokens"`
		Ephemeral1h int `json:"ephemeral_1h_input_tokens"`
	} `json:"cache_creation"`
	ServiceTier string `json:"service_tier"`
	Speed       string `json:"speed"`
}

type cwin struct {
	ContextWindow int `json:"contextWindow"`
	// Lo que el `result` ya trae por modelo y hasta hoy se tiraba.
	CostUSD        float64 `json:"costUSD"`
	CanonicalModel string  `json:"canonicalModel"`
	Provider       string  `json:"provider"`
	InputTokens    int     `json:"inputTokens"`
	OutputTokens   int     `json:"outputTokens"`
}

// pump reads stdout NDJSON, translates each frame, and fans it out on events. It uses
// a bufio.Reader (not Scanner) because init frames exceed Scanner's default 64KB line
// cap. It closes events when stdout ends.
func (s *ccSession) pump(stdout io.Reader) {
	defer close(s.events)
	r := bufio.NewReaderSize(stdout, 1<<20)

	// lastUsage: usage del ÚLTIMO API call del turno (frames assistant) — la ocupación
	// REAL de la ventana. El usage del frame result es ACUMULADO (cache_read re-contado
	// por cada tool-call) y sobreestima brutalmente (100 % espurio medido en vivo,
	// paquete mejorar-arnes-conversando T4/T7).
	var lastUsage *usage
	for {
		line, err := readLine(r)
		if len(line) > 0 {
			for _, ev := range translate(line, &lastUsage) {
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

// translate convierte una línea NDJSON en cero o más AgentEvents (nil = frame que se
// ignora adrede: hooks, rate limits, status, deltas no-texto). Un frame `assistant` puede
// producir VARIOS eventos en orden de bloques (CH-D2/CH-D3) — ver assistantEvents.
// lastUsage (estado del pump, un solo goroutine) captura el usage del último frame
// assistant del turno — el que ctxPct usa en el result (RF-194): el usage del result es
// acumulado y miente sobre la ventana. nil lastUsage es legal (tests de frames sueltos).
func translate(line []byte, lastUsage **usage) []ports.AgentEvent {
	var f rawFrame
	if err := json.Unmarshal(line, &f); err != nil {
		// A malformed line is not fatal; skip it.
		return nil
	}

	switch f.Type {
	case "system":
		if f.Subtype == "init" {
			return []ports.AgentEvent{{
				Kind:            ports.EventInit,
				ClaudeSessionID: f.SessionID,
				Model:           f.Model,
				Raw:             line,
			}}
		}
		return nil

	case "stream_event":
		if f.Event != nil && f.Event.Type == "content_block_delta" &&
			f.Event.Delta != nil && f.Event.Delta.Type == "text_delta" {
			return []ports.AgentEvent{{Kind: ports.EventDelta, Text: f.Event.Delta.Text, Raw: line}}
		}
		return nil

	case "assistant":
		if lastUsage != nil && f.Message != nil && f.Message.Usage != nil {
			*lastUsage = f.Message.Usage
		}
		return assistantEvents(f.Message, line)

	case "result":
		var last *usage
		if lastUsage != nil {
			last = *lastUsage
		}
		return []ports.AgentEvent{{
			Kind: ports.EventResult, Text: f.Result, Subtype: f.Subtype,
			CtxPct: ctxPct(f, last), Uso: parseResult(f, last), Raw: line,
		}}

	case "control_request":
		// Forward can_use_tool VERBATIM instead of discarding it (Fase E): the daemon —
		// role permission-set + Dock human-in-the-loop — resolves it and answers via
		// RespondControl. Other control subtypes (hook_callback, mcp_message…) stay
		// out of scope: not forwarded, honestly ignored.
		if f.Request != nil && f.Request.Subtype == "can_use_tool" {
			return []ports.AgentEvent{{
				Kind:      ports.EventControlRequest,
				RequestID: f.RequestID,
				Tool:      f.Request.ToolName,
				Input:     f.Request.Input,
				ToolUseID: f.Request.ToolUseID,
				Raw:       line,
			}}
		}
		return nil

	default:
		return nil
	}
}

// assistantEvents desarma un mensaje assistant completo en eventos EN ORDEN de bloques
// (CH-D2/CH-D3): thinking → actividad «thinking» (jamás su contenido), texto contiguo →
// UNA burbuja (EventMessage), tool_use → actividad con el blanco legible del input.
func assistantEvents(m *assistantMsg, raw []byte) []ports.AgentEvent {
	if m == nil {
		return nil
	}
	var out []ports.AgentEvent
	var sb strings.Builder
	flush := func() {
		if sb.Len() > 0 {
			out = append(out, ports.AgentEvent{Kind: ports.EventMessage, Text: sb.String(), Raw: raw})
			sb.Reset()
		}
	}
	for _, b := range m.Content {
		switch b.Type {
		case "text":
			sb.WriteString(b.Text)
		case "thinking":
			flush()
			out = append(out, ports.AgentEvent{Kind: ports.EventActivity, Tool: "thinking", Raw: raw})
		case "tool_use":
			flush()
			out = append(out, ports.AgentEvent{Kind: ports.EventActivity, Tool: b.Name, Text: blanco(b.Input), Raw: raw})
		}
	}
	flush()
	return out
}

// blanco extrae el blanco legible del input de un tool_use — lo que la tarjeta de
// actividad muestra al lado del tool. Vacío si no hay clave conocida.
func blanco(input json.RawMessage) string {
	var m map[string]any
	if json.Unmarshal(input, &m) != nil {
		return ""
	}
	for _, k := range []string{"file_path", "path", "command", "pattern", "url", "query", "description"} {
		if v, ok := m[k].(string); ok && v != "" {
			if r := []rune(v); len(r) > 80 {
				return string(r[:80]) + "…"
			}
			return v
		}
	}
	return ""
}

// parseResult extrae el uso del turno del frame `result`.
//
// **Se extrae con este nombre a propósito**: es el único lugar del árbol que decodifica ese
// frame, y no se crea un segundo parser de stream-json en `adapters/telemetria/` (decisión
// A5). Dos decodificadores del mismo frame driftan por separado, y el adaptador del agente es
// el dueño del protocolo — un adaptador de agente por runtime, no dos.
//
// Devuelve nil cuando el frame no trae uso: **no se fabrica un uso en ceros**, que se leería
// como «este turno no consumió nada».
func parseResult(f rawFrame, last *usage) *domain.UsoDelTurno {
	u := f.Usage
	if u == nil {
		u = last
	}
	if u == nil {
		return nil
	}
	uso := &domain.UsoDelTurno{
		Entrada:        int64(u.InputTokens),
		Salida:         int64(u.OutputTokens),
		CacheLectura:   int64(u.CacheReadInputTokens),
		CacheEscritura: int64(u.CacheCreationInputTokens),
		ServiceTier:    u.ServiceTier,
		Speed:          u.Speed,
		CostoUSD:       f.TotalCost,
		Modelo:         f.Model,
	}
	if u.CacheCreation != nil {
		// El split. Un 0 acá es un DATO —el runtime lo dijo—, no una ausencia: por eso los
		// campos no son punteros y por eso `ephemeral_5m = 0` con `ephemeral_1h = 8257` es
		// información, no un hueco.
		uso.Ephemeral5m = int64(u.CacheCreation.Ephemeral5m)
		uso.Ephemeral1h = int64(u.CacheCreation.Ephemeral1h)
	}
	// `modelUsage` trae el costo, el nombre canónico y el proveedor por modelo.
	if cw, ok := f.ModelUsage[f.Model]; ok {
		aplicarModelUsage(uso, cw)
	} else {
		for _, cw := range f.ModelUsage {
			aplicarModelUsage(uso, cw)
			break
		}
	}
	return uso
}

func aplicarModelUsage(uso *domain.UsoDelTurno, cw cwin) {
	uso.ContextWindow = cw.ContextWindow
	uso.ModeloCanonico = cw.CanonicalModel
	uso.Proveedor = cw.Provider
	if uso.CostoUSD == 0 && cw.CostUSD > 0 {
		uso.CostoUSD = cw.CostUSD
	}
}

// ctxPct estimates context-window usage (0–100). Prefiere el usage del ÚLTIMO API call
// del turno (last, capturado de los frames assistant): esa ES la ocupación de la ventana.
// El usage del frame result es la SUMA de todos los API calls del turno (cache_read
// re-contado por tool-call) — solo se usa de fallback cuando ningún assistant trajo usage
// (RF-194; hallazgo del E2E T4: 100 % espurio en un turno con ~10 tool-calls).
func ctxPct(f rawFrame, last *usage) int {
	u := f.Usage
	if last != nil && (last.InputTokens+last.CacheReadInputTokens+last.CacheCreationInputTokens) > 0 {
		u = last
	}
	if u == nil {
		return 0
	}
	used := u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
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
