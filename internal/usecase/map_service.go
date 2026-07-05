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

// Rebuild reconstructs the disposable index from the JSONL corpus.
func (s *MapService) Rebuild(ctx context.Context) error {
	return s.index.Rebuild(ctx)
}
