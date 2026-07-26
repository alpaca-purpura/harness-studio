package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// telemetria_retencion.go es el TTL y el botón «borrar la telemetría de este arnés».
//
// ⚠️ **El número del TTL NO está firmado** (J-6 · parada P2 del plan): D15.3 firmó «TTL por
// default» **sin número**. El 90 de acá es un valor **PROPUESTO**: viaja por flag, la config
// lo lleva rotulado y `arnesia telemetria salud` lo muestra como propuesto. **La UI nunca lo
// hardcodea.** El número lo pone el operador.

// RetencionDefaultDias es el TTL propuesto de la tabla cruda. Ver el aviso de arriba: es
// PROPUESTO, no firmado.
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
	HorasAfectadas(ctx context.Context, arnesID string) ([]string, error)
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
		return s.BorrarArnes(ctx, p.ArnesID)
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

// BorrarArnes borra TODO lo de un arnés: el detalle **y el agregado**, en una transacción, y
// después rehace las horas afectadas.
//
// Borrar solo el detalle dejaría una cifra huérfana en el tablero alimentándose de filas que
// ya no existen — y esa cifra sobreviviría a la operación que el usuario pidió justamente
// para hacerla desaparecer.
func (s *TelemetriaService) BorrarArnes(ctx context.Context, arnesID string) (int64, error) {
	if s.purga == nil {
		return 0, fmt.Errorf("telemetria: retención sin almacén cableado")
	}
	if arnesID == "" {
		return 0, fmt.Errorf("telemetria: borrar sin arnés no es un borrado, es un wipe")
	}
	// Se anotan ANTES de borrar: después de borrar, las horas afectadas ya no se pueden
	// deducir de los datos.
	horas, err := s.purga.HorasAfectadas(ctx, arnesID)
	if err != nil {
		return 0, err
	}
	n, err := s.purga.Purgar(ctx, ports.PurgaTelemetria{ArnesID: arnesID})
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
