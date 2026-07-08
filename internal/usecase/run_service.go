package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// runEventType is the SSE event name for run lifecycle frames. It mirrors
// sse.EventRun; kept as a literal so this package needs no transport import (same
// pattern as dockEventType — arch/boundaries/dominio-independiente-de-transporte.md).
const runEventType = "run"

// Sentinels the transport maps to status codes (404 / 422).
var (
	// ErrArnesNoIndexado — the harness id resolves to no indexed graph.
	ErrArnesNoIndexado = errors.New("el arnés no está indexado")
	// ErrNodoNoExiste — the harness is indexed but has no such node.
	ErrNodoNoExiste = errors.New("el nodo no existe en el grafo del arnés")
	// ErrNoEsCaja — the node carries no caja contract; only a process box runs T3.
	ErrNoEsCaja = errors.New("el nodo no es una caja con contrato — solo una caja de proceso corre T3")
)

// runFrame is one run-lifecycle payload published on the multiplexed SSE stream
// (event: run) — the Corridas surface (S8) consumes it.
type runFrame struct {
	RunID       string `json:"run_id"`
	Harness     string `json:"harness"`
	Box         string `json:"box"`
	Kind        string `json:"kind"` // started | finished | error
	Rol         string `json:"rol,omitempty"`
	Estado      string `json:"estado,omitempty"`
	Siguiente   string `json:"siguiente,omitempty"`
	Handoff     bool   `json:"handoff,omitempty"`
	Iteraciones int    `json:"iteraciones,omitempty"`
	Detalle     string `json:"detalle,omitempty"`
}

// RunResult is the HTTP-facing outcome of one T3 box run.
type RunResult struct {
	RunID       string            `json:"run_id"`
	Box         string            `json:"box"`
	Rol         string            `json:"rol,omitempty"`
	Estado      domain.EstadoCaja `json:"estado"`
	Iteraciones int               `json:"iteraciones"`
	Siguiente   string            `json:"siguiente,omitempty"`
	Handoff     bool              `json:"handoff"`
	// Advertencias — lecturas no fatales del run (RF-111): p.ej. un error al leer el
	// status del artefacto. Visibles, jamás descartadas en silencio.
	Advertencias []string `json:"advertencias,omitempty"`
}

// RunService is the daemon entry to the T3 conductor (D2: POST /harnesses/{id}/boxes/
// {boxId}/run). It realizes the doctrine wiring end to end: rol del arnés (META de
// enganche) → PermissionPort → spawn confinado al cwd del arnés con --max-turns +
// Permisos + Injection → loop determinista del BoxConductor → eventos `run` por SSE.
type RunService struct {
	index     ports.IndexPort
	conductor *BoxConductor
	perms     ports.PermissionPort
	resolver  ports.WorkdirResolver
	injector  ports.InjectionProvisioner // nil = corre sin doctrina (degradación honesta).
	pub       EventPublisher

	mu  sync.Mutex
	seq int
}

// NewRunService wires the run entry point. injector may be nil.
func NewRunService(index ports.IndexPort, conductor *BoxConductor, perms ports.PermissionPort, resolver ports.WorkdirResolver, injector ports.InjectionProvisioner, pub EventPublisher) *RunService {
	return &RunService{index: index, conductor: conductor, perms: perms, resolver: resolver, injector: injector, pub: pub}
}

// RunBox runs one box's T3 loop to a terminal state, SYNCHRONOUSLY. Decisión: la firma
// real de BoxConductor.Run es bloqueante y acotada (repair-cap × --max-turns), así que
// el resultado ES la respuesta HTTP (el error de spawn le llega al caller, no se pierde
// en un 202); el progreso vivo viaja en paralelo por SSE (event: run).
func (s *RunService) RunBox(ctx context.Context, harnessID, boxID string) (RunResult, error) {
	g, err := s.index.Query(ctx, harnessID)
	if err != nil {
		return RunResult{}, fmt.Errorf("%w: %q (%w)", ErrArnesNoIndexado, harnessID, err)
	}
	box, ok := g.NodeByID(boxID)
	if !ok {
		return RunResult{}, fmt.Errorf("%w: %q en %q", ErrNodoNoExiste, boxID, harnessID)
	}
	if !box.IsCaja() {
		return RunResult{}, fmt.Errorf("%w (nodo %q)", ErrNoEsCaja, boxID)
	}

	// El rol viene de la META de enganche del grafo indexado (arnes.rol) — el
	// permission-set deriva del rol, jamás de un default (permisos-derivan-del-rol).
	rol := ""
	if g.Arnes != nil {
		rol = g.Arnes.Rol
	}
	ps, err := s.perms.ResolveForRole(ctx, rol)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolver permisos del rol: %w", err)
	}

	// Confinamiento: el spawn corre en el árbol del arnés (sesion-aislada-por-cwd).
	cwd, _, err := s.resolver.Resolve(harnessID)
	if err != nil {
		return RunResult{}, fmt.Errorf("resolver workdir del arnés %q: %w", harnessID, err)
	}

	// Inyección de doctrina: un fallo degrada honesto (corre sin doctrina + warn),
	// jamás bloquea la corrida (principio 6: guía sin bloqueo).
	var inj ports.Injection
	if s.injector != nil {
		if inj, err = s.injector.Provision(ctx); err != nil {
			slog.Warn("run: provisión de doctrina falló — corrida sin inyección", "err", err)
			inj = ports.Injection{}
		}
	}

	runID := s.nextRunID(harnessID, boxID)
	s.publish(runFrame{RunID: runID, Harness: harnessID, Box: boxID, Kind: "started", Rol: ps.Rol})

	out, err := s.conductor.RunWith(ctx, box, ports.SpawnOpts{
		Cwd:       cwd,
		Permisos:  ps,
		Injection: inj,
		// MaxTurns lo garantiza RunWith (max-turns-siempre): config del conductor o default.
	})
	if err != nil {
		s.publish(runFrame{RunID: runID, Harness: harnessID, Box: boxID, Kind: "error", Rol: ps.Rol, Detalle: err.Error()})
		return RunResult{}, fmt.Errorf("correr caja %q: %w", boxID, err)
	}

	for _, adv := range out.Advertencias {
		slog.Warn("run: advertencia del conductor", "run_id", runID, "box", boxID, "detalle", adv)
	}
	s.publish(runFrame{
		RunID: runID, Harness: harnessID, Box: boxID, Kind: "finished", Rol: ps.Rol,
		Estado: string(out.Estado), Siguiente: out.Siguiente, Handoff: out.Handoff, Iteraciones: out.Iteraciones,
	})
	return RunResult{
		RunID: runID, Box: out.Box, Rol: ps.Rol, Estado: out.Estado,
		Iteraciones: out.Iteraciones, Siguiente: out.Siguiente, Handoff: out.Handoff,
		Advertencias: out.Advertencias,
	}, nil
}

// nextRunID mints a monotonic run id (frames idempotentes por run, como el Dock).
func (s *RunService) nextRunID(harnessID, boxID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	return fmt.Sprintf("%s-%s-r%d", harnessID, boxID, s.seq)
}

// publish marshals a run frame onto the SSE stream (event: run).
func (s *RunService) publish(f runFrame) {
	b, err := json.Marshal(f)
	if err != nil {
		slog.Error("run service: marshal run frame", "err", err)
		return
	}
	s.pub.Publish(runEventType, b)
}
