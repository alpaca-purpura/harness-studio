package forja_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/forja"
)

// arbolContrato son las 19 piezas EXACTAS del contrato semilla-arnesia.md §2 (rutas
// relativas a .arnesia/). Cambiar el árbol es cambiar el contrato — este test lo fija.
var arbolContrato = []string{
	"semilla.lock.json",
	"terreno/INDEX.md",
	"terreno/proposito/INDEX.md",
	"terreno/producto/INDEX.md",
	"terreno/organizacion/INDEX.md",
	"product/backlog.md",
	"product/roadmap.md",
	"product/stories/INDEX.md",
	"wip/INDEX.md",
	"wip/activo/.gitkeep",
	"wip/done/.gitkeep",
	"proceso/historia/00-research.md",
	"proceso/historia/01-spec.md",
	"proceso/historia/02-ux.md",
	"proceso/historia/03-arch.md",
	"proceso/historia/04-build.md",
	"proceso/historia/05-paridad.md",
	"proceso/spike/00-investigar.md",
	"proceso/spike/01-decidir.md",
}

// TestSembrarCreaElArbolDelContrato: siembra completa en un dir virgen = EXACTAMENTE el
// árbol §2, sin placeholders sin resolver, con el lock schema 0 y sus sha256 reales.
func TestSembrarCreaElArbolDelContrato(t *testing.T) {
	dir := t.TempDir()
	a := forja.New(semillaReal(t))

	inf, err := a.Sembrar(dir)
	if err != nil {
		t.Fatalf("Sembrar: %v", err)
	}
	if len(inf.YaExistian) != 0 {
		t.Errorf("dir virgen: YaExistian = %v, quiero vacío", inf.YaExistian)
	}
	if len(inf.Creados) != len(arbolContrato) {
		t.Errorf("Creados = %d, quiero %d (árbol §2)", len(inf.Creados), len(arbolContrato))
	}
	for _, ruta := range arbolContrato {
		abs := filepath.Join(dir, ".arnesia", filepath.FromSlash(ruta))
		b, rerr := os.ReadFile(abs) //nolint:gosec // test: rutas del contrato bajo t.TempDir.
		if rerr != nil {
			t.Errorf("falta %s: %v", ruta, rerr)
			continue
		}
		if bytes.Contains(b, []byte("{{")) {
			t.Errorf("%s: placeholder sin resolver (contiene «{{»)", ruta)
		}
	}

	// El render del territorio deriva del arnes.yaml: label + dimensiones del territorio.
	prop, err := os.ReadFile(filepath.Join(dir, ".arnesia", "terreno", "proposito", "INDEX.md")) //nolint:gosec // test: ruta fija bajo t.TempDir.
	if err != nil {
		t.Fatal(err)
	}
	for _, quiero := range []string{"territorio: proposito", "Propósito", "| encargo | Why |", "| stakeholders | Who |", "| necesidad | What |"} {
		if !bytes.Contains(prop, []byte(quiero)) {
			t.Errorf("terreno/proposito/INDEX.md sin %q", quiero)
		}
	}

	// Lock schema 0 (contrato §3): 5 campos + hashes de lo que ESTA siembra escribió
	// (todo el árbol salvo el lock mismo), verificados contra el disco.
	var lock struct {
		Schema         int               `json:"schema"`
		ArnesID        string            `json:"arnes_id"`
		VersionPineada int               `json:"version_pineada"`
		Fecha          string            `json:"fecha"`
		Archivos       map[string]string `json:"archivos"`
	}
	lb, err := os.ReadFile(filepath.Join(dir, ".arnesia", "semilla.lock.json")) //nolint:gosec // test: ruta fija bajo t.TempDir.
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(lb, &lock); err != nil {
		t.Fatalf("lock ilegible: %v", err)
	}
	if lock.Schema != 0 || lock.ArnesID != "dev" || lock.VersionPineada != 0 {
		t.Errorf("lock = schema %d arnes_id %q version_pineada %d, quiero 0/dev/0", lock.Schema, lock.ArnesID, lock.VersionPineada)
	}
	if len(lock.Fecha) != len("2026-07-30") {
		t.Errorf("lock.Fecha = %q, quiero AAAA-MM-DD", lock.Fecha)
	}
	if len(lock.Archivos) != len(arbolContrato)-1 {
		t.Errorf("lock.Archivos = %d entradas, quiero %d (todo menos el lock)", len(lock.Archivos), len(arbolContrato)-1)
	}
	for ruta, quiero := range lock.Archivos {
		b, rerr := os.ReadFile(filepath.Join(dir, ".arnesia", filepath.FromSlash(ruta))) //nolint:gosec // test: rutas del lock bajo t.TempDir.
		if rerr != nil {
			t.Errorf("lock apunta a %s que no existe: %v", ruta, rerr)
			continue
		}
		sum := sha256.Sum256(b)
		if got := hex.EncodeToString(sum[:]); got != quiero {
			t.Errorf("lock hash de %s no coincide con el disco", ruta)
		}
	}
}

// TestSembrarIdempotenteJamasPisa: re-correr sobre una instalación sana ⇒ TODO ya-existia,
// cero re-escritura — incluso un archivo EDITADO por el usuario se respeta byte a byte
// (contrato §4.1).
func TestSembrarIdempotenteJamasPisa(t *testing.T) {
	dir := t.TempDir()
	a := forja.New(semillaReal(t))
	if _, err := a.Sembrar(dir); err != nil {
		t.Fatalf("primera siembra: %v", err)
	}

	editado := filepath.Join(dir, ".arnesia", "product", "backlog.md")
	delUsuario := []byte("# mi backlog, editado a mano\n")
	if err := os.WriteFile(editado, delUsuario, 0o600); err != nil {
		t.Fatal(err)
	}

	inf, err := a.Sembrar(dir)
	if err != nil {
		t.Fatalf("re-siembra: %v", err)
	}
	if len(inf.Creados) != 0 {
		t.Errorf("re-siembra: Creados = %v, quiero cero escrituras", inf.Creados)
	}
	if len(inf.YaExistian) != len(arbolContrato) {
		t.Errorf("re-siembra: YaExistian = %d, quiero %d (todo el árbol)", len(inf.YaExistian), len(arbolContrato))
	}
	tras, err := os.ReadFile(editado) //nolint:gosec // test: ruta fija bajo t.TempDir.
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tras, delUsuario) {
		t.Error("la re-siembra PISÓ un archivo del usuario — viola §4.1 (jamás se pisa)")
	}
}

// TestSembrarParcialSoloCreaLoAusente: con la mitad del árbol ya presente, la siembra
// crea SOLO lo ausente y el lock lleva hashes SOLO de lo que esta siembra escribió.
func TestSembrarParcialSoloCreaLoAusente(t *testing.T) {
	dir := t.TempDir()
	pre := filepath.Join(dir, ".arnesia", "product", "backlog.md")
	if err := os.MkdirAll(filepath.Dir(pre), 0o750); err != nil {
		t.Fatal(err)
	}
	previo := []byte("# backlog previo del usuario\n")
	if err := os.WriteFile(pre, previo, 0o600); err != nil {
		t.Fatal(err)
	}

	a := forja.New(semillaReal(t))
	inf, err := a.Sembrar(dir)
	if err != nil {
		t.Fatalf("Sembrar: %v", err)
	}
	if len(inf.YaExistian) != 1 || inf.YaExistian[0] != ".arnesia/product/backlog.md" {
		t.Errorf("YaExistian = %v, quiero solo .arnesia/product/backlog.md", inf.YaExistian)
	}
	tras, _ := os.ReadFile(pre) //nolint:gosec // test: ruta fija bajo t.TempDir.
	if !bytes.Equal(tras, previo) {
		t.Error("el archivo preexistente se pisó")
	}

	var lock struct {
		Archivos map[string]string `json:"archivos"`
	}
	lb, _ := os.ReadFile(filepath.Join(dir, ".arnesia", "semilla.lock.json")) //nolint:gosec // test: ruta fija bajo t.TempDir.
	if err := json.Unmarshal(lb, &lock); err != nil {
		t.Fatal(err)
	}
	if _, tiene := lock.Archivos["product/backlog.md"]; tiene {
		t.Error("el lock lleva hash de un archivo que ESTA siembra no escribió")
	}
	if len(lock.Archivos) != len(arbolContrato)-2 { // sin el lock ni el preexistente
		t.Errorf("lock.Archivos = %d, quiero %d", len(lock.Archivos), len(arbolContrato)-2)
	}
}
