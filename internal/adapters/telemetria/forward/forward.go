// Package forward es el escape hatch del OPERADOR (C2 / D13), **apagado por default**.
//
// Dos reglas que lo definen y que no son negociables:
//
//  1. **Solo el operador lo enciende** — por flag o variable de entorno del daemon. Nunca
//     desde la API, nunca desde un arnés. Un arnés no puede alcanzar ni configurar esto.
//  2. **Reenvía el evento YA PROYECTADO**, jamás el cuerpo OTLP crudo. Reenviar crudo
//     exportaría el email y los identificadores de cuenta de quien corra el arnés, que
//     llegan en CADA punto (medido, INFORME §V6).
//
// Ojo con no confundir los dos niveles de egreso (D13): `telemetria-no-egresa` ata al ARNÉS;
// esto es el daemon, y es del operador.
package forward

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Forward implementa ports.ForwardOTLP.
type Forward struct {
	destino string
	cliente *http.Client
}

var _ ports.ForwardOTLP = (*Forward)(nil)

// New arma el forward. **Un destino vacío devuelve un forward APAGADO** — que es el default y
// el único estado que se alcanza sin una decisión explícita del operador.
func New(destino string, cliente *http.Client) *Forward {
	if cliente == nil {
		cliente = &http.Client{Timeout: 10 * time.Second}
	}
	return &Forward{destino: destino, cliente: cliente}
}

// Activo reporta si el forward está encendido. Se muestra en la UI mientras lo esté: un
// egreso silencioso sería exactamente lo que este módulo existe para no hacer.
func (f *Forward) Activo() bool { return f.destino != "" }

// Destino devuelve el endpoint configurado, para mostrarlo.
func (f *Forward) Destino() string { return f.destino }

// Enviar reenvía los eventos **ya proyectados**.
//
// Apagado: devuelve nil **sin abrir un socket**. No es una optimización — es el invariante:
// sin la decisión del operador, cero conexiones salientes.
func (f *Forward) Enviar(ctx context.Context, evs []domain.EventoTelemetria) error {
	if !f.Activo() || len(evs) == 0 {
		return nil
	}
	// Lo que sale es el evento canónico, que por construcción no tiene identidad de cuenta
	// ni contenido: el filtro ya ocurrió en la puerta de ingesta y acá no se deshace.
	cuerpo, err := json.Marshal(evs)
	if err != nil {
		return fmt.Errorf("forward: codificar: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.destino, bytes.NewReader(cuerpo))
	if err != nil {
		return fmt.Errorf("forward: request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := f.cliente.Do(req)
	if err != nil {
		return fmt.Errorf("forward: enviar a %s: %w", f.destino, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("forward: %s respondió %s", f.destino, resp.Status)
	}
	return nil
}
