// Package httpapi is the HTTP delivery adapter: it mounts the daemon's routes
// (arch/contracts/api/openapi.yaml) onto a net/http mux and translates requests to
// use-case calls. It holds no business rules (see
// arch/boundaries/dominio-independiente-de-transporte.md). The SSE stream is injected
// as a plain http.Handler so this package need not know the concrete broker.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// errorBody is the JSON envelope for error responses.
type errorBody struct {
	Error string `json:"error"`
}

// NewHandler builds the daemon's router. maps serves the Map/graph endpoints; sessions
// drives the multisesión Dock; arneses is the arnés→path registry (per-session workdir
// confinement); events is the SSE broker mounted at /events. auth confines the whole
// surface (Host+Origin+token, boundary superficie-local-confinada).
func NewHandler(maps *usecase.MapService, sessions *usecase.SessionService, arneses ports.ArnesRegistry, events http.Handler, auth AuthConfig) http.Handler {
	mux := http.NewServeMux()

	// Liveness + the multiplexed SSE stream (the two endpoints the shell polls first).
	mux.HandleFunc("GET /healthz", healthz)
	mux.Handle("GET /events", events)
	mux.Handle("GET /api/events", events) // OpenAPI server base is /api.

	// Portfolio / Map / Inspector / Runs (S1–S3, S8).
	mux.HandleFunc("GET /api/harnesses", listHarnesses(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/graph", getHarnessGraph(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/nodes/{nodeId}", getNode(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/runs", listRuns)

	// Arnés registry (S2) — maps an arnés to the working dir its sessions run claude in.
	mux.HandleFunc("GET /api/arneses", listArneses(arneses))
	mux.HandleFunc("PUT /api/arneses/{id}", registerArnes(arneses))

	// Multisesión + Dock (S4). Every conductor turn streams back over /events.
	mux.HandleFunc("GET /api/sessions", listSessions(sessions))
	mux.HandleFunc("POST /api/sessions", createSession(sessions, arneses))
	mux.HandleFunc("GET /api/sessions/{id}", getSession(sessions))
	mux.HandleFunc("PATCH /api/sessions/{id}", patchSession(sessions))
	mux.HandleFunc("DELETE /api/sessions/{id}", deleteSession(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/turn", sessionTurn(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/permission", resolvePermission)

	return withAuth(auth, mux)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// harnessSummary is one row of GET /api/harnesses — the lightweight portfolio entry the Map
// picker (RF-72) consumes; the full graph is a separate call (getHarnessGraph).
type harnessSummary struct {
	ID      string `json:"id"`
	Rol     string `json:"rol,omitempty"`
	Proceso string `json:"proceso,omitempty"`
	Empresa string `json:"empresa,omitempty"`
}

// listHarnesses (S1) — portfolio, from the index (RF-72).
func listHarnesses(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gs, err := maps.Harnesses(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		out := make([]harnessSummary, 0, len(gs))
		for _, g := range gs {
			if g.Arnes == nil {
				continue
			}
			out = append(out, harnessSummary{
				ID:      g.Arnes.ID,
				Rol:     g.Arnes.Rol,
				Proceso: g.Arnes.Proceso,
				Empresa: g.Arnes.Empresa,
			})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// getHarnessGraph (S2) — the agnostic graph of one harness (real, from the index).
func getHarnessGraph(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		g, err := maps.Graph(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, g)
	}
}

// getNode (S3) — inspector: one node (Box + fused Contract) of a harness (RF-71).
func getNode(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		box, ok, err := maps.Node(r.Context(), r.PathValue("id"), r.PathValue("nodeId"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "node not found"})
			return
		}
		writeJSON(w, http.StatusOK, box)
	}
}

// listRuns (S8) — runs derived from the JSONL. Stub.
func listRuns(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

// resolvePermission — diff-approval (human-in-the-loop). Stub (fase 4 spec S4).
func resolvePermission(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

func notImplemented(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotImplemented, errorBody{Error: "not implemented (fase 5)"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("http: encode response", "err", err)
	}
}
