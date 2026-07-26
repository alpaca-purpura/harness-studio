package fitness

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	telcatalogo "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/hooks"
	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/otlp"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// telemetria_test.go son los tests de boundary del módulo de telemetría
// (arquitectura-modulo.md §10.5). Los invariantes que valen para el MÓDULO ENTERO viven acá;
// los que valen para un paquete viven colocados con él.
//
// 🔴 **Ningún test de este archivo toca el `~/.arnesia` real del operador.** Escribir eventos
// de prueba en la base real contaminaría los totales que la propia feature muestra — el pecado
// exacto que este módulo existe para no cometer.

// evidencia devuelve la ruta de los payloads MEDIDOS.
func evidencia(t *testing.T, nombre string) string {
	t.Helper()
	root := repoRoot()
	if root == "" {
		t.Fatal("repoRoot vacío")
	}
	return filepath.Join(root, "docs/product/stories/2026-07-24-telemetria-embebida-otel",
		"verificacion-2026-07-26/evidencia", nombre)
}

// TestAllowlistNoPersistePII — **el test central de RF-282**, y la única forma de asertar una
// ausencia sin confiar en la forma del código: se ingiere por la puerta real, se cierra el
// almacén, y se busca la PII **como SUBCADENA DEL ARCHIVO `.db`**.
//
// Los golden traen la PII redactada (son datos reales del operador), así que además se inyecta
// una fixture con PII SIMULADA y buscable — sin un valor que se pueda buscar, «no aparece» no
// se puede asertar.
//
// ⚠ Control positivo en la misma corrida: un marcador que **sí** tiene que estar en el
// archivo. Sin eso, un almacén que no escribió nada pasaría todos los asserts de ausencia.
func TestAllowlistNoPersistePII(t *testing.T) {
	const (
		marcaEmail = "PII-EMAIL-fitness@ejemplo.test"
		marcaCta   = "PII-CUENTA-fitness-9d02"
		marcaOrg   = "PII-ORG-fitness-4b8e"
		marcaUser  = "PII-USER-fitness-2c55"
		marcaUUID  = "PII-UUID-fitness-0f13"
		marcaOK    = "SESION-CONTROL-POSITIVO-fitness"
	)
	dir := t.TempDir()
	ruta := filepath.Join(dir, "telemetria.db")
	st, err := telstore.New(ruta, telstore.Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(), otlp.PerfilClaudeCode(), time.Now)
	receptor := otlp.NewReceptor(svc, otlp.Opciones{})
	srv := httptest.NewServer(receptor)
	defer srv.Close()

	// 1 · los payloads REALES, tal cual se midieron.
	for _, n := range []string{"logs-run1.json", "logs-run2-con-skill.json", "logs-run4-con-tools.json"} {
		crudo, rerr := os.ReadFile(evidencia(t, n))
		if rerr != nil {
			t.Fatal(rerr)
		}
		postearFitness(t, srv.URL+"/v1/logs", crudo)
	}
	// 2 · una fixture con PII SIMULADA y buscable, con la MISMA forma del wire.
	fixture := `{"resourceLogs":[{"resource":{"attributes":[
	  {"key":"user.email","value":{"stringValue":"` + marcaEmail + `"}},
	  {"key":"user.account_id","value":{"stringValue":"` + marcaCta + `"}},
	  {"key":"user.account_uuid","value":{"stringValue":"` + marcaUUID + `"}},
	  {"key":"user.id","value":{"stringValue":"` + marcaUser + `"}},
	  {"key":"organization.id","value":{"stringValue":"` + marcaOrg + `"}}]},
	  "scopeLogs":[{"scope":{"name":"com.anthropic.claude_code.events"},"logRecords":[{
	    "timeUnixNano":"1785048577058000000","attributes":[
	      {"key":"session.id","value":{"stringValue":"` + marcaOK + `"}},
	      {"key":"prompt.id","value":{"stringValue":"turno-1"}},
	      {"key":"event.name","value":{"stringValue":"api_request"}},
	      {"key":"model","value":{"stringValue":"claude-haiku-4-5"}},
	      {"key":"input_tokens","value":{"intValue":10}},
	      {"key":"cost_usd_micros","value":{"intValue":18473}}]}]}]}]}`
	postearFitness(t, srv.URL+"/v1/logs", []byte(fixture))

	if serr := st.Sincronizar(context.Background()); serr != nil {
		t.Fatal(serr)
	}
	if cerr := st.Close(); cerr != nil {
		t.Fatal(cerr)
	}

	// 3 · la búsqueda de SUBCADENA sobre el archivo. No sobre structs: sobre los BYTES.
	crudo, err := os.ReadFile(ruta)
	if err != nil {
		t.Fatal(err)
	}
	texto := string(crudo)
	for _, m := range []string{marcaEmail, marcaCta, marcaUUID, marcaUser, marcaOrg,
		"user.email", "account_uuid", "organization.id", "REDACTADO"} {
		if strings.Contains(texto, m) {
			t.Errorf("la identidad llegó al almacén: %q aparece como subcadena del .db", m)
		}
	}
	// ── control positivo, misma corrida ──
	if !strings.Contains(texto, marcaOK) {
		t.Fatalf("control positivo: el sesion_id declarado NO está en el .db — el almacén no escribió nada "+
			"y los asserts de ausencia no prueban nada (%d bytes)", len(crudo))
	}
	if !strings.Contains(texto, "claude-haiku-4-5") {
		t.Fatal("control positivo: el modelo declarado tiene que estar en el .db")
	}
	t.Logf("archivo auditado: %d bytes, 0 ocurrencias de identidad, control positivo presente", len(crudo))
}

func postearFitness(t *testing.T, url string, cuerpo []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, strings.NewReader(string(cuerpo)))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s dio %d", url, resp.StatusCode)
	}
}

// TestHookNoReenviaContenido — el otro camino de RF-282, con los payloads del ANEXO H3.
func TestHookNoReenviaContenido(t *testing.T) {
	const (
		marcaPrompt = "FITNESS-PROMPT-9f2a"
		marcaResp   = "FITNESS-RESPUESTA-4c81"
		marcaTool   = "FITNESS-TOOLRESP-77de"
		marcaRuta   = "/home/usuario-real/Proyectos/vitalia"
	)
	payloads := []string{
		`{"session_id":"s-1","prompt_id":"t-1","hook_event_name":"UserPromptSubmit","cwd":"` + marcaRuta + `","prompt":"` + marcaPrompt + `"}`,
		`{"session_id":"s-1","prompt_id":"t-1","hook_event_name":"Stop","cwd":"` + marcaRuta + `","last_assistant_message":"` + marcaResp + `"}`,
		`{"session_id":"s-1","prompt_id":"t-1","hook_event_name":"PostToolUse","cwd":"` + marcaRuta + `","tool_name":"Read","tool_response":"` + marcaTool + `","duration_ms":7}`,
		`{"session_id":"s-1","prompt_id":"t-1","hook_event_name":"SessionEnd","cwd":"` + marcaRuta + `","reason":"clear"}`,
	}
	proyectados := 0
	for _, p := range payloads {
		ev, err := hooks.Proyectar([]byte(p), nil, time.Now())
		if err != nil {
			t.Fatalf("proyectar %s: %v", p[:40], err)
		}
		proyectados++
		raw, _ := json.Marshal(ev)
		for _, m := range []string{marcaPrompt, marcaResp, marcaTool, marcaRuta, "usuario-real"} {
			if strings.Contains(string(raw), m) {
				t.Errorf("el hook reenvió %q: %s", m, raw)
			}
		}
		// ── control positivo ──
		if ev.SesionID != "s-1" || ev.TurnoID != "t-1" {
			t.Errorf("control positivo: los identificadores declarados tienen que llegar: %+v", ev.LlaveJoin)
		}
	}
	if proyectados != len(payloads) {
		t.Fatalf("se proyectaron %d de %d payloads", proyectados, len(payloads))
	}
}

// TestAllowlistEsListaNoSugerencia — **source-scan**: no existe ninguna ruta de código que
// copie un mapa entero de atributos al evento canónico. Es la única forma de que la allowlist
// sea una garantía y no una intención: si `domain.NuevoEvento(map[string]any)` existiera,
// tarde o temprano alguien la usaría.
func TestAllowlistEsListaNoSugerencia(t *testing.T) {
	root := repoRoot()
	if root == "" {
		t.Fatal("repoRoot vacío")
	}
	// El detector busca funciones exportadas del dominio que reciban un mapa: ese es el
	// constructor-por-mapa que no debe existir.
	fset := token.NewFileSet()
	base := filepath.Join(root, "internal", "domain")
	var hallazgos []string
	revisados := 0
	if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if !strings.Contains(filepath.Base(path), "telemetria") {
			return nil
		}
		revisados++
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				return true
			}
			for _, p := range fn.Type.Params.List {
				if _, esMapa := p.Type.(*ast.MapType); esMapa {
					hallazgos = append(hallazgos, filepath.Base(path)+":"+fn.Name.Name)
				}
			}
			return true
		})
		return nil
	}); werr != nil {
		t.Fatal(werr)
	}
	if revisados == 0 {
		t.Fatal("el scan no revisó ningún archivo del módulo: verde por vacío")
	}
	for _, h := range hallazgos {
		t.Errorf("%s recibe un mapa: un constructor-por-mapa reintroduce la copia en masa "+
			"y convierte la allowlist en una intención", h)
	}
	// ── control positivo: el detector encuentra lo que busca ──
	// Se le da un archivo sintético con exactamente el patrón prohibido.
	fset2 := token.NewFileSet()
	f2, err := parser.ParseFile(fset2, "sintetico.go",
		"package domain\nfunc NuevoEvento(campos map[string]any) EventoTelemetria { return EventoTelemetria{} }\n", 0)
	if err != nil {
		t.Fatal(err)
	}
	encontrado := false
	ast.Inspect(f2, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || !fn.Name.IsExported() {
			return true
		}
		for _, p := range fn.Type.Params.List {
			if _, esMapa := p.Type.(*ast.MapType); esMapa {
				encontrado = true
			}
		}
		return true
	})
	if !encontrado {
		t.Fatal("control positivo: el detector no encuentra un constructor-por-mapa evidente — está roto")
	}
}

// TestNoAplicaNoEsCeroEnElWire — boundary no-aplica-no-es-cero, a nivel de los DTO de salida.
func TestNoAplicaNoEsCeroEnElWire(t *testing.T) {
	// Un evento con buckets nil no puede serializar las claves.
	entrada := int64(10)
	ev := domain.EventoTelemetria{
		LlaveJoin: domain.LlaveJoin{SesionID: "s"},
		Tokens:    domain.Tokens{Entrada: &entrada},
	}
	raw, _ := json.Marshal(ev)
	for _, k := range []string{"razonamiento", "cache_escritura_1h", "cache_lectura"} {
		if strings.Contains(string(raw), k) {
			t.Errorf("un bucket nil no viaja al wire: %q está en %s", k, raw)
		}
	}
	if !strings.Contains(string(raw), `"entrada":10`) {
		t.Fatalf("control positivo: el bucket con dato tiene que viajar: %s", raw)
	}

	// Un `GastoCaja` no atribuible lleva `costo_micros: null` **explícito** (sin omitempty)
	// y motivo no vacío.
	g := domain.GastoCaja{CajaID: "paso-3", Atribuible: false, Motivo: "sin dato atribuible"}
	raw2, _ := json.Marshal(g)
	if !strings.Contains(string(raw2), `"costo_micros":null`) {
		t.Errorf("una caja sin dato lleva costo_micros null EXPLÍCITO: %s", raw2)
	}
	if strings.Contains(string(raw2), `"costo_micros":0`) {
		t.Errorf("jamás 0: %s", raw2)
	}
	// Y la Cobertura sin denominador viaja con `esperados: null`, no 0.
	c := domain.Cobertura{Exacta: 3}
	raw3, _ := json.Marshal(c)
	if !strings.Contains(string(raw3), `"esperados":null`) {
		t.Errorf("sin denominador, esperados viaja null: %s", raw3)
	}
}

// TestCifraLlevaConfianza — boundary cifra-viaja-con-su-confianza, por REFLEXIÓN: todo DTO
// que lleva un campo de dinero tiene que llevar un campo de confianza en la misma struct.
// Si no, la cifra podría viajar sola hasta la pantalla.
func TestCifraLlevaConfianza(t *testing.T) {
	tipos := []any{
		domain.EventoTelemetria{}, domain.ResumenTelemetria{}, domain.GastoCaja{},
		domain.TurnoUnido{}, domain.DetalleCaja{}, domain.FilaPortafolio{},
		domain.PuntoDeMejora{},
	}
	revisados := 0
	for _, x := range tipos {
		tp := reflect.TypeOf(x)
		conDinero, conConfianza := false, false
		for i := 0; i < tp.NumField(); i++ {
			n := strings.ToLower(tp.Field(i).Name)
			if strings.Contains(n, "costo") || strings.Contains(n, "micros") || strings.Contains(n, "gasto") {
				conDinero = true
			}
			if strings.Contains(n, "confianza") || strings.Contains(n, "atribucion") ||
				strings.Contains(n, "paridad") || strings.Contains(n, "cobertura") {
				conConfianza = true
			}
		}
		if !conDinero {
			continue
		}
		revisados++
		if !conConfianza {
			t.Errorf("%s lleva dinero y no lleva confianza: la cifra podría llegar sola a la pantalla", tp.Name())
		}
	}
	if revisados == 0 {
		t.Fatal("ningún tipo con dinero revisado: verde por vacío")
	}
	t.Logf("tipos con dinero verificados: %d", revisados)
}

// TestTelemetriaDBNoEsElIndice — A2 por source-scan: el almacén de telemetría nunca resuelve
// a `index.db`. Compartir archivo haría que un bump del índice —que es DESECHABLE por
// doctrina— borre una historia que no se puede reconstruir de ningún lado.
func TestTelemetriaDBNoEsElIndice(t *testing.T) {
	root := repoRoot()
	if root == "" {
		t.Fatal("repoRoot vacío")
	}
	base := filepath.Join(root, "internal", "adapters", "telemetria", "store")
	revisados := 0
	if werr := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		revisados++
		// Se miran los LITERALES DE CADENA, no el archivo crudo: los comentarios de este
		// módulo explican justamente por qué la base NO es `index.db`, y un scan textual
		// los cazaría a ellos en vez de cazar código.
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return perr
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			if strings.Contains(lit.Value, "index.db") {
				t.Errorf("%s tiene un literal que nombra index.db: la telemetría vive en su "+
					"propio archivo (A2) — compartirlo haría que un bump del índice borre "+
					"una historia irrecuperable", filepath.Base(path))
			}
			return true
		})
		return nil
	}); werr != nil {
		t.Fatal(werr)
	}
	if revisados == 0 {
		t.Fatal("el scan no revisó ningún archivo del almacén: verde por vacío")
	}
	// ── control positivo del detector: encuentra un literal prohibido evidente ──
	fset2 := token.NewFileSet()
	fuenteSintetica := "package store\nconst ruta = `~/.arnesia/index.db`\n"
	f2, perr := parser.ParseFile(fset2, "sintetico.go", fuenteSintetica, 0)
	if perr != nil {
		t.Fatal(perr)
	}
	hallado := false
	ast.Inspect(f2, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING &&
			strings.Contains(lit.Value, "index.db") {
			hallado = true
		}
		return true
	})
	if !hallado {
		t.Fatal("control positivo: el detector no encuentra un literal `index.db` evidente — está roto")
	}

	// Control positivo funcional: un store real resuelve a `telemetria.db`, no a `index.db`.
	home := t.TempDir()
	t.Setenv("HOME", home)
	st, err := telstore.New("", telstore.Opciones{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	if !strings.HasSuffix(st.Ruta(), "telemetria.db") {
		t.Errorf("la ruta por default es telemetria.db, es %q", st.Ruta())
	}
}

// TestS2DegradadoApagaLosDetectoresDeDinero — a nivel de MÓDULO, no solo de detector: con
// solo eventos de hook, el resumen trae los costos en **null** y B4/B6/B3 salen en
// `no_aplican` con motivo. Nunca en 0.
func TestS2DegradadoApagaLosDetectoresDeDinero(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin:  domain.LlaveJoin{SesionID: "s-1", TurnoID: "t-1", ArnesID: "vitalia"},
		Emisor:     domain.EmisorHook,
		Runtime:    "claude-code",
		TSRecibido: time.Now().UTC(),
		TipoEvento: domain.EventoTurnoFin,
		Escenario:  domain.EscenarioS2Degradado,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros != nil || r.CostoCalculadoMicros != nil {
		t.Errorf("sin señal de dinero los costos viajan null: %v / %v",
			r.CostoReportadoMicros, r.CostoCalculadoMicros)
	}
	mej, err := svc.Mejoras(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	apagados := map[domain.DetectorID]string{}
	for _, na := range mej.NoAplican {
		apagados[na.Detector] = na.Motivo
	}
	for _, d := range []domain.DetectorID{domain.DetB4, domain.DetB6, domain.DetB3} {
		motivo, ok := apagados[d]
		if !ok {
			t.Errorf("%s tiene que salir en no_aplican", d)
			continue
		}
		if motivo == "" {
			t.Errorf("%s se apagó sin motivo", d)
		}
	}
}

// TestS2InstrumentadoTieneDineroYNoTieneSplit — con `api_request` y sin cierre de turno del
// subproceso, B4/B6/B3 aplican y B1 no — **con el motivo del split**, no con «corrió fuera».
func TestS2InstrumentadoTieneDineroYNoTieneSplit(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	micros := int64(18473)
	var evs []domain.EventoTelemetria
	for i, modelo := range []string{"claude-haiku-4-5", "claude-sonnet-4-5"} {
		evs = append(evs, domain.EventoTelemetria{
			LlaveJoin:  domain.LlaveJoin{SesionID: "s-1", TurnoID: "t-" + string(rune('a'+i)), ArnesID: "vitalia"},
			Emisor:     domain.EmisorOTLP,
			Runtime:    "claude-code",
			TSRecibido: time.Now().UTC(),
			TipoEvento: domain.EventoAPIRequest,
			Escenario:  domain.EscenarioS2Instrumentado,
			Modelo:     modelo,
			// El split NO viene por OTel: `CacheEscritura1h` queda nil.
			Tokens:               domain.Tokens{Entrada: i64f(10), CacheEscritura5m: i64f(8257)},
			CostoReportadoMicros: &micros,
		})
	}
	if _, err := svc.Ingerir(ctx, evs); err != nil {
		t.Fatal(err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}
	mej, err := svc.Mejoras(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if mej.Escenario != domain.EscenarioS2Instrumentado {
		t.Fatalf("escenario = %q, se esperaba s2-instrumentado", mej.Escenario)
	}
	noAplican := map[domain.DetectorID]string{}
	for _, na := range mej.NoAplican {
		if !na.Aplica {
			noAplican[na.Detector] = na.Motivo
		}
	}
	for _, d := range []domain.DetectorID{domain.DetB4, domain.DetB6, domain.DetB3} {
		if m, ok := noAplican[d]; ok {
			t.Errorf("%s tiene que APLICAR con señal de dinero; se apagó con: %q", d, m)
		}
	}
	motivoB1, ok := noAplican[domain.DetB1]
	if !ok {
		t.Fatal("B1 no puede aplicar sin el desglose por vencimiento")
	}
	if strings.Contains(strings.ToLower(motivoB1), "fuera de arnesia") {
		t.Errorf("el motivo de B1 confunde los dos casos: %q", motivoB1)
	}
}

// TestToolResultBytesSePersiste — A21: los dos tamaños se guardan y **ningún campo de
// contenido** entra con ellos.
func TestToolResultBytesSePersiste(t *testing.T) {
	crudo, err := os.ReadFile(evidencia(t, "logs-run4-con-tools.json"))
	if err != nil {
		t.Fatal(err)
	}
	regs, err := otlp.DecodificarLogs(crudo)
	if err != nil {
		t.Fatal(err)
	}
	visto := false
	for _, r := range regs {
		if r.Attrs.Texto("event.name") != "tool_result" {
			continue
		}
		ev, _, merr := otlp.MapearLogRecord(r, otlp.PerfilClaudeCode(), time.Now())
		if merr != nil {
			t.Fatal(merr)
		}
		visto = true
		if ev.ToolInputBytes == nil || ev.ToolResultBytes == nil {
			t.Fatalf("los dos tamaños se persisten desde el día 1 (A21): %v / %v",
				ev.ToolInputBytes, ev.ToolResultBytes)
		}
		raw, _ := json.Marshal(ev)
		for _, prohibido := range []string{"tool_response", "contenido", "REDACTED"} {
			if strings.Contains(string(raw), prohibido) {
				t.Errorf("son BYTES, no contenido: %q está en %s", prohibido, raw)
			}
		}
	}
	if !visto {
		t.Fatal("el golden no produjo ningún tool_result: el test no verificó nada")
	}
}

// TestEscenarioSeDerivaDeLaSenal — boundary telemetria-de-nacimiento: **un arnés no puede
// declarar su propio nivel de instrumentación**. Los tres lotes traen un `escenario`
// MENTIROSO y el módulo lo ignora.
func TestEscenarioSeDerivaDeLaSenal(t *testing.T) {
	casos := []struct {
		nombre   string
		ev       domain.EventoTelemetria
		esperado domain.Escenario
	}{
		{"con corrida nuestra", domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{SesionID: "s", TurnoID: "t", ArnesID: "a", CorridaID: "run-1"},
			Emisor:    domain.EmisorOTLP, TipoEvento: domain.EventoAPIRequest,
			Escenario: domain.EscenarioS2Degradado, // mentira
		}, domain.EscenarioS1},
		{"dinero sin corrida", domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{SesionID: "s", TurnoID: "t", ArnesID: "a"},
			Emisor:    domain.EmisorOTLP, TipoEvento: domain.EventoAPIRequest,
			Escenario: domain.EscenarioS1, // mentira
		}, domain.EscenarioS2Instrumentado},
		{"solo hook", domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{SesionID: "s", TurnoID: "t", ArnesID: "a"},
			Emisor:    domain.EmisorHook, TipoEvento: domain.EventoTurnoFin,
			Escenario: domain.EscenarioS1, // mentira
		}, domain.EscenarioS2Degradado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, st := servicioFitness(t)
			ctx := context.Background()
			c.ev.Runtime = "claude-code"
			c.ev.TSRecibido = time.Now().UTC()
			if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{c.ev}); err != nil {
				t.Fatal(err)
			}
			if err := st.Sincronizar(ctx); err != nil {
				t.Fatal(err)
			}
			r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
			if err != nil {
				t.Fatal(err)
			}
			if r.Escenario != c.esperado {
				t.Errorf("escenario = %q, se esperaba %q — el emisor NO puede elegirlo",
					r.Escenario, c.esperado)
			}
		})
	}
}

// TestNoAplicaSobreviveAlRollup — el NULL sobrevive a la agregación, verificado por el módulo
// entero: se ingiere por el servicio y se lee por el agregado.
func TestNoAplicaSobreviveAlRollup(t *testing.T) {
	dir := t.TempDir()
	st, err := telstore.New(filepath.Join(dir, "telemetria.db"),
		telstore.Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	roll := telstore.NewRollup(st, time.Hour)
	defer roll.Detener()
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(), otlp.PerfilClaudeCode(), time.Now)
	ctx := context.Background()
	if _, ierr := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin: domain.LlaveJoin{SesionID: "s", TurnoID: "t", ArnesID: "a"},
		Emisor:    domain.EmisorOTLP, Runtime: "claude-code", TipoEvento: domain.EventoAPIRequest,
		Escenario: domain.EscenarioS2Instrumentado, TSRecibido: time.Now().UTC(),
		Tokens: domain.Tokens{Entrada: i64f(10)},
	}}); ierr != nil {
		t.Fatal(ierr)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}
	if aerr := roll.Actualizar(ctx); aerr != nil {
		t.Fatal(aerr)
	}
	// La verificación fina vive colocada en el paquete del almacén; acá se asegura que el
	// camino COMPLETO (servicio → almacén → agregado) no rompe la propiedad.
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros != nil {
		t.Errorf("nadie reportó costo: el total viaja null, no %d", *r.CostoReportadoMicros)
	}
	if r.Turnos != 1 {
		t.Errorf("control positivo: el turno tiene que estar contado: %d", r.Turnos)
	}
}

// TestDetectorQueNoAplicaTraeMotivo — a nivel de módulo: los seis con contexto vacío.
func TestDetectorQueNoAplicaTraeMotivo(t *testing.T) {
	ds := domain.DetectoresMVP()
	if len(ds) != 6 {
		t.Fatalf("el MVP son 6 detectores, hay %d", len(ds))
	}
	for _, d := range ds {
		ap := d.Aplica(domain.ContextoDeteccion{})
		if ap.Aplica {
			t.Errorf("%s: con contexto vacío no puede aplicar", d.ID())
		}
		if strings.TrimSpace(ap.Motivo) == "" {
			t.Errorf("%s: Motivo es OBLIGATORIO cuando no aplica", d.ID())
		}
	}
	// Control positivo: con señal completa, los seis aplican.
	completo := domain.ContextoDeteccion{
		Runtime: "claude-code", TieneCosto: true, TieneSplitTTL: true, TieneSenalProceso: true,
		TieneGateHumano: true, TieneEventoRotacion: true, ModelosDistintos: 2,
	}
	for _, d := range ds {
		if !d.Aplica(completo).Aplica {
			t.Errorf("%s: con señal completa tiene que aplicar", d.ID())
		}
	}
}

// PresupuestoBaselineMB es el peso del daemon **sin el módulo de telemetría**, medido sobre un
// árbol limpio (`git archive HEAD`, sin `web/dist`) el 2026-07-26: **18,95 MB**.
//
// Está acá y no en un archivo aparte para que quien cambie el número tenga que tocar el test que
// lo usa, y para que el `git blame` diga cuándo y por qué se movió.
const PresupuestoBaselineMB = 18.95

// TestPresupuestoDeBinario verifica los DOS presupuestos de §11, no solo uno.
//
// 🔴 M8 de la auditoría: este test decía verificar «delta ≤ 1,5 MB contra el release anterior» y
// solo comparaba contra el techo absoluto de 25 MB. El presupuesto PRINCIPAL del boundary
// `peso-del-binario-es-presupuesto` no estaba enforced en ningún lado — que el módulo entrara
// era un accidente feliz, no un hecho verificado.
//
// Y el auditor encontró algo peor: **el resultado dependía del working tree.** Árbol limpio →
// 19,41 MB; con la SPA construida en `web/dist` → 23,38 MB, o sea el 93,5 % del techo. El mismo
// test daba dos respuestas muy distintas según qué hubiera compilado el desarrollador antes.
//
// Por eso ahora se compila desde un **`git archive HEAD`** en un directorio limpio: sin
// `web/dist`, sin artefactos locales, mismo resultado en cualquier máquina y en CI.
func TestPresupuestoDeBinario(t *testing.T) {
	if testing.Short() {
		t.Skip("compila el daemon desde un árbol limpio; se salta en -short y corre en CI")
	}
	root := repoRoot()
	if root == "" {
		t.Fatal("repoRoot vacío")
	}
	limpio := t.TempDir()
	// `git archive` da el árbol EXACTO del commit: sin `web/dist`, sin binarios locales, sin
	// nada que el desarrollador haya compilado antes. Es lo que hace al número reproducible.
	tar := exec.Command("git", "archive", "--format=tar", "HEAD")
	tar.Dir = root
	untar := exec.Command("tar", "-x", "-C", limpio)
	untar.Stdin, _ = tar.StdoutPipe()
	if err := untar.Start(); err != nil {
		t.Fatalf("desempaquetar el árbol limpio: %v", err)
	}
	if err := tar.Run(); err != nil {
		t.Skipf("no se pudo armar el árbol limpio (¿working tree sin commitear?): %v", err)
	}
	if err := untar.Wait(); err != nil {
		t.Fatalf("desempaquetar: %v", err)
	}

	bin := filepath.Join(t.TempDir(), "arnesia")
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/arnesia")
	cmd.Dir = limpio
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compilar desde el árbol limpio: %v\n%s", err, out)
	}
	fi, err := os.Stat(bin)
	if err != nil {
		t.Fatal(err)
	}
	mb := float64(fi.Size()) / (1 << 20)

	// 1 · techo ABSOLUTO.
	const topeAbsolutoMB = 25.0
	if mb > topeAbsolutoMB {
		t.Errorf("el daemon pesa %.2f MB, tope absoluto %.0f MB (§11)", mb, topeAbsolutoMB)
	}
	// 2 · **el presupuesto que este test decía verificar y no verificaba**: el delta.
	const topeDeltaMB = 1.5
	delta := mb - PresupuestoBaselineMB
	if delta > topeDeltaMB {
		t.Errorf("el daemon creció %.2f MB sobre la línea base de %.2f MB (tope +%.1f MB, §11). "+
			"Si el crecimiento es legítimo, mové PresupuestoBaselineMB **con la medición nueva "+
			"en el commit** — no bajes el tope", delta, PresupuestoBaselineMB, topeDeltaMB)
	}
	t.Logf("daemon (árbol limpio) = %.2f MB · delta sobre la base %.2f MB = %+.2f MB "+
		"(tope +%.1f) · %.0f %% del techo absoluto",
		mb, PresupuestoBaselineMB, delta, topeDeltaMB, mb/topeAbsolutoMB*100)
}

// servicioFitness arma servicio + almacén en un directorio TEMPORAL.
func servicioFitness(t *testing.T) (*usecase.TelemetriaService, *telstore.Store) {
	t.Helper()
	st, err := telstore.New(filepath.Join(t.TempDir(), "telemetria.db"),
		telstore.Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(), otlp.PerfilClaudeCode(), time.Now)
	svc.SetRetencion(st, telstore.NewRollup(st, time.Hour), 90, 24)
	return svc, st
}

func i64f(v int64) *int64 { return &v }

// compilarDaemon compila `cmd/arnesia` con el entorno REAL (caché de módulos incluida). Si se
// compilara bajo un HOME de prueba, `go build` re-descargaría el módulo entero — mismo gotcha
// que ya documenta TestDaemonServableHeadless.
func compilarDaemon(t *testing.T, root, salida string) error {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", salida, "./cmd/arnesia")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

// TestTierDesconocidoNoSeAsumeBarato es **el test que ata el defecto que el oráculo de doble
// costo cazó** (2026-07-26, corriendo el módulo contra el payload medido).
//
// El defecto: `cache_creation_tokens` llega por OTLP **sin decir a qué vencimiento se
// escribió**, y se estaba plegando al bucket de 5 minutos «porque es el default». Nuestra
// propia evidencia lo falsifica — el `result` de esa misma corrida dice
// `ephemeral_1h = 8257, ephemeral_5m = 0`, y el costo reportado (18 473 micros) coincide
// exacto con la tarifa de 1 h. Cotizarlo a 5 m daba **12 280**: un 33 % por debajo. Es
// `phoenix#14314`, el bug nº 3 de los que este módulo existe para no reproducir.
//
// Lo que el test asegura, y por qué nada lo cazaba antes:
//
//  1. **El calculado NO se presenta como completo.** `Completo:false` solo se encendía cuando
//     un bucket no tenía tarifa; el de 5 m SÍ la tiene, así que el número salía «completo» y
//     mentía.
//  2. **La divergencia contra el reportado queda VISIBLE.** El flag existía y no llegaba a
//     ninguna salida: un operador veía dos cifras que difieren un tercio y nada que le
//     dijera por qué.
//  3. **El motivo se nombra.** Un «incompleto» sin razón es un aviso que nadie puede accionar.
func TestTierDesconocidoNoSeAsumeBarato(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	receptor := otlp.NewReceptor(svc, otlp.Opciones{})
	srv := httptest.NewServer(receptor)
	defer srv.Close()

	crudo, err := os.ReadFile(evidencia(t, "logs-run1.json"))
	if err != nil {
		t.Fatal(err)
	}
	postearFitness(t, srv.URL+"/v1/logs", crudo)
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros == nil {
		t.Fatal("control positivo: el costo reportado del golden tiene que llegar")
	}
	if *r.CostoReportadoMicros != 18473 {
		t.Fatalf("el costo reportado medido es 18473 micros, llegó %d", *r.CostoReportadoMicros)
	}

	// 1 · el calculado NO puede presentarse como completo.
	if r.CostoCompleto == nil {
		t.Fatal("con un costo calculado presente, `costo_completo` tiene que viajar (no puede ser nil)")
	}
	if *r.CostoCompleto {
		t.Error("el canal OTLP no dice a qué vencimiento se escribió el cache: el costo calculado " +
			"NO puede salir marcado como completo — así se coló phoenix#14314")
	}
	// 3 · y el motivo se nombra.
	if len(r.SinTarifa) == 0 {
		t.Error("un «incompleto» sin razón es un aviso que nadie puede accionar")
	}
	nombrado := false
	for _, s := range r.SinTarifa {
		if s == "cache_escritura_sin_tier" {
			nombrado = true
		}
	}
	if !nombrado {
		t.Errorf("el motivo tiene que nombrar el bucket sin tier: %v", r.SinTarifa)
	}

	// 2 · **la divergencia queda visible, y marcada.**
	if r.CostoCalculadoMicros == nil {
		t.Fatal("el calculado tiene que existir igual: es una cota inferior declarada, no una ausencia")
	}
	if r.DivergenciaPct == nil {
		t.Fatal("con los dos costos presentes, la divergencia tiene que viajar")
	}
	// M9 · esta divergencia está EXPLICADA (el motivo viaja en `sin_tarifa`), así que NO se
	// marca sospechosa: una alarma que suena en todas las corridas no es una alarma. Lo que
	// se exige acá es que la divergencia sea VISIBLE y que el motivo la acompañe.
	if r.DivergenciaSospechosa {
		t.Errorf("una divergencia con motivo declarado (%v) no es sospechosa: el oráculo tiene "+
			"que apuntar a lo INEXPLICADO, o suena siempre y se aprende a ignorarlo", r.SinTarifa)
	}
	// Y NO se cotizó al tramo barato: el calculado tiene que ser el que EXCLUYE el bucket sin
	// tier (1 959 micros), no los 12 280 de asumir 5 minutos.
	if *r.CostoCalculadoMicros == 12280 {
		t.Fatalf("REGRESIÓN de phoenix#14314: el calculado dio %d micros, que es exactamente el "+
			"valor de aplanar el tier de cache write a 5 minutos", *r.CostoCalculadoMicros)
	}
	t.Logf("reportado=%d calculado=%d divergencia=%.1f %% completo=%v motivo=%v",
		*r.CostoReportadoMicros, *r.CostoCalculadoMicros, *r.DivergenciaPct, *r.CostoCompleto, r.SinTarifa)
}

// TestB1NoDetectaReWarmSobreUnTierFabricado — la otra mitad de la corrección: si el bucket sin
// tier se hubiera plegado a 5 m, B1 vería escrituras de vencimiento corto que **nunca
// ocurrieron** y «detectaría» un re-warm inexistente, recomendando un cambio sobre un dato
// inventado.
func TestB1NoDetectaReWarmSobreUnTierFabricado(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	receptor := otlp.NewReceptor(svc, otlp.Opciones{})
	srv := httptest.NewServer(receptor)
	defer srv.Close()

	crudo, err := os.ReadFile(evidencia(t, "logs-run1.json"))
	if err != nil {
		t.Fatal(err)
	}
	postearFitness(t, srv.URL+"/v1/logs", crudo)
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	mej, err := svc.Mejoras(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range mej.Puntos {
		if p.Detector == domain.DetB1 {
			t.Fatalf("B1 no puede hallar re-warm sin saber a qué vencimiento se escribió: %+v", p)
		}
	}
	// ── control positivo: B1 aparece en `no_aplican` CON su motivo ──
	var motivo string
	visto := false
	for _, na := range mej.NoAplican {
		if na.Detector == domain.DetB1 {
			visto, motivo = true, na.Motivo
		}
	}
	if !visto {
		t.Fatal("B1 tiene que declararse: apagado sin decirlo sería un hueco escondido")
	}
	if motivo == "" {
		t.Error("B1 se apagó sin motivo")
	}
	t.Logf("B1 no aplica, y lo dice: %q", motivo)
}

// TestElDineroSeCuentaUnaSolaVez es **el test del defecto más grave que la auditoría encontró**
// (C1, 2026-07-26), reproducido tal cual lo reprodujo el auditor.
//
// El defecto: `claude_code.cost.usage` (canal `/v1/metrics`) y `api_request.cost_usd_micros`
// (canal `/v1/logs`) son **el mismo gasto de la misma llamada**. Los dos entraban como
// `api_request` y `SUM()` los sumaba: el mismo turno reportaba **18 473 con un exportador y
// 36 946 con los dos**. Y no era un caso de laboratorio — `SpawnEnv` enciende los DOS
// exportadores en todo spawn de S1, así que la configuración que el propio módulo prescribe
// era la que producía el doble conteo.
//
// Para un producto cuya única promesa es decir cuánto cuesta algo, sobre-reportar el doble es
// el peor defecto posible: dispara decisiones de gasto sobre un número inventado hacia arriba.
//
// La regla que este test enforcea: **una unidad de gasto se cuenta UNA sola vez.**
func TestElDineroSeCuentaUnaSolaVez(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	srv := httptest.NewServer(otlp.NewReceptor(svc, otlp.Opciones{}))
	defer srv.Close()

	logs, err := os.ReadFile(evidencia(t, "logs-run1.json"))
	if err != nil {
		t.Fatal(err)
	}
	metricas, err := os.ReadFile(evidencia(t, "metrics-run1.json"))
	if err != nil {
		t.Fatal(err)
	}

	// PASO 1 · solo el canal primario.
	postearFitness(t, srv.URL+"/v1/logs", logs)
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}
	soloLogs, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if soloLogs.CostoReportadoMicros == nil {
		t.Fatal("control positivo: el canal primario tiene que traer el costo")
	}
	primario := *soloLogs.CostoReportadoMicros
	if primario != 18473 {
		t.Fatalf("el costo del golden es 18473 micros, llegó %d", primario)
	}

	// PASO 2 · la MISMA corrida, ahora también por el canal secundario. Es exactamente lo
	// que pasa en producción: `SpawnEnv` enciende los dos exportadores.
	postearFitness(t, srv.URL+"/v1/metrics", metricas)
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}
	conAmbos, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if conAmbos.CostoReportadoMicros == nil {
		t.Fatal("con los dos canales el costo tiene que seguir estando")
	}
	if got := *conAmbos.CostoReportadoMicros; got != primario {
		t.Fatalf("EL DINERO SE CONTÓ DOS VECES: con un canal %d micros, con los dos %d. "+
			"`claude_code.cost.usage` y `api_request.cost_usd_micros` son EL MISMO gasto — "+
			"una unidad de gasto se cuenta una sola vez", primario, got)
	}
	// Y tampoco se duplican los tokens ni los turnos.
	if conAmbos.Turnos != soloLogs.Turnos {
		t.Errorf("los turnos se duplicaron: %d → %d", soloLogs.Turnos, conAmbos.Turnos)
	}

	// ── control positivo: el dato del canal secundario NO se tiró, está guardado ──
	// Se guarda porque tirarlo sería perder señal; simplemente no suma.
	var metricasGuardadas int
	if qerr := st.ReaderParaTest().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM evento WHERE tipo_evento = ?`, string(domain.EventoMetrica)).
		Scan(&metricasGuardadas); qerr != nil {
		t.Fatal(qerr)
	}
	if metricasGuardadas == 0 {
		t.Fatal("control positivo: el canal secundario se GUARDA (no se tira), solo que no suma — " +
			"si no hay filas, el test de arriba pasaría porque el dato desapareció, que es otro bug")
	}
	t.Logf("un canal: %d micros · dos canales: %d micros · filas del canal secundario guardadas: %d",
		primario, *conAmbos.CostoReportadoMicros, metricasGuardadas)
}

// TestUnCostoQueNoSePudoCotizarNoViajaComoCero — **C2 de la auditoría, en el wire**.
//
// Un `api_request` sin desglose de tokens —el caso real del canal de métricas— salía con
// `costo_calculado_micros: 0` y `costo_completo: true`. Un turno que costó USD 0,018473
// reportados se mostraba como **USD 0,000000 calculado**, sin ninguna marca. Y el 0 no se
// quedaba quieto: entraba al `SUM()` del resumen y al veredicto de completitud como un voto a
// favor.
//
// Cero tokens no es cero pesos: es «no sé cuánto».
func TestUnCostoQueNoSePudoCotizarNoViajaComoCero(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	micros := int64(18473)
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin:  domain.LlaveJoin{SesionID: "s-1", TurnoID: "t-1", ArnesID: "vitalia", CajaID: "paso-3"},
		Emisor:     domain.EmisorOTLP,
		Runtime:    "claude-code",
		TSRecibido: time.Now().UTC(),
		TipoEvento: domain.EventoAPIRequest,
		Escenario:  domain.EscenarioS2Instrumentado,
		Modelo:     "claude-haiku-4-5",
		// Sin ningún bucket con tokens: exactamente lo que trae el canal de métricas.
		CostoReportadoMicros: &micros,
	}}); err != nil {
		t.Fatal(err)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	// En la FILA: `costo_calculado_micros` tiene que ser NULL, no 0.
	var calc, completo interface{}
	if qerr := st.ReaderParaTest().QueryRowContext(ctx,
		`SELECT costo_calculado_micros, costo_completo FROM evento LIMIT 1`).Scan(&calc, &completo); qerr != nil {
		t.Fatal(qerr)
	}
	if calc != nil {
		t.Errorf("sin tokens no hay nada que cotizar: costo_calculado_micros = %v, tiene que ser NULL", calc)
	}
	if completo != nil {
		t.Errorf("un costo que no existe no puede declararse completo: costo_completo = %v", completo)
	}

	// En el RESUMEN: el 0 no puede entrar al total ni votar «completo».
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoCalculadoMicros != nil {
		t.Errorf("el resumen no puede reportar un calculado de %d cuando no se cotizó nada",
			*r.CostoCalculadoMicros)
	}
	// ── control positivo: lo que SÍ se pudo medir sigue viajando ──
	if r.CostoReportadoMicros == nil || *r.CostoReportadoMicros != 18473 {
		t.Fatalf("control positivo: el costo REPORTADO tiene que seguir estando: %v", r.CostoReportadoMicros)
	}
}

// TestElTotalDeLaCajaEsElTotal — **C4 de la auditoría**: el truncado silencioso.
//
// `Store.Turnos` devuelve una PÁGINA (500 por default) y el detalle sumaba **sobre esa lista
// recortada**, presentando el resultado como total de la caja. Con 600 turnos de 1 000 micros:
//
//	resumen  → 600 000 micros · 600 turnos
//	detalle  → 500 000 micros · 500 turnos · **sin ningún campo que avisara**
//
// Y el detector reportaba `parte_del_total = 0,8333` para una caja que se lleva el **100 %**,
// con «500 corridas totales» donde hubo 600 — numerador recortado, denominador completo.
//
// Dos pantallas del mismo dato mostrando cifras distintas, y ninguna diciendo por qué, es
// exactamente lo que la doctrina de honestidad de este paquete prohíbe.
func TestElTotalDeLaCajaEsElTotal(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()

	const n = 600
	evs := make([]domain.EventoTelemetria, 0, n)
	base := time.Now().UTC()
	for i := 0; i < n; i++ {
		micros := int64(1000)
		ts := base.Add(-time.Duration(i) * time.Second)
		evs = append(evs, domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{
				SesionID: "s-masiva", TurnoID: fmt.Sprintf("t-%04d", i),
				ArnesID: "masivo", CajaID: "caja-unica",
			},
			Emisor: domain.EmisorOTLP, Runtime: "claude-code",
			TSRecibido: ts, TSEmisor: &ts,
			TipoEvento: domain.EventoAPIRequest, Escenario: domain.EscenarioS2Instrumentado,
			CostoReportadoMicros: &micros,
		})
	}
	if _, err := svc.Ingerir(ctx, evs); err != nil {
		t.Fatal(err)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	q := ports.ConsultaTelemetria{ArnesID: "masivo"}
	resumen, err := svc.Resumen(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	if resumen.Turnos != n {
		t.Fatalf("control positivo: el resumen tiene que ver los %d turnos, ve %d", n, resumen.Turnos)
	}

	q.CajaID = "caja-unica"
	det, err := svc.DetalleCaja(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	// 1 · el total del detalle es EL TOTAL, no el de la página.
	if det.Paridad.ReportadoMicros == nil {
		t.Fatal("el detalle tiene que traer el costo")
	}
	if *det.Paridad.ReportadoMicros != *resumen.CostoReportadoMicros {
		t.Errorf("dos pantallas del mismo dato no coinciden: resumen %d, detalle %d",
			*resumen.CostoReportadoMicros, *det.Paridad.ReportadoMicros)
	}
	if det.TurnosTotales != n {
		t.Errorf("turnos_totales = %d, hubo %d", det.TurnosTotales, n)
	}
	// 2 · y el recorte se DECLARA.
	if len(det.Turnos) >= n {
		t.Skip("la página no se recortó; el resto del test no aplica")
	}
	if !det.Truncado {
		t.Errorf("la página trae %d de %d turnos y no lo declara — una degradación que no se "+
			"declara es un total parcial disfrazado de total", len(det.Turnos), det.TurnosTotales)
	}

	// 3 · los detectores no producen porcentajes con numerador recortado.
	mej, err := svc.Mejoras(ctx, q)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range mej.Puntos {
		if p.ParteDelTotal > 0 && p.ParteDelTotal < 0.99 {
			t.Errorf("%s: parte_del_total = %.4f sobre una caja que se lleva el 100 %% — "+
				"numerador recortado, denominador completo", p.Detector, p.ParteDelTotal)
		}
		if p.CorridasTotales < n && p.CorridasTotales > 0 {
			t.Errorf("%s: declara %d corridas totales donde hubo %d", p.Detector, p.CorridasTotales, n)
		}
	}
	// ── control positivo: los detectores se declararon apagados CON MOTIVO ──
	if len(mej.Puntos) == 0 {
		conMotivo := 0
		for _, na := range mej.NoAplican {
			if na.Motivo != "" {
				conMotivo++
			}
		}
		if conMotivo == 0 {
			t.Fatal("si no hay puntos, los detectores tienen que estar declarados apagados con motivo — " +
				"callarlos sería el hueco escondido que el paquete prohíbe")
		}
		t.Logf("con la ventana truncada los %d detectores se declaran apagados con motivo", conMotivo)
	}
}

// TestNingunIdentificadorEsUnaRutaDelUsuario — **S1 de la auditoría**, y la clase de fuga que
// la allowlist NO puede atrapar.
//
// El campo `instalacion_id` es legítimo y está en la lista de permitidos; lo que estaba mal era
// su **contenido**: se le asignaba `inst.ProyectoPath`, un path del filesystem del usuario
// (`/home/<usuario>/Proyectos/<lo-que-sea>`), que se persistía crudo y **egresaba por el
// forward**. Contradecía el comentario de A14 tres líneas más arriba de la propia función.
//
// La allowlist protege por NOMBRE de campo. Esto protege por FORMA del valor: ningún
// identificador que se persiste puede parecerse a una ruta absoluta ni contener el nombre de
// un usuario. Es la única barrera que atrapa «campo permitido con contenido prohibido».
func TestNingunIdentificadorEsUnaRutaDelUsuario(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()

	// Los dos caminos que producen un identificador de instalación: el hook (que ya usaba
	// huella) y el Portafolio (que usaba la ruta cruda).
	const rutaReal = "/home/un-usuario-real/Proyectos/vitalia"
	svc.SetPortafolio(
		func(ctx context.Context, arnesID string) string { return "" },
		func(ctx context.Context) []usecase.FilaInstalacion {
			// Se simula lo que el composition root arma. Si alguien vuelve a poner el path
			// crudo acá, el assert de abajo lo caza.
			return []usecase.FilaInstalacion{{
				ArnesID: "vitalia", InstalacionID: hooks.HuellaCWD(rutaReal), Clave: "k1",
			}}
		},
	)
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin: domain.LlaveJoin{
			SesionID: "s-1", TurnoID: "t-1", ArnesID: "vitalia",
			InstalacionID: hooks.HuellaCWD(rutaReal),
		},
		Emisor: domain.EmisorOTLP, Runtime: "claude-code",
		TSRecibido: time.Now().UTC(), TipoEvento: domain.EventoAPIRequest,
		Escenario: domain.EscenarioS2Instrumentado,
	}}); err != nil {
		t.Fatal(err)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	// La FORMA de cada identificador persistido: ninguno puede parecer una ruta.
	rows, err := st.ReaderParaTest().QueryContext(ctx,
		`SELECT COALESCE(instalacion_id,''), COALESCE(cwd_huella,''), COALESCE(arnes_id,''),
		        COALESCE(caja_id,'') FROM evento`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	filas := 0
	for rows.Next() {
		var inst, huella, arnes, caja string
		if serr := rows.Scan(&inst, &huella, &arnes, &caja); serr != nil {
			t.Fatal(serr)
		}
		filas++
		for campo, v := range map[string]string{
			"instalacion_id": inst, "cwd_huella": huella, "arnes_id": arnes, "caja_id": caja,
		} {
			if v == "" {
				continue
			}
			if strings.HasPrefix(v, "/") || strings.HasPrefix(v, "~") ||
				strings.Contains(v, `:\`) || strings.Count(v, "/") > 1 {
				t.Errorf("%s = %q tiene forma de RUTA: un campo permitido con contenido "+
					"prohibido es la fuga que la allowlist no atrapa (A14)", campo, v)
			}
			for _, marca := range []string{"/home/", "/Users/", "Proyectos", "un-usuario-real"} {
				if strings.Contains(v, marca) {
					t.Errorf("%s = %q contiene %q: la ruta del usuario nunca entra al almacén", campo, v, marca)
				}
			}
		}
	}
	if filas == 0 {
		t.Fatal("no se leyó ninguna fila: el test no verificó nada")
	}

	// ── control positivo: el identificador SÍ existe y agrupa ──
	// No alcanza con que no sea una ruta: tiene que servir para agrupar corridas del mismo
	// lugar, o lo habríamos «arreglado» borrando el dato.
	var conInstalacion int
	if qerr := st.ReaderParaTest().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM evento WHERE instalacion_id = ?`, hooks.HuellaCWD(rutaReal)).
		Scan(&conInstalacion); qerr != nil {
		t.Fatal(qerr)
	}
	if conInstalacion == 0 {
		t.Fatal("control positivo: la huella tiene que estar y agrupar — si no, el arreglo " +
			"habría sido borrar el dato, no protegerlo")
	}
	// Y la huella es estable y no reversible.
	if hooks.HuellaCWD(rutaReal) == hooks.HuellaCWD(rutaReal+"-otro") {
		t.Error("dos lugares distintos no pueden compartir huella")
	}
	if strings.Contains(hooks.HuellaCWD(rutaReal), "vitalia") {
		t.Error("la huella no puede contener el nombre del proyecto")
	}
}

// TestElAgregadoNoConservaLoQueNadiePuedeLeer — **A2/C3 de la auditoría**: el agregado
// afirmaba una retención que no existía.
//
// La doc prometía *«el rollup sobrevive más que el detalle: tras purgar, el drill-down dice
// detalle purgado, resumen conservado»*. Era falso: **ninguna consulta de lectura toca
// `rollup_hora`**, así que tras la purga quedaban filas que el producto no muestra,
// sosteniendo una promesa vacía.
//
// Este test enforcea la coherencia elegida: **lo que las lecturas no devuelven, el agregado no
// lo conserva.** El día que el rollup entre al camino de lectura, este test tiene que cambiar
// — y ese cambio es justamente la señal de que la promesa pasó a ser verdad.
func TestElAgregadoNoConservaLoQueNadiePuedeLeer(t *testing.T) {
	dir := t.TempDir()
	ahora := time.Now().UTC()
	st, err := telstore.New(filepath.Join(dir, "telemetria.db"),
		telstore.Opciones{LoteEspera: 10 * time.Millisecond, Reloj: func() time.Time { return ahora }})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	roll := telstore.NewRollup(st, time.Hour)
	defer roll.Detener()
	svc := usecase.NewTelemetriaService(st, telcatalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(), otlp.PerfilClaudeCode(), func() time.Time { return ahora })
	svc.SetRetencion(st, roll, 90, 24)
	ctx := context.Background()

	viejo := ahora.AddDate(0, 0, -120) // más allá del TTL
	micros := int64(50_000)
	if _, ierr := svc.Ingerir(ctx, []domain.EventoTelemetria{{
		LlaveJoin: domain.LlaveJoin{SesionID: "s-vieja", TurnoID: "t-1", ArnesID: "viejo"},
		Emisor:    domain.EmisorOTLP, Runtime: "claude-code", TSRecibido: viejo,
		TipoEvento: domain.EventoAPIRequest, Escenario: domain.EscenarioS2Instrumentado,
		CostoReportadoMicros: &micros,
	}}); ierr != nil {
		t.Fatal(ierr)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}
	if aerr := roll.Actualizar(ctx); aerr != nil {
		t.Fatal(aerr)
	}
	// Control positivo previo: el agregado SÍ tenía la fila.
	var antes int
	if qerr := st.ReaderParaTest().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rollup_hora WHERE arnes_id = 'viejo'`).Scan(&antes); qerr != nil {
		t.Fatal(qerr)
	}
	if antes == 0 {
		t.Fatal("control positivo: el agregado tenía que tener la fila antes de purgar")
	}

	if _, perr := svc.Purgar(ctx, ports.PurgaTelemetria{}); perr != nil {
		t.Fatal(perr)
	}

	// Lo que las lecturas devuelven: nada.
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{
		ArnesID: "viejo", Desde: viejo.AddDate(0, 0, -1), Hasta: ahora,
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros != nil {
		t.Fatalf("tras la purga la lectura no devuelve nada: %d", *r.CostoReportadoMicros)
	}
	// Y el agregado no conserva lo que la lectura no devuelve.
	var despues int
	if qerr := st.ReaderParaTest().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rollup_hora WHERE arnes_id = 'viejo'`).Scan(&despues); qerr != nil {
		t.Fatal(qerr)
	}
	if despues != 0 {
		t.Errorf("quedaron %d filas de agregado de una hora purgada: afirman una retención que "+
			"ninguna pantalla puede mostrar", despues)
	}
}

// TestElOraculoApuntaALoInexplicado — **M9 de la auditoría**: el oráculo estaba saturado.
//
// En el camino normal la divergencia ronda el 90 % **siempre** (OTLP nunca dice el vencimiento
// del cache write), así que `divergencia_sospechosa` se encendía en todas las corridas. Una
// alarma que suena siempre no es una alarma: entrena a ignorarla, y entonces no sirve para lo
// único que existe — cazar el caso en que el costeo está mal de verdad.
//
// La regla: sospechosa es la divergencia **sin explicación**. Con motivo declarado, no.
func TestElOraculoApuntaALoInexplicado(t *testing.T) {
	svc, st := servicioFitness(t)
	ctx := context.Background()
	base := func(sesion string, tk domain.Tokens, reportado int64) domain.EventoTelemetria {
		r := reportado
		return domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{SesionID: sesion, TurnoID: "t-1", ArnesID: sesion},
			Emisor:    domain.EmisorOTLP, Runtime: "claude-code", TSRecibido: time.Now().UTC(),
			TipoEvento: domain.EventoAPIRequest, Escenario: domain.EscenarioS2Instrumentado,
			Modelo: "claude-haiku-4-5", Tokens: tk, CostoReportadoMicros: &r,
		}
	}
	// A · divergencia EXPLICADA: el cache write llegó sin vencimiento declarado.
	sinTier := int64(8257)
	// B · divergencia INEXPLICADA: todos los buckets cotizables, y aun así no coincide.
	entrada := int64(10)
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{
		base("explicada", domain.Tokens{Entrada: &entrada, CacheEscrituraSinTier: &sinTier}, 18473),
		base("inexplicada", domain.Tokens{Entrada: &entrada}, 999_999),
	}); err != nil {
		t.Fatal(err)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	expl, err := svc.Resumen(ctx, ports.ConsultaTelemetria{ArnesID: "explicada"})
	if err != nil {
		t.Fatal(err)
	}
	if expl.DivergenciaPct == nil {
		t.Fatal("la divergencia tiene que ser VISIBLE aunque esté explicada")
	}
	if expl.DivergenciaSospechosa {
		t.Errorf("divergencia con motivo (%v) marcada sospechosa: el oráculo suena siempre y "+
			"se aprende a ignorarlo", expl.SinTarifa)
	}
	if len(expl.SinTarifa) == 0 {
		t.Error("si no es sospechosa es porque hay motivo, y el motivo tiene que viajar")
	}

	// ── control positivo: la INEXPLICADA sí dispara ──
	inex, err := svc.Resumen(ctx, ports.ConsultaTelemetria{ArnesID: "inexplicada"})
	if err != nil {
		t.Fatal(err)
	}
	if inex.DivergenciaPct == nil {
		t.Fatal("control positivo: la divergencia inexplicada tiene que calcularse")
	}
	if !inex.DivergenciaSospechosa {
		t.Fatalf("control positivo: una divergencia del %.1f %% SIN motivo declarado tiene que "+
			"encender el aviso — es para lo que el oráculo existe", *inex.DivergenciaPct)
	}
	t.Logf("explicada: %.1f %% (no suena, motivo %v) · inexplicada: %.1f %% (SUENA)",
		*expl.DivergenciaPct, expl.SinTarifa, *inex.DivergenciaPct)
}
