package ports

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// IndexPort is the disposable index over the JSONL source of truth: it can always be
// rebuilt from scratch (see docs/architecture/boundaries/indice-desechable-jsonl-es-verdad.md).
type IndexPort interface {
	// Rebuild reconstructs the whole index from the JSONL corpus.
	Rebuild(ctx context.Context) error
	// Query returns the agnostic graph of one harness.
	Query(ctx context.Context, harnessID string) (domain.Graph, error)
	// List returns every harness graph currently indexed (portfolio, S1).
	List(ctx context.Context) ([]domain.Graph, error)
	// Upsert inserts or replaces one harness graph (HS-11: la vía del loader real —
	// «Cargar» un arnés del disco lo indexa en vivo). El índice sigue desechable: lo
	// upserteado se reconstruye re-cargando el directorio, jamás es fuente de verdad.
	Upsert(ctx context.Context, g domain.Graph) error
}
