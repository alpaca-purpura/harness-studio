package mechanism

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPkgOf: el enforced_by decide el paquete donde corre el arch-test — alias
// histórico "arch_test.go:TestX" = paquete fitness (fallback); ruta repo-relativa
// = el paquete de ese archivo; archivo inexistente = colgante (ok=false), jamás
// un pass sobre el paquete equivocado.
func TestPkgOf(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "adapters", "portafolio"), 0o750); err != nil {
		t.Fatal(err)
	}
	colocado := filepath.Join(root, "internal", "adapters", "portafolio", "store_test.go")
	if err := os.WriteFile(colocado, []byte("package portafolio\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	fallback := "./docs/architecture/fitness/"
	cases := []struct {
		name       string
		enforcedBy string
		wantPkg    string
		wantOK     bool
	}{
		{"alias fitness", "arch_test.go:TestX", fallback, true},
		{"sin dos-puntos", "arch_test.go", fallback, true},
		{
			"colocado existente", "internal/adapters/portafolio/store_test.go:TestStoreDegradaHonesto",
			"./internal/adapters/portafolio/", true,
		},
		{"colocado colgante", "internal/adapters/portafolio/no_existe_test.go:TestX", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pkg, ok := pkgOf(root, c.enforcedBy, fallback)
			if pkg != c.wantPkg || ok != c.wantOK {
				t.Fatalf("pkgOf(%q) = (%q, %v), quiero (%q, %v)", c.enforcedBy, pkg, ok, c.wantPkg, c.wantOK)
			}
		})
	}
}

func TestTestNameOf(t *testing.T) {
	cases := []struct {
		enforcedBy string
		want       string
	}{
		{"arch_test.go:TestX", "TestX"},
		{"internal/adapters/portafolio/deriva_test.go:TestDerivaNuncaSemver", "TestDerivaNuncaSemver"},
		{"arch_test.go", ""},
		{"arch_test.go:noEsTest", ""},
	}
	for _, c := range cases {
		if got := testNameOf(c.enforcedBy); got != c.want {
			t.Errorf("testNameOf(%q) = %q, quiero %q", c.enforcedBy, got, c.want)
		}
	}
}
