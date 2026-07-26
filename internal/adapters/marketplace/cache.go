package marketplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// cache.go guarda la última lectura EXITOSA de cada catálogo en
// `~/.arnesia/catalogos/<slug(nombre)>.json` — UN archivo por marketplace (design.md §4.2).
//
// Por qué existe y no es opcional: el plano necesita «N entradas» y «leído hace <t>» POR FILA, y
// uno de los catálogos reales son 159 KB / 273 entradas (AG-D16). Sin caché, cada
// `GET /api/marketplaces` parsearía los 5 `marketplace.json`. El valor offline es real pero
// SECUNDARIO: para los marketplaces que CC ya clonó, el checkout local ES la fuente offline.
//
// Por qué un archivo por marketplace: (a) el radio de daño de una corrupción es un marketplace,
// no todos; (b) refrescar uno no reescribe ~200 KB; (c) `Olvidar` es un `os.Remove`.

// cacheVersionActual es la versión del envelope del caché. Un mismatch DESCARTA el caché y lo
// reconstruye (E-49) — acá sí es legítimo, a diferencia del registro: este dato ES derivable.
const cacheVersionActual = 1

// ErrCacheSinEntradas se devuelve cuando `Guardar` recibe un catálogo sin lectura válida
// (Entradas==nil) y lo RECHAZA.
// Cachear un «no sé» convertiría un fallo transitorio en un estado persistente (E-50).
var ErrCacheSinEntradas = errors.New("marketplace cache: no se cachea un catálogo sin lectura (entradas nil)")

// envelopeCache es la forma en disco de una lectura cacheada.
type envelopeCache struct {
	Version  int             `json:"version"`
	Nombre   string          `json:"nombre"`
	LeidoEn  string          `json:"leido_en"`
	Fuente   string          `json:"fuente"`
	Catalogo domain.Catalogo `json:"catalogo"`
}

// Cache implementa ports.CatalogoCache sobre JSON atómico, un archivo por marketplace.
type Cache struct {
	dir   string
	ahora func() time.Time

	mu sync.Mutex
}

var _ ports.CatalogoCache = (*Cache)(nil)

// NewCache abre (o crea) el dir del caché (default `~/.arnesia/catalogos`).
func NewCache(dir string) (*Cache, error) {
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("marketplace cache: resolve home: %w", err)
		}
		dir = filepath.Join(home, ".arnesia", "catalogos")
	}
	return &Cache{dir: dir, ahora: time.Now}, nil
}

// SetAhora inyecta el reloj (tests): el caché estampa `leido_en` y los tests no pueden depender
// del reloj real.
func (c *Cache) SetAhora(f func() time.Time) { c.ahora = f }

// rutaDe resuelve el archivo de `nombre`. El slug usa `domain.Slug` (la MISMA regla del
// Portafolio: `[a-z0-9-_]`, runs colapsados), así un nombre con `@`, espacios, unicode o `..`
// no puede escapar del directorio (E-52). Una colisión de slug entre dos nombres distintos se
// resuelve con sufijo de huella: NUNCA se pisa el caché de otro.
func (c *Cache) rutaDe(nombre string) string {
	slug := domain.Slug(nombre)
	if slug == "" {
		// Nombre íntegramente no-ASCII (o vacío): la huella es el desempate, jamás un archivo
		// sin nombre que colisionaría con cualquier otro igual de anónimo.
		slug = "mkt~" + domain.HuellaPath(nombre)
	}
	base := filepath.Join(c.dir, slug+".json")
	if dueño, ok := nombreDelArchivo(base); ok && dueño != nombre {
		return filepath.Join(c.dir, slug+"~"+domain.HuellaPath(nombre)+".json")
	}
	return base
}

// nombreDelArchivo lee solo el `nombre` de un archivo de caché existente (para detectar
// colisión de slug). ok=false si no existe o no parsea.
func nombreDelArchivo(path string) (string, bool) {
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta del propio territorio ~/.arnesia.
	if err != nil {
		return "", false
	}
	var env struct {
		Nombre string `json:"nombre"`
	}
	if json.Unmarshal(b, &env) != nil {
		return "", false
	}
	return env.Nombre, true
}

// Leer devuelve el catálogo cacheado con su `Lectura` ya poblada desde el envelope (Tipo=leido,
// Cuando=leido_en, Fuente). ok=false tanto si no hay caché (motivo "") como si está corrupto o
// es de una versión desconocida (motivo poblado, para que el degradado sea VISIBLE).
func (c *Cache) Leer(nombre string) (domain.Catalogo, string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.rutaDe(nombre)
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta del propio territorio ~/.arnesia.
	if err != nil {
		return domain.Catalogo{}, "", false // ausente: la fila dice `no-leido`, nunca «0 entradas».
	}
	var env envelopeCache
	if uerr := json.Unmarshal(b, &env); uerr != nil {
		return domain.Catalogo{}, fmt.Sprintf("caché de catálogo ilegible: %v", uerr), false
	}
	if env.Version != cacheVersionActual {
		return domain.Catalogo{}, fmt.Sprintf("caché de versión %d desconocida (se re-leerá)", env.Version), false
	}
	if env.Catalogo.Entradas == nil {
		// Defensa en profundidad: `Guardar` ya lo rechaza, pero un archivo editado a mano no
		// puede convertir un «no sé» en un estado persistente.
		return domain.Catalogo{}, "caché sin lectura válida (entradas null): se re-leerá", false
	}
	cat := env.Catalogo
	cat.Lectura = domain.EstadoLectura{
		Tipo:     domain.LecturaLeida,
		Cuando:   env.LeidoEn,
		Entradas: len(cat.Entradas),
		Fuente:   env.Fuente,
	}
	return cat, "", true
}

// Guardar persiste la lectura con su timestamp. RECHAZA un catálogo sin lectura válida
// (ErrCacheSinEntradas, E-50) y no escribe ningún archivo en ese caso.
func (c *Cache) Guardar(nombre string, cat domain.Catalogo) error {
	if cat.Entradas == nil {
		return fmt.Errorf("%w: %s", ErrCacheSinEntradas, nombre)
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	cuando := cat.Lectura.Cuando
	if cuando == "" {
		cuando = c.ahora().UTC().Format(time.RFC3339)
	}
	env := envelopeCache{
		Version:  cacheVersionActual,
		Nombre:   nombre,
		LeidoEn:  cuando,
		Fuente:   cat.Lectura.Fuente,
		Catalogo: cat,
	}
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("marketplace cache: encode %s: %w", nombre, err)
	}
	return escribirAtomico(c.rutaDe(nombre), ".catalogo-*.json", b)
}

// Olvidar borra el archivo de caché de nombre (un `os.Remove`). Ausente no es error.
func (c *Cache) Olvidar(nombre string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := os.Remove(c.rutaDe(nombre)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("marketplace cache: olvidar %s: %w", nombre, err)
	}
	return nil
}
