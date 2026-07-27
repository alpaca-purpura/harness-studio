package domain

import (
	"strings"
	"testing"
	"time"
)

func ventanaBase(turnos []TurnoUnido, c ContextoDeteccion) Ventana {
	var total int64
	for _, t := range turnos {
		if t.CostoReportadoMicros != nil && t.Atribucion != ConfianzaSinDato {
			total += *t.CostoReportadoMicros
		}
	}
	return Ventana{
		Desde:   time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		Hasta:   time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		ArnesID: "vitalia", Turnos: turnos, TotalMicros: total, Contexto: c,
		Precios: preciosDeTest(turnos),
	}
}

// preciosDeTest le da precio a los modelos de la ventana, como hace el servicio (D26.1). Las
// tarifas son redondas a propósito —1 micro por token de escritura a 5 min, 1,6 a 1 h— para
// que el assert diga qué se está probando y no dependa del catálogo real.
//
// Un modelo que un test NO quiera cotizar simplemente no entra acá: la ausencia es el caso
// «no está en el catálogo», y el detector la declara en vez de cobrarla a cero.
func preciosDeTest(turnos []TurnoUnido) map[string]PrecioModelo {
	tarifa := func(v float64) *float64 { return &v }
	out := map[string]PrecioModelo{}
	for _, t := range turnos {
		if t.Modelo == "" {
			continue
		}
		out[t.Modelo] = PrecioModelo{
			ModeloCanonico:   t.Modelo,
			Entrada:          tarifa(0.0000008),
			Salida:           tarifa(0.000004),
			CacheLectura:     tarifa(0.0000001),
			CacheEscritura5m: tarifa(0.000001),
			CacheEscritura1h: tarifa(0.0000016),
		}
	}
	return out
}

func turno(id string, mod func(*TurnoUnido)) TurnoUnido {
	// `Modelo` viene con default porque desde D26.1 **cotizar exige saber de qué modelo se
	// habla**: un turno sin modelo no tiene tarifa, y eso es un caso de borde propio, no el
	// caso base.
	t := TurnoUnido{
		SesionID: "s-1", TurnoID: id, ArnesID: "vitalia",
		Modelo: "claude-haiku-4-5", Atribucion: ConfianzaExacta,
	}
	if mod != nil {
		mod(&t)
	}
	return t
}

func ptr(v int64) *int64 { return &v }

// TestDetectorQueNoAplicaTraeMotivo — con un contexto VACÍO los seis se apagan, y **los seis
// dicen por qué**. Un detector apagado sin razón es un gap escondido; uno que devuelve cero
// porque no pudo correr es una mentira.
func TestDetectorQueNoAplicaTraeMotivo(t *testing.T) {
	ds := DetectoresMVP()
	if len(ds) != 6 {
		t.Fatalf("el MVP son 6 detectores, hay %d", len(ds))
	}
	vacio := ContextoDeteccion{}
	for _, d := range ds {
		ap := d.Aplica(vacio)
		if ap.Aplica {
			t.Errorf("%s: con contexto vacío no puede aplicar", d.ID())
		}
		if strings.TrimSpace(ap.Motivo) == "" {
			t.Errorf("%s: Motivo es OBLIGATORIO cuando no aplica", d.ID())
		}
	}
	// ── control positivo: con la señal completa, la mayoría SÍ aplica ──
	completo := ContextoDeteccion{
		Runtime: "claude-code", Escenario: EscenarioS1,
		TieneCosto: true, TieneSplitTTL: true, TieneSenalProceso: true,
		TieneGateHumano: true, TieneEventoRotacion: true, ModelosDistintos: 2,
		CatalogoDisponible: true,
	}
	aplicaron := 0
	for _, d := range ds {
		if d.Aplica(completo).Aplica {
			aplicaron++
		}
	}
	if aplicaron != 6 {
		t.Fatalf("con señal completa (s1) tienen que aplicar los 6, aplicaron %d — "+
			"si no, el assert de arriba podría estar pasando porque NINGUNO aplica nunca", aplicaron)
	}
	// Y los motivos son DISTINTOS entre sí: seis copias del mismo texto no explicarían nada.
	motivos := map[string]bool{}
	for _, d := range ds {
		motivos[d.Aplica(vacio).Motivo] = true
	}
	if len(motivos) < 4 {
		t.Errorf("los motivos deben explicar cada caso, no repetir uno genérico: %d distintos", len(motivos))
	}
}

// TestS2DegradadoApagaLosDetectoresDeDinero — solo hook ⇒ B4, B6 y B3 se apagan CON MOTIVO.
// Nunca en 0: un 0 se leería como «no hay problema».
func TestS2DegradadoApagaLosDetectoresDeDinero(t *testing.T) {
	c := ContextoDeteccion{
		Runtime: "claude-code", Escenario: EscenarioS2Degradado,
		TieneCosto: false, TieneSenalProceso: true,
	}
	apagados := map[DetectorID]bool{DetB4: true, DetB6: true, DetB3: true}
	for _, d := range DetectoresMVP() {
		ap := d.Aplica(c)
		if apagados[d.ID()] {
			if ap.Aplica {
				t.Errorf("%s tiene que apagarse sin señal de dinero", d.ID())
			}
			if ap.Motivo == "" {
				t.Errorf("%s se apagó sin decir por qué", d.ID())
			}
		}
	}
	// ── control positivo: P1 SÍ corre en degradado — es el único que puede ──
	var p1 Detector
	for _, d := range DetectoresMVP() {
		if d.ID() == DetP1 {
			p1 = d
		}
	}
	ap := p1.Aplica(c)
	if !ap.Aplica {
		t.Fatal("P1 es el detector que sobrevive al modo degradado: tiene que aplicar")
	}
	if !ap.Parcial {
		t.Error("sin gate humano, P1 corre PARCIAL y lo dice")
	}
}

// TestS2InstrumentadoTieneDineroYNoTieneSplit — el matiz que la verificación H9 dejó: con el
// bloque de telemetría encendido fuera de ArnesIA **sí hay dinero**; lo que falta es el
// desglose por vencimiento. B1 sigue sin aplicar, **pero por otra razón**, y el motivo lo
// dice.
func TestS2InstrumentadoTieneDineroYNoTieneSplit(t *testing.T) {
	c := ContextoDeteccion{
		Runtime: "claude-code", Escenario: EscenarioS2Instrumentado,
		TieneCosto: true, TieneSplitTTL: false, TieneSenalProceso: true, ModelosDistintos: 2,
		CatalogoDisponible: true,
	}
	encendidos := map[DetectorID]bool{DetB4: true, DetB6: true, DetB3: true}
	for _, d := range DetectoresMVP() {
		ap := d.Aplica(c)
		if encendidos[d.ID()] && !ap.Aplica {
			t.Errorf("%s tiene que aplicar: con el bloque de telemetría hay dinero (motivo dado: %q)",
				d.ID(), ap.Motivo)
		}
		if d.ID() == DetB1 {
			if ap.Aplica {
				t.Error("B1 no puede aplicar sin el desglose por vencimiento")
			}
			// **El motivo NO es «corrió fuera de ArnesIA»**: es la falta del split. Escrito
			// así, el detector se enciende solo el día que el split llegue por otro canal.
			if strings.Contains(strings.ToLower(ap.Motivo), "fuera de arnesia") {
				t.Errorf("el motivo de B1 confunde los dos casos: %q", ap.Motivo)
			}
			if !strings.Contains(strings.ToLower(ap.Motivo), "desglose") &&
				!strings.Contains(strings.ToLower(ap.Motivo), "vencimiento") {
				t.Errorf("el motivo de B1 tiene que nombrar lo que falta: %q", ap.Motivo)
			}
		}
	}
	// Y el día que el split llegue, B1 se enciende SIN tocar código.
	c.TieneSplitTTL = true
	for _, d := range DetectoresMVP() {
		if d.ID() == DetB1 && !d.Aplica(c).Aplica {
			t.Error("con el split presente B1 tiene que encenderse solo: la regla es de la ventana, no del escenario")
		}
	}
}

// TestB1BreakEvenTTL — B1 cita su umbral ENTERO, y es álgebra, no una medición de esta
// corrida (RF-250).
func TestB1BreakEvenTTL(t *testing.T) {
	var b1 Detector
	for _, d := range DetectoresMVP() {
		if d.ID() == DetB1 {
			b1 = d
		}
	}
	// **DOS escrituras del mismo contexto**: eso es lo que el vencimiento largo evita. Con una
	// sola no hay re-warm y no hay punto — es la condición de existencia de B1 desde D26.1.
	ctx := ContextoDeteccion{
		Runtime: "claude-code", TieneCosto: true, TieneSplitTTL: true,
		CatalogoDisponible: true,
	}
	turnoQueEscribe := func(id string) TurnoUnido {
		return turno(id, func(t *TurnoUnido) {
			t.Tokens.CacheEscritura5m = ptr(10_000)
			t.Tokens.CacheEscritura1h = ptr(0)
			t.CostoReportadoMicros = ptr(20_000)
		})
	}
	// Control negativo primero: UNA escritura no es un hallazgo chico, no es un hallazgo.
	if p := b1.Evaluar(ventanaBase([]TurnoUnido{turnoQueEscribe("t-1")}, ctx)); len(p) != 0 {
		t.Errorf("con una sola escritura no hay re-warm que evitar: %+v", p)
	}

	v := ventanaBase([]TurnoUnido{turnoQueEscribe("t-1"), turnoQueEscribe("t-2")}, ctx)
	puntos := b1.Evaluar(v)
	if len(puntos) != 1 {
		t.Fatalf("se esperaba 1 punto, hay %d", len(puntos))
	}
	p := puntos[0]
	// El umbral cita la desigualdad completa.
	for _, frag := range []string{"2", "1,25", "0,1", "39,47"} {
		if !strings.Contains(p.Umbral, frag) {
			t.Errorf("el umbral debe citar la desigualdad entera; falta %q en %q", frag, p.Umbral)
		}
	}
	// El contrafactual NO es «gastaste X»: es «con el cambio habrías gastado Y» — y desde
	// D26.1 es un mundo MÁS BARATO, cotizado con el catálogo.
	//
	// La aritmética que el assert fija: 2 × 10 000 tokens a 5 min = 20 000 micros; con 1 h se
	// escribe UNA sola vez (10 000 tokens) a 1,6 = 16 000 micros. Ahorro 4 000.
	if p.GastoMicros != 20_000 {
		t.Errorf("gasto = %d micros, se esperaba 20000 (2 × 10 000 tokens × 1 micro)", p.GastoMicros)
	}
	if p.ContrafactualMicros != 16_000 {
		t.Errorf("contrafactual = %d micros, se esperaba 16000 (una sola escritura a 1 h)",
			p.ContrafactualMicros)
	}
	if p.DiferenciaMicros <= 0 {
		t.Errorf("el ahorro de B1 tiene que ser POSITIVO: %d — un mundo alternativo más caro "+
			"no es una recomendación", p.DiferenciaMicros)
	}
	if p.Fix == "" {
		t.Error("regla A4: un punto sin fix concreto no se muestra")
	}
	if p.ScoreVersion != ScoreVersionMVP {
		t.Errorf("score_version = %d, se esperaba %d", p.ScoreVersion, ScoreVersionMVP)
	}
}

// TestSesgoTieneDireccion — RF-252: `direccion_sesgo` existe, **no es neutro** y el texto del
// sesgo no está vacío. Un sesgo sin dirección no se puede evaluar.
func TestSesgoTieneDireccion(t *testing.T) {
	c := ContextoDeteccion{
		Runtime: "claude-code", Escenario: EscenarioS1, TieneCosto: true, TieneSplitTTL: true,
		TieneSenalProceso: true, TieneGateHumano: true, TieneEventoRotacion: true, ModelosDistintos: 2,
	}
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.CajaID = "paso-3"
			t.CostoReportadoMicros = ptr(10_000)
			t.Modelo = "claude-haiku-4-5"
			t.Rotaciones = 1
			t.TieneProceso = true
			t.Tokens.CacheEscritura5m = ptr(5_000)
			t.Tokens.CacheEscritura1h = ptr(0)
			t.Resultados = []Resultado{ResultadoRechazado}
		}),
		turno("t-2", func(t *TurnoUnido) {
			t.CajaID = "paso-3"
			t.CostoReportadoMicros = ptr(1_000)
			t.Modelo = "claude-sonnet-4-5"
			t.Tokens.CacheEscritura5m = ptr(2_000)
			t.TieneProceso = true
		}),
	}, c)

	total := 0
	for _, d := range DetectoresMVP() {
		if !d.Aplica(c).Aplica {
			continue
		}
		for _, p := range d.Evaluar(v) {
			total++
			if p.DireccionSesgo != "subestima" && p.DireccionSesgo != "sobreestima" {
				t.Errorf("%s: direccion_sesgo = %q — tiene que ser subestima o sobreestima, jamás neutro",
					p.Detector, p.DireccionSesgo)
			}
			if strings.TrimSpace(p.Sesgo) == "" {
				t.Errorf("%s: el sesgo se declara con texto, no solo con una etiqueta", p.Detector)
			}
			// Regla A4 completa: sin las cinco cosas el punto no se muestra. El «por qué lo
			// creemos» es el umbral **o** el patrón (RF-250): P1 no se decide por una
			// desigualdad, y pintarle una inventada sería fabricarle rigor. Lo que no se
			// permite es que falten los dos.
			if p.Umbral == "" && p.Patron == "" {
				t.Errorf("%s: sin umbral citado ni patrón", p.Detector)
			}
			if p.Umbral != "" && p.Patron != "" {
				t.Errorf("%s: umbral y patrón a la vez — la tarjeta pinta uno solo", p.Detector)
			}
			if p.Fix == "" {
				t.Errorf("%s: sin fix concreto", p.Detector)
			}
			if p.Titulo == "" {
				t.Errorf("%s: sin titular", p.Detector)
			}
		}
	}
	if total == 0 {
		t.Fatal("ningún punto evaluado: el test no comparó nada")
	}
	t.Logf("puntos de mejora evaluados con sesgo declarado: %d", total)
}

// TestP1CajaQueSeRechaza — el desperdicio de las corridas rechazadas se cuantifica en dinero,
// no en un «⚠ hay rechazos».
func TestP1CajaQueSeRechaza(t *testing.T) {
	var p1 Detector
	for _, d := range DetectoresMVP() {
		if d.ID() == DetP1 {
			p1 = d
		}
	}
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.CajaID = "paso-11"
			t.CostoReportadoMicros = ptr(3_000)
			t.Resultados = []Resultado{ResultadoRechazado}
			t.TieneProceso = true
		}),
		turno("t-2", func(t *TurnoUnido) {
			t.CajaID = "paso-11"
			t.CostoReportadoMicros = ptr(2_000)
			t.Resultados = []Resultado{ResultadoRechazado}
			t.TieneProceso = true
		}),
		turno("t-3", func(t *TurnoUnido) {
			t.CajaID = "paso-11"
			t.CostoReportadoMicros = ptr(1_000)
			t.Resultados = []Resultado{ResultadoOK}
			t.TieneProceso = true
		}),
	}, ContextoDeteccion{TieneSenalProceso: true, TieneGateHumano: true})

	puntos := p1.Evaluar(v)
	if len(puntos) != 1 {
		t.Fatalf("se esperaba 1 punto, hay %d", len(puntos))
	}
	p := puntos[0]
	if p.GastoMicros != 5_000 {
		t.Errorf("el desperdicio son los 2 rechazados (3000+2000): %d", p.GastoMicros)
	}
	// El contrafactual de P1 es CERO, y es el único de los seis donde eso es correcto: una
	// corrida rechazada literalmente no entregó nada.
	if p.ContrafactualMicros != 0 {
		t.Errorf("una corrida rechazada no entregó nada: contrafactual = %d", p.ContrafactualMicros)
	}
	if p.CorridasUsadas != 2 || p.CorridasTotales != 3 {
		t.Errorf("la confianza lleva denominador: %d de %d", p.CorridasUsadas, p.CorridasTotales)
	}
}

// TestP1ParcialSinGate — con señal de proceso pero sin gate humano, P1 corre y **lo declara
// parcial**. No es un visto bueno liso.
func TestP1ParcialSinGate(t *testing.T) {
	var p1 Detector
	for _, d := range DetectoresMVP() {
		if d.ID() == DetP1 {
			p1 = d
		}
	}
	ap := p1.Aplica(ContextoDeteccion{TieneSenalProceso: true, TieneGateHumano: false})
	if !ap.Aplica {
		t.Fatal("con señal de proceso, P1 corre")
	}
	if !ap.Parcial {
		t.Fatal("sin gate humano corre PARCIAL, y eso viaja explícito")
	}
	if ap.Motivo == "" {
		t.Error("la cobertura parcial se explica, no se insinúa")
	}
	// Control positivo: CON gate humano, no es parcial.
	if ap2 := p1.Aplica(ContextoDeteccion{TieneSenalProceso: true, TieneGateHumano: true}); ap2.Parcial {
		t.Error("con gate humano P1 corre completo")
	}
}

// TestB2CostoDeLaRotacion / TestB6SesionAbandonada / TestB3CambioDeModelo — los tres que
// faltan, cada uno con su señal mínima.
func TestB2CostoDeLaRotacion(t *testing.T) {
	d := detectorPorID(t, DetB2)
	if ap := d.Aplica(ContextoDeteccion{TieneCosto: true}); ap.Aplica {
		t.Error("sin rotaciones registradas, B2 no aplica")
	} else if !strings.Contains(strings.ToLower(ap.Motivo), "rotaci") {
		t.Errorf("el motivo debe nombrar la rotación: %q", ap.Motivo)
	}
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.Rotaciones = 2
			t.CostoReportadoMicros = ptr(4_000)
		}),
	}, ContextoDeteccion{TieneCosto: true, TieneEventoRotacion: true})
	puntos := d.Evaluar(v)
	if len(puntos) != 1 || puntos[0].GastoMicros != 4_000 {
		t.Fatalf("B2 debe cuantificar la rotación: %+v", puntos)
	}
	if puntos[0].DireccionSesgo != "sobreestima" {
		t.Errorf("el sesgo de B2 sobreestima el ahorro: %q", puntos[0].DireccionSesgo)
	}
}

func TestB6SesionAbandonada(t *testing.T) {
	d := detectorPorID(t, DetB6)
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.SesionID = "s-abandonada"
			t.CostoReportadoMicros = ptr(7_000)
			t.TieneProceso = false // nadie cerró nada
		}),
		turno("t-2", func(t *TurnoUnido) {
			t.SesionID = "s-sana"
			t.CostoReportadoMicros = ptr(1_000)
			t.TieneProceso = true
		}),
	}, ContextoDeteccion{TieneCosto: true})
	puntos := d.Evaluar(v)
	if len(puntos) != 1 {
		t.Fatalf("se esperaba 1 punto, hay %d", len(puntos))
	}
	if puntos[0].GastoMicros != 7_000 {
		t.Errorf("solo cuenta la sesión abandonada: %d", puntos[0].GastoMicros)
	}
}

func TestB3CambioDeModelo(t *testing.T) {
	d := detectorPorID(t, DetB3)
	if ap := d.Aplica(ContextoDeteccion{
		TieneCosto: true, ModelosDistintos: 1,
		CatalogoDisponible: true,
	}); ap.Aplica {
		t.Error("con un solo modelo no hubo cambio que invalidara el cache")
	}
	// Sin catálogo NO corre: el número saldría en tokens y se mostraría como dinero (D26.1).
	if ap := d.Aplica(ContextoDeteccion{TieneCosto: true, ModelosDistintos: 2}); ap.Aplica {
		t.Error("sin catálogo B3 no puede poner un monto: cotizar es la mitad del hallazgo")
	}
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.Modelo = "claude-haiku-4-5"
			t.CostoReportadoMicros = ptr(1_000)
		}),
		turno("t-2", func(t *TurnoUnido) {
			t.Modelo = "claude-sonnet-4-5"
			t.Tokens.CacheEscritura5m = ptr(9_000)
			t.CostoReportadoMicros = ptr(5_000)
		}),
	}, ContextoDeteccion{TieneCosto: true, ModelosDistintos: 2, CatalogoDisponible: true})
	puntos := d.Evaluar(v)
	// 9 000 tokens de re-escritura **cotizados**, no contados: × 1 micro/token = 9 000 micros.
	// El número coincide con el conteo viejo por la tarifa redonda del fixture, y esa
	// coincidencia es a propósito: hace visible que lo que cambió es la UNIDAD, no la magnitud.
	if len(puntos) != 1 || puntos[0].GastoMicros != 9_000 {
		t.Fatalf("B3 cuantifica la re-escritura tras el cambio: %+v", puntos)
	}
	if !strings.Contains(puntos[0].Calculo, "catálogo") {
		t.Errorf("el cálculo tiene que decir que el monto se cotizó: %q", puntos[0].Calculo)
	}
}

// TestNingunPuntoSinContrafactualNiFix — la regla A4 aplicada a TODOS los puntos que los seis
// puedan producir: sin las cinco cosas, no se muestra.
func TestNingunPuntoSinContrafactualNiFix(t *testing.T) {
	c := ContextoDeteccion{
		Runtime: "claude-code", Escenario: EscenarioS1, TieneCosto: true, TieneSplitTTL: true,
		TieneSenalProceso: true, TieneGateHumano: true, TieneEventoRotacion: true, ModelosDistintos: 2,
	}
	v := ventanaBase([]TurnoUnido{
		turno("t-1", func(t *TurnoUnido) {
			t.SesionID = "s-a"
			t.CajaID = "paso-3"
			t.Modelo = "claude-haiku-4-5"
			t.CostoReportadoMicros = ptr(10_000)
			t.Rotaciones = 1
			t.Tokens.CacheEscritura5m = ptr(4_000)
			t.Tokens.CacheEscritura1h = ptr(0)
			t.Resultados = []Resultado{ResultadoReintento}
		}),
		turno("t-2", func(t *TurnoUnido) {
			t.SesionID = "s-b"
			t.Modelo = "claude-sonnet-4-5"
			t.CostoReportadoMicros = ptr(500)
			t.Tokens.CacheEscritura5m = ptr(1_000)
		}),
	}, c)
	vistos := 0
	for _, d := range DetectoresMVP() {
		if !d.Aplica(c).Aplica {
			continue
		}
		for _, p := range d.Evaluar(v) {
			vistos++
			if (p.Umbral == "" && p.Patron == "") || p.Sesgo == "" || p.DireccionSesgo == "" ||
				p.Fix == "" || p.Titulo == "" {
				t.Errorf("%s: punto incompleto (A4 pide número, contrafactual, umbral o patrón, "+
					"sesgo y UN fix): %+v", p.Detector, p)
			}
			if p.Confianza == "" {
				t.Errorf("%s: un número sin su confianza no se muestra", p.Detector)
			}
		}
	}
	if vistos == 0 {
		t.Fatal("ningún punto producido: el test no verificó nada")
	}
}

func detectorPorID(t *testing.T, id DetectorID) Detector {
	t.Helper()
	for _, d := range DetectoresMVP() {
		if d.ID() == id {
			return d
		}
	}
	t.Fatalf("detector %q no está en el MVP", id)
	return nil
}

// TestElMontoDelCacheSaleDeLaTarifaNoDelConteo — el candado del defecto M2 (D26.1).
//
// B1 y B3 sumaban **conteos de tokens** en campos `micros` y la tarjeta mostraba ese número
// como dinero. La forma de probar que ahora se COTIZA es cambiar la tarifa dejando los tokens
// iguales: si el monto no se mueve, el número no es dinero.
func TestElMontoDelCacheSaleDeLaTarifaNoDelConteo(t *testing.T) {
	b1 := detectorPorID(t, DetB1)
	ctx := ContextoDeteccion{
		Runtime: "claude-code", TieneCosto: true, TieneSplitTTL: true,
		CatalogoDisponible: true,
	}
	turnos := []TurnoUnido{
		turno("t-1", func(t *TurnoUnido) { t.Tokens.CacheEscritura5m = ptr(10_000) }),
		turno("t-2", func(t *TurnoUnido) { t.Tokens.CacheEscritura5m = ptr(10_000) }),
	}

	conTarifa := func(mult float64) int64 {
		v := ventanaBase(turnos, ctx)
		for m, p := range v.Precios {
			e5 := 0.000001 * mult
			e1 := 0.0000016 * mult
			p.CacheEscritura5m, p.CacheEscritura1h = &e5, &e1
			v.Precios[m] = p
		}
		puntos := b1.Evaluar(v)
		if len(puntos) != 1 {
			t.Fatalf("mult %.1f: se esperaba 1 punto, hay %d", mult, len(puntos))
		}
		return puntos[0].GastoMicros
	}

	base, doble := conTarifa(1), conTarifa(2)
	if base <= 0 {
		t.Fatal("el monto base es 0: el test no comparó nada")
	}
	if doble != base*2 {
		t.Errorf("con la tarifa al doble el monto tiene que duplicarse: %d vs %d — si no se "+
			"mueve, el número son tokens disfrazados de dinero", doble, base)
	}

	// Control negativo: un modelo fuera del catálogo **no vale cero**. Sin tarifa no hay
	// hallazgo cotizable, y el detector no inventa uno.
	v := ventanaBase(turnos, ctx)
	v.Precios = nil
	if puntos := b1.Evaluar(v); len(puntos) != 0 {
		t.Errorf("sin catálogo el detector no puede poner un monto y no lo inventa: %+v", puntos)
	}
}
