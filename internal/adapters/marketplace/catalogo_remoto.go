package marketplace

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// catalogo_remoto.go lee el `marketplace.json` de un repo SIN clonarlo: `gh api` (shell-out al
// CLI YA AUTENTICADO del operador) con PAT como fallback.
//
// Por qué `gh api` y no un clone shallow (design.md §5.3): 1-2 requests / ~160 KB peor caso vs
// el árbol completo; cero escrituras en disco; cero duplicación de lo que CC ya clona en
// `installLocation`. Y sobre todo AUTH-TERMS: **la app es conductor, no proxy** — invoca el CLI
// del operador y JAMÁS ve, pide ni guarda una credencial propia (BR-18). Precedentes en el
// árbol: `adapters/selfupdate/updater.go` y `adapters/publish/publisher.go` ya hacen shell-out.

// TipoCredencial es la vía de acceso elegida. El orden es doctrina (auth-terms): el CLI del
// operador primero — la app JAMÁS proxya un login ni guarda una credencial propia.
type TipoCredencial string

// Las tres vías posibles (la tercera es la ausencia honesta de vía).
const (
	CredencialGH      TipoCredencial = "gh"
	CredencialPAT     TipoCredencial = "pat"
	CredencialNinguna TipoCredencial = ""
)

// elegirCredencial es PURA: decide con los hechos que el caller ya averiguó. Se testea sin
// ningún login real (E-55, spec §8).
func elegirCredencial(ghDisponible, ghAutenticado bool, pat string) (TipoCredencial, string) {
	if ghDisponible && ghAutenticado {
		return CredencialGH, ""
	}
	if pat != "" {
		return CredencialPAT, ""
	}
	switch {
	case !ghDisponible:
		return CredencialNinguna, "`gh` no está en el PATH y ARNESIA_GH_TOKEN está vacío"
	default:
		return CredencialNinguna, "`gh` está instalado pero no autenticado, y ARNESIA_GH_TOKEN está vacío"
	}
}

// LectorRemoto satisface ports.CatalogoReader y ports.CatalogoValidador.
type LectorRemoto struct {
	GHBin      string        // default "gh" resuelto por PATH; inyectable (shim en tests).
	Token      string        // PAT fallback (ARNESIA_GH_TOKEN); "" = no usar.
	Timeout    time.Duration // default 10 s: un `gh` colgado no cuelga el daemon (E-41/E-56).
	MaxBytes   int64         // default 8 MiB (io.LimitReader): un archivo hostil no OOMea (E-57).
	HTTPClient *http.Client  // solo la rama PAT; nil ⇒ default con Timeout.
}

var (
	_ ports.CatalogoReader    = (*LectorRemoto)(nil)
	_ ports.CatalogoValidador = (*LectorRemoto)(nil)
)

// Fuente identifica la vía para EstadoLectura.Fuente (BR-3, trazabilidad).
func (l *LectorRemoto) Fuente() string { return "remoto" }

func (l *LectorRemoto) ghBin() string {
	if l.GHBin != "" {
		return l.GHBin
	}
	return "gh"
}

func (l *LectorRemoto) timeout() time.Duration {
	if l.Timeout > 0 {
		return l.Timeout
	}
	return 10 * time.Second
}

func (l *LectorRemoto) maxBytes() int64 {
	if l.MaxBytes > 0 {
		return l.MaxBytes
	}
	return 8 << 20
}

// Leer lee el catálogo del repo declarado por m (canonicalizado). Requiere `Repo`: sin repo no
// hay a dónde mirar.
func (l *LectorRemoto) Leer(ctx context.Context, m domain.MarketplaceConocido) (domain.Catalogo, error) {
	cat, err := l.leerDeRepo(ctx, m.Repo)
	if err != nil {
		return domain.Catalogo{}, err
	}
	cat.Clase = m.Clase
	cat.Repo = m.Repo
	if m.Nombre != "" {
		cat.Marketplace = m.Nombre
	}
	return cat, nil
}

// Validar prueba una URL ANTES de registrar (S6, BR-5/G3): lee el archivo real y devuelve lo que
// LEYÓ. Nunca persiste nada. El nombre del catálogo es el del ARCHIVO (todavía no hay registro).
func (l *LectorRemoto) Validar(ctx context.Context, url string) (domain.Catalogo, error) {
	canon, ok := domain.CanonicalizarRepo(url)
	if !ok {
		return domain.Catalogo{}, fmt.Errorf("marketplace: url %q no resuelve a host/owner/repo", url)
	}
	cat, err := l.leerDeRepo(ctx, canon)
	if err != nil {
		return domain.Catalogo{}, err
	}
	cat.Repo = canon
	return cat, nil
}

// leerDeRepo es el mecanismo compartido: resuelve owner/repo, elige la credencial, baja el
// archivo y lo parsea. El enriquecimiento `catalogo.json` es best-effort (404 = ausente, normal).
func (l *LectorRemoto) leerDeRepo(ctx context.Context, repoCanon string) (domain.Catalogo, error) {
	owner, repo, err := ownerRepoDe(repoCanon)
	if err != nil {
		return domain.Catalogo{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, l.timeout())
	defer cancel()

	cred, motivo := elegirCredencial(l.ghDisponible(), l.ghAutenticado(ctx), l.Token)
	if cred == CredencialNinguna {
		return domain.Catalogo{}, fmt.Errorf("%w: %s", ErrSinViaDeLectura, motivo)
	}

	b, err := l.bajarArchivo(ctx, cred, owner, repo, ".claude-plugin/marketplace.json")
	if err != nil {
		// E-16 vs E-17: `gh api` devuelve 404 tanto si el REPO no existe como si existe pero no
		// tiene el archivo, y el plan de pruebas exige que los dos mensajes sean DISTINGUIBLES.
		// Una sonda extra al repo (SOLO en la rama de error, 1 request) resuelve la ambigüedad:
		// sin ella el operador no puede saber si escribió mal la url o si el repo no es un
		// marketplace (C25 de design.md).
		return domain.Catalogo{}, l.desambiguar404(ctx, cred, owner, repo, err)
	}
	cat, motivoParse, perr := parsearMarketplaceJSON(b)
	if perr != nil {
		return domain.Catalogo{}, perr
	}

	// Enriquecimiento OPCIONAL: un 404 es lo NORMAL (es convención de prenter, no del estándar) ⇒
	// degrada sin ruido. En remoto NO se verifican las rutas relativas de las filas: exigiría N
	// requests para recorrer el árbol ajeno, y eso sí sería la app haciendo de proxy (§5.2.6).
	if eb, eerr := l.bajarArchivo(ctx, cred, owner, repo, "catalogo.json"); eerr == nil {
		if aviso := aplicarEnriquecimiento(&cat, eb); aviso != "" {
			motivoParse = strings.TrimSpace(motivoParse + " " + aviso)
		}
	}

	cat.Lectura = domain.EstadoLectura{
		Tipo: domain.LecturaLeida, Entradas: len(cat.Entradas), Motivo: motivoParse, Fuente: l.Fuente(),
	}
	return cat, nil
}

// desambiguar404 distingue «el repo no existe / no tengo acceso» de «el repo existe pero no expone
// el catálogo» (E-16 vs E-17). Solo corre en la rama de error y solo si el error es un 404: para
// un 401/403 la respuesta ya es inequívoca («no puedo mirar»), y sondear de nuevo no aportaría.
// Si la SONDA misma falla por otra razón, se dice — jamás se atribuye el 404 al repo sin haberlo
// comprobado.
func (l *LectorRemoto) desambiguar404(ctx context.Context, cred TipoCredencial, owner, repo string, original error) error {
	if !errors.Is(original, ErrNoEsMarketplace) || !strings.Contains(original.Error(), "404") {
		return original
	}
	existe, motivoSonda := l.repoExiste(ctx, cred, owner, repo)
	switch {
	case existe:
		return fmt.Errorf("%w: el repo %s/%s EXISTE pero no tiene `.claude-plugin/marketplace.json`: no es un marketplace",
			ErrNoEsMarketplace, owner, repo)
	case motivoSonda != "":
		// La sonda no pudo comprobar nada: se informa el 404 original Y por qué no se pudo
		// desambiguar, en vez de afirmar que el repo no existe.
		return fmt.Errorf("%w: 404 en `.claude-plugin/marketplace.json` de %s/%s; no se pudo comprobar si el repo existe (%s)",
			ErrNoEsMarketplace, owner, repo, motivoSonda)
	default:
		return fmt.Errorf("%w: el repo %s/%s no existe o no tenés acceso con la credencial actual (404 al repo, no al archivo)",
			ErrNoEsMarketplace, owner, repo)
	}
}

// repoExiste sondea `repos/<owner>/<repo>` (un campo que SIEMPRE está, `full_name`) para saber si
// el 404 fue del repo o del archivo. Devuelve (existe, motivo-si-no-se-pudo-comprobar).
func (l *LectorRemoto) repoExiste(ctx context.Context, cred TipoCredencial, owner, repo string) (bool, string) {
	if cred != CredencialGH {
		// Por PAT la sonda sería un request más contra api.github.com; se declara el límite en vez
		// de afirmar algo no comprobado.
		return false, "la sonda de existencia del repo solo está implementada por `gh`"
	}
	cmd := exec.CommandContext(ctx, l.ghBin(), "api", fmt.Sprintf("repos/%s/%s", owner, repo), "--jq", ".full_name") //nolint:gosec // G204: bin inyectable, endpoint armado de un repo canonicalizado.
	var out, errBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errBuf
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return false, "timeout de la sonda"
		}
		low := strings.ToLower(errBuf.String())
		if strings.Contains(low, "404") || strings.Contains(low, "not found") {
			return false, "" // comprobado: el repo NO existe (o no es visible con esta credencial).
		}
		return false, strings.TrimSpace(errBuf.String())
	}
	return strings.TrimSpace(out.String()) != "", ""
}

// ownerRepoDe parte un repo canónico "host/owner/repo" y exige github.com: no se inventa una API
// para un host desconocido (E-59, límite DECLARADO, no un fallo genérico).
func ownerRepoDe(repoCanon string) (owner, repo string, err error) {
	if repoCanon == "" {
		return "", "", fmt.Errorf("%w: el marketplace no declara repo", ErrSinViaDeLectura)
	}
	partes := strings.Split(repoCanon, "/")
	if len(partes) < 3 {
		return "", "", fmt.Errorf("%w: repo %q no tiene forma host/owner/repo", ErrSinViaDeLectura, repoCanon)
	}
	if partes[0] != "github.com" {
		return "", "", fmt.Errorf("%w: solo se sabe leer catálogos de github.com por ahora (host: %s)", ErrSinViaDeLectura, partes[0])
	}
	return partes[1], strings.Join(partes[2:], "/"), nil
}

// ghDisponible reporta si el binario de `gh` se resuelve. No lo ejecuta.
func (l *LectorRemoto) ghDisponible() bool {
	bin := l.ghBin()
	if strings.ContainsRune(bin, '/') {
		// Ruta explícita (shim de test o binario fuera del PATH): se prueba tal cual.
		_, err := exec.LookPath(bin)
		return err == nil
	}
	_, err := exec.LookPath(bin)
	return err == nil
}

// ghAutenticado corre `gh auth status`. Exit 0 ⇒ autenticado. NO lee ni guarda el token: solo
// pregunta al CLI del operador si él lo tiene (auth-terms, BR-18).
func (l *LectorRemoto) ghAutenticado(ctx context.Context) bool {
	if !l.ghDisponible() {
		return false
	}
	cmd := exec.CommandContext(ctx, l.ghBin(), "auth", "status") //nolint:gosec // G204: bin inyectable, args fijos.
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	cmd.WaitDelay = 2 * time.Second
	return cmd.Run() == nil
}

// bajarArchivo trae `path` del repo por la credencial elegida y devuelve sus bytes YA decodeados.
func (l *LectorRemoto) bajarArchivo(ctx context.Context, cred TipoCredencial, owner, repo, path string) ([]byte, error) {
	if cred == CredencialPAT {
		return l.bajarPorPAT(ctx, owner, repo, path)
	}
	return l.bajarPorGH(ctx, owner, repo, path)
}

// bajarPorGH corre `gh api repos/<owner>/<repo>/contents/<path> --jq .content` y decodea el
// base64. El stderr real viaja como motivo (E-16/E-17/E-60 exigen mensajes DISTINGUIBLES).
//
// El stdout se lee por PIPE con io.LimitReader (nunca un `bytes.Buffer` sin techo ni un ReadAll a
// pelo): un archivo hostil no puede OOMear el daemon (E-57), y pasado el techo el proceso se MATA
// en vez de esperar a que llene la pipe.
func (l *LectorRemoto) bajarPorGH(ctx context.Context, owner, repo, path string) ([]byte, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/contents/%s", owner, repo, path)
	cmd := exec.CommandContext(ctx, l.ghBin(), "api", endpoint, "--jq", ".content") //nolint:gosec // G204: bin inyectable, endpoint armado de un repo canonicalizado.
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	// WaitDelay: un `gh` colgado (o un nieto que heredó la pipe) no puede colgar el daemon
	// esperando que el pipe cierre después de matar al proceso (E-56).
	cmd.WaitDelay = 2 * time.Second

	stdout, perr := cmd.StdoutPipe()
	if perr != nil {
		return nil, fmt.Errorf("%w: gh: pipe: %w", ErrSinViaDeLectura, perr)
	}
	if serr := cmd.Start(); serr != nil {
		return nil, fmt.Errorf("%w: gh: no se pudo ejecutar: %w", ErrSinViaDeLectura, serr)
	}
	raw, rerr := io.ReadAll(io.LimitReader(stdout, l.maxBytes()+1))
	excedido := int64(len(raw)) > l.maxBytes()
	if excedido && cmd.Process != nil {
		_ = cmd.Process.Kill() // no esperamos a que llene la pipe: ya sabemos que no lo queremos.
	}
	werr := cmd.Wait()

	switch {
	case ctx.Err() != nil:
		// E-56: timeout / cancelación ⇒ sin-acceso con el motivo real y los segundos. El caché
		// viejo se conserva intacto (el usecase no llama Guardar en esta rama).
		return nil, fmt.Errorf("%w: gh: timeout de %s: %w", ErrSinViaDeLectura, l.timeout(), ctx.Err())
	case excedido:
		return nil, fmt.Errorf("%w: la respuesta excede el techo de %d MiB", ErrNoEsMarketplace, l.maxBytes()>>20)
	case werr != nil:
		return nil, clasificarErrorGH(errBuf.String(), werr)
	case rerr != nil:
		return nil, fmt.Errorf("%w: gh: leer respuesta: %w", ErrSinViaDeLectura, rerr)
	}
	return decodificarContenido(raw, l.maxBytes())
}

// clasificarErrorGH mapea el stderr REAL de `gh` a un centinela: 404 ⇒ no-es-marketplace (400),
// 401/403 ⇒ sin-vía (503). Los motivos quedan DISTINGUIBLES, ninguno genérico.
func clasificarErrorGH(stderr string, err error) error {
	s := strings.TrimSpace(stderr)
	if s == "" {
		s = err.Error()
	}
	low := strings.ToLower(s)
	switch {
	case strings.Contains(low, "http 401"), strings.Contains(low, "http 403"),
		strings.Contains(low, "authentication"), strings.Contains(low, "not logged"):
		return fmt.Errorf("%w: gh: %s", ErrSinViaDeLectura, s)
	default:
		// 404 y cualquier otro: el repo/archivo no expone el catálogo. El motivo REAL viaja.
		return fmt.Errorf("%w: gh: %s", ErrNoEsMarketplace, s)
	}
}

// bajarPorPAT es el fallback de auth-terms (spec §8): el PAT es del OPERADOR y viaja en el header
// del request, jamás se persiste ni se loguea.
func (l *LectorRemoto) bajarPorPAT(ctx context.Context, owner, repo, path string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: armar request: %w", ErrSinViaDeLectura, err)
	}
	req.Header.Set("Authorization", "token "+l.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	cli := l.HTTPClient
	if cli == nil {
		cli = &http.Client{Timeout: l.timeout()}
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: pat: %w", ErrSinViaDeLectura, err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch {
	case resp.StatusCode == http.StatusUnauthorized, resp.StatusCode == http.StatusForbidden:
		return nil, fmt.Errorf("%w: pat: HTTP %d — la credencial no da acceso a %s/%s", ErrSinViaDeLectura, resp.StatusCode, owner, repo)
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: pat: HTTP %d en %s", ErrNoEsMarketplace, resp.StatusCode, path)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, l.maxBytes()+1))
	if err != nil {
		return nil, fmt.Errorf("%w: pat: leer respuesta: %w", ErrSinViaDeLectura, err)
	}
	if int64(len(b)) > l.maxBytes() {
		return nil, fmt.Errorf("%w: la respuesta excede el techo de %d MiB", ErrNoEsMarketplace, l.maxBytes()>>20)
	}
	var cuerpo struct {
		Content string `json:"content"`
	}
	if uerr := json.Unmarshal(b, &cuerpo); uerr != nil {
		return nil, fmt.Errorf("%w: pat: respuesta ilegible: %w", ErrNoEsMarketplace, uerr)
	}
	return decodificarContenido([]byte(strconvQuote(cuerpo.Content)), l.maxBytes())
}

// strconvQuote envuelve s en comillas JSON para reusar decodificarContenido (que acepta la salida
// cruda de `--jq .content`, o sea un string JSON con saltos de línea escapados).
func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// decodificarContenido acepta la salida de `gh api … --jq .content` (un string JSON con base64,
// o el texto crudo si el shim lo emitió pelado) y devuelve los bytes del archivo. El techo se
// aplica con io.LimitReader — nunca un ReadAll a pelo (E-57).
func decodificarContenido(raw []byte, maxBytes int64) ([]byte, error) {
	texto := strings.TrimSpace(string(raw))
	if texto == "" {
		return nil, fmt.Errorf("%w: gh devolvió una respuesta vacía", ErrNoEsMarketplace)
	}
	if strings.HasPrefix(texto, `"`) {
		var s string
		if err := json.Unmarshal([]byte(texto), &s); err == nil {
			texto = s
		}
	}
	excede := func() error {
		return fmt.Errorf("%w: la respuesta excede el techo de %d MiB", ErrNoEsMarketplace, maxBytes>>20)
	}
	// El `content` de la API de GitHub viene base64 con saltos de línea cada 60 chars.
	limpio := strings.NewReplacer("\n", "", "\r", "", " ", "").Replace(texto)
	dec := base64.NewDecoder(base64.StdEncoding, strings.NewReader(limpio))
	if b, err := io.ReadAll(io.LimitReader(dec, maxBytes+1)); err == nil && len(bytes.TrimSpace(b)) > 0 {
		if int64(len(b)) > maxBytes {
			return nil, excede()
		}
		return b, nil
	}
	// No era base64: el shim (o un endpoint futuro) devolvió el JSON pelado. Mismo techo.
	if int64(len(texto)) > maxBytes {
		return nil, excede()
	}
	return []byte(texto), nil
}
