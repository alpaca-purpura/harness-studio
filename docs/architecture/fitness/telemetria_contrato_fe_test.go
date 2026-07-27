package fitness

// telemetria_contrato_fe_test.go ata el tipo del FE al JSON que el Go realmente manda — el
// candado que **faltaba** y que dejó pasar el hueco V-5 de `PARIDAD.md`:
//
//	«El FE tipa `PuntoMejora` con campos que el Go no manda con esos nombres (`contrafactual`
//	como prosa, `patron`, `fix_codigo`, `calculo`, `caja_nombre`, `id`, `solo_s1`,
//	`confianza_detalle`) … ningún test la ata al JSON real.»
//
// La composición del Tramo B estaba escrita contra el tipo del FE, y `tsc` la validaba entera
// contra fixtures que también eran del FE. **Nada comparaba los dos lados del wire**, así que el
// tipo podía prometer campos que nadie producía y todos los gates quedaban verdes.
//
// La dirección del invariante importa: **todo campo que el FE TIPA tiene que existir en el JSON
// del Go.** Al revés no: que el dominio mande más de lo que una superficie usa es normal (el FE no
// necesita `contrafactual_micros` para pintar la frase), y exigir simetría convertiría cada campo
// nuevo del dominio en una edición obligatoria del FE.
//
// Es un SOURCE-SCAN a propósito, igual que `marketplace_shape_test.go`: el invariante no es «qué
// devuelve el endpoint» sino «qué FORMA tienen los dos tipos», y esa forma se puede romper sin
// levantar un daemon.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// contratosDelWire son los pares (struct de Go ↔ interface de TypeScript) del módulo de
// telemetría cuyo desalineo YA costó un hueco. Crece cuando crezca la superficie.
var contratosDelWire = []struct {
	goArchivo string
	goStruct  string
	tsInterfa string
}{
	{"internal/domain/telemetria_deteccion.go", "PuntoDeMejora", "PuntoMejora"},
	{"internal/domain/telemetria_vistas.go", "EstadoDetector", "EstadoDetector"},
	// `SaludTelemetria` entra por D26.3: al firmar el TTL salió `retencion_propuesta` del wire,
	// y sin este candado el FE podía seguir tipándolo —y rotulando «propuesto» sobre un número
	// decidido— sin que nada se pusiera rojo.
	{"internal/domain/telemetria_vistas.go", "SaludTelemetria", "SaludTelemetria"},
	// `ResumenTelemetria` y `RespuestaMejoras` entran con D26.4: los estados 1b, 4, 5 y
	// «runtime no soportado» estaban storiados y eran INALCANZABLES porque el wire no traía
	// el dato. Ahora lo trae, y este candado impide que el FE vuelva a tipar campos que nadie
	// manda — que es exactamente cómo esas cuatro filas quedaron muertas.
	{"internal/domain/telemetria_vistas.go", "ResumenTelemetria", "ResumenTelemetria"},
	{"internal/domain/telemetria_vistas.go", "RespuestaMejoras", "RespuestaMejoras"},
}

const tsTipos = "web/src/entities/telemetria/model/types.ts"

var (
	reTagJSON   = regexp.MustCompile("`json:\"([^\",]+)")
	reCampoTS   = regexp.MustCompile(`(?m)^\s{2}([a-z_][a-z0-9_]*)\??\s*:`)
	reInterfaTS = regexp.MustCompile(`export interface ([A-Za-z0-9_]+)\s*{`)
)

func TestElFeNoTipaCamposQueElWireNoManda(t *testing.T) {
	root := repoRoot()
	tsSrc := leerArchivo(t, filepath.Join(root, tsTipos))

	for _, c := range contratosDelWire {
		t.Run(c.tsInterfa, func(t *testing.T) {
			goSrc := leerArchivo(t, filepath.Join(root, c.goArchivo))
			claves := clavesJSONDe(t, goSrc, c.goStruct)
			if len(claves) == 0 {
				t.Fatalf("no se encontraron tags json en %s#%s (¿se renombró?)", c.goArchivo, c.goStruct)
			}
			campos := camposDeInterface(t, tsSrc, c.tsInterfa)
			if len(campos) == 0 {
				t.Fatalf("no se encontraron campos en %s#%s (¿se renombró?)", tsTipos, c.tsInterfa)
			}
			for _, campo := range campos {
				if !claves[campo] {
					t.Errorf("`%s.%s` no existe en el JSON de `%s`: el FE tipa un campo que "+
						"nadie manda. O lo produce el dominio, o el tipo no lo promete.",
						c.tsInterfa, campo, c.goStruct)
				}
			}
		})
	}
}

// TestLaProsaSeArmaEnElDominio — D25 como forma, no como intención. El FE **presenta**: ninguna
// pieza suya puede rearmar la frase del contrafactual a partir de los micros. Si alguien vuelve a
// hacerlo, el campo que delata el intento es `contrafactual_micros` usado fuera del tipo.
func TestLaProsaSeArmaEnElDominio(t *testing.T) {
	root := repoRoot()
	var culpables []string
	err := filepath.Walk(filepath.Join(root, "web/src"), func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(p, ".ts") && !strings.HasSuffix(p, ".tsx") {
			return nil
		}
		// El tipo SÍ nombra el campo: es su declaración. Lo que no puede es consumirse.
		if strings.HasSuffix(p, filepath.FromSlash(tsTipos)) {
			return nil
		}
		src := string(leerBytes(t, p))
		if strings.Contains(src, ".contrafactual_micros") {
			rel, _ := filepath.Rel(root, p)
			culpables = append(culpables, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("recorrer web/src: %v", err)
	}
	if len(culpables) > 0 {
		t.Errorf("estos archivos leen `contrafactual_micros` para armar prosa en la UI — la frase "+
			"se resuelve en el dominio Go y viaja resuelta (D25 · design.md §1.3): %v", culpables)
	}
}

// clavesJSONDe devuelve el conjunto de claves `json:"…"` del struct pedido. Corta en la primera
// llave de cierre a nivel de columna 0, que es donde termina un struct de nivel superior.
func clavesJSONDe(t *testing.T, src, nombre string) map[string]bool {
	t.Helper()
	inicio := strings.Index(src, "type "+nombre+" struct {")
	if inicio < 0 {
		t.Fatalf("no se encontró `type %s struct`", nombre)
	}
	resto := src[inicio:]
	fin := strings.Index(resto, "\n}")
	if fin < 0 {
		t.Fatalf("no se encontró el cierre de `type %s struct`", nombre)
	}
	claves := map[string]bool{}
	for _, m := range reTagJSON.FindAllStringSubmatch(resto[:fin], -1) {
		if m[1] == "-" {
			continue // insumo de redacción, no viaja al wire (BaseContrafactual)
		}
		claves[m[1]] = true
	}
	return claves
}

// camposDeInterface devuelve los nombres de campo de una interface de TypeScript, cortando en la
// siguiente declaración de interface.
func camposDeInterface(t *testing.T, src, nombre string) []string {
	t.Helper()
	inicio := strings.Index(src, "export interface "+nombre+" {")
	if inicio < 0 {
		t.Fatalf("no se encontró `export interface %s`", nombre)
	}
	resto := src[inicio+len("export interface "+nombre+" {"):]
	if sig := reInterfaTS.FindStringIndex(resto); sig != nil {
		resto = resto[:sig[0]]
	}
	if fin := strings.Index(resto, "\n}"); fin >= 0 {
		resto = resto[:fin]
	}
	var campos []string
	for _, m := range reCampoTS.FindAllStringSubmatch(resto, -1) {
		campos = append(campos, m[1])
	}
	return campos
}

func leerArchivo(t *testing.T, p string) string {
	t.Helper()
	return string(leerBytes(t, p))
}

func leerBytes(t *testing.T, p string) []byte {
	t.Helper()
	b, err := os.ReadFile(p) //nolint:gosec // rutas del propio árbol del repo
	if err != nil {
		t.Fatalf("leer %s: %v", p, err)
	}
	return b
}
