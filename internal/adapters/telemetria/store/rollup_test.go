package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// TestNoAplicaSobreviveAlRollup — **el test central del ticket**. Una hora donde ningún
// evento tuvo `tok_cache_1h` queda `NULL` en el agregado, no `0`.
//
// Sin esto, el tablero mostraría «0 tokens de cache de una hora» en vez de «este runtime no
// reporta ese bucket», que son afirmaciones distintas: la primera es medición, la segunda es
// ausencia. Y una vez agregado a 0, la diferencia ya no se puede recuperar.
func TestNoAplicaSobreviveAlRollup(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour) // debounce largo: en el test se dispara a mano.
	ctx := context.Background()

	ingerirYEsperar(t, s,
		evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.Tokens.Entrada = i64(10)
			e.Tokens.Salida = i64(5)
			// tok_cache_1h y tok_razonamiento quedan nil: el runtime no los reportó.
		}),
		evento("s-1", "t-2", func(e *domain.EventoTelemetria) {
			e.Tokens.Entrada = i64(20)
		}),
	)
	if err := r.Actualizar(ctx); err != nil {
		t.Fatalf("Actualizar: %v", err)
	}

	var entrada, salida, c1h, raz sql.NullInt64
	if err := s.reader.QueryRow(
		`SELECT SUM(tok_entrada), SUM(tok_salida), SUM(tok_cache_1h), SUM(tok_razonamiento) FROM rollup_hora`).
		Scan(&entrada, &salida, &c1h, &raz); err != nil {
		t.Fatal(err)
	}
	if c1h.Valid {
		t.Errorf("tok_cache_1h agregó a %d; nadie lo reportó, tiene que seguir NULL", c1h.Int64)
	}
	if raz.Valid {
		t.Errorf("tok_razonamiento agregó a %d; tiene que seguir NULL", raz.Int64)
	}
	// ── control positivo, misma corrida ──
	if !entrada.Valid || entrada.Int64 != 30 {
		t.Fatalf("control positivo: tok_entrada debe sumar 30, dio %v", entrada)
	}
	if !salida.Valid || salida.Int64 != 5 {
		t.Fatalf("control positivo: tok_salida debe sumar 5, dio %v", salida)
	}

	// Y el NULL sobrevive a un SEGUNDO lote que tampoco lo trae: el `ON CONFLICT` es donde
	// un COALESCE(...,0) apurado rompería la propiedad.
	ingerirYEsperar(t, s, evento("s-1", "t-3", func(e *domain.EventoTelemetria) {
		e.Tokens.Entrada = i64(7)
	}))
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.reader.QueryRow(
		`SELECT SUM(tok_entrada), SUM(tok_cache_1h) FROM rollup_hora`).Scan(&entrada, &c1h); err != nil {
		t.Fatal(err)
	}
	if c1h.Valid {
		t.Errorf("tras acumular un segundo lote, tok_cache_1h pasó a %d — el ON CONFLICT rompió el NULL", c1h.Int64)
	}
	if !entrada.Valid || entrada.Int64 != 37 {
		t.Errorf("control positivo tras el segundo lote: tok_entrada = %v, se esperaba 37", entrada)
	}
}

// TestNULLSeVuelveDatoCuandoAlguienLoMide — el contraste que el test anterior necesita para
// no ser ambiguo: cuando un evento SÍ trae el bucket, el agregado deja de ser NULL. Si no
// fuera así, el «NULL sobrevive» sería «NULL siempre», que es otro bug.
func TestNULLSeVuelveDatoCuandoAlguienLoMide(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	ingerirYEsperar(t, s, evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.Tokens.Entrada = i64(10)
	}))
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	ingerirYEsperar(t, s, evento("s-1", "t-2", func(e *domain.EventoTelemetria) {
		e.Tokens.Entrada = i64(10)
		e.Tokens.CacheEscritura1h = i64(8_257) // el split llegó por el stream-json
	}))
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	var c1h sql.NullInt64
	if err := s.reader.QueryRow(`SELECT SUM(tok_cache_1h) FROM rollup_hora`).Scan(&c1h); err != nil {
		t.Fatal(err)
	}
	if !c1h.Valid || c1h.Int64 != 8_257 {
		t.Fatalf("cuando alguien SÍ mide el bucket, el agregado deja de ser NULL: %v", c1h)
	}
}

// TestRollupReanudaDesdeElCursor — el rollup se interrumpe y se retoma; ninguna fila se
// cuenta dos veces.
func TestRollupReanudaDesdeElCursor(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	ingerirYEsperar(t, s,
		evento("s-1", "t-1", func(e *domain.EventoTelemetria) { e.CostoReportadoMicros = i64(100) }),
		evento("s-1", "t-2", func(e *domain.EventoTelemetria) { e.CostoReportadoMicros = i64(200) }),
	)
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	var total sql.NullInt64
	if err := s.reader.QueryRow(`SELECT SUM(costo_reportado_micros) FROM rollup_hora`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total.Int64 != 300 {
		t.Fatalf("primera pasada: %d, se esperaba 300", total.Int64)
	}

	// Correrla otra vez SIN eventos nuevos no puede cambiar nada: es idempotente.
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.reader.QueryRow(`SELECT SUM(costo_reportado_micros) FROM rollup_hora`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total.Int64 != 300 {
		t.Fatalf("una segunda pasada sin eventos nuevos duplicó el total: %d", total.Int64)
	}

	// Y una tercera pasada CON eventos nuevos agrega solo los nuevos.
	ingerirYEsperar(t, s, evento("s-1", "t-3", func(e *domain.EventoTelemetria) { e.CostoReportadoMicros = i64(50) }))
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.reader.QueryRow(`SELECT SUM(costo_reportado_micros) FROM rollup_hora`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if total.Int64 != 350 {
		t.Fatalf("tras el tercer evento: %d, se esperaba 350", total.Int64)
	}
	// El cursor avanzó.
	var cursor int64
	if err := s.reader.QueryRow(`SELECT ultimo_evento FROM rollup_cursor WHERE unico=1`).Scan(&cursor); err != nil {
		t.Fatal(err)
	}
	if cursor != 3 {
		t.Errorf("cursor = %d, se esperaba 3", cursor)
	}
}

// TestTurnoCruzaHora — asserta **LA DIRECCIÓN del sesgo**, no su ausencia. Un turno que cruza
// la frontera horaria se cuenta en las dos horas: el denominador se INFLA, o sea el
// costo-por-turno que mostramos BAJA. El sesgo va **en contra** de nuestra propia
// recomendación, que es la única dirección aceptable para un sesgo que no se puede eliminar.
func TestTurnoCruzaHora(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	// El MISMO turno, con eventos a las 14:59 y a las 15:01.
	antes := time.Date(2026, 7, 26, 14, 59, 0, 0, time.UTC)
	despues := time.Date(2026, 7, 26, 15, 1, 0, 0, time.UTC)
	ingerirYEsperar(t, s,
		evento("s-1", "t-cruza", func(e *domain.EventoTelemetria) {
			e.TSRecibido = antes
			e.CostoReportadoMicros = i64(600)
		}),
		evento("s-1", "t-cruza", func(e *domain.EventoTelemetria) {
			e.TSRecibido = despues
			e.CostoReportadoMicros = i64(400)
			ts := despues
			e.TSEmisor = &ts // distinto ts_emisor: no es un duplicado
		}),
	)
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	var turnosAgregados int
	var costo sql.NullInt64
	if err := s.reader.QueryRow(`SELECT SUM(turnos), SUM(costo_reportado_micros) FROM rollup_hora`).
		Scan(&turnosAgregados, &costo); err != nil {
		t.Fatal(err)
	}
	// El costo NO se duplica (cada evento suma una vez)...
	if !costo.Valid || costo.Int64 != 1000 {
		t.Fatalf("el costo se contó mal: %v, se esperaba 1000", costo)
	}
	// ...pero el turno SÍ se cuenta dos veces, una por hora. Ese es el sesgo.
	if turnosAgregados != 2 {
		t.Fatalf("el sesgo declarado es que un turno a caballo cuenta en las DOS horas: turnos=%d", turnosAgregados)
	}
	// Y la DIRECCIÓN: el costo-por-turno agregado (1000/2 = 500) es MENOR que el real
	// (1000/1 = 1000). El sesgo subestima el gasto por turno — en contra de recomendar un
	// cambio, nunca a favor.
	real := 1000.0
	agregado := float64(costo.Int64) / float64(turnosAgregados)
	if agregado >= real {
		t.Errorf("el sesgo tiene que ir EN CONTRA: costo-por-turno agregado %.0f, real %.0f", agregado, real)
	}
	t.Logf("sesgo declarado: costo-por-turno agregado %.0f vs real %.0f (subestima, en contra de la recomendación)",
		agregado, real)
}

// TestRollupColapsaCardinalidad — al superar el tope, `caja_id` colapsa a `(otros)` y la fila
// queda MARCADA. Degradación declarada, no silenciosa: perder la dimensión y no decirlo sería
// mostrar un desglose incompleto como si fuera completo.
func TestRollupColapsaCardinalidad(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	// Sembrar el rollup por encima del tope sería carísimo en un test. Se verifica el
	// MECANISMO: con la tabla ya por encima del tope, la próxima pasada colapsa.
	if _, err := s.writer.ExecContext(ctx, `INSERT INTO rollup_hora
	  (hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion, eventos, turnos)
	  SELECT '2026-01-01T00', 'a', '', 'caja-' || value, 'claude-code', '', 'otlp', 'exacta', 1, 1
	    FROM (WITH RECURSIVE serie(value) AS (SELECT 1 UNION ALL SELECT value+1 FROM serie WHERE value < ?)
	          SELECT value FROM serie)`, topeCardinalidad); err != nil {
		t.Fatalf("sembrar cardinalidad: %v", err)
	}
	var filas int
	if err := s.reader.QueryRow(`SELECT COUNT(*) FROM rollup_hora`).Scan(&filas); err != nil {
		t.Fatal(err)
	}
	if filas < topeCardinalidad {
		t.Fatalf("el sembrado no llegó al tope: %d filas", filas)
	}

	ingerirYEsperar(t, s, evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.CajaID = "una-caja-mas"
		e.CostoReportadoMicros = i64(10)
	}))
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}
	if !r.Colapsada() {
		t.Error("superado el tope, el rollup tiene que reportar que colapsó")
	}
	var colapsadas int
	if err := s.reader.QueryRow(
		`SELECT COUNT(*) FROM rollup_hora WHERE caja_id = ? AND cardinalidad_colapsada = 1`,
		CajaColapsada).Scan(&colapsadas); err != nil {
		t.Fatal(err)
	}
	if colapsadas == 0 {
		t.Error("la fila colapsada debe existir y estar MARCADA — un desglose incompleto que no se declara es una mentira")
	}
	// Y la caja original NO aparece con su nombre: se colapsó, no se duplicó.
	var conNombre int
	if err := s.reader.QueryRow(`SELECT COUNT(*) FROM rollup_hora WHERE caja_id = 'una-caja-mas'`).
		Scan(&conNombre); err != nil {
		t.Fatal(err)
	}
	if conNombre != 0 {
		t.Error("al colapsar, la caja no puede seguir apareciendo con su nombre propio")
	}
}

// TestRecomputarTrasBorrado — borrar un arnés obliga a rehacer sus horas: si no, el agregado
// seguiría mostrando un total alimentado por filas que ya no existen.
func TestRecomputarTrasBorrado(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	ingerirYEsperar(t, s,
		evento("s-a", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "arnes-a"
			e.CostoReportadoMicros = i64(1000)
		}),
		evento("s-b", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "arnes-b"
			e.CostoReportadoMicros = i64(500)
		}),
	)
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}

	horas, err := s.HorasAfectadas(ctx, "arnes-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(horas) == 0 {
		t.Fatal("no se detectaron horas afectadas por arnes-a")
	}
	if _, perr := s.Purgar(ctx, ports.PurgaTelemetria{ArnesID: "arnes-a"}); perr != nil {
		t.Fatal(perr)
	}
	if rerr := r.Recomputar(ctx, horas); rerr != nil {
		t.Fatal(rerr)
	}

	var quedanA int
	if err := s.reader.QueryRow(`SELECT COUNT(*) FROM rollup_hora WHERE arnes_id='arnes-a'`).Scan(&quedanA); err != nil {
		t.Fatal(err)
	}
	if quedanA != 0 {
		t.Errorf("tras borrar arnes-a no puede quedar agregado suyo: %d filas", quedanA)
	}
	// ── control positivo: las de B siguen intactas ──
	var costoB sql.NullInt64
	if err := s.reader.QueryRow(
		`SELECT SUM(costo_reportado_micros) FROM rollup_hora WHERE arnes_id='arnes-b'`).Scan(&costoB); err != nil {
		t.Fatal(err)
	}
	if !costoB.Valid || costoB.Int64 != 500 {
		t.Errorf("control positivo: el agregado de arnes-b debe seguir en 500, está en %v", costoB)
	}
}

// TestRollupYCrudoDicenLoMismo — anti-drift: el agregado y la tabla cruda tienen que dar el
// mismo total. Un rollup que diverge del detalle es peor que no tener rollup: muestra un
// número que nadie puede reproducir mirando las filas.
func TestRollupYCrudoDicenLoMismo(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	var evs []domain.EventoTelemetria
	for i := 0; i < 20; i++ {
		evs = append(evs, evento("s-"+itoa(i%3), "t-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.CostoReportadoMicros = i64(int64(100 + i))
			e.CajaID = "paso-" + itoa(i%4)
		}))
	}
	ingerirYEsperar(t, s, evs...)
	if err := r.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}

	var crudo, agregado sql.NullInt64
	if err := s.reader.QueryRow(`SELECT SUM(costo_reportado_micros) FROM evento`).Scan(&crudo); err != nil {
		t.Fatal(err)
	}
	if err := s.reader.QueryRow(`SELECT SUM(costo_reportado_micros) FROM rollup_hora`).Scan(&agregado); err != nil {
		t.Fatal(err)
	}
	if !crudo.Valid || !agregado.Valid {
		t.Fatalf("los dos lados deben tener dato: crudo=%v agregado=%v", crudo, agregado)
	}
	if crudo.Int64 != agregado.Int64 {
		t.Fatalf("el agregado divergió del detalle: crudo=%d agregado=%d", crudo.Int64, agregado.Int64)
	}
	t.Logf("crudo == agregado == %d micros sobre %d eventos", crudo.Int64, len(evs))
}

// BenchmarkRollup mide la pasada incremental. El presupuesto del tablero es p95 ≤ 50 ms
// sobre el agregado (§11); esto mide el costo de MANTENERLO.
func BenchmarkRollup(b *testing.B) {
	s, err := New(b.TempDir()+"/telemetria.db", Opciones{LoteEspera: 5 * time.Millisecond})
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	r := NewRollup(s, time.Hour)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var evs []domain.EventoTelemetria
		for j := 0; j < 100; j++ {
			ts := time.Date(2026, 7, 26, 15, 0, i%60, j, time.UTC)
			evs = append(evs, domain.EventoTelemetria{
				LlaveJoin: domain.LlaveJoin{
					SesionID: "s", TurnoID: "t-" + itoa(i*100+j), ArnesID: "a", CajaID: "c-" + itoa(j%5),
				},
				Emisor: domain.EmisorOTLP, Runtime: "claude-code", AdaptadorVersion: "cc-otlp/1",
				TSRecibido: ts, TSEmisor: &ts, TipoEvento: domain.EventoAPIRequest,
				Escenario: domain.EscenarioS1, Atribucion: domain.ConfianzaExacta,
				CostoReportadoMicros: i64(100),
			})
		}
		if _, ierr := s.Ingerir(ctx, evs); ierr != nil {
			b.Fatal(ierr)
		}
		if serr := s.Sincronizar(ctx); serr != nil {
			b.Fatal(serr)
		}
		if aerr := r.Actualizar(ctx); aerr != nil {
			b.Fatal(aerr)
		}
	}
}
