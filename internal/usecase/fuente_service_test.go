package usecase_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/artifact"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// resolverFijo is a WorkdirResolver stub: one arnés id maps to dir (registered);
// everything else resolves to an unregistered fallback.
type resolverFijo struct {
	arnes string
	dir   string
}

func (r resolverFijo) Resolve(arnesID string) (string, bool, error) {
	if arnesID == r.arnes {
		return r.dir, true, nil
	}
	return filepath.Join(r.dir, "fallback", arnesID), false, nil
}

// arnesConFuente seeds the index with one harness whose nodes exercise the fuente
// taxonomy: a node with a relative fuente_path, one without, and one that escapes.
func arnesConFuente(t *testing.T, id, dir string) *index.Store {
	t.Helper()
	idx, err := index.New(filepath.Join(t.TempDir(), "index.db"), nil, nil)
	if err != nil {
		t.Fatalf("index.New: %v", err)
	}
	t.Cleanup(func() { _ = idx.Close() })
	g := domain.Graph{
		Arnes: &domain.Arnes{ID: id, ReportaA: nil},
		Nodes: []domain.Box{
			{ID: "con-fuente", Clase: domain.ClaseSkill, Nombre: "con fuente", FuentePath: "skills/x/SKILL.md"},
			{ID: "sin-fuente", Clase: domain.ClaseHook, Nombre: "sin fuente"},
			{ID: "escapista", Clase: domain.ClaseSkill, Nombre: "escapa", FuentePath: "../fuera.md"},
			{ID: "absoluta", Clase: domain.ClaseRule, Nombre: "abs", FuentePath: filepath.Join(dir, "CLAUDE.md")},
		},
	}
	if err := idx.Upsert(context.Background(), id, g); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	return idx
}

// TestFuenteService covers RF-93 end-to-end at the use-case level: happy path
// (relative + absolute-inside fuente_path), and the honest error taxonomy — no
// fuente_path, unknown node, unregistered arnés, traversal escape, unknown harness.
func TestFuenteService(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "skills", "x"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "skills", "x", "SKILL.md"), []byte("---\nname: x\n---\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# regla"), 0o600); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	idx := arnesConFuente(t, "af", dir)
	svc := usecase.NewFuenteService(idx, resolverFijo{arnes: "af", dir: dir}, artifact.NewFuenteReader())

	path, contenido, err := svc.Fuente(ctx, "af", "con-fuente")
	if err != nil {
		t.Fatalf("Fuente(con-fuente) = %v, want contenido", err)
	}
	if path != "skills/x/SKILL.md" || string(contenido) != "---\nname: x\n---\n" {
		t.Errorf("Fuente(con-fuente) = %q %q — path/contenido inesperados", path, contenido)
	}

	if _, contenido, err = svc.Fuente(ctx, "af", "absoluta"); err != nil || string(contenido) != "# regla" {
		t.Errorf("Fuente(absoluta dentro del dir) = %q err:%v, want «# regla»", contenido, err)
	}

	if _, _, err = svc.Fuente(ctx, "af", "sin-fuente"); !errors.Is(err, usecase.ErrFuenteNoDeclarada) {
		t.Errorf("Fuente(sin-fuente) err = %v, want ErrFuenteNoDeclarada", err)
	}
	if _, _, err = svc.Fuente(ctx, "af", "nope"); !errors.Is(err, usecase.ErrNodoNoEncontrado) {
		t.Errorf("Fuente(nope) err = %v, want ErrNodoNoEncontrado", err)
	}
	if _, _, err = svc.Fuente(ctx, "af", "escapista"); !errors.Is(err, ports.ErrFueraDelArnes) {
		t.Errorf("Fuente(escapista) err = %v, want ErrFueraDelArnes", err)
	}
	if _, _, err = svc.Fuente(ctx, "ghost", "x"); err == nil {
		t.Error("Fuente(ghost) err = nil, want error (arnés desconocido)")
	}

	// dev-full-cycle vive en el índice embebido pero SIN directorio registrado → 404 honesto.
	if _, _, err = svc.Fuente(ctx, "dev-full-cycle", "spec-writer"); !errors.Is(err, usecase.ErrArnesSinDirectorio) {
		t.Errorf("Fuente(no registrado) err = %v, want ErrArnesSinDirectorio", err)
	}
}
