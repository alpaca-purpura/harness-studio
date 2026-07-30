package domain

import (
	"strings"
	"testing"
)

// Helpers de fixture: entradas del Portafolio con shape real (identidad provisional cruzada
// por registry es el caso DOMINANTE de esta máquina — design.md §6.2).

const mktPrenter = "github.com/alpacapurpura/prenter-marketplace"

func filaHarness(version string) EntradaCatalogo { //nolint:unparam // la firma toma la versión a propósito: los casos de E-20/E-21 la varían en su propio helper.
	return EntradaCatalogo{
		Nombre:  "harness",
		Source:  SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./plugins/harness/" + version, Ruta: "./plugins/harness/" + version},
		Version: version,
	}
}

func entradaConCanonico(id, versionCanonico string, insts ...Instalacion) EntradaPortafolio {
	return EntradaPortafolio{
		Identidad:     IdentidadArnes{Home: mktPrenter, ID: id},
		Canonico:      &Canonico{Path: "/home/x/.arnesia/checkouts/prenter-marketplace/" + id, Version: versionCanonico},
		Instalaciones: insts,
	}
}

func inst(path string, deriva EstadoDeriva, detalle, aviso string) Instalacion {
	return Instalacion{InstallPath: path, Tipo: InstReferenciadaCC, Deriva: deriva, DerivaDetalle: detalle, Aviso: aviso}
}

func coincidencias(e EntradaPortafolio, via string) []CoincidenciaPortafolio { //nolint:unparam // la vía es parte del contrato que se prueba; fijarla escondería el dato.
	return []CoincidenciaPortafolio{{Entrada: e, Via: via}}
}

// E-07 · BR-1: un marketplace `referencia` JAMÁS habilita una acción, en NINGUNA de las 6
// ramas, y el motivo es el literal firmado en el mockup. Enforcer del check
// `referencia-no-opera` del boundary marketplace-referencia-es-solo-procedencia.
func TestReferenciaNuncaHabilitaAccion(t *testing.T) {
	las6 := []TipoSituacion{
		SituacionNoLoTengo, SituacionAlHilo, SituacionMiCopiaAdelantada,
		SituacionEstanteAdelantado, SituacionEnDeriva, SituacionNoComparable,
	}
	for _, tipo := range las6 {
		a := AccionDeSituacion(SituacionCatalogo{Tipo: tipo, Motivo: "algo"}, ClaseReferencia)
		if a.Habilitada {
			t.Fatalf("situación %q en clase referencia quedó HABILITADA (BR-1 roto)", tipo)
		}
		if a.Motivo != MotivoSoloArnesesPropios {
			t.Fatalf("situación %q: Motivo = %q, want %q", tipo, a.Motivo, MotivoSoloArnesesPropios)
		}
	}
	// Una clase fuera del enum se trata como referencia (fail-safe, boundary L2 punto 2).
	if a := AccionDeSituacion(SituacionCatalogo{Tipo: SituacionNoLoTengo}, "tienda"); a.Habilitada {
		t.Fatal("una clase desconocida habilitó una acción (fail-safe roto)")
	}
	// Las DOS celdas habilitadas: propio × no-lo-tengo (AG-D17) y propio × mi-copia-adelantada
	// (B2, paquete 2026-07-30-volverlo-de-arnesia-y-publicar).
	a := AccionDeSituacion(SituacionCatalogo{Tipo: SituacionNoLoTengo}, ClasePropio)
	if !a.Habilitada || a.Verbo != AccionTraerCanonico || a.Motivo != "" {
		t.Fatalf("propio × no-lo-tengo = %+v, want {traer-canonico true \"\"}", a)
	}
	p := AccionDeSituacion(SituacionCatalogo{Tipo: SituacionMiCopiaAdelantada}, ClasePropio)
	if !p.Habilitada || p.Verbo != AccionPublicar || p.Motivo != "" {
		t.Fatalf("propio × mi-copia-adelantada = %+v, want {publicar true \"\"}", p)
	}
	// Y las otras filas siguen deshabilitadas (spec §0: el entregable es la situación).
	for _, c := range []struct {
		tipo  TipoSituacion
		verbo Accion
	}{
		{SituacionAlHilo, AccionNinguna},
		{SituacionEstanteAdelantado, AccionActualizarMiCopia},
		{SituacionEnDeriva, AccionReparar},
		{SituacionNoComparable, AccionNinguna},
	} {
		got := AccionDeSituacion(SituacionCatalogo{Tipo: c.tipo, Motivo: "m"}, ClasePropio)
		if got.Habilitada {
			t.Fatalf("propio × %q quedó habilitada (fuera de alcance, BR-10)", c.tipo)
		}
		if got.Verbo != c.verbo {
			t.Fatalf("propio × %q: Verbo = %q, want %q", c.tipo, got.Verbo, c.verbo)
		}
	}
	// El motivo de `no-comparable` se PROPAGA, no se duplica (§6.3).
	if got := AccionDeSituacion(SituacionCatalogo{Tipo: SituacionNoComparable, Motivo: "porque X"}, ClasePropio); got.Motivo != "porque X" {
		t.Fatalf("Motivo = %q, want la propagación de Situacion.Motivo", got.Motivo)
	}
}

// E-20 · canónico 0.5.4 vs estante 0.5.3 ⇒ mi-copia-adelantada + `Publicar` HABILITADO (B2,
// paquete 2026-07-30-volverlo-de-arnesia-y-publicar: el write-side existe).
func TestSituacionMiCopiaAdelantada(t *testing.T) {
	e := entradaConCanonico("harness", "0.5.4", inst("/i/1", DerivaAlHilo, "", ""))
	s := CalcularSituacion(filaHarness("0.5.3"), coincidencias(e, ViaHomeDeclarado))
	if s.Tipo != SituacionMiCopiaAdelantada || s.Mia != "0.5.4" || s.Estante != "0.5.3" {
		t.Fatalf("situación = %+v, want {mi-copia-adelantada 0.5.4 0.5.3}", s)
	}
	a := AccionDeSituacion(s, ClasePropio)
	want := AccionCatalogo{Verbo: AccionPublicar, Habilitada: true}
	if a != want {
		t.Fatalf("accion = %+v, want %+v", a, want)
	}
}

// E-21 · canónico 0.5.2 vs estante 0.5.3 ⇒ estante-adelantado + tooltip del ítem 4.
func TestSituacionEstanteAdelantado(t *testing.T) {
	e := entradaConCanonico("harness", "0.5.2", inst("/i/1", DerivaAlHilo, "", ""))
	s := CalcularSituacion(filaHarness("0.5.3"), coincidencias(e, ViaHomeDeclarado))
	if s.Tipo != SituacionEstanteAdelantado || s.Mia != "0.5.2" || s.Estante != "0.5.3" {
		t.Fatalf("situación = %+v, want {estante-adelantado 0.5.2 0.5.3}", s)
	}
	a := AccionDeSituacion(s, ClasePropio)
	want := AccionCatalogo{Verbo: AccionActualizarMiCopia, Motivo: "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)"}
	if a != want {
		t.Fatalf("accion = %+v, want %+v", a, want)
	}
}

// E-22 · la deriva GANA a la divergencia de versión (precedencia §6.1 fila 5): es un hecho duro
// de hash y es lo más accionable.
func TestSituacionEnDerivaGanaAVersion(t *testing.T) {
	e := entradaConCanonico("harness", "0.5.2",
		inst("/i/1", DerivaAlHilo, "", ""),
		inst("/i/2", DerivaEnDeriva, "hash de contenido distinto de la referencia /ref", ""),
	)
	s := CalcularSituacion(filaHarness("0.5.3"), coincidencias(e, ViaHomeDeclarado))
	if s.Tipo != SituacionEnDeriva {
		t.Fatalf("Tipo = %q, want instalaciones-en-deriva (la deriva gana a la versión)", s.Tipo)
	}
	if s.Cuantas != 1 {
		t.Fatalf("Cuantas = %d, want 1", s.Cuantas)
	}
	a := AccionDeSituacion(s, ClasePropio)
	want := AccionCatalogo{Verbo: AccionReparar, Motivo: "Reparar se construye en su propio paquete (ítem 5 del outcome)"}
	if a != want {
		t.Fatalf("accion = %+v, want %+v", a, want)
	}
}

// E-23 · sin versión del estante ⇒ no-comparable con motivo (BR-9), acción sin verbo.
func TestSituacionNoComparableSinVersionEstante(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre: "42crunch-api-security-testing",
		Source: SourceCatalogo{Tipo: SourceGitSubdir, Crudo: "git-subdir:x#p@v1.5.5", Ruta: "plugins/api-security-testing", Ref: "v1.5.5", SHA: "30287f5e"},
	}
	e := entradaConCanonico("42crunch-api-security-testing", "0.5.2")
	s := CalcularSituacion(fila, coincidencias(e, ViaHomeDeclarado))
	const want = "el catálogo no declara versión de esta entrada ni se puede derivar de su source"
	if s.Tipo != SituacionNoComparable || s.Motivo != want {
		t.Fatalf("situación = %+v, want no-comparable con motivo %q", s, want)
	}
	if a := AccionDeSituacion(s, ClasePropio); a.Verbo != AccionNinguna {
		t.Fatalf("Verbo = %q, want vacío", a.Verbo)
	}
}

// E-43 · versión no-semver ⇒ no-comparable citando las DOS versiones. Jamás una comparación
// de strings.
func TestVersionNoSemverEsNoComparable(t *testing.T) {
	fila := filaHarness("0.5.3")
	fila.Version, fila.VersionDe = "latest", VersionDeCampo
	e := entradaConCanonico("harness", "0.5.2")
	s := CalcularSituacion(fila, coincidencias(e, ViaHomeDeclarado))
	const want = "versiones no comparables (no-semver): «0.5.2» vs «latest»"
	if s.Tipo != SituacionNoComparable || s.Motivo != want {
		t.Fatalf("situación = %+v, want no-comparable con motivo %q", s, want)
	}
}

// TestSituacionNoComparableGana — enforcer del check
// `situacion-no-comparable-de-primera-clase` (boundary v1.2 §8): falta CUALQUIER insumo ⇒
// no-comparable con motivo, y `al-hilo` es la ÚLTIMA rama de la tabla.
func TestSituacionNoComparableGana(t *testing.T) {
	casos := []struct {
		nombre string
		fila   EntradaCatalogo
		e      EntradaPortafolio
		motivo string
	}{
		{
			"source no reconocido",
			EntradaCatalogo{Nombre: "harness", Source: SourceCatalogo{Tipo: SourceDesconocido, Crudo: "42"}, Version: "0.5.3"},
			entradaConCanonico("harness", "0.5.3"),
			"el catálogo declara un `source` que no reconozco: 42",
		},
		{
			"nombre duplicado en el catálogo",
			EntradaCatalogo{Nombre: "harness", Source: SourceCatalogo{Tipo: SourceRutaRelativa, Crudo: "./x"}, Version: "0.5.3", Aviso: []string{AvisoNombreDuplicado + "harness"}},
			entradaConCanonico("harness", "0.5.3"),
			AvisoNombreDuplicado + "harness",
		},
		{
			"sin canónico",
			filaHarness("0.5.3"),
			EntradaPortafolio{Identidad: IdentidadArnes{Home: mktPrenter, ID: "harness"}, Instalaciones: []Instalacion{inst("/i/1", DerivaAlHilo, "", "")}},
			"no tenés canónico de este arnés (solo instalaciones read-only): no hay copia editable que comparar contra el estante",
		},
		{
			"canónico sin versión",
			filaHarness("0.5.3"),
			entradaConCanonico("harness", ""),
			"tu canónico no declara versión",
		},
		{
			"deriva no evaluable con versión igual",
			filaHarness("0.5.3"),
			entradaConCanonico("harness", "0.5.3", inst("/i/1", DerivaNoEvaluable, "sin referencia local accesible", "")),
			"versión igual al estante, pero 1 instalación(es) con deriva no evaluable: /i/1: sin referencia local accesible",
		},
		{
			"aviso de instalación con versión igual",
			filaHarness("0.5.3"),
			entradaConCanonico("harness", "0.5.3", inst("/i/1", DerivaAlHilo, "", "no-reconocible: /i/1 sin .claude-plugin/plugin.json")),
			"versión igual al estante, pero 1 instalación(es) con aviso: no-reconocible: /i/1 sin .claude-plugin/plugin.json",
		},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s := CalcularSituacion(c.fila, coincidencias(c.e, ViaHomeDeclarado))
			if s.Tipo != SituacionNoComparable {
				t.Fatalf("Tipo = %q, want no-comparable", s.Tipo)
			}
			if s.Motivo != c.motivo {
				t.Fatalf("Motivo = %q, want %q", s.Motivo, c.motivo)
			}
		})
	}

	// `al-hilo` SOLO cuando ningún insumo falta y ninguna instalación tiene señal.
	sano := CalcularSituacion(filaHarness("0.5.3"), coincidencias(
		entradaConCanonico("harness", "0.5.3", inst("/i/1", DerivaAlHilo, "", "")), ViaHomeDeclarado))
	if sano.Tipo != SituacionAlHilo {
		t.Fatalf("Tipo = %q, want al-hilo (todos los insumos presentes)", sano.Tipo)
	}
	// Y sin instalaciones también es legítimo (solo canónico).
	soloCanonico := CalcularSituacion(filaHarness("0.5.3"), coincidencias(entradaConCanonico("harness", "0.5.3"), ViaHomeDeclarado))
	if soloCanonico.Tipo != SituacionAlHilo {
		t.Fatalf("Tipo = %q, want al-hilo (canónico igual, cero instalaciones)", soloCanonico.Tipo)
	}
}

// E-39 · una forma de `source` no reconocida es VISIBLE (crudo en el motivo), no un descarte.
func TestSourceDesconocidoEsVisible(t *testing.T) {
	fila := EntradaCatalogo{
		Nombre:  "raro",
		Source:  SourceCatalogo{Tipo: SourceDesconocido, Crudo: "42"},
		Version: "0.5.3",
		Aviso:   []string{"el catálogo declara un `source` con una forma que no reconozco: 42"},
	}
	s := CalcularSituacion(fila, coincidencias(entradaConCanonico("raro", "0.5.3"), ViaHomeDeclarado))
	if s.Tipo != SituacionNoComparable || !strings.Contains(s.Motivo, "42") {
		t.Fatalf("situación = %+v, want no-comparable con el crudo visible", s)
	}
}

// E-64 · Portafolio vacío ⇒ TODO `no-lo-tengo`; cero `al-hilo`. `al-hilo` no es el default.
func TestPortafolioVacioTodoNoLoTengo(t *testing.T) {
	for i := range 273 {
		fila := EntradaCatalogo{Nombre: "p" + string(rune('a'+i%26)), Version: "1.0.0"}
		s := CalcularSituacion(fila, CruzarConPortafolio(mktPrenter, fila, nil))
		if s.Tipo != SituacionNoLoTengo {
			t.Fatalf("con Portafolio vacío la fila %d dio %q, want no-lo-tengo", i, s.Tipo)
		}
	}
}

// E-65 · N coincidencias ⇒ no-comparable citando LAS DOS claves; no se elige ninguna (C-ID-2).
func TestNCoincidenciasEsNoComparable(t *testing.T) {
	a := EntradaPortafolio{Identidad: IdentidadArnes{ID: "harness", Scope: "a"}, Registries: []string{mktPrenter}}
	b := EntradaPortafolio{Identidad: IdentidadArnes{ID: "harness", Scope: "b"}, Registries: []string{mktPrenter}}
	fila := filaHarness("0.5.3")
	coin := CruzarConPortafolio(mktPrenter, fila, []EntradaPortafolio{a, b})
	if len(coin) != 2 {
		t.Fatalf("len(coincidencias) = %d, want 2", len(coin))
	}
	s := CalcularSituacion(fila, coin)
	if s.Tipo != SituacionNoComparable {
		t.Fatalf("Tipo = %q, want no-comparable", s.Tipo)
	}
	for _, clave := range []string{a.Identidad.Clave(), b.Identidad.Clave()} {
		if !strings.Contains(s.Motivo, clave) {
			t.Fatalf("el motivo debe citar la clave %q: %q", clave, s.Motivo)
		}
	}
	if s.ClavePortafolio != "" {
		t.Fatalf("con N coincidencias no se elige ninguna, got ClavePortafolio=%q", s.ClavePortafolio)
	}
}

// E-66 · el cruce DÉBIL por faceta `registry` es el caso REAL dominante: sin él, toda la
// columna de situación diría `no-lo-tengo` (design.md §6.2).
func TestCruceFacetaRegistry(t *testing.T) {
	// Caso real: harness@prenter-marketplace instalado, dir sin arnes.l0.json ⇒ identidad
	// provisional, pero origen.Registry SÍ resolvió el repo vía cc-plugins.
	e := EntradaPortafolio{
		Identidad:  IdentidadArnes{ID: "harness", Scope: ""},
		Registries: []string{"github.com/alpacapurpura/prenter-marketplace"},
	}
	fila := filaHarness("0.5.3")
	coin := CruzarConPortafolio(mktPrenter, fila, []EntradaPortafolio{e})
	if len(coin) != 1 {
		t.Fatalf("len(coincidencias) = %d, want 1 (cruce por registry)", len(coin))
	}
	if coin[0].Via != ViaFacetaRegistry {
		t.Fatalf("Via = %q, want %q", coin[0].Via, ViaFacetaRegistry)
	}
	s := CalcularSituacion(fila, coin)
	if s.Via != ViaFacetaRegistry || s.ClavePortafolio != e.Identidad.Clave() {
		t.Fatalf("la situación debe propagar la vía y la clave: %+v", s)
	}

	// El registry también se toma de las instalaciones (unión de facetas, S1-D3).
	e2 := EntradaPortafolio{
		Identidad:     IdentidadArnes{ID: "harness"},
		Instalaciones: []Instalacion{{InstallPath: "/i/1", Origen: OrigenPortafolio{Registry: "alpacapurpura/prenter-marketplace"}}},
	}
	if got := CruzarConPortafolio(mktPrenter, fila, []EntradaPortafolio{e2}); len(got) != 1 {
		t.Fatalf("el registry de una instalación (crudo `owner/repo`) debe cruzar canonicalizado, got %d", len(got))
	}
}

// E-106 · el canónico traído cruza con SU PROPIA fila del catálogo: es la prueba de que la
// identidad usa el repo canonicalizado y NO el nombre del marketplace (C17 de design.md).
func TestCanonicoTraidoCruzaPorHome(t *testing.T) {
	traido := entradaConCanonico("harness", "0.5.3")
	traido.Registries = []string{mktPrenter}
	fila := filaHarness("0.5.3")
	coin := CruzarConPortafolio(mktPrenter, fila, []EntradaPortafolio{traido})
	if len(coin) != 1 || coin[0].Via != ViaHomeDeclarado {
		t.Fatalf("coincidencias = %+v, want 1 por home-declarado", coin)
	}
	if s := CalcularSituacion(fila, coin); s.Tipo != SituacionAlHilo {
		t.Fatalf("Tipo = %q, want al-hilo tras traer", s.Tipo)
	}

	// Contraprueba de C17: registrado con Home = NOMBRE del marketplace (no el repo), la misma
	// fila daría `no-lo-tengo` sobre algo recién traído — el bug que E-106 vigila.
	malRegistrado := EntradaPortafolio{
		Identidad: IdentidadArnes{Home: "prenter-marketplace", ID: "harness"},
		Canonico:  &Canonico{Path: "/x", Version: "0.5.3"},
	}
	if got := CruzarConPortafolio(mktPrenter, fila, []EntradaPortafolio{malRegistrado}); len(got) != 0 {
		t.Fatalf("un Home por nombre NO debería cruzar (si cruza, el test no vigila nada): %+v", got)
	}
}

// RegistriesDe canonicaliza los dos lados y conserva el crudo que no canonicaliza (S1-D3).
func TestRegistriesDeUneYConservaCrudo(t *testing.T) {
	e := EntradaPortafolio{
		Registries: []string{"alpacapurpura/prenter-marketplace", "no-es-un-repo"},
		Instalaciones: []Instalacion{
			{Origen: OrigenPortafolio{Registry: "https://github.com/alpacapurpura/prenter-marketplace.git"}},
			{Origen: OrigenPortafolio{Registry: ""}},
		},
	}
	got := RegistriesDe(e)
	want := []string{"github.com/alpacapurpura/prenter-marketplace", "no-es-un-repo"}
	if len(got) != len(want) {
		t.Fatalf("RegistriesDe = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("RegistriesDe = %v, want %v", got, want)
		}
	}
}
