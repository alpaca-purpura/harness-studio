package domain

import "math"

// telemetria_costo.go es el costeo con NUESTRO catálogo (arquitectura-modulo.md §2.3).
// Es puro: no toca red, ni disco, ni reloj. Acá viven los tres modos de fallo que toda la
// industria tiene y que este módulo no reproduce (D9.6) — cada uno con su test.

// PrecioModelo es una fila del catálogo, en USD por token. Punteros: un modelo que no
// publica tarifa de cache 1h tiene nil, NO 0. Aplanar los tiers es el bug phoenix#14314, y
// cotizar a 0 lo que no tiene tarifa es peor: se lee como «gratis».
type PrecioModelo struct {
	ModeloCanonico   string   `json:"modelo"`
	Proveedor        string   `json:"proveedor,omitempty"`
	Entrada          *float64 `json:"entrada,omitempty"`
	Salida           *float64 `json:"salida,omitempty"`
	CacheLectura     *float64 `json:"cache_lectura,omitempty"`
	CacheEscritura5m *float64 `json:"cache_escritura_5m,omitempty"`
	CacheEscritura1h *float64 `json:"cache_escritura_1h,omitempty"`
	Razonamiento     *float64 `json:"razonamiento,omitempty"`
	// UmbralContextoTok + SobreUmbral son el tier largo. NO se aplanan: si el prompt pasa
	// el umbral, se cotiza con SobreUmbral (phoenix#14314 aplana y subestima).
	UmbralContextoTok *int64        `json:"umbral_contexto_tokens,omitempty"`
	SobreUmbral       *PrecioModelo `json:"sobre_umbral,omitempty"`
}

// CostoCalculado es el resultado del costeo con nuestro catálogo.
//
// `Completo: false` marca que ALGÚN bucket tenía tokens y no tenía tarifa: el número es
// PARCIAL y la UI lo dice, no lo redondea ni lo presenta como total. El bucket sin tarifa
// sale nombrado en `SinTarifa` — **no se cobra a 0 en silencio**, que es la forma más
// barata de mentir sobre dinero.
type CostoCalculado struct {
	Micros    int64    `json:"micros"`
	Completo  bool     `json:"completo"`
	SinTarifa []string `json:"sin_tarifa,omitempty"`
	Version   string   `json:"catalogo_version,omitempty"`
	// SinNingunaTarifa distingue «no pude cotizar NADA» (el modelo no está en el catálogo)
	// de «cotizé 0 porque no hubo tokens». El primero debe viajar como null al wire.
	SinNingunaTarifa bool `json:"sin_ninguna_tarifa,omitempty"`
}

// CalcularCosto cotiza un uso con un precio. Es **pura**: no toca red, ni disco, ni reloj.
//
// Acá viven los tres modos de fallo que este módulo NO reproduce (D9.6, E8):
//
//  1. **El cache WRITE se cobra.** langfuse#14249 lo olvida y subestima ~28 %.
//  2. **No se suman buckets que se solapan** — de ahí el parámetro `a`: con
//     `AritmeticaInclusiva`, `cache_lectura ⊂ entrada`, y no restar duplica el conteo
//     (langfuse#12306, 2×).
//  3. **Los tiers NO se aplanan**: si hay `SobreUmbral` y el prompt pasa `UmbralContextoTok`,
//     se cotiza con ese (phoenix#14314 aplana y subestima justo los prompts caros).
//
// Y una cuarta, que es de honestidad y no de aritmética: **un bucket con tokens y sin tarifa
// NO se cobra a 0**. Sale nombrado en `SinTarifa` y `Completo` queda en false.
func CalcularCosto(t Tokens, p PrecioModelo, a Aritmetica) CostoCalculado {
	// El tramo se elige por el tamaño del PROMPT (lo que entra al modelo), que es el eje
	// sobre el que los proveedores tarifan el contexto largo: entrada + lo que se leyó de
	// cache + lo que se escribió a cache. La salida no cuenta para el umbral.
	prompt := valor(t.Entrada) + valor(t.CacheLectura) + valor(t.CacheEscritura5m) + valor(t.CacheEscritura1h)
	efectivo := p
	if p.SobreUmbral != nil && p.UmbralContextoTok != nil && prompt > *p.UmbralContextoTok {
		efectivo = *p.SobreUmbral
		efectivo.ModeloCanonico = p.ModeloCanonico
	}

	// Aritmética inclusiva: `cache_lectura` viene DENTRO de `entrada`. Se descuenta para no
	// cobrar los mismos tokens dos veces, y nunca por debajo de cero — un descuento que
	// deja negativo significa que el emisor reportó algo incoherente, y regalar crédito
	// sería tan falso como cobrar de más.
	entrada := valor(t.Entrada)
	if a == AritmeticaInclusiva {
		entrada -= valor(t.CacheLectura)
		if entrada < 0 {
			entrada = 0
		}
	}

	var usd float64
	var sinTarifa []string
	cobrados := 0

	cobrar := func(nombre string, tokens int64, tarifa *float64) {
		if tokens <= 0 {
			return // sin tokens no hay nada que cobrar ni nada que reportar como faltante.
		}
		if tarifa == nil {
			// Cuarta regla: sin tarifa NO es gratis. Se nombra y se marca incompleto.
			sinTarifa = append(sinTarifa, nombre)
			return
		}
		usd += float64(tokens) * *tarifa
		cobrados++
	}

	cobrar("entrada", entrada, efectivo.Entrada)
	cobrar("salida", valor(t.Salida), efectivo.Salida)
	cobrar("cache_lectura", valor(t.CacheLectura), efectivo.CacheLectura)
	// Los dos tramos del cache write se cobran por separado, cada uno con SU tarifa. Es la
	// regla 1 (langfuse#14249 no cobra ninguno) y a la vez lo que hace posible B1.
	cobrar("cache_escritura_5m", valor(t.CacheEscritura5m), efectivo.CacheEscritura5m)
	cobrar("cache_escritura_1h", valor(t.CacheEscritura1h), efectivo.CacheEscritura1h)
	cobrar("razonamiento", valor(t.Razonamiento), efectivo.Razonamiento)

	return CostoCalculado{
		Micros:    int64(math.Round(usd * 1e6)),
		Completo:  len(sinTarifa) == 0,
		SinTarifa: sinTarifa,
		// SinNingunaTarifa: había tokens que cobrar y no se pudo cobrar NINGUNO. El caller
		// debe mandar `null` al wire, no un 0 — un 0 se lee como «salió gratis».
		SinNingunaTarifa: cobrados == 0 && len(sinTarifa) > 0,
	}
}

func valor(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}
