package watch_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/watch"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// regFija is a minimal ports.ArnesRegistry fake: a fixed set of entries, Register/
// Resolve unused (Watch only calls List).
type regFija []ports.ArnesPath

func (r regFija) Resolve(string) (string, bool, error) {
	return "", false, errors.New("regFija: Resolve no implementado")
}

func (r regFija) Register(string, string) error {
	return errors.New("regFija: Register no implementado")
}

func (r regFija) List() []ports.ArnesPath { return r }

// awaitEvent reads from events until it sees one whose Path matches want, or times
// out — fsnotify can emit more than one raw event per FS operation (e.g. a Write
// alongside a Chmod), so a strict single-recv assertion would be flaky.
func awaitEvent(t *testing.T, events <-chan ports.WatchEvent, want string) ports.WatchEvent {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				t.Fatalf("channel de eventos cerrado esperando %q", want)
			}
			if ev.Path == want {
				return ev
			}
		case <-deadline:
			t.Fatalf("timeout esperando WatchEvent para %q", want)
		}
	}
}

// TestWatchEmitsEventUnderRegisteredPath cubre RF-210 escenario 1: un archivo creado
// bajo el árbol de un arnés registrado produce un WatchEvent con ese path.
func TestWatchEmitsEventUnderRegisteredPath(t *testing.T) {
	dir := t.TempDir()
	reg := regFija{{Arnes: "a", Path: dir}}
	w := watch.New(reg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	target := filepath.Join(dir, "nuevo.txt")
	if werr := os.WriteFile(target, []byte("hola"), 0o600); werr != nil {
		t.Fatalf("WriteFile: %v", werr)
	}

	ev := awaitEvent(t, events, target)
	if ev.Op != ports.WatchCreate && ev.Op != ports.WatchWrite {
		t.Errorf("Op = %q, want create o write", ev.Op)
	}
}

// TestWatchObservesNestedSubdirCreatedAfterStart cubre el manejo de subdirectorios
// nuevos: fsnotify no es recursivo, así que un Mkdir posterior al arranque del watcher
// debe sumarse al watch-set para que un archivo creado DENTRO también se vea.
func TestWatchObservesNestedSubdirCreatedAfterStart(t *testing.T) {
	dir := t.TempDir()
	reg := regFija{{Arnes: "a", Path: dir}}
	w := watch.New(reg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	// Un solo MkdirAll crearía ambos niveles casi atómicamente — más rápido de lo que
	// el consumidor puede reaccionar al Create de "skills" y sumarlo al watch, así que
	// el nivel más profundo nace en un directorio todavía no observado (carrera real
	// de cualquier watch-recursivo-perezoso). Crear un nivel por vez, esperando el
	// evento de cada uno, prueba el mecanismo sin depender de ganar esa carrera.
	level1 := filepath.Join(dir, "skills")
	if merr := os.Mkdir(level1, 0o750); merr != nil {
		t.Fatalf("Mkdir: %v", merr)
	}
	awaitEvent(t, events, level1)

	sub := filepath.Join(level1, "nuevo")
	if merr := os.Mkdir(sub, 0o750); merr != nil {
		t.Fatalf("Mkdir: %v", merr)
	}
	awaitEvent(t, events, sub)

	target := filepath.Join(sub, "SKILL.md")
	if werr := os.WriteFile(target, []byte("---\nname: x\n---\n"), 0o600); werr != nil {
		t.Fatalf("WriteFile: %v", werr)
	}
	awaitEvent(t, events, target)
}

// TestWatchIgnoresUnregisteredTree cubre el caso base: un directorio que NADIE
// registró nunca observa eventos (el watcher solo suma lo que reg.List() devuelve).
func TestWatchIgnoresUnregisteredTree(t *testing.T) {
	watched := t.TempDir()
	unwatched := t.TempDir()
	reg := regFija{{Arnes: "a", Path: watched}}
	w := watch.New(reg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}

	if werr := os.WriteFile(filepath.Join(unwatched, "ignorado.txt"), []byte("x"), 0o600); werr != nil {
		t.Fatalf("WriteFile: %v", werr)
	}
	// Confirma que el watcher sigue vivo/atento al árbol correcto tras el ruido.
	target := filepath.Join(watched, "si-importa.txt")
	if werr := os.WriteFile(target, []byte("x"), 0o600); werr != nil {
		t.Fatalf("WriteFile: %v", werr)
	}
	ev := awaitEvent(t, events, target)
	if ev.Path == filepath.Join(unwatched, "ignorado.txt") {
		t.Fatal("el watcher emitió un evento del árbol NO registrado")
	}
}

// TestWatchClosesWhenContextDone cubre el contrato del puerto: el canal se cierra
// cuando el ctx se cancela, sin importar si reg es nil.
func TestWatchClosesWhenContextDone(t *testing.T) {
	w := watch.New(nil)
	ctx, cancel := context.WithCancel(context.Background())
	events, err := w.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	cancel()
	select {
	case _, ok := <-events:
		if ok {
			t.Fatal("recibí un evento inesperado tras cancelar ctx")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("el canal no se cerró tras cancelar ctx")
	}
}
