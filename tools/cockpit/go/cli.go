// cli.go — subcomandos: start/stop/status/add/remove/list/auto-add/story search.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const cliHelp = `cockpit — Cockpit SDD standalone (filesystem-as-DB)

Uso:
  cockpit [start] [-port 4000] [-workspace PATH] [-d]
      start         server (default). Sin -workspace usa el registry
                    ~/.cockpit/cockpit.yaml (multi-workspace, sistemas proyecto/sistema).
      -d            daemon (background · PID en ~/.cockpit/cockpit.pid · logs en ~/.cockpit/logs/)

  cockpit stop                  detiene el daemon
  cockpit status                estado del daemon + registry
  cockpit add PATH [-name N]    registra un workspace en el registry
  cockpit remove NAME           desregistra (no borra archivos)
  cockpit auto-add DIR          escanea DIR/* y registra todo workspace SDD
  cockpit list                  proyectos registrados + sistemas + stories
  cockpit story search TEXTO    búsqueda global de stories (id/título/goal)
  cockpit version
`

func runCLI() {
	args := os.Args[1:]
	cmd := "start"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cmd = args[0]
		args = args[1:]
	}

	switch cmd {
	case "start", "serve":
		cmdStart(args)
	case "stop":
		cmdStop()
	case "status":
		cmdStatus()
	case "add":
		cmdAdd(args)
	case "remove":
		cmdRemove(args)
	case "auto-add":
		cmdAutoAdd(args)
	case "list":
		cmdList()
	case "story":
		if len(args) >= 2 && args[0] == "search" {
			cmdStorySearch(strings.Join(args[1:], " "))
			return
		}
		fmt.Fprintln(os.Stderr, "uso: cockpit story search TEXTO")
		os.Exit(2)
	case "version":
		fmt.Println("cockpit-go " + version)
	case "help", "-h", "--help":
		fmt.Print(cliHelp)
	default:
		fmt.Fprintf(os.Stderr, "subcomando desconocido: %s\n\n%s", cmd, cliHelp)
		os.Exit(2)
	}
}

// ── start (foreground o daemon) ─────────────────────────────────────────────

func cmdStart(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	port := fs.Int("port", 4000, "puerto HTTP")
	workspace := fs.String("workspace", "", "workspace root único (sin registry)")
	daemonize := fs.Bool("d", false, "correr como daemon")
	sidecar := fs.String("sidecar", "", "dir del sidecar de delivery (default: junto al binario — RN-35)")
	allowPlatform := fs.Bool("allow-platform-writes", false, "habilita escrituras sobre el pseudo-sistema `platform` (workspace cuyo docs/product vive en la raíz)")
	_ = fs.Parse(args)
	flagSidecarDir = *sidecar
	flagAllowPlatformWrites = *allowPlatform

	if *daemonize {
		if pid, running := daemonRunning(); running {
			fmt.Fprintf(os.Stderr, "daemon ya corriendo (pid %d) · cockpit stop primero\n", pid)
			os.Exit(1)
		}
		self, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, "✗", err)
			os.Exit(1)
		}
		if err := os.MkdirAll(filepath.Dir(logPath()), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "✗", err)
			os.Exit(1)
		}
		logFile, err := os.OpenFile(logPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "✗", err)
			os.Exit(1)
		}
		childArgs := []string{"start", "-port", strconv.Itoa(*port)}
		if *workspace != "" {
			childArgs = append(childArgs, "-workspace", *workspace)
		}
		if *sidecar != "" {
			childArgs = append(childArgs, "-sidecar", *sidecar)
		}
		// El hijo es un proceso NUEVO: sin propagar el flag, `start -d
		// -allow-platform-writes` arrancaría un daemon read-only y el drag del board
		// fallaría con 400 sin explicación.
		if *allowPlatform {
			childArgs = append(childArgs, "-allow-platform-writes")
		}
		childPid, err := spawnDaemon(self, childArgs, logFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "✗", err)
			os.Exit(1)
		}
		_ = writeFileAtomic(pidPath(), strconv.Itoa(childPid)+"\n")
		time.Sleep(500 * time.Millisecond)
		if !pidAlive(childPid) {
			fmt.Fprintf(os.Stderr, "✗ daemon murió al arrancar · revisa %s\n", logPath())
			os.Exit(1)
		}
		fmt.Printf("✓ daemon corriendo (pid %d) · http://localhost:%d · logs: %s\n",
			childPid, *port, logPath())
		return
	}

	serve(*port, *workspace)
}

func cmdStop() {
	pid, running := daemonRunning()
	if !running {
		fmt.Println("daemon no está corriendo")
		_ = os.Remove(pidPath())
		return
	}
	if err := terminatePID(pid); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
	_ = os.Remove(pidPath())
	fmt.Printf("✓ daemon detenido (pid %d)\n", pid)
}

func cmdStatus() {
	if pid, running := daemonRunning(); running {
		fmt.Printf("daemon: ✓ corriendo (pid %d)\n", pid)
	} else {
		fmt.Println("daemon: ✗ detenido")
	}
	reg := loadRegistry()
	active := reg.activeProjects()
	fmt.Printf("registry: %s (%d proyectos, %d activos)\n", registryPath(), len(reg.Projects), len(active))
	for _, p := range active {
		fmt.Printf("  · %-22s %s (%d sistemas)\n", p.Name, p.Path, len(selectableSistemasIn(p.Path)))
	}
}

// ── registry mgmt ───────────────────────────────────────────────────────────

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	name := fs.String("name", "", "nombre del proyecto (default: basename del path)")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "uso: cockpit add PATH [-name N]")
		os.Exit(2)
	}
	reg := loadRegistry()
	proj, err := reg.add(fs.Arg(0), *name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s → %s (%d sistemas)\n", proj.Name, proj.Path, len(selectableSistemasIn(proj.Path)))
}

func cmdRemove(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "uso: cockpit remove NAME")
		os.Exit(2)
	}
	reg := loadRegistry()
	var kept []registryProject
	found := false
	for _, p := range reg.Projects {
		if p.Name == args[0] {
			found = true
			continue
		}
		kept = append(kept, p)
	}
	if !found {
		fmt.Fprintf(os.Stderr, "✗ proyecto no registrado: %s\n", args[0])
		os.Exit(1)
	}
	reg.Projects = kept
	if err := saveRegistry(reg); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s desregistrado (archivos intactos)\n", args[0])
}

func cmdAutoAdd(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "uso: cockpit auto-add DIR")
		os.Exit(2)
	}
	base, err := filepath.Abs(args[0])
	if err != nil || !dirExists(base) {
		fmt.Fprintf(os.Stderr, "✗ dir inválido: %s\n", args[0])
		os.Exit(1)
	}
	reg := loadRegistry()
	added := 0
	entries, _ := os.ReadDir(base)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		path := filepath.Join(base, e.Name())
		if !isWorkspaceLike(path) {
			continue
		}
		if proj, err := reg.add(path, ""); err == nil {
			fmt.Printf("✓ %s → %s\n", proj.Name, proj.Path)
			added++
		}
	}
	if added == 0 {
		fmt.Println("(nada nuevo para registrar)")
	}
}

func cmdList() {
	reg := loadRegistry()
	if len(reg.Projects) == 0 {
		fmt.Println("registry vacío · cockpit add PATH")
		return
	}
	for _, p := range reg.Projects {
		state := "✓"
		if !p.Active {
			state = "·"
		}
		if !dirExists(p.Path) {
			state = "✗(path no existe)"
		}
		fmt.Printf("%s %-22s %s\n", state, p.Name, p.Path)
		if dirExists(p.Path) {
			for _, b := range selectableSistemasIn(p.Path) {
				n := 0
				if entries, err := os.ReadDir(filepath.Join(sistemaDocsRootIn(p.Path, b), "product", "stories")); err == nil {
					for _, s := range entries {
						if s.IsDir() {
							n++
						}
					}
				}
				fmt.Printf("    %-20s %d stories\n", b, n)
			}
		}
	}
}

// sistemaDocsRootIn: como sistemaDocsRoot pero con root explícito (CLI, sin modo global).
func sistemaDocsRootIn(root, sistema string) string {
	if sistema == platformSlug {
		return filepath.Join(root, "docs")
	}
	return filepath.Join(root, sistema, "docs")
}

// ── story search global ─────────────────────────────────────────────────────

func cmdStorySearch(query string) {
	q := strings.ToLower(query)
	reg := loadRegistry()
	projects := reg.activeProjects()
	if len(projects) == 0 {
		// fallback: workspace actual
		if root, err := getWorkspaceRoot(); err == nil {
			projects = []registryProject{{Name: filepath.Base(root), Path: root, Active: true}}
		}
	}
	if len(projects) == 0 {
		fmt.Fprintln(os.Stderr, "✗ sin proyectos (cockpit add PATH) ni workspace actual")
		os.Exit(1)
	}

	matches := 0
	for _, proj := range projects {
		for _, sistema := range selectableSistemasIn(proj.Path) {
			base := filepath.Join(sistemaDocsRootIn(proj.Path, sistema), "product", "stories")
			dirs := []string{base}
			archBase := filepath.Join(sistemaDocsRootIn(proj.Path, sistema), "archive")
			if years, err := os.ReadDir(archBase); err == nil {
				for _, y := range years {
					if y.IsDir() {
						dirs = append(dirs, filepath.Join(archBase, y.Name(), "stories"))
					}
				}
			}
			for _, dir := range dirs {
				entries, err := os.ReadDir(dir)
				if err != nil {
					continue
				}
				for _, e := range entries {
					if !e.IsDir() {
						continue
					}
					doc, err := readMarkdownWithFrontmatter(filepath.Join(dir, e.Name(), "checkpoint.md"))
					if err != nil {
						continue
					}
					id, _ := doc.Frontmatter["story_id"].(string)
					if id == "" {
						id = e.Name()
					}
					title, _ := doc.Frontmatter["title"].(string)
					goal, _ := doc.Frontmatter["goal"].(string)
					state, _ := doc.Frontmatter["state"].(string)
					haystack := strings.ToLower(id + " " + title + " " + goal)
					if !strings.Contains(haystack, q) {
						continue
					}
					matches++
					label := title
					if label == "" {
						label = goal
					}
					fmt.Printf("%-30s %-12s %-12s %s\n", proj.Name+"/"+sistema, id, state, label)
				}
			}
		}
	}
	if matches == 0 {
		fmt.Printf("(sin matches para %q)\n", query)
	}
}
