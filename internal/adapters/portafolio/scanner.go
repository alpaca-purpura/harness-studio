package portafolio

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Scanner es el walker físico READ-ONLY del Portafolio (D-DOM-3/6, S0-D1/D2): descubre
// las tres formas de instalación bajo un proyecto (S0-D2) sin escribir nada en disco.
type Scanner struct {
	// CCPluginsDir es el dir de metadata de Claude Code (default ~/.claude/plugins) —
	// INYECTABLE para tests, jamás la máquina real fuera de T9.
	CCPluginsDir string
	// MaxDepth acota el descenso recursivo por subcarpetas (default 4, C-P-11 monorepo).
	MaxDepth int
}

func (s *Scanner) ccPluginsDir() string {
	if s.CCPluginsDir != "" {
		return s.CCPluginsDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "plugins")
}

func (s *Scanner) maxDepth() int {
	if s.MaxDepth > 0 {
		return s.MaxDepth
	}
	return 4
}

// Escanear recorre root (validado abs+existe+dir; un bare repo no es escaneable, C-N-1) y
// devuelve TODOS los hallazgos crudos de las tres formas de instalación (S0-D2) + los
// subdirectorios anidados hasta MaxDepth (monorepo, C-P-11/C-N-13). NO carga ni persiste
// nada — eso es del usecase (§2.7).
func (s *Scanner) Escanear(ctx context.Context, root string) ([]domain.HallazgoInstalacion, error) {
	if !filepath.IsAbs(root) {
		return nil, fmt.Errorf("portafolio: root %q debe ser absoluto", root)
	}
	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("portafolio: root %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("portafolio: root %q no es un directorio", root)
	}
	if gd, ok := gitDir(root); ok && esBare(gd) {
		return nil, fmt.Errorf("portafolio: root %q es un bare repo (sin working tree), no escaneable", root)
	}

	var hallazgos []domain.HallazgoInstalacion

	// 1. proyecto-instalado: root/.claude/ poblado (nomenclatura §1, forma-instalada).
	if fi, serr := os.Stat(filepath.Join(root, ".claude")); serr == nil && fi.IsDir() {
		hallazgos = append(hallazgos, domain.HallazgoInstalacion{Dir: root, Tipo: domain.InstProyectoInstalado})
	}

	// 2. materializada: root/.claude/plugins/<id>/ (convención DevStudio, HS-12).
	hallazgos = append(hallazgos, s.escanearMaterializadas(root)...)

	// 3. lock DevStudio (.devstudio/arneses.yaml, detector 3° read-only, HS-12).
	hallazgos = append(hallazgos, s.escanearLock(root)...)

	// 4. referenciada-cc (enabledPlugins × installed_plugins.json × known_marketplaces.json, S0-D1).
	hallazgos = append(hallazgos, s.escanearCC(root)...)

	// 5. git-proyecto: el remote del PROYECTO se adjunta a cada hallazgo ya emitido con
	//    Dir bajo root (describe el proyecto, no cada instalación — RN-GIT-1, jamás se
	//    confunde con el home/registry del arnés).
	if gr, ok := ResolverRemotesGit(root); ok && gr.Origin != "" {
		for i := range hallazgos {
			if hallazgos[i].Dir == "" {
				continue
			}
			hallazgos[i].Eslabones = append(hallazgos[i].Eslabones, domain.EslabonOrigen{
				Fuente: "git-proyecto", Campo: "proyecto-remote", Valor: gr.Origin,
			})
		}
	}

	// 6. subcarpetas anidadas (monorepo).
	sub, serr := s.escanearSubcarpetas(ctx, root, root, 1)
	if serr != nil {
		return nil, serr
	}
	hallazgos = append(hallazgos, sub...)

	return hallazgos, nil
}

// escanearMaterializadas cubre root/.claude/plugins/<id>/ (S0-D2 `materializada`): sin
// `.claude-plugin/plugin.json` legal, el dir sigue siendo un hallazgo VISIBLE con Aviso
// no-reconocible (C-P-5) — jamás se omite ni crashea el resto del scan.
func (s *Scanner) escanearMaterializadas(root string) []domain.HallazgoInstalacion {
	dirPlugins := filepath.Join(root, ".claude", "plugins")
	entries, err := os.ReadDir(dirPlugins)
	if err != nil {
		return nil
	}
	var out []domain.HallazgoInstalacion
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pdir := filepath.Join(dirPlugins, e.Name())
		h := domain.HallazgoInstalacion{Dir: pdir, Tipo: domain.InstMaterializada}
		if _, perr := os.Stat(filepath.Join(pdir, ".claude-plugin", "plugin.json")); perr != nil {
			h.Aviso = fmt.Sprintf("no-reconocible: %s sin .claude-plugin/plugin.json", pdir) // C-P-5
		}
		if gr, gok := ResolverRemotesGit(pdir); gok && gr.Origin != "" {
			h.Eslabones = append(h.Eslabones, domain.EslabonOrigen{Fuente: "git-plugin", Campo: "registry", Valor: gr.Origin})
		}
		out = append(out, h)
	}
	return out
}

// lockEntry es una fila de `.devstudio/arneses.yaml` (HS-12: contrato estable
// `id·versión·canal·registry`, evolución aditiva — pedido recíproco fichado con DevStudio).
type lockEntry struct {
	ID       string `yaml:"id"`
	Version  string `yaml:"version"`
	Canal    string `yaml:"canal"`
	Registry string `yaml:"registry"`
}

type lockFile struct {
	Arneses []lockEntry `yaml:"arneses"`
}

// devStudioCacheDir es el caché local de rehidratación de DevStudio (nomenclatura-arnes.md
// §1: `~/.dev-studio/arneses/`). Slice 0 NO rehidrata desde el marketplace (sin red,
// S0-D7/P4) — solo comprueba si la entrada YA está ahí.
func devStudioCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dev-studio", "arneses")
}

// escanearLock lee el lock DevStudio READ-ONLY (detector 3°, HS-12): cada entrada aporta
// eslabones `lock-devstudio`; si el dir en el caché local no existe, el hallazgo sale
// IGUAL con Aviso visible (C-P-14 — "declarada-en-lock, ausente"), jamás se omite. Un
// lock ilegible (YAML roto) emite un único hallazgo-aviso, no crashea el resto del scan.
func (s *Scanner) escanearLock(root string) []domain.HallazgoInstalacion {
	ruta := filepath.Join(root, ".devstudio", "arneses.yaml")
	b, err := os.ReadFile(ruta) //nolint:gosec // G304: ruta bajo el proyecto que el caller eligió escanear.
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []domain.HallazgoInstalacion{{Aviso: fmt.Sprintf("lock %s ilegible: %v", ruta, err)}}
	}
	var lf lockFile
	if yerr := yaml.Unmarshal(b, &lf); yerr != nil {
		return []domain.HallazgoInstalacion{{Aviso: fmt.Sprintf("lock %s ilegible: %v", ruta, yerr)}}
	}

	var out []domain.HallazgoInstalacion
	for _, e := range lf.Arneses {
		h := domain.HallazgoInstalacion{
			Eslabones: []domain.EslabonOrigen{
				{Fuente: "lock-devstudio", Campo: "registry", Valor: e.Registry},
				{Fuente: "lock-devstudio", Campo: "version", Valor: e.Version},
				{Fuente: "lock-devstudio", Campo: "canal", Valor: e.Canal},
			},
		}
		dir := filepath.Join(devStudioCacheDir(), e.ID, e.Version)
		if fi, serr := os.Stat(dir); serr == nil && fi.IsDir() {
			h.Dir = dir
			h.Tipo = domain.InstMaterializada
		} else {
			h.Aviso = fmt.Sprintf("declarada en lock (%s@%s), dir ausente en el caché: %s", e.ID, e.Version, dir) // C-P-14
		}
		out = append(out, h)
	}
	return out
}

// settingsJSON es el subset de `.claude/settings.json` que el cruce CC necesita.
type settingsJSON struct {
	EnabledPlugins map[string]bool `json:"enabledPlugins"`
}

// ccInstalledPlugin es una entrada de `installed_plugins.json` (S0-D1).
type ccInstalledPlugin struct {
	Scope       string `json:"scope"`
	ProjectPath string `json:"projectPath"`
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
}

// leerInstalledPlugins lee `installed_plugins.json`: exige `"version":2` (S0-D1); shape
// inesperada o versión distinta → ok=false — el caller lo convierte en eslabón
// no-legible, JAMÁS crashea (el formato interno de CC puede driftar).
func leerInstalledPlugins(path string) (map[string][]ccInstalledPlugin, bool) {
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta fija de metadata CC, inyectable en tests.
	if err != nil {
		return nil, false
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, false
	}
	verRaw, ok := raw["version"]
	if !ok {
		return nil, false
	}
	var ver int
	if err := json.Unmarshal(verRaw, &ver); err != nil || ver != 2 {
		return nil, false
	}
	out := map[string][]ccInstalledPlugin{}
	for k, v := range raw {
		if k == "version" {
			continue
		}
		var entries []ccInstalledPlugin
		if err := json.Unmarshal(v, &entries); err != nil {
			continue // una clave individual ilegible no invalida el resto del archivo.
		}
		out[k] = entries
	}
	return out, true
}

// ccMarketplace es una entrada de `known_marketplaces.json` (S0-D1).
type ccMarketplace struct {
	Source struct {
		Source string `json:"source"`
		Repo   string `json:"repo"`
	} `json:"source"`
	InstallLocation string `json:"installLocation"`
}

func leerKnownMarketplaces(path string) (map[string]ccMarketplace, bool) {
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta fija de metadata CC, inyectable en tests.
	if err != nil {
		return nil, false
	}
	var m map[string]ccMarketplace
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}

// escanearCC cruza `.claude/settings.json#enabledPlugins` del proyecto con el registro
// global de Claude Code (S0-D1): cada `"<id>@<mkt>": true` con un install-record cuyo
// `projectPath == root` aporta un hallazgo `referenciada-cc`, Dir = installPath del caché
// global. Metadata CC ilegible → eslabón `no-legible` visible, jamás crash.
func (s *Scanner) escanearCC(root string) []domain.HallazgoInstalacion {
	settingsPath := filepath.Join(root, ".claude", "settings.json")
	b, err := os.ReadFile(settingsPath) //nolint:gosec // G304: ruta bajo el proyecto que el caller eligió escanear.
	if err != nil {
		return nil // sin settings.json, nada que cruzar — legal.
	}
	var st settingsJSON
	if uerr := json.Unmarshal(b, &st); uerr != nil {
		return []domain.HallazgoInstalacion{{
			Eslabones: []domain.EslabonOrigen{{Fuente: "no-legible", Campo: "registry", Valor: settingsPath}},
			Aviso:     fmt.Sprintf("%s ilegible: %v", settingsPath, uerr),
		}}
	}
	if len(st.EnabledPlugins) == 0 {
		return nil
	}

	installed, iok := leerInstalledPlugins(filepath.Join(s.ccPluginsDir(), "installed_plugins.json"))
	marketplaces, mok := leerKnownMarketplaces(filepath.Join(s.ccPluginsDir(), "known_marketplaces.json"))

	var out []domain.HallazgoInstalacion
	for key, enabled := range st.EnabledPlugins {
		if !enabled {
			continue
		}
		_, mkt, ok := strings.Cut(key, "@")
		if !ok {
			continue
		}
		if !iok {
			out = append(out, domain.HallazgoInstalacion{
				Eslabones: []domain.EslabonOrigen{{Fuente: "no-legible", Campo: "version", Valor: key}},
				Aviso:     fmt.Sprintf("enabledPlugins declara %q pero installed_plugins.json es ilegible", key),
			})
			continue
		}
		var record *ccInstalledPlugin
		for i, e := range installed[key] {
			if e.ProjectPath == root {
				record = &installed[key][i]
				break
			}
		}
		if record == nil {
			out = append(out, domain.HallazgoInstalacion{
				Aviso: fmt.Sprintf("enabledPlugins declara %q pero sin record de instalación para %s", key, root),
			})
			continue
		}
		h := domain.HallazgoInstalacion{
			Dir:  record.InstallPath,
			Tipo: domain.InstReferenciadaCC,
			Eslabones: []domain.EslabonOrigen{
				{Fuente: "cc-plugins", Campo: "version", Valor: record.Version},
			},
		}
		switch {
		case !mok:
			h.Eslabones = append(h.Eslabones, domain.EslabonOrigen{Fuente: "no-legible", Campo: "registry", Valor: mkt})
		default:
			if m, mfound := marketplaces[mkt]; mfound && m.Source.Repo != "" {
				h.Eslabones = append(h.Eslabones, domain.EslabonOrigen{Fuente: "cc-plugins", Campo: "registry", Valor: m.Source.Repo})
			}
		}
		out = append(out, h)
	}
	return out
}

// escanearSubcarpetas desciende recursivamente desde dir (acotado a MaxDepth, C-P-11)
// buscando `.claude/` anidados (monorepo, C-N-13): cada uno es OTRO ámbito de instalación
// `proyecto-instalado`. No sigue symlinks que escapen de base (C-P-12); salta
// `node_modules`/`.git`/dirs ocultos salvo `.claude`/`.devstudio` (ya cubiertos al nivel
// raíz, no se re-procesan anidados); cancelable vía ctx (C-P-13).
func (s *Scanner) escanearSubcarpetas(ctx context.Context, base, dir string, depth int) ([]domain.HallazgoInstalacion, error) {
	if depth > s.maxDepth() {
		return nil, nil
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil //nolint:nilerr // subdir ilegible (permisos/etc): no aborta el scan completo del root.
	}

	var out []domain.HallazgoInstalacion
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "node_modules" || name == ".git" || name == ".devstudio" {
			continue
		}
		sub := filepath.Join(dir, name)

		if name == ".claude" {
			if sub == filepath.Join(base, ".claude") {
				continue // el .claude/ del root ya lo cubrió Escanear (paso 1) — no duplicar.
			}
			out = append(out, domain.HallazgoInstalacion{Dir: filepath.Dir(sub), Tipo: domain.InstProyectoInstalado})
			continue // no descender DENTRO de un .claude/ buscando más .claude/.
		}
		if strings.HasPrefix(name, ".") {
			continue // resto de dirs ocultos: fuera del alcance del walker.
		}

		if fi, lerr := os.Lstat(sub); lerr == nil && fi.Mode()&os.ModeSymlink != 0 {
			resolved, everr := filepath.EvalSymlinks(sub)
			if everr != nil || !dentroDe(base, resolved) {
				continue // symlink roto o que escapa de base (C-P-12): no se sigue.
			}
		}

		children, cerr := s.escanearSubcarpetas(ctx, base, sub, depth+1)
		if cerr != nil {
			return nil, cerr
		}
		out = append(out, children...)
	}
	return out, nil
}

// dentroDe reporta si child es igual a o está contenido dentro de parent.
func dentroDe(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
