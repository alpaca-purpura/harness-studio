package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/marketplace"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// publisher_test.go prueba el publisher contra REPOS GIT REALES en t.TempDir (AC-a/b/c del
// spec §B2): la semilla es el fixture prenter de internal/adapters/marketplace/testdata (git
// init + commit + clone --bare como origin) y la verificación se hace CLONANDO el bare — se
// asserta lo que el remoto tiene, no lo que el workdir cree.

const repoHomeTest = "github.com/alpacapurpura/prenter-marketplace"

// requiereGit salta HONESTO si no hay git en PATH (spec: skip honesto, jamás un pass fabricado).
func requiereGit(t *testing.T) string {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("sin `git` en PATH: no se puede probar el publisher real")
	}
	return git
}

// gitT corre git con identidad fija y falla el test si el comando falla.
func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	base := []string{"-C", dir, "-c", "user.name=test", "-c", "user.email=test@test"}
	out, err := exec.CommandContext(t.Context(), "git", append(base, args...)...).CombinedOutput() //nolint:gosec // G204: helper de test, args fijos del propio test
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// copiarFixture copia el árbol del fixture prenter (marketplace.json + catalogo.json +
// plugins/harness/0.5.x) a dst.
func copiarFixture(t *testing.T, dst string) {
	t.Helper()
	src := filepath.Join("..", "marketplace", "testdata", "prenter")
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		dest := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dest, 0o750)
		}
		b, rerr2 := os.ReadFile(path) //nolint:gosec // fixture del repo
		if rerr2 != nil {
			return rerr2
		}
		return os.WriteFile(dest, b, 0o600) //nolint:gosec // G703: dest deriva del walk del fixture del propio repo
	})
	if err != nil {
		t.Fatal(err)
	}
}

// semillaBare arma el origin: fixture → repo de trabajo → clone --bare. Devuelve la ruta del
// bare y mapea la URL https del repoHome al bare vía GIT_CONFIG_* (env del proceso de test, que
// el publisher hereda — el hardening GIT_CONFIG_GLOBAL=/dev/null no apaga las vars).
func semillaBare(t *testing.T) (bare string) {
	t.Helper()
	requiereGit(t)
	tmp := t.TempDir()
	seed := filepath.Join(tmp, "seed")
	if err := os.MkdirAll(seed, 0o750); err != nil {
		t.Fatal(err)
	}
	copiarFixture(t, seed)
	gitT(t, seed, "init", "-q", "-b", "main")
	gitT(t, seed, "add", "-A")
	gitT(t, seed, "commit", "-q", "-m", "semilla prenter")
	bare = filepath.Join(tmp, "origin.git")
	gitT(t, tmp, "clone", "-q", "--bare", seed, bare)

	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "url."+bare+".insteadOf")
	t.Setenv("GIT_CONFIG_VALUE_0", "https://"+repoHomeTest+".git")
	return bare
}

// canonicoSintetico arma un canónico 0.6.0 con .git/.in_use adentro (el set de exclusión de
// HashFormaPlugin tiene que dejarlos afuera de la publicación).
func canonicoSintetico(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "canonico")
	if err := os.MkdirAll(filepath.Join(dir, ".claude-plugin"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o750); err != nil {
		t.Fatal(err)
	}
	escribir := func(rel, contenido string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	escribir(filepath.Join(".claude-plugin", "plugin.json"), `{"name":"harness","version":"0.6.0"}`+"\n")
	escribir("README.md", "arnés harness v0.6.0\n")
	escribir(filepath.Join(".git", "config"), "no debe publicarse\n")
	escribir(".in_use", "no debe publicarse\n")
	escribir(".orphaned_at", "no debe publicarse\n")
	return dir
}

func publisherDePrueba(t *testing.T) *Publisher {
	t.Helper()
	return &Publisher{RaizArnesia: filepath.Join(t.TempDir(), "arnesia")}
}

func solicitud060(canonico string) domain.SolicitudPublicacion {
	return domain.SolicitudPublicacion{
		RepoHome: repoHomeTest, ID: "harness", Version: "0.6.0", OrigenDir: canonico,
	}
}

// AC-a · el camino feliz, verificado EN EL BARE (clonándolo): árbol publicado sin el set de
// exclusión, marketplace.json re-parseable por parse.go con el source re-apuntado y las filas
// ajenas intactas, catalogo.json con la fila nueva (sello AAMMDDHHMM aditivo), tag <id>/vX.Y.Z.
func TestPublicarFelizContraBare(t *testing.T) {
	bare := semillaBare(t)
	p := publisherDePrueba(t)

	res, err := p.Publicar(context.Background(), solicitud060(canonicoSintetico(t)))
	if err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(res.Commit) {
		t.Fatalf("Commit = %q, want 40 hex", res.Commit)
	}
	if res.Tag != "harness/v0.6.0" {
		t.Fatalf("Tag = %q, want harness/v0.6.0", res.Tag)
	}

	// La verificación clona el BARE: se asserta lo que el remoto tiene de verdad.
	verif := filepath.Join(t.TempDir(), "verif")
	gitT(t, t.TempDir(), "clone", "-q", bare, verif)

	b, rerr := os.ReadFile(filepath.Join(verif, "plugins", "harness", "0.6.0", ".claude-plugin", "plugin.json")) //nolint:gosec // G304: ruta del clone de verificación del test
	if rerr != nil {
		t.Fatalf("el árbol publicado no está en el remoto: %v", rerr)
	}
	if !strings.Contains(string(b), `"0.6.0"`) {
		t.Fatalf("plugin.json publicado = %s", b)
	}
	// El set de exclusión de HashFormaPlugin quedó FUERA de la publicación.
	for _, excl := range []string{".git", ".in_use", ".orphaned_at"} {
		if _, serr := os.Stat(filepath.Join(verif, "plugins", "harness", "0.6.0", excl)); serr == nil {
			t.Fatalf("%s viajó al remoto (el set de exclusión de HashFormaPlugin se rompió)", excl)
		}
	}

	// Round-trip por parse.go (LectorLocal lee el checkout como lo haría el estante).
	lector := &marketplace.LectorLocal{}
	cat, lerr := lector.Leer(context.Background(), domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", InstallLocation: verif,
	})
	if lerr != nil {
		t.Fatalf("el marketplace.json publicado no re-parsea: %v", lerr)
	}
	var harness, beta *domain.EntradaCatalogo
	for i := range cat.Entradas {
		switch cat.Entradas[i].Nombre {
		case "harness":
			harness = &cat.Entradas[i]
		case "harness-beta":
			beta = &cat.Entradas[i]
		}
	}
	if harness == nil || harness.Source.Crudo != "./plugins/harness/0.6.0" {
		t.Fatalf("source de harness no re-apuntado: %+v", harness)
	}
	if beta == nil || beta.Source.Crudo != "./plugins/harness/0.5.3" {
		t.Fatalf("la fila ajena harness-beta se tocó: %+v", beta)
	}

	// catalogo.json: la fila nueva con el sello AAMMDDHHMM ADITIVO (jamás dentro del semver) y
	// los canales INTACTOS (B-D6).
	raw, rerr2 := os.ReadFile(filepath.Join(verif, "catalogo.json")) //nolint:gosec // G304: ruta del clone de verificación del test
	if rerr2 != nil {
		t.Fatal(rerr2)
	}
	var catJSON struct {
		Canales   map[string]string `json:"canales"`
		Versiones []map[string]any  `json:"versiones"`
	}
	if err := json.Unmarshal(raw, &catJSON); err != nil {
		t.Fatalf("catalogo.json publicado ilegible: %v", err)
	}
	if catJSON.Canales["estable"] != "0.5.3" || catJSON.Canales["beta"] != "0.5.3" {
		t.Fatalf("los canales se mutaron (B-D6 los prohíbe): %v", catJSON.Canales)
	}
	var fila map[string]any
	for _, v := range catJSON.Versiones {
		if v["version"] == "0.6.0" {
			fila = v
		}
	}
	if fila == nil {
		t.Fatalf("falta la fila 0.6.0 en catalogo.json: %s", raw)
	}
	if fila["estado"] != "habilitada" || fila["fuente"] != "harness@v0.6.0" {
		t.Fatalf("fila = %v", fila)
	}
	sello, _ := fila["sello"].(string)
	if !regexp.MustCompile(`^\d{10}$`).MatchString(sello) {
		t.Fatalf("sello = %q, want AAMMDDHHMM (10 dígitos)", sello)
	}
	if fecha, _ := fila["fecha"].(string); !regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(fecha) {
		t.Fatalf("fecha = %q", fecha)
	}

	// El tag vive en el REMOTO.
	tags := gitT(t, bare, "tag", "-l")
	if !strings.Contains(tags, "harness/v0.6.0") {
		t.Fatalf("el tag no llegó al remoto: %q", tags)
	}
}

// AC-b · idempotencia: republicar la MISMA versión ⇒ ErrPublicarVersionYaPublicada, el working
// tree queda intacto (status limpio) y el remoto no gana ni un commit.
func TestPublicarIdempotente(t *testing.T) {
	bare := semillaBare(t)
	p := publisherDePrueba(t)
	canonico := canonicoSintetico(t)

	if _, err := p.Publicar(context.Background(), solicitud060(canonico)); err != nil {
		t.Fatalf("primer Publicar: %v", err)
	}
	antes := strings.TrimSpace(gitT(t, bare, "rev-list", "--count", "HEAD"))

	_, err := p.Publicar(context.Background(), solicitud060(canonico))
	if !errors.Is(err, domain.ErrPublicarVersionYaPublicada) {
		t.Fatalf("err = %v, want ErrPublicarVersionYaPublicada", err)
	}
	despues := strings.TrimSpace(gitT(t, bare, "rev-list", "--count", "HEAD"))
	if antes != despues {
		t.Fatalf("el remoto ganó commits en un republish (%s → %s)", antes, despues)
	}
	slug, _ := domain.SlugRepo(repoHomeTest)
	work := filepath.Join(domain.RaizPublicaciones(p.RaizArnesia), slug)
	if status := gitT(t, work, "status", "--porcelain"); strings.TrimSpace(status) != "" {
		t.Fatalf("el working tree quedó tocado tras el 409: %q", status)
	}
}

// La guarda de idempotencia también protege lo que YA estaba en la semilla (0.5.3 poblado).
func TestPublicarVersionDeLaSemillaYaPublicada(t *testing.T) {
	semillaBare(t)
	p := publisherDePrueba(t)
	sol := solicitud060(canonicoSintetico(t))
	sol.Version = "0.5.3"

	if _, err := p.Publicar(context.Background(), sol); !errors.Is(err, domain.ErrPublicarVersionYaPublicada) {
		t.Fatalf("err = %v, want ErrPublicarVersionYaPublicada", err)
	}
}

// AC-c · push rechazado: un COMPETIDOR pushea entre el fetch y el push (shim de git que, ante
// el push del publisher, primero avanza el remoto desde otro clon) ⇒ ErrPublicarPushRechazado
// — jamás un --force que lo pise.
func TestPublicarPushRechazado(t *testing.T) {
	bare := semillaBare(t)
	p := publisherDePrueba(t)

	// El clon competidor, listo para pushear.
	competidor := filepath.Join(t.TempDir(), "competidor")
	gitT(t, t.TempDir(), "clone", "-q", bare, competidor)

	// Shim: delega TODO a git real, pero cuando ve `push … origin HEAD` hace pushear al
	// competidor primero. Reproduce la carrera real de forma determinista.
	shimDir := t.TempDir()
	shim := filepath.Join(shimDir, "git-shim.sh")
	script := fmt.Sprintf(`#!/bin/sh
case "$*" in
  *"push --quiet origin HEAD"*)
    git -C %q -c user.name=rival -c user.email=rival@test commit -q --allow-empty -m rival >/dev/null 2>&1
    git -C %q push -q origin HEAD >/dev/null 2>&1
    ;;
esac
exec git "$@"
`, competidor, competidor)
	if err := os.WriteFile(shim, []byte(script), 0o700); err != nil { //nolint:gosec // shim de test
		t.Fatal(err)
	}
	p.GitBin = shim

	_, err := p.Publicar(context.Background(), solicitud060(canonicoSintetico(t)))
	if !errors.Is(err, domain.ErrPublicarPushRechazado) {
		t.Fatalf("err = %v, want ErrPublicarPushRechazado", err)
	}
}

// La tabla de clasificación calcada de traer (B-D5): auth ≠ rechazo ≠ genérico — motivos
// DISTINGUIBLES, ninguno se traga.
func TestClasificarErrorGitPublish(t *testing.T) {
	casos := []struct {
		stderr string
		want   error
	}{
		{"! [rejected]        main -> main (fetch first)", domain.ErrPublicarPushRechazado},
		{"error: failed to push some refs: non-fast-forward", domain.ErrPublicarPushRechazado},
		{"fatal: could not read Username for 'https://github.com'", domain.ErrPublicarSinAuth},
		{"remote: HTTP 403", domain.ErrPublicarSinAuth},
		{"fatal: Authentication failed for 'https://github.com'", domain.ErrPublicarSinAuth},
		{"fatal: terminal prompts disabled", domain.ErrPublicarSinAuth},
	}
	for _, c := range casos {
		got := clasificarErrorGitPublish([]string{"push"}, c.stderr, errors.New("exit status 1"))
		if !errors.Is(got, c.want) {
			t.Fatalf("stderr %q → %v, want %v", c.stderr, got, c.want)
		}
	}
	generico := clasificarErrorGitPublish([]string{"push"}, "fatal: algo raro", errors.New("exit status 1"))
	if errors.Is(generico, domain.ErrPublicarSinAuth) || errors.Is(generico, domain.ErrPublicarPushRechazado) {
		t.Fatalf("un stderr desconocido no puede clasificarse a un centinela: %v", generico)
	}
}

// Los symlinks NO viajan al estante: se omiten con aviso VISIBLE (jamás en silencio).
func TestPublicarSymlinkOmitidoConAviso(t *testing.T) {
	semillaBare(t)
	p := publisherDePrueba(t)
	canonico := canonicoSintetico(t)
	if err := os.Symlink("/etc/hostname", filepath.Join(canonico, "link-al-fs")); err != nil {
		t.Skipf("sin symlinks en este FS: %v", err)
	}

	res, err := p.Publicar(context.Background(), solicitud060(canonico))
	if err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	var hallado bool
	for _, a := range res.Avisos {
		if strings.Contains(a, "symlink omitido") && strings.Contains(a, "link-al-fs") {
			hallado = true
		}
	}
	if !hallado {
		t.Fatalf("falta el aviso del symlink omitido: %v", res.Avisos)
	}
}
