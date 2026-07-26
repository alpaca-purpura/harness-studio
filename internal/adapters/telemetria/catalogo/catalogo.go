// Package catalogo embebe el catálogo de tarifas por token en el binario. Implementa
// ports.CatalogoPrecios (arquitectura-modulo.md §8).
//
// Cero post-install (D11): el catálogo viaja compilado, no se descarga al instalar ni al
// arrancar. El refresco por red existe pero va APAGADO por default (A11) — es egreso del
// daemon, y D13 dejó firmado que el egreso del daemon lo decide el operador.
//
// Lo que este paquete NUNCA hace: inventar una tarifa. Un modelo desconocido devuelve
// `ok=false` y el costo calculado queda `nil`, que la UI muestra como «sin dato». Un
// PrecioModelo en ceros costearía todo gratis en silencio, que es peor que no cotizar.
package catalogo

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// preciosJSON es el catálogo filtrado de LiteLLM (MIT). Se COMMITEA: el build es offline y
// reproducible, y `go generate` es un acto deliberado del que actualiza precios.
//
//go:embed precios.json
var preciosJSON []byte

// tarifa es la forma compacta del archivo embebido. Las claves cortas son para el
// presupuesto de binario (§11), no por gusto.
type tarifa struct {
	Proveedor        string   `json:"p,omitempty"`
	Entrada          *float64 `json:"in,omitempty"`
	Salida           *float64 `json:"out,omitempty"`
	CacheLectura     *float64 `json:"cr,omitempty"`
	CacheEscritura5m *float64 `json:"cw5,omitempty"`
	CacheEscritura1h *float64 `json:"cw1h,omitempty"`
	Razonamiento     *float64 `json:"rz,omitempty"`
	UmbralTokens     *int64   `json:"umbral,omitempty"`
	SobreUmbral      *tarifa  `json:"sobre,omitempty"`
}

type archivo struct {
	Version string            `json:"version"`
	Rev     string            `json:"rev"`
	Fuente  string            `json:"fuente"`
	Modelos map[string]tarifa `json:"modelos"`
}

// Catalogo es una vista de solo lectura sobre el archivo embebido (o sobre uno refrescado).
// Seguro para uso concurrente: el mapa se arma una vez y no se muta; `Refrescado` sí se
// protege con mutex porque el refresco opcional lo escribe.
type Catalogo struct {
	datos  archivo
	sha256 string

	mu         sync.RWMutex
	refrescado *time.Time
}

var _ ports.CatalogoPrecios = (*Catalogo)(nil)

var (
	unaVez   sync.Once
	embebido *Catalogo
	errEmbed error
)

// Embebido devuelve el catálogo compilado en el binario. Se decodifica una sola vez.
// Un archivo embebido ilegible es un bug de build, no un dato del usuario: el error se
// propaga por `Version().Modelos == 0` y `Precio` devuelve siempre `ok=false` — el daemon
// arranca igual y la UI dice que no hay catálogo, en vez de cotizar todo a cero.
func Embebido() *Catalogo {
	unaVez.Do(func() {
		suma := sha256.Sum256(preciosJSON)
		c := &Catalogo{sha256: hex.EncodeToString(suma[:])}
		if err := json.Unmarshal(preciosJSON, &c.datos); err != nil {
			errEmbed = fmt.Errorf("catalogo: precios.json embebido ilegible: %w", err)
			c.datos = archivo{Modelos: map[string]tarifa{}}
		}
		if c.datos.Modelos == nil {
			c.datos.Modelos = map[string]tarifa{}
		}
		embebido = c
	})
	return embebido
}

// ErrEmbebido devuelve el error de decodificación del catálogo embebido, o nil. Existe para
// que el composition root pueda loguearlo: un catálogo roto degrada, pero no en silencio.
func ErrEmbebido() error { Embebido(); return errEmbed }

// TamanoEmbebido son los bytes del `go:embed`. Presupuesto: ≤ 256 KB (§11).
func TamanoEmbebido() int { return len(preciosJSON) }

// Version describe con qué precios se está costeando. `Refrescado` nil = NUNCA se refrescó,
// y así se muestra: «precios del release», no «hace 0 h».
func (c *Catalogo) Version() domain.VersionCatalogoPrecios {
	c.mu.RLock()
	r := c.refrescado
	c.mu.RUnlock()
	return domain.VersionCatalogoPrecios{
		Version:    c.datos.Version,
		Rev:        c.datos.Rev,
		SHA256:     c.sha256,
		Modelos:    len(c.datos.Modelos),
		Refrescado: r,
	}
}

// MarcarRefrescado registra que el refresco opcional trajo datos. Lo llama el caller que
// refresca; el catálogo embebido nunca se marca solo.
func (c *Catalogo) MarcarRefrescado(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	u := t.UTC()
	c.refrescado = &u
}

// reAliasCloud reconoce los nombres de Bedrock/Vertex con prefijo de región. Portado del
// regex de Langfuse (MIT, E6): `(eu.|us.|apac.|global.)?anthropic.claude-…`.
var reAliasCloud = regexp.MustCompile(`^(?:(?:eu|us|apac|global|us-gov)\.)?anthropic\.(.+?)(?:-v\d+:\d+|:\d+)?$`)

// reFecha reconoce el sufijo de fecha que Anthropic pone en los ids concretos
// (`claude-haiku-4-5-20251001` → `claude-haiku-4-5`) y Vertex con `@`.
var reFecha = regexp.MustCompile(`[-@]\d{8}$`)

// Canonizar aplica los alias de cloud al nombre crudo del runtime.
//
// **Un nombre que no matcha se devuelve TAL CUAL.** No se inventa una canonización: si no
// sabemos a qué modelo corresponde, decirlo es más útil que adivinar y cotizar mal.
func (c *Catalogo) Canonizar(modelo string) string {
	m := strings.TrimSpace(modelo)
	if m == "" {
		return modelo
	}
	// 1. El nombre tal cual, que es el caso normal (LiteLLM ya trae las filas de Bedrock
	//    con su prefijo de región).
	if _, ok := c.datos.Modelos[m]; ok {
		return m
	}
	candidatos := []string{m}
	// 2. Prefijo de proveedor estilo `anthropic/claude-…` o `bedrock/…`.
	if i := strings.LastIndex(m, "/"); i >= 0 && i+1 < len(m) {
		candidatos = append(candidatos, m[i+1:])
	}
	// 3. Alias de cloud: `us.anthropic.claude-…-v1:0` → `claude-…`.
	for _, cand := range append([]string{}, candidatos...) {
		if g := reAliasCloud.FindStringSubmatch(cand); g != nil {
			candidatos = append(candidatos, g[1])
		}
	}
	// 4. Sufijo de fecha: `claude-haiku-4-5-20251001` → `claude-haiku-4-5`.
	for _, cand := range append([]string{}, candidatos...) {
		if sin := reFecha.ReplaceAllString(cand, ""); sin != cand {
			candidatos = append(candidatos, sin)
		}
	}
	for _, cand := range candidatos {
		if _, ok := c.datos.Modelos[cand]; ok {
			return cand
		}
	}
	return modelo // no matchea: se devuelve crudo y `Precio` dirá ok=false.
}

// Precio devuelve la tarifa de un modelo YA canonizado. `ok=false` cuando no está: jamás un
// PrecioModelo en ceros, que cotizaría todo gratis sin que nadie se entere.
func (c *Catalogo) Precio(modeloCanonico string) (domain.PrecioModelo, bool) {
	t, ok := c.datos.Modelos[modeloCanonico]
	if !ok {
		// Segundo intento canonizando: un caller que pase el nombre crudo obtiene lo
		// mismo que uno que canonizó primero. No es laxitud, es que una API con dos
		// resultados distintos para el mismo modelo es una trampa.
		if can := c.Canonizar(modeloCanonico); can != modeloCanonico {
			if t, ok = c.datos.Modelos[can]; ok {
				modeloCanonico = can
			}
		}
	}
	if !ok {
		return domain.PrecioModelo{}, false
	}
	return aPrecio(modeloCanonico, t), true
}

func aPrecio(nombre string, t tarifa) domain.PrecioModelo {
	p := domain.PrecioModelo{
		ModeloCanonico:    nombre,
		Proveedor:         t.Proveedor,
		Entrada:           t.Entrada,
		Salida:            t.Salida,
		CacheLectura:      t.CacheLectura,
		CacheEscritura5m:  t.CacheEscritura5m,
		CacheEscritura1h:  t.CacheEscritura1h,
		Razonamiento:      t.Razonamiento,
		UmbralContextoTok: t.UmbralTokens,
	}
	if t.SobreUmbral != nil {
		sobre := aPrecio(nombre, *t.SobreUmbral)
		p.SobreUmbral = &sobre
	}
	return p
}
