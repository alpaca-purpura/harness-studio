package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// dockEventType is the SSE event name multiplexed for the Dock. It mirrors
// sse.EventDock; kept as a literal so this package needs no transport import (see
// docs/architecture/boundaries/dominio-independiente-de-transporte.md).
const dockEventType = "dock"

// askUserQuestionTool is the CC tool name the Dock renders as a question card (not a
// generic permission ask) — RF-113 bugfix. It never mints a session grant (each question
// is unique; blind-approving a FUTURE different question with no answer would echo a
// stale/wrong updatedInput back to the conductor).
const askUserQuestionTool = "AskUserQuestion"

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
//
// kind=permission is the ask→UI card (D3: el canal del control_request es el Dock): it
// carries request_id + tool + the raw input so the shell paints the diff and answers via
// POST /sessions/{id}/permission. kind=permission_result closes the card.
type dockFrame struct {
	SessionID       string          `json:"session_id"`
	RunID           string          `json:"run_id,omitempty"`
	Kind            string          `json:"kind"` // status|init|delta|message|act|result|error|permission|permission_result
	Text            string          `json:"text,omitempty"`
	Status          string          `json:"status,omitempty"`
	CtxPct          int             `json:"ctx_pct,omitempty"`
	Model           string          `json:"model,omitempty"`
	ClaudeSessionID string          `json:"claude_session_id,omitempty"`
	RequestID       string          `json:"request_id,omitempty"`
	Tool            string          `json:"tool,omitempty"`
	Input           json.RawMessage `json:"input,omitempty"`
	Decision        string          `json:"decision,omitempty"`
}

// pendingPermission is a forwarded control_request waiting for the human (or the role
// authority) to resolve it.
type pendingPermission struct {
	Tool      string
	Input     []byte
	ToolUseID string
}

// RoleSource resolves the ROLE an arnés hydrates (its graph.l0 META) — the authority
// permission-sets derive from (boundary permisos-derivan-del-rol). Empty = unknown
// arnés/rol: the spawn degrades honestly (no permission flags) and a permission
// resolution without a role fails loudly. It is a func, not a port: the composition
// root closes it over the MapService (the index already owns the graph).
type RoleSource func(ctx context.Context, arnesID string) string

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

	// human-in-the-loop (Fase E): forwarded control_requests awaiting a decision, and
	// the ephemeral grants already approved (least temporal privilege — they expire).
	pendingPerm map[string]pendingPermission // request_id → pending ask.
	grants      map[string]domain.Grant      // tool → live grant.

	// CH-D3: algún EventMessage ya cerró burbuja este turno → el result no re-arma el
	// texto desde su propio campo (duplicaría lo flusheado).
	msgFlushed bool

	// cwd real del conductor (resuelto al spawn) — insumo del reindex-tras-turno (RF-184).
	cwd string
}

// activa devuelve la conversación activa de la sesión — el hilo que tiene conductor
// (CV-D7). Nunca devuelve nil: si el registro llegó sin conversaciones (una sesión
// persistida antes de CV-D3, una sembrada, o un archivo editado a mano) repara la
// invariante acá mismo y lo DICE en el log. Reparar en silencio está prohibido (E-01).
//
// Caller holds s.mu — igual que todo lo que toca r.meta.
func (r *sessionRuntime) activa() *domain.Conversacion {
	if c, ok := r.meta.Activa(); ok {
		return c
	}
	for _, arreglo := range r.meta.NormalizarConversaciones(time.Now().UTC()) {
		slog.Warn("session: invariante de conversaciones reparada", "arreglo", arreglo)
	}
	c, _ := r.meta.Activa()
	return c
}

// SessionService owns the registry of work-fronts and drives their Claude Code
// conductors (multisesión). It is the daemon's source of truth for open sessions; the
// conversation content is rehydrated from Claude Code via --resume.
type SessionService struct {
	mu        sync.Mutex
	rt        map[string]*sessionRuntime
	order     []string // stable creation order for List.
	agent     ports.AgentPort
	store     ports.SessionStore
	pub       EventPublisher
	resolver  ports.WorkdirResolver
	injector  ports.InjectionProvisioner // nil = spawns sin doctrina (degradación honesta).
	perms     ports.PermissionPort       // resuelve el set del rol al responder un control_request.
	roleFor   RoleSource                 // rol del arnés (server-side, jamás del FE) — RF-112/RF-114.
	reindex   Reindexer                  // reindex del Mapa tras cada turno (RF-184); nil = sin reindex.
	grounding GroundingSource            // tarjeta de identidad por sesión (RF-189); nil = doctrina compartida.
	umbralRot int                        // % de contexto que marca rotación pendiente (RF-195); 0 = apagado.
	cerradas  ports.SessionStore         // registro de sesiones cerradas (RF-200); nil = borrado seco.
	historial HistoryReader              // lector del corpus JSONL nativo (RF-201); nil = sin historial.
	cerrado   []string                   // marcas del paquete propio (CH-D6); vacío = guardrail apagado.
	baseCtx   context.Context
	maxTurns  int
}

// maxCtxHist acota el histórico de ctxPct por sesión (RF-194) — suficiente para cualquier
// conversación real, sin crecer sin techo en el JSON persistido.
const maxCtxHist = 500

// SetUmbralRotacion cablea el umbral de rotación de contexto (RF-195, default del daemon
// 40 %). 0 apaga el disparador. Se llama una vez en el composition root.
func (s *SessionService) SetUmbralRotacion(pct int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.umbralRot = pct
}

// SetReindexer cablea el reindex-tras-turno (RF-184). Se llama una vez en el composition
// root, antes de servir; nil lo apaga (comportamiento previo).
func (s *SessionService) SetReindexer(r Reindexer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reindex = r
}

// ProtegerPaqueteCerrado cablea el árbol del paquete propio de arnesia (el kit embebido
// materializado, CH-D6): desde ahí, cualquier control_request cuyo input lo mencione se
// deniega en el gate — sin tarjeta, por encima de grants y del click humano. Se llama una
// vez en el composition root; también registra la forma `~/…` para atrapar comandos Bash.
func (s *SessionService) ProtegerPaqueteCerrado(dir string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cerrado = []string{dir}
	if home, err := os.UserHomeDir(); err == nil {
		if rel, err := filepath.Rel(home, dir); err == nil && !strings.HasPrefix(rel, "..") {
			s.cerrado = append(s.cerrado, "~/"+rel)
		}
	}
}

// refiereAlguno reporta si el input crudo de un tool_use menciona alguna marca.
// ponytail: substring sobre el JSON crudo — blunt adrede (deny es el lado seguro y cubre
// file_path/command/notebook de una); si un falso positivo molesta, el upgrade es parsear
// el path por tool.
func refiereAlguno(input []byte, marcas []string) bool {
	in := string(input)
	for _, m := range marcas {
		if m != "" && strings.Contains(in, m) {
			return true
		}
	}
	return false
}

// NewSessionService loads the persisted registry and returns a ready service. baseCtx
// bounds every conductor's lifetime (cancel it to stop all sessions on shutdown); resolver
// maps each session's arnés to the working directory its conductor runs in (per-session
// confinement, never a shared cwd); maxTurns caps every turn's agent loop; injector
// materializa la doctrina/kit e inyecta los flags a cada spawn (nil = sin inyección).
// perms resuelve el permission-set del rol al responder un control_request (nil = el
// endpoint de permisos responde honesto que no hay autoridad cableada). roleFor resuelve
// el rol del arnés de cada sesión — con él el spawn del Dock materializa los flags de
// permisos (RF-112) y una resolución sin rol explícito usa la autoridad del arnés
// (RF-114); nil = spawns sin flags (comportamiento previo).
func NewSessionService(baseCtx context.Context, agent ports.AgentPort, store ports.SessionStore, pub EventPublisher, resolver ports.WorkdirResolver, maxTurns int, injector ports.InjectionProvisioner, perms ports.PermissionPort, roleFor RoleSource) (*SessionService, error) {
	s := &SessionService{
		rt:       map[string]*sessionRuntime{},
		agent:    agent,
		store:    store,
		pub:      pub,
		resolver: resolver,
		injector: injector,
		perms:    perms,
		roleFor:  roleFor,
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
	ahora := time.Now().UTC()
	for i := range persisted {
		m := persisted[i]
		// Processes are gone after a restart: no conductor is live, so nothing is
		// streaming. The conversation resumes via --resume on the next turn.
		m.Status = domain.StatusIdle
		// Segunda línea de defensa de la invariante (CV-D3): un registro escrito antes de
		// que la conversación existiera, una migración parcial o un archivo editado a mano
		// no pueden meter una sesión sin conversación activa en memoria. Lo reparado se
		// DICE — reparar en silencio está prohibido (E-01).
		for _, arreglo := range m.NormalizarConversaciones(ahora) {
			slog.Warn("session service: invariante de conversaciones reparada al cargar", "arreglo", arreglo)
		}
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
			// Instantanea, no *r.meta: el valor cruza el candado y el conductor sigue
			// escribiendo en las conversaciones del runtime.
			out = append(out, r.meta.Instantanea())
		}
	}
	return out
}

// Get returns one session by id.
func (s *SessionService) Get(id string) (domain.Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r := s.rt[id]; r != nil {
		return r.meta.Instantanea(), true
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
	return sess.Instantanea(), nil
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
	return r.meta.Instantanea(), nil
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
	return r.meta.Instantanea(), nil
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
	// Archivar ANTES de borrar (RF-200): la metadata (cadena de ClaudeSessionIDs, cwd,
	// fechas) es el join que el historial B2 necesita para leer las JSONL nativas.
	s.archivarLocked(r.meta.Instantanea())
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
	// `await` also counts — the turn is still in flight, parked on a permission ask; a new
	// user message would interleave with the conductor blocked on the control channel.
	if r.meta.Status == domain.StatusStreaming || r.meta.Status == domain.StatusAwait {
		s.mu.Unlock()
		return ErrBusy
	}

	// Rotación invisible (RF-197): con el umbral cruzado, ESTE turno arranca en un
	// proceso fresco — checkpoint + cadena + breadcrumb; el Session.ID no cambia y el
	// Conv sigue sin cortes. Se rota ENTRE turnos por construcción (nunca streaming acá).
	if r.activa().RotacionPendiente {
		s.rotarLocked(r)
	}

	// Auto-derive the front name from the first user message.
	if r.meta.Frente == "" || r.meta.Frente == "nuevo frente" {
		r.meta.Frente = deriveFrente(text)
	}
	// El turno cae en la conversación ACTIVA, que es la única con conductor (CV-D7).
	conv := r.activa()
	conv.Conv = append(conv.Conv, domain.Turn{Rol: domain.RolUser, Text: text})
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
	resume := r.activa().ClaudeSessionID
	// Inyección de doctrina (HS-11 puente 2): cuerpos ①+② entran por flags desde
	// ~/.arnesia. Un fallo de provisión DEGRADA honesto (spawn sin doctrina + warn),
	// jamás bloquea la sesión (principio 6: guía sin bloqueo).
	var inj ports.Injection
	if s.injector != nil {
		// Tarjeta de identidad por sesión (RF-189): la doctrina compartida + quién es este
		// arnés y qué copia edita la sesión. Sin grounding cableado (o tarjeta vacía),
		// ProvisionSession degrada al archivo compartido.
		tarjeta := ""
		if s.grounding != nil {
			tarjeta = s.grounding(s.baseCtx, r.meta.Arnes, cwd)
		}
		// Checkpoint de rotación (RF-196/197): el proceso fresco arranca sabiendo dónde
		// quedó la conversación — viaja en el MISMO system-prompt por sesión que la
		// tarjeta (un mecanismo, dos usos — MC-D6).
		// El checkpoint es de la CONVERSACIÓN activa, no de la sesión: la key del
		// system-prompt sigue siendo la sesión porque sólo la activa spawnea (§4.6).
		if ck := r.activa().Checkpoint; ck != "" {
			tarjeta = strings.TrimSpace(tarjeta + "\n\n## Checkpoint de rotación (la conversación CONTINÚA)\n\n" + ck)
		}
		var ierr error
		if inj, ierr = s.injector.ProvisionSession(s.baseCtx, id, tarjeta); ierr != nil {
			slog.Warn("session: provisión de doctrina falló — spawn sin inyección", "err", ierr)
			inj = ports.Injection{}
		}
	}
	// Permisos del ROL del arnés (RF-112, permisos-derivan-del-rol): el spawn del Dock
	// materializa --permission-mode/--allowedTools/--permission-prompt-tool stdio para que
	// cada escritura llegue como control_request→tarjeta. Sin rol resoluble se degrada
	// honesto (sin flags = comportamiento previo del Dock), jamás bloquea la sesión.
	var permisos domain.PermissionSet
	if s.perms != nil && s.roleFor != nil {
		if rol := s.roleFor(s.baseCtx, r.meta.Arnes); rol != "" {
			if ps, perr := s.perms.ResolveForRole(s.baseCtx, rol); perr != nil {
				slog.Warn("session: rol no resoluble — spawn sin flags de permisos", "arnes", r.meta.Arnes, "rol", rol, "err", perr)
			} else {
				permisos = ps
			}
		}
	}
	live, err := s.agent.Spawn(s.baseCtx, ports.SpawnOpts{
		Resume: resume,
		// El modelo es de la conversación (§1.2): se captura del init de CADA proceso
		// suyo, así que el que se pide al spawn sale de la misma — un solo dueño.
		Model:     r.activa().Model,
		Cwd:       cwd,
		MaxTurns:  s.maxTurns,
		Injection: inj,
		Permisos:  permisos,
	})
	if err != nil {
		return err
	}
	r.live = live
	r.cwd = cwd
	r.meta.Cwd = cwd // persistido: el join del historial B2 (RF-200) sobrevive al cierre.
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
				conv := r.activa()
				conv.ClaudeSessionID = ev.ClaudeSessionID
				if ev.Model != "" {
					conv.Model = ev.Model
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
			var reindex Reindexer
			var reindexArnes, reindexCwd string
			if r := s.rt[id]; r != nil && r.live == live {
				// CH-D3: los EventMessage ya cerraron sus burbujas; acá solo cae el
				// remanente (deltas sin message — p. ej. resume viejo) o el fallback.
				final := strings.TrimSpace(r.assembling.String())
				if final == "" && !r.msgFlushed {
					final = ev.Text
				}
				conv := r.activa()
				if final != "" {
					conv.Conv = append(conv.Conv, domain.Turn{Rol: domain.RolAssistant, Text: final})
				}
				r.msgFlushed = false
				r.meta.Status = domain.StatusIdle
				if ev.CtxPct > 0 {
					conv.CtxPct = ev.CtxPct
					// Histórico + umbral de rotación (RF-194/195): se marca DESPUÉS de
					// responder el turno; el próximo Turn rota — nunca a mitad de nada.
					conv.CtxHist = append(conv.CtxHist, ev.CtxPct)
					if len(conv.CtxHist) > maxCtxHist {
						conv.CtxHist = conv.CtxHist[len(conv.CtxHist)-maxCtxHist:]
					}
					if s.umbralRot > 0 && ev.CtxPct >= s.umbralRot && !conv.RotacionPendiente {
						conv.RotacionPendiente = true
						slog.Info("session: umbral de contexto cruzado — rotación pendiente",
							"session", id, "ctx_pct", ev.CtxPct, "umbral", s.umbralRot)
					}
				}
				r.assembling.Reset()
				r.pendingTurn = ""
				runID = r.curRun
				reindex, reindexArnes, reindexCwd = s.reindex, r.meta.Arnes, r.cwd
				s.persistLocked()
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "result", Text: ev.Text, CtxPct: ev.CtxPct, Status: string(domain.StatusIdle)})
			// Reindex-tras-turno (RF-184): el Mapa refleja lo que el chat acaba de editar.
			// Fuera del lock (hace IO) y best-effort (RF-185): jamás rompe el turno.
			if reindex != nil && reindexCwd != "" {
				reindex(s.baseCtx, reindexArnes, reindexCwd)
			}

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

		case ports.EventControlRequest:
			s.onControlRequest(id, live, ev)

		case ports.EventMessage:
			// CH-D3: cada mensaje assistant completo CIERRA su burbuja — el texto (ya
			// streameado por deltas) se congela como Turn propio y el buffer se resetea;
			// el result del turno ya no lo re-arma.
			txt := strings.TrimSpace(ev.Text)
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				runID = r.curRun
				r.assembling.Reset()
				if txt != "" {
					conv := r.activa()
					conv.Conv = append(conv.Conv, domain.Turn{Rol: domain.RolAssistant, Text: txt})
					r.msgFlushed = true
					s.persistLocked()
				}
			}
			s.mu.Unlock()
			if txt != "" {
				s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "message", Text: txt, Status: string(domain.StatusStreaming)})
			}

		case ports.EventActivity:
			// CH-D2: paso visible del turno — rastro RolAct persistido (el FE agrupa
			// consecutivos en la tarjeta desplegable) + frame `act` en vivo.
			paso := strings.TrimSpace(ev.Tool + " " + ev.Text)
			s.mu.Lock()
			runID := ""
			if r := s.rt[id]; r != nil && r.live == live {
				runID = r.curRun
				conv := r.activa()
				conv.Conv = append(conv.Conv, domain.Turn{Rol: domain.RolAct, Text: paso})
				s.persistLocked()
			}
			s.mu.Unlock()
			s.publish(dockFrame{SessionID: id, RunID: runID, Kind: "act", Tool: ev.Tool, Text: ev.Text, Status: string(domain.StatusStreaming)})
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
	conv := r.activa()
	slog.Info("session service: resume failed, restarting fresh", "session", id, "stale_cc", conv.ClaudeSessionID)
	r.resumeRetried = true
	conv.ClaudeSessionID = "" // the id was stale; next spawn starts fresh.
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

// onControlRequest handles one forwarded control_request (human-in-the-loop, Fase E).
// A live, unexpired grant for the same tool auto-allows without re-asking (permiso
// efímero: mientras Vigente no se re-pregunta; expirado → re-aprobación). Anything else
// parks the ask as pending, flips the session to `await` and emits the Dock card (D3);
// POST /sessions/{id}/permission resolves it.
func (s *SessionService) onControlRequest(id string, live ports.AgentSession, ev ports.AgentEvent) {
	now := time.Now()
	autoAllow := false

	s.mu.Lock()
	r := s.rt[id]
	if r == nil || r.live != live {
		s.mu.Unlock()
		return // stale process: nothing to answer against.
	}
	runID := r.curRun
	// CH-D6: el paquete propio de arnesia es CERRADO — se deniega acá, ANTES del
	// auto-allow por grant: ni un grant vigente ni el click humano lo abren.
	if len(s.cerrado) > 0 && refiereAlguno(ev.Input, s.cerrado) {
		s.mu.Unlock()
		const motivo = "paquete cerrado: los arneses de arnesia no se modifican desde el chat"
		if err := live.RespondControl(s.baseCtx, ev.RequestID, ports.ControlDecision{Message: motivo, ToolUseID: ev.ToolUseID}); err != nil {
			slog.Error("session: denegar tool_use sobre el paquete cerrado", "session", id, "err", err)
		}
		s.publish(dockFrame{
			SessionID: id, RunID: runID, Kind: "permission_result",
			RequestID: ev.RequestID, Tool: ev.Tool,
			Decision: string(domain.DecisionDeny), Text: motivo,
		})
		return
	}
	if g, ok := r.grants[ev.Tool]; ok && g.Vigente(now) {
		autoAllow = true
	} else {
		if r.pendingPerm == nil {
			r.pendingPerm = map[string]pendingPermission{}
		}
		r.pendingPerm[ev.RequestID] = pendingPermission{Tool: ev.Tool, Input: ev.Input, ToolUseID: ev.ToolUseID}
		r.meta.Status = domain.StatusAwait
		s.persistLocked()
	}
	s.mu.Unlock()

	if autoAllow {
		if err := live.RespondControl(s.baseCtx, ev.RequestID, ports.ControlDecision{Allow: true, UpdatedInput: ev.Input, ToolUseID: ev.ToolUseID}); err != nil {
			slog.Error("session: responder control_request con grant vigente", "session", id, "err", err)
			return
		}
		s.publish(dockFrame{
			SessionID: id, RunID: runID, Kind: "permission_result",
			RequestID: ev.RequestID, Tool: ev.Tool, Decision: string(domain.DecisionAllow),
			Text: "grant vigente — sin re-pregunta",
		})
		return
	}
	s.publish(dockFrame{
		SessionID: id, RunID: runID, Kind: "permission",
		RequestID: ev.RequestID, Tool: ev.Tool, Input: ev.Input,
		Status: string(domain.StatusAwait),
	})
}

// PermissionResolution is the effective outcome of resolving a control_request.
type PermissionResolution struct {
	RequestID string          `json:"request_id"`
	Tool      string          `json:"tool"`
	Efectiva  domain.Decision `json:"efectiva"`
	Motivo    string          `json:"motivo,omitempty"`
	Expira    *time.Time      `json:"expira,omitempty"` // fin del grant cuando Efectiva=allow.
}

// Permission sentinels the transport maps to status codes.
var (
	// ErrPermisoNoPendiente — no session/control_request matches (HTTP 404).
	ErrPermisoNoPendiente = errors.New("no hay control_request pendiente con ese id")
	// ErrEnvioControl — the answer could not reach the conductor's stdin (HTTP 500).
	ErrEnvioControl = errors.New("no pude entregar la respuesta al conductor")
)

// ResolvePermission resolves a pending control_request with the ROLE's authority plus
// the human decision (permisos-derivan-del-rol + permisos-gui-human-in-the-loop):
//
//   - el set del rol manda: si ps.Decide(tool) es deny, la respuesta efectiva es deny
//     aunque el humano haya aprobado (la autoridad se impone fuera del razonamiento —
//     y fuera del click);
//   - allow mintea un Grant efímero (TTL del rol, o ttl si el operador lo acota más);
//     mientras Vigente, el mismo tool no re-pregunta en esta sesión — EXCEPTO
//     askUserQuestionTool, que jamás mintea grant (cada pregunta es distinta);
//   - deny-by-default: sin rol no hay resolución (ResolveForRole exige rol).
//
// answers (RF-113 bugfix) son las respuestas humanas a un AskUserQuestion — question text
// → label elegido (multi-select: labels separados por coma, mismo shape que la propia
// herramienta usa internamente). nil/vacío para cualquier otro tool.
func (s *SessionService) ResolvePermission(id, requestID, decision, rol string, ttl time.Duration, answers map[string]string) (PermissionResolution, error) {
	if decision != string(domain.DecisionAllow) && decision != string(domain.DecisionDeny) {
		return PermissionResolution{}, fmt.Errorf("decision %q inválida: allow|deny", decision)
	}
	if s.perms == nil {
		return PermissionResolution{}, errors.New("sin PermissionPort cableado — no hay autoridad de rol para resolver")
	}

	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return PermissionResolution{}, fmt.Errorf("%w (sesión %q)", ErrPermisoNoPendiente, id)
	}
	p, ok := r.pendingPerm[requestID]
	live := r.live
	arnes := r.meta.Arnes
	s.mu.Unlock()
	if !ok {
		return PermissionResolution{}, fmt.Errorf("%w (request %q)", ErrPermisoNoPendiente, requestID)
	}

	// RF-114: sin rol explícito, la autoridad ES la del arnés de la sesión (META de
	// enganche) — el FE no elige autoridad, la hereda (decisión #6).
	if rol == "" && s.roleFor != nil {
		rol = s.roleFor(s.baseCtx, arnes)
	}

	ps, err := s.perms.ResolveForRole(s.baseCtx, rol)
	if err != nil {
		return PermissionResolution{}, fmt.Errorf("resolver rol: %w", err)
	}

	res := PermissionResolution{RequestID: requestID, Tool: p.Tool, Efectiva: domain.Decision(decision)}
	switch {
	case ps.Decide(p.Tool) == domain.DecisionDeny:
		// Deny del rol gana SIEMPRE (deny > ask > allow) — incluso sobre el click humano.
		res.Efectiva = domain.DecisionDeny
		res.Motivo = fmt.Sprintf("el rol %q deniega %s — la autoridad del rol gana sobre la aprobación", ps.Rol, p.Tool)
	case res.Efectiva == domain.DecisionDeny:
		res.Motivo = "denegado por el operador"
	default:
		res.Motivo = "aprobado por el operador"
	}

	var grant domain.Grant
	mintaGrant := res.Efectiva == domain.DecisionAllow && p.Tool != askUserQuestionTool
	if mintaGrant {
		grantSet := ps
		if ttl > 0 && (grantSet.TTL == 0 || ttl < grantSet.TTL) {
			grantSet.TTL = ttl // el operador solo puede ACOTAR el TTL del rol, nunca ampliarlo.
		}
		grant = grantSet.NuevoGrant(p.Tool, time.Now())
		res.Expira = &grant.Expira
	}

	s.mu.Lock()
	if r2 := s.rt[id]; r2 != nil {
		delete(r2.pendingPerm, requestID)
		if mintaGrant {
			if r2.grants == nil {
				r2.grants = map[string]domain.Grant{}
			}
			r2.grants[p.Tool] = grant
		}
		if len(r2.pendingPerm) == 0 && r2.meta.Status == domain.StatusAwait {
			r2.meta.Status = domain.StatusStreaming // el turno sigue vivo tras la decisión.
		}
		s.persistLocked()
	}
	s.mu.Unlock()

	if live == nil {
		return res, fmt.Errorf("%w: la sesión ya no tiene conductor vivo", ErrEnvioControl)
	}
	d := ports.ControlDecision{Allow: res.Efectiva == domain.DecisionAllow, Message: res.Motivo, ToolUseID: p.ToolUseID}
	if d.Allow {
		if p.Tool == askUserQuestionTool && len(answers) > 0 {
			d.UpdatedInput = askUserQuestionUpdatedInput(p.Input, answers)
		} else {
			d.UpdatedInput = p.Input // echo del input original (control protocol: requerido en allow).
		}
	}
	if err := live.RespondControl(s.baseCtx, requestID, d); err != nil {
		return res, fmt.Errorf("%w: %w", ErrEnvioControl, err)
	}

	s.publish(dockFrame{
		SessionID: id, Kind: "permission_result",
		RequestID: requestID, Tool: p.Tool,
		Decision: string(res.Efectiva), Text: res.Motivo,
		Status: string(domain.StatusStreaming),
	})
	return res, nil
}

// askUserQuestionUpdatedInput builds the allow updatedInput for an AskUserQuestion
// control_request: el `questions` original (y cualquier otro campo) + `answers` (question
// text → label elegido). HONESTIDAD sobre el wire: no está documentado oficialmente:
// inferido del zod schema `{questions, answers, response?, annotations?, afkTimeoutMs?}`
// leído del binario `claude` instalado (2.1.220) — misma incertidumbre best-effort que
// controlResponseLine (claudecode/conductor.go). Si el original no parsea como objeto,
// cae a echo crudo (nunca inventa un shape sobre datos que no pudo leer).
func askUserQuestionUpdatedInput(original []byte, answers map[string]string) []byte {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(original, &m); err != nil || m == nil {
		return original
	}
	raw, err := json.Marshal(answers)
	if err != nil {
		return original
	}
	m["answers"] = raw
	out, err := json.Marshal(m)
	if err != nil {
		return original
	}
	return out
}

// ErrNadaQueInterrumpir — Interrupt on a session with no live in-flight turn (HTTP 409).
var ErrNadaQueInterrumpir = errors.New("no hay turno en vuelo que interrumpir")

// Interrupt stops a session's in-flight turn (RF-116): pending permission asks are
// denied first (motivo «interrumpido por el operador» — the conductor must not stay
// blocked on an ask nobody will answer), then the in-band interrupt rides the control
// channel. The conductor emits its result frame and the session settles idle by the
// normal consume path; the subprocess stays alive for the next turn.
func (s *SessionService) Interrupt(id string) error {
	s.mu.Lock()
	r := s.rt[id]
	if r == nil {
		s.mu.Unlock()
		return errNotFound(id)
	}
	live := r.live
	status := r.meta.Status
	if live == nil || (status != domain.StatusStreaming && status != domain.StatusAwait) {
		s.mu.Unlock()
		return ErrNadaQueInterrumpir
	}
	pend := r.pendingPerm
	r.pendingPerm = nil
	if status == domain.StatusAwait {
		r.meta.Status = domain.StatusStreaming // el turno sigue vivo hasta su result.
		s.persistLocked()
	}
	runID := r.curRun
	s.mu.Unlock()

	for reqID, p := range pend {
		if err := live.RespondControl(s.baseCtx, reqID, ports.ControlDecision{Message: "interrumpido por el operador", ToolUseID: p.ToolUseID}); err != nil {
			slog.Warn("session: denegar ask pendiente al interrumpir", "session", id, "request", reqID, "err", err)
		}
		s.publish(dockFrame{
			SessionID: id, RunID: runID, Kind: "permission_result",
			RequestID: reqID, Tool: p.Tool, Decision: string(domain.DecisionDeny),
			Text: "interrumpido por el operador", Status: string(domain.StatusStreaming),
		})
	}
	if err := live.Interrupt(s.baseCtx); err != nil {
		return fmt.Errorf("%w: %w", ErrEnvioControl, err)
	}
	return nil
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
			// El store puede retener el slice (los in-memory lo hacen): va instantánea.
			snap = append(snap, r.meta.Instantanea())
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
