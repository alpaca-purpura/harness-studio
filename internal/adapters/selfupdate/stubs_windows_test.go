//go:build windows

package selfupdate

// stubs_windows_test.go — fixtures por-OS de los tests compartidos (split del paquete
// 2026-08-13-compilacion-windows): en Windows el bundle es python y los ejecutables se
// resuelven por PATHEXT (sin bits de modo). La contraparte vive en stubs_unix_test.go.

import (
	"path/filepath"
	"testing"
)

// stubTool crea un ejecutable falso que lookPathEn resuelve — extensión .bat (PATHEXT),
// jamás se ejecuta (lookPathEn solo stat-ea).
func stubTool(t *testing.T, dir, name string) {
	t.Helper()
	escribe(t, filepath.Join(dir, name+".bat"), "@exit /b 0\r\n", 0o644)
}

// aislarHome — os.UserHomeDir lee USERPROFILE; candidatosPATHOS además lee APPDATA y
// LOCALAPPDATA (npm/pnpm). Se redirigen los tres para que la toolchain real del host
// no se cuele en los casos negativos.
func aislarHome(t *testing.T) {
	t.Helper()
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("LOCALAPPDATA", t.TempDir())
}

// Stubs del script de bundle (scriptDeBundle = scripts/bundle.py → python real del host).
func stubBundleNoop() string { return "raise SystemExit(0)\n" }

func stubBundleOK() string {
	return "import os\n" +
		"os.makedirs('bin', exist_ok=True)\n" +
		"open(os.path.join('bin', '" + nombreBinNuevo() + "'), 'w').write('compilado')\n"
}

func stubBundleFalla() string {
	return "import sys\nprint('error TS2345: tipo roto', file=sys.stderr)\nsys.exit(1)\n"
}

func stubBundleCuelga() string { return "import time\ntime.sleep(10)\n" }
