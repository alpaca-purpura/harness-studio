package httpapi

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// fakeScan/fakeLoad/fakeDeriva: mismo patrón de puertos fake que internal/usecase (no se
// comparten entre paquetes de test — cada capa prueba el wire con su propio doble mínimo).
type fakeScan struct {
	hallazgos []domain.HallazgoInstalacion
}

func (f *fakeScan) Escanear(context.Context, string) ([]domain.HallazgoInstalacion, error) {
	return f.hallazgos, nil
}

type fakeLoad struct{ porDir map[string]domain.Graph }

func (f *fakeLoad) Load(dir string) (domain.Graph, error) { return f.porDir[dir], nil }

type fakeDeriva struct{}

func (fakeDeriva) Evaluar(string, string, string, string) (domain.EstadoDeriva, string) {
	return domain.DerivaNoEvaluable, "fake"
}

func svcConTienda(t *testing.T, root string) (*usecase.PortafolioService, *portafolio.Store) {
	t.Helper()
	st, err := portafolio.NewStore(filepath.Join(t.TempDir(), "portafolio.json"))
	if err != nil {
		t.Fatal(err)
	}
	scan := &fakeScan{hallazgos: []domain.HallazgoInstalacion{{Dir: root, Tipo: domain.InstProyectoInstalado}}}
	ldr := &fakeLoad{porDir: map[string]domain.Graph{root: {Arnes: &domain.Arnes{ID: "harness-x", Marketplace: "owner/repo"}}}}
	svc := usecase.NewPortafolioService(st, scan, ldr, fakeDeriva{})
	return svc, st
}

func TestPortafolioListIncluyeCorruptas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	st, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if uerr := st.Upsert(domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "x"}}); uerr != nil {
		t.Fatal(uerr)
	}
	svc := usecase.NewPortafolioService(st, &fakeScan{}, &fakeLoad{}, fakeDeriva{})

	w := httptest.NewRecorder()
	listPortafolio(svc)(w, httptest.NewRequestWithContext(context.Background(), "GET", "/api/portafolio", nil))
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var body portafolioListResponse
	if derr := json.Unmarshal(w.Body.Bytes(), &body); derr != nil {
		t.Fatal(derr)
	}
	if len(body.Entradas) != 1 || body.Entradas[0].Clave == "" {
		t.Fatalf("wire infiel: %+v", body)
	}
}

func TestPortafolioEscanearNoPersiste(t *testing.T) {
	root := t.TempDir()
	svc, st := svcConTienda(t, root)

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/escaneos",
		strings.NewReader(`{"path":"`+root+`"}`))
	postEscanear(svc)(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var cands []usecase.Candidato
	if err := json.Unmarshal(w.Body.Bytes(), &cands); err != nil {
		t.Fatal(err)
	}
	if len(cands) != 1 || cands[0].Clave == "" {
		t.Fatalf("wire infiel: %+v", cands)
	}
	sanas, _ := st.Listar()
	if len(sanas) != 0 {
		t.Errorf("escanear NO debe persistir nada, el store tiene %d entradas", len(sanas))
	}
}

func TestPortafolioAgregar(t *testing.T) {
	root := t.TempDir()
	svc, st := svcConTienda(t, root)

	// primero escanear para conocer la clave (igual que un cliente real).
	cands, err := svc.Escanear(context.Background(), root)
	if err != nil || len(cands) != 1 {
		t.Fatalf("setup: %v %d", err, len(cands))
	}
	clave := cands[0].Clave

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/proyectos",
		strings.NewReader(`{"path":"`+root+`","elegidos":["`+clave+`"]}`))
	postAgregar(svc)(w, r)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	sanas, _ := st.Listar()
	if len(sanas) != 1 {
		t.Fatalf("tras agregar, store = %d entradas, quiero 1", len(sanas))
	}
}

func TestPortafolioDesvincular404(t *testing.T) {
	root := t.TempDir()
	svc, _ := svcConTienda(t, root)

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/portafolio/arneses/no-existe", nil)
	r.SetPathValue("clave", "no-existe")
	deleteDesvincular(svc)(w, r)
	if w.Code != 404 {
		t.Fatalf("status %d, quiero 404 para una clave desconocida", w.Code)
	}
}
