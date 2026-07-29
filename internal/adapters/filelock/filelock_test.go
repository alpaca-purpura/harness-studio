package filelock_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/filelock"
)

// TestGuardSerializesWithinProcess prueba la exclusión mutua dentro de UN proceso (varias
// goroutines): cada una hace read-increment-write de un contador en disco sin Guard sería
// una carrera clásica de lost-update; con Guard, el resultado final tiene que ser exacto.
func TestGuardSerializesWithinProcess(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contador.txt")
	if err := os.WriteFile(path, []byte("0"), 0o600); err != nil {
		t.Fatal(err)
	}

	const goroutines, incrementosPorGoroutine = 8, 50
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range incrementosPorGoroutine {
				if err := filelock.Guard(path, func() error { return incrementar(path) }); err != nil {
					t.Error(err)
				}
			}
		}()
	}
	wg.Wait()

	got := leerContador(t, path)
	want := goroutines * incrementosPorGoroutine
	if got != want {
		t.Fatalf("contador = %d, want %d (una escritura se perdió: Guard no serializó)", got, want)
	}
}

// TestGuardSerializesAcrossProcesses es la prueba REAL de la Fase 1: dos procesos del
// SISTEMA OPERATIVO — no dos goroutines del mismo proceso — compitiendo por el mismo
// contador vía Guard. Esto es exactamente la forma del bug confirmado en producción
// (`arnesia serve` vs `arnesia portafolio agregar`, dos binarios separados). Se ejecuta
// reinvocándose a sí mismo (`os.Args[0]`) con una variable de entorno que dispara el modo
// "helper" — patrón estándar de Go para tests que necesitan un proceso de verdad.
func TestGuardSerializesAcrossProcesses(t *testing.T) {
	if os.Getenv("FILELOCK_HELPER") != "" {
		t.Skip("soy el proceso hijo, no un test — ver TestHelperProcess")
	}

	path := filepath.Join(t.TempDir(), "contador.txt")
	if err := os.WriteFile(path, []byte("0"), 0o600); err != nil {
		t.Fatal(err)
	}

	const procesos, incrementosPorProceso = 4, 40
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	errs := make(chan error, procesos)
	for range procesos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestHelperProcess", "-test.v") //nolint:gosec // G204: os.Args[0] es el propio binario de test, patrón estándar de re-exec.
			cmd.Env = append(os.Environ(),
				"FILELOCK_HELPER=1",
				"FILELOCK_PATH="+path,
				"FILELOCK_N="+strconv.Itoa(incrementosPorProceso),
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				errs <- fmt.Errorf("proceso hijo: %w\n%s", err, out)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}

	got := leerContador(t, path)
	want := procesos * incrementosPorProceso
	if got != want {
		t.Fatalf("contador = %d, want %d (una escritura se perdió ENTRE PROCESOS: el lock no es cross-proceso)", got, want)
	}
}

// TestHelperProcess no es un test: es el cuerpo del proceso hijo que
// TestGuardSerializesAcrossProcesses reinvoca. Sale inmediatamente si no lo llamó ese test.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("FILELOCK_HELPER") == "" {
		return
	}
	path := os.Getenv("FILELOCK_PATH")
	n, err := strconv.Atoi(os.Getenv("FILELOCK_N"))
	if err != nil {
		t.Fatalf("FILELOCK_N inválido: %v", err)
	}
	for range n {
		if err := filelock.Guard(path, func() error { return incrementar(path) }); err != nil {
			t.Fatal(err)
		}
	}
}

// incrementar hace el read-modify-write MÁS ingenuo posible a propósito: sin esto adentro de
// un Guard, es una carrera de lost-update de manual.
func incrementar(path string) error {
	b, err := os.ReadFile(path) //nolint:gosec // ruta de test, dentro de t.TempDir().
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}
	// Ensancha la ventana de la carrera a propósito: sin Guard, esto la haría casi segura.
	time.Sleep(time.Millisecond)
	return os.WriteFile(path, []byte(strconv.Itoa(n+1)), 0o600) //nolint:gosec // G304: ruta de test, dentro de t.TempDir().
}

func leerContador(t *testing.T, path string) int {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // ruta de test.
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		t.Fatal(err)
	}
	return n
}
