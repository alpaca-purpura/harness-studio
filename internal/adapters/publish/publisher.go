// Package publish implementa ports.PublishPort shelleando `git`: el write-side del
// prenter-marketplace (B2 del paquete 2026-07-30-volverlo-de-arnesia-y-publicar, B-D5).
//
// MECANISMO puro — la política (guardas + gate de conformance) vive en el usecase. Secuencia:
// clone/pull del repo del marketplace en `~/.arnesia/publicaciones/<slug>/` → guarda de
// idempotencia (`plugins/<id>/<version>/` poblado ⇒ ya publicada, NADA se toca) → copia del
// árbol del canónico (mismo set de exclusión que HashFormaPlugin) → read-modify-write de
// `marketplace.json` + `catalogo.json` con map[string]any (los structs de parse.go son lossy a
// propósito: re-serializarlos destruiría claves ajenas) → commit → push SIN force → tag
// `<id>/vX.Y.Z` TRAS el push (fallo de tag = aviso visible, no rollback fantasma).
//
// Auth calcada de traer/externo.go (BR-18): `gh` como credential helper cuando existe, PAT vía
// GIT_ASKPASS si no — el token JAMÁS en argv ni en la URL remota.
package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Publisher es el ports.PublishPort real por `git`. Mismo shape que traer.ClonadorExterno:
// binarios inyectables (shim en tests), timeout acotado, raíz inyectable.
type Publisher struct {
	GitBin      string        // default "git".
	GHBin       string        // default "": sin credential helper de gh.
	Token       string        // PAT fallback (ARNESIA_GH_TOKEN); "" ⇒ no se usa.
	Timeout     time.Duration // default 120 s: acá hay clone + push, más que leer un catálogo.
	RaizArnesia string        // "" ⇒ ~/.arnesia — inyectable para tests.
}

var _ ports.PublishPort = (*Publisher)(nil)

func (p *Publisher) gitBin() string {
	if p.GitBin != "" {
		return p.GitBin
	}
	return "git"
}

func (p *Publisher) timeout() time.Duration {
	if p.Timeout > 0 {
		return p.Timeout
	}
	return 120 * time.Second
}

func (p *Publisher) raiz() (string, error) {
	if p.RaizArnesia != "" {
		return p.RaizArnesia, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("publicar: resolver home: %w", err)
	}
	return filepath.Join(home, ".arnesia"), nil
}

// Publicar ejecuta la secuencia completa (B-D5). El único error que deja el remoto tocado es el
// del tag — y ese NO es error: es éxito parcial con aviso (el branch ya viajó; deshacerlo
// fabricaría historia).
func (p *Publisher) Publicar(ctx context.Context, sol domain.SolicitudPublicacion) (domain.ResultadoPublicacion, error) {
	var vacio domain.ResultadoPublicacion
	if sol.RepoHome == "" || sol.ID == "" || sol.Version == "" || sol.OrigenDir == "" {
		return vacio, fmt.Errorf("publicar: solicitud incompleta (repo=%q id=%q version=%q origen=%q)",
			sol.RepoHome, sol.ID, sol.Version, sol.OrigenDir)
	}
	// El id y la versión forman segmentos de path del repo destino: un separador adentro
	// escaparía del árbol `plugins/`. La versión ya viene semver-validada por el usecase; esto
	// es defensa en profundidad, no la guarda principal.
	if strings.ContainsAny(sol.ID, `/\`) || strings.ContainsAny(sol.Version, `/\`) || sol.ID == ".." {
		return vacio, fmt.Errorf("publicar: id/version con separadores de path: %q / %q", sol.ID, sol.Version)
	}

	raiz, err := p.raiz()
	if err != nil {
		return vacio, err
	}
	slug, ok := domain.SlugRepo(sol.RepoHome)
	if !ok {
		return vacio, errors.New("publicar: repo home vacío")
	}
	base := domain.RaizPublicaciones(raiz)
	if merr := os.MkdirAll(base, 0o750); merr != nil {
		return vacio, fmt.Errorf("publicar: crear %s: %w", base, merr)
	}
	work := filepath.Join(base, slug)

	ctx, cancel := context.WithTimeout(ctx, p.timeout())
	defer cancel()

	askpass, limpiar, aerr := p.prepararAuth(base)
	if aerr != nil {
		return vacio, aerr
	}
	defer limpiar()

	// (a) workdir espejo del remoto: clone si falta; si existe, pull-antes-de-push (`fetch` +
	// `reset --hard origin/HEAD` + `clean -fdx`). El clean importa: un run anterior que falló
	// entre la copia y el push deja archivos UNTRACKED que el reset no toca — sin limpiarlos, la
	// guarda de idempotencia daría un 409 fabricado sobre una versión que el remoto NO tiene.
	url := "https://" + sol.RepoHome + ".git"
	if _, serr := os.Stat(filepath.Join(work, ".git")); serr != nil {
		if _, gerr := p.git(ctx, base, askpass, "clone", "--quiet", url, work); gerr != nil {
			return vacio, gerr
		}
	} else {
		if _, gerr := p.git(ctx, work, askpass, "fetch", "--quiet", "origin"); gerr != nil {
			return vacio, gerr
		}
		// set-head --auto asegura que origin/HEAD exista (un clone viejo puede no traerlo).
		if _, gerr := p.git(ctx, work, askpass, "remote", "set-head", "origin", "--auto"); gerr != nil {
			return vacio, gerr
		}
		if _, gerr := p.git(ctx, work, askpass, "reset", "-q", "--hard", "origin/HEAD"); gerr != nil {
			return vacio, gerr
		}
		if _, gerr := p.git(ctx, work, askpass, "clean", "-q", "-fdx"); gerr != nil {
			return vacio, gerr
		}
	}

	// (b) guarda de idempotencia: la versión YA está en el estante ⇒ nada se toca.
	destino := filepath.Join(work, "plugins", sol.ID, sol.Version)
	if poblado, perr := dirPoblado(destino); perr != nil {
		return vacio, fmt.Errorf("publicar: revisar %s: %w", destino, perr)
	} else if poblado {
		return vacio, fmt.Errorf("%w: plugins/%s/%s", domain.ErrPublicarVersionYaPublicada, sol.ID, sol.Version)
	}

	// (c) copiar el árbol del canónico, con el MISMO set de exclusión que HashFormaPlugin
	// (deriva.go): .git/, .in_use, .orphaned_at — así la copia publicada hashea igual que el
	// canónico y la deriva del que la instale da al-hilo.
	avisos, cerr := copiarArbolPublicable(sol.OrigenDir, destino)
	if cerr != nil {
		return vacio, fmt.Errorf("publicar: copiar el canónico: %w", cerr)
	}

	// (d) read-modify-write de los dos índices del marketplace.
	ahora := time.Now()
	if merr := reapuntarMarketplaceJSON(work, sol.ID, sol.Version); merr != nil {
		return vacio, merr
	}
	if cerr := estamparCatalogoJSON(work, sol.ID, sol.Version, ahora); cerr != nil {
		return vacio, cerr
	}

	// (e) commit. La identidad va explícita porque el env neutraliza la config global (mismo
	// hardening que traer): sin -c el commit fallaría en una máquina sin user.email.
	if _, gerr := p.git(ctx, work, askpass, "add", "-A"); gerr != nil {
		return vacio, gerr
	}
	mensaje := fmt.Sprintf("publish %s v%s", sol.ID, sol.Version)
	if _, gerr := p.git(ctx, work, askpass,
		"-c", "user.name=arnesia", "-c", "user.email=arnesia@alpacapurpura.lat",
		"commit", "--quiet", "-m", mensaje); gerr != nil {
		return vacio, gerr
	}

	// (f) push SIN force — innegociable: un rechazo es «alguien publicó antes», jamás se pisa.
	if _, gerr := p.git(ctx, work, askpass, "push", "--quiet", "origin", "HEAD"); gerr != nil {
		return vacio, gerr
	}
	commit, gerr := p.git(ctx, work, askpass, "rev-parse", "HEAD")
	if gerr != nil {
		return vacio, gerr
	}
	commit = strings.TrimSpace(commit)

	// (g) tag TRAS el push del branch: fallo ⇒ éxito parcial con aviso, NO rollback (B-D5).
	res := domain.ResultadoPublicacion{Commit: commit, Avisos: avisos}
	tag := sol.ID + "/v" + sol.Version
	if _, terr := p.git(ctx, work, askpass,
		"-c", "user.name=arnesia", "-c", "user.email=arnesia@alpacapurpura.lat",
		"tag", tag); terr != nil {
		res.Avisos = append(res.Avisos, "publicado sin tag: "+terr.Error())
		return res, nil
	}
	if _, terr := p.git(ctx, work, askpass, "push", "--quiet", "origin", "refs/tags/"+tag); terr != nil {
		res.Avisos = append(res.Avisos, "publicado sin tag: "+terr.Error())
		return res, nil
	}
	res.Tag = tag
	return res, nil
}

// prepararAuth arma la credencial SIN poner el token en argv (BR-18) — patrón EXACTO de
// traer.ClonadorExterno.prepararAuth (externo.go). DUPLICACIÓN DOCUMENTADA: go-arch-lint
// prohíbe `publish`→`traer` (los adapters no se importan entre sí); precedente
// `validarRootPortafolio` duplicada en usecase/portafolio.go.
func (p *Publisher) prepararAuth(dir string) (askpass string, limpiar func(), err error) {
	limpiar = func() {}
	if p.Token == "" {
		return "", limpiar, nil
	}
	ruta := filepath.Join(dir, "askpass.sh")
	script := "#!/bin/sh\nprintf '%s' \"$ARNESIA_GH_TOKEN\"\n"
	if werr := os.WriteFile(ruta, []byte(script), 0o700); werr != nil { //nolint:gosec // G306: tiene que ser ejecutable; vive bajo ~/.arnesia/publicaciones.
		return "", limpiar, fmt.Errorf("publicar: escribir GIT_ASKPASS: %w", werr)
	}
	return ruta, func() { _ = os.Remove(ruta) }, nil
}

// git corre un subcomando en dir y devuelve su stdout; el stderr real se clasifica a un
// centinela. Mismo hardening de env que traer (terminal prompts apagados, config global
// neutralizada — la config del operador no cambia el comportamiento).
func (p *Publisher) git(ctx context.Context, dir, askpass string, args ...string) (string, error) {
	base := []string{"-C", dir}
	if askpass == "" && p.GHBin != "" {
		base = append(base, "-c", "credential.helper=!"+p.GHBin+" auth git-credential")
	}
	base = append(base, args...)

	cmd := exec.CommandContext(ctx, p.gitBin(), base...) //nolint:gosec // G204: bin inyectable, args armados de una solicitud ya sancionada por el usecase.
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	cmd.WaitDelay = 2 * time.Second
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
	)
	if askpass != "" {
		cmd.Env = append(cmd.Env, "GIT_ASKPASS="+askpass, "ARNESIA_GH_TOKEN="+p.Token)
	}

	err := cmd.Run()
	if ctx.Err() != nil {
		return "", fmt.Errorf("publicar: git %s: timeout de %s", primerArg(args), p.timeout())
	}
	if err != nil {
		return "", clasificarErrorGitPublish(args, errBuf.String(), err)
	}
	return out.String(), nil
}

// primerArg devuelve el subcomando real (saltando los -c de identidad) para los mensajes.
func primerArg(args []string) string {
	for i := 0; i < len(args); i++ {
		if args[i] == "-c" {
			i++
			continue
		}
		return args[i]
	}
	return "git"
}

// clasificarErrorGitPublish mapea el stderr real de git a un centinela. CALCA la tabla de
// clasificarErrorGit (internal/adapters/traer/externo.go) + los rechazos de push que solo
// existen del lado write. DUPLICACIÓN DOCUMENTADA: go-arch-lint prohíbe `publish`→`traer` —
// mismo precedente que prepararAuth arriba.
func clasificarErrorGitPublish(args []string, stderr string, err error) error {
	s := strings.TrimSpace(stderr)
	if s == "" {
		s = err.Error()
	}
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "rejected"),
		strings.Contains(low, "non-fast-forward"),
		strings.Contains(low, "fetch first"):
		return fmt.Errorf("%w: git %s: %s", domain.ErrPublicarPushRechazado, primerArg(args), s)
	case strings.Contains(low, "could not read username"),
		strings.Contains(low, "authentication failed"),
		strings.Contains(low, "terminal prompts disabled"),
		strings.Contains(low, "403"):
		return fmt.Errorf("%w: git %s: %s", domain.ErrPublicarSinAuth, primerArg(args), s)
	default:
		return fmt.Errorf("publicar: git %s: %s", primerArg(args), s)
	}
}

// dirPoblado reporta si dir existe con contenido (mismo criterio que destinoPoblado de Traer:
// un dir vacío no cuenta; un archivo en la ruta sí).
func dirPoblado(dir string) (bool, error) {
	fi, err := os.Lstat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !fi.IsDir() {
		return true, nil
	}
	entradas, rerr := os.ReadDir(dir)
	if rerr != nil {
		return false, rerr
	}
	return len(entradas) > 0, nil
}

// copiarArbolPublicable copia origen→destino excluyendo EXACTAMENTE el set de HashFormaPlugin
// (internal/adapters/portafolio/deriva.go): `.git/`, `.in_use`, `.orphaned_at`. Los symlinks no
// se publican (un link al FS del operador no significa nada en el estante): se omiten con aviso
// VISIBLE, jamás en silencio.
func copiarArbolPublicable(origen, destino string) ([]string, error) {
	var avisos []string
	werr := filepath.WalkDir(origen, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(origen, path)
		if rerr != nil {
			return rerr
		}
		if rel == "." {
			return os.MkdirAll(destino, 0o750)
		}
		nombre := d.Name()
		if d.IsDir() && nombre == ".git" {
			return filepath.SkipDir
		}
		if nombre == ".in_use" || nombre == ".orphaned_at" {
			return nil
		}
		dest := filepath.Join(destino, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o750)
		}
		if d.Type()&fs.ModeSymlink != 0 {
			avisos = append(avisos, "symlink omitido de la publicación: "+rel)
			return nil
		}
		if !d.Type().IsRegular() {
			avisos = append(avisos, "archivo no-regular omitido de la publicación: "+rel)
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return ierr
		}
		return copiarArchivo(path, dest, info.Mode().Perm())
	})
	if werr != nil {
		return avisos, werr
	}
	return avisos, nil
}

// copiarArchivo copia un archivo regular preservando el bit de ejecución.
func copiarArchivo(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src) //nolint:gosec // G304: ruta bajo el canónico que el usecase sancionó.
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm) //nolint:gosec // G304: ruta bajo el workdir de publicaciones.
	if err != nil {
		return err
	}
	if _, cerr := io.Copy(out, in); cerr != nil {
		_ = out.Close()
		return cerr
	}
	return out.Close()
}

// reapuntarMarketplaceJSON hace el read-modify-write de `.claude-plugin/marketplace.json` con
// map[string]any — JAMÁS con los structs de parse.go, que son lossy a propósito (descartan las
// claves que el dominio no modela; re-serializarlos destruiría el archivo ajeno). Si existe la
// fila `name==id` se re-apunta su `source`; si no, se agrega una fila mínima. Las claves de las
// filas ajenas quedan intactas (el orden de claves lo normaliza el encoder; el CONTENIDO no se
// toca).
func reapuntarMarketplaceJSON(work, id, version string) error {
	ruta := filepath.Join(work, ".claude-plugin", "marketplace.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta dentro del workdir de publicaciones.
	if err != nil {
		return fmt.Errorf("publicar: el clone del marketplace no expone .claude-plugin/marketplace.json: %w", err)
	}
	var m map[string]any
	if uerr := json.Unmarshal(b, &m); uerr != nil {
		return fmt.Errorf("publicar: marketplace.json ilegible: %w", uerr)
	}
	source := "./plugins/" + id + "/" + version

	plugins, _ := m["plugins"].([]any)
	hallada := false
	for _, fila := range plugins {
		obj, ok := fila.(map[string]any)
		if !ok {
			continue
		}
		if nombre, _ := obj["name"].(string); nombre == id {
			obj["source"] = source
			hallada = true
		}
	}
	if !hallada {
		plugins = append(plugins, map[string]any{"name": id, "source": source})
	}
	m["plugins"] = plugins
	return escribirJSON(ruta, m)
}

// estamparCatalogoJSON agrega la fila de la versión a `catalogo.json` (convención de prenter):
// `versiones[] += {version, estado: "habilitada", fuente: "<id>@v<version>", fecha, sello}`.
// El sello AAMMDDHHMM es el sello de EXTRACCIÓN (conventions/versionado-arnes.md): campo
// ADITIVO, jamás dentro del string semver — CC compara strings. Los `canales` NO se tocan
// (B-D6). Si el archivo falta, se crea mínimo.
func estamparCatalogoJSON(work, id, version string, ahora time.Time) error {
	ruta := filepath.Join(work, "catalogo.json")
	cat := map[string]any{}
	if b, err := os.ReadFile(ruta); err == nil { //nolint:gosec // G304: ruta dentro del workdir.
		if uerr := json.Unmarshal(b, &cat); uerr != nil {
			return fmt.Errorf("publicar: catalogo.json ilegible: %w", uerr)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("publicar: leer catalogo.json: %w", err)
	} else if nombre := nombreDelMarketplace(work); nombre != "" {
		cat["marketplace"] = nombre
	}

	versiones, _ := cat["versiones"].([]any)
	versiones = append(versiones, map[string]any{
		"version": version,
		"estado":  "habilitada",
		"fuente":  id + "@v" + version,
		"fecha":   ahora.Format("2006-01-02"),
		"sello":   ahora.Format("0601021504"),
	})
	cat["versiones"] = versiones
	return escribirJSON(ruta, cat)
}

// nombreDelMarketplace lee el `name` del marketplace.json del workdir (para el catalogo.json
// recién creado). "" si no se puede — el campo es informativo, no se fabrica.
func nombreDelMarketplace(work string) string {
	b, err := os.ReadFile(filepath.Join(work, ".claude-plugin", "marketplace.json")) //nolint:gosec // G304: ruta dentro del workdir.
	if err != nil {
		return ""
	}
	var m struct {
		Name string `json:"name"`
	}
	if uerr := json.Unmarshal(b, &m); uerr != nil {
		return ""
	}
	return m.Name
}

// escribirJSON serializa con indent de 2 (la forma de los archivos reales) + newline final.
func escribirJSON(ruta string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, append(b, '\n'), 0o644) //nolint:gosec // G306: índice público del marketplace, versionado en git.
}
