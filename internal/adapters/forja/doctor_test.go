package forja_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/forja"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestChequearTresEstados cubre el doctor v0 completo (contrato §4, A-D4): ausente en un
// dir virgen → sana tras sembrar → incompleta al borrar una pieza (con el faltante
// LISTADO), y el lock ilegible como malformado visible.
func TestChequearTresEstados(t *testing.T) {
	dir := t.TempDir()
	a := forja.New(semillaReal(t))

	// 1. ausente: no existe .arnesia/.
	salud, err := a.Chequear(dir)
	if err != nil {
		t.Fatalf("Chequear (virgen): %v", err)
	}
	if salud.Estado != domain.SemillaAusente {
		t.Fatalf("estado = %q, quiero ausente", salud.Estado)
	}

	// 2. sana: tras la siembra completa.
	if _, serr := a.Sembrar(dir); serr != nil {
		t.Fatalf("Sembrar: %v", serr)
	}
	salud, err = a.Chequear(dir)
	if err != nil {
		t.Fatalf("Chequear (sembrado): %v", err)
	}
	if salud.Estado != domain.SemillaSana || len(salud.Faltantes) != 0 {
		t.Fatalf("salud = %+v, quiero sana sin faltantes", salud)
	}

	// 3. incompleta: se borra una pieza y el veredicto la LISTA.
	if rerr := os.Remove(filepath.Join(dir, ".arnesia", "proceso", "historia", "01-spec.md")); rerr != nil {
		t.Fatal(rerr)
	}
	salud, err = a.Chequear(dir)
	if err != nil {
		t.Fatalf("Chequear (roto): %v", err)
	}
	if salud.Estado != domain.SemillaIncompleta {
		t.Errorf("estado = %q, quiero incompleta", salud.Estado)
	}
	if !slices.Contains(salud.Faltantes, ".arnesia/proceso/historia/01-spec.md") {
		t.Errorf("Faltantes = %v, quiero listado .arnesia/proceso/historia/01-spec.md", salud.Faltantes)
	}
}

// TestChequearLockMalformado: un lock presente pero ilegible NO pasa por sana — es
// malformado VISIBLE con motivo (diseño A-T2: lock ilegible = malformado).
func TestChequearLockMalformado(t *testing.T) {
	dir := t.TempDir()
	a := forja.New(semillaReal(t))
	if _, err := a.Sembrar(dir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".arnesia", "semilla.lock.json"), []byte("{esto no es json"), 0o600); err != nil {
		t.Fatal(err)
	}
	salud, err := a.Chequear(dir)
	if err != nil {
		t.Fatalf("Chequear: %v", err)
	}
	if salud.Estado != domain.SemillaIncompleta {
		t.Errorf("estado = %q, quiero incompleta (lock malformado)", salud.Estado)
	}
	if !slices.Contains(salud.Faltantes, ".arnesia/semilla.lock.json (malformado)") {
		t.Errorf("Faltantes = %v, quiero el lock marcado malformado", salud.Faltantes)
	}
	if salud.Detalle == "" {
		t.Error("Detalle vacío — el motivo del malformado debe ser visible")
	}
}
