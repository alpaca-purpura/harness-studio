//go:build !windows

package selfupdate

// stubs_unix_test.go — fixtures por-OS de los tests compartidos (split del paquete
// 2026-08-13-compilacion-windows): en unix el bundle es bash y los ejecutables se
// reconocen por bit de ejecución. La contraparte vive en stubs_windows_test.go.

import (
	"path/filepath"
	"testing"
)

// stubTool crea un ejecutable falso que lookPathEn resuelve (bit de ejecución).
func stubTool(t *testing.T, dir, name string) {
	t.Helper()
	escribe(t, filepath.Join(dir, name), "#!/bin/sh\nexit 0\n", 0o755)
}

// aislarHome apunta el home del proceso a un TempDir (os.UserHomeDir lee HOME) para
// que la toolchain real del host no se cuele por candidatosPATHOS en casos negativos.
func aislarHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

// Stubs del script de bundle (scriptDeBundle = scripts/bundle.sh → bash).
func stubBundleNoop() string { return "#!/usr/bin/env bash\nexit 0\n" }

func stubBundleOK() string {
	return "#!/usr/bin/env bash\nmkdir -p bin\necho compilado > bin/" + nombreBinNuevo() + "\n"
}

func stubBundleFalla() string {
	return "#!/usr/bin/env bash\necho 'error TS2345: tipo roto' >&2\nexit 1\n"
}

func stubBundleCuelga() string { return "#!/usr/bin/env bash\nsleep 10\n" }
