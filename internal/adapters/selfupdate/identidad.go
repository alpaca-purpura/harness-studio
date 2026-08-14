package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// identidad.go — QUÉ build es este binario (paquete 2026-07-26-identidad-de-build, RF-231).
//
// **El agujero que tapa.** Hasta acá `GET /api/version` reportaba solo la huella del commit
// (`vcs.revision[:7]`). Eso NO alcanza para la pregunta que el operador se hace de verdad —
// «¿estoy corriendo lo último que compilé?»— por dos motivos medidos el 2026-07-26:
//
//  1. **El semver no estaba en el binario.** `0.2.21` vive en `Cargo.toml`/`package.json`; el
//     daemon no lo conocía, así que Ajustes no podía decir qué versión corría.
//  2. **La huella no distingue dos compilaciones del mismo commit.** Ese mismo día se
//     bundleó v0.2.21 dos veces desde el mismo árbol: dos binarios distintos, huella idéntica.
//
// Las tres variables las inyecta `scripts/bundle.sh` por `-ldflags -X`. Un `go build` pelado
// (CI, `go run`) las deja vacías y todo reporta **`dev`** — degradación honesta: mejor decir
// «no sé qué build soy» que inventar un número.

// Version es el semver del release (0.2.21), inyectado desde Cargo.toml —la fuente de verdad
// que ya usa el Makefile— al compilar. Vacío = build de desarrollo.
var Version = ""

// Build es el sello de compilación en formato `AAMMDDHHMM` (2607260225 = 26/07/2026 02:25).
//
// **Por qué un timestamp y no un contador:** cambia SIEMPRE (dos builds del mismo commit dan
// números distintos, que es exactamente el caso que hacía falta distinguir), es monótono —más
// grande = más nuevo, sin excepciones que el operador tenga que recordar— y no necesita
// guardar estado en el repo, así que no hay archivo que commitear ni contador que se
// desincronice al compilar en otra máquina.
//
// Se sella en hora LOCAL de la máquina de build, no UTC: el número se compara contra «cuándo
// compilé», y esa referencia es el reloj que el operador tiene delante. El costo es que un
// cambio de huso o el fin del horario de verano puede repetir una hora una vez al año — a
// cambio de que el número no esté corrido cinco horas de lo que dice el reloj.
var Build = ""

// Compilado es la misma marca en formato legible (`2026-07-26 02:25`). Se inyecta aparte en
// vez de derivarse de Build para que el formato humano no dependa de parsear un número.
var Compilado = ""

// VersionCompleta arma `0.2.21.2607260225` — el identificador que se compara de un vistazo.
// Sin identidad inyectada devuelve "dev".
func VersionCompleta() string {
	switch {
	case Version == "" && Build == "":
		return "dev"
	case Build == "":
		return Version
	case Version == "":
		return "dev." + Build
	default:
		return Version + "." + Build
	}
}

// selladoEn parsea Build a un instante. `ok=false` en un binario sin identidad inyectada:
// sin sello no hay con qué comparar, y ahí NO se avisa nada (callar es honesto; inventar una
// fecha para poder comparar no).
func selladoEn() (time.Time, bool) {
	if len(Build) != 10 {
		return time.Time{}, false
	}
	t, err := time.ParseInLocation("0601021504", Build, time.Local)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// margenDeSello absorbe la distancia entre el instante en que `bundle.sh` sella el número y
// el instante en que el linker termina de escribir el archivo.
//
// Sin este margen el binario recién instalado se denunciaría a sí mismo como «viejo»: su
// mtime es SIEMPRE posterior a su propio sello. 5 minutos cubre un build lento sin tapar un
// rebuild de verdad, que en la práctica llega minutos u horas después.
const margenDeSello = 5 * time.Minute

// masNuevoQueEsteBuild dice si el archivo en path se escribió después de que este binario se
// sellara — o sea, si en disco hay un build más nuevo que el que está corriendo.
//
// Compara mtime contra el sello, y no dos sellos entre sí, a propósito: leer el sello del OTRO
// binario obligaría a ejecutarlo (arbitrario, lento) o a rastrear sus bytes (frágil). El mtime
// lo responde con un `stat`.
func masNuevoQueEsteBuild(path string) (bool, string) {
	sello, ok := selladoEn()
	if !ok || path == "" {
		return false, ""
	}
	// Tras el rename atómico del self-update, /proc/self/exe reporta «(deleted)» (decisión #8
	// del paquete boton-actualizar). Un path así no se stat-ea: no es un archivo.
	if strings.HasSuffix(path, "(deleted)") {
		return false, ""
	}
	st, err := os.Stat(path)
	if err != nil || st.IsDir() {
		return false, ""
	}
	if !st.ModTime().After(sello.Add(margenDeSello)) {
		return false, ""
	}
	return true, st.ModTime().Local().Format("2006-01-02 15:04")
}

// avisoDeBuildViejo devuelve el aviso para Ajustes cuando hay un build más nuevo que el que
// corre, o "" cuando el que corre ES el último.
//
// Mira dos lugares, en orden de urgencia — los DOS son casos reales de esta casa:
//
//  1. **El propio ejecutable instalado.** `make dev-sync` reemplaza `~/.local/bin/arnesia`
//     mientras la app está abierta: el archivo ya es nuevo y el proceso sigue siendo el viejo.
//     Se arregla cerrando y reabriendo, y sin este aviso el operador no tiene cómo saberlo.
//  2. **El binario compilado en el repo** (`<repo>/bin/arnesia`, lo que deja `bundle.sh`).
//     Compilaste pero no instalaste.
func avisoDeBuildViejo(exePath, repo string) string {
	if nuevo, cuando := masNuevoQueEsteBuild(exePath); nuevo {
		return "el binario instalado es más nuevo (" + cuando +
			") que el que está corriendo — cerrá y reabrí la app"
	}
	if repo == "" {
		return ""
	}
	binRepo := filepath.Join(repo, "bin", nombreBinNuevo())
	if nuevo, cuando := masNuevoQueEsteBuild(binRepo); nuevo {
		return fmt.Sprintf("hay un build más nuevo sin instalar (%s en %s) — corré `make dev-sync`",
			cuando, binRepo)
	}
	return ""
}
