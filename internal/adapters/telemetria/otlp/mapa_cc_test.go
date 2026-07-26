package otlp

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

var reloj = time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)

// TestMapeoCCCompleto corre el mapeo contra los TRES golden reales y verifica, evento por
// evento, que el `api_request` trae lo que el ANEXO midió: la llave del join completa, los
// buckets, el costo reportado y la duración.
func TestMapeoCCCompleto(t *testing.T) {
	perfil := PerfilClaudeCode()
	apiRequests, herramientas, ignorados := 0, 0, 0

	for _, f := range []string{"logs-run1.json", "logs-run2-con-skill.json", "logs-run4-con-tools.json"} {
		regs, err := DecodificarLogs(golden(t, f))
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		for _, r := range regs {
			ev, d, merr := MapearLogRecord(r, perfil, reloj)
			if errors.Is(merr, ErrEventoIgnorado) {
				ignorados++
				continue
			}
			if merr != nil {
				t.Fatalf("%s: %v", f, merr)
			}
			// El descarte SE CUENTA: los cinco campos de identidad llegan en cada record.
			if d.Identidad != len(NegadosDuros()) {
				t.Errorf("%s: se esperaban %d campos de identidad descartados, se contaron %d",
					f, len(NegadosDuros()), d.Identidad)
			}
			if ev.SesionID == "" || ev.TurnoID == "" {
				t.Errorf("%s: la llave del join llegó incompleta: sesion=%q turno=%q", f, ev.SesionID, ev.TurnoID)
			}
			if ev.AdaptadorVersion != AdaptadorVersionCC {
				t.Errorf("%s: adaptador_version=%q, se esperaba %q", f, ev.AdaptadorVersion, AdaptadorVersionCC)
			}
			switch ev.TipoEvento {
			case domain.EventoAPIRequest:
				apiRequests++
				if ev.Modelo == "" {
					t.Errorf("%s: api_request sin modelo", f)
				}
				if ev.CostoReportadoMicros == nil {
					t.Errorf("%s: api_request sin cost_usd_micros — es el canal del dinero", f)
				}
				if ev.Tokens.Entrada == nil || ev.Tokens.Salida == nil {
					t.Errorf("%s: api_request sin los buckets básicos", f)
				}
				// El split NO viene por OTel: `CacheEscritura1h` tiene que quedar nil.
				// Ponerle el total sería inventar en qué tramo se escribió.
				if ev.Tokens.CacheEscritura1h != nil {
					t.Errorf("%s: OTel no trae el split 5m/1h; CacheEscritura1h debe quedar nil, es %d",
						f, *ev.Tokens.CacheEscritura1h)
				}
				// Razonamiento tampoco: Claude Code no lo reporta.
				if ev.Tokens.Razonamiento != nil {
					t.Errorf("%s: Razonamiento debe quedar nil (ausencia ≠ 0)", f)
				}
			case domain.EventoHerramienta:
				herramientas++
			}
		}
	}
	if apiRequests == 0 {
		t.Fatal("ningún api_request mapeado en los tres golden — el mapeo no está leyendo el canal del dinero")
	}
	if herramientas == 0 {
		t.Fatal("ningún evento de herramienta mapeado (logs-run4 trae tool_decision y tool_result)")
	}
	if ignorados == 0 {
		t.Fatal("ningún evento ignorado: los golden traen assistant_response y hook_execution_*, " +
			"que el MVP descarta a propósito. Si no se ignoró nada, el default-deny no está corriendo")
	}
	t.Logf("mapeados: %d api_request · %d herramienta · %d ignorados por default-deny", apiRequests, herramientas, ignorados)
}

// TestMapeoAtribuyeExactoConLosArnesiaAttrs — los cuatro `arnesia.*` los inyectamos nosotros
// y viajan copiados en cada log record (V4). Su presencia ES la atribución exacta.
func TestMapeoAtribuyeExactoConLosArnesiaAttrs(t *testing.T) {
	regs, err := DecodificarLogs(golden(t, "logs-run1.json"))
	if err != nil {
		t.Fatal(err)
	}
	exactos := 0
	for _, r := range regs {
		ev, _, merr := MapearLogRecord(r, PerfilClaudeCode(), reloj)
		if merr != nil {
			continue
		}
		if ev.ArnesID != "vitalia" {
			t.Errorf("arnes_id=%q, el golden trae vitalia", ev.ArnesID)
		}
		if ev.Atribucion != domain.ConfianzaExacta {
			t.Errorf("con arnesia.arnes presente la atribución es exacta, es %q", ev.Atribucion)
		}
		exactos++
	}
	if exactos == 0 {
		t.Fatal("ningún evento atribuido — el test no comparó nada")
	}
	// Control negativo con el mismo mapeo: sin los `arnesia.*`, la atribución NO se inventa.
	sin := RegistroLog{Attrs: Atributos{
		"session.id": txt("s-1"), "prompt.id": txt("t-1"), "event.name": txt("api_request"),
	}}
	ev, _, err := MapearLogRecord(sin, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Atribucion != domain.ConfianzaSinDato {
		t.Errorf("sin arnesia.* la atribución es sin-dato, es %q", ev.Atribucion)
	}
}

// TestEscenarioSeDerivaDeLaSenalEnElMapeo — `s1` si hay corrida nuestra, `s2-instrumentado`
// si hay `api_request` sin corrida. **Ningún campo del emisor puede elegirlo.**
func TestEscenarioSeDerivaDeLaSenalEnElMapeo(t *testing.T) {
	base := Atributos{
		"session.id": txt("s"), "prompt.id": txt("t"), "event.name": txt("api_request"),
		"model": txt("claude-haiku-4-5"),
	}
	// Con corrida nuestra ⇒ s1.
	conCorrida := Atributos{}
	for k, v := range base {
		conCorrida[k] = v
	}
	conCorrida["arnesia.corrida"] = txt("run-9")
	conCorrida["arnesia.arnes"] = txt("vitalia")
	ev1, _, err := MapearLogRecord(RegistroLog{Attrs: conCorrida}, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev1.Escenario != domain.EscenarioS1 {
		t.Errorf("con arnesia.corrida el escenario es s1, es %q", ev1.Escenario)
	}
	// Sin corrida ⇒ s2-instrumentado.
	ev2, _, err := MapearLogRecord(RegistroLog{Attrs: base}, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev2.Escenario != domain.EscenarioS2Instrumentado {
		t.Errorf("sin corrida nuestra el escenario es s2-instrumentado, es %q", ev2.Escenario)
	}
	// Y el emisor NO puede declararlo: un atributo `escenario` se descarta como cualquier
	// otro fuera de lista.
	mintiendo := Atributos{}
	for k, v := range base {
		mintiendo[k] = v
	}
	mintiendo["escenario"] = txt("s1")
	mintiendo["arnesia.escenario"] = txt("s1")
	ev3, d, err := MapearLogRecord(RegistroLog{Attrs: mintiendo}, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatal(err)
	}
	if ev3.Escenario != domain.EscenarioS2Instrumentado {
		t.Errorf("un arnés no puede declarar su propio escenario; quedó %q", ev3.Escenario)
	}
	if d.FueraDeLista < 2 {
		t.Errorf("los dos atributos inventados debieron contarse como descarte, se contaron %d", d.FueraDeLista)
	}
}

// TestLaIdentidadNoCruzaLaPuerta — el test central de RF-282 en el camino OTLP. Se alimenta
// el mapeo con un payload que trae los CINCO campos de identidad con valores BUSCABLES (los
// golden los traen redactados, así que no servirían para asertar sobre el valor) y se
// verifica que ninguno aparece en el evento serializado.
//
// ⚠ Control positivo en la misma corrida: los campos de la allowlist SÍ están. Sin eso, un
// mapeo que devolviera un evento vacío pasaría el assert de ausencia.
func TestLaIdentidadNoCruzaLaPuerta(t *testing.T) {
	const (
		marcaEmail = "PII-EMAIL-a71f@ejemplo.test"
		marcaCta   = "PII-CUENTA-9d02"
		marcaOrg   = "PII-ORG-4b8e"
		marcaUser  = "PII-USER-2c55"
		marcaUUID  = "PII-UUID-0f13"
	)
	attrs := Atributos{
		"session.id": txt("SES-CONTROL-POSITIVO"), "prompt.id": txt("TUR-CONTROL-POSITIVO"),
		"event.name": txt("api_request"), "model": txt("claude-haiku-4-5"),
		"input_tokens": num(10), "cost_usd_micros": num(18473),
		"arnesia.arnes": txt("vitalia"),
		// los cinco, con valores que se pueden buscar
		"user.email":        txt(marcaEmail),
		"user.account_id":   txt(marcaCta),
		"user.account_uuid": txt(marcaUUID),
		"user.id":           txt(marcaUser),
		"organization.id":   txt(marcaOrg),
	}
	ev, d, err := MapearLogRecord(RegistroLog{Attrs: attrs}, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatalf("mapear: %v", err)
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{marcaEmail, marcaCta, marcaUUID, marcaUser, marcaOrg} {
		if bytes.Contains(raw, []byte(m)) {
			t.Errorf("la identidad cruzó la puerta: %q aparece en el evento serializado\n%s", m, raw)
		}
	}
	// ── control positivo, misma corrida ──
	if !bytes.Contains(raw, []byte("SES-CONTROL-POSITIVO")) {
		t.Fatal("control positivo: el session_id declarado NO llegó al evento — el assert de ausencia no prueba nada")
	}
	if !bytes.Contains(raw, []byte("TUR-CONTROL-POSITIVO")) {
		t.Fatal("control positivo: el prompt_id declarado NO llegó al evento")
	}
	if ev.CostoReportadoMicros == nil || *ev.CostoReportadoMicros != 18473 {
		t.Fatal("control positivo: el costo declarado NO llegó al evento")
	}
	if d.Identidad != 5 {
		t.Errorf("los 5 campos de identidad deben CONTARSE al descartarse, se contaron %d", d.Identidad)
	}
}

// TestAtributoNuevoCaeAfueraYSeCuenta — default-deny: un atributo que no está en la lista no
// entra, y el descarte incrementa un contador de salud. Un descarte silencioso es una señal
// perdida sin rastro.
func TestAtributoNuevoCaeAfueraYSeCuenta(t *testing.T) {
	attrs := Atributos{
		"session.id": txt("s"), "prompt.id": txt("t"), "event.name": txt("api_request"),
		"model": txt("claude-haiku-4-5"), "input_tokens": num(5),
		"cosa.nueva": txt("MARCA-COSA-NUEVA-3a91"),
	}
	ev, d, err := MapearLogRecord(RegistroLog{Attrs: attrs}, PerfilClaudeCode(), reloj)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(ev)
	if bytes.Contains(raw, []byte("MARCA-COSA-NUEVA-3a91")) {
		t.Errorf("un atributo fuera de lista entró al evento: %s", raw)
	}
	if d.FueraDeLista != 1 {
		t.Errorf("el descarte debe contarse: FueraDeLista=%d, se esperaba 1", d.FueraDeLista)
	}
	// Control positivo: lo declarado sí entró.
	if ev.Tokens.Entrada == nil || *ev.Tokens.Entrada != 5 {
		t.Error("control positivo: el bucket declarado no llegó al evento")
	}
}

// TestToolResultBytesSePersiste — A21: los dos tamaños se guardan desde el día 1 aunque B11
// esté fuera del MVP. Son BYTES, no contenido, y un dato que no se guarda hoy no se recupera
// mañana.
func TestToolResultBytesSePersiste(t *testing.T) {
	regs, err := DecodificarLogs(golden(t, "logs-run4-con-tools.json"))
	if err != nil {
		t.Fatal(err)
	}
	visto := false
	for _, r := range regs {
		if r.Attrs.Texto("event.name") != "tool_result" {
			continue
		}
		ev, _, merr := MapearLogRecord(r, PerfilClaudeCode(), reloj)
		if merr != nil {
			t.Fatalf("tool_result: %v", merr)
		}
		visto = true
		// Medido en la corrida real: input 143 bytes, result 9 bytes.
		if ev.ToolInputBytes == nil || *ev.ToolInputBytes != 143 {
			t.Errorf("tool_input_size_bytes = %v, el golden mide 143", ev.ToolInputBytes)
		}
		if ev.ToolResultBytes == nil || *ev.ToolResultBytes != 9 {
			t.Errorf("tool_result_size_bytes = %v, el golden mide 9", ev.ToolResultBytes)
		}
		if ev.Herramienta != "Read" {
			t.Errorf("herramienta=%q, el golden mide Read", ev.Herramienta)
		}
		// Y NINGÚN campo de contenido: son tamaños, no lo que la herramienta leyó.
		raw, _ := json.Marshal(ev)
		for _, prohibido := range []string{"tool_response", "tool_input\"", "contenido"} {
			if strings.Contains(string(raw), prohibido) {
				t.Errorf("el evento de tool_result no puede llevar contenido: %s", raw)
			}
		}
	}
	if !visto {
		t.Fatal("logs-run4-con-tools.json no produjo ningún tool_result — el golden o el mapeo cambiaron")
	}
}

// TestEventoSinSesionSeRechaza — sin `session.id` el evento no es atribuible ni deduplicable.
func TestEventoSinSesionSeRechaza(t *testing.T) {
	_, _, err := MapearLogRecord(RegistroLog{Attrs: Atributos{"event.name": txt("api_request")}}, PerfilClaudeCode(), reloj)
	if !errors.Is(err, domain.ErrEventoSinSesion) {
		t.Fatalf("se esperaba ErrEventoSinSesion, got %v", err)
	}
	// Control positivo: con sesión, entra.
	if _, _, err := MapearLogRecord(RegistroLog{Attrs: Atributos{
		"session.id": txt("s"), "event.name": txt("api_request"),
	}}, PerfilClaudeCode(), reloj); err != nil {
		t.Fatalf("con session.id el evento debe entrar: %v", err)
	}
}

// TestMetricasMapeanSusBuckets — el canal secundario. Las métricas llegan como `asDouble`
// (hallazgo medido) y desagregadas por `type`.
func TestMetricasMapeanSusBuckets(t *testing.T) {
	puntos, err := DecodificarMetricas(golden(t, "metrics-run1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var conCosto, conTokens int
	for _, p := range puntos {
		ev, _, merr := MapearPuntoMetrica(p, PerfilClaudeCode(), reloj)
		if errors.Is(merr, ErrEventoIgnorado) {
			continue
		}
		if merr != nil {
			t.Fatalf("%s: %v", p.Metrica, merr)
		}
		if ev.CostoReportadoMicros != nil {
			conCosto++
			// 0,0184726 USD → 18 473 micros, el mismo número que el log record.
			if *ev.CostoReportadoMicros != 18473 {
				t.Errorf("cost.usage → %d micros, se esperaba 18473", *ev.CostoReportadoMicros)
			}
		}
		if ev.Tokens.Entrada != nil || ev.Tokens.Salida != nil ||
			ev.Tokens.CacheLectura != nil || ev.Tokens.CacheEscritura5m != nil {
			conTokens++
		}
	}
	if conCosto == 0 {
		t.Error("ningún punto de costo mapeado")
	}
	t.Logf("métricas mapeadas: %d con costo · %d con tokens", conCosto, conTokens)
}

func txt(s string) valorAtributo { return valorAtributo{StringValue: &s} }
func num(n int64) valorAtributo {
	v := json.Number(strings.TrimSpace(json.Number(itoa(n)).String()))
	return valorAtributo{IntValue: &v}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
