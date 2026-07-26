package marketplace

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

func cacheEn(t *testing.T) (*Cache, string) {
	t.Helper()
	dir := t.TempDir()
	c, err := NewCache(dir)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	return c, dir
}

func catSano(nombre string, cuando string) domain.Catalogo {
	return domain.Catalogo{
		Marketplace: nombre,
		Clase:       domain.ClasePropio,
		Lectura:     domain.EstadoLectura{Tipo: domain.LecturaLeida, Cuando: cuando, Entradas: 2, Fuente: "local"},
		Entradas: []domain.EntradaCatalogo{
			{Nombre: "harness", Version: "0.5.3", VersionDe: domain.VersionDeSource},
			{Nombre: "harness-beta", Version: "0.5.3", VersionDe: domain.VersionDeSource},
		},
	}
}

// Round-trip: se guarda con su timestamp y se lee con la Lectura repoblada desde el envelope.
func TestCacheRoundTrip(t *testing.T) {
	c, _ := cacheEn(t)
	if err := c.Guardar("prenter-marketplace", catSano("prenter-marketplace", "2026-07-25T14:07:33Z")); err != nil {
		t.Fatal(err)
	}
	got, motivo, ok := c.Leer("prenter-marketplace")
	if !ok || motivo != "" {
		t.Fatalf("Leer = (ok=%v, motivo=%q), want (true, \"\")", ok, motivo)
	}
	if len(got.Entradas) != 2 {
		t.Fatalf("len(Entradas) = %d, want 2", len(got.Entradas))
	}
	if got.Lectura.Cuando != "2026-07-25T14:07:33Z" || got.Lectura.Fuente != "local" {
		t.Fatalf("Lectura = %+v, want el timestamp y la fuente del envelope", got.Lectura)
	}
	if err := c.Olvidar("prenter-marketplace"); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := c.Leer("prenter-marketplace"); ok {
		t.Fatal("Leer tras Olvidar devolvió ok=true")
	}
	if err := c.Olvidar("prenter-marketplace"); err != nil {
		t.Fatalf("Olvidar de algo ausente no debe ser error: %v", err)
	}
}

// Caché AUSENTE ⇒ ok=false con motivo "" (la fila dice `no-leido`, nunca «0 entradas»).
func TestCacheAusenteNoTieneMotivo(t *testing.T) {
	c, _ := cacheEn(t)
	cat, motivo, ok := c.Leer("nadie")
	if ok || motivo != "" || cat.Entradas != nil {
		t.Fatalf("Leer = (%+v, %q, %v), want (vacío, \"\", false)", cat, motivo, ok)
	}
}

// E-48 · caché CORRUPTO ⇒ ok=false con el motivo VISIBLE; el próximo refresco lo sobreescribe.
// Nunca «0 entradas».
func TestCacheCorruptoNoEsCeroEntradas(t *testing.T) {
	c, dir := cacheEn(t)
	if err := os.WriteFile(filepath.Join(dir, "prenter-marketplace.json"), []byte(`{"version":1,"catalo`), 0o600); err != nil {
		t.Fatal(err)
	}
	cat, motivo, ok := c.Leer("prenter-marketplace")
	if ok {
		t.Fatal("Leer devolvió ok=true con caché corrupto")
	}
	if !strings.HasPrefix(motivo, "caché de catálogo ilegible:") {
		t.Fatalf("motivo = %q, want prefijo «caché de catálogo ilegible:»", motivo)
	}
	if cat.Entradas != nil {
		t.Fatalf("Entradas = %v, want nil", cat.Entradas)
	}
	// El próximo refresco lo sobreescribe sin intervención.
	if err := c.Guardar("prenter-marketplace", catSano("prenter-marketplace", "2026-07-25T15:00:00Z")); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := c.Leer("prenter-marketplace"); !ok {
		t.Fatal("tras el refresco el caché sigue ilegible")
	}
}

// E-49 · versión de caché desconocida ⇒ se DESCARTA y se reconstruye. Acá sí es legítimo: el
// dato ES derivable (a diferencia del registro, donde la clase declarada no lo es).
func TestCacheVersionDesconocidaSeDescarta(t *testing.T) {
	c, dir := cacheEn(t)
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(`{"version":99,"nombre":"x","catalogo":{"entradas":[]}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, motivo, ok := c.Leer("x")
	if ok {
		t.Fatal("un caché de versión desconocida no debe usarse")
	}
	if motivo != "caché de versión 99 desconocida (se re-leerá)" {
		t.Fatalf("motivo = %q", motivo)
	}
}

// E-50 · `Guardar` de un catálogo SIN lectura ⇒ error centinela y CERO archivos escritos.
// Cachear un «no sé» convertiría un fallo transitorio en un estado persistente.
func TestCacheRechazaEntradasNil(t *testing.T) {
	c, dir := cacheEn(t)
	err := c.Guardar("x", domain.Catalogo{Marketplace: "x", Lectura: domain.EstadoLectura{Tipo: domain.LecturaSinAcceso, Motivo: "gh: HTTP 404"}})
	if !errors.Is(err, ErrCacheSinEntradas) {
		t.Fatalf("err = %v, want ErrCacheSinEntradas", err)
	}
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 0 {
		t.Fatalf("se escribieron %d archivo(s): %v", len(entries), entries)
	}
}

// E-52 · nombres con `@`, espacios, unicode o `..`: el archivo cae SIEMPRE dentro del dir del
// caché, dos nombres que colapsan al mismo slug no se pisan, y el `nombre` se conserva crudo.
func TestCacheSlugSeguroYSinColision(t *testing.T) {
	c, dir := cacheEn(t)
	nombres := []string{"mi mkt", "mi-mkt", "ñandú@v2", "../escape", "…"}
	for _, n := range nombres {
		if err := c.Guardar(n, catSano(n, "2026-07-25T00:00:00Z")); err != nil {
			t.Fatalf("Guardar(%q): %v", n, err)
		}
	}
	// Todos los archivos caen DENTRO del dir del caché: ningún `..` sobrevive a domain.Slug.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(nombres) {
		t.Fatalf("archivos = %d, want %d (ninguno pisó a otro): %v", len(entries), len(nombres), nombresDe(entries))
	}
	padre, err := os.ReadDir(filepath.Dir(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range padre {
		if e.Name() != filepath.Base(dir) {
			t.Fatalf("algo escapó del dir del caché: %s", e.Name())
		}
	}
	// Cada uno se relee con SU nombre crudo intacto (el slug es de disco, no de presentación).
	for _, n := range nombres {
		cat, motivo, ok := c.Leer(n)
		if !ok {
			t.Fatalf("Leer(%q) = ok=false, motivo=%q", n, motivo)
		}
		if cat.Marketplace != n {
			t.Fatalf("Leer(%q).Marketplace = %q, want el nombre crudo", n, cat.Marketplace)
		}
	}
}

func nombresDe(entries []os.DirEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

// E-61 · dos escrituras + dos lecturas concurrentes del mismo catálogo: ninguna respuesta
// partida (temp+rename es atómico) y el archivo final parsea. `go test -race` limpio.
func TestCacheDosEscriturasConcurrentes(t *testing.T) {
	c, dir := cacheEn(t)
	if err := c.Guardar("x", catSano("x", "2026-07-25T00:00:00Z")); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cat := catSano("x", time.Date(2026, 7, 25, n, 0, 0, 0, time.UTC).Format(time.RFC3339))
			if err := c.Guardar("x", cat); err != nil {
				t.Errorf("Guardar concurrente: %v", err)
			}
		}(i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if cat, _, ok := c.Leer("x"); ok && len(cat.Entradas) != 2 {
				t.Errorf("lectura partida: %d entradas", len(cat.Entradas))
			}
		}()
	}
	wg.Wait()

	b, err := os.ReadFile(filepath.Join(dir, "x.json")) //nolint:gosec // G304: temporal del propio test.
	if err != nil {
		t.Fatal(err)
	}
	var env envelopeCache
	if uerr := json.Unmarshal(b, &env); uerr != nil {
		t.Fatalf("el archivo final no parsea (escritura partida): %v", uerr)
	}
	if len(env.Catalogo.Entradas) != 2 {
		t.Fatalf("el archivo final tiene %d entradas, want 2", len(env.Catalogo.Entradas))
	}
	// El último gana: `singleflight` queda como deuda con razón (design.md §4.3).
	entries, rerr := os.ReadDir(dir)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if len(entries) != 1 {
		t.Fatalf("quedaron temporales sin limpiar: %v", nombresDe(entries))
	}
}

// Un caché editado a mano con `entradas: null` no puede convertirse en un estado persistente.
func TestCacheConEntradasNullSeDescarta(t *testing.T) {
	c, dir := cacheEn(t)
	if err := os.WriteFile(filepath.Join(dir, "x.json"), []byte(`{"version":1,"nombre":"x","catalogo":{"entradas":null}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, motivo, ok := c.Leer("x")
	if ok || !strings.Contains(motivo, "sin lectura válida") {
		t.Fatalf("Leer = (ok=%v, motivo=%q), want descarte visible", ok, motivo)
	}
}
