package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	telcatalogo "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/descubrimiento"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	httpapi "github.com/alpacapurpura/arnesia/internal/adapters/transport/http"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// homeDePrueba fija el home y el config-dir del proceso a un temporal.
//
// 🔴 **Ningún test de este paquete puede tocar el `~/.arnesia` real del operador**: escribir
// eventos de prueba ahí contaminaría los totales que la propia feature muestra, que es
// exactamente el pecado que este módulo existe para no cometer. En Windows,
// os.UserHomeDir lee USERPROFILE (no HOME) y os.UserConfigDir lee APPDATA — sin
// redirigirlos, el test escribiría en el ~/.arnesia REAL (bug de aislamiento detectado
// en el port 2026-08-13).
func homeDePrueba(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, ".config"))
	t.Setenv("APPDATA", filepath.Join(dir, ".config"))
	t.Setenv("LOCALAPPDATA", filepath.Join(dir, ".local"))
	return dir
}

// TestTelemetriaCLIReusaElUsecase — el comando y la API leen el MISMO caso de uso sobre el
// MISMO almacén, y dan lo mismo campo a campo. Dos implementaciones de la misma consulta
// divergen el día que alguien arregla una sola.
func TestTelemetriaCLIReusaElUsecase(t *testing.T) {
	home := homeDePrueba(t)
	ctx := context.Background()

	// Se siembra la base por la vía normal del servicio.
	st, err := telstore.New(filepath.Join(home, ".arnesia", "telemetria.db"),
		telstore.Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(), perfilCLI(), time.Now)
	svc.SetRetencion(st, telstore.NewRollup(st, time.Hour), 90, 24)
	micros := int64(18473)
	if _, ierr := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin:            domain.LlaveJoin{SesionID: "s-1", TurnoID: "t-1", ArnesID: "vitalia", CajaID: "paso-3"},
		Emisor:               domain.EmisorOTLP,
		Runtime:              "claude-code",
		AdaptadorVersion:     "cc-otlp/1",
		TSRecibido:           time.Now().UTC(),
		TipoEvento:           domain.EventoAPIRequest,
		Escenario:            domain.EscenarioS2Instrumentado,
		Modelo:               "claude-haiku-4-5",
		CostoReportadoMicros: &micros,
	}}); ierr != nil {
		t.Fatal(ierr)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}
	_ = st.Close()

	// ── vía HTTP ──
	svcHTTP, cerrarHTTP, err := servicioTelemetriaCLI(90)
	if err != nil {
		t.Fatal(err)
	}
	defer cerrarHTTP()
	vacio := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := httpapi.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
		svcHTTP, nil, vacio, vacio, httpapi.AuthConfigFor("127.0.0.1:4200", ""))
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:4200/api/telemetria/resumen", nil)
	req.Host = "127.0.0.1:4200"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /api/telemetria/resumen dio %d: %s", w.Code, w.Body.String())
	}
	var porHTTP domain.ResumenTelemetria
	if derr := json.Unmarshal(w.Body.Bytes(), &porHTTP); derr != nil {
		t.Fatal(derr)
	}

	// ── vía CLI, mismo servicio ──
	porCLI, err := svcHTTP.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}

	if porHTTP.CostoReportadoMicros == nil || porCLI.CostoReportadoMicros == nil {
		t.Fatalf("las dos vías tienen que traer el costo: http=%v cli=%v",
			porHTTP.CostoReportadoMicros, porCLI.CostoReportadoMicros)
	}
	if *porHTTP.CostoReportadoMicros != *porCLI.CostoReportadoMicros {
		t.Errorf("costo: http=%d cli=%d", *porHTTP.CostoReportadoMicros, *porCLI.CostoReportadoMicros)
	}
	if porHTTP.Turnos != porCLI.Turnos || porHTTP.Sesiones != porCLI.Sesiones {
		t.Errorf("conteos distintos: http=%+v cli=%+v", porHTTP.Cobertura, porCLI.Cobertura)
	}
	if porHTTP.Escenario != porCLI.Escenario {
		t.Errorf("escenario: http=%q cli=%q", porHTTP.Escenario, porCLI.Escenario)
	}
	// Control positivo: el número no es cero por accidente.
	if *porCLI.CostoReportadoMicros != 18473 {
		t.Errorf("el costo sembrado tiene que llegar entero: %d", *porCLI.CostoReportadoMicros)
	}
}

// TestCLICatalogoNoAbreLaBase — `catalogo` lee solo el binario. No puede crear un `.db` en el
// HOME de quien lo corra solo por preguntar con qué precios se costea.
func TestCLICatalogoNoAbreLaBase(t *testing.T) {
	home := homeDePrueba(t)
	if err := runTelemetria([]string{"catalogo"}); err != nil {
		t.Fatalf("catalogo: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".arnesia", "telemetria.db")); err == nil {
		t.Error("`arnesia telemetria catalogo` no puede crear la base: solo lee el binario")
	}
}

// ── el hook (T20) ────────────────────────────────────────────────────────────────────────

// TestHookStdoutVacio — la salida estándar queda VACÍA. Un hook que imprime le inyecta texto
// al contexto del agente.
func TestHookStdoutVacio(t *testing.T) {
	homeDePrueba(t)
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	hookProceso(strings.NewReader(`{"session_id":"s-1","prompt_id":"t-1","hook_event_name":"Stop","last_assistant_message":"CONTENIDO"}`))
	_ = w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if buf.Len() != 0 {
		t.Errorf("stdout del hook tiene que quedar vacía, escribió %d bytes: %q", buf.Len(), buf.String())
	}
}

// TestHookSaleCeroEnTodasSusRamas — las CUATRO ramas de error del contrato. `runHook` no
// tiene una sola por la que se pueda salir mal: un hook que termina mal al empezar un turno
// bloquea el turno del usuario.
func TestHookSaleCeroEnTodasSusRamas(t *testing.T) {
	homeDePrueba(t)
	casos := map[string]string{
		"stdin vacío":             "",
		"JSON roto":               "{esto no es json",
		"evento fuera de alcance": `{"session_id":"s","hook_event_name":"PreToolUse"}`,
		"sin session_id":          `{"hook_event_name":"Stop"}`,
	}
	for nombre, entrada := range casos {
		if err := runHook([]string{"proceso"}); err != nil {
			t.Fatalf("%s: runHook nunca devuelve error: %v", nombre, err)
		}
		hookProceso(strings.NewReader(entrada)) // no puede entrar en pánico
	}
	// Un subcomando equivocado tampoco rompe el turno.
	if err := runHook([]string{"inventado"}); err != nil {
		t.Errorf("ni un subcomando equivocado justifica romper un turno: %v", err)
	}
	if err := runHook(nil); err != nil {
		t.Errorf("sin subcomando tampoco: %v", err)
	}
}

// TestHookFailOpenSinDaemon — **el arquetipo del control positivo** (§4). Dos marcadores
// DISTINTOS, un servidor de puerto EFÍMERO, y el assert `len(conA)==0 && len(conB)==1`.
func TestHookFailOpenSinDaemon(t *testing.T) {
	homeDePrueba(t)
	var recibidos []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ev domain.EventoTelemetria
		_ = json.NewDecoder(r.Body).Decode(&ev)
		recibidos = append(recibidos, ev.SesionID)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	// ── variante A · SIN daemon publicado ──
	inicio := time.Now()
	hookProceso(strings.NewReader(`{"session_id":"MARCADOR-A-SIN-DAEMON","prompt_id":"p","hook_event_name":"Stop"}`))
	sinDaemon := time.Since(inicio)
	if sinDaemon > topeHook {
		t.Errorf("sin daemon el hook no puede tardar: %v (tope %v)", sinDaemon, topeHook)
	}

	// ── variante B · CON daemon, en puerto efímero ──
	f, err := descubrimiento.New("")
	if err != nil {
		t.Fatal(err)
	}
	d := domain.FichaDaemon{Version: 1, Endpoint: srv.URL, RutaProceso: "/", Desde: time.Now()}
	if perr := f.Publicar(context.Background(), d); perr != nil {
		t.Fatal(perr)
	}
	hookProceso(strings.NewReader(`{"session_id":"MARCADOR-B-CON-DAEMON","prompt_id":"p","hook_event_name":"Stop"}`))

	// ── el assert canónico ──
	var conA, conB int
	for _, s := range recibidos {
		switch s {
		case "MARCADOR-A-SIN-DAEMON":
			conA++
		case "MARCADOR-B-CON-DAEMON":
			conB++
		}
	}
	if conA != 0 || conB != 1 {
		t.Fatalf("conA=%d (debe ser 0) conB=%d (debe ser 1) — recibidos: %v", conA, conB, recibidos)
	}
}

// TestHookFichaHuerfana — el daemon murió sin retirar la ficha. El hook falla rápido dentro de
// su presupuesto y sale sin ruido; no se detecta con el pid (que además se puede reusar).
func TestHookFichaHuerfana(t *testing.T) {
	homeDePrueba(t)
	// Un servidor que se levanta y se baja: la ficha queda apuntando a un puerto muerto.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	f, err := descubrimiento.New("")
	if err != nil {
		t.Fatal(err)
	}
	if perr := f.Publicar(context.Background(), domain.FichaDaemon{
		Version: 1, Endpoint: url, RutaProceso: "/", Desde: time.Now(),
	}); perr != nil {
		t.Fatal(perr)
	}
	inicio := time.Now()
	hookProceso(strings.NewReader(`{"session_id":"s-huerfana","prompt_id":"p","hook_event_name":"Stop"}`))
	if d := time.Since(inicio); d > topeHook*2 {
		t.Errorf("con ficha huérfana el hook falla RÁPIDO: %v", d)
	}
}

// TestHookNoTardaNiFalla — el contrato completo, contra el BINARIO REAL como subproceso: exit
// 0, stdout vacía y menos del presupuesto, con contenido en la entrada.
func TestHookNoTardaNiFalla(t *testing.T) {
	if testing.Short() {
		t.Skip("compila el binario; se salta en -short")
	}
	// ⚠️ El BUILD usa el entorno real (caché de módulos incluida) y solo el SUBPROCESO
	// recibe el HOME de prueba. Al revés, `go build` re-descargaría el módulo entero bajo un
	// HOME falso — mismo gotcha que ya documenta TestDaemonServableHeadless.
	bin := filepath.Join(t.TempDir(), "arnesia")
	if runtime.GOOS == "windows" {
		bin += ".exe" // exec exige extensión ejecutable en Windows
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := homeDePrueba(t)
	const marca = "SECRETO-QUE-NO-DEBE-SALIR-7f21"
	cmd := exec.Command(bin, "hook", "proceso")
	cmd.Env = append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+filepath.Join(home, ".config"))
	cmd.Stdin = strings.NewReader(
		`{"session_id":"s-bin","prompt_id":"p","hook_event_name":"Stop","last_assistant_message":"` + marca + `"}`)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	inicio := time.Now()
	err := cmd.Run()
	transcurrido := time.Since(inicio)

	if err != nil {
		t.Fatalf("el hook tiene que salir 0 SIEMPRE: %v (stderr: %s)", err, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout tiene que quedar vacía: %q", stdout.String())
	}
	// El presupuesto es de PARED del hook (250 ms); el arranque del proceso suma, así que se
	// verifica un techo generoso. Lo que se prueba es que no cuelga.
	if transcurrido > 3*time.Second {
		t.Errorf("el hook colgó: %v", transcurrido)
	}
	if strings.Contains(stdout.String()+stderr.String(), marca) {
		t.Error("el contenido de la conversación no puede salir por ninguna salida del hook")
	}
}
