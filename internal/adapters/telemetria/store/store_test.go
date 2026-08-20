package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

func i64(v int64) *int64 { return &v }

// nuevoStore abre un store en un directorio temporal. **Nunca en el HOME real del operador**:
// escribir eventos de prueba en la base real contaminaría los totales que la propia feature
// muestra, que es exactamente el pecado que este módulo existe para no cometer.
func nuevoStore(t *testing.T, o Opciones) *Store {
	t.Helper()
	if o.Reloj == nil {
		o.Reloj = func() time.Time { return time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC) }
	}
	if o.LoteEspera == 0 {
		o.LoteEspera = 10 * time.Millisecond
	}
	s, err := New(filepath.Join(t.TempDir(), "telemetria.db"), o)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func evento(sesion, turno string, mod func(*domain.EventoTelemetria)) domain.EventoTelemetria {
	e := domain.EventoTelemetria{
		LlaveJoin:        domain.LlaveJoin{SesionID: sesion, TurnoID: turno, ArnesID: "vitalia"},
		Emisor:           domain.EmisorOTLP,
		Runtime:          "claude-code",
		AdaptadorVersion: "cc-otlp/1",
		TSRecibido:       time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC),
		TipoEvento:       domain.EventoAPIRequest,
		Escenario:        domain.EscenarioS2Instrumentado,
		Atribucion:       domain.ConfianzaExacta,
	}
	if mod != nil {
		mod(&e)
	}
	return e
}

func ingerirYEsperar(t *testing.T, s *Store, evs ...domain.EventoTelemetria) {
	t.Helper()
	ctx := context.Background()
	if _, err := s.Ingerir(ctx, evs); err != nil {
		t.Fatalf("Ingerir: %v", err)
	}
	if err := s.Sincronizar(ctx); err != nil {
		t.Fatalf("Sincronizar: %v", err)
	}
}

func contarEventos(t *testing.T, s *Store) int {
	t.Helper()
	var n int
	if err := s.reader.QueryRow(`SELECT COUNT(*) FROM evento`).Scan(&n); err != nil {
		t.Fatalf("contar: %v", err)
	}
	return n
}

// TestRutasPorHome — la base vive en `~/.arnesia/telemetria.db` y **jamás** es `index.db`.
func TestRutasPorHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	s, err := New("", Opciones{})
	if err != nil {
		t.Fatalf("New con ruta vacía: %v", err)
	}
	defer func() { _ = s.Close() }()

	want := filepath.Join(home, ".arnesia", "telemetria.db")
	if s.Ruta() != want {
		t.Errorf("ruta = %q, se esperaba %q", s.Ruta(), want)
	}
	if strings.Contains(s.Ruta(), "index.db") {
		t.Fatal("la telemetría JAMÁS puede vivir en index.db: un bump del índice borraría la historia")
	}
	if _, serr := os.Stat(want); serr != nil {
		t.Errorf("el archivo debería existir: %v", serr)
	}
}

// TestTelemetriaDBNoEsElIndice — el invariante A2 con el escenario que lo motiva: se borra el
// índice (que es DESECHABLE por doctrina) y la telemetría sigue entera.
func TestTelemetriaDBNoEsElIndice(t *testing.T) {
	dir := t.TempDir()
	rutaIndice := filepath.Join(dir, "index.db")
	rutaTel := filepath.Join(dir, "telemetria.db")

	// Un `index.db` cualquiera, para tener qué borrar.
	idx, err := sql.Open(driverName, "file:"+rutaIndice)
	if err != nil {
		t.Fatal(err)
	}
	if _, eerr := idx.Exec(`CREATE TABLE graphs (clave TEXT PRIMARY KEY)`); eerr != nil {
		t.Fatal(eerr)
	}
	_ = idx.Close()

	s, err := New(rutaTel, Opciones{LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	ingerirYEsperar(t, s, evento("s-1", "t-1", nil), evento("s-1", "t-2", nil))
	antes := contarEventos(t, s)
	if antes != 2 {
		t.Fatalf("se esperaban 2 eventos, hay %d", antes)
	}

	// El índice se borra entero — es su doctrina (RF-208).
	if rerr := os.Remove(rutaIndice); rerr != nil {
		t.Fatal(rerr)
	}
	if _, serr := os.Stat(rutaIndice); serr == nil {
		t.Fatal("control del test: el índice no se borró")
	}
	// Y la telemetría no se enteró.
	if despues := contarEventos(t, s); despues != antes {
		t.Errorf("un wipe del índice no puede tocar la telemetría: %d → %d eventos", antes, despues)
	}
	if _, serr := os.Stat(rutaTel); serr != nil {
		t.Errorf("telemetria.db debe seguir existiendo: %v", serr)
	}
}

// TestMigracionesSonAditivas — una migración destructiva no compila el merge.
func TestMigracionesSonAditivas(t *testing.T) {
	ms := Migraciones()
	if len(ms) == 0 {
		t.Fatal("no hay migraciones: el escáner no compararía nada")
	}
	sentencias := 0
	for _, m := range ms {
		for _, s := range m.SQL {
			sentencias++
			if !EsAditiva(s) {
				t.Errorf("migración %d contiene SQL destructivo:\n%s", m.Version, s)
			}
		}
	}
	if sentencias == 0 {
		t.Fatal("las migraciones no tienen sentencias: verde por vacío")
	}
	// ── control positivo: el detector encuentra lo que busca ──
	for _, malo := range []string{
		"DROP TABLE evento",
		"ALTER TABLE evento DROP COLUMN modelo",
		"ALTER TABLE evento RENAME TO viejo",
		"DELETE FROM evento",
	} {
		if EsAditiva(malo) {
			t.Errorf("control positivo: %q debería detectarse como destructivo", malo)
		}
	}
	// Y no marca como destructivo lo que sí es aditivo.
	for _, bueno := range []string{
		"CREATE TABLE IF NOT EXISTS x (a INTEGER) STRICT",
		"ALTER TABLE evento ADD COLUMN nuevo TEXT",
		"CREATE INDEX IF NOT EXISTS i ON evento (arnes_id)",
	} {
		if !EsAditiva(bueno) {
			t.Errorf("falso positivo: %q es aditivo", bueno)
		}
	}
}

// TestGeneracionArchivaNoBorra — un binario de otra generación **archiva** la historia, no la
// borra, y la salud lo dice.
func TestGeneracionArchivaNoBorra(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "telemetria.db")
	reloj := func() time.Time { return time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC) }

	s, err := New(ruta, Opciones{Reloj: reloj, LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	ingerirYEsperar(t, s, evento("s-vieja", "t-1", nil))
	if contarEventos(t, s) != 1 {
		t.Fatal("el evento de la generación vieja no se escribió")
	}
	_ = s.Close()

	// Se simula que el archivo es de una generación anterior.
	db, err := sql.Open(driverName, "file:"+ruta)
	if err != nil {
		t.Fatal(err)
	}
	if _, eerr := db.Exec(`UPDATE schema_meta SET generacion = ?`, generacionActual-1); eerr != nil {
		t.Fatal(eerr)
	}
	_ = db.Close()

	s2, err := New(ruta, Opciones{Reloj: reloj, LoteEspera: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("New tras mismatch de generación: %v", err)
	}
	defer func() { _ = s2.Close() }()

	// 1. La base nueva arranca VACÍA.
	if n := contarEventos(t, s2); n != 0 {
		t.Errorf("la base de la generación nueva arranca vacía, tiene %d eventos", n)
	}
	// 2. El archivo viejo SIGUE EN DISCO, renombrado.
	if s2.Archivada() == "" {
		t.Fatal("no se registró el nombre del archivo archivado")
	}
	if _, serr := os.Stat(filepath.Join(dir, s2.Archivada())); serr != nil {
		t.Errorf("el archivo archivado debe seguir en disco: %v", serr)
	}
	// 3. Y la salud lo DICE: una historia archivada que nadie menciona es una pérdida
	//    silenciosa.
	sal, err := s2.Salud(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sal.HistoriaArchivada != s2.Archivada() {
		t.Errorf("salud.historia_archivada = %q, se esperaba %q", sal.HistoriaArchivada, s2.Archivada())
	}
	// 4. Control positivo: los eventos viejos SIGUEN dentro del archivo archivado.
	viejo, err := sql.Open(driverName, "file:"+filepath.Join(dir, s2.Archivada()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = viejo.Close() }()
	var n int
	if serr := viejo.QueryRow(`SELECT COUNT(*) FROM evento`).Scan(&n); serr != nil {
		t.Fatalf("leer el archivo archivado: %v", serr)
	}
	if n != 1 {
		t.Errorf("el archivo archivado debe conservar sus %d eventos, tiene %d", 1, n)
	}
}

// TestDBCorruptaSeArchiva — mismo trato que el mismatch de generación: se archiva y se
// arranca vacía, nunca se borra en silencio.
func TestDBCorruptaSeArchiva(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "telemetria.db")
	// Un archivo que no es una base SQLite.
	if err := os.WriteFile(ruta, []byte("esto no es una base de datos, ni de casualidad"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := New(ruta, Opciones{
		Reloj:      func() time.Time { return time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC) },
		LoteEspera: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("una base corrupta no puede tumbar al daemon: %v", err)
	}
	defer func() { _ = s.Close() }()
	if s.Archivada() == "" {
		t.Fatal("una base ilegible se archiva, y el nombre se reporta")
	}
	if _, serr := os.Stat(filepath.Join(dir, s.Archivada())); serr != nil {
		t.Errorf("el archivo corrupto debe seguir en disco: %v", serr)
	}
	// Control positivo: la base nueva funciona.
	ingerirYEsperar(t, s, evento("s-nueva", "t-1", nil))
	if contarEventos(t, s) != 1 {
		t.Error("tras archivar la corrupta, la base nueva debe aceptar escrituras")
	}
}

// TestEventoDuplicadoNoSeCuentaDosVeces — un emisor que se reinicia y re-manda su lote no
// infla los totales (escenario C11).
func TestEventoDuplicadoNoSeCuentaDosVeces(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ts := time.Date(2026, 7, 26, 14, 59, 0, 0, time.UTC)
	e := evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.TSEmisor = &ts
		e.CostoReportadoMicros = i64(1000)
	})
	ingerirYEsperar(t, s, e)
	ingerirYEsperar(t, s, e) // el mismo, reenviado
	if n := contarEventos(t, s); n != 1 {
		t.Fatalf("el mismo evento reenviado no puede duplicarse: hay %d filas", n)
	}
	// Y el total no se infló.
	r, err := s.Resumen(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros == nil || *r.CostoReportadoMicros != 1000 {
		t.Errorf("el costo se duplicó: %v", r.CostoReportadoMicros)
	}
	// ── control positivo: un evento DISTINTO del mismo turno sí entra ──
	otro := evento("s-1", "t-1", func(x *domain.EventoTelemetria) {
		otroTS := ts.Add(time.Second)
		x.TSEmisor = &otroTS
		x.CostoReportadoMicros = i64(500)
	})
	ingerirYEsperar(t, s, otro)
	if n := contarEventos(t, s); n != 2 {
		t.Fatalf("un evento distinto del mismo turno SÍ debe entrar: hay %d filas", n)
	}
	// Y el dedupe se cuenta en la salud.
	sal, err := s.Salud(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if sal.Aceptados != 2 {
		t.Errorf("aceptados = %d, se esperaban 2 (el duplicado no cuenta)", sal.Aceptados)
	}
}

// TestEventoSinSesionSeRechazaEnElStore — sin `sesion_id` el evento no es atribuible ni
// deduplicable: no entra.
func TestEventoSinSesionSeRechazaEnElStore(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ctx := context.Background()
	if _, err := s.Ingerir(ctx, []domain.EventoTelemetria{evento("", "t-1", nil)}); err == nil {
		t.Error("un evento sin sesión debe rechazarse")
	}
	// Control positivo: con sesión, entra.
	n, err := s.Ingerir(ctx, []domain.EventoTelemetria{evento("s-1", "t-1", nil)})
	if err != nil || n != 1 {
		t.Fatalf("con sesión el evento debe aceptarse: n=%d err=%v", n, err)
	}
}

// TestRelojHaciaAtrasNoRompeLaVentana — A10: un `ts_emisor` absurdo NO descarta el evento (el
// reloj del daemon manda) pero SÍ marca la fila, para poder explicar después una ventana rara.
func TestRelojHaciaAtrasNoRompeLaVentana(t *testing.T) {
	ahora := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	s := nuevoStore(t, Opciones{Reloj: func() time.Time { return ahora }})

	haceTresDias := ahora.AddDate(0, 0, -3)
	enElFuturo := ahora.Add(2 * time.Hour)
	reciente := ahora.Add(-time.Minute)

	ingerirYEsperar(t, s,
		evento("s-viejo", "t-1", func(e *domain.EventoTelemetria) { e.TSEmisor = &haceTresDias }),
		evento("s-futuro", "t-1", func(e *domain.EventoTelemetria) { e.TSEmisor = &enElFuturo }),
		evento("s-ok", "t-1", func(e *domain.EventoTelemetria) { e.TSEmisor = &reciente }),
	)

	if n := contarEventos(t, s); n != 3 {
		t.Fatalf("los tres eventos deben guardarse: hay %d", n)
	}
	var sospechosos int
	if err := s.reader.QueryRow(`SELECT COUNT(*) FROM evento WHERE reloj_sospechoso = 1`).Scan(&sospechosos); err != nil {
		t.Fatal(err)
	}
	if sospechosos != 2 {
		t.Errorf("se esperaban 2 filas con reloj sospechoso (3 días atrás y 2 h adelante), hay %d", sospechosos)
	}
	// Control positivo: el reciente NO está marcado.
	var okSospechoso int
	if err := s.reader.QueryRow(
		`SELECT reloj_sospechoso FROM evento WHERE sesion_id = 's-ok'`).Scan(&okSospechoso); err != nil {
		t.Fatal(err)
	}
	if okSospechoso != 0 {
		t.Error("un ts_emisor razonable no puede marcarse sospechoso")
	}
	// Y la ventana usa el reloj del DAEMON: los tres entran a una ventana de ±1 min.
	r, err := s.Resumen(context.Background(), ports.ConsultaTelemetria{
		Desde: ahora.Add(-time.Minute), Hasta: ahora.Add(time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Sesiones != 3 {
		t.Errorf("la ventana usa ts_recibido, no ts_emisor: %d sesiones en la ventana, se esperaban 3", r.Sesiones)
	}
}

// TestDosEscritoresSobreElMismoDB — dos goroutines escribiendo en paralelo no se pisan ni
// dan SQLITE_BUSY: el writer tiene una sola conexión y database/sql serializa.
func TestDosEscritoresSobreElMismoDB(t *testing.T) {
	s := nuevoStore(t, Opciones{LoteMax: 8})
	ctx := context.Background()
	var wg sync.WaitGroup
	const goroutines, porGoroutine = 8, 25
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < porGoroutine; i++ {
				ts := time.Date(2026, 7, 26, 14, 0, i, g, time.UTC)
				e := evento("s-conc", "", func(e *domain.EventoTelemetria) {
					e.TSEmisor = &ts
					e.TurnoID = "t-" + itoa(g) + "-" + itoa(i)
				})
				if _, err := s.Ingerir(ctx, []domain.EventoTelemetria{e}); err != nil {
					t.Errorf("goroutine %d: %v", g, err)
					return
				}
			}
		}(g)
	}
	wg.Wait()
	if err := s.Sincronizar(ctx); err != nil {
		t.Fatal(err)
	}
	if n := contarEventos(t, s); n != goroutines*porGoroutine {
		t.Errorf("se esperaban %d eventos, hay %d", goroutines*porGoroutine, n)
	}
}

// TestDiscoLlenoNoTumbaElDaemon — un almacén que no puede escribir degrada honesto: el error
// se cuenta y el proceso sigue vivo. Se pierde telemetría, no la sesión del usuario.
func TestDiscoLlenoNoTumbaElDaemon(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ctx := context.Background()
	// Se cierra el writer por debajo: cualquier escritura posterior falla, que es el efecto
	// observable de un disco lleno o de un archivo que se volvió inescribible.
	if err := s.writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Ingerir(ctx, []domain.EventoTelemetria{evento("s-1", "t-1", nil)}); err != nil {
		t.Fatalf("Ingerir no puede fallar por un almacén roto: encolar es todo lo que hace (%v)", err)
	}
	// El writer va a fallar en background; lo que importa es que el proceso sigue vivo.
	time.Sleep(60 * time.Millisecond)
	// Control positivo del contraste: la salud sigue respondiendo (degradada) en vez de
	// tumbar al llamador.
	if _, err := s.Salud(ctx); err != nil {
		t.Errorf("Salud debe responder aunque el almacén esté roto: %v", err)
	}
}

// TestColaLlenaDescartaYCuenta — la cola acotada descarta y lo CUENTA. Nunca bloquea al
// emisor: un receptor que hace esperar al agente degradaría el trabajo que mide.
func TestColaLlenaDescartaYCuenta(t *testing.T) {
	// Cola de 1 y un writer que tarda: el segundo lote no entra.
	s := nuevoStore(t, Opciones{Cola: 1, LoteEspera: 5 * time.Second, LoteMax: 10_000})
	ctx := context.Background()
	descartados := 0
	for i := 0; i < 50; i++ {
		if _, err := s.Ingerir(ctx, []domain.EventoTelemetria{evento("s-1", "t-"+itoa(i), nil)}); err != nil {
			descartados++
		}
	}
	if descartados == 0 {
		t.Fatal("con cola de 1 y writer lento, algo tuvo que descartarse — el test no probó nada")
	}
	sal, err := s.Salud(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if sal.DescartadosColaLlena == 0 {
		t.Error("el descarte por cola llena tiene que CONTARSE: un descarte silencioso es una señal perdida sin rastro")
	}
}

// TestNoAplicaSeGuardaComoNULL — el invariante que atraviesa el módulo entero: un bucket
// ausente se persiste como NULL, no como 0.
func TestNoAplicaSeGuardaComoNULL(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ingerirYEsperar(t, s, evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
		e.Tokens.Entrada = i64(10)
		// Razonamiento y los dos tramos de cache write quedan nil.
	}))
	var raz, c5, c1h sql.NullInt64
	var entrada sql.NullInt64
	if err := s.reader.QueryRow(
		`SELECT tok_razonamiento, tok_cache_5m, tok_cache_1h, tok_entrada FROM evento`).
		Scan(&raz, &c5, &c1h, &entrada); err != nil {
		t.Fatal(err)
	}
	if raz.Valid || c5.Valid || c1h.Valid {
		t.Errorf("los buckets ausentes deben ser NULL, no 0: raz=%v c5=%v c1h=%v", raz, c5, c1h)
	}
	// Control positivo: el bucket con dato SÍ está.
	if !entrada.Valid || entrada.Int64 != 10 {
		t.Errorf("el bucket con dato debe persistirse: %v", entrada)
	}
}

// TestSinDatoNoSumaAlTotalEnElStore — A15: lo no atribuido se guarda y NO suma. Aparece en
// Cobertura.SinDato.
func TestSinDatoNoSumaAlTotalEnElStore(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	var evs []domain.EventoTelemetria
	for i := 0; i < 5; i++ {
		evs = append(evs, evento("s-ok", "t-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.CostoReportadoMicros = i64(100)
		}))
	}
	for i := 0; i < 3; i++ {
		evs = append(evs, evento("s-huerfano", "u-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.ArnesID = ""
			e.Atribucion = domain.ConfianzaSinDato
			e.CostoReportadoMicros = i64(1_000_000) // ruido caro, para que se note si suma
		}))
	}
	ingerirYEsperar(t, s, evs...)

	r, err := s.Resumen(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.CostoReportadoMicros == nil || *r.CostoReportadoMicros != 500 {
		t.Fatalf("el total suma SOLO lo atribuido (5×100): %v", r.CostoReportadoMicros)
	}
	if r.Cobertura.SinDato != 3 {
		t.Errorf("Cobertura.SinDato = %d, se esperaban 3", r.Cobertura.SinDato)
	}
	if r.Cobertura.Exacta != 5 {
		t.Errorf("Cobertura.Exacta = %d, se esperaban 5", r.Cobertura.Exacta)
	}
	// La confianza del agregado es la PEOR de sus partes.
	if r.Confianza != domain.ConfianzaSinDato {
		t.Errorf("confianza del agregado = %q, se esperaba sin-dato (la peor de las partes)", r.Confianza)
	}
}

// TestCoberturaSinDenominadorEnS2 — en S2 no hay forma independiente de saber cuántos turnos
// hubo: `Esperados` viaja **nil**, no 0. Un 0 diría «no hubo turnos», que es otra cosa.
func TestCoberturaSinDenominadorEnS2(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ingerirYEsperar(t, s, evento("s-1", "t-1", nil)) // s2-instrumentado por default
	r, err := s.Resumen(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Escenario != domain.EscenarioS2Instrumentado {
		t.Fatalf("escenario = %q, se esperaba s2-instrumentado", r.Escenario)
	}
	if r.Cobertura.Esperados != nil {
		t.Errorf("fuera de ArnesIA no hay denominador: Esperados debe ser nil, es %d", *r.Cobertura.Esperados)
	}
	if r.Cobertura.NoLlegaron != nil {
		t.Errorf("sin denominador no se puede decir cuántos no llegaron: %d", *r.Cobertura.NoLlegaron)
	}
}

// TestConciliacionCuentaLosNoLlegadosEnElStore — con denominador (S1): 7 turnos ocurrieron,
// 5 se midieron ⇒ 2 no llegaron. Eso convierte el agujero de red de «un 0 que parece un
// dato» en «2 de 7 turnos no reportaron telemetría».
func TestConciliacionCuentaLosNoLlegadosEnElStore(t *testing.T) {
	ahora := time.Date(2026, 7, 26, 15, 0, 0, 0, time.UTC)
	s := nuevoStore(t, Opciones{Reloj: func() time.Time { return ahora }})
	ctx := context.Background()

	// El daemon declara 7 turnos que OCURRIERON. Se escriben con el reloj fijado 10 min
	// atrás para que pasen el corte de 5 min.
	viejo := nuevoStore(t, Opciones{Reloj: func() time.Time { return ahora.Add(-10 * time.Minute) }})
	_ = viejo.Close()
	for i := 0; i < 7; i++ {
		if _, err := s.writer.ExecContext(ctx,
			`INSERT INTO turno_esperado (sesion_id, turno_id, arnes_id, ts, medido) VALUES (?,?,?,?,0)`,
			"s-1", "t-"+itoa(i), "vitalia", ahora.Add(-10*time.Minute).Format(time.RFC3339Nano)); err != nil {
			t.Fatal(err)
		}
	}
	// Llegan 5 api_request, con corrida nuestra (⇒ escenario s1, que es el que tiene
	// denominador).
	var evs []domain.EventoTelemetria
	for i := 0; i < 5; i++ {
		evs = append(evs, evento("s-1", "t-"+itoa(i), func(e *domain.EventoTelemetria) {
			e.CorridaID = "run-1"
			e.Escenario = domain.EscenarioS1
			e.CostoReportadoMicros = i64(100)
		}))
	}
	ingerirYEsperar(t, s, evs...)

	r, err := s.Resumen(ctx, ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Escenario != domain.EscenarioS1 {
		t.Fatalf("con arnesia.corrida el escenario es s1, es %q", r.Escenario)
	}
	if r.Cobertura.Esperados == nil || *r.Cobertura.Esperados != 7 {
		t.Fatalf("Esperados = %v, se esperaban 7", r.Cobertura.Esperados)
	}
	if r.Cobertura.NoLlegaron == nil || *r.Cobertura.NoLlegaron != 2 {
		t.Fatalf("NoLlegaron = %v, se esperaban 2", r.Cobertura.NoLlegaron)
	}
	// Y el total NO se presenta como completo: la suma de los medidos no es el denominador.
	if *r.Cobertura.Esperados == r.Cobertura.Exacta {
		t.Error("el denominador no puede coincidir con lo medido cuando hubo turnos sin medir")
	}
}

// TestPorCajaIncluyeLasCajasSinDato — omitir una caja sin dato obligaría al FE a inventar por
// qué falta; ponerle 0 diría que corrió gratis.
func TestPorCajaIncluyeLasCajasSinDato(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ingerirYEsperar(t, s,
		evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.CajaID = "paso-3"
			e.CostoReportadoMicros = i64(1000)
		}),
		evento("s-2", "t-2", func(e *domain.EventoTelemetria) {
			e.CajaID = "paso-11"
			e.ArnesID = ""
			e.Atribucion = domain.ConfianzaSinDato
		}),
	)
	cajas, err := s.PorCaja(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cajas) != 2 {
		t.Fatalf("las cajas sin dato NO se omiten: hay %d", len(cajas))
	}
	var conDato, sinDato *domain.GastoCaja
	for i := range cajas {
		if cajas[i].Atribuible {
			conDato = &cajas[i]
		} else {
			sinDato = &cajas[i]
		}
	}
	if conDato == nil || sinDato == nil {
		t.Fatalf("se esperaba una caja con dato y una sin: %+v", cajas)
	}
	if sinDato.CostoMicros != nil {
		t.Errorf("una caja sin dato lleva costo null, no 0: %d", *sinDato.CostoMicros)
	}
	if sinDato.Motivo == "" {
		t.Error("el motivo es OBLIGATORIO cuando no es atribuible: un «sin dato» sin razón es un gap escondido")
	}
	// Control positivo: la que sí tiene dato lo tiene.
	if conDato.CostoMicros == nil || *conDato.CostoMicros != 1000 {
		t.Errorf("la caja atribuible lleva su costo: %v", conDato.CostoMicros)
	}
}

// TestJoinPorSesionYTurnoEnElStore — el join es la igualdad de dos campos. Un `api_request` y
// un `Stop` del mismo par producen UN turno con dinero y proceso.
func TestJoinPorSesionYTurnoEnElStore(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ingerirYEsperar(t, s,
		evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.CostoReportadoMicros = i64(1234)
			e.Tokens.Entrada = i64(10)
		}),
		evento("s-1", "t-1", func(e *domain.EventoTelemetria) {
			e.Emisor = domain.EmisorHook
			e.TipoEvento = domain.EventoTurnoFin
			e.CostoReportadoMicros = nil
			e.DuracionMs = i64(900)
		}),
		// Otro turno de la misma sesión: NO se fusiona con el primero.
		evento("s-1", "t-2", func(e *domain.EventoTelemetria) { e.CostoReportadoMicros = i64(50) }),
	)
	turnos, err := s.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 2 {
		t.Fatalf("se esperaban 2 turnos unidos, hay %d", len(turnos))
	}
	var t1 *domain.TurnoUnido
	for i := range turnos {
		if turnos[i].TurnoID == "t-1" {
			t1 = &turnos[i]
		}
	}
	if t1 == nil {
		t.Fatal("no se encontró el turno t-1")
	}
	if !t1.TieneDinero || !t1.TieneProceso {
		t.Errorf("t-1 debe tener las dos mitades: dinero=%v proceso=%v", t1.TieneDinero, t1.TieneProceso)
	}
	if t1.CostoReportadoMicros == nil || *t1.CostoReportadoMicros != 1234 {
		t.Errorf("el costo del turno: %v", t1.CostoReportadoMicros)
	}
	if t1.DuracionMs == nil || *t1.DuracionMs != 900 {
		t.Errorf("la duración del turno: %v", t1.DuracionMs)
	}
	// Los buckets que ningún lado trajo siguen nil.
	if t1.Tokens.Razonamiento != nil {
		t.Error("un bucket que nadie trajo sigue nil tras el join, no 0")
	}
}

// TestTurnoSoloConProcesoOSoloConDinero — las dos mitades son independientes y viajan
// declaradas. Un turno con solo una de las dos es un DATO, no un error.
func TestTurnoSoloConProcesoOSoloConDinero(t *testing.T) {
	s := nuevoStore(t, Opciones{})
	ingerirYEsperar(t, s,
		evento("s-1", "solo-dinero", func(e *domain.EventoTelemetria) { e.CostoReportadoMicros = i64(10) }),
		evento("s-1", "solo-proceso", func(e *domain.EventoTelemetria) {
			e.Emisor = domain.EmisorHook
			e.TipoEvento = domain.EventoTurnoFin
		}),
	)
	turnos, err := s.Turnos(context.Background(), ports.ConsultaTelemetria{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tu := range turnos {
		switch tu.TurnoID {
		case "solo-dinero":
			if !tu.TieneDinero || tu.TieneProceso {
				t.Errorf("solo-dinero: dinero=%v proceso=%v", tu.TieneDinero, tu.TieneProceso)
			}
		case "solo-proceso":
			if tu.TieneDinero || !tu.TieneProceso {
				t.Errorf("solo-proceso: dinero=%v proceso=%v", tu.TieneDinero, tu.TieneProceso)
			}
			if tu.CostoReportadoMicros != nil {
				t.Error("un turno sin dinero tiene costo null, no 0")
			}
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
