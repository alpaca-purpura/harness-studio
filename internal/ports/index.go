package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// IndexPort is the disposable index over the arnés-tree source of truth (the harness
// directories the Portafolio knows about, ports.PortafolioStore/PortafolioScanner +
// ports.ArnesLoader — NOT the ~/.claude conversation JSONL, a separate corpus read by
// internal/adapters/history): it can always be rebuilt from scratch (see
// docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md v1.2).
type IndexPort interface {
	// Rebuild reconstructs the whole index from the known arnés tree (Portafolio).
	Rebuild(ctx context.Context) error
	// Query returns the agnostic graph of one harness.
	Query(ctx context.Context, harnessID string) (domain.Graph, error)
	// List returns every harness graph currently indexed (portfolio, S1).
	List(ctx context.Context) ([]domain.Graph, error)
	// Upsert inserts or replaces one harness graph bajo `clave` (HS-11: la vía del loader
	// real — «Cargar» un arnés del disco lo indexa en vivo). `clave` es la llave AUTORITATIVA
	// (deuda BACKLOG «re-key (home,id,scope)», 2026-07-23): el caller la decide — el
	// `Clave()` calificado del Portafolio para flujos que lo conocen, o un id simple para el
	// registro manual — NUNCA se re-deriva del propio `g.Arnes.ID` del grafo, que puede
	// colisionar entre dos arneses distintos (docs/architecture/boundaries/
	// portafolio-identidad-y-deriva-honesta.md). El índice sigue desechable: lo upserteado se
	// reconstruye re-cargando el directorio, jamás es fuente de verdad.
	Upsert(ctx context.Context, clave string, g domain.Graph) error
}
