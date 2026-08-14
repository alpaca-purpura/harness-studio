package fitness

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// leerChangelog lee CHANGELOG.md sin abortar si falta: la ausencia del archivo es EL hallazgo
// que este test tiene que reportar con nombre y apellido, no un `read: no such file`.
func leerChangelog(t *testing.T) string {
	t.Helper()
	root := repoRoot()
	if root == "" {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(root, "CHANGELOG.md")) //nolint:gosec // G304: ruta literal del repo, jamás input externo.
	if err != nil {
		return ""
	}
	return string(b)
}

// changelog_test.go — guarda docs/architecture/conventions/versionado.md §changelog-y-bump
// (paquete 2026-07-26-versionado-y-changelog-metodologicos, VC-D5).
//
// El gate de `scripts/bump.sh` solo muerde a quien corre `make bump-*`. Estos tests son la
// segunda capa: corren en CI y agarran el caso que el gate no ve — un manifiesto bumpeado a mano
// (sed, editor, merge) que deja la versión publicada sin decir qué trajo. Hasta 2026-07-26 el
// repo tenía 20 releases en instaladores/ y cero registro de qué había en cada uno; la convención
// es exigible desde el marcador `convencion-desde`, y lo anterior NO se valida ni se inventa
// (VC-D4, honestidad §4: hueco declarado antes que relleno plausible).

var (
	reSeccionCL   = regexp.MustCompile(`(?m)^## \[([^\]]+)\](?:\s*[—-]\s*(\S+))?\s*$`)
	reCategoriaCL = regexp.MustCompile(`(?m)^### (.+?)\s*$`)
	reConvencion  = regexp.MustCompile(`<!--\s*convencion-desde:\s*(\d+\.\d+\.\d+)\s*-->`)
	reSemverCL    = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
	reCargoVerCL  = regexp.MustCompile(`(?m)^version = "([^"]+)"`)
)

// categoriasCL son las 6 de Keep a Changelog traducidas. El set es cerrado a propósito: una
// categoría libre ("Notas", "Varios") es por donde se cuela el changelog que no dice nada.
var categoriasCL = map[string]bool{
	"Agregado": true, "Cambiado": true, "Deprecado": true,
	"Eliminado": true, "Corregido": true, "Seguridad": true,
}

// rellenoCL son las entradas que simulan contenido. Se rechazan explícito.
var rellenoCL = map[string]bool{
	"tbd": true, "todo": true, "pendiente": true, "n/a": true, "wip": true, "por definir": true,
}

func semverMenor(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, _ := strconv.Atoi(pa[i])
		y, _ := strconv.Atoi(pb[i])
		if x != y {
			return x < y
		}
	}
	return false
}

// seccionesCL parte el changelog en (versión → cuerpo), en orden de aparición.
func seccionesCL(t *testing.T, texto string) []struct {
	Version, Fecha, Cuerpo string
} {
	t.Helper()
	idx := reSeccionCL.FindAllStringSubmatchIndex(texto, -1)
	out := make([]struct{ Version, Fecha, Cuerpo string }, 0, len(idx))
	for i, m := range idx {
		fin := len(texto)
		if i+1 < len(idx) {
			fin = idx[i+1][0]
		}
		var fecha string
		if m[4] >= 0 {
			fecha = texto[m[4]:m[5]]
		}
		out = append(out, struct{ Version, Fecha, Cuerpo string }{
			Version: texto[m[2]:m[3]], Fecha: fecha, Cuerpo: texto[m[1]:fin],
		})
	}
	return out
}

// TestChangelogExisteYTieneForma: sin CHANGELOG.md no hay convención que valer. Verifica el
// marcador `convencion-desde`, la sección [Sin publicar] única, semver plano con fecha, orden
// descendente y categorías canónicas.
func TestChangelogExisteYTieneForma(t *testing.T) {
	texto := leerChangelog(t)
	if texto == "" {
		t.Fatal("falta CHANGELOG.md en la raíz — obligatorio desde 2026-07-26 (versionado.md §changelog-y-bump)")
	}

	conv := reConvencion.FindStringSubmatch(texto)
	if conv == nil {
		t.Fatal("CHANGELOG.md sin marcador `<!-- convencion-desde: X.Y.Z -->`: no se sabe desde qué versión es exigible")
	}

	secs := seccionesCL(t, texto)
	if len(secs) == 0 {
		t.Fatal("CHANGELOG.md no tiene ninguna sección `## [...]`")
	}

	sinPublicar := 0
	var previa string
	for _, s := range secs {
		if s.Version == "Sin publicar" {
			sinPublicar++
			continue
		}
		if !reSemverCL.MatchString(s.Version) {
			// Se tolera SOLO el bloque de historia no reconstruida, rotulado como tal.
			if !strings.Contains(s.Version, "anteriores") {
				t.Errorf("sección `[%s]`: no es semver plano X.Y.Z", s.Version)
			}
			continue
		}
		if s.Fecha == "" {
			t.Errorf("[%s] sin fecha — el formato es `## [%s] — AAAA-MM-DD`", s.Version, s.Version)
		} else if !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(s.Fecha) {
			t.Errorf("[%s]: fecha %q no es AAAA-MM-DD", s.Version, s.Fecha)
		}
		if previa != "" && !semverMenor(s.Version, previa) {
			t.Errorf("[%s] rompe el orden descendente (viene después de [%s]; la más nueva va arriba)", s.Version, previa)
		}
		previa = s.Version

		if !semverMenor(s.Version, conv[1]) && !strings.Contains(s.Cuerpo, "\n- ") {
			t.Errorf("[%s] no tiene ni una entrada — una versión sin cambios no se publica", s.Version)
		}
	}
	if sinPublicar != 1 {
		t.Errorf("hay %d secciones `[Sin publicar]`; debe haber exactamente una (es donde se acumula la próxima versión)", sinPublicar)
	}

	for _, s := range secs {
		for _, c := range reCategoriaCL.FindAllStringSubmatch(s.Cuerpo, -1) {
			if !categoriasCL[c[1]] {
				t.Errorf("[%s]: categoría %q no es una de las 6 canónicas (Agregado·Cambiado·Deprecado·Eliminado·Corregido·Seguridad)", s.Version, c[1])
			}
		}
		for _, l := range strings.Split(s.Cuerpo, "\n") {
			if e := strings.TrimSpace(strings.TrimPrefix(l, "- ")); strings.HasPrefix(l, "- ") {
				if rellenoCL[strings.ToLower(strings.TrimSuffix(e, "."))] {
					t.Errorf("[%s]: entrada de relleno %q — un changelog que no dice nada miente", s.Version, e)
				}
			}
		}
	}
}

// TestChangelogCubreLaVersionDeLosManifiestos es el diente que importa: la versión que hoy
// declaran los manifiestos —la que va a salir en un instalador— tiene que tener su sección.
// Bumpear a mano sin pasar por `make bump-*` deja este test rojo.
func TestChangelogCubreLaVersionDeLosManifiestos(t *testing.T) {
	cargo := readSourceFile(t, filepath.Join("web", "src-tauri", "Cargo.toml"))
	texto := leerChangelog(t)
	if cargo == "" || texto == "" {
		return
	}
	m := reCargoVerCL.FindStringSubmatch(cargo)
	if m == nil {
		t.Fatal("no [package].version en web/src-tauri/Cargo.toml")
	}
	version := m[1]

	conv := reConvencion.FindStringSubmatch(texto)
	if conv == nil {
		return // ya lo reporta TestChangelogExisteYTieneForma
	}
	if semverMenor(version, conv[1]) {
		t.Skipf("versión %s es anterior a la convención (%s): historia no reconstruida a propósito (VC-D4)", version, conv[1])
	}

	for _, s := range seccionesCL(t, texto) {
		if s.Version == version {
			if !strings.Contains(s.Cuerpo, "\n- ") {
				t.Fatalf("[%s] existe pero está vacía — no se publica una versión muda", version)
			}
			return
		}
	}
	t.Fatalf("los manifiestos dicen %s y CHANGELOG.md no tiene su sección — bumpeá con `make bump-patch|bump-minor|bump-major`, "+
		"que promueve [Sin publicar] (VC-D2); nunca a mano", version)
}

// pythonDelHost resuelve el intérprete: python3 (unix) o python (Windows, donde el
// python3 de WindowsApps es un stub falso que abre la Store) — sondeando por EJECUCIÓN,
// no por presencia en PATH (mismo criterio que el shim PYTHON del Makefile).
func pythonDelHost() string {
	for _, p := range []string{"python3", "python"} {
		if exec.Command(p, "-c", "pass").Run() == nil {
			return p
		}
	}
	return "python3"
}

// TestChangelogScriptSeValidaASiMismo corre el gate real. Si `scripts/changelog.py` se rompe,
// el bump dejaría de proteger nada y nadie se enteraría hasta el próximo release.
func TestChangelogScriptSeValidaASiMismo(t *testing.T) {
	root := repoRoot()
	cmd := exec.Command(pythonDelHost(), filepath.Join(root, "scripts", "changelog.py"), "check")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scripts/changelog.py check falló:\n%s", out)
	}
}
