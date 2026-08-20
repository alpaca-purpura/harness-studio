package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestRunInitSiembraChequeaYExitCode cubre el contrato del comando (spec RF-A.3 / AC a-b):
// siembra + doctor en un solo tiro, re-run idempotente, y `--check` con exit≠0 al romper
// una pieza («no avanzamos si no está sana», A-D4).
func TestRunInitSiembraChequeaYExitCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows: os.UserHomeDir lee USERPROFILE, no HOME
	proyecto := filepath.Join(home, "proyecto")
	if err := os.MkdirAll(proyecto, 0o750); err != nil {
		t.Fatal(err)
	}

	// Siembra + salud sana ⇒ exit 0.
	if err := runInit([]string{"--json", proyecto}); err != nil {
		t.Fatalf("init (virgen): %v", err)
	}
	if _, err := os.Stat(filepath.Join(proyecto, ".arnesia", "semilla.lock.json")); err != nil {
		t.Fatalf("la siembra no dejó el lock: %v", err)
	}

	// Re-run idempotente ⇒ sigue sana, exit 0.
	if err := runInit([]string{"--json", proyecto}); err != nil {
		t.Fatalf("init (re-run): %v", err)
	}

	// --check sobre instalación sana ⇒ exit 0.
	if err := runInit([]string{"--check", "--json", proyecto}); err != nil {
		t.Fatalf("init --check (sana): %v", err)
	}

	// Se rompe una pieza ⇒ --check exit≠0 con el error sentinela.
	if err := os.Remove(filepath.Join(proyecto, ".arnesia", "wip", "INDEX.md")); err != nil {
		t.Fatal(err)
	}
	err := runInit([]string{"--check", "--json", proyecto})
	if !errors.Is(err, errSemillaInsana) {
		t.Fatalf("init --check (rota) = %v, quiero errSemillaInsana (exit≠0)", err)
	}
}
