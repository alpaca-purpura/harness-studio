package claudecode

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// CH-D2/CH-D3 (paquete chat-dock-ux): un mensaje assistant completo se desarma EN ORDEN
// de bloques — thinking → actividad «thinking», texto contiguo → UNA burbuja
// (EventMessage), tool_use → actividad con el blanco legible del input.
func TestAssistantSeDesarmaEnBurbujasYActividad(t *testing.T) {
	line := []byte(`{"type":"assistant","message":{"content":[` +
		`{"type":"thinking","thinking":"…"},` +
		`{"type":"text","text":"Voy a mirar el hook."},` +
		`{"type":"tool_use","name":"Read","input":{"file_path":"hooks/sellar.sh"}},` +
		`{"type":"text","text":"Listo, lo reparo."}]}}`)
	evs := translate(line, nil)
	want := []struct {
		kind ports.AgentEventKind
		tool string
		text string
	}{
		{ports.EventActivity, "thinking", ""},
		{ports.EventMessage, "", "Voy a mirar el hook."},
		{ports.EventActivity, "Read", "hooks/sellar.sh"},
		{ports.EventMessage, "", "Listo, lo reparo."},
	}
	if len(evs) != len(want) {
		t.Fatalf("got %d eventos, want %d (%+v)", len(evs), len(want), evs)
	}
	for i, w := range want {
		if evs[i].Kind != w.kind || evs[i].Tool != w.tool || evs[i].Text != w.text {
			t.Errorf("evs[%d] = kind %q tool %q text %q, want %+v", i, evs[i].Kind, evs[i].Tool, evs[i].Text, w)
		}
	}

	// El blanco de un Bash sale de `command` y se recorta a algo legible.
	largo := strings.Repeat("x", 200)
	evs = translate([]byte(`{"type":"assistant","message":{"content":[`+
		`{"type":"tool_use","name":"Bash","input":{"command":"`+largo+`"}}]}}`), nil)
	if len(evs) != 1 || evs[0].Tool != "Bash" || len([]rune(evs[0].Text)) > 81 {
		t.Errorf("blanco de Bash = %q (%d eventos), want command recortado", evs[0].Text, len(evs))
	}
}

// Regresión (cazada en E2E vivo): contentText se comparte entre PARSEAR frames y MANDAR
// el turno user por stdin — los campos de tool_use jamás pueden viajar en un bloque text
// (la API responde 400 «Extra inputs are not permitted»).
func TestUserTurnWireSinCamposExtra(t *testing.T) {
	b, err := json.Marshal(contentText{Type: "text", Text: "hola"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := string(b); got != `{"type":"text","text":"hola"}` {
		t.Errorf("bloque text por el wire = %s, want sin campos extra", got)
	}
}

func TestTranslateForwardsControlRequest(t *testing.T) {
	line := []byte(`{"type":"control_request","request_id":"req-7",` +
		`"request":{"subtype":"can_use_tool","tool_name":"Write",` +
		`"input":{"file_path":"spec.md","content":"x"}}}`)
	evs := translate(line, nil)
	if len(evs) != 1 {
		t.Fatal("un control_request:can_use_tool debe reenviarse, no descartarse (Fase E)")
	}
	ev := evs[0]
	if ev.Kind != ports.EventControlRequest {
		t.Errorf("kind = %q, want control_request", ev.Kind)
	}
	if ev.RequestID != "req-7" || ev.Tool != "Write" {
		t.Errorf("request_id/tool = %q/%q, want req-7/Write", ev.RequestID, ev.Tool)
	}
	var input map[string]any
	if err := json.Unmarshal(ev.Input, &input); err != nil || input["file_path"] != "spec.md" {
		t.Errorf("input crudo no viaja: %s (err %v)", ev.Input, err)
	}

	// Otros subtipos de control quedan honestamente fuera de alcance.
	if evs := translate([]byte(`{"type":"control_request","request_id":"r","request":{"subtype":"hook_callback"}}`), nil); len(evs) != 0 {
		t.Error("un control subtype ajeno a can_use_tool no debe reenviarse")
	}
}

func TestControlResponseLineWireFormat(t *testing.T) {
	// allow: echo del input original como updatedInput.
	line, err := controlResponseLine("req-7", ports.ControlDecision{
		Allow:        true,
		UpdatedInput: []byte(`{"file_path":"spec.md"}`),
	})
	if err != nil {
		t.Fatalf("allow: %v", err)
	}
	var env struct {
		Type     string `json:"type"`
		Response struct {
			Subtype   string `json:"subtype"`
			RequestID string `json:"request_id"`
			Response  struct {
				Behavior     string         `json:"behavior"`
				UpdatedInput map[string]any `json:"updatedInput"`
				Message      string         `json:"message"`
			} `json:"response"`
		} `json:"response"`
	}
	if uerr := json.Unmarshal(line, &env); uerr != nil {
		t.Fatalf("unmarshal: %v", uerr)
	}
	if env.Type != "control_response" || env.Response.Subtype != "success" || env.Response.RequestID != "req-7" {
		t.Errorf("envelope = %+v, want control_response/success/req-7", env)
	}
	if env.Response.Response.Behavior != "allow" || env.Response.Response.UpdatedInput["file_path"] != "spec.md" {
		t.Errorf("inner = %+v, want behavior allow + echo del input", env.Response.Response)
	}
	if !strings.HasSuffix(string(line), "\n") {
		t.Error("la línea NDJSON debe terminar en \\n")
	}

	// deny: lleva message, no updatedInput.
	line, err = controlResponseLine("req-8", ports.ControlDecision{Allow: false, Message: "el rol reviewer deniega Write"})
	if err != nil {
		t.Fatalf("deny: %v", err)
	}
	if uerr := json.Unmarshal(line, &env); uerr != nil {
		t.Fatalf("unmarshal deny: %v", uerr)
	}
	if env.Response.Response.Behavior != "deny" || env.Response.Response.Message == "" {
		t.Errorf("deny inner = %+v, want behavior deny + message", env.Response.Response)
	}
}

func TestPermissionArgsMaterialization(t *testing.T) {
	// Zero value = canal SIEMPRE cableado (DD-1, deuda D): deny-by-default + prompt-tool,
	// nada pre-aprobado — una sesión sin rol sellado sigue teniendo HITL, no headless mudo.
	vacio := strings.Join(permissionArgs(domain.PermissionSet{}), " ")
	if vacio != "--permission-mode default --permission-prompt-tool stdio" {
		t.Errorf("set vacío debe cablear solo el canal (mode+prompt-tool), got %q", vacio)
	}

	ps := domain.PermissionSet{
		Rol:   "backend-dev",
		Allow: []string{"Read", "Grep", "Glob", "Edit", "Write"},
		Ask:   []string{"Bash"},
		Deny:  []string{"WebSearch"},
		TTL:   15 * time.Minute,
	}
	args := permissionArgs(ps)
	joined := " " + strings.Join(args, " ") + " "
	for _, must := range []string{
		" --permission-mode default ",
		" --disallowedTools WebSearch ",
		" --permission-prompt-tool stdio ",
	} {
		if !strings.Contains(joined, must) {
			t.Errorf("flags %q no contienen %q", joined, must)
		}
	}
	// El valor exacto de --allowedTools: SOLO el allow read-only genuino. Write/Edit
	// quedan fuera aunque el rol los permita (pasan por el diff-approval del GUI) y
	// Bash (ask) jamás se pre-aprueba.
	if got := flagValue(args, "--allowedTools"); got != "Read,Grep,Glob" {
		t.Errorf("--allowedTools = %q, want Read,Grep,Glob (sin Write/Edit/Bash)", got)
	}
}

// TestSpawnArgsMCPAislado enforces HS-17 D2 (docs/architecture/boundaries/superficie-local-confinada.md):
// every spawn with a populated Injection locks MCP to ONLY the aggregated config file —
// never the operator's `~/.claude` servers or claude.ai account connectors.
func TestSpawnArgsMCPAislado(t *testing.T) {
	args := SpawnArgs(ports.SpawnOpts{Injection: ports.Injection{MCPConfigFile: "/tmp/mcp.json"}})
	joined := " " + strings.Join(args, " ") + " "
	if !strings.Contains(joined, " --mcp-config /tmp/mcp.json --strict-mcp-config ") {
		t.Errorf("argv = %q, want --mcp-config <archivo> --strict-mcp-config contiguos", joined)
	}

	// Injection vacía (degradación honesta, provisioning falló): cero flags de MCP —
	// nunca emitir --mcp-config con ruta vacía.
	args = SpawnArgs(ports.SpawnOpts{})
	if flagValue(args, "--mcp-config") != "" || strings.Contains(strings.Join(args, " "), "--strict-mcp-config") {
		t.Errorf("argv = %v, want cero flags de MCP sin Injection.MCPConfigFile", args)
	}
}

// TestSpawnArgsSettingSourcesExcludeUser enforces HS-17 D3 (docs/architecture/boundaries/
// superficie-local-confinada.md): EVERY spawn fixes --setting-sources to
// "project,local" — "user" (the operator's own ~/.claude/settings.json, with THEIR
// enabledPlugins/hooks) never rides, and there is no SpawnOpts field to reintroduce it.
func TestSpawnArgsSettingSourcesExcludeUser(t *testing.T) {
	for _, opts := range []ports.SpawnOpts{
		{},
		{Injection: ports.Injection{MCPConfigFile: "/tmp/mcp.json"}},
		{Permisos: domain.PermissionSet{Rol: "backend-dev", Allow: []string{"Read"}}},
	} {
		if got := flagValue(SpawnArgs(opts), "--setting-sources"); got != "project,local" {
			t.Errorf("--setting-sources = %q, want project,local (opts=%+v)", got, opts)
		}
	}
}

// flagValue returns the argument following flag, or "".
func flagValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// TestCtxPctUsaUltimoUsage (RF-194): el ctx del result sale del usage del ÚLTIMO API call
// (frames assistant), no del acumulado del result — que re-cuenta cache_read por tool-call
// y dio 100 % espurio en vivo (E2E T4 del paquete mejorar-arnes-conversando).
func TestCtxPctUsaUltimoUsage(t *testing.T) {
	var last *usage
	// Dos API calls: el turno ocupa 60k al final (30 % de 200k), no la suma 460k. El
	// contenido vacío no emite eventos, pero el usage SÍ se captura.
	translate([]byte(`{"type":"assistant","message":{"content":[],"usage":{"input_tokens":500,"cache_read_input_tokens":30000,"cache_creation_input_tokens":9500}}}`), &last)
	translate([]byte(`{"type":"assistant","message":{"content":[],"usage":{"input_tokens":1000,"cache_read_input_tokens":50000,"cache_creation_input_tokens":9000}}}`), &last)
	evs := translate([]byte(`{"type":"result","result":"listo","usage":{"input_tokens":10000,"cache_read_input_tokens":420000,"cache_creation_input_tokens":30000}}`), &last)
	if len(evs) != 1 || evs[0].Kind != ports.EventResult {
		t.Fatalf("result no tradujo: %+v", evs)
	}
	if evs[0].CtxPct != 30 {
		t.Errorf("CtxPct = %d, quiero 30 (60k del último call / 200k default)", evs[0].CtxPct)
	}

	// Sin ningún assistant con usage → fallback honesto al usage del result.
	var vacio *usage
	evs = translate([]byte(`{"type":"result","result":"x","usage":{"input_tokens":20000,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`), &vacio)
	if len(evs) != 1 || evs[0].CtxPct != 10 {
		t.Errorf("fallback CtxPct = %+v, quiero 10", evs)
	}
}
