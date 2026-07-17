// Package portafolio implementa el adapter físico del Portafolio de arneses (Slice 0):
// el walker que descubre instalaciones (scanner.go), el parser manual de `.git/config`
// (gitconfig.go, S0-D8 — sin go-git) y el hasher de deriva (deriva.go, T5) + el store
// (store.go, T6). Depende SOLO de domain (+ yaml para el lock DevStudio) — go-arch-lint
// componente `portafolio`.
package portafolio

import (
	"os"
	"path/filepath"
	"strings"
)

// GitRemotes son los remotes resueltos de un repo git (RN-GIT-1: esto describe el
// PROYECTO, jamás el home/registry del arnés). Origin gana como hogar del proyecto;
// Upstream es dato secundario (multi-remote, C-N-2) — nunca se elige uno en silencio.
type GitRemotes struct {
	Origin   string
	Upstream string
}

// gitDir resuelve el `.git` de dir: si es un directorio, ES el gitdir; si es un archivo
// `gitdir: <ruta>` (worktree, C-N-1), resuelve la ruta (relativa al padre del archivo si
// no es absoluta). ok=false si dir no tiene ningún `.git` resoluble.
func gitDir(dir string) (string, bool) {
	p := filepath.Join(dir, ".git")
	info, err := os.Lstat(p)
	if err != nil {
		return "", false
	}
	if info.IsDir() {
		return p, true
	}
	b, err := os.ReadFile(p) //nolint:gosec // G304: ruta bajo el dir del proyecto que el caller eligió escanear.
	if err != nil {
		return "", false
	}
	const prefix = "gitdir:"
	line := strings.TrimSpace(string(b))
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	ref := strings.TrimSpace(strings.TrimPrefix(line, prefix))
	if ref == "" {
		return "", false
	}
	if !filepath.IsAbs(ref) {
		ref = filepath.Join(dir, ref)
	}
	return filepath.Clean(ref), true
}

// gitCommonDir resuelve el gitdir COMÚN del repo de dir (S1-D26): dos working trees del
// mismo repo (worktree linkeado + principal) comparten el común. Un worktree linkeado tiene
// `<gitdir>/commondir` apuntando (relativo o absoluto) al `.git` principal — el mecanismo
// canónico de git, no una heurística de nombres; un repo normal ES su propio común.
func gitCommonDir(dir string) (string, bool) {
	gd, ok := gitDir(dir)
	if !ok {
		return "", false
	}
	b, err := os.ReadFile(filepath.Join(gd, "commondir")) //nolint:gosec // G304: ruta derivada del dir que el caller eligió escanear.
	if err != nil {
		return gd, true
	}
	common := strings.TrimSpace(string(b))
	if common == "" {
		return gd, true
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(gd, common)
	}
	return filepath.Clean(common), true
}

// esBare reporta si el repo cuyo gitDir es gitDirPath es un bare repo (C-N-1: sin
// working tree, no escaneable) — vía `core.bare` de su config.
func esBare(gitDirPath string) bool {
	cfg, err := parseGitConfig(filepath.Join(gitDirPath, "config"))
	if err != nil {
		return false
	}
	return strings.EqualFold(cfg["core"]["bare"], "true")
}

// ResolverRemotesGit lee el remote origin/upstream del repo git en dir. ok=false si dir
// no es un repo git resoluble (sin `.git`, o `.git` roto). Lee el config del gitdir COMÚN
// (S1-D26): los remotes son repo-level y en un worktree linkeado viven ahí — el config del
// gitdir propio del worktree no los tiene (por eso el root de luana-vitalia salía sin scope).
func ResolverRemotesGit(dir string) (GitRemotes, bool) {
	gd, ok := gitCommonDir(dir)
	if !ok {
		return GitRemotes{}, false
	}
	cfg, err := parseGitConfig(filepath.Join(gd, "config"))
	if err != nil {
		return GitRemotes{}, false
	}
	return GitRemotes{
		Origin:   cfg[`remote "origin"`]["url"],
		Upstream: cfg[`remote "upstream"`]["url"],
	}, true
}

// parseGitConfig parsea el subset INI de `.git/config` que necesitamos: secciones
// `[nombre]` / `[nombre "sub"]`, claves `k = v`. Sin dependencia externa (S0-D8: nada de
// go-git — el módulo se queda con sus 2 deps).
func parseGitConfig(path string) (map[string]map[string]string, error) {
	b, err := os.ReadFile(path) //nolint:gosec // G304: ruta bajo el dir del proyecto que el caller eligió escanear.
	if err != nil {
		return nil, err
	}
	cfg := map[string]map[string]string{}
	var section string
	for _, raw := range strings.Split(string(b), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
			if _, ok := cfg[section]; !ok {
				cfg[section] = map[string]string{}
			}
			continue
		}
		if section == "" {
			continue
		}
		if i := strings.Index(line, "="); i >= 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			cfg[section][k] = v
		}
	}
	return cfg, nil
}
