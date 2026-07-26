package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
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
	mu        sync.Mutex
	entradas  map[string]domain.EntradaPortafolio
	checkouts []string
	// upsertErr fuerza el fallo del paso 9 de Traer (E-100): el ÚNICO parcial posible.
	upsertErr error
}

func newFakePortafolioStore() *fakePortafolioStore {
	return &fakePortafolioStore{entradas: map[string]domain.EntradaPortafolio{}}
}

func (f *fakePortafolioStore) Listar() ([]domain.EntradaPortafolio, []domain.EntradaCorrupta) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.EntradaPortafolio, 0, len(f.entradas))
	for _, e := range f.entradas {
		out = append(out, e)
	}
	return out, nil
}

func (f *fakePortafolioStore) Upsert(e domain.EntradaPortafolio) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.entradas[e.Identidad.Clave()] = e
	return nil
}

func (f *fakePortafolioStore) Desvincular(clave string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.entradas[clave]; !ok {
		return false, nil
	}
	delete(f.entradas, clave)
	return true, nil
}

func (f *fakePortafolioStore) Checkouts() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.checkouts
}

// fakeIndexPort satisface ports.IndexPort para testear ObservarEnMapa (S1-D1, Slice 1): el
// único método que el usecase ejercita es Upsert — Rebuild/Query/List son stubs no
// llamados desde este service (el 5° puerto es de solo-escritura acá).
type fakeIndexPort struct {
	upserted []domain.Graph
	claves   []string
	err      error
}

func (f *fakeIndexPort) Rebuild(context.Context) error { return nil }
func (f *fakeIndexPort) Query(context.Context, string) (domain.Graph, error) {
	return domain.Graph{}, nil
}
func (f *fakeIndexPort) List(context.Context) ([]domain.Graph, error) { return nil, nil }
func (f *fakeIndexPort) Upsert(_ context.Context, clave string, g domain.Graph) error {
	f.upserted = append(f.upserted, g)
	f.claves = append(f.claves, clave)
	return f.err
}

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
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, nil)

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
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, nil)

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
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, nil)

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
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, nil)

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

// ── ObservarEnMapa (Slice 1, S1-D1 — cierra GAP-1) ──

// TestObservarEnMapaIndexaSinRegistro cubre el camino feliz: una entrada YA PERSISTIDA con
// una instalación cuyo install_path coincide se publica al índice del Mapa — el fake index
// recibe exactamente el grafo cargado por el loader, indexado bajo la CLAVE calificada
// (deuda BACKLOG «re-key (home,id,scope)», cerrada 2026-07-23 — antes era el bare
// g.Arnes.ID, colisionable). "cero interacción con ArnesRegistry" es estructural:
// PortafolioService no recibe ese puerto en su constructor en absoluto — observar NO
// registra cwd.
func TestObservarEnMapaIndexaSinRegistro(t *testing.T) {
	store := newFakePortafolioStore()
	installPath := filepath.Join(t.TempDir(), "instalacion")
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Home: "owner/repo", ID: "harness-x"},
		Instalaciones: []domain.Instalacion{{InstallPath: installPath, Tipo: domain.InstReferenciadaCC}},
	}
	if err := store.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	clave := entrada.Identidad.Clave()

	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		installPath: {Arnes: &domain.Arnes{ID: "harness-x", Marketplace: "owner/repo"}},
	}}
	idx := &fakeIndexPort{}
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{}, ldr, fakeDerivaEvaluator{}, idx)

	id, err := svc.ObservarEnMapa(context.Background(), clave, installPath)
	if err != nil {
		t.Fatalf("ObservarEnMapa: %v", err)
	}
	if id != clave {
		t.Errorf("id = %q, quiero la clave calificada %q", id, clave)
	}
	if len(idx.upserted) != 1 {
		t.Fatalf("indice.Upsert llamado %d veces, quiero 1", len(idx.upserted))
	}
	if idx.claves[0] != clave {
		t.Errorf("Upsert indexó bajo %q, want la clave calificada %q", idx.claves[0], clave)
	}
	if idx.upserted[0].Arnes == nil || idx.upserted[0].Arnes.ID != "harness-x" {
		t.Errorf("grafo indexado = %+v, no es el que cargó el loader", idx.upserted[0])
	}
}

// TestObservarEnMapaDegradadoSintetiza cubre S1-D27 (Opción A, honra contrato §2): una
// presencia SIN sello (loader devuelve Arnes==nil) NO se rechaza — se sintetiza un arnés
// mínimo con la huella de path como llave sintética estable, se marca Degradado, y los nodos
// que el loader reconoció se preservan. Antes esto era un 400 «no resolvió un arnés cargable».
func TestObservarEnMapaDegradadoSintetiza(t *testing.T) {
	store := newFakePortafolioStore()
	installPath := filepath.Join(t.TempDir(), "cruda")
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Scope: ".", Disc: "deadbeef1234"}, // anónima persistida
		Instalaciones: []domain.Instalacion{{InstallPath: installPath, Tipo: domain.InstProyectoInstalado}},
	}
	if err := store.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	clave := entrada.Identidad.Clave()

	// Loader degradado: sin sello (Arnes==nil) pero con un nodo reconocido.
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		installPath: {Arnes: nil, Degradado: true, Nodes: []domain.Box{{ID: "std", Clase: domain.ClaseRule}}},
	}}
	idx := &fakeIndexPort{}
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{}, ldr, fakeDerivaEvaluator{}, idx)

	id, err := svc.ObservarEnMapa(context.Background(), clave, installPath)
	if err != nil {
		t.Fatalf("ObservarEnMapa degradado NO debe fallar (S1-D27): %v", err)
	}
	// El id devuelto (y la llave del índice) es la CLAVE calificada de la entrada persistida
	// (deuda BACKLOG «re-key», cerrada 2026-07-23) — no la huella de path recién sintetizada
	// para el `g.Arnes.ID` interno del grafo (esa sigue siendo un fallback de DISPLAY, ya no
	// la llave del índice).
	if id != clave {
		t.Errorf("id = %q, quiero la clave calificada %q", id, clave)
	}
	if len(idx.upserted) != 1 {
		t.Fatalf("indice.Upsert llamado %d veces, quiero 1", len(idx.upserted))
	}
	if idx.claves[0] != clave {
		t.Errorf("Upsert indexó bajo %q, want la clave calificada %q", idx.claves[0], clave)
	}
	g := idx.upserted[0]
	if g.Arnes == nil || g.Arnes.ID == "" {
		t.Errorf("grafo indexado sin arnés sintético estable: %+v", g.Arnes)
	}
	if !g.Degradado {
		t.Error("el grafo indexado debe seguir marcado Degradado")
	}
	if g.Arnes.Nombre != "cruda" {
		t.Errorf("Nombre = %q, quiero el basename del install_path como fallback", g.Arnes.Nombre)
	}
	if len(g.Nodes) != 1 {
		t.Errorf("los nodos reconocidos por el loader deben preservarse, got %d", len(g.Nodes))
	}
}

// TestIdentificarSellaYRekey cubre S1-D28: Identificar escribe el sello `arnes.l0.json`
// IN-SITU (scaffold mínimo) y re-keya la entrada anónima con su identidad ya sellada,
// desvinculando la clave vieja. El fake loader simula que, tras sellar, el loader lee el
// manifiesto y resuelve id — la escritura del sello, las guardas y el re-key son reales.
func TestIdentificarSellaYRekey(t *testing.T) {
	dir := t.TempDir()

	store := newFakePortafolioStore()
	anon := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Scope: ".", Disc: domain.HuellaPath(dir)},
		Instalaciones: []domain.Instalacion{{ProyectoPath: dir, InstallPath: dir, Tipo: domain.InstProyectoInstalado}},
	}
	if err := store.Upsert(anon); err != nil {
		t.Fatal(err)
	}
	claveVieja := anon.Identidad.Clave()

	scan := &fakePortafolioScanner{hallazgos: []domain.HallazgoInstalacion{
		{Dir: dir, Tipo: domain.InstProyectoInstalado},
	}}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		dir: {Arnes: &domain.Arnes{ID: "mi-arnes", Nombre: "Mi Arnés"}},
	}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, &fakeIndexPort{})

	nueva, err := svc.Identificar(context.Background(), claveVieja, dir, "mi-arnes", "Mi Arnés")
	if err != nil {
		t.Fatalf("Identificar: %v", err)
	}

	// 1. El sello se escribió con el scaffold mínimo (id · version · reporta_a null).
	b, rerr := os.ReadFile(filepath.Join(dir, "arnes.l0.json")) //nolint:gosec // G304: ruta de fixture del test.
	if rerr != nil {
		t.Fatalf("el sello no se escribió: %v", rerr)
	}
	var sello map[string]any
	if jerr := json.Unmarshal(b, &sello); jerr != nil {
		t.Fatalf("sello inválido: %v", jerr)
	}
	if sello["id"] != "mi-arnes" || sello["version"] != "0.1.0" {
		t.Errorf("sello = %v, quiero id=mi-arnes version=0.1.0", sello)
	}
	if v, ok := sello["reporta_a"]; !ok || v != nil {
		t.Errorf("el sello debe emitir reporta_a: null, got %v (ok=%v)", v, ok)
	}

	// 2. Re-key: la entrada nueva tiene la identidad sellada; la vieja anónima se desvinculó.
	if nueva.Identidad.ID != "mi-arnes" {
		t.Errorf("identidad re-keyed = %+v", nueva.Identidad)
	}
	if _, ok := store.entradas[claveVieja]; ok {
		t.Error("la clave vieja anónima debe desvincularse tras el re-key")
	}
	if _, ok := store.entradas[nueva.Identidad.Clave()]; !ok {
		t.Error("la entrada sellada debe quedar persistida bajo su clave nueva")
	}
}

// TestIdentificarNoPisaSelloExistente cubre la guarda 2 de S1-D28: un dir que YA tiene
// arnes.l0.json no se re-sella (ErrIdentificarYaSellado) — editar un sello es otra operación.
func TestIdentificarNoPisaSelloExistente(t *testing.T) {
	dir := t.TempDir()
	if werr := os.WriteFile(filepath.Join(dir, "arnes.l0.json"), []byte(`{"id":"ya"}`), 0o600); werr != nil {
		t.Fatal(werr)
	}
	store := newFakePortafolioStore()
	e := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Scope: ".", Disc: domain.HuellaPath(dir)},
		Instalaciones: []domain.Instalacion{{ProyectoPath: dir, InstallPath: dir}},
	}
	if err := store.Upsert(e); err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{}, &fakePortafolioLoader{porDir: map[string]domain.Graph{}}, fakeDerivaEvaluator{}, &fakeIndexPort{})

	_, err := svc.Identificar(context.Background(), e.Identidad.Clave(), dir, "", "")
	if !errors.Is(err, usecase.ErrIdentificarYaSellado) {
		t.Fatalf("err = %v, quiero ErrIdentificarYaSellado", err)
	}
}

// TestObservarEnMapaInstallPathAjeno: un install_path que no es ni una instalación ni el
// canónico de la entrada es ajeno — 400-style, jamás carga un directorio arbitrario.
func TestObservarEnMapaInstallPathAjeno(t *testing.T) {
	store := newFakePortafolioStore()
	instReal := filepath.Join(t.TempDir(), "real")
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Home: "owner/repo", ID: "harness-y"},
		Instalaciones: []domain.Instalacion{{InstallPath: instReal}},
	}
	if err := store.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	clave := entrada.Identidad.Clave()

	idx := &fakeIndexPort{}
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{},
		&fakePortafolioLoader{porDir: map[string]domain.Graph{}}, fakeDerivaEvaluator{}, idx)

	ajeno := filepath.Join(t.TempDir(), "ajeno")
	if _, err := svc.ObservarEnMapa(context.Background(), clave, ajeno); !errors.Is(err, usecase.ErrObservarInstallPathAjeno) {
		t.Fatalf("err = %v, quiero ErrObservarInstallPathAjeno", err)
	}
	if len(idx.upserted) != 0 {
		t.Error("un install_path ajeno NO debe indexar nada")
	}
}

// TestObservarEnMapaClaveInexistente: solo se observa lo YA PERSISTIDO — una clave que no
// está en el store es 404-style, nunca un candidato de escaneo.
func TestObservarEnMapaClaveInexistente(t *testing.T) {
	store := newFakePortafolioStore()
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{},
		&fakePortafolioLoader{porDir: map[string]domain.Graph{}}, fakeDerivaEvaluator{}, &fakeIndexPort{})

	if _, err := svc.ObservarEnMapa(context.Background(), "no-existe", "/cualquier/path"); !errors.Is(err, usecase.ErrObservarClaveNoEncontrada) {
		t.Fatalf("err = %v, quiero ErrObservarClaveNoEncontrada", err)
	}
}

// TestObservarEnMapaSinIndice: el subcomando CLI cablea el 5° puerto en nil — el error es
// honesto («requiere el daemon»), jamás un nil-pointer panic.
func TestObservarEnMapaSinIndice(t *testing.T) {
	store := newFakePortafolioStore()
	installPath := filepath.Join(t.TempDir(), "instalacion")
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Home: "owner/repo", ID: "harness-z"},
		Instalaciones: []domain.Instalacion{{InstallPath: installPath}},
	}
	if err := store.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	clave := entrada.Identidad.Clave()

	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{},
		&fakePortafolioLoader{porDir: map[string]domain.Graph{}}, fakeDerivaEvaluator{}, nil)

	if _, err := svc.ObservarEnMapa(context.Background(), clave, installPath); !errors.Is(err, usecase.ErrObservarSinIndice) {
		t.Fatalf("err = %v, quiero ErrObservarSinIndice", err)
	}
}

// TestObservarEnMapaNoCargable: install_path pertenece a la entrada pero el loader falla —
// el motivo real del loader viaja en el error, jamás se indexa un grafo inventado.
func TestObservarEnMapaNoCargable(t *testing.T) {
	store := newFakePortafolioStore()
	installPath := filepath.Join(t.TempDir(), "instalacion")
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Home: "owner/repo", ID: "harness-w"},
		Instalaciones: []domain.Instalacion{{InstallPath: installPath}},
	}
	if err := store.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	clave := entrada.Identidad.Clave()

	idx := &fakeIndexPort{}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{}} // sin fixture: Load falla.
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{}, ldr, fakeDerivaEvaluator{}, idx)

	_, err := svc.ObservarEnMapa(context.Background(), clave, installPath)
	if err == nil {
		t.Fatal("un dir no cargable debe fallar, jamás indexar un grafo inventado")
	}
	if errors.Is(err, usecase.ErrObservarClaveNoEncontrada) || errors.Is(err, usecase.ErrObservarInstallPathAjeno) || errors.Is(err, usecase.ErrObservarSinIndice) {
		t.Fatalf("err = %v: debe ser el motivo del loader, no uno de los otros 3 casos", err)
	}
	if len(idx.upserted) != 0 {
		t.Error("un loader que falla NO debe llegar a indexar nada")
	}
}

// ── AgregarProyecto: Registries (Slice 1, S1-D3 — cierra GAP-3) ──

// TestAgregarProyectoPueblaRegistries: el registry de origen resuelto se puebla como facet
// — canonicalizado cuando CanonicalizarRepo lo reconoce, crudo VISIBLE cuando no parsea
// (el dato no se descarta por no parsear).
func TestAgregarProyectoPueblaRegistries(t *testing.T) {
	root := t.TempDir()
	dirCanon := filepath.Join(root, "canon")
	dirCrudo := filepath.Join(root, "crudo")

	store := newFakePortafolioStore()
	scan := &fakePortafolioScanner{hallazgos: []domain.HallazgoInstalacion{
		{Dir: dirCanon, Tipo: domain.InstProyectoInstalado, Eslabones: []domain.EslabonOrigen{
			{Fuente: "lock-devstudio", Campo: "registry", Valor: "owner/repo"},
		}},
		{Dir: dirCrudo, Tipo: domain.InstProyectoInstalado, Eslabones: []domain.EslabonOrigen{
			{Fuente: "lock-devstudio", Campo: "registry", Valor: "esto no es un repo"},
		}},
	}}
	ldr := &fakePortafolioLoader{porDir: map[string]domain.Graph{
		dirCanon: {Arnes: &domain.Arnes{ID: "harness-canon"}},
		dirCrudo: {Arnes: &domain.Arnes{ID: "harness-crudo"}},
	}}
	svc := usecase.NewPortafolioService(store, scan, ldr, fakeDerivaEvaluator{}, nil)

	cands, err := svc.Escanear(context.Background(), root)
	if err != nil || len(cands) != 2 {
		t.Fatalf("setup: %v %d", err, len(cands))
	}
	elegidos := []string{cands[0].Identidad.Clave(), cands[1].Identidad.Clave()}
	persistidas, err := svc.AgregarProyecto(context.Background(), root, elegidos)
	if err != nil {
		t.Fatal(err)
	}
	if len(persistidas) != 2 {
		t.Fatalf("persistidas = %d, quiero 2", len(persistidas))
	}

	sanas, _, _ := svc.Listar(context.Background())
	var canon, crudo domain.EntradaPortafolio
	for _, e := range sanas {
		switch {
		case len(e.Instalaciones) == 1 && e.Instalaciones[0].InstallPath == dirCanon:
			canon = e
		case len(e.Instalaciones) == 1 && e.Instalaciones[0].InstallPath == dirCrudo:
			crudo = e
		}
	}
	if len(canon.Registries) != 1 || canon.Registries[0] != "github.com/owner/repo" {
		t.Errorf("registries canónico = %v, quiero [github.com/owner/repo]", canon.Registries)
	}
	if len(crudo.Registries) != 1 || crudo.Registries[0] != "esto no es un repo" {
		t.Errorf("registries crudo = %v, quiero el valor crudo visible (no parsea a host/owner/repo)", crudo.Registries)
	}
}

// derivaFija — evaluador configurable para probar re-evaluación (RF-193).
type derivaFija struct {
	estado  domain.EstadoDeriva
	detalle string
}

func (d derivaFija) Evaluar(string, string, string, string) (domain.EstadoDeriva, string) {
	return d.estado, d.detalle
}

// TestReevaluarDerivaTrasEdicion (RF-193): tras una edición por chat, la deriva de la
// instalación se re-evalúa y persiste si cambió; el Portafolio nunca finge al-hilo.
func TestReevaluarDerivaTrasEdicion(t *testing.T) {
	store := newFakePortafolioStore()
	if err := store.Upsert(domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{ID: "vitalia", Scope: "."},
		Instalaciones: []domain.Instalacion{{
			ProyectoPath: "/proj", InstallPath: "/proj/vitalia",
			Deriva: domain.DerivaNoEvaluable, DerivaDetalle: "sin version",
		}},
	}); err != nil {
		t.Fatal(err)
	}
	svc := usecase.NewPortafolioService(store, &fakePortafolioScanner{}, &fakePortafolioLoader{}, derivaFija{domain.DerivaEnDeriva, "hash difiere"}, nil)

	cambio, err := svc.ReevaluarDeriva(context.Background(), "/proj/vitalia")
	if err != nil || !cambio {
		t.Fatalf("cambio=%v err=%v", cambio, err)
	}
	entradas, _ := store.Listar()
	if got := entradas[0].Instalaciones[0].Deriva; got != domain.DerivaEnDeriva {
		t.Errorf("deriva persistida = %q, quiero en-deriva", got)
	}

	// Segunda pasada sin cambio → false, sin re-persistir.
	if cambio, err = svc.ReevaluarDeriva(context.Background(), "/proj/vitalia"); err != nil || cambio {
		t.Errorf("sin cambio: cambio=%v err=%v", cambio, err)
	}
	// Path fuera del Portafolio → no-op honesto.
	if cambio, err = svc.ReevaluarDeriva(context.Background(), "/otro/lado"); err != nil || cambio {
		t.Errorf("fuera del portafolio: cambio=%v err=%v", cambio, err)
	}
}

// ── S7 · AsignarOrigen (paquete 2026-07-23-portafolio-agregar-marketplace) ──

// derivaConHome finge un evaluador cuya referencia SOLO resuelve cuando el home está declarado:
// es exactamente el efecto útil de reconciliar (E-70).
type derivaConHome struct{ visto []string }

func (d *derivaConHome) Evaluar(installDir, home, _, version string) (domain.EstadoDeriva, string) {
	d.visto = append(d.visto, home)
	if home == "" || version == "" {
		return domain.DerivaNoEvaluable, "sin referencia local accesible (marketplace no clonado, o esa versión ausente del checkout)"
	}
	return domain.DerivaAlHilo, ""
}

// svcConStore cablea un PortafolioService mínimo (sin scanner ni loader: AsignarOrigen no los usa).
func svcConStore(st *fakePortafolioStore, deriva ports.DerivaEvaluator) *usecase.PortafolioService {
	return usecase.NewPortafolioService(st, &fakePortafolioScanner{}, &fakePortafolioLoader{}, deriva, nil)
}

const homeVitalia = "github.com/vitalia/arneses"

func entradaProvisional(id string, insts ...domain.Instalacion) domain.EntradaPortafolio {
	return domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{ID: id},
		Nombre:        id,
		Instalaciones: insts,
	}
}

// E-25 · AsignarOrigen re-keyea por el home declarado, anota el eslabón en CADA instalación y NO
// escribe ningún archivo fuera del store (el dir de la instalación está en modo 0555).
func TestAsignarOrigenReKeyeaYAnotaEslabon(t *testing.T) {
	st := newFakePortafolioStore()
	dirInst := t.TempDir()
	if err := os.Chmod(dirInst, 0o555); err != nil { //nolint:gosec // G302: el dir read-only es EL insumo de E-25 (nada se escribe ahí).
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dirInst, 0o750) }) //nolint:gosec // G302: restaura para que t.TempDir() pueda limpiar.

	vieja := entradaProvisional("legal-administrativo", domain.Instalacion{
		InstallPath: dirInst, Tipo: domain.InstReferenciadaCC,
		Origen: domain.OrigenPortafolio{Version: "1.2.0", Registry: "github.com/otro/registry"},
	})
	claveVieja := vieja.Identidad.Clave()
	st.entradas[claveVieja] = vieja

	deriva := &derivaConHome{}
	svc := svcConStore(st, deriva)

	nueva, err := svc.AsignarOrigen(context.Background(), claveVieja, homeVitalia)
	if err != nil {
		t.Fatalf("AsignarOrigen: %v", err)
	}

	claveNueva := domain.IdentidadArnes{Home: homeVitalia, ID: "legal-administrativo"}.Clave()
	if nueva.Identidad.Clave() != claveNueva {
		t.Fatalf("clave nueva = %q, want %q", nueva.Identidad.Clave(), claveNueva)
	}
	// El slug de `github.com/vitalia/arneses` colapsa '.' y '/' a '-' (misma regla que todo el
	// Portafolio): la clave real es `github-com-vitalia-arneses~legal-administrativo~`.
	if claveNueva != "github-com-vitalia-arneses~legal-administrativo~" {
		t.Fatalf("clave nueva = %q (el slug del Portafolio colapsa '.' y '/')", claveNueva)
	}
	if _, sigue := st.entradas[claveVieja]; sigue {
		t.Fatalf("la clave vieja %q sigue en el store", claveVieja)
	}
	if _, ok := st.entradas[claveNueva]; !ok {
		t.Fatalf("la clave nueva %q no está en el store: %v", claveNueva, st.entradas)
	}

	var vioEslabon bool
	for _, inst := range nueva.Instalaciones {
		for _, e := range inst.Origen.Eslabones {
			if e.Fuente == "declarado-por-operador" && e.Campo == "home" && e.Valor == homeVitalia {
				vioEslabon = true
			}
		}
	}
	if !vioEslabon {
		t.Fatalf("falta el eslabón {declarado-por-operador,home,%s}: %+v", homeVitalia, nueva.Instalaciones)
	}
	// El home declarado se UNE a los registries (faceta N:M, S1-D3): no reemplaza el existente.
	if len(nueva.Registries) != 1 || nueva.Registries[0] != homeVitalia {
		t.Fatalf("Registries = %v, want [%s]", nueva.Registries, homeVitalia)
	}
	// El dir de la instalación en 0555 no impidió nada: NO se escribe ahí (C3).
	if entradas, rerr := os.ReadDir(dirInst); rerr != nil || len(entradas) != 0 {
		t.Fatalf("se escribió en el dir de la instalación: %v / %v", entradas, rerr)
	}
}

// E-26 · «ninguno — dejarlo sin origen»: la identidad NO cambia, se estampa el sello temporal y el
// contador baja en 1. JAMÁS se inventa un home.
func TestAsignarOrigenNingunoNoInventaHome(t *testing.T) {
	st := newFakePortafolioStore()
	vieja := entradaProvisional("legal-administrativo")
	clave := vieja.Identidad.Clave()
	st.entradas[clave] = vieja
	svc := svcConStore(st, fakeDerivaEvaluator{})

	antes, err := svc.SinOrigenResuelto(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if antes != 1 {
		t.Fatalf("SinOrigenResuelto antes = %d, want 1", antes)
	}

	nueva, err := svc.AsignarOrigen(context.Background(), clave, "")
	if err != nil {
		t.Fatalf("AsignarOrigen(\"\"): %v", err)
	}
	if nueva.Identidad.Clave() != clave {
		t.Fatalf("la identidad CAMBIÓ: %q → %q", clave, nueva.Identidad.Clave())
	}
	if nueva.Identidad.Home != "" {
		t.Fatalf("se inventó un home: %q", nueva.Identidad.Home)
	}
	if nueva.OrigenSinResolverDesde == "" {
		t.Fatal("OrigenSinResolverDesde vacío: sin el sello la fila sigue reclamando atención")
	}
	if _, perr := time.Parse(time.RFC3339, nueva.OrigenSinResolverDesde); perr != nil {
		t.Fatalf("OrigenSinResolverDesde = %q, want RFC3339: %v", nueva.OrigenSinResolverDesde, perr)
	}

	despues, err := svc.SinOrigenResuelto(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if despues != 0 {
		t.Fatalf("SinOrigenResuelto después = %d, want 0", despues)
	}
}

// E-69 · colisión al asignar origen: NO se fusionan las dos identidades, las DOS siguen intactas.
func TestAsignarOrigenColisionaNoFusiona(t *testing.T) {
	st := newFakePortafolioStore()
	ya := domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{Home: homeVitalia, ID: "legal-administrativo"},
		Nombre:    "la que ya estaba",
	}
	st.entradas[ya.Identidad.Clave()] = ya
	otra := entradaProvisional("legal-administrativo")
	st.entradas[otra.Identidad.Clave()] = otra

	svc := svcConStore(st, fakeDerivaEvaluator{})
	_, err := svc.AsignarOrigen(context.Background(), otra.Identidad.Clave(), homeVitalia)
	if !errors.Is(err, usecase.ErrAsignarOrigenColisiona) {
		t.Fatalf("err = %v, want ErrAsignarOrigenColisiona", err)
	}
	if len(st.entradas) != 2 {
		t.Fatalf("entradas = %d, want 2 (las dos intactas)", len(st.entradas))
	}
	if st.entradas[ya.Identidad.Clave()].Nombre != "la que ya estaba" {
		t.Fatal("la entrada existente se pisó")
	}
	if _, ok := st.entradas[otra.Identidad.Clave()]; !ok {
		t.Fatal("la provisional desapareció")
	}
}

// E-70 · asignar origen HABILITA la deriva: la instalación pasa de `no-evaluable` a un veredicto
// real por hash, y el cambio se persiste. Es el efecto útil de reconciliar, sin un botón nuevo.
func TestAsignarOrigenReevaluaDeriva(t *testing.T) {
	st := newFakePortafolioStore()
	vieja := entradaProvisional("legal-administrativo", domain.Instalacion{
		InstallPath: "/proj/.claude/plugins/legal-administrativo", Tipo: domain.InstReferenciadaCC,
		Deriva:        domain.DerivaNoEvaluable,
		DerivaDetalle: "sin referencia local accesible (marketplace no clonado, o esa versión ausente del checkout)",
		Origen:        domain.OrigenPortafolio{Version: "1.2.0"},
	})
	clave := vieja.Identidad.Clave()
	st.entradas[clave] = vieja

	deriva := &derivaConHome{}
	svc := svcConStore(st, deriva)
	nueva, err := svc.AsignarOrigen(context.Background(), clave, homeVitalia)
	if err != nil {
		t.Fatal(err)
	}
	if len(nueva.Instalaciones) != 1 {
		t.Fatalf("instalaciones = %d, want 1", len(nueva.Instalaciones))
	}
	if nueva.Instalaciones[0].Deriva != domain.DerivaAlHilo {
		t.Fatalf("Deriva = %q, want al-hilo (el home nuevo habilitó la referencia)", nueva.Instalaciones[0].Deriva)
	}
	if len(deriva.visto) != 1 || deriva.visto[0] != homeVitalia {
		t.Fatalf("la deriva se evaluó con home %v, want %q", deriva.visto, homeVitalia)
	}
	persistida := st.entradas[nueva.Identidad.Clave()]
	if len(persistida.Instalaciones) != 1 || persistida.Instalaciones[0].Deriva != domain.DerivaAlHilo {
		t.Fatalf("el cambio no se persistió: %+v", persistida.Instalaciones)
	}
}

// Un home que no canonicaliza es 400 explícito y no toca nada.
func TestAsignarOrigenHomeNoCanonicalizable(t *testing.T) {
	st := newFakePortafolioStore()
	vieja := entradaProvisional("x")
	st.entradas[vieja.Identidad.Clave()] = vieja
	svc := svcConStore(st, fakeDerivaEvaluator{})

	_, err := svc.AsignarOrigen(context.Background(), vieja.Identidad.Clave(), "solo-un-nombre")
	if !errors.Is(err, usecase.ErrAsignarOrigenNoCanonicalizable) {
		t.Fatalf("err = %v, want ErrAsignarOrigenNoCanonicalizable", err)
	}
	if len(st.entradas) != 1 || st.entradas[vieja.Identidad.Clave()].Identidad.Home != "" {
		t.Fatal("se tocó el store con un home inválido")
	}
}

// Una clave desconocida es 404.
func TestAsignarOrigenClaveDesconocida(t *testing.T) {
	svc := svcConStore(newFakePortafolioStore(), fakeDerivaEvaluator{})
	_, err := svc.AsignarOrigen(context.Background(), "no-existe", homeVitalia)
	if !errors.Is(err, usecase.ErrObservarClaveNoEncontrada) {
		t.Fatalf("err = %v, want ErrObservarClaveNoEncontrada", err)
	}
}
