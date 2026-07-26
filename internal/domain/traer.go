package domain

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// traer.go es la decisión PURA de `↧ Traer canónico` (AG-D17 FIRMADA 🧑‍⚖️, design.md §13):
// QUÉ se materializa, DESDE dónde y HACIA dónde. Sin un solo `os.` — así BR-13 (el destino) y
// las precondiciones de BR-1/BR-14 se prueban sin tocar disco: son DECISIONES, no efectos.

// CaminoTraer son las dos vías de AG-D17 más el rechazo explícito.
type CaminoTraer string

// Las dos vías + la ausencia de vía.
const (
	CaminoLocal   CaminoTraer = "local"   // copia de subcarpeta del checkout de CC. Sin red.
	CaminoExterno CaminoTraer = "externo" // fetch shallow por sha + extracción de subruta.
	CaminoNinguno CaminoTraer = ""        // no materializable: el motivo dice por qué.
)

// Errores de PLANIFICACIÓN (puros, sin I/O). El handler los mapea por errors.Is (§13.8).
var (
	// ErrTraerClaseReferencia: BR-1 — en `referencia` la acción NO EXISTE. Es el mismo
	// invariante que AccionDeSituacion ya enforça; acá se re-chequea porque un POST puede
	// llegar sin haber pasado por la UI (el botón deshabilitado NO es un control de acceso).
	// Es el 5º check del boundary marketplace-referencia-es-solo-procedencia.
	ErrTraerClaseReferencia = errors.New("traer: " + MotivoSoloArnesesPropios)
	// ErrTraerSourceNoMaterializable: `SourceDesconocido`, o un objeto sin `sha` NI `ref` (nada
	// que pinear), o una ruta relativa sin checkout en disco. El motivo lleva el crudo visible (E-90).
	ErrTraerSourceNoMaterializable = errors.New("traer: el catálogo declara un source que no sé materializar")
	// ErrTraerDestinoEscapa: BR-13 — el destino calculado no cae dentro de ~/.arnesia/checkouts.
	ErrTraerDestinoEscapa = errors.New("traer: el destino calculado escapa de ~/.arnesia/checkouts")
)

// Errores de EJECUCIÓN de Traer. Viven en el dominio por la misma razón que los de lectura
// (C23 de design.md): los PRODUCE el adapter `internal/adapters/traer` y los CLASIFICA el
// transporte, y go-arch-lint no permite que ninguno de los dos importe al otro. `usecase` los
// re-exporta con los nombres del contrato de design.md §13.3.
var (
	// ErrTraerRemotoNoTiene: el repo/sha/ref que el catálogo declara no existe en el remoto, o el
	// commit no trae la subruta, o el fetch dio timeout → 502 («miré y el estante mintió»).
	ErrTraerRemotoNoTiene = errors.New("traer: el remoto no tiene lo que el catálogo declara")
	// ErrTraerSinAuth: BR-18 — sin `gh` autenticado ni PAT para un repo que los exige → 503
	// («no puedo mirar»), jamás 400. ArnesIA no pide, guarda ni proxya credenciales.
	ErrTraerSinAuth = errors.New("traer: sin credencial para acceder al repo (gh no autenticado y sin PAT)")
	// ErrTraerLocal: fallo NUESTRO — disco lleno, permisos, rename, Upsert → 500.
	ErrTraerLocal = errors.New("traer: fallo local al materializar")
)

// PlanTraer es la decisión pura de CÓMO materializar una fila del catálogo: qué camino, de
// dónde, qué se verifica y a qué destino.
type PlanTraer struct {
	Camino  CaminoTraer `json:"camino"`
	Destino string      `json:"destino"` // absoluto, YA validado dentro de checkouts (BR-13).
	// Identidad con la que se registrará el canónico. Home = repo CANONICALIZADO (RN-IDENT-1),
	// NO el nombre del marketplace — ver C17 de design.md y la advertencia de RutaCanonico.
	Identidad IdentidadArnes `json:"identidad"`

	// ── camino A ──
	OrigenLocal string `json:"origen_local,omitempty"` // <installLocation>/<ruta del source>.
	// RaizDeMarketplace marca el caso `source: "./"`: el arnés declarado ES la raíz del repo del
	// marketplace. Se materializa igual (el catálogo declara lo que declara), con Aviso visible.
	RaizDeMarketplace bool `json:"raiz_de_marketplace,omitempty"`

	// ── camino B ──
	URL string `json:"url,omitempty"` // https://host/owner/repo.git ya armada.
	// SHAEsperado es el pin del catálogo y la AUTORIDAD (§13.1 hecho 2): se hace fetch de ESTE
	// commit y se verifica que HEAD lo iguale (BR-16). "" ⇒ no hay pin: se usa Ref.
	SHAEsperado string `json:"sha_esperado,omitempty"`
	Ref         string `json:"ref,omitempty"`     // pista/fallback, NUNCA la autoridad.
	Subruta     string `json:"subruta,omitempty"` // `path` del source; "" ⇒ la raíz del repo.

	// Avisos son problemas del DATO que no impiden traer y que se MUESTRAN (BR-8): `commit`≠`sha`
	// en un source `github`, `ref` que no coincide con `sha`, `source: "./"`, sin subruta.
	Avisos []string `json:"avisos,omitempty"`
}

// PlanificarTraer decide el plan SIN tocar disco (`installLocationExiste` lo averigua el usecase
// y lo pasa como HECHO). Reglas, en orden:
//  1. clase != propio                                            → ErrTraerClaseReferencia (BR-1)
//  2. Source.Tipo == SourceDesconocido                            → ErrTraerSourceNoMaterializable (E-90)
//  3. SourceRutaRelativa ∧ installLocationExiste                  → CaminoLocal
//  4. SourceRutaRelativa ∧ ¬existe                                → ErrTraerSourceNoMaterializable
//     («el catálogo declara una ruta relativa pero el marketplace no está clonado en disco»:
//     una ruta relativa no dice a qué repo pertenece — inventarlo sería fabricar el origen)
//  5. Source objeto sin SHA ni Ref                                → ErrTraerSourceNoMaterializable
//  6. Source objeto                                               → CaminoExterno
//
// El destino sale de RutaCanonico (§13.4) y se valida ahí mismo.
func PlanificarTraer(mkt MarketplaceConocido, fila EntradaCatalogo, installLocationExiste bool, raizArnesia string) (PlanTraer, error) {
	// 1 · BR-1 en el DOMINIO, con cero I/O. Un `disabled` de la UI es render, no control de acceso.
	if ClaseSegura(mkt.Clase) != ClasePropio {
		return PlanTraer{}, fmt.Errorf("%w (marketplace %q es de referencia)", ErrTraerClaseReferencia, mkt.Nombre)
	}
	// La identidad del canónico exige un repo canonicalizable: sin él, el canónico quedaría con
	// identidad provisional y NO cruzaría con su propia fila del catálogo (C17/E-106).
	home, ok := CanonicalizarRepo(mkt.Repo)
	if !ok {
		return PlanTraer{}, fmt.Errorf("%w: el marketplace %q no declara un repo canonicalizable (%q): sin él la identidad del canónico sería provisional y no cruzaría con su propia fila",
			ErrTraerSourceNoMaterializable, mkt.Nombre, mkt.Repo)
	}
	if fila.Nombre == "" {
		return PlanTraer{}, fmt.Errorf("%w: la fila del catálogo no declara nombre", ErrTraerSourceNoMaterializable)
	}

	destino, dok := RutaCanonico(raizArnesia, mkt.Nombre, fila.Nombre)
	if !dok {
		return PlanTraer{}, fmt.Errorf("%w: %s / %s", ErrTraerDestinoEscapa, mkt.Nombre, fila.Nombre)
	}

	plan := PlanTraer{
		Destino:   destino,
		Identidad: IdentidadArnes{Home: home, ID: fila.Nombre},
	}
	src := fila.Source

	switch src.Tipo {
	case SourceDesconocido:
		return PlanTraer{}, fmt.Errorf("%w: %s", ErrTraerSourceNoMaterializable, src.Crudo)

	case SourceRutaRelativa:
		if !installLocationExiste {
			return PlanTraer{}, fmt.Errorf("%w: el catálogo declara la ruta relativa %q pero el marketplace no está clonado en disco: una ruta relativa no dice a qué repo pertenece",
				ErrTraerSourceNoMaterializable, src.Ruta)
		}
		plan.Camino = CaminoLocal
		limpia := strings.TrimSpace(src.Ruta)
		plan.RaizDeMarketplace = limpia == "" || limpia == "./" || limpia == "." || limpia == "/"
		if plan.RaizDeMarketplace {
			plan.OrigenLocal = filepath.Clean(mkt.InstallLocation)
			plan.Avisos = append(plan.Avisos, `el catálogo declara la raíz del marketplace como el arnés (source "./"): el canónico es una copia de ese árbol sin su .git`)
		} else {
			plan.OrigenLocal = filepath.Join(mkt.InstallLocation, filepath.FromSlash(limpia))
		}
		// El catálogo AJENO no dicta dónde leemos: una ruta relativa que escapa del checkout con
		// `..` se rechaza acá, en el dominio y sin I/O (§13.6 paso 1).
		if !dentroDeRuta(mkt.InstallLocation, plan.OrigenLocal) {
			return PlanTraer{}, fmt.Errorf("%w: la ruta relativa %q escapa del checkout del marketplace (%s)",
				ErrTraerSourceNoMaterializable, src.Ruta, mkt.InstallLocation)
		}
		return plan, nil

	case SourceGitSubdir, SourceURL, SourceGitHub:
		if src.SHA == "" && src.Ref == "" {
			return PlanTraer{}, fmt.Errorf("%w: el catálogo declara un source objeto sin `sha` ni `ref`: no hay nada que pinear (%s)",
				ErrTraerSourceNoMaterializable, src.Crudo)
		}
		url, uok := urlDeSource(src)
		if !uok {
			return PlanTraer{}, fmt.Errorf("%w: el source objeto no declara url ni repo resoluble (%s)", ErrTraerSourceNoMaterializable, src.Crudo)
		}
		plan.Camino = CaminoExterno
		plan.URL = url
		plan.SHAEsperado = src.SHA
		plan.Ref = src.Ref
		plan.Subruta = strings.Trim(strings.TrimSpace(src.Ruta), "/")
		plan.Avisos = append(plan.Avisos, avisosDeSourceObjeto(src)...)
		return plan, nil

	default:
		return PlanTraer{}, fmt.Errorf("%w: %s", ErrTraerSourceNoMaterializable, src.Crudo)
	}
}

// avisosDeSourceObjeto arma los avisos VISIBLES del dato (BR-8): nunca bloquean, nunca se callan.
func avisosDeSourceObjeto(src SourceCatalogo) []string {
	var out []string
	if src.Tipo == SourceGitHub && src.Commit != "" && src.SHA != "" && src.Commit != src.SHA {
		out = append(out, fmt.Sprintf(
			"el catálogo declara dos hashes distintos (commit %s y sha %s) sin semántica documentada: se materializa el sha, que es el único campo presente en todas las formas",
			src.Commit, src.SHA))
	}
	if src.Ref != "" && src.SHA != "" {
		out = append(out, fmt.Sprintf(
			"el ref declarado (%s) puede no apuntar al sha pineado (%s): manda el sha", src.Ref, src.SHA))
	}
	if strings.TrimSpace(src.Ruta) == "" {
		out = append(out, "el catálogo no declara subruta: el canónico es la raíz del repo")
	}
	return out
}

// urlDeSource arma la url de clone del source objeto: la `url` declarada manda; con solo `repo`
// (forma `github`) se canonicaliza y se arma `https://<host>/<owner>/<repo>.git`.
func urlDeSource(src SourceCatalogo) (string, bool) {
	if u := strings.TrimSpace(src.URL); u != "" {
		return u, true
	}
	canon, ok := CanonicalizarRepo(src.Repo)
	if !ok {
		return "", false
	}
	return "https://" + canon + ".git", true
}

// RaizCheckouts es el único territorio donde Traer escribe: `<raizArnesia>/checkouts`.
func RaizCheckouts(raizArnesia string) string {
	return filepath.Join(raizArnesia, "checkouts")
}

// RaizTemporales es donde vive el staging de Traer: `<raizArnesia>/tmp`.
//
// ⚠ NO es `os.TempDir()` A PROPÓSITO (design.md §12.2 riesgo 11): `os.Rename` entre filesystems
// distintos falla con EXDEV, y `/tmp` es `tmpfs` en muchas instalaciones. Bajo `~/.arnesia/` el
// staging comparte filesystem con el destino ⇒ el rename es atómico DE VERDAD (BR-15). E-101 lo
// vigila por prefijo de path.
func RaizTemporales(raizArnesia string) string {
	return filepath.Join(raizArnesia, "tmp")
}

// RutaCanonico calcula el destino de un Traer y GARANTIZA que cae dentro de RaizCheckouts
// (BR-13). Reglas:
//  1. Los dos segmentos pasan por Slug() — la MISMA regla del Portafolio: minúsculas, runs de
//     [^a-z0-9-_] colapsados a un '-', trim de '-'. Un `..` colapsa y un "../../etc" da "etc":
//     estructuralmente NO puede haber un segmento de escape ni un separador.
//  2. Un segmento cuyo Slug PIERDE información (no es igual al crudo) lleva sufijo de huella
//     `~<HuellaPath>`: así "mi mkt" y "mi-mkt" —que colapsan al mismo slug— NUNCA dan el mismo
//     destino, y un nombre íntegramente no-ASCII no produce un segmento vacío.
//  3. Se re-verifica el resultado con DentroDeCheckouts(): defensa en profundidad, porque la
//     garantía de (1) es un argumento sobre Slug y este chequeo es una aserción sobre el path.
//
// slugHome sale del NOMBRE del marketplace (legible, y es la clave única del registro);
// `id` es EntradaCatalogo.Nombre. ok=false ⇒ ErrTraerDestinoEscapa.
//
// ⚠ LA TRAMPA (C17 de design.md): el **path** usa el nombre del marketplace, pero la
// **identidad** usa `Home = CanonicalizarRepo(mkt.Repo)`. Son dos cosas distintas y colapsarlas
// rompe el cruce de §6.2: un escaneo produce `Registries: ["github.com/owner/repo"]`, así que un
// canónico registrado con `Home = "prenter-marketplace"` NO cruzaría con la fila del catálogo de
// la que vino. El path es almacenamiento; la identidad es el contrato (RN-IDENT-1).
func RutaCanonico(raizArnesia, nombreMarketplace, id string) (destino string, ok bool) {
	segHome, hok := segmentoDeDestino(nombreMarketplace)
	segID, iok := segmentoDeDestino(id)
	if !hok || !iok {
		return "", false
	}
	destino = filepath.Join(RaizCheckouts(raizArnesia), segHome, segID)
	if !DentroDeCheckouts(raizArnesia, destino) {
		return "", false
	}
	return destino, true
}

// segmentoDeDestino convierte un nombre crudo en UN segmento de path seguro y sin colisiones
// (regla 2 de RutaCanonico). ok=false solo con un crudo vacío: sin nombre no hay destino.
func segmentoDeDestino(crudo string) (string, bool) {
	if strings.TrimSpace(crudo) == "" {
		return "", false
	}
	slug := Slug(crudo)
	if slug == crudo && slug != "" {
		return slug, true // lossless: el nombre ES su slug ⇒ path legible y único por definición.
	}
	// El slug perdió información: se desempata con la huella del crudo, así dos nombres distintos
	// que colapsan al mismo slug jamás comparten destino.
	huella := HuellaPath("segmento:" + crudo)
	if slug == "" {
		return "h~" + huella, true
	}
	return slug + "~" + huella, true
}

// DentroDeCheckouts reporta si `p` está contenido en RaizCheckouts(raizArnesia) — comparación
// sobre paths Clean-eados y con `..` ya resuelto. NO llama a EvalSymlinks: el caller decide si el
// path ya existe (destino) o todavía no (a crear). La variante que sí resuelve symlinks es
// DentroDeCheckoutsResuelto, para el chequeo POST-creación (§13.9).
func DentroDeCheckouts(raizArnesia, p string) bool {
	raiz := filepath.Clean(RaizCheckouts(raizArnesia))
	limpio := filepath.Clean(p)
	if limpio == raiz {
		return true
	}
	rel, err := filepath.Rel(raiz, limpio)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// dentroDeRuta reporta si child es igual a, o está contenido en, parent (paths Clean-eados, sin
// I/O). Mismo criterio que el walker del Portafolio (C-P-12).
func dentroDeRuta(parent, child string) bool {
	rel, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
