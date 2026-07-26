package marketplace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

func storeEn(t *testing.T, contenido string) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "marketplaces.json")
	if contenido != "" {
		if err := os.WriteFile(path, []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	st, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return st, path
}

// Archivo ausente ⇒ store vacío sin error: el plano nace igual poblado por el detector CC.
func TestStoreArchivoAusenteAbreVacio(t *testing.T) {
	st, _ := storeEn(t, "")
	sanas, corruptas := st.Listar()
	if len(sanas) != 0 || len(corruptas) != 0 {
		t.Fatalf("Listar = (%v, %v), want vacío", sanas, corruptas)
	}
}

// E-44 · envelope ilegible ⇒ NewStore no falla, Listar devuelve 1 corrupta con el blob entero, y
// el plano sigue poblado por el detector CC + banner de corrupta.
func TestStoreEnvelopeIlegibleDegradaHonesto(t *testing.T) {
	st, _ := storeEn(t, `{"version": 1, "marketplaces": [`)
	sanas, corruptas := st.Listar()
	if len(sanas) != 0 {
		t.Fatalf("sanas = %v, want 0", sanas)
	}
	if len(corruptas) != 1 {
		t.Fatalf("corruptas = %d, want 1", len(corruptas))
	}
	if !strings.HasPrefix(corruptas[0].Motivo, "envelope ilegible:") {
		t.Fatalf("Motivo = %q, want prefijo «envelope ilegible:»", corruptas[0].Motivo)
	}
	if len(corruptas[0].Raw) == 0 {
		t.Fatal("Raw vacío: el blob entero se conserva")
	}
}

// E-45 · una fila ilegible no contagia al resto; un Upsert posterior RE-SERIALIZA la corrupta
// cruda junto a las sanas (misma decisión y misma limitación honesta que portafolio.Store).
func TestStoreFilaCorruptaNoContagia(t *testing.T) {
	st, path := storeEn(t, `{"version":1,"marketplaces":[
		{"nombre":"a","repo":"github.com/a/a","clase":"propio"},
		{"nombre":"b","clase":123},
		{"nombre":"c","repo":"github.com/c/c","clase":"referencia"}
	]}`)
	sanas, corruptas := st.Listar()
	if len(sanas) != 2 {
		t.Fatalf("sanas = %d, want 2", len(sanas))
	}
	if len(corruptas) != 1 {
		t.Fatalf("corruptas = %d, want 1", len(corruptas))
	}

	if err := st.Upsert(domain.MarketplaceConocido{Nombre: "d", Clase: domain.ClasePropio}); err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	b, err := os.ReadFile(path) //nolint:gosec // G304: temporal del propio test.
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"clase":123`) && !strings.Contains(string(b), `"clase": 123`) {
		t.Fatalf("la corrupta no sobrevivió al ciclo load→save:\n%s", b)
	}
	st2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	sanas2, corruptas2 := st2.Listar()
	if len(sanas2) != 3 || len(corruptas2) != 1 {
		t.Fatalf("tras reabrir: sanas=%d corruptas=%d, want 3/1", len(sanas2), len(corruptas2))
	}
}

// E-46 · fila sin `nombre` ⇒ corrupta con el motivo exacto. NO se le inventa un nombre
// derivándolo del repo.
func TestStoreFilaSinNombreEsCorrupta(t *testing.T) {
	st, _ := storeEn(t, `{"version":1,"marketplaces":[{"repo":"github.com/a/b","clase":"propio"}]}`)
	sanas, corruptas := st.Listar()
	if len(sanas) != 0 {
		t.Fatalf("sanas = %v, want 0", sanas)
	}
	if len(corruptas) != 1 || corruptas[0].Motivo != "fila sin nombre: no hay clave de merge" {
		t.Fatalf("corruptas = %+v, want el motivo exacto", corruptas)
	}
}

// Upsert/Olvidar hacen el round-trip completo, y Listar siempre anota el eslabón declarado.
func TestStoreUpsertOlvidarRoundTrip(t *testing.T) {
	st, path := storeEn(t, "")
	m := domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: "github.com/alpacapurpura/prenter-marketplace",
		Clase: domain.ClasePropio, Registrado: "2026-07-25T14:02:11Z",
	}
	if err := st.Upsert(m); err != nil {
		t.Fatal(err)
	}
	st2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	sanas, _ := st2.Listar()
	if len(sanas) != 1 {
		t.Fatalf("sanas = %d, want 1", len(sanas))
	}
	if sanas[0].Clase != domain.ClasePropio || sanas[0].Registrado != m.Registrado {
		t.Fatalf("fila = %+v, want la persistida", sanas[0])
	}
	if len(sanas[0].Eslabones) != 1 || sanas[0].Eslabones[0] != domain.EslabonDeclarado {
		t.Fatalf("Eslabones = %v, want [declarado-por-operador]", sanas[0].Eslabones)
	}

	ok, err := st2.Olvidar("prenter-marketplace")
	if err != nil || !ok {
		t.Fatalf("Olvidar = (%v, %v), want (true, nil)", ok, err)
	}
	if ok, _ := st2.Olvidar("prenter-marketplace"); ok {
		t.Fatal("Olvidar dos veces devolvió true la segunda")
	}
	sanas, _ = st2.Listar()
	if len(sanas) != 0 {
		t.Fatalf("tras Olvidar: sanas = %v, want 0", sanas)
	}
}

// Upsert sin nombre es un error explícito, no una fila fantasma.
func TestStoreUpsertSinNombreEsError(t *testing.T) {
	st, _ := storeEn(t, "")
	if err := st.Upsert(domain.MarketplaceConocido{}); err == nil {
		t.Fatal("Upsert sin nombre devolvió nil error")
	}
}

// El archivo del registro guarda SOLO lo que el registro posee (design.md §4.1): los `eslabones`
// (collect-all) y la `lectura` (caché) son DERIVADOS y no deben ensuciar un archivo que el
// operador tiene que poder abrir y entender a mano. Bug cosmético cazado en el E2E vivo.
func TestStorePersisteSoloLoQuePosee(t *testing.T) {
	st, path := storeEn(t, "")
	if err := st.Upsert(domain.MarketplaceConocido{
		Nombre: "prenter-marketplace", Repo: "github.com/alpacapurpura/prenter-marketplace",
		Clase: domain.ClasePropio, Registrado: "2026-07-25T14:02:11Z",
		// Derivados que NO deben viajar al archivo:
		Eslabones:       []domain.EslabonMarketplace{domain.EslabonDeclarado},
		Lectura:         domain.EstadoLectura{Tipo: domain.LecturaLeida, Entradas: 2},
		InstallLocation: "/checkout", CCActualizado: "2026-07-10T00:36:43.459Z",
	}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path) //nolint:gosec // G304: temporal del propio test.
	if err != nil {
		t.Fatal(err)
	}
	for _, ruido := range []string{"eslabones", "lectura", "install_location", "cc_actualizado"} {
		if strings.Contains(string(b), ruido) {
			t.Fatalf("el archivo persiste el derivado %q:\n%s", ruido, b)
		}
	}
	for _, esperado := range []string{"nombre", "repo", "clase", "registrado"} {
		if !strings.Contains(string(b), esperado) {
			t.Fatalf("falta %q en el archivo:\n%s", esperado, b)
		}
	}
}
