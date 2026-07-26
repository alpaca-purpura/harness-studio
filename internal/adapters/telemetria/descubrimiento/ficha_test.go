package descubrimiento

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

func fichaDe(t *testing.T) (*Ficha, string) {
	t.Helper()
	dir := t.TempDir()
	f, err := New(filepath.Join(dir, "arnesia", "daemon.json"))
	if err != nil {
		t.Fatal(err)
	}
	return f, dir
}

func daemonFicha(endpoint string) domain.FichaDaemon {
	d := domain.FichaDaemon{
		Version: 1, PID: os.Getpid(), Endpoint: endpoint,
		RutaProceso:  "/api/telemetria/proceso",
		TokenIngesta: "TOKEN-ACOTADO", Binario: "/usr/bin/arnesia",
		Desde: time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC),
	}
	d.OTLP.Logs = "/v1/logs"
	d.OTLP.Metrics = "/v1/metrics"
	return d
}

// TestFichaSoloTrasEscuchar — la ficha no existe antes de que el listener acepte; existe
// después Y su endpoint responde; y desaparece en el shutdown.
//
// El control positivo es el que le da sentido: no alcanza con que el archivo exista — el
// endpoint que nombra tiene que estar VIVO. Una ficha que apunta a un puerto muerto es una
// mentira que el hook cobra en timeouts.
func TestFichaSoloTrasEscuchar(t *testing.T) {
	f, _ := fichaDe(t)
	ctx := context.Background()

	// 1 · antes de escuchar: no hay ficha.
	if _, err := f.Leer(); !errors.Is(err, ports.ErrSinFicha) {
		t.Fatalf("antes de escuchar no puede haber ficha: %v", err)
	}

	// 2 · el listener acepta.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	// 3 · recién ahora se publica.
	if err := f.Publicar(ctx, daemonFicha(srv.URL)); err != nil {
		t.Fatal(err)
	}
	leida, err := f.Leer()
	if err != nil {
		t.Fatalf("tras publicar la ficha tiene que leerse: %v", err)
	}
	if leida.Endpoint != srv.URL {
		t.Errorf("endpoint = %q, se esperaba %q", leida.Endpoint, srv.URL)
	}
	// ── CONTROL POSITIVO: el endpoint que la ficha nombra responde de verdad ──
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, leida.Endpoint+"/healthz", nil)
	resp, herr := http.DefaultClient.Do(req)
	if herr != nil {
		t.Fatalf("el endpoint de la ficha tiene que estar vivo: %v", herr)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("el endpoint de la ficha respondió %d", resp.StatusCode)
	}

	// 4 · shutdown ordenado: la ficha se va.
	if err := f.Retirar(); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Leer(); !errors.Is(err, ports.ErrSinFicha) {
		t.Fatalf("tras el shutdown no puede quedar ficha: %v", err)
	}
	// Retirar dos veces no es un error: un shutdown que corre dos veces no tiene que fallar.
	if err := f.Retirar(); err != nil {
		t.Errorf("retirar una ficha ya retirada no es un error: %v", err)
	}
}

// TestFichaPermisos0600 — el archivo lleva un token: `0600` el archivo, `0700` el directorio.
func TestFichaPermisos0600(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("los permisos POSIX no aplican en Windows; el confinamiento ahí lo da la ACL del %APPDATA% del usuario")
	}
	f, _ := fichaDe(t)
	if err := f.Publicar(context.Background(), daemonFicha("http://127.0.0.1:4200")); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(f.Ruta())
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Errorf("la ficha lleva un token: permisos %o, se esperaba 600", perm)
	}
	di, err := os.Stat(filepath.Dir(f.Ruta()))
	if err != nil {
		t.Fatal(err)
	}
	if perm := di.Mode().Perm(); perm != 0o700 {
		t.Errorf("el directorio de la ficha: permisos %o, se esperaba 700", perm)
	}
}

// TestFichaSeRelePorInvocacion — el daemon se reinicia en OTRO puerto entre dos lecturas. La
// segunda lectura tiene que ver el puerto nuevo: cachearla apuntaría al viejo para siempre.
func TestFichaSeRelePorInvocacion(t *testing.T) {
	f, _ := fichaDe(t)
	ctx := context.Background()
	if err := f.Publicar(ctx, daemonFicha("http://127.0.0.1:4200")); err != nil {
		t.Fatal(err)
	}
	primera, err := f.Leer()
	if err != nil {
		t.Fatal(err)
	}
	// El daemon se reinicia en otro puerto.
	if err := f.Publicar(ctx, daemonFicha("http://127.0.0.1:4299")); err != nil {
		t.Fatal(err)
	}
	segunda, err := f.Leer()
	if err != nil {
		t.Fatal(err)
	}
	if primera.Endpoint == segunda.Endpoint {
		t.Fatal("la ficha se relee en cada invocación: una cacheada apuntaría al puerto viejo")
	}
	if segunda.Endpoint != "http://127.0.0.1:4299" {
		t.Errorf("la segunda lectura debe traer el puerto nuevo: %q", segunda.Endpoint)
	}
}

// TestFichaNoEscribibleDegradaHonesto — un directorio de config no escribible NO tumba el
// daemon. Publicar falla, el caller lo reporta en la salud, y S1 y el modo instrumentado
// siguen funcionando: van por env y por settings, no por ficha.
func TestFichaNoEscribibleDegradaHonesto(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("como root o en Windows, un directorio 0500 sigue siendo escribible")
	}
	dir := t.TempDir()
	confinado := filepath.Join(dir, "solo-lectura")
	if err := os.MkdirAll(confinado, 0o500); err != nil {
		t.Fatal(err)
	}
	f, err := New(filepath.Join(confinado, "arnesia", "daemon.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perr := f.Publicar(context.Background(), daemonFicha("http://127.0.0.1:4200")); perr == nil {
		t.Fatal("publicar en un directorio no escribible tiene que fallar, no fingir éxito")
	}
	// Y leer devuelve la señal de fail-open, no un pánico.
	if _, lerr := f.Leer(); !errors.Is(lerr, ports.ErrSinFicha) {
		t.Errorf("sin ficha, Leer devuelve la señal de fail-open: %v", lerr)
	}
}

// TestFichaIlegibleEsComoNoTenerla — una ficha corrupta no puede hacer que el hook explote:
// es lo mismo que no tener ficha, y el hook sale fail-open.
func TestFichaIlegibleEsComoNoTenerla(t *testing.T) {
	f, _ := fichaDe(t)
	if err := os.MkdirAll(filepath.Dir(f.Ruta()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f.Ruta(), []byte("{esto no es json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Leer(); !errors.Is(err, ports.ErrSinFicha) {
		t.Fatalf("una ficha ilegible es como no tenerla: %v", err)
	}
	// Y una ficha sin endpoint tampoco sirve: nombra un daemon que no se puede alcanzar.
	sin, _ := json.Marshal(domain.FichaDaemon{Version: 1})
	if err := os.WriteFile(f.Ruta(), sin, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Leer(); !errors.Is(err, ports.ErrSinFicha) {
		t.Fatalf("una ficha sin endpoint es como no tenerla: %v", err)
	}
	// ── control positivo: una ficha buena SÍ se lee ──
	if err := f.Publicar(context.Background(), daemonFicha("http://127.0.0.1:4200")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Leer(); err != nil {
		t.Fatalf("control positivo: una ficha bien formada tiene que leerse: %v", err)
	}
}

// TestFichaEsAtomica — la escritura usa temp + rename en el MISMO directorio. Un lector que
// mire a mitad de la escritura ve la ficha vieja entera o la nueva entera, nunca media.
func TestFichaEsAtomica(t *testing.T) {
	f, _ := fichaDe(t)
	ctx := context.Background()
	if err := f.Publicar(ctx, daemonFicha("http://127.0.0.1:4200")); err != nil {
		t.Fatal(err)
	}
	if err := f.Publicar(ctx, daemonFicha("http://127.0.0.1:4299")); err != nil {
		t.Fatal(err)
	}
	// No quedan temporales tirados en el directorio.
	entradas, err := os.ReadDir(filepath.Dir(f.Ruta()))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entradas {
		if e.Name() != "daemon.json" {
			t.Errorf("quedó un archivo intermedio en el directorio de la ficha: %q", e.Name())
		}
	}
}
