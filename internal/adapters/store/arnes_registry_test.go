package store

import (
	"path/filepath"
	"testing"
)

// TestArnesRegistryRegisterRoundTrip cubre el camino básico: Register persiste, List lo ve,
// y una segunda instancia sobre el mismo archivo lee lo mismo.
func TestArnesRegistryRegisterRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arneses.json")
	proyecto := t.TempDir()

	r1, err := NewArnesRegistry(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if uerr := r1.Register("harness-x", proyecto); uerr != nil {
		t.Fatal(uerr)
	}

	r2, err := NewArnesRegistry(path, "")
	if err != nil {
		t.Fatal(err)
	}
	list := r2.List()
	if len(list) != 1 || list[0].Arnes != "harness-x" {
		t.Fatalf("List() = %+v, quiero 1 entrada harness-x", list)
	}
}

// TestArnesRegistryRegisterNoPisaEscrituraDeOtraInstancia es la regresión de la Fase 1
// (D2/D3) para arneses.json: dos instancias sobre el mismo archivo (mismo shape que
// `runServe` vs cualquier futuro CLI standalone que registre un arnés) no deben pisarse.
func TestArnesRegistryRegisterNoPisaEscrituraDeOtraInstancia(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arneses.json")
	proyectoA := t.TempDir()
	proyectoB := t.TempDir()
	proyectoC := t.TempDir()

	r1, err := NewArnesRegistry(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if uerr := r1.Register("harness-a", proyectoA); uerr != nil {
		t.Fatal(uerr)
	}

	// r2 carga el archivo TAL COMO ESTÁ ahora (solo harness-a) — simula un segundo proceso
	// que arranca mientras el primero ya corría.
	r2, err := NewArnesRegistry(path, "")
	if err != nil {
		t.Fatal(err)
	}

	// r1 registra ALGO MÁS mientras r2 ya está vivo — r2 todavía no lo sabe.
	if uerr := r1.Register("harness-b", proyectoB); uerr != nil {
		t.Fatal(uerr)
	}

	// r2 registra lo suyo. Sin el reload-under-lock, esto pisaría el archivo con solo
	// [harness-a, harness-c], perdiendo harness-b para siempre.
	if uerr := r2.Register("harness-c", proyectoC); uerr != nil {
		t.Fatal(uerr)
	}

	r3, err := NewArnesRegistry(path, "")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, e := range r3.List() {
		got[e.Arnes] = true
	}
	for _, arnes := range []string{"harness-a", "harness-b", "harness-c"} {
		if !got[arnes] {
			t.Errorf("tras el round-trip, falta %q — se perdió una escritura entre instancias (got=%v)", arnes, got)
		}
	}
	if len(got) != 3 {
		t.Fatalf("quiero 3 arneses registrados en el archivo final, got %d: %v", len(got), got)
	}
}
