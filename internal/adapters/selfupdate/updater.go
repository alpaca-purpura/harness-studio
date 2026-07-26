// Package selfupdate implements ports.SelfUpdater: el self-update sin sudo del binario
// instalado (paquete boton-actualizar, RF-104). Mecánica pura — build del repo
// configurado, inspección del binario producido SIN ejecutarlo, instalación
// write-tmp→rename atómico y re-exec; las reglas («ya al día», corte al primer fallo,
// un update a la vez) viven en usecase.SelfUpdateService.
package selfupdate

import (
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

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
	if _, err := os.Stat(filepath.Join(repo, "scripts", "bundle.sh")); err != nil {
		return "", fmt.Errorf("el repo no trae scripts/bundle.sh: %w", err)
	}
	for _, tool := range []string{"go", "pnpm", "bash"} {
		if _, err := lookPathEn(tool, pathAumentado()); err != nil {
			return "", fmt.Errorf("toolchain incompleta: %q no está en PATH (feature de operador-dev)", tool)
		}
	}
	return fmt.Sprintf("repo %s · módulo esperado · go/pnpm/bash presentes", repo), nil
}

// candidatosPATH — ubicaciones ESTÁNDAR de instalación de go/pnpm que un proceso
// lanzado desde el launcher gráfico no hereda: el `.desktop` arranca con el PATH de la
// sesión (systemd/display-manager), fijado al iniciar sesión — jamás lee
// ~/.profile/~/.bashrc como sí hace una terminal interactiva. Root cause del bug real
// (no un problema de build viejo): "toolchain incompleta: X no está en PATH" con la
// herramienta instalada y en el PATH de CUALQUIER terminal del mismo usuario. El fix
// original (HS) solo cubrió go; pnpm pegaba el mismo síntoma vía nvm (ver
// nvmBinsInstalados) y quedó afuera hasta ahora.
func candidatosPATH() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	cands := []string{
		filepath.Join(home, ".local", "go", "bin"), // convención de este repo (ver ~/.profile)
		filepath.Join(home, "go", "bin"),           // convención oficial go.dev (go install)
		"/usr/local/go/bin",                        // convención oficial go.dev (tarball)
	}
	return append(cands, nvmBinsInstalados(home)...)
}

// nvmBinsInstalados — mismo root cause que go (ver candidatosPATH): el launcher gráfico
// no corre ~/.bashrc, así que nvm.sh nunca carga y el pnpm de la versión de node activa
// queda fuera del PATH aunque esté instalado (ver ~/.nvm/versions/node/*/bin). nvm no
// fija un symlink "current" — cada versión vive en su propio dir versionado — así que se
// agregan TODOS los bin/ instalados: cualquiera resuelve pnpm/node/npm igual de bien
// para bundle.sh.
func nvmBinsInstalados(home string) []string {
	nvmDir := os.Getenv("NVM_DIR")
	if nvmDir == "" {
		nvmDir = filepath.Join(home, ".nvm")
	}
	versionesDir := filepath.Join(nvmDir, "versions", "node")
	entries, err := os.ReadDir(versionesDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(versionesDir, e.Name(), "bin"))
		}
	}
	return out
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
// pathAumentado() sin mutar el proceso real).
func lookPathEn(tool, pathEnv string) (string, error) {
	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == "" {
			continue
		}
		candidato := filepath.Join(dir, tool)
		if st, err := os.Stat(candidato); err == nil && !st.IsDir() && st.Mode().Perm()&0o111 != 0 {
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

// Build compila el árbol local (paso ②): scripts/bundle.sh --daemon-only con cwd=repo,
// cancelable por ctx. Un fallo devuelve la cola de la salida como detalle honesto.
func (u *Updater) Build(ctx context.Context) (string, error) {
	//nolint:gosec // G204: es EL diseño (RF-104 ②): correr el bundle.sh del repo CONFIGURADO
	// al daemon (flag/env, jamás del request — RF-106); Verificar ya ancló el árbol al módulo esperado.
	cmd := exec.CommandContext(ctx, "bash", filepath.Join("scripts", "bundle.sh"), "--daemon-only")
	cmd.Dir = u.repoAtual()
	cmd.Env = envConPathAumentado()
	// La cancelación mata el GRUPO entero (bash + go/pnpm/vite hijos): matar solo a
	// bash dejaría a los hijos corriendo Y sosteniendo el pipe (CombinedOutput jamás
	// retornaría). WaitDelay corta el pipe si algún nieto sobrevive al SIGKILL.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	out, err := cmd.CombinedOutput()
	if err != nil {
		return cola(string(out)), fmt.Errorf("bundle.sh --daemon-only: %w", err)
	}
	return "scripts/bundle.sh --daemon-only OK (SPA + go build)", nil
}

// VerificarBinario inspecciona repo/bin/arnesia SIN ejecutarlo (paso ③): existe,
// ejecutable, buildinfo legible → huella del binario NUEVO.
func (u *Updater) VerificarBinario(_ context.Context) (string, string, error) {
	path := u.binNuevo()
	st, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("el build no dejó bin/arnesia: %w", err)
	}
	if !st.Mode().IsRegular() || st.Mode().Perm()&0o111 == 0 {
		return "", "", fmt.Errorf("bin/arnesia no es un ejecutable regular (modo %v)", st.Mode())
	}
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("bin/arnesia no trae buildinfo de Go legible: %w", err)
	}
	huella, _, _ := huellaDeSettings(bi.Settings)
	return huella, fmt.Sprintf("bin/arnesia · huella %s · %d bytes", huella, st.Size()), nil
}

// Instalar reemplaza el ejecutable corriente (paso ④): write a tmp EN EL MISMO
// directorio (mismo filesystem) → chmod 0755 → rename atómico. Cualquier fallo limpia
// el tmp y deja el binario instalado INTACTO.
func (u *Updater) Instalar(_ context.Context) (string, error) {
	src, err := os.Open(u.binNuevo())
	if err != nil {
		return "", fmt.Errorf("abrir el binario nuevo: %w", err)
	}
	defer func() { _ = src.Close() }() // lectura: el Close sin error que importe.

	dir := filepath.Dir(u.exePath)
	tmp, err := os.CreateTemp(dir, ".arnesia-nuevo-*")
	if err != nil {
		return "", fmt.Errorf("el directorio de instalación %s no acepta escritura: %w", dir, err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name()) // no-op tras el rename; limpia el resto de caminos.
	}()
	if _, err := io.Copy(tmp, src); err != nil {
		return "", fmt.Errorf("copiar el binario nuevo: %w", err)
	}
	if err := tmp.Chmod(0o755); err != nil {
		return "", fmt.Errorf("chmod del binario nuevo: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("cerrar el binario nuevo: %w", err)
	}
	if err := os.Rename(tmp.Name(), u.exePath); err != nil {
		return "", fmt.Errorf("rename atómico sobre %s: %w", u.exePath, err)
	}
	return fmt.Sprintf("instalado en %s (rename atómico, sin sudo)", u.exePath), nil
}

// Reiniciar re-ejecuta el daemon (paso ⑤): mismo path (el capturado al construir),
// mismos args, mismo entorno. En Linux exec reemplaza el proceso — no retorna si va bien.
func (u *Updater) Reiniciar() error {
	//nolint:gosec // G204: re-exec de SÍ MISMO (RF-105) — ruta capturada al boot, args propios.
	return syscall.Exec(u.exePath, os.Args, os.Environ())
}

func (u *Updater) binNuevo() string {
	return filepath.Join(u.repoAtual(), "bin", "arnesia")
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
