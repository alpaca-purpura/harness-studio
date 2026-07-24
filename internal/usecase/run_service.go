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
// pattern as dockEventType — docs/architecture/boundaries/dominio-independiente-de-transporte.md).
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
	// Conformance — el gate post-run (deuda BACKLOG «run async + gate post-run»,
	// 2026-07-23): SOLO se corre si el run auto-permitió ≥1 escritura
	// (BoxOutcome.EscribioAlgo) — mismo patrón que CAP-71 (el FE llama a conformance tras
	// un turno de chat con escrituras), acá lo dispara el backend porque un run headless
	// no tiene un turno de FE del que colgarse. nil = no se corrió (run 100% lectura).
	Conformance *domain.ConformanceReport `json:"conformance,omitempty"`
}

// RunStatus is the async-facing record for GET /harnesses/{id}/boxes/{boxId}/runs/{runId}
// (deuda BACKLOG «run async», 2026-07-23): el 202 devuelve solo el run_id; este es el
// «desenlace final» que el poller consulta.
type RunStatus struct {
	RunID  string     `json:"run_id"`
	Estado string     `json:"estado"` // corriendo | terminado | error
	Result *RunResult `json:"result,omitempty"`
	Error  string     `json:"error,omitempty"`
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
	conf      ports.ConformancePort // nil = gate post-run desactivado (degradación honesta).
	confBase  func(id string) string

	mu   sync.Mutex
	seq  int
	runs map[string]*RunStatus
}

// NewRunService wires the run entry point. injector/conf may be nil (degradación honesta:
// sin doctrina / sin gate post-run, jamás bloquea el run).
func NewRunService(index ports.IndexPort, conductor *BoxConductor, perms ports.PermissionPort, resolver ports.WorkdirResolver, injector ports.InjectionProvisioner, pub EventPublisher, conf ports.ConformancePort, confBase func(id string) string) *RunService {
	return &RunService{
		index: index, conductor: conductor, perms: perms, resolver: resolver,
		injector: injector, pub: pub, conf: conf, confBase: confBase,
		runs: map[string]*RunStatus{},
	}
}

// StartRun valida SÍNCRONO (arnés indexado · nodo existe · es caja · permisos del rol ·
// workdir resoluble — todo barato, cero tokens quemados) y devuelve el run_id de
// inmediato; el loop del conductor (spawn + repair-cap × --max-turns, lo que de verdad
// tarda) corre en un goroutine — deuda BACKLOG «run async del /boxes/{id}/run» (2026-07-23,
// diseño del operador): 202+run_id ya, `GET .../runs/{runId}` para el desenlace, progreso
// en paralelo por SSE (event: run, sin cambios). Los errores de VALIDACIÓN siguen
// síncronos (404/422/409 — el caller los ve en la respuesta del POST, nunca escondidos
// detrás de un 202 que después falla en silencio).
func (s *RunService) StartRun(ctx context.Context, harnessID, boxID string) (runID string, err error) {
	g, err := s.index.Query(ctx, harnessID)
	if err != nil {
		return "", fmt.Errorf("%w: %q (%w)", ErrArnesNoIndexado, harnessID, err)
	}
	box, ok := g.NodeByID(boxID)
	if !ok {
		return "", fmt.Errorf("%w: %q en %q", ErrNodoNoExiste, boxID, harnessID)
	}
	if !box.IsCaja() {
		return "", fmt.Errorf("%w (nodo %q)", ErrNoEsCaja, boxID)
	}

	// El rol viene de la META de enganche del grafo indexado (arnes.rol) — el
	// permission-set deriva del rol, jamás de un default (permisos-derivan-del-rol).
	rol := ""
	if g.Arnes != nil {
		rol = g.Arnes.Rol
	}
	ps, err := s.perms.ResolveForRole(ctx, rol)
	if err != nil {
		return "", fmt.Errorf("resolver permisos del rol: %w", err)
	}

	// Confinamiento: el spawn corre en el árbol del arnés (sesion-aislada-por-cwd). Se
	// valida acá (barato, cero tokens); runAsync lo vuelve a resolver — no vale la pena
	// cargarlo por el canal del goroutine por un valor que casi nunca cambia entremedio.
	if _, _, err := s.resolver.Resolve(harnessID); err != nil {
		return "", fmt.Errorf("resolver workdir del arnés %q: %w", harnessID, err)
	}

	runID = s.nextRunID(harnessID, boxID)
	s.mu.Lock()
	s.runs[runID] = &RunStatus{RunID: runID, Estado: "corriendo"}
	s.mu.Unlock()
	s.publish(runFrame{RunID: runID, Harness: harnessID, Box: boxID, Kind: "started", Rol: ps.Rol})

	// El ctx del request HTTP muere con la respuesta 202 — el run necesita su propio ctx,
	// independiente del caller (sigue vivo aunque el cliente se desconecte).
	go s.runAsync(context.WithoutCancel(ctx), runID, harnessID, boxID, g, box, ps)

	return runID, nil
}

// GetRun devuelve el estado/desenlace de un run ya arrancado (ok=false si el run_id no
// existe — 404 en el transporte).
func (s *RunService) GetRun(runID string) (RunStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.runs[runID]
	if !ok {
		return RunStatus{}, false
	}
	cp := *st // copia: el caller no debe mutar el registro vivo.
	return cp, true
}

// runAsync ejecuta el loop del conductor (la parte lenta) y publica el desenlace tanto al
// registro (GET .../runs/{runId}) como al stream SSE (event: run, kind=finished/error) —
// exactamente el mismo evento que RunBox publicaba antes de esta deuda, sin cambio de
// contrato para quien ya escuchaba /events.
func (s *RunService) runAsync(ctx context.Context, runID, harnessID, boxID string, g domain.Graph, box domain.Box, ps domain.PermissionSet) {
	// Inyección de doctrina: un fallo degrada honesto (corre sin doctrina + warn),
	// jamás bloquea la corrida (principio 6: guía sin bloqueo).
	var inj ports.Injection
	if s.injector != nil {
		if provisioned, ierr := s.injector.Provision(ctx); ierr != nil {
			slog.Warn("run: provisión de doctrina falló — corrida sin inyección", "err", ierr)
		} else {
			inj = provisioned
		}
	}

	cwd, _, err := s.resolver.Resolve(harnessID)
	if err != nil { // ya se resolvió en StartRun; solo puede fallar por una carrera rarísima.
		s.finishError(runID, harnessID, boxID, ps.Rol, fmt.Errorf("resolver workdir del arnés %q: %w", harnessID, err))
		return
	}

	// Insumos resueltos contra el grafo (D7): el conductor statea los requeridos antes
	// de spawnear y cita rutas+digest en tarea() — jamás el documento entero.
	insumos := domain.InsumosDe(g, box)

	out, err := s.conductor.RunWith(ctx, box, insumos, ports.SpawnOpts{
		Cwd:       cwd,
		Permisos:  ps,
		Injection: inj,
		// MaxTurns lo garantiza RunWith (max-turns-siempre): config del conductor o default.
	})
	if err != nil {
		var pre *PrecondicionError
		if errors.As(err, &pre) {
			s.finishError(runID, harnessID, boxID, ps.Rol, fmt.Errorf("caja %q: %w", boxID, pre))
			return
		}
		s.finishError(runID, harnessID, boxID, ps.Rol, fmt.Errorf("correr caja %q: %w", boxID, err))
		return
	}

	for _, adv := range out.Advertencias {
		slog.Warn("run: advertencia del conductor", "run_id", runID, "box", boxID, "detalle", adv)
	}

	// Gate post-run (CAP-70/71, mismo patrón que el chat corre tras un turno con
	// escrituras — acá lo dispara el backend, un run headless no tiene turno de FE del
	// que colgarse). Solo si el run auto-permitió algo Y hay ConformancePort cableado
	// (degradación honesta: sin conf, el run igual termina — no se bloquea el desenlace
	// por un gate que no puede correr).
	var rep *domain.ConformanceReport
	if out.EscribioAlgo && s.conf != nil && s.confBase != nil {
		raw, merr := json.Marshal(g)
		if merr != nil {
			slog.Error("run: gate post-run — marshal del grafo", "run_id", runID, "err", merr)
		} else if r, cerr := s.conf.RunGraph(ctx, raw, s.confBase(harnessID)); cerr != nil {
			slog.Error("run: gate post-run — conformance falló", "run_id", runID, "err", cerr)
		} else {
			rep = &r
		}
	}

	res := RunResult{
		RunID: runID, Box: out.Box, Rol: ps.Rol, Estado: out.Estado,
		Iteraciones: out.Iteraciones, Siguiente: out.Siguiente, Handoff: out.Handoff,
		Advertencias: out.Advertencias, Conformance: rep,
	}
	s.mu.Lock()
	s.runs[runID] = &RunStatus{RunID: runID, Estado: "terminado", Result: &res}
	s.mu.Unlock()
	s.publish(runFrame{
		RunID: runID, Harness: harnessID, Box: boxID, Kind: "finished", Rol: ps.Rol,
		Estado: string(out.Estado), Siguiente: out.Siguiente, Handoff: out.Handoff, Iteraciones: out.Iteraciones,
	})
}

// finishError registra + publica un run terminado en error — un solo punto de salida
// para que el registro y el SSE nunca queden en "corriendo" para siempre.
func (s *RunService) finishError(runID, harnessID, boxID, rol string, err error) {
	s.mu.Lock()
	s.runs[runID] = &RunStatus{RunID: runID, Estado: "error", Error: err.Error()}
	s.mu.Unlock()
	s.publish(runFrame{RunID: runID, Harness: harnessID, Box: boxID, Kind: "error", Rol: rol, Detalle: err.Error()})
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
