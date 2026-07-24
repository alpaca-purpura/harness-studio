// Package index implements ports.IndexPort in memory (a map). The JSONL under
// ~/.claude is the source of truth; this index is disposable and rebuildable.
//
// TODO(fase 5): modernc.org/sqlite (pure-Go, WAL) as the on-disk disposable index —
// deleting the .db and reindexing must reproduce the same queryable state.
package index

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"

	"github.com/alpacapurpura/arnesia/dogfood"
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

// List returns every indexed harness graph, ordered by clave for a stable portfolio (S1).
func (s *Store) List(_ context.Context) ([]domain.Graph, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	claves := make([]string, 0, len(s.graphs))
	for k := range s.graphs {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	out := make([]domain.Graph, 0, len(claves))
	for _, k := range claves {
		out = append(out, s.graphs[k])
	}
	return out, nil
}

// Upsert inserts or replaces one harness graph bajo `clave` — la llave AUTORITATIVA que el
// caller decide (deuda BACKLOG «re-key (home,id,scope)», 2026-07-23: nunca se re-deriva del
// propio `g.Arnes.ID`, que dos arneses distintos pueden compartir). Un grafo sin manifiesto
// no es indexable (no hay nada que mostrar) — error explícito, jamás un grafo inventado.
func (s *Store) Upsert(_ context.Context, clave string, g domain.Graph) error {
	if clave == "" {
		return errors.New("index: clave vacía — no indexable")
	}
	if g.Arnes == nil {
		return errors.New("index: grafo sin manifiesto (arnes.id) — no indexable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphs[clave] = g
	return nil
}

// seed loads a minimal demo harness so the wiring is observable.
func (s *Store) seed() {
	demo := domain.Graph{
		Arnes: &domain.Arnes{
			ID:       "demo",
			Rol:      "backend",
			Proceso:  "desarrollo",
			Empresas: []string{"alpacapurpura"},
			ReportaA: nil, // raíz (emite null, válido contra graph.l0).
			Canal:    domain.CanalBeta,
			Fases:    []domain.Fase{"spec"},
			Spine: &domain.Spine{
				Inicial:      "grill",
				Terminales:   []string{"spec"},
				Estados:      []string{"grill", "spec"},
				Transiciones: []domain.Transicion{{De: "grill", A: "spec"}},
			},
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
					Why:       "convertir la conversación en un spec ejecutable",
					Clase:     domain.ClaseSkill,
					Arquetipo: domain.ArqExcepcion,
					Perfil:    domain.PerfilT2,
					Caja:      true,
					Fase:      "spec",
					Estado:    "grill -> spec",
					Gate:      &domain.Gate{Tipo: domain.GateManual, Detalle: "revisión humana del spec"},
				},
			},
			{
				ID:     "std-go",
				Clase:  domain.ClaseRule,
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
	// Seed the embedded arnés graphs so GET /api/harnesses/{id}/graph serves them and the
	// picker can alternate: dev-full-cycle (the REAL, honest dogfood) + content-studio-full
	// (the showcase kitchen-sink, HS-09 Hito 2). A decode failure of an embedded, tested
	// asset is a programmer error, but we log-and-continue so the demo still serves.
	for name, raw := range map[string][]byte{
		"dev-full-cycle":      dogfood.DevFullCycleJSON,
		"content-studio-full": dogfood.ContentStudioFullJSON,
	} {
		if g, err := decodeGraph(raw); err != nil {
			slog.Error("index: seed embedded graph", "arnes", name, "err", err)
		} else if g.Arnes != nil {
			s.graphs[g.Arnes.ID] = g
		}
	}
	s.mu.Unlock()
}

// decodeGraph unmarshals an embedded L0 graph. The JSON keys mirror the domain tags
// exactly, so it decodes straight into a Graph with no field mapping.
func decodeGraph(raw []byte) (domain.Graph, error) {
	var g domain.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		return domain.Graph{}, fmt.Errorf("decode embedded graph: %w", err)
	}
	return g, nil
}
