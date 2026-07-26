// Package marketplace es el adapter del ESTANTE: registro declarado + caché de catálogo +
// detector de `known_marketplaces.json` + los dos lectores de catálogo (local y remoto).
//
// Es el ANTI-CORRUPTION LAYER del boundary `marketplace-referencia-es-solo-procedencia`
// (L2 punto 4): el shape crudo del `marketplace.json` ajeno se traduce ACÁ a `domain.*` y el
// resto de las claves se DESCARTA — el dominio no crece un campo por cada extra del formato de
// Claude Code. Nunca escribe en `~/.claude/**` ni en el checkout de un marketplace.
package marketplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Alias locales de los centinelas del dominio (declarados en `internal/domain/marketplace.go`
// para que adapter y usecase compartan LA MISMA instancia: ninguno de los dos puede importar al
// otro — go-arch-lint). Los alias existen solo para que este paquete se lea natural.
var (
	ErrNoEsMarketplace = domain.ErrNoEsMarketplace
	ErrSinViaDeLectura = domain.ErrSinViaDeLectura
)

// maxEntradas es el techo de filas de un catálogo (§5.3). Lo que se descarta viaja en
// Catalogo.Truncado y se MUESTRA: recorte visible, jamás silencioso (E-58).
const maxEntradas = 5000

// crudoMarketplace es el shape AJENO de `.claude-plugin/marketplace.json`, tal como los 5
// marketplaces reales de esta máquina lo escriben (AG-D13/D15, verificado 2026-07-25). Las
// claves que el dominio no modela (`$schema`, `strict`, `skills`, `lspServers`, `homepage`,
// `author`, `displayName`, `keywords`, `category`, `tags`) NO se declaran acá a propósito:
// `encoding/json` las ignora y el dominio no las hereda.
type crudoMarketplace struct {
	Name        string `json:"name"`
	Description string `json:"description"` // top-level: 4 de los 5 marketplaces reales la usan.
	Metadata    struct {
		Description string `json:"description"` // convención de prenter (el 5º).
	} `json:"metadata"`
	Owner struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		URL   string `json:"url"` // caveman/ponytail/warp usan `url` en vez de `email` (AG-D15).
	} `json:"owner"`
	Renames map[string]string `json:"renames"` // viejo → nuevo (6 entradas reales en el oficial).
	// Plugins es *[]…: distingue «la clave falta» (nil ⇒ E-31, Entradas=nil + motivo) de
	// «declara una lista vacía» ([] ⇒ E-32, Entradas=[] con su fecha). Es BR-4 en el parser.
	Plugins *[]crudoEntrada `json:"plugins"`
}

// crudoEntrada es una fila de `plugins[]`. `Source` viaja como RawMessage porque tiene CUATRO
// formas reales (AG-D13); `Version` como *string para distinguir ausente / null / "".
type crudoEntrada struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     *string         `json:"version"`
	Source      json.RawMessage `json:"source"`
}

// crudoSourceObjeto es la forma OBJETO de `plugins[].source` (220 de 273 filas del catálogo
// oficial). `commit` es un extra que SOLO la forma `github` declara, y con valor distinto del
// `sha` (C18): se conserva normalizado para poder citar los dos en el aviso.
type crudoSourceObjeto struct {
	Source string `json:"source"`
	URL    string `json:"url"`
	Repo   string `json:"repo"`
	Path   string `json:"path"`
	Ref    string `json:"ref"`
	SHA    string `json:"sha"`
	Commit string `json:"commit"`
}

// crudoCatalogoJSON es el enriquecimiento OPCIONAL `catalogo.json` (convención de prenter, NO
// del estándar): un marketplace ajeno no tiene por qué tenerlo.
type crudoCatalogoJSON struct {
	Marketplace string                   `json:"marketplace"`
	Canales     map[string]string        `json:"canales"`
	Versiones   []domain.VersionCatalogo `json:"versiones"`
}

// parsearMarketplaceJSON traduce el shape ajeno a domain.Catalogo (§5.4). Devuelve el catálogo
// con Entradas/Truncado/avisos ya poblados y `motivo` con lo que haya que mostrar además de la
// lectura exitosa (clave `plugins` ausente, filas descartadas). Un archivo sin `name` o que no
// parsea ⇒ ErrNoEsMarketplace envuelto con el motivo real.
func parsearMarketplaceJSON(b []byte) (domain.Catalogo, string, error) {
	var crudo crudoMarketplace
	if err := json.Unmarshal(b, &crudo); err != nil {
		// El OFFSET del error viaja en el motivo (E-33): `json.SyntaxError` lo trae pero no lo
		// pone en su mensaje, y sin él el operador no sabe DÓNDE está roto el archivo ajeno.
		var syn *json.SyntaxError
		if errors.As(err, &syn) {
			return domain.Catalogo{}, "", fmt.Errorf("%w: JSON inválido en offset %d: %w", ErrNoEsMarketplace, syn.Offset, err)
		}
		return domain.Catalogo{}, "", fmt.Errorf("%w: JSON inválido: %w", ErrNoEsMarketplace, err)
	}
	if strings.TrimSpace(crudo.Name) == "" {
		return domain.Catalogo{}, "", fmt.Errorf("%w: marketplace.json sin `name`", ErrNoEsMarketplace)
	}

	cat := domain.Catalogo{
		Marketplace: crudo.Name,
		OwnerNombre: crudo.Owner.Name,
		OwnerEmail:  crudo.Owner.Email,
		OwnerURL:    crudo.Owner.URL,
		Descripcion: primerNoVacio(crudo.Description, crudo.Metadata.Description),
	}

	// `plugins` ausente ⇒ Entradas nil (null en el cable) + motivo VISIBLE: «leí y el archivo
	// está incompleto», no «no tiene arneses» (E-31, BR-4).
	if crudo.Plugins == nil {
		return cat, "marketplace.json sin la clave plugins", nil
	}

	// A partir de acá HUBO lectura: el slice se inicializa (aunque quede en 0) porque `[]` es
	// una afirmación evidenciada — «leí y no declara ninguno» (E-32).
	filas := *crudo.Plugins
	anterioresDe := invertirRenames(crudo.Renames)

	entradas := make([]domain.EntradaCatalogo, 0, len(filas))
	sinNombre, truncadas := 0, 0
	for _, f := range filas {
		if strings.TrimSpace(f.Name) == "" {
			sinNombre++ // sin `name` no hay nada instalable que nombrar: se descarta VISIBLE (E-40).
			continue
		}
		if len(entradas) >= maxEntradas {
			truncadas++
			continue
		}
		e := domain.EntradaCatalogo{Nombre: f.Name, Descripcion: f.Description}
		var avisoSource string
		e.Source, avisoSource = normalizarSource(f.Source)
		if avisoSource != "" {
			e.Aviso = append(e.Aviso, avisoSource)
		}
		var declarada string
		if f.Version != nil {
			declarada = *f.Version
		}
		e.Version, e.VersionDe = domain.VersionDeEntrada(declarada, e.Source)
		if ant := anterioresDe[f.Name]; len(ant) > 0 {
			e.NombreAnterior = ant
		}
		entradas = append(entradas, e)
	}

	entradas = anotarDuplicados(entradas)
	cat.Entradas = domain.AgruparCanales(entradas)
	cat.Truncado = truncadas + sinNombre

	var motivos []string
	if sinNombre > 0 {
		motivos = append(motivos, fmt.Sprintf("%d fila(s) sin name descartada(s): sin nombre no hay nada instalable", sinNombre))
	}
	if truncadas > 0 {
		motivos = append(motivos, fmt.Sprintf("%d fila(s) por encima del techo de %d entradas (recorte visible)", truncadas, maxEntradas))
	}
	return cat, strings.Join(motivos, " · "), nil
}

// normalizarSource traduce `plugins[].source` (string u objeto) a domain.SourceCatalogo. Una
// forma no reconocida NO se descarta: queda SourceDesconocido con el crudo recortado visible
// (E-39) y su aviso.
func normalizarSource(raw json.RawMessage) (domain.SourceCatalogo, string) {
	if len(raw) == 0 {
		return domain.SourceCatalogo{Tipo: domain.SourceDesconocido, Crudo: ""},
			"el catálogo no declara `source` para esta fila"
	}

	var ruta string
	if err := json.Unmarshal(raw, &ruta); err == nil {
		return domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: ruta, Ruta: ruta}, ""
	}

	var obj crudoSourceObjeto
	if err := json.Unmarshal(raw, &obj); err == nil && obj.Source != "" {
		tipo := domain.TipoSource(obj.Source)
		switch tipo {
		case domain.SourceGitSubdir, domain.SourceURL, domain.SourceGitHub:
		default:
			return domain.SourceCatalogo{Tipo: domain.SourceDesconocido, Crudo: recortar(string(raw))},
				"el catálogo declara un `source` de tipo desconocido: " + obj.Source
		}
		destino := obj.URL
		if destino == "" {
			destino = obj.Repo
		}
		return domain.SourceCatalogo{
			Tipo:   tipo,
			Crudo:  domain.CrudoDeSource(tipo, destino, obj.Path, obj.Ref),
			Ruta:   obj.Path,
			URL:    obj.URL,
			Repo:   obj.Repo,
			Ref:    obj.Ref,
			SHA:    obj.SHA,
			Commit: obj.Commit,
		}, ""
	}

	return domain.SourceCatalogo{Tipo: domain.SourceDesconocido, Crudo: recortar(string(raw))},
		"el catálogo declara un `source` con una forma que no reconozco: " + recortar(string(raw))
}

// recortar acota un crudo a 120 chars para que un archivo hostil no infle el wire; el recorte
// se MARCA (nunca se hace en silencio).
func recortar(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= 120 {
		return s
	}
	return s[:120] + "…"
}

// anotarDuplicados cubre E-37: dos filas con el mismo `name` se CONSERVAN las dos, las dos con
// el aviso. El catálogo ajeno está mal y eso se muestra, en vez de elegir una o dedupear el
// error del otro.
func anotarDuplicados(es []domain.EntradaCatalogo) []domain.EntradaCatalogo {
	cuenta := map[string]int{}
	for _, e := range es {
		cuenta[e.Nombre]++
	}
	for i := range es {
		if cuenta[es[i].Nombre] > 1 {
			es[i].Aviso = append(es[i].Aviso, domain.AvisoNombreDuplicado+es[i].Nombre)
		}
	}
	return es
}

// invertirRenames convierte el mapa `viejo → nuevo` del catálogo en `nuevo → [viejos]`, que es
// lo que la fila destino necesita para seguir cruzando con una entrada keyeada por el nombre
// viejo (AG-D15, E-38).
func invertirRenames(renames map[string]string) map[string][]string {
	if len(renames) == 0 {
		return nil
	}
	out := map[string][]string{}
	for viejo, nuevo := range renames {
		if viejo == "" || nuevo == "" {
			continue
		}
		out[nuevo] = append(out[nuevo], viejo)
	}
	for k := range out {
		sort.Strings(out[k]) // orden estable: el map de Go no lo garantiza.
	}
	return out
}

// aplicarEnriquecimiento cruza el `catalogo.json` opcional con el catálogo ya parseado: puebla
// Canales/Versiones y el EstadoCanal de cada fila (por su Version). Un archivo de OTRO
// marketplace se ignora (E-73): el enriquecimiento no se aplica a ciegas.
func aplicarEnriquecimiento(cat *domain.Catalogo, b []byte) string {
	var crudo crudoCatalogoJSON
	if err := json.Unmarshal(b, &crudo); err != nil {
		return fmt.Sprintf("(catalogo.json ignorado: %v)", err)
	}
	if crudo.Marketplace != "" && crudo.Marketplace != cat.Marketplace {
		return fmt.Sprintf("(catalogo.json ignorado: declara marketplace %q, no %q)", crudo.Marketplace, cat.Marketplace)
	}
	cat.Canales = crudo.Canales
	cat.Versiones = crudo.Versiones
	estadoDe := map[string]string{}
	for _, v := range crudo.Versiones {
		estadoDe[v.Version] = v.Estado
	}
	for i := range cat.Entradas {
		if est, ok := estadoDe[cat.Entradas[i].Version]; ok {
			cat.Entradas[i].EstadoCanal = est
		}
	}
	return ""
}

// primerNoVacio devuelve el primer string no vacío de vals.
func primerNoVacio(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
