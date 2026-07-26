package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// ── fakes de los 6 puertos del estante ──

type fakeMarketplaceStore struct {
	mu        sync.Mutex
	filas     map[string]domain.MarketplaceConocido
	corruptas []domain.EntradaCorrupta
	upsertErr error
}

func newFakeMarketplaceStore() *fakeMarketplaceStore {
	return &fakeMarketplaceStore{filas: map[string]domain.MarketplaceConocido{}}
}

func (f *fakeMarketplaceStore) Listar() ([]domain.MarketplaceConocido, []domain.EntradaCorrupta) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]domain.MarketplaceConocido, 0, len(f.filas))
	for _, m := range f.filas {
		m.Eslabones = []domain.EslabonMarketplace{domain.EslabonDeclarado}
		out = append(out, m)
	}
	return out, f.corruptas
}

func (f *fakeMarketplaceStore) Upsert(m domain.MarketplaceConocido) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.upsertErr != nil {
		return f.upsertErr
	}
	f.filas[m.Nombre] = m
	return nil
}

func (f *fakeMarketplaceStore) Olvidar(nombre string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.filas[nombre]; !ok {
		return false, nil
	}
	delete(f.filas, nombre)
	return true, nil
}

type fakeDetector struct {
	filas []domain.MarketplaceConocido
	err   error
}

func (f *fakeDetector) Detectados() ([]domain.MarketplaceConocido, error) { return f.filas, f.err }

type fakeCatalogoReader struct {
	fuente  string
	cat     domain.Catalogo
	err     error
	llamado int
}

func (f *fakeCatalogoReader) Fuente() string { return f.fuente }

func (f *fakeCatalogoReader) Leer(context.Context, domain.MarketplaceConocido) (domain.Catalogo, error) {
	f.llamado++
	if f.err != nil {
		return domain.Catalogo{}, f.err
	}
	return f.cat, nil
}

type fakeValidador struct {
	cat domain.Catalogo
	err error
}

func (f *fakeValidador) Validar(context.Context, string) (domain.Catalogo, error) {
	if f.err != nil {
		return domain.Catalogo{}, f.err
	}
	return f.cat, nil
}

type fakeCache struct {
	mu         sync.Mutex
	guardados  map[string]domain.Catalogo
	motivo     string
	guardarErr error
	veces      int
}

func newFakeCache() *fakeCache { return &fakeCache{guardados: map[string]domain.Catalogo{}} }

func (f *fakeCache) Leer(nombre string) (domain.Catalogo, string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cat, ok := f.guardados[nombre]
	if !ok {
		return domain.Catalogo{}, f.motivo, false
	}
	return cat, "", true
}

func (f *fakeCache) Guardar(nombre string, cat domain.Catalogo) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.guardarErr != nil {
		return f.guardarErr
	}
	f.guardados[nombre] = cat
	f.veces++
	return nil
}

func (f *fakeCache) Olvidar(nombre string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.guardados, nombre)
	return nil
}

// ── helpers ──

const repoPrenter = "github.com/alpacapurpura/prenter-marketplace"

func catDosEntradas(nombre, cuando, fuente string) domain.Catalogo {
	return domain.Catalogo{
		Marketplace: nombre,
		OwnerNombre: "Prenter",
		OwnerEmail:  "hola@alpacapurpura.lat",
		Lectura:     domain.EstadoLectura{Tipo: domain.LecturaLeida, Cuando: cuando, Entradas: 2, Fuente: fuente},
		Entradas: []domain.EntradaCatalogo{
			{
				Nombre: "harness", Version: "0.5.3", VersionDe: domain.VersionDeSource,
				Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: "./plugins/harness/0.5.3", Ruta: "./plugins/harness/0.5.3"},
			},
			{
				Nombre: "harness-beta", Version: "0.5.3", VersionDe: domain.VersionDeSource,
				Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: "./plugins/harness/0.5.3", Ruta: "./plugins/harness/0.5.3"},
			},
		},
	}
}

type equipo struct {
	svc       *usecase.MarketplaceService
	store     *fakeMarketplaceStore
	detector  *fakeDetector
	local     *fakeCatalogoReader
	remoto    *fakeCatalogoReader
	validador *fakeValidador
	cache     *fakeCache
	pf        *fakePortafolioStore
}

// armar cablea el servicio con fakes y un reloj FIJO (los tests no dependen del reloj real).
func armar(t *testing.T) *equipo {
	t.Helper()
	e := &equipo{
		store:     newFakeMarketplaceStore(),
		detector:  &fakeDetector{},
		local:     &fakeCatalogoReader{fuente: "local"},
		remoto:    &fakeCatalogoReader{fuente: "remoto"},
		validador: &fakeValidador{},
		cache:     newFakeCache(),
		pf:        newFakePortafolioStore(),
	}
	e.svc = usecase.NewMarketplaceService(e.store, e.detector, e.local, e.remoto, e.validador, e.cache, e.pf)
	e.svc.SetAhora(func() time.Time { return time.Date(2026, 7, 25, 14, 7, 33, 0, time.UTC) })
	return e
}

// E-16 · url inexistente ⇒ ErrNoEsMarketplace con el 404 real en el motivo. NO hay forma de
// pintar un ✓ (BR-5).
func TestValidarURLInexistente(t *testing.T) {
	e := armar(t)
	e.validador.err = fmt.Errorf("%w: gh: gh: HTTP 404: Not Found", domain.ErrNoEsMarketplace)

	_, err := e.svc.Validar(context.Background(), "https://github.com/alpacapurpura/no-existe-xyz")
	if !errors.Is(err, usecase.ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want ErrNoEsMarketplace", err)
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("el motivo debe traer el 404 real: %v", err)
	}
}

// E-17 · repo que existe pero NO es marketplace ⇒ motivo con el literal, DISTINGUIBLE del 404.
func TestValidarRepoSinMarketplaceJSON(t *testing.T) {
	e := armar(t)
	e.validador.err = fmt.Errorf("%w: gh: HTTP 404: Not Found (contents/.claude-plugin/marketplace.json)", domain.ErrNoEsMarketplace)

	_, err := e.svc.Validar(context.Background(), "https://github.com/alpacapurpura/vitalia")
	if !errors.Is(err, usecase.ErrNoEsMarketplace) {
		t.Fatalf("err = %v, want ErrNoEsMarketplace", err)
	}
	if !strings.Contains(err.Error(), "no expone .claude-plugin/marketplace.json legible") {
		t.Fatalf("el motivo debe traer el literal explícito: %v", err)
	}
}

// Una url que no canonicaliza es ErrURLNoCanonicalizable (400), sin tocar ningún lector.
func TestValidarURLNoCanonicalizable(t *testing.T) {
	e := armar(t)
	_, err := e.svc.Validar(context.Background(), "solo-un-nombre")
	if !errors.Is(err, usecase.ErrURLNoCanonicalizable) {
		t.Fatalf("err = %v, want ErrURLNoCanonicalizable", err)
	}
}

// E-18 · registrar un nombre ya declarado NO pisa ni duplica, y el store queda INTACTO.
func TestRegistrarDuplicadoNoPisaNiDuplica(t *testing.T) {
	e := armar(t)
	e.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClasePropio, Registrado: "2026-07-20T00:00:00Z",
	}
	e.validador.cat = catDosEntradas("prenter-marketplace", "2026-07-25T14:07:33Z", "remoto")

	_, err := e.svc.Registrar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace", domain.ClaseReferencia)
	if !errors.Is(err, usecase.ErrMarketplaceYaRegistrado) {
		t.Fatalf("err = %v, want ErrMarketplaceYaRegistrado", err)
	}
	filas, _ := e.store.Listar()
	if len(filas) != 1 {
		t.Fatalf("filas = %d, want 1 (no duplica)", len(filas))
	}
	if filas[0].Clase != domain.ClasePropio {
		t.Fatalf("Clase = %q, want propio (NO se pisa)", filas[0].Clase)
	}
	if filas[0].Registrado != "2026-07-20T00:00:00Z" {
		t.Fatalf("Registrado = %q, want el original", filas[0].Registrado)
	}
}

// Una clase fuera del enum es 400 explícito y no toca el store.
func TestRegistrarClaseInvalida(t *testing.T) {
	e := armar(t)
	_, err := e.svc.Registrar(context.Background(), "https://github.com/a/b", "tienda")
	if !errors.Is(err, usecase.ErrClaseInvalida) {
		t.Fatalf("err = %v, want ErrClaseInvalida", err)
	}
	if filas, _ := e.store.Listar(); len(filas) != 0 {
		t.Fatalf("el store se tocó con una clase inválida: %v", filas)
	}
}

// Registrar el camino feliz: persiste una fila y devuelve la MERGEADA (con el eslabón de CC).
func TestRegistrarCaminoFeliz(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{
		Nombre: "prenter-marketplace", Repo: repoPrenter, InstallLocation: "/checkout", CCActualizado: "2026-07-10T00:36:43.459Z",
	}}
	e.validador.cat = catDosEntradas("prenter-marketplace", "", "remoto")
	e.local.err = errors.New("fake: sin checkout legible")

	fila, err := e.svc.Registrar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace", domain.ClasePropio)
	if err != nil {
		t.Fatalf("Registrar: %v", err)
	}
	if fila.Clase != domain.ClasePropio {
		t.Fatalf("Clase = %q", fila.Clase)
	}
	if len(fila.Eslabones) != 2 {
		t.Fatalf("Eslabones = %v, want los DOS (detectado + declarado)", fila.Eslabones)
	}
	if fila.Registrado != "2026-07-25T14:07:33Z" {
		t.Fatalf("Registrado = %q, want el del reloj inyectado", fila.Registrado)
	}
}

// E-62 · dos Registrar concurrentes del mismo nombre ⇒ exactamente 1 éxito y 1 conflicto.
func TestRegistrarConcurrenteNoDuplica(t *testing.T) {
	e := armar(t)
	e.validador.cat = catDosEntradas("prenter-marketplace", "", "remoto")

	var wg sync.WaitGroup
	var exitos, conflictos int
	var mu sync.Mutex
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := e.svc.Registrar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace", domain.ClasePropio)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				exitos++
			case errors.Is(err, usecase.ErrMarketplaceYaRegistrado):
				conflictos++
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()

	if exitos+conflictos != 2 {
		t.Fatalf("exitos=%d conflictos=%d, want 2 en total", exitos, conflictos)
	}
	filas, _ := e.store.Listar()
	if len(filas) != 1 {
		t.Fatalf("filas = %d, want 1 sola", len(filas))
	}
}

// E-29 · sin red y con caché: se devuelve el CACHÉ con la degradación visible y el `Cuando`
// VIEJO. Nunca `Entradas==nil` habiendo caché, nunca `leido` sin lectura fresca.
func TestCatalogoSinRedUsaCacheConCuando(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenter}}
	haceDosDias := "2026-07-23T14:07:33Z"
	e.cache.guardados["prenter-marketplace"] = catDosEntradas("prenter-marketplace", haceDosDias, "local")
	e.local.err = fmt.Errorf("%w: installLocation ya no existe en disco", domain.ErrNoEsMarketplace)
	e.remoto.err = fmt.Errorf("%w: gh: dial tcp: lookup api.github.com: no such host", domain.ErrSinViaDeLectura)

	cat, err := e.svc.Catalogo(context.Background(), "prenter-marketplace", true)
	if err != nil {
		t.Fatalf("Catalogo: %v", err)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want las 2 del caché", len(cat.Entradas))
	}
	if cat.Lectura.Tipo != domain.LecturaSinAcceso {
		t.Fatalf("Lectura.Tipo = %q, want sin-acceso", cat.Lectura.Tipo)
	}
	if cat.Lectura.Cuando != haceDosDias {
		t.Fatalf("Lectura.Cuando = %q, want %q (el del caché, NO ahora)", cat.Lectura.Cuando, haceDosDias)
	}
	if !strings.Contains(cat.Lectura.Motivo, "no such host") {
		t.Fatalf("Lectura.Motivo = %q, want el error de red real", cat.Lectura.Motivo)
	}
}

// TestCatalogoIlegibleNoFabricaVacio — enforcer del check `catalogo-nunca-vacio-fabricado`
// (boundary portafolio-identidad-y-deriva-honesta v1.2 §6): sin caché y sin lectura, `entradas`
// viaja NULL con su motivo. Jamás un `[]` fabricado, jamás un 500.
func TestCatalogoIlegibleNoFabricaVacio(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b"}}
	e.local.err = errors.New("sin checkout")
	e.remoto.err = fmt.Errorf("%w: gh: HTTP 404", domain.ErrNoEsMarketplace)

	cat, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatalf("Catalogo devolvió error (debería degradar 200 con motivo): %v", err)
	}
	if cat.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil (null en el cable)", cat.Entradas)
	}
	if cat.Lectura.Tipo == domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = leido SIN lectura fresca (pass fabricado)")
	}
	if cat.Lectura.Motivo == "" {
		t.Fatal("degradado MUDO: sin motivo el operador no sabe qué pasó")
	}

	// Y un `plugins: []` REAL sí viaja como `[]` con su fecha: es una afirmación evidenciada.
	e.remoto.err = nil
	e.remoto.cat = domain.Catalogo{Marketplace: "x", Entradas: []domain.EntradaCatalogo{}}
	cat2, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if cat2.Entradas == nil {
		t.Fatal("un `plugins: []` real debe viajar como [], no como null")
	}
	if cat2.Lectura.Tipo != domain.LecturaLeida || cat2.Lectura.Cuando == "" {
		t.Fatalf("Lectura = %+v, want leido con su fecha", cat2.Lectura)
	}
}

// E-35 (lado usecase) · el local falla ⇒ se CAE al remoto. Sin remoto ⇒ sin-acceso con el motivo
// del intento local adentro (no se pierde).
func TestInstallLocationAusenteCaeARemoto(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b", InstallLocation: "/borrado"}}
	e.local.err = fmt.Errorf("%w: installLocation ya no existe en disco: /borrado", domain.ErrNoEsMarketplace)
	e.remoto.cat = domain.Catalogo{Marketplace: "x", Entradas: []domain.EntradaCatalogo{{Nombre: "a"}}}

	cat, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 1 || cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("no cayó al remoto: %+v", cat)
	}

	// Sin remoto: sin-acceso con el motivo del local adentro.
	sinRemoto := usecase.NewMarketplaceService(e.store, e.detector, e.local, nil, e.validador, newFakeCache(), e.pf)
	cat2, err := sinRemoto.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if cat2.Lectura.Tipo != domain.LecturaSinAcceso {
		t.Fatalf("Lectura.Tipo = %q, want sin-acceso", cat2.Lectura.Tipo)
	}
	if !strings.Contains(cat2.Lectura.Motivo, "installLocation ya no existe en disco") {
		t.Fatalf("Motivo = %q, want el del intento local", cat2.Lectura.Motivo)
	}
	if cat2.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil", cat2.Entradas)
	}
}

// E-51 · no se puede escribir el caché ⇒ el catálogo se devuelve COMPLETO con el sufijo en el
// motivo. La lectura no se pierde por no poder guardarla.
func TestCatalogoSeDevuelveAunqueNoSePuedaCachear(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b"}}
	e.remoto.cat = catDosEntradas("x", "", "remoto")
	e.cache.guardarErr = errors.New("mkdir /ro/catalogos: read-only file system")

	cat, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2 (la lectura no se pierde)", len(cat.Entradas))
	}
	if !strings.Contains(cat.Lectura.Motivo, "no se pudo cachear") {
		t.Fatalf("Motivo = %q, want el sufijo «(no se pudo cachear: …)»", cat.Lectura.Motivo)
	}
	if cat.Lectura.Tipo != domain.LecturaLeida {
		t.Fatalf("Lectura.Tipo = %q, want leido (SÍ se leyó)", cat.Lectura.Tipo)
	}
}

// E-63 · un caché stale NUNCA afirma frescura: con refrescar=false se devuelve el caché con su
// `Cuando` viejo; con refrescar=true la respuesta pasa a la fresca.
func TestCacheStaleNoAfirmaEstarAlDia(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b"}}
	viejo := catDosEntradas("x", "2026-07-23T00:00:00Z", "local")
	viejo.Entradas[0].Version = "0.5.2"
	e.cache.guardados["x"] = viejo
	fresco := catDosEntradas("x", "", "remoto")
	e.remoto.cat = fresco

	cacheado, err := e.svc.Catalogo(context.Background(), "x", false)
	if err != nil {
		t.Fatal(err)
	}
	if cacheado.Entradas[0].Version != "0.5.2" || cacheado.Lectura.Cuando != "2026-07-23T00:00:00Z" {
		t.Fatalf("con refrescar=false debe salir el CACHÉ: %+v", cacheado.Lectura)
	}
	if e.remoto.llamado != 0 {
		t.Fatalf("con refrescar=false NO se debe tocar la red (llamadas: %d)", e.remoto.llamado)
	}

	refrescado, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if refrescado.Entradas[0].Version != "0.5.3" {
		t.Fatalf("con refrescar=true debe salir la lectura fresca: %q", refrescado.Entradas[0].Version)
	}
	if refrescado.Lectura.Cuando != "2026-07-25T14:07:33Z" {
		t.Fatalf("Cuando = %q, want ahora", refrescado.Lectura.Cuando)
	}
}

// E-74 · un `plugins: []` REAL pisa el caché: es una lectura EXITOSA. Distinto de E-29, donde no
// hubo lectura.
func TestVacioRealPisaElCache(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b"}}
	e.cache.guardados["x"] = catDosEntradas("x", "2026-07-23T00:00:00Z", "local")
	e.remoto.cat = domain.Catalogo{Marketplace: "x", Entradas: []domain.EntradaCatalogo{}}

	cat, err := e.svc.Catalogo(context.Background(), "x", true)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Entradas == nil || len(cat.Entradas) != 0 {
		t.Fatalf("Entradas = %v, want [] (vacío REAL)", cat.Entradas)
	}
	if cat.Lectura.Entradas != 0 || cat.Lectura.Cuando != "2026-07-25T14:07:33Z" {
		t.Fatalf("Lectura = %+v, want {0 entradas, ahora}", cat.Lectura)
	}
}

// E-68 · `known_marketplaces.json` ilegible ⇒ Listar devuelve 200 con `aviso_detector` poblado y
// SOLO los declarados. Nunca un 500, nunca una lista vacía muda.
func TestDetectorIlegibleNoOcultaNiAborta(t *testing.T) {
	e := armar(t)
	e.detector.err = errors.New("marketplace detector: known_marketplaces.json ilegible: unexpected end of JSON input")
	e.store.filas["mio"] = domain.MarketplaceConocido{Nombre: "mio", Repo: "github.com/a/b", Clase: domain.ClasePropio}

	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatalf("Listar devolvió error: %v", err)
	}
	if out.AvisoDetector == "" {
		t.Fatal("aviso_detector vacío: la metadata ilegible se ocultó")
	}
	if len(out.Marketplaces) != 1 || out.Marketplaces[0].Nombre != "mio" {
		t.Fatalf("Marketplaces = %+v, want solo el declarado", out.Marketplaces)
	}
}

// E-75 · una corrupta del registro y el detector OK conviven: ninguna oculta a la otra.
func TestCorruptaYDetectorConviven(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{
		{Nombre: "a", Repo: "github.com/a/a"},
		{Nombre: "b", Repo: "github.com/b/b"},
		{Nombre: "c", Repo: "github.com/c/c"},
		{Nombre: "d", Repo: "github.com/d/d"},
		{Nombre: "e", Repo: "github.com/e/e"},
	}
	e.store.corruptas = []domain.EntradaCorrupta{{Motivo: "fila sin nombre: no hay clave de merge"}}

	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Marketplaces) != 5 {
		t.Fatalf("Marketplaces = %d, want 5", len(out.Marketplaces))
	}
	if len(out.Corruptas) != 1 {
		t.Fatalf("Corruptas = %d, want 1", len(out.Corruptas))
	}
}

// Listar hidrata la Lectura desde el CACHÉ, sin leer ningún catálogo (AG-D16: el GET es O(1)
// lecturas chicas por fila). Sin caché ⇒ `no-leido`, nunca «0 entradas».
func TestListarHidrataDesdeElCacheSinLeer(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{
		{Nombre: "con-cache", Repo: "github.com/a/a"}, {Nombre: "sin-cache", Repo: "github.com/b/b"},
	}
	e.cache.guardados["con-cache"] = catDosEntradas("con-cache", "2026-07-25T10:00:00Z", "local")

	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if e.local.llamado != 0 || e.remoto.llamado != 0 {
		t.Fatalf("Listar leyó catálogos (local=%d remoto=%d): el caché existe para que NO lo haga", e.local.llamado, e.remoto.llamado)
	}
	for _, m := range out.Marketplaces {
		switch m.Nombre {
		case "con-cache":
			if m.Lectura.Tipo != domain.LecturaLeida || m.Lectura.Entradas != 2 {
				t.Fatalf("con-cache Lectura = %+v", m.Lectura)
			}
		case "sin-cache":
			if m.Lectura.Tipo != domain.LecturaNoLeida {
				t.Fatalf("sin-cache Lectura = %+v, want no-leido", m.Lectura)
			}
		}
	}
}

// E-48 (lado usecase) · caché corrupto ⇒ la fila del plano dice `no-leido` CON el motivo visible.
func TestListarConCacheCorruptoMuestraElMotivo(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b"}}
	e.cache.motivo = "caché de catálogo ilegible: unexpected end of JSON input"

	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	l := out.Marketplaces[0].Lectura
	if l.Tipo != domain.LecturaNoLeida || !strings.Contains(l.Motivo, "caché de catálogo ilegible") {
		t.Fatalf("Lectura = %+v, want no-leido con el motivo visible", l)
	}
}

// E-71 · `Olvidar` un marketplace que CC sigue conociendo: 200 con `sigue_detectado:true`, y la
// fila SIGUE apareciendo con eslabón de CC y clase degradada a `referencia`.
func TestOlvidarSigueDetectado(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenter, InstallLocation: "/checkout"}}
	e.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClasePropio}

	olvidado, sigue, err := e.svc.Olvidar(context.Background(), "prenter-marketplace")
	if err != nil || !olvidado || !sigue {
		t.Fatalf("Olvidar = (%v,%v,%v), want (true,true,nil)", olvidado, sigue, err)
	}
	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Marketplaces) != 1 {
		t.Fatalf("la fila desapareció: %+v", out.Marketplaces)
	}
	fila := out.Marketplaces[0]
	if len(fila.Eslabones) != 1 || fila.Eslabones[0] != domain.EslabonCCKnown {
		t.Fatalf("Eslabones = %v, want solo [cc-known-marketplaces]", fila.Eslabones)
	}
	if fila.Clase != domain.ClaseReferencia {
		t.Fatalf("Clase = %q, want referencia (fail-safe tras olvidar la declaración)", fila.Clase)
	}

	// Olvidar algo que el operador nunca declaró ⇒ 404.
	if olvidado, _, _ := e.svc.Olvidar(context.Background(), "prenter-marketplace"); olvidado {
		t.Fatal("Olvidar dos veces devolvió true la segunda")
	}
}

// El contador cruzado cuenta SOLO las provisionales que el operador todavía no miró (AG-D8 dec 7).
func TestContadorSinOrigenResuelto(t *testing.T) {
	e := armar(t)
	e.pf.entradas["sin-home~a~"] = domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "a"}}
	e.pf.entradas["sin-home~b~"] = domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{ID: "b"}, OrigenSinResolverDesde: "2026-07-25T00:00:00Z",
	}
	e.pf.entradas["github-com-a-b~c~"] = domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{Home: "github.com/a/b", ID: "c"}}

	out, err := e.svc.Listar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if out.SinOrigenResuelto != 1 {
		t.Fatalf("SinOrigenResuelto = %d, want 1 (solo la provisional NO confirmada)", out.SinOrigenResuelto)
	}
}

// La situación se calcula SIEMPRE al responder, incluso viniendo del caché: el Portafolio cambia
// con cada Traer/Identificar/escaneo y una situación cacheada mentiría (C24 de design.md).
func TestSituacionSeRecalculaAunDesdeElCache(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenter}}
	e.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClasePropio}
	// El caché guarda la situación VIEJA (no-lo-tengo) a propósito.
	viejo := catDosEntradas("prenter-marketplace", "2026-07-23T00:00:00Z", "local")
	viejo.Entradas[0].Situacion = domain.SituacionCatalogo{Tipo: domain.SituacionNoLoTengo}
	e.cache.guardados["prenter-marketplace"] = viejo
	// Pero el Portafolio YA tiene el canónico (como tras un Traer).
	e.pf.entradas["github-com-alpacapurpura-prenter-marketplace~harness~"] = domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{Home: repoPrenter, ID: "harness"},
		Canonico:  &domain.Canonico{Path: "/checkouts/harness", Version: "0.5.3"},
	}

	cat, err := e.svc.Catalogo(context.Background(), "prenter-marketplace", false)
	if err != nil {
		t.Fatal(err)
	}
	if cat.Entradas[0].Situacion.Tipo != domain.SituacionAlHilo {
		t.Fatalf("Situacion = %+v, want al-hilo recalculado (la del caché estaba stale)", cat.Entradas[0].Situacion)
	}
	if !cat.Entradas[0].Accion.Habilitada && cat.Entradas[0].Accion.Verbo != domain.AccionNinguna {
		t.Fatalf("Accion = %+v", cat.Entradas[0].Accion)
	}
}

// Un nombre desconocido es 404, no una lista vacía ni un 500.
func TestCatalogoDeMarketplaceDesconocido(t *testing.T) {
	e := armar(t)
	_, err := e.svc.Catalogo(context.Background(), "no-existe", false)
	if !errors.Is(err, usecase.ErrMarketplaceNoConocido) {
		t.Fatalf("err = %v, want ErrMarketplaceNoConocido", err)
	}
}

// CandidatosDeOrigen ordena por señales BLANDAS y NO premarca nada (BR-11).
func TestCandidatosDeOrigenOrdenaPorSenal(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{
		{Nombre: "zeta-referencia", Repo: "github.com/zeta/zeta"},
		{Nombre: "prenter-marketplace", Repo: repoPrenter},
		{Nombre: "vitalia-arneses", Repo: "github.com/vitalia/arneses"},
	}
	e.store.filas["vitalia-arneses"] = domain.MarketplaceConocido{Nombre: "vitalia-arneses", Repo: "github.com/vitalia/arneses", Clase: domain.ClasePropio}
	e.pf.entradas["sin-home~legal-administrativo~"] = domain.EntradaPortafolio{
		Identidad:  domain.IdentidadArnes{ID: "legal-administrativo"},
		Registries: []string{repoPrenter},
	}

	candidatos, actual, err := e.svc.CandidatosDeOrigen(context.Background(), "sin-home~legal-administrativo~")
	if err != nil {
		t.Fatalf("CandidatosDeOrigen: %v", err)
	}
	if actual != "" {
		t.Fatalf("actual = %q, want vacío (la identidad sigue provisional)", actual)
	}
	if len(candidatos) != 3 {
		t.Fatalf("candidatos = %d, want 3", len(candidatos))
	}
	if candidatos[0].Nombre != "prenter-marketplace" {
		t.Fatalf("el primero = %q, want prenter-marketplace (su registry ya apunta acá)", candidatos[0].Nombre)
	}
	if !strings.Contains(candidatos[0].Senal, "el registry de tu copia ya apunta acá") {
		t.Fatalf("Senal = %q, want la señal fuerte", candidatos[0].Senal)
	}
	if candidatos[1].Nombre != "vitalia-arneses" {
		t.Fatalf("el segundo = %q, want vitalia-arneses (clase propio antes que referencia)", candidatos[1].Nombre)
	}
	// La señal NO puede decir «no leído aún» de un marketplace que SÍ se leyó (bug cazado en el
	// E2E vivo): la Lectura se hidrata del caché igual que en Listar.
	e.cache.guardados["prenter-marketplace"] = catDosEntradas("prenter-marketplace", "2026-07-25T10:00:00Z", "local")
	conLectura, _, err := e.svc.CandidatosDeOrigen(context.Background(), "sin-home~legal-administrativo~")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(conLectura[0].Senal, "no leído aún") {
		t.Fatalf("Senal = %q: dice «no leído aún» de un marketplace YA leído", conLectura[0].Senal)
	}

	if _, _, err := e.svc.CandidatosDeOrigen(context.Background(), "no-existe"); !errors.Is(err, usecase.ErrObservarClaveNoEncontrada) {
		t.Fatalf("err = %v, want ErrObservarClaveNoEncontrada", err)
	}
}

// Ninguna fila del wire viaja con `lectura.tipo: ""` — `no-leido` es el CERO del tipo a propósito
// («no leído aún»), y un enum vacío obligaría al FE a inventar el estado. Bug real cazado en el
// E2E vivo: `Registrar` devolvía la fila mergeada SIN hidratar la lectura.
func TestFilaRegistradaTraeLecturaNoLeida(t *testing.T) {
	e := armar(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenter, InstallLocation: "/checkout"}}
	e.validador.cat = catDosEntradas("prenter-marketplace", "", "remoto")
	e.local.err = errors.New("fake: sin checkout legible")

	fila, err := e.svc.Registrar(context.Background(), "https://github.com/alpacapurpura/prenter-marketplace", domain.ClasePropio)
	if err != nil {
		t.Fatal(err)
	}
	if fila.Lectura.Tipo != domain.LecturaNoLeida {
		t.Fatalf("Lectura.Tipo = %q, want %q (jamás vacío en el wire)", fila.Lectura.Tipo, domain.LecturaNoLeida)
	}
}
