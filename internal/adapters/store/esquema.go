package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// EsquemaActual es la forma que ESTE binario escribe en el registro de sesiones.
// Bumpearlo exige sumar un migrador a `migradores`: sin él, la cadena no llega y el
// arranque falla ruidoso en vez de leer a medias.
//
//	v1 — array desnudo de sesiones (la forma que se escribió hasta 2026-07-26)
//	v2 — sobre versionado; cada sesión lleva sus conversaciones (CV-D3)
const EsquemaActual = 2

// sobre es el envelope: la forma dice de qué versión es, quién la escribió y cuándo.
// Sin él, un archivo que cambia de forma es indistinguible de uno corrupto.
type sobre struct {
	SchemaVersion int             `json:"schema_version"`
	EscritoPor    string          `json:"escrito_por,omitempty"` // sello de build del que escribió.
	EscritoEn     string          `json:"escrito_en"`            // RFC3339 UTC.
	Sesiones      json.RawMessage `json:"sesiones"`
}

// ErrEsquemaSinVersion — un objeto JSON sin `schema_version` usable. NO se asume v1: la
// forma de v1 es un array, así que un objeto sin versión es un archivo de otra cosa.
var ErrEsquemaSinVersion = errors.New("store: el archivo es un objeto JSON sin schema_version — no se asume una versión")

// ErrNoEsJSON — el contenido no arranca con `[` ni con `{`. Va a cuarentena.
var ErrNoEsJSON = errors.New("store: el archivo no es JSON")

// detectarVersion mira el primer byte no-blanco y decide. No adivina: cada rama tiene un
// motivo y la que no lo tiene devuelve error.
//
//	'[' → array desnudo = v1 implícito, la forma que el registro escribía antes
//	'{' → sobre; manda su schema_version, y 0 o ausente es un error, no un v1 tácito
//	otro / vacío → no es JSON
func detectarVersion(b []byte) (int, error) {
	t := bytes.TrimSpace(b)
	if len(t) == 0 {
		return 0, fmt.Errorf("%w: está vacío", ErrNoEsJSON)
	}
	switch t[0] {
	case '[':
		return 1, nil
	case '{':
		var s sobre
		if err := json.Unmarshal(t, &s); err != nil {
			return 0, fmt.Errorf("%w: %w", ErrNoEsJSON, err)
		}
		if s.SchemaVersion <= 0 {
			return 0, ErrEsquemaSinVersion
		}
		return s.SchemaVersion, nil
	default:
		return 0, fmt.Errorf("%w: empieza con %q", ErrNoEsJSON, string(t[0]))
	}
}

// envolver arma el sobre de EsquemaActual alrededor de un payload ya serializado.
func envolver(payload json.RawMessage, sello string) ([]byte, error) {
	b, err := json.MarshalIndent(sobre{
		SchemaVersion: EsquemaActual,
		EscritoPor:    sello,
		EscritoEn:     time.Now().UTC().Format(time.RFC3339),
		Sesiones:      payload,
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("store: encode del sobre: %w", err)
	}
	return b, nil
}
