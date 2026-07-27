package usecase_test

import (
	"context"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// TestGetDevuelveInstantaneaNoLaSesionViva (CV-D3): la sesión que devuelve el servicio
// cruza el candado y la lee cualquiera. Con el estado del conductor bajado a un slice de
// conversaciones, devolver `*r.meta` compartía el backing array: cada campo que el
// conductor movía (rotación pendiente, ctx, id de Claude Code) se escribía sobre la
// memoria que el lector YA tenía en la mano.
//
// Con la sesión plana esto no podía pasar — esos campos se copiaban por valor. Es una
// consecuencia de partir el agregado, y el test la fija: la copia entregada es un
// instante, no una ventana al registro vivo.
func TestGetDevuelveInstantaneaNoLaSesionViva(t *testing.T) {
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
	if turnErr := svc.Turn(s.ID, "primer pedido"); turnErr != nil {
		t.Fatal(turnErr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-uno"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 12}
	espera(t, func() bool {
		m, _ := svc.Get(s.ID)
		return m.Status == domain.StatusIdle && activaSinFallar(m).CtxPct == 12
	})

	// El instante que el lector se lleva.
	antes, _ := svc.Get(s.ID)
	convAntes := activaDe(t, antes)

	// El conductor sigue trabajando: otro turno mueve ctx y cruza el umbral.
	if turnErr := svc.Turn(s.ID, "segundo pedido"); turnErr != nil {
		t.Fatal(turnErr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 77}
	espera(t, func() bool {
		m, _ := svc.Get(s.ID)
		return m.Status == domain.StatusIdle && activaSinFallar(m).CtxPct == 77
	})

	// El instante NO se movió.
	if got := antes.Conversaciones[0].CtxPct; got != 12 {
		t.Errorf("la copia entregada mutó: ctx_pct = %d, quiero el 12 del instante", got)
	}
	if antes.Conversaciones[0].RotacionPendiente {
		t.Error("la copia entregada mutó: se marcó la rotación que ocurrió DESPUÉS de entregarla")
	}
	if convAntes.CtxPct != 12 {
		t.Errorf("la conversación leída del instante mutó: ctx_pct = %d", convAntes.CtxPct)
	}

	// Y la sesión viva sí avanzó — el instante es viejo, no es una copia congelada de todo.
	ahora, _ := svc.Get(s.ID)
	if c := activaDe(t, ahora); c.CtxPct != 77 || !c.RotacionPendiente {
		t.Errorf("la sesión viva no avanzó: ctx_pct=%d pendiente=%v", c.CtxPct, c.RotacionPendiente)
	}
}

// TestInstantaneaNoCopiaElTranscript fija la otra mitad de la decisión: Instantanea copia
// el slice de conversaciones y NO los slices de adentro. Copiarlos sería copiar el
// transcript entero en cada List(), que es el costo que el método existe para no pagar.
// Es seguro porque a esos slices sólo se les appendea: el lector ve un prefijo estable.
func TestInstantaneaNoCopiaElTranscript(t *testing.T) {
	s := domain.Session{ID: "s1", Conversaciones: []domain.Conversacion{{
		ID: "cv1", Activa: true, Conv: []domain.Turn{{Rol: domain.RolUser, Text: "uno"}},
	}}}
	copia := s.Instantanea()

	// El slice de conversaciones es otro: tocar la copia no toca el original.
	copia.Conversaciones[0].Titulo = "cambiado en la copia"
	if s.Conversaciones[0].Titulo != "" {
		t.Errorf("la copia escribió sobre el original: título = %q", s.Conversaciones[0].Titulo)
	}

	// Los turnos se comparten a propósito — declarado, no accidental.
	if &copia.Conversaciones[0].Conv[0] != &s.Conversaciones[0].Conv[0] {
		t.Error("Instantanea copió el transcript: eso es el costo que el método evita")
	}
}
