// Package descubrimiento publica dónde está el daemon para que un hook que no conoce nuestras
// convenciones lo encuentre, en las tres plataformas (decisión A8).
//
// Vive en `os.UserConfigDir()/arnesia/daemon.json` y **no** en `~/.arnesia`: es el ÚNICO
// archivo que tiene que encontrar un proceso ajeno, y esa es la ruta que el sistema operativo
// publica para eso (`%APPDATA%\arnesia\` en Windows, `~/Library/Application Support/arnesia/`
// en macOS). El resto del estado sigue en `~/.arnesia`; migrarlo entero es otra deuda.
package descubrimiento

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// Ficha implementa ports.DescubrimientoDaemon sobre un archivo.
type Ficha struct{ ruta string }

var _ ports.DescubrimientoDaemon = (*Ficha)(nil)

// New arma la ficha. Una ruta vacía resuelve a `os.UserConfigDir()/arnesia/daemon.json`.
func New(ruta string) (*Ficha, error) {
	r, err := Resolver(ruta)
	if err != nil {
		return nil, err
	}
	return &Ficha{ruta: r}, nil
}

// Resolver devuelve la ruta canónica de la ficha. Se exporta porque el hook standalone la
// necesita sin construir el resto del adaptador.
func Resolver(ruta string) (string, error) {
	if ruta != "" {
		return ruta, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("descubrimiento: resolver config dir: %w", err)
	}
	return filepath.Join(dir, "arnesia", "daemon.json"), nil
}

// Ruta devuelve el archivo que esta ficha maneja.
func (f *Ficha) Ruta() string { return f.ruta }

// Publicar escribe la ficha de forma ATÓMICA (temp en el mismo directorio + rename), con
// permisos `0600` y el directorio `0700`.
//
// 🔴 **Se llama DESPUÉS de que el listener acepta conexiones, nunca antes.** Una ficha que
// nombra un puerto donde no escucha nadie es una mentira que el hook cobra en timeouts — y el
// hook tiene 250 ms de presupuesto total.
//
// El temp va en el MISMO directorio a propósito: `os.Rename` solo es atómico dentro del mismo
// sistema de archivos, y `/tmp` puede estar en otro.
func (f *Ficha) Publicar(_ context.Context, d domain.FichaDaemon) error {
	dir := filepath.Dir(f.ruta)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("descubrimiento: mkdir %s: %w", dir, err)
	}
	crudo, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("descubrimiento: codificar ficha: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".daemon-*.json")
	if err != nil {
		return fmt.Errorf("descubrimiento: temp en %s: %w", dir, err)
	}
	nombreTmp := tmp.Name()
	defer func() { _ = os.Remove(nombreTmp) }() // no-op si el rename ya lo movió.

	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("descubrimiento: permisos del temp: %w", err)
	}
	if _, err := tmp.Write(append(crudo, '\n')); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("descubrimiento: escribir temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("descubrimiento: cerrar temp: %w", err)
	}
	if err := os.Rename(nombreTmp, f.ruta); err != nil {
		return fmt.Errorf("descubrimiento: publicar %s: %w", f.ruta, err)
	}
	// El rename puede conservar los permisos del destino previo: se re-afirman.
	if err := os.Chmod(f.ruta, 0o600); err != nil {
		return fmt.Errorf("descubrimiento: permisos de la ficha: %w", err)
	}
	return nil
}

// Leer devuelve la ficha publicada, o `ErrSinFicha` si no hay.
//
// **Se relee en CADA invocación del hook, jamás se cachea**: el daemon puede haberse
// reiniciado en otro puerto entre dos turnos, y una ficha cacheada apuntaría al viejo.
func (f *Ficha) Leer() (domain.FichaDaemon, error) {
	crudo, err := os.ReadFile(f.ruta) //nolint:gosec // ruta derivada de os.UserConfigDir().
	if err != nil {
		return domain.FichaDaemon{}, ports.ErrSinFicha
	}
	var d domain.FichaDaemon
	if uerr := json.Unmarshal(crudo, &d); uerr != nil {
		// Una ficha ilegible es lo mismo que no tener ficha: el hook sale fail-open.
		return domain.FichaDaemon{}, ports.ErrSinFicha
	}
	if d.Endpoint == "" {
		return domain.FichaDaemon{}, ports.ErrSinFicha
	}
	return d, nil
}

// Retirar borra la ficha en el shutdown ordenado.
//
// Una ficha HUÉRFANA (daemon muerto sin shutdown) **no se detecta con el pid**: el hook
// simplemente falla el POST dentro de su presupuesto y sale fail-open. Es más simple y más
// robusto que un liveness check, que además podría dar un falso negativo si el pid se reusó.
func (f *Ficha) Retirar() error {
	if err := os.Remove(f.ruta); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("descubrimiento: retirar %s: %w", f.ruta, err)
	}
	return nil
}
