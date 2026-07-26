package marketplace

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// shimGH escribe un script ejecutable que se hace pasar por `gh`: responde `auth status` con el
// exit code pedido y `api <endpoint>` con el cuerpo pedido. Es cómo se ejercita el camino remoto
// sin red y sin ningún login real (spec §8: el camino PAT se cubre con el selector puro).
func shimGH(t *testing.T, script string) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "gh")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"+script), 0o700); err != nil { //nolint:gosec // G306: shim de test, tiene que ser ejecutable.
		t.Fatal(err)
	}
	return bin
}

// shimConCatalogo emite el `marketplace.json` del fixture en base64 (igual que
// `gh api … --jq .content`) y 404 para todo lo demás.
func shimConCatalogo(t *testing.T, fixture string) string {
	t.Helper()
	b, err := os.ReadFile(fixture) //nolint:gosec // G304: fixture del propio repo.
	if err != nil {
		t.Fatal(err)
	}
	b64 := base64.StdEncoding.EncodeToString(b)
	return shimGH(t, fmt.Sprintf(`
case "$1" in
  auth) exit 0 ;;
  api)
    case "$2" in
      *".claude-plugin/marketplace.json") printf '%%s' '%s'; exit 0 ;;
      *) echo "gh: HTTP 404: Not Found" >&2; exit 1 ;;
    esac ;;
esac
exit 1
`, b64))
}

// E-15 · validar una url REAL contra el shim: el `gh` real nunca se invoca y los 3 datos
// (nombre · owner · nº de plugins) salen del ARCHIVO leído, no inferidos.
func TestLectorRemotoGHShimPrenter(t *testing.T) {
	l := &LectorRemoto{GHBin: shimConCatalogo(t, "testdata/prenter/.claude-plugin/marketplace.json")}
	cat, err := l.Validar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace")
	if err != nil {
		t.Fatalf("Validar: %v", err)
	}
	if cat.Marketplace != "prenter-marketplace" {
		t.Fatalf("Marketplace = %q", cat.Marketplace)
	}
	if cat.OwnerNombre != "Prenter" {
		t.Fatalf("OwnerNombre = %q, want Prenter", cat.OwnerNombre)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2", len(cat.Entradas))
	}
	if cat.Lectura.Fuente != "remoto" || cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura = %+v, want {leido … remoto}", cat.Lectura)
	}
	if cat.Repo != "github.com/alpacapurpura/prenter-marketplace" {
		t.Fatalf("Repo = %q, want canonicalizado", cat.Repo)
	}
	// El enriquecimiento ausente (404) degrada SIN ruido.
	if cat.Canales != nil || strings.Contains(cat.Lectura.Motivo, "catalogo.json") {
		t.Fatalf("un catalogo.json ausente no debe hacer ruido: %+v / %q", cat.Canales, cat.Lectura.Motivo)
	}
}

// E-54 · sin `gh` y sin PAT ⇒ ErrSinViaDeLectura (503, NUNCA 400): el sistema dice «no puedo
// mirar», jamás «tu url está mal».
func TestSinViaDeLectura503(t *testing.T) {
	l := &LectorRemoto{GHBin: "/no/existe/gh", Token: ""}
	_, err := l.Validar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace")
	if !errors.Is(err, ErrSinViaDeLectura) {
		t.Fatalf("err = %v, want ErrSinViaDeLectura", err)
	}
	if errors.Is(err, ErrNoEsMarketplace) {
		t.Fatal("«no puedo mirar» no debe clasificar como «no es un marketplace»")
	}
	if !strings.Contains(err.Error(), "ARNESIA_GH_TOKEN") {
		t.Fatalf("el motivo debe decir QUÉ falta: %v", err)
	}
}

// E-55 · el selector de credencial es PURO: tabla completa, cero login real.
func TestElegirCredencial(t *testing.T) {
	casos := []struct {
		gh, auth bool
		pat      string
		want     TipoCredencial
	}{
		{true, true, "", CredencialGH},
		{true, true, "x", CredencialGH},   // el CLI del operador PRIMERO (auth-terms).
		{true, false, "x", CredencialPAT}, // gh sin auth + PAT ⇒ PAT.
		{false, false, "x", CredencialPAT},
		{true, false, "", CredencialNinguna},
		{false, false, "", CredencialNinguna},
	}
	for _, c := range casos {
		got, motivo := elegirCredencial(c.gh, c.auth, c.pat)
		if got != c.want {
			t.Fatalf("elegirCredencial(%v,%v,%q) = %q, want %q", c.gh, c.auth, c.pat, got, c.want)
		}
		if got == CredencialNinguna && motivo == "" {
			t.Fatalf("elegirCredencial(%v,%v,%q): sin vía SIN motivo (degradado mudo)", c.gh, c.auth, c.pat)
		}
		if got != CredencialNinguna && motivo != "" {
			t.Fatalf("elegirCredencial(%v,%v,%q): motivo %q con vía elegida", c.gh, c.auth, c.pat, motivo)
		}
	}
}

// E-56 · timeout ⇒ sin-acceso con el motivo real y los segundos. El daemon sigue vivo.
func TestTimeoutEsSinAccesoConMotivo(t *testing.T) {
	l := &LectorRemoto{
		GHBin:   shimGH(t, `case "$1" in auth) exit 0 ;; api) sleep 30 ;; esac`),
		Timeout: 200 * time.Millisecond,
	}
	inicio := time.Now()
	_, err := l.Validar(context.Background(), "https://github.com/a/b")
	if err == nil {
		t.Fatal("Validar devolvió nil con el shim colgado")
	}
	if !errors.Is(err, ErrSinViaDeLectura) {
		t.Fatalf("err = %v, want ErrSinViaDeLectura (sin-acceso)", err)
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("el motivo debe decir timeout y los segundos: %v", err)
	}
	if time.Since(inicio) > 10*time.Second {
		t.Fatal("el timeout inyectado no se respetó")
	}
}

// E-57 · respuesta gigante ⇒ se corta por el techo de bytes, con el motivo visible. El proceso no
// crece sin control (io.LimitReader, no ReadAll a pelo).
func TestTechoDeBytes(t *testing.T) {
	// El shim emite ~9 MiB de base64 válido.
	l := &LectorRemoto{
		GHBin:    shimGH(t, `case "$1" in auth) exit 0 ;; api) head -c 9437184 /dev/zero | tr '\0' 'A' ;; esac`),
		MaxBytes: 8 << 20,
		Timeout:  30 * time.Second,
	}
	_, err := l.Validar(context.Background(), "https://github.com/a/b")
	if err == nil {
		t.Fatal("Validar aceptó una respuesta por encima del techo")
	}
	if !strings.Contains(err.Error(), "excede el techo de 8 MiB") {
		t.Fatalf("motivo = %v, want «excede el techo de 8 MiB»", err)
	}
}

// E-58 · más filas que el techo de entradas ⇒ recorte VISIBLE (`Truncado`), jamás silencioso.
func TestTechoDeEntradasEsVisible(t *testing.T) {
	filas := make([]map[string]string, 0, 6000)
	for i := range 6000 {
		filas = append(filas, map[string]string{"name": fmt.Sprintf("p%04d", i), "source": fmt.Sprintf("./plugins/p%04d/1.0.0", i)})
	}
	doc, err := json.Marshal(map[string]any{"name": "gigante", "plugins": filas})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	fixture := filepath.Join(dir, "marketplace.json")
	if werr := os.WriteFile(fixture, doc, 0o600); werr != nil {
		t.Fatal(werr)
	}

	cat, _, perr := parsearMarketplaceJSON(doc)
	if perr != nil {
		t.Fatal(perr)
	}
	if len(cat.Entradas) != maxEntradas {
		t.Fatalf("len(Entradas) = %d, want %d", len(cat.Entradas), maxEntradas)
	}
	if cat.Truncado != 6000-maxEntradas {
		t.Fatalf("Truncado = %d, want %d", cat.Truncado, 6000-maxEntradas)
	}
}

// E-59 · host que no es github.com ⇒ límite DECLARADO (503), no un fallo genérico.
func TestHostNoGitHub(t *testing.T) {
	l := &LectorRemoto{GHBin: shimConCatalogo(t, "testdata/prenter/.claude-plugin/marketplace.json")}
	_, err := l.Validar(context.Background(), "https://gitlab.com/a/b")
	if !errors.Is(err, ErrSinViaDeLectura) {
		t.Fatalf("err = %v, want ErrSinViaDeLectura", err)
	}
	if !strings.Contains(err.Error(), "solo se sabe leer catálogos de github.com por ahora") {
		t.Fatalf("motivo = %v, want el límite declarado", err)
	}
}

// E-60 · `gh` con 403 ⇒ sin-acceso (503) con el stderr REAL como motivo — DISTINGUIBLE del 404
// de E-16 (que clasifica ErrNoEsMarketplace ⇒ 400).
func TestGH401EsSinAcceso(t *testing.T) {
	l403 := &LectorRemoto{GHBin: shimGH(t, `case "$1" in auth) exit 0 ;; api) echo "gh: HTTP 403: Forbidden (https://api.github.com/…)" >&2; exit 1 ;; esac`)}
	_, err := l403.Validar(context.Background(), "https://github.com/a/privado")
	if !errors.Is(err, ErrSinViaDeLectura) {
		t.Fatalf("403: err = %v, want ErrSinViaDeLectura", err)
	}
	if !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("403: el stderr real debe viajar como motivo: %v", err)
	}

	l404 := &LectorRemoto{GHBin: shimGH(t, `case "$1" in auth) exit 0 ;; api) echo "gh: HTTP 404: Not Found" >&2; exit 1 ;; esac`)}
	_, err = l404.Validar(context.Background(), "https://github.com/alpacapurpura/no-existe-xyz")
	if !errors.Is(err, ErrNoEsMarketplace) {
		t.Fatalf("404: err = %v, want ErrNoEsMarketplace", err)
	}
	if errors.Is(err, ErrSinViaDeLectura) {
		t.Fatal("404 y 403 tienen que ser DISTINGUIBLES")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("404: el motivo debe traer el código real: %v", err)
	}
}

// Una url que no canonicaliza es un 400, no un 503 (E-16 lado adapter).
func TestValidarURLNoCanonicalizable(t *testing.T) {
	l := &LectorRemoto{GHBin: shimConCatalogo(t, "testdata/prenter/.claude-plugin/marketplace.json")}
	_, err := l.Validar(context.Background(), "solo-un-nombre")
	if err == nil || !strings.Contains(err.Error(), "no resuelve a host/owner/repo") {
		t.Fatalf("err = %v, want el motivo de url no canonicalizable", err)
	}
}

// La rama PAT (fallback de auth-terms) baja el archivo por HTTP sin `gh` — el token va en el
// header, jamás en el argv ni en la url.
func TestBajarPorPATFallback(t *testing.T) {
	b, err := os.ReadFile("testdata/prenter/.claude-plugin/marketplace.json")
	if err != nil {
		t.Fatal(err)
	}
	var authVisto, urlVisto string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authVisto, urlVisto = r.Header.Get("Authorization"), r.URL.String()
		if strings.HasSuffix(r.URL.Path, "catalogo.json") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"content": base64.StdEncoding.EncodeToString(b)})
	}))
	defer srv.Close()

	l := &LectorRemoto{GHBin: "/no/existe/gh", Token: "tok-secreto", HTTPClient: srv.Client()}
	// Se apunta el cliente al server de test reescribiendo el transporte.
	l.HTTPClient = &http.Client{Transport: redirigirA(srv.URL)}

	cat, err := l.Validar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace")
	if err != nil {
		t.Fatalf("Validar por PAT: %v", err)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2", len(cat.Entradas))
	}
	if authVisto != "token tok-secreto" {
		t.Fatalf("Authorization = %q, want el header con el PAT", authVisto)
	}
	if strings.Contains(urlVisto, "tok-secreto") {
		t.Fatalf("el token NO puede viajar en la url (queda en logs/reflog): %q", urlVisto)
	}
}

// redirigirA manda cualquier request al host del server de test, conservando el path.
type redirector struct {
	base string
}

func redirigirA(base string) http.RoundTripper { return &redirector{base: base} }

func (r *redirector) RoundTrip(req *http.Request) (*http.Response, error) {
	nueva := req.Clone(req.Context())
	base := strings.TrimPrefix(r.base, "http://")
	nueva.URL.Scheme = "http"
	nueva.URL.Host = base
	nueva.Host = base
	return http.DefaultTransport.RoundTrip(nueva)
}
