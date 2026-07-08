package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// listSessions (S4) — the multisesión rail.
func listSessions(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, svc.List())
	}
}

// createSessionBody is the POST /api/sessions payload. Path, when present, registers the
// arnés's working directory in the same call (S2), so a new front is confined from turn one.
type createSessionBody struct {
	Arnes   string       `json:"arnes"`
	Frente  string       `json:"frente"`
	Empresa string       `json:"empresa"`
	Puesto  string       `json:"puesto"`
	Salud   domain.Salud `json:"salud"`
	View    string       `json:"view"`
	Parked  string       `json:"parked"`
	Path    string       `json:"path"`
}

// createSession opens a new work-front. If a path is supplied it registers the arnés's
// working directory first; a bad path fails the create (400) rather than opening a session
// that would fall back to a scratch dir.
func createSession(svc *usecase.SessionService, reg ports.ArnesRegistry) http.HandlerFunc {
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
		}
		sess, err := svc.Create(domain.Session{
			Arnes:   body.Arnes,
			Frente:  body.Frente,
			Empresa: body.Empresa,
			Puesto:  body.Puesto,
			Salud:   body.Salud,
			View:    body.View,
			Parked:  body.Parked,
		})
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusCreated, sess)
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
		writeJSON(w, http.StatusOK, sess)
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
		writeJSON(w, http.StatusOK, sess)
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
type permissionBody struct {
	RequestID   string `json:"request_id"`
	Decision    string `json:"decision"` // allow | deny
	Role        string `json:"role"`
	TTLSegundos int    `json:"ttl_segundos"`
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
			time.Duration(body.TTLSegundos)*time.Second)
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
