package otlp

import (
	"fmt"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// mapa_cc.go traduce `claude_code.*` al evento canónico (arquitectura-modulo.md §4.4) con la
// **allowlist default-deny** de §6.1.
//
// 🔴 Por qué la allowlist no es opcional: la telemetría de Claude Code trae `user.email`,
// `user.account_uuid`, `user.account_id`, `user.id` y `organization.id` **en cada log record
// y en cada punto de métrica** (medido, INFORME §V6). Nada de eso hace falta para el join:
// alcanzan `session.id`, `prompt.id` y los `arnesia.*`.
//
// **Forma del enforcement:** `domain.NuevoEvento(campos map[string]any)` **no existe y no debe
// existir**. El evento se arma campo por campo desde las claves declaradas abajo. No hay
// ninguna ruta de código donde un mapa entero se copie al evento — que es la única forma de
// que la allowlist sea una garantía y no una intención.

// AdaptadorVersionCC versiona el MAPEO, no el runtime. Cambiar qué atributo va a qué campo
// obliga a bumpearlo (D7.5): es lo que permite explicar una divergencia histórica sin
// adivinar cuál de las dos versiones del código produjo cada fila.
const AdaptadorVersionCC = "cc-otlp/1"

// ScopeClaudeCode es el `scope.name` que Claude Code estampa en sus log records.
const ScopeClaudeCode = "com.anthropic.claude_code.events"

// allowlistOTLP es LA LISTA. Default-deny: un atributo que no está acá no entra al evento,
// venga de donde venga y se llame como se llame. Agregar uno es tocar este archivo.
//
// No se usa para copiar en masa (eso sería reintroducir el mapa-entero por la ventana): se
// usa para CONTAR los descartes, que es dato de salud, y para que el test de source-scan
// pueda contrastar la lista contra lo que el mapeo realmente lee.
var allowlistOTLP = map[string]bool{
	// llave del join
	"session.id": true, "prompt.id": true,
	// atribución exacta (la inyectamos nosotros al spawn)
	"arnesia.arnes": true, "arnesia.instalacion": true, "arnesia.caja": true, "arnesia.corrida": true,
	// qué evento es
	"event.name": true,
	// dinero
	"model": true, "input_tokens": true, "output_tokens": true,
	"cache_read_tokens": true, "cache_creation_tokens": true,
	"cost_usd_micros": true, "duration_ms": true,
	"speed": true, "service_tier": true, "query_source": true,
	// atribución por huella
	"plugin_id_hash": true, "plugin.scope": true,
	// proceso por OTel (ANEXO H8): decisión de permiso, éxito y duración POR HERRAMIENTA
	"tool_name": true, "tool_source": true, "decision": true, "source": true, "success": true,
	"tool_input_size_bytes": true, "tool_result_size_bytes": true,
	// versión del runtime, cuando viene versionada
	"terminal.type": true,
}

// negadosDuros son los cinco campos de identidad que el runtime manda en CADA punto. No
// están en la allowlist —así que ya estarían fuera— pero se nombran aparte para poder
// **contarlos** y para que el día que alguien agregue uno a la allowlist por accidente,
// `TestAllowlistNoPersistePII` lo cace por nombre y no por casualidad.
var negadosDuros = []string{"user.email", "user.account_uuid", "user.account_id", "user.id", "organization.id"}

// NegadosDuros expone la denylist para los tests de fitness y para el chequeo de conformance
// del arnés. Devuelve una copia: nadie de afuera muta la lista.
func NegadosDuros() []string { return append([]string(nil), negadosDuros...) }

// AllowlistOTLP expone la lista declarada, para el mismo uso.
func AllowlistOTLP() []string {
	out := make([]string, 0, len(allowlistOTLP))
	for k := range allowlistOTLP {
		out = append(out, k)
	}
	return out
}

// PerfilClaudeCode es el perfil que este adaptador ESTAMPA en cada evento (decisión A18):
// la aritmética y la acumulación son propiedad del adaptador, no de un agregador central.
func PerfilClaudeCode() domain.PerfilRuntime {
	return domain.PerfilRuntime{
		Runtime:          "claude-code",
		AdaptadorVersion: AdaptadorVersionCC,
		// Anthropic reporta los 4 buckets DISJUNTOS: se suman tal cual (verificado contra
		// el cost_usd_micros real en TestParidadConElCostoReportadoReal).
		Aritmetica:  domain.AritmeticaDisjunta,
		Acumulacion: domain.AcumulacionPorRequest,
		Soporta: []domain.DetectorID{
			domain.DetB4, domain.DetP1, domain.DetB2, domain.DetB6, domain.DetB3, domain.DetB1,
		},
	}
}

// Descartes cuenta lo que quedó afuera de la puerta. Es dato de SALUD: un atributo nuevo del
// runtime que se descarta en silencio y nadie cuenta es una señal perdida sin rastro.
type Descartes struct {
	FueraDeLista int
	Identidad    int
}

// ErrEventoIgnorado marca los `event.name` que el MVP no instrumenta a propósito
// (`hook_execution_start`/`_complete`: los cubre el hook propio con más contexto). No es un
// fallo — es una decisión, y por eso tiene su propio centinela en vez de un error genérico.
var ErrEventoIgnorado = fmt.Errorf("otlp: event.name fuera del alcance del MVP")

// MapearLogRecord arma el evento canónico de UN log record, campo por campo.
//
// Devuelve `ErrEventoIgnorado` para los eventos fuera de alcance y `ErrEventoSinSesion` para
// los que no traen `session.id` — sin sesión el evento no es atribuible ni deduplicable.
func MapearLogRecord(r RegistroLog, perfil domain.PerfilRuntime, ahora time.Time) (domain.EventoTelemetria, Descartes, error) {
	var d Descartes
	for k := range r.Attrs {
		if allowlistOTLP[k] {
			continue
		}
		d.FueraDeLista++
		for _, n := range negadosDuros {
			if k == n {
				d.Identidad++
			}
		}
	}

	nombre := r.Attrs.Texto("event.name")
	sesion := r.Attrs.Texto("session.id")
	if sesion == "" {
		return domain.EventoTelemetria{}, d, domain.ErrEventoSinSesion
	}

	ev := domain.EventoTelemetria{
		LlaveJoin: domain.LlaveJoin{
			// Los cuatro `arnesia.*` los inyectamos nosotros al spawn y viajan COPIADOS en
			// cada log record (V4) ⇒ su presencia es la atribución exacta.
			ArnesID:       r.Attrs.Texto("arnesia.arnes"),
			InstalacionID: r.Attrs.Texto("arnesia.instalacion"),
			CajaID:        r.Attrs.Texto("arnesia.caja"),
			SesionID:      sesion,
			// `prompt.id` es la OTRA MITAD de la llave del join (ANEXO H1). Sin él, el join
			// sería a nivel sesión y no alcanzaría para decir dónde se va el gasto.
			TurnoID:   r.Attrs.Texto("prompt.id"),
			CorridaID: r.Attrs.Texto("arnesia.corrida"),
		},
		Emisor:           domain.EmisorOTLP,
		Runtime:          perfil.Runtime,
		RuntimeVersion:   r.ScopeVersion,
		AdaptadorVersion: perfil.AdaptadorVersion,
		TSRecibido:       ahora.UTC(),
		TSEmisor:         r.TS,
		Aritmetica:       perfil.Aritmetica,
		Acumulacion:      perfil.Acumulacion,
		PluginIDHash:     r.Attrs.Texto("plugin_id_hash"),
	}

	switch nombre {
	case "api_request":
		ev.TipoEvento = domain.EventoAPIRequest
		ev.Modelo = r.Attrs.Texto("model")
		ev.Speed = r.Attrs.Texto("speed")
		ev.ServiceTier = r.Attrs.Texto("service_tier")
		// Los 4 buckets. `CacheEscritura5m`/`1h` quedan **nil** acá a propósito: el split no
		// viene por OTel, lo aporta el `result` del stream-json (§4.4). Poner el total en
		// uno de los dos sería inventar en qué tramo se escribió — y, como el de 1 h cuesta
		// 1,6×, inventar hacia abajo.
		ev.Tokens.Entrada = enteroOpcional(r.Attrs, "input_tokens")
		ev.Tokens.Salida = enteroOpcional(r.Attrs, "output_tokens")
		ev.Tokens.CacheLectura = enteroOpcional(r.Attrs, "cache_read_tokens")
		ev.CostoReportadoMicros = enteroOpcional(r.Attrs, "cost_usd_micros")
		ev.DuracionMs = enteroOpcional(r.Attrs, "duration_ms")
		// 🔴 `cache_creation_tokens` es el TOTAL de escritura de cache **sin desagregar**, y
		// va a su bucket propio — NO al de 5 minutos.
		//
		// Plegarlo a 5 m «porque es el TTL por default» es el bug `phoenix#14314`, y nuestra
		// propia evidencia lo falsifica: en la corrida del 2026-07-26 el `result` dice
		// `ephemeral_1h = 8257, ephemeral_5m = 0` — fue 1 hora, y el costo reportado (18 473
		// micros) coincide exacto con la tarifa de 1 h. Cotizarlo a 5 m da 12 280: un 33 %
		// por debajo. **Elegir la tarifa barata en la duda es inventar hacia abajo.**
		ev.Tokens.CacheEscrituraSinTier = enteroOpcional(r.Attrs, "cache_creation_tokens")

	case "tool_decision":
		// ANEXO H8: es señal de PROCESO que llega por OTel, no por hook.
		ev.TipoEvento = domain.EventoHerramienta
		ev.Herramienta = r.Attrs.Texto("tool_name")
		ev.Decision = r.Attrs.Texto("decision")
		ev.Motivo = r.Attrs.Texto("source")
		ev.Resultado = resultadoDeDecision(r.Attrs.Texto("decision"))

	case "tool_result":
		ev.TipoEvento = domain.EventoHerramienta
		ev.Herramienta = r.Attrs.Texto("tool_name")
		ev.DuracionMs = enteroOpcional(r.Attrs, "duration_ms")
		// A21: los dos tamaños se persisten desde el día 1 aunque B11 esté fuera del MVP.
		// Son BYTES, no contenido, y un dato que no se guarda hoy no se recupera mañana.
		ev.ToolInputBytes = enteroOpcional(r.Attrs, "tool_input_size_bytes")
		ev.ToolResultBytes = enteroOpcional(r.Attrs, "tool_result_size_bytes")
		if r.Attrs.Texto("success") == "false" {
			ev.Resultado = domain.ResultadoRechazado
		} else {
			ev.Resultado = domain.ResultadoOK
		}

	case "hook_execution_start", "hook_execution_complete":
		// Ignorados en el MVP: los cubre el hook propio, con más contexto. Se anota como
		// reserva, no como olvido.
		return domain.EventoTelemetria{}, d, ErrEventoIgnorado

	default:
		// **Default-deny también a nivel evento.** `assistant_response`,
		// `mcp_server_connection`, `user_prompt` y cualquier `event.name` futuro caen acá:
		// no se persisten. El descarte ya quedó contado arriba.
		return domain.EventoTelemetria{}, d, ErrEventoIgnorado
	}

	// La atribución se decide con lo que HAY, en el orden de §6.3. El resto —por hash y por
	// proceso— lo resuelve el servicio, que es quien tiene la tabla y el Portafolio.
	if ev.ArnesID != "" {
		ev.Atribucion = domain.ConfianzaExacta
	} else {
		ev.Atribucion = domain.ConfianzaSinDato
	}

	// El escenario se DERIVA de la señal (A19 · §7.0.2): hay corrida nuestra ⇒ s1; hay
	// api_request sin corrida nuestra ⇒ s2-instrumentado. Un arnés no puede declararlo.
	if ev.CorridaID != "" {
		ev.Escenario = domain.EscenarioS1
	} else {
		ev.Escenario = domain.EscenarioS2Instrumentado
	}

	return ev, d, nil
}

// MapearPuntoMetrica arma el evento canónico de un punto de métrica. El canal de métricas es
// SECUNDARIO: `/v1/logs` ya trae el dinero por request con más detalle. Se mapea igual
// porque un runtime que solo exporte métricas sigue siendo medible, degradado pero honesto.
func MapearPuntoMetrica(p PuntoMetrica, perfil domain.PerfilRuntime, ahora time.Time) (domain.EventoTelemetria, Descartes, error) {
	var d Descartes
	for k := range p.Attrs {
		if allowlistOTLP[k] {
			continue
		}
		d.FueraDeLista++
		for _, n := range negadosDuros {
			if k == n {
				d.Identidad++
			}
		}
	}

	sesion := p.Attrs.Texto("session.id")
	if sesion == "" {
		return domain.EventoTelemetria{}, d, domain.ErrEventoSinSesion
	}
	if !p.TemporalidadSoportada || !p.Legible {
		// Ni se adivina un delta desde un acumulado, ni se cuenta un valor ilegible como 0.
		// El receptor lo suma a `salud.temporalidad_no_soportada`.
		return domain.EventoTelemetria{}, d, ErrEventoIgnorado
	}

	ev := domain.EventoTelemetria{
		LlaveJoin: domain.LlaveJoin{
			ArnesID:       p.Attrs.Texto("arnesia.arnes"),
			InstalacionID: p.Attrs.Texto("arnesia.instalacion"),
			CajaID:        p.Attrs.Texto("arnesia.caja"),
			SesionID:      sesion,
			TurnoID:       p.Attrs.Texto("prompt.id"),
			CorridaID:     p.Attrs.Texto("arnesia.corrida"),
		},
		Emisor:           domain.EmisorOTLP,
		Runtime:          perfil.Runtime,
		AdaptadorVersion: perfil.AdaptadorVersion,
		TSRecibido:       ahora.UTC(),
		TSEmisor:         p.TS,
		// 🔴 NO es `api_request`: es el canal SECUNDARIO, y su costo es EL MISMO que el del
		// canal primario. Marcarlo `api_request` haría que `SUM(costo)` contara el gasto dos
		// veces — con los dos exportadores encendidos, que es la configuración que este
		// mismo módulo prescribe para S1.
		TipoEvento:  domain.EventoMetrica,
		Modelo:      p.Attrs.Texto("model"),
		Aritmetica:  perfil.Aritmetica,
		Acumulacion: perfil.Acumulacion,
	}
	if ev.ArnesID != "" {
		ev.Atribucion = domain.ConfianzaExacta
	} else {
		ev.Atribucion = domain.ConfianzaSinDato
	}
	if ev.CorridaID != "" {
		ev.Escenario = domain.EscenarioS1
	} else {
		ev.Escenario = domain.EscenarioS2Instrumentado
	}

	// `claude_code.token.usage` desagrega por `type`; `claude_code.cost.usage` trae USD.
	valor := p.ValorDecimal
	if !p.EsDecimal {
		valor = float64(p.Valor)
	}
	switch p.Metrica {
	case "claude_code.token.usage":
		n := int64(valor)
		switch strings.ToLower(p.Attrs.Texto("type")) {
		case "input":
			ev.Tokens.Entrada = &n
		case "output":
			ev.Tokens.Salida = &n
		case "cacheread", "cache_read":
			ev.Tokens.CacheLectura = &n
		case "cachecreation", "cache_creation":
			// Misma razón que en el log record: la métrica tampoco dice el tramo.
			ev.Tokens.CacheEscrituraSinTier = &n
		default:
			return domain.EventoTelemetria{}, d, ErrEventoIgnorado
		}
	case "claude_code.cost.usage":
		micros := int64(valor*1e6 + 0.5)
		ev.CostoReportadoMicros = &micros
	default:
		return domain.EventoTelemetria{}, d, ErrEventoIgnorado
	}
	return ev, d, nil
}

// enteroOpcional devuelve un puntero al valor, o **nil** cuando el atributo no está o no es
// legible. Es la mitad del contrato «no aplica ≠ 0» que vive en la ingesta: devolver un 0
// acá lo haría indistinguible de una medición de cero.
func enteroOpcional(a Atributos, k string) *int64 {
	v, ok := a.Entero(k)
	if !ok {
		return nil
	}
	return &v
}

func resultadoDeDecision(d string) domain.Resultado {
	switch strings.ToLower(d) {
	case "accept", "allow":
		return domain.ResultadoOK
	case "reject", "deny":
		return domain.ResultadoRechazado
	default:
		return ""
	}
}
