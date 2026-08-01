package marketplace

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// git corre un git en dir y falla el test si el comando falla (setup de fixtures).
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "git", append([]string{"-C", dir}, args...)...) //nolint:gosec // G204: fixture de test, args literales del propio test.
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v en %s: %v\n%s", args, dir, err, out)
	}
}

// armarOrigenYClone crea un repo origen con un commit y su clone (el «checkout de CC»).
func armarOrigenYClone(t *testing.T) (origen, clone string) {
	t.Helper()
	base := t.TempDir()
	origen = filepath.Join(base, "origen")
	clone = filepath.Join(base, "clone")
	if err := os.MkdirAll(origen, 0o750); err != nil {
		t.Fatal(err)
	}
	git(t, origen, "init", "-b", "main")
	if err := os.WriteFile(filepath.Join(origen, "v1.txt"), []byte("v1"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, origen, "add", ".")
	git(t, origen, "commit", "-m", "v1")
	git(t, base, "clone", "--quiet", origen, clone)
	return origen, clone
}

// TestSincronizadorAvanzaElClone — DD-2/E-bis: un commit nuevo en el origen aparece en el
// clone tras Sincronizar (esto es lo que «↻ Refrescar» promete: ver la versión nueva sin
// `git pull` manual).
func TestSincronizadorAvanzaElClone(t *testing.T) {
	origen, clone := armarOrigenYClone(t)
	if err := os.WriteFile(filepath.Join(origen, "v2.txt"), []byte("v2"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, origen, "add", ".")
	git(t, origen, "commit", "-m", "v2")

	s := &SincronizadorGit{}
	detalle, err := s.Sincronizar(context.Background(), domain.MarketplaceConocido{Nombre: "mio", InstallLocation: clone})
	if err != nil {
		t.Fatalf("Sincronizar: %v", err)
	}
	if !strings.HasPrefix(detalle, "avanzó a ") {
		t.Errorf("detalle = %q, want «avanzó a <sha>»", detalle)
	}
	if _, serr := os.Stat(filepath.Join(clone, "v2.txt")); serr != nil {
		t.Errorf("v2.txt no llegó al clone tras Sincronizar: %v", serr)
	}

	// Segunda pasada sin cambios: «ya al día», jamás error.
	detalle, err = s.Sincronizar(context.Background(), domain.MarketplaceConocido{Nombre: "mio", InstallLocation: clone})
	if err != nil || detalle != "ya al día" {
		t.Errorf("segunda pasada = (%q, %v), want («ya al día», nil)", detalle, err)
	}
}

// TestSincronizadorNoMergeaHistoriaDivergente — `--ff-only` a propósito: un clone editado a
// mano (commit local) NO se mergea en silencio; el pull falla con motivo visible.
func TestSincronizadorNoMergeaHistoriaDivergente(t *testing.T) {
	origen, clone := armarOrigenYClone(t)
	// Historia divergente: un commit distinto en cada lado.
	if err := os.WriteFile(filepath.Join(origen, "suyo.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, origen, "add", ".")
	git(t, origen, "commit", "-m", "suyo")
	if err := os.WriteFile(filepath.Join(clone, "mano.txt"), []byte("y"), 0o600); err != nil {
		t.Fatal(err)
	}
	git(t, clone, "add", ".")
	git(t, clone, "commit", "-m", "a mano")

	s := &SincronizadorGit{}
	if _, err := s.Sincronizar(context.Background(), domain.MarketplaceConocido{Nombre: "mio", InstallLocation: clone}); err == nil {
		t.Fatal("historia divergente debe FALLAR el pull ff-only, no mergear en silencio")
	}
}

// TestSincronizadorSinCheckout — sin InstallLocation o sin repo git: error con motivo, sin pánico.
func TestSincronizadorSinCheckout(t *testing.T) {
	s := &SincronizadorGit{}
	if _, err := s.Sincronizar(context.Background(), domain.MarketplaceConocido{Nombre: "x"}); err == nil {
		t.Error("sin InstallLocation debe dar error")
	}
	noGit := t.TempDir()
	if _, err := s.Sincronizar(context.Background(), domain.MarketplaceConocido{Nombre: "x", InstallLocation: noGit}); err == nil {
		t.Error("un dir sin .git debe dar error")
	}
}
