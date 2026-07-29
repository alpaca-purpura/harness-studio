package main

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// indiceFalso implementa ports.IndexPort con un set fijo de claves "conocidas" — lo único
// que avisarCoberturaPortafolioIndice necesita (Query), el resto no se llama en el test.
type indiceFalso struct{ conocidas map[string]bool }

func (f indiceFalso) Rebuild(context.Context) error { return nil }
func (f indiceFalso) Query(_ context.Context, id string) (domain.Graph, error) {
	if f.conocidas[id] {
		return domain.Graph{Arnes: &domain.Arnes{ID: id}}, nil
	}
	return domain.Graph{}, errors.New("no encontrado")
}
func (f indiceFalso) List(context.Context) ([]domain.Graph, error)       { return nil, nil }
func (f indiceFalso) Upsert(context.Context, string, domain.Graph) error { return nil }

func capturarLogs(t *testing.T, fn func()) string {
	t.Helper()
	previo := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previo) })
	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	fn()
	return buf.String()
}

// TestAvisarCoberturaPortafolioIndiceCallaSiTodoCubierto: cero drift, cero ruido — mismo
// criterio que loguearRecalibracion.
func TestAvisarCoberturaPortafolioIndiceCallaSiTodoCubierto(t *testing.T) {
	store, err := portafolio.NewStore(filepath.Join(t.TempDir(), "portafolio.json"))
	if err != nil {
		t.Fatal(err)
	}
	id := domain.IdentidadArnes{Home: "github.com/o/r", ID: "harness-x"}
	if err := store.Upsert(domain.EntradaPortafolio{Identidad: id}); err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewPortafolioService(store, nil, nil, nil, nil)
	idx := indiceFalso{conocidas: map[string]bool{id.Clave(): true}}

	out := capturarLogs(t, func() { avisarCoberturaPortafolioIndice(context.Background(), svc, idx) })
	if strings.Contains(out, "sin observar en el Mapa") {
		t.Fatalf("cobertura completa no debía loguear nada, got: %s", out)
	}
}

// TestAvisarCoberturaPortafolioIndiceDiceCuantasFaltan: la brecha real (D4) — N entradas
// sanas del Portafolio que el índice no conoce se reportan, con la cuenta y las claves.
func TestAvisarCoberturaPortafolioIndiceDiceCuantasFaltan(t *testing.T) {
	store, err := portafolio.NewStore(filepath.Join(t.TempDir(), "portafolio.json"))
	if err != nil {
		t.Fatal(err)
	}
	cubierta := domain.IdentidadArnes{Home: "github.com/o/r", ID: "cubierta"}
	sinCubrir := domain.IdentidadArnes{Home: "github.com/o/r", ID: "sin-cubrir"}
	if err := store.Upsert(domain.EntradaPortafolio{Identidad: cubierta}); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert(domain.EntradaPortafolio{Identidad: sinCubrir}); err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewPortafolioService(store, nil, nil, nil, nil)
	idx := indiceFalso{conocidas: map[string]bool{cubierta.Clave(): true}}

	out := capturarLogs(t, func() { avisarCoberturaPortafolioIndice(context.Background(), svc, idx) })
	if !strings.Contains(out, "sin observar en el Mapa") {
		t.Fatalf("quiero el aviso de cobertura, got: %s", out)
	}
	if !strings.Contains(out, "sin_indice=1") {
		t.Fatalf("quiero sin_indice=1 (una sola entrada sin cubrir), got: %s", out)
	}
	if !strings.Contains(out, "de_un_total=2") {
		t.Fatalf("quiero de_un_total=2, got: %s", out)
	}
	if !strings.Contains(out, sinCubrir.Clave()) {
		t.Fatalf("quiero la clave concreta sin cubrir en el log, got: %s", out)
	}
	if strings.Contains(out, cubierta.Clave()) {
		t.Fatalf("la clave YA cubierta no debía aparecer en la lista de faltantes, got: %s", out)
	}
}
