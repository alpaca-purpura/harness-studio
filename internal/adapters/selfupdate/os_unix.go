//go:build !windows

package selfupdate

// os_unix.go — la mitad unix de la mecánica por-OS del self-update (paquete
// 2026-08-13-compilacion-windows): nombre del binario, script de bundle, toolchain
// exigida, veredicto de ejecutabilidad, resolución en PATH y el subproceso de build
// con kill-de-grupo. La contraparte vive en os_windows.go — mismas firmas, semántica
// del OS real (jamás un stub que finja paridad).

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// nombreBinNuevo — nombre del binario que el bundle deja en <repo>/bin/.
func nombreBinNuevo() string { return "arnesia" }

// scriptDeBundle — ruta relativa del script de build que Verificar exige y Build corre.
func scriptDeBundle() string { return filepath.Join("scripts", "bundle.sh") }

// descripcionBundle — cómo se nombra el paso de build en los detalles humanos.
const descripcionBundle = "scripts/bundle.sh --daemon-only"

// toolchainRequerida — herramientas que Verificar exige resolubles en PATH (el bundle
// unix es bash → bash es parte del terreno).
func toolchainRequerida() []string { return []string{"go", "pnpm", "bash"} }

// esEjecutable — en unix el veredicto es el bit de ejecución del modo.
func esEjecutable(modo fs.FileMode, _ string) bool { return modo.Perm()&0o111 != 0 }

// ejecutableEnDir — candidato dir/tool con bit de ejecución (semántica exec.LookPath
// unix, contra un PATH explícito).
func ejecutableEnDir(dir, tool string) (string, bool) {
	candidato := filepath.Join(dir, tool)
	if st, err := os.Stat(candidato); err == nil && !st.IsDir() && st.Mode().Perm()&0o111 != 0 {
		return candidato, true
	}
	return "", false
}

// candidatosPATHOS — ubicaciones ESTÁNDAR de instalación de go/pnpm que un proceso
// lanzado desde el launcher gráfico no hereda: el `.desktop` arranca con el PATH de la
// sesión (systemd/display-manager), fijado al iniciar sesión — jamás lee
// ~/.profile/~/.bashrc como sí hace una terminal interactiva. Root cause del bug real
// (no un problema de build viejo): "toolchain incompleta: X no está en PATH" con la
// herramienta instalada y en el PATH de CUALQUIER terminal del mismo usuario. El fix
// original (HS) solo cubrió go; pnpm pegaba el mismo síntoma vía nvm (ver
// nvmBinsInstalados) y quedó afuera hasta ahora.
func candidatosPATHOS(home string) []string {
	cands := []string{
		filepath.Join(home, ".local", "go", "bin"), // convención de este repo (ver ~/.profile)
		filepath.Join(home, "go", "bin"),           // convención oficial go.dev (go install)
		"/usr/local/go/bin",                        // convención oficial go.dev (tarball)
	}
	return append(cands, nvmBinsInstalados(home)...)
}

// nvmBinsInstalados — mismo root cause que go (ver candidatosPATHOS): el launcher gráfico
// no corre ~/.bashrc, así que nvm.sh nunca carga y el pnpm de la versión de node activa
// queda fuera del PATH aunque esté instalado (ver ~/.nvm/versions/node/*/bin). nvm no
// fija un symlink "current" — cada versión vive en su propio dir versionado — así que se
// agregan TODOS los bin/ instalados: cualquiera resuelve pnpm/node/npm igual de bien
// para bundle.sh.
func nvmBinsInstalados(home string) []string {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(home, ".nvm")
	}
	versionesDir := filepath.Join(nvmDir, "versions", "node")
	entries, err := os.ReadDir(versionesDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(versionesDir, e.Name(), "bin"))
		}
	}
	return out
}

// comandoBundle arma el subproceso del paso ② (RF-104): bash scripts/bundle.sh
// --daemon-only con cwd=repo. La cancelación mata el GRUPO entero (bash + go/pnpm/vite
// hijos): matar solo a bash dejaría a los hijos corriendo Y sosteniendo el pipe
// (CombinedOutput jamás retornaría). WaitDelay corta el pipe si algún nieto sobrevive
// al SIGKILL.
func comandoBundle(ctx context.Context, repo string) *exec.Cmd {
	//nolint:gosec // G204: es EL diseño (RF-104 ②): correr el bundle del repo CONFIGURADO
	// al daemon (flag/env, jamás del request — RF-106); Verificar ya ancló el árbol al módulo esperado.
	cmd := exec.CommandContext(ctx, "bash", scriptDeBundle(), "--daemon-only")
	cmd.Dir = repo
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	return cmd
}
