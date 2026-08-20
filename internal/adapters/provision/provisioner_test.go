package provision_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/provision"
)

// TestProvisionMaterializesAndIsIdempotent verifica el puente 2 (HS-11): la doctrina y
// el kit embebidos se materializan bajo el dir de la app, la Injection apunta a rutas
// reales FUERA de cualquier arnés (② ↛ ③), y una segunda provisión no re-escribe (la
// huella de contenido gobierna el refresh).
func TestProvisionMaterializesAndIsIdempotent(t *testing.T) {
	base := t.TempDir()
	p, err := provision.New(base, doctrina.Kit, doctrina.Files)
	if err != nil {
		t.Fatal(err)
	}

	inj, err := p.Provision(context.Background())
	if err != nil {
		t.Fatalf("provision: %v", err)
	}

	// La Injection apunta a los tres cuerpos materializados.
	wantFiles := []string{
		filepath.Join(base, "kit", ".claude-plugin", "plugin.json"),
		filepath.Join(base, "kit", "skills", "forjar-arnes", "SKILL.md"),
		filepath.Join(base, "kit", "skills", "forjar-caja", "SKILL.md"),
		filepath.Join(base, "kit", "skills", "auditar-arnes", "SKILL.md"),
		inj.SystemPromptFile, // doctrine.md suelto (--append-system-prompt-file).
		filepath.Join(base, "knowhow", "skills.md"),
		filepath.Join(base, "knowhow", "harness-profile.md"),
		inj.MCPConfigFile, // mcp.json (HS-17 D2): --mcp-config apunta acá, nunca a ~/.claude.
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("falta materializado: %s (%v)", f, err)
		}
	}
	if len(inj.PluginDirs) != 1 || inj.PluginDirs[0] != filepath.Join(base, "kit") {
		t.Errorf("PluginDirs = %v", inj.PluginDirs)
	}
	if len(inj.AddDirs) != 1 || inj.AddDirs[0] != filepath.Join(base, "knowhow") {
		t.Errorf("AddDirs = %v", inj.AddDirs)
	}
	if inj.MCPConfigFile != filepath.Join(base, "mcp.json") {
		t.Errorf("MCPConfigFile = %q", inj.MCPConfigFile)
	}
	mcpContent, rerr := os.ReadFile(inj.MCPConfigFile)
	if rerr != nil {
		t.Fatalf("mcp.json: %v", rerr)
	}
	if got := string(mcpContent); got != `{"mcpServers":{}}` {
		t.Errorf("mcp.json = %q, want {\"mcpServers\":{}} (el kit hoy no declara uno propio)", got)
	}

	// Idempotencia: mismo binario → misma huella → el stamp no cambia y un archivo
	// tocado a mano NO se re-escribe en el mismo proceso (cache) ni en otro (huella).
	stamp := filepath.Join(base, ".doctrina-version")
	before, rerr := os.ReadFile(stamp) //nolint:gosec // G304: ruta dentro del TempDir del propio test.
	if rerr != nil {
		t.Fatal(rerr)
	}
	if _, perr := p.Provision(context.Background()); perr != nil {
		t.Fatal(perr)
	}
	after, rerr2 := os.ReadFile(stamp) //nolint:gosec // G304: ídem.
	if rerr2 != nil {
		t.Fatal(rerr2)
	}
	if string(before) != string(after) {
		t.Error("el stamp cambió entre provisiones idénticas")
	}
}

// TestProvisionSessionEscribeTarjeta (RF-189, mejorar-arnes-conversando T10): con extra,
// nace ~/.arnesia/sessions/<id>/system.md = doctrina ② + tarjeta, y la Injection apunta
// ahí; sin extra, degrada al archivo compartido (comportamiento previo). El archivo se
// RE-escribe por spawn (la tarjeta refleja el estado actual).
func TestProvisionSessionEscribeTarjeta(t *testing.T) {
	base := t.TempDir()
	p, err := provision.New(base, doctrina.Kit, doctrina.Files)
	if err != nil {
		t.Fatal(err)
	}

	inj, err := p.ProvisionSession(context.Background(), "s123", "## Tarjeta\n\n- arnés: vitalia")
	if err != nil {
		t.Fatalf("provision session: %v", err)
	}
	want := filepath.Join(base, "sessions", "s123", "system.md")
	if inj.SystemPromptFile != want {
		t.Fatalf("SystemPromptFile = %q, quiero %q", inj.SystemPromptFile, want)
	}
	b, err := os.ReadFile(want) //nolint:gosec // G304: ruta de fixture del test (t.TempDir()), no input externo.
	if err != nil {
		t.Fatal(err)
	}
	contenido := string(b)
	if !strings.Contains(contenido, "arnés: vitalia") {
		t.Error("la tarjeta no quedó en el system.md por sesión")
	}
	base2, err := os.ReadFile(filepath.Join(base, "doctrine.md")) //nolint:gosec // G304: ruta de fixture del test.
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(contenido, strings.TrimSpace(string(base2))[:40]) {
		t.Error("el system.md por sesión debe CONTENER la doctrina compartida")
	}

	// Re-provisión con tarjeta nueva pisa el archivo (estado actual, no el del turno 1).
	if _, provisionErr := p.ProvisionSession(context.Background(), "s123", "tarjeta-v2"); provisionErr != nil {
		t.Fatal(provisionErr)
	}
	b, _ = os.ReadFile(want) //nolint:gosec // G304: ruta de fixture del test.
	if !strings.Contains(string(b), "tarjeta-v2") {
		t.Error("re-provisión no actualizó la tarjeta")
	}

	// Sin extra → el archivo compartido, sin tocar el de la sesión.
	inj2, err := p.ProvisionSession(context.Background(), "s123", "")
	if err != nil {
		t.Fatal(err)
	}
	if inj2.SystemPromptFile != filepath.Join(base, "doctrine.md") {
		t.Errorf("sin tarjeta debe degradar al doctrine.md compartido, got %q", inj2.SystemPromptFile)
	}
}
