package domain

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// marketplace.go modela el ESTANTE (AG-D8): los marketplaces que ArnesIA conoce y el
// catálogo que cada uno declara. Puro — sin I/O, sin reloj, sin transporte
// (`dominio-independiente-de-transporte`). El shape crudo del `marketplace.json` ajeno se
// traduce en `internal/adapters/marketplace/parse.go` y NO se adopta acá
// (boundary `marketplace-referencia-es-solo-procedencia`, L2 punto 4).

// Errores centinela de la LECTURA de un catálogo. Viven en el dominio (no en el usecase ni en
// el adapter) porque los produce el adapter y los clasifica el usecase, y `usecase` no puede
// importar un adapter ni al revés (go-arch-lint: los dos solo ven `domain`+`ports`). Son valores
// puros — `errors.New` es stdlib, cero transporte. El usecase los RE-EXPORTA con los nombres del
// contrato de design.md §3.5 (C23 de design.md).
var (
	// ErrNoEsMarketplace: el repo/dir existe pero no expone un `.claude-plugin/marketplace.json`
	// legible (ausente, JSON inválido, sin `name`). El handler lo mapea a 400 con motivo
	// explícito (E-17) — jamás un 500 genérico.
	ErrNoEsMarketplace = errors.New("marketplace: el repo no expone .claude-plugin/marketplace.json legible")
	// ErrSinViaDeLectura: no hay checkout local NI vía remota (gh ausente/no autenticado y sin
	// PAT), o el host no es github.com. DISTINTO de ErrNoEsMarketplace: no es que la url esté
	// mal, es que NO PUEDO MIRAR ⇒ 503, nunca 400 (E-54/E-59).
	ErrSinViaDeLectura = errors.New("marketplace: sin vía de lectura (ni checkout local ni gh/PAT)")
)

// ClaseMarketplace parte el universo en dos, y NO son simétricas (AG-D8 decisión 1):
// `propio` = publicamos ahí (participa de traer/publicar/actualizar/reparar);
// `referencia` = solo resuelve la procedencia de un arnés ajeno — catálogo read-only,
// jamás operable (boundary marketplace-referencia-es-solo-procedencia).
type ClaseMarketplace string

// Las dos clases conocidas. Una tercera no existe: lo desconocido degrada a `referencia`.
const (
	ClasePropio     ClaseMarketplace = "propio"
	ClaseReferencia ClaseMarketplace = "referencia"
)

// ClaseValida reporta si c es una de las dos clases conocidas. Una clase desconocida NUNCA
// degrada a `propio` (eso habilitaría operar lo ajeno): degrada a `referencia`, el lado seguro.
func ClaseValida(c ClaseMarketplace) bool {
	return c == ClasePropio || c == ClaseReferencia
}

// ClaseSegura devuelve c si es válida, y ClaseReferencia si no lo es — fail-safe explícito.
func ClaseSegura(c ClaseMarketplace) ClaseMarketplace {
	if ClaseValida(c) {
		return c
	}
	return ClaseReferencia
}

// EslabonMarketplace dice de dónde salió el CONOCIMIENTO de que este marketplace existe
// (AG-D9, collect-all: no exclusivo, ≥1, los dos = detectado + confirmado). Eje DISTINTO de
// EslabonOrigen.Fuente (portafolio.go), que dice de dónde salió un dato de UNA COPIA
// instalada; no unificar (C9 de design.md): son dos preguntas.
type EslabonMarketplace string

// Los dos eslabones reales de conocimiento de un marketplace (AG-D9).
const (
	// EslabonCCKnown: `~/.claude/plugins/known_marketplaces.json` — registro propio de Claude
	// Code. Da checkout local gratis (`installLocation`) y fecha de última actualización real.
	EslabonCCKnown EslabonMarketplace = "cc-known-marketplaces"
	// EslabonDeclarado: lo registró el operador por el wizard (S6).
	EslabonDeclarado EslabonMarketplace = "declarado-por-operador"
)

// TipoLectura son los 4 estados explícitos de AG-D8 decisión 4. `no-leido` es el CERO del tipo
// a propósito: un EstadoLectura sin poblar dice «no leído aún», nunca «leído con 0 entradas».
type TipoLectura string

// Los 4 estados de lectura de un catálogo (AG-D8 decisión 4).
const (
	LecturaNoLeida       TipoLectura = "no-leido"
	LecturaLeida         TipoLectura = "leido"
	LecturaSinAcceso     TipoLectura = "sin-acceso"
	LecturaURLNoResuelve TipoLectura = "url-no-resuelve"
)

// EstadoLectura es el veredicto de la última lectura de catálogo de un marketplace.
// Invariante (BR-4): Tipo==LecturaLeida ⟺ Entradas es una cuenta REAL y Cuando≠"".
// Cualquier otro Tipo EXIGE Motivo≠"" — un degradado sin motivo es un degradado mudo.
type EstadoLectura struct {
	Tipo TipoLectura `json:"tipo"`
	// Cuando es RFC3339 UTC de la última lectura EXITOSA (aunque Tipo ya no sea `leido`:
	// así la fila puede decir «leído hace 2 días · ahora sin acceso» — BR-3 + BR-4 juntas).
	Cuando string `json:"cuando,omitempty"`
	// Entradas es la cuenta de filas del catálogo de esa última lectura exitosa. Solo
	// significativa con Cuando≠"".
	Entradas int `json:"entradas,omitempty"`
	// Motivo es el texto que la UI muestra tal cual (gh no autenticado, 404, JSON inválido…).
	Motivo string `json:"motivo,omitempty"`
	// Fuente de esa lectura: "local" (checkout de CC) | "remoto" (gh api) | "" (nunca leído).
	Fuente string `json:"fuente,omitempty"`
}

// TipoSource son las formas reales que `plugins[].source` toma en los marketplaces de esta
// máquina (AG-D13, verificado sobre 273+2+1+1+1 entradas). El adapter traduce; el dominio
// solo ve esto normalizado.
type TipoSource string

// Las 4 formas reales de `plugins[].source` + el degradado visible (AG-D13).
const (
	SourceRutaRelativa TipoSource = "ruta-relativa" // "./plugins/harness/0.5.3", "./"
	SourceGitSubdir    TipoSource = "git-subdir"
	SourceURL          TipoSource = "url"
	SourceGitHub       TipoSource = "github"
	SourceDesconocido  TipoSource = "desconocido" // shape que el adapter no reconoció: VISIBLE, no descartado.
)

// SourceCatalogo es `plugins[].source` normalizado. Crudo es la representación textual estable
// que se usa para agrupar canales (AG-D11): para una ruta relativa es la ruta tal cual; para un
// objeto es "<tipo>:<url|repo>#<path>@<ref>" — determinista, comparable, y jamás se muestra como
// si fuera una ruta local.
type SourceCatalogo struct {
	Tipo  TipoSource `json:"tipo"`
	Crudo string     `json:"crudo"`
	Ruta  string     `json:"ruta,omitempty"` // ruta relativa, o `path` del git-subdir.
	URL   string     `json:"url,omitempty"`
	Repo  string     `json:"repo,omitempty"`
	Ref   string     `json:"ref,omitempty"`
	SHA   string     `json:"sha,omitempty"`
	// Commit es el 2º hash que SOLO la forma `github` declara, y con valor DISTINTO del `sha`
	// en las 2 entradas reales que la usan (C18/C19 de design.md, verificado en vivo:
	// fullstorydev/fullstory-skills y jfrog/claude-plugin). Se conserva normalizado para poder
	// citar los dos en el aviso de divergencia (E-95) — la autoridad es SIEMPRE el `sha`, que
	// es el único campo presente en las 220 formas objeto.
	Commit string `json:"commit,omitempty"`
}

// CrudoDeSource arma la representación textual estable de un source objeto:
// "<tipo>:<url|repo>#<path>@<ref>" — determinista y comparable (AG-D11: agrupar canales).
// Los segmentos ausentes se omiten con su separador, así dos objetos idénticos dan la misma
// cadena y dos distintos jamás colapsan.
func CrudoDeSource(tipo TipoSource, urlOrRepo, path, ref string) string {
	var b strings.Builder
	b.WriteString(string(tipo))
	b.WriteString(":")
	b.WriteString(urlOrRepo)
	if path != "" {
		b.WriteString("#")
		b.WriteString(path)
	}
	if ref != "" {
		b.WriteString("@")
		b.WriteString(ref)
	}
	return b.String()
}

// MarketplaceConocido es UNA fila del plano Marketplaces (S2), ya mergeada collect-all
// (AG-D9). La clave de merge es Nombre — el mismo nombre con el que Claude Code keyea
// `installed_plugins.json` (`"<id>@<nombre>"`), o sea el nombre INSTALABLE, no el repo.
type MarketplaceConocido struct {
	Nombre string `json:"nombre"`
	// Repo canonicalizado con domain.CanonicalizarRepo → "host/owner/repo" (RN-IDENT-1).
	// "" si ninguna fuente lo declara o ninguna canonicaliza (el crudo queda en Discrepancias).
	Repo  string           `json:"repo,omitempty"`
	Clase ClaseMarketplace `json:"clase"`
	// Eslabones ordenados y dedupeados (detectado antes que declarado), ≥1 siempre.
	Eslabones []EslabonMarketplace `json:"eslabones"`
	// InstallLocation es el checkout local que Claude Code ya mantiene — la ruta contra la que
	// `deriva` compara (`<InstallLocation>/plugins/<id>/<version>/`). Solo del eslabón CC.
	InstallLocation string `json:"install_location,omitempty"`
	// CCActualizado es el `lastUpdated` que CC declara (RFC3339). Dato de CC, NO nuestra lectura:
	// nunca se muestra como «leído hace…» (eso es Lectura.Cuando).
	CCActualizado string        `json:"cc_actualizado,omitempty"`
	Lectura       EstadoLectura `json:"lectura"`
	// Discrepancias entre eslabones (repo distinto, clase distinta): se MUESTRAN, jamás se
	// eligen en silencio (BR-8, mismo criterio que OrigenPortafolio.Discrepancias / C-OR-6).
	Discrepancias []string `json:"discrepancias,omitempty"`
	// Registrado es RFC3339 de cuándo el operador lo declaró; "" si solo lo detectó CC.
	Registrado string `json:"registrado,omitempty"`
}

// ProcedenciaVersion dice de DÓNDE salió Version — BR-2 con AG-D14: nunca se inventa, y
// siempre se puede auditar cuál de las dos fuentes ganó.
type ProcedenciaVersion string

// La precedencia de BR-2/AG-D14: campo estándar > último segmento semver de la ruta > ausente.
const (
	VersionDeCampo  ProcedenciaVersion = "campo-version"      // plugins[].version
	VersionDeSource ProcedenciaVersion = "derivada-de-source" // último segmento semver de la ruta
	VersionAusente  ProcedenciaVersion = ""                   // no hay dato: Version==""
)

// AvisoNombreDuplicado es el prefijo LITERAL del aviso de fila duplicada. Vive acá, en el
// dominio, porque lo ESCRIBE el adapter (al parsear el catálogo ajeno) y lo LEE
// CalcularSituacion (fila 4 de la tabla de verdad, §6.1): un solo literal, dos consumidores,
// cero drift entre lo que se muestra y lo que se decide.
const AvisoNombreDuplicado = "nombre duplicado en el catálogo: "

// EntradaCatalogo es UNA fila del catálogo (AG-D11: la ENTRADA DE ÍNDICE, o sea el canal —
// lo que el cliente instala con `/plugin install <nombre>@<marketplace>`; dos entradas pueden
// compartir Source y NO son dos arneses).
type EntradaCatalogo struct {
	Nombre      string             `json:"nombre"`
	Source      SourceCatalogo     `json:"source"`
	Version     string             `json:"version,omitempty"`
	VersionDe   ProcedenciaVersion `json:"version_de,omitempty"`
	Descripcion string             `json:"descripcion,omitempty"`
	// ComparteSourceCon son los OTROS Nombre del mismo catálogo con idéntico Source.Crudo
	// (AG-D11: canales). Vacío = fila única.
	ComparteSourceCon []string `json:"comparte_source_con,omitempty"`
	// EstadoCanal viene SOLO del enriquecimiento opcional `catalogo.json` (convención de
	// prenter, NO del estándar): "habilitada" | "deprecada" | "" si no hay archivo.
	EstadoCanal string `json:"estado_canal,omitempty"`
	// NombreAnterior: si `renames` del marketplace mapea algún nombre viejo a este, queda acá
	// (AG-D15) — así un arnés del Portafolio keyeado con el nombre viejo sigue cruzando.
	NombreAnterior []string `json:"nombre_anterior,omitempty"`
	// Aviso son problemas de CALIDAD DEL DATO de esta fila, no de nuestro cálculo: nombre
	// duplicado en el catálogo, `source` que apunta a un dir ausente del checkout, shape de
	// `source` no reconocido. Se MUESTRAN completos (mismo criterio que Instalacion.Aviso /
	// BR-8) y jamás se convierten en un descarte silencioso de la fila.
	Aviso []string `json:"aviso,omitempty"`
	// Situacion y Accion las calcula el dominio al cruzar con el Portafolio (§6). Nunca las
	// calcula el widget (spec §4.3).
	Situacion SituacionCatalogo `json:"situacion"`
	Accion    AccionCatalogo    `json:"accion"`
}

// TieneAvisoDuplicado reporta si esta fila trae el aviso de nombre duplicado en el catálogo
// (fila 4 de la tabla de verdad de §6.1).
func (e EntradaCatalogo) TieneAvisoDuplicado() bool {
	for _, a := range e.Aviso {
		if strings.HasPrefix(a, AvisoNombreDuplicado) {
			return true
		}
	}
	return false
}

// Catalogo es el catálogo leído de UN marketplace. Entradas==nil ⟺ NO hay lectura válida
// (C4 de design.md: viaja `null`, no `[]` — «no sé» ≠ «no tiene arneses», BR-4).
// Entradas==[]EntradaCatalogo{} ⟺ el marketplace declara `plugins: []` de verdad.
type Catalogo struct {
	Marketplace string           `json:"marketplace"`
	Clase       ClaseMarketplace `json:"clase"`
	Repo        string           `json:"repo,omitempty"`
	OwnerNombre string           `json:"owner_nombre,omitempty"`
	OwnerEmail  string           `json:"owner_email,omitempty"`
	OwnerURL    string           `json:"owner_url,omitempty"`
	Descripcion string           `json:"descripcion,omitempty"`
	Lectura     EstadoLectura    `json:"lectura"`
	// Entradas SIN omitempty a propósito: nil DEBE viajar como `null` (BR-4/C4).
	Entradas []EntradaCatalogo `json:"entradas"`
	// Canales/Versiones: enriquecimiento opcional de `catalogo.json` (prenter). nil = ausente,
	// degradado SIN ruido (AG-D11 sub-hallazgo).
	Canales   map[string]string `json:"canales,omitempty"`
	Versiones []VersionCatalogo `json:"versiones,omitempty"`
	// Truncado: cuántas entradas se descartaron por el techo de seguridad (§5.3). >0 es un dato
	// VISIBLE, jamás un recorte silencioso (E-28).
	Truncado int `json:"truncado,omitempty"`
}

// VersionCatalogo es una fila de `catalogo.json#versiones[]` (convención de prenter).
type VersionCatalogo struct {
	Version string `json:"version"`
	Estado  string `json:"estado"` // habilitada | deprecada (crudo, sin enum: es formato ajeno).
	Fuente  string `json:"fuente,omitempty"`
	Fecha   string `json:"fecha,omitempty"`
}

// MergeMarketplaces aplica el collect-all de AG-D9: mergea por Nombre las filas detectadas
// (CC) con las declaradas (store del operador). Reglas, en este orden:
//  1. Un Nombre presente en los dos lados produce UNA fila con Eslabones = [cc-known, declarado].
//  2. Repo: si los dos lados canonicalizan al MISMO valor, se usa. Si difieren, se usa el
//     DECLARADO (el operador es la autoridad de lo que él registró) y la divergencia queda en
//     Discrepancias con LOS DOS valores crudos visibles — jamás se elige en silencio (BR-8).
//  3. Clase: CC no modela clase. La declarada manda; sin declaración, ClaseReferencia
//     (fail-safe: un marketplace que solo apareció por procedencia NO es «propio»).
//  4. InstallLocation/CCActualizado: solo del lado CC (el operador no los declara).
//  5. Orden de salida: por Nombre ascendente (estable, sin depender del orden del map).
//
// Es pura: no lee disco, no consulta red. `detectados` y `declarados` los traen los puertos.
func MergeMarketplaces(detectados, declarados []MarketplaceConocido) []MarketplaceConocido {
	porNombre := map[string]*MarketplaceConocido{}
	orden := make([]string, 0, len(detectados)+len(declarados))

	for _, d := range detectados {
		if d.Nombre == "" {
			continue // sin nombre no hay clave de merge (el store lo reporta como corrupta).
		}
		fila := d
		fila.Eslabones = []EslabonMarketplace{EslabonCCKnown}
		// CC no modela clase: el lado detectado nace en el lado seguro (regla 3).
		fila.Clase = ClaseReferencia
		if _, ok := porNombre[d.Nombre]; !ok {
			orden = append(orden, d.Nombre)
		}
		porNombre[d.Nombre] = &fila
	}

	for _, dec := range declarados {
		if dec.Nombre == "" {
			continue
		}
		existente, ok := porNombre[dec.Nombre]
		if !ok {
			fila := dec
			fila.Eslabones = []EslabonMarketplace{EslabonDeclarado}
			fila.Clase = claseDeclarada(&fila, dec.Clase)
			orden = append(orden, dec.Nombre)
			porNombre[dec.Nombre] = &fila
			continue
		}
		existente.Eslabones = append(existente.Eslabones, EslabonDeclarado)
		existente.Registrado = dec.Registrado
		existente.Repo = repoMergeado(existente, dec)
		existente.Clase = claseDeclarada(existente, dec.Clase)
		existente.Discrepancias = append(existente.Discrepancias, dec.Discrepancias...)
	}

	sort.Strings(orden)
	out := make([]MarketplaceConocido, 0, len(orden))
	for _, n := range orden {
		out = append(out, *porNombre[n])
	}
	anotarRepoCompartido(out)
	return out
}

// claseDeclarada aplica la regla 3 de MergeMarketplaces: la clase declarada manda, y una clase
// fuera del enum degrada a `referencia` con la degradación VISIBLE (E-47) — jamás a `propio`.
func claseDeclarada(fila *MarketplaceConocido, declarada ClaseMarketplace) ClaseMarketplace {
	if declarada == "" {
		return ClaseReferencia
	}
	if !ClaseValida(declarada) {
		fila.Discrepancias = append(fila.Discrepancias,
			fmt.Sprintf("clase desconocida %q: se trata como de referencia", string(declarada)))
		return ClaseReferencia
	}
	return declarada
}

// repoMergeado aplica la regla 2: mismo repo canónico ⇒ se usa; distinto ⇒ manda el DECLARADO
// y la divergencia queda visible con los dos valores crudos (BR-8).
func repoMergeado(detectada *MarketplaceConocido, declarada MarketplaceConocido) string {
	if declarada.Repo == "" {
		return detectada.Repo
	}
	if detectada.Repo == "" {
		return declarada.Repo
	}
	cd, _ := CanonicalizarRepo(detectada.Repo)
	cl, _ := CanonicalizarRepo(declarada.Repo)
	if cd != "" && cd == cl {
		return declarada.Repo
	}
	detectada.Discrepancias = append(detectada.Discrepancias, fmt.Sprintf(
		"repo en conflicto entre eslabones: %s (cc-known-marketplaces) vs %s (declarado-por-operador): manda el declarado",
		detectada.Repo, declarada.Repo))
	return declarada.Repo
}

// anotarRepoCompartido cubre E-67: dos marketplaces con NOMBRES distintos y el MISMO repo NO
// se fusionan (la clave de merge es el nombre, el que CC usa para keyear) — pero la
// coincidencia se anota como discrepancia informativa en cada uno, jamás se resuelve sola.
func anotarRepoCompartido(filas []MarketplaceConocido) {
	porRepo := map[string][]string{}
	for _, f := range filas {
		if f.Repo == "" {
			continue
		}
		canon, ok := CanonicalizarRepo(f.Repo)
		if !ok {
			canon = f.Repo
		}
		porRepo[canon] = append(porRepo[canon], f.Nombre)
	}
	for i := range filas {
		if filas[i].Repo == "" {
			continue
		}
		canon, ok := CanonicalizarRepo(filas[i].Repo)
		if !ok {
			canon = filas[i].Repo
		}
		otros := make([]string, 0, len(porRepo[canon]))
		for _, n := range porRepo[canon] {
			if n != filas[i].Nombre {
				otros = append(otros, n)
			}
		}
		if len(otros) == 0 {
			continue
		}
		sort.Strings(otros)
		filas[i].Discrepancias = append(filas[i].Discrepancias,
			"otro marketplace conocido apunta al mismo repo: "+strings.Join(otros, ", "))
	}
}

// VersionDeEntrada aplica la precedencia de AG-D14/BR-2. `declarada` es plugins[].version tal
// como vino (puede ser ""). Devuelve ("","") cuando no hay dato — jamás una versión fabricada.
func VersionDeEntrada(declarada string, src SourceCatalogo) (version string, de ProcedenciaVersion) {
	if d := strings.TrimSpace(declarada); d != "" {
		// El dato se CONSERVA tal cual vino, incluso si no es semver ("latest"): descartarlo
		// sería perder información del estante; compararlo como string sería el pass fabricado
		// que BR-2 mata. La incomparabilidad la resuelve CalcularSituacion (fila 9).
		return d, VersionDeCampo
	}
	if src.Ruta == "" {
		return "", VersionAusente
	}
	segs := strings.Split(strings.Trim(strings.ReplaceAll(src.Ruta, "\\", "/"), "/"), "/")
	ultimo := segs[len(segs)-1]
	if ultimo != "" && ultimo != "." && EsSemver(ultimo) {
		return ultimo, VersionDeSource
	}
	return "", VersionAusente
}

// EsSemver reporta si s parsea como MAJOR.MINOR.PATCH con pre-release/build opcionales
// (subconjunto laxo de semver 2.0.0, suficiente para comparar versiones de plugin). Sin
// dependencia externa: el repo no tiene golang.org/x/mod y no se agrega una por esto.
func EsSemver(s string) bool {
	_, ok := parsearSemver(s)
	return ok
}

// semver es la descomposición mínima que CompararSemver necesita.
type semver struct {
	major, minor, patch int
	pre                 string
}

// parsearSemver acepta "1.2.3", "1.2.3-rc.1", "1.2.3+build", "v1.2.3" (la `v` de tag se
// tolera: los catálogos reales la usan en `ref`). Rechaza todo lo demás — sin fabricar.
func parsearSemver(s string) (semver, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return semver{}, false
	}
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i] // build metadata no participa de la precedencia (semver 2.0.0 §10).
	}
	var pre string
	if i := strings.IndexByte(s, '-'); i >= 0 {
		pre, s = s[i+1:], s[:i]
	}
	partes := strings.Split(s, ".")
	if len(partes) != 3 {
		return semver{}, false
	}
	nums := make([]int, 3)
	for i, p := range partes {
		if p == "" || strings.TrimLeft(p, "0123456789") != "" {
			return semver{}, false
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return semver{}, false
		}
		nums[i] = n
	}
	return semver{major: nums[0], minor: nums[1], patch: nums[2], pre: pre}, true
}

// CompararSemver ordena a vs b como semver (−1/0/+1). ok=false si alguno no parsea — el caller
// NUNCA cae a comparación de strings (eso es justo el pass fabricado que BR-4 mata en deriva).
func CompararSemver(a, b string) (cmp int, ok bool) {
	sa, oka := parsearSemver(a)
	sb, okb := parsearSemver(b)
	if !oka || !okb {
		return 0, false
	}
	for _, par := range [][2]int{{sa.major, sb.major}, {sa.minor, sb.minor}, {sa.patch, sb.patch}} {
		switch {
		case par[0] < par[1]:
			return -1, true
		case par[0] > par[1]:
			return 1, true
		}
	}
	// Pre-release: una version con pre-release precede a la misma sin él (semver 2.0.0 §11).
	switch {
	case sa.pre == sb.pre:
		return 0, true
	case sa.pre == "":
		return 1, true
	case sb.pre == "":
		return -1, true
	case sa.pre < sb.pre:
		return -1, true
	default:
		return 1, true
	}
}

// AgruparCanales puebla ComparteSourceCon en su: dos entradas con el mismo Source.Crudo son
// canales del mismo arnés físico (AG-D11), y cada una lista a las otras. Muta su in-place y lo
// devuelve para poder encadenar. Orden de los nombres: como vienen en el catálogo (estable).
func AgruparCanales(su []EntradaCatalogo) []EntradaCatalogo {
	porCrudo := map[string][]string{}
	for _, e := range su {
		if e.Source.Crudo == "" {
			continue
		}
		porCrudo[e.Source.Crudo] = append(porCrudo[e.Source.Crudo], e.Nombre)
	}
	for i := range su {
		crudo := su[i].Source.Crudo
		if crudo == "" {
			su[i].ComparteSourceCon = nil
			continue
		}
		otros := make([]string, 0, len(porCrudo[crudo]))
		for _, n := range porCrudo[crudo] {
			if n != su[i].Nombre {
				otros = append(otros, n)
			}
		}
		if len(otros) == 0 {
			su[i].ComparteSourceCon = nil
			continue
		}
		su[i].ComparteSourceCon = otros
	}
	return su
}
