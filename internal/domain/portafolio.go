package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// IdentidadArnes es la clave del Portafolio (BR-1, S0-D del paquete Slice 0). Home = el
// marketplace autor-declarado, canonicalizado (RN-IDENT-1); Home=="" ⇒ identidad
// provisional (RN-IDENT-2) y Scope discrimina (install-path relativo al proyecto en
// local, o remote canonicalizado del proyecto en git).
type IdentidadArnes struct {
	Home  string `json:"home,omitempty"`
	ID    string `json:"id"`
	Scope string `json:"scope,omitempty"`
	// Disc es el desempate de última instancia (S1-D29): la huella de la ruta física
	// (HuellaPath) que se puebla SOLO cuando la identidad es totalmente anónima —sin home,
	// sin id, sin scope discriminante— para que dos proyectos crudos escaneados en su raíz
	// no colapsen a la misma clave "sin-home~~". Vacío en toda identidad ya discriminable.
	Disc string `json:"disc,omitempty"`
}

// Provisional reporta si i carece de home resuelto (RN-IDENT-2): sin marketplace
// autor-declarado no hay identidad estable — solo un scope que discrimina.
func (i IdentidadArnes) Provisional() bool {
	return i.Home == ""
}

// Clave es el slug estable de i, apto para URL y dedup: "<home-slug>~<id>~<scope-slug>".
// Cada segmento se reduce a [a-z0-9-_] (el resto colapsa a '-'); una identidad
// provisional usa el literal "sin-home" (RN-IDENT-2: "(⟂sin-home, id, scope)") en vez de
// un segmento vacío, para que la clave siga siendo legible y estable.
func (i IdentidadArnes) Clave() string {
	home := i.Home
	if home == "" {
		home = "sin-home"
	}
	base := slugPortafolio(home) + "~" + slugPortafolio(i.ID) + "~" + slugPortafolio(i.Scope)
	// Desempate por huella de path (S1-D29): una identidad totalmente anónima (sin home, sin
	// id, sin scope discriminante) colapsa a "sin-home~~" — idéntica para CUALQUIER proyecto
	// crudo escaneado en su raíz, lo que las pisa en silencio al hacer Upsert por clave.
	// Cuando hay huella de la ruta física, se anexa para que dos roots distintos nunca
	// compartan clave. Las claves ya discriminables NO cambian (Disc vacío) → cero migración.
	if i.Home == "" && slugPortafolio(i.ID) == "" && slugPortafolio(i.Scope) == "" && i.Disc != "" {
		return base + i.Disc
	}
	return base
}

// HuellaPath es la huella estable de una ruta física (S1-D29): sha256 del path Clean-eado,
// primeros 12 hex. Sirve de desempate de la clave de una identidad anónima (Clave) y —Slice
// 2 T2— de llave sintética del índice del Mapa para un arnés degradado sin id. El caller pasa
// la ruta YA canonicalizada (EvalSymlinks vive en el usecase, domain no hace I/O); acá solo
// se normaliza con Clean y se hashea. "" → "" (sin dir físico no hay huella).
func HuellaPath(abs string) string {
	if abs == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(filepath.Clean(abs)))
	return hex.EncodeToString(sum[:])[:12]
}

// Slug expone slugPortafolio al usecase (Identificar S1-D28: sluggear el id del sello, sea
// el dado por el usuario o el basename del install-path) — misma regla de segmento estable
// que usa Clave, para que un id sellado sea siempre una clave válida.
func Slug(s string) string { return slugPortafolio(s) }

// slugPortafolio reduce s a un segmento de clave estable: minúsculas, runs de caracteres
// fuera de [a-z0-9-_] colapsados a un solo '-', sin '-' en los bordes.
func slugPortafolio(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
			prevDash = false
		case !prevDash:
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

// EstadoDeriva es el veredicto de BR-4: nunca semver-string ni `git status`, siempre hash
// de contenido contra la referencia inmutable — o el default honesto cuando no hay
// referencia accesible (S0-D7).
type EstadoDeriva string

// Los tres veredictos posibles de deriva (BR-4): jamás un cuarto valor inventado.
const (
	DerivaAlHilo      EstadoDeriva = "al-hilo"
	DerivaEnDeriva    EstadoDeriva = "en-deriva"
	DerivaNoEvaluable EstadoDeriva = "deriva-no-evaluable"
)

// TipoInstalacion son las tres formas físicas que el walker distingue (S0-D2).
type TipoInstalacion string

// Las tres formas físicas que el walker distingue (S0-D2).
const (
	InstMaterializada     TipoInstalacion = "materializada"      // raíz/.claude/plugins/<id>/ (convención DevStudio, HS-12).
	InstReferenciadaCC    TipoInstalacion = "referenciada-cc"    // enabledPlugins:true → cache global de CC.
	InstProyectoInstalado TipoInstalacion = "proyecto-instalado" // el .claude/ poblado del proyecto mismo.
)

// EslabonOrigen es un dato crudo de la cadena de resolución de origen (spec §6) con su
// fuente anotada — trazabilidad honesta (BR-3): jamás se pierde de dónde salió un dato.
type EslabonOrigen struct {
	Fuente string `json:"fuente"` // manifiesto | lock-devstudio | cc-plugins | git-plugin | git-proyecto | no-legible
	Campo  string `json:"campo"`  // home | registry | version | empresa | canal | proyecto-remote
	Valor  string `json:"valor"`
}

// OrigenPortafolio es la reconciliación collect-all de una instalación (spec §6).
// Registry/Version vacíos son desconocido/`?` honestos, jamás inventados; Discrepancias
// nunca se resuelven en silencio (C-OR-6). Nombrado distinto del `domain.Origen` existente
// (nodo.origen L0: estandar/del-puesto, box.go) — mismo vocablo, conceptos distintos; el
// plan de implementación usaba `Origen` a secas, pero ese identificador ya está tomado.
type OrigenPortafolio struct {
	Registry      string          `json:"registry,omitempty"`
	Version       string          `json:"version,omitempty"`
	Eslabones     []EslabonOrigen `json:"eslabones,omitempty"`
	Discrepancias []string        `json:"discrepancias,omitempty"`
}

// rangoRegistry y rangoVersion son las tablas de autoridad de S0-D1: los records de
// install-time (lock DevStudio y CC installed_plugins/enabledPlugins) mandan a la par
// (cada uno autoritativo en su carril de gestión); el remote git del plugin es secundario;
// el manifiesto NUNCA da procedencia de la copia (solo declara `home` de autor). Rango
// menor = más autoridad; una fuente ausente de la tabla no participa del campo (p.ej.
// git-proyecto jamás aporta `registry`: BR-5).
var rangoRegistry = map[string]int{
	"lock-devstudio": 1,
	"cc-plugins":     1,
	"git-plugin":     2,
	"manifiesto":     3,
}

var rangoVersion = map[string]int{
	"lock-devstudio": 1,
	"cc-plugins":     1,
	"manifiesto":     2,
}

// ResolverOrigen aplica el orden de autoridad de S0-D1 sobre los eslabones crudos
// recolectados (collect-all, jamás "para en el primer hit"): registry = lock-devstudio ≙
// cc-plugins (install-time) > git-plugin > manifiesto (que en la práctica no lo aporta);
// version = lock/cc > manifiesto (plugin.json). Todo conflicto entre valores de un mismo
// campo, y toda discrepancia home(manifiesto)≠registry(adquisición), queda anotada en
// Discrepancias — nunca una elección silenciosa (C-OR-6).
func ResolverOrigen(eslabones []EslabonOrigen) OrigenPortafolio {
	o := OrigenPortafolio{Eslabones: eslabones}

	var disc []string
	o.Registry, disc = resolverCampo(eslabones, "registry", rangoRegistry, disc)
	o.Version, disc = resolverCampo(eslabones, "version", rangoVersion, disc)

	if home, hok := valorDe(eslabones, "home"); hok && o.Registry != "" {
		hc, _ := CanonicalizarRepo(home)
		rc, _ := CanonicalizarRepo(o.Registry)
		if hc != "" && rc != "" && hc != rc {
			disc = append(disc, fmt.Sprintf(
				"home declarado (%s) ≠ registry de adquisición (%s): normal en el modelo N:M, visible por trazabilidad",
				home, o.Registry))
		}
	}

	o.Discrepancias = disc
	return o
}

// resolverCampo elige el valor de mayor autoridad para campo entre los eslabones que lo
// declaran (según rango; una fuente sin entrada en rango no participa) y registra una
// discrepancia si hay más de un valor distinto en juego — la elección nunca es silenciosa.
func resolverCampo(eslabones []EslabonOrigen, campo string, rango map[string]int, disc []string) (string, []string) {
	var ganador string
	mejorRango := int(^uint(0) >> 1) // max int
	valores := map[string]bool{}
	for _, e := range eslabones {
		if e.Campo != campo || e.Valor == "" {
			continue
		}
		valores[e.Valor] = true
		r, ok := rango[e.Fuente]
		if !ok {
			continue
		}
		if r < mejorRango {
			mejorRango = r
			ganador = e.Valor
		}
	}
	if len(valores) > 1 {
		vs := make([]string, 0, len(valores))
		for v := range valores {
			vs = append(vs, v)
		}
		sort.Strings(vs)
		disc = append(disc, fmt.Sprintf("%s: valores en conflicto entre eslabones (%s)", campo, strings.Join(vs, " vs ")))
	}
	return ganador, disc
}

// valorDe devuelve el primer valor no vacío de campo entre eslabones.
func valorDe(eslabones []EslabonOrigen, campo string) (string, bool) {
	for _, e := range eslabones {
		if e.Campo == campo && e.Valor != "" {
			return e.Valor, true
		}
	}
	return "", false
}

// ResolverIdentidad deriva la clave (home,id) de un arnés cargado: home =
// CanonicalizarRepo(a.Marketplace) si resuelve; el id sigue la precedencia RN-IDENT-3
// (a.ID, que el loader ya llenó vía arnes.l0→plugin.json, gana sobre idFallback —
// típicamente el nombre del dir cuando ningún manifiesto trae id); una discrepancia entre
// ambos se anota como aviso, nunca se elige en silencio. Sin home resoluble, la identidad
// es provisional con scope (RN-IDENT-2): scopeRemote (proyecto-remote canonicalizado)
// manda sobre scopeLocal (install-path relativo) cuando ambos están disponibles.
func ResolverIdentidad(a *Arnes, idFallback, scopeLocal, scopeRemote, installPath string) (identidad IdentidadArnes, aviso string) {
	id := idFallback
	if a != nil && a.ID != "" {
		id = a.ID
		if idFallback != "" && idFallback != a.ID {
			aviso = fmt.Sprintf("id discrepante: %q (manifiesto) ≠ %q (fallback)", a.ID, idFallback)
		}
	}

	var home string
	if a != nil && a.Marketplace != "" {
		if canon, ok := CanonicalizarRepo(a.Marketplace); ok {
			home = canon
		}
	}
	if home == "" {
		scope := scopeRemote
		if scope == "" {
			scope = scopeLocal
		}
		ident := IdentidadArnes{ID: id, Scope: scope}
		// Anónima total (sin id ni scope que discrimine, p.ej. un proyecto crudo escaneado en
		// su raíz: scope "." → slug vacío) → huella de la ruta física como desempate (S1-D29),
		// para no colapsar a la clave degenerada "sin-home~~".
		if slugPortafolio(id) == "" && slugPortafolio(scope) == "" && installPath != "" {
			ident.Disc = HuellaPath(installPath)
		}
		return ident, aviso
	}
	return IdentidadArnes{Home: home, ID: id}, aviso
}

// Instalacion es la materialización read-only de un arnés en un proyecto (spec §2, N por
// identidad).
type Instalacion struct {
	ProyectoPath  string           `json:"proyecto_path"`
	InstallPath   string           `json:"install_path"` // canonicalizado (EvalSymlinks+Clean, C-N-3).
	Tipo          TipoInstalacion  `json:"tipo"`
	Origen        OrigenPortafolio `json:"origen"`
	Deriva        EstadoDeriva     `json:"deriva"`
	DerivaDetalle string           `json:"deriva_detalle,omitempty"`
	Aviso         string           `json:"aviso,omitempty"` // C-P-5 no-reconocible · C-P-14 declarada-ausente · C-P-9 no-encontrada.
}

// Canonico es la única copia editable de una identidad (spec §2, 0 o 1 por identidad).
// Slice 0 solo modela su forma; el clone llega en Slice 2 (S0-D10).
type Canonico struct {
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}

// HallazgoInstalacion es un hallazgo crudo del walker (D-DOM-3/6): el adapter
// `portafolio.Scanner` lo emite, el usecase `PortafolioService` lo consume (cargar vía
// loader → ResolverIdentidad/ResolverOrigen → persistir). Vive en domain (no en el
// adapter ni en ports) para no duplicar la definición entre ambos lados del puerto.
// Dir=="" es legal: una entrada de lock/CC declarada pero sin dir físico resoluble
// (C-P-14) sigue siendo un hallazgo VISIBLE vía Aviso, solo que no cargable.
type HallazgoInstalacion struct {
	Dir       string          // dir a cargar con el loader (forma-plugin o raíz de proyecto); "" si no resoluble.
	Tipo      TipoInstalacion // "" si Dir=="" (sin forma física que clasificar).
	Eslabones []EslabonOrigen // crudos: lock-devstudio | cc-plugins | git-plugin | git-proyecto | no-legible.
	Aviso     string          // C-P-5 no-reconocible · C-P-14 declarada-ausente · CC ilegible.
	// IDConocido es el id declarado por la fuente del hallazgo (el `id` de una fila del
	// lock DevStudio, la parte-id de una clave `enabledPlugins`) — best-effort: la fuente
	// SABE el id aunque el manifiesto físico falte o no cargue (C-P-14, S1-D26); si el
	// manifiesto sí carga y discrepa, ResolverIdentidad lo hace visible como aviso. "" si
	// ninguna fuente lo declara.
	IDConocido string
}

// EntradaCorrupta es una fila del store del Portafolio que no parseó como
// EntradaPortafolio (BR-11, S0-D5): se conserva CRUDA y visible, jamás se pierde ni
// impide el arranque (C-N-4).
type EntradaCorrupta struct {
	Raw    json.RawMessage `json:"raw"`
	Motivo string          `json:"motivo"`
}

// EntradaPortafolio es 1 card por identidad (BR-1): empresas y registries son facetas
// N:M, nunca dueños de la identidad.
type EntradaPortafolio struct {
	Identidad     IdentidadArnes `json:"identidad"`
	Nombre        string         `json:"nombre,omitempty"`
	Descripcion   string         `json:"descripcion,omitempty"`
	Empresas      []string       `json:"empresas,omitempty"`
	Registries    []string       `json:"registries,omitempty"`
	Canonico      *Canonico      `json:"canonico,omitempty"`
	Instalaciones []Instalacion  `json:"instalaciones,omitempty"`
	Agregado      string         `json:"agregado,omitempty"` // RFC3339.
}
