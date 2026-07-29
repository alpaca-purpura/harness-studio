//go:build windows

package filelock

import (
	"os"

	"golang.org/x/sys/windows"
)

// lock toma un LockFileEx exclusivo, bloqueante (sin LOCKFILE_FAIL_IMMEDIATELY), sobre f.
// El rango cubierto (0, ^uint32(0), ^uint32(0)) alcanza para un archivo `.lock` que ningún
// otro código toca ni lee: solo existe para que su lock exista.
func lock(f *os.File) error {
	ol := new(windows.Overlapped)
	return windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, ^uint32(0), ^uint32(0), ol)
}

func unlock(f *os.File) error {
	ol := new(windows.Overlapped)
	return windows.UnlockFileEx(windows.Handle(f.Fd()), 0, ^uint32(0), ^uint32(0), ol)
}
