package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// BoxConductor runs a T3 box's autonomous loop. It is the realization of
// arch/boundaries/orquestacion-determinista-entre-cajas.md: the Go conductor OWNS the
// control flow — it spawns `claude` with a hard --max-turns, applies an explicit repair
// cap (double red), reads the `result` subtype + the artifact `status` (document-as-
// cache) to advance the internal FSM, terminates in `blocked`→handoff, and executes
// contract.ruta to pick the next box. The LLM never decides the hand-off; code does.
type BoxConductor struct {
	agent     ports.AgentPort
	artifacts ports.ArtifactReader
	repairCap int // explicit iteration cap, on top of --max-turns.
	maxTurns  int // per-iteration --max-turns.
}

// NewBoxConductor wires the conductor. repairCap ≥ 1; maxTurns caps each iteration.
func NewBoxConductor(agent ports.AgentPort, artifacts ports.ArtifactReader, repairCap, maxTurns int) *BoxConductor {
	if repairCap < 1 {
		repairCap = 1
	}
	return &BoxConductor{agent: agent, artifacts: artifacts, repairCap: repairCap, maxTurns: maxTurns}
}

// BoxOutcome is the deterministic result of running a box: the terminal internal state,
// how many iterations it took, and where CODE routed next (the next box or a handoff).
type BoxOutcome struct {
	Box         string
	Estado      domain.EstadoCaja
	Iteraciones int
	Siguiente   string // next box id / handoff target, from contract.ruta — decided by code.
	Handoff     bool   // true when Estado=blocked → escalation.
	Senal       domain.SenalIteracion
	// Advertencias — non-fatal reads that used to be silent (RF-111): an artifact
	// status error does not break the loop (empty status stays valid) but is VISIBLE.
	Advertencias []string
}

// defaultMaxTurns is the last-resort turn cap: a T3 run NEVER spawns uncapped
// (boundary permisos-gui `max-turns-siempre`), even if the config left it 0.
const defaultMaxTurns = 40

// Run drives the box loop to a terminal state and returns the routing decision, with
// the conductor's default spawn parameters (no cwd, no permission-set).
func (c *BoxConductor) Run(ctx context.Context, box domain.Box) (BoxOutcome, error) {
	return c.RunWith(ctx, box, ports.SpawnOpts{})
}

// RunWith drives the box loop with explicit spawn parameters — the daemon path (Fase E):
// opts carries the arnés cwd (confinement), the role-derived Permisos and the doctrine
// Injection. MaxTurns is ALWAYS enforced: an unset cap falls back to the conductor's,
// then to defaultMaxTurns — never unbounded.
func (c *BoxConductor) RunWith(ctx context.Context, box domain.Box, opts ports.SpawnOpts) (BoxOutcome, error) {
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = c.maxTurns
	}
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = defaultMaxTurns
	}
	sess, err := c.agent.Spawn(ctx, opts)
	if err != nil {
		return BoxOutcome{}, fmt.Errorf("conductor: spawn: %w", err)
	}
	defer func() { _ = sess.Close() }()

	estado := domain.CajaDraft
	var senal domain.SenalIteracion
	var advertencias []string
	artifact := artifactRef(box)

	iters := 0
	for iters < c.repairCap && !estado.EsTerminal() {
		if err := sess.Send(ctx, c.tarea(box, iters)); err != nil {
			return BoxOutcome{}, fmt.Errorf("conductor: send: %w", err)
		}
		iters++
		res, ok := awaitResult(ctx, sess.Events())
		if !ok {
			estado = domain.CajaBlocked // stream died mid-turn: stop, do not spin.
			break
		}
		// Read ONLY machine signals: the result subtype + the artifact's document-as-cache
		// status. The chat text (res.Text) is deliberately ignored for control flow. The
		// read is confined under the run's cwd (the arnés tree the spawn ran in).
		status, _, statusErr := c.artifacts.Status(ctx, opts.Cwd, artifact)
		if statusErr != nil {
			// RF-111: no rompe el loop (status vacío sigue siendo válido) pero jamás
			// se descarta en silencio — viaja al resultado del run.
			advertencias = append(advertencias,
				fmt.Sprintf("iteración %d: leer status de %q: %v", iters, artifact, statusErr))
		}
		senal = domain.SenalIteracion{Subtipo: mapSubtipo(res), EstadoArtefacto: status}
		estado = domain.AvanzarCaja(estado, senal)
	}
	if !estado.EsTerminal() {
		// Repair cap exhausted without done/blocked → block (bounded-error, not whack-a-mole).
		estado = domain.CajaBlocked
	}

	return BoxOutcome{
		Box:          box.ID,
		Estado:       estado,
		Iteraciones:  iters,
		Siguiente:    domain.RutaSiguiente(box.Contract, estado, senal),
		Handoff:      estado == domain.CajaBlocked,
		Senal:        senal,
		Advertencias: advertencias,
	}, nil
}

// tarea builds the per-iteration prompt. Iteration 0 = the box's task; later = a repair
// nudge. The content is not a control signal — routing never depends on it.
func (c *BoxConductor) tarea(box domain.Box, iter int) string {
	if box.Contract != nil && box.Contract.Why != "" && iter == 0 {
		return "Tarea de la caja " + box.ID + ": " + box.Contract.Why
	}
	return "Continúa la caja " + box.ID + " (iteración de reparación " + strconv.Itoa(iter) + ")."
}

// artifactRef returns the box's primary output artifact (document-as-cache target).
// An explicit `path` (D3, artefacto-archivo) wins; the `art` label is the fallback.
func artifactRef(box domain.Box) string {
	if box.Contract != nil && len(box.Contract.Entrega) > 0 {
		if p := box.Contract.Entrega[0].Path; p != "" {
			return p
		}
		return box.Contract.Entrega[0].Art
	}
	return box.ID
}

// awaitResult reads events until the turn's `result` (or an error/closed stream). It
// consumes the machine signal, never the chat text.
func awaitResult(ctx context.Context, ch <-chan ports.AgentEvent) (ports.AgentEvent, bool) {
	for {
		select {
		case <-ctx.Done():
			return ports.AgentEvent{}, false
		case ev, ok := <-ch:
			if !ok {
				return ports.AgentEvent{}, false
			}
			if ev.Kind == ports.EventResult || ev.Kind == ports.EventError {
				return ev, true
			}
		}
	}
}

// mapSubtipo maps a normalized result event to the domain result subtype.
func mapSubtipo(ev ports.AgentEvent) domain.ResultadoSubtipo {
	if ev.Kind == ports.EventError {
		return domain.SubtipoError
	}
	s := strings.ToLower(ev.Subtype)
	switch {
	case strings.Contains(s, "max") || strings.Contains(s, "limit"):
		return domain.SubtipoLimite
	case strings.Contains(s, "error"):
		return domain.SubtipoError
	default:
		return domain.SubtipoOK
	}
}
