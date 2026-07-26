// Command bajar-catalogo baja `model_prices_and_context_window.json` de un REV FIJADO de
// LiteLLM (MIT, raíz del repo — el carve-out `enterprise/` no toca este archivo), lo filtra
// a los proveedores que este árbol puede ver y lo compacta a `precios.json`.
//
// Se corre a mano (`go generate ./internal/adapters/telemetria/catalogo/`), NUNCA como paso
// del build: el binario se compila offline y reproducible, y actualizar precios es un acto
// deliberado del que los actualiza — con su diff a la vista.
//
// Lo que este generador NO hace, y es el punto: **no aplana los tiers**. ccusage colapsa
// todo a un solo `cache_create` y pierde `cache_creation_input_token_cost_above_1hr`, que es
// justo el campo que habilita el detector B1 (re-warm por TTL). Acá viaja entero.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

// proveedoresEnUso es el filtro. La lista completa de LiteLLM son ~2 984 modelos (1,6 MB):
// embeber eso sería pagar peso por proveedores que este árbol no ve nunca.
var proveedoresEnUso = map[string]bool{
	"anthropic":                  true,
	"bedrock":                    true,
	"bedrock_converse":           true,
	"vertex_ai-anthropic_models": true,
	"openai":                     true,
	"gemini":                     true,
	"vertex_ai-language-models":  true,
	"deepseek":                   true,
	"xai":                        true,
	"mistral":                    true,
	"groq":                       true,
}

// filaLiteLLM es el subconjunto del schema de LiteLLM que nos interesa. Todo lo demás
// (supports_*, max_tokens, search_context_cost…) se descarta: no cotiza tokens.
type filaLiteLLM struct {
	Proveedor string `json:"litellm_provider"`
	Modo      string `json:"mode"`

	Entrada      *float64 `json:"input_cost_per_token"`
	Salida       *float64 `json:"output_cost_per_token"`
	CacheLectura *float64 `json:"cache_read_input_token_cost"`
	// CacheEscritura es el tramo de 5 minutos (el default de Anthropic).
	CacheEscritura *float64 `json:"cache_creation_input_token_cost"`
	// CacheEscritura1h es EL campo que ccusage descarta (E3). Sin él, B1 no existe.
	CacheEscritura1h *float64 `json:"cache_creation_input_token_cost_above_1hr"`
	Razonamiento     *float64 `json:"reasoning_cost_per_token"`

	// Tramo largo de contexto. LiteLLM lo expresa con umbrales distintos por familia
	// (200k · 272k · 512k · 256k · 128k); se toma el primero que exista, en ese orden.
	Entrada200k          *float64 `json:"input_cost_per_token_above_200k_tokens"`
	Salida200k           *float64 `json:"output_cost_per_token_above_200k_tokens"`
	CacheLectura200k     *float64 `json:"cache_read_input_token_cost_above_200k_tokens"`
	CacheEscritura200k   *float64 `json:"cache_creation_input_token_cost_above_200k_tokens"`
	CacheEscritura1h200k *float64 `json:"cache_creation_input_token_cost_above_1hr_above_200k_tokens"`

	Entrada272k        *float64 `json:"input_cost_per_token_above_272k_tokens"`
	Salida272k         *float64 `json:"output_cost_per_token_above_272k_tokens"`
	CacheLectura272k   *float64 `json:"cache_read_input_token_cost_above_272k_tokens"`
	CacheEscritura272k *float64 `json:"cache_creation_input_token_cost_above_272k_tokens"`

	Entrada128k *float64 `json:"input_cost_per_token_above_128k_tokens"`
	Salida128k  *float64 `json:"output_cost_per_token_above_128k_tokens"`
}

// tarifaSalida es la forma compacta que se embebe. Claves cortas a propósito: el archivo
// viaja dentro del binario y cada byte cuenta contra el presupuesto (§11).
type tarifaSalida struct {
	Proveedor        string        `json:"p,omitempty"`
	Entrada          *float64      `json:"in,omitempty"`
	Salida           *float64      `json:"out,omitempty"`
	CacheLectura     *float64      `json:"cr,omitempty"`
	CacheEscritura5m *float64      `json:"cw5,omitempty"`
	CacheEscritura1h *float64      `json:"cw1h,omitempty"`
	Razonamiento     *float64      `json:"rz,omitempty"`
	UmbralTokens     *int64        `json:"umbral,omitempty"`
	SobreUmbral      *tarifaSalida `json:"sobre,omitempty"`
}

type archivoSalida struct {
	Version string                  `json:"version"`
	Rev     string                  `json:"rev"`
	Fuente  string                  `json:"fuente"`
	Bajado  string                  `json:"bajado"`
	Modelos map[string]tarifaSalida `json:"modelos"`
}

func main() {
	rev := flag.String("rev", "", "sha del commit de LiteLLM a fijar (obligatorio: sin rev no es reproducible)")
	fecha := flag.String("fecha", "", "fecha del rev, YYYY-MM-DD (obligatorio: es la `version` que se muestra en la UI)")
	out := flag.String("out", "precios.json", "archivo de salida")
	flag.Parse()

	if *rev == "" || *fecha == "" {
		fmt.Fprintln(os.Stderr, "bajar-catalogo: -rev y -fecha son obligatorios — un catálogo sin procedencia no se puede auditar")
		os.Exit(2)
	}

	url := "https://raw.githubusercontent.com/BerriAI/litellm/" + *rev + "/model_prices_and_context_window.json"
	crudo, err := bajar(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bajar-catalogo: %v\n", err)
		os.Exit(1)
	}

	var origen map[string]json.RawMessage
	if uerr := json.Unmarshal(crudo, &origen); uerr != nil {
		fmt.Fprintf(os.Stderr, "bajar-catalogo: decodificar: %v\n", uerr)
		os.Exit(1)
	}

	salida := archivoSalida{
		Version: *fecha,
		Rev:     *rev,
		Fuente:  "BerriAI/litellm · model_prices_and_context_window.json (MIT)",
		Bajado:  time.Now().UTC().Format(time.RFC3339),
		Modelos: map[string]tarifaSalida{},
	}
	for nombre, raw := range origen {
		if nombre == "sample_spec" {
			continue
		}
		var f filaLiteLLM
		if json.Unmarshal(raw, &f) != nil {
			continue // una fila con forma rara se descarta; no se adivina.
		}
		if !proveedoresEnUso[f.Proveedor] || (f.Modo != "chat" && f.Modo != "responses") {
			continue
		}
		t := convertir(f)
		if t.Entrada == nil && t.Salida == nil {
			continue // sin tarifa de entrada ni de salida no cotiza nada: no ocupa bytes.
		}
		salida.Modelos[nombre] = t
	}

	if len(salida.Modelos) == 0 {
		fmt.Fprintln(os.Stderr, "bajar-catalogo: el filtro dejó 0 modelos — no se escribe un catálogo vacío")
		os.Exit(1)
	}

	buf, merr := json.Marshal(salida)
	if merr != nil {
		fmt.Fprintf(os.Stderr, "bajar-catalogo: codificar: %v\n", merr)
		os.Exit(1)
	}
	if werr := os.WriteFile(*out, append(buf, '\n'), 0o644); werr != nil { //nolint:gosec // dato público, no secreto.
		fmt.Fprintf(os.Stderr, "bajar-catalogo: escribir %s: %v\n", *out, werr)
		os.Exit(1)
	}
	nombres := make([]string, 0, len(salida.Modelos))
	for n := range salida.Modelos {
		nombres = append(nombres, n)
	}
	sort.Strings(nombres)
	fmt.Printf("bajar-catalogo: %d modelos · %d bytes · rev %s\n", len(nombres), len(buf)+1, *rev)
}

func bajar(url string) ([]byte, error) {
	cl := &http.Client{Timeout: 60 * time.Second}
	resp, err := cl.Get(url) //nolint:noctx // herramienta de generación manual, no runtime.
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

// convertir arma la tarifa compacta y, si la familia declara un tramo largo, lo conserva
// COMO TRAMO — no lo aplana ni lo promedia (phoenix#14314).
func convertir(f filaLiteLLM) tarifaSalida {
	t := tarifaSalida{
		Proveedor:        proveedorCorto(f.Proveedor),
		Entrada:          f.Entrada,
		Salida:           f.Salida,
		CacheLectura:     f.CacheLectura,
		CacheEscritura5m: f.CacheEscritura,
		CacheEscritura1h: f.CacheEscritura1h,
		Razonamiento:     f.Razonamiento,
	}
	type tramo struct {
		umbral                 int64
		in, out, cr, cw5, cw1h *float64
	}
	candidatos := []tramo{
		{200_000, f.Entrada200k, f.Salida200k, f.CacheLectura200k, f.CacheEscritura200k, f.CacheEscritura1h200k},
		{272_000, f.Entrada272k, f.Salida272k, f.CacheLectura272k, f.CacheEscritura272k, nil},
		{128_000, f.Entrada128k, f.Salida128k, nil, nil, nil},
	}
	for _, c := range candidatos {
		if c.in == nil && c.out == nil {
			continue
		}
		u := c.umbral
		t.UmbralTokens = &u
		t.SobreUmbral = &tarifaSalida{
			Entrada: primero(c.in, f.Entrada), Salida: primero(c.out, f.Salida),
			CacheLectura: primero(c.cr, f.CacheLectura), CacheEscritura5m: primero(c.cw5, f.CacheEscritura),
			CacheEscritura1h: primero(c.cw1h, f.CacheEscritura1h),
		}
		break
	}
	return t
}

func primero(a, b *float64) *float64 {
	if a != nil {
		return a
	}
	return b
}

// proveedorCorto colapsa las variantes de un mismo proveedor a una etiqueta legible. Es
// solo para mostrar: la canonización de NOMBRES DE MODELO vive en catalogo.go.
func proveedorCorto(p string) string {
	switch {
	case strings.HasPrefix(p, "bedrock"):
		return "bedrock"
	case strings.HasPrefix(p, "vertex_ai"):
		return "vertex"
	default:
		return p
	}
}
