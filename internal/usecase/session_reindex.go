package usecase

import (
	"context"
	"log/slog"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Reindexer re-carga el árbol del arnés de una sesión y actualiza el índice del Mapa —
// se dispara DESPUÉS de cada turno del conductor (RF-184, paquete mejorar-arnes-conversando;
// vive en el usecase y NO en el conductor: boundary adaptadores-de-agente-intercambiables).
// No devuelve error: un reindex fallido degrada honesto y jamás rompe el turno (RF-185).
// nil = sin reindex (tests / composición parcial).
type Reindexer func(ctx context.Context, arnesID, cwd string)

// NewTurnReindexer arma el Reindexer real sobre el índice y el loader (inyectado como func
// para que usecase no importe adapters — misma razón que RoleSource). Tres salidas, todas
// visibles en el Mapa:
//   - carga OK con sello → el grafo nuevo entra tal cual;
//   - carga OK sin sello → misma síntesis que ObservarEnMapa (id = HuellaPath del cwd,
//     Degradado) — el grafo fino que el loader SÍ reconoció sobrevive;
//   - carga rota (el chat rompió el árbol) → grafo vacío Degradado bajo el id del registro:
//     el Mapa muestra el estado real, nunca la foto vieja como si nada.
func NewTurnReindexer(idx ports.IndexPort, load func(dir string) (domain.Graph, error)) Reindexer {
	return func(ctx context.Context, arnesID, cwd string) {
		g, err := load(cwd)
		if err != nil {
			slog.Warn("reindex tras turno: arnés no cargable — se indexa degradado", "arnes", arnesID, "cwd", cwd, "err", err)
			g = domain.Graph{Nodes: []domain.Box{}, Degradado: true}
			g.Arnes = &domain.Arnes{ID: arnesID, Nombre: arnesID}
		}
		if g.Arnes == nil {
			g.Arnes = &domain.Arnes{ID: domain.HuellaPath(cwd), Nombre: arnesID}
			g.Degradado = true
		}
		if uerr := idx.Upsert(ctx, g); uerr != nil {
			slog.Warn("reindex tras turno: upsert falló", "arnes", arnesID, "err", uerr)
		}
	}
}
