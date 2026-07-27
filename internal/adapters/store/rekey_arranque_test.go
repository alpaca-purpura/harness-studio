package store

// CV-D18 (FIRMADA 2026-07-27) — el re-key se CABLEA al arranque del daemon. Cierra A-10.
//
// Antes: `cmd/arnesia/main.go` pasaba `nil` como `ClaveCalificada`, así que `reKey` no corría
// nunca en el arranque y CV-D16 —firmada— no estaba construida. El comando manual que la
// reemplazó no se lo sugiere nadie al operador, así que 3 de sus 5 sesiones quedaban
// invisibles sin aviso (N-23). Una decisión firmada que no corre no está construida.
//
// Estos tests cubren las cuatro condiciones que la decisión declara NO opcionales:
//
//	1 · no hay llaves a medias  ⇒ no toca nada y NO respalda de gusto
//	2 · hay N a medias          ⇒ respalda, recalibra y el informe dice el número
//	3 · falla el respaldo       ⇒ NO recalibra (primero la red, después el cambio)
//	4 · idempotencia            ⇒ arrancar dos veces no duplica respaldos ni vuelve a mover
//
// Y una quinta que el E2E ya conocía y la decisión manda no callar: `sin-candidata` se
// cuenta aparte, nunca se confunde con «no la miré».

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// resolverFalso imita al resolvedor real: `~` ⇒ ya calificada; un id conocido ⇒ se resuelve;
// cualquier otro ⇒ sin-candidata (su arnés no está en el Portafolio).
func resolverFalso(conocidos map[string]string) ClaveCalificada {
	return func(idPelado, _ string) (string, bool, string) {
		if strings.Contains(idPelado, "~") {
			return "", false, "ya-calificada"
		}
		if c, ok := conocidos[idPelado]; ok {
			return c, true, "resuelta-por-id"
		}
		return "", false, "sin-candidata"
	}
}

// registroCon escribe un registro con esas llaves de arnés y devuelve su ruta + el Registry.
func registroCon(t *testing.T, arneses ...string) (string, *Registry) {
	t.Helper()
	dir := t.TempDir()
	ruta := filepath.Join(dir, "sessions.json")
	reg := &Registry{path: ruta, sello: "0.2.25.2607270900"}
	ses := make([]domain.Session, 0, len(arneses))
	for i, a := range arneses {
		ses = append(ses, domain.Session{
			ID:             "s" + string(rune('a'+i)),
			Arnes:          a,
			Conversaciones: []domain.Conversacion{{ID: "cv", Activa: true, Titulo: "t"}},
		})
	}
	if err := reg.Save(context.Background(), ses); err != nil {
		t.Fatal(err)
	}
	return ruta, reg
}

func baks(t *testing.T, ruta string) []string {
	t.Helper()
	m, err := filepath.Glob(ruta + ".bak-*")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// 1 · Sin llaves a medias no se toca nada — y sobre todo NO se respalda de gusto. Un
// respaldo por arranque llenaría el disco del operador de copias idénticas.
func TestRecalibrarAlArrancarSinLlavesAMediasNoTocaNiRespalda(t *testing.T) {
	ruta, reg := registroCon(t, "sin-home~vitalia~vitalia", "sin-home~arnesia~arnesia")
	antes, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}

	inf, err := reg.RecalibrarAlArrancar(resolverFalso(nil), "sello-1")
	if err != nil {
		t.Fatalf("recalibrar: %v", err)
	}
	if inf.Recalibradas != 0 {
		t.Errorf("recalibró %d con todas las llaves ya calificadas", inf.Recalibradas)
	}
	if inf.YaCalificadas != 2 {
		t.Errorf("ya-calificadas = %d, quiero 2", inf.YaCalificadas)
	}
	if inf.RespaldoEn != "" {
		t.Errorf("respaldó sin tener nada que mover: %q", inf.RespaldoEn)
	}
	if b := baks(t, ruta); len(b) != 0 {
		t.Errorf("dejó respaldos de gusto: %v", b)
	}
	despues, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}
	if string(antes) != string(despues) {
		t.Error("el archivo cambió sin que hubiera nada que recalibrar")
	}
	if inf.Hubo() {
		t.Error("un arranque sin novedad no tiene que loguear")
	}
}

// 2 · Con N a medias: respalda ANTES, recalibra, y el informe dice el número. Además el
// respaldo tiene el estado ANTERIOR — que es lo único que lo hace servir para revertir.
func TestRecalibrarAlArrancarRespaldaRecalibraYCuenta(t *testing.T) {
	ruta, reg := registroCon(t, "vitalia", "vitalia", "sin-home~arnesia~arnesia", "fantasma")
	original, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}

	inf, err := reg.RecalibrarAlArrancar(resolverFalso(map[string]string{
		"vitalia": "sin-home~vitalia~vitalia",
	}), "sello-2")
	if err != nil {
		t.Fatalf("recalibrar: %v", err)
	}

	if inf.Recalibradas != 2 {
		t.Errorf("recalibradas = %d, quiero 2", inf.Recalibradas)
	}
	// `fantasma` no está en el Portafolio: se cuenta APARTE y no se calla (CV-D18).
	if inf.SinCandidata != 1 {
		t.Errorf("sin-candidata = %d, quiero 1 — el caso que el E2E ya conoce", inf.SinCandidata)
	}
	if inf.YaCalificadas != 1 {
		t.Errorf("ya-calificadas = %d, quiero 1", inf.YaCalificadas)
	}
	if len(inf.Filas) != 4 {
		t.Errorf("filas = %d, quiero UNA por sesión (4)", len(inf.Filas))
	}
	if !inf.Hubo() {
		t.Error("con recalibradas y sin-candidata, el arranque TIENE que loguear")
	}

	// El respaldo existe, es el que la decisión nombra, y tiene el estado ANTERIOR.
	if inf.RespaldoEn != ruta+".bak-sello-2" {
		t.Errorf("respaldo en %q, quiero %q (el nombre que nombra CV-D16/CV-D18)", inf.RespaldoEn, ruta+".bak-sello-2")
	}
	b, err := os.ReadFile(inf.RespaldoEn)
	if err != nil {
		t.Fatalf("sin respaldo previo: %v", err)
	}
	if string(b) != string(original) {
		t.Error("el respaldo NO tiene el estado anterior: no serviría para revertir")
	}

	// Y el archivo vivo quedó recalibrado.
	vivas, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	llaves := map[string]int{}
	for _, s := range vivas {
		llaves[s.Arnes]++
	}
	if llaves["sin-home~vitalia~vitalia"] != 2 {
		t.Errorf("quedaron %d sesiones con la llave calificada de vitalia, quiero 2", llaves["sin-home~vitalia~vitalia"])
	}
	if llaves["fantasma"] != 1 {
		t.Error("la sesión sin candidata tiene que quedar TAL CUAL: jamás se le adivina una llave")
	}
	if llaves["vitalia"] != 0 {
		t.Errorf("quedaron %d llaves peladas de vitalia sin recalibrar", llaves["vitalia"])
	}
}

// 3 · Si el respaldo falla, NO se recalibra. Primero la red, después el cambio — el orden
// ES la decisión, no un detalle de implementación.
func TestRecalibrarAlArrancarNoRecalibraSiFallaElRespaldo(t *testing.T) {
	ruta, reg := registroCon(t, "vitalia", "vitalia")
	original, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}

	// El respaldo se hace imposible: en el destino ya hay un DIRECTORIO con ese nombre, así
	// que el WriteFile falla. (Un `.bak-` que ya es archivo NO sirve para esto: ese caso es
	// el de la idempotencia y no se pisa a propósito.)
	if merr := os.Mkdir(ruta+".bak-sello-3", 0o750); merr != nil {
		t.Fatal(merr)
	}

	inf, err := reg.RecalibrarAlArrancar(resolverFalso(map[string]string{
		"vitalia": "sin-home~vitalia~vitalia",
	}), "sello-3")
	if err == nil {
		t.Fatal("el respaldo falló y la recalibración siguió igual: eso pisa el registro del operador sin red")
	}
	if inf.RespaldoEn != "" {
		t.Errorf("informa un respaldo que no se pudo hacer: %q", inf.RespaldoEn)
	}

	// Lo que importa: el archivo NO se tocó.
	despues, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}
	if string(despues) != string(original) {
		t.Error("se recalibró sin respaldo: el registro del operador quedó sin red")
	}
	vivas, lerr := reg.Load(context.Background())
	if lerr != nil {
		t.Fatal(lerr)
	}
	for _, s := range vivas {
		if s.Arnes != "vitalia" {
			t.Errorf("la sesión %q se movió a %q pese a que el respaldo falló", s.ID, s.Arnes)
		}
	}
}

// 4 · Idempotencia: arrancar dos veces no duplica respaldos ni vuelve a mover nada.
func TestRecalibrarAlArrancarEsIdempotente(t *testing.T) {
	ruta, reg := registroCon(t, "vitalia", "arnesia")
	clave := resolverFalso(map[string]string{
		"vitalia": "sin-home~vitalia~vitalia",
		"arnesia": "sin-home~arnesia~arnesia",
	})

	inf1, err := reg.RecalibrarAlArrancar(clave, "sello-4")
	if err != nil {
		t.Fatalf("1er arranque: %v", err)
	}
	if inf1.Recalibradas != 2 {
		t.Fatalf("1er arranque recalibró %d, quiero 2", inf1.Recalibradas)
	}
	trasUno, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}

	// Segundo arranque, mismo binario (mismo sello) — y también uno con sello distinto,
	// que es el caso real de una actualización.
	inf2, err := reg.RecalibrarAlArrancar(clave, "sello-4")
	if err != nil {
		t.Fatalf("2do arranque: %v", err)
	}
	inf3, err := reg.RecalibrarAlArrancar(clave, "sello-5")
	if err != nil {
		t.Fatalf("3er arranque: %v", err)
	}
	if inf2.Recalibradas != 0 || inf3.Recalibradas != 0 {
		t.Errorf("volvió a mover llaves ya calificadas: %d y %d", inf2.Recalibradas, inf3.Recalibradas)
	}
	if inf2.YaCalificadas != 2 || inf3.YaCalificadas != 2 {
		t.Errorf("ya-calificadas = %d y %d, quiero 2 en los dos", inf2.YaCalificadas, inf3.YaCalificadas)
	}
	if b := baks(t, ruta); len(b) != 1 {
		t.Errorf("respaldos = %d, quiero 1: arrancar de nuevo no puede duplicarlos (%v)", len(b), b)
	}
	trasTres, err := os.ReadFile(ruta) //nolint:gosec // t.TempDir().
	if err != nil {
		t.Fatal(err)
	}
	if string(trasUno) != string(trasTres) {
		t.Error("el archivo cambió en un arranque que no tenía nada que hacer")
	}
}

// 5 · Sin resolvedor no se inventa nada: es un error, no un no-op silencioso. Es
// exactamente el `nil` que main.go pasaba y que dejó a CV-D16 sin construir.
func TestRecalibrarAlArrancarSinResolvedorEsError(t *testing.T) {
	_, reg := registroCon(t, "vitalia")
	if _, err := reg.RecalibrarAlArrancar(nil, "s"); err == nil {
		t.Error("sin resolvedor tiene que fallar ruidoso: pasar nil es lo que dejó A-10 vivo")
	}
}

// 6 · El respaldo se puede REVERTIR — «reversible» tiene que ser observable, no una promesa.
func TestElRespaldoDelArranqueSirveParaRevertir(t *testing.T) {
	ruta, reg := registroCon(t, "vitalia", "arnesia")
	inf, err := reg.RecalibrarAlArrancar(resolverFalso(map[string]string{
		"vitalia": "sin-home~vitalia~vitalia",
		"arnesia": "sin-home~arnesia~arnesia",
	}), "sello-6")
	if err != nil {
		t.Fatal(err)
	}
	previas, err := LeerRespaldo(inf.RespaldoEn)
	if err != nil {
		t.Fatalf("el respaldo del arranque no se puede leer: %v", err)
	}
	filas, err := reg.RevertirLlaves(previas)
	if err != nil {
		t.Fatalf("revertir: %v", err)
	}
	revertidas := 0
	for _, f := range filas {
		if f.Motivo == "revertida" {
			revertidas++
		}
	}
	if revertidas != 2 {
		t.Errorf("revertidas = %d, quiero 2", revertidas)
	}
	vivas, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range vivas {
		if strings.Contains(s.Arnes, "~") {
			t.Errorf("la sesión %q quedó en %q: la reversión no la devolvió a su llave pelada", s.ID, s.Arnes)
		}
	}
	_ = ruta
}
