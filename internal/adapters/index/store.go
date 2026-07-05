// Package index implements ports.IndexPort in memory (a map). The JSONL under
// ~/.claude is the source of truth; this index is disposable and rebuildable.
//
// TODO(fase 5): modernc.org/sqlite (pure-Go, WAL) as the on-disk disposable index —
// deleting the .db and reindexing must reproduce the same queryable state.
package index

import (
	"context"
	"errors"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// ErrNotFound is returned when a harness id is unknown to the index.
var ErrNotFound = errors.New("harness not found")

// Store is an in-memory ports.IndexPort. Safe for concurrent use.
type Store struct {
	mu     sync.RWMutex
	graphs map[string]domain.Graph
}

var _ ports.IndexPort = (*Store)(nil)

// New returns an index seeded with a small demo graph so the daemon is runnable
// end-to-end before the real JSONL indexer lands.
func New() *Store {
	s := &Store{graphs: map[string]domain.Graph{}}
	s.seed()
	return s
}

// Rebuild reconstructs the index. Stub: re-seeds the demo graph.
// TODO(fase 5): walk the ~/.claude JSONL corpus and rebuild from it.
func (s *Store) Rebuild(_ context.Context) error {
	s.seed()
	return nil
}

// Query returns the graph of one harness, or ErrNotFound.
func (s *Store) Query(_ context.Context, harnessID string) (domain.Graph, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	g, ok := s.graphs[harnessID]
	if !ok {
		return domain.Graph{}, ErrNotFound
	}
	return g, nil
}

// seed loads a minimal demo harness so the wiring is observable.
func (s *Store) seed() {
	demo := domain.Graph{
		Arnes: domain.Arnes{
			ID:      "demo",
			Puesto:  "backend",
			Empresa: "alpacapurpura",
			Canal:   domain.CanalBeta,
		},
		Nodes: []domain.Box{
			{
				ID:     "guard-format",
				Clase:  domain.ClaseHook,
				Nombre: "gofmt guard",
				Banda:  domain.BandaGuardia,
			},
			{
				ID:     "spec",
				Clase:  domain.ClaseSkill,
				Nombre: "escribir spec",
				Banda:  domain.BandaFase,
				Fase:   "spec",
				Estado: "grill -> spec",
				Contract: &domain.Contract{
					Caja:   true,
					Fase:   "spec",
					Estado: "grill -> spec",
					Gate:   &domain.Gate{Tipo: domain.GateManual},
				},
			},
			{
				ID:     "std-go",
				Clase:  domain.ClaseKnowledge,
				Nombre: "estándar Go",
				Banda:  domain.BandaBase,
			},
		},
		Edges: []domain.Edge{
			{De: "spec", A: "std-go", Tipo: domain.EdgeLee},
		},
	}
	s.mu.Lock()
	s.graphs[demo.Arnes.ID] = demo
	s.mu.Unlock()
}
