// Package selfupdate implements ports.SelfUpdater: el self-update sin sudo del binario
// instalado (paquete boton-actualizar, RF-104). Mecánica pura — build del repo
// configurado, inspección del binario producido SIN ejecutarlo, instalación y re-exec;
// las reglas («ya al día», corte al primer fallo, un update a la vez) viven en
// usecase.SelfUpdateService. La mecánica atada al OS (nombre del binario, script de
// bundle, ejecutabilidad, PATH, instalar/reiniciar) vive en os_{unix,windows}.go +
// instalar_{unix,windows}.go — en unix instala con rename atómico + re-exec; en
// Windows degrada honesto (paquete 2026-08-13-compilacion-windows, decisión ①).
package selfupdate

import (
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// moduloEsperado ancla el paso Verificar al árbol de ESTE producto (decisión #4c del
// paquete: «es el árbol esperado») — un --repo apuntando a otro módulo Go se rechaza.
const moduloEsperado = "module github.com/alpacapurpura/arnesia"

// colaDetalle acota el stderr que viaja como detalle de un build fallido (RF-104: el
// error viaja con stderr, pero la respuesta no arrastra megabytes de log).
const colaDetalle = 1600

// Updater implements ports.SelfUpdater over a configured repo path and the executable
// path captured at construction time (decisión #8: tras el rename, /proc/self/exe
// reporta «(deleted)» — leerlo tarde re-ejecutaría una ruta que ya no existe).
type Updater struct {
	exePath string

	// mu guarda repo: ConfigurarRepo (bugfix fix-repo-self-update, RF-109) lo fija en
	// caliente mientras Version/Verificar/Build pueden estar leyéndolo concurrentemente.
	mu   sync.RWMutex
	repo string
}

// New returns an Updater for the daemon binary now running. repo may be empty (sin
// --repo): Version lo reporta honesto y el usecase corta antes de cualquier paso.
func New(repo string) (*Updater, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("selfupdate: resolver el ejecutable corriente: %w", err)
	}
	return &Updater{repo: repo, exePath: exe}, nil
}

// Version reporta la identidad del binario corriendo (RF-107): huella del buildinfo
// del PROPIO proceso, ruta real del ejecutable, escribibilidad del directorio y repo
// configurado.
func (u *Updater) Version() ports.VersionInfo {
	repo := u.repoAtual()
	v := ports.VersionInfo{
		Huella:      "dev",
		InstaladoEn: u.exePath,
		Repo:        repo,
		Escribible:  dirEscribible(filepath.Dir(u.exePath)),
		// La identidad del BUILD (semver + sello) es aparte de la del COMMIT (huella): dos
		// compilaciones del mismo árbol comparten huella y no comparten sello — RF-231.
		Version:    VersionCompleta(),
		Compilado:  Compilado,
		AvisoBuild: avisoDeBuildViejo(u.exePath, repo),
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		v.Huella, v.Fecha, v.Sucio = huellaDeSettings(bi.Settings)
	}
	return v
}

// Verificar valida el terreno (paso ①): repo presente y del módulo esperado, script de
// build presente, toolchain (go · pnpm) resoluble en PATH.
func (u *Updater) Verificar(_ context.Context) (string, error) {
	return validarRepo(u.repoAtual())
}

// ConfigurarRepo valida path (mismas reglas que Verificar, corriendo SOBRE el
// candidato — no sobre u.repo) y, si pasa, lo fija como repo activo (bugfix
// fix-repo-self-update, RF-109). No persiste — eso es responsabilidad del usecase vía
// ports.RepoConfigStore.
func (u *Updater) ConfigurarRepo(_ context.Context, path string) (string, error) {
	detalle, err := validarRepo(path)
	if err != nil {
		return "", err
	}
	u.mu.Lock()
	u.repo = path
	u.mu.Unlock()
	return detalle, nil
}

// repoAtual devuelve el repo activo bajo lock de lectura (ConfigurarRepo puede
// cambiarlo concurrentemente con Version/Verificar/Build/binNuevo).
func (u *Updater) repoAtual() string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.repo
}

// validarRepo es el terreno del paso ① (RF-104), factorizado para que Verificar (sobre
// el repo YA activo) y ConfigurarRepo (sobre un candidato AÚN no activo) compartan
// exactamente las mismas reglas — un path que no pasa esto tampoco debería poder
// guardarse como «el repo configurado» (decisión #5 del bugfix fix-repo-self-update).
func validarRepo(repo string) (string, error) {
	if repo == "" {
		return "", errors.New("repo no configurado (arranca el daemon con --repo o ARNESIA_REPO)")
	}
	gomod, err := os.ReadFile(filepath.Join(repo, "go.mod")) //nolint:gosec // G304: el repo lo configura el operador (--repo/ARNESIA_REPO/Ajustes) y esta función ES su validación.
	if err != nil {
		return "", fmt.Errorf("el repo configurado no es un árbol Go legible: %w", err)
	}
	if !strings.Contains(string(gomod), moduloEsperado) {
		return "", fmt.Errorf("%s/go.mod no es el módulo esperado (%s)", repo, moduloEsperado)
	}
	if _, err := os.Stat(filepath.Join(repo, scriptDeBundle())); err != nil {
		return "", fmt.Errorf("el repo no trae %s: %w", scriptDeBundle(), err)
	}
	for _, tool := range toolchainRequerida() {
		if _, err := lookPathEn(tool, pathAumentado()); err != nil {
			return "", fmt.Errorf("toolchain incompleta: %q no está en PATH (feature de operador-dev)", tool)
		}
	}
	return fmt.Sprintf("repo %s · módulo esperado · %s presentes", repo, strings.Join(toolchainRequerida(), "/")), nil
}

// candidatosPATH — ubicaciones ESTÁNDAR de instalación de la toolchain que un proceso
// lanzado desde el launcher gráfico (unix) o un acceso directo (Windows) puede no
// heredar. Root cause del bug real: "toolchain incompleta: X no está en PATH" con la
// herramienta instalada y en el PATH de CUALQUIER terminal del mismo usuario. La lista
// concreta es por-OS (candidatosPATHOS en os_{unix,windows}.go).
func candidatosPATH() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return candidatosPATHOS(home)
}

// pathAumentado agrega los candidatosPATH() que EXISTEN al PATH heredado — nunca lo
// reemplaza, nunca muta os.Environ() del proceso (concurrente con otros pasos).
func pathAumentado() string {
	base := os.Getenv("PATH")
	var extra []string
	for _, dir := range candidatosPATH() {
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			extra = append(extra, dir)
		}
	}
	if len(extra) == 0 {
		return base
	}
	return strings.Join(extra, string(os.PathListSeparator)) + string(os.PathListSeparator) + base
}

// lookPathEn busca tool en cada directorio de pathEnv (semántica de exec.LookPath, pero
// contra un PATH explícito en vez de leer os.Getenv("PATH") — necesario para probar
// pathAumentado() sin mutar el proceso real). El veredicto por-candidato es del OS:
// bit de ejecución en unix, PATHEXT en Windows (ejecutableEnDir).
func lookPathEn(tool, pathEnv string) (string, error) {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		if candidato, ok := ejecutableEnDir(dir, tool); ok {
			return candidato, nil
		}
	}
	return "", fmt.Errorf("%s: %w", tool, exec.ErrNotFound)
}

// envConPathAumentado copia el entorno del proceso reemplazando SOLO PATH — para el
// subproceso de Build(), que hereda env por defecto (mismo PATH roto que validarRepo
// ya corrigió para el check; sin esto el check pasa pero bundle.sh igual falla al
// invocar go/pnpm adentro).
func envConPathAumentado() []string {
	env := os.Environ()
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, "PATH=") {
			out = append(out, e)
		}
	}
	return append(out, "PATH="+pathAumentado())
}

// Build compila el árbol local (paso ②): el script de bundle del OS (comandoBundle —
// bash bundle.sh en unix, python bundle.py en Windows) con cwd=repo, cancelable por
// ctx. Un fallo devuelve la cola de la salida como detalle honesto.
func (u *Updater) Build(ctx context.Context) (string, error) {
	cmd := comandoBundle(ctx, u.repoAtual())
	cmd.Env = envConPathAumentado()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return cola(string(out)), fmt.Errorf("%s: %w", descripcionBundle, err)
	}
	return descripcionBundle + " OK (SPA + go build)", nil
}

// VerificarBinario inspecciona repo/bin/<binario> SIN ejecutarlo (paso ③): existe,
// ejecutable (bit de ejecución en unix, extensión .exe en Windows — esEjecutable),
// buildinfo legible → huella del binario NUEVO.
func (u *Updater) VerificarBinario(_ context.Context) (string, string, error) {
	path := u.binNuevo()
	st, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("el build no dejó bin/%s: %w", nombreBinNuevo(), err)
	}
	if !st.Mode().IsRegular() || !esEjecutable(st.Mode(), path) {
		return "", "", fmt.Errorf("bin/%s no es un ejecutable regular (modo %v)", nombreBinNuevo(), st.Mode())
	}
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("bin/%s no trae buildinfo de Go legible: %w", nombreBinNuevo(), err)
	}
	huella, _, _ := huellaDeSettings(bi.Settings)
	return huella, fmt.Sprintf("bin/%s · huella %s · %d bytes", nombreBinNuevo(), huella, st.Size()), nil
}

// Instalar (paso ④) y Reiniciar (paso ⑤) viven en instalar_{unix,windows}.go: unix
// hace rename atómico + re-exec; Windows degrada honesto (staged + error accionable).

func (u *Updater) binNuevo() string {
	return filepath.Join(u.repoAtual(), "bin", nombreBinNuevo())
}

// huellaDeSettings extrae (huella, fecha, sucio) de los settings del buildinfo:
// vcs.revision[:7] + sufijo "+sucio" si vcs.modified (decisión #7); sin VCS → "dev".
func huellaDeSettings(settings []debug.BuildSetting) (string, string, bool) {
	var rev, fecha string
	var sucio bool
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.time":
			fecha, _, _ = strings.Cut(s.Value, "T")
		case "vcs.modified":
			sucio = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev", "", false
	}
	if len(rev) > 7 {
		rev = rev[:7]
	}
	if sucio {
		rev += "+sucio"
	}
	return rev, fecha, sucio
}

// dirEscribible prueba la escritura REAL (crear y borrar un tmp): cubre permisos, ACLs
// y filesystems read-only — el rename del paso ④ exige exactamente esto.
func dirEscribible(dir string) bool {
	f, err := os.CreateTemp(dir, ".arnesia-w-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

// cola devuelve los últimos colaDetalle bytes de s (el final del stderr es donde vive
// el error real de un build).
func cola(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= colaDetalle {
		return s
	}
	return "…" + s[len(s)-colaDetalle:]
}
