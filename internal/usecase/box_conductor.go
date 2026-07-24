package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// BoxConductor runs a T3 box's autonomous loop. It is the realization of
// docs/architecture/boundaries/orquestacion-determinista-entre-cajas.md: the Go conductor OWNS the
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
	// EscribioAlgo — ≥1 control_request se auto-resolvió Allow durante el run (deuda
	// BACKLOG «run async + gate post-run», 2026-07-23): drives el gate post-run (RunService
	// corre `conformance --arnes` solo si esto es true — un run 100% de lectura no lo necesita).
	EscribioAlgo bool
}

// defaultMaxTurns is the last-resort turn cap: a T3 run NEVER spawns uncapped
// (boundary permisos-gui `max-turns-siempre`), even if the config left it 0.
const defaultMaxTurns = 40

// PrecondicionError reports the unmet `necesita` of a box BEFORE any spawn (RF-120):
// the run did not start, no tokens were burned. Faltantes lists what is missing.
type PrecondicionError struct {
	Box       string
	Faltantes []string
}

func (e *PrecondicionError) Error() string {
	return "precondición incumplida en la caja " + e.Box + ": falta " + strings.Join(e.Faltantes, "; ")
}

// Run drives the box loop to a terminal state and returns the routing decision, with
// the conductor's default spawn parameters (no cwd, no permission-set, no insumos).
func (c *BoxConductor) Run(ctx context.Context, box domain.Box) (BoxOutcome, error) {
	return c.RunWith(ctx, box, nil, ports.SpawnOpts{})
}

// RunWith drives the box loop with explicit spawn parameters — the daemon path (Fase E):
// opts carries the arnés cwd (confinement), the role-derived Permisos and the doctrine
// Injection; insumos are the box's graph-resolved inputs (D7). MaxTurns is ALWAYS
// enforced: an unset cap falls back to the conductor's, then to defaultMaxTurns —
// never unbounded.
func (c *BoxConductor) RunWith(ctx context.Context, box domain.Box, insumos []domain.Insumo, opts ports.SpawnOpts) (BoxOutcome, error) {
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = c.maxTurns
	}
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = defaultMaxTurns
	}
	// Precondición D7 (RF-120): todo `necesita` requerido con path resoluble debe
	// existir bajo el cwd ANTES de spawnear — faltante = el run no arranca, cero
	// tokens. Los orígenes externos (usuario/terceros/base) no se statean: su control
	// es admisión de la skill consumidora (D10). requerido:false jamás bloquea.
	if faltantes := c.precondiciones(ctx, insumos, opts.Cwd); len(faltantes) > 0 {
		return BoxOutcome{}, &PrecondicionError{Box: box.ID, Faltantes: faltantes}
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
	escribioAlgo := false
	for iters < c.repairCap && !estado.EsTerminal() {
		if err := sess.Send(ctx, c.tarea(ctx, box, insumos, opts.Cwd, iters)); err != nil {
			return BoxOutcome{}, fmt.Errorf("conductor: send: %w", err)
		}
		iters++
		res, escribio, ok := awaitResult(ctx, sess, opts.Permisos, sess.Events())
		if escribio {
			escribioAlgo = true
		}
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
		EscribioAlgo: escribioAlgo,
		Estado:       estado,
		Iteraciones:  iters,
		Siguiente:    domain.RutaSiguiente(box.Contract, estado, senal),
		Handoff:      estado == domain.CajaBlocked,
		Senal:        senal,
		Advertencias: advertencias,
	}, nil
}

// tarea builds the per-iteration prompt. Iteration 0 = the box's task + the insumos
// block (D7: RUTAS + digest/frontmatter — jamás el documento entero, RF-121); later =
// a repair nudge. The content is not a control signal — routing never depends on it.
func (c *BoxConductor) tarea(ctx context.Context, box domain.Box, insumos []domain.Insumo, cwd string, iter int) string {
	if iter > 0 || box.Contract == nil || box.Contract.Why == "" {
		return "Continúa la caja " + box.ID + " (iteración de reparación " + strconv.Itoa(iter) + ")."
	}
	var b strings.Builder
	b.WriteString("Tarea de la caja " + box.ID + ": " + box.Contract.Why)
	if bloque := c.bloqueInsumos(ctx, insumos, cwd); bloque != "" {
		b.WriteString("\n\nInsumos (lee cada uno desde su ruta bajo el directorio de trabajo):\n")
		b.WriteString(bloque)
	}
	return b.String()
}

// bloqueInsumos renders the deterministic context of the box's inputs: one line per
// insumo (ruta u origen) + its Resumen (digest sidecar o frontmatter). Economía de
// contexto (p11): citations, never the document body.
func (c *BoxConductor) bloqueInsumos(ctx context.Context, insumos []domain.Insumo, cwd string) string {
	var b strings.Builder
	for _, in := range insumos {
		b.WriteString("- " + in.Art)
		if in.Path != "" {
			b.WriteString(" → ruta: " + in.Path)
		} else {
			b.WriteString(" ← " + in.De)
		}
		if !in.Requerido {
			b.WriteString(" (opcional)")
		}
		b.WriteString("\n")
		if in.Path == "" || cwd == "" {
			continue
		}
		if resumen, err := c.artifacts.Resumen(ctx, cwd, in.Path); err == nil && resumen != "" {
			b.WriteString("  resumen:\n  " + strings.ReplaceAll(resumen, "\n", "\n  ") + "\n")
		}
	}
	return b.String()
}

// precondiciones stats every REQUIRED insumo with a resolvable path under cwd; the
// missing ones are returned (RF-120). Without a cwd there is no confined tree to stat —
// nothing to verify (the daemon path always sets it).
func (c *BoxConductor) precondiciones(ctx context.Context, insumos []domain.Insumo, cwd string) []string {
	if cwd == "" {
		return nil
	}
	var faltantes []string
	for _, in := range insumos {
		if !in.Requerido || in.Path == "" {
			continue
		}
		_, exists, err := c.artifacts.Status(ctx, cwd, in.Path)
		switch {
		case err != nil:
			faltantes = append(faltantes, in.Art+" ("+in.Path+"): "+err.Error())
		case !exists:
			faltantes = append(faltantes, in.Art+" ("+in.Path+") de "+in.De)
		}
	}
	return faltantes
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
// awaitResult drena eventos hasta uno terminal (result/error), auto-resolviendo cualquier
// control_request contra el PermissionSet del rol — deuda BACKLOG «run async + gate
// post-run» (2026-07-23): un box run es autónomo (agencia-dentro-del-frame,
// orquestacion-determinista-entre-cajas), no hay humano en el loop para el click de
// permisos-gui. Allow del rol ⇒ auto-allow; Deny/Ask ⇒ auto-deny (Ask no tiene a quién
// preguntarle en un run headless — deny-by-default: nunca cuelga en silencio esperando un
// control_response que nadie manda, nunca auto-aprueba lo que el rol no listó explícito).
// escribio reporta si ≥1 control_request se auto-permitió (proxy honesto de "el run tocó
// algo" — permissionArgs rutea TODO tool de escritura por acá sin excepción, y todo tool de
// lectura pre-aprobado nunca llega a esta función).
func awaitResult(
	ctx context.Context, sess ports.AgentSession, ps domain.PermissionSet, ch <-chan ports.AgentEvent,
) (ev ports.AgentEvent, escribio bool, ok bool) {
	for {
		select {
		case <-ctx.Done():
			return ports.AgentEvent{}, escribio, false
		case e, chOK := <-ch:
			if !chOK {
				return ports.AgentEvent{}, escribio, false
			}
			if e.Kind == ports.EventResult || e.Kind == ports.EventError {
				return e, escribio, true
			}
			if e.Kind == ports.EventControlRequest {
				allow := ps.Decide(e.Tool) == domain.DecisionAllow
				if allow {
					escribio = true
				}
				cd := ports.ControlDecision{Allow: allow, UpdatedInput: e.Input, ToolUseID: e.ToolUseID}
				if !allow {
					cd.Message = "run autónomo (T3): sin humano en el loop — denegado por el permission-set del rol"
				}
				if respErr := sess.RespondControl(ctx, e.RequestID, cd); respErr != nil {
					slog.Error("conductor: responder control_request en run autónomo", "err", respErr)
				}
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
