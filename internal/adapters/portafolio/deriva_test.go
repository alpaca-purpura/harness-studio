package portafolio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
)

// armaArbol crea un mini-árbol golden bajo un TempDir nuevo: 2 archivos + un .git/ y un
// .in_use que deben quedar excluidos del hash.
func armaArbol(t *testing.T, extra map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	escribir(t, filepath.Join(dir, ".claude-plugin", "plugin.json"), `{"name":"x","version":"1.0.0"}`)
	escribir(t, filepath.Join(dir, "skills", "a", "SKILL.md"), "# a\n")
	escribir(t, filepath.Join(dir, ".git", "HEAD"), "ref: refs/heads/main\n")
	escribir(t, filepath.Join(dir, ".in_use"), "")
	for rel, contenido := range extra {
		escribir(t, filepath.Join(dir, rel), contenido)
	}
	return dir
}

func TestHashFormaPluginDeterminista(t *testing.T) {
	a := armaArbol(t, nil)
	b := armaArbol(t, nil) // árbol idéntico, dir físico distinto.

	ha, err := portafolio.HashFormaPlugin(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := portafolio.HashFormaPlugin(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Errorf("dos árboles idénticos deben hashear igual: %s ≠ %s", ha, hb)
	}
	if ha == "" {
		t.Error("el hash no debe ser vacío")
	}

	// Solo el .in_use difiere (excluido) → mismo hash.
	c := armaArbol(t, nil)
	if werr := os.WriteFile(filepath.Join(c, ".in_use"), []byte("otro contenido"), 0o600); werr != nil {
		t.Fatal(werr)
	}
	hc, err := portafolio.HashFormaPlugin(c)
	if err != nil {
		t.Fatal(err)
	}
	if hc != ha {
		t.Error(".in_use debe quedar excluido del hash (S0-D7)")
	}

	// Un byte distinto en un archivo REAL sí cambia el hash.
	d := armaArbol(t, map[string]string{"skills/a/SKILL.md": "# a modificado\n"})
	hd, err := portafolio.HashFormaPlugin(d)
	if err != nil {
		t.Fatal(err)
	}
	if hd == ha {
		t.Error("un byte distinto en un archivo real debe cambiar el hash")
	}
}

func TestEvaluarDerivaAlHilo(t *testing.T) {
	ccDir := t.TempDir()
	mktDir := t.TempDir()
	escribir(t, filepath.Join(ccDir, "known_marketplaces.json"), `{
		"kit-mkt": {"source":{"source":"github","repo":"owner/kit-mkt"},"installLocation":"`+mktDir+`"}
	}`)

	// La referencia y la instalación son EL MISMO contenido (árboles gemelos).
	contenido := map[string]string{"skills/a/SKILL.md": "# a\n"}
	refDir := filepath.Join(mktDir, "plugins", "harness-x", "1.0.0")
	escribir(t, filepath.Join(refDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"1.0.0"}`)
	escribir(t, filepath.Join(refDir, "skills", "a", "SKILL.md"), contenido["skills/a/SKILL.md"])

	installDir := t.TempDir()
	escribir(t, filepath.Join(installDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"1.0.0"}`)
	escribir(t, filepath.Join(installDir, "skills", "a", "SKILL.md"), contenido["skills/a/SKILL.md"])

	refs := &portafolio.Referencias{CCPluginsDir: ccDir}
	estado, motivo := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/kit-mkt", "harness-x", "1.0.0")
	if estado != "al-hilo" {
		t.Errorf("estado = %q (motivo %q), quiero al-hilo", estado, motivo)
	}
}

func TestEvaluarDerivaEnDeriva(t *testing.T) {
	ccDir := t.TempDir()
	mktDir := t.TempDir()
	escribir(t, filepath.Join(ccDir, "known_marketplaces.json"), `{
		"kit-mkt": {"source":{"source":"github","repo":"owner/kit-mkt"},"installLocation":"`+mktDir+`"}
	}`)

	refDir := filepath.Join(mktDir, "plugins", "harness-x", "1.0.0")
	escribir(t, filepath.Join(refDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"1.0.0"}`)
	escribir(t, filepath.Join(refDir, "skills", "a", "SKILL.md"), "# a original\n")

	installDir := t.TempDir()
	escribir(t, filepath.Join(installDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"1.0.0"}`)
	escribir(t, filepath.Join(installDir, "skills", "a", "SKILL.md"), "# a editado localmente\n")

	refs := &portafolio.Referencias{CCPluginsDir: ccDir}
	estado, motivo := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/kit-mkt", "harness-x", "1.0.0")
	if estado != "en-deriva" {
		t.Errorf("estado = %q, quiero en-deriva", estado)
	}
	if motivo == "" {
		t.Error("quiero un motivo visible")
	}
}

// TestDerivaNuncaSemver: MISMA version declarada en plugin.json, árboles con contenido
// DISTINTO ⇒ en-deriva. El veredicto sale del hash, jamás del string de versión (BR-4).
func TestDerivaNuncaSemver(t *testing.T) {
	ccDir := t.TempDir()
	mktDir := t.TempDir()
	escribir(t, filepath.Join(ccDir, "known_marketplaces.json"), `{
		"kit-mkt": {"source":{"source":"github","repo":"owner/kit-mkt"},"installLocation":"`+mktDir+`"}
	}`)
	refDir := filepath.Join(mktDir, "plugins", "harness-x", "9.9.9")
	escribir(t, filepath.Join(refDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"9.9.9"}`)
	escribir(t, filepath.Join(refDir, "skills", "a", "SKILL.md"), "# referencia\n")

	installDir := t.TempDir()
	// mismo string de version "9.9.9" que la referencia, contenido DIFERENTE.
	escribir(t, filepath.Join(installDir, ".claude-plugin", "plugin.json"), `{"name":"harness-x","version":"9.9.9"}`)
	escribir(t, filepath.Join(installDir, "skills", "a", "SKILL.md"), "# esto NO es lo mismo\n")

	refs := &portafolio.Referencias{CCPluginsDir: ccDir}
	estado, _ := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/kit-mkt", "harness-x", "9.9.9")
	if estado != "en-deriva" {
		t.Errorf("misma version, contenido distinto ⇒ en-deriva (nunca al-hilo por coincidencia de string); got %q", estado)
	}
}

func TestEvaluarDerivaNoEvaluable(t *testing.T) {
	ccDir := t.TempDir()
	installDir := t.TempDir()
	escribir(t, filepath.Join(installDir, ".claude-plugin", "plugin.json"), `{"name":"x"}`)

	t.Run("sin version", func(t *testing.T) {
		refs := &portafolio.Referencias{CCPluginsDir: ccDir}
		estado, motivo := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/mkt", "x", "")
		if estado != "deriva-no-evaluable" || motivo == "" {
			t.Errorf("estado=%q motivo=%q, quiero deriva-no-evaluable con motivo", estado, motivo)
		}
	})

	t.Run("sin marketplace local conocido", func(t *testing.T) {
		refs := &portafolio.Referencias{CCPluginsDir: ccDir} // sin known_marketplaces.json.
		estado, motivo := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/mkt", "x", "1.0.0")
		if estado != "deriva-no-evaluable" || motivo == "" {
			t.Errorf("estado=%q motivo=%q, quiero deriva-no-evaluable con motivo", estado, motivo)
		}
	})

	t.Run("version inexistente en el checkout local", func(t *testing.T) {
		mktDir := t.TempDir()
		escribir(t, filepath.Join(ccDir, "known_marketplaces.json"), `{
			"kit-mkt": {"source":{"source":"github","repo":"owner/mkt"},"installLocation":"`+mktDir+`"}
		}`)
		// el marketplace SÍ está clonado, pero la versión pedida no existe ahí.
		refs := &portafolio.Referencias{CCPluginsDir: ccDir}
		estado, motivo := portafolio.EvaluarDeriva(installDir, refs, "github.com/owner/mkt", "x", "999.0.0")
		if estado != "deriva-no-evaluable" || motivo == "" {
			t.Errorf("estado=%q motivo=%q, quiero deriva-no-evaluable con motivo", estado, motivo)
		}
	})
}
