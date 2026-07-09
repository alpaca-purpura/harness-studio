// HS-17: fitness tests for the config-source axis of superficie-local-confinada (v1.3) —
// the MCP + settings-sources isolation SpawnArgs enforces on every spawn. Kept in its own
// file (not arch_test.go) so it lands independent of any other in-flight edit to that file;
// `arnesia conformance` discovers it the same way — Go test discovery is package-scoped,
// not filename-scoped, and the ruleset's `arch_test.go:TestX` enforcer strings are a
// convention the parser matches textually, not a literal path.
package fitness

import (
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/agent/claudecode"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// TestConfigSourceMCPAislado enforces checklist row `mcp-config-siempre`: every spawn with
// a populated Injection locks MCP to ONLY the aggregated config file.
func TestConfigSourceMCPAislado(t *testing.T) {
	args := claudecode.SpawnArgs(ports.SpawnOpts{Injection: ports.Injection{MCPConfigFile: "/tmp/mcp.json"}})
	joined := " " + strings.Join(args, " ") + " "
	if !strings.Contains(joined, " --mcp-config /tmp/mcp.json --strict-mcp-config ") {
		t.Errorf("argv = %q, want --mcp-config <archivo> --strict-mcp-config", joined)
	}
}

// TestConfigSourceSettingSourcesExcludeUser enforces checklist row
// `setting-sources-siempre`: every spawn fixes --setting-sources to "project,local" —
// "user" (the operator's own ~/.claude/settings.json) never rides.
func TestConfigSourceSettingSourcesExcludeUser(t *testing.T) {
	args := claudecode.SpawnArgs(ports.SpawnOpts{})
	joined := " " + strings.Join(args, " ") + " "
	if !strings.Contains(joined, " --setting-sources project,local ") {
		t.Errorf("argv = %q, want --setting-sources project,local", joined)
	}
}
