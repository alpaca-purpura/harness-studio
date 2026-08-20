// registry.go — ~/.cockpit/: registry de workspaces (cockpit.yaml), PID file, logs.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// directorioMeta / sistemaMeta / servicioCompartido: copia LOCAL (Stage 4 · CK-07)
// del shape que products/cockpit/go/types.go define para estos mismos campos —
// devhub deja de importar el paquete cockpit (dos binarios independientes), pero
// SIGUE siendo dueño de `cockpit add`/`remove`/`auto-add`, que reescriben el
// registry ENTERO en cada save (saveRegistry marshalea reg.Projects completo):
// sin este shape exacto, cualquier add/remove futuro borraría silenciosamente
// kind/sistemas/repo de TODOS los demás proyectos. Mismo criterio de duplicación
// consciente que el glue de CK-05/06 — red: TestRegistryWireFormatRoundTrip.
type directorioMeta struct {
	Kind     string        `yaml:"kind,omitempty" json:"kind,omitempty"`
	Sistemas []sistemaMeta `yaml:"sistemas,omitempty" json:"sistemas,omitempty"`
	Repo     string        `yaml:"repo,omitempty" json:"repo,omitempty"`
}

type sistemaMeta struct {
	Slug        string `yaml:"slug" json:"slug"`
	Procedencia string `yaml:"procedencia" json:"procedencia"`
	Ref         string `yaml:"ref,omitempty" json:"ref,omitempty"`
	Descripcion string `yaml:"descripcion,omitempty" json:"descripcion,omitempty"`
	Workspace   string `yaml:"workspace,omitempty" json:"workspace,omitempty"`
}

type servicioCompartido struct {
	Slug         string   `yaml:"slug" json:"slug"`
	Nombre       string   `yaml:"nombre,omitempty" json:"nombre,omitempty"`
	ConsumidoPor []string `yaml:"consumido_por,omitempty" json:"consumido_por,omitempty"`
}

type registryProject struct {
	// Genéricos (workspace/multi-mode) — devhub-owned.
	Name   string `yaml:"name" json:"name"`
	Path   string `yaml:"path" json:"path"`
	Active bool   `yaml:"active" json:"active"`
	// Torre de control (F2 · DH-04/D2) — campos ADITIVOS del wire format que emite
	// products/devhub/scripts/gen_registro.py desde el registro curado (RN-06): display
	// de la fila + comando del gate de fábrica por sistema. Flat a nivel del proyecto;
	// deben sobrevivir el round-trip de saveRegistry — red: TestRegistryWireFormatRoundTrip.
	Nombre    string `yaml:"nombre,omitempty" json:"nombre,omitempty"`
	GateCheck string `yaml:"gate_check,omitempty" json:"gate_check,omitempty"`
	// Directorio-específicos (I-36 · vista de directorio) — cockpit-owned desde
	// CK-02 Stage 2. Named (no embed): cada call site marca explícito que toca
	// dato directorio-owned. `,inline` preserva el wire format YAML plano
	// (kind/sistemas/repo a nivel del proyecto) — red: TestRegistryWireFormatRoundTrip.
	Directorio directorioMeta `yaml:",inline"`
}

type registry struct {
	Projects             []registryProject   `yaml:"projects"`
	ServiciosCompartidos []servicioCompartido `yaml:"servicios_compartidos,omitempty"`
}

func cockpitHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".cockpit"
	}
	return filepath.Join(home, ".cockpit")
}

func registryPath() string { return filepath.Join(cockpitHome(), "cockpit.yaml") }
func pidPath() string      { return filepath.Join(cockpitHome(), "cockpit.pid") }
func logPath() string      { return filepath.Join(cockpitHome(), "logs", "cockpit.log") }

func loadRegistry() *registry {
	raw, err := os.ReadFile(registryPath())
	if err != nil {
		return &registry{}
	}
	var reg registry
	if err := yaml.Unmarshal(raw, &reg); err != nil {
		return &registry{}
	}
	return &reg
}

func saveRegistry(reg *registry) error {
	if err := os.MkdirAll(cockpitHome(), 0o755); err != nil {
		return err
	}
	out, err := yaml.Marshal(reg)
	if err != nil {
		return err
	}
	return writeFileAtomic(registryPath(), string(out))
}

func (reg *registry) activeProjects() []registryProject {
	var out []registryProject
	for _, p := range reg.Projects {
		if p.Active && dirExists(p.Path) {
			out = append(out, p)
		}
	}
	return out
}

func (reg *registry) find(name string) *registryProject {
	for i := range reg.Projects {
		if reg.Projects[i].Name == name {
			return &reg.Projects[i]
		}
	}
	return nil
}

// projectNameFor sanea un nombre a slug seguro y resuelve colisiones con sufijo.
func (reg *registry) projectNameFor(path, explicit string) string {
	name := explicit
	if name == "" {
		name = filepath.Base(path)
	}
	name = strings.ToLower(name)
	name = regexpNonSlug.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "project"
	}
	base := name
	for i := 2; reg.find(name) != nil; i++ {
		name = base + "-" + strconv.Itoa(i)
	}
	return name
}

// isWorkspaceLike: tiene docs/product/ propio o algún subdir {slug}/docs/product/.
func isWorkspaceLike(path string) bool {
	if dirExists(filepath.Join(path, "docs", "product")) {
		return true
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") &&
			dirExists(filepath.Join(path, e.Name(), "docs", "product")) {
			return true
		}
	}
	return false
}

func (reg *registry) add(path, explicitName string) (registryProject, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return registryProject{}, err
	}
	if !dirExists(abs) {
		return registryProject{}, fmt.Errorf("el path no existe: %s", abs)
	}
	if !isWorkspaceLike(abs) {
		return registryProject{}, fmt.Errorf("no parece un workspace SDD (sin docs/product/ ni subdirs {sistema}/docs/product/): %s", abs)
	}
	for _, p := range reg.Projects {
		if p.Path == abs {
			return p, fmt.Errorf("ya registrado como %q", p.Name)
		}
	}
	proj := registryProject{Name: reg.projectNameFor(abs, explicitName), Path: abs, Active: true}
	reg.Projects = append(reg.Projects, proj)
	return proj, saveRegistry(reg)
}

// ── PID helpers (daemon) ────────────────────────────────────────────────────

func readDaemonPID() int {
	raw, err := os.ReadFile(pidPath())
	if err != nil {
		return 0
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
	return pid
}

func daemonRunning() (int, bool) {
	pid := readDaemonPID()
	if pid > 0 && pidAlive(pid) {
		return pid, true
	}
	return pid, false
}
