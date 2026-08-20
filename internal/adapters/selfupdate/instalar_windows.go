//go:build windows

package selfupdate

// instalar_windows.go — pasos ④ y ⑤ en Windows: DEGRADACIÓN HONESTA (decisión ① del
// paquete 2026-08-13-compilacion-windows). Windows no permite reemplazar un .exe
// mientras corre (la imagen queda mapeada como sección — el rename del camino unix
// falla con Access Denied), así que el paso ④ NO instala: deja el binario nuevo
// staged AL LADO del instalado y falla con la acción concreta. El reporte de 5 pasos
// muestra la verdad — build ✓ · verificar binario ✓ · instalar ✗ (con el path staged)
// · reiniciar no-corrido — jamás un «actualizado» que no actualizó («gris ≠ verde»).
// El swap automático (rename-trick NTFS: renombrar el .exe corriente SÍ es legal,
// sobrescribirlo no) es tech-debt registrada, no está construido.

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// nombreStaged — el binario nuevo queda ACÁ, junto al instalado, esperando el swap manual.
const nombreStaged = "arnesia-nuevo.exe"

// Instalar (Windows): copia el binario nuevo a <dir del exe>\arnesia-nuevo.exe
// (write-tmp → rename, igual de atómico que unix) y devuelve el error honesto con la
// acción. Un fallo de copia deja el directorio como estaba.
func (u *Updater) Instalar(_ context.Context) (string, error) {
	src, err := os.Open(u.binNuevo())
	if err != nil {
		return "", fmt.Errorf("abrir el binario nuevo: %w", err)
	}
	defer func() { _ = src.Close() }() // lectura: el Close sin error que importe.

	dir := filepath.Dir(u.exePath)
	staged := filepath.Join(dir, nombreStaged)
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
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("cerrar el binario nuevo: %w", err)
	}
	if err := os.Rename(tmp.Name(), staged); err != nil {
		return "", fmt.Errorf("dejar el binario staged en %s: %w", staged, err)
	}
	return "", fmt.Errorf(
		"self-update automático no disponible en Windows (un .exe en ejecución no se puede reemplazar): "+
			"el binario nuevo quedó en %s — cerrá la app y renombralo sobre %s",
		staged, u.exePath)
}

// Reiniciar (Windows): sin re-exec — unix reemplaza el proceso con syscall.Exec;
// relanzar acá el MISMO exe viejo confundiría («reinició pero no cambió nada»). El
// paso ⑤ nunca se agenda porque Instalar corta el flujo antes; si algo lo llama
// igual, el error es honesto.
func (u *Updater) Reiniciar() error {
	return errors.New("reinicio automático no disponible en Windows — cerrá y reabrí la app")
}
