// Package filelock da exclusión mutua CROSS-PROCESO alrededor del ciclo
// recargar-desde-disco→mutar→guardar de los stores JSON de arnesia (Fase 1, decisiones.md D2).
//
// Por qué existe: `portafolio.Store`, `marketplace.Store` y `store.ArnesRegistry` cargan el
// archivo entero a un mapa en memoria una vez, en `New*`, y lo reescriben entero en cada
// mutación. Eso es correcto dentro de UN proceso (un `sync.Mutex` alcanza), pero el daemon
// (`arnesia serve`) y un subcomando CLI standalone (`arnesia portafolio agregar`,
// `arnesia sesiones recalibrar-llaves`) son DOS procesos con DOS cachés en memoria
// independientes escribiendo el mismo archivo — confirmado leyendo `cmd/arnesia/main.go`:
// `runPortafolio` abre su propia instancia del store, separada de la que vive en `runServe`.
//
// `Guard` toma un lock exclusivo, bloqueante, sobre un sidecar `<path>.lock` — nunca el
// archivo de datos en sí, que sigue reemplazándose por temp+rename dentro de la sección
// crítica. El lock es advisory (POSIX flock / Windows LockFileEx): solo sirve entre
// procesos que lo pidan, que es exactamente el conjunto de procesos que este binario
// controla (los 4 stores file-backed).
package filelock

import (
	"fmt"
	"os"
)

// Guard abre (o crea) `<path>.lock`, toma el lock exclusivo — bloqueando hasta obtenerlo,
// nunca fallando por contención — corre fn() adentro, y libera el lock siempre, incluso si
// fn() devuelve error. El archivo de datos en `path` no se toca acá: fn() es responsable de
// leerlo/escribirlo.
func Guard(path string, fn func() error) error {
	lockPath := path + ".lock"
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600) //nolint:gosec // sidecar propio, ruta derivada del store que llama.
	if err != nil {
		return fmt.Errorf("filelock: abrir %s: %w", lockPath, err)
	}
	defer func() { _ = f.Close() }()

	if err := lock(f); err != nil {
		return fmt.Errorf("filelock: lock %s: %w", lockPath, err)
	}
	defer func() { _ = unlock(f) }()

	return fn()
}
