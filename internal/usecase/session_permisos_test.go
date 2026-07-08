package usecase_test

// Tests del paquete chat-cc-funcional (RF-112 · RF-114 · RF-116): el spawn del Dock
// materializa los permisos del ROL del arnés; una resolución sin rol usa la autoridad
// del arnés; Interrupt deniega los asks pendientes y viaja in-band.

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/permission"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

type stubSession struct {
	mu         sync.Mutex
	events     chan ports.AgentEvent
	responded  []ports.ControlDecision
	interrupts int
}

func (s *stubSession) Send(context.Context, string) error { return nil }
func (s *stubSession) Events() <-chan ports.AgentEvent    { return s.events }
func (s *stubSession) Close() error                       { return nil }
func (s *stubSession) Interrupt(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interrupts++
	return nil
}

func (s *stubSession) RespondControl(_ context.Context, _ string, d ports.ControlDecision) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responded = append(s.responded, d)
	return nil
}

func (s *stubSession) snapshot() (int, []ports.ControlDecision) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ports.ControlDecision, len(s.responded))
	copy(out, s.responded)
	return s.interrupts, out
}

type stubAgent struct {
	mu       sync.Mutex
	spawns   []ports.SpawnOpts
	sessions []*stubSession
}

func (a *stubAgent) Spawn(_ context.Context, opts ports.SpawnOpts) (ports.AgentSession, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.spawns = append(a.spawns, opts)
	s := &stubSession{events: make(chan ports.AgentEvent, 16)}
	a.sessions = append(a.sessions, s)
	return s, nil
}

type stubStore struct{}

func (stubStore) Load(context.Context) ([]domain.Session, error) { return nil, nil }
func (stubStore) Save(context.Context, []domain.Session) error   { return nil }

type stubResolver struct{ path string }

func (r stubResolver) Resolve(string) (string, bool, error) { return r.path, true, nil }

type stubPub struct{}

func (stubPub) Publish(string, []byte) {}

func newSvc(t *testing.T, agent ports.AgentPort, roleFor usecase.RoleSource) *usecase.SessionService {
	t.Helper()
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, permission.NewKitProvisioner(), roleFor)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	return svc
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condition not met within timeout")
}

// RF-112: el spawn del Dock lleva el permission-set del rol del arnés — resuelto
// server-side por la RoleSource, jamás por el FE.
func TestDockSpawnCarriesRolePermissions(t *testing.T) {
	agent := &stubAgent{}
	roleFor := func(_ context.Context, arnesID string) string {
		if arnesID == "dev-full-cycle" {
			return "Ingeniería · Desarrollo full-cycle"
		}
		return ""
	}
	svc := newSvc(t, agent, roleFor)
	id := svc.List()[0].ID // seed: la primera sesión apunta a dev-full-cycle.
	if err := svc.Turn(id, "edita el gate de builder"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	agent.mu.Lock()
	defer agent.mu.Unlock()
	if len(agent.spawns) != 1 {
		t.Fatalf("want 1 spawn, got %d", len(agent.spawns))
	}
	ps := agent.spawns[0].Permisos
	if ps.Rol != "Ingeniería · Desarrollo full-cycle" {
		t.Errorf("Permisos.Rol = %q, want el rol del arnés (RF-112)", ps.Rol)
	}
	if ps.Decide("Edit") != domain.DecisionAllow {
		t.Errorf("el rol full-cycle debe poder aprobar Edit (allow=aprobable), got %v", ps.Decide("Edit"))
	}
}

// Degradación honesta: rol irresoluble ⇒ spawn SIN flags de permisos (zero-value), la
// sesión jamás se bloquea por la autoridad ausente.
func TestDockSpawnDegradesWithoutRole(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "hola"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	agent.mu.Lock()
	defer agent.mu.Unlock()
	if got := agent.spawns[0].Permisos; got.Rol != "" || len(got.Allow) != 0 {
		t.Errorf("sin rol el spawn debe ir sin permisos, got %+v", got)
	}
}

// RF-114: resolver un permiso SIN rol en el request usa la autoridad del arnés.
func TestResolvePermissionUsesArnesRole(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "Ingeniería · Desarrollo full-cycle" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "edita"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	sess := agent.sessions[0]
	sess.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-1", Tool: "Edit", Input: []byte(`{"file_path":"x"}`), ToolUseID: "toolu_1"}
	waitUntil(t, func() bool {
		s, _ := svc.Get(id)
		return s.Status == domain.StatusAwait
	})

	res, err := svc.ResolvePermission(id, "cr-1", "allow", "", 0) // rol vacío ⇒ del arnés.
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if res.Efectiva != domain.DecisionAllow || res.Expira == nil {
		t.Fatalf("resolución = %+v, want allow con grant del rol del arnés", res)
	}
	_, responded := sess.snapshot()
	if len(responded) != 1 || !responded[0].Allow || responded[0].ToolUseID != "toolu_1" {
		t.Errorf("control_response = %+v, want allow con toolUseID ecoado", responded)
	}
	if len(responded[0].UpdatedInput) == 0 {
		t.Error("allow sin updatedInput — el wire lo requiere (echo del input original)")
	}
}

// RF-116: Interrupt deniega los asks pendientes (con motivo) y manda el interrupt
// in-band; la sesión no queda await.
func TestInterruptDeniesPendingAndSignalsConductor(t *testing.T) {
	agent := &stubAgent{}
	svc := newSvc(t, agent, func(context.Context, string) string { return "Ingeniería · Desarrollo full-cycle" })
	id := svc.List()[0].ID
	if err := svc.Turn(id, "edita"); err != nil {
		t.Fatalf("turn: %v", err)
	}
	sess := agent.sessions[0]
	sess.events <- ports.AgentEvent{Kind: ports.EventControlRequest, RequestID: "cr-9", Tool: "Write", Input: []byte(`{}`), ToolUseID: "toolu_9"}
	waitUntil(t, func() bool {
		s, _ := svc.Get(id)
		return s.Status == domain.StatusAwait
	})

	if err := svc.Interrupt(id); err != nil {
		t.Fatalf("interrupt: %v", err)
	}
	interrupts, responded := sess.snapshot()
	if interrupts != 1 {
		t.Errorf("interrupts = %d, want 1 (in-band, no kill)", interrupts)
	}
	if len(responded) != 1 || responded[0].Allow || !strings.Contains(responded[0].Message, "interrumpido") {
		t.Errorf("el ask pendiente debía denegarse con motivo, got %+v", responded)
	}
	s, _ := svc.Get(id)
	if s.Status == domain.StatusAwait {
		t.Error("tras interrupt la sesión no puede quedar await")
	}

	// Sin turno en vuelo, Interrupt es 409 honesto.
	sess.events <- ports.AgentEvent{Kind: ports.EventResult, Text: "done"}
	waitUntil(t, func() bool {
		s2, _ := svc.Get(id)
		return s2.Status == domain.StatusIdle
	})
	if err := svc.Interrupt(id); err == nil {
		t.Error("interrupt sin turno en vuelo debía fallar (ErrNadaQueInterrumpir)")
	}
}
