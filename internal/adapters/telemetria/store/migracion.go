package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// migracion.go implementa el versionado del esquema (decisión A3). **Dos números, no uno**, y
// la diferencia es la razón de existir de este archivo:
//
//   - `version` crece por migraciones **aditivas**: `CREATE TABLE`, `CREATE INDEX`,
//     `ALTER TABLE … ADD COLUMN`. Se aplican en orden, dentro de una transacción.
//   - `generacion` cambia SOLO ante una ruptura que no se puede expresar de forma aditiva. Y
//     entonces el store **archiva** el archivo viejo y arranca vacío. **Nunca borra.**
//
// Por qué NO se calca el wipe del índice: `index.db` se puede reconstruir del árbol de
// arneses; esto no. El runtime no guarda la historia de telemetría en ningún otro lado — si se
// pierde, se perdió. Archivar cuesta un `os.Rename` y conserva la posibilidad de recuperarla a
// mano.

// generacionActual es el número de generación que este binario entiende. Bumpearlo hace que
// todo `.db` de una generación anterior se archive al abrirlo.
const generacionActual = 1

// Migracion es un escalón del esquema. `SQL` **solo puede ser aditivo** —
// `TestMigracionesSonAditivas` escanea la lista y falla si aparece `DROP`, `RENAME` o
// `ALTER … DROP COLUMN`, para que el merge no compile una migración destructiva.
type Migracion struct {
	Version int
	SQL     []string
}

// migraciones es la lista, en orden. Agregar un escalón es agregar una entrada acá; NUNCA
// editar una ya aplicada, que dejaría bases idénticas en versión y distintas en forma.
var migraciones = []Migracion{
	{Version: 1, SQL: []string{
		// evento: la fila cruda. UNA tabla para los cuatro emisores, a propósito:
		//   · el join dinero×proceso es un self-join por (sesion_id, turno_id) — un índice,
		//     no dos tablas que se pueden desincronizar;
		//   · la purga y el borrado por arnés son UNA sentencia;
		//   · las columnas de dinero quedan NULL en las filas de proceso, que es exactamente
		//     «no aplica» a nivel de almacenamiento (los hooks NO traen dinero, ANEXO H2).
		`CREATE TABLE IF NOT EXISTS evento (
  id                     INTEGER PRIMARY KEY AUTOINCREMENT,
  ts_recibido            TEXT    NOT NULL,
  ts_emisor              TEXT,
  reloj_sospechoso       INTEGER NOT NULL DEFAULT 0,
  emisor                 TEXT    NOT NULL,
  runtime                TEXT    NOT NULL,
  runtime_version        TEXT,
  adaptador_version      TEXT    NOT NULL,
  sesion_id              TEXT    NOT NULL,
  turno_id               TEXT,
  arnes_id               TEXT,
  instalacion_id         TEXT,
  caja_id                TEXT,
  corrida_id             TEXT,
  atribucion             TEXT    NOT NULL,
  plugin_id_hash         TEXT,
  cwd_huella             TEXT,
  modelo                 TEXT,
  modelo_canonico        TEXT,
  proveedor              TEXT,
  speed                  TEXT,
  service_tier           TEXT,
  tok_entrada            INTEGER,
  tok_salida             INTEGER,
  tok_cache_lectura      INTEGER,
  tok_cache_5m           INTEGER,
  tok_cache_1h           INTEGER,
  tok_razonamiento       INTEGER,
  aritmetica             TEXT,
  acumulacion            TEXT,
  costo_reportado_micros INTEGER,
  costo_calculado_micros INTEGER,
  costo_completo         INTEGER,
  catalogo_version       TEXT,
  duracion_ms            INTEGER,
  escenario              TEXT    NOT NULL,
  tipo_evento            TEXT    NOT NULL,
  resultado              TEXT,
  gate                   TEXT,
  motivo                 TEXT,
  herramienta            TEXT,
  decision               TEXT,
  tool_input_bytes       INTEGER,
  tool_result_bytes      INTEGER
) STRICT`,
		// EL índice del módulo: la consulta que une dinero y proceso (ANEXO H1) es
		// `WHERE sesion_id = ? AND turno_id = ?`.
		`CREATE INDEX IF NOT EXISTS idx_evento_join    ON evento (sesion_id, turno_id)`,
		`CREATE INDEX IF NOT EXISTS idx_evento_ventana ON evento (ts_recibido)`,
		`CREATE INDEX IF NOT EXISTS idx_evento_arnes   ON evento (arnes_id, instalacion_id, ts_recibido)`,
		`CREATE INDEX IF NOT EXISTS idx_evento_caja    ON evento (arnes_id, caja_id, ts_recibido)`,
		// Dedupe: el mismo evento reenviado por un emisor que se reinició no se cuenta dos
		// veces. La llave es (sesion, turno, tipo, ts_emisor) — el ts_recibido NO entra
		// porque es distinto en el reenvío, que es justo lo que hay que ignorar.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_evento_dedupe
  ON evento (sesion_id, COALESCE(turno_id,''), tipo_evento, COALESCE(ts_emisor,''), emisor, COALESCE(herramienta,''))`,
		// rollup_hora: la proyección que hace que el tablero cueste milisegundos. Las sumas
		// son NULLABLE a propósito: SUM() sobre un grupo donde nadie tenía el bucket
		// devuelve NULL, y así el «no aplica» sobrevive a la agregación.
		`CREATE TABLE IF NOT EXISTS rollup_hora (
  hora                   TEXT    NOT NULL,
  arnes_id               TEXT    NOT NULL DEFAULT '',
  instalacion_id         TEXT    NOT NULL DEFAULT '',
  caja_id                TEXT    NOT NULL DEFAULT '',
  runtime                TEXT    NOT NULL,
  modelo_canonico        TEXT    NOT NULL DEFAULT '',
  emisor                 TEXT    NOT NULL,
  atribucion             TEXT    NOT NULL,
  eventos                INTEGER NOT NULL,
  turnos                 INTEGER NOT NULL,
  tok_entrada            INTEGER,
  tok_salida             INTEGER,
  tok_cache_lectura      INTEGER,
  tok_cache_5m           INTEGER,
  tok_cache_1h           INTEGER,
  tok_razonamiento       INTEGER,
  costo_reportado_micros INTEGER,
  costo_calculado_micros INTEGER,
  duracion_ms_suma       INTEGER,
  duracion_ms_cuenta     INTEGER,
  rechazados             INTEGER NOT NULL DEFAULT 0,
  cardinalidad_colapsada INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion)
) STRICT, WITHOUT ROWID`,
		// Cursor del rollup: hasta qué evento.id se agregó. Reanudable tras una caída.
		`CREATE TABLE IF NOT EXISTS rollup_cursor (
  unico         INTEGER PRIMARY KEY CHECK (unico = 1),
  ultimo_evento INTEGER NOT NULL,
  actualizado   TEXT    NOT NULL
) STRICT`,
		// atribucion_hash: la tabla `plugin_id_hash → arnés` que rescata la atribución cuando
		// el runtime redacta el nombre del plugin. Se APRENDE del spawn controlado: nosotros
		// instalamos el arnés, así que podemos observar qué hash le corresponde.
		`CREATE TABLE IF NOT EXISTS atribucion_hash (
  plugin_id_hash   TEXT PRIMARY KEY,
  arnes_id         TEXT NOT NULL,
  instalacion_id   TEXT NOT NULL DEFAULT '',
  visto            TEXT NOT NULL,
  como_se_aprendio TEXT NOT NULL
) STRICT`,
		// turno_esperado: el DENOMINADOR de la cobertura (A9). Lo escribe el daemon cuando el
		// stream-json declara un turno; `medido` se marca cuando llega su api_request. La
		// diferencia es el «no llegaron» honesto — sin esto, «cuánto medimos» se leería como
		// «cuánto hubo».
		`CREATE TABLE IF NOT EXISTS turno_esperado (
  sesion_id TEXT NOT NULL,
  turno_id  TEXT NOT NULL,
  arnes_id  TEXT,
  caja_id   TEXT,
  ts        TEXT NOT NULL,
  medido    INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (sesion_id, turno_id)
) STRICT, WITHOUT ROWID`,
		// salud: contadores del receptor. Se PERSISTEN para que sobrevivan a un reinicio —
		// «cuántos descarté» es dato de honestidad, no un gauge en memoria que se pierde.
		`CREATE TABLE IF NOT EXISTS salud (
  clave       TEXT PRIMARY KEY,
  valor       INTEGER NOT NULL,
  actualizado TEXT NOT NULL
) STRICT`,
	}},
}

// Migraciones expone la lista para el test que verifica que son aditivas. Devuelve una copia
// superficial: nadie de afuera reordena la lista.
func Migraciones() []Migracion { return append([]Migracion(nil), migraciones...) }

// GeneracionActual expone la generación para los tests y para el reporte de salud.
func GeneracionActual() int { return generacionActual }

// patronesDestructivos son las formas de SQL que una migración aditiva NO puede tener.
var patronesDestructivos = []string{"drop table", "drop index", "drop column", "rename to",
	"rename column", "delete from", "truncate"}

// EsAditiva reporta si una sentencia es aditiva. Se expone porque el test la usa sobre la
// lista real Y sobre un control positivo: un escáner que no encuentra nada daría verde por
// vacío.
func EsAditiva(sqlTexto string) bool {
	bajo := strings.ToLower(sqlTexto)
	for _, p := range patronesDestructivos {
		if strings.Contains(bajo, p) {
			return false
		}
	}
	return true
}

// Aplicar lleva el esquema a la versión más alta declarada, dentro de UNA transacción por
// escalón. Devuelve la generación y versión resultantes.
//
// Un mismatch de GENERACIÓN no se resuelve acá: lo detecta `New` antes de llamar a esta
// función, porque la respuesta es archivar el archivo, no tocar su contenido.
func Aplicar(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_meta (
  generacion INTEGER NOT NULL,
  version    INTEGER NOT NULL,
  aplicado   TEXT    NOT NULL
) STRICT`); err != nil {
		return fmt.Errorf("store: schema_meta: %w", err)
	}

	gen, ver, err := leerMeta(ctx, db)
	if err != nil {
		return err
	}
	if gen == 0 {
		gen = generacionActual
	}

	for _, m := range migraciones {
		if m.Version <= ver {
			continue
		}
		tx, terr := db.BeginTx(ctx, nil)
		if terr != nil {
			return fmt.Errorf("store: migración %d begin: %w", m.Version, terr)
		}
		for _, s := range m.SQL {
			if _, eerr := tx.ExecContext(ctx, s); eerr != nil {
				_ = tx.Rollback()
				return fmt.Errorf("store: migración %d: %w", m.Version, eerr)
			}
		}
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM schema_meta`); eerr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("store: migración %d meta: %w", m.Version, eerr)
		}
		if _, eerr := tx.ExecContext(ctx,
			`INSERT INTO schema_meta (generacion, version, aplicado) VALUES (?,?,?)`,
			gen, m.Version, time.Now().UTC().Format(time.RFC3339)); eerr != nil {
			_ = tx.Rollback()
			return fmt.Errorf("store: migración %d meta: %w", m.Version, eerr)
		}
		if cerr := tx.Commit(); cerr != nil {
			return fmt.Errorf("store: migración %d commit: %w", m.Version, cerr)
		}
		ver = m.Version
	}
	return nil
}

// leerMeta devuelve (generacion, version) del archivo, o (0,0) si nunca se escribió.
func leerMeta(ctx context.Context, db *sql.DB) (gen, ver int, err error) {
	row := db.QueryRowContext(ctx, `SELECT generacion, version FROM schema_meta LIMIT 1`)
	switch err = row.Scan(&gen, &ver); {
	case err == sql.ErrNoRows:
		return 0, 0, nil
	case err != nil:
		return 0, 0, fmt.Errorf("store: leer schema_meta: %w", err)
	}
	return gen, ver, nil
}
