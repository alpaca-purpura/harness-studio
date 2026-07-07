package artifact_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/artifact"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// TestFuenteReaderConfinamiento covers the S2 confinement contract of LeerConfinado:
// relative refs join under dir; absolute refs must already fall inside dir; anything
// that escapes (.., absolute-outside) is ErrFueraDelArnes; empty dir/ref refuse.
func TestFuenteReaderConfinamiento(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "adentro.md"), []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	fuera := filepath.Join(t.TempDir(), "fuera.md")
	if err := os.WriteFile(fuera, []byte("secreto"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := artifact.NewFuenteReader()

	if b, err := r.LeerConfinado(dir, "adentro.md"); err != nil || string(b) != "ok" {
		t.Errorf("relativa dentro = %q err:%v, want ok", b, err)
	}
	if b, err := r.LeerConfinado(dir, filepath.Join(dir, "adentro.md")); err != nil || string(b) != "ok" {
		t.Errorf("absoluta dentro = %q err:%v, want ok", b, err)
	}

	for name, ref := range map[string]string{
		"traversal relativo": "../fuera.md",
		"absoluta fuera":     fuera,
		"traversal anidado":  "skills/../../fuera.md",
	} {
		if _, err := r.LeerConfinado(dir, ref); !errors.Is(err, ports.ErrFueraDelArnes) {
			t.Errorf("%s: err = %v, want ErrFueraDelArnes", name, err)
		}
	}

	if _, err := r.LeerConfinado("", "x.md"); err == nil {
		t.Error("dir vacío: err = nil, want error (lectura sin confinamiento prohibida)")
	}
	if _, err := r.LeerConfinado(dir, ""); err == nil {
		t.Error("ref vacía: err = nil, want error")
	}
	if _, err := r.LeerConfinado(dir, "no-existe.md"); err == nil || errors.Is(err, ports.ErrFueraDelArnes) {
		t.Errorf("archivo ausente: err = %v, want error de lectura (no confinamiento)", err)
	}
}
