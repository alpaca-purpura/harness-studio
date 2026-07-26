package traer

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// planLocalDe arma el plan del camino A contra un checkout de fixture.
func planLocalDe(t *testing.T, checkout, ruta, id, raiz string) domain.PlanTraer {
	t.Helper()
	abs, err := filepath.Abs(checkout)
	if err != nil {
		t.Fatal(err)
	}
	mkt := domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: "github.com/alpacapurpura/prenter-marketplace",
		Clase: domain.ClasePropio, InstallLocation: abs,
	}
	fila := domain.EntradaCatalogo{
		Nombre: id,
		Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: ruta, Ruta: ruta},
	}
	plan, err := domain.PlanificarTraer(mkt, fila, true, raiz)
	if err != nil {
		t.Fatalf("PlanificarTraer: %v", err)
	}
	return plan
}

// E-76 (lado adapter) · camino A feliz contra el fixture del checkout REAL de prenter: el árbol
// completo llega al staging, sin `.git`/`.in_use`/`.orphaned_at`, y el sha efectivo es "" (en el
// camino local NO hay commit que reportar: fabricar uno sería inventar procedencia).
func TestCopiadorLocalPrenter(t *testing.T) {
	staging := filepath.Join(t.TempDir(), "staging")
	plan := planLocalDe(t, "testdata/mkt-prenter", "./plugins/harness/0.5.3", "harness", t.TempDir())

	sha, avisos, err := (&CopiadorLocal{}).Materializar(context.Background(), plan, staging)
	if err != nil {
		t.Fatalf("Materializar: %v", err)
	}
	if sha != "" {
		t.Fatalf("sha efectivo = %q, want vacío (el camino local no pinea un commit)", sha)
	}
	if len(avisos) != 0 {
		t.Fatalf("avisos = %v, want ninguno", avisos)
	}

	copiados := listar(t, staging)
	for _, esperado := range []string{"VERSION", "README.md", "hooks/hooks.json", "hooks/session-start-rules.sh", ".claude-plugin/plugin.json", ".gitignore"} {
		if !contiene(copiados, esperado) {
			t.Fatalf("%q no llegó al staging: %v", esperado, copiados)
		}
	}
	if contiene(copiados, ".orphaned_at") {
		t.Fatalf(".orphaned_at se copió (set de exclusión): %v", copiados)
	}
	if fi, serr := os.Stat(filepath.Join(staging, "hooks", "session-start-rules.sh")); serr != nil || fi.Mode().Perm()&0o111 == 0 {
		t.Fatalf("el hook perdió el bit de ejecución: %v", serr)
	}
	// Cero red: este adapter no importa nada de red — verificable por inspección de imports
	// (TestCaminoLocalNoImportaRed).
}

// El camino A no puede usar red ni `git`: se asserta por INSPECCIÓN DE IMPORTS del archivo, no por
// buena voluntad (§13.6 punto 6).
func TestCaminoLocalNoImportaRed(t *testing.T) {
	b, err := os.ReadFile("local.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, prohibido := range []string{`"net/http"`, `"os/exec"`, `"net"`} {
		if strings.Contains(string(b), prohibido) {
			t.Fatalf("local.go importa %s: el camino A tiene que ser cero-red, cero-git", prohibido)
		}
	}
}

// E-91 · `source: "./"`: el arnés ES la raíz del marketplace; se copia sin su `.git`.
func TestCopiadorLocalRaizDeMarketplace(t *testing.T) {
	staging := filepath.Join(t.TempDir(), "staging")
	abs, _ := filepath.Abs("testdata/mkt-caveman")
	mkt := domain.MarketplaceConocido{
		Nombre: "caveman", Repo: "github.com/juliusbrussee/caveman", Clase: domain.ClasePropio, InstallLocation: abs,
	}
	fila := domain.EntradaCatalogo{Nombre: "caveman", Source: domain.SourceCatalogo{Tipo: domain.SourceRutaRelativa, Crudo: "./", Ruta: "./"}}
	plan, err := domain.PlanificarTraer(mkt, fila, true, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !plan.RaizDeMarketplace {
		t.Fatal("RaizDeMarketplace = false")
	}
	if _, _, merr := (&CopiadorLocal{}).Materializar(context.Background(), plan, staging); merr != nil {
		t.Fatalf("Materializar: %v", merr)
	}
	for _, rel := range listar(t, staging) {
		if strings.HasPrefix(rel, ".git") {
			t.Fatalf("%q se copió: el canónico heredaría el remote del marketplace", rel)
		}
	}
}

// Un plan del camino equivocado, o con un origen inexistente, falla explícito (nunca copia nada).
func TestCopiadorLocalRechazaPlanAjeno(t *testing.T) {
	staging := filepath.Join(t.TempDir(), "staging")
	plan := planLocalDe(t, "testdata/mkt-prenter", "./plugins/harness/0.5.3", "harness", t.TempDir())
	plan.Camino = domain.CaminoExterno
	if _, _, err := (&CopiadorLocal{}).Materializar(context.Background(), plan, staging); err == nil {
		t.Fatal("el copiador local aceptó un plan del camino externo")
	}

	plan.Camino = domain.CaminoLocal
	plan.OrigenLocal = filepath.Join(t.TempDir(), "no-existe")
	if _, _, err := (&CopiadorLocal{}).Materializar(context.Background(), plan, staging); err == nil {
		t.Fatal("el copiador local aceptó un origen inexistente")
	}
	if _, serr := os.Stat(staging); serr == nil {
		t.Fatal("se creó el staging pese a fallar la precondición")
	}
}
