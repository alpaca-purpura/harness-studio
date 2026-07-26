package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// consultas.go son las lecturas del almacén. La regla que las gobierna a todas:
// **`NULL` sobrevive**. Un `SUM()` sobre un grupo donde nadie tenía el bucket devuelve NULL,
// y ese NULL viaja hasta el wire — no se `COALESCE`a a 0 por comodidad del scanner.
//
// `A15`: los eventos con `atribucion='sin-dato'` **se guardan y NO se suman al total**.
// Aparecen en `Cobertura.SinDato` y en el drill-down. Sumarlos sería atribuir por adivinanza.

// filtro arma el WHERE común a las consultas.
func filtro(q ports.ConsultaTelemetria) (string, []any) {
	var cond []string
	var args []any
	if q.ArnesID != "" {
		cond = append(cond, "arnes_id = ?")
		args = append(args, q.ArnesID)
	}
	if q.InstalacionID != "" {
		cond = append(cond, "instalacion_id = ?")
		args = append(args, q.InstalacionID)
	}
	if q.CajaID != "" {
		cond = append(cond, "caja_id = ?")
		args = append(args, q.CajaID)
	}
	if q.SesionID != "" {
		cond = append(cond, "sesion_id = ?")
		args = append(args, q.SesionID)
	}
	if !q.Desde.IsZero() {
		cond = append(cond, "ts_recibido >= ?")
		args = append(args, q.Desde.UTC().Format(time.RFC3339Nano))
	}
	if !q.Hasta.IsZero() {
		cond = append(cond, "ts_recibido <= ?")
		args = append(args, q.Hasta.UTC().Format(time.RFC3339Nano))
	}
	if len(cond) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(cond, " AND "), args
}

// Resumen agrega el gasto de la ventana.
//
// **Solo suma lo atribuido** (A15): las filas `sin-dato` cuentan en `Cobertura.SinDato` y no
// entran a ningún total. Los costos viajan como punteros: `null` cuando ningún turno los
// trajo, jamás 0 — un 0 en dinero se lee como «salió gratis».
func (s *Store) Resumen(ctx context.Context, q ports.ConsultaTelemetria) (domain.ResumenTelemetria, error) {
	out := domain.ResumenTelemetria{
		Desde: q.Desde.UTC(), Hasta: q.Hasta.UTC(),
		// `Estimado` es SIEMPRE true mientras la fuente sea `cost_usd_micros`: el runtime lo
		// documenta como "Estimated cost", no como facturación (ANEXO H4).
		Estimado:  true,
		Confianza: domain.ConfianzaSinDato,
	}
	where, args := filtro(q)
	// Los agregados de dinero excluyen `sin-dato`; los conteos de sesiones/turnos NO,
	// porque un turno no atribuido igual ocurrió.
	sinDato := " atribucion <> 'sin-dato'"
	whereAtrib := where
	if whereAtrib == "" {
		whereAtrib = " WHERE" + sinDato
	} else {
		whereAtrib += " AND" + sinDato
	}

	row := s.reader.QueryRowContext(ctx,
		`SELECT SUM(costo_reportado_micros), SUM(costo_calculado_micros),
		        COUNT(DISTINCT corrida_id), COUNT(DISTINCT sesion_id), COUNT(DISTINCT turno_id)
		   FROM evento`+whereAtrib, args...)
	var rep, calc sql.NullInt64
	if err := row.Scan(&rep, &calc, &out.Corridas, &out.Sesiones, &out.Turnos); err != nil {
		return out, fmt.Errorf("store: resumen: %w", err)
	}
	if rep.Valid {
		v := rep.Int64
		out.CostoReportadoMicros = &v
	}
	if calc.Valid {
		v := calc.Int64
		out.CostoCalculadoMicros = &v
	}

	cob, esc, conf, err := s.cobertura(ctx, q)
	if err != nil {
		return out, err
	}
	out.Cobertura = cob
	out.Escenario = esc
	out.Confianza = conf
	return out, nil
}

// cobertura cuenta los turnos por confianza y los concilia contra los esperados (A9).
func (s *Store) cobertura(ctx context.Context, q ports.ConsultaTelemetria) (domain.Cobertura, domain.Escenario, domain.Confianza, error) {
	var cob domain.Cobertura
	where, args := filtro(q)
	rows, err := s.reader.QueryContext(ctx,
		`SELECT atribucion, COUNT(DISTINCT COALESCE(turno_id, 'sin-turno:' || id)) FROM evento`+where+
			` GROUP BY atribucion`, args...)
	if err != nil {
		return cob, "", domain.ConfianzaSinDato, fmt.Errorf("store: cobertura: %w", err)
	}
	defer func() { _ = rows.Close() }()
	peor := domain.Confianza("")
	for rows.Next() {
		var a string
		var n int
		if serr := rows.Scan(&a, &n); serr != nil {
			return cob, "", domain.ConfianzaSinDato, fmt.Errorf("store: cobertura scan: %w", serr)
		}
		switch domain.Confianza(a) {
		case domain.ConfianzaExacta:
			cob.Exacta = n
		case domain.ConfianzaPorHash:
			cob.PorHash = n
		case domain.ConfianzaPorProceso:
			cob.PorProceso = n
		default:
			cob.SinDato = n
		}
		if peor == "" {
			peor = domain.Confianza(a)
		} else {
			peor = domain.PeorConfianza(peor, domain.Confianza(a))
		}
	}
	if rerr := rows.Err(); rerr != nil {
		return cob, "", domain.ConfianzaSinDato, fmt.Errorf("store: cobertura rows: %w", rerr)
	}
	if peor == "" {
		peor = domain.ConfianzaSinDato
	}

	// ── el denominador (A9) ──
	// Solo existe cuando el daemon fue el proceso padre: en S2 no hay forma independiente de
	// saber cuántos turnos hubo, así que `Esperados` viaja **nil** y la UI dice «cobertura
	// desconocida fuera de ArnesIA». Un 0 ahí sería una mentira aritmética.
	esc, err := s.escenario(ctx, q)
	if err != nil {
		return cob, "", peor, err
	}
	if esc == domain.EscenarioS1 {
		var esperados, noLlegaron int
		w, a := filtroTurnoEsperado(q)
		if err := s.reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM turno_esperado`+w, a...).Scan(&esperados); err != nil {
			return cob, esc, peor, fmt.Errorf("store: esperados: %w", err)
		}
		// Solo cuentan como «no llegaron» los que ya tuvieron tiempo de llegar: un turno de
		// hace 10 segundos que todavía no reportó no es un agujero, es un turno en vuelo.
		corte := s.opts.Reloj().UTC().Add(-5 * time.Minute).Format(time.RFC3339Nano)
		w2 := w
		if w2 == "" {
			w2 = " WHERE medido = 0 AND ts < ?"
		} else {
			w2 += " AND medido = 0 AND ts < ?"
		}
		if err := s.reader.QueryRowContext(ctx, `SELECT COUNT(*) FROM turno_esperado`+w2,
			append(append([]any{}, a...), corte)...).Scan(&noLlegaron); err != nil {
			return cob, esc, peor, fmt.Errorf("store: no llegaron: %w", err)
		}
		cob.Esperados = &esperados
		cob.NoLlegaron = &noLlegaron
	}
	return cob, esc, peor, nil
}

func filtroTurnoEsperado(q ports.ConsultaTelemetria) (string, []any) {
	var cond []string
	var args []any
	if q.ArnesID != "" {
		cond = append(cond, "arnes_id = ?")
		args = append(args, q.ArnesID)
	}
	if q.SesionID != "" {
		cond = append(cond, "sesion_id = ?")
		args = append(args, q.SesionID)
	}
	if !q.Desde.IsZero() {
		cond = append(cond, "ts >= ?")
		args = append(args, q.Desde.UTC().Format(time.RFC3339Nano))
	}
	if !q.Hasta.IsZero() {
		cond = append(cond, "ts <= ?")
		args = append(args, q.Hasta.UTC().Format(time.RFC3339Nano))
	}
	if len(cond) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(cond, " AND "), args
}

// escenario DERIVA el nivel de instrumentación de la señal que llegó (A19 · §7.0.2):
//
//	hay corrida nuestra o evento de stream-json ⇒ s1
//	hay api_request sin corrida nuestra         ⇒ s2-instrumentado
//	solo hay eventos de hook                    ⇒ s2-degradado
//
// **Un arnés no puede declarar su propio escenario**: se calcula de lo que hay.
func (s *Store) escenario(ctx context.Context, q ports.ConsultaTelemetria) (domain.Escenario, error) {
	where, args := filtro(q)
	var s1, dinero, total int
	err := s.reader.QueryRowContext(ctx,
		`SELECT
		   SUM(CASE WHEN corrida_id IS NOT NULL OR emisor = 'streamjson' THEN 1 ELSE 0 END),
		   SUM(CASE WHEN tipo_evento = 'api_request' THEN 1 ELSE 0 END),
		   COUNT(*)
		 FROM evento`+where, args...).Scan(&sqlNullInt{&s1}, &sqlNullInt{&dinero}, &total)
	if err != nil {
		return "", fmt.Errorf("store: escenario: %w", err)
	}
	switch {
	case total == 0:
		return "", nil // sin señal no hay escenario que declarar; tampoco se inventa uno.
	case s1 > 0:
		return domain.EscenarioS1, nil
	case dinero > 0:
		return domain.EscenarioS2Instrumentado, nil
	default:
		return domain.EscenarioS2Degradado, nil
	}
}

// sqlNullInt escanea un `SUM()` que puede volver NULL a un int, sin fingir que el NULL era 0
// en el modelo: acá el 0 es el neutro correcto de un conteo, no de una medición.
type sqlNullInt struct{ dst *int }

func (s sqlNullInt) Scan(v any) error {
	switch t := v.(type) {
	case nil:
		*s.dst = 0
	case int64:
		*s.dst = int(t)
	case float64:
		*s.dst = int(t)
	default:
		return fmt.Errorf("store: valor entero inesperado %T", v)
	}
	return nil
}

// PorCaja devuelve el gasto por caja. **Incluye las cajas SIN dato atribuible**, con
// `Atribuible:false` + motivo y `CostoMicros: null` — omitirlas obligaría al FE a inventar
// por qué faltan, y un `0` diría que corrieron gratis.
func (s *Store) PorCaja(ctx context.Context, q ports.ConsultaTelemetria) ([]domain.GastoCaja, error) {
	where, args := filtro(q)
	rows, err := s.reader.QueryContext(ctx,
		`SELECT COALESCE(caja_id,''), SUM(costo_reportado_micros), SUM(costo_calculado_micros),
		        COUNT(DISTINCT corrida_id), atribucion
		   FROM evento`+where+`
		  GROUP BY COALESCE(caja_id,''), atribucion
		  ORDER BY 1`, args...)
	if err != nil {
		return nil, fmt.Errorf("store: por caja: %w", err)
	}
	defer func() { _ = rows.Close() }()

	acc := map[string]*domain.GastoCaja{}
	var orden []string
	var total int64
	for rows.Next() {
		var caja, atrib string
		var rep, calc sql.NullInt64
		var corridas int
		if serr := rows.Scan(&caja, &rep, &calc, &corridas, &atrib); serr != nil {
			return nil, fmt.Errorf("store: por caja scan: %w", serr)
		}
		g, ok := acc[caja]
		if !ok {
			g = &domain.GastoCaja{CajaID: caja, Nombre: caja, Confianza: domain.Confianza(atrib)}
			acc[caja] = g
			orden = append(orden, caja)
		} else {
			g.Confianza = domain.PeorConfianza(g.Confianza, domain.Confianza(atrib))
		}
		g.Corridas += corridas
		if domain.Confianza(atrib) == domain.ConfianzaSinDato {
			g.SinAtribucion = true
			continue // A15: no suma al total. Se cuenta aparte, no se descarta.
		}
		if rep.Valid {
			// `Atribuible` significa «hay una cifra que mostrar acá», no «hubo actividad».
			// Una caja con actividad registrada y SIN dinero es un estado honesto distinto
			// —el normal en modo degradado— y merece su propio motivo.
			g.Atribuible = true
			v := rep.Int64
			if g.CostoMicros == nil {
				g.CostoMicros = &v
			} else {
				*g.CostoMicros += v
			}
			total += v
		}
	}
	if rerr := rows.Err(); rerr != nil {
		return nil, fmt.Errorf("store: por caja rows: %w", rerr)
	}

	out := make([]domain.GastoCaja, 0, len(orden))
	for _, k := range orden {
		g := acc[k]
		if !g.Atribuible {
			// El motivo es OBLIGATORIO cuando no hay cifra: un «sin dato» sin razón es un
			// gap escondido. Y los tres casos son DISTINTOS — decirlos igual sería tapar
			// tres cosas bajo una.
			switch {
			case g.CajaID == "":
				g.Motivo = "gasto sin caja: la corrida no declaró a qué caja pertenece"
			case g.SinAtribucion:
				g.Motivo = "sin dato atribuible: ningún evento de esta ventana pudo asignarse a esta caja"
			default:
				g.Motivo = "esta caja tuvo actividad pero ningún evento con costo: " +
					"el arnés no reporta dinero en esta ventana"
			}
		}
		if total > 0 && g.CostoMicros != nil {
			p := float64(*g.CostoMicros) / float64(total)
			g.Parte = &p
		}
		out = append(out, *g)
	}
	return out, nil
}

// Turnos hace el JOIN dinero×proceso por `(sesion_id, turno_id)` — igualdad de dos campos,
// **sin heurística de tiempo ni de orden** (ANEXO H1). Es el drill-down: toca la tabla cruda,
// no el rollup.
func (s *Store) Turnos(ctx context.Context, q ports.ConsultaTelemetria) ([]domain.TurnoUnido, error) {
	where, args := filtro(q)
	limite := q.Limite
	if limite <= 0 {
		limite = 500 // el default del adaptador; jamás «sin límite».
	}
	rows, err := s.reader.QueryContext(ctx,
		`SELECT sesion_id, COALESCE(turno_id,''), COALESCE(arnes_id,''), COALESCE(instalacion_id,''),
		        COALESCE(caja_id,''), COALESCE(corrida_id,''), ts_recibido, COALESCE(modelo,''),
		        tok_entrada, tok_salida, tok_cache_lectura, tok_cache_5m, tok_cache_1h, tok_razonamiento,
		        costo_reportado_micros, costo_calculado_micros, duracion_ms,
		        atribucion, escenario, tipo_evento, COALESCE(resultado,''), COALESCE(gate,''),
		        COALESCE(herramienta,''), COALESCE(aritmetica,'')
		   FROM evento`+where+` ORDER BY ts_recibido`, args...)
	if err != nil {
		return nil, fmt.Errorf("store: turnos: %w", err)
	}
	defer func() { _ = rows.Close() }()

	acc := map[string]*domain.TurnoUnido{}
	var orden []string
	for rows.Next() {
		var (
			sesion, turno, arnes, inst, caja, corrida, ts, modelo string
			atrib, esc, tipo, res, gate, herr, arit               string
			e, sal, cr, c5, c1h, raz                              sql.NullInt64
			rep, calc, dur                                        sql.NullInt64
		)
		if serr := rows.Scan(&sesion, &turno, &arnes, &inst, &caja, &corrida, &ts, &modelo,
			&e, &sal, &cr, &c5, &c1h, &raz, &rep, &calc, &dur,
			&atrib, &esc, &tipo, &res, &gate, &herr, &arit); serr != nil {
			return nil, fmt.Errorf("store: turnos scan: %w", serr)
		}
		clave := sesion + "\x00" + turno
		t, ok := acc[clave]
		if !ok {
			cuando, _ := time.Parse(time.RFC3339Nano, ts)
			t = &domain.TurnoUnido{
				SesionID: sesion, TurnoID: turno, ArnesID: arnes, InstalacionID: inst,
				CajaID: caja, CorridaID: corrida, TS: cuando.UTC(),
				Atribucion: domain.Confianza(atrib), Escenario: domain.Escenario(esc),
				Aritmetica: domain.Aritmetica(arit),
			}
			acc[clave] = t
			orden = append(orden, clave)
		} else {
			t.Atribucion = domain.PeorConfianza(t.Atribucion, domain.Confianza(atrib))
			if t.CajaID == "" {
				t.CajaID = caja
			}
			if t.ArnesID == "" {
				t.ArnesID = arnes
			}
		}
		if modelo != "" {
			t.Modelo = modelo
		}
		acumularToken(&t.Tokens.Entrada, e)
		acumularToken(&t.Tokens.Salida, sal)
		acumularToken(&t.Tokens.CacheLectura, cr)
		acumularToken(&t.Tokens.CacheEscritura5m, c5)
		acumularToken(&t.Tokens.CacheEscritura1h, c1h)
		acumularToken(&t.Tokens.Razonamiento, raz)
		acumularToken(&t.CostoReportadoMicros, rep)
		acumularToken(&t.CostoCalculadoMicros, calc)
		acumularToken(&t.DuracionMs, dur)

		switch domain.TipoEvento(tipo) {
		case domain.EventoAPIRequest:
			t.TieneDinero = true
		case domain.EventoRotacion:
			t.Rotaciones++
			t.TieneProceso = true
		case domain.EventoGate:
			if gate != "" {
				t.Gates = append(t.Gates, gate)
			}
			t.TieneProceso = true
		default:
			t.TieneProceso = true
		}
		if res != "" {
			t.Resultados = append(t.Resultados, domain.Resultado(res))
		}
		if herr != "" {
			if t.Herramientas == nil {
				t.Herramientas = map[string]int{}
			}
			t.Herramientas[herr]++
		}
	}
	if rerr := rows.Err(); rerr != nil {
		return nil, fmt.Errorf("store: turnos rows: %w", rerr)
	}

	out := make([]domain.TurnoUnido, 0, len(orden))
	for i, k := range orden {
		if i >= limite {
			break
		}
		out = append(out, *acc[k])
	}
	return out, nil
}

// acumularToken suma un valor SQL que puede ser NULL a un puntero.
//
// **Si los dos lados son nil, sigue nil.** Es la mitad de la regla que hace que el «no
// aplica» sobreviva a la agregación: un `COALESCE(...,0)` acá convertiría cada ausencia en un
// cero medido, y ya no habría forma de distinguirlos aguas abajo.
func acumularToken(dst **int64, v sql.NullInt64) {
	if !v.Valid {
		return
	}
	if *dst == nil {
		n := v.Int64
		*dst = &n
		return
	}
	**dst += v.Int64
}
