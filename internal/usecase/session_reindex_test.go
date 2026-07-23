package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// fakeIndex registra los Upsert que recibe (RF-184).
type fakeIndex struct{ upserts []domain.Graph }

func (f *fakeIndex) Rebuild(context.Context) error                       { return nil }
func (f *fakeIndex) Query(context.Context, string) (domain.Graph, error) { return domain.Graph{}, nil }
func (f *fakeIndex) List(context.Context) ([]domain.Graph, error)        { return nil, nil }
func (f *fakeIndex) Upsert(_ context.Context, g domain.Graph) error {
	f.upserts = append(f.upserts, g)
	return nil
}

// TestTurnReindexerUpsertaGrafo: carga OK → el grafo entra al índice tal cual (RF-184).
func TestTurnReindexerUpsertaGrafo(t *testing.T) {
	idx := &fakeIndex{}
	re := usecase.NewTurnReindexer(idx, func(string) (domain.Graph, error) {
		return domain.Graph{Arnes: &domain.Arnes{ID: "vitalia"}, Nodes: []domain.Box{{ID: "r1"}}}, nil
	})
	re(context.Background(), "vitalia", "/tmp/x")
	if len(idx.upserts) != 1 || idx.upserts[0].Arnes.ID != "vitalia" || len(idx.upserts[0].Nodes) != 1 {
		t.Fatalf("upserts = %+v", idx.upserts)
	}
}

// TestTurnReindexerSinSello: carga OK pero sin manifiesto (Arnes nil) → misma síntesis que
// ObservarEnMapa (id = HuellaPath del cwd) + Degradado, jamás Upsert con Arnes nil.
func TestTurnReindexerSinSello(t *testing.T) {
	idx := &fakeIndex{}
	re := usecase.NewTurnReindexer(idx, func(string) (domain.Graph, error) {
		return domain.Graph{Nodes: []domain.Box{{ID: "r1"}}, Degradado: true}, nil
	})
	re(context.Background(), "vitalia", "/tmp/x")
	if len(idx.upserts) != 1 {
		t.Fatalf("upserts = %+v", idx.upserts)
	}
	g := idx.upserts[0]
	if g.Arnes == nil || g.Arnes.ID != domain.HuellaPath("/tmp/x") || !g.Degradado {
		t.Errorf("degradado sin sello mal sintetizado: %+v", g.Arnes)
	}
	if len(g.Nodes) != 1 {
		t.Errorf("los nodos reconocidos deben sobrevivir: %+v", g.Nodes)
	}
}

// TestTurnReindexerCargaRota: el chat rompió el árbol (LoadArnes error) → el índice refleja
// estado DEGRADADO real (nodos vacíos + Degradado), nunca la foto vieja ni silencio (RF-184).
func TestTurnReindexerCargaRota(t *testing.T) {
	idx := &fakeIndex{}
	re := usecase.NewTurnReindexer(idx, func(string) (domain.Graph, error) {
		return domain.Graph{}, errors.New("boom")
	})
	re(context.Background(), "vitalia", "/tmp/x")
	if len(idx.upserts) != 1 {
		t.Fatalf("upserts = %+v", idx.upserts)
	}
	g := idx.upserts[0]
	if g.Arnes == nil || g.Arnes.ID != "vitalia" || !g.Degradado || len(g.Nodes) != 0 {
		t.Errorf("carga rota debe indexar degradado honesto: %+v (nodos %d)", g.Arnes, len(g.Nodes))
	}
}

// TestReindexTrasTurno: el servicio dispara el reindex DESPUÉS de cada EventResult, con el
// cwd real de la sesión (RF-184) — y un reindex que falla jamás rompe el turno (RF-185:
// el Reindexer no devuelve error; acá se prueba el disparo).
func TestReindexTrasTurno(t *testing.T) {
	agent := &stubAgent{}
	cwd := t.TempDir()
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: cwd}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	type llamada struct{ arnes, cwd string }
	llamadas := make(chan llamada, 4)
	svc.SetReindexer(func(_ context.Context, arnesID, c string) {
		llamadas <- llamada{arnesID, c}
	})

	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "arregla la rule"); err != nil {
		t.Fatal(err)
	}
	if len(agent.sessions) != 1 {
		t.Fatalf("sesiones spawneadas: %d", len(agent.sessions))
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo", CtxPct: 10}

	select {
	case l := <-llamadas:
		if l.arnes != "vitalia" || l.cwd != cwd {
			t.Errorf("reindex con (%q,%q), quiero (vitalia,%q)", l.arnes, l.cwd, cwd)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("el reindex no se disparó tras EventResult")
	}
}
