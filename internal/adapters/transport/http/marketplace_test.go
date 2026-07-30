package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/marketplace"
	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/adapters/traer"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// marketplace_test.go prueba el WIRE: método · ruta · status · shape del JSON. Las reglas de
// negocio viven en D/U y se prueban allá (capa H del plan de pruebas).

// ── dobles mínimos de los puertos del estante ──

type mktDetector struct {
	filas []domain.MarketplaceConocido
	err   error
}

func (m *mktDetector) Detectados() ([]domain.MarketplaceConocido, error) { return m.filas, m.err }

type mktReader struct {
	fuente string
	cat    domain.Catalogo
	err    error
}

func (m *mktReader) Fuente() string { return m.fuente }
func (m *mktReader) Leer(context.Context, domain.MarketplaceConocido) (domain.Catalogo, error) {
	return m.cat, m.err
}

type mktValidador struct {
	cat domain.Catalogo
	err error
}

func (m *mktValidador) Validar(context.Context, string) (domain.Catalogo, error) {
	return m.cat, m.err
}

const repoPrenterHTTP = "github.com/alpacapurpura/prenter-marketplace"

type escenarioHTTP struct {
	svc       *usecase.MarketplaceService
	pfSvc     *usecase.PortafolioService
	pfStore   *portafolio.Store
	store     *marketplace.Store
	detector  *mktDetector
	local     *mktReader
	remoto    *mktReader
	validador *mktValidador
	raiz      string
}

// armarHTTP cablea el servicio con adapters REALES de persistencia (store + caché sobre temp) y
// dobles solo en lo que sale de la máquina (lectores). La raíz de Traer es inyectada.
func armarHTTP(t *testing.T) *escenarioHTTP {
	t.Helper()
	tmp := t.TempDir()
	st, err := marketplace.NewStore(filepath.Join(tmp, "marketplaces.json"))
	if err != nil {
		t.Fatal(err)
	}
	cache, err := marketplace.NewCache(filepath.Join(tmp, "catalogos"))
	if err != nil {
		t.Fatal(err)
	}
	pfStore, err := portafolio.NewStore(filepath.Join(tmp, "portafolio.json"))
	if err != nil {
		t.Fatal(err)
	}
	e := &escenarioHTTP{
		pfStore:   pfStore,
		store:     st,
		detector:  &mktDetector{},
		local:     &mktReader{fuente: "local", err: errors.New("sin checkout local en el test")},
		remoto:    &mktReader{fuente: "remoto"},
		validador: &mktValidador{},
		raiz:      filepath.Join(tmp, "arnesia"),
	}
	e.svc = usecase.NewMarketplaceService(st, e.detector, e.local, e.remoto, e.validador, cache, pfStore)
	e.svc.SetTraer(&traer.CopiadorLocal{}, &traer.ClonadorExterno{GitBin: "/no/existe/git"}, fakeDeriva{}, e.raiz)
	e.pfSvc = usecase.NewPortafolioService(pfStore, &fakeScan{}, &fakeLoad{}, fakeDeriva{}, nil, fakeSchemas{})
	return e
}

func catPrenterHTTP() domain.Catalogo {
	return domain.Catalogo{
		Marketplace: "prenter-marketplace",
		OwnerNombre: "Prenter",
		OwnerEmail:  "hola@alpacapurpura.lat",
		Lectura:     domain.EstadoLectura{Tipo: domain.LecturaLeida, Fuente: "remoto"},
		Entradas: []domain.EntradaCatalogo{{
			Nombre: "harness", Version: "0.5.3", VersionDe: domain.VersionDeSource,
			Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: "./plugins/harness/0.5.3", Ruta: "./plugins/harness/0.5.3"},
		}},
	}
}

// GET /api/marketplaces siempre responde 200 con el shape del plano; `marketplaces` NUNCA es null.
func TestGetMarketplacesShape(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{
		Nombre: "prenter-marketplace", Repo: repoPrenterHTTP, InstallLocation: "/checkout",
		CCActualizado: "2026-07-10T00:36:43.459Z",
	}}

	rec := httptest.NewRecorder()
	listMarketplaces(e.svc)(rec, httptest.NewRequestWithContext(context.Background(), "GET", "/api/marketplaces", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var out usecase.ListadoMarketplaces
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v — %s", err, rec.Body.String())
	}
	if len(out.Marketplaces) != 1 || out.Marketplaces[0].Nombre != "prenter-marketplace" {
		t.Fatalf("marketplaces = %+v", out.Marketplaces)
	}
	if !strings.Contains(rec.Body.String(), `"marketplaces":[`) {
		t.Fatalf("`marketplaces` debe viajar como array, nunca null: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"sin_origen_resuelto"`) {
		t.Fatalf("falta el contador cruzado: %s", rec.Body.String())
	}
}

// E-68 (lado HTTP) · detector ilegible ⇒ 200 con `aviso_detector`, jamás un 500.
func TestGetMarketplacesDetectorIlegible200(t *testing.T) {
	e := armarHTTP(t)
	e.detector.err = errors.New("known_marketplaces.json ilegible: unexpected end of JSON input")

	rec := httptest.NewRecorder()
	listMarketplaces(e.svc)(rec, httptest.NewRequestWithContext(context.Background(), "GET", "/api/marketplaces", nil))
	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "aviso_detector") {
		t.Fatalf("falta aviso_detector: %s", rec.Body.String())
	}
}

// El servicio no cableado responde 500 con motivo, no un panic.
func TestGetMarketplacesSinServicio500(t *testing.T) {
	rec := httptest.NewRecorder()
	listMarketplaces(nil)(rec, httptest.NewRequestWithContext(context.Background(), "GET", "/api/marketplaces", nil))
	if rec.Code != 500 {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

// E-16 · repo inexistente ⇒ 400 con el 404 en el body, y el body NO trae nada que se lea como ✓.
func TestPostValidacionesRepoInexistente400(t *testing.T) {
	e := armarHTTP(t)
	e.validador.err = fmt.Errorf("%w: gh: gh: HTTP 404: Not Found", domain.ErrNoEsMarketplace)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/validaciones",
		strings.NewReader(`{"url":"https://github.com/alpacapurpura/no-existe-xyz"}`))
	postValidarMarketplace(e.svc)(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "404") {
		t.Fatalf("el body debe traer el 404 real: %s", rec.Body.String())
	}
	for _, prohibido := range []string{`"valido"`, `"entradas"`, `"nombre"`} {
		if strings.Contains(rec.Body.String(), prohibido) {
			t.Fatalf("el body de error trae %s: no hay forma de pintar un ✓ sin lectura (BR-5): %s", prohibido, rec.Body.String())
		}
	}
}

// E-17 · repo que existe pero no es marketplace ⇒ 400 con el literal, DISTINGUIBLE del 404.
func TestPostValidacionesNoEsMarketplace400(t *testing.T) {
	e := armarHTTP(t)
	e.validador.err = fmt.Errorf("%w: gh: HTTP 404 en .claude-plugin/marketplace.json", domain.ErrNoEsMarketplace)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/validaciones",
		strings.NewReader(`{"url":"https://github.com/alpacapurpura/vitalia"}`))
	postValidarMarketplace(e.svc)(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "no expone .claude-plugin/marketplace.json legible") {
		t.Fatalf("el body debe traer el literal explícito: %s", rec.Body.String())
	}
}

// E-54 (lado HTTP) · sin `gh` ni PAT ⇒ **503**, no 400: el sistema dice «no puedo mirar».
func TestPostValidacionesSinViaDeLectura503(t *testing.T) {
	e := armarHTTP(t)
	e.validador.err = fmt.Errorf("%w: `gh` no está en el PATH y ARNESIA_GH_TOKEN está vacío", domain.ErrSinViaDeLectura)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/validaciones",
		strings.NewReader(`{"url":"https://github.com/a/b"}`))
	postValidarMarketplace(e.svc)(rec, req)

	if rec.Code != 503 {
		t.Fatalf("status = %d, want 503 (nunca 400: la url no está mal): %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "sin vía de lectura") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

// Validar con éxito ⇒ 200 con los 3 datos LEÍDOS del archivo real.
func TestPostValidaciones200ConDatosLeidos(t *testing.T) {
	e := armarHTTP(t)
	e.validador.cat = catPrenterHTTP()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/validaciones",
		strings.NewReader(`{"url":"https://github.com/alpacapurpura/prenter-marketplace"}`))
	postValidarMarketplace(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var v usecase.Validacion
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Nombre != "prenter-marketplace" || v.OwnerNombre != "Prenter" || v.Entradas != 1 {
		t.Fatalf("Validacion = %+v", v)
	}
	if v.URLCanonica != repoPrenterHTTP {
		t.Fatalf("URLCanonica = %q, want canonicalizada", v.URLCanonica)
	}
}

// E-18 · registrar duplicado ⇒ **409** con `{"error":…,"nombre":…}` para que la UI ofrezca «ir a él».
func TestPostMarketplaces409ConNombre(t *testing.T) {
	e := armarHTTP(t)
	e.validador.cat = catPrenterHTTP()
	if err := e.store.Upsert(domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenterHTTP, Clase: domain.ClasePropio,
	}); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces",
		strings.NewReader(`{"url":"https://github.com/alpacapurpura/prenter-marketplace","clase":"referencia"}`))
	postRegistrarMarketplace(e.svc)(rec, req)

	if rec.Code != 409 {
		t.Fatalf("status = %d, want 409: %s", rec.Code, rec.Body.String())
	}
	var body conflictoMarketplaceBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Nombre != "prenter-marketplace" {
		t.Fatalf("nombre = %q, want el existente (para «ir a él» sin parsear texto)", body.Nombre)
	}
	// El store queda INTACTO: sigue `propio`.
	filas, _ := e.store.Listar()
	if len(filas) != 1 || filas[0].Clase != domain.ClasePropio {
		t.Fatalf("el store se pisó: %+v", filas)
	}
}

// Una clase fuera del enum ⇒ 400.
func TestPostMarketplacesClaseInvalida400(t *testing.T) {
	e := armarHTTP(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces", strings.NewReader(`{"url":"https://github.com/a/b","clase":"tienda"}`))
	postRegistrarMarketplace(e.svc)(rec, req)
	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

// E-71 (lado HTTP) · DELETE de un declarado que CC sigue conociendo ⇒ 200
// `{olvidado:true, sigue_detectado:true}`. Uno no declarado ⇒ 404.
func TestDeleteMarketplaceSigueDetectado(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenterHTTP}}
	if err := e.store.Upsert(domain.MarketplaceConocido{Nombre: "prenter-marketplace", Repo: repoPrenterHTTP, Clase: domain.ClasePropio}); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/marketplaces/prenter-marketplace", nil)
	req.SetPathValue("nombre", "prenter-marketplace")
	deleteOlvidarMarketplace(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var body map[string]bool
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body["olvidado"] || !body["sigue_detectado"] {
		t.Fatalf("body = %v, want olvidado+sigue_detectado", body)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), "DELETE", "/api/marketplaces/prenter-marketplace", nil)
	req2.SetPathValue("nombre", "prenter-marketplace")
	deleteOlvidarMarketplace(e.svc)(rec2, req2)
	if rec2.Code != 404 {
		t.Fatalf("status = %d, want 404", rec2.Code)
	}
}

// BR-4 en el CABLE · un catálogo no legible viaja **200 con `entradas: null`** + motivo. No hay
// 5xx por «no pude leer el catálogo».
func TestGetCatalogoDegradado200ConEntradasNull(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "vitalia-arneses", Repo: "github.com/vitalia/arneses"}}
	e.remoto.err = fmt.Errorf("%w: gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual", domain.ErrNoEsMarketplace)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/marketplaces/vitalia-arneses/catalogo", nil)
	req.SetPathValue("nombre", "vitalia-arneses")
	getCatalogoMarketplace(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200 (nunca 5xx por no poder leer): %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"entradas":null`) {
		t.Fatalf("el cable debe traer `entradas: null`, no `[]`: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "HTTP 404") {
		t.Fatalf("falta el motivo real: %s", rec.Body.String())
	}
}

// Un marketplace desconocido ⇒ 404.
func TestGetCatalogoDesconocido404(t *testing.T) {
	e := armarHTTP(t)
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/marketplaces/nadie/catalogo", nil)
	req.SetPathValue("nombre", "nadie")
	getCatalogoMarketplace(e.svc)(rec, req)
	if rec.Code != 404 {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

// El POST de lecturas es el REFRESCO explícito y responde el mismo shape que el GET.
func TestPostLecturasRefrescaYDevuelveElMismoShape(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenterHTTP}}
	e.remoto.cat = catPrenterHTTP()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/prenter-marketplace/lecturas", nil)
	req.SetPathValue("nombre", "prenter-marketplace")
	postLeerCatalogo(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var cat domain.Catalogo
	if err := json.Unmarshal(rec.Body.Bytes(), &cat); err != nil {
		t.Fatal(err)
	}
	if len(cat.Entradas) != 1 || cat.Lectura.Tipo != domain.LecturaLeida || cat.Lectura.Cuando == "" {
		t.Fatalf("Catalogo = %+v", cat)
	}
}

// ── `↧ Traer canónico` (§13.8) ──

// prepararTraerLocal deja un marketplace `propio` con checkout REAL en disco y su catálogo
// cacheado, listo para el camino A.
func prepararTraerLocal(t *testing.T, e *escenarioHTTP) {
	t.Helper()
	checkout := t.TempDir()
	sub := filepath.Join(checkout, "plugins", "harness", "0.5.3")
	if err := os.MkdirAll(sub, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "VERSION"), []byte("KIT_VERSION=0.5.3\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e.detector.filas = []domain.MarketplaceConocido{{
		Nombre: "prenter-marketplace", Repo: repoPrenterHTTP, InstallLocation: checkout,
	}}
	if err := e.store.Upsert(domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenterHTTP, Clase: domain.ClasePropio,
	}); err != nil {
		t.Fatal(err)
	}
	e.local.err = nil
	e.local.cat = catPrenterHTTP()
}

// E-76 (lado HTTP) · el camino A feliz responde 200 con el shape de ResultadoTraer.
func TestPostTraidos200(t *testing.T) {
	e := armarHTTP(t)
	prepararTraerLocal(t, e)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/prenter-marketplace/traidos", strings.NewReader(`{"entrada":"harness"}`))
	req.SetPathValue("nombre", "prenter-marketplace")
	postTraerCanonico(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	var res usecase.ResultadoTraer
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Camino != domain.CaminoLocal {
		t.Fatalf("camino = %q, want local", res.Camino)
	}
	esperado := filepath.Join(e.raiz, "checkouts", "prenter-marketplace", "harness")
	if res.Destino != esperado {
		t.Fatalf("destino = %q, want %q", res.Destino, esperado)
	}
	// BR-17 · un 200 puede traer un veredicto incómodo: la deriva se muestra tal cual salga.
	if res.Deriva != domain.DerivaNoEvaluable {
		t.Fatalf("deriva = %q, want el veredicto del fake (no-evaluable) — 200 igual", res.Deriva)
	}
	if !strings.Contains(rec.Body.String(), `"deriva"`) {
		t.Fatalf("la deriva debe viajar SIEMPRE: %s", rec.Body.String())
	}
}

// E-79 (lado HTTP) · repetir ⇒ **409** con `{"error":…,"destino":…}` para ofrecer «abrir el
// canónico que ya tenés».
func TestPostTraidos409ConDestino(t *testing.T) {
	e := armarHTTP(t)
	prepararTraerLocal(t, e)
	for i := range 2 {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/prenter-marketplace/traidos", strings.NewReader(`{"entrada":"harness"}`))
		req.SetPathValue("nombre", "prenter-marketplace")
		postTraerCanonico(e.svc)(rec, req)
		if i == 0 {
			if rec.Code != 200 {
				t.Fatalf("el primero = %d, want 200: %s", rec.Code, rec.Body.String())
			}
			continue
		}
		if rec.Code != 409 {
			t.Fatalf("el segundo = %d, want 409: %s", rec.Code, rec.Body.String())
		}
		var body conflictoTraerBody
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Destino == "" {
			t.Fatalf("falta `destino` en el 409 (la UI no puede ofrecer abrirlo): %s", rec.Body.String())
		}
		if !strings.HasSuffix(body.Destino, filepath.Join("checkouts", "prenter-marketplace", "harness")) {
			t.Fatalf("destino = %q", body.Destino)
		}
	}
}

// E-80 (lado HTTP) · clase `referencia` ⇒ **400** con el literal firmado.
func TestPostTraidosReferencia400(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{
		Nombre: "claude-plugins-official", Repo: "github.com/anthropics/claude-plugins-official", InstallLocation: t.TempDir(),
	}}
	e.local.err = nil
	e.local.cat = domain.Catalogo{
		Marketplace: "claude-plugins-official",
		Lectura:     domain.EstadoLectura{Tipo: domain.LecturaLeida, Fuente: "local"},
		Entradas: []domain.EntradaCatalogo{{
			Nombre: "frontend-design",
			Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: "./plugins/frontend-design", Ruta: "./plugins/frontend-design"},
		}},
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/marketplaces/claude-plugins-official/traidos", strings.NewReader(`{"entrada":"frontend-design"}`))
	req.SetPathValue("nombre", "claude-plugins-official")
	postTraerCanonico(e.svc)(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), domain.MotivoSoloArnesesPropios) {
		t.Fatalf("el body debe traer el literal %q: %s", domain.MotivoSoloArnesesPropios, rec.Body.String())
	}
	// Cero I/O: no se creó nada bajo la raíz inyectada.
	if _, err := os.Stat(e.raiz); err == nil {
		t.Fatal("se creó la raíz de trabajo con una clase referencia (debía abortar en el dominio)")
	}
}

// E-84/E-85 (lado HTTP) · los tres códigos de Traer son DISTINTOS y ninguno es genérico.
func TestTraerStatusPorCentinela(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		want   int
	}{
		{"marketplace desconocido ⇒ 404", fmt.Errorf("%w: x", usecase.ErrMarketplaceNoConocido), 404},
		{"clase referencia ⇒ 400", fmt.Errorf("%w: x", domain.ErrTraerClaseReferencia), 400},
		{"source no materializable ⇒ 400", fmt.Errorf("%w: 42", domain.ErrTraerSourceNoMaterializable), 400},
		{"destino que escapa ⇒ 400", fmt.Errorf("%w: x", domain.ErrTraerDestinoEscapa), 400},
		{"entrada ajena al catálogo ⇒ 400", fmt.Errorf("%w: x", usecase.ErrEntradaNoEnCatalogo), 400},
		{"destino poblado ⇒ 409", fmt.Errorf("%w: /x/y", usecase.ErrTraerDestinoPoblado), 409},
		{"sin credencial ⇒ 503", fmt.Errorf("%w: could not read Username", usecase.ErrTraerSinAuth), 503},
		{"sin materializador ⇒ 503", fmt.Errorf("%w: externo", usecase.ErrTraerSinMaterializador), 503},
		{"el remoto no tiene ⇒ 502", fmt.Errorf("%w: Repository not found", usecase.ErrTraerRemotoNoTiene), 502},
		{"sha que no coincide ⇒ 502", fmt.Errorf("%w: aaa vs bbb", usecase.ErrTraerSHANoCoincide), 502},
		{"fallo local ⇒ 500", fmt.Errorf("%w: ENOSPC", usecase.ErrTraerLocal), 500},
	}
	vistos := map[int]bool{}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			rec := httptest.NewRecorder()
			escribirErrorTraer(rec, c.err)
			if rec.Code != c.want {
				t.Fatalf("status = %d, want %d", rec.Code, c.want)
			}
			if rec.Body.Len() == 0 || !strings.Contains(rec.Body.String(), "error") {
				t.Fatalf("el body debe traer el motivo real: %s", rec.Body.String())
			}
			vistos[c.want] = true
		})
	}
	// Los 3 códigos que el FE usa distinto tienen que existir de verdad en la tabla.
	for _, code := range []int{400, 502, 503} {
		if !vistos[code] {
			t.Fatalf("la tabla no cubre el status %d", code)
		}
	}
}

// ── S7 · candidatos + asignar origen ──

// GET …/origen/candidatos: 200 con `candidatos` array (nunca null) + `actual`; 404 si la clave no
// existe.
func TestGetCandidatosOrigen(t *testing.T) {
	e := armarHTTP(t)
	e.detector.filas = []domain.MarketplaceConocido{{Nombre: "prenter-marketplace", Repo: repoPrenterHTTP}}
	entrada := domain.EntradaPortafolio{
		Identidad:  domain.IdentidadArnes{ID: "harness"},
		Registries: []string{repoPrenterHTTP},
	}
	if err := e.pfStore.Upsert(entrada); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/api/portafolio/arneses/x/origen/candidatos", nil)
	req.SetPathValue("clave", entrada.Identidad.Clave())
	getCandidatosOrigen(e.svc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"candidatos":[`) {
		t.Fatalf("`candidatos` debe viajar como array: %s", rec.Body.String())
	}
	var out candidatosOrigenResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Candidatos) != 1 || out.Candidatos[0].Senal == "" {
		t.Fatalf("candidatos = %+v, want 1 con señal armada por el backend", out.Candidatos)
	}
	if out.Actual != "" {
		t.Fatalf("actual = %q, want vacío (identidad provisional)", out.Actual)
	}

	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), "GET", "/api/portafolio/arneses/no-existe/origen/candidatos", nil)
	req2.SetPathValue("clave", "no-existe")
	getCandidatosOrigen(e.svc)(rec2, req2)
	if rec2.Code != 404 {
		t.Fatalf("status = %d, want 404", rec2.Code)
	}
}

// POST …/origen: 200 con la entrada RE-KEYED; 400 sin `home` ni `sin_origen`; 409 en colisión.
func TestPostAsignarOrigen(t *testing.T) {
	e := armarHTTP(t)
	entrada := domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "legal-administrativo"}}
	if err := e.pfStore.Upsert(entrada); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/x/origen", strings.NewReader(`{"home":"github.com/vitalia/arneses"}`))
	req.SetPathValue("clave", entrada.Identidad.Clave())
	postAsignarOrigen(e.pfSvc)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}
	var out entradaWire
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Clave != "github-com-vitalia-arneses~legal-administrativo~" {
		t.Fatalf("clave = %q, want la re-keyed", out.Clave)
	}
	if out.Identidad.Home != "github.com/vitalia/arneses" {
		t.Fatalf("home = %q", out.Identidad.Home)
	}

	// Sin `home` ni `sin_origen`: 400 explícito — elegir «ninguno» es EXPLÍCITO, no un default.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/x/origen", strings.NewReader(`{}`))
	req2.SetPathValue("clave", out.Clave)
	postAsignarOrigen(e.pfSvc)(rec2, req2)
	if rec2.Code != 400 {
		t.Fatalf("status = %d, want 400: %s", rec2.Code, rec2.Body.String())
	}

	// `sin_origen: true` ⇒ 200 sin tocar la identidad.
	otra := domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "otra"}}
	if err := e.pfStore.Upsert(otra); err != nil {
		t.Fatal(err)
	}
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/x/origen", strings.NewReader(`{"sin_origen":true}`))
	req3.SetPathValue("clave", otra.Identidad.Clave())
	postAsignarOrigen(e.pfSvc)(rec3, req3)
	if rec3.Code != 200 {
		t.Fatalf("status = %d: %s", rec3.Code, rec3.Body.String())
	}
	if !strings.Contains(rec3.Body.String(), "origen_sin_resolver_desde") {
		t.Fatalf("falta el sello de «ninguno»: %s", rec3.Body.String())
	}

	// E-69 (lado HTTP) · colisión ⇒ 409.
	tercera := domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "legal-administrativo", Scope: "otro"}}
	if err := e.pfStore.Upsert(tercera); err != nil {
		t.Fatal(err)
	}
	rec4 := httptest.NewRecorder()
	req4 := httptest.NewRequestWithContext(context.Background(), "POST", "/api/portafolio/arneses/x/origen", strings.NewReader(`{"home":"github.com/vitalia/arneses"}`))
	req4.SetPathValue("clave", tercera.Identidad.Clave())
	postAsignarOrigen(e.pfSvc)(rec4, req4)
	if rec4.Code != 409 {
		t.Fatalf("status = %d, want 409: %s", rec4.Code, rec4.Body.String())
	}
}
