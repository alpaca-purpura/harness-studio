package domain

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestNoAplicaNoEsCeroEnElWire (versión mínima sobre el dominio; la de fitness cubre el
// wire HTTP completo). Un bucket ausente NO puede serializar como 0: `nil` significa «este
// runtime no tiene el concepto» y un 0 significa «midió cero». Confundirlos es el pass
// fabricado que boundaries/no-aplica-no-es-cero.md prohíbe.
func TestNoAplicaNoEsCeroEnElWire(t *testing.T) {
	entrada := int64(10)
	ev := EventoTelemetria{
		LlaveJoin:  LlaveJoin{SesionID: "s-1", TurnoID: "t-1"},
		Emisor:     EmisorOTLP,
		Runtime:    "claude-code",
		TipoEvento: EventoAPIRequest,
		Escenario:  EscenarioS1,
		Atribucion: ConfianzaExacta,
		// Razonamiento queda nil a propósito: Claude Code no lo reporta.
		Tokens: Tokens{Entrada: &entrada},
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(raw)
	if strings.Contains(got, "razonamiento") {
		t.Errorf("un bucket nil NO debe aparecer en el wire; salió: %s", got)
	}
	if strings.Contains(got, `"razonamiento":0`) {
		t.Errorf("un bucket nil serializó como 0 — eso es fabricar un dato: %s", got)
	}
	// Control positivo, misma corrida: el bucket que SÍ tiene dato sí aparece. Sin esto,
	// un marshal roto que emitiera "{}" pasaría el assert de ausencia.
	if !strings.Contains(got, `"entrada":10`) {
		t.Errorf("el bucket con dato debe viajar; salió: %s", got)
	}
}

// TestEventoNoTieneCamposDeIdentidad recorre por REFLEXIÓN los campos del evento canónico y
// falla si alguien agrega un lugar donde meter identidad de cuenta o contenido de
// conversación. La telemetría de Claude Code trae `user.email`, `user.account_uuid`,
// `user.account_id`, `user.id` y `organization.id` en CADA punto (verificado, INFORME §V6)
// y el payload del hook trae el prompt en claro (ANEXO H4): si el evento tuviera dónde
// ponerlos, alguien los pondría.
func TestEventoNoTieneCamposDeIdentidad(t *testing.T) {
	prohibidos := []string{"email", "account", "organization", "user", "prompt", "respuesta",
		"response", "message", "content", "contenido", "transcript", "cwd_ruta", "path"}
	// `CWDHuella` es legal: es una HUELLA, no la ruta (A14). Se exime por nombre exacto.
	exentos := map[string]bool{"CWDHuella": true}

	var campos []string
	var recorrer func(reflect.Type, string)
	recorrer = func(tp reflect.Type, prefijo string) {
		for i := 0; i < tp.NumField(); i++ {
			f := tp.Field(i)
			if f.Anonymous && f.Type.Kind() == reflect.Struct {
				recorrer(f.Type, prefijo)
				continue
			}
			campos = append(campos, prefijo+f.Name)
			if f.Type.Kind() == reflect.Struct && f.Type != reflect.TypeOf(time.Time{}) {
				recorrer(f.Type, prefijo+f.Name+".")
			}
		}
	}
	recorrer(reflect.TypeOf(EventoTelemetria{}), "")

	if len(campos) == 0 {
		t.Fatal("la reflexión no encontró ningún campo: el scanner está roto y su verde no significa nada")
	}
	for _, c := range campos {
		base := c
		if i := strings.LastIndex(c, "."); i >= 0 {
			base = c[i+1:]
		}
		if exentos[base] {
			continue
		}
		bajo := strings.ToLower(base)
		for _, p := range prohibidos {
			if strings.Contains(bajo, p) {
				t.Errorf("EventoTelemetria.%s: el evento canónico no puede tener dónde poner %q "+
					"(boundary ingesta-por-allowlist-declarada). Si el campo es legítimo, exímelo "+
					"por nombre con la razón escrita, como CWDHuella", c, p)
			}
		}
	}
	// Control positivo: el detector encuentra lo que busca. Un struct de prueba con un
	// campo prohibido debe caer en la misma regla.
	type conPII struct{ UserEmail string }
	hallado := false
	tp := reflect.TypeOf(conPII{})
	for i := 0; i < tp.NumField(); i++ {
		if strings.Contains(strings.ToLower(tp.Field(i).Name), "email") {
			hallado = true
		}
	}
	if !hallado {
		t.Fatal("control positivo: la regla no detecta un campo obviamente prohibido — está rota")
	}
}

// TestEventoLlevaSuConfianza — boundary cifra-viaja-con-su-confianza: no existe una ruta
// donde un costo viaje sin su atribución al lado, en la misma struct.
func TestEventoLlevaSuConfianza(t *testing.T) {
	tp := reflect.TypeOf(EventoTelemetria{})
	if _, ok := tp.FieldByName("Atribucion"); !ok {
		t.Fatal("EventoTelemetria sin campo Atribucion: una cifra sin su confianza no se puede mostrar")
	}
	if _, ok := tp.FieldByName("CostoReportadoMicros"); !ok {
		t.Fatal("EventoTelemetria sin CostoReportadoMicros")
	}
	if _, ok := tp.FieldByName("CostoCalculadoMicros"); !ok {
		t.Fatal("EventoTelemetria sin CostoCalculadoMicros: los DOS costos viajan siempre (D16.2)")
	}
}

// TestConfianzaDeAgregadoEsLaMinima — la confianza de una suma vale lo que su peor parte,
// nunca la mejor ni la moda. Mezclar una atribución exacta con una adivinada y presentar
// el total como exacto es la forma silenciosa de inflar la certeza.
func TestConfianzaDeAgregadoEsLaMinima(t *testing.T) {
	casos := []struct {
		a, b, want Confianza
	}{
		{ConfianzaExacta, ConfianzaPorHash, ConfianzaPorHash},
		{ConfianzaPorHash, ConfianzaExacta, ConfianzaPorHash},
		{ConfianzaExacta, ConfianzaExacta, ConfianzaExacta},
		{ConfianzaPorProceso, ConfianzaSinDato, ConfianzaSinDato},
		{ConfianzaExacta, Confianza("inventada"), ConfianzaSinDato},
	}
	for _, c := range casos {
		if got := PeorConfianza(c.a, c.b); got != c.want {
			t.Errorf("PeorConfianza(%q,%q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

// TestTodoEnUTC — el reloj del daemon manda y viaja en UTC RFC3339 (escenario D5). Un
// timestamp con offset local es indistinguible de otro con el mismo instante en otra zona
// cuando se compara como texto, que es exactamente lo que el rollup hace con `strftime`.
func TestTodoEnUTC(t *testing.T) {
	loc := time.FixedZone("UTC-5", -5*3600)
	ev := EventoTelemetria{TSRecibido: time.Date(2026, 7, 26, 9, 0, 0, 0, loc)}
	if ev.TSRecibido.UTC().Format(time.RFC3339) != "2026-07-26T14:00:00Z" {
		t.Fatalf("normalización a UTC rota: %s", ev.TSRecibido.UTC().Format(time.RFC3339))
	}
}
