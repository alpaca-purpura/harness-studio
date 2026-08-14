package selfupdate

// updater_test.go — casos VÁLIDOS EN AMBOS OS (split 2026-08-13-compilacion-windows):
// los fixtures por-OS (stubTool, aislarHome, stubBundle*) viven en stubs_{unix,windows}_test.go
// y los casos cuya semántica solo existe en un OS (bit de ejecución, chmod 0555,
// instalar con rename/degradación) en updater_{unix,windows}_test.go.

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
// en silencio. El modo lo decide el caso (los stubs de bundle DEBEN ser ejecutables).
func escribe(t *testing.T, path string, contenido string, modo os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contenido), modo); err != nil {
		t.Fatal(err)
	}
}

// repoFake arma un árbol mínimo que pasa Verificar: go.mod del módulo esperado + el
// script de bundle del OS (el stub que se le pase, o el noop OK).
func repoFake(t *testing.T, bundleStub string) string {
	t.Helper()
	dir := t.TempDir()
	escribe(t, filepath.Join(dir, "go.mod"), "module github.com/alpacapurpura/arnesia\n\ngo 1.23.0\n", 0o600)
	if bundleStub == "" {
		bundleStub = stubBundleNoop()
	}
	escribe(t, filepath.Join(dir, scriptDeBundle()), bundleStub, 0o755)
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
	t.Run("sin script de bundle", func(t *testing.T) {
		dir := t.TempDir()
		escribe(t, filepath.Join(dir, "go.mod"), "module github.com/alpacapurpura/arnesia\n", 0o600)
		u := &Updater{repo: dir}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatalf("sin %s debe fallar", scriptDeBundle())
		}
	})
	t.Run("repo completo", func(t *testing.T) {
		// Toolchain hermética: el runner de CI NO trae pnpm — stubs que lookPathEn
		// resuelve (por bit en unix, por PATHEXT en windows) en un dir prepended a PATH
		// hacen al test independiente del host.
		stubs := t.TempDir()
		for _, tool := range toolchainRequerida() {
			stubTool(t, stubs, tool)
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
		// PATH sin toolchain → Verificar corta honesto (feature de operador-dev).
		// Home/env aislados: los fallbacks reales de candidatosPATHOS del host no
		// pueden decidir el caso negativo.
		t.Setenv("PATH", t.TempDir())
		aislarHome(t)
		u := &Updater{repo: repoFake(t, "")}
		if _, err := u.Verificar(ctx); err == nil {
			t.Fatal("sin toolchain en PATH debe fallar honesto")
		}
	})
}

// TestConfigurarRepo (bugfix fix-repo-self-update, RF-109): mismas reglas que
// Verificar, mas corriendo sobre un CANDIDATO — nunca sobre u.repo hasta que pasa.
func TestConfigurarRepo(t *testing.T) {
	ctx := context.Background()
	stubs := t.TempDir()
	for _, tool := range toolchainRequerida() {
		stubTool(t, stubs, tool)
	}
	t.Setenv("PATH", stubs+string(os.PathListSeparator)+os.Getenv("PATH"))

	t.Run("candidato valido activa el repo", func(t *testing.T) {
		u := &Updater{}
		candidato := repoFake(t, "")
		detalle, err := u.ConfigurarRepo(ctx, candidato)
		if err != nil {
			t.Fatalf("candidato válido debe pasar: %v", err)
		}
		if !strings.Contains(detalle, "módulo esperado") {
			t.Fatalf("detalle sin sustancia: %q", detalle)
		}
		if got := u.repoAtual(); got != candidato {
			t.Fatalf("repo activo = %q, quiero %q", got, candidato)
		}
	})

	t.Run("candidato invalido NO toca el repo activo", func(t *testing.T) {
		activo := repoFake(t, "")
		u := &Updater{repo: activo}
		invalido := t.TempDir() // sin go.mod: falla validarRepo.
		if _, err := u.ConfigurarRepo(ctx, invalido); err == nil {
			t.Fatal("un dir sin go.mod debe rechazarse")
		}
		if got := u.repoAtual(); got != activo {
			t.Fatalf("un candidato inválido NO debe mover el repo activo: tengo %q, quiero %q", got, activo)
		}
	})

	t.Run("modulo ajeno rechazado", func(t *testing.T) {
		u := &Updater{}
		dir := t.TempDir()
		escribe(t, filepath.Join(dir, "go.mod"), "module github.com/otra/cosa\n", 0o600)
		if _, err := u.ConfigurarRepo(ctx, dir); err == nil {
			t.Fatal("módulo ajeno debe rechazarse igual que Verificar (decisión #4c)")
		}
		if got := u.repoAtual(); got != "" {
			t.Fatalf("repo activo debe seguir vacío, tengo %q", got)
		}
	})
}

func TestBuildStubOK(t *testing.T) {
	repo := repoFake(t, stubBundleOK())
	u := &Updater{repo: repo}
	if _, err := u.Build(context.Background()); err != nil {
		t.Fatalf("stub OK debe pasar: %v", err)
	}
	if _, err := os.Stat(filepath.Join(repo, "bin", nombreBinNuevo())); err != nil {
		t.Fatalf("el stub debió dejar bin/%s: %v", nombreBinNuevo(), err)
	}
}

func TestBuildFallaConStderr(t *testing.T) {
	repo := repoFake(t, stubBundleFalla())
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
	repo := repoFake(t, stubBundleCuelga())
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
// deja como repo/bin/<binario> — un binario Go REAL cuyo buildinfo se puede leer sin
// ejecutarlo (RF-104 ③).
func binarioGoTest(ctx context.Context, t *testing.T, repo string) {
	t.Helper()
	src := t.TempDir()
	escribe(t, filepath.Join(src, "go.mod"), "module tiny\n\ngo 1.23.0\n", 0o600)
	escribe(t, filepath.Join(src, "main.go"), "package main\n\nfunc main() {}\n", 0o600)
	out := filepath.Join(repo, "bin", nombreBinNuevo())
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
			t.Fatal("sin el binario nuevo debe fallar honesto")
		}
	})
	t.Run("sin buildinfo", func(t *testing.T) {
		repo := t.TempDir()
		escribe(t, filepath.Join(repo, "bin", nombreBinNuevo()), "#!/bin/sh\n", 0o755)
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

func TestInstalarSinBinarioNuevo(t *testing.T) {
	u := &Updater{repo: t.TempDir(), exePath: exeFake(t, "VIEJO")}
	if _, err := u.Instalar(context.Background()); err == nil {
		t.Fatal("sin el binario nuevo el paso instalar debe fallar")
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
