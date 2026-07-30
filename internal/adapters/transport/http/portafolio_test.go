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
	"github.com/alpacapurpura/arnesia/internal/ports"
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

// fakeIndex satisface ports.IndexPort — el único método que el endpoint de observar
// ejercita es Upsert (S1-D1).
type fakeIndex struct {
	upserted []domain.Graph
	claves   []string
}

func (f *fakeIndex) Rebuild(context.Context) error { return nil }
func (f *fakeIndex) Query(context.Context, string) (domain.Graph, error) {
	return domain.Graph{}, nil
}
func (f *fakeIndex) List(context.Context) ([]ports.EntradaIndice, error) { return nil, nil }
func (f *fakeIndex) Upsert(_ context.Context, clave string, g domain.Graph) error {
	f.upserted = append(f.upserted, g)
	f.claves = append(f.claves, clave)
	return nil
}

func svcConTienda(t *testing.T, root string) (*usecase.PortafolioService, *portafolio.Store, *fakeIndex) {
	t.Helper()
	st, err := portafolio.NewStore(filepath.Join(t.TempDir(), "portafolio.json"))
	if err != nil {
		t.Fatal(err)
	}
	scan := &fakeScan{hallazgos: []domain.HallazgoInstalacion{{Dir: root, Tipo: domain.InstProyectoInstalado}}}
	ldr := &fakeLoad{porDir: map[string]domain.Graph{root: {Arnes: &domain.Arnes{ID: "harness-x", Marketplace: "owner/repo"}}}}
	idx := &fakeIndex{}
	svc := usecase.NewPortafolioService(st, scan, ldr, fakeDeriva{}, idx)
	return svc, st, idx
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
	svc := usecase.NewPortafolioService(st, &fakeScan{}, &fakeLoad{}, fakeDeriva{}, &fakeIndex{})

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
	svc, st, _ := svcConTienda(t, root)

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
	svc, st, _ := svcConTienda(t, root)

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
	svc, _, _ := svcConTienda(t, root)

	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/portafolio/arneses/no-existe", nil)
	r.SetPathValue("clave", "no-existe")
	deleteDesvincular(svc)(w, r)
	if w.Code != 404 {
		t.Fatalf("status %d, quiero 404 para una clave desconocida", w.Code)
	}
}

// TestPortafolioObservar cubre el endpoint POST /api/portafolio/arneses/{clave}/mapa
// (S1-D1): 200 indexa la instalación real vía el fake index, 404 clave desconocida, 400
// install_path ajeno.
func TestPortafolioObservar(t *testing.T) {
	root := t.TempDir()
	svc, _, idx := svcConTienda(t, root)

	cands, err := svc.Escanear(context.Background(), root)
	if err != nil || len(cands) != 1 {
		t.Fatalf("setup: %v %d", err, len(cands))
	}
	clave := cands[0].Clave
	if _, aerr := svc.AgregarProyecto(context.Background(), root, []string{clave}); aerr != nil {
		t.Fatalf("setup agregar: %v", aerr)
	}

	t.Run("200 indexa", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/"+clave+"/mapa",
			strings.NewReader(`{"install_path":"`+root+`"}`))
		r.SetPathValue("clave", clave)
		postObservarEnMapa(svc)(w, r)
		if w.Code != 200 {
			t.Fatalf("status %d: %s", w.Code, w.Body)
		}
		var body map[string]any
		if derr := json.Unmarshal(w.Body.Bytes(), &body); derr != nil {
			t.Fatal(derr)
		}
		// El "id" devuelto ES la clave calificada (deuda BACKLOG «re-key», cerrada
		// 2026-07-23) — nunca el g.Arnes.ID pelado, que otro arnés distinto podría compartir.
		if body["id"] != clave || body["indexed"] != true {
			t.Fatalf("wire infiel: %+v (want id=%q)", body, clave)
		}
		if len(idx.upserted) != 1 {
			t.Fatalf("indice.Upsert llamado %d veces, quiero 1", len(idx.upserted))
		}
		if idx.claves[0] != clave {
			t.Errorf("Upsert indexó bajo %q, want la clave calificada %q", idx.claves[0], clave)
		}
	})

	t.Run("404 clave desconocida", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/no-existe/mapa",
			strings.NewReader(`{"install_path":"`+root+`"}`))
		r.SetPathValue("clave", "no-existe")
		postObservarEnMapa(svc)(w, r)
		if w.Code != 404 {
			t.Fatalf("status %d, quiero 404 para una clave desconocida: %s", w.Code, w.Body)
		}
	})

	t.Run("400 install_path ajeno", func(t *testing.T) {
		ajeno := filepath.Join(t.TempDir(), "ajeno")
		w := httptest.NewRecorder()
		r := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/"+clave+"/mapa",
			strings.NewReader(`{"install_path":"`+ajeno+`"}`))
		r.SetPathValue("clave", clave)
		postObservarEnMapa(svc)(w, r)
		if w.Code != 400 {
			t.Fatalf("status %d, quiero 400 para un install_path ajeno: %s", w.Code, w.Body)
		}
	})
}
