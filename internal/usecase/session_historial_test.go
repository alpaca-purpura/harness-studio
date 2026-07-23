package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// memSessionStore es un SessionStore en memoria para el registro de cerradas.
type memSessionStore struct{ sesiones []domain.Session }

func (m *memSessionStore) Load(context.Context) ([]domain.Session, error) {
	return append([]domain.Session{}, m.sesiones...), nil
}

func (m *memSessionStore) Save(_ context.Context, s []domain.Session) error {
	m.sesiones = append([]domain.Session{}, s...)
	return nil
}

// fakeHistory devuelve turnos por (cwd,ccid) o error si no existe.
type fakeHistory struct{ corpus map[string][]domain.Turn }

func (f *fakeHistory) Turnos(cwd, ccid string) ([]domain.Turn, error) {
	t, ok := f.corpus[cwd+"|"+ccid]
	if !ok {
		return nil, fmt.Errorf("sin jsonl para %s", ccid)
	}
	return t, nil
}

// TestCloseArchivaMetadata (RF-200): cerrar deja metadata liviana (cadena completa, cwd,
// fecha, cantidad de turnos) SIN Conv — el historial deja de morir con la pestaña.
func TestCloseArchivaMetadata(t *testing.T) {
	agent := &stubAgent{}
	cerradas := &memSessionStore{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetArchivoCerradas(cerradas)

	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Turn(s.ID, "hola"); err != nil {
		t.Fatal(err)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-vivo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo", CtxPct: 5}
	espera(t, func() bool {
		m, _ := svc.Get(s.ID)
		return m.Status == domain.StatusIdle && m.ClaudeSessionID == "cc-vivo"
	})

	if err := svc.Close(s.ID); err != nil {
		t.Fatal(err)
	}
	lista, err := svc.Cerradas(context.Background(), "vitalia")
	if err != nil || len(lista) != 1 {
		t.Fatalf("cerradas = %v (err %v)", lista, err)
	}
	c := lista[0]
	if c.ID != s.ID || len(c.CadenaCC) != 1 || c.CadenaCC[0] != "cc-vivo" || c.Cwd == "" || c.CerradaEn == "" || c.Turnos != 2 {
		t.Errorf("metadata archivada incompleta: %+v", c)
	}
	if c.Conv != nil {
		t.Error("el Conv NO viaja al archivo (B2: la JSONL nativa es la verdad)")
	}
	if otros, _ := svc.Cerradas(context.Background(), "otro-arnes"); len(otros) != 0 {
		t.Errorf("el filtro por arnés debe excluir: %v", otros)
	}
}

// TestHistorialCerradaCoseCadena (RF-202): la conversación se reconstruye cosiendo las
// JSONL de la cadena; las ausentes van a `faltantes` — jamás se inventa.
func TestHistorialCerradaCoseCadena(t *testing.T) {
	cerradas := &memSessionStore{sesiones: []domain.Session{{
		ID: "s1", Arnes: "vitalia", Cwd: "/proj/vitalia",
		CadenaCC: []string{"cc-a", "cc-borrada", "cc-b"},
	}}}
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetArchivoCerradas(cerradas)
	svc.SetHistoryReader(&fakeHistory{corpus: map[string][]domain.Turn{
		"/proj/vitalia|cc-a": {{Rol: domain.RolUser, Text: "uno"}},
		"/proj/vitalia|cc-b": {{Rol: domain.RolAssistant, Text: "dos"}},
	}})

	turnos, faltantes, err := svc.HistorialCerrada(context.Background(), "s1")
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 2 || turnos[0].Text != "uno" || turnos[1].Text != "dos" {
		t.Errorf("turnos cosidos = %+v", turnos)
	}
	if len(faltantes) != 1 || faltantes[0] != "cc-borrada" {
		t.Errorf("faltantes = %v", faltantes)
	}
	if _, _, err := svc.HistorialCerrada(context.Background(), "no-existe"); err == nil {
		t.Error("cerrada inexistente debe ser not-found")
	}
}
