package usecase

import (
	"context"
	"encoding/json"
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

// mapEventType es el evento SSE del Mapa (espeja sse.EventMap; literal por la misma razón
// que dockEventType — usecase sin imports de transporte).
const mapEventType = "map"

// mapFrame es el aviso «este arnés se reindexó» que viaja por `event: map` (RF-186). El FE
// que esté mirando ese arnés refetchea su grafo al recibirlo.
type mapFrame struct {
	HarnessID string `json:"harness_id"`
	Degradado bool   `json:"degradado,omitempty"`
}

// NewTurnReindexer arma el Reindexer real sobre el índice y el loader (inyectado como func
// para que usecase no importe adapters — misma razón que RoleSource). Tres salidas, todas
// visibles en el Mapa:
//   - carga OK con sello → el grafo nuevo entra tal cual;
//   - carga OK sin sello → el grafo fino que el loader SÍ reconoció sobrevive, Degradado,
//     bajo el ID DEL REGISTRO (la llave que la sesión y el Mapa ya usan — sintetizar una
//     llave nueva acá duplicaría la entrada del índice y dejaría stale la vigente);
//   - carga rota (el chat rompió el árbol) → grafo vacío Degradado bajo el id del registro:
//     el Mapa muestra el estado real, nunca la foto vieja como si nada.
//
// Tras cada Upsert (sano o degradado) publica `event: map` con el harness_id (RF-186) para
// que el FE refetchee en vivo. pub nil = sin aviso (tests/composición parcial).
func NewTurnReindexer(idx ports.IndexPort, load func(dir string) (domain.Graph, error), pub EventPublisher) Reindexer {
	return func(ctx context.Context, arnesID, cwd string) {
		g, err := load(cwd)
		if err != nil {
			slog.Warn("reindex tras turno: arnés no cargable — se indexa degradado", "arnes", arnesID, "cwd", cwd, "err", err)
			g = domain.Graph{Nodes: []domain.Box{}, Degradado: true}
		}
		if g.Arnes == nil {
			g.Arnes = &domain.Arnes{ID: arnesID, Nombre: arnesID}
			g.Degradado = true
		}
		if uerr := idx.Upsert(ctx, g); uerr != nil {
			slog.Warn("reindex tras turno: upsert falló", "arnes", arnesID, "err", uerr)
			return
		}
		if pub != nil {
			if b, merr := json.Marshal(mapFrame{HarnessID: g.Arnes.ID, Degradado: g.Degradado}); merr == nil {
				pub.Publish(mapEventType, b)
			}
		}
	}
}
