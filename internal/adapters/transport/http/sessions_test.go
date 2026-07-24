package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

type nilAgent struct{}

func (nilAgent) Spawn(context.Context, ports.SpawnOpts) (ports.AgentSession, error) { return nil, nil }

type nilStore struct{}

func (nilStore) Load(context.Context) ([]domain.Session, error) { return nil, nil }
func (nilStore) Save(context.Context, []domain.Session) error   { return nil }

type nilPub struct{}

func (nilPub) Publish(string, []byte) {}

type memRegistry struct{ regs map[string]string }

func (m *memRegistry) Register(id, path string) error {
	m.regs[id] = path
	return nil
}
func (m *memRegistry) List() []ports.ArnesPath              { return nil }
func (m *memRegistry) Resolve(string) (string, bool, error) { return "", false, nil }

// TestCreateSessionIndexaAlRegistrar cubre el gap destapado por el E2E de T4 (paquete
// mejorar-arnes-conversando): crear una sesión con path debe INDEXAR el árbol en la misma
// llamada (onRegistered) — sin eso, roleFor no resuelve rol en el primer spawn y la sesión
// nace sin flags de permisos. Un onRegistered que falla NO rompe el create (degradación
// honesta: carpeta cruda sigue siendo sesión legal).
func TestCreateSessionIndexaAlRegistrar(t *testing.T) {
	svc, err := usecase.NewSessionService(context.Background(), nilAgent{}, nilStore{}, nilPub{}, nil, 1, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	reg := &memRegistry{regs: map[string]string{}}
	var indexado []string
	h := createSession(svc, reg, func(id, path string) error {
		indexado = append(indexado, id+"→"+path)
		return nil
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/sessions",
		strings.NewReader(`{"arnes":"vitalia","frente":"reparación","path":"/tmp/x"}`))
	rec := httptest.NewRecorder()
	h(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if len(indexado) != 1 || indexado[0] != "vitalia→/tmp/x" {
		t.Errorf("onRegistered no se llamó con el árbol registrado: %v", indexado)
	}
	if reg.regs["vitalia"] != "/tmp/x" {
		t.Errorf("registro: %v", reg.regs)
	}
}
