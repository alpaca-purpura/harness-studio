package selfupdate

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// escribe es os.WriteFile con el fallo cortando el test — los fixtures jamás fallan
// en silencio. El modo lo decide el caso (los stubs de bundle.sh DEBEN ser ejecutables).
func escribe(t *testing.T, path string, contenido string, modo os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contenido), modo); err != nil {
		t.Fatal(err)
	}
}

// repoFake arma un árbol mínimo que pasa Verificar: go.mod del módulo esperado +
// scripts/bundle.sh (el stub que se le pase, o uno vacío OK).
func repoFake(t *testing.T, bundleStub string) string {
	t.Helper()
	dir := t.TempDir()
	escribe(t, filepath.Join(dir, "go.mod"), "module github.com/alpacapurpura/arnesia\n\ngo 1.23.0\n", 0o600)
	if bundleStub == "" {
		bundleStub = "#!/usr/bin/env bash\nexit 0\n"
	}
	escribe(t, filepath.Join(dir, "scripts", "bundle.sh"), bundleStub, 0o755)
	return dir
}

// exeFake instala un "binario corriente" en su propio directorio y devuelve su ruta.
func exeFake(t *testing.T, contenido string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "arnesia")
	escribe(t, path, contenido, 0o755)
	return path
}

// leeExe lee el "binario" instalado (ruta del propio test).
func leeExe(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta dentro del TempDir del propio test.
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

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

func TestVersionHonesta(t *testing.T) {
	exe := exeFake(t, "binario corriente")
	u := &Updater{repo: "", exePath: exe}
	v := u.Version()
	if v.InstaladoEn != exe {
		t.Fatalf("InstaladoEn = %q, quiero %q", v.InstaladoEn, exe)
	}
	if !v.Escribible {
		t.Fatal("un t.TempDir debe ser escribible")
	}
	if v.Repo != "" {
		t.Fatalf("Repo = %q, quiero vacío (sin --repo)", v.Repo)
	}
	if v.Huella == "" {
		t.Fatal("Huella jamás vacía — sin VCS info debe decir 'dev'")
	}
}

func TestVersionNoEscribible(t *testing.T) {
	exe := exeFake(t, "binario corriente")
	soloLectura(t, filepath.Dir(exe))
	u := &Updater{exePath: exe}
	if u.Version().Escribible {
		t.Fatal("un dir 0555 debe reportar Escribible=false")
	}
}

func TestVerificar(t *testing.T) {
	ctx := context.Background()

	t.Run("sin repo", func(t *testing.T) {
		u := &Updater{repo: ""}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("sin repo debe fallar honesto")
		}
	})
	t.Run("repo sin go.mod", func(t *testing.T) {
		u := &Updater{repo: t.TempDir()}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("un dir pelado no es el árbol esperado")
		}
	})
	t.Run("modulo ajeno", func(t *testing.T) {
		dir := t.TempDir()
		escribe(t, filepath.Join(dir, "go.mod"), "module github.com/otra/cosa\n", 0o600)
		u := &Updater{repo: dir}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("un módulo ajeno debe rechazarse (decisión #4c)")
		}
	})
	t.Run("sin bundle.sh", func(t *testing.T) {
		dir := t.TempDir()
		escribe(t, filepath.Join(dir, "go.mod"), "module github.com/alpacapurpura/arnesia\n", 0o600)
		u := &Updater{repo: dir}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("sin scripts/bundle.sh debe fallar")
		}
	})
	t.Run("repo completo", func(t *testing.T) {
		// Toolchain hermética: el runner de CI del job go NO trae pnpm — un stub
		// ejecutable en un dir prepended a PATH hace al test independiente del host.
		stubs := t.TempDir()
		for _, tool := range []string{"go", "pnpm", "bash"} {
			escribe(t, filepath.Join(stubs, tool), "#!/bin/sh\nexit 0\n", 0o755)
		}
		t.Setenv("PATH", stubs+string(os.PathListSeparator)+os.Getenv("PATH"))
		u := &Updater{repo: repoFake(t, "")}
		detalle, err := u.Verificar(ctx)
		if err != nil {
			t.Fatalf("repo fake completo debe pasar: %v", err)
		}
		if !strings.Contains(detalle, "módulo esperado") {
			t.Fatalf("detalle sin sustancia: %q", detalle)
		}
	})
	t.Run("toolchain incompleta", func(t *testing.T) {
		// PATH SIN pnpm/go/bash → Verificar corta honesto (feature de operador-dev).
		t.Setenv("PATH", t.TempDir())
		u := &Updater{repo: repoFake(t, "")}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("sin toolchain en PATH debe fallar honesto")
		}
	})
}

func TestBuildStubOK(t *testing.T) {
	repo := repoFake(t, "#!/usr/bin/env bash\nmkdir -p bin\necho compilado > bin/arnesia\n")
	u := &Updater{repo: repo}
	if _, err := u.Build(context.Background()); err != nil {
		t.Fatalf("stub OK debe pasar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "bin", "arnesia")); err != nil {
		t.Fatalf("el stub debió dejar bin/arnesia: %v", err)
	}
}

func TestBuildFallaConStderr(t *testing.T) {
	repo := repoFake(t, "#!/usr/bin/env bash\necho 'error TS2345: tipo roto' >&2\nexit 1\n")
	u := &Updater{repo: repo}
	detalle, err := u.Build(context.Background())
	if err == nil {
		t.Fatal("exit 1 debe fallar")
	}
	if !strings.Contains(detalle, "TS2345") {
		t.Fatalf("el detalle debe llevar la cola del stderr, tengo %q", detalle)
	}
}

func TestBuildCancelable(t *testing.T) {
	repo := repoFake(t, "#!/usr/bin/env bash\nsleep 10\n")
	u := &Updater{repo: repo}
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	inicio := time.Now()
	if _, err := u.Build(ctx); err == nil {
		t.Fatal("un ctx cancelado debe abortar el build")
	}
	if time.Since(inicio) > 5*time.Second {
		t.Fatal("el build no respetó la cancelación")
	}
}

// binarioGoTest compila un main trivial FUERA del repo (sin VCS → huella "dev") y lo
// deja como repo/bin/arnesia — un binario Go REAL cuyo buildinfo se puede leer sin
// ejecutarlo (RF-104 ③).
func binarioGoTest(ctx context.Context, t *testing.T, repo string) {
	t.Helper()
	src := t.TempDir()
	escribe(t, filepath.Join(src, "go.mod"), "module tiny\n\ngo 1.23.0\n", 0o600)
	escribe(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() {}\n", 0o600)
	out := filepath.Join(repo, "bin", "arnesia")
	if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, "go", "build", "-o", out, ".") //nolint:gosec // G204: compila el fixture del test con rutas de TempDir.
	cmd.Dir = src
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build del binario de prueba: %v\n%s", err, b)
	}
}

func TestVerificarBinario(t *testing.T) {
	ctx := context.Background()

	t.Run("inexistente", func(t *testing.T) {
		u := &Updater{repo: t.TempDir()}
		if _, _, err := u.VerificarBinario(ctx); err == nil {
			t.Fatal("sin bin/arnesia debe fallar honesto")
		}
	})
	t.Run("no ejecutable", func(t *testing.T) {
		repo := t.TempDir()
		escribe(t, filepath.Join(repo, "bin", "arnesia"), "x", 0o600)
		u := &Updater{repo: repo}
		if _, _, err := u.VerificarBinario(ctx); err == nil {
			t.Fatal("un archivo sin bit de ejecución debe rechazarse")
		}
	})
	t.Run("sin buildinfo", func(t *testing.T) {
		repo := t.TempDir()
		escribe(t, filepath.Join(repo, "bin", "arnesia"), "#!/bin/sh\n", 0o755)
		u := &Updater{repo: repo}
		if _, _, err := u.VerificarBinario(ctx); err == nil {
			t.Fatal("un no-binario-Go debe rechazarse (buildinfo ilegible)")
		}
	})
	t.Run("binario go real", func(t *testing.T) {
		repo := t.TempDir()
		binarioGoTest(ctx, t, repo)
		u := &Updater{repo: repo}
		huella, detalle, err := u.VerificarBinario(ctx)
		if err != nil {
			t.Fatalf("binario Go real debe pasar: %v", err)
		}
		if huella != "dev" {
			t.Fatalf("compilado fuera de git → huella 'dev', tengo %q", huella)
		}
		if !strings.Contains(detalle, "bin/arnesia") {
			t.Fatalf("detalle sin sustancia: %q", detalle)
		}
	})
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

func TestInstalarSinBinarioNuevo(t *testing.T) {
	u := &Updater{repo: t.TempDir(), exePath: exeFake(t, "VIEJO")}
	if _, err := u.Instalar(context.Background()); err == nil {
		t.Fatal("sin bin/arnesia el paso instalar debe fallar")
	}
}

func TestHuellaDeSettings(t *testing.T) {
	casos := []struct {
		nombre   string
		settings []debug.BuildSetting
		huella   string
		fecha    string
		sucio    bool
	}{
		{"sin vcs", nil, "dev", "", false},
		{"limpio", []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1c7443fabcdef0123456789"},
			{Key: "vcs.time", Value: "2026-07-07T12:00:00Z"},
			{Key: "vcs.modified", Value: "false"},
		}, "1c7443f", "2026-07-07", false},
		{"sucio", []debug.BuildSetting{
			{Key: "vcs.revision", Value: "abcdef012345"},
			{Key: "vcs.modified", Value: "true"},
		}, "abcdef0+sucio", "", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			huella, fecha, sucio := huellaDeSettings(c.settings)
			if huella != c.huella || fecha != c.fecha || sucio != c.sucio {
				t.Fatalf("(%q,%q,%v), quiero (%q,%q,%v)", huella, fecha, sucio, c.huella, c.fecha, c.sucio)
			}
		})
	}
}
