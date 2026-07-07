// Command arnesia is the ArnesIA daemon. It is the composition root: the only place
// that wires concrete adapters (agent, index, watch, publish) to the use cases and the
// transport. Subcommands: serve | open | index | publish.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	iofs "io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	doctrina "github.com/alpacapurpura/arnesia"
	"github.com/alpacapurpura/arnesia/internal/adapters/agent/claudecode"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
	"github.com/alpacapurpura/arnesia/internal/adapters/provision"
	"github.com/alpacapurpura/arnesia/internal/adapters/publish"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	httpapi "github.com/alpacapurpura/arnesia/internal/adapters/transport/http"
	"github.com/alpacapurpura/arnesia/internal/adapters/transport/sse"
	"github.com/alpacapurpura/arnesia/internal/adapters/watch"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch cmd := os.Args[1]; cmd {
	case "serve":
		err = runServe(os.Args[2:])
	case "open":
		err = runOpen(os.Args[2:])
	case "index":
		err = runIndex(os.Args[2:])
	case "publish":
		err = runPublish(os.Args[2:])
	case "conformance":
		err = runConformance(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "arnesia: unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		slog.Error("arnesia", "cmd", os.Args[1], "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `arnesia — la fábrica de arneses (daemon)

usage: arnesia <command> [flags]

commands:
  serve     watcher + index + HTTP/SSE API on :4200 (la UI la sirve Tauri/vite; go:embed de la SPA = deuda)
  open      open the UI (stub)
  index     load an arnés directory into a graph.l0 (nomenclatura-arnes.md)
  publish   publish a harness to its marketplace repo (stub)
  conformance  run the ruleset against an element or an arnés (METODOLOGIA §6)
`)
}

// runServe wires the adapters + use cases + HTTP/SSE transport and listens on :4200.
// It is the only real subcommand in this skeleton.
func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:4200", "listen address")
	claudeBin := fs.String("claude", "claude", "path to the claude binary (the Dock conductor)")
	sessionsPath := fs.String("sessions", "", "session registry file (default ~/.arnesia/sessions.json)")
	arnesesPath := fs.String("arneses", "", "arnés→path registry file (default ~/.arnesia/arneses.json)")
	arnesRoot := fs.String("arnes-root", "", "root for unregistered-arnés fallback dirs (default ~/.arnesia/arneses)")
	maxTurns := fs.Int("max-turns", 40, "cap on the agent loop per turn (--max-turns); 0 disables the cap")
	authToken := fs.String("auth-token", os.Getenv("ARNESIA_AUTH_TOKEN"),
		"capability token required on the API (default $ARNESIA_AUTH_TOKEN; empty = Host+Origin only, dev)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Outbound adapters (concrete — wired only here, the composition root).
	idx := index.New()
	if err := idx.Rebuild(ctx); err != nil {
		return fmt.Errorf("index rebuild: %w", err)
	}
	agent := claudecode.New(resolveClaudeBin(*claudeBin)) // the Dock conductor.

	sessionStore, err := store.NewRegistry(*sessionsPath)
	if err != nil {
		return fmt.Errorf("session store: %w", err)
	}

	// Arnés→path registry: resolves each session's cwd so a conductor is confined to its
	// arnés's tree, never a shared global cwd (boundary permisos-gui `sesion-aislada-por-cwd`).
	arnesReg, err := store.NewArnesRegistry(*arnesesPath, *arnesRoot)
	if err != nil {
		return fmt.Errorf("arnes registry: %w", err)
	}

	watcher := watch.New()
	events, err := watcher.Watch(ctx)
	if err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	// Transport + services.
	broker := sse.NewBroker()
	mapSvc := usecase.NewMapService(idx)
	// Inyección de doctrina (HS-11, puente 2): materializa kit+knowhow embebidos a
	// ~/.arnesia (idempotente por huella) e inyecta --plugin-dir/--append-system-prompt-
	// file/--add-dir a cada spawn. El usuario no configura nada.
	injector, err := provision.New("", doctrina.Kit, doctrina.Files)
	if err != nil {
		return fmt.Errorf("provisioner: %w", err)
	}

	// Conformance del daemon (HS-11, puente 3): SIEMPRE el ruleset/schemas embebidos —
	// el binario instalado se comporta igual que el de dev; el scope `fabrica`
	// (arch-test/go-arch-lint, repoRoot="") difiere honesto, el scope `arnes` (schema +
	// spine + escritor único + firewall) da pass/fail real vía RunGraph.
	schemaFS, err := iofs.Sub(doctrina.Files, "arch/contracts/schema")
	if err != nil {
		return fmt.Errorf("schemas embebidos: %w", err)
	}
	confSvc := usecase.NewConformanceService("", ruleset.NewFromFS(doctrina.Files),
		mechanism.NewSchemaSetFS(schemaFS), []ports.MechanismAdapter{
			mechanism.NLJudge{}, mechanism.StaticScan{}, mechanism.SchemaAdapter{},
		})

	sessionSvc, err := usecase.NewSessionService(ctx, agent, sessionStore, brokerPublisher{broker}, arnesReg, *maxTurns, injector)
	if err != nil {
		return fmt.Errorf("session service: %w", err)
	}
	// baseFor: el dir del arnés registrado resuelve los fuente_path relativos del
	// firewall; un arnés no registrado (fixtures embebidos) usa el repo si existe.
	confBase := func(id string) string {
		for _, ap := range arnesReg.List() {
			if ap.Arnes == id {
				return ap.Path
			}
		}
		if r, err := findRepoRoot(); err == nil {
			return r
		}
		return ""
	}

	handler := httpapi.NewHandler(mapSvc, sessionSvc, arnesReg, confSvc, confBase, broker, httpapi.AuthConfigFor(*addr, *authToken))

	// Filesystem changes drive incremental reindex + a map delta on the SSE bus.
	go func() {
		for range events {
			// TODO(fase 5): idx.Rebuild(ctx) then broker.Publish(sse.EventMap, delta).
		}
	}()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		// WithoutCancel: the parent ctx is already done (that is why we are here); the
		// shutdown needs its own deadline, inheriting values but not the cancellation.
		shutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown", "err", err)
		}
	}()

	slog.Info("arnesia serve", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// brokerPublisher adapts the SSE broker (whose Publish returns the stored event) to
// the usecase.EventPublisher interface (which ignores it), keeping usecase free of any
// transport import.
type brokerPublisher struct{ b *sse.Broker }

func (p brokerPublisher) Publish(eventType string, data []byte) { p.b.Publish(eventType, data) }

// resolveClaudeBin hardens `claude` discovery for GUI launches: una app de escritorio
// Linux (lanzada desde .desktop, no desde una shell) frecuentemente NO lleva
// ~/.local/bin en su PATH — el binario existe pero LookPath no lo ve y el spawn moriría
// silencioso hacia el Dock (hallazgo de la auditoría 2026-07-07). Un nombre pelado que
// PATH no resuelve se sondea en las rutas de instalación conocidas; el error del spawn
// sigue siendo la autoridad final.
func resolveClaudeBin(bin string) string {
	if strings.ContainsRune(bin, os.PathSeparator) {
		return bin // ruta explícita del operador: se respeta tal cual.
	}
	if _, err := exec.LookPath(bin); err == nil {
		return bin
	}
	home, _ := os.UserHomeDir()
	for _, p := range []string{
		filepath.Join(home, ".local", "bin", bin),
		filepath.Join(home, ".claude", "local", bin),
		filepath.Join("/usr/local/bin", bin),
		filepath.Join("/opt/homebrew/bin", bin),
	} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			slog.Info("claude resuelto fuera de PATH (entorno GUI)", "path", p)
			return p
		}
	}
	slog.Warn("claude no está en PATH ni en rutas conocidas — instala Claude Code o pasa --claude", "bin", bin)
	return bin
}

// runOpen opens the UI. Stub.
func runOpen(_ []string) error {
	fmt.Println("arnesia open: TODO(fase 5) — open the embedded UI / Tauri shell")
	return nil
}

// runIndex loads an arnés DIRECTORY into its graph.l0 via the nomenclatura loader
// (HS-11, puente 1: archivo-por-archivo → grafo, nomenclatura-arnes.md v1) and emits
// the graph. Sin argumento, re-seedea el índice in-memory (comportamiento previo).
func runIndex(args []string) error {
	fs := flag.NewFlagSet("index", flag.ExitOnError)
	out := fs.String("o", "", "write the graph.l0 JSON to this file (default: stdout)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia index [-o out.graph.json] [<dir-del-arnés>]   (flags ANTES del dir — semántica flag de Go)

  <dir>   raíz de un arnés (plugin CC con .claude-plugin/, o proyecto con .claude/):
          se reconoce archivo-por-archivo según arch/contracts/nomenclatura-arnes.md
          y se emite su graph.l0 (fuente_path ESTAMPADOS; no-reconocido VISIBLE).
  sin dir: re-seedea el índice in-memory embebido (fixtures dogfood).
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := fs.Arg(0)
	if dir == "" {
		idx := index.New()
		if err := idx.Rebuild(context.Background()); err != nil {
			return err
		}
		fmt.Println("arnesia index: rebuilt (in-memory; el índice SQLite llega con el indexer JSONL)")
		return nil
	}

	g, err := loader.LoadArnes(dir)
	if err != nil {
		return fmt.Errorf("index %s: %w", dir, err)
	}
	b, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if *out != "" {
		if werr := os.WriteFile(*out, b, 0o600); werr != nil {
			return werr
		}
		fmt.Fprintf(os.Stderr, "arnesia index: %s → %s (%d nodos, %d edges)\n", dir, *out, len(g.Nodes), len(g.Edges))
		return nil
	}
	_, err = os.Stdout.Write(b)
	return err
}

// runPublish publishes a harness. Stub over the git publisher.
func runPublish(args []string) error {
	fs := flag.NewFlagSet("publish", flag.ExitOnError)
	harness := fs.String("harness", "", "harness id to publish")
	target := fs.String("target", "", "marketplace repo (remote)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	pub := publish.New("git")
	return pub.Publish(context.Background(), *harness, *target)
}
