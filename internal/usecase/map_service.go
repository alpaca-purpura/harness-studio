// Package usecase holds the services that orchestrate the domain over the ports. It
// depends only on domain and ports — never on transport or concrete adapters.
package usecase

import (
	"context"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// MapService assembles the agnostic graph that the Map (S2) renders. It reads through
// the index port; the JSONL corpus remains the source of truth.
type MapService struct {
	index ports.IndexPort
}

// NewMapService returns a MapService backed by index.
func NewMapService(index ports.IndexPort) *MapService {
	return &MapService{index: index}
}

// Graph returns the agnostic graph of one harness.
func (s *MapService) Graph(ctx context.Context, harnessID string) (domain.Graph, error) {
	return s.index.Query(ctx, harnessID)
}

// Harnesses returns every indexed harness graph — the portfolio the picker lists (S1, RF-72).
func (s *MapService) Harnesses(ctx context.Context) ([]domain.Graph, error) {
	return s.index.List(ctx)
}

// Node returns one node of a harness graph for the inspector (S3, RF-71). The bool is false
// when the harness exists but has no such node; err is set when the harness is unknown.
func (s *MapService) Node(ctx context.Context, harnessID, nodeID string) (domain.Box, bool, error) {
	g, err := s.index.Query(ctx, harnessID)
	if err != nil {
		return domain.Box{}, false, err
	}
	b, ok := g.NodeByID(nodeID)
	return b, ok, nil
}

// Rebuild reconstructs the disposable index from the JSONL corpus.
func (s *MapService) Rebuild(ctx context.Context) error {
	return s.index.Rebuild(ctx)
}
