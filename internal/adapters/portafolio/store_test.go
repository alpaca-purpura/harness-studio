package portafolio_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/domain"
)

func entrada(id string) domain.EntradaPortafolio {
	return domain.EntradaPortafolio{Identidad: domain.IdentidadArnes{ID: id}}
}

// TestUpsertNoPisaEscrituraDeOtraInstancia es la regresión de la Fase 1 (D2/D3) a nivel de
// Store: reproduce el bug CONFIRMADO en `cmd/arnesia/main.go` — `runServe` y `runPortafolio`
// abren cada uno su PROPIA instancia de `portafolio.Store` sobre el mismo archivo, como dos
// procesos independientes harían. Sin el reload-antes-de-mutar (D2), el Upsert de `s2`
// pisaría el archivo con SOLO lo que `s2` tenía en memoria al abrir + su propio cambio,
// perdiendo silenciosamente lo que `s1` escribió después.
func TestUpsertNoPisaEscrituraDeOtraInstancia(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")

	s1, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	if uerr := s1.Upsert(entrada("de-s1-inicial")); uerr != nil {
		t.Fatal(uerr)
	}

	// s2 abre y carga el archivo TAL COMO ESTÁ ahora (solo "de-s1-inicial") — simula el
	// segundo proceso (CLI standalone) que arranca mientras el primero (daemon) ya corría.
	s2, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}

	// s1 sigue vivo y agrega una entrada MÁS — s2 todavía no lo sabe: su caché en memoria
	// quedó desactualizado en el instante en que s1 escribe esto.
	if uerr := s1.Upsert(entrada("de-s1-despues")); uerr != nil {
		t.Fatal(uerr)
	}

	// s2 muta algo propio. Sin el reload-under-lock, esto reescribiría el archivo con
	// SOLO ["de-s1-inicial", "de-s2"] — perdiendo "de-s1-despues" para siempre.
	if uerr := s2.Upsert(entrada("de-s2")); uerr != nil {
		t.Fatal(uerr)
	}

	s3, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	sanas, _ := s3.Listar()
	got := make(map[string]bool, len(sanas))
	for _, e := range sanas {
		got[e.Identidad.Clave()] = true
	}
	for _, clave := range []string{"sin-home~de-s1-inicial~", "sin-home~de-s1-despues~", "sin-home~de-s2~"} {
		if !got[clave] {
			t.Errorf("tras el round-trip, falta %q — se perdió una escritura entre instancias (got=%v)", clave, got)
		}
	}
	if len(sanas) != 3 {
		t.Fatalf("quiero 3 entradas sanas en el archivo final, got %d: %v", len(sanas), got)
	}
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
	if uerr := s.Upsert(entrada("sana-3")); uerr != nil {
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
// abre igual (0 sanas, 0 corruptas EN MEMORIA — D1) porque el blob entero se pone en
// CUARENTENA en disco (`<ruta>.corrupto-<sello>`), intacto, en vez de intentar reinsertar
// bytes no-JSON dentro de un envelope que sí debe serlo. Un write posterior nunca lo toca.
func TestStoreArchivoTotalmenteIlegible(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	original := "{esto ni siquiera es JSON válido, sin comillas ni estructura"
	escribir(t, path, original)

	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore no debe fallar ni con el archivo totalmente roto: %v", err)
	}
	sanas, corruptas := s.Listar()
	if len(sanas) != 0 || len(corruptas) != 0 {
		t.Fatalf("sanas=%d corruptas=%d, quiero 0 y 0 (el blob se fue a cuarentena, no queda en memoria)", len(sanas), len(corruptas))
	}

	// El archivo original desapareció de su ruta (se RENOMBRÓ, no se copió)...
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("la ruta original debía quedar libre tras la cuarentena, stat err=%v", statErr)
	}
	// ...y sus bytes sobreviven intactos en algún `<path>.corrupto-*`.
	entries, rerr := os.ReadDir(filepath.Dir(path))
	if rerr != nil {
		t.Fatal(rerr)
	}
	var cuarentena string
	for _, e := range entries {
		if strings.Contains(e.Name(), ".corrupto-") {
			cuarentena = filepath.Join(filepath.Dir(path), e.Name())
		}
	}
	if cuarentena == "" {
		t.Fatal("quiero un archivo `.corrupto-*` en el mismo directorio")
	}
	b, rerr := os.ReadFile(cuarentena) //nolint:gosec // G304: temporal del propio test.
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != original {
		t.Errorf("la cuarentena debe conservar los bytes originales byte a byte, got %q", string(b))
	}

	// Un write posterior arranca de cero (store vacío) y no debe crashear.
	if uerr := s.Upsert(entrada("nueva")); uerr != nil {
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

// TestUpsertMergeUneFacetas cubre S1-D3 (cierra GAP-3): Empresas/Registries se UNEN entre
// upserts de la misma identidad (orden estable, sin duplicados) y un upsert con la faceta
// vacía NO borra lo ya persistido — el escaneo fresco nunca degrada el dato acumulado.
func TestUpsertMergeUneFacetas(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	id := domain.IdentidadArnes{Home: "github.com/o/r", ID: "harness-x"}

	e1 := domain.EntradaPortafolio{
		Identidad:  id,
		Empresas:   []string{"alpacapurpura"},
		Registries: []string{"github.com/o/r"},
	}
	if err := s.Upsert(e1); err != nil {
		t.Fatal(err)
	}

	// e2 trae una faceta NUEVA (empresa distinta) y repite el mismo registry: unión sin
	// duplicar, orden estable (existente primero).
	e2 := domain.EntradaPortafolio{
		Identidad:  id,
		Empresas:   []string{"vitalia"},
		Registries: []string{"github.com/o/r"},
	}
	if err := s.Upsert(e2); err != nil {
		t.Fatal(err)
	}

	sanas, _ := s.Listar()
	if len(sanas) != 1 {
		t.Fatalf("quiero 1 entrada, got %d", len(sanas))
	}
	if got := sanas[0].Empresas; len(got) != 2 || got[0] != "alpacapurpura" || got[1] != "vitalia" {
		t.Fatalf("Empresas = %v, quiero unión estable [alpacapurpura vitalia]", got)
	}
	if got := sanas[0].Registries; len(got) != 1 || got[0] != "github.com/o/r" {
		t.Fatalf("Registries = %v, quiero dedup [github.com/o/r]", got)
	}

	// e3 llega con las facetas VACÍAS (p.ej. un re-escaneo que no resolvió empresa ni
	// registry): NO debe borrar lo ya persistido.
	e3 := domain.EntradaPortafolio{Identidad: id}
	if err := s.Upsert(e3); err != nil {
		t.Fatal(err)
	}
	sanas, _ = s.Listar()
	if got := sanas[0].Empresas; len(got) != 2 {
		t.Fatalf("un upsert con Empresas vacío NO debe borrar lo existente, got %v", got)
	}
	if got := sanas[0].Registries; len(got) != 1 {
		t.Fatalf("un upsert con Registries vacío NO debe borrar lo existente, got %v", got)
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
	if uerr := s.Upsert(entrada("x")); uerr != nil {
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

// TestStoreEsquemaFuturoEsSoloLectura cubre D1 Modo E: un `version` mayor al que este
// binario entiende significa que lo escribió un binario más nuevo. El store abre en
// solo-lectura (Listar sigue andando) y CUALQUIER escritura se rechaza — nunca se pisa un
// archivo que este binario no sabe interpretar completo.
func TestStoreEsquemaFuturoEsSoloLectura(t *testing.T) {
	path := filepath.Join(t.TempDir(), "portafolio.json")
	original := `{"version":99,"entradas":[{"identidad":{"id":"del-futuro"}}]}`
	escribir(t, path, original)

	s, err := portafolio.NewStore(path)
	if err != nil {
		t.Fatalf("NewStore no debe fallar por un esquema futuro: %v", err)
	}
	sanas, corruptas := s.Listar()
	if len(sanas) != 0 || len(corruptas) != 0 {
		t.Fatalf("esquema futuro: sanas=%d corruptas=%d, quiero 0 y 0 (no se interpretan filas de una forma que no se conoce)", len(sanas), len(corruptas))
	}

	if uerr := s.Upsert(entrada("nueva")); uerr == nil {
		t.Fatal("Upsert sobre un esquema futuro debe rechazarse, no pisarlo en silencio")
	}

	b, rerr := os.ReadFile(path) //nolint:gosec // G304: temporal del propio test.
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != original {
		t.Errorf("el archivo de un esquema futuro no debe tocarse ni un byte, got %q", string(b))
	}
}
