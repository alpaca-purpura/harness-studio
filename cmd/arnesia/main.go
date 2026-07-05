// Command arnesia is the ArnesIA daemon. It is the composition root: the only place
// that wires concrete adapters (agent, index, watch, publish) to the use cases and the
// transport. Subcommands: serve | open | index | publish.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/agent/claudecode"
	"github.com/alpacapurpura/arnesia/internal/adapters/index"
	"github.com/alpacapurpura/arnesia/internal/adapters/publish"
	"github.com/alpacapurpura/arnesia/internal/adapters/store"
	httpapi "github.com/alpacapurpura/arnesia/internal/adapters/transport/http"
	"github.com/alpacapurpura/arnesia/internal/adapters/transport/sse"
	"github.com/alpacapurpura/arnesia/internal/adapters/watch"
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
  serve     watcher + index + HTTP/SSE API + embedded UI on :4200
  open      open the UI (stub)
  index     rebuild the disposable index from the JSONL corpus (stub)
  publish   publish a harness to its marketplace repo (stub)
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
	agent := claudecode.New(*claudeBin) // the Dock conductor.

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
	sessionSvc, err := usecase.NewSessionService(ctx, agent, sessionStore, brokerPublisher{broker}, arnesReg, *maxTurns)
	if err != nil {
		return fmt.Errorf("session service: %w", err)
	}
	handler := httpapi.NewHandler(mapSvc, sessionSvc, arnesReg, broker, httpapi.AuthConfigFor(*addr, *authToken))

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
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

// runOpen opens the UI. Stub.
func runOpen(_ []string) error {
	fmt.Println("arnesia open: TODO(fase 5) — open the embedded UI / Tauri shell")
	return nil
}

// runIndex rebuilds the disposable index. Stub over the in-memory store.
func runIndex(_ []string) error {
	idx := index.New()
	if err := idx.Rebuild(context.Background()); err != nil {
		return err
	}
	fmt.Println("arnesia index: rebuilt (in-memory; TODO fase 5: modernc.org/sqlite)")
	return nil
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
