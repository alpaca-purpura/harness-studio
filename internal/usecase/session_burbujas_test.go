package usecase_test

// Tests del paquete chat-dock-ux (CH-D2/CH-D3): el turno se parte en burbujas por
// actividad — cada EventMessage cierra su burbuja, cada EventActivity deja rastro RolAct,
// y el result NO duplica lo ya flusheado.

import (
	"context"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

func TestTurnoSeParteEnBurbujasPorActividad(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "repara el hook"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	sess := agent.sessions[0]
	sess.events <- ports.AgentEvent{Kind: ports.EventDelta, Text: "Voy a mirar."}
	sess.events <- ports.AgentEvent{Kind: ports.EventMessage, Text: "Voy a mirar."}
	sess.events <- ports.AgentEvent{Kind: ports.EventActivity, Tool: "thinking"}
	sess.events <- ports.AgentEvent{Kind: ports.EventActivity, Tool: "Read", Text: "hooks/sellar.sh"}
	sess.events <- ports.AgentEvent{Kind: ports.EventDelta, Text: "Listo."}
	sess.events <- ports.AgentEvent{Kind: ports.EventMessage, Text: "Listo."}
	sess.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "Voy a mirar.Listo."}
	waitUntil(t, func() bool {
		s, _ := svc.Get(id)
		return s.Status == domain.StatusIdle
	})

	s, _ := svc.Get(id)
	want := []domain.Turn{
		{Rol: domain.RolUser, Text: "repara el hook"},
		{Rol: domain.RolAssistant, Text: "Voy a mirar."},
		{Rol: domain.RolAct, Text: "thinking"},
		{Rol: domain.RolAct, Text: "Read hooks/sellar.sh"},
		{Rol: domain.RolAssistant, Text: "Listo."},
	}
	c := activaDe(t, s)
	if len(c.Conv) != len(want) {
		t.Fatalf("conv = %+v, want %+v", c.Conv, want)
	}
	for i, w := range want {
		if c.Conv[i] != w {
			t.Errorf("conv[%d] = %+v, want %+v", i, c.Conv[i], w)
		}
	}
}

// Fallback: un turno SIN EventMessage (p. ej. resume viejo) conserva el comportamiento
// previo — el result arma la única burbuja desde los deltas o su propio texto.
func TestResultSinMensajesConservaFallback(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "hola"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	sess := agent.sessions[0]
	sess.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hola humano"}
	waitUntil(t, func() bool {
		s, _ := svc.Get(id)
		return s.Status == domain.StatusIdle
	})
	s, _ := svc.Get(id)
	c := activaDe(t, s)
	if n := len(c.Conv); n != 2 || c.Conv[1] != (domain.Turn{Rol: domain.RolAssistant, Text: "hola humano"}) {
		t.Errorf("conv = %+v, want [user, assistant hola humano]", c.Conv)
	}
}
