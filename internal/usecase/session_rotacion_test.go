package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// TestCtxHistYUmbralRotacion (RF-194/195): el histórico crece por turno y cruzar el umbral
// marca RotacionPendiente — después de responder, jamás a mitad del turno.
func TestCtxHistYUmbralRotacion(t *testing.T) {
	agent := &stubAgent{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)

	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	turno := func(ctx int) {
		if err := svc.Turn(s.ID, "hola"); err != nil {
			t.Fatal(err)
		}
		agent.sessions[len(agent.sessions)-1].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "ok", CtxPct: ctx}
		espera(t, func() bool { m, _ := svc.Get(s.ID); return m.Status == domain.StatusIdle && m.CtxPct == ctx })
	}

	turno(10)
	m, _ := svc.Get(s.ID)
	if len(m.CtxHist) != 1 || m.CtxHist[0] != 10 || m.RotacionPendiente {
		t.Fatalf("tras turno 1: hist=%v pendiente=%v", m.CtxHist, m.RotacionPendiente)
	}

	turno(45)
	m, _ = svc.Get(s.ID)
	if len(m.CtxHist) != 2 || m.CtxHist[1] != 45 {
		t.Errorf("hist=%v, quiero [10 45]", m.CtxHist)
	}
	if !m.RotacionPendiente {
		t.Error("45 >= umbral 40 debe marcar RotacionPendiente")
	}
}

func espera(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timeout esperando condición")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
