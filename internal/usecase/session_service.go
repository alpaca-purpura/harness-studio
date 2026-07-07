package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// dockEventType is the SSE event name multiplexed for the Dock. It mirrors
// sse.EventDock; kept as a literal so this package needs no transport import (see
// arch/boundaries/dominio-independiente-de-transporte.md).
const dockEventType = "dock"

// ErrBusy is returned by Turn when the session's conductor is already generating. A turn
// must not race a live one (boundary sesion-viva-consistente `un-turno-a-la-vez`): two
// turns would interleave on one stdin and mix the assembling buffer. The transport maps
// this to HTTP 409.
var ErrBusy = errors.New("session is streaming — a turn is already in flight")

// EventPublisher is the minimal fan-out this service needs (the SSE broker satisfies
// it via a thin wrapper in the composition root). Declaring it here keeps usecase free
// of any transport dependency.
type EventPublisher interface {
	Publish(eventType string, data []byte)
}

// dockFrame is one Dock payload published to the multiplexed SSE stream. Every frame
// carries session_id so the shell can route N parallel conversations off one connection,
// and run_id so each frame is attributable to a single turn (idempotent apply on the FE —
// boundary sesion-viva-consistente `frames-idempotentes-run-id`).
type dockFrame struct {
	SessionID       string `json:"session_id"`
	RunID           string `json:"run_id,omitempty"`
	Kind            string `json:"kind"` // status|init|delta|message|result|error
	Text            string `json:"text,omitempty"`
	Status          string `json:"status,omitempty"`
	CtxPct          int    `json:"ctx_pct,omitempty"`
	Model           string `json:"model,omitempty"`
	ClaudeSessionID string `json:"claude_session_id,omitempty"`
}

// sessionRuntime pairs a persisted session with its live conductor (nil until the first
// turn spawns/resumes it) and the per-turn streaming state.
type sessionRuntime struct {
	meta       *domain.Session
	live       ports.AgentSession
	runSeq     int
	curRun     string // run id of the in-flight turn; stamped on every frame.
	assembling strings.Builder

	// resume self-heal (boundary sesion-viva-consistente `resume-auto-sana`):
	pendingTurn   string // the in-flight user turn, for resend after a heal.
	wasResume     bool   // this process life was spawned with --resume.
	sawInit       bool   // an init frame arrived this life (resume/spawn succeeded).
	resumeRetried bool   // a heal already happened this turn — do not loop.
}

// SessionService owns the registry of work-fronts and drives their Claude Code
// conductors (multisesión). It is the daemon's source of truth for open sessions; the
// conversation content is rehydrated from Claude Code via --resume.
type SessionService struct {
	mu       sync.Mutex
	rt       map[string]*sessionRuntime
	order    []string // stable creation order for List.
	agent    ports.AgentPort
	store    ports.SessionStore
	pub      EventPublisher
	resolver ports.WorkdirResolver
	baseCtx  context.Context
	maxTurns int
}

// NewSessionService loads the persisted registry and returns a ready service. baseCtx
// bounds every conductor's lifetime (cancel it to stop all sessions on shutdown); resolver
// maps each session's arnés to the working directory its conductor runs in (per-session
// confinement, never a shared cwd); maxTurns caps every turn's agent loop.
func NewSessionService(baseCtx context.Context, agent ports.AgentPort, store ports.SessionStore, pub EventPublisher, resolver ports.WorkdirResolver, maxTurns int) (*SessionService, error) {
	s := &SessionService{
		rt:       map[string]*sessionRuntime{},
		agent:    agent,
		store:    store,
		pub:      pub,
		resolver: resolver,
		baseCtx:  baseCtx,
		maxTurns: maxTurns,
	}
	persisted, err := store.Load(baseCtx)
	if err != nil {
		return nil, fmt.Errorf("session service: load: %w", err)
	}
	if len(persisted) == 0 {
		persisted = seedSessions()
	}
	for i := range persisted {
		m := persisted[i]
		// Processes are gone after a restart: no conductor is live, so nothing is
		// streaming. The conversation resumes via --resume on the next turn.
		m.Status = domain.StatusIdle
		s.order = append(s.order, m.ID)
		s.rt[m.ID] = &sessionRuntime{meta: &m}
	}
	return s, nil
}

// List returns the sessions in stable creation order.
func (s *SessionService) List() []domain.Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.Session, 0, len(s.order))
	for _, id := range s.order {
		if r := s.rt[id]; r != nil {
			out = append(out, *r.meta)
		}
	}
	return out
}

// Get returns one session by id.
func (s *SessionService) Get(id string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.rt[id]; r != nil {
		return *r.meta, true
	}
	return domain.Session{}, false
}

// Create opens a new work-front over arnés and persists it. frente may be empty — it
// is auto-derived from the first user turn.
func (s *SessionService) Create(sess domain.Session) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.ID = newID()
	sess.Status = domain.StatusIdle
	if sess.View == "" {
		sess.View = "Mapa"
	}
	if sess.Frente == "" {
		sess.Frente = "nuevo frente"
	}
	s.order = append(s.order, sess.ID)
	s.rt[sess.ID] = &sessionRuntime{meta: &sess}
	s.persistLocked()
	return sess, nil
}

// Rename sets a session's frente (the editable work-front name).
func (s *SessionService) Rename(id, frente string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rt[id]
	if r == nil {
		return domain.Session{}, errNotFound(id)
	}
	if frente = strings.TrimSpace(frente); frente != "" {
		r.meta.Frente = frente
	}
	s.persistLocked()
	return *r.meta, nil
}

// SetView records the view a session is parked on (Mapa|Diag|…).
func (s *SessionService) SetView(id, view string) (domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rt[id]
	if r == nil {
		return domain.Session{}, errNotFound(id)
	}
	if view != "" {
		r.meta.View = view
	}
	s.persistLocked()
	return *r.meta, nil
}

// Close ends a session: it stops the conductor and drops the registry entry.
func (s *SessionService) Close(id string) error {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return errNotFound(id)
	}
	live := r.live
	delete(s.rt, id)
	for i, oid := range s.order {
		if oid == id {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	s.persistLocked()
	s.mu.Unlock()

	if live != nil {
		_ = live.Close()
	}
	return nil
}

// Turn streams one user message to a session's conductor, spawning (or resuming) it on
// first use. It rejects a turn that would race a live one (ErrBusy). The conductor's
// events are pumped to the Dock asynchronously.
func (s *SessionService) Turn(id, text string) error {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return errNotFound(id)
	}
	// One turn at a time: a second turn while streaming would interleave stdin + assembling.
	if r.meta.Status == domain.StatusStreaming {
		s.mu.Unlock()
		return ErrBusy
	}

	// Auto-derive the front name from the first user message.
	if r.meta.Frente == "" || r.meta.Frente == "nuevo frente" {
		r.meta.Frente = deriveFrente(text)
	}
	r.meta.Conv = append(r.meta.Conv, domain.Turn{Rol: domain.RolUser, Text: text})
	r.meta.Status = domain.StatusStreaming
	r.runSeq++
	runID := fmt.Sprintf("%s-r%d", id, r.runSeq)
	r.curRun = runID
	r.pendingTurn = text
	r.resumeRetried = false
	r.assembling.Reset()

	if r.live == nil {
		if err := s.spawnLocked(id, r); err != nil {
			r.meta.Status = domain.StatusIdle
			s.persistLocked()
			s.mu.Unlock()
			return fmt.Errorf("session service: spawn %s: %w", id, err)
		}
	}
	live := r.live
	s.persistLocked()
	s.mu.Unlock()

	s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "status", Status: string(domain.StatusStreaming)})
	if err := live.Send(s.baseCtx, text); err != nil {
		return fmt.Errorf("session service: send %s: %w", id, err)
	}
	return nil
}

// spawnLocked resolves the session's arnés working directory, spawns a conductor confined
// to it, and starts consuming its events. Caller holds s.mu.
func (s *SessionService) spawnLocked(id string, r *sessionRuntime) error {
	cwd, _, err := s.resolver.Resolve(r.meta.Arnes)
	if err != nil {
		return fmt.Errorf("resolve arnés %q workdir: %w", r.meta.Arnes, err)
	}
	resume := r.meta.ClaudeSessionID
	live, err := s.agent.Spawn(s.baseCtx, ports.SpawnOpts{
		Resume:   resume,
		Model:    r.meta.Model,
		Cwd:      cwd,
		MaxTurns: s.maxTurns,
	})
	if err != nil {
		return err
	}
	r.live = live
	r.wasResume = resume != ""
	r.sawInit = false
	go s.consume(id, live)
	return nil
}

// consume pumps one conductor's normalized events onto the Dock and updates the session's
// live state. Every frame is stamped with the in-flight run id. It runs until the
// conductor's channel closes (or a resume-heal hands off to a fresh consume).
func (s *SessionService) consume(id string, live ports.AgentSession) {
	for ev := range live.Events() {
		switch ev.Kind {
		case ports.EventInit:
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				r.sawInit = true
				r.meta.ClaudeSessionID = ev.ClaudeSessionID
				if ev.Model != "" {
					r.meta.Model = ev.Model
				}
				runID = r.curRun
				s.persistLocked()
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "init", ClaudeSessionID: ev.ClaudeSessionID, Model: ev.Model})

		case ports.EventDelta:
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				r.assembling.WriteString(ev.Text)
				runID = r.curRun
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "delta", Text: ev.Text, Status: string(domain.StatusStreaming)})

		case ports.EventResult:
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				final := strings.TrimSpace(r.assembling.String())
				if final == "" {
					final = ev.Text
				}
				r.meta.Conv = append(r.meta.Conv, domain.Turn{Rol: domain.RolAssistant, Text: final})
				r.meta.Status = domain.StatusIdle
				if ev.CtxPct > 0 {
					r.meta.CtxPct = ev.CtxPct
				}
				r.assembling.Reset()
				r.pendingTurn = ""
				runID = r.curRun
				s.persistLocked()
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "result", Text: ev.Text, CtxPct: ev.CtxPct, Status: string(domain.StatusIdle)})

		case ports.EventError:
			// A resume that never initialized → self-heal by restarting fresh.
			if s.tryHealResume(id, live) {
				return
			}
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				r.meta.Status = domain.StatusIdle
				r.live = nil
				runID = r.curRun
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "error", Text: ev.Text, Status: string(domain.StatusIdle)})

		case ports.EventMessage:
			// Full assistant message: ignored when deltas already assembled the text;
			// kept for protocol completeness (tool-only turns emit no text deltas).
		}
	}

	// Channel closed: the subprocess exited. If it was a failed resume, heal; otherwise
	// drop the live handle so the next turn resumes in a fresh process.
	if s.tryHealResume(id, live) {
		return
	}
	s.mu.Lock()
	if r := s.rt[id]; r != nil && r.live == live {
		r.live = nil
		if r.meta.Status == domain.StatusStreaming {
			r.meta.Status = domain.StatusIdle
		}
	}
	s.mu.Unlock()
}

// tryHealResume detects a failed --resume (the process died or errored before ever
// emitting init) and restarts the session fresh ONCE, resending the in-flight turn. It
// returns true if it took over — the caller must stop, since a new consume goroutine is now
// running for the fresh process. The heal is silent on success (the resent turn yields a
// normal result); only a failed restart surfaces an error frame.
func (s *SessionService) tryHealResume(id string, live ports.AgentSession) bool {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil || r.live != live || !r.wasResume || r.sawInit || r.resumeRetried {
		s.mu.Unlock()
		return false
	}
	slog.Info("session service: resume failed, restarting fresh", "session", id, "stale_cc", r.meta.ClaudeSessionID)
	r.resumeRetried = true
	r.meta.ClaudeSessionID = "" // the id was stale; next spawn starts fresh.
	pending := r.pendingTurn
	if err := s.spawnLocked(id, r); err != nil {
		r.meta.Status = domain.StatusIdle
		r.live = nil
		runID := r.curRun
		s.persistLocked()
		s.mu.Unlock()
		s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "error", Text: "no pude reiniciar la sesión: " + err.Error(), Status: string(domain.StatusIdle)})
		return true
	}
	newLive := r.live
	s.persistLocked()
	s.mu.Unlock()

	if pending != "" {
		if err := newLive.Send(s.baseCtx, pending); err != nil {
			s.publish(dockFrame{SessionID: id, Kind: "error", Text: "reenvío tras reinicio falló: " + err.Error(), Status: string(domain.StatusIdle)})
		}
	}
	return true
}

// publish marshals a Dock frame and fans it out on the SSE stream.
func (s *SessionService) publish(f dockFrame) {
	b, err := json.Marshal(f)
	if err != nil {
		slog.Error("session service: marshal dock frame", "err", err)
		return
	}
	s.pub.Publish(dockEventType, b)
}

// persistLocked snapshots the registry to the store. Caller must hold s.mu.
func (s *SessionService) persistLocked() {
	snap := make([]domain.Session, 0, len(s.order))
	for _, id := range s.order {
		if r := s.rt[id]; r != nil {
			snap = append(snap, *r.meta)
		}
	}
	if err := s.store.Save(s.baseCtx, snap); err != nil {
		slog.Error("session service: persist", "err", err)
	}
}

// deriveFrente turns the first user message into a short work-front label.
func deriveFrente(text string) string {
	text = strings.TrimSpace(strings.Join(strings.Fields(text), " "))
	const maxLen = 48
	if len(text) > maxLen {
		return text[:maxLen] + "…"
	}
	if text == "" {
		return "nuevo frente"
	}
	return text
}

// newID returns a short, collision-resistant session id (distinct from the Claude
// Code session id).
func newID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read never fails on supported platforms; fall back deterministically.
		return "s00000000"
	}
	return "s" + hex.EncodeToString(b[:])
}

func errNotFound(id string) error {
	return fmt.Errorf("session %q not found", id)
}

// seedSessions returns illustrative work-fronts shown on first run (empty registry) so
// the rail is populated, mirroring the signed mockup (it.14). The FIRST session points at
// the real indexed dogfood arnés (dev-full-cycle) with the Mapa view, so the map loads
// immediately on first run; the other two are illustrative (their arnés/company labels are
// placeholders until a real graph is indexed, and their views don't fetch a graph).
func seedSessions() []domain.Session {
	return []domain.Session{
		{ID: newID(), Frente: "ciclo full-cycle · spec→released", Arnes: "dev-full-cycle", Empresa: "alpacapurpura", Puesto: "Ingeniería · Desarrollo full-cycle", Salud: domain.SaludInfo, Status: domain.StatusIdle, View: "Mapa", Parked: "spec-writer"},
		{ID: newID(), Frente: "eval-gate de po-ux", Arnes: "ux-nordia", Empresa: "Nordia", Puesto: "Diseño · UX", Salud: domain.SaludWarn, Status: domain.StatusIdle, View: "Diag", Parked: "hallazgo éxito 87%"},
		{ID: newID(), Frente: "corrida r3 · 3 hallazgos", Arnes: "backend-nordia", Empresa: "Nordia", Puesto: "Backend", Salud: domain.SaludCrit, Status: domain.StatusIdle, View: "Corridas", Parked: "corrida r3"},
	}
}
