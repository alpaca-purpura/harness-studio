package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// TestArchivoCorruptoSePreservaYSeDice (E-44) — un registro ilegible se pone en cuarentena
// con su sello, el registro arranca vacío y el informe trae la RUTA para que el arranque la
// loguee. Lo que NO pasa: sobreescribirlo. Un JSON roto puede ser recuperable a mano, y
// pisarlo lo vuelve irrecuperable. Check `corrupto-se-preserva`.
func TestArchivoCorruptoSePreservaYSeDice(t *testing.T) {
	casos := []struct {
		nombre string
		crudo  string
	}{
		{"json truncado", `[{"id":"s1","frente":"repar`},
		{"basura", "\x00\x01binario"},
		{"objeto sin schema_version", `{"sesiones":[{"id":"s1"}]}`},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			dir := t.TempDir()
			v2 := filepath.Join(dir, "sesiones.json")
			if err := os.WriteFile(v2, []byte(c.crudo), 0o600); err != nil {
				t.Fatal(err)
			}

			reg, inf, err := AbrirRegistro(v2, "", "2607262100", nil)
			if err != nil {
				t.Fatalf("un archivo roto no puede tumbar el arranque: %v", err)
			}
			if !inf.Corrupto || inf.CuarentenaEn == "" {
				t.Fatalf("informe = %+v — la cuarentena tiene que viajar con su ruta", inf)
			}
			if !strings.Contains(inf.CuarentenaEn, ".corrupto-2607262100") {
				t.Errorf("la cuarentena no lleva el sello: %s", inf.CuarentenaEn)
			}
			// El contenido original está entero en la cuarentena.
			guardado, rerr := os.ReadFile(inf.CuarentenaEn)
			if rerr != nil {
				t.Fatalf("el archivo en cuarentena no existe: %v", rerr)
			}
			if string(guardado) != c.crudo {
				t.Error("el archivo en cuarentena no es el original byte a byte")
			}
			// Y la ruta original quedó libre, no con una copia del roto.
			if _, serr := os.Stat(v2); !os.IsNotExist(serr) {
				t.Error("el archivo roto sigue en su ruta: el próximo arranque lo volvería a encontrar")
			}
			sesiones, lerr := reg.Load(context.Background())
			if lerr != nil || len(sesiones) != 0 {
				t.Errorf("registro tras la cuarentena = %+v (err %v), quiero vacío", sesiones, lerr)
			}
		})
	}
}

// TestCuarentenaNoPisaLaAnterior — dos corrupciones con el mismo sello son dos archivos,
// y la segunda no puede borrar la evidencia de la primera.
func TestCuarentenaNoPisaLaAnterior(t *testing.T) {
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")

	rutas := make([]string, 0, 2)
	for i, contenido := range []string{"roto-uno", "roto-dos"} {
		if err := os.WriteFile(v2, []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
		_, inf, err := AbrirRegistro(v2, "", "2607262100", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !inf.Corrupto {
			t.Fatalf("corrida %d: no se encuarentenó", i)
		}
		rutas = append(rutas, inf.CuarentenaEn)
	}
	if rutas[0] == rutas[1] {
		t.Fatal("la segunda cuarentena pisó a la primera")
	}
	for i, r := range rutas {
		b, err := os.ReadFile(r) //nolint:gosec // ruta que devolvió el propio informe.
		if err != nil {
			t.Fatalf("falta la cuarentena %d: %v", i, err)
		}
		quiero := []string{"roto-uno", "roto-dos"}[i]
		if string(b) != quiero {
			t.Errorf("cuarentena %d = %q, quiero %q", i, b, quiero)
		}
	}
}

// TestCorrupcionNoEsSembrable (E-44) — un registro vacío POR CORRUPCIÓN no es un primer
// arranque, y el que decide si se siembra tiene que poder distinguirlos. Si no, la semilla
// ilustrativa tapa la corrupción y el operador ve tres sesiones de ejemplo donde antes
// tenía las suyas.
func TestCorrupcionNoEsSembrable(t *testing.T) {
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")
	if err := os.WriteFile(v2, []byte("roto"), 0o600); err != nil {
		t.Fatal(err)
	}
	reg, inf, err := AbrirRegistro(v2, "", "sello", nil)
	if err != nil || !inf.Corrupto {
		t.Fatalf("inf=%+v err=%v", inf, err)
	}
	if reg.Sembrable() {
		t.Error("un registro en cuarentena se declara sembrable: la semilla taparía la corrupción")
	}

	// Contraste: un primer arranque de verdad SÍ es sembrable.
	limpio, _, err := AbrirRegistro(filepath.Join(t.TempDir(), "sesiones.json"), "", "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !limpio.Sembrable() {
		t.Error("un primer arranque tiene que poder sembrarse")
	}
}

// TestEsquemaFuturoNoSePisa (modo E) — un archivo escrito por un binario más nuevo deja el
// registro en solo-lectura y NO se toca un byte. Degradar a vacío y después persistir
// destruiría el archivo nuevo con el binario viejo: el único modo de fallo irreversible.
// Check `esquema-futuro-no-se-degrada`.
func TestEsquemaFuturoNoSePisa(t *testing.T) {
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")
	delFuturo := `{"schema_version":99,"escrito_por":"9.9.9","sesiones":[{"id":"s-del-futuro"}]}`
	if err := os.WriteFile(v2, []byte(delFuturo), 0o600); err != nil {
		t.Fatal(err)
	}

	reg, inf, err := AbrirRegistro(v2, "", "sello", nil)
	if err != nil {
		t.Fatalf("un esquema futuro no tumba el arranque, lo bloquea: %v", err)
	}
	if !inf.EsquemaFuturo {
		t.Fatalf("informe = %+v — el esquema futuro tiene que decirse", inf)
	}
	bloqueado, motivo := reg.SoloLectura()
	if !bloqueado {
		t.Fatal("el registro no quedó en solo-lectura")
	}
	if !strings.Contains(motivo, "99") || !strings.Contains(motivo, "2") {
		t.Errorf("el motivo no nombra las dos versiones: %q", motivo)
	}
	if reg.Sembrable() {
		t.Error("un registro bloqueado no se siembra: taparía el archivo del futuro")
	}

	// Ninguna escritura pasa, y el archivo no cambió un byte.
	err = reg.Save(context.Background(), []domain.Session{{ID: "s-del-pasado"}})
	if !errors.Is(err, ErrSoloLectura) {
		t.Errorf("Save en solo-lectura = %v, quiero ErrSoloLectura", err)
	}
	despues, rerr := os.ReadFile(v2) //nolint:gosec // ruta del propio test.
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(despues) != delFuturo {
		t.Error("el archivo del futuro cambió: es exactamente lo que este modo existe para impedir")
	}
}
