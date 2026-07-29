//go:build unix

package filelock

import (
	"os"

	"golang.org/x/sys/unix"
)

// lock toma un flock(2) exclusivo, bloqueante, sobre f.
func lock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_EX)
}

// unlock libera el flock. Cerrar el fd (que Guard ya hace via defer) también lo liberaría,
// pero ser explícito documenta la simetría con lock y no depende del orden de los defer.
func unlock(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}
