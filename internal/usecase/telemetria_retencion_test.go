package usecase_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/telemetria/catalogo"
	telstore "github.com/alpacapurpura/arnesia/internal/adapters/telemetria/store"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// servicioConRetencion arma servicio + store + rollup con un reloj movible y el TTL dado.
func servicioConRetencion(t *testing.T, dias int, reloj *time.Time) (*usecase.TelemetriaService, *telstore.Store) {
	t.Helper()
	st, err := telstore.New(filepath.Join(t.TempDir(), "telemetria.db"), telstore.Opciones{
		LoteEspera: 10 * time.Millisecond,
		Reloj:      func() time.Time { return *reloj },
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	roll := telstore.NewRollup(st, time.Hour) // debounce largo: se dispara a mano.
	t.Cleanup(roll.Detener)

	svc := usecase.NewTelemetriaService(st, catalogo.Embebido(), nil, nil,
		domain.DetectoresMVP(),
		domain.PerfilRuntime{Runtime: "claude-code", Aritmetica: domain.AritmeticaDisjunta},
		func() time.Time { return *reloj })
	svc.SetRetencion(st, roll, dias, 24)
	return svc, st
}

// TestPurgaRespetaTTL — lo más viejo que el TTL se va; lo reciente se queda.
func TestPurgaRespetaTTL(t *testing.T) {
	reloj := ahoraFijo
	svc, st := servicioConRetencion(t, 90, &reloj)
	ctx := context.Background()

	viejo := ahoraFijo.AddDate(0, 0, -120) // más allá de los 90
	nuevo := ahoraFijo.AddDate(0, 0, -3)
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{
		ev("s-viejo", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.TSRecibido = viejo
			e.CostoReportadoMicros = i64(100)
		}),
		ev("s-nuevo", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "vitalia"
			e.TSRecibido = nuevo
			e.CostoReportadoMicros = i64(200)
		}),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}

	n, err := svc.Purgar(ctx, ports.PurgaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("se esperaba 1 fila purgada (la de 120 días), se purgaron %d", n)
	}
	// ── control positivo: lo reciente sigue ahí ──
	turnos, err := svc.Turnos(ctx, ports.ConsultaTelemetria{Desde: ahoraFijo.AddDate(0, 0, -365)})
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 1 || turnos[0].SesionID != "s-nuevo" {
		t.Fatalf("el evento reciente tiene que sobrevivir a la purga: %+v", turnos)
	}
}

// TestRetencionNoEsUnaConstante — el TTL sale de la CONFIG, no de un número escrito en el
// código ni en la UI. Con 45 días, la salud reporta 45.
func TestRetencionNoEsUnaConstante(t *testing.T) {
	reloj := ahoraFijo
	svc, _ := servicioConRetencion(t, 45, &reloj)
	if svc.RetencionDias() != 45 {
		t.Fatalf("RetencionDias() = %d, se esperaba 45", svc.RetencionDias())
	}
	sal, err := svc.Salud(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sal.RetencionDias != 45 {
		t.Errorf("la salud reporta %d, se esperaba 45 — el TTL no puede estar hardcodeado", sal.RetencionDias)
	}
	// El número está FIRMADO en 90 (D26.3) y aun así esta corrida reporta 45: firmar el
	// default no lo clava, y el candado de que no está hardcodeado sigue siendo este.
	// Control positivo del contraste: con el default, reporta el default.
	reloj2 := ahoraFijo
	svc2, _ := servicioConRetencion(t, 0, &reloj2)
	if svc2.RetencionDias() != usecase.RetencionDefaultDias {
		t.Errorf("sin config, el default: %d", svc2.RetencionDias())
	}
}

// TestBorradoPorArnesTambienLimpiaRollup — el botón «borrar la telemetría de este arnés»
// borra el detalle **y el agregado**. Si no, quedaría una cifra huérfana en el tablero
// alimentándose de filas que ya no existen — y sobreviviría justo a la operación que el
// usuario pidió para hacerla desaparecer.
func TestBorradoPorArnesTambienLimpiaRollup(t *testing.T) {
	reloj := ahoraFijo
	svc, st := servicioConRetencion(t, 90, &reloj)
	ctx := context.Background()

	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{
		ev("s-a", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "arnes-a"
			e.CostoReportadoMicros = i64(1000)
		}),
		ev("s-b", "t-1", func(e *domain.EventoTelemetria) {
			e.ArnesID = "arnes-b"
			e.CostoReportadoMicros = i64(500)
		}),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}
	roll := telstore.NewRollup(st, time.Hour)
	defer roll.Detener()
	if err := roll.Actualizar(ctx); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.BorrarArnes(ctx, "arnes-a", time.Time{}, time.Time{}); err != nil {
		t.Fatal(err)
	}

	// El resumen de A vuelve al estado inicial: sin costo (null, no 0).
	rA, err := svc.Resumen(ctx, ports.ConsultaTelemetria{ArnesID: "arnes-a"})
	if err != nil {
		t.Fatal(err)
	}
	if rA.CostoReportadoMicros != nil {
		t.Errorf("tras borrar, el costo de A vuelve a NULL, no a %d", *rA.CostoReportadoMicros)
	}
	if rA.Turnos != 0 {
		t.Errorf("tras borrar, A no tiene turnos: %d", rA.Turnos)
	}
	// ── control positivo: las de B siguen intactas ──
	rB, err := svc.Resumen(ctx, ports.ConsultaTelemetria{ArnesID: "arnes-b"})
	if err != nil {
		t.Fatal(err)
	}
	if rB.CostoReportadoMicros == nil || *rB.CostoReportadoMicros != 500 {
		t.Fatalf("control positivo: borrar A no puede tocar a B: %v", rB.CostoReportadoMicros)
	}
}

// TestBorrarSinArnesNoEsUnBorrado — un borrado por arnés con arnés vacío sería un wipe. Se
// rechaza: la operación que el usuario pidió es acotada, y ejecutarla sin acotar es otra cosa.
func TestBorrarSinArnesNoEsUnBorrado(t *testing.T) {
	reloj := ahoraFijo
	svc, _ := servicioConRetencion(t, 90, &reloj)
	if _, err := svc.BorrarArnes(context.Background(), "", time.Time{}, time.Time{}); err == nil {
		t.Fatal("borrar «todos los arneses» pasando el vacío tiene que rechazarse")
	}
}

// TestAvisoDeTamano — el aviso de tamaño sale del archivo real, no de una estimación.
func TestAvisoDeTamano(t *testing.T) {
	reloj := ahoraFijo
	svc, st := servicioConRetencion(t, 90, &reloj)
	ctx := context.Background()
	if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{
		ev("s-1", "t-1", func(e *domain.EventoTelemetria) { e.ArnesID = "vitalia" }),
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}
	sal, err := svc.Salud(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if sal.TamanoBytes <= 0 {
		t.Error("el tamaño sale del archivo real: tiene que ser > 0")
	}
	// Con una base chica NO hay aviso: el aviso significa algo solo si no es constante.
	if sal.AvisoTamano {
		t.Errorf("una base de %d bytes no puede disparar el aviso de 500 MB", sal.TamanoBytes)
	}
}

// TestElBorradoRespetaLaVentana — D26.5 · A-4. Es la ÚNICA acción irreversible de la
// superficie, así que se prueba por los **dos** bordes en la misma corrida:
//
//   - con ventana, sobrevive lo de afuera (si no, la confirmación subdeclara la destrucción);
//   - sin ventana, no sobrevive nada (si no, el botón «borrar todo» no borra todo).
//
// Un test de un solo borde pasa con una implementación que ignora la ventana **o** con una que
// no borra nunca. Los dos juntos no.
func TestElBorradoRespetaLaVentana(t *testing.T) {
	reloj := ahoraFijo
	svc, st := servicioConRetencion(t, 90, &reloj)
	ctx := context.Background()

	viejo := ahoraFijo.AddDate(0, 0, -30)
	reciente := ahoraFijo.AddDate(0, 0, -1)
	ingerirEn := func(sesion, turno string, ts time.Time) {
		t.Helper()
		e := domain.EventoTelemetria{
			LlaveJoin: domain.LlaveJoin{SesionID: sesion, TurnoID: turno},
			Emisor:    domain.EmisorOTLP, Runtime: "claude-code", AdaptadorVersion: "cc-otlp/1",
			TSRecibido: ts, TipoEvento: domain.EventoAPIRequest,
			Escenario: domain.EscenarioS1,
			Modelo:    "claude-haiku-4-5",
		}
		e.ArnesID = "vitalia"
		e.CostoReportadoMicros = i64(1_000)
		if _, err := svc.Ingerir(ctx, []domain.EventoTelemetria{e}); err != nil {
			t.Fatalf("Ingerir: %v", err)
		}
		if err := st.Sincronizar(ctx); err != nil {
			t.Fatalf("Sincronizar: %v", err)
		}
	}
	cuantosQuedan := func() int {
		t.Helper()
		turnos, err := svc.Turnos(ctx, ports.ConsultaTelemetria{
			ArnesID: "vitalia",
			Desde:   ahoraFijo.AddDate(0, 0, -365), Hasta: ahoraFijo.AddDate(0, 0, 1),
		})
		if err != nil {
			t.Fatal(err)
		}
		return len(turnos)
	}

	ingerirEn("s-viejo", "t-1", viejo)
	ingerirEn("s-nuevo", "t-1", reciente)
	if n := cuantosQuedan(); n != 2 {
		t.Fatalf("el fixture no quedó armado: hay %d turnos, se esperaban 2", n)
	}

	// ── Borde 1: con ventana, lo de afuera SOBREVIVE ──
	n, err := svc.BorrarArnes(ctx, "vitalia", ahoraFijo.AddDate(0, 0, -7), ahoraFijo.AddDate(0, 0, 1))
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Error("el borrado acotado no borró nada: la ventana se está aplicando de más")
	}
	if q := cuantosQuedan(); q != 1 {
		t.Errorf("quedan %d turnos, se esperaba 1 — el borrado acotado se llevó puesto lo de "+
			"afuera de la ventana, que es justo lo que la confirmación NO declara", q)
	}

	// ── Borde 2: sin ventana, no sobrevive nada ──
	if _, err := svc.BorrarArnes(ctx, "vitalia", time.Time{}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if q := cuantosQuedan(); q != 0 {
		t.Errorf("quedan %d turnos tras el borrado total: «borrar todo» tiene que borrar todo", q)
	}
}
