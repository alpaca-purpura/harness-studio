package marketplace

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// E-01 · el plano nace poblado por Claude Code, no vacío. Insumo: COPIA LITERAL de
// `~/.claude/plugins/known_marketplaces.json` de esta máquina (5 entradas reales).
func TestDetectorCCCincoEntradasReales(t *testing.T) {
	d := &DetectorCC{CCPluginsDir: "testdata"}
	out, err := d.Detectados()
	if err != nil {
		t.Fatalf("Detectados: %v", err)
	}
	if len(out) != 5 {
		t.Fatalf("len(out) = %d, want 5", len(out))
	}

	nombres := make([]string, 0, len(out))
	for _, m := range out {
		nombres = append(nombres, m.Nombre)
	}
	sort.Strings(nombres)
	want := []string{"caveman", "claude-code-warp", "claude-plugins-official", "ponytail", "prenter-marketplace"}
	for i := range want {
		if nombres[i] != want[i] {
			t.Fatalf("nombres = %v, want %v", nombres, want)
		}
	}

	repos := map[string]string{
		"caveman":                 "github.com/juliusbrussee/caveman",
		"claude-code-warp":        "github.com/warpdotdev/claude-code-warp",
		"claude-plugins-official": "github.com/anthropics/claude-plugins-official",
		"ponytail":                "github.com/dietrichgebert/ponytail",
		"prenter-marketplace":     "github.com/alpacapurpura/prenter-marketplace",
	}
	for _, m := range out {
		if len(m.Eslabones) != 1 || m.Eslabones[0] != domain.EslabonCCKnown {
			t.Fatalf("%s: Eslabones = %v, want [cc-known-marketplaces]", m.Nombre, m.Eslabones)
		}
		if m.InstallLocation == "" {
			t.Fatalf("%s: InstallLocation vacío (CC lo declara siempre)", m.Nombre)
		}
		if m.Repo != repos[m.Nombre] {
			t.Fatalf("%s: Repo = %q, want %q (canonicalizado, RN-IDENT-1)", m.Nombre, m.Repo, repos[m.Nombre])
		}
		if m.CCActualizado == "" {
			t.Fatalf("%s: CCActualizado vacío (es el lastUpdated crudo de CC)", m.Nombre)
		}
		if m.Clase != domain.ClaseReferencia {
			t.Fatalf("%s: Clase = %q, want referencia (CC no modela clase: fail-safe)", m.Nombre, m.Clase)
		}
	}
}

// Metadata ausente ⇒ lista vacía SIN error: «CC no conoce ninguno» es una afirmación verdadera.
func TestDetectorSinMetadataNoEsError(t *testing.T) {
	d := &DetectorCC{CCPluginsDir: t.TempDir()}
	out, err := d.Detectados()
	if err != nil || out != nil {
		t.Fatalf("Detectados = (%v, %v), want (nil, nil)", out, err)
	}
}

// E-68 (lado adapter) · metadata ILEGIBLE ⇒ error, jamás una lista vacía silenciosa. El usecase
// lo convierte en `aviso_detector` visible y responde 200 con lo declarado.
func TestDetectorIlegibleDevuelveError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "known_marketplaces.json"), []byte("{roto"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := (&DetectorCC{CCPluginsDir: dir}).Detectados()
	if err == nil {
		t.Fatal("Detectados devolvió nil error con metadata ilegible (lista vacía muda)")
	}
	if out != nil {
		t.Fatalf("out = %v, want nil", out)
	}
}

// Un repo que NO canonicaliza no se descarta: queda visible como discrepancia (S1-D3).
func TestDetectorRepoNoCanonicalizableEsVisible(t *testing.T) {
	dir := t.TempDir()
	crudo := `{"raro":{"source":{"source":"github","repo":"solo-un-nombre"},"installLocation":"/tmp/raro"}}`
	if err := os.WriteFile(filepath.Join(dir, "known_marketplaces.json"), []byte(crudo), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := (&DetectorCC{CCPluginsDir: dir}).Detectados()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Repo != "" {
		t.Fatalf("out = %+v, want 1 fila con Repo vacío", out)
	}
	if len(out[0].Discrepancias) != 1 {
		t.Fatalf("el crudo no canonicalizable debe quedar VISIBLE: %v", out[0].Discrepancias)
	}
}
