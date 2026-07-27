package usecase_test

import (
	"context"
	"fmt"
	"strings"
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

// TestCloseArchivaConTranscript — el MISMO test que antes se llamaba
// `TestCloseArchivaMetadata` y afirmaba lo CONTRARIO: que el transcript NO viajaba al
// archivo. Se renombra e invierte en vez de borrarse, para que el diff muestre que lo que
// cambió es una LEY y no una implementación (CV-D8 + enmienda F-3, ledger HS-29).
//
// La ley vieja decía «sin Conv: la JSONL nativa es la verdad». La JSONL sigue siendo la
// verdad del CONTENIDO — pero `Conv` es la copia de presentación, la única que sobrevive a
// un GC del corpus, la única sobre la que se puede buscar, y la que se repinta al retomar.
// Archivar sin ella era perder lo que el operador vio.
func TestCloseArchivaConTranscript(t *testing.T) {
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
	if turnErr := svc.Turn(s.ID, "hola"); turnErr != nil {
		t.Fatal(turnErr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-vivo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "listo", CtxPct: 5}
	espera(t, func() bool {
		m, _ := svc.Get(s.ID)
		return m.Status == domain.StatusIdle && activaSinFallar(m).ClaudeSessionID == "cc-vivo"
	})

	if closeErr := svc.Close(s.ID); closeErr != nil {
		t.Fatal(closeErr)
	}
	lista, err := svc.Cerradas(context.Background(), "vitalia")
	if err != nil || len(lista) != 1 {
		t.Fatalf("cerradas = %v (err %v)", lista, err)
	}
	c := lista[0]
	if c.ID != s.ID || c.Cwd == "" || c.CerradaEn == "" {
		t.Errorf("metadata archivada incompleta: %+v", c)
	}
	// La cadena y el transcript son de la CONVERSACIÓN, no de la sesión (CV-D3).
	conv := activaDe(t, c)
	if len(conv.CadenaCC) != 1 || conv.CadenaCC[0] != "cc-vivo" {
		t.Errorf("cadena archivada = %v, quiero [cc-vivo]", conv.CadenaCC)
	}
	// ── LA ASERCIÓN INVERTIDA ──
	if conv.Conv == nil {
		t.Fatal("el transcript NO viajó al archivo: es lo que el operador vio, y es lo único que sobrevive a un GC del corpus")
	}
	if conv.NumTurnos() != 2 {
		t.Errorf("turnos archivados = %d, quiero 2 (el user y el assistant)", conv.NumTurnos())
	}
	if conv.Conv[0].Rol != domain.RolUser || conv.Conv[0].Text != "hola" {
		t.Errorf("el primer turno archivado no es el del operador: %+v", conv.Conv[0])
	}
	if otros, _ := svc.Cerradas(context.Background(), "otro-arnes"); len(otros) != 0 {
		t.Errorf("el filtro por arnés debe excluir: %v", otros)
	}
}

// TestCloseConservaCheckpoint (enmienda F-3) — la mitad que NINGUNA decisión previa cubría.
// CV-D11 promete que retomar una conversación la devuelve con su checkpoint intacto, pero
// el archivado lo borraba junto con el transcript. Sin él, retomar una conversación que ya
// había rotado arranca sin el digest de su propia rotación — exactamente el estado que el
// checkpoint existe para no perder.
func TestCloseConservaCheckpoint(t *testing.T) {
	agent := &stubAgent{}
	cerradas := &memSessionStore{}
	svc, err := usecase.NewSessionService(context.Background(), agent, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetArchivoCerradas(cerradas)
	svc.SetUmbralRotacion(40)

	s, err := svc.Create(domain.Session{Arnes: "vitalia", Frente: "reparar"})
	if err != nil {
		t.Fatal(err)
	}
	// Turno 1 cruza el umbral; turno 2 rota y escribe el checkpoint.
	if turnErr := svc.Turn(s.ID, "el pedido que quedó a medias"); turnErr != nil {
		t.Fatal(turnErr)
	}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventInit, ClaudeSessionID: "cc-viejo"}
	agent.sessions[0].events <- ports.AgentEvent{Kind: ports.EventResult, Text: "hecho", CtxPct: 45}
	espera(t, func() bool { m, _ := svc.Get(s.ID); return activaSinFallar(m).RotacionPendiente })
	if turnErr := svc.Turn(s.ID, "seguimos"); turnErr != nil {
		t.Fatal(turnErr)
	}
	vivo, _ := svc.Get(s.ID)
	if activaDe(t, vivo).Checkpoint == "" {
		t.Fatal("la rotación no dejó checkpoint: el test no está midiendo lo que dice")
	}

	if closeErr := svc.Close(s.ID); closeErr != nil {
		t.Fatal(closeErr)
	}
	lista, err := svc.Cerradas(context.Background(), "vitalia")
	if err != nil || len(lista) != 1 {
		t.Fatalf("cerradas = %v (err %v)", lista, err)
	}
	conv := activaDe(t, lista[0])
	if conv.Checkpoint == "" {
		t.Fatal("el checkpoint NO sobrevivió al archivado: retomar arrancaría sin el digest de su propia rotación")
	}
	if !strings.Contains(conv.Checkpoint, "el pedido que quedó a medias") {
		t.Errorf("el checkpoint archivado no trae los últimos turnos: %.120q", conv.Checkpoint)
	}
	// Y la cadena de rotaciones también: el join hacia las JSONL nativas sigue entero.
	if len(conv.CadenaCC) != 1 || conv.CadenaCC[0] != "cc-viejo" {
		t.Errorf("cadena archivada = %v, quiero [cc-viejo]", conv.CadenaCC)
	}
}

// TestHistorialArchivadoUsaElTranscriptPropio — desde que el transcript se archiva, leer el
// historial de una sesión archivada NO depende de que el corpus de Claude Code siga en
// disco. El lector JSONL queda de fallback para las archivadas antes de la migración.
func TestHistorialArchivadoUsaElTranscriptPropio(t *testing.T) {
	cerradas := &memSessionStore{sesiones: []domain.Session{{
		ID: "s-nueva", Arnes: "vitalia", Cwd: "/proj/vitalia",
		Conversaciones: []domain.Conversacion{{
			ID: "cv1", Titulo: "archivada con transcript", Activa: true,
			CadenaCC: []string{"cc-a"},
			Conv:     []domain.Turn{{Rol: domain.RolUser, Text: "lo que el operador vio"}},
		}},
	}}}
	svc, err := usecase.NewSessionService(context.Background(), &stubAgent{}, stubStore{}, stubPub{}, stubResolver{path: t.TempDir()}, 40, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	svc.SetArchivoCerradas(cerradas)
	// El corpus JSONL ya no existe: si el historial dependiera de él, esto devolvería vacío.
	svc.SetHistoryReader(&fakeHistory{corpus: map[string][]domain.Turn{}})

	turnos, faltantes, err := svc.HistorialCerrada(context.Background(), "s-nueva")
	if err != nil {
		t.Fatal(err)
	}
	if len(turnos) != 1 || turnos[0].Text != "lo que el operador vio" {
		t.Errorf("turnos = %+v, quiero el transcript archivado", turnos)
	}
	if len(faltantes) != 0 {
		t.Errorf("faltantes = %v — no falta nada: el transcript estaba guardado", faltantes)
	}
}

// TestHistorialCerradaCoseCadena (RF-202): el FALLBACK — una sesión archivada ANTES de la
// migración no tiene transcript propio, así que se reconstruye cosiendo las JSONL de la
// cadena; las ausentes van a `faltantes` y jamás se inventa. Se conserva a propósito: el
// registro archivado del operador tiene entradas de esa época.
func TestHistorialCerradaCoseCadena(t *testing.T) {
	cerradas := &memSessionStore{sesiones: []domain.Session{{
		ID: "s1", Arnes: "vitalia", Cwd: "/proj/vitalia",
		Conversaciones: []domain.Conversacion{{
			ID: "cv0000000a", Titulo: "hilo archivado", Activa: true,
			CadenaCC: []string{"cc-a", "cc-borrada", "cc-b"},
		}},
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
