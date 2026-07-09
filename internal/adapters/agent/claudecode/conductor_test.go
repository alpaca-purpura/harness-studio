package claudecode

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

func TestTranslateForwardsControlRequest(t *testing.T) {
	line := []byte(`{"type":"control_request","request_id":"req-7",` +
		`"request":{"subtype":"can_use_tool","tool_name":"Write",` +
		`"input":{"file_path":"spec.md","content":"x"}}}`)
	ev, ok := translate(line)
	if !ok {
		t.Fatal("un control_request:can_use_tool debe reenviarse, no descartarse (Fase E)")
	}
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
	if _, ok := translate([]byte(`{"type":"control_request","request_id":"r","request":{"subtype":"hook_callback"}}`)); ok {
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
	// Zero value = sin flags extra: el spawn interactivo del Dock no cambia.
	if got := permissionArgs(domain.PermissionSet{}); got != nil {
		t.Errorf("set vacío debe emitir cero flags, got %v", got)
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

// TestSpawnArgsMCPAislado enforces HS-17 D2 (arch/boundaries/superficie-local-confinada.md):
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

// TestSpawnArgsSettingSourcesExcludeUser enforces HS-17 D3 (arch/boundaries/
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
