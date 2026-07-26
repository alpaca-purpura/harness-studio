package traer

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
)

// copiarDir duplica un árbol de testdata en un dir temporal para poder mutarlo (agregar symlinks,
// cambiar permisos) sin tocar el fixture del repo. Los symlinks y el bit setuid se crean ACÁ y no
// en `testdata/` a propósito: un symlink a `/etc/hostname` commiteado en el repo es ruido, y el
// checkout real de prenter no tiene ninguno — el caso es sintético en cualquier escenario.
func copiarDir(t *testing.T, origen string) string {
	t.Helper()
	destino := t.TempDir()
	err := filepath.WalkDir(origen, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(origen, p)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		b, rerr := os.ReadFile(p) //nolint:gosec // G304: fixture del propio repo.
		if rerr != nil {
			return rerr
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		return os.WriteFile(target, b, info.Mode().Perm()) //nolint:gosec // G703/G306: copia de un fixture del repo a t.TempDir(), preservando permisos a propósito.
	})
	if err != nil {
		t.Fatalf("copiar %s: %v", origen, err)
	}
	return destino
}

func listar(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if rel == "." {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("listar %s: %v", dir, err)
	}
	sort.Strings(out)
	return out
}

func contiene(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// E-94 · el set de exclusión es EXACTAMENTE el de `HashFormaPlugin`, y la aserción es CRUZADA:
// `HashFormaPlugin(origen) == HashFormaPlugin(staging)`. Es la garantía estructural de que un
// traído sano dé `al-hilo` (BR-17). El test importa el adapter hermano a propósito: go-arch-lint
// excluye `_test.go` — en producción sigue prohibido (C15).
func TestExclusionCoincideConHashFormaPlugin(t *testing.T) {
	origen := copiarDir(t, "testdata/mkt-prenter/plugins/harness/0.5.3")
	// El fixture ya trae `.orphaned_at` y `.gitignore`; se agrega `.git/` y `.in_use` acá.
	if err := os.MkdirAll(filepath.Join(origen, ".git", "refs"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, ".git", "config"), []byte("[remote \"origin\"]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, ".in_use"), nil, 0o600); err != nil {
		t.Fatal(err)
	}

	staging := filepath.Join(t.TempDir(), "staging")
	avisos, err := CopiarArbol(origen, staging)
	if err != nil {
		t.Fatalf("CopiarArbol: %v", err)
	}
	if len(avisos) != 0 {
		t.Fatalf("avisos = %v, want ninguno (árbol sano)", avisos)
	}

	copiados := listar(t, staging)
	for _, prohibido := range []string{".git", ".git/config", ".in_use", ".orphaned_at"} {
		if contiene(copiados, prohibido) {
			t.Fatalf("%q se copió y NO debería (set de exclusión de HashFormaPlugin)", prohibido)
		}
	}
	if !contiene(copiados, ".gitignore") {
		t.Fatalf(".gitignore NO se copió y debería (no está en el set): %v", copiados)
	}

	hOrigen, err := portafolio.HashFormaPlugin(origen)
	if err != nil {
		t.Fatalf("HashFormaPlugin(origen): %v", err)
	}
	hStaging, err := portafolio.HashFormaPlugin(staging)
	if err != nil {
		t.Fatalf("HashFormaPlugin(staging): %v", err)
	}
	if hOrigen != hStaging {
		t.Fatalf("HashFormaPlugin difiere ⇒ BR-17 reportaría en-deriva sobre algo recién traído:\norigen:  %s\nstaging: %s\ncopiados: %v", hOrigen, hStaging, copiados)
	}
}

// E-92 · symlink interno se copia COMO symlink; el que escapa y el roto NO se copian y dejan
// aviso VISIBLE. El contenido de afuera JAMÁS aparece en el staging.
func TestCopiaSymlinks(t *testing.T) {
	origen := t.TempDir()
	if err := os.MkdirAll(filepath.Join(origen, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, "sub", "x"), []byte("adentro\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("./sub/x", filepath.Join(origen, "dentro")); err != nil {
		t.Skipf("symlinks no disponibles: %v", err)
	}
	if err := os.Symlink("/etc/hostname", filepath.Join(origen, "fuera")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("./nada", filepath.Join(origen, "roto")); err != nil {
		t.Fatal(err)
	}

	staging := filepath.Join(t.TempDir(), "staging")
	avisos, err := CopiarArbol(origen, staging)
	if err != nil {
		t.Fatalf("CopiarArbol: %v", err)
	}

	fi, lerr := os.Lstat(filepath.Join(staging, "dentro"))
	if lerr != nil {
		t.Fatalf("el symlink interno no se copió: %v", lerr)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("el symlink interno se copió como archivo, no como symlink")
	}
	b, rerr := os.ReadFile(filepath.Join(staging, "dentro")) //nolint:gosec // G304: staging del propio test.
	if rerr != nil || string(b) != "adentro\n" {
		t.Fatalf("el symlink interno no resuelve dentro del staging: %v / %q", rerr, b)
	}

	if _, serr := os.Lstat(filepath.Join(staging, "fuera")); serr == nil {
		t.Fatal("el symlink que ESCAPA se copió (metería contenido de otro árbol en el canónico)")
	}
	if _, serr := os.Lstat(filepath.Join(staging, "roto")); serr == nil {
		t.Fatal("el symlink ROTO se copió")
	}

	// /etc/hostname no puede aparecer en NINGÚN archivo del staging.
	for _, rel := range listar(t, staging) {
		p := filepath.Join(staging, rel)
		if fi, serr := os.Lstat(p); serr == nil && fi.Mode().IsRegular() {
			contenido, _ := os.ReadFile(p) //nolint:gosec // G304: dentro del staging del test.
			hostname, _ := os.ReadFile("/etc/hostname")
			if len(hostname) > 0 && strings.Contains(string(contenido), strings.TrimSpace(string(hostname))) {
				t.Fatalf("%s trae contenido de /etc/hostname", rel)
			}
		}
	}

	var vioFuera, vioRoto bool
	for _, a := range avisos {
		if strings.Contains(a, "fuera") && strings.Contains(a, "escapa") {
			vioFuera = true
		}
		if strings.Contains(a, "roto") && strings.Contains(a, "está roto") {
			vioRoto = true
		}
	}
	if !vioFuera || !vioRoto {
		t.Fatalf("avisos = %v, want uno por el symlink que escapa y uno por el roto", avisos)
	}
}

// E-93 · permisos: dirs 0o750, archivos 0o640, el bit de ejecución del origen se CONSERVA y
// setuid/setgid/sticky se DESCARTAN.
func TestCopiaPreservaEjecutable(t *testing.T) {
	origen := t.TempDir()
	if err := os.MkdirAll(filepath.Join(origen, "hooks"), 0o777); err != nil { //nolint:gosec // G301: el test verifica que la copia NORMALICE a 0750.
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, "hooks", "pre.sh"), []byte("#!/bin/sh\n"), 0o755); err != nil { //nolint:gosec // G306: el ejecutable es EL insumo de E-93.
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, "README.md"), []byte("hola\n"), 0o644); err != nil { //nolint:gosec // G306: insumo de E-93 (0644 → 0640).
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(origen, "raro"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(origen, "raro"), 0o4755); err != nil { //nolint:gosec // G302: el setuid es EL insumo de E-93 (se verifica que se DESCARTE).
		t.Fatal(err)
	}

	staging := filepath.Join(t.TempDir(), "staging")
	if _, err := CopiarArbol(origen, staging); err != nil {
		t.Fatalf("CopiarArbol: %v", err)
	}

	fiDir, err := os.Stat(filepath.Join(staging, "hooks"))
	if err != nil {
		t.Fatal(err)
	}
	if fiDir.Mode().Perm() != 0o750 {
		t.Fatalf("dir hooks perm = %o, want 750", fiDir.Mode().Perm())
	}

	fiSh, err := os.Stat(filepath.Join(staging, "hooks", "pre.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if fiSh.Mode().Perm()&0o111 == 0 {
		t.Fatalf("pre.sh perdió el bit de ejecución: %o", fiSh.Mode().Perm())
	}
	if fiSh.Mode().Perm() != 0o640|0o111 {
		t.Fatalf("pre.sh perm = %o, want %o (0640 + bits de ejecución del origen)", fiSh.Mode().Perm(), 0o640|0o111)
	}

	fiMd, err := os.Stat(filepath.Join(staging, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if fiMd.Mode().Perm() != 0o640 {
		t.Fatalf("README.md perm = %o, want 640", fiMd.Mode().Perm())
	}

	fiRaro, err := os.Stat(filepath.Join(staging, "raro"))
	if err != nil {
		t.Fatal(err)
	}
	if fiRaro.Mode()&os.ModeSetuid != 0 || fiRaro.Mode()&os.ModeSetgid != 0 || fiRaro.Mode()&os.ModeSticky != 0 {
		t.Fatalf("setuid/setgid/sticky sobrevivió la copia: %v", fiRaro.Mode())
	}
}

// E-91 (lado copia) · `source: "./"` (caveman/ponytail lo usan): se copia el árbol SIN su `.git`,
// así el canónico no hereda el remote del marketplace.
func TestCopiaRaizDeMarketplaceExcluyeGit(t *testing.T) {
	staging := filepath.Join(t.TempDir(), "staging")
	avisos, err := CopiarArbol("testdata/mkt-caveman", staging)
	if err != nil {
		t.Fatalf("CopiarArbol: %v", err)
	}
	if len(avisos) != 0 {
		t.Fatalf("avisos = %v, want ninguno", avisos)
	}
	copiados := listar(t, staging)
	for _, rel := range copiados {
		if strings.HasPrefix(rel, ".git") {
			t.Fatalf("%q se copió: el canónico heredaría el remote del MARKETPLACE", rel)
		}
	}
	for _, esperado := range []string{".claude-plugin/marketplace.json", "README.md", "skills/SKILL.md"} {
		if !contiene(copiados, esperado) {
			t.Fatalf("%q no se copió: %v", esperado, copiados)
		}
	}
}

// TamanoDeArbol suma solo archivos regulares (guarda de techo del camino B).
func TestTamanoDeArbol(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "b"), []byte("123"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := TamanoDeArbol(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != 8 {
		t.Fatalf("TamanoDeArbol = %d, want 8", got)
	}
}
