package usecase_test

import (
	"context"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// TestMapServiceNode covers the inspector read (S3, RF-71): a found node carries its fused
// contract; a missing node of a known harness is (false, nil); an unknown harness is an error.
func TestMapServiceNode(t *testing.T) {
	svc := usecase.NewMapService(index.New())
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
	gs, err := usecase.NewMapService(index.New()).Harnesses(context.Background())
	if err != nil {
		t.Fatalf("Harnesses() = %v", err)
	}
	if len(gs) != 3 {
		t.Errorf("Harnesses len = %d, want 3 (demo + dev-full-cycle + content-studio-full)", len(gs))
	}
}
