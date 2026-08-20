//go:build !windows

package selfupdate

// instalar_unix.go — pasos ④ (instalar) y ⑤ (reiniciar) en unix: el rename atómico
// sobre el ejecutable corriente ES legal (el inode viejo sigue mapeado; /proc/self/exe
// pasa a «(deleted)») y el re-exec reemplaza el proceso. Movidos verbatim desde
// updater.go en el split por-OS (paquete 2026-08-13-compilacion-windows).

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// Instalar reemplaza el ejecutable corriente (paso ④): write a tmp EN EL MISMO
// directorio (mismo filesystem) → chmod 0755 → rename atómico. Cualquier fallo limpia
// el tmp y deja el binario instalado INTACTO.
func (u *Updater) Instalar(_ context.Context) (string, error) {
	src, err := os.Open(u.binNuevo())
	if err != nil {
		return "", fmt.Errorf("abrir el binario nuevo: %w", err)
	}
	defer func() { _ = src.Close() }() // lectura: el Close sin error que importe.

	dir := filepath.Dir(u.exePath)
	tmp, err := os.CreateTemp(dir, ".arnesia-nuevo-*")
	if err != nil {
		return "", fmt.Errorf("el directorio de instalación %s no acepta escritura: %w", dir, err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) // no-op tras el rename; limpia el resto de caminos.
	}()
	if _, err := io.Copy(tmp, src); err != nil {
		return "", fmt.Errorf("copiar el binario nuevo: %w", err)
	}
	if err := tmp.Chmod(0o755); err != nil {
		return "", fmt.Errorf("chmod del binario nuevo: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("cerrar el binario nuevo: %w", err)
	}
	if err := os.Rename(tmp.Name(), u.exePath); err != nil {
		return "", fmt.Errorf("rename atómico sobre %s: %w", u.exePath, err)
	}
	return fmt.Sprintf("instalado en %s (rename atómico, sin sudo)", u.exePath), nil
}

// Reiniciar re-ejecuta el daemon (paso ⑤): mismo path (el capturado al construir),
// mismos args, mismo entorno. En Linux exec reemplaza el proceso — no retorna si va bien.
func (u *Updater) Reiniciar() error {
	//nolint:gosec // G204: re-exec de SÍ MISMO (RF-105) — ruta capturada al boot, args propios.
	return syscall.Exec(u.exePath, os.Args, os.Environ())
}
