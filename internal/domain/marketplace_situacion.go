package domain

import (
	"fmt"
	"sort"
	"strings"
)

// marketplace_situacion.go es el ENTREGABLE CENTRAL del paquete (spec §0: «lo que este paquete
// SÍ entrega es el cálculo honesto de la situación»). Puro: sin I/O, sin reloj, sin transporte.
// El orden de los `case` de CalcularSituacion ES la precedencia (design.md §6.1) — reordenarlos
// cambia el veredicto.

// TipoSituacion son las 6 ramas de spec §4.3. La 6ta (`no-comparable`) es de PRIMERA CLASE:
// sin ella el sistema tendría que elegir entre mentir (al-hilo por defecto) o callarse (BR-9).
type TipoSituacion string

// Las 6 ramas de situación. `no-comparable` gana sobre toda afirmación positiva.
const (
	SituacionNoLoTengo         TipoSituacion = "no-lo-tengo"
	SituacionAlHilo            TipoSituacion = "al-hilo"
	SituacionMiCopiaAdelantada TipoSituacion = "mi-copia-adelantada"
	SituacionEstanteAdelantado TipoSituacion = "estante-adelantado"
	SituacionEnDeriva          TipoSituacion = "instalaciones-en-deriva"
	SituacionNoComparable      TipoSituacion = "no-comparable"
)

// Las tres vías de cruce catálogo × Portafolio, en orden de fuerza (design.md §6.2).
const (
	ViaHomeDeclarado  = "home-declarado"
	ViaFacetaRegistry = "faceta-registry"
	ViaRename         = "rename"
)

// SituacionCatalogo es el veredicto del cruce catálogo × Portafolio para UNA fila.
// Invariante: Tipo==SituacionNoComparable ⟹ Motivo≠"" (una rama muda no informa nada).
type SituacionCatalogo struct {
	Tipo    TipoSituacion `json:"tipo"`
	Mia     string        `json:"mia,omitempty"`     // versión del canónico propio.
	Estante string        `json:"estante,omitempty"` // versión de la fila del catálogo.
	Cuantas int           `json:"cuantas,omitempty"` // instalaciones en-deriva.
	Motivo  string        `json:"motivo,omitempty"`
	// ClavePortafolio es la Identidad.Clave() de la entrada cruzada ("" si no-lo-tengo) — el FE
	// la usa para el link «ver ficha» sin re-derivar el slug.
	ClavePortafolio string `json:"clave_portafolio,omitempty"`
	// Via dice CÓMO se cruzó (BR-3, trazabilidad): "home-declarado" (identidad.home == el repo
	// del marketplace) | "faceta-registry" (identidad provisional, pero el registry de la copia
	// resuelve a este marketplace) | "rename" (cruzó por nombre_anterior). La UI muestra el
	// cruce débil ("faceta-registry") como tal — no lo presenta como identidad resuelta.
	Via string `json:"via,omitempty"`
}

// Accion es el verbo que la fila del catálogo ofrece (AG-D8 decisión 6). El vocabulario ya
// existe: `traer-canonico` es el `↧ Traer canónico` del drawer (portafolio-drawer.tsx) —
// dos puertas, un acto, cero vocabulario nuevo (mockups/INDEX.md regla dura 4).
type Accion string

// Los 4 verbos + la ausencia de verbo. Solo `traer-canonico` se ejecuta en este paquete (AG-D17).
const (
	AccionNinguna           Accion = ""
	AccionTraerCanonico     Accion = "traer-canonico"
	AccionPublicar          Accion = "publicar"
	AccionActualizarMiCopia Accion = "actualizar-mi-copia"
	AccionReparar           Accion = "reparar"
)

// MotivoSoloArnesesPropios es el tooltip LITERAL de la clase `referencia` — el único par
// `disabled`+`title` que el mockup firmado ya traía correcto (C5 de design.md). Se conserva
// tal cual: es el enforcement de BR-1 hecho texto.
const MotivoSoloArnesesPropios = "no aplica: solo arneses propios"

// AccionCatalogo es el verbo + si está habilitado + POR QUÉ no. En ESTE paquete Habilitada es
// `true` en UNA sola celda (`propio` × `no-lo-tengo` ⇒ `↧ Traer canónico`, AG-D17): el resto
// del entregable es la situación honesta, no la ejecución. El motivo es el tooltip LITERAL —
// vive acá, en el dominio, para que exista un solo texto testeable y para que ningún widget
// pueda pintar un botón habilitado por su cuenta.
type AccionCatalogo struct {
	Verbo      Accion `json:"verbo"`
	Habilitada bool   `json:"habilitada"`
	Motivo     string `json:"motivo,omitempty"`
}

// CoincidenciaPortafolio es una entrada del Portafolio que cruzó con una fila del catálogo,
// con el CÓMO anotado.
type CoincidenciaPortafolio struct {
	Entrada EntradaPortafolio
	Via     string // home-declarado | faceta-registry | rename
}

// CalcularSituacion cruza UNA fila del catálogo contra las entradas del Portafolio que le
// corresponden. `coincidencias` viene de CruzarConPortafolio (0, 1 o N). Tabla de verdad y
// precedencia completas en design.md §6.1 — el orden de los `case` ES la precedencia.
func CalcularSituacion(fila EntradaCatalogo, coincidencias []CoincidenciaPortafolio) SituacionCatalogo {
	// Filas 1-4: integridad del insumo. Si no sabemos de qué entrada hablamos, cualquier
	// veredicto sería inventado.
	if len(coincidencias) == 0 {
		return SituacionCatalogo{Tipo: SituacionNoLoTengo}
	}
	if len(coincidencias) > 1 {
		claves := make([]string, 0, len(coincidencias))
		for _, c := range coincidencias {
			claves = append(claves, c.Entrada.Identidad.Clave())
		}
		sort.Strings(claves)
		return SituacionCatalogo{Tipo: SituacionNoComparable, Motivo: fmt.Sprintf(
			"%d entradas de tu portafolio coinciden con esta fila (%s): resolvé el origen para desambiguar",
			len(claves), strings.Join(claves, ", "))}
	}

	elegida := coincidencias[0]
	base := SituacionCatalogo{
		ClavePortafolio: elegida.Entrada.Identidad.Clave(),
		Via:             elegida.Via,
	}
	noComparable := func(motivo string) SituacionCatalogo {
		s := base
		s.Tipo = SituacionNoComparable
		s.Motivo = motivo
		return s
	}

	if fila.Source.Tipo == SourceDesconocido {
		return noComparable("el catálogo declara un `source` que no reconozco: " + fila.Source.Crudo)
	}
	if fila.TieneAvisoDuplicado() {
		return noComparable(AvisoNombreDuplicado + fila.Nombre)
	}

	insts := elegida.Entrada.Instalaciones

	// Fila 5: la deriva es un hecho DURO de hash (no le falta ningún insumo) y es lo más
	// accionable — gana a cualquier comparación de versión (spec §4.3).
	if n := contarDeriva(insts, DerivaEnDeriva); n > 0 {
		s := base
		s.Tipo = SituacionEnDeriva
		s.Cuantas = n
		return s
	}

	// Filas 6-9: falta un insumo para comparar ⇒ no-comparable, jamás una afirmación.
	vEstante := fila.Version
	if vEstante == "" {
		return noComparable("el catálogo no declara versión de esta entrada ni se puede derivar de su source")
	}
	if elegida.Entrada.Canonico == nil {
		return noComparable("no tenés canónico de este arnés (solo instalaciones read-only): no hay copia editable que comparar contra el estante")
	}
	vMia := elegida.Entrada.Canonico.Version
	if vMia == "" {
		return noComparable("tu canónico no declara versión")
	}
	cmp, ok := CompararSemver(vMia, vEstante)
	if !ok {
		return noComparable(fmt.Sprintf("versiones no comparables (no-semver): «%s» vs «%s»", vMia, vEstante))
	}

	// Filas 10-11: divergencia de versión, con los dos valores visibles.
	if cmp != 0 {
		s := base
		s.Mia, s.Estante = vMia, vEstante
		if cmp > 0 {
			s.Tipo = SituacionMiCopiaAdelantada
		} else {
			s.Tipo = SituacionEstanteAdelantado
		}
		return s
	}

	// Filas 12-13: versión igual, pero alguna instalación no permite afirmar nada.
	if n := contarDeriva(insts, DerivaNoEvaluable); n > 0 {
		return noComparable(fmt.Sprintf(
			"versión igual al estante, pero %d instalación(es) con deriva no evaluable: %s",
			n, strings.Join(detallesDeriva(insts, DerivaNoEvaluable), " · ")))
	}
	if avisos := avisosDeInstalaciones(insts); len(avisos) > 0 {
		return noComparable(fmt.Sprintf(
			"versión igual al estante, pero %d instalación(es) con aviso: %s",
			len(avisos), strings.Join(avisos, " · ")))
	}

	// Fila 14: la ÚNICA afirmación positiva del set, y es la última — solo se llega acá cuando
	// ningún insumo falta y ninguna instalación tiene señal.
	s := base
	s.Tipo = SituacionAlHilo
	s.Mia, s.Estante = vMia, vEstante
	return s
}

// contarDeriva cuenta las instalaciones con el veredicto de deriva pedido.
func contarDeriva(insts []Instalacion, estado EstadoDeriva) int {
	n := 0
	for _, i := range insts {
		if i.Deriva == estado {
			n++
		}
	}
	return n
}

// detallesDeriva junta los motivos de las instalaciones con el veredicto pedido (visibles
// completos, BR-8: un motivo truncado no informa).
func detallesDeriva(insts []Instalacion, estado EstadoDeriva) []string {
	var out []string
	for _, i := range insts {
		if i.Deriva != estado {
			continue
		}
		detalle := i.DerivaDetalle
		if detalle == "" {
			detalle = "sin motivo declarado"
		}
		out = append(out, i.InstallPath+": "+detalle)
	}
	return out
}

// avisosDeInstalaciones junta los avisos no vacíos de las instalaciones.
func avisosDeInstalaciones(insts []Instalacion) []string {
	var out []string
	for _, i := range insts {
		if i.Aviso != "" {
			out = append(out, i.Aviso)
		}
	}
	return out
}

// CruzarConPortafolio busca en `todas` las entradas que corresponden a (repoMarketplace, fila).
// Tres vías, en orden de fuerza (design.md §6.2):
//  1. home-declarado: identidad.Home == repoMarketplace ∧ identidad.ID == fila.Nombre.
//  2. faceta-registry: identidad.Home == "" (provisional) ∧ identidad.ID == fila.Nombre ∧
//     repoMarketplace ∈ RegistriesDe(entrada).
//  3. rename: idem 1 o 2 pero con identidad.ID ∈ fila.NombreAnterior.
//
// Devuelve TODAS las coincidencias — el caller nunca elige una en silencio (C-ID-2).
func CruzarConPortafolio(repoMarketplace string, fila EntradaCatalogo, todas []EntradaPortafolio) []CoincidenciaPortafolio {
	canonMkt, _ := CanonicalizarRepo(repoMarketplace)
	if canonMkt == "" {
		canonMkt = repoMarketplace
	}
	if canonMkt == "" {
		return nil
	}

	var out []CoincidenciaPortafolio
	for _, e := range todas {
		idCoincide := e.Identidad.ID != "" && e.Identidad.ID == fila.Nombre
		idEsRename := !idCoincide && e.Identidad.ID != "" && contieneStr(fila.NombreAnterior, e.Identidad.ID)
		if !idCoincide && !idEsRename {
			continue
		}

		var via string
		switch {
		case e.Identidad.Home != "" && mismoRepo(e.Identidad.Home, canonMkt):
			via = ViaHomeDeclarado
		case e.Identidad.Home == "" && contieneStr(RegistriesDe(e), canonMkt):
			via = ViaFacetaRegistry
		default:
			continue // ni el home ni el registry de la copia apuntan a este marketplace.
		}
		if idEsRename {
			via = ViaRename
		}
		out = append(out, CoincidenciaPortafolio{Entrada: e, Via: via})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Entrada.Identidad.Clave() < out[j].Entrada.Identidad.Clave()
	})
	return out
}

// mismoRepo compara dos referencias git canonicalizando LOS DOS lados (RN-IDENT-1) — el bug
// clásico de este cruce es comparar un `owner/repo` corto contra un `host/owner/repo`.
func mismoRepo(a, b string) bool {
	ca, oka := CanonicalizarRepo(a)
	cb, okb := CanonicalizarRepo(b)
	if oka && okb {
		return ca == cb
	}
	return a == b
}

// RegistriesDe es el equivalente Go del selector FE homónimo: unión de entrada.Registries ∪
// cada instalaciones[].Origen.Registry, canonicalizada y dedupeada, orden de primera aparición.
// Un crudo que no canonicaliza se conserva tal cual (S1-D3: visible, nunca descartado).
func RegistriesDe(e EntradaPortafolio) []string {
	seen := map[string]bool{}
	var out []string
	agregar := func(v string) {
		if v == "" {
			return
		}
		canon, ok := CanonicalizarRepo(v)
		if !ok {
			canon = v
		}
		if seen[canon] {
			return
		}
		seen[canon] = true
		out = append(out, canon)
	}
	for _, r := range e.Registries {
		agregar(r)
	}
	for _, i := range e.Instalaciones {
		agregar(i.Origen.Registry)
	}
	return out
}

// contieneStr reporta si v está en xs.
func contieneStr(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// AccionDeSituacion mapea (situación, clase) → verbo + habilitación + tooltip literal.
// ES EL PUNTO DE ENFORCEMENT de BR-1 y del boundary marketplace-referencia-es-solo-procedencia:
// para ClaseReferencia devuelve Habilitada:false en TODA rama, sin excepción.
func AccionDeSituacion(s SituacionCatalogo, clase ClaseMarketplace) AccionCatalogo {
	// Fail-safe: una clase fuera del enum se trata como `referencia` (nunca habilita operar).
	if ClaseSegura(clase) != ClasePropio {
		verbo := AccionNinguna
		if s.Tipo == SituacionNoLoTengo {
			verbo = AccionTraerCanonico // se PINTA el verbo, deshabilitado con su motivo (mockup firmado).
		}
		return AccionCatalogo{Verbo: verbo, Habilitada: false, Motivo: MotivoSoloArnesesPropios}
	}

	switch s.Tipo {
	case SituacionNoLoTengo:
		// La ÚNICA celda habilitada del paquete (AG-D17 FIRMADA 🧑‍⚖️, mecanismo en design.md §13):
		// Traer escribe SOLO en ~/.arnesia/checkouts/, a diferencia de los otros tres verbos.
		return AccionCatalogo{Verbo: AccionTraerCanonico, Habilitada: true}
	case SituacionMiCopiaAdelantada:
		// Habilitada desde B2 (paquete 2026-07-30-volverlo-de-arnesia-y-publicar): el write-side
		// existe — el drawer ejecuta; la celda del catálogo pinta lo que el dominio manda (B-D6).
		return AccionCatalogo{Verbo: AccionPublicar, Habilitada: true}
	case SituacionEstanteAdelantado:
		return AccionCatalogo{Verbo: AccionActualizarMiCopia, Motivo: "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)"}
	case SituacionEnDeriva:
		return AccionCatalogo{Verbo: AccionReparar, Motivo: "Reparar se construye en su propio paquete (ítem 5 del outcome)"}
	case SituacionNoComparable:
		// El motivo se PROPAGA, no se duplica: un solo texto, escrito una vez (§6.3).
		return AccionCatalogo{Verbo: AccionNinguna, Motivo: s.Motivo}
	case SituacionAlHilo:
		return AccionCatalogo{Verbo: AccionNinguna}
	default:
		return AccionCatalogo{Verbo: AccionNinguna}
	}
}
