//go:build windows

package selfupdate

// updater_windows_test.go — la degradación honesta de Instalar en Windows (decisión ①
// del paquete 2026-08-13-compilacion-windows): el paso ④ NO reemplaza el .exe en
// ejecución — deja el binario staged al lado y falla con la acción concreta.

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstalarWindowsDegradaHonesto(t *testing.T) {
	repo := t.TempDir()
	escribe(t, filepath.Join(repo, "bin", nombreBinNuevo()), "NUEVO", 0o755)
	exe := exeFake(t, "VIEJO")
	u := &Updater{repo: repo, exePath: exe}

	_, err := u.Instalar(context.Background())
	if err == nil {
		t.Fatal("en Windows instalar debe degradar honesto (error con la acción), jamás decir que instaló")
	}
	if !strings.Contains(err.Error(), nombreStaged) {
		t.Fatalf("el error debe nombrar el binario staged (%s), tengo %q", nombreStaged, err)
	}

	staged := filepath.Join(filepath.Dir(exe), nombreStaged)
	if got := leeExe(t, staged); got != "NUEVO" {
		t.Fatalf("el staged debe ser el binario nuevo, tengo %q", got)
	}
	if got := leeExe(t, exe); got != "VIEJO" {
		t.Fatalf("el ejecutable corriente debe quedar INTACTO, tengo %q", got)
	}
	// Sin residuos: el tmp del write-tmp→rename no queda tirado (el glob con punto
	// inicial NO matchea arnesia-nuevo.exe).
	restos, err := filepath.Glob(filepath.Join(filepath.Dir(exe), ".arnesia-nuevo-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(restos) != 0 {
		t.Fatalf("quedaron tmp sin limpiar: %v", restos)
	}
}

func TestReiniciarWindowsErrorHonesto(t *testing.T) {
	u := &Updater{exePath: exeFake(t, "VIEJO")}
	if err := u.Reiniciar(); err == nil {
		t.Fatal("Reiniciar en Windows debe devolver error honesto (sin re-exec)")
	}
}
