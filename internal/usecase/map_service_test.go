package usecase_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// newTestIndex opens a disposable SQLite index (Store) for these usecase-level tests:
// no ArnesRegistry/loader wired since none call Rebuild — they exercise Query/List
// straight off the seeded demo/dogfood graphs, same as index.New always provides.
func newTestIndex(t *testing.T) *index.Store {
	t.Helper()
	idx, err := index.New(filepath.Join(t.TempDir(), "index.db"), nil, nil)
	if err != nil {
		t.Fatalf("index.New: %v", err)
	}
	t.Cleanup(func() { _ = idx.Close() })
	return idx
}

// TestMapServiceNode covers the inspector read (S3, RF-71): a found node carries its fused
// contract; a missing node of a known harness is (false, nil); an unknown harness is an error.
func TestMapServiceNode(t *testing.T) {
	svc := usecase.NewMapService(newTestIndex(t))
	ctx := context.Background()

	box, ok, err := svc.Node(ctx, "dev-full-cycle", "spec-writer")
	if err != nil || !ok {
		t.Fatalf("Node(spec-writer) = ok:%v err:%v, want ok:true", ok, err)
	}
	if !box.IsCaja() {
		t.Error("spec-writer.IsCaja() = false, want true")
	}
	if box.Contract == nil || box.Contract.Why == "" {
		t.Error("spec-writer contract/why missing (inspector needs the fused contract)")
	}

	if _, ok, err := svc.Node(ctx, "dev-full-cycle", "nope"); ok || err != nil {
		t.Errorf("Node(nope) = ok:%v err:%v, want ok:false err:nil", ok, err)
	}
	if _, _, err := svc.Node(ctx, "ghost-arnes", "x"); err == nil {
		t.Error("Node(ghost-arnes) err = nil, want error (unknown harness)")
	}
}

// TestMapServiceHarnesses covers the portfolio read (S1, RF-72).
func TestMapServiceHarnesses(t *testing.T) {
	gs, err := usecase.NewMapService(newTestIndex(t)).Harnesses(context.Background())
	if err != nil {
		t.Fatalf("Harnesses() = %v", err)
	}
	if len(gs) != 3 {
		t.Errorf("Harnesses len = %d, want 3 (demo + dev-full-cycle + content-studio-full)", len(gs))
	}
}
