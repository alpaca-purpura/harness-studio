package usecase_test

import (
	"context"
	"strings"
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

// TestRotacionInvisible (RF-196/197/198): con RotacionPendiente, el próximo Turn cierra el
// proceso viejo, encadena el ClaudeSessionID, spawnea FRESCO (sin --resume) con el
// checkpoint en el system-prompt por sesión, deja breadcrumb RolSys y el Conv no pierde
// ningún turno. Session.ID no cambia.
func TestRotacionInvisible(t *testing.T) {
	agent := &stubAgent{}
	inj := &stubInjector{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, inj, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetUmbralRotacion(40)
	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}

	// Turno 1: init con ClaudeSessionID real + result que cruza el umbral.
	if err := svc.Turn(s.ID, "primer pedido"); err != nil {
		t.Fatal(err)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-viejo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return m.RotacionPendiente })

	// Turno 2: rota por detrás y sigue.
	if err := svc.Turn(s.ID, "segundo pedido"); err != nil {
		t.Fatal(err)
	}
	if len(agent.spawns) != 2 {
		t.Fatalf("spawns = %d, quiero 2 (proceso fresco)", len(agent.spawns))
	}
	if agent.spawns[1].Resume != "" {
		t.Errorf("el spawn post-rotación debe ir SIN --resume, got %q", agent.spawns[1].Resume)
	}
	m, _ := svc.Get(s.ID)
	if m.ID != s.ID {
		t.Error("Session.ID no debe cambiar")
	}
	if len(m.CadenaCC) != 1 || m.CadenaCC[0] != "cc-viejo" {
		t.Errorf("CadenaCC = %v, quiero [cc-viejo]", m.CadenaCC)
	}
	if m.RotacionPendiente {
		t.Error("la rotación debe consumir la marca")
	}
	if m.Checkpoint == "" || !strings.Contains(m.Checkpoint, "primer pedido") {
		t.Errorf("checkpoint mecánico sin los últimos turnos: %q", m.Checkpoint)
	}
	// El checkpoint viaja en el system-prompt por sesión del spawn 2.
	if len(inj.extras) != 2 || !strings.Contains(inj.extras[1], "Checkpoint de rotación") || !strings.Contains(inj.extras[1], "primer pedido") {
		t.Errorf("extras del injector: %d, último sin checkpoint: %.120q", len(inj.extras), inj.extras[len(inj.extras)-1])
	}
	// Conv íntegro + breadcrumb: user1, assistant1, sys, user2.
	roles := []string{}
	for _, tu := range m.Conv {
		roles = append(roles, string(tu.Rol))
	}
	quiero := []string{"user", "assistant", "sys", "user"}
	if len(roles) != len(quiero) {
		t.Fatalf("roles del Conv = %v, quiero %v", roles, quiero)
	}
	for i := range quiero {
		if roles[i] != quiero[i] {
			t.Fatalf("roles del Conv = %v, quiero %v", roles, quiero)
		}
	}
}
