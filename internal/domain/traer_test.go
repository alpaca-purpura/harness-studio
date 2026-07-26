package domain

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func mktPropio(installLocation string) MarketplaceConocido {
	return MarketplaceConocido{
		Nombre:          "prenter-marketplace",
		Repo:            "github.com/alpacapurpura/prenter-marketplace",
		Clase:           ClasePropio,
		InstallLocation: installLocation,
	}
}

// Camino A · el plan del caso real de prenter: copia de subcarpeta, sin red, con la identidad
// keyeada por el REPO canonicalizado (no por el nombre del marketplace — C17).
func TestPlanificarTraerCaminoLocal(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre:  "harness",
		Source:  SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./plugins/harness/0.5.3", Ruta: "./plugins/harness/0.5.3"},
		Version: "0.5.3",
	}
	plan, err := PlanificarTraer(mktPropio("/home/x/.claude/plugins/marketplaces/prenter-marketplace"), fila, true, "/home/x/.arnesia")
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	if plan.Camino != CaminoLocal {
		t.Fatalf("Camino = %q, want local", plan.Camino)
	}
	want := "/home/x/.arnesia/checkouts/prenter-marketplace/harness"
	if plan.Destino != want {
		t.Fatalf("Destino = %q, want %q", plan.Destino, want)
	}
	if plan.OrigenLocal != "/home/x/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.3" {
		t.Fatalf("OrigenLocal = %q", plan.OrigenLocal)
	}
	if plan.Identidad.Home != "github.com/alpacapurpura/prenter-marketplace" || plan.Identidad.ID != "harness" {
		t.Fatalf("Identidad = %+v, want home = el REPO canonicalizado (C17)", plan.Identidad)
	}
	if plan.RaizDeMarketplace {
		t.Fatal("RaizDeMarketplace = true con una subruta declarada")
	}
	if plan.SHAEsperado != "" || plan.URL != "" {
		t.Fatalf("el camino local no pinea nada ni arma url: %+v", plan)
	}
}

// E-91 (lado plan) · `source: "./"` (caveman/ponytail lo usan de verdad) ⇒ el arnés ES la raíz del
// marketplace: se materializa igual, con el aviso VISIBLE.
func TestPlanificarTraerRaizDeMarketplace(t *testing.T) {
	fila := EntradaCatalogo{Nombre: "caveman", Source: SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./", Ruta: "./"}}
	mkt := mktPropio("/checkouts/caveman")
	mkt.Nombre, mkt.Repo = "caveman", "github.com/juliusbrussee/caveman"
	plan, err := PlanificarTraer(mkt, fila, true, "/raiz")
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	if !plan.RaizDeMarketplace {
		t.Fatal("RaizDeMarketplace = false con source «./»")
	}
	if plan.OrigenLocal != "/checkouts/caveman" {
		t.Fatalf("OrigenLocal = %q, want la raíz del checkout", plan.OrigenLocal)
	}
	if len(plan.Avisos) != 1 || !strings.Contains(plan.Avisos[0], `source "./"`) {
		t.Fatalf("Avisos = %v, want el literal de §13.2", plan.Avisos)
	}
}

// Camino B · el plan de una entrada real `git-subdir` del catálogo oficial: pin por SHA, subruta
// extraída, url armada.
func TestPlanificarTraerCaminoExterno(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre: "42crunch-api-security-testing",
		Source: SourceCatalogo{
			Tipo: SourceGitSubdir, Crudo: "git-subdir:x", URL: "https://github.com/42Crunch-AI/claude-plugins.git",
			Ruta: "plugins/api-security-testing", Ref: "v1.5.5", SHA: "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
		},
	}
	mkt := mktPropio("")
	plan, err := PlanificarTraer(mkt, fila, false, "/raiz")
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	if plan.Camino != CaminoExterno {
		t.Fatalf("Camino = %q, want externo", plan.Camino)
	}
	if plan.SHAEsperado != "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad" {
		t.Fatalf("SHAEsperado = %q", plan.SHAEsperado)
	}
	if plan.Ref != "v1.5.5" {
		t.Fatalf("Ref = %q (pista, no autoridad)", plan.Ref)
	}
	if plan.Subruta != "plugins/api-security-testing" {
		t.Fatalf("Subruta = %q", plan.Subruta)
	}
	if plan.URL != "https://github.com/42Crunch-AI/claude-plugins.git" {
		t.Fatalf("URL = %q", plan.URL)
	}
}

// La forma `github` solo trae `repo`: la url se ARMA canonicalizando, no se inventa.
func TestPlanificarTraerFormaGitHubArmaURL(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre: "fullstory",
		Source: SourceCatalogo{
			Tipo: SourceGitHub, Crudo: "github:fullstorydev/fullstory-skills",
			Repo: "fullstorydev/fullstory-skills", SHA: "b20614e2d08d7a7c70775bb62b5af640f60b024b",
		},
	}
	plan, err := PlanificarTraer(mktPropio(""), fila, false, "/raiz")
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	if plan.URL != "https://github.com/fullstorydev/fullstory-skills.git" {
		t.Fatalf("URL = %q", plan.URL)
	}
}

// E-80 · clase `referencia` ⇒ rechazo en el DOMINIO con el literal firmado, y CERO I/O (la
// función es pura: no hay nada que instrumentar). 5º check del boundary.
func TestPlanificarTraerRechazaReferencia(t *testing.T) {
	mkt := MarketplaceConocido{
		Nombre: "claude-plugins-official", Repo: "github.com/anthropics/claude-plugins-official", Clase: ClaseReferencia,
		InstallLocation: "/checkout",
	}
	fila := EntradaCatalogo{Nombre: "frontend-design", Source: SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./plugins/frontend-design"}}
	_, err := PlanificarTraer(mkt, fila, true, "/raiz")
	if !errors.Is(err, ErrTraerClaseReferencia) {
		t.Fatalf("err = %v, want ErrTraerClaseReferencia", err)
	}
	if !strings.Contains(err.Error(), MotivoSoloArnesesPropios) {
		t.Fatalf("el motivo debe traer el literal firmado %q: %v", MotivoSoloArnesesPropios, err)
	}
	// Una clase FUERA del enum también se rechaza (fail-safe: degrada a referencia).
	mkt.Clase = "tienda"
	if _, err2 := PlanificarTraer(mkt, fila, true, "/raiz"); !errors.Is(err2, ErrTraerClaseReferencia) {
		t.Fatalf("clase desconocida: err = %v, want ErrTraerClaseReferencia", err2)
	}
}

// E-90 · las tres formas de source no materializable, cada una con el crudo VISIBLE y sin
// adivinar nada.
func TestPlanificarTraerSourceNoMaterializable(t *testing.T) {
	casos := []struct {
		nombre                string
		fila                  EntradaCatalogo
		installLocationExiste bool
		enElMotivo            string
	}{
		{
			"source desconocido",
			EntradaCatalogo{Nombre: "raro", Source: SourceCatalogo{Tipo: SourceDesconocido, Crudo: "42"}},
			true, "42",
		},
		{
			"objeto sin sha ni ref",
			EntradaCatalogo{Nombre: "x", Source: SourceCatalogo{
				Tipo: SourceURL, Crudo: "url:https://github.com/a/b.git", URL: "https://github.com/a/b.git",
			}},
			true, "sin `sha` ni `ref`",
		},
		{
			"ruta relativa sin checkout",
			EntradaCatalogo{Nombre: "x", Source: SourceCatalogo{
				Tipo: SourceRutaRelativa, Crudo: "./plugins/x", Ruta: "./plugins/x",
			}},
			false, "no está clonado en disco",
		},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := PlanificarTraer(mktPropio("/checkout"), c.fila, c.installLocationExiste, "/raiz")
			if !errors.Is(err, ErrTraerSourceNoMaterializable) {
				t.Fatalf("err = %v, want ErrTraerSourceNoMaterializable", err)
			}
			if !strings.Contains(err.Error(), c.enElMotivo) {
				t.Fatalf("motivo = %v, want que contenga %q", err, c.enElMotivo)
			}
		})
	}
}

// E-95 · `source: github` con DOS hashes distintos (dato real de fullstorydev/fullstory-skills):
// el pin es el `sha` y el aviso cita LOS DOS. No se elige en silencio (C18).
func TestGitHubDosHashesGanaSha(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre: "fullstory",
		Source: SourceCatalogo{
			Tipo: SourceGitHub, Crudo: "github:fullstorydev/fullstory-skills", Repo: "fullstorydev/fullstory-skills",
			Commit: "1ec5865e7ab1449f9a0859d164c4b6a8c53b6e2f", SHA: "b20614e2d08d7a7c70775bb62b5af640f60b024b",
		},
	}
	plan, err := PlanificarTraer(mktPropio(""), fila, false, "/raiz")
	if err != nil {
		t.Fatal(err)
	}
	if plan.SHAEsperado != "b20614e2d08d7a7c70775bb62b5af640f60b024b" {
		t.Fatalf("SHAEsperado = %q, want el `sha` (el único campo presente en las 220 formas objeto)", plan.SHAEsperado)
	}
	var citaLosDos bool
	for _, a := range plan.Avisos {
		if strings.Contains(a, "1ec5865e7ab1449f9a0859d164c4b6a8c53b6e2f") && strings.Contains(a, "b20614e2d08d7a7c70775bb62b5af640f60b024b") {
			citaLosDos = true
		}
	}
	if !citaLosDos {
		t.Fatalf("Avisos = %v, want uno que cite LOS DOS hashes", plan.Avisos)
	}
}

// E-96 · `ref` que no coincide con el `sha` (dato real de 42Crunch: v1.5.5 → faf53053… vs sha
// 30287f5e…): el pin es el `sha`, la divergencia es AVISO, NO aborto (C16 — abortar rechazaría
// una entrada sana).
func TestRefDivergenteEsAvisoNoAborto(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre: "42crunch-api-security-testing",
		Source: SourceCatalogo{
			Tipo: SourceGitSubdir, URL: "https://github.com/42Crunch-AI/claude-plugins.git",
			Ruta: "plugins/api-security-testing", Ref: "v1.5.5", SHA: "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
		},
	}
	plan, err := PlanificarTraer(mktPropio(""), fila, false, "/raiz")
	if err != nil {
		t.Fatalf("una divergencia ref/sha NO debe abortar: %v", err)
	}
	if plan.SHAEsperado != "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad" {
		t.Fatalf("SHAEsperado = %q, want el sha", plan.SHAEsperado)
	}
	var hayAviso bool
	for _, a := range plan.Avisos {
		if strings.Contains(a, "v1.5.5") && strings.Contains(a, "manda el sha") {
			hayAviso = true
		}
	}
	if !hayAviso {
		t.Fatalf("Avisos = %v, want el literal de divergencia ref/sha", plan.Avisos)
	}
}

// E-87 · BR-13 · ningún input puede producir un destino fuera de `<raiz>/checkouts`, y dos
// nombres distintos NUNCA dan el mismo destino.
func TestRutaCanonicoNoEscapa(t *testing.T) {
	const raiz = "/home/x/.arnesia"
	crudos := []string{"../../.claude", "..%2F..", "/etc", "a/../../b", "…", "mi mkt", "mi-mkt", "ñandú@v2", "prenter-marketplace"}

	vistos := map[string]string{}
	for _, home := range crudos {
		destino, ok := RutaCanonico(raiz, home, "harness")
		if !ok {
			t.Fatalf("RutaCanonico(%q) = ok=false; want un destino seguro (no un rechazo)", home)
		}
		if !DentroDeCheckouts(raiz, destino) {
			t.Fatalf("RutaCanonico(%q) = %q — ESCAPA de %s", home, destino, RaizCheckouts(raiz))
		}
		if !strings.HasPrefix(filepath.Clean(destino), filepath.Clean(RaizCheckouts(raiz))+string(filepath.Separator)) {
			t.Fatalf("RutaCanonico(%q) = %q sin el prefijo de checkouts", home, destino)
		}
		if strings.Contains(destino, "..") {
			t.Fatalf("RutaCanonico(%q) = %q contiene `..`", home, destino)
		}
		if otro, dup := vistos[destino]; dup {
			t.Fatalf("colisión de destino: %q y %q dan %q", otro, home, destino)
		}
		vistos[destino] = home
	}

	// Un nombre vacío NO produce un segmento vacío: se rechaza (sin nombre no hay destino).
	if _, ok := RutaCanonico(raiz, "", "harness"); ok {
		t.Fatal("RutaCanonico con nombre vacío devolvió ok=true")
	}
	if _, ok := RutaCanonico(raiz, "prenter-marketplace", ""); ok {
		t.Fatal("RutaCanonico con id vacío devolvió ok=true")
	}
	// El caso feliz es LEGIBLE (E-76 lo espera exacto).
	if d, _ := RutaCanonico(raiz, "prenter-marketplace", "harness"); d != raiz+"/checkouts/prenter-marketplace/harness" {
		t.Fatalf("destino del caso real = %q", d)
	}
}

// DentroDeCheckouts es una aserción sobre paths, no una expresión de deseo.
func TestDentroDeCheckouts(t *testing.T) {
	const raiz = "/r"
	casos := map[string]bool{
		"/r/checkouts":            true,
		"/r/checkouts/a":          true,
		"/r/checkouts/a/b":        true,
		"/r/checkouts/../otro":    false,
		"/r/tmp/x":                false,
		"/home/x/.claude/plugins": false,
		"/":                       false,
	}
	for p, want := range casos {
		if got := DentroDeCheckouts(raiz, p); got != want {
			t.Fatalf("DentroDeCheckouts(%q) = %v, want %v", p, got, want)
		}
	}
}

// El temporal vive bajo `<raiz>/tmp`, NUNCA en os.TempDir() (E-101, riesgo EXDEV).
func TestRaizTemporalesBajoArnesia(t *testing.T) {
	if got := RaizTemporales("/home/x/.arnesia"); got != "/home/x/.arnesia/tmp" {
		t.Fatalf("RaizTemporales = %q", got)
	}
}
