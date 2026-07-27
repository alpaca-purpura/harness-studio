package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// sessionWire es la sesión TAL COMO VIAJA. No es `domain.Session` serializada, y la
// diferencia es la razón de que el tipo exista: el agregado tiene N conversaciones con sus
// transcripts completos, y mandarlas todas haría que `GET /api/sessions` costara doce
// kilobytes por conversación por sesión. Lo que viaja es la ACTIVA —la única que la superficie
// pinta— y las demás llegan por `GET /api/sessions/{id}/conversaciones`, sin transcript.
//
// `activa` NO es opcional: la invariante garantiza que existe (BR-CV-1), y un campo opcional
// invitaría al `if (!activa)` defensivo que escondería el día en que no exista.
type sessionWire struct {
	ID         string               `json:"id"`
	Frente     string               `json:"frente"`
	Arnes      string               `json:"arnes"`
	Empresa    string               `json:"empresa,omitempty"`
	Puesto     string               `json:"puesto,omitempty"`
	Salud      domain.Salud         `json:"salud,omitempty"`
	Status     domain.SessionStatus `json:"status"`
	View       string               `json:"view"`
	Parked     string               `json:"parked,omitempty"`
	Reparacion bool                 `json:"reparacion,omitempty"`
	Cwd        string               `json:"cwd,omitempty"`
	CerradaEn  string               `json:"cerrada_en,omitempty"`

	Activa usecase.ConversacionActiva `json:"activa"`
}

// aWire proyecta una sesión del dominio a lo que viaja.
//
// Una sesión sin conversación activa no debería existir —la invariante se repara al cargar y
// `Create` la produce— pero si llegara una, esto NO inventa una: manda el cero con la lista de
// turnos vacía y lo DICE en el log. Fabricarle un id acá haría que el defecto se viera como un
// dato normal en la interfaz, que es la forma más cara de esconder un bug.
func aWire(s domain.Session) sessionWire {
	w := sessionWire{
		ID: s.ID, Frente: s.Frente, Arnes: s.Arnes,
		Empresa: s.Empresa, Puesto: s.Puesto, Salud: s.Salud,
		Status: s.Status, View: s.View, Parked: s.Parked,
		Reparacion: s.Reparacion, Cwd: s.Cwd, CerradaEn: s.CerradaEn,
	}
	c, ok := s.Activa()
	if !ok {
		slog.Error("sessions: una sesión sin conversación activa llegó al wire", "session", s.ID)
		w.Activa = usecase.ConversacionActiva{Conv: []domain.Turn{}}
		return w
	}
	w.Activa = usecase.ConversacionActivaDe(*c)
	return w
}

func aWireLista(ss []domain.Session) []sessionWire {
	out := make([]sessionWire, 0, len(ss))
	for _, s := range ss {
		out = append(out, aWire(s))
	}
	return out
}

// listSessions (S4) — el rail multisesión. `?arnes=<id>` filtra por arnés con comparación
// EXACTA. `?cerradas=1` se RETIRÓ (RF-345 CA-1): bajo el modelo nuevo las conversaciones no se
// cierran, se desactivan, y las de una sesión viven en su propio endpoint. Responde 400 con el
// puntero en vez de ignorar el parámetro en silencio o devolver 200 con una lista vacía —
// las dos alternativas le hacen creer al cliente que preguntó bien.
//
// Y la respuesta vuelve a ser SIEMPRE un array. Era la única bimorfa del daemon: la misma ruta
// devolvía una lista o un objeto según un query, lo que hace la operación indocumentable.
func listSessions(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cerradas") != "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "el parámetro `cerradas` se retiró: " +
				"las conversaciones de una sesión viven en GET /api/sessions/{id}/conversaciones"})
			return
		}
		out := svc.List()
		if arnes := r.URL.Query().Get("arnes"); arnes != "" {
			filtradas := make([]domain.Session, 0, len(out))
			for _, s := range out {
				if s.Arnes == arnes {
					filtradas = append(filtradas, s)
				}
			}
			out = filtradas
		}
		writeJSON(w, http.StatusOK, aWireLista(out))
	}
}

// createSessionBody is the POST /api/sessions payload. Path, when present, registers the
// arnés's working directory in the same call (S2), so a new front is confined from turn one.
type createSessionBody struct {
	Arnes      string       `json:"arnes"`
	Frente     string       `json:"frente"`
	Empresa    string       `json:"empresa"`
	Puesto     string       `json:"puesto"`
	Salud      domain.Salud `json:"salud"`
	View       string       `json:"view"`
	Parked     string       `json:"parked"`
	Path       string       `json:"path"`
	Reparacion bool         `json:"reparacion"`
}

// createSession opens a new work-front. If a path is supplied it registers the arnés's
// working directory first; a bad path fails the create (400) rather than opening a session
// that would fall back to a scratch dir. onRegistered (el mismo closure del composition
// root que usa PUT /api/arneses) indexa el árbol registrado ANTES del primer spawn —
// sin esto, roleFor no resuelve el rol del arnés en el primer turno y la sesión spawnea
// sin flags de permisos (gap real destapado por el E2E de T4, paquete
// mejorar-arnes-conversando). Best-effort: un árbol no indexable (carpeta cruda sin sello)
// degrada honesto con warn — la sesión abre igual, como siempre.
func createSession(svc *usecase.SessionService, reg ports.ArnesRegistry, onRegistered func(id, path string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body createSessionBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if body.Path != "" {
			if err := reg.Register(body.Arnes, body.Path); err != nil {
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
				return
			}
			if onRegistered != nil {
				if err := onRegistered(body.Arnes, body.Path); err != nil {
					slog.Warn("createSession: árbol registrado no indexable aún", "arnes", body.Arnes, "err", err)
				}
			}
		}
		sess, err := svc.Create(domain.Session{
			Arnes:      body.Arnes,
			Frente:     body.Frente,
			Empresa:    body.Empresa,
			Puesto:     body.Puesto,
			Salud:      body.Salud,
			View:       body.View,
			Parked:     body.Parked,
			Reparacion: body.Reparacion,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, aWire(sess))
	}
}

// getSession returns one session.
func getSession(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, ok := svc.Get(r.PathValue("id"))
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "session not found"})
			return
		}
		writeJSON(w, http.StatusOK, aWire(sess))
	}
}

// patchSessionBody carries the mutable fields of a session (rename / park view).
type patchSessionBody struct {
	Frente *string `json:"frente"`
	View   *string `json:"view"`
}

// patchSession renames a front and/or records its parked view.
func patchSession(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body patchSessionBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		var (
			sess domain.Session
			err  error
			ok   bool
		)
		if body.Frente != nil {
			sess, err = svc.Rename(id, *body.Frente)
			ok = true
		}
		if body.View != nil {
			sess, err = svc.SetView(id, *body.View)
			ok = true
		}
		if !ok {
			sess, _ = svc.Get(id)
		}
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, aWire(sess))
	}
}

// deleteSession closes a session.
func deleteSession(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Close(r.PathValue("id")); err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// turnBody is the POST /api/sessions/{id}/turn payload.
type turnBody struct {
	Text string `json:"text"`
}

// sessionTurn streams one user message to a session's conductor. The response is
// immediate (202); the assistant reply arrives over the SSE Dock stream.
func sessionTurn(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body turnBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if body.Text == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "empty turn"})
			return
		}
		if err := svc.Turn(id, body.Text); err != nil {
			// A turn while the session is already streaming is a conflict, not a failure
			// (boundary sesion-viva-consistente `un-turno-a-la-vez`).
			if errors.Is(err, usecase.ErrBusy) {
				writeJSON(w, http.StatusConflict, errorBody{Error: err.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// sessionInterrupt (RF-116) stops a session's in-flight turn in-band: pending asks are
// denied and the conductor receives control_request subtype=interrupt. 202 = accepted;
// the result frame arrives over the SSE Dock stream like any turn end.
func sessionInterrupt(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := svc.Interrupt(r.PathValue("id")); err != nil {
			switch {
			case errors.Is(err, usecase.ErrNadaQueInterrumpir):
				writeJSON(w, http.StatusConflict, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrEnvioControl):
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			default:
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			}
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// permissionBody is the POST /api/sessions/{id}/permission payload: the human answer
// to a control_request card of the Dock (D3). role names the authority the decision
// runs under; ttl_segundos optionally NARROWS the role's grant TTL (never widens it).
// answers (RF-113 bugfix) solo lo manda la tarjeta de AskUserQuestion — question text →
// label elegido.
type permissionBody struct {
	RequestID   string            `json:"request_id"`
	Decision    string            `json:"decision"` // allow | deny
	Role        string            `json:"role"`
	TTLSegundos int               `json:"ttl_segundos"`
	Answers     map[string]string `json:"answers,omitempty"`
}

// resolvePermission resolves a pending control_request: role permission-set + human
// decision → deny-wins → ephemeral Grant on allow → control_response al conductor.
// Realiza permisos-gui-human-in-the-loop + permisos-derivan-del-rol (reemplaza el 501).
func resolvePermission(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body permissionBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if body.RequestID == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "request_id requerido"})
			return
		}
		res, err := svc.ResolvePermission(r.PathValue("id"), body.RequestID, body.Decision, body.Role,
			time.Duration(body.TTLSegundos)*time.Second, body.Answers)
		if err != nil {
			switch {
			case errors.Is(err, usecase.ErrPermisoNoPendiente):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrEnvioControl):
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			default:
				// decisión inválida / rol vacío o irresoluble → petición mal formada.
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

// decodeJSON decodes a request body, writing a 400 on failure. It returns the error so
// callers can bail early.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body: " + err.Error()})
		return err
	}
	return nil
}
