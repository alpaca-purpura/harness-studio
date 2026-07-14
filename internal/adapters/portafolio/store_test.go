package portafolio_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

func entrada(home, id string) domain.EntradaPortafolio {
	return domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{Home: home, ID: id}}
}

func TestStoreDegradaHonesto(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	escribir(t, path, `{"version":1,"entradas":[
		{"identidad":{"id":"sana-1"}},
		{"identidad":{"id":"sana-2"}},
		{"identidad":"esto debería ser un objeto, no un string"}
	]}`)

	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore no debe fallar por contenido corrupto: %v", err)
	}
	sanas, corruptas := s.Listar()
	if len(sanas) != 2 {
		t.Errorf("sanas = %d, quiero 2", len(sanas))
	}
	if len(corruptas) != 1 {
		t.Fatalf("corruptas = %d, quiero 1 (visible)", len(corruptas))
	}
	if corruptas[0].Motivo == "" {
		t.Error("quiero un motivo visible en la corrupta")
	}

	// Un Upsert posterior (dispara save) NO debe perder la corrupta cruda (decisión: conservarla).
	if uerr := s.Upsert(entrada("", "sana-3")); uerr != nil {
		t.Fatal(uerr)
	}
	s2, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	sanas2, corruptas2 := s2.Listar()
	if len(sanas2) != 3 {
		t.Errorf("tras reabrir: sanas = %d, quiero 3", len(sanas2))
	}
	if len(corruptas2) != 1 {
		t.Errorf("tras reabrir: corruptas = %d, quiero 1 (la corrupta sobrevive al save)", len(corruptas2))
	}
}

// TestStoreArchivoTotalmenteIlegible cubre C-N-4 en su forma más dura: un archivo editado
// a mano hasta romper la sintaxis JSON entera (ni siquiera tokeniza como array). El store
// abre igual (0 sanas, 1 corrupta visible con el blob entero) y — a diferencia de una fila
// individual corrupta pero sintácticamente válida — un write posterior NO puede
// re-insertar bytes no-JSON en el envelope nuevo: se pierde tras el primer save, pero
// jamás crashea ni bloquea el arranque.
func TestStoreArchivoTotalmenteIlegible(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	escribir(t, path, `{esto ni siquiera es JSON válido, sin comillas ni estructura`)

	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore no debe fallar ni con el archivo totalmente roto: %v", err)
	}
	sanas, corruptas := s.Listar()
	if len(sanas) != 0 || len(corruptas) != 1 {
		t.Fatalf("sanas=%d corruptas=%d, quiero 0 y 1 (el blob entero visible)", len(sanas), len(corruptas))
	}

	// Un write posterior NO debe crashear (aunque la corrupta irrecuperable no sobreviva).
	if uerr := s.Upsert(entrada("", "nueva")); uerr != nil {
		t.Fatalf("Upsert tras un archivo roto no debe fallar: %v", uerr)
	}
	sanas, _ = s.Listar()
	if len(sanas) != 1 {
		t.Errorf("tras el Upsert quiero 1 sana, got %d", len(sanas))
	}
}

func TestStoreUpsertMergePorIdentidad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.IdentidadArnes{Home: "github.com/o/r", ID: "harness-x"}

	e1 := domain.EntradaPortafolio{
		Identidad:     id,
		Instalaciones: []domain.Instalacion{{InstallPath: "/proyecto-a/.claude"}},
	}
	if err := s.Upsert(e1); err != nil {
		t.Fatal(err)
	}
	e2 := domain.EntradaPortafolio{
		Identidad:     id,
		Instalaciones: []domain.Instalacion{{InstallPath: "/proyecto-b/.claude"}},
	}
	if err := s.Upsert(e2); err != nil {
		t.Fatal(err)
	}

	sanas, _ := s.Listar()
	if len(sanas) != 1 {
		t.Fatalf("misma identidad en 2 proyectos debe ser 1 entrada (C-P-10), got %d", len(sanas))
	}
	if len(sanas[0].Instalaciones) != 2 {
		t.Fatalf("quiero 2 instalaciones bajo la misma entrada, got %d", len(sanas[0].Instalaciones))
	}

	// Re-upsert del MISMO installPath no duplica (C-P-8/C-N-3).
	if err := s.Upsert(e1); err != nil {
		t.Fatal(err)
	}
	sanas, _ = s.Listar()
	if len(sanas[0].Instalaciones) != 2 {
		t.Fatalf("re-upsert del mismo installPath no debe duplicar, got %d", len(sanas[0].Instalaciones))
	}
}

func TestStoreNoFusionaProvisional(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	provisional := domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: "harness-x", Scope: "proyectos/x"}}
	resuelta := domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{Home: "github.com/o/r", ID: "harness-x"}}

	if err := s.Upsert(provisional); err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(resuelta); err != nil {
		t.Fatal(err)
	}
	sanas, _ := s.Listar()
	if len(sanas) != 2 {
		t.Fatalf("provisional y resuelta NUNCA se fusionan automáticamente (C-ID-2): quiero 2 entradas, got %d", len(sanas))
	}
}

func TestStoreDobleCanonico(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.IdentidadArnes{Home: "github.com/o/r", ID: "harness-x"}
	e1 := domain.EntradaPortafolio{Identidad: id, Canonico: &domain.Canonico{Path: "/checkouts/a"}}
	e2 := domain.EntradaPortafolio{Identidad: id, Canonico: &domain.Canonico{Path: "/checkouts/b"}}

	if err := s.Upsert(e1); err != nil {
		t.Fatal(err)
	}
	if err := s.Upsert(e2); err == nil {
		t.Fatal("dos canónicos distintos para la misma identidad debe ser error explícito (C-N-5)")
	}
}

func TestStoreSaveAtomico(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if uerr := s.Upsert(entrada("", "x")); uerr != nil {
		t.Fatal(uerr)
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if len(e.Name()) > 12 && e.Name()[:12] == ".portafolio-" {
			t.Errorf("quedó un temp file huérfano tras el save atómico: %s", e.Name())
		}
	}
	if _, serr := os.Stat(path); serr != nil {
		t.Fatalf("el archivo final debe existir: %v", serr)
	}
}
