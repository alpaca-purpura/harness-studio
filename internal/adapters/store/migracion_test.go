package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// fixtureReal copia el registro v1 REAL del operador (17 202 B, 5 sesiones) a un directorio
// temporal y devuelve la ruta de la copia. El fixture es el archivo de verdad y no uno
// sintético a propósito: una migración probada contra datos inventados prueba el invento.
func fixtureReal(t *testing.T, dir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "sessions-v1-real.json"))
	if err != nil {
		t.Fatalf("fixture real: %v", err)
	}
	ruta := filepath.Join(dir, "sessions.json")
	if err := os.WriteFile(ruta, b, 0o600); err != nil { //nolint:gosec // ruta bajo t.TempDir().
		t.Fatalf("fixture real: %v", err)
	}
	return ruta
}

// TestSaveEstampaEsquemaActual — todo lo que este binario escribe dice de qué versión es.
// Check `envelope-versionado` del boundary archivo-durable-declara-su-esquema.
func TestSaveEstampaEsquemaActual(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "sesiones.json")
	reg := &Registry{path: ruta, sello: "0.2.17.2607262100"}
	if err := reg.Save(context.Background(), []domain.Session{{ID: "s1"}}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	var s sobre
	if uerr := json.Unmarshal(b, &s); uerr != nil {
		t.Fatalf("lo escrito no es un sobre: %v", uerr)
	}
	if s.SchemaVersion != EsquemaActual {
		t.Errorf("schema_version = %d, quiero %d", s.SchemaVersion, EsquemaActual)
	}
	if s.EscritoPor != "0.2.17.2607262100" {
		t.Errorf("escrito_por = %q — el sello del binario tiene que viajar", s.EscritoPor)
	}
	if s.EscritoEn == "" {
		t.Error("escrito_en vacío: un archivo durable dice cuándo se escribió")
	}
	// Y se relee sin ayuda: el sobre no es sólo metadata decorativa.
	vuelta, err := reg.Load(context.Background())
	if err != nil || len(vuelta) != 1 || vuelta[0].ID != "s1" {
		t.Errorf("releer lo escrito = %+v (err %v)", vuelta, err)
	}
}

// TestDetectarVersionArrayDesnudoEsV1 — el array desnudo es la forma vieja, y se reconoce
// por la forma, no por adivinanza.
func TestDetectarVersionArrayDesnudoEsV1(t *testing.T) {
	casos := []struct {
		nombre string
		crudo  string
	}{
		{"array vacío", `[]`},
		{"array con sesiones", `[{"id":"s1"}]`},
		{"array con blancos adelante", "  \n\t[{\"id\":\"s1\"}]"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			v, err := detectarVersion([]byte(c.crudo))
			if err != nil || v != 1 {
				t.Errorf("detectarVersion = %d (err %v), quiero 1", v, err)
			}
		})
	}
}

// TestDetectarVersionSobre — el sobre manda su propia versión, y un objeto SIN versión es
// un error: no se asume v1 tácito, que es exactamente la adivinanza que el envelope existe
// para eliminar.
func TestDetectarVersionSobre(t *testing.T) {
	casos := []struct {
		nombre  string
		crudo   string
		quiero  int
		esError bool
	}{
		{"v2 declarada", `{"schema_version":2,"sesiones":[]}`, 2, false},
		{"v99 futura", `{"schema_version":99,"sesiones":[]}`, 99, false},
		{"objeto sin versión", `{"sesiones":[]}`, 0, true},
		{"versión cero", `{"schema_version":0,"sesiones":[]}`, 0, true},
		{"no es JSON", `hola`, 0, true},
		{"vacío", ``, 0, true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			v, err := detectarVersion([]byte(c.crudo))
			if c.esError {
				if err == nil {
					t.Errorf("detectarVersion(%q) = %d sin error — tiene que fallar ruidoso", c.crudo, v)
				}
				return
			}
			if err != nil || v != c.quiero {
				t.Errorf("detectarVersion = %d (err %v), quiero %d", v, err, c.quiero)
			}
		})
	}
}

// TestMigracionRespaldaAntesDeEscribir — el respaldo existe y es BYTE A BYTE el original,
// y el original sigue intacto. Check `migracion-forward-only-con-respaldo`.
func TestMigracionRespaldaAntesDeEscribir(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)
	original, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}

	_, inf, err := AbrirRegistro(filepath.Join(dir, "sesiones.json"), legado, "0.2.17.2607262100", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !inf.Migro || inf.DesdeVersion != 1 {
		t.Fatalf("informe = %+v, quiero migró desde la 1", inf)
	}
	if inf.RespaldoEn == "" {
		t.Fatal("no hay respaldo: migrar sin copia previa es lo que el boundary prohíbe")
	}
	respaldo, err := os.ReadFile(inf.RespaldoEn)
	if err != nil {
		t.Fatalf("el respaldo no se puede leer: %v", err)
	}
	if string(respaldo) != string(original) {
		t.Error("el respaldo no es byte a byte el original")
	}
	// El legado NO se toca, NO se borra, NO se renombra: volver atrás es gratis.
	ahora, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatalf("el registro legado desapareció: %v", err)
	}
	if string(ahora) != string(original) {
		t.Error("el registro legado cambió: la migración no puede tocarlo")
	}
}

// TestMigracionEsIdempotente — el segundo arranque no vuelve a migrar y deja el archivo
// IGUAL byte a byte. E-43: la idempotencia se ve en el disco, no se promete.
func TestMigracionEsIdempotente(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)
	v2 := filepath.Join(dir, "sesiones.json")

	if _, inf, err := AbrirRegistro(v2, legado, "sello-1", nil); err != nil || !inf.Migro {
		t.Fatalf("primer arranque: inf=%+v err=%v", inf, err)
	}
	primera, err := os.ReadFile(v2) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}

	_, inf, err := AbrirRegistro(v2, legado, "sello-2", nil)
	if err != nil {
		t.Fatal(err)
	}
	if inf.Migro {
		t.Error("el segundo arranque volvió a migrar: el sobre está para que eso no pase")
	}
	if len(inf.Reparaciones) != 0 {
		t.Errorf("el segundo arranque reparó algo: %v — la primera migración dejó el registro sano", inf.Reparaciones)
	}
	segunda, err := os.ReadFile(v2) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	if string(primera) != string(segunda) {
		t.Error("el archivo cambió en un arranque sin novedades")
	}
}

// TestDeV1aV2ConservaTodoElFixtureReal — el corazón del ticket. Se compara CAMPO POR CAMPO
// contra el archivo real, no con un len(): un campo que el migrador olvide llevar se
// pierde en silencio, y contar sesiones no lo detecta.
func TestDeV1aV2ConservaTodoElFixtureReal(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)
	crudo, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}
	var viejas []sesionV1
	if uerr := json.Unmarshal(crudo, &viejas); uerr != nil {
		t.Fatal(uerr)
	}
	if len(viejas) != 5 {
		t.Fatalf("el fixture real dejó de tener 5 sesiones (tiene %d) — revisá qué se copió", len(viejas))
	}

	reg, inf, err := AbrirRegistro(filepath.Join(dir, "sesiones.json"), legado, "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Reparaciones) != 0 {
		t.Errorf("la migración dejó sesiones que hubo que reparar: %v", inf.Reparaciones)
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
			// Lo que se queda en la sesión.
			if n.ID != v.ID || n.Frente != v.Frente || n.Arnes != v.Arnes || n.Cwd != v.Cwd {
				t.Errorf("identidad de la sesión cambió: %+v", n)
			}
			if n.Empresa != v.Empresa || n.Puesto != v.Puesto || n.Salud != v.Salud ||
				n.View != v.View || n.Parked != v.Parked || n.Reparacion != v.Reparacion ||
				n.CerradaEn != v.CerradaEn {
				t.Errorf("metadata del frente cambió: %+v", n)
			}
			// Exactamente una conversación, activa.
			if verr := domain.VerificarUnaActiva(n); verr != nil {
				t.Fatalf("la migrada viola la invariante: %v", verr)
			}
			if len(n.Conversaciones) != 1 {
				t.Fatalf("conversaciones = %d, quiero 1", len(n.Conversaciones))
			}
			c := n.Conversaciones[0]

			// Lo que baja, campo por campo.
			if c.ClaudeSessionID != v.ClaudeSessionID || c.Model != v.Model {
				t.Errorf("id de Claude Code / modelo perdidos: %q %q", c.ClaudeSessionID, c.Model)
			}
			if c.CtxPct != v.CtxPct || c.RotacionPendiente != v.RotacionPendiente || c.Checkpoint != v.Checkpoint {
				t.Errorf("estado de contexto perdido: ctx=%d rot=%v ck=%q", c.CtxPct, c.RotacionPendiente, c.Checkpoint)
			}
			if len(c.CtxHist) != len(v.CtxHist) || len(c.CadenaCC) != len(v.CadenaCC) {
				t.Errorf("histórico/cadena perdidos: hist=%d cadena=%d", len(c.CtxHist), len(c.CadenaCC))
			}
			if len(c.Conv) != len(v.Conv) {
				t.Fatalf("turnos = %d, quiero %d", len(c.Conv), len(v.Conv))
			}
			for j := range v.Conv {
				if c.Conv[j] != v.Conv[j] {
					t.Fatalf("turno %d cambió: %+v vs %+v", j, c.Conv[j], v.Conv[j])
				}
			}

			// Lo que NACE, y lo que deliberadamente no.
			if c.ID == "" || c.CreadaEn == "" || !c.Activa {
				t.Errorf("la conversación migrada nació incompleta: %+v", c)
			}
			if c.UltimaInteraccion != "" {
				t.Errorf("ultima_interaccion = %q — esa fecha NO existe en el archivo viejo y no se inventa", c.UltimaInteraccion)
			}
			if c.TituloEditado {
				t.Error("una conversación migrada no la tituló nadie a mano")
			}
			if c.Conv == nil {
				t.Error("conv es null: «leí y no había turnos» y «no cargué» tienen que distinguirse")
			}
			if len(v.Conv) == 0 && c.Titulo != domain.TituloConversacionNueva {
				t.Errorf("sin turnos el título tiene que ser el default, es %q", c.Titulo)
			}
			if len(v.Conv) > 0 && c.Titulo == "" {
				t.Error("con turnos la conversación tiene que quedar bautizada")
			}
		})
	}

	// El archivo del operador no se tocó.
	despues, err := os.ReadFile(legado) //nolint:gosec // ruta del propio test.
	if err != nil || string(despues) != string(crudo) {
		t.Error("el registro v1 cambió: la migración escribe al lado, no encima")
	}
}

// TestCrashAMitadDejaElViejoEntero (E-47) — un Save que falla no deja medio archivo: o el
// viejo entero o el nuevo entero. Se fuerza con un directorio sin permiso de escritura.
func TestCrashAMitadDejaElViejoEntero(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("como root el permiso de escritura no frena nada")
	}
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")
	reg := &Registry{path: v2, sello: "sello"}
	if err := reg.Save(context.Background(), []domain.Session{{ID: "s-viejo"}}); err != nil {
		t.Fatal(err)
	}
	antes, err := os.ReadFile(v2) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatal(err)
	}

	// 0o500 es el punto del test: un directorio donde el Save NO puede escribir.
	if cerr := os.Chmod(dir, 0o500); cerr != nil { //nolint:gosec // permisos de directorio, no de archivo.
		t.Fatal(cerr)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) }) //nolint:gosec // ídem: es un directorio temporal.

	if serr := reg.Save(context.Background(), []domain.Session{{ID: "s-nuevo"}}); serr == nil {
		t.Fatal("guardar en un directorio de sólo lectura tiene que fallar, no fallar en silencio")
	}
	despues, err := os.ReadFile(v2) //nolint:gosec // ruta del propio test.
	if err != nil {
		t.Fatalf("el registro viejo desapareció: %v", err)
	}
	if string(despues) != string(antes) {
		t.Error("el registro quedó a medias: el temp+rename existe justo para que eso no pase")
	}
	vuelta, err := reg.Load(context.Background())
	if err != nil || len(vuelta) != 1 || vuelta[0].ID != "s-viejo" {
		t.Errorf("releer tras el fallo = %+v (err %v), quiero el viejo entero", vuelta, err)
	}
}

// TestPrimerArranqueNoInventaArchivo — sin registro previo no se migra nada y no se escribe
// nada. Modo A de los seis de fallo.
func TestPrimerArranqueNoInventaArchivo(t *testing.T) {
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")
	reg, inf, err := AbrirRegistro(v2, filepath.Join(dir, "sessions.json"), "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if inf.Hubo() {
		t.Errorf("un primer arranque no tiene nada que contar: %+v", inf)
	}
	if _, serr := os.Stat(v2); !os.IsNotExist(serr) {
		t.Error("se escribió un registro que nadie pidió")
	}
	sesiones, err := reg.Load(context.Background())
	if err != nil || len(sesiones) != 0 {
		t.Errorf("registro inicial = %+v (err %v), quiero vacío", sesiones, err)
	}
}

// TestNormalizarAlCargarSeReportaYSeArregla — un registro v2 con la invariante rota se
// repara al abrir y el arreglo VIAJA al informe para que el arranque lo loguee.
func TestNormalizarAlCargarSeReportaYSeArregla(t *testing.T) {
	dir := t.TempDir()
	v2 := filepath.Join(dir, "sesiones.json")
	roto := `{"schema_version":2,"sesiones":[
	  {"id":"s-sin-conversaciones","arnes":"vitalia","conversaciones":[]},
	  {"id":"s-dos-activas","arnes":"vitalia","conversaciones":[
	     {"id":"cv1","titulo":"a","activa":true,"creada_en":"2026-07-01T00:00:00Z","conv":[]},
	     {"id":"cv2","titulo":"b","activa":true,"creada_en":"2026-07-02T00:00:00Z","conv":[]}]}]}`
	if err := os.WriteFile(v2, []byte(roto), 0o600); err != nil {
		t.Fatal(err)
	}
	reg, inf, err := AbrirRegistro(v2, "", "sello", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(inf.Reparaciones) != 2 {
		t.Errorf("reparaciones = %v, quiero una por cada sesión rota", inf.Reparaciones)
	}
	sesiones, err := reg.Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range sesiones {
		if verr := domain.VerificarUnaActiva(s); verr != nil {
			t.Errorf("quedó rota tras normalizar: %v", verr)
		}
	}
	// La que tenía dos activas conserva la más reciente por fecha de creación.
	if c, ok := sesiones[1].Activa(); !ok || c.ID != "cv2" {
		t.Errorf("desempate equivocado: quedó activa %+v", c)
	}
}

// TestRespaldoNoSePisaEntreMigraciones — el sello va en el nombre justo para eso.
func TestRespaldoNoSePisaEntreMigraciones(t *testing.T) {
	dir := t.TempDir()
	legado := fixtureReal(t, dir)
	uno, err := respaldar(legado, 1, "sello-uno")
	if err != nil {
		t.Fatal(err)
	}
	dos, err := respaldar(legado, 1, "sello-dos")
	if err != nil {
		t.Fatal(err)
	}
	if uno == dos {
		t.Fatal("dos respaldos con el mismo nombre: el segundo pisó al primero")
	}
	for _, r := range []string{uno, dos} {
		if _, err := os.Stat(r); err != nil {
			t.Errorf("falta el respaldo %s: %v", r, err)
		}
	}
}

// TestMigradorNoLeeElReloj — la misma entrada y el mismo instante dan la misma salida,
// salvo el id de la conversación, que es aleatorio por diseño.
func TestMigradorNoLeeElReloj(t *testing.T) {
	entrada := json.RawMessage(`[{"id":"s1","arnes":"vitalia","conv":[{"rol":"user","text":"arreglá el hook"}]}]`)
	ahora := time.Date(2026, 7, 26, 21, 0, 0, 0, time.UTC)

	var salidas []domain.Session
	for range 2 {
		out, err := deV1aV2(entrada, ahora)
		if err != nil {
			t.Fatal(err)
		}
		var s []domain.Session
		if err := json.Unmarshal(out, &s); err != nil {
			t.Fatal(err)
		}
		salidas = append(salidas, s[0])
	}
	a, b := salidas[0].Conversaciones[0], salidas[1].Conversaciones[0]
	if a.CreadaEn != "2026-07-26T21:00:00Z" || a.CreadaEn != b.CreadaEn {
		t.Errorf("creada_en = %q / %q — el migrador tiene que usar el instante que le dan", a.CreadaEn, b.CreadaEn)
	}
	if a.Titulo != "arreglá el hook" {
		t.Errorf("título = %q, quiero el primer mensaje del usuario", a.Titulo)
	}
	if a.ID == b.ID {
		t.Error("dos corridas dieron el mismo id de conversación: el generador no es aleatorio")
	}
}
