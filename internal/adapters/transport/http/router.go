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

// NewHandler builds the daemon's router. maps serves the Map/graph endpoints; events
// is the SSE broker mounted at /events (and its OpenAPI path /api/events).
func NewHandler(maps *usecase.MapService, events http.Handler) http.Handler {
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

	// Dock (S4): turn + human-in-the-loop permission.
	mux.HandleFunc("POST /api/dock/{sessionId}/turn", sendTurn)
	mux.HandleFunc("POST /api/dock/{sessionId}/permission", resolvePermission)

	return mux
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

// sendTurn (S4) — enqueue a turn to the live conductor session. Stub.
func sendTurn(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

// resolvePermission — diff-approval (human-in-the-loop). Stub.
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
