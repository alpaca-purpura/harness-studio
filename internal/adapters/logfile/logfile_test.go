package logfile

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestEscribeYPersiste(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "arnesia.log")
	w, err := Abrir(path, 0)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	defer func() { _ = w.Close() }()

	if _, err := w.Write([]byte("una línea\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer: %v", err)
	}
	if string(b) != "una línea\n" {
		t.Errorf("contenido = %q", b)
	}
}

// El dir padre puede no existir en la primera corrida de una instalación nueva: crearlo es
// parte del contrato, no del wiring del llamador.
func TestAbrirCreaElDirPadre(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "arnesia.log")
	w, err := Abrir(path, 0)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	_ = w.Close()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("no se creó el archivo: %v", err)
	}
}

func TestRotaYConservaUnaGeneracion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arnesia.log")
	w, err := Abrir(path, 16)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	defer func() { _ = w.Close() }()

	if _, err := w.Write([]byte("0123456789\n")); err != nil { // 11 bytes
		t.Fatalf("Write: %v", err)
	}
	if _, err := w.Write([]byte("nueva generación\n")); err != nil { // pasa el tope → rota
		t.Fatalf("Write: %v", err)
	}

	viejo, err := os.ReadFile(path + ".1")
	if err != nil {
		t.Fatalf("no quedó la generación vieja: %v", err)
	}
	if !strings.Contains(string(viejo), "0123456789") {
		t.Errorf(".1 = %q", viejo)
	}
	nuevo, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("leer actual: %v", err)
	}
	if !strings.Contains(string(nuevo), "nueva generación") {
		t.Errorf("actual = %q", nuevo)
	}
	if strings.Contains(string(nuevo), "0123456789") {
		t.Error("el archivo actual conserva lo viejo: no rotó, appendeó")
	}
}

// Reabrir NO trunca: un segundo arranque del daemon no puede borrar el log del incidente que
// se está investigando.
func TestReabrirAppendea(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arnesia.log")
	w1, err := Abrir(path, 0)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	_, _ = w1.Write([]byte("primera\n"))
	_ = w1.Close()

	w2, err := Abrir(path, 0)
	if err != nil {
		t.Fatalf("reabrir: %v", err)
	}
	_, _ = w2.Write([]byte("segunda\n"))
	_ = w2.Close()

	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "primera") || !strings.Contains(string(b), "segunda") {
		t.Errorf("contenido = %q", b)
	}
}

// slog escribe desde cualquier goroutine del daemon: si el writer no fuera seguro, el log se
// entreveraría justo cuando hay concurrencia — que es cuando más se lo necesita.
func TestWriteConcurrenteNoRompe(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arnesia.log")
	w, err := Abrir(path, 64)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	defer func() { _ = w.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = w.Write([]byte("linea concurrente\n"))
		}()
	}
	wg.Wait()
}

func TestWriteTrasCloseNoPanica(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arnesia.log")
	w, err := Abrir(path, 0)
	if err != nil {
		t.Fatalf("Abrir: %v", err)
	}
	_ = w.Close()
	if _, err := w.Write([]byte("x")); err == nil {
		t.Error("escribir tras Close debería fallar, no romper")
	}
}
