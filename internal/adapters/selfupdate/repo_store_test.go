package selfupdate

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRepoStore (bugfix fix-repo-self-update, RF-108): persistencia atómica del path
// configurado vía UI — sobrevive un "reinicio" (una segunda apertura del mismo path).
func TestRepoStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "self-update.json")

	t.Run("leer sin archivo devuelve vacio", func(t *testing.T) {
		s, err := NewRepoStore(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s.Leer()
		if err != nil {
			t.Fatalf("archivo ausente NO es error: %v", err)
		}
		if got != "" {
			t.Fatalf("Leer() = %q, quiero vacío", got)
		}
	})

	t.Run("guardar y releer, incluso tras reabrir el store", func(t *testing.T) {
		s, err := NewRepoStore(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = s.Guardar("/repo/uno"); err != nil {
			t.Fatal(err)
		}
		// "reinicio" = una instancia NUEVA sobre el mismo archivo.
		s2, err := NewRepoStore(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s2.Leer()
		if err != nil {
			t.Fatal(err)
		}
		if got != "/repo/uno" {
			t.Fatalf("Leer() = %q, quiero /repo/uno", got)
		}
	})

	t.Run("guardar reemplaza el valor previo", func(t *testing.T) {
		s, err := NewRepoStore(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = s.Guardar("/repo/dos"); err != nil {
			t.Fatal(err)
		}
		got, err := s.Leer()
		if err != nil {
			t.Fatal(err)
		}
		if got != "/repo/dos" {
			t.Fatalf("Leer() = %q, quiero /repo/dos (el guardado más reciente)", got)
		}
	})

	t.Run("sin residuos tmp tras guardar", func(t *testing.T) {
		restos, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".self-update-*.json"))
		if err != nil {
			t.Fatal(err)
		}
		if len(restos) != 0 {
			t.Fatalf("quedaron tmp sin limpiar: %v", restos)
		}
	})

	t.Run("archivo corrupto degrada a vacio, no aborta el boot", func(t *testing.T) {
		corrupto := filepath.Join(t.TempDir(), "self-update.json")
		if err := os.WriteFile(corrupto, []byte("{no es json"), 0o600); err != nil {
			t.Fatal(err)
		}
		s, err := NewRepoStore(corrupto)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s.Leer()
		if err != nil {
			t.Fatalf("un archivo corrupto NO debe abortar el boot: %v", err)
		}
		if got != "" {
			t.Fatalf("Leer() sobre corrupto = %q, quiero vacío (honesto, no inventado)", got)
		}
	})
}
