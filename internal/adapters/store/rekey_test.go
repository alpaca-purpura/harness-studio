package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// portafolioFake modela el resolvedor que el composition root inyecta: para cada id pelado,
// las claves calificadas candidatas y el cwd al que resuelve cada una.
type portafolioFake struct {
	// porID: id pelado → claves calificadas de las entradas con ese id.
	porID map[string][]string
	// cwdDe: clave calificada → cwd al que resuelve.
	cwdDe map[string]string
}

// clave implementa el algoritmo de CV-D16, el mismo que cablea el daemon: primero por cwd
// (la vía fuerte), después por id (el fallback para las que nunca spawnearon), y si no hay
// exactamente una candidata NO se decide.
func (p portafolioFake) clave(idPelado, cwd string) (string, bool, string) {
	if strings.Contains(idPelado, "~") {
		return "", false, "ya-calificada"
	}
	candidatas := p.porID[idPelado]
	if cwd != "" {
		var porCwd []string
		for _, c := range candidatas {
			if p.cwdDe[c] == cwd {
				porCwd = append(porCwd, c)
			}
		}
		if len(porCwd) == 1 {
			return porCwd[0], true, "resuelta-por-cwd"
		}
	}
	switch len(candidatas) {
	case 1:
		return candidatas[0], true, "resuelta-por-id"
	case 0:
		return "", false, "sin-candidata"
	default:
		return "", false, "ambigua"
	}
}

// registroReal arma un registro v2 con las 5 sesiones del caso real de CV-D16 y devuelve el
// Registry ya escrito.
func registroReal(t *testing.T) (*Registry, string) {
	t.Helper()
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sesiones.json")
	reg := &Registry{path: ruta, sello: "sello"}
	if err := reg.Save(context.Background(), []domain.Session{
		{ID: "s25123a2c", Arnes: "vitalia", Cwd: "/home/x/luana-vitalia/vitalia"},
		{ID: "s6165ac75", Arnes: "sin-home~vitalia~vitalia", Cwd: "/home/x/luana-vitalia/vitalia"},
		{ID: "s0fec7798", Arnes: "vitalia"},
		{ID: "sfc512b15", Arnes: "vitalia"},
		{ID: "s78b3aeeb", Arnes: "arnesia"},
	}); err != nil {
		t.Fatal(err)
	}
	return reg, ruta
}

func portafolioDelCasoReal() portafolioFake {
	return portafolioFake{
		porID: map[string][]string{
			"vitalia": {"sin-home~vitalia~vitalia"},
			"arnesia": {"sin-home~arnesia~"},
		},
		cwdDe: map[string]string{
			"sin-home~vitalia~vitalia": "/home/x/luana-vitalia/vitalia",
			"sin-home~arnesia~":        "/home/x/Proyectos/harness-studio",
		},
	}
}

func filaDe(filas []Recalibracion, id string) (Recalibracion, bool) {
	for _, f := range filas {
		if f.SesionID == id {
			return f, true
		}
	}
	return Recalibracion{}, false
}

// TestReKeyPorCwd — la vía fuerte: con cwd, la llave se resuelve contra la entrada que
// realmente vive en ese directorio.
func TestReKeyPorCwd(t *testing.T) {
	sesiones := []domain.Session{
		{ID: "s1", Arnes: "vitalia", Cwd: "/home/x/luana-vitalia/vitalia"},
	}
	pf := portafolioFake{
		porID: map[string][]string{"vitalia": {"sin-home~vitalia~vitalia", "acme~vitalia~otra"}},
		cwdDe: map[string]string{
			"sin-home~vitalia~vitalia": "/home/x/luana-vitalia/vitalia",
			"acme~vitalia~otra":        "/home/x/otro-lado",
		},
	}
	filas := reKey(sesiones, pf.clave)
	if sesiones[0].Arnes != "sin-home~vitalia~vitalia" {
		t.Errorf("arnes = %q, quiero la calificada del cwd", sesiones[0].Arnes)
	}
	if filas[0].Motivo != "resuelta-por-cwd" {
		t.Errorf("motivo = %q — el cwd desambigua entre dos candidatas y tiene que decirlo", filas[0].Motivo)
	}
}

// TestReKeyPorIdCuandoNoHayCwd — el caso REAL de 3 de las 4: nunca spawnearon, así que no
// tienen cwd. Sin fallback por id se quedarían peladas para siempre.
func TestReKeyPorIdCuandoNoHayCwd(t *testing.T) {
	sesiones := []domain.Session{{ID: "s1", Arnes: "vitalia"}}
	filas := reKey(sesiones, portafolioDelCasoReal().clave)
	if sesiones[0].Arnes != "sin-home~vitalia~vitalia" {
		t.Errorf("arnes = %q, quiero la calificada resuelta por id", sesiones[0].Arnes)
	}
	if filas[0].Motivo != "resuelta-por-id" {
		t.Errorf("motivo = %q, quiero que diga que fue por el fallback", filas[0].Motivo)
	}
}

// TestReKeyAmbiguaNoDecide y TestReKeySinCandidataNoDecide — lo que no se puede decidir NO
// se decide: la llave queda como está y el motivo viaja. Elegir una de dos candidatas al
// azar fusionaría dos identidades por coincidencia.
func TestReKeyAmbiguaNoDecide(t *testing.T) {
	sesiones := []domain.Session{{ID: "s1", Arnes: "vitalia"}}
	pf := portafolioFake{porID: map[string][]string{"vitalia": {"a~vitalia~x", "b~vitalia~y"}}}
	filas := reKey(sesiones, pf.clave)
	if sesiones[0].Arnes != "vitalia" {
		t.Errorf("arnes = %q — con dos candidatas la llave NO se toca", sesiones[0].Arnes)
	}
	if filas[0].Motivo != "ambigua" || filas[0].Movio() {
		t.Errorf("fila = %+v, quiero que declare la ambigüedad sin moverse", filas[0])
	}
}

func TestReKeySinCandidataNoDecide(t *testing.T) {
	sesiones := []domain.Session{{ID: "s1", Arnes: "fantasma"}}
	filas := reKey(sesiones, portafolioDelCasoReal().clave)
	if sesiones[0].Arnes != "fantasma" {
		t.Errorf("arnes = %q — sin candidata la llave NO se toca", sesiones[0].Arnes)
	}
	if filas[0].Motivo != "sin-candidata" || filas[0].Movio() {
		t.Errorf("fila = %+v, quiero que declare el hueco sin moverse", filas[0])
	}
}

// TestReKeyEmiteUnaFilaPorSesionSeMuevaONo — un silencio sería indistinguible de «no la
// miré». El caso real completo: 5 sesiones, 4 se mueven, 1 ya estaba calificada.
func TestReKeyEmiteUnaFilaPorSesionSeMuevaONo(t *testing.T) {
	reg, _ := registroReal(t)
	filas, err := reg.Recalibrar(portafolioDelCasoReal().clave, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(filas) != 5 {
		t.Fatalf("filas = %d, quiero una por sesión", len(filas))
	}
	quiero := map[string]string{
		"s25123a2c": "resuelta-por-cwd",
		"s6165ac75": "ya-calificada",
		"s0fec7798": "resuelta-por-id",
		"sfc512b15": "resuelta-por-id",
		"s78b3aeeb": "resuelta-por-id",
	}
	for id, motivo := range quiero {
		f, ok := filaDe(filas, id)
		if !ok {
			t.Errorf("falta la fila de %s", id)
			continue
		}
		if f.Motivo != motivo {
			t.Errorf("%s: motivo = %q, quiero %q", id, f.Motivo, motivo)
		}
	}
	// La ya calificada no se movió; las otras cuatro sí, y nada se perdió.
	sesiones, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(sesiones) != 5 {
		t.Fatalf("sesiones = %d — el re-key no puede perder ninguna", len(sesiones))
	}
	for _, s := range sesiones {
		if !strings.Contains(s.Arnes, "~") {
			t.Errorf("la sesión %s quedó con llave pelada: %q", s.ID, s.Arnes)
		}
	}
}

// TestReKeyEsIdempotente — la segunda corrida no mueve una sola fila y no reescribe.
func TestReKeyEsIdempotente(t *testing.T) {
	reg, ruta := registroReal(t)
	pf := portafolioDelCasoReal()
	if _, err := reg.Recalibrar(pf.clave, false); err != nil {
		t.Fatal(err)
	}
	primera, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}

	filas, err := reg.Recalibrar(pf.clave, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range filas {
		if f.Movio() {
			t.Errorf("la segunda corrida movió %s: %+v", f.SesionID, f)
		}
		if f.Motivo != "ya-calificada" {
			t.Errorf("%s: motivo = %q, quiero ya-calificada", f.SesionID, f.Motivo)
		}
	}
	segunda, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	if string(primera) != string(segunda) {
		t.Error("la segunda corrida reescribió el archivo sin haber movido nada")
	}
}

// TestDryRunNoEscribe — el dry-run es el paso que convierte «nada se borra» en algo
// observable: el operador ve la tabla ANTES de que se aplique.
func TestDryRunNoEscribe(t *testing.T) {
	reg, ruta := registroReal(t)
	antes, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	filas, err := reg.Recalibrar(portafolioDelCasoReal().clave, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(filas) != 5 {
		t.Errorf("el dry-run tiene que reportar las 5 filas, reportó %d", len(filas))
	}
	movidas := 0
	for _, f := range filas {
		if f.Movio() {
			movidas++
		}
	}
	if movidas != 4 {
		t.Errorf("el dry-run reporta %d movimientos, quiero 4", movidas)
	}
	despues, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	if string(antes) != string(despues) {
		t.Error("el dry-run escribió: entonces no es un dry-run")
	}
}

// TestRevertirRestauraSoloElArnes (procedimiento R2 de CV-D16) — deshacer el re-key sin
// perder lo que se hizo después. Es lo que hace que «reversible» no sea un eufemismo de
// «restaurá el backup y perdé los turnos de hoy».
func TestRevertirRestauraSoloElArnes(t *testing.T) {
	reg, _ := registroReal(t)
	respaldo, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, rerr := reg.Recalibrar(portafolioDelCasoReal().clave, false); rerr != nil {
		t.Fatal(rerr)
	}

	// Después de migrar, el operador siguió trabajando: turnos nuevos y una sesión nueva.
	sesiones, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	sesiones[0].Conversaciones = []domain.Conversacion{{
		ID: "cv-post", Titulo: "trabajo posterior a la migración", Activa: true,
		Conv: []domain.Turn{{Rol: domain.RolUser, Text: "esto pasó DESPUÉS"}},
	}}
	sesiones = append(sesiones, domain.Session{ID: "s-nueva", Arnes: "sin-home~nueva~"})
	if serr := reg.Save(context.Background(), sesiones); serr != nil {
		t.Fatal(serr)
	}

	filas, err := reg.RevertirLlaves(respaldo)
	if err != nil {
		t.Fatal(err)
	}

	vuelta, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// Las llaves volvieron.
	if vuelta[0].Arnes != "vitalia" {
		t.Errorf("arnes tras revertir = %q, quiero el pelado original", vuelta[0].Arnes)
	}
	// Y el trabajo posterior SOBREVIVIÓ.
	if len(vuelta[0].Conversaciones) != 1 || len(vuelta[0].Conversaciones[0].Conv) != 1 {
		t.Fatalf("revertir se llevó puesto el trabajo posterior: %+v", vuelta[0].Conversaciones)
	}
	if vuelta[0].Conversaciones[0].Conv[0].Text != "esto pasó DESPUÉS" {
		t.Error("el turno posterior a la migración se perdió al revertir")
	}
	// La sesión que nació después no estaba en el respaldo: se deja como está, con motivo.
	if len(vuelta) != 6 {
		t.Errorf("sesiones tras revertir = %d, quiero 6 — la nueva no se borra", len(vuelta))
	}
	f, ok := filaDe(filas, "s-nueva")
	if !ok || f.Motivo != "no-estaba-en-el-respaldo" || f.Movio() {
		t.Errorf("fila de la sesión nueva = %+v — revertir no puede inventarle una llave anterior", f)
	}
}

// TestReKeyEnLaMigracion — el re-key es el paso 4 de la migración, no un comando aparte:
// arrancar sobre el registro v1 con el resolvedor cableado deja las llaves calificadas y
// el informe con sus 5 filas.
func TestReKeyEnLaMigracion(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)

	reg, inf, err := AbrirRegistro(filepath.Join(dir, "sesiones.json"), legado, "sello",
		portafolioFake{
			porID: map[string][]string{
				"vitalia": {"sin-home~vitalia~vitalia"},
				"arnesia": {"sin-home~arnesia~"},
			},
			cwdDe: map[string]string{
				"sin-home~vitalia~vitalia": "/home/chalreme/Proyectos/luana-vitalia/vitalia",
			},
		}.clave)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Recalibradas) != 5 {
		t.Fatalf("recalibradas = %d, quiero una fila por sesión del registro real", len(inf.Recalibradas))
	}
	if !inf.Hubo() {
		t.Error("el informe no tiene nada que contar pese a haber migrado y recalibrado")
	}
	sesiones, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range sesiones {
		if !strings.Contains(s.Arnes, "~") {
			t.Errorf("la sesión %s quedó pelada tras migrar: %q", s.ID, s.Arnes)
		}
		// Y el re-key no tocó nada más que la llave.
		if len(s.Conversaciones) != 1 {
			t.Errorf("la sesión %s perdió su conversación en el re-key", s.ID)
		}
	}
}

// TestReKeySinResolvedorNoRompeElArranque — el Portafolio puede no estar disponible.
// Entonces no se recalibra nada, el arranque sigue, y `--aplicar` existe para después.
func TestReKeySinResolvedorNoRompeElArranque(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)
	reg, inf, err := AbrirRegistro(filepath.Join(dir, "sesiones.json"), legado, "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Recalibradas) != 0 {
		t.Errorf("sin resolvedor no se recalibra nada, se reportaron %d", len(inf.Recalibradas))
	}
	sesiones, err := reg.Load(context.Background())
	if err != nil || len(sesiones) != 5 {
		t.Fatalf("el arranque tiene que seguir con las 5 sesiones: %d (err %v)", len(sesiones), err)
	}
	if _, rerr := reg.Recalibrar(nil, true); rerr == nil {
		t.Error("recalibrar sin resolvedor tiene que fallar ruidoso, no devolver una tabla vacía")
	}
}
