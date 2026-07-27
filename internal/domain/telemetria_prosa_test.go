package domain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// telemetria_prosa_test.go cementa D25: **la prosa del punto de mejora se arma en el dominio**, y
// el formato del dinero es UNO SOLO de los dos lados del wire.

// TestElFormatoDeDineroEsUnoSolo — el gate de RF-281 que faltaba. `MontoMicros` (Go, prosa) y
// `usd()` (FE, `CifraUsd`) tienen que producir **exactamente lo mismo**, y la única forma de
// asertarlo sin ejecutar TypeScript desde Go es que los dos asserten **la misma tabla**.
//
// El archivo es golden: este test NO lo escribe. Si el formateador cambia, el test se pone rojo y
// alguien decide — que es lo contrario de un golden que se auto-actualiza y nunca falla.
func TestElFormatoDeDineroEsUnoSolo(t *testing.T) {
	ruta := filepath.Join(repoRootDominio(t),
		"web/src/entities/telemetria/testing/montos.golden.json")
	b, err := os.ReadFile(ruta) //nolint:gosec // ruta del propio árbol del repo
	if err != nil {
		t.Fatalf("leer la tabla golden: %v", err)
	}
	var casos []struct {
		Micros int64  `json:"micros"`
		Texto  string `json:"texto"`
		Porque string `json:"porque"`
	}
	if err := json.Unmarshal(b, &casos); err != nil {
		t.Fatalf("parsear la tabla golden: %v", err)
	}
	if len(casos) < 6 {
		t.Fatalf("la tabla golden tiene %d casos: no cubre los bordes que RF-281 nombra", len(casos))
	}
	for _, c := range casos {
		if got := MontoMicros(c.Micros); got != c.Texto {
			t.Errorf("MontoMicros(%d) = %q, la tabla dice %q — %s", c.Micros, got, c.Texto, c.Porque)
		}
	}
}

// TestUnMontoQueExisteNuncaSaleCero — la mentira barata que este módulo existe para no decir. Un
// monto de 4 000 micros es USD 0,004 y **no** «0,00»: un cero se lee como «no costó nada».
func TestUnMontoQueExisteNuncaSaleCero(t *testing.T) {
	for _, micros := range []int64{1, 10, 100, 1_000, 4_000, 9_999} {
		got := MontoMicros(micros)
		if got == "0,00" {
			t.Errorf("MontoMicros(%d) = %q — un monto que existe se muestra con la precisión que "+
				"haga falta, jamás redondeado a cero", micros, got)
		}
	}
	// Control positivo: el cero REAL sí se escribe como cero. Sin esto, un formateador que
	// nunca devuelve «0,00» pasaría el assert de arriba sin ser correcto.
	if got := MontoMicros(0); got != "0,00" {
		t.Errorf("MontoMicros(0) = %q, se esperaba %q: el cero real es un dato", got, "0,00")
	}
}

// TestElContrafactualDeclaraSuUnidad — RF-249. La iteración 1 del mockup decía «USD 0,53 por
// corrida» sobre una caja de 14 corridas y se leyó de dos maneras. La frase nombra **las dos**
// unidades y su denominador, así que no queda ambigüedad que interpretar.
func TestElContrafactualDeclaraSuUnidad(t *testing.T) {
	p := PuntoDeMejora{
		Detector:            DetB1,
		BaseContrafactual:   "Con el vencimiento del cache en 1 h",
		ContrafactualMicros: 310_000,
		DiferenciaMicros:    530_000,
		CorridasUsadas:      14,
	}.Redactar()

	for _, frag := range []string{
		"Con el vencimiento del cache en 1 h", // la cláusula del detector
		"las mismas 14 corridas",              // el denominador, explícito
		"USD 0,31",                            // el mundo alternativo
		"USD 0,53 en la ventana",              // la diferencia, con SU unidad
		"por corrida, sobre 14",               // y la otra unidad, con su base
	} {
		if !strings.Contains(p.Contrafactual, frag) {
			t.Errorf("falta %q en el contrafactual: %q", frag, p.Contrafactual)
		}
	}
}

// TestSinAhorroPositivoNoHayContrafactual — la regla A4 donde de verdad muerde. Un «mundo
// alternativo» que sale MÁS caro no es una recomendación: es un reproche. El punto se queda sin
// prosa a propósito, y `Mejoras()` lo publica como detector «sin fix propuesto».
//
// Hasta D26.1 el caso era el de B1, cuyo contrafactual daba negativo por construcción. Ese
// detector se arregló, así que hoy **ningún detector real produce este caso** — y el invariante
// se prueba igual, porque es del REDACTOR, no de un detector.
func TestSinAhorroPositivoNoHayContrafactual(t *testing.T) {
	for _, diferencia := range []int64{-7_500, 0} {
		p := PuntoDeMejora{
			Detector:            DetB1,
			BaseContrafactual:   "Con el vencimiento del cache en 1 h",
			ContrafactualMicros: 20_000,
			DiferenciaMicros:    diferencia,
			CorridasUsadas:      3,
		}.Redactar()
		if p.Contrafactual != "" {
			t.Errorf("diferencia %d: se redactó %q — sin ahorro positivo no hay recomendación",
				diferencia, p.Contrafactual)
		}
	}
	// Control positivo en la misma corrida: con ahorro, la frase sí se arma. Sin esto, un
	// redactor que devuelve vacío siempre pasaría el assert de ausencia.
	p := PuntoDeMejora{
		Detector: DetB1, BaseContrafactual: "Con el vencimiento del cache en 1 h",
		ContrafactualMicros: 10_000, DiferenciaMicros: 5_000, CorridasUsadas: 3,
	}.Redactar()
	if p.Contrafactual == "" {
		t.Error("con ahorro positivo la frase tiene que armarse: el redactor está mudo siempre")
	}
}

// TestElIdDelPuntoEsEstable — la superficie descarta UN punto, no «el tercero de la lista»: el
// orden es por ahorro y cambia entre consultas. Dos evaluaciones del mismo hallazgo dan el mismo
// id, y dos cajas distintas del mismo detector NO lo comparten.
func TestElIdDelPuntoEsEstable(t *testing.T) {
	nuevo := func(caja string) PuntoDeMejora {
		return PuntoDeMejora{Detector: DetB4, CajaID: caja}.Redactar()
	}
	if a, b := nuevo("paso-3"), nuevo("paso-3"); a.ID != b.ID {
		t.Errorf("el mismo hallazgo dio dos ids: %q vs %q", a.ID, b.ID)
	}
	if a, b := nuevo("paso-3"), nuevo("paso-4"); a.ID == b.ID {
		t.Errorf("dos cajas comparten id %q: descartar una descartaría la otra", a.ID)
	}
	if sin := nuevo(""); sin.ID != string(DetB4) {
		t.Errorf("un punto sin caja tiene id %q, se esperaba %q", sin.ID, DetB4)
	}
}

// TestRedactarNoPisaLoQueElDetectorEscribio — un detector puede redactar su propio contrafactual
// si el genérico no le sirve. El redactor completa, no gobierna.
func TestRedactarNoPisaLoQueElDetectorEscribio(t *testing.T) {
	const propio = "Con el arreglo, esto no se pagaba: USD 0,00 → diferencia USD 0,10 en la ventana."
	p := PuntoDeMejora{
		Detector: DetP1, CajaID: "paso-3", ID: "propio-1", Contrafactual: propio,
		BaseContrafactual: "ignorado", DiferenciaMicros: 100_000, CorridasUsadas: 2,
	}.Redactar()
	if p.Contrafactual != propio {
		t.Errorf("el redactor pisó la frase del detector: %q", p.Contrafactual)
	}
	if p.ID != "propio-1" {
		t.Errorf("el redactor pisó el id del detector: %q", p.ID)
	}
}

// repoRootDominio sube hasta encontrar el `go.mod`. El dominio no conoce rutas del repo en
// producción; esto vive solo en el test.
func repoRootDominio(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for range 8 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("no se encontró la raíz del repo (go.mod)")
	return ""
}
