package usecase_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// publicar_test.go prueba la POLÍTICA de `▲ Publicar` (RF-B2.3) con fakes: las guardas corren
// TODAS antes del publisher, el gate de conformance bloquea con el reporte adjunto, y el
// post-éxito deja el registro local al día. El mecanismo git se prueba en
// internal/adapters/publish (repos reales).

// ── fakes de los 3 puertos de Publicar ──

type fakePublisher struct {
	solicitudes []domain.SolicitudPublicacion
	res         domain.ResultadoPublicacion
	err         error
}

func (f *fakePublisher) Publicar(_ context.Context, sol domain.SolicitudPublicacion) (domain.ResultadoPublicacion, error) {
	f.solicitudes = append(f.solicitudes, sol)
	if f.err != nil {
		return domain.ResultadoPublicacion{}, f.err
	}
	return f.res, nil
}

type fakeLoaderPub struct {
	g   domain.Graph
	err error
}

func (f *fakeLoaderPub) Load(string) (domain.Graph, error) { return f.g, f.err }

type fakeConformance struct {
	rep     domain.ConformanceReport
	err     error
	corrido int
}

func (f *fakeConformance) Run(context.Context, ports.Target) (domain.ConformanceReport, error) {
	return f.rep, f.err
}

func (f *fakeConformance) RunGraph(context.Context, []byte, string) (domain.ConformanceReport, error) {
	f.corrido++
	return f.rep, f.err
}

func reporteVerde() domain.ConformanceReport {
	return domain.ConformanceReport{Results: []domain.CheckResult{
		{Check: domain.Check{ID: "ok", Severidad: domain.SevError}, Veredicto: domain.VeredictoPass},
	}}
}

func reporteRojo() domain.ConformanceReport {
	return domain.ConformanceReport{Results: []domain.CheckResult{
		{Check: domain.Check{ID: "escritor-unico", Severidad: domain.SevError}, Veredicto: domain.VeredictoFail, Detalle: "dos cajas escriben el mismo artefacto"},
	}}
}

// equipoPublicar arma el escenario feliz completo: marketplace propio declarado, entrada con
// canónico + home, loader con versión semver, gate verde. Cada test rompe UNA pieza.
type equipoPublicar struct {
	*equipo
	pub    *fakePublisher
	loader *fakeLoaderPub
	conf   *fakeConformance
	clave  string
}

func armarPublicar(t *testing.T) *equipoPublicar {
	t.Helper()
	e := armar(t)
	e.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClasePropio,
	}
	entrada := domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{Home: repoPrenter, ID: "harness"},
		Nombre:    "harness",
		Canonico:  &domain.Canonico{Path: "/checkouts/prenter/harness", Version: "0.5.3"},
	}
	if err := e.pf.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	ep := &equipoPublicar{
		equipo: e,
		pub:    &fakePublisher{res: domain.ResultadoPublicacion{Commit: "a1b2c3", Tag: "harness/v0.6.0"}},
		loader: &fakeLoaderPub{g: domain.Graph{Arnes: &domain.Arnes{ID: "harness", Version: "0.6.0"}, Nodes: []domain.Box{}}},
		conf:   &fakeConformance{rep: reporteVerde()},
		clave:  entrada.Identidad.Clave(),
	}
	ep.svc.SetPublicar(ep.pub, ep.loader, ep.conf)
	return ep
}

// Sin SetPublicar el servicio degrada honesto (503), jamás un nil-pointer panic.
func TestPublicarNoDisponible(t *testing.T) {
	e := armar(t)
	_, err := e.svc.Publicar(context.Background(), "cualquiera")
	if !errors.Is(err, usecase.ErrPublicarNoDisponible) {
		t.Fatalf("err = %v, want ErrPublicarNoDisponible", err)
	}
}

// Clave desconocida ⇒ 404 (solo se publica lo YA persistido — la autoridad es el store).
func TestPublicarClaveNoEncontrada(t *testing.T) {
	ep := armarPublicar(t)
	_, err := ep.svc.Publicar(context.Background(), "no-existe~x~")
	if !errors.Is(err, usecase.ErrObservarClaveNoEncontrada) {
		t.Fatalf("err = %v, want ErrObservarClaveNoEncontrada", err)
	}
}

// Guarda 2 · sin canónico no hay copia editable que publicar (ley anti-drift).
func TestPublicarSinCanonico(t *testing.T) {
	ep := armarPublicar(t)
	entrada := domain.EntradaPortafolio{
		Identidad:     domain.IdentidadArnes{Home: repoPrenter, ID: "solo-instalado"},
		Instalaciones: []domain.Instalacion{{InstallPath: "/p/.claude/plugins/solo-instalado"}},
	}
	if err := ep.pf.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	_, err := ep.svc.Publicar(context.Background(), entrada.Identidad.Clave())
	if !errors.Is(err, domain.ErrPublicarSinCanonico) {
		t.Fatalf("err = %v, want ErrPublicarSinCanonico", err)
	}
	if len(ep.pub.solicitudes) != 0 {
		t.Fatal("el publisher se llamó con una guarda rota")
	}
}

// Guarda 3 · identidad provisional (sin home) ⇒ no hay estante destino.
func TestPublicarSinHome(t *testing.T) {
	ep := armarPublicar(t)
	entrada := domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{ID: "provisional", Scope: "proyecto"},
		Canonico:  &domain.Canonico{Path: "/checkouts/x/provisional"},
	}
	if err := ep.pf.Upsert(entrada); err != nil {
		t.Fatal(err)
	}
	_, err := ep.svc.Publicar(context.Background(), entrada.Identidad.Clave())
	if !errors.Is(err, domain.ErrPublicarSinHome) {
		t.Fatalf("err = %v, want ErrPublicarSinHome", err)
	}
	if len(ep.pub.solicitudes) != 0 {
		t.Fatal("el publisher se llamó sin home")
	}
}

// Guarda 4 · BR-1 lado write: un home cuyo marketplace es `referencia` (o desconocido) JAMÁS
// recibe un push nuestro.
func TestPublicarNoPropio(t *testing.T) {
	ep := armarPublicar(t)
	ep.store.filas["prenter-marketplace"] = domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: repoPrenter, Clase: domain.ClaseReferencia,
	}
	_, err := ep.svc.Publicar(context.Background(), ep.clave)
	if !errors.Is(err, domain.ErrPublicarNoPropio) {
		t.Fatalf("err = %v, want ErrPublicarNoPropio", err)
	}

	// Home sin NINGÚN marketplace conocido: mismo centinela, con la pista de registrarlo.
	delete(ep.store.filas, "prenter-marketplace")
	_, err = ep.svc.Publicar(context.Background(), ep.clave)
	if !errors.Is(err, domain.ErrPublicarNoPropio) {
		t.Fatalf("err = %v, want ErrPublicarNoPropio (home sin marketplace)", err)
	}
	if len(ep.pub.solicitudes) != 0 {
		t.Fatal("el publisher se llamó contra un marketplace no propio")
	}
}

// Guarda 5 · la versión sale del canónico REAL y tiene que ser semver — "latest" o vacío no se
// publican (la SoT es plugin.json.version, jamás se fabrica).
func TestPublicarVersionNoSemver(t *testing.T) {
	ep := armarPublicar(t)
	for _, version := range []string{"latest", ""} {
		ep.loader.g = domain.Graph{Arnes: &domain.Arnes{ID: "harness", Version: version}}
		_, err := ep.svc.Publicar(context.Background(), ep.clave)
		if !errors.Is(err, domain.ErrPublicarVersionInvalida) {
			t.Fatalf("version %q: err = %v, want ErrPublicarVersionInvalida", version, err)
		}
	}
	if len(ep.pub.solicitudes) != 0 {
		t.Fatal("el publisher se llamó con versión inválida")
	}
}

// Guarda 6 · INNEGOCIABLE: el gate de conformance en rojo bloquea ANTES de tocar el remoto, y
// el reporte viaja adjunto (el operador ve QUÉ bloqueó, no solo «rojo»).
func TestPublicarConformanceRojoBloquea(t *testing.T) {
	ep := armarPublicar(t)
	ep.conf.rep = reporteRojo()

	_, err := ep.svc.Publicar(context.Background(), ep.clave)
	if !errors.Is(err, domain.ErrPublicarConformanceRojo) {
		t.Fatalf("err = %v, want ErrPublicarConformanceRojo", err)
	}
	var confErr *usecase.ErrorConformancePublicar
	if !errors.As(err, &confErr) {
		t.Fatalf("el reporte no viaja adjunto: %v", err)
	}
	if len(confErr.Reporte.Results) != 1 || confErr.Reporte.Results[0].Check.ID != "escritor-unico" {
		t.Fatalf("reporte = %+v", confErr.Reporte)
	}
	if len(ep.pub.solicitudes) != 0 {
		t.Fatal("el publisher se llamó con el gate rojo (innegociable roto)")
	}
}

// Un warn en fail NO bloquea (OK() solo mira severidad error) — el gate es el mismo del daemon.
func TestPublicarWarnNoBloquea(t *testing.T) {
	ep := armarPublicar(t)
	ep.conf.rep = domain.ConformanceReport{Results: []domain.CheckResult{
		{Check: domain.Check{ID: "olor", Severidad: domain.SevWarn}, Veredicto: domain.VeredictoFail},
	}}
	if _, err := ep.svc.Publicar(context.Background(), ep.clave); err != nil {
		t.Fatalf("un warn bloqueó el publish: %v", err)
	}
}

// Camino feliz · la solicitud lleva EXACTAMENTE la identidad del store + la versión del loader,
// y el post-éxito invalida el caché del catálogo y sube Canonico.Version.
func TestPublicarActualizaCanonicoEInvalidaCache(t *testing.T) {
	ep := armarPublicar(t)
	// Caché poblado del catálogo del marketplace destino: tiene que invalidarse tras publicar
	// (el estante cambió — servir el catálogo viejo diría «mi copia adelantada» de una versión
	// que ya está publicada).
	if err := ep.cache.Guardar("prenter-marketplace", catDosEntradas("prenter-marketplace", "2026-07-30T00:00:00Z", "local")); err != nil {
		t.Fatal(err)
	}

	res, err := ep.svc.Publicar(context.Background(), ep.clave)
	if err != nil {
		t.Fatalf("Publicar: %v", err)
	}
	if res.ID != "harness" || res.Version != "0.6.0" || res.Marketplace != "prenter-marketplace" ||
		res.Commit != "a1b2c3" || res.Tag != "harness/v0.6.0" || res.Clave != ep.clave {
		t.Fatalf("res = %+v", res)
	}
	if ep.conf.corrido != 1 {
		t.Fatalf("el gate corrió %d veces, want 1", ep.conf.corrido)
	}
	if len(ep.pub.solicitudes) != 1 {
		t.Fatalf("solicitudes = %d, want 1", len(ep.pub.solicitudes))
	}
	sol := ep.pub.solicitudes[0]
	want := domain.SolicitudPublicacion{
		RepoHome: repoPrenter, ID: "harness", Version: "0.6.0", OrigenDir: "/checkouts/prenter/harness",
	}
	if sol != want {
		t.Fatalf("solicitud = %+v, want %+v", sol, want)
	}

	if _, _, ok := ep.cache.Leer("prenter-marketplace"); ok {
		t.Fatal("el caché del catálogo NO se invalidó tras publicar")
	}
	entradas, _ := ep.pf.Listar()
	var hallada bool
	for _, e := range entradas {
		if e.Identidad.Clave() == ep.clave {
			hallada = true
			if e.Canonico == nil || e.Canonico.Version != "0.6.0" {
				t.Fatalf("Canonico.Version = %+v, want 0.6.0", e.Canonico)
			}
		}
	}
	if !hallada {
		t.Fatal("la entrada desapareció del store")
	}
}

// El error del publisher pasa TAL CUAL (los centinelas del adapter llegan al transporte sin
// re-envolver en genéricos).
func TestPublicarPropagaCentinelasDelPublisher(t *testing.T) {
	for _, centinela := range []error{
		domain.ErrPublicarVersionYaPublicada,
		domain.ErrPublicarPushRechazado,
		domain.ErrPublicarSinAuth,
	} {
		ep := armarPublicar(t)
		ep.pub.err = centinela
		_, err := ep.svc.Publicar(context.Background(), ep.clave)
		if !errors.Is(err, centinela) {
			t.Fatalf("err = %v, want %v", err, centinela)
		}
		// Un fallo del publisher NO toca el registro local: la versión del canónico queda.
		entradas, _ := ep.pf.Listar()
		for _, e := range entradas {
			if e.Identidad.Clave() == ep.clave && e.Canonico.Version != "0.5.3" {
				t.Fatalf("Canonico.Version = %q tras un fallo, want 0.5.3 intacta", e.Canonico.Version)
			}
		}
	}
}

// El mensaje del error de conformance nombra el gate y las cuentas (fail/error) — lo que el
// operador lee en el 409 sin abrir el reporte.
func TestErrorConformancePublicarMensaje(t *testing.T) {
	e := &usecase.ErrorConformancePublicar{Reporte: reporteRojo()}
	if !strings.Contains(e.Error(), "conformance") || !strings.Contains(e.Error(), "fail 1") {
		t.Fatalf("mensaje = %q", e.Error())
	}
}
