package selfupdate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// conIdentidad fija las variables de ldflags durante un test y las restaura al salir.
func conIdentidad(t *testing.T, version, build, compilado string) {
	t.Helper()
	v, b, c := Version, Build, Compilado
	Version, Build, Compilado = version, build, compilado
	t.Cleanup(func() { Version, Build, Compilado = v, b, c })
}

func TestVersionCompletaArmaSemverMasSello(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	if got := VersionCompleta(); got != "0.2.21.2607260225" {
		t.Errorf("VersionCompleta() = %q", got)
	}
}

// Un `go build` pelado (CI, `go run`) no inyecta nada. Decir «dev» es honesto; inventar un
// número sería justo la mentira que RF-231 vino a sacar.
func TestVersionCompletaSinIdentidadDiceDev(t *testing.T) {
	conIdentidad(t, "", "", "")
	if got := VersionCompleta(); got != "dev" {
		t.Errorf("VersionCompleta() = %q, quiero dev", got)
	}
}

func TestVersionCompletaTolerantesAMediaIdentidad(t *testing.T) {
	conIdentidad(t, "0.2.21", "", "")
	if got := VersionCompleta(); got != "0.2.21" {
		t.Errorf("solo semver = %q", got)
	}
	conIdentidad(t, "", "2607260225", "")
	if got := VersionCompleta(); got != "dev.2607260225" {
		t.Errorf("solo sello = %q", got)
	}
}

// El sello es monótono por construcción: más grande = más nuevo. Es la propiedad por la que se
// eligió un timestamp y no un contador, así que se prueba.
func TestSelloMasGrandeEsMasNuevo(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "")
	viejo, ok := selladoEn()
	if !ok {
		t.Fatal("no parseó el sello")
	}
	conIdentidad(t, "0.2.21", "2607260318", "")
	nuevo, ok := selladoEn()
	if !ok {
		t.Fatal("no parseó el sello")
	}
	if !nuevo.After(viejo) {
		t.Errorf("2607260318 no quedó después de 2607260225 (%v vs %v)", nuevo, viejo)
	}
}

func TestSelloIlegibleNoRompe(t *testing.T) {
	for _, malo := range []string{"", "abc", "26072602251", "no-soy-fecha"} {
		conIdentidad(t, "0.2.21", malo, "")
		if _, ok := selladoEn(); ok {
			t.Errorf("selladoEn() aceptó %q", malo)
		}
	}
}

// archivoConMtime crea un archivo y le fija el mtime.
func archivoConMtime(t *testing.T, path string, mt time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("binario"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chtimes(path, mt, mt); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

// EL caso del operador: `make dev-sync` reemplaza el binario mientras la app está abierta. El
// archivo ya es nuevo, el proceso sigue siendo el viejo, y sin aviso no hay cómo enterarse.
func TestAvisaCuandoElInstaladoEsMasNuevo(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	sello, _ := selladoEn()

	exe := filepath.Join(t.TempDir(), "arnesia")
	archivoConMtime(t, exe, sello.Add(2*time.Hour))

	aviso := avisoDeBuildViejo(exe, "")
	if !strings.Contains(aviso, "más nuevo") || !strings.Contains(aviso, "reabrí") {
		t.Errorf("aviso = %q", aviso)
	}
}

func TestAvisaCuandoElRepoTieneUnBuildSinInstalar(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	sello, _ := selladoEn()

	dir := t.TempDir()
	exe := filepath.Join(dir, "instalado", "arnesia")
	archivoConMtime(t, exe, sello.Add(-time.Hour)) // el instalado ES este build

	repo := filepath.Join(dir, "repo")
	archivoConMtime(t, filepath.Join(repo, "bin", nombreBinNuevo()), sello.Add(3*time.Hour))

	aviso := avisoDeBuildViejo(exe, repo)
	if !strings.Contains(aviso, "sin instalar") || !strings.Contains(aviso, "dev-sync") {
		t.Errorf("aviso = %q", aviso)
	}
}

// El caso normal —corro lo último— NO puede avisar nada: un aviso que aparece siempre no se
// lee nunca.
func TestSinNadaMasNuevoNoAvisa(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	sello, _ := selladoEn()

	dir := t.TempDir()
	exe := filepath.Join(dir, "arnesia")
	archivoConMtime(t, exe, sello.Add(-time.Minute))

	if aviso := avisoDeBuildViejo(exe, ""); aviso != "" {
		t.Errorf("avisó sin motivo: %q", aviso)
	}
}

// El binario recién instalado tiene mtime POSTERIOR a su propio sello (el linker escribe
// después de que bundle.sh sella). Sin margen se denunciaría a sí mismo como viejo en cada
// arranque — el falso positivo que volvería inútil al aviso.
func TestElPropioBuildNoSeDenunciaASiMismo(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	sello, _ := selladoEn()

	exe := filepath.Join(t.TempDir(), "arnesia")
	archivoConMtime(t, exe, sello.Add(90*time.Second)) // el build tardó minuto y medio

	if aviso := avisoDeBuildViejo(exe, ""); aviso != "" {
		t.Errorf("el binario se denunció a sí mismo: %q", aviso)
	}
}

// Sin sello no hay con qué comparar: callar es honesto, inventar una fecha para poder
// comparar no lo sería.
func TestSinSelloNoAvisaNada(t *testing.T) {
	conIdentidad(t, "", "", "")
	exe := filepath.Join(t.TempDir(), "arnesia")
	archivoConMtime(t, exe, time.Now().Add(48*time.Hour))

	if aviso := avisoDeBuildViejo(exe, ""); aviso != "" {
		t.Errorf("un build dev avisó: %q", aviso)
	}
}

// Tras el rename atómico del self-update, /proc/self/exe reporta «(deleted)» (decisión #8 del
// paquete boton-actualizar). Eso no es un archivo: no se stat-ea ni se avisa por él.
func TestRutaBorradaNoRompeNiAvisa(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	if aviso := avisoDeBuildViejo("/usr/bin/arnesia (deleted)", ""); aviso != "" {
		t.Errorf("aviso sobre una ruta borrada: %q", aviso)
	}
}

func TestRutaInexistenteNoRompe(t *testing.T) {
	conIdentidad(t, "0.2.21", "2607260225", "2026-07-26 02:25")
	if aviso := avisoDeBuildViejo("/no/existe/arnesia", "/tampoco/existe"); aviso != "" {
		t.Errorf("aviso = %q", aviso)
	}
}
