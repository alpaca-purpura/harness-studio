package usecase

import (
	"context"
	"sort"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// telemetria_mejoras.go corre el motor de detección (arquitectura-modulo.md §9.1).
//
// El orden importa y es la razón de ser del archivo:
//
//  1. se arma la `Ventana` (turnos unidos, ya costeados);
//  2. se arma el `ContextoDeteccion` **del dato real de esa ventana**, jamás de una config;
//  3. por cada detector: **`Aplica(c)` PRIMERO**. Si no aplica, va a `no_aplican[]` con su
//     motivo. Si aplica, `Evaluar(v)` y sus puntos van a `puntos[]`;
//  4. los detectores fuera del MVP se emiten en `no_medidos[]` con «no medido todavía».
//
// **Las tres listas viajan SIEMPRE** (A16). Omitir `no_aplican` obligaría al FE a elegir
// entre no mostrar nada (gap escondido) o mostrar 0 (mentira).

// detectoresNoMedidos son los otros siete de la familia B. No pasan la regla A4 —les falta el
// fix concreto o la cotización— así que **no se implementan y se declaran**. Un detector
// vacío que devuelve «sin hallazgos» sería peor que decir «todavía no lo medimos».
var detectoresNoMedidos = []struct {
	ID     domain.DetectorID
	Nombre string
}{
	{"b5-cache-escrito-y-nunca-leido", "cache escrito y nunca leído"},
	{"b7-costo-de-compact", "costo de /compact"},
	{"b8-overhead-de-subagentes", "overhead de subagentes"},
	{"b9-turnos-sin-avance", "turnos sin avance"},
	{"b10-modelo-caro-para-tarea-simple", "modelo caro para tarea simple"},
	{"b11-resultados-de-herramienta-obesos", "resultados de herramienta obesos"},
	{"b12-ttl-largo-desaprovechado", "TTL largo desaprovechado"},
	{"b13-reintentos-en-cascada", "reintentos en cascada"},
}

// Mejoras devuelve los puntos de mejora de una ventana, más los detectores que no aplican
// (con motivo) y los que todavía no medimos.
func (s *TelemetriaService) Mejoras(ctx context.Context, q ports.ConsultaTelemetria) (domain.RespuestaMejoras, error) {
	v := s.ventana(q)
	out := domain.RespuestaMejoras{
		Puntos:    []domain.PuntoDeMejora{},
		NoAplican: []domain.EstadoDetector{},
		NoMedidos: []domain.EstadoDetector{},
	}
	out.Ventana.Desde = v.Desde
	out.Ventana.Hasta = v.Hasta

	turnos, err := s.store.Turnos(ctx, v)
	if err != nil {
		return out, err
	}
	resumen, err := s.store.Resumen(ctx, v)
	if err != nil {
		return out, err
	}
	out.Escenario = resumen.Escenario

	ventana := domain.Ventana{
		Desde: v.Desde, Hasta: v.Hasta, ArnesID: v.ArnesID, Turnos: turnos,
		Contexto: s.contextoDe(turnos, resumen),
	}
	if resumen.CostoReportadoMicros != nil {
		ventana.TotalMicros = *resumen.CostoReportadoMicros
	}

	for _, d := range s.detectores {
		// **Aplica() SIEMPRE antes que Evaluar().** Un detector que corre sin poder correr
		// devuelve 0, y un 0 se lee como «no hay problema».
		ap := d.Aplica(ventana.Contexto)
		if !ap.Aplica {
			out.NoAplican = append(out.NoAplican, domain.EstadoDetector{
				Detector: d.ID(), Nombre: d.Nombre(), Aplica: false, Motivo: ap.Motivo,
			})
			continue
		}
		puntos := d.Evaluar(ventana)
		if ap.Parcial {
			// Corre, pero no ve todo. Viaja como matiz EXPLÍCITO con su motivo, nunca como
			// un visto bueno liso.
			out.NoAplican = append(out.NoAplican, domain.EstadoDetector{
				Detector: d.ID(), Nombre: d.Nombre(), Aplica: true,
				CoberturaParcial: true, Motivo: ap.Motivo, Hallazgos: len(puntos),
			})
		}
		out.Puntos = append(out.Puntos, puntos...)
	}

	for _, nm := range detectoresNoMedidos {
		out.NoMedidos = append(out.NoMedidos, domain.EstadoDetector{
			Detector: nm.ID, Nombre: nm.Nombre, Aplica: false,
			Motivo: "no medido todavía",
		})
	}

	// Orden estable por impacto: lo más caro primero. Empates por id, para que dos corridas
	// del mismo dato den la misma lista.
	sort.SliceStable(out.Puntos, func(i, j int) bool {
		if out.Puntos[i].DiferenciaMicros != out.Puntos[j].DiferenciaMicros {
			return out.Puntos[i].DiferenciaMicros > out.Puntos[j].DiferenciaMicros
		}
		return out.Puntos[i].Detector < out.Puntos[j].Detector
	})
	return out, nil
}

// contextoDe arma el contexto de detección **del dato real** de la ventana. Ningún campo sale
// de una configuración ni del escenario declarado: si saliera, un arnés podría decir que está
// mejor medido de lo que está.
func (s *TelemetriaService) contextoDe(turnos []domain.TurnoUnido, r domain.ResumenTelemetria) domain.ContextoDeteccion {
	c := domain.ContextoDeteccion{
		Runtime:            s.perfil.Runtime,
		Escenario:          r.Escenario,
		CatalogoDisponible: s.catalogo != nil && s.catalogo.Version().Modelos > 0,
	}
	modelos := map[string]bool{}
	for _, t := range turnos {
		if t.TieneDinero {
			c.TieneCosto = true
		}
		if t.TieneProceso {
			c.TieneSenalProceso = true
		}
		// TieneSplitTTL es un hecho de la VENTANA, no del escenario: el día que el split
		// llegue por otro canal, B1 se enciende SOLO, sin tocar código.
		if t.Tokens.CacheEscritura1h != nil && t.Tokens.CacheEscritura5m != nil {
			c.TieneSplitTTL = true
		}
		if len(t.Gates) > 0 {
			c.TieneGateHumano = true
		}
		if t.Rotaciones > 0 {
			c.TieneEventoRotacion = true
		}
		if t.Modelo != "" {
			modelos[t.Modelo] = true
		}
	}
	c.ModelosDistintos = len(modelos)
	return c
}
