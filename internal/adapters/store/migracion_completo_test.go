package store

// A-3 (auditoría 2026-07-26): `TestDeV1aV2ConservaTodoElFixtureReal` se presenta como
// «CAMPO POR CAMPO», y lo es — pero sobre el registro REAL del operador, que no contiene 5
// de los 20 campos de `sesionV1` (`cadena_cc`, `cerrada_en`, `checkpoint`, `parked`,
// `puesto`). Un test campo-por-campo no puede comparar un campo que el insumo no trae:
// tirar `Checkpoint` o `CadenaCC` del migrador dejaba `go test ./...` VERDE.
//
// Son justo los dos que la enmienda F-3 volvió portantes (CV-D11 promete retomar con
// `CadenaCC`/`Checkpoint`/`Cwd` intactos). El síntoma para el operador sería mudo: una
// conversación rotada que al retomarse arranca amnésica.
//
// Acá va el fixture SINTÉTICO con los 20 campos poblados. El real NO se reemplaza: su
// valor es otro — prueba que el archivo del operador sobrevive byte a byte.
//
// La red no es sólo la comparación: `TestElFixtureCompletoPueblaTodosLosCamposDeV1` es un
// guard REFLEXIVO que falla si alguien agrega un campo a `sesionV1` y no lo puebla acá.
// Sin ese guard, este archivo envejecería igual que el que vino a reparar.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fixtureCompleto copia el fixture sintético a `dir` y devuelve su ruta.
func fixtureCompleto(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-completo.json"))
	if err != nil {
		t.Fatalf("fixture completo: %v", err)
	}
	ruta := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(ruta, b, 0o600); err != nil { //nolint:gosec // ruta bajo t.TempDir().
		t.Fatalf("fixture completo: %v", err)
	}
	return ruta
}

// TestElFixtureCompletoPueblaTodosLosCamposDeV1 — el guard que impide que A-3 vuelva.
// Recorre por reflexión TODOS los campos de `sesionV1` y exige que al menos una sesión del
// fixture los traiga NO-CERO. Un campo nuevo sin poblar rompe acá, con su nombre, antes de
// que ningún test campo-por-campo pueda mentir que lo comparó.
func TestElFixtureCompletoPueblaTodosLosCamposDeV1(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-completo.json"))
	if err != nil {
		t.Fatal(err)
	}
	var viejas []sesionV1
	if uerr := json.Unmarshal(crudo, &viejas); uerr != nil {
		t.Fatal(uerr)
	}
	if len(viejas) == 0 {
		t.Fatal("el fixture completo quedó vacío")
	}

	tipo := reflect.TypeOf(sesionV1{})
	for i := range tipo.NumField() {
		campo := tipo.Field(i)
		poblado := false
		for j := range viejas {
			v := reflect.ValueOf(viejas[j]).Field(i)
			if !v.IsZero() {
				poblado = true
				break
			}
		}
		if !poblado {
			t.Errorf("el campo %q (json %q) NO está poblado en ninguna sesión del fixture completo — "+
				"un test campo-por-campo no puede comparar lo que el insumo no trae (A-3)",
				campo.Name, campo.Tag.Get("json"))
		}
	}

	// El número se afirma para que agregar un campo a `sesionV1` sea una decisión visible:
	// la struct está congelada a propósito («esta struct no se toca nunca más»).
	if n := tipo.NumField(); n != 20 {
		t.Errorf("`sesionV1` tiene %d campos, el fixture se armó para 20 — repoblá el fixture y actualizá el número", n)
	}
}

// TestDeV1aV2ConservaLos20CamposDeV1 — la comparación campo por campo sobre un insumo que
// SÍ trae los 20. Tirar cualquiera de ellos del migrador pone este test rojo (verificado
// por mutación sobre los 20, incluidos `Checkpoint` y `CadenaCC`, que sobrevivían).
func TestDeV1aV2ConservaLos20CamposDeV1(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureCompleto(t, dir)
	crudo, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	var viejas []sesionV1
	if uerr := json.Unmarshal(crudo, &viejas); uerr != nil {
		t.Fatal(uerr)
	}

	reg, _, err := AbrirRegistro(filepath.Join(dir, "sesiones.json"), legado, "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	nuevas, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(nuevas) != len(viejas) {
		t.Fatalf("sesiones migradas = %d, quiero %d", len(nuevas), len(viejas))
	}

	for i, v := range viejas {
		n := nuevas[i]
		t.Run(v.ID, func(t *testing.T) {
			// ── Los 12 que se quedan en la SESIÓN ────────────────────────────────
			igual(t, "id", n.ID, v.ID)
			igual(t, "frente", n.Frente, v.Frente)
			igual(t, "arnes", n.Arnes, v.Arnes)
			igual(t, "empresa", n.Empresa, v.Empresa)
			igual(t, "puesto", n.Puesto, v.Puesto)
			igual(t, "salud", n.Salud, v.Salud)
			igual(t, "status", string(n.Status), v.Status)
			igual(t, "view", n.View, v.View)
			igual(t, "parked", n.Parked, v.Parked)
			igual(t, "reparacion", n.Reparacion, v.Reparacion)
			igual(t, "cwd", n.Cwd, v.Cwd)
			igual(t, "cerrada_en", n.CerradaEn, v.CerradaEn)

			if len(n.Conversaciones) != 1 {
				t.Fatalf("conversaciones = %d, quiero 1", len(n.Conversaciones))
			}
			c := n.Conversaciones[0]

			// ── Los 8 que BAJAN a la conversación ────────────────────────────────
			igual(t, "claude_session_id", c.ClaudeSessionID, v.ClaudeSessionID)
			igual(t, "model", c.Model, v.Model)
			igual(t, "ctx_pct", c.CtxPct, v.CtxPct)
			igual(t, "rotacion_pendiente", c.RotacionPendiente, v.RotacionPendiente)
			igual(t, "checkpoint", c.Checkpoint, v.Checkpoint)
			igualSlice(t, "ctx_hist", c.CtxHist, v.CtxHist)
			igualSlice(t, "cadena_cc", c.CadenaCC, v.CadenaCC)
			igualSlice(t, "conv", c.Conv, v.Conv)

			// ── Lo que NACE, y lo que deliberadamente NO ─────────────────────────
			if c.ID == "" || c.CreadaEn == "" || !c.Activa {
				t.Errorf("la conversación migrada nació incompleta: %+v", c)
			}
			if c.UltimaInteraccion != "" {
				t.Errorf("ultima_interaccion = %q — esa fecha no existe en v1 y no se inventa (BR-CV-14)", c.UltimaInteraccion)
			}
			if c.TituloEditado {
				t.Error("a una conversación migrada no la tituló nadie a mano")
			}
			if c.Conv == nil {
				t.Error("conv es null: «leí y no había turnos» y «no cargué» tienen que distinguirse")
			}
		})
	}
}

func igual[T comparable](t *testing.T, campo string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s se perdió en la migración: got %v, want %v", campo, got, want)
	}
}

func igualSlice[T comparable](t *testing.T, campo string, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s se perdió en la migración: len %d, want %d (got %v, want %v)", campo, len(got), len(want), got, want)
		return
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] cambió: got %v, want %v", campo, i, got[i], want[i])
		}
	}
}

// TestElFixtureRealNoTraeLosCamposQueElSinteticoCubre deja el hecho de A-3 ESCRITO como
// test, no como prosa: si algún día el fixture real llega a traerlos, este test avisa y el
// sintético deja de ser necesario para esos campos. Es la documentación que no puede
// envejecer en silencio.
func TestElFixtureRealNoTraeLosCamposQueElSinteticoCubre(t *testing.T) {
	crudo, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-real.json"))
	if err != nil {
		t.Fatal(err)
	}
	var filas []map[string]json.RawMessage
	if uerr := json.Unmarshal(crudo, &filas); uerr != nil {
		t.Fatal(uerr)
	}
	ausentes := map[string]bool{
		"cadena_cc": true, "cerrada_en": true, "checkpoint": true, "parked": true, "puesto": true,
	}
	for _, f := range filas {
		for k := range f {
			delete(ausentes, k)
		}
	}
	if len(ausentes) != 5 {
		t.Errorf("el fixture real cambió: los campos que NO traía ahora son %v — "+
			"revisá si el sintético sigue haciendo falta para todos", ausentes)
	}
	// Y el sintético sí los trae: si no, la red de A-3 no existe.
	compl, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-completo.json"))
	if err != nil {
		t.Fatal(err)
	}
	var completas []map[string]json.RawMessage
	if uerr := json.Unmarshal(compl, &completas); uerr != nil {
		t.Fatal(uerr)
	}
	for campo := range ausentes {
		visto := false
		for _, f := range completas {
			if _, ok := f[campo]; ok {
				visto = true
				break
			}
		}
		if !visto {
			t.Errorf("el fixture sintético dejó de traer %q — es justo el campo que el real no cubre", campo)
		}
	}
}
