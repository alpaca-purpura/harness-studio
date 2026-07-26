package domain

import (
	"strings"
	"testing"
)

// marketplace_test.go — capa D del plan de pruebas del paquete
// `2026-07-23-portafolio-agregar-marketplace`. Tests COLOCADOS junto al código (patrón
// sancionado en la auditoría de carga 2026-07-14; el motor `arnesia conformance` sabe
// correrlos). Insumos = filas REALES de los marketplaces de la máquina, no invenciones.

// E-03 · dos entradas con el mismo `source` son CANALES del mismo arnés físico (AG-D11).
// Insumo: las 2 filas reales de prenter-marketplace, ambas `./plugins/harness/0.5.3`.
func TestAgruparCanalesMismoSource(t *testing.T) {
	src := SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./plugins/harness/0.5.3", Ruta: "./plugins/harness/0.5.3"}
	unico := SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./plugins/otro", Ruta: "./plugins/otro"}
	out := AgruparCanales([]EntradaCatalogo{
		{Nombre: "harness", Source: src},
		{Nombre: "harness-beta", Source: src},
		{Nombre: "solo", Source: unico},
	})

	if got := out[0].ComparteSourceCon; len(got) != 1 || got[0] != "harness-beta" {
		t.Fatalf("harness.ComparteSourceCon = %v, want [harness-beta]", got)
	}
	if got := out[1].ComparteSourceCon; len(got) != 1 || got[0] != "harness" {
		t.Fatalf("harness-beta.ComparteSourceCon = %v, want [harness]", got)
	}
	if out[2].ComparteSourceCon != nil {
		t.Fatalf("una fila con source único debe tener ComparteSourceCon nil, got %v", out[2].ComparteSourceCon)
	}
}

// E-04 · la versión se DERIVA del último segmento semver de la ruta (BR-2 punto 2).
func TestVersionDeEntradaDerivaDeSource(t *testing.T) {
	v, de := VersionDeEntrada("", SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./plugins/harness/0.5.3"})
	if v != "0.5.3" || de != VersionDeSource {
		t.Fatalf("VersionDeEntrada = (%q,%q), want (0.5.3, derivada-de-source)", v, de)
	}
}

// TestVersionCatalogoNuncaInventada — enforcer del check homónimo del boundary
// `portafolio-identidad-y-deriva-honesta` v1.2 (§5-§8): la versión sale del campo, o del
// último segmento semver de la ruta, o NO EXISTE. Y la comparación es semver o nada.
func TestVersionCatalogoNuncaInventada(t *testing.T) {
	casos := []struct {
		nombre    string
		declarada string
		src       SourceCatalogo
		wantV     string
		wantDe    ProcedenciaVersion
	}{
		// AG-D14: el campo estándar gana (14 entradas reales del catálogo oficial lo traen).
		{"campo gana a la ruta", "1.0.0", SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./plugins/clangd-lsp/0.9.0"}, "1.0.0", VersionDeCampo},
		{"campo no-semver se conserva", "latest", SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./plugins/x"}, "latest", VersionDeCampo},
		{"ruta sin semver ⇒ ausente", "", SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./plugins/agent-sdk-dev"}, "", VersionAusente},
		{"raíz del marketplace ⇒ ausente", "", SourceCatalogo{Tipo: SourceRutaRelativa, Ruta: "./"}, "", VersionAusente},
		// Un source objeto lleva ref/sha, que NO son semver: jamás una versión falsa (BR-2b).
		{"git-subdir sin version ⇒ ausente", "", SourceCatalogo{
			Tipo: SourceGitSubdir, Ruta: "plugins/api-security-testing", Ref: "v1.5.5",
			SHA: "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
		}, "", VersionAusente},
		{"version null explícito ⇒ ausente", "", SourceCatalogo{Tipo: SourceURL, URL: "https://github.com/get-convex/convex-backend-skill.git"}, "", VersionAusente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			v, de := VersionDeEntrada(c.declarada, c.src)
			if v != c.wantV || de != c.wantDe {
				t.Fatalf("VersionDeEntrada(%q,%+v) = (%q,%q), want (%q,%q)", c.declarada, c.src, v, de, c.wantV, c.wantDe)
			}
		})
	}

	// Comparación: semver o nada. NUNCA strings.
	if _, ok := CompararSemver("0.5.2", "latest"); ok {
		t.Fatal("CompararSemver aceptó un no-semver: eso habilitaría comparar strings")
	}
	if cmp, ok := CompararSemver("0.5.4", "0.5.3"); !ok || cmp != 1 {
		t.Fatalf("CompararSemver(0.5.4, 0.5.3) = (%d,%v), want (1,true)", cmp, ok)
	}
	if cmp, ok := CompararSemver("1.0.0-rc.1", "1.0.0"); !ok || cmp != -1 {
		t.Fatalf("CompararSemver(1.0.0-rc.1, 1.0.0) = (%d,%v), want (-1,true)", cmp, ok)
	}
	for _, mal := range []string{"", "1.2", "1.2.3.4", "x.y.z", "01a.2.3", "latest", "main"} {
		if EsSemver(mal) {
			t.Fatalf("EsSemver(%q) = true, want false", mal)
		}
	}
	for _, bien := range []string{"0.5.3", "1.0.0", "2.1.0", "v1.5.5", "1.0.0-rc.1", "1.0.0+build.7"} {
		if !EsSemver(bien) {
			t.Fatalf("EsSemver(%q) = false, want true", bien)
		}
	}
}

// E-24 · merge collect-all: un repo distinto entre eslabones se MUESTRA, no se elige en
// silencio (BR-8). Enforcer del check `registro-marketplaces-collect-all` (boundary v1.2).
func TestMergeMarketplacesAnotaDiscrepancia(t *testing.T) {
	out := MergeMarketplaces(
		[]MarketplaceConocido{{Nombre: "x", Repo: "github.com/a/b", InstallLocation: "/tmp/x", CCActualizado: "2026-07-10T00:36:43.459Z"}},
		[]MarketplaceConocido{{Nombre: "x", Repo: "github.com/c/d", Clase: ClasePropio, Registrado: "2026-07-25T14:02:11Z"}},
	)
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1 (merge por nombre)", len(out))
	}
	f := out[0]
	if f.Repo != "github.com/c/d" {
		t.Fatalf("Repo = %q, want github.com/c/d (el declarado manda)", f.Repo)
	}
	if len(f.Discrepancias) != 1 {
		t.Fatalf("len(Discrepancias) = %d, want 1: %v", len(f.Discrepancias), f.Discrepancias)
	}
	if !strings.Contains(f.Discrepancias[0], "github.com/a/b") || !strings.Contains(f.Discrepancias[0], "github.com/c/d") {
		t.Fatalf("la discrepancia debe citar LOS DOS valores crudos, got %q", f.Discrepancias[0])
	}
	if len(f.Eslabones) != 2 || f.Eslabones[0] != EslabonCCKnown || f.Eslabones[1] != EslabonDeclarado {
		t.Fatalf("Eslabones = %v, want [cc-known-marketplaces declarado-por-operador]", f.Eslabones)
	}
	if f.Clase != ClasePropio {
		t.Fatalf("Clase = %q, want propio (la declarada manda)", f.Clase)
	}
	if f.InstallLocation != "/tmp/x" || f.CCActualizado == "" {
		t.Fatalf("InstallLocation/CCActualizado deben venir SOLO del lado CC: %+v", f)
	}
	if f.Registrado != "2026-07-25T14:02:11Z" {
		t.Fatalf("Registrado = %q, want el del lado declarado", f.Registrado)
	}
}

// Un nombre que solo CC conoce NO es «propio»: fail-safe hacia el lado que no habilita operar.
func TestMergeSoloDetectadoEsReferencia(t *testing.T) {
	out := MergeMarketplaces([]MarketplaceConocido{{Nombre: "caveman", Repo: "JuliusBrussee/caveman"}}, nil)
	if len(out) != 1 || out[0].Clase != ClaseReferencia {
		t.Fatalf("un marketplace solo detectado debe ser `referencia`, got %+v", out)
	}
	if len(out[0].Eslabones) != 1 || out[0].Eslabones[0] != EslabonCCKnown {
		t.Fatalf("Eslabones = %v, want [cc-known-marketplaces]", out[0].Eslabones)
	}
}

// E-47 · una clase fuera del enum degrada a `referencia`, con la degradación VISIBLE.
func TestClaseDesconocidaDegradaAReferencia(t *testing.T) {
	if got := ClaseSegura("tienda"); got != ClaseReferencia {
		t.Fatalf("ClaseSegura(tienda) = %q, want referencia", got)
	}
	if ClaseValida("tienda") {
		t.Fatal("ClaseValida(tienda) = true, want false")
	}
	out := MergeMarketplaces(nil, []MarketplaceConocido{{Nombre: "x", Clase: "tienda"}})
	if len(out) != 1 || out[0].Clase != ClaseReferencia {
		t.Fatalf("la fila debe degradar a referencia, got %+v", out)
	}
	if len(out[0].Discrepancias) != 1 || !strings.Contains(out[0].Discrepancias[0], `clase desconocida "tienda": se trata como de referencia`) {
		t.Fatalf("la degradación debe quedar visible, got %v", out[0].Discrepancias)
	}
}

// E-67 · dos marketplaces con nombres distintos y el mismo repo NO se fusionan (la clave de
// merge es el nombre, el que CC usa para keyear `installed_plugins.json`) — pero la
// coincidencia se anota, nunca se resuelve sola.
func TestMismoRepoNombresDistintosNoMergean(t *testing.T) {
	out := MergeMarketplaces([]MarketplaceConocido{
		{Nombre: "a", Repo: "github.com/x/y"},
		{Nombre: "b", Repo: "https://github.com/x/y.git"},
	}, nil)
	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2 (no se fusionan por repo)", len(out))
	}
	for _, f := range out {
		if len(f.Discrepancias) != 1 || !strings.Contains(f.Discrepancias[0], "otro marketplace conocido apunta al mismo repo") {
			t.Fatalf("fila %q sin la discrepancia informativa: %v", f.Nombre, f.Discrepancias)
		}
	}
	if !strings.Contains(out[0].Discrepancias[0], "b") || !strings.Contains(out[1].Discrepancias[0], "a") {
		t.Fatalf("cada fila debe citar la otra: %v / %v", out[0].Discrepancias, out[1].Discrepancias)
	}
}

// E-38 · `renames` del catálogo cruza por el nombre viejo. Insumo REAL: el oficial declara
// `renames: {"convex-backend":"convex"}`.
func TestRenamesCruzaPorNombreAnterior(t *testing.T) {
	mkt := "github.com/anthropics/claude-plugins-official"
	entrada := EntradaPortafolio{
		Identidad:  IdentidadArnes{ID: "convex-backend"},
		Registries: []string{mkt},
	}

	conRename := EntradaCatalogo{Nombre: "convex", NombreAnterior: []string{"convex-backend"}}
	got := CruzarConPortafolio(mkt, conRename, []EntradaPortafolio{entrada})
	if len(got) != 1 {
		t.Fatalf("con renames debe encontrar 1 coincidencia, got %d", len(got))
	}
	if got[0].Via != ViaRename {
		t.Fatalf("Via = %q, want %q", got[0].Via, ViaRename)
	}

	// Sin `renames`, la MISMA entrada no cruza: la fila diría `no-lo-tengo`.
	sinRename := EntradaCatalogo{Nombre: "convex"}
	if got := CruzarConPortafolio(mkt, sinRename, []EntradaPortafolio{entrada}); len(got) != 0 {
		t.Fatalf("sin renames no debe cruzar, got %v", got)
	}
}

// CrudoDeSource es determinista y no colapsa dos objetos distintos (AG-D11: es la clave de
// agrupamiento de canales).
func TestCrudoDeSourceEsDeterministaYDiscrimina(t *testing.T) {
	a := CrudoDeSource(SourceGitSubdir, "https://github.com/42Crunch-AI/claude-plugins.git", "plugins/api-security-testing", "v1.5.5")
	b := CrudoDeSource(SourceGitSubdir, "https://github.com/42Crunch-AI/claude-plugins.git", "plugins/api-security-testing", "v1.5.5")
	c := CrudoDeSource(SourceGitSubdir, "https://github.com/42Crunch-AI/claude-plugins.git", "plugins/otro", "v1.5.5")
	if a != b {
		t.Fatalf("CrudoDeSource no es determinista: %q vs %q", a, b)
	}
	if a == c {
		t.Fatalf("dos paths distintos colapsaron al mismo crudo: %q", a)
	}
	if want := "git-subdir:https://github.com/42Crunch-AI/claude-plugins.git#plugins/api-security-testing@v1.5.5"; a != want {
		t.Fatalf("crudo = %q, want %q", a, want)
	}
}
