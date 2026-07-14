package portafolio_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

func escribir(t *testing.T, ruta, contenido string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(ruta), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
}

func hallazgoConTipo(hs []domain.HallazgoInstalacion, tipo domain.TipoInstalacion) (domain.HallazgoInstalacion, bool) {
	for _, h := range hs {
		if h.Tipo == tipo {
			return h, true
		}
	}
	return domain.HallazgoInstalacion{}, false
}

func TestScannerProyectoInstalado(t *testing.T) {
	root := t.TempDir()
	escribir(t, filepath.Join(root, ".claude", "settings.json"), `{}`)

	s := &portafolio.Scanner{}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := hallazgoConTipo(hs, domain.InstProyectoInstalado)
	if !ok {
		t.Fatalf("quiero un hallazgo proyecto-instalado, got %+v", hs)
	}
	if h.Dir != root {
		t.Errorf("Dir = %q, quiero %q", h.Dir, root)
	}
}

func TestScannerMaterializada(t *testing.T) {
	root := t.TempDir()
	escribir(t, filepath.Join(root, ".claude", "plugins", "harness-x", ".claude-plugin", "plugin.json"), `{"name":"harness-x"}`)

	s := &portafolio.Scanner{}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := hallazgoConTipo(hs, domain.InstMaterializada)
	if !ok {
		t.Fatalf("quiero un hallazgo materializada, got %+v", hs)
	}
	if h.Aviso != "" {
		t.Errorf("plugin.json presente no debe llevar aviso, got %q", h.Aviso)
	}
}

func TestScannerPluginRotoVisible(t *testing.T) {
	root := t.TempDir()
	// dir de plugin SIN plugin.json (C-P-5): no-reconocible, visible, no crashea.
	escribir(t, filepath.Join(root, ".claude", "plugins", "roto", "algo.txt"), "no es un plugin")

	s := &portafolio.Scanner{}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := hallazgoConTipo(hs, domain.InstMaterializada)
	if !ok {
		t.Fatalf("quiero un hallazgo materializada (aunque roto), got %+v", hs)
	}
	if h.Aviso == "" {
		t.Error("quiero un aviso no-reconocible (C-P-5) visible")
	}
}

func TestScannerLockDevstudio(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	// entrada resoluble: el dir existe en el caché local.
	cacheDir := filepath.Join(home, ".dev-studio", "arneses", "harness-a", "1.0.0")
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		t.Fatal(err)
	}
	escribir(t, filepath.Join(root, ".devstudio", "arneses.yaml"), `
arneses:
  - id: harness-a
    version: "1.0.0"
    canal: beta
    registry: github.com/owner/repo
  - id: harness-ausente
    version: "2.0.0"
    canal: beta
    registry: github.com/owner/repo2
`)

	s := &portafolio.Scanner{}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	var resuelta, ausente bool
	for _, h := range hs {
		for _, e := range h.Eslabones {
			if e.Fuente == "lock-devstudio" && e.Campo == "version" && e.Valor == "1.0.0" {
				resuelta = true
				if h.Dir == "" {
					t.Error("la entrada resoluble debe traer Dir")
				}
			}
			if e.Fuente == "lock-devstudio" && e.Campo == "version" && e.Valor == "2.0.0" {
				ausente = true
				if h.Aviso == "" {
					t.Error("la entrada sin dir en caché debe traer Aviso visible (C-P-14)")
				}
			}
		}
	}
	if !resuelta {
		t.Error("falta el hallazgo de la entrada resoluble")
	}
	if !ausente {
		t.Error("falta el hallazgo de la entrada ausente")
	}
}

func TestScannerReferenciadaCC(t *testing.T) {
	root := t.TempDir()
	ccDir := t.TempDir()
	escribir(t, filepath.Join(root, ".claude", "settings.json"),
		`{"enabledPlugins":{"harness@kit-mkt":true,"otro@kit-mkt":false}}`)
	installDir := filepath.Join(ccDir, "cache", "kit-mkt", "harness", "1.2.3")
	escribir(t, filepath.Join(ccDir, "installed_plugins.json"), `{
		"version": 2,
		"harness@kit-mkt": [{"scope":"project","projectPath":"`+root+`","installPath":"`+installDir+`","version":"1.2.3"}]
	}`)
	escribir(t, filepath.Join(ccDir, "known_marketplaces.json"), `{
		"kit-mkt": {"source":{"source":"github","repo":"owner/kit-mkt"},"installLocation":"`+ccDir+`/marketplaces/kit-mkt"}
	}`)

	s := &portafolio.Scanner{CCPluginsDir: ccDir}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := hallazgoConTipo(hs, domain.InstReferenciadaCC)
	if !ok {
		t.Fatalf("quiero un hallazgo referenciada-cc, got %+v", hs)
	}
	if h.Dir != installDir {
		t.Errorf("Dir = %q, quiero %q", h.Dir, installDir)
	}
	var version, registry string
	for _, e := range h.Eslabones {
		if e.Campo == "version" {
			version = e.Valor
		}
		if e.Campo == "registry" {
			registry = e.Valor
		}
	}
	if version != "1.2.3" {
		t.Errorf("version = %q, quiero 1.2.3", version)
	}
	if registry != "owner/kit-mkt" {
		t.Errorf("registry = %q, quiero owner/kit-mkt", registry)
	}
	// enabledPlugins:false NO se lista (S0-D2).
	if _, ok := hallazgoConTipo(hs, domain.InstReferenciadaCC); ok && len(hs) > 0 {
		for _, h := range hs {
			for _, e := range h.Eslabones {
				if e.Valor == "otro" {
					t.Error("un plugin con enabledPlugins:false no debe listarse")
				}
			}
		}
	}
}

func TestScannerCCMetadataIlegible(t *testing.T) {
	root := t.TempDir()
	ccDir := t.TempDir()
	escribir(t, filepath.Join(root, ".claude", "settings.json"), `{"enabledPlugins":{"harness@kit-mkt":true}}`)
	// installed_plugins.json con version distinta de 2 → no-legible (S0-D1).
	escribir(t, filepath.Join(ccDir, "installed_plugins.json"), `{"version": 99, "harness@kit-mkt": []}`)

	s := &portafolio.Scanner{CCPluginsDir: ccDir}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	var vistoNoLegible bool
	for _, h := range hs {
		for _, e := range h.Eslabones {
			if e.Fuente == "no-legible" {
				vistoNoLegible = true
			}
		}
	}
	if !vistoNoLegible {
		t.Error("quiero un eslabón no-legible visible, jamás un crash")
	}
}

func TestScannerMonorepoAcotado(t *testing.T) {
	root := t.TempDir()
	// .claude anidado dentro de un paquete del monorepo.
	escribir(t, filepath.Join(root, "packages", "a", ".claude", "settings.json"), `{}`)
	// symlink circular: no debe colgar el scan ni seguirse fuera de root.
	circular := filepath.Join(root, "packages", "circular")
	if err := os.MkdirAll(filepath.Dir(circular), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(root, circular); err != nil {
		t.Skipf("symlinks no soportados en este entorno: %v", err)
	}

	s := &portafolio.Scanner{}
	hs, err := s.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, h := range hs {
		if h.Tipo == domain.InstProyectoInstalado && h.Dir == filepath.Join(root, "packages", "a") {
			found = true
		}
	}
	if !found {
		t.Errorf("quiero encontrar el .claude anidado en packages/a, got %+v", hs)
	}

	// cancelación de ctx: un ctx ya cancelado debe devolver error, no colgar.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Escanear(ctx, root); err == nil {
		t.Error("quiero un error honesto con ctx cancelado, no un scan silencioso")
	}
}

func TestScannerRootInvalido(t *testing.T) {
	s := &portafolio.Scanner{}

	if _, err := s.Escanear(context.Background(), "relativo/no/absoluto"); err == nil {
		t.Error("root relativo debe rechazarse")
	}
	if _, err := s.Escanear(context.Background(), filepath.Join(t.TempDir(), "no-existe")); err == nil {
		t.Error("root inexistente debe rechazarse")
	}
}
