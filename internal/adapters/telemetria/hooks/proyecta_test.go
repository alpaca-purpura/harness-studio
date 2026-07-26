package hooks

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

var reloj = time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)

// payloadsReales son los SEIS eventos con la forma que el ANEXO H3 midió, incluidos los
// campos de contenido en claro que el runtime manda de verdad. Los marcadores son buscables
// a propósito: sin un valor que se pueda buscar, «no aparece» no se puede asertar.
const (
	marcaPrompt     = "SECRETO-PROMPT-9f2a"
	marcaRespuesta  = "SECRETO-RESPUESTA-4c81"
	marcaToolResp   = "SECRETO-TOOLRESP-77de"
	marcaToolInput  = "SECRETO-TOOLINPUT-1b60"
	marcaRuta       = "/home/usuario-real/Proyectos/vitalia"
	marcaTranscript = "/home/usuario-real/.claude/projects/-home-usuario-real-Proyectos-vitalia"
)

func payloadsReales() map[string]string {
	return map[string]string{
		"SessionStart": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"SessionStart","source":"startup"}`,
		"UserPromptSubmit": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"UserPromptSubmit","prompt_id":"t-1","permission_mode":"acceptEdits",
			"prompt":"` + marcaPrompt + `"}`,
		"PreToolUse": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"PreToolUse","prompt_id":"t-1","permission_mode":"acceptEdits",
			"tool_name":"Read","tool_input":{"file_path":"` + marcaToolInput + `"},"tool_use_id":"toolu_1"}`,
		"PostToolUse": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"PostToolUse","prompt_id":"t-1","permission_mode":"acceptEdits",
			"tool_name":"Read","tool_input":{"file_path":"` + marcaToolInput + `"},
			"tool_response":"` + marcaToolResp + `","duration_ms":7,"tool_use_id":"toolu_1"}`,
		"Stop": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"Stop","prompt_id":"t-1","permission_mode":"acceptEdits","stop_hook_active":false,
			"last_assistant_message":"` + marcaRespuesta + `","background_tasks":[],"session_crons":[]}`,
		"SessionEnd": `{"session_id":"s-1","transcript_path":"` + marcaTranscript + `","cwd":"` + marcaRuta + `",
			"hook_event_name":"SessionEnd","prompt_id":"t-1","reason":"clear"}`,
	}
}

// TestHookNoReenviaContenido — RF-282, camino del hook. Los SEIS payloads reales; ninguno de
// los marcadores de contenido puede sobrevivir a la proyección.
//
// ⚠ Control positivo en la misma corrida: `session_id` y `prompt_id` SÍ están. Sin eso, una
// proyección que devolviera un evento vacío pasaría todos los asserts de ausencia.
func TestHookNoReenviaContenido(t *testing.T) {
	marcas := []string{marcaPrompt, marcaRespuesta, marcaToolResp, marcaToolInput, marcaRuta, marcaTranscript}
	proyectados := 0

	for nombre, raw := range payloadsReales() {
		ev, err := Proyectar([]byte(raw), nil, reloj)
		if errors.Is(err, ErrHookFueraDeAlcance) {
			continue // PreToolUse y SessionStart: fuera del MVP a propósito.
		}
		if err != nil {
			t.Fatalf("%s: %v", nombre, err)
		}
		proyectados++
		serializado, merr := json.Marshal(ev)
		if merr != nil {
			t.Fatal(merr)
		}
		for _, m := range marcas {
			if strings.Contains(string(serializado), m) {
				t.Errorf("%s: el contenido sobrevivió a la proyección — %q está en %s", nombre, m, serializado)
			}
		}
		// ── control positivo, misma corrida ──
		if ev.SesionID != "s-1" {
			t.Errorf("%s: control positivo — el session_id declarado NO llegó (%q)", nombre, ev.SesionID)
		}
		if nombre != "SessionEnd" && ev.TurnoID != "t-1" {
			t.Errorf("%s: control positivo — el prompt_id declarado NO llegó (%q)", nombre, ev.TurnoID)
		}
	}
	if proyectados != 4 {
		t.Fatalf("se esperaban 4 eventos proyectados (UserPromptSubmit·PostToolUse·Stop·SessionEnd), hubo %d", proyectados)
	}
}

// TestProyeccionPorEvento — cada `hook_event_name` mapea a su tipo, y los dos que el MVP no
// instrumenta se rechazan explícitamente (no se cuelan como «otro»).
func TestProyeccionPorEvento(t *testing.T) {
	esperado := map[string]domain.TipoEvento{
		"UserPromptSubmit": domain.EventoTurnoInicio,
		"PostToolUse":      domain.EventoHerramienta,
		"Stop":             domain.EventoTurnoFin,
		"SessionEnd":       domain.EventoSesionFin,
	}
	for nombre, raw := range payloadsReales() {
		ev, err := Proyectar([]byte(raw), nil, reloj)
		want, instrumentado := esperado[nombre]
		if !instrumentado {
			if !errors.Is(err, ErrHookFueraDeAlcance) {
				t.Errorf("%s no se instrumenta en el MVP: se esperaba ErrHookFueraDeAlcance, got %v", nombre, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", nombre, err)
		}
		if ev.TipoEvento != want {
			t.Errorf("%s → %q, se esperaba %q", nombre, ev.TipoEvento, want)
		}
		if ev.Emisor != domain.EmisorHook {
			t.Errorf("%s: emisor=%q, se esperaba hook", nombre, ev.Emisor)
		}
		// El hook JAMÁS trae dinero (ANEXO H2): los buckets quedan nil, que es «no aplica».
		if ev.Tokens.Entrada != nil || ev.Tokens.Salida != nil || ev.CostoReportadoMicros != nil {
			t.Errorf("%s: el hook no trae dinero; los buckets deben quedar nil, no 0", nombre)
		}
	}
	// PostToolUse trae `duration_ms` medido — control positivo de que lo declarado entra.
	ev, err := Proyectar([]byte(payloadsReales()["PostToolUse"]), nil, reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev.DuracionMs == nil || *ev.DuracionMs != 7 {
		t.Errorf("PostToolUse.duration_ms debe llegar: %v", ev.DuracionMs)
	}
	if ev.Herramienta != "Read" {
		t.Errorf("PostToolUse.tool_name debe llegar: %q", ev.Herramienta)
	}
}

// TestCwdDesconocidoNoSeGuardaCrudo — A14. Un `cwd` que no resuelve a una instalación
// conocida se guarda como HUELLA, jamás como ruta.
func TestCwdDesconocidoNoSeGuardaCrudo(t *testing.T) {
	raw := payloadsReales()["Stop"]
	ev, err := Proyectar([]byte(raw), nil, reloj)
	if err != nil {
		t.Fatal(err)
	}
	serializado, _ := json.Marshal(ev)
	if strings.Contains(string(serializado), marcaRuta) {
		t.Errorf("la ruta cruda no puede persistirse: %s", serializado)
	}
	if strings.Contains(string(serializado), "usuario-real") {
		t.Errorf("ni siquiera un fragmento de la ruta: %s", serializado)
	}
	// Control positivo: la huella SÍ está, y es estable.
	if ev.CWDHuella == "" {
		t.Fatal("control positivo: sin huella no se pueden agrupar corridas del mismo lugar")
	}
	if ev.Atribucion != domain.ConfianzaSinDato {
		t.Errorf("un cwd no reconocido no atribuye nada: atribucion=%q", ev.Atribucion)
	}
	// La huella normaliza: `/a/b` y `/a/b/` son el mismo lugar.
	if HuellaCWD("/a/b") != HuellaCWD("/a/b/") {
		t.Error("la huella debe normalizar la ruta; si no, la misma instalación cuenta doble")
	}
	if HuellaCWD("/a/b") == HuellaCWD("/a/c") {
		t.Error("dos lugares distintos no pueden compartir huella")
	}
	// Y no es reversible ni contiene la ruta.
	if strings.Contains(HuellaCWD(marcaRuta), "vitalia") {
		t.Error("la huella no puede contener el nombre del proyecto")
	}
}

// TestCwdConocidoAtribuyePorProceso — el mismo `cwd`, con el Portafolio a mano, sí atribuye:
// da arnés + instalación (no la caja), con confianza `por-proceso`.
func TestCwdConocidoAtribuyePorProceso(t *testing.T) {
	resolver := func(cwd string) (string, string, bool) {
		if cwd == marcaRuta {
			return "vitalia", "home-local", true
		}
		return "", "", false
	}
	ev, err := Proyectar([]byte(payloadsReales()["Stop"]), resolver, reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev.ArnesID != "vitalia" || ev.InstalacionID != "home-local" {
		t.Errorf("el cwd conocido debe resolver: arnes=%q instalacion=%q", ev.ArnesID, ev.InstalacionID)
	}
	if ev.Atribucion != domain.ConfianzaPorProceso {
		t.Errorf("atribucion=%q, se esperaba por-proceso", ev.Atribucion)
	}
	// **Y la ruta igual se tira**: resolverla no es excusa para guardarla.
	serializado, _ := json.Marshal(ev)
	if strings.Contains(string(serializado), marcaRuta) {
		t.Errorf("la ruta cruda NO se guarda ni cuando resolvió: %s", serializado)
	}
	// La caja NO se atribuye por proceso: el cwd no la conoce. Inventarla sería adivinar.
	if ev.CajaID != "" {
		t.Errorf("por-proceso da arnés + instalación, jamás la caja: caja=%q", ev.CajaID)
	}
}

// TestHookPayloadInvalido — un stdin roto no produce un evento a medias.
func TestHookPayloadInvalido(t *testing.T) {
	for _, raw := range []string{``, `{`, `no soy json`, `[1,2,3]`} {
		if _, err := Proyectar([]byte(raw), nil, reloj); !errors.Is(err, domain.ErrPayloadInvalido) {
			t.Errorf("payload %q: se esperaba ErrPayloadInvalido, got %v", raw, err)
		}
	}
	// Sin session_id no es atribuible ni deduplicable.
	if _, err := Proyectar([]byte(`{"hook_event_name":"Stop"}`), nil, reloj); !errors.Is(err, domain.ErrEventoSinSesion) {
		t.Errorf("sin session_id se esperaba ErrEventoSinSesion, got %v", err)
	}
	// Control positivo: uno bien formado entra.
	if _, err := Proyectar([]byte(payloadsReales()["Stop"]), nil, reloj); err != nil {
		t.Fatalf("control positivo: un payload real debe proyectar (%v)", err)
	}
}

// TestLaAllowlistNoTieneCamposDeContenido — las dos listas no se pisan: ningún campo negado
// está en la allowlist. Suena obvio; es exactamente el error que se comete al agregar un
// campo apurado.
func TestLaAllowlistNoTieneCamposDeContenido(t *testing.T) {
	permitidos := map[string]bool{}
	for _, k := range AllowlistHook() {
		permitidos[k] = true
	}
	if len(permitidos) == 0 {
		t.Fatal("la allowlist está vacía: el test no compara nada")
	}
	for _, n := range NegadosHook() {
		if permitidos[n] {
			t.Errorf("%q está en la allowlist Y en la denylist — uno de los dos está mal", n)
		}
	}
	// Control positivo: los campos de la llave del join SÍ están permitidos.
	for _, k := range []string{"session_id", "prompt_id"} {
		if !permitidos[k] {
			t.Errorf("%q debe estar en la allowlist: es media llave del join", k)
		}
	}
}
