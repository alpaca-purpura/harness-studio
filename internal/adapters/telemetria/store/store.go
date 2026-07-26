// Package store persiste la telemetría en `~/.arnesia/telemetria.db` — una base **propia**,
// jamás dentro de `index.db` (decisión A2).
//
// Por qué separada, y no es una preferencia de organización:
//
//	                index.db                      telemetria.db
//	doctrina        DESECHABLE — un mismatch      NO reconstruible — el runtime no
//	                de schema borra el archivo    guarda esto en ningún lado; si se
//	                entero y se reconstruye       pierde, se perdió
//
// Compartir archivo significaría que **un bump del esquema del índice borra la historia de
// telemetría**. Y no hay ganancia: no existe una sola consulta que haga join SQL entre el
// grafo y la telemetría — ese join lo hace el caso de uso por `arnes_id`.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	_ "modernc.org/sqlite" // driver "sqlite": pure Go, sin CGO (check sin-cgo).
)

const driverName = "sqlite"

// Claves de la tabla `salud`. Se persisten (no son gauges en memoria) porque «cuántos
// descarté» es dato de honestidad y tiene que sobrevivir a un reinicio.
const (
	SaludRecibidos        = "recibidos"
	SaludAceptados        = "aceptados"
	SaludColaLlena        = "descartados_cola_llena"
	SaludFormato          = "rechazados_formato"
	SaludTamano           = "rechazados_tamano"
	SaludFueraDeLista     = "atributos_fuera_de_lista"
	SaludTemporalidad     = "temporalidad_no_soportada"
	SaludDuplicados       = "eventos_duplicados"
	SaludErroresEscritura = "errores_escritura"
)

// Opciones parametriza el store.
type Opciones struct {
	// RetencionDias es el TTL de la tabla cruda. ⚠️ El default de 90 es un valor
	// **PROPUESTO, no firmado** (J-6 · parada P2): D15.3 firmó «TTL por default» SIN número.
	// Viaja rotulado como propuesto en la salud y en el CLI.
	RetencionDias int
	// RollupMeses es el TTL del agregado.
	RollupMeses int
	// Cola es el tamaño del canal del writer. 0 = 4096 (§11).
	Cola int
	// LoteMax / LoteEspera: el writer escribe cuando junta LoteMax eventos o pasa
	// LoteEspera, lo que llegue primero.
	LoteMax    int
	LoteEspera time.Duration
	// Reloj permite fijar el tiempo en los tests. nil = time.Now.
	Reloj func() time.Time
}

func (o Opciones) conDefaults() Opciones {
	if o.RetencionDias <= 0 {
		o.RetencionDias = 90
	}
	if o.RollupMeses <= 0 {
		o.RollupMeses = 24
	}
	if o.Cola <= 0 {
		o.Cola = 4096
	}
	if o.LoteMax <= 0 {
		o.LoteMax = 256
	}
	if o.LoteEspera <= 0 {
		o.LoteEspera = 250 * time.Millisecond
	}
	if o.Reloj == nil {
		o.Reloj = time.Now
	}
	return o
}

// Store implementa ports.TelemetriaStore sobre modernc.org/sqlite.
//
// Dos handles como en `internal/adapters/index/store.go`, calcado a propósito: writer con
// `SetMaxOpenConns(1)` (database/sql serializa toda escritura, cero SQLITE_BUSY entre
// goroutines de este proceso) y reader pooled bajo WAL.
type Store struct {
	writer *sql.DB
	reader *sql.DB
	ruta   string
	opts   Opciones

	cola    chan []domain.EventoTelemetria
	cerrar  chan struct{}
	drenado sync.WaitGroup
	unaVez  sync.Once
	// encolados/escritos son la BARRERA real de `Sincronizar`. Antes era un sleep de
	// `LoteEspera + 50 ms`, que no es una barrera: es una apuesta a que el lote termine en
	// ese rato. Con 600 filas bajo `-race` la apuesta se pierde y el test falla por el
	// reloj, no por el código — la peor clase de test intermitente, porque enseña a
	// re-correr en vez de a mirar.
	encolados atomic.Int64
	escritos  atomic.Int64

	// archivada es el nombre del `.db` que se archivó al abrir, si hubo. Viaja a la salud:
	// una historia archivada que nadie menciona es una pérdida silenciosa.
	archivada string

	// trasLote lo llama el writer después de cada lote escrito. Lo usa el rollup para
	// engancharse sin que este archivo lo conozca.
	mu       sync.Mutex
	trasLote func()
}

var _ ports.TelemetriaStore = (*Store)(nil)

// New abre (o crea) la base. Una ruta vacía resuelve a `~/.arnesia/telemetria.db`.
//
// Dos rupturas se manejan ACÁ y las dos **archivan, nunca borran**:
//
//  1. **mismatch de generación** — el binario entiende otra forma de los datos;
//  2. **archivo corrupto** — el `.db` no abre ni migra.
//
// En los dos casos el archivo viejo queda en disco renombrado, la base nueva arranca vacía, y
// `Salud()` **lo dice**. Borrar en silencio para simplificar una migración sería destruir el
// único registro que existe.
func New(ruta string, o Opciones) (*Store, error) {
	o = o.conDefaults()
	ruta, err := resolverRuta(ruta)
	if err != nil {
		return nil, err
	}
	if dir := filepath.Dir(ruta); dir != "." {
		if merr := os.MkdirAll(dir, 0o750); merr != nil {
			return nil, fmt.Errorf("store: mkdir %s: %w", dir, merr)
		}
	}

	archivada, err := archivarSiCorresponde(ruta, o.Reloj())
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	writer, reader, err := abrirYMigrar(ctx, ruta)
	if err != nil {
		// El archivo no abre o no migra: es corrupción. **Se archiva, no se borra** — el
		// mismo trato que un mismatch de generación, por la misma razón: esto no se puede
		// reconstruir de ningún lado.
		if _, existe := os.Stat(ruta); existe != nil {
			return nil, err // no es corrupción: el problema es del directorio, y eso sí sube.
		}
		nombre, rerr := archivar(ruta, "corrupta", o.Reloj())
		if rerr != nil {
			return nil, fmt.Errorf("store: %w (y no se pudo archivar: %v)", err, rerr)
		}
		slog.Warn("telemetria: base ilegible — se archivó y se arranca vacía",
			"archivada", nombre, "err", err)
		archivada = nombre
		if writer, reader, err = abrirYMigrar(ctx, ruta); err != nil {
			return nil, fmt.Errorf("store: esquema tras archivar: %w", err)
		}
	}

	s := &Store{
		writer: writer, reader: reader, ruta: ruta, opts: o,
		cola:      make(chan []domain.EventoTelemetria, o.Cola),
		cerrar:    make(chan struct{}),
		archivada: archivada,
	}
	s.drenado.Add(1)
	go s.drenar()
	return s, nil
}

// ReaderParaTest expone el handle de lectura para los tests de fitness, que necesitan
// verificar la FORMA de las filas —no solo lo que las consultas devuelven—. Un test que solo
// mira la salida de la consulta no puede distinguir «el dato no suma» de «el dato no está».
func (s *Store) ReaderParaTest() *sql.DB { return s.reader }

// Ruta devuelve el archivo que este store abrió. Lo usan el reporte de salud y el test que
// verifica que jamás es `index.db`.
func (s *Store) Ruta() string { return s.ruta }

// Archivada devuelve el nombre del `.db` archivado al abrir, o "".
func (s *Store) Archivada() string { return s.archivada }

// AlEscribirLote registra un callback que corre después de cada lote persistido. Lo usa el
// rollup; el store no sabe qué es un rollup.
func (s *Store) AlEscribirLote(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trasLote = f
}

// Ingerir encola los eventos. **No bloquea**: si la cola está llena descarta, lo CUENTA y lo
// dice en el valor de retorno — la diferencia con `len(evs)` son descartes contados, no
// silencio. Un receptor que hace esperar al agente degradaría el trabajo que mide.
func (s *Store) Ingerir(ctx context.Context, evs []domain.EventoTelemetria) (int, error) {
	if len(evs) == 0 {
		return 0, nil
	}
	// Se validan ANTES de encolar: un evento sin sesión no es atribuible ni deduplicable, y
	// encolarlo solo para tirarlo después escondería el descarte en el writer.
	buenos := make([]domain.EventoTelemetria, 0, len(evs))
	for _, e := range evs {
		if e.SesionID == "" {
			continue
		}
		buenos = append(buenos, e)
	}
	if len(buenos) == 0 {
		return 0, domain.ErrEventoSinSesion
	}
	select {
	case s.cola <- buenos:
		s.encolados.Add(int64(len(buenos)))
		return len(buenos), nil
	default:
		s.sumarSalud(ctx, SaludColaLlena, int64(len(buenos)))
		return 0, ports.ErrColaLlena
	}
}

// Sincronizar espera a que TODO lo encolado esté escrito. Existe para los tests y para el
// shutdown ordenado; el camino caliente nunca la llama.
//
// Es una barrera de verdad —compara un contador de encolados contra uno de escritos— y no un
// sleep: un sleep es una apuesta a que el lote termine a tiempo, y una apuesta que se pierde
// bajo carga produce un test intermitente, que es peor que un test que falla.
func (s *Store) Sincronizar(ctx context.Context) error {
	for {
		if s.escritos.Load() >= s.encolados.Load() {
			return nil
		}
		select {
		case <-time.After(2 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Close drena lo pendiente y libera los dos handles. A diferencia del índice, este `.db` NO
// es desechable: cerrar bien importa.
func (s *Store) Close() error {
	s.unaVez.Do(func() { close(s.cerrar) })
	s.drenado.Wait()
	werr := s.writer.Close()
	rerr := s.reader.Close()
	if werr != nil {
		return werr
	}
	return rerr
}

// drenar es la goroutine del writer: junta hasta LoteMax eventos o espera LoteEspera, lo que
// llegue primero, y escribe el lote en UNA transacción.
func (s *Store) drenar() {
	defer s.drenado.Done()
	tick := time.NewTicker(s.opts.LoteEspera)
	defer tick.Stop()

	var lote []domain.EventoTelemetria
	escribir := func() {
		if len(lote) == 0 {
			return
		}
		// El contador sube pase lo que pase: un lote que no se pudo escribir es una pérdida
		// contada (`errores_escritura`), no una razón para que la barrera espere para siempre.
		// `n` se captura ACÁ: más abajo `lote` se vacía, y un `len(lote)` diferido sumaría 0.
		n := int64(len(lote))
		defer func() { s.escritos.Add(n) }()
		if err := s.escribirLote(context.Background(), lote); err != nil {
			// **El daemon NO se cae por esto** (disco lleno, base bloqueada): se pierde
			// telemetría, no la sesión del usuario. Y se cuenta, para que la pérdida sea
			// visible en la salud en vez de silenciosa.
			slog.Warn("telemetria: lote no escrito", "eventos", len(lote), "err", err)
			s.sumarSalud(context.Background(), SaludErroresEscritura, int64(len(lote)))
		}
		lote = lote[:0]
		s.mu.Lock()
		f := s.trasLote
		s.mu.Unlock()
		if f != nil {
			f()
		}
	}

	for {
		select {
		case evs := <-s.cola:
			lote = append(lote, evs...)
			if len(lote) >= s.opts.LoteMax {
				escribir()
			}
		case <-tick.C:
			escribir()
		case <-s.cerrar:
			// Drenado final: lo que quedó en la cola se escribe antes de cerrar.
			for {
				select {
				case evs := <-s.cola:
					lote = append(lote, evs...)
				default:
					escribir()
					return
				}
				if len(lote) >= s.opts.LoteMax {
					escribir()
				}
			}
		}
	}
}

const insertEvento = `INSERT OR IGNORE INTO evento (
  ts_recibido, ts_emisor, reloj_sospechoso, emisor, runtime, runtime_version, adaptador_version,
  sesion_id, turno_id, arnes_id, instalacion_id, caja_id, corrida_id, atribucion,
  plugin_id_hash, cwd_huella, modelo, modelo_canonico, proveedor, speed, service_tier,
  tok_entrada, tok_salida, tok_cache_lectura, tok_cache_5m, tok_cache_1h, tok_razonamiento,
  tok_cache_sin_tier, aritmetica, acumulacion, costo_reportado_micros, costo_calculado_micros, costo_completo,
  catalogo_version, duracion_ms, escenario, tipo_evento, resultado, gate, motivo, herramienta,
  decision, tool_input_bytes, tool_result_bytes
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`

func (s *Store) escribirLote(ctx context.Context, evs []domain.EventoTelemetria) error {
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	st, err := tx.PrepareContext(ctx, insertEvento)
	if err != nil {
		return fmt.Errorf("store: prepare: %w", err)
	}
	defer func() { _ = st.Close() }()

	var insertados, duplicados int64
	for _, e := range evs {
		sospechoso := 0
		if e.RelojSospechoso || relojSospechoso(e) {
			sospechoso = 1
		}
		res, eerr := st.ExecContext(ctx,
			e.TSRecibido.UTC().Format(time.RFC3339Nano), tsOpcional(e.TSEmisor), sospechoso,
			string(e.Emisor), e.Runtime, textoOpcional(e.RuntimeVersion), e.AdaptadorVersion,
			e.SesionID, textoOpcional(e.TurnoID), textoOpcional(e.ArnesID),
			textoOpcional(e.InstalacionID), textoOpcional(e.CajaID), textoOpcional(e.CorridaID),
			string(e.Atribucion), textoOpcional(e.PluginIDHash), textoOpcional(e.CWDHuella),
			textoOpcional(e.Modelo), textoOpcional(e.ModeloCanonico), textoOpcional(e.Proveedor),
			textoOpcional(e.Speed), textoOpcional(e.ServiceTier),
			e.Tokens.Entrada, e.Tokens.Salida, e.Tokens.CacheLectura,
			e.Tokens.CacheEscritura5m, e.Tokens.CacheEscritura1h, e.Tokens.Razonamiento,
			e.Tokens.CacheEscrituraSinTier,
			textoOpcional(string(e.Aritmetica)), textoOpcional(string(e.Acumulacion)),
			e.CostoReportadoMicros, e.CostoCalculadoMicros, boolOpcional(e.CostoCompleto),
			textoOpcional(e.CatalogoVersion), e.DuracionMs,
			string(e.Escenario), string(e.TipoEvento), textoOpcional(string(e.Resultado)),
			textoOpcional(e.Gate), textoOpcional(e.Motivo), textoOpcional(e.Herramienta),
			textoOpcional(e.Decision), e.ToolInputBytes, e.ToolResultBytes,
		)
		if eerr != nil {
			return fmt.Errorf("store: insert: %w", eerr)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			// `INSERT OR IGNORE` + índice único = dedupe. Un emisor que se reinicia y
			// re-manda su lote no infla los totales (escenario C11).
			duplicados++
		} else {
			insertados++
		}
		// El api_request de un turno esperado lo marca como medido: es la mitad viva de la
		// conciliación de cobertura (A9).
		if e.TipoEvento == domain.EventoAPIRequest && e.TurnoID != "" {
			if _, uerr := tx.ExecContext(ctx,
				`UPDATE turno_esperado SET medido = 1 WHERE sesion_id = ? AND turno_id = ?`,
				e.SesionID, e.TurnoID); uerr != nil {
				return fmt.Errorf("store: marcar turno medido: %w", uerr)
			}
		}
	}
	if err := sumarSaludTx(ctx, tx, SaludAceptados, insertados, s.opts.Reloj()); err != nil {
		return err
	}
	if duplicados > 0 {
		if err := sumarSaludTx(ctx, tx, SaludDuplicados, duplicados, s.opts.Reloj()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// relojSospechoso marca los `ts_emisor` fuera de `[ts_recibido − 24 h, ts_recibido + 5 min]`
// (A10). No se descarta el evento —el reloj del daemon manda igual— pero la fila queda
// marcada: un timestamp de emisor absurdo explica después una ventana rara.
func relojSospechoso(e domain.EventoTelemetria) bool {
	if e.TSEmisor == nil {
		return false
	}
	d := e.TSEmisor.Sub(e.TSRecibido)
	return d > 5*time.Minute || d < -24*time.Hour
}

// EsperarTurno registra que un turno OCURRIÓ, se sepa medir o no. Es el denominador de la
// cobertura: sin esto, «cuánto medimos» se leería como «cuánto hubo».
func (s *Store) EsperarTurno(ctx context.Context, sesionID, turnoID, arnesID, cajaID string) error {
	if sesionID == "" || turnoID == "" {
		return domain.ErrEventoSinSesion
	}
	_, err := s.writer.ExecContext(ctx,
		`INSERT INTO turno_esperado (sesion_id, turno_id, arnes_id, caja_id, ts, medido)
		 VALUES (?,?,?,?,?,COALESCE((SELECT medido FROM turno_esperado WHERE sesion_id=? AND turno_id=?),0))
		 ON CONFLICT(sesion_id, turno_id) DO UPDATE SET arnes_id=excluded.arnes_id, caja_id=excluded.caja_id`,
		sesionID, turnoID, textoOpcional(arnesID), textoOpcional(cajaID),
		s.opts.Reloj().UTC().Format(time.RFC3339Nano), sesionID, turnoID)
	if err != nil {
		return fmt.Errorf("store: esperar turno: %w", err)
	}
	return nil
}

// AprenderHash registra un par `plugin_id_hash → arnés` observado. `como` documenta de dónde
// salió, para poder explicar después una atribución rara en vez de tener que adivinarla.
func (s *Store) AprenderHash(ctx context.Context, hash, arnesID, instalacionID, como string) error {
	if hash == "" || arnesID == "" {
		return errors.New("store: hash y arnés son obligatorios para aprender una atribución")
	}
	_, err := s.writer.ExecContext(ctx,
		`INSERT INTO atribucion_hash (plugin_id_hash, arnes_id, instalacion_id, visto, como_se_aprendio)
		 VALUES (?,?,?,?,?)
		 ON CONFLICT(plugin_id_hash) DO UPDATE SET arnes_id=excluded.arnes_id,
		   instalacion_id=excluded.instalacion_id, visto=excluded.visto,
		   como_se_aprendio=excluded.como_se_aprendio`,
		hash, arnesID, instalacionID, s.opts.Reloj().UTC().Format(time.RFC3339Nano), como)
	if err != nil {
		return fmt.Errorf("store: aprender hash: %w", err)
	}
	return nil
}

// BuscarHash resuelve un `plugin_id_hash` a su arnés aprendido.
func (s *Store) BuscarHash(ctx context.Context, hash string) (arnesID, instalacionID string, ok bool) {
	row := s.reader.QueryRowContext(ctx,
		`SELECT arnes_id, instalacion_id FROM atribucion_hash WHERE plugin_id_hash = ?`, hash)
	if err := row.Scan(&arnesID, &instalacionID); err != nil {
		return "", "", false
	}
	return arnesID, instalacionID, true
}

// SumarSalud incrementa un contador persistido. Lo usan el receptor (descartes de formato, de
// tamaño, de cola) y el writer.
func (s *Store) SumarSalud(ctx context.Context, clave string, n int64) {
	s.sumarSalud(ctx, clave, n)
}

func (s *Store) sumarSalud(ctx context.Context, clave string, n int64) {
	if n == 0 {
		return
	}
	if _, err := s.writer.ExecContext(ctx,
		`INSERT INTO salud (clave, valor, actualizado) VALUES (?,?,?)
		 ON CONFLICT(clave) DO UPDATE SET valor = salud.valor + excluded.valor, actualizado = excluded.actualizado`,
		clave, n, s.opts.Reloj().UTC().Format(time.RFC3339Nano)); err != nil {
		slog.Warn("telemetria: contador de salud no persistido", "clave", clave, "err", err)
	}
}

func sumarSaludTx(ctx context.Context, tx *sql.Tx, clave string, n int64, ahora time.Time) error {
	if n == 0 {
		return nil
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO salud (clave, valor, actualizado) VALUES (?,?,?)
		 ON CONFLICT(clave) DO UPDATE SET valor = salud.valor + excluded.valor, actualizado = excluded.actualizado`,
		clave, n, ahora.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("store: salud %s: %w", clave, err)
	}
	return nil
}

// Salud arma el reporte. `RetencionPropuesta` viaja en true mientras el número no esté
// firmado (J-6): la UI y el CLI lo rotulan como propuesto en vez de presentarlo como política.
func (s *Store) Salud(ctx context.Context) (domain.SaludTelemetria, error) {
	out := domain.SaludTelemetria{
		AlmacenDisponible:  true,
		RetencionDias:      s.opts.RetencionDias,
		RetencionPropuesta: true,
		RollupMeses:        s.opts.RollupMeses,
		HistoriaArchivada:  s.archivada,
	}
	rows, err := s.reader.QueryContext(ctx, `SELECT clave, valor FROM salud`)
	if err != nil {
		out.AlmacenDisponible = false
		out.AlmacenMotivo = err.Error()
		return out, nil //nolint:nilerr // el daemon sigue vivo; la salud lo DICE.
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var clave string
		var valor int64
		if serr := rows.Scan(&clave, &valor); serr != nil {
			return out, fmt.Errorf("store: salud scan: %w", serr)
		}
		switch clave {
		case SaludRecibidos:
			out.Recibidos = int(valor)
		case SaludAceptados:
			out.Aceptados = int(valor)
		case SaludColaLlena:
			out.DescartadosColaLlena = int(valor)
		case SaludFormato:
			out.RechazadosFormato = int(valor)
		case SaludTamano:
			out.RechazadosTamano = int(valor)
		case SaludFueraDeLista:
			out.AtributosFueraDeLista = int(valor)
		case SaludTemporalidad:
			out.TemporalidadNoSoportada = int(valor)
		}
	}
	if rerr := rows.Err(); rerr != nil {
		return out, fmt.Errorf("store: salud rows: %w", rerr)
	}

	var ultima sql.NullString
	if err := s.reader.QueryRowContext(ctx, `SELECT MAX(ts_recibido) FROM evento`).Scan(&ultima); err == nil && ultima.Valid {
		if t, perr := time.Parse(time.RFC3339Nano, ultima.String); perr == nil {
			u := t.UTC()
			out.UltimaRecepcion = &u
		}
	}
	if fi, serr := os.Stat(s.ruta); serr == nil {
		out.TamanoBytes = fi.Size()
		// Aviso a partir de 500 MB (§11): el umbral de incomodidad medido está ~940 MB, así
		// que avisar a la mitad deja margen para actuar.
		out.AvisoTamano = fi.Size() > 500<<20
	}
	return out, nil
}

// Purgar aplica el TTL o borra un arnés. **Los dos modos borran también el agregado**, en la
// MISMA transacción: si no, quedaría una cifra en el tablero alimentándose de filas que ya no
// existen. Devuelve cuántas filas de `evento` se fueron.
func (s *Store) Purgar(ctx context.Context, p ports.PurgaTelemetria) (int64, error) {
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("store: purga begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var n int64
	switch {
	case p.ArnesID != "":
		res, eerr := tx.ExecContext(ctx, `DELETE FROM evento WHERE arnes_id = ?`, p.ArnesID)
		if eerr != nil {
			return 0, fmt.Errorf("store: purga por arnés: %w", eerr)
		}
		n, _ = res.RowsAffected()
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM rollup_hora WHERE arnes_id = ?`, p.ArnesID); eerr != nil {
			return 0, fmt.Errorf("store: purga rollup por arnés: %w", eerr)
		}
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM turno_esperado WHERE arnes_id = ?`, p.ArnesID); eerr != nil {
			return 0, fmt.Errorf("store: purga turnos por arnés: %w", eerr)
		}
	case !p.AntesDe.IsZero():
		corte := p.AntesDe.UTC().Format(time.RFC3339Nano)
		res, eerr := tx.ExecContext(ctx, `DELETE FROM evento WHERE ts_recibido < ?`, corte)
		if eerr != nil {
			return 0, fmt.Errorf("store: purga por TTL: %w", eerr)
		}
		n, _ = res.RowsAffected()
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM turno_esperado WHERE ts < ?`, corte); eerr != nil {
			return 0, fmt.Errorf("store: purga turnos por TTL: %w", eerr)
		}
		// 🔴 El agregado de las horas purgadas se va CON el detalle.
		//
		// La versión anterior lo conservaba, apoyada en una promesa que era falsa: «el rollup
		// sobrevive más que el detalle, así que tras purgar el resumen se conserva». No se
		// conservaba nada — **ninguna consulta de lectura toca `rollup_hora`** (A2 de la
		// auditoría), así que lo único que quedaba eran filas que el producto no muestra,
		// afirmando una retención que no existe.
		//
		// Entre dejar la mentira y sacarla, se saca: el agregado no puede afirmar que
		// conserva algo que ninguna pantalla puede devolver. El día que el rollup entre al
		// camino de lectura, esta línea se revierte **y la promesa pasa a ser verdad** —
		// deuda registrada en el BACKLOG y en CAP-123, que por eso está en `parcial`.
		if _, eerr := tx.ExecContext(ctx,
			`DELETE FROM rollup_hora WHERE hora < ?`, p.AntesDe.UTC().Format("2006-01-02T15")); eerr != nil {
			return 0, fmt.Errorf("store: purga rollup por TTL: %w", eerr)
		}
	default:
		return 0, errors.New("store: purga sin criterio (ni TTL ni arnés)")
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("store: purga commit: %w", err)
	}
	return n, nil
}

// PurgarRollup aplica el TTL del agregado (por default 24 meses). Va aparte del de la tabla
// cruda a propósito: el agregado sobrevive al detalle.
func (s *Store) PurgarRollup(ctx context.Context, antesDe time.Time) (int64, error) {
	res, err := s.writer.ExecContext(ctx, `DELETE FROM rollup_hora WHERE hora < ?`,
		antesDe.UTC().Format("2006-01-02T15"))
	if err != nil {
		return 0, fmt.Errorf("store: purga rollup: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// ── apertura, rutas y archivado ──────────────────────────────────────────────────────────

func resolverRuta(ruta string) (string, error) {
	if ruta != "" {
		return ruta, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("store: resolver home: %w", err)
	}
	// `telemetria.db`, hermana de `index.db` pero **archivo propio** (A2).
	return filepath.Join(home, ".arnesia", "telemetria.db"), nil
}

// abrirHandles abre el par writer/reader con los MISMOS pragmas que el índice.
func abrirHandles(ruta string) (writer, reader *sql.DB, err error) {
	writerDSN := "file:" + ruta +
		"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate"
	writer, err = sql.Open(driverName, writerDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("store: abrir writer: %w", err)
	}
	// Un solo escritor: database/sql serializa todo Exec sobre él, así que no hay
	// SQLITE_BUSY entre goroutines de este proceso.
	writer.SetMaxOpenConns(1)

	reader, err = sql.Open(driverName, "file:"+ruta+"?_pragma=busy_timeout(5000)")
	if err != nil {
		_ = writer.Close()
		return nil, nil, fmt.Errorf("store: abrir reader: %w", err)
	}
	if perr := writer.Ping(); perr != nil {
		_ = writer.Close()
		_ = reader.Close()
		return nil, nil, fmt.Errorf("store: ping: %w", perr)
	}
	return writer, reader, nil
}

// abrirYMigrar abre el par de handles y lleva el esquema a la versión actual. Devuelve error
// —con los handles ya cerrados— si el archivo no es una base o si la migración falla; el
// caller decide si eso amerita archivar.
func abrirYMigrar(ctx context.Context, ruta string) (writer, reader *sql.DB, err error) {
	writer, reader, err = abrirHandles(ruta)
	if err != nil {
		return nil, nil, err
	}
	if aerr := Aplicar(ctx, writer); aerr != nil {
		_ = writer.Close()
		_ = reader.Close()
		return nil, nil, aerr
	}
	return writer, reader, nil
}

// archivarSiCorresponde mira la generación del archivo existente. Un mismatch archiva; el
// archivo nuevo arranca vacío. **Nunca borra.**
func archivarSiCorresponde(ruta string, ahora time.Time) (string, error) {
	if _, err := os.Stat(ruta); err != nil {
		return "", nil //nolint:nilerr // no hay archivo: nada que archivar.
	}
	db, err := sql.Open(driverName, "file:"+ruta+"?_pragma=busy_timeout(2000)")
	if err != nil {
		return "", fmt.Errorf("store: inspeccionar generación: %w", err)
	}
	var gen int
	row := db.QueryRow(`SELECT generacion FROM schema_meta LIMIT 1`)
	scanErr := row.Scan(&gen)
	_ = db.Close()
	if scanErr != nil {
		// Sin `schema_meta` legible puede ser una base nueva a medio crear o una corrupta.
		// No se archiva acá: se deja que `Aplicar` lo intente y, si falla, `New` archiva con
		// el motivo correcto. Archivar por las dudas destruiría una base sana.
		return "", nil
	}
	if gen == generacionActual {
		return "", nil
	}
	nombre, err := archivar(ruta, fmt.Sprintf("gen%d", gen), ahora)
	if err != nil {
		return "", err
	}
	slog.Warn("telemetria: generación distinta — la historia se ARCHIVA, no se borra",
		"generacion_archivo", gen, "generacion_binario", generacionActual, "archivada", nombre)
	return nombre, nil
}

// archivar renombra el `.db` y sus sidecars `-wal`/`-shm`. Devuelve el nombre del archivo
// resultante (no la ruta completa: es lo que la UI muestra).
func archivar(ruta, sufijo string, ahora time.Time) (string, error) {
	base := strings.TrimSuffix(ruta, ".db")
	destino := fmt.Sprintf("%s-%s-%s.db.archivada", base, sufijo, ahora.UTC().Format("2006-01-02T150405"))
	if err := os.Rename(ruta, destino); err != nil {
		return "", fmt.Errorf("store: archivar %s: %w", ruta, err)
	}
	for _, side := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(ruta + side); err == nil {
			_ = os.Rename(ruta+side, destino+side)
		}
	}
	return filepath.Base(destino), nil
}

// ── helpers de nulos ─────────────────────────────────────────────────────────────────────
//
// Los tres existen por la MISMA razón: una cadena vacía y un NULL no son lo mismo, y un
// puntero nil no puede convertirse en 0 al bajar a SQL.

func textoOpcional(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func tsOpcional(t *time.Time) any {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func boolOpcional(b *bool) any {
	if b == nil {
		return nil
	}
	if *b {
		return 1
	}
	return 0
}
