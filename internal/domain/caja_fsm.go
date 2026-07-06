package domain

import "strings"

// This file models the INTERNAL FSM of a running box — the *adaptation* state machine
// (draft→…→blocked) that lives DENTRO de una caja, distinct from the arnés-level work
// *spine* (evolution, cross-caja). It is a doctrine axis (a product constant, unlike the
// per-arnés spine states). The conductor advances it by reading MACHINE signals — the
// `result` event subtype + the artifact `status` (document-as-cache) — never the chat
// text (boundary orquestacion-determinista-entre-cajas, DAOP-A2 / 12-Factor).

// EstadoCaja is the internal state of a box's run loop.
type EstadoCaja string

const (
	// CajaDraft — the loop has not produced a first artifact yet.
	CajaDraft EstadoCaja = "draft"
	// CajaWorking — the box is iterating (repair loop).
	CajaWorking EstadoCaja = "working"
	// CajaReview — an artifact exists and is being gated.
	CajaReview EstadoCaja = "review"
	// CajaDone — terminal success: the artifact passed its gate.
	CajaDone EstadoCaja = "done"
	// CajaBlocked — terminal stop: repair cap hit, an error, or an explicit block → handoff.
	CajaBlocked EstadoCaja = "blocked"
)

// EsTerminal reports whether the box loop has ended.
func (e EstadoCaja) EsTerminal() bool { return e == CajaDone || e == CajaBlocked }

// ResultadoSubtipo is the machine-readable outcome of one conductor iteration, taken from
// the stream-json `result` frame subtype — never inferred from the chat text.
type ResultadoSubtipo string

const (
	// SubtipoOK — the turn completed within its budget.
	SubtipoOK ResultadoSubtipo = "success"
	// SubtipoLimite — the turn hit --max-turns (a bounded, recoverable stop).
	SubtipoLimite ResultadoSubtipo = "max-turns"
	// SubtipoError — the turn errored.
	SubtipoError ResultadoSubtipo = "error"
)

// SenalIteracion is the machine signal the conductor reads after an iteration: the result
// subtype plus the artifact's document-as-cache `status`. The chat text is deliberately
// absent — the conductor must not scrape it.
type SenalIteracion struct {
	Subtipo ResultadoSubtipo
	// EstadoArtefacto is the `status:` read from the box's output artifact (document-as-
	// cache). Agnostic string; the doctrine reads a small set of terminal markers.
	EstadoArtefacto string
}

// artifact status markers the conductor recognizes (document-as-cache convention). These
// are OUTCOME markers of the box's own artifact, not the arnés spine states.
const (
	ArtefactoDone    = "done"
	ArtefactoBlocked = "blocked"
)

// AvanzarCaja is the pure FSM transition: given the current internal state and the
// machine signal, it returns the next state. It reads ONLY the signal (result subtype +
// artifact status), never chat text — that is what makes routing deterministic and owned
// by code. An error subtype or an explicit `blocked` artifact stops the loop; a `done`
// artifact promotes to review→done; otherwise the box keeps working (repair).
func AvanzarCaja(actual EstadoCaja, s SenalIteracion) EstadoCaja {
	if actual.EsTerminal() {
		return actual
	}
	switch {
	case s.Subtipo == SubtipoError, s.EstadoArtefacto == ArtefactoBlocked:
		return CajaBlocked
	case s.EstadoArtefacto == ArtefactoDone:
		// The artifact declares itself complete: gate it and finish.
		return CajaDone
	default:
		// Not done and not blocked → keep iterating (max-turns is a recoverable pause).
		return CajaWorking
	}
}

// RutaSiguiente picks the next hand-off target of a box deterministically from its
// contract.ruta — CODE reads the DAG, not the LLM. On a blocked terminal it returns the
// handoff target; on done it takes the first route whose `si` matches the outcome signal,
// else the happy path (the first route with no `si`). Returns "" when nothing routes.
func RutaSiguiente(c *Contract, estado EstadoCaja, senal SenalIteracion) string {
	if c == nil {
		return ""
	}
	if estado == CajaBlocked {
		if c.Handoff != nil && c.Handoff.A != "" {
			return c.Handoff.A
		}
		return ""
	}
	if estado != CajaDone {
		return ""
	}
	var happy string
	for _, r := range c.Ruta {
		if r.Si == "" {
			if happy == "" {
				happy = r.A
			}
			continue
		}
		if rutaCoincide(r.Si, senal) {
			return r.A
		}
	}
	return happy
}

// rutaCoincide reports whether a route's `si` condition matches the outcome signal. The
// condition language is intentionally minimal (substring against the artifact status /
// subtype) — enough to prove code owns the branch without a DSL. Extensible later.
func rutaCoincide(si string, senal SenalIteracion) bool {
	return containsFold(si, senal.EstadoArtefacto) || containsFold(si, string(senal.Subtipo))
}

// containsFold reports whether sub occurs in s, case-insensitively (sub non-empty).
func containsFold(s, sub string) bool {
	return sub != "" && strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
