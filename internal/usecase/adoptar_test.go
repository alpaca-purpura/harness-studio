package usecase_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// ── DD-2/E-a · «Adoptar como canónico»: el camino a canónico de un plugin forjado local ──

type escenarioAdoptar struct {
	*escenarioTraer
	loader *fakeLoaderPub
	origen string
}

// armarAdoptar cablea Traer (staging/raiz/deriva) + el loader del sello. `origen` es el
// dir forjado del operador (existe en disco: Adoptar lo valida con Stat).
func armarAdoptar(t *testing.T) *escenarioAdoptar {
	t.Helper()
	esc := armarTraer(t)
	origen := t.TempDir()
	if err := os.WriteFile(filepath.Join(origen, "plugin.json"), []byte(`{"name":"developer-vitalia","version":"0.1.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := &fakeLoaderPub{g: domain.Graph{Arnes: &domain.Arnes{
		ID: "developer-vitalia", Nombre: "developer-vitalia", Version: "0.1.0",
		Marketplace: "github.com/alpacapurpura/prenter-marketplace",
	}}}
	esc.svc.SetPublicar(nil, loader, nil)
	return &escenarioAdoptar{escenarioTraer: esc, loader: loader, origen: origen}
}

// El camino feliz: el dir forjado termina como canónico bajo checkouts/<mkt>/<id>, la
// entrada queda registrada con canónico y la deriva viaja (BR-17). El marketplace sale del
// HOME DEL SELLO — sin nombrarlo en el request.
func TestAdoptarSellaCanonicoDesdeElSello(t *testing.T) {
	esc := armarAdoptar(t)

	res, err := esc.svc.Adoptar(context.Background(), esc.origen, "")
	if err != nil {
		t.Fatalf("Adoptar: %v", err)
	}
	want := filepath.Join(esc.raiz, "checkouts", "prenter-marketplace", "developer-vitalia")
	if res.Destino != want {
		t.Errorf("Destino = %q, want %q", res.Destino, want)
	}
	if fi, serr := os.Stat(res.Destino); serr != nil || !fi.IsDir() {
		t.Errorf("el canónico no quedó en disco: %v", serr)
	}
	if res.Entrada.Canonico == nil || res.Entrada.Canonico.Path != want || res.Entrada.Canonico.Version != "0.1.0" {
		t.Errorf("entrada.Canonico = %+v, want path+version del sello", res.Entrada.Canonico)
	}
	if res.Marketplace != "prenter-marketplace" || res.Clave == "" {
		t.Errorf("res = %+v, want marketplace del home del sello + clave", res)
	}
	if res.Deriva != domain.DerivaAlHilo {
		t.Errorf("Deriva = %q, want la del evaluador (BR-17)", res.Deriva)
	}
	// Y quedó PERSISTIDA: la próxima Publicar la encuentra por clave.
	if got, ok := esc.pf.entradas[res.Clave]; !ok || got.Canonico == nil {
		t.Errorf("la entrada no quedó en el store: %+v", got)
	}
}

// Idempotencia de identidad: una identidad con canónico registrado NO se pisa (ley
// anti-drift — una sola copia editable), y el segundo intento lo dice.
func TestAdoptarNoPisaCanonicoExistente(t *testing.T) {
	esc := armarAdoptar(t)
	if _, err := esc.svc.Adoptar(context.Background(), esc.origen, ""); err != nil {
		t.Fatalf("primera adopción: %v", err)
	}
	_, err := esc.svc.Adoptar(context.Background(), esc.origen, "")
	if !errors.Is(err, usecase.ErrAdoptarYaCanonico) {
		t.Fatalf("err = %v, want ErrAdoptarYaCanonico", err)
	}
}

// BR-1 lado write: adoptar hacia un marketplace de referencia se rechaza en el dominio.
func TestAdoptarRechazaMarketplaceAjeno(t *testing.T) {
	esc := armarAdoptar(t)
	esc.store.filas["ajeno"] = domain.MarketplaceConocido{
		Nombre: "ajeno", Repo: "github.com/x/ajeno", Clase: domain.ClaseReferencia,
	}
	esc.loader.g.Arnes.Marketplace = "" // sello mudo: el request nombra.
	_, err := esc.svc.Adoptar(context.Background(), esc.origen, "ajeno")
	if !errors.Is(err, usecase.ErrAdoptarNoPropio) {
		t.Fatalf("err = %v, want ErrAdoptarNoPropio", err)
	}
}

// Sello y request en desacuerdo ⇒ 400 discrepante, jamás adoptar «a donde caiga».
func TestAdoptarHomeDiscrepante(t *testing.T) {
	esc := armarAdoptar(t)
	esc.store.filas["otro"] = domain.MarketplaceConocido{
		Nombre: "otro", Repo: "github.com/alpacapurpura/otro", Clase: domain.ClasePropio,
	}
	_, err := esc.svc.Adoptar(context.Background(), esc.origen, "otro")
	if !errors.Is(err, usecase.ErrAdoptarHomeDiscrepante) {
		t.Fatalf("err = %v, want ErrAdoptarHomeDiscrepante", err)
	}
}

// Sin home en el sello y sin marketplace en el request no hay destino que inventar.
func TestAdoptarSinMarketplace(t *testing.T) {
	esc := armarAdoptar(t)
	esc.loader.g.Arnes.Marketplace = ""
	_, err := esc.svc.Adoptar(context.Background(), esc.origen, "")
	if !errors.Is(err, usecase.ErrAdoptarSinMarketplace) {
		t.Fatalf("err = %v, want ErrAdoptarSinMarketplace", err)
	}
}

// Una versión no-semver ADOPTA igual (el material crudo es el caso de uso) pero el aviso
// dice que Publicar la va a exigir — jamás sorpresa después.
func TestAdoptarVersionNoSemverAvisa(t *testing.T) {
	esc := armarAdoptar(t)
	esc.loader.g.Arnes.Version = "en-forja"
	res, err := esc.svc.Adoptar(context.Background(), esc.origen, "")
	if err != nil {
		t.Fatalf("Adoptar: %v", err)
	}
	var hallado bool
	for _, a := range res.Avisos {
		if strings.Contains(a, "no es semver") {
			hallado = true
		}
	}
	if !hallado {
		t.Errorf("Avisos = %v, want el aviso de versión no-semver", res.Avisos)
	}
}

// La entrada previa del proyecto (instalaciones del wizard) NO se pierde: adoptar solo le
// SELLA el canónico encima.
func TestAdoptarConservaEntradaExistente(t *testing.T) {
	esc := armarAdoptar(t)
	clave := domain.IdentidadArnes{Home: repoPrenter, ID: "developer-vitalia"}.Clave()
	esc.pf.entradas[clave] = domain.EntradaPortafolio{
		Identidad: domain.IdentidadArnes{Home: repoPrenter, ID: "developer-vitalia"},
		Nombre:    "developer-vitalia",
		Instalaciones: []domain.Instalacion{{
			ProyectoPath: "/proyectos/vitalia-app", InstallPath: "/proyectos/vitalia-app/.claude/plugins/developer-vitalia",
		}},
	}

	res, err := esc.svc.Adoptar(context.Background(), esc.origen, "")
	if err != nil {
		t.Fatalf("Adoptar: %v", err)
	}
	if len(res.Entrada.Instalaciones) != 1 {
		t.Errorf("las instalaciones previas se perdieron: %+v", res.Entrada)
	}
	if res.Entrada.Canonico == nil {
		t.Error("el canónico no quedó sellado sobre la entrada existente")
	}
}
