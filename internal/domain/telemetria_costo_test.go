package domain

import "testing"

func i64(v int64) *int64     { return &v }
func f64(v float64) *float64 { return &v }

// precioHaiku45 son las tarifas REALES de claude-haiku-4-5 (catálogo embebido, rev fijado de
// LiteLLM). No son inventadas: se cruzan contra el `cost_usd_micros` medido en vivo, ver
// TestParidadConElCostoReportadoReal.
func precioHaiku45() PrecioModelo {
	return PrecioModelo{
		ModeloCanonico:   "claude-haiku-4-5",
		Proveedor:        "anthropic",
		Entrada:          f64(1e-6),
		Salida:           f64(5e-6),
		CacheLectura:     f64(1e-7),
		CacheEscritura5m: f64(1.25e-6),
		CacheEscritura1h: f64(2e-6),
	}
}

// TestCosteoCobraElCacheWrite — el bug langfuse#14249: olvidarse de cobrar la ESCRITURA de
// cache subestima ~28 %. El assert no es «el número es X»: es que el costo con el bucket es
// estrictamente mayor que sin él. Un cambio de tarifas no rompe el test; olvidarse de cobrar
// el bucket sí.
func TestCosteoCobraElCacheWrite(t *testing.T) {
	p := precioHaiku45()
	base := Tokens{Entrada: i64(1000), Salida: i64(500)}
	conWrite := Tokens{Entrada: i64(1000), Salida: i64(500), CacheEscritura1h: i64(8000)}

	sin := CalcularCosto(base, p, AritmeticaDisjunta)
	con := CalcularCosto(conWrite, p, AritmeticaDisjunta)

	if con.Micros <= sin.Micros {
		t.Fatalf("la escritura de cache no se cobró: con=%d micros, sin=%d", con.Micros, sin.Micros)
	}
	// Y se cobró con SU tarifa, no con la de entrada.
	esperado := sin.Micros + int64(8000*2e-6*1e6)
	if con.Micros != esperado {
		t.Errorf("cache write 1h cobrado con la tarifa equivocada: %d micros, se esperaba %d", con.Micros, esperado)
	}
	if !con.Completo {
		t.Errorf("con todas las tarifas presentes el costo es completo: %+v", con)
	}
}

// TestCosteoNoSumaBucketsQueSeSolapan — el bug langfuse#12306: con aritmética inclusiva,
// `cache_lectura` viene DENTRO de `entrada`, y no restarlo duplica el conteo (2×).
func TestCosteoNoSumaBucketsQueSeSolapan(t *testing.T) {
	p := precioHaiku45()
	// 10 000 de entrada, de los cuales 8 000 fueron leídos de cache.
	tk := Tokens{Entrada: i64(10_000), CacheLectura: i64(8_000)}

	disj := CalcularCosto(tk, p, AritmeticaDisjunta)
	incl := CalcularCosto(tk, p, AritmeticaInclusiva)

	if incl.Micros >= disj.Micros {
		t.Fatalf("con aritmética inclusiva hay que RESTAR el solapamiento: incl=%d, disj=%d", incl.Micros, disj.Micros)
	}
	// disjunta: 10 000 × 1e-6 + 8 000 × 1e-7 = 0,0108 USD = 10 800 micros
	if disj.Micros != 10_800 {
		t.Errorf("disjunta: %d micros, se esperaba 10800", disj.Micros)
	}
	// inclusiva: (10 000 − 8 000) × 1e-6 + 8 000 × 1e-7 = 0,0028 USD = 2 800 micros
	if incl.Micros != 2_800 {
		t.Errorf("inclusiva: %d micros, se esperaba 2800", incl.Micros)
	}
}

// TestCosteoInclusivaNoRegalaCredito — un emisor que reporta más cache_lectura que entrada
// está reportando algo incoherente. Restar hasta negativo daría un costo con signo, que se
// leería como un crédito. Se corta en 0: mentir a favor del usuario sigue siendo mentir.
func TestCosteoInclusivaNoRegalaCredito(t *testing.T) {
	p := precioHaiku45()
	tk := Tokens{Entrada: i64(100), CacheLectura: i64(9_000)}
	c := CalcularCosto(tk, p, AritmeticaInclusiva)
	if c.Micros < 0 {
		t.Fatalf("un costo negativo es un crédito inventado: %d micros", c.Micros)
	}
	// Control positivo: la lectura de cache SÍ se cobró.
	if c.Micros != int64(9_000*1e-7*1e6) {
		t.Errorf("%d micros, se esperaba solo la lectura de cache (900)", c.Micros)
	}
}

// TestCosteoNoAplanaLosTiers — el bug phoenix#14314: aplanar el tramo largo subestima justo
// los prompts caros, que son los que importan.
func TestCosteoNoAplanaLosTiers(t *testing.T) {
	base := PrecioModelo{
		ModeloCanonico:    "claude-sonnet-4-5",
		Entrada:           f64(3e-6),
		Salida:            f64(15e-6),
		UmbralContextoTok: i64(200_000),
		SobreUmbral: &PrecioModelo{
			Entrada: f64(6e-6),
			Salida:  f64(22.5e-6),
		},
	}
	corto := CalcularCosto(Tokens{Entrada: i64(100_000), Salida: i64(1_000)}, base, AritmeticaDisjunta)
	largo := CalcularCosto(Tokens{Entrada: i64(300_000), Salida: i64(1_000)}, base, AritmeticaDisjunta)

	// corto: 100 000 × 3e-6 + 1 000 × 15e-6 = 0,315 USD = 315 000 micros
	if corto.Micros != 315_000 {
		t.Errorf("bajo el umbral se cotiza con la tarifa base: %d micros, se esperaba 315000", corto.Micros)
	}
	// largo: 300 000 × 6e-6 + 1 000 × 22,5e-6 = 1,8225 USD = 1 822 500 micros
	if largo.Micros != 1_822_500 {
		t.Errorf("sobre el umbral se cotiza con el TRAMO, no con la base: %d micros, se esperaba 1822500", largo.Micros)
	}
	// El assert que mata el aplanado: si se aplanara a la tarifa base, el prompt largo
	// costaría 300 000 × 3e-6 + 1 000 × 15e-6 = 915 000 micros. Tiene que ser MÁS.
	if largo.Micros <= 915_000 {
		t.Errorf("el tramo largo se aplanó a la tarifa base: %d micros", largo.Micros)
	}
	// Control positivo del contraste: un precio SIN tramo no cambia con el tamaño.
	sinTramo := precioHaiku45()
	a := CalcularCosto(Tokens{Entrada: i64(100_000)}, sinTramo, AritmeticaDisjunta)
	b := CalcularCosto(Tokens{Entrada: i64(300_000)}, sinTramo, AritmeticaDisjunta)
	if b.Micros != a.Micros*3 {
		t.Errorf("sin tramo declarado, el costo es lineal: %d vs %d×3", b.Micros, a.Micros)
	}
}

// TestSinNingunCosto — la cuarta regla, la de honestidad: un bucket con tokens y SIN tarifa
// no se cobra a 0. Sale nombrado y el total se marca parcial.
func TestSinNingunCosto(t *testing.T) {
	// Un precio que no publica tarifa de razonamiento (el caso real de Anthropic).
	p := precioHaiku45()
	tk := Tokens{Entrada: i64(1_000), Razonamiento: i64(5_000)}
	c := CalcularCosto(tk, p, AritmeticaDisjunta)

	if c.Completo {
		t.Error("con un bucket sin tarifa el costo NO es completo")
	}
	if len(c.SinTarifa) != 1 || c.SinTarifa[0] != "razonamiento" {
		t.Errorf("el bucket sin tarifa debe salir NOMBRADO: %v", c.SinTarifa)
	}
	// Y no se cobró a 0 en silencio: lo que sí tenía tarifa sí se cobró (control positivo).
	if c.Micros != 1_000 {
		t.Errorf("los buckets con tarifa se cobran igual: %d micros, se esperaba 1000", c.Micros)
	}
	if c.SinNingunaTarifa {
		t.Error("hubo al menos un bucket cotizado: SinNingunaTarifa no corresponde")
	}

	// El caso extremo: NINGÚN bucket tiene tarifa. El caller debe mandar null al wire.
	vacio := CalcularCosto(Tokens{Entrada: i64(1_000)}, PrecioModelo{ModeloCanonico: "desconocido"}, AritmeticaDisjunta)
	if !vacio.SinNingunaTarifa {
		t.Error("sin ninguna tarifa aplicable, SinNingunaTarifa debe ser true — si no, el 0 viajaría como dato")
	}
	if vacio.Micros != 0 || vacio.Completo {
		t.Errorf("un costo sin tarifas es 0 micros pero INCOMPLETO: %+v", vacio)
	}
}

// TestCosteoCeroTokensEsCompleto — el contraste que el test anterior necesita para no ser
// ambiguo: un uso sin tokens cuesta 0 y ESO SÍ es un dato completo. `Micros: 0` con
// `Completo: true` significa «costó cero»; con `SinNingunaTarifa: true` significa «no se
// pudo cotizar». Son cosas distintas y el wire las distingue.
func TestCosteoCeroTokensEsCompleto(t *testing.T) {
	c := CalcularCosto(Tokens{}, precioHaiku45(), AritmeticaDisjunta)
	if c.Micros != 0 {
		t.Errorf("sin tokens el costo es 0: %d", c.Micros)
	}
	if !c.Completo || c.SinNingunaTarifa {
		t.Errorf("un uso vacío cotiza COMPLETO en 0, no «sin tarifa»: %+v", c)
	}
}

// TestParidadConElCostoReportadoReal cruza nuestro cálculo contra el `cost_usd_micros` que
// Claude Code reportó en la corrida REAL del 2026-07-26
// (verificacion-2026-07-26/evidencia/logs-run1.json + result-envelope.json).
//
// Es el test que convierte al catálogo de «datos que bajé» en «datos verificados»: con el
// split 5m/1h que solo trae el stream-json, el número tiene que dar exactamente el que el
// runtime reportó. Si diera otro, o el catálogo está mal o el costeo está mal — y no
// tendríamos forma de saber cuál sin este cruce.
func TestParidadConElCostoReportadoReal(t *testing.T) {
	// Datos MEDIDOS (result-envelope.json): input 10 · output 39 · cache_read 17 536 ·
	// cache_creation 8 257, todo en el tramo de 1 h (ephemeral_5m = 0, que es un CERO
	// LEGÍTIMO: el runtime lo dijo, no es una ausencia).
	tk := Tokens{
		Entrada:          i64(10),
		Salida:           i64(39),
		CacheLectura:     i64(17_536),
		CacheEscritura5m: i64(0),
		CacheEscritura1h: i64(8_257),
	}
	const reportadoMicros = 18_473 // `cost_usd_micros` de logs-run1.json, medido en vivo.

	c := CalcularCosto(tk, precioHaiku45(), AritmeticaDisjunta)
	if !c.Completo {
		t.Fatalf("con las tarifas reales de haiku-4-5 el costo debe ser completo: %+v", c)
	}
	if c.Micros != reportadoMicros {
		t.Fatalf("nuestro costo = %d micros, el runtime reportó %d. "+
			"El catálogo o el costeo divergen del dato medido en vivo (evidencia/logs-run1.json)",
			c.Micros, reportadoMicros)
	}

	// El contraste que prueba que el split IMPORTA: cotizando los mismos 8 257 tokens con
	// la tarifa de 5 min (que es lo único que se sabe cuando solo llega OTLP), el número
	// NO cierra. Esa divergencia es exactamente la señal que el detector B1 explota.
	soloOTLP := Tokens{
		Entrada: i64(10), Salida: i64(39), CacheLectura: i64(17_536),
		CacheEscritura5m: i64(8_257),
	}
	c5m := CalcularCosto(soloOTLP, precioHaiku45(), AritmeticaDisjunta)
	if c5m.Micros == reportadoMicros {
		t.Fatal("cotizar el cache write a la tarifa de 5 min no puede dar el mismo número que a la de 1 h: " +
			"si diera, el split no serviría para nada y B1 no existiría")
	}
	if c5m.Micros >= c.Micros {
		t.Errorf("la tarifa de 1 h es más cara que la de 5 min: 1h=%d, 5m=%d", c.Micros, c5m.Micros)
	}
}
