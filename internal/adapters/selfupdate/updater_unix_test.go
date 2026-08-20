//go:build !windows

package selfupdate

// updater_unix_test.go — casos cuya semántica SOLO existe en unix (split
// 2026-08-13-compilacion-windows): chmod 0555 como dir no-escribible, el bit de
// ejecución como veredicto, el rename atómico de Instalar y los fallbacks de PATH
// del launcher gráfico (~/.local/go/bin, ~/.nvm).

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func soloLectura(t *testing.T, dir string) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root escribe en cualquier lado — el caso no-escribible no aplica")
	}
	if err := os.Chmod(dir, 0o555); err != nil { //nolint:gosec // G302: el 0555 ES el caso probado (dir no escribible).
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) }) //nolint:gosec // G302: restaura el TempDir del test para su limpieza.
}

func TestVersionNoEscribible(t *testing.T) {
	exe := exeFake(t, "binario corriente")
	soloLectura(t, filepath.Dir(exe))
	u := &Updater{exePath: exe}
	if u.Version().Escribible {
		t.Fatal("un dir 0555 debe reportar Escribible=false")
	}
}

func TestVerificarFallbackGoLocal(t *testing.T) {
	// Bug real (no de staleness): un proceso lanzado desde el launcher gráfico NO
	// hereda ~/.profile/~/.bashrc — go queda fuera de PATH aunque cualquier
	// terminal lo vea. pathAumentado() debe encontrarlo igual vía la convención
	// ~/.local/go/bin (ver ~/.profile) sin que el operador reinicie sesión.
	home := t.TempDir()
	t.Setenv("HOME", home)
	golangStub := filepath.Join(home, ".local", "go", "bin", "go")
	escribe(t, golangStub, "#!/bin/sh\nexit 0\n", 0o755)
	stubs := t.TempDir()
	for _, tool := range []string{"pnpm", "bash"} { // go NO va acá: debe venir del fallback.
		escribe(t, filepath.Join(stubs, tool), "#!/bin/sh\nexit 0\n", 0o755)
	}
	t.Setenv("PATH", stubs)
	u := &Updater{repo: repoFake(t, "")}
	if _, err := u.Verificar(context.Background()); err != nil {
		t.Fatalf("go en ~/.local/go/bin debe resolverse vía fallback: %v", err)
	}
}

func TestVerificarFallbackNvmPnpm(t *testing.T) {
	// Mismo root cause que el caso go de arriba: pnpm instalado vía nvm vive en
	// ~/.nvm/versions/node/vX.Y.Z/bin, invisible para un proceso lanzado sin
	// ~/.bashrc. pathAumentado() debe encontrarlo igual, sin symlink "current".
	home := t.TempDir()
	t.Setenv("HOME", home)
	// NVM_DIR manda sobre ~/.nvm (comportamiento correcto de producción) — pinnearlo
	// al home fake: los runners de GitHub traen nvm SIN versiones con NVM_DIR global
	// (/home/runner/.nvm), que secuestraba el fallback y el stub nunca se encontraba
	// (CI rojo 2026-07-17); en dev, un NVM_DIR real con pnpm enmascaraba el caso.
	t.Setenv("NVM_DIR", filepath.Join(home, ".nvm"))
	pnpmStub := filepath.Join(home, ".nvm", "versions", "node", "v24.18.0", "bin", "pnpm")
	escribe(t, pnpmStub, "#!/bin/sh\nexit 0\n", 0o755)
	stubs := t.TempDir()
	for _, tool := range []string{"go", "bash"} { // pnpm NO va acá: debe venir del fallback nvm.
		escribe(t, filepath.Join(stubs, tool), "#!/bin/sh\nexit 0\n", 0o755)
	}
	t.Setenv("PATH", stubs)
	u := &Updater{repo: repoFake(t, "")}
	if _, err := u.Verificar(context.Background()); err != nil {
		t.Fatalf("pnpm en ~/.nvm/versions/node/*/bin debe resolverse vía fallback: %v", err)
	}
}

func TestBuildVePathAumentado(t *testing.T) {
	// Regresión del bug real: validarRepo podía encontrar go vía fallback pero Build()
	// heredaba el PATH SIN aumentar (cmd.Env nil = os.Environ() tal cual) — bundle.sh
	// fallaba igual adentro. bundle.sh acá resuelve "go" vía `command -v`, que solo
	// existe en el fallback ~/.local/go/bin, nunca en el PATH base del test.
	home := t.TempDir()
	t.Setenv("HOME", home)
	escribe(t, filepath.Join(home, ".local", "go", "bin", "go"), "#!/bin/sh\nexit 0\n", 0o755)
	// base PATH fiel al bug real: bash SÍ resuelve (systemd/session lo trae de
	// /usr/bin — por eso el error reportado era solo sobre "go"), pero go NO — solo
	// el fallback ~/.local/go/bin lo resuelve. bashDir real (no un stub): así el
	// bash que corre bundle.sh es el de verdad, no otra capa de fixture.
	bashReal, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("host sin bash resoluble — no puedo probar el PATH aumentado")
	}
	t.Setenv("PATH", filepath.Dir(bashReal))

	repo := repoFake(t, "#!/usr/bin/env bash\ncommand -v go >/dev/null || { echo 'go no resuelto' >&2; exit 1; }\nmkdir -p bin\necho compilado > bin/arnesia\n")
	u := &Updater{repo: repo}
	detalle, err := u.Build(context.Background())
	if err != nil {
		t.Fatalf("Build debe heredar el PATH aumentado (go vía fallback): %v — detalle: %s", err, detalle)
	}
}

func TestVerificarBinarioNoEjecutable(t *testing.T) {
	// El bit de ejecución como veredicto es semántica unix (en Windows el veredicto
	// es la extensión .exe — ver esEjecutable de os_windows.go).
	repo := t.TempDir()
	escribe(t, filepath.Join(repo, "bin", "arnesia"), "x", 0o600)
	u := &Updater{repo: repo}
	if _, _, err := u.VerificarBinario(context.Background()); err == nil {
		t.Fatal("un archivo sin bit de ejecución debe rechazarse")
	}
}

func TestInstalarAtomico(t *testing.T) {
	repo := t.TempDir()
	escribe(t, filepath.Join(repo, "bin", "arnesia"), "NUEVO", 0o755)
	exe := exeFake(t, "VIEJO")
	u := &Updater{repo: repo, exePath: exe}

	if _, err := u.Instalar(context.Background()); err != nil {
		t.Fatalf("instalar debe pasar: %v", err)
	}
	if got := leeExe(t, exe); got != "NUEVO" {
		t.Fatalf("el ejecutable debe ser el binario nuevo, tengo %q", got)
	}
	st, err := os.Stat(exe)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm() != 0o755 {
		t.Fatalf("modo = %v, quiero 0755", st.Mode().Perm())
	}
	// Sin residuos: el tmp del write-tmp→rename no queda tirado.
	restos, err := filepath.Glob(filepath.Join(filepath.Dir(exe), ".arnesia-nuevo-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(restos) != 0 {
		t.Fatalf("quedaron tmp sin limpiar: %v", restos)
	}
}

func TestInstalarNoEscribibleDejaIntacto(t *testing.T) {
	repo := t.TempDir()
	escribe(t, filepath.Join(repo, "bin", "arnesia"), "NUEVO", 0o755)
	exe := exeFake(t, "VIEJO")
	soloLectura(t, filepath.Dir(exe))
	u := &Updater{repo: repo, exePath: exe}

	if _, err := u.Instalar(context.Background()); err == nil {
		t.Fatal("un dir no escribible debe fallar honesto")
	}
	if got := leeExe(t, exe); got != "VIEJO" {
		t.Fatalf("el binario instalado debe quedar INTACTO tras el fallo, tengo %q", got)
	}
}

// TestVerificarUnixMenciona la toolchain unix en el detalle (bash es parte del terreno).
func TestVerificarDetalleTools(t *testing.T) {
	stubs := t.TempDir()
	for _, tool := range toolchainRequerida() {
		stubTool(t, stubs, tool)
	}
	t.Setenv("PATH", stubs+string(os.PathListSeparator)+os.Getenv("PATH"))
	u := &Updater{repo: repoFake(t, "")}
	detalle, err := u.Verificar(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(detalle, "bash") {
		t.Fatalf("el detalle unix debe nombrar bash, tengo %q", detalle)
	}
}
