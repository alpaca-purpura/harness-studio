package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// conformance.go expone el botón de auditoría del Mapa (HS-11, puente 3): el motor
// corre el scope `arnes` (schema + spine + escritor único + firewall — pass/fail real,
// sin repo fuente) sobre el grafo que el índice ya sirve. El scope `fabrica`
// (arch-tests/go-arch-lint sobre el código de ArnesIA) NO pasa por aquí: vive en el CI
// del repo.

// getConformance (S9) — GET /api/harnesses/{id}/conformance: audita el grafo indexado.
// baseFor resuelve el directorio del arnés para el firewall scan (fuente_path
// relativos); un id sin directorio conocido difiere el firewall honesto.
func getConformance(maps *usecase.MapService, conf ports.ConformancePort, baseFor func(id string) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		g, err := maps.Graph(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		raw, err := json.Marshal(g)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		rep, err := conf.RunGraph(r.Context(), raw, baseFor(id))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, rep)
	}
}
