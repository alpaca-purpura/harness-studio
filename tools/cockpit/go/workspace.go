// workspace.go — resolución de root + sistemas. Dos modos:
//
//   single  (paridad con cockpit/lib/workspace.ts): un workspace root; sistemas =
//           subdirs con docs/product + pseudo-sistema platform.
//   multi   (Fase 3 · daemon): N workspaces del registry ~/.cockpit/cockpit.yaml;
//           cada sistema se expone COMPUESTA como "{proyecto}/{sistema}" — el sistema
//           switcher de la UI se convierte en selector global sin cambios de UI.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

const platformSlug = "platform"

var (
	rootOnce   sync.Once
	cachedRoot string
	rootErr    error

	// multiProjects ≠ nil → modo multi-workspace.
	multiProjects []registryProject
)

var slugRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var regexpNonSlug = regexp.MustCompile(`[^a-z0-9_-]+`)

func isValidSlug(slug string) bool { return slugRe.MatchString(slug) }

func enableMultiMode(projects []registryProject) {
	multiProjects = projects
}

func isMultiMode() bool { return multiProjects != nil }

// ── Fase 3 · sistemas con workspace externo (cross-repo) ─────────────────────
// Un sistema puede declarar `workspace` (un path/repo distinto al de su empresa,
// ej. prenter→prenter-harness). systemWorkspaces es la tabla de reenvío: la key
// compuesta "empresa/sistema" resuelve a (root externo, sistema local). Se arma de
// TODOS los proyectos (incl. inactivos) en serve() → resolveSistema la consulta
// ANTES del lookup normal. La key sigue siendo `empresa/sistema` (contrato UI intacto).

type sysWorkspace struct {
	root    string
	sistema string // sistema local (subdir con docs/product) dentro del workspace, o "platform"
}

var systemWorkspaces = map[string]sysWorkspace{}

func enableSystemWorkspaces(projects []registryProject) {
	m := map[string]sysWorkspace{}
	for _, p := range projects {
		for _, s := range p.Directorio.Sistemas {
			if s.Workspace == "" {
				continue
			}
			inner := soleSistemaIn(s.Workspace)
			if inner == "" || !dirExists(boardDirFor(s.Workspace, inner)) {
				continue // workspace sin board navegable → el sistema queda como gap
			}
			m[p.Name+"/"+s.Slug] = sysWorkspace{root: s.Workspace, sistema: inner}
		}
	}
	systemWorkspaces = m
}

// systemWorkspaceKeys: keys compuestas de la tabla de reenvío, ordenadas (para
// agregarlas a getSistemas/getSelectableSistemas — validación de endpoints + CLI).
func systemWorkspaceKeys() []string {
	out := make([]string, 0, len(systemWorkspaces))
	for k := range systemWorkspaces {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// soleSistemaIn: el único sistema navegable de un workspace (auto-discover). "" si
// hay 0 o >1 (ambiguo → en el futuro se declara cuál explícito).
func soleSistemaIn(workspace string) string {
	ss := selectableSistemasIn(workspace)
	if len(ss) == 1 {
		return ss[0]
	}
	return ""
}

// boardDirFor: dónde vive el docs/product de (workspace, sistema). platform → raíz.
func boardDirFor(workspace, sistema string) string {
	if sistema == platformSlug {
		return filepath.Join(workspace, "docs", "product")
	}
	return filepath.Join(workspace, sistema, "docs", "product")
}

// ── Modo single: root único (paridad TS) ────────────────────────────────────

func getWorkspaceRoot() (string, error) {
	rootOnce.Do(func() {
		if env := os.Getenv("WORKSPACE_ROOT"); env != "" {
			abs, err := filepath.Abs(env)
			if err == nil {
				if st, statErr := os.Stat(abs); statErr == nil && st.IsDir() {
					cachedRoot = abs
					return
				}
			}
			rootErr = fmt.Errorf("WORKSPACE_ROOT no existe en disco: %s", env)
			return
		}

		if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
			root := strings.TrimSpace(string(out))
			if root != "" {
				if _, statErr := os.Stat(root); statErr == nil {
					cachedRoot = root
					return
				}
			}
		}

		dir, _ := os.Getwd()
		for dir != filepath.Dir(dir) {
			if fileExists(filepath.Join(dir, "pnpm-workspace.yaml")) || fileExists(filepath.Join(dir, ".git")) {
				cachedRoot = dir
				return
			}
			dir = filepath.Dir(dir)
		}

		rootErr = fmt.Errorf("no se pudo resolver WORKSPACE_ROOT · setea env var WORKSPACE_ROOT o ejecuta desde dentro de un git repo")
	})
	return cachedRoot, rootErr
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// workspaceRoots: todos los roots en juego (1 en single, N en multi).
func workspaceRoots() []string {
	if isMultiMode() {
		var roots []string
		for _, p := range multiProjects {
			roots = append(roots, p.Path)
		}
		return roots
	}
	root, err := getWorkspaceRoot()
	if err != nil {
		return nil
	}
	return []string{root}
}

func isBootstrappedSistema(root, slug string) bool {
	return dirExists(filepath.Join(root, slug, "docs", "product"))
}

// ── Sistemas por root (lógica original) ───────────────────────────────────────

func sistemasIn(root string) []string {
	var fromSeam []string
	for _, slug := range seamSistemaSlugs(root) {
		if isValidSlug(slug) && isBootstrappedSistema(root, slug) {
			fromSeam = append(fromSeam, slug)
		}
	}
	seen := map[string]bool{}
	for _, s := range fromSeam {
		seen[s] = true
	}

	var discovered []string
	entries, err := os.ReadDir(root)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if !e.IsDir() || strings.HasPrefix(name, ".") || !isValidSlug(name) || seen[name] {
				continue
			}
			if isBootstrappedSistema(root, name) {
				discovered = append(discovered, name)
			}
		}
		sort.Strings(discovered)
	}

	return append(fromSeam, discovered...)
}

func selectableSistemasIn(root string) []string {
	real := sistemasIn(root)
	if dirExists(filepath.Join(root, "docs", "product")) {
		return append(real, platformSlug)
	}
	return real
}

// ── API pública (composite-aware) ───────────────────────────────────────────

// getSistemas: sistemas reales. Multi → "{proyecto}/{sistema}".
func getSistemas() []string {
	if !isMultiMode() {
		root, err := getWorkspaceRoot()
		if err != nil {
			return nil
		}
		return sistemasIn(root)
	}
	var out []string
	for _, p := range multiProjects {
		for _, b := range sistemasIn(p.Path) {
			out = append(out, p.Name+"/"+b)
		}
	}
	out = append(out, systemWorkspaceKeys()...) // Fase 3: sistemas con workspace externo
	return out
}

// getSelectableSistemas: reales + platform. Multi → compuestas por proyecto.
func getSelectableSistemas() []string {
	if !isMultiMode() {
		root, err := getWorkspaceRoot()
		if err != nil {
			return nil
		}
		return selectableSistemasIn(root)
	}
	var out []string
	for _, p := range multiProjects {
		for _, b := range selectableSistemasIn(p.Path) {
			out = append(out, p.Name+"/"+b)
		}
	}
	out = append(out, systemWorkspaceKeys()...) // Fase 3: sistemas con workspace externo
	return out
}

// ── escrituras sobre el pseudo-sistema platform (vendored ArnesIA · DA-2) ───
//
// Upstream, `platform` es SOLO-LECTURA: getSelectableSistemas() lo incluye para ver, y
// getSistemas() lo excluye para escribir. Eso asume que el workspace tiene sistemas
// reales (subdirs con docs/product) y que la raíz es un agregado transversal.
//
// harness-studio NO tiene sistemas reales: su `docs/product/` vive en la raíz, así que
// TODO su contenido es `platform`. Con la regla upstream el board sería un museo — un
// drag idea→refining daría 400 en handlers_stories.go:271.
//
// El flag `-allow-platform-writes` (o COCKPIT_ALLOW_PLATFORM_WRITES=1) habilita las
// escrituras sobre el pseudo-sistema. Es opt-in a propósito: sin él el binario se
// comporta EXACTO como upstream, y el drift queda acotado a esta función.
var flagAllowPlatformWrites bool

// writableSistemas: los sistemas donde el operador PUEDE escribir desde el cockpit.
// Default = getSistemas() (paridad upstream). Con el flag = getSelectableSistemas()
// (incluye `platform`).
func writableSistemas() []string {
	if flagAllowPlatformWrites || os.Getenv("COCKPIT_ALLOW_PLATFORM_WRITES") == "1" {
		return getSelectableSistemas()
	}
	return getSistemas()
}

// defaultSistemaFromEnv: sistema inicial preferido (DEFAULT_SISTEMA por-worktree, vía
// cockpit-up.sh). "cross-sistema" (hub) o vacío → sin preferencia. Compartido por
// /api/sistemas y /api/portfolio (una sola lectura del env).
func defaultSistemaFromEnv() string {
	if env := os.Getenv("DEFAULT_SISTEMA"); env != "" && env != "cross-sistema" {
		return env
	}
	return ""
}

func containsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// resolveSistema: sistema (simple o compuesto) → (root del workspace, sistema local).
func resolveSistema(sistema string) (string, string, error) {
	if isMultiMode() {
		// Fase 3: ¿sistema con workspace externo? La tabla de reenvío gana.
		if sw, ok := systemWorkspaces[sistema]; ok {
			return sw.root, sw.sistema, nil
		}
		parts := strings.SplitN(sistema, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("sistema inválido (esperado proyecto/sistema): %s", sistema)
		}
		for _, p := range multiProjects {
			if p.Name == parts[0] {
				return p.Path, parts[1], nil
			}
		}
		return "", "", fmt.Errorf("proyecto desconocido: %s", parts[0])
	}
	root, err := getWorkspaceRoot()
	if err != nil {
		return "", "", err
	}
	return root, sistema, nil
}

// sistemaDocsRoot: {root}/{sistema}/docs · platform → {root}/docs.
func sistemaDocsRoot(sistema string) string {
	root, local, err := resolveSistema(sistema)
	if err != nil {
		return ""
	}
	if local == platformSlug {
		return filepath.Join(root, "docs")
	}
	return filepath.Join(root, local, "docs")
}

func sistemaPath(sistema string) (string, error) {
	_, local, err := resolveSistema(sistema)
	if err != nil {
		return "", err
	}
	if local != platformSlug && !isValidSlug(local) {
		return "", fmt.Errorf("sistema inválido: %s", sistema)
	}
	p := filepath.Join(sistemaDocsRoot(sistema), "product")
	if !dirExists(p) {
		return "", fmt.Errorf("contexto %s no bootstrapeado · falta %s", sistema, p)
	}
	return p, nil
}

func storiesPath(sistema string) string {
	p, _ := sistemaPath(sistema)
	return filepath.Join(p, "stories")
}

func capabilitiesPath(sistema string) string {
	p, _ := sistemaPath(sistema)
	return filepath.Join(p, "capabilities")
}

func releasesPath(sistema string) string {
	p, _ := sistemaPath(sistema)
	return filepath.Join(p, "releases")
}

func learningsPath(sistema string) string {
	return filepath.Join(sistemaDocsRoot(sistema), "learnings")
}

func archiveRootPath(sistema string) string {
	return filepath.Join(sistemaDocsRoot(sistema), "archive")
}

func archivePath(sistema string, year string) string {
	return filepath.Join(sistemaDocsRoot(sistema), "archive", year, "stories")
}

// ── seam (project.config.yaml por root) ─────────────────────────────────────

const fillSentinel = "__FILL_ME__"

func readSeam(root string) map[string]any {
	raw, err := os.ReadFile(filepath.Join(root, "project.config.yaml"))
	if err != nil {
		return nil
	}
	m, err := parseYAMLMap(raw)
	if err != nil {
		return nil
	}
	return m
}

func seamSistemaSlugs(root string) []string {
	seam := readSeam(root)
	if seam == nil {
		return nil
	}
	// el seam (project.config.yaml) puede declarar la key nueva `sistemas:` (I-52) o la
	// legacy `brands:` (seams pre-rename) — el cockpit lee ambas (el loader python igual).
	sistemas, _ := seam["sistemas"].(map[string]any)
	if sistemas == nil {
		sistemas, _ = seam["brands"].(map[string]any)
	}
	active, _ := sistemas["active"].([]any)
	var out []string
	for _, item := range active {
		entry, _ := item.(map[string]any)
		if slug, ok := entry["slug"].(string); ok && slug != "" && slug != fillSentinel {
			out = append(out, slug)
		}
	}
	return out
}

// getValueStreamStages: slot value_stream del seam DEL ROOT de el sistema.
func getValueStreamStages(sistema string) any {
	root, local, err := resolveSistema(sistema)
	if err != nil {
		return nil
	}
	seam := readSeam(root)
	if seam == nil {
		return nil
	}
	vs, ok := seam["value_stream"]
	if !ok || vs == nil {
		return nil
	}
	if _, isStr := vs.(string); isStr {
		return nil
	}
	if bySistema, isMap := vs.(map[string]any); isMap {
		if stages, found := bySistema[local]; found {
			return stages
		}
		return nil
	}
	return vs
}
