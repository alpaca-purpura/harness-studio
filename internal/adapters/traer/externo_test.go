package traer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// El sha y la subruta REALES de la entrada `42crunch-api-security-testing` del catálogo oficial
// (verificados en vivo el 2026-07-25: `HEAD == 30287f5e…`, 664 KB, solo el subdir pedido).
const (
	shaReal     = "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad"
	refReal     = "v1.5.5"
	subrutaReal = "plugins/api-security-testing"
	urlReal     = "https://github.com/42Crunch-AI/claude-plugins.git"
)

func planExterno(t *testing.T, raiz string) domain.PlanTraer {
	t.Helper()
	fila := domain.EntradaCatalogo{
		Nombre: "42crunch-api-security-testing",
		Source: domain.SourceCatalogo{
			Tipo: domain.SourceGitSubdir, URL: urlReal, Ruta: subrutaReal, Ref: refReal, SHA: shaReal,
			Crudo: domain.CrudoDeSource(domain.SourceGitSubdir, urlReal, subrutaReal, refReal),
		},
	}
	mkt := domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: "github.com/alpacapurpura/prenter-marketplace", Clase: domain.ClasePropio,
	}
	plan, err := domain.PlanificarTraer(mkt, fila, false, raiz)
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	return plan
}

// entorno arma un temporal con `staging/` al lado y devuelve (staging, log, setEnv).
func entorno(t *testing.T) (staging, log string) {
	t.Helper()
	tmp := t.TempDir()
	return filepath.Join(tmp, "staging"), filepath.Join(tmp, "shim.log")
}

func conEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func lineasDelLog(t *testing.T, log string) []string {
	t.Helper()
	b, err := os.ReadFile(log) //nolint:gosec // G304: log del shim, en un temporal del test.
	if err != nil {
		t.Fatalf("leer log del shim: %v", err)
	}
	var out []string
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// E-81 · camino B feliz contra el shim: la secuencia de comandos es la de §13.7, el pin es el
// **`sha`, NO el `ref`** (C16), y el staging queda con el contenido de la subruta y SIN `.git`
// ni los archivos de la raíz que el cone mode arrastra (§13.1 hecho 7).
func TestClonadorExternoShim(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":     log,
		"ARNESIA_SHIM_SHA":     shaReal,
		"ARNESIA_SHIM_SUBRUTA": subrutaReal,
	})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second}
	plan := planExterno(t, t.TempDir())

	sha, avisos, err := c.Materializar(context.Background(), plan, staging)
	if err != nil {
		t.Fatalf("Materializar: %v", err)
	}
	if sha != shaReal {
		t.Fatalf("sha efectivo = %q, want %q", sha, shaReal)
	}
	if len(avisos) != 0 {
		t.Fatalf("avisos = %v, want ninguno en el camino feliz", avisos)
	}

	lineas := lineasDelLog(t, log)
	esperados := []string{
		"init -q",
		"remote add origin " + urlReal,
		"sparse-checkout init --cone",
		"sparse-checkout set " + subrutaReal,
		"fetch --depth 1 --filter=blob:none origin " + shaReal,
		"checkout -q FETCH_HEAD",
		"rev-parse HEAD",
	}
	if len(lineas) != len(esperados) {
		t.Fatalf("invocaciones = %d, want %d:\n%s", len(lineas), len(esperados), strings.Join(lineas, "\n"))
	}
	for i, esperado := range esperados {
		if !strings.HasSuffix(lineas[i], esperado) {
			t.Fatalf("invocación %d = %q, want que termine en %q", i, lineas[i], esperado)
		}
	}
	// El pin es el SHA, jamás el ref (C16: en una entrada sana apuntan a commits DISTINTOS).
	for _, l := range lineas {
		if strings.Contains(l, "fetch") && strings.Contains(l, refReal) {
			t.Fatalf("el fetch usó el ref como pin: %q", l)
		}
	}

	copiados := listar(t, staging)
	for _, esperado := range []string{"skills/SKILL.md", "references/nota.md", "run.sh"} {
		if !contiene(copiados, esperado) {
			t.Fatalf("%q no llegó al staging: %v", esperado, copiados)
		}
	}
	for _, prohibido := range []string{".git", ".git/HEAD", "README.md"} {
		if contiene(copiados, prohibido) {
			t.Fatalf("%q llegó al staging (cone mode arrastra la raíz del repo AJENO): %v", prohibido, copiados)
		}
	}
	// El bit de ejecución del script sobrevive (E-93 en el camino B).
	if fi, serr := os.Stat(filepath.Join(staging, "run.sh")); serr != nil || fi.Mode().Perm()&0o111 == 0 {
		t.Fatalf("run.sh perdió el bit de ejecución: %v", serr)
	}
}

// E-83 · el remoto no tiene lo que el catálogo declara ⇒ ErrTraerRemotoNoTiene con el stderr REAL
// en el motivo.
func TestTraerRemotoNoTiene(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":       log,
		"ARNESIA_SHIM_FETCH_ERR": "fatal: couldn't find remote ref 30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
	})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second}
	plan := planExterno(t, t.TempDir())
	plan.Ref = "" // sin ref no hay reintento: se aborta con el stderr real.

	_, _, err := c.Materializar(context.Background(), plan, staging)
	if !errors.Is(err, ErrTraerRemotoNoTiene) {
		t.Fatalf("err = %v, want ErrTraerRemotoNoTiene", err)
	}
	if !strings.Contains(err.Error(), "couldn't find remote ref") {
		t.Fatalf("el stderr real debe viajar en el motivo: %v", err)
	}
}

// E-84/E-85 · auth vs no-existe: códigos y motivos DISTINTOS, ninguno genérico. Y el token NUNCA
// aparece en el argv capturado por el shim ni en ningún archivo del temporal (BR-18).
func TestTraerClasificaAuthVsNoExiste(t *testing.T) {
	casos := []struct {
		nombre   string
		stderr   string
		wantErr  error
		enMotivo string
	}{
		{"sin credencial", "fatal: could not read Username for 'https://github.com': terminal prompts disabled", ErrTraerSinAuth, "could not read Username"},
		{"403", "remote: Write access to repository not granted.\nfatal: unable to access: The requested URL returned error: 403", ErrTraerSinAuth, "403"},
		{"repo inexistente", "remote: Repository not found.\nfatal: repository 'https://github.com/a/b.git/' not found", ErrTraerRemotoNoTiene, "Repository not found"},
		{"ref inexistente", "fatal: couldn't find remote ref v9.9.9", ErrTraerRemotoNoTiene, "couldn't find remote ref"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			staging, log := entorno(t)
			conEnv(t, map[string]string{"ARNESIA_SHIM_LOG": log, "ARNESIA_SHIM_FETCH_ERR": c.stderr})
			cl := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second}
			plan := planExterno(t, t.TempDir())
			plan.Ref = ""
			_, _, err := cl.Materializar(context.Background(), plan, staging)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("err = %v, want %v", err, c.wantErr)
			}
			if !strings.Contains(err.Error(), c.enMotivo) {
				t.Fatalf("motivo = %v, want que contenga %q", err, c.enMotivo)
			}
		})
	}
}

// E-84 (2ª mitad) · con PAT, el token viaja por GIT_ASKPASS (env del hijo) y NUNCA en el argv.
func TestTraerTokenNuncaEnArgv(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":     log,
		"ARNESIA_SHIM_SHA":     shaReal,
		"ARNESIA_SHIM_SUBRUTA": subrutaReal,
	})
	const token = "ghp-token-secretisimo"
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Token: token, Timeout: 30 * time.Second}
	plan := planExterno(t, t.TempDir())
	if _, _, err := c.Materializar(context.Background(), plan, staging); err != nil {
		t.Fatalf("Materializar: %v", err)
	}

	for _, l := range lineasDelLog(t, log) {
		if strings.Contains(l, token) {
			t.Fatalf("el token apareció en el argv (visible en `ps`): %q", l)
		}
	}
	// Y el script de GIT_ASKPASS se limpia al terminar (no queda un archivo con el token).
	tmp := filepath.Dir(staging)
	if _, err := os.Stat(filepath.Join(tmp, "askpass.sh")); err == nil {
		t.Fatal("el script de GIT_ASKPASS quedó en disco tras terminar")
	}
	// Ningún archivo del temporal contiene el token.
	for _, rel := range listar(t, tmp) {
		p := filepath.Join(tmp, rel)
		if fi, serr := os.Lstat(p); serr == nil && fi.Mode().IsRegular() {
			b, _ := os.ReadFile(p) //nolint:gosec // G304: dentro del temporal del test.
			if strings.Contains(string(b), token) {
				t.Fatalf("%s contiene el token", rel)
			}
		}
	}
}

// E-97 · el remoto rechaza el fetch por sha ⇒ UN solo reintento con `--depth 1 origin <ref>`, y
// BR-16 SIGUE aplicando (el sha se verifica igual). Sin `ref` ⇒ aborta con el stderr real.
func TestFetchPorShaFallaReintentaConRef(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":       log,
		"ARNESIA_SHIM_SHA":       shaReal,
		"ARNESIA_SHIM_SUBRUTA":   subrutaReal,
		"ARNESIA_SHIM_FETCH_ERR": "error: Server does not allow request for unadvertised object " + shaReal,
	})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second}
	plan := planExterno(t, t.TempDir())

	sha, avisos, err := c.Materializar(context.Background(), plan, staging)
	if err != nil {
		t.Fatalf("Materializar tras el reintento: %v", err)
	}
	if sha != shaReal {
		t.Fatalf("sha = %q, want %q (BR-16 sigue aplicando en el reintento)", sha, shaReal)
	}

	var fetches []string
	for _, l := range lineasDelLog(t, log) {
		if strings.Contains(l, " fetch ") || strings.HasPrefix(l, "fetch ") {
			fetches = append(fetches, l)
		}
	}
	if len(fetches) != 2 {
		t.Fatalf("fetches = %d, want 2 (UN solo reintento):\n%s", len(fetches), strings.Join(fetches, "\n"))
	}
	if !strings.HasSuffix(fetches[1], "fetch --depth 1 origin "+refReal) {
		t.Fatalf("el reintento = %q, want «fetch --depth 1 origin %s»", fetches[1], refReal)
	}
	var vio bool
	for _, a := range avisos {
		if strings.Contains(a, "rechazó el fetch por sha") && strings.Contains(a, refReal) {
			vio = true
		}
	}
	if !vio {
		t.Fatalf("avisos = %v, want el aviso del reintento (visible, BR-8)", avisos)
	}
}

// E-98 · techo de tamaño del clone: aborta con el motivo, temporal limpio (el `defer` es del
// usecase; acá se verifica que el staging queda vacío/inexistente).
func TestTechoDeBytesEnClone(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":     log,
		"ARNESIA_SHIM_SHA":     shaReal,
		"ARNESIA_SHIM_SUBRUTA": subrutaReal,
		"ARNESIA_SHIM_BYTES":   "3000000",
	})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second, MaxBytes: 1 << 20}
	plan := planExterno(t, t.TempDir())

	_, _, err := c.Materializar(context.Background(), plan, staging)
	if err == nil {
		t.Fatal("Materializar aceptó un árbol por encima del techo")
	}
	if !strings.Contains(err.Error(), "excede el techo de 1 MiB") {
		t.Fatalf("motivo = %v, want «excede el techo de 1 MiB»", err)
	}
	if _, serr := os.Stat(staging); serr == nil {
		t.Fatal("el staging quedó poblado tras abortar por techo")
	}
}

// E-99 · timeout del clone: ErrTraerRemotoNoTiene con motivo de timeout, sin colgar el proceso.
func TestTimeoutDeClone(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{"ARNESIA_SHIM_LOG": log, "ARNESIA_SHIM_SLEEP": "30"})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 200 * time.Millisecond}
	plan := planExterno(t, t.TempDir())

	inicio := time.Now()
	_, _, err := c.Materializar(context.Background(), plan, staging)
	if err == nil {
		t.Fatal("Materializar devolvió nil con el shim colgado")
	}
	if !errors.Is(err, ErrTraerRemotoNoTiene) {
		t.Fatalf("err = %v, want ErrTraerRemotoNoTiene", err)
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("el motivo debe decir timeout: %v", err)
	}
	if time.Since(inicio) > 20*time.Second {
		t.Fatalf("el timeout inyectado no se respetó (%s)", time.Since(inicio))
	}
}

// El commit materializado sin la subruta declarada es un dato que NO coincide con el catálogo:
// error honesto, no un canónico con el árbol equivocado.
func TestSubrutaAusenteEnElCommit(t *testing.T) {
	staging, log := entorno(t)
	conEnv(t, map[string]string{
		"ARNESIA_SHIM_LOG":     log,
		"ARNESIA_SHIM_SHA":     shaReal,
		"ARNESIA_SHIM_SUBRUTA": "otra/ruta",
	})
	c := &ClonadorExterno{GitBin: "testdata/bin/git", Timeout: 30 * time.Second}
	plan := planExterno(t, t.TempDir())

	_, _, err := c.Materializar(context.Background(), plan, staging)
	if !errors.Is(err, ErrTraerRemotoNoTiene) {
		t.Fatalf("err = %v, want ErrTraerRemotoNoTiene", err)
	}
	if !strings.Contains(err.Error(), subrutaReal) {
		t.Fatalf("el motivo debe citar la subruta que el catálogo declara: %v", err)
	}
}

// E-81b · camino B contra la RED REAL. Opt-in (`-tags red`): no corre en CI. La aserción es la
// verificada a mano el 2026-07-25 antes de escribir el diseño.
func TestClonadorExternoRedReal(t *testing.T) {
	if os.Getenv("ARNESIA_TEST_RED") == "" {
		t.Skip("E-81b es opt-in: exportá ARNESIA_TEST_RED=1 para correrlo contra la red real")
	}
	staging, _ := entorno(t)
	c := &ClonadorExterno{GHBin: "gh", Timeout: 120 * time.Second}
	plan := planExterno(t, t.TempDir())

	sha, _, err := c.Materializar(context.Background(), plan, staging)
	if err != nil {
		t.Fatalf("Materializar (red real): %v", err)
	}
	if sha != shaReal {
		t.Fatalf("HEAD = %q, want %q exacto", sha, shaReal)
	}
	copiados := listar(t, staging)
	if !contiene(copiados, "skills") && !contiene(copiados, "references") {
		t.Fatalf("el staging no trae skills/ ni references/: %v", copiados)
	}
	if contiene(copiados, "README.md") {
		t.Fatalf("el staging se llevó el README.md de la RAÍZ del repo ajeno: %v", copiados)
	}
	tam, terr := TamanoDeArbol(staging)
	if terr != nil {
		t.Fatal(terr)
	}
	if tam > 5<<20 {
		t.Fatalf("el staging pesa %d bytes, want < 5 MiB", tam)
	}
}
