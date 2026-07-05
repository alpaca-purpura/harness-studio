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

	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// errorBody is the JSON envelope for error responses.
type errorBody struct {
	Error string `json:"error"`
}

// NewHandler builds the daemon's router. maps serves the Map/graph endpoints; sessions
// drives the multisesión Dock; events is the SSE broker mounted at /events.
func NewHandler(maps *usecase.MapService, sessions *usecase.SessionService, events http.Handler) http.Handler {
	mux := http.NewServeMux()

	// Liveness + the multiplexed SSE stream (the two endpoints the shell polls first).
	mux.HandleFunc("GET /healthz", healthz)
	mux.Handle("GET /events", events)
	mux.Handle("GET /api/events", events) // OpenAPI server base is /api.

	// Portfolio / Map / Inspector / Runs (S1–S3, S8).
	mux.HandleFunc("GET /api/harnesses", listHarnesses)
	mux.HandleFunc("GET /api/harnesses/{id}/graph", getHarnessGraph(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/nodes/{nodeId}", getNode)
	mux.HandleFunc("GET /api/harnesses/{id}/runs", listRuns)

	// Multisesión + Dock (S4). Every conductor turn streams back over /events.
	mux.HandleFunc("GET /api/sessions", listSessions(sessions))
	mux.HandleFunc("POST /api/sessions", createSession(sessions))
	mux.HandleFunc("GET /api/sessions/{id}", getSession(sessions))
	mux.HandleFunc("PATCH /api/sessions/{id}", patchSession(sessions))
	mux.HandleFunc("DELETE /api/sessions/{id}", deleteSession(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/turn", sessionTurn(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/permission", resolvePermission)

	return withCORS(mux)
}

// withCORS allows the Vite dev origin (and the Tauri WebView) to call the daemon
// cross-origin during development, and answers preflight requests.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Last-Event-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// listHarnesses (S1) — portfolio. Stub: empty list.
func listHarnesses(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, []any{})
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

// getNode (S3) — inspector. Stub.
func getNode(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
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
