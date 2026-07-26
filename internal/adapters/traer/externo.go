package traer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// externo.go es el camino B (§13.7): `git init` + `sparse-checkout --cone` +
// `fetch --depth 1 --filter=blob:none origin <sha>` + `checkout FETCH_HEAD`. Probado en vivo
// contra `42Crunch-AI/claude-plugins`: HEAD == el sha declarado, solo el subdir pedido, 664 KB.
//
// `git`, NO `gh`, para traer — y `gh` SOLO como credential helper: `git rev-parse HEAD` da los 40
// hex exactos que BR-16 necesita (el tarball de `gh` trae 7), `--filter=blob:none` baja solo los
// blobs de la subruta, y el pin por `sha` SIN `ref` (143 de 220 entradas reales) es soportado.
//
// **El token NUNCA en `argv`** (riesgo 17): `-c credential.helper='!gh auth git-credential'`
// cuando `gh` está autenticado, y `GIT_ASKPASS` a un script en el temporal (modo 0700) que lee de
// su propio env cuando hay PAT. Un `-c http.extraHeader=…` quedaría en `ps` y una url con token
// en el reflog.

// Alias locales de los centinelas de EJECUCIÓN del dominio: adapter y transporte comparten LA
// MISMA instancia porque ninguno puede importar al otro (C23 de design.md).
var (
	ErrTraerRemotoNoTiene = domain.ErrTraerRemotoNoTiene
	ErrTraerSinAuth       = domain.ErrTraerSinAuth
	ErrTraerLocalFS       = domain.ErrTraerLocal
)

// ClonadorExterno satisface ports.Materializador para el camino B.
type ClonadorExterno struct {
	GitBin   string        // default "git"; inyectable (shim en tests).
	GHBin    string        // default "gh": credential helper. "" ⇒ no se usa.
	Token    string        // PAT fallback (ARNESIA_GH_TOKEN); "" ⇒ no se usa.
	Timeout  time.Duration // default 60 s (más generoso que leer un catálogo: acá se baja un árbol).
	MaxBytes int64         // default 64 MiB, chequeado DESPUÉS del fetch midiendo el work tree.
}

var _ ports.Materializador = (*ClonadorExterno)(nil)

// Camino identifica la vía.
func (c *ClonadorExterno) Camino() domain.CaminoTraer { return domain.CaminoExterno }

func (c *ClonadorExterno) gitBin() string {
	if c.GitBin != "" {
		return c.GitBin
	}
	return "git"
}

func (c *ClonadorExterno) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 60 * time.Second
}

func (c *ClonadorExterno) maxBytes() int64 {
	if c.MaxBytes > 0 {
		return c.MaxBytes
	}
	return 64 << 20
}

// Materializar baja el árbol pineado y deja SOLO la subruta en `staging`, con el mismo set de
// exclusión y las mismas reglas de symlink/permisos del camino A (§13.6 pasos 2-4) — eso descarta
// `.git` y también los archivos de la RAÍZ que el cone mode arrastra (§13.1 hecho 7).
func (c *ClonadorExterno) Materializar(ctx context.Context, plan domain.PlanTraer, staging string) (string, []string, error) {
	if plan.Camino != domain.CaminoExterno {
		return "", nil, fmt.Errorf("traer externo: el plan pide el camino %q", plan.Camino)
	}
	if plan.URL == "" {
		return "", nil, errors.New("traer externo: el plan no declara url")
	}
	pin := plan.SHAEsperado
	if pin == "" {
		pin = plan.Ref // sin `sha`, el `ref` es lo único que hay (PlanificarTraer ya exigió uno).
	}
	if pin == "" {
		return "", nil, errors.New("traer externo: el plan no declara sha ni ref")
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout())
	defer cancel()

	// El work tree vive AL LADO del staging (dentro del temporal del usecase): mismo filesystem,
	// y lo limpia el `defer` del usecase junto con todo lo demás (BR-15).
	work := filepath.Join(filepath.Dir(staging), "work")
	if err := os.MkdirAll(work, 0o750); err != nil {
		return "", nil, fmt.Errorf("%w: crear work tree: %w", ErrTraerLocalFS, err)
	}

	askpass, limpiarAskpass, aerr := c.prepararAuth(filepath.Dir(staging))
	if aerr != nil {
		return "", nil, aerr
	}
	defer limpiarAskpass()

	var avisos []string
	pasos := [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", plan.URL},
		{"sparse-checkout", "init", "--cone"},
	}
	if plan.Subruta != "" {
		pasos = append(pasos, []string{"sparse-checkout", "set", plan.Subruta})
	}
	for _, args := range pasos {
		if _, err := c.git(ctx, work, askpass, args...); err != nil {
			return "", avisos, err
		}
	}

	// Fetch pineado por el `sha` (la AUTORIDAD, §13.1 hecho 2). GitHub permite fetch por sha
	// alcanzable — verificado en vivo.
	fetch := []string{"fetch", "--depth", "1", "--filter=blob:none", "origin", pin}
	if _, err := c.git(ctx, work, askpass, fetch...); err != nil {
		// Si el remoto NO permite fetch por sha (`allowReachableSHA1InWant` apagado), se reintenta
		// UNA sola vez por `ref` — y BR-16 SIGUE aplicando sobre el sha (si HEAD no lo iguala, 502).
		if plan.Ref == "" || !rechazaFetchPorSHA(err) {
			return "", avisos, err
		}
		avisos = append(avisos, fmt.Sprintf("el remoto rechazó el fetch por sha; se reintentó por ref %q (el sha se sigue verificando)", plan.Ref))
		if _, err2 := c.git(ctx, work, askpass, "fetch", "--depth", "1", "origin", plan.Ref); err2 != nil {
			return "", avisos, err2
		}
	}
	if _, err := c.git(ctx, work, askpass, "checkout", "-q", "FETCH_HEAD"); err != nil {
		return "", avisos, err
	}

	// Techo de tamaño DESPUÉS del fetch (§13.7): el `defer` del usecase limpia el temporal.
	tam, terr := TamanoDeArbol(work)
	if terr != nil {
		return "", avisos, fmt.Errorf("%w: medir el árbol traído: %w", ErrTraerLocalFS, terr)
	}
	if tam > c.maxBytes() {
		return "", avisos, fmt.Errorf("%w: el árbol traído (%d bytes) excede el techo de %d MiB", ErrTraerRemotoNoTiene, tam, c.maxBytes()>>20)
	}

	sha, err := c.git(ctx, work, askpass, "rev-parse", "HEAD")
	if err != nil {
		return "", avisos, err
	}
	sha = strings.TrimSpace(sha)

	// El contenido canónico es `<work>/<Subruta>`, JAMÁS `<work>`: el cone mode arrastra los
	// archivos de la raíz del repo ajeno (§13.1 hecho 7) — sin este staging el canónico se
	// llevaría el README.md ajeno y el .git.
	origen := work
	if plan.Subruta != "" {
		origen = filepath.Join(work, filepath.FromSlash(plan.Subruta))
		if fi, serr := os.Stat(origen); serr != nil || !fi.IsDir() {
			return "", avisos, fmt.Errorf("%w: el commit %s no contiene la subruta %q que el catálogo declara", ErrTraerRemotoNoTiene, sha, plan.Subruta)
		}
	}
	avisosCopia, cerr := CopiarArbol(origen, staging)
	avisos = append(avisos, avisosCopia...)
	if cerr != nil {
		return "", avisos, fmt.Errorf("%w: %w", ErrTraerLocalFS, cerr)
	}
	return sha, avisos, nil
}

// prepararAuth arma la credencial SIN poner el token en `argv` (BR-18 · riesgo 17). Devuelve la
// ruta del script `GIT_ASKPASS` ("" si no hace falta) y su limpiador.
func (c *ClonadorExterno) prepararAuth(dirTemporal string) (askpass string, limpiar func(), err error) {
	limpiar = func() {}
	if c.Token == "" {
		return "", limpiar, nil
	}
	ruta := filepath.Join(dirTemporal, "askpass.sh")
	// El token viaja por el ENV DEL HIJO, jamás en el argv (visible en `ps`) ni en la url
	// (queda en el reflog). El script se escribe en el temporal, modo 0700.
	script := "#!/bin/sh\nprintf '%s' \"$ARNESIA_GH_TOKEN\"\n"
	if werr := os.WriteFile(ruta, []byte(script), 0o700); werr != nil { //nolint:gosec // G306: tiene que ser ejecutable; vive en el temporal del usecase.
		return "", limpiar, fmt.Errorf("%w: escribir GIT_ASKPASS: %w", ErrTraerLocalFS, werr)
	}
	return ruta, func() { _ = os.Remove(ruta) }, nil
}

// git corre un subcomando de git en `work` y devuelve su stdout. El stderr REAL se clasifica a un
// centinela (§13.7): los motivos quedan DISTINGUIBLES, ninguno genérico.
func (c *ClonadorExterno) git(ctx context.Context, work, askpass string, args ...string) (string, error) {
	base := []string{"-C", work}
	// `gh` como credential helper: ArnesIA NUNCA ve el token (BR-18 literal).
	if askpass == "" && c.GHBin != "" {
		base = append(base, "-c", "credential.helper=!"+c.GHBin+" auth git-credential")
	}
	base = append(base, args...)

	cmd := exec.CommandContext(ctx, c.gitBin(), base...) //nolint:gosec // G204: bin inyectable, args armados del plan del dominio.
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	cmd.WaitDelay = 2 * time.Second
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0", // jamás pedir credenciales por tty (BR-18: la app no pide nada).
		"GIT_CONFIG_NOSYSTEM=1", // la config global del operador no cambia el comportamiento.
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	if askpass != "" {
		cmd.Env = append(cmd.Env, "GIT_ASKPASS="+askpass, "ARNESIA_GH_TOKEN="+c.Token)
	}

	err := cmd.Run()
	if ctx.Err() != nil {
		return "", fmt.Errorf("%w: git %s: timeout de %s", ErrTraerRemotoNoTiene, args[0], c.timeout())
	}
	if err != nil {
		return "", clasificarErrorGit(args, errBuf.String(), err)
	}
	return out.String(), nil
}

// rechazaFetchPorSHA reporta si el error corresponde a un remoto que no admite fetch por sha
// (el único caso que habilita el reintento único por `ref`).
func rechazaFetchPorSHA(err error) bool {
	low := strings.ToLower(err.Error())
	return strings.Contains(low, "not our ref") ||
		strings.Contains(low, "unadvertised object") ||
		strings.Contains(low, "server does not allow request for unadvertised object")
}

// clasificarErrorGit mapea el stderr REAL de git a un centinela (tabla de §13.7). E-84/E-85
// exigen motivos DISTINTOS, ninguno genérico.
func clasificarErrorGit(args []string, stderr string, err error) error {
	s := strings.TrimSpace(stderr)
	if s == "" {
		s = err.Error()
	}
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "could not read username"),
		strings.Contains(low, "authentication failed"),
		strings.Contains(low, "terminal prompts disabled"),
		strings.Contains(low, "403"):
		return fmt.Errorf("%w: git %s: %s", ErrTraerSinAuth, args[0], s)
	case strings.Contains(low, "repository not found"),
		strings.Contains(low, "404"),
		strings.Contains(low, "couldn't find remote ref"),
		strings.Contains(low, "not our ref"),
		strings.Contains(low, "unadvertised object"),
		strings.Contains(low, "did not match any file"):
		return fmt.Errorf("%w: git %s: %s", ErrTraerRemotoNoTiene, args[0], s)
	case strings.Contains(low, "no space left"), strings.Contains(low, "permission denied"):
		return fmt.Errorf("%w: git %s: %s", ErrTraerLocalFS, args[0], s)
	default:
		return fmt.Errorf("%w: git %s: %s", ErrTraerRemotoNoTiene, args[0], s)
	}
}
