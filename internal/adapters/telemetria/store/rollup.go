package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// rollup.go es la proyección por hora que hace que el tablero cueste milisegundos en vez de
// cientos (190× medido). Es **incremental y reanudable**: `rollup_cursor` guarda el último
// `evento.id` agregado, así que una caída a mitad no obliga a recalcular todo ni cuenta nada
// dos veces.
//
// 🔴 La regla que define este archivo: **el `NULL` sobrevive a la agregación.** Un `SUM()`
// sobre un grupo donde ningún evento tenía el bucket devuelve `NULL`, y al acumularlo con
// `ON CONFLICT` hay que preservarlo — un `COALESCE(...,0)` liso convertiría cada ausencia en
// un cero medido y ya no habría forma de distinguirlos aguas abajo. De ahí el `CASE WHEN los
// dos lados son NULL THEN NULL ELSE suma END` de cada columna.

// topeCardinalidad es el máximo de filas de `rollup_hora` por mes (§11). La clave tiene ocho
// dimensiones: sin tope, un arnés con cajas generadas dinámicamente la explota.
//
// Al superarlo, `caja_id` colapsa a `(otros)` y la fila queda marcada
// `cardinalidad_colapsada` — degradación **declarada**, no silenciosa.
const topeCardinalidad = 50_000

// CajaColapsada es la etiqueta que reemplaza a `caja_id` cuando se supera el tope. Es visible
// a propósito: una fila que dice `(otros)` es honesta; una fila que se pierde, no.
const CajaColapsada = "(otros)"

// Rollup mantiene la proyección horaria de un Store.
type Rollup struct {
	s *Store

	mu        sync.Mutex
	debounce  *time.Timer
	espera    time.Duration
	colapsada bool
}

// NewRollup arma el rollup y lo engancha al writer del store: cada lote escrito dispara una
// actualización con debounce. `espera` en 0 usa 2 s (§5.5).
func NewRollup(s *Store, espera time.Duration) *Rollup {
	if espera <= 0 {
		espera = 2 * time.Second
	}
	r := &Rollup{s: s, espera: espera}
	s.AlEscribirLote(r.programar)
	return r
}

// programar agenda una actualización con debounce. Varios lotes seguidos producen UNA
// pasada: el rollup no es el camino caliente y no tiene por qué correr por cada lote.
func (r *Rollup) programar() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debounce != nil {
		r.debounce.Stop()
	}
	r.debounce = time.AfterFunc(r.espera, func() {
		if err := r.Actualizar(context.Background()); err != nil {
			slog.Warn("telemetria: rollup no actualizado", "err", err)
		}
	})
}

// Detener cancela el debounce pendiente. Se llama en el shutdown.
func (r *Rollup) Detener() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.debounce != nil {
		r.debounce.Stop()
		r.debounce = nil
	}
}

// sumaPreservandoNULL arma el fragmento `ON CONFLICT DO UPDATE` de una columna sumable.
//
// Es literalmente el corazón de `no-aplica-no-es-cero` a nivel de agregación: si los dos
// lados son NULL, el resultado sigue NULL. Solo cuando ALGUNO tiene dato se suma tratando al
// otro como 0 — que ahí sí es el neutro correcto, porque hay al menos una medición real.
func sumaPreservandoNULL(col string) string {
	return fmt.Sprintf(
		`%[1]s = CASE WHEN rollup_hora.%[1]s IS NULL AND excluded.%[1]s IS NULL THEN NULL
		              ELSE COALESCE(rollup_hora.%[1]s,0) + COALESCE(excluded.%[1]s,0) END`, col)
}

var columnasSumables = []string{
	"tok_entrada", "tok_salida", "tok_cache_lectura", "tok_cache_5m", "tok_cache_1h",
	"tok_razonamiento", "tok_cache_sin_tier", "costo_reportado_micros", "costo_calculado_micros",
	"duracion_ms_suma", "duracion_ms_cuenta",
}

// insertRollup se arma una vez: el SQL es fijo, solo el cursor cambia.
func insertRollup(colapsar bool) string {
	caja := "COALESCE(caja_id,'')"
	if colapsar {
		caja = "'" + CajaColapsada + "'"
	}
	sets := []string{
		"eventos = rollup_hora.eventos + excluded.eventos",
		"turnos  = rollup_hora.turnos  + excluded.turnos",
	}
	for _, c := range columnasSumables {
		sets = append(sets, sumaPreservandoNULL(c))
	}
	if colapsar {
		sets = append(sets, "cardinalidad_colapsada = 1")
	}
	conflicto := ""
	for i, s := range sets {
		if i > 0 {
			conflicto += ",\n  "
		}
		conflicto += s
	}
	return `
INSERT INTO rollup_hora (
  hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion,
  eventos, turnos, tok_entrada, tok_salida, tok_cache_lectura, tok_cache_5m, tok_cache_1h,
  tok_razonamiento, tok_cache_sin_tier, costo_reportado_micros, costo_calculado_micros,
  duracion_ms_suma, duracion_ms_cuenta, rechazados, cardinalidad_colapsada)
SELECT
  substr(ts_recibido, 1, 13),
  COALESCE(arnes_id,''), COALESCE(instalacion_id,''), ` + caja + `,
  runtime, COALESCE(modelo_canonico, modelo, ''), emisor, atribucion,
  COUNT(*),
  -- ⚠️ SESGO DECLARADO Y EN CONTRA (A3): COUNT(DISTINCT turno_id) SOBREESTIMA si un turno
  -- cruza la frontera horaria — el mismo turno cuenta en dos horas. Eso INFLA el
  -- denominador, o sea BAJA el costo-por-turno que mostramos. El sesgo va en contra de
  -- nuestra propia recomendación, que es la única dirección aceptable para un sesgo que no
  -- se puede eliminar.
  COUNT(DISTINCT turno_id),
  SUM(tok_entrada), SUM(tok_salida), SUM(tok_cache_lectura), SUM(tok_cache_5m),
  SUM(tok_cache_1h), SUM(tok_razonamiento), SUM(tok_cache_sin_tier),
  SUM(costo_reportado_micros), SUM(costo_calculado_micros),
  SUM(duracion_ms), SUM(CASE WHEN duracion_ms IS NOT NULL THEN 1 ELSE 0 END),
  0, ` + boolSQL(colapsar) + `
FROM evento
-- El canal SECUNDARIO queda afuera del agregado: su costo es el mismo del primario y
-- sumarlo contaría el gasto dos veces (ver domain.EventoMetrica).
WHERE id > ? AND id <= ? AND tipo_evento <> 'metrica'
GROUP BY 1,2,3,4,5,6,7,8
ON CONFLICT (hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion)
DO UPDATE SET
  ` + conflicto
}

func boolSQL(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// Actualizar agrega los eventos nuevos desde el cursor. Es idempotente: correrla dos veces
// seguidas no cambia nada, porque el cursor avanza dentro de la misma transacción.
func (r *Rollup) Actualizar(ctx context.Context) error {
	tx, err := r.s.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rollup: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var cursor int64
	row := tx.QueryRowContext(ctx, `SELECT ultimo_evento FROM rollup_cursor WHERE unico = 1`)
	switch err := row.Scan(&cursor); {
	case err == sql.ErrNoRows:
		cursor = 0
	case err != nil:
		return fmt.Errorf("rollup: leer cursor: %w", err)
	}

	// El techo se fija ANTES de agregar: si llegan eventos nuevos mientras corre, quedan
	// para la próxima pasada. Sin techo, el cursor podría saltarse filas escritas entre el
	// SELECT y el UPDATE.
	var techo sql.NullInt64
	if serr := tx.QueryRowContext(ctx, `SELECT MAX(id) FROM evento`).Scan(&techo); serr != nil {
		return fmt.Errorf("rollup: techo: %w", serr)
	}
	if !techo.Valid || techo.Int64 <= cursor {
		return nil // nada nuevo que agregar.
	}

	colapsar, err := r.debeColapsar(ctx, tx)
	if err != nil {
		return err
	}
	if _, eerr := tx.ExecContext(ctx, insertRollup(colapsar), cursor, techo.Int64); eerr != nil {
		return fmt.Errorf("rollup: agregar: %w", eerr)
	}
	if _, eerr := tx.ExecContext(ctx,
		`INSERT INTO rollup_cursor (unico, ultimo_evento, actualizado) VALUES (1,?,?)
		 ON CONFLICT(unico) DO UPDATE SET ultimo_evento = excluded.ultimo_evento, actualizado = excluded.actualizado`,
		techo.Int64, r.s.opts.Reloj().UTC().Format(time.RFC3339Nano)); eerr != nil {
		return fmt.Errorf("rollup: cursor: %w", eerr)
	}
	if cerr := tx.Commit(); cerr != nil {
		return fmt.Errorf("rollup: commit: %w", cerr)
	}
	r.mu.Lock()
	r.colapsada = r.colapsada || colapsar
	r.mu.Unlock()
	return nil
}

// debeColapsar mira cuántas filas tiene ya el rollup. Superado el tope, `caja_id` colapsa a
// `(otros)` y la fila se marca — degradación DECLARADA. Perder la dimensión y no decirlo
// sería mostrar un desglose incompleto como si fuera completo.
func (r *Rollup) debeColapsar(ctx context.Context, tx *sql.Tx) (bool, error) {
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM rollup_hora`).Scan(&n); err != nil {
		return false, fmt.Errorf("rollup: cardinalidad: %w", err)
	}
	return n >= topeCardinalidad, nil
}

// Colapsada reporta si alguna pasada tuvo que colapsar la dimensión de caja.
func (r *Rollup) Colapsada() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.colapsada
}

// Recomputar rehace desde cero las horas indicadas (formato `2006-01-02T15`, UTC). Se dispara
// **solo** tras un borrado por arnés o un archivado: una hora cuyo detalle cambió tiene que
// re-agregarse o el total quedaría alimentándose de filas que ya no existen.
//
// Con `horas` vacío recomputa TODO — es la salida de emergencia, no el camino normal.
func (r *Rollup) Recomputar(ctx context.Context, horas []string) error {
	tx, err := r.s.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rollup: recomputar begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if len(horas) == 0 {
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM rollup_hora`); eerr != nil {
			return fmt.Errorf("rollup: recomputar limpiar: %w", eerr)
		}
		if _, eerr := tx.ExecContext(ctx, `DELETE FROM rollup_cursor`); eerr != nil {
			return fmt.Errorf("rollup: recomputar cursor: %w", eerr)
		}
	} else {
		for _, h := range horas {
			if _, eerr := tx.ExecContext(ctx, `DELETE FROM rollup_hora WHERE hora = ?`, h); eerr != nil {
				return fmt.Errorf("rollup: recomputar hora %s: %w", h, eerr)
			}
		}
	}
	if cerr := tx.Commit(); cerr != nil {
		return fmt.Errorf("rollup: recomputar commit: %w", cerr)
	}

	if len(horas) == 0 {
		return r.Actualizar(ctx)
	}
	// Re-agregar solo las horas afectadas, desde la tabla cruda.
	tx2, err := r.s.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("rollup: reagregar begin: %w", err)
	}
	defer func() { _ = tx2.Rollback() }()
	for _, h := range horas {
		if _, eerr := tx2.ExecContext(ctx, insertRollupPorHora(), h); eerr != nil {
			return fmt.Errorf("rollup: reagregar hora %s: %w", h, eerr)
		}
	}
	return tx2.Commit()
}

// insertRollupPorHora reagrega UNA hora entera desde la tabla cruda. No lleva `ON CONFLICT`
// porque la fila se borró justo antes: si hubiera conflicto, sería un bug del caller y es
// mejor que se note.
func insertRollupPorHora() string {
	return `
INSERT INTO rollup_hora (
  hora, arnes_id, instalacion_id, caja_id, runtime, modelo_canonico, emisor, atribucion,
  eventos, turnos, tok_entrada, tok_salida, tok_cache_lectura, tok_cache_5m, tok_cache_1h,
  tok_razonamiento, tok_cache_sin_tier, costo_reportado_micros, costo_calculado_micros,
  duracion_ms_suma, duracion_ms_cuenta, rechazados, cardinalidad_colapsada)
SELECT
  substr(ts_recibido,1,13), COALESCE(arnes_id,''), COALESCE(instalacion_id,''), COALESCE(caja_id,''),
  runtime, COALESCE(modelo_canonico, modelo, ''), emisor, atribucion,
  COUNT(*), COUNT(DISTINCT turno_id),
  SUM(tok_entrada), SUM(tok_salida), SUM(tok_cache_lectura), SUM(tok_cache_5m),
  SUM(tok_cache_1h), SUM(tok_razonamiento), SUM(tok_cache_sin_tier),
  SUM(costo_reportado_micros), SUM(costo_calculado_micros),
  SUM(duracion_ms), SUM(CASE WHEN duracion_ms IS NOT NULL THEN 1 ELSE 0 END),
  0, 0
FROM evento
WHERE substr(ts_recibido,1,13) = ? AND tipo_evento <> 'metrica'
GROUP BY 1,2,3,4,5,6,7,8`
}

// HorasAfectadas devuelve las horas (UTC, `2006-01-02T15`) que tienen eventos de un arnés.
// Se llama ANTES de borrarlo, para saber qué recomputar después.
func (s *Store) HorasAfectadas(ctx context.Context, arnesID string) ([]string, error) {
	rows, err := s.reader.QueryContext(ctx,
		`SELECT DISTINCT substr(ts_recibido,1,13) FROM evento WHERE arnes_id = ?`, arnesID)
	if err != nil {
		return nil, fmt.Errorf("store: horas afectadas: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var h string
		if serr := rows.Scan(&h); serr != nil {
			return nil, fmt.Errorf("store: horas afectadas scan: %w", serr)
		}
		out = append(out, h)
	}
	return out, rows.Err()
}
