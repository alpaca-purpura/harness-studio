package store

// A-8 (auditoría 2026-07-26): `respaldarSiEsViejoLocked` es la red que impide pisar el
// registro v1 del operador por un camino de escritura no previsto, y era **la única línea
// del paquete sin test**. Su propio comentario dice cuánto vale: «el respaldo va acá y no
// en el llamador para que no haya un camino de escritura que se lo saltee — el que se lo
// saltea es el que borra el archivo del operador».
//
// La mutación que sobrevivía: reemplazar el bloque por `if false { return nil }` dejaba
// VERDE `./internal/adapters/store/` y `./docs/architecture/fitness/`. El check
// `migracion-forward-only-con-respaldo` sí se cae si se saca el respaldo del camino de
// `AbrirRegistro` — pero NO si se saca el de `Save`, que es el que cubre todos los demás.
//
// Por eso estos tests entran por `Save()` DIRECTO, sin pasar por `AbrirRegistro`: es el
// camino que no tenía red.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestSaveRespaldaElV1EnDiscoAntesDePisarlo — el caso que la auditoría pidió: registro v1
// en disco + `Save()` directo ⇒ existe `<ruta>.v1-<sello>.bak` con el contenido ORIGINAL.
func TestSaveRespaldaElV1EnDiscoAntesDePisarlo(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sessions.json")
	// Un v1 legítimo: array desnudo, sin sobre versionado (`detectarVersion` lo lee como 1).
	original, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-real.json"))
	if err != nil {
		t.Fatal(err)
	}
	if werr := os.WriteFile(ruta, original, 0o600); werr != nil { //nolint:gosec // t.TempDir().
		t.Fatal(werr)
	}

	// Save DIRECTO — sin AbrirRegistro, que es el camino que ya tenía red.
	reg := &Registry{path: ruta, sello: "0.2.25.2607270900"}
	if serr := reg.Save(context.Background(), []domain.Session{{
		ID:             "s-nueva",
		Conversaciones: []domain.Conversacion{{ID: "cv1", Activa: true, Titulo: "t"}},
	}}); serr != nil {
		t.Fatalf("save: %v", serr)
	}

	bak := ruta + ".v1-0.2.25.2607270900.bak"
	respaldado, err := os.ReadFile(bak) //nolint:gosec // ruta derivada bajo t.TempDir().
	if err != nil {
		t.Fatalf("NO se respaldó el v1 antes de pisarlo — este es el camino que borra el "+
			"archivo del operador (A-8): %v", err)
	}
	if string(respaldado) != string(original) {
		t.Errorf("el respaldo no es el archivo original: %d B vs %d B", len(respaldado), len(original))
	}

	// Y lo que quedó en la ruta viva es lo nuevo, con su sobre: el respaldo no reemplaza
	// la escritura, la precede.
	ahora, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(ahora), "s-nueva") {
		t.Error("la escritura nueva no llegó al archivo vivo")
	}
	if string(ahora) == string(original) {
		t.Error("el archivo vivo no cambió: Save no escribió")
	}
}

// TestSaveNoRespaldaDosVecesElMismoEsquema — el respaldo corre porque en disco hay una
// versión ANTERIOR, no en cada escritura. Sin esta mitad, un test que sólo mira «existe el
// .bak» pasaría igual con un respaldo incondicional, que llenaría el disco del operador.
func TestSaveNoRespaldaDosVecesElMismoEsquema(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sessions.json")
	reg := &Registry{path: ruta, sello: "sello-x"}
	ses := []domain.Session{{
		ID:             "s1",
		Conversaciones: []domain.Conversacion{{ID: "cv1", Activa: true, Titulo: "t"}},
	}}

	// 1.ª escritura: no hay nada en disco ⇒ nada que respaldar.
	if err := reg.Save(context.Background(), ses); err != nil {
		t.Fatal(err)
	}
	// 2.ª escritura: en disco ya está el esquema ACTUAL ⇒ tampoco se respalda.
	if err := reg.Save(context.Background(), ses); err != nil {
		t.Fatal(err)
	}

	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entradas {
		if strings.HasSuffix(e.Name(), ".bak") {
			t.Errorf("se respaldó un esquema que ya estaba al día: %s — el respaldo es por "+
				"versión anterior, no por cada escritura", e.Name())
		}
	}
}

// TestSaveNoRespaldaUnArchivoIlegible — un archivo roto lo preserva la CUARENTENA de
// `AbrirRegistro`, entero; respaldarlo también dejaría dos copias del mismo problema.
func TestSaveNoRespaldaUnArchivoIlegible(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(ruta, []byte("{no soy json"), 0o600); err != nil {
		t.Fatal(err)
	}
	reg := &Registry{path: ruta, sello: "sello-y"}
	if err := reg.Save(context.Background(), []domain.Session{{
		ID:             "s1",
		Conversaciones: []domain.Conversacion{{ID: "cv1", Activa: true, Titulo: "t"}},
	}}); err != nil {
		t.Fatalf("save sobre un ilegible tiene que seguir: %v", err)
	}
	entradas, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entradas {
		if strings.HasSuffix(e.Name(), ".bak") {
			t.Errorf("se respaldó un archivo ilegible (%s): de eso se ocupa la cuarentena", e.Name())
		}
	}
}

// TestElRespaldoDeSaveConservaLasSesionesDelOperador — el respaldo no vale por existir,
// vale por lo que se puede recuperar de él. Se decodifica y se cuentan las sesiones.
func TestElRespaldoDeSaveConservaLasSesionesDelOperador(t *testing.T) {
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sessions.json")
	original, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-real.json"))
	if err != nil {
		t.Fatal(err)
	}
	if werr := os.WriteFile(ruta, original, 0o600); werr != nil { //nolint:gosec // t.TempDir().
		t.Fatal(werr)
	}
	var antes []sesionV1
	if uerr := json.Unmarshal(original, &antes); uerr != nil {
		t.Fatal(uerr)
	}

	reg := &Registry{path: ruta, sello: "s"}
	if serr := reg.Save(context.Background(), nil); serr != nil {
		t.Fatalf("save: %v", serr)
	}

	b, err := os.ReadFile(ruta + ".v1-s.bak") //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatalf("sin respaldo: %v", err)
	}
	var despues []sesionV1
	if uerr := json.Unmarshal(b, &despues); uerr != nil {
		t.Fatalf("el respaldo no se puede volver a leer: %v", uerr)
	}
	if len(despues) != len(antes) {
		t.Fatalf("el respaldo tiene %d sesiones, el original %d", len(despues), len(antes))
	}
	for i := range antes {
		if despues[i].ID != antes[i].ID || len(despues[i].Conv) != len(antes[i].Conv) {
			t.Errorf("sesión %d cambió en el respaldo: %q/%d vs %q/%d",
				i, despues[i].ID, len(despues[i].Conv), antes[i].ID, len(antes[i].Conv))
		}
	}
}
