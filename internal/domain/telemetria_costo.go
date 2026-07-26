package domain

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
