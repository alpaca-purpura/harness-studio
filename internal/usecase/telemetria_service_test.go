package usecase_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

var ahoraFijo = time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)

func i64(v int64) *int64 { return &v }

// nuevoServicio arma servicio + store sobre un directorio TEMPORAL. **Nunca el HOME real**:
// escribir eventos de prueba en la base real contaminaría los totales que la feature muestra.
func nuevoServicio(t *testing.T) (*usecase.TelemetriaService, *telstore.Store) {
	t.Helper()
	return nuevoServicioConDetectores(t, domain.DetectoresMVP())
}

// nuevoServicioConDetectores es el mismo servicio con la lista de detectores inyectada. Los
// detectores son un puerto del servicio, así que un test puede probar el MOTOR sin depender
// del comportamiento de ninguno de los seis reales.
func nuevoServicioConDetectores(t *testing.T, ds []domain.Detector) (*usecase.TelemetriaService, *telstore.Store) {
	t.Helper()
	st, err := telstore.New(filepath.Join(t.TempDir(), "telemetria.db"), telstore.Opciones{
		LoteEspera: 10 * time.Millisecond,
		Reloj:      func() time.Time { return ahoraFijo },
	})
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	reg := usecase.NuevoRegistroAtribucion(
		func(ctx context.Context, hash string) (string, string, bool) { return st.BuscarHash(ctx, hash) },
		func(huella string) (string, string, bool) { return "", "", false },
		func(ctx context.Context, h, a, i, c string) error { return st.AprenderHash(ctx, h, a, i, c) },
	)
	svc := usecase.NewTelemetriaService(st, catalogo.Embebido(), reg, nil, ds,
		domain.PerfilRuntime{Runtime: "claude-code", Aritmetica: domain.AritmeticaDisjunta},
		func() time.Time { return ahoraFijo })
	return svc, st
}

func ev(sesion, turno string, mod func(*domain.EventoTelemetria)) domain.EventoTelemetria {
	e := domain.EventoTelemetria{
		LlaveJoin:        domain.LlaveJoin{SesionID: sesion, TurnoID: turno},
		Emisor:           domain.EmisorOTLP,
		Runtime:          "claude-code",
		AdaptadorVersion: "cc-otlp/1",
		TSRecibido:       ahoraFijo,
		TipoEvento:       domain.EventoAPIRequest,
		Escenario:        domain.EscenarioS2Instrumentado,
	}
	if mod != nil {
		mod(&e)
	}
	return e
}

func ingerir(t *testing.T, svc *usecase.TelemetriaService, st *telstore.Store, evs ...domain.EventoTelemetria) {
	t.Helper()
	ctx := context.Background()
	if _, err := svc.Ingerir(ctx, evs); err != nil {
		t.Fatalf("Ingerir: %v", err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatalf("Sincronizar: %v", err)
	}
}

// TestAtribucionPorHash — el segundo escalón del orden de preferencia: sin los `arnesia.*`,
// el `plugin_id_hash` resuelve en la tabla APRENDIDA (no adivinada) y da el arnés.
func TestAtribucionPorHash(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()
	if err := st.AprenderHash(ctx, "hash-abc", "vitalia", "home-local", "spawn-controlado"); err != nil {
		t.Fatal(err)
	}
	e := ev("s-1", "t-1", func(e *domain.EventoTelemetria) { e.PluginIDHash = "hash-abc" })
	svc.Atribuir(&e)
	if e.Atribucion != domain.ConfianzaPorHash {
		t.Errorf("atribucion = %q, se esperaba por-hash", e.Atribucion)
	}
	if e.ArnesID != "vitalia" {
		t.Errorf("arnes = %q, se esperaba vitalia", e.ArnesID)
	}
	// **Por hash NO se atribuye la caja**: el hash identifica el paquete instalado, no qué
	// parte de él corrió. Inventarla sería adivinar.
	if e.CajaID != "" {
		t.Errorf("por-hash da el arnés, jamás la caja: caja=%q", e.CajaID)
	}
}

// TestHashDesconocido — un hash que no está en la tabla NO inventa un arnés. Queda sin-dato.
func TestHashDesconocido(t *testing.T) {
	svc, _ := nuevoServicio(t)
	e := ev("s-1", "t-1", func(e *domain.EventoTelemetria) { e.PluginIDHash = "hash-que-nadie-aprendio" })
	svc.Atribuir(&e)
	if e.Atribucion != domain.ConfianzaSinDato {
		t.Errorf("un hash desconocido no atribuye nada: %q", e.Atribucion)
	}
	if e.ArnesID != "" {
		t.Errorf("no se puede inventar un arnés: %q", e.ArnesID)
	}
}

// TestAtribucionExactaGanaSobreLasDemas — el primero que acierta FIJA la confianza: con los
// `arnesia.*` presentes ni se consulta la tabla de hashes.
func TestAtribucionExactaGanaSobreLasDemas(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()
	// Se aprende un hash que apunta a OTRO arnés, a propósito.
	if err := st.AprenderHash(ctx, "hash-abc", "otro-arnes", "", "spawn-controlado"); err != nil {
		t.Fatal(err)
	}
	e := ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.PluginIDHash = "hash-abc"
	})
	svc.Atribuir(&e)
	if e.Atribucion != domain.ConfianzaExacta {
		t.Errorf("atribucion = %q, se esperaba exacta", e.Atribucion)
	}
	if e.ArnesID != "vitalia" {
		t.Errorf("la atribución exacta no puede ser pisada por el hash: %q", e.ArnesID)
	}
}

// TestSinDatoNoSumaAlTotal — A15 a nivel de servicio: lo no atribuido se guarda, aparece en
// la cobertura y NO entra al total.
func TestSinDatoNoSumaAlTotal(t *testing.T) {
	svc, st := nuevoServicio(t)
	var evs []domain.EventoTelemetria
	for i := 0; i < 5; i++ {
		evs = append(evs, ev("s-ok", "t-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CostoReportadoMicros = i64(100)
		}))
	}
	for i := 0; i < 3; i++ {
		evs = append(evs, ev("s-otro", "u-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.CostoReportadoMicros = i64(999_999) // ruido caro: si sumara, se vería
		}))
	}
	ingerir(t, svc, st, evs...)

	r, err := svc.Resumen(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros == nil || *r.CostoReportadoMicros != 500 {
		t.Fatalf("el total suma SOLO lo atribuido: %v", r.CostoReportadoMicros)
	}
	if r.Cobertura.SinDato != 3 {
		t.Errorf("Cobertura.SinDato = %d, se esperaban 3", r.Cobertura.SinDato)
	}
	// El denominador NO es la suma de los medidos.
	if r.Cobertura.Exacta+r.Cobertura.SinDato != 8 {
		t.Errorf("los 8 turnos tienen que estar contados en algún lado: %+v", r.Cobertura)
	}
}

// TestJoinPorSesionYTurno — el join es la igualdad de dos campos. Sin heurística de tiempo.
func TestJoinPorSesionYTurno(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st,
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CostoReportadoMicros = i64(1234)
		}),
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.Emisor = domain.EmisorHook
			e.TipoEvento = domain.EventoTurnoFin
			e.DuracionMs = i64(900)
			// El evento de proceso llega 3 horas DESPUÉS: si el join usara tiempo, no
			// uniría. Une igual, porque la llave es el par.
			e.TSRecibido = ahoraFijo.Add(-3 * time.Hour)
			ts := ahoraFijo.Add(-3 * time.Hour)
			e.TSEmisor = &ts
		}),
	)
	turnos, err := svc.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 1 {
		t.Fatalf("los dos eventos del mismo par son UN turno: hay %d", len(turnos))
	}
	if !turnos[0].TieneDinero || !turnos[0].TieneProceso {
		t.Errorf("el turno unido debe tener las dos mitades: %+v", turnos[0])
	}
}

// TestEscenarioSeDerivaDeLaSenal — tres lotes, tres escenarios. **Ningún campo del emisor
// pudo elegirlo**: se calcula de lo que llegó.
func TestEscenarioSeDerivaDeLaSenal(t *testing.T) {
	casos := []struct {
		nombre string
		evs    []domain.EventoTelemetria
		want   domain.Escenario
	}{
		{"con corrida nuestra", []domain.EventoTelemetria{
			ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
				e.ArnesID = "vitalia"
				e.CorridaID = "run-1"
				// El emisor MIENTE sobre su escenario, a propósito.
				e.Escenario = domain.EscenarioS2Degradado
				e.CostoReportadoMicros = i64(10)
			}),
		}, domain.EscenarioS1},
		{"api_request sin corrida", []domain.EventoTelemetria{
			ev("s-2", "t-1", func(e *domain.EventoTelemetria) {
				e.ArnesID = "vitalia"
				e.Escenario = domain.EscenarioS1 // otra mentira
				e.CostoReportadoMicros = i64(10)
			}),
		}, domain.EscenarioS2Instrumentado},
		{"solo hook", []domain.EventoTelemetria{
			ev("s-3", "t-1", func(e *domain.EventoTelemetria) {
				e.ArnesID = "vitalia"
				e.Emisor = domain.EmisorHook
				e.TipoEvento = domain.EventoTurnoFin
				e.Escenario = domain.EscenarioS1 // otra más
			}),
		}, domain.EscenarioS2Degradado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			svc, st := nuevoServicio(t)
			ingerir(t, svc, st, c.evs...)
			r, err := svc.Resumen(context.Background(), ports.ConsultaTelemetria{})
			if err != nil {
				t.Fatal(err)
			}
			if r.Escenario != c.want {
				t.Errorf("escenario = %q, se esperaba %q — el emisor NO puede elegirlo", r.Escenario, c.want)
			}
		})
	}
}

// TestDobleCostoSePersisteEntero — los DOS costos conviven; ninguno pisa al otro.
func TestDobleCostoSePersisteEntero(t *testing.T) {
	svc, st := nuevoServicio(t)
	// El uso REAL medido: el costo calculado tiene que dar el mismo número que el reportado
	// cuando el split viene completo.
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.Modelo = "claude-haiku-4-5"
		e.Tokens = domain.Tokens{
			Entrada: i64(10), Salida: i64(39), CacheLectura: i64(17_536),
			CacheEscritura5m: i64(0), CacheEscritura1h: i64(8_257),
		}
		e.CostoReportadoMicros = i64(18_473)
	}))
	turnos, err := svc.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 1 {
		t.Fatalf("se esperaba 1 turno, hay %d", len(turnos))
	}
	tu := turnos[0]
	if tu.CostoReportadoMicros == nil || *tu.CostoReportadoMicros != 18_473 {
		t.Errorf("el costo reportado se pisó: %v", tu.CostoReportadoMicros)
	}
	if tu.CostoCalculadoMicros == nil {
		t.Fatal("el costo calculado no se persistió: los DOS viajan siempre")
	}
	if *tu.CostoCalculadoMicros != 18_473 {
		t.Errorf("costo calculado = %d, se esperaba 18473 (paridad con el dato medido)", *tu.CostoCalculadoMicros)
	}
}

// TestCostoSoloCalculado — un runtime que no reporta costo igual se cotiza con nuestro
// catálogo: falta uno de los dos, no los dos.
func TestCostoSoloCalculado(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.Modelo = "claude-haiku-4-5"
		e.Tokens = domain.Tokens{Entrada: i64(1000), Salida: i64(500)}
		e.CostoReportadoMicros = nil // el runtime no lo dijo
	}))
	turnos, err := svc.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if turnos[0].CostoReportadoMicros != nil {
		t.Error("lo que el runtime no dijo sigue en null, no en 0")
	}
	if turnos[0].CostoCalculadoMicros == nil || *turnos[0].CostoCalculadoMicros != 3500 {
		t.Errorf("el costo calculado debe existir igual: %v", turnos[0].CostoCalculadoMicros)
	}
}

// TestModeloDesconocidoNoCuestaCero — un modelo que el catálogo no conoce deja el calculado
// en null. Un 0 se leería como «salió gratis».
func TestModeloDesconocidoNoCuestaCero(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.Modelo = "un-modelo-de-otro-mundo"
		e.Tokens = domain.Tokens{Entrada: i64(1_000_000)}
	}))
	turnos, err := svc.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if turnos[0].CostoCalculadoMicros != nil {
		t.Errorf("un modelo desconocido deja el costo en null, no en %d", *turnos[0].CostoCalculadoMicros)
	}
}

// TestConfianzaDeAgregadoEsLaMinima — el total de una caja con partes de confianza distinta
// vale lo que su peor parte.
func TestConfianzaDeAgregadoEsLaMinima(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()
	if err := st.AprenderHash(ctx, "h1", "vitalia", "", "spawn-controlado"); err != nil {
		t.Fatal(err)
	}
	ingerir(t, svc, st,
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CajaID = "paso-3"
			e.CostoReportadoMicros = i64(100)
		}),
		ev("s-1", "t-2", func(e *domain.EventoTelemetria) {
			e.PluginIDHash = "h1" // resolverá por-hash, que es peor que exacta
			e.CajaID = "paso-3"
			e.CostoReportadoMicros = i64(50)
		}),
	)
	cajas, err := svc.PorCaja(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	var caja *domain.GastoCaja
	for i := range cajas {
		if cajas[i].CajaID == "paso-3" {
			caja = &cajas[i]
		}
	}
	if caja == nil {
		t.Fatalf("no se encontró paso-3: %+v", cajas)
	}
	if caja.Confianza != domain.ConfianzaPorHash {
		t.Errorf("la confianza del agregado es la PEOR de sus partes: %q, se esperaba por-hash", caja.Confianza)
	}
}

// TestConciliacionCuentaLosNoLlegados — 7 turnos ocurrieron, 5 se midieron ⇒ 2 no llegaron.
// Eso convierte el agujero de red de «un 0 que parece un dato» en «2 de 7 turnos no
// reportaron telemetría».
func TestConciliacionCuentaLosNoLlegados(t *testing.T) {
	// Reloj movible: los turnos se declaran 10 min antes para que pasen el corte de 5 min
	// (un turno recién declarado que todavía no reportó no es un agujero, es un turno en
	// vuelo — y contarlo sería fabricar una pérdida).
	reloj := ahoraFijo.Add(-10 * time.Minute)
	st, err := telstore.New(filepath.Join(t.TempDir(), "telemetria.db"), telstore.Opciones{
		LoteEspera: 10 * time.Millisecond,
		Reloj:      func() time.Time { return reloj },
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	svc := usecase.NewTelemetriaService(st, catalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(),
		domain.PerfilRuntime{Runtime: "claude-code", Aritmetica: domain.AritmeticaDisjunta},
		func() time.Time { return reloj })
	ctx := context.Background()

	// El daemon declara los SIETE turnos que ocurrieron, se sepan medir o no.
	for i := 0; i < 7; i++ {
		if cerr := svc.Conciliar(ctx, "s-1", "t-"+itoa(i), "vitalia", ""); cerr != nil {
			t.Fatal(cerr)
		}
	}
	// Pasan 10 minutos y llegan solo CINCO api_request.
	reloj = ahoraFijo
	var evs []domain.EventoTelemetria
	for i := 0; i < 5; i++ {
		evs = append(evs, ev("s-1", "t-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CorridaID = "run-1" // ⇒ escenario s1, el único que tiene denominador
			e.CostoReportadoMicros = i64(100)
		}))
	}
	if _, ierr := svc.Ingerir(ctx, evs); ierr != nil {
		t.Fatal(ierr)
	}
	if serr := st.Sincronizar(ctx); serr != nil {
		t.Fatal(serr)
	}

	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Cobertura.Esperados == nil || *r.Cobertura.Esperados != 7 {
		t.Fatalf("Esperados = %v, se esperaban 7", r.Cobertura.Esperados)
	}
	if r.Cobertura.NoLlegaron == nil || *r.Cobertura.NoLlegaron != 2 {
		t.Fatalf("NoLlegaron = %v, se esperaban 2 — el agujero se ve, no se disimula", r.Cobertura.NoLlegaron)
	}
	// ── control positivo: los 5 que SÍ llegaron están contados como medidos ──
	if r.Cobertura.Exacta != 5 {
		t.Errorf("Cobertura.Exacta = %d, se esperaban 5", r.Cobertura.Exacta)
	}
	// Y el total NO se presenta como completo.
	if *r.Cobertura.Esperados == r.Cobertura.Exacta {
		t.Error("el denominador no puede coincidir con lo medido cuando hubo turnos sin medir")
	}
}

// TestPuestoSaleDelRolDelArnes — D20: la fila del Portafolio saca el puesto del `rol` del
// arnés indexado, y viaja **null** cuando no lo declara. Ese es el caso NORMAL hoy.
func TestPuestoSaleDelRolDelArnes(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()
	svc.SetPortafolio(
		func(ctx context.Context, arnesID string) string {
			if arnesID == "con-rol" {
				return "arquitecto"
			}
			return "" // sin rol declarado
		},
		func(ctx context.Context) []usecase.FilaInstalacion {
			return []usecase.FilaInstalacion{
				{ArnesID: "con-rol", InstalacionID: "i1", Clave: "k1", Nombre: "Con rol"},
				{ArnesID: "sin-rol", InstalacionID: "i2", Clave: "k2", Nombre: "Sin rol"},
			}
		},
	)
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "con-rol"
		e.InstalacionID = "i1"
		e.CorridaID = "run-1"
		e.CostoReportadoMicros = i64(1000)
	}))

	filas, err := svc.Portafolio(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(filas) != 2 {
		t.Fatalf("una fila por (arnés, instalación): hay %d", len(filas))
	}
	for _, f := range filas {
		switch f.ArnesID {
		case "con-rol":
			if f.Puesto == nil || *f.Puesto != "arquitecto" {
				t.Errorf("el puesto sale del rol del arnés: %v", f.Puesto)
			}
			if f.CostoPorCorrida == nil {
				t.Error("con corridas, el costo por corrida existe")
			}
		case "sin-rol":
			if f.Puesto != nil {
				t.Errorf("sin rol declarado el puesto viaja NULL, no una cadena vacía: %q", *f.Puesto)
			}
			// Y el que nunca corrió lleva `costo_por_corrida: null`, no 0.
			if f.CostoPorCorrida != nil {
				t.Errorf("el que nunca corrió lleva null, no %d", *f.CostoPorCorrida)
			}
		}
	}
}

// TestInstalacionesNoSeSumanSolas — dos instalaciones del MISMO arnés son dos filas. Sumarlas
// escondería que el mismo arnés cuesta distinto en dos lugares, que es el punto del eje.
func TestInstalacionesNoSeSumanSolas(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()
	svc.SetPortafolio(
		func(ctx context.Context, arnesID string) string { return "" },
		func(ctx context.Context) []usecase.FilaInstalacion {
			return []usecase.FilaInstalacion{
				{ArnesID: "acme-cli", InstalacionID: "proyecto-a", Clave: "k1"},
				{ArnesID: "acme-cli", InstalacionID: "proyecto-b", Clave: "k2"},
			}
		},
	)
	ingerir(t, svc, st,
		ev("s-a", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "acme-cli"
			e.InstalacionID = "proyecto-a"
			e.CostoReportadoMicros = i64(1000)
		}),
		ev("s-b", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "acme-cli"
			e.InstalacionID = "proyecto-b"
			e.CostoReportadoMicros = i64(250)
		}),
	)
	filas, err := svc.Portafolio(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(filas) != 2 {
		t.Fatalf("dos instalaciones = dos filas: hay %d", len(filas))
	}
	for _, f := range filas {
		switch f.InstalacionID {
		case "proyecto-a":
			if f.CostoMicros == nil || *f.CostoMicros != 1000 {
				t.Errorf("proyecto-a: %v", f.CostoMicros)
			}
		case "proyecto-b":
			if f.CostoMicros == nil || *f.CostoMicros != 250 {
				t.Errorf("proyecto-b: %v", f.CostoMicros)
			}
		}
	}
}

// TestGastoSinCaja — el gasto que no se pudo asignar a una caja NO desaparece: sale en su
// propia fila, con motivo.
func TestGastoSinCaja(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st,
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CajaID = "paso-3"
			e.CostoReportadoMicros = i64(100)
		}),
		ev("s-1", "t-2", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CostoReportadoMicros = i64(70) // sin caja
		}),
	)
	cajas, err := svc.PorCaja(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	var sinCaja *domain.GastoCaja
	for i := range cajas {
		if cajas[i].CajaID == "" {
			sinCaja = &cajas[i]
		}
	}
	if sinCaja == nil {
		t.Fatalf("el gasto sin caja no puede desaparecer: %+v", cajas)
	}
	if sinCaja.CostoMicros == nil || *sinCaja.CostoMicros != 70 {
		t.Errorf("el gasto sin caja conserva su monto: %v", sinCaja.CostoMicros)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestNoMedidosSeDeclaran — A16: la respuesta lleva las TRES listas SIEMPRE. Los otros siete
// detectores de la familia se declaran «no medidos todavía» en vez de callarse. Omitirlos
// obligaría al FE a elegir entre no mostrar nada (gap escondido) o mostrar 0 (mentira).
func TestNoMedidosSeDeclaran(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.CostoReportadoMicros = i64(100)
	}))
	r, err := svc.Mejoras(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.NoMedidos) != 8 {
		t.Errorf("los detectores fuera del MVP se declaran, no se callan: hay %d", len(r.NoMedidos))
	}
	for _, nm := range r.NoMedidos {
		if nm.Motivo == "" {
			t.Errorf("%s: «no medido todavía» es un motivo, y tiene que estar escrito", nm.Detector)
		}
		if nm.Aplica {
			t.Errorf("%s: un detector no medido no puede figurar como aplicando", nm.Detector)
		}
	}
	// Las tres listas viajan SIEMPRE, aunque estén vacías: un `null` obligaría al FE a
	// distinguir «no hay» de «no vino».
	if r.Puntos == nil || r.NoAplican == nil || r.NoMedidos == nil {
		t.Error("las tres listas viajan siempre, aunque vacías")
	}
	// ── control positivo: los que SÍ aplican no están en no_medidos ──
	for _, nm := range r.NoMedidos {
		for _, d := range domain.DetectoresMVP() {
			if nm.Detector == d.ID() {
				t.Errorf("%s es del MVP: no puede estar en no_medidos", nm.Detector)
			}
		}
	}
}

// TestS2DegradadoElResumenNoTraeCeros — con solo eventos de hook, el resumen trae los costos
// en **null**. Un 0 diría «este arnés no gastó nada», que es una afirmación distinta de «no
// pudimos medir cuánto gastó».
func TestS2DegradadoElResumenNoTraeCeros(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.Emisor = domain.EmisorHook
		e.TipoEvento = domain.EventoTurnoFin
		e.CostoReportadoMicros = nil
	}))
	ctx := context.Background()
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Escenario != domain.EscenarioS2Degradado {
		t.Fatalf("escenario = %q, se esperaba s2-degradado", r.Escenario)
	}
	if r.CostoReportadoMicros != nil {
		t.Errorf("sin señal de dinero el costo viaja NULL, no %d", *r.CostoReportadoMicros)
	}
	if r.CostoCalculadoMicros != nil {
		t.Errorf("sin tokens no hay costo calculado: %d", *r.CostoCalculadoMicros)
	}
	// Y los detectores de dinero salen en no_aplican CON MOTIVO.
	mej, err := svc.Mejoras(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	apagados := map[domain.DetectorID]bool{domain.DetB4: true, domain.DetB6: true, domain.DetB3: true}
	vistos := 0
	for _, na := range mej.NoAplican {
		if apagados[na.Detector] {
			vistos++
			if na.Motivo == "" {
				t.Errorf("%s se apagó sin motivo", na.Detector)
			}
		}
	}
	if vistos != 3 {
		t.Errorf("B4, B6 y B3 tienen que salir en no_aplican: hay %d de 3", vistos)
	}
}

// detectorSinFix es un detector de prueba que ENCUENTRA algo y no puede proponer nada: emite
// un punto sin contrafactual. Existe porque desde D26.1 **ningún detector real produce ese
// caso** —B1 se arregló—, y el invariante que se prueba acá es del MOTOR, no de un detector:
// un punto sin recomendación se declara, no se borra.
//
// Un test que dependiera de que B1 esté roto se pondría verde por la razón equivocada y se
// caería el día que B1 se arregle. Eso ya pasó una vez en este paquete.
type detectorSinFix struct{}

func (detectorSinFix) ID() domain.DetectorID { return "x9-encuentra-y-no-propone" }
func (detectorSinFix) Nombre() string        { return "encuentra y no propone" }
func (detectorSinFix) Aplica(domain.ContextoDeteccion) domain.Aplicabilidad {
	return domain.Aplicabilidad{Aplica: true}
}

func (detectorSinFix) Evaluar(v domain.Ventana) []domain.PuntoDeMejora {
	if len(v.Turnos) == 0 {
		return nil
	}
	return []domain.PuntoDeMejora{{
		Detector: "x9-encuentra-y-no-propone", ScoreVersion: domain.ScoreVersionMVP,
		Titulo: "algo cotizable, sin arreglo que proponer", Lede: "se midió; no hay fix",
		GastoMicros: 50_000, Umbral: "≥ 1", Sesgo: "n/a en prueba", DireccionSesgo: "subestima",
		Fix: "—", Confianza: domain.ConfianzaExacta, CorridasUsadas: 1, CorridasTotales: 1,
		// Sin `BaseContrafactual` y sin ahorro ⇒ `Redactar()` no arma prosa (regla A4).
	}}
}

// TestUnPuntoSinContrafactualSeDeclaraNoSeBorra — D25 en la puerta de salida del wire.
//
// La regla A4 dice que una tarjeta sin contrafactual es un reproche, no una recomendación; y el
// boundary `no-aplica-no-es-cero` dice que lo que no se puede mostrar **se declara**. Juntas: el
// punto no viaja como tarjeta, y el detector viaja con `sin_fix` y su motivo. Si solo se filtrara,
// «no lo mostramos» se leería como «no encontró nada» — el hueco exacto que este módulo existe
// para no dejar.
func TestUnPuntoSinContrafactualSeDeclaraNoSeBorra(t *testing.T) {
	svc, st := nuevoServicioConDetectores(t, []domain.Detector{detectorSinFix{}})
	ingerir(t, svc, st, ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.CajaID = "paso-3"
		e.Modelo = "claude-haiku-4-5"
		e.CostoReportadoMicros = i64(20_000)
	}))
	r, err := svc.Mejoras(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Puntos) != 0 {
		t.Errorf("un punto sin contrafactual no se publica como tarjeta: %+v", r.Puntos)
	}
	var declarado *domain.EstadoDetector
	for i := range r.NoAplican {
		if r.NoAplican[i].SinFix {
			declarado = &r.NoAplican[i]
		}
	}
	if declarado == nil {
		t.Fatal("encontró y no puede proponer: tiene que viajar con sin_fix, no desaparecer")
	}
	if declarado.Motivo == "" {
		t.Error("sin_fix sin motivo es un gap escondido")
	}
	if !declarado.Aplica {
		t.Error("SÍ aplicó — decir lo contrario confundiría «no pudo correr» con «corrió y no propone»")
	}
	if declarado.Hallazgos == 0 {
		t.Error("un detector sin fix igual dice CUÁNTOS hallazgos tuvo: 0 se leería como «no encontró»")
	}
}

// TestTodoPuntoPublicadoLlevaSuProsa — el control positivo del test de arriba, y el candado de
// V-5: **ningún punto sale al wire con las piezas numéricas y sin la frase**. Es la forma de que
// el FE no pueda volver a quedarse tipando un campo que nadie manda.
func TestTodoPuntoPublicadoLlevaSuProsa(t *testing.T) {
	svc, st := nuevoServicio(t)
	ingerir(t, svc, st,
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CajaID = "paso-3"
			e.Modelo = "claude-haiku-4-5"
			e.CostoReportadoMicros = i64(90_000)
		}),
		ev("s-1", "t-2", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.CajaID = "paso-4"
			e.Modelo = "claude-haiku-4-5"
			e.CostoReportadoMicros = i64(1_000)
		}),
	)
	r, err := svc.Mejoras(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Puntos) == 0 {
		t.Fatal("el control positivo no produjo ningún punto: el test no comparó nada")
	}
	for _, p := range r.Puntos {
		if p.ID == "" {
			t.Errorf("%s: punto sin id — la superficie no puede descartar «el tercero de la lista»", p.Detector)
		}
		if p.Contrafactual == "" {
			t.Errorf("%s: punto publicado sin contrafactual en prosa (D25 · RF-249)", p.Detector)
		}
		if p.DiferenciaMicros <= 0 {
			t.Errorf("%s: punto publicado con ahorro %d", p.Detector, p.DiferenciaMicros)
		}
		if !strings.Contains(p.Contrafactual, "en la ventana") {
			t.Errorf("%s: el contrafactual no declara su unidad: %q", p.Detector, p.Contrafactual)
		}
		if p.Calculo == "" {
			t.Errorf("%s: sin cálculo resuelto no hay «ver el cálculo» que auditar (H-4)", p.Detector)
		}
	}
}

// TestUltimaCorridaIgnoraLaVentana — D26.4, estado 1b. **«No corrió en estos 7 días» y «nunca
// corrió» son afirmaciones distintas**, y sobre un arnés con historial la segunda es falsa.
//
// El dato que las separa es la última corrida mirando TODO el historial, no la ventana. Se
// prueba por los dos lados: con historial fuera de la ventana viaja la fecha; sin ningún
// evento viaja `nil`, que es el estado 1 de verdad.
func TestUltimaCorridaIgnoraLaVentana(t *testing.T) {
	svc, st := nuevoServicio(t)
	ctx := context.Background()

	// Sin un solo evento: «nunca corrió» es cierto y viaja como nil.
	vacio, err := svc.Resumen(ctx, ports.ConsultaTelemetria{ArnesID: "vitalia"})
	if err != nil {
		t.Fatal(err)
	}
	if vacio.UltimaCorrida != nil {
		t.Errorf("sin eventos la última corrida es nil, no %v — nil es «nunca corrió»", vacio.UltimaCorrida)
	}

	viejo := ahoraFijo.AddDate(0, 0, -30)
	e := ev("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.ArnesID = "vitalia"
		e.Modelo = "claude-haiku-4-5"
		e.TSRecibido = viejo
		e.CostoReportadoMicros = i64(1_000)
	})
	ingerir(t, svc, st, e)

	// Ventana de 7 días: 0 corridas adentro, PERO la última corrida existe y viaja.
	r, err := svc.Resumen(ctx, ports.ConsultaTelemetria{
		ArnesID: "vitalia", Desde: ahoraFijo.AddDate(0, 0, -7), Hasta: ahoraFijo,
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Corridas != 0 {
		t.Errorf("la ventana de 7 días no puede contener la corrida de hace 30: corridas = %d", r.Corridas)
	}
	if r.UltimaCorrida == nil {
		t.Fatal("hay historial fuera de la ventana: decir «nunca corrió» sería falso")
	}
	if !r.UltimaCorrida.Equal(viejo.UTC()) {
		t.Errorf("última corrida = %v, se esperaba %v", r.UltimaCorrida, viejo.UTC())
	}
	// Y el runtime viaja con su veredicto: sabemos medir este, y lo decimos.
	if r.Runtime != "claude-code" || !r.RuntimeSoportado {
		t.Errorf("runtime = %q soportado = %v — el veredicto sale del dato, no de una config",
			r.Runtime, r.RuntimeSoportado)
	}
}
