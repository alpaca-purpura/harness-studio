package claudecode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// TestSpawnInyectaTelemetria — el contrato del spawn de S1, **assertado por nombre y valor**.
//
// 🔴 `OTEL_LOGS_EXPORTER=otlp` está acá porque se MIDIÓ (ANEXO H10.4): sin ella llegan **0**
// log events contra 2 del control positivo, misma corrida y mismo receptor. Como el canal
// primario es `/v1/logs` —por donde viaja el dinero— omitirla apaga la señal entera **sin un
// solo error visible**. Este test es lo que impide que alguien la saque «porque parece
// redundante».
func TestSpawnInyectaTelemetria(t *testing.T) {
	env := SpawnEnv(ports.SpawnOpts{}, "TOKEN-INGESTA", "http://127.0.0.1:4200",
		AtribucionSpawn{ArnesID: "vitalia", InstalacionID: "home-local", CajaID: "paso-3", CorridaID: "run-1"})

	tiene := func(k string) string {
		for _, e := range env {
			if strings.HasPrefix(e, k+"=") {
				return strings.TrimPrefix(e, k+"=")
			}
		}
		return ""
	}
	obligatorias := map[string]string{
		"CLAUDE_CODE_ENABLE_TELEMETRY": "1",
		// La medida. No se toca sin volver a medir.
		"OTEL_LOGS_EXPORTER":    "otlp",
		"OTEL_METRICS_EXPORTER": "otlp",
		// http/json y NO http/protobuf: el default del runtime es gRPC en otro puerto.
		"OTEL_EXPORTER_OTLP_PROTOCOL": "http/json",
		"OTEL_EXPORTER_OTLP_ENDPOINT": "http://127.0.0.1:4200",
		"OTEL_METRIC_EXPORT_INTERVAL": "10000",
		"OTEL_LOGS_EXPORT_INTERVAL":   "5000",
	}
	for k, v := range obligatorias {
		got := tiene(k)
		if got == "" {
			t.Errorf("falta %s: sin ella la señal se apaga en silencio", k)
			continue
		}
		if got != v {
			t.Errorf("%s = %q, se esperaba %q", k, got, v)
		}
	}
	// El endpoint va SIN `/v1/...`: el exportador concatena la ruta por spec. Ponérsela
	// haría que el runtime postee a `/v1/logs/v1/logs`.
	if ep := tiene("OTEL_EXPORTER_OTLP_ENDPOINT"); strings.Contains(ep, "/v1/") {
		t.Errorf("el endpoint no lleva la ruta: %q", ep)
	}
	if !strings.Contains(tiene("OTEL_EXPORTER_OTLP_ENDPOINT"), "127.0.0.1") {
		t.Error("el endpoint tiene que ser loopback: la telemetría no sale de la máquina")
	}
	// El vector de atribución, con las cuatro etiquetas.
	ra := tiene("OTEL_RESOURCE_ATTRIBUTES")
	for _, k := range []string{"arnesia.arnes=vitalia", "arnesia.instalacion=home-local",
		"arnesia.caja=paso-3", "arnesia.corrida=run-1"} {
		if !strings.Contains(ra, k) {
			t.Errorf("falta %q en OTEL_RESOURCE_ATTRIBUTES: %q", k, ra)
		}
	}
	// El token va por header, nunca por query string.
	if h := tiene("OTEL_EXPORTER_OTLP_HEADERS"); !strings.Contains(h, "x-arnesia-token=TOKEN-INGESTA") {
		t.Errorf("el token va por header: %q", h)
	}
}

// TestSpawnNoFiltraElTokenDeAPI — el env lleva el token de INGESTA y **jamás** el de la API.
// Filtrar el primero concede «escribime telemetría»; el segundo concede «conducí un agente
// con acceso al filesystem».
func TestSpawnNoFiltraElTokenDeAPI(t *testing.T) {
	t.Setenv("ARNESIA_AUTH_TOKEN", "TOKEN-DE-LA-API-QUE-NO-DEBE-SALIR")
	env := SpawnEnv(ports.SpawnOpts{}, "TOKEN-INGESTA", "http://127.0.0.1:4200", AtribucionSpawn{ArnesID: "a"})
	junto := strings.Join(env, "\n")
	if strings.Contains(junto, "TOKEN-DE-LA-API-QUE-NO-DEBE-SALIR") {
		t.Fatalf("el token de la API se filtró al subproceso:\n%s", junto)
	}
	if strings.Contains(junto, "ARNESIA_AUTH_TOKEN") {
		t.Errorf("ni siquiera el nombre de la variable:\n%s", junto)
	}
	// ── control positivo: el token de ingesta SÍ viaja ──
	if !strings.Contains(junto, "TOKEN-INGESTA") {
		t.Fatal("control positivo: el token de ingesta tiene que viajar; sin esto el assert de ausencia no prueba nada")
	}
}

// TestSpawnSinReceptorNoInstrumenta — sin endpoint no se inyecta nada. Apuntar a la nada solo
// agregaría latencia a cada request del agente.
func TestSpawnSinReceptorNoInstrumenta(t *testing.T) {
	if env := SpawnEnv(ports.SpawnOpts{}, "tok", "", AtribucionSpawn{ArnesID: "a"}); len(env) != 0 {
		t.Errorf("sin receptor no se instrumenta: %v", env)
	}
	// Control positivo: con endpoint sí.
	if env := SpawnEnv(ports.SpawnOpts{}, "tok", "http://127.0.0.1:4200", AtribucionSpawn{}); len(env) == 0 {
		t.Error("con endpoint tiene que instrumentar")
	}
}

// TestSpawnNoMandaEtiquetasVacias — una etiqueta vacía es ruido con forma de dato: llegaría al
// receptor como `arnesia.caja=` y la atribución diría que hay caja cuando no la hay.
func TestSpawnNoMandaEtiquetasVacias(t *testing.T) {
	// El caso del Dock: hay arnés, no hay caja ni corrida.
	env := SpawnEnv(ports.SpawnOpts{}, "tok", "http://127.0.0.1:4200",
		AtribucionSpawn{ArnesID: "vitalia"})
	var ra string
	for _, e := range env {
		if strings.HasPrefix(e, "OTEL_RESOURCE_ATTRIBUTES=") {
			ra = strings.TrimPrefix(e, "OTEL_RESOURCE_ATTRIBUTES=")
		}
	}
	if ra != "arnesia.arnes=vitalia" {
		t.Errorf("solo la etiqueta con dato: %q", ra)
	}
	if strings.Contains(ra, "caja=") || strings.Contains(ra, "corrida=") {
		t.Error("una etiqueta vacía haría creer que hay caja cuando no la hay")
	}
}

// TestResultTraeUsoDelTurno — el frame `result` REAL de la corrida del 2026-07-26. Los datos
// no son inventados: salen de `testdata/result-envelope.json`, copia byte por byte de la
// evidencia.
func TestResultTraeUsoDelTurno(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join("testdata", "result-envelope.json"))
	if err != nil {
		t.Fatal(err)
	}
	// El golden es el CUERPO del result; se envuelve en el frame como llega por stream-json.
	var cuerpo map[string]any
	if uerr := json.Unmarshal(crudo, &cuerpo); uerr != nil {
		t.Fatal(uerr)
	}
	cuerpo["type"] = "result"
	cuerpo["subtype"] = "success"
	cuerpo["model"] = "claude-haiku-4-5-20251001"
	linea, err := json.Marshal(cuerpo)
	if err != nil {
		t.Fatal(err)
	}

	var last *usage
	evs := translate(linea, &last)
	if len(evs) != 1 || evs[0].Kind != ports.EventResult {
		t.Fatalf("se esperaba un EventResult, hubo %d eventos", len(evs))
	}
	uso := evs[0].Uso
	if uso == nil {
		t.Fatal("el `result` trae uso y hasta hoy se tiraba: Uso no puede ser nil")
	}
	// Los valores MEDIDOS.
	if uso.Entrada != 10 || uso.Salida != 39 {
		t.Errorf("entrada/salida = %d/%d, se medía 10/39", uso.Entrada, uso.Salida)
	}
	if uso.CacheLectura != 17_536 {
		t.Errorf("cache lectura = %d, se medía 17536", uso.CacheLectura)
	}
	if uso.CacheEscritura != 8_257 {
		t.Errorf("cache escritura = %d, se medía 8257", uso.CacheEscritura)
	}
	if uso.ServiceTier != "standard" || uso.Speed != "standard" {
		t.Errorf("service_tier/speed = %q/%q", uso.ServiceTier, uso.Speed)
	}
	if uso.ModeloCanonico != "claude-haiku-4-5" || uso.Proveedor != "firstParty" {
		t.Errorf("modelo canónico/proveedor = %q/%q", uso.ModeloCanonico, uso.Proveedor)
	}
	if uso.ContextWindow != 200_000 {
		t.Errorf("ventana de contexto = %d", uso.ContextWindow)
	}
}

// TestSplitTTLLlegaDelStreamJSON — **la razón de ser de todo esto**: el split por vencimiento
// del cache NO viaja por telemetría, solo por este frame. Es lo que permite cotizar la
// escritura de cache al tramo correcto (asumir el barato subestima un 33 %, medido) y lo único
// que habilita al detector de re-warm.
func TestSplitTTLLlegaDelStreamJSON(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join("testdata", "result-envelope.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cuerpo map[string]any
	if uerr := json.Unmarshal(crudo, &cuerpo); uerr != nil {
		t.Fatal(uerr)
	}
	cuerpo["type"] = "result"
	linea, _ := json.Marshal(cuerpo)
	var last *usage
	evs := translate(linea, &last)
	uso := evs[0].Uso
	if uso == nil {
		t.Fatal("sin uso no hay split que verificar")
	}
	// ⚠️ El plan decía 4099; el archivo REAL dice **8257**. Manda el archivo.
	if uso.Ephemeral1h != 8_257 {
		t.Errorf("ephemeral_1h = %d, el golden mide 8257", uso.Ephemeral1h)
	}
	// **El 0 es un DATO, no una ausencia**: el runtime dijo que no se escribió nada al tramo
	// corto. Por eso el campo no es puntero y por eso este assert existe.
	if uso.Ephemeral5m != 0 {
		t.Errorf("ephemeral_5m = %d, el golden mide 0 — y ese 0 es un dato", uso.Ephemeral5m)
	}
	// Y los dos tramos suman el total: si no, el split estaría midiendo otra cosa.
	if uso.Ephemeral5m+uso.Ephemeral1h != uso.CacheEscritura {
		t.Errorf("el split no cierra con el total: %d + %d ≠ %d",
			uso.Ephemeral5m, uso.Ephemeral1h, uso.CacheEscritura)
	}
	// Control positivo del contraste: un `result` SIN el bloque de split no fabrica ceros.
	sinSplit := `{"type":"result","usage":{"input_tokens":5,"output_tokens":1}}`
	var l2 *usage
	evs2 := translate([]byte(sinSplit), &l2)
	if evs2[0].Uso == nil {
		t.Fatal("un result con uso pero sin split igual trae uso")
	}
	if evs2[0].Uso.CacheEscritura != 0 {
		t.Errorf("sin cache write el total es 0: %d", evs2[0].Uso.CacheEscritura)
	}
}

// TestResultSinUsoNoFabricaCeros — un `result` sin uso devuelve **nil**, no un uso en ceros
// que se leería como «este turno no consumió nada».
func TestResultSinUsoNoFabricaCeros(t *testing.T) {
	var last *usage
	evs := translate([]byte(`{"type":"result","subtype":"success","result":"listo"}`), &last)
	if len(evs) != 1 {
		t.Fatalf("se esperaba 1 evento, hubo %d", len(evs))
	}
	if evs[0].Uso != nil {
		t.Errorf("sin uso reportado, Uso es nil — un uso en ceros diría que no consumió nada: %+v", evs[0].Uso)
	}
}
