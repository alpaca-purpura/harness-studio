package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// telemetria_retencion.go es el TTL y el botón «borrar la telemetría de este arnés».
//
// El número del TTL está **FIRMADO en 90 días** (D26.3, 2026-07-27). D15.3 había firmado «TTL
// por default» sin número y el 90 viajaba rotulado «propuesto»; el rótulo salió del wire y de
// la UI. **Firmarlo no lo clava:** sigue viajando por flag, y **la UI nunca lo hardcodea** —
// lo lee de `GET /api/telemetria/salud`, que es lo que arregló A-2.
//
// ⚠️ Bajar este número más adelante **borra datos que hoy existen**; subirlo no los recupera.

// RetencionDefaultDias es el TTL de la tabla cruda: 90 días, firmado.
const RetencionDefaultDias = 90

// RollupDefaultMeses es el TTL del agregado. El agregado sobrevive al detalle a propósito:
// tras purgar, el drill-down de un turno viejo dice «detalle purgado, resumen conservado» en
// vez de perder el resumen también.
const RollupDefaultMeses = 24

// purgador es lo que la retención necesita del almacén. Interfaz mínima en el consumidor: el
// caso de uso no pide el store entero para borrar filas.
type purgador interface {
	Purgar(ctx context.Context, p ports.PurgaTelemetria) (int64, error)
	PurgarRollup(ctx context.Context, antesDe time.Time) (int64, error)
	HorasAfectadas(ctx context.Context, arnesID string, desde, hasta time.Time) ([]string, error)
	Descartar(ctx context.Context, arnesID, puntoID string, ahora time.Time) error
	Recuperar(ctx context.Context, arnesID, puntoID string) error
}

// recomputador rehace el agregado de las horas que un borrado dejó huérfanas.
type recomputador interface {
	Recomputar(ctx context.Context, horas []string) error
}

// SetRetencion cablea el almacén concreto y el recomputador. Se llama desde el composition
// root; `rec` puede ser nil (sin agregado que rehacer).
func (s *TelemetriaService) SetRetencion(p purgador, rec recomputador, dias, meses int) {
	if dias <= 0 {
		dias = RetencionDefaultDias
	}
	if meses <= 0 {
		meses = RollupDefaultMeses
	}
	s.purga = p
	s.recomputa = rec
	s.retencionDias = dias
	s.rollupMeses = meses
}

// Purgar aplica el TTL: borra el detalle más viejo que la retención y el agregado más viejo
// que su propio TTL. Devuelve cuántas filas de detalle se fueron.
func (s *TelemetriaService) Purgar(ctx context.Context, p ports.PurgaTelemetria) (int64, error) {
	if s.purga == nil {
		return 0, fmt.Errorf("telemetria: retención sin almacén cableado")
	}
	if p.ArnesID != "" {
		return s.BorrarArnes(ctx, p.ArnesID, p.Desde, p.Hasta)
	}
	if p.AntesDe.IsZero() {
		dias := s.retencionDias
		if dias <= 0 {
			dias = RetencionDefaultDias
		}
		p.AntesDe = s.reloj().UTC().AddDate(0, 0, -dias)
	}
	n, err := s.purga.Purgar(ctx, p)
	if err != nil {
		return 0, err
	}
	meses := s.rollupMeses
	if meses <= 0 {
		meses = RollupDefaultMeses
	}
	if m, rerr := s.purga.PurgarRollup(ctx, s.reloj().UTC().AddDate(0, -meses, 0)); rerr != nil {
		slog.Warn("telemetria: TTL del agregado no aplicado", "err", rerr)
	} else if m > 0 {
		slog.Info("telemetria: agregado purgado por TTL", "filas", m, "meses", meses)
	}
	return n, nil
}

// BorrarArnes borra la telemetría de un arnés: el detalle **y el agregado**, en una
// transacción, y después rehace las horas afectadas.
//
// Borrar solo el detalle dejaría una cifra huérfana en el tablero alimentándose de filas que
// ya no existen — y esa cifra sobreviviría a la operación que el usuario pidió justamente
// para hacerla desaparecer.
//
// **`desde`/`hasta` acotan el borrado** (D26.5 · A-4). En cero, borra todo el historial del
// arnés — que sigue siendo un caso legítimo, ahora explícito. La razón de que exista la
// ventana: la confirmación de la UI declara el conteo de la ventana activa, y una acción
// irreversible cuyo alcance declarado no es su alcance real es la peor clase de mentira que
// esta superficie podía tener.
func (s *TelemetriaService) BorrarArnes(ctx context.Context, arnesID string, desde, hasta time.Time) (int64, error) {
	if s.purga == nil {
		return 0, fmt.Errorf("telemetria: retención sin almacén cableado")
	}
	if arnesID == "" {
		return 0, fmt.Errorf("telemetria: borrar sin arnés no es un borrado, es un wipe")
	}
	// Se anotan ANTES de borrar: después de borrar, las horas afectadas ya no se pueden
	// deducir de los datos.
	horas, err := s.purga.HorasAfectadas(ctx, arnesID, desde, hasta)
	if err != nil {
		return 0, err
	}
	n, err := s.purga.Purgar(ctx, ports.PurgaTelemetria{ArnesID: arnesID, Desde: desde, Hasta: hasta})
	if err != nil {
		return 0, err
	}
	if s.recomputa != nil && len(horas) > 0 {
		if rerr := s.recomputa.Recomputar(ctx, horas); rerr != nil {
			return n, fmt.Errorf("telemetria: agregado no recomputado tras borrar %s: %w", arnesID, rerr)
		}
	}
	return n, nil
}

// RetencionDias devuelve el TTL vigente, para que el CLI y el HTTP lo muestren **desde la
// config** y no lo hardcodeen.
func (s *TelemetriaService) RetencionDias() int {
	if s.retencionDias <= 0 {
		return RetencionDefaultDias
	}
	return s.retencionDias
}

// RollupMeses devuelve el TTL del agregado.
func (s *TelemetriaService) RollupMeses() int {
	if s.rollupMeses <= 0 {
		return RollupDefaultMeses
	}
	return s.rollupMeses
}

// DescartarPunto y RecuperarPunto son las decisiones del operador sobre un punto de mejora
// (D26.4). Están acá, con el borrado, porque son la misma clase de cosa: **lo que el operador
// decide sobre sus datos**, no lo que el sistema mide.
//
// Un descarte se guarda; no se recuerda en memoria. Un descarte que se pierde al reiniciar no
// es un descarte: el punto vuelve solo y el operador vuelve a descartarlo, para siempre.
func (s *TelemetriaService) DescartarPunto(ctx context.Context, arnesID, puntoID string) error {
	if arnesID == "" || puntoID == "" {
		return errors.New("telemetria: descartar exige arnés y punto")
	}
	if s.purga == nil {
		return errors.New("telemetria: descartes sin almacén cableado")
	}
	return s.purga.Descartar(ctx, arnesID, puntoID, s.reloj().UTC())
}

// RecuperarPunto deshace un descarte. Existe porque un descarte sin vuelta atrás convierte un
// clic distraído en la pérdida permanente de un hallazgo que costó dinero producir.
func (s *TelemetriaService) RecuperarPunto(ctx context.Context, arnesID, puntoID string) error {
	if arnesID == "" || puntoID == "" {
		return errors.New("telemetria: recuperar exige arnés y punto")
	}
	if s.purga == nil {
		return errors.New("telemetria: descartes sin almacén cableado")
	}
	return s.purga.Recuperar(ctx, arnesID, puntoID)
}
