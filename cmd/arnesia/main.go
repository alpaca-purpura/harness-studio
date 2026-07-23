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
	"github.com/alpacapurpura/arnesia/internal/adapters/artifact"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/mechanism"
	"github.com/alpacapurpura/arnesia/internal/adapters/conformance/ruleset"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/adapters/loader"
	"github.com/alpacapurpura/arnesia/internal/adapters/permission"
	"github.com/alpacapurpura/arnesia/internal/adapters/portafolio"
	"github.com/alpacapurpura/arnesia/internal/adapters/provision"
	"github.com/alpacapurpura/arnesia/internal/adapters/publish"
	"github.com/alpacapurpura/arnesia/internal/adapters/selfupdate"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	httpapi "github.com/alpacapurpura/arnesia/internal/adapters/transport/http"
	"github.com/alpacapurpura/arnesia/internal/adapters/transport/sse"
	"github.com/alpacapurpura/arnesia/internal/adapters/watch"
	"github.com/alpacapurpura/arnesia/internal/domain"
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
	case "portafolio":
		err = runPortafolio(os.Args[2:])
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
  serve     watcher + index + HTTP/SSE API + UI embebida on :4200 (dev sin dist: solo API; el bundle la trae)
  open      open the UI (stub)
  index     load an arnés directory into a graph.l0 (nomenclatura-arnes.md)
  publish   publish a harness to its marketplace repo (stub)
  conformance  run the ruleset against an element or an arnés (METODOLOGIA §6)
  portafolio   escanear/listar/agregar/desvincular arneses del Portafolio (Slice 0)
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
	repairCap := fs.Int("repair-cap", 3, "iteraciones máximas de reparación de una caja T3 (BoxConductor)")
	authToken := fs.String("auth-token", os.Getenv("ARNESIA_AUTH_TOKEN"),
		"capability token required on the API (default $ARNESIA_AUTH_TOKEN; empty = Host+Origin only, dev)")
	repo := fs.String("repo", os.Getenv("ARNESIA_REPO"),
		"ruta del repo para el self-update (default $ARNESIA_REPO; vacío = botón Actualizar deshabilitado)")
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
	// loadArnesDir: la vía viva del loader real (HS-11) — un directorio registrado se
	// reconoce archivo-por-archivo (nomenclatura v1) y entra al índice. Falla honesto
	// (el registro queda; el índice no inventa).
	loadArnesDir := func(id, path string) error {
		g, err := loader.LoadArnes(path)
		if err != nil {
			return fmt.Errorf("cargar arnés %s desde %s: %w", id, path, err)
		}
		return idx.Upsert(ctx, g)
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
	// Al boot, los arneses ya registrados se re-cargan del disco al índice (el índice
	// es desechable; el directorio es la verdad).
	for _, ap := range arnesReg.List() {
		if lerr := loadArnesDir(ap.Arnes, ap.Path); lerr != nil {
			slog.Warn("boot: arnés registrado no indexable aún", "arnes", ap.Arnes, "err", lerr)
		}
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
	schemaFS, err := iofs.Sub(doctrina.Files, "docs/architecture/contracts/schema")
	if err != nil {
		return fmt.Errorf("schemas embebidos: %w", err)
	}
	confSvc := usecase.NewConformanceService("", ruleset.NewFromFS(doctrina.Files),
		mechanism.NewSchemaSetFS(schemaFS), []ports.MechanismAdapter{
			mechanism.NLJudge{}, mechanism.StaticScan{}, mechanism.SchemaAdapter{},
		})

	// Permisos por rol (Fase E): el KitProvisioner resuelve el permission-set del rol
	// que hidrata (permisos-derivan-del-rol); las sesiones lo usan al responder un
	// control_request y el conductor T3 lo materializa en flags CC-native al spawn.
	perms := permission.NewKitProvisioner()

	// El rol de una sesión ES el del arnés que hidrata (graph.l0 META, decisión #6 del
	// paquete chat-cc-funcional): el índice lo conoce; el FE jamás elige autoridad.
	roleFor := func(rctx context.Context, arnesID string) string {
		g, gerr := mapSvc.Graph(rctx, arnesID)
		if gerr != nil || g.Arnes == nil {
			return ""
		}
		return g.Arnes.Rol
	}

	sessionSvc, err := usecase.NewSessionService(ctx, agent, sessionStore, brokerPublisher{broker}, arnesReg, *maxTurns, injector, perms, roleFor)
	if err != nil {
		return fmt.Errorf("session service: %w", err)
	}
	// Reindex-tras-turno (RF-184): el Mapa refleja lo que el chat edita, turno a turno.
	sessionSvc.SetReindexer(usecase.NewTurnReindexer(idx, loader.LoadArnes))

	// Conductor T3 (Fase E): el loop determinista de una caja, con el lector de
	// `status:` del artefacto (document-as-cache) confinado al árbol del arnés.
	conductor := usecase.NewBoxConductor(agent, artifact.NewReader(), *repairCap, *maxTurns)
	runSvc := usecase.NewRunService(idx, conductor, perms, arnesReg, injector, brokerPublisher{broker})
	// Fuente del nodo (RF-93, tab Contenido del inspector): lectura confinada al dir
	// registrado del arnés (S2) — el drawer muestra el archivo REAL, jamás reconstruye.
	fuenteSvc := usecase.NewFuenteService(idx, arnesReg, artifact.NewFuenteReader())
	// baseFor: el dir del arnés registrado resuelve los fuente_path relativos del
	// firewall; un arnés no registrado (fixtures embebidos) usa el repo si existe.
	confBase := func(id string) string {
		for _, ap := range arnesReg.List() {
			if ap.Arnes == id {
				return ap.Path
			}
		}
		if r, rerr := findRepoRoot(); rerr == nil {
			return r
		}
		return ""
	}

	// Self-update sin sudo (paquete boton-actualizar): el repo llega por flag/env —
	// JAMÁS del request (RF-106); sin repo la tarjeta lo dice y el botón queda disabled.
	// Bugfix fix-repo-self-update (RF-108): si no vino por flag/env, se intenta el
	// último configurado vía UI (persistido en ~/.arnesia/self-update.json) — así una
	// instalación empaquetada (Tauri/.deb, sin --repo en su sidecar) puede actualizar
	// tras configurarlo UNA vez desde Ajustes, sin volver a tocar CLI.
	repoStore, err := selfupdate.NewRepoStore("")
	if err != nil {
		return fmt.Errorf("self-update: repo store: %w", err)
	}
	repoInicial := *repo
	if repoInicial == "" {
		if persistido, perr := repoStore.Leer(); perr == nil {
			repoInicial = persistido
		} else {
			slog.Warn("self-update: repo store ilegible al boot — arranca sin repo", "err", perr)
		}
	}
	updater, err := selfupdate.New(repoInicial)
	if err != nil {
		return fmt.Errorf("self-update: %w", err)
	}
	updSvc := usecase.NewSelfUpdateService(updater, repoStore)

	// Portafolio de arneses (Slice 0, HS-22): store separado de arneses.json (A1) + walker
	// físico READ-ONLY + evaluador de deriva local + wrapper del loader real. Mismo wiring
	// que usa el subcomando `portafolio` — newPortafolioService lo factoriza para no
	// duplicarlo.
	portafolioSvc, err := newPortafolioService(idx)
	if err != nil {
		return err
	}

	handler := httpapi.NewHandler(mapSvc, sessionSvc, runSvc, fuenteSvc, arnesReg, confSvc, confBase, loadArnesDir, updSvc, portafolioSvc, embeddedUI(), broker, httpapi.AuthConfigFor(*addr, *authToken))

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

// arnesLoaderFunc adapta una función libre (loader.LoadArnes) a ports.ArnesLoader — el
// usecase del Portafolio no puede importar el paquete loader (go-arch-lint), así que cmd
// cablea la función concreta detrás del puerto.
type arnesLoaderFunc func(dir string) (domain.Graph, error)

func (f arnesLoaderFunc) Load(dir string) (domain.Graph, error) { return f(dir) }

// embeddedUI returns the SPA handler when this build carries web/dist (scripts/
// bundle.sh la compila antes del daemon), or nil for an honest dev build without UI.
// Fallback SPA: cualquier ruta sin archivo sirve index.html (la app navega por
// hash-state, sin router — pero los deep-links no deben 404).
func embeddedUI() http.Handler {
	ui, err := iofs.Sub(doctrina.WebDist, "web/dist")
	if err != nil {
		return nil
	}
	if _, err := iofs.Stat(ui, "index.html"); err != nil {
		return nil // build de dev: el dist embebido solo trae .gitkeep.
	}
	files := http.FileServerFS(ui)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" && p != "index.html" {
			if _, err := iofs.Stat(ui, p); err == nil {
				files.ServeHTTP(w, r)
				return
			}
		}
		// index.html (raíz, explícito o fallback SPA) va sin caché: tras un self-update
		// el WebView del shell y el browser deben recoger los assets nuevos (hasheados
		// por Vite) en la próxima carga, no cuando el heurístico de caché quiera.
		w.Header().Set("Cache-Control", "no-cache")
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		files.ServeHTTP(w, r2)
	})
}

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
          se reconoce archivo-por-archivo según docs/architecture/contracts/nomenclatura-arnes.md
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

// newPortafolioService cablea el usecase del Portafolio con sus adapters por default
// (store en ~/.arnesia/portafolio.json, Scanner/Referencias con CCPluginsDir/MaxDepth
// default) — compartido entre `serve` y el subcomando `portafolio`, sin duplicar wiring.
// indice es el 5° puerto (S1-D1, Observar en Mapa): `serve` pasa el `idx` real que ya
// construyó; el subcomando CLI pasa nil — no necesita indexar (ObservarEnMapa con índice
// nil da el error honesto «requiere el daemon», jamás un nil-pointer panic).
func newPortafolioService(indice ports.IndexPort) (*usecase.PortafolioService, error) {
	st, err := portafolio.NewStore("")
	if err != nil {
		return nil, fmt.Errorf("portafolio store: %w", err)
	}
	return usecase.NewPortafolioService(st, &portafolio.Scanner{}, arnesLoaderFunc(loader.LoadArnes), &portafolio.Referencias{}, indice), nil
}

// runPortafolio es la vía de verificación E2E del Portafolio sin FE (S0-D9): reusa el
// MISMO usecase que el HTTP — cero lógica propia acá.
func runPortafolio(args []string) error {
	fs := flag.NewFlagSet("portafolio", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: arnesia portafolio <subcomando> [args]

  escanear <dir>            escanea dir (abs o relativo a cwd) y emite los candidatos crudos (JSON); NO persiste.
  listar                    emite las entradas del registro + las corruptas visibles (JSON).
  agregar <dir> <clave>...  re-escanea dir y persiste SOLO los candidatos con esas claves.
  desvincular <clave>       quita la entrada del registro; no borra nada de disco.
`)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return errors.New("arnesia portafolio: falta el subcomando")
	}

	// nil: el CLI no observa en Mapa (esa vía es HTTP-only, S1-D1) — no necesita el 5° puerto.
	svc, err := newPortafolioService(nil)
	if err != nil {
		return err
	}
	ctx := context.Background()

	switch sub, rest := fs.Arg(0), fs.Args()[1:]; sub {
	case "escanear":
		if len(rest) < 1 {
			return errors.New("uso: arnesia portafolio escanear <dir>")
		}
		dir, aerr := filepath.Abs(rest[0])
		if aerr != nil {
			return aerr
		}
		cands, serr := svc.Escanear(ctx, dir)
		if serr != nil {
			return serr
		}
		return imprimirJSON(cands)

	case "listar":
		sanas, corruptas, lerr := svc.Listar(ctx)
		if lerr != nil {
			return lerr
		}
		return imprimirJSON(struct {
			Entradas  []domain.EntradaPortafolio `json:"entradas"`
			Corruptas []domain.EntradaCorrupta   `json:"corruptas,omitempty"`
		}{sanas, corruptas})

	case "agregar":
		if len(rest) < 2 {
			return errors.New("uso: arnesia portafolio agregar <dir> <clave> [<clave>...]")
		}
		dir, aerr := filepath.Abs(rest[0])
		if aerr != nil {
			return aerr
		}
		persistidas, perr := svc.AgregarProyecto(ctx, dir, rest[1:])
		if perr != nil {
			return perr
		}
		return imprimirJSON(persistidas)

	case "desvincular":
		if len(rest) < 1 {
			return errors.New("uso: arnesia portafolio desvincular <clave>")
		}
		ok, derr := svc.Desvincular(ctx, rest[0])
		if derr != nil {
			return derr
		}
		return imprimirJSON(map[string]bool{"desvinculado": ok})

	default:
		fs.Usage()
		return fmt.Errorf("arnesia portafolio: subcomando desconocido %q", sub)
	}
}

// imprimirJSON emite v indentado a stdout — la forma común de las 4 salidas de
// `arnesia portafolio`.
func imprimirJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = os.Stdout.Write(b)
	return err
}
