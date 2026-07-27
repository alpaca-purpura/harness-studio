package usecase

import (
	"context"
	"fmt"
	"log/slog"
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

	// 🔴 Defecto C4: `Turnos` devuelve una PÁGINA. Si hay más turnos que el límite, los
	// detectores tendrían el numerador recortado y el denominador completo — y una caja se
	// llevaría el 83 % cuando se lleva el 100 %. Antes que producir un porcentaje falso, se
	// **declara que no se puede evaluar** y se dice por qué.
	ag, aerr := s.agregado(ctx, v)
	if aerr == nil && ag.Turnos > len(turnos) {
		motivo := fmt.Sprintf("la ventana tiene %d turnos y la consulta devuelve %d: "+
			"evaluar sobre la página daría porcentajes calculados con numerador recortado y "+
			"denominador completo. Acotá la ventana con --desde/--hasta", ag.Turnos, len(turnos))
		for _, d := range s.detectores {
			out.NoAplican = append(out.NoAplican, domain.EstadoDetector{
				Detector: d.ID(), Nombre: d.Nombre(), Aplica: false, Motivo: motivo,
			})
		}
		for _, nm := range detectoresNoMedidos {
			out.NoMedidos = append(out.NoMedidos, domain.EstadoDetector{
				Detector: nm.ID, Nombre: nm.Nombre, Aplica: false, Motivo: "no medido todavía",
			})
		}
		return out, nil
	}

	ventana := domain.Ventana{
		Desde: v.Desde, Hasta: v.Hasta, ArnesID: v.ArnesID, Turnos: turnos,
		Contexto: s.contextoDe(turnos, resumen),
		Precios:  s.preciosDe(turnos),
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
		// Regla A4 en la puerta de salida (D25): **un punto sin contrafactual no es tarjeta.**
		// Pero tampoco desaparece: el detector que encontró algo y no puede proponer un
		// arreglo se declara `sin_fix` con su motivo. Filtrarlo en silencio dejaría el mismo
		// hueco que este módulo existe para no dejar — «no lo mostramos» leído como «no hay».
		var conFix []domain.PuntoDeMejora
		var sinFix int
		for _, p := range puntos {
			if p.Contrafactual == "" {
				sinFix++
				continue
			}
			conFix = append(conFix, p)
		}
		if sinFix > 0 {
			out.NoAplican = append(out.NoAplican, domain.EstadoDetector{
				Detector: d.ID(), Nombre: d.Nombre(), Aplica: true,
				SinFix: true, Hallazgos: sinFix, Motivo: domain.MotivoSinContrafactual,
			})
		}
		out.Puntos = append(out.Puntos, conFix...)
	}

	for _, nm := range detectoresNoMedidos {
		out.NoMedidos = append(out.NoMedidos, domain.EstadoDetector{
			Detector: nm.ID, Nombre: nm.Nombre, Aplica: false,
			Motivo: "no medido todavía",
		})
	}

	// D26.4 — los puntos que el operador ya descartó no vuelven a la lista. Se filtran ACÁ y
	// no en el FE: filtrar en la pantalla dejaría el contador y el «✓ sin fugas» del
	// Portafolio contando cosas que el operador ya dijo que no quiere ver.
	if desc, derr := s.store.Descartados(ctx, v.ArnesID); derr == nil && len(desc) > 0 {
		vivos := out.Puntos[:0]
		for _, p := range out.Puntos {
			if !desc[p.ID] {
				vivos = append(vivos, p)
			}
		}
		out.Descartados = len(out.Puntos) - len(vivos)
		out.Puntos = vivos
	} else if derr != nil {
		// Si no se pueden leer los descartes, se muestran TODOS los puntos. Esconder por un
		// error de lectura sería esconder hallazgos por una falla de infraestructura.
		slog.Warn("telemetria: descartes no legibles — se muestran todos los puntos", "err", derr)
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

// preciosDe arma el catálogo aplicable a los modelos de ESTA ventana (D26.1).
//
// Existe porque un detector **no conoce puertos**: no puede consultar el catálogo. Antes de
// esto, B1 y B3 se las arreglaban sumando conteos de tokens en campos `micros` — una cifra que
// no era dinero y se mostraba como si lo fuera.
//
// Un modelo que el catálogo no conoce **no entra al mapa**, y eso es deliberado: un
// `PrecioModelo` en cero costearía todo gratis en silencio, que es exactamente lo que
// `ports.CatalogoPrecios` evita al devolver `ok=false`. El detector ve la ausencia y la
// declara en su sesgo.
//
// La clave es el nombre CANÓNICO y también el crudo cuando difieren: el turno guarda el nombre
// tal como lo dijo el runtime, y `PrecioDe` busca por ese.
func (s *TelemetriaService) preciosDe(turnos []domain.TurnoUnido) map[string]domain.PrecioModelo {
	if s.catalogo == nil {
		return nil
	}
	out := map[string]domain.PrecioModelo{}
	for _, t := range turnos {
		if t.Modelo == "" {
			continue
		}
		if _, ya := out[t.Modelo]; ya {
			continue
		}
		canonico := s.catalogo.Canonizar(t.Modelo)
		precio, ok := s.catalogo.Precio(canonico)
		if !ok {
			continue // desconocido ≠ gratis: se omite y el detector lo declara.
		}
		out[t.Modelo] = precio
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
