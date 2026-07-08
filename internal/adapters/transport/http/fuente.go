package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// fuente.go expone la tab Contenido del inspector (RF-93): la fuente REAL del nodo,
// leída CONFINADA al directorio registrado del arnés (S2). El contenido viaja tal cual
// (text/plain) — el drawer lo muestra read-only con números de línea; jamás se
// reconstruye ni se inventa.

// getNodeFuente — GET /api/harnesses/{id}/nodes/{nodeId}/fuente.
// 200 text/plain (el archivo, con su ruta en X-Arnesia-Fuente-Path) · 403 si la ruta
// escapa del dir del arnés · 404 en todo lo demás (arnés/nodo/fuente/dir/archivo
// ausentes — el cuerpo dice CUÁL, estado honesto).
func getNodeFuente(fuentes *usecase.FuenteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path, contenido, err := fuentes.Fuente(r.Context(), r.PathValue("id"), r.PathValue("nodeId"))
		if err != nil {
			status := http.StatusNotFound
			if errors.Is(err, ports.ErrFueraDelArnes) {
				status = http.StatusForbidden
			}
			writeJSON(w, status, errorBody{Error: err.Error()})
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Arnesia-Fuente-Path", path)
		// (El nolint:gosec G705 que vivía aquí quedó huérfano: con selfupdate.go en el
		// paquete, gosec ya no reporta este Write — Content-Type text/plain sigue fijado.)
		if _, err := w.Write(contenido); err != nil {
			slog.Error("http: write fuente", "err", err)
		}
	}
}
