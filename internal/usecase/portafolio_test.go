package usecase_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// ── fakes de los 4 puertos del Portafolio ──

type fakePortafolioScanner struct {
	hallazgos []domain.HallazgoInstalacion
	err       error
}

func (f *fakePortafolioScanner) Escanear(context.Context, string) ([]domain.HallazgoInstalacion, error) {
	return f.hallazgos, f.err
}

type fakePortafolioLoader struct {
	porDir map[string]domain.Graph
}

func (f *fakePortafolioLoader) Load(dir string) (domain.Graph, error) {
	g, ok := f.porDir[dir]
	if !ok {
		return domain.Graph{}, fmt.Errorf("fakePortafolioLoader: sin fixture para %s", dir)
	}
	return g, nil
}

type fakeDerivaEvaluator struct{}

func (fakeDerivaEvaluator) Evaluar(string, string, string, string) (domain.EstadoDeriva, string) {
	return domain.DerivaNoEvaluable, "fake: no evaluado"
}

type fakePortafolioStore struct {
	entradas  map[string]domain.EntradaPortafolio
	checkouts []string
}

func newFakePortafolioStore() *fakePortafolioStore {
	return &fakePortafolioStore{entradas: map[string]domain.EntradaPortafolio{}}
}

func (f *fakePortafolioStore) Listar() ([]domain.EntradaPortafolio, []domain.EntradaCorrupta) {
	out := make([]domain.EntradaPortafolio, 0, len(f.entradas))
	for _, e := range f.entradas {
		out = append(out, e)
	}
	return out, nil
}

func (f *fakePortafolioStore) Upsert(e domain.EntradaPortafolio) error {
	f.entradas[e.Identidad.Clave()] = e
	return nil
}

func (f *fakePortafolioStore) Desvincular(clave string) (bool, error) {
	if _, ok := f.entradas[clave]; !ok {
		return false, nil
	}
	delete(f.entradas, clave)
	return true, nil
}

func (f *fakePortafolioStore) Checkouts() []string { return f.checkouts }

// TestServiceEscanearClasificaCheckout cubre RN-IDENT-4/C-N-12: un hallazgo cuyo Dir ES
// un checkout conocido del store se clasifica CANÓNICO, jamás instalación (cierra el
// doble-conteo/deriva auto-referencial del dogfood).
func TestServiceEscanearClasificaCheckout(t *testing.T) {
	root := t.TempDir()
	checkoutDir := filepath.Join(root, "checkout-conocido")

	store := newFakePortafolioStore()
	store.checkouts = []string{checkoutDir}
	scan := &fakePortafolioScanner{hallazgos: []domain.HallazgoInstalacion{
		{Dir: checkoutDir, Tipo: domain.InstProyectoInstalado},
	}}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		checkoutDir: {Arnes: &domain.Arnes{ID: "x", Marketplace: "owner/repo"}},
	}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{})

	cands, err := svc.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 1 {
		t.Fatalf("candidatos = %d, quiero 1", len(cands))
	}
	if !cands[0].EsCanonico {
		t.Error("un dir ∈ Checkouts() debe clasificarse CANÓNICO, jamás instalación (RN-IDENT-4)")
	}
}

// TestServiceAgregarSoloElegidos: de N candidatos, AgregarProyecto persiste SOLO los
// cuya Identidad.Clave() está en elegidos.
func TestServiceAgregarSoloElegidos(t *testing.T) {
	root := t.TempDir()
	dirA := filepath.Join(root, "a")
	dirB := filepath.Join(root, "b")

	store := newFakePortafolioStore()
	scan := &fakePortafolioScanner{hallazgos: []domain.HallazgoInstalacion{
		{Dir: dirA, Tipo: domain.InstProyectoInstalado},
		{Dir: dirB, Tipo: domain.InstProyectoInstalado},
	}}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		dirA: {Arnes: &domain.Arnes{ID: "harness-a", Marketplace: "owner/repo-a"}},
		dirB: {Arnes: &domain.Arnes{ID: "harness-b", Marketplace: "owner/repo-b"}},
	}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{})

	cands, err := svc.Escanear(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cands) != 2 {
		t.Fatalf("candidatos = %d, quiero 2", len(cands))
	}
	elegido := cands[0].Identidad.Clave()

	persistidas, err := svc.AgregarProyecto(context.Background(), root, []string{elegido})
	if err != nil {
		t.Fatal(err)
	}
	if len(persistidas) != 1 {
		t.Fatalf("persistidas = %d, quiero 1 (solo el elegido)", len(persistidas))
	}
	sanas, _, _ := svc.Listar(context.Background())
	if len(sanas) != 1 {
		t.Fatalf("store final = %d entradas, quiero 1", len(sanas))
	}
}

// TestServiceDesvincularNoTocaDisco cubre C-UNL-3: desvincular solo saca del registro, los
// archivos del proyecto fixture siguen intactos.
func TestServiceDesvincularNoTocaDisco(t *testing.T) {
	root := t.TempDir()
	archivoProyecto := filepath.Join(root, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(archivoProyecto), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archivoProyecto, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}

	store := newFakePortafolioStore()
	scan := &fakePortafolioScanner{hallazgos: []domain.HallazgoInstalacion{
		{Dir: root, Tipo: domain.InstProyectoInstalado},
	}}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		root: {Arnes: &domain.Arnes{ID: "harness-x", Marketplace: "owner/repo"}},
	}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{})

	cands, err := svc.Escanear(context.Background(), root)
	if err != nil || len(cands) != 1 {
		t.Fatalf("setup: %v %d", err, len(cands))
	}
	clave := cands[0].Identidad.Clave()
	if _, aerr := svc.AgregarProyecto(context.Background(), root, []string{clave}); aerr != nil {
		t.Fatal(aerr)
	}

	ok, err := svc.Desvincular(context.Background(), clave)
	if err != nil || !ok {
		t.Fatalf("Desvincular: ok=%v err=%v", ok, err)
	}
	sanas, _, _ := svc.Listar(context.Background())
	if len(sanas) != 0 {
		t.Errorf("tras desvincular, store debe quedar vacío, got %d", len(sanas))
	}
	if _, serr := os.Stat(archivoProyecto); serr != nil {
		t.Errorf("el archivo del proyecto NO debe tocarse al desvincular (C-UNL-3): %v", serr)
	}
}

// TestServiceRootProtegido cubre la reutilización de la política de arnes_registry: un
// root que ES $HOME (o se superpone con una ubicación protegida) se rechaza.
func TestServiceRootProtegido(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store := newFakePortafolioStore()
	scan := &fakePortafolioScanner{}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{})

	if _, err := svc.Escanear(context.Background(), home); err == nil {
		t.Error("root == $HOME debe rechazarse")
	}
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Escanear(context.Background(), sshDir); err == nil {
		t.Error("root dentro de una ubicación protegida debe rechazarse")
	}
}
