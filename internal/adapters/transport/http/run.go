package httpapi

import (
	"errors"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// run.go expone el conductor T3 como recurso propio (D2, gate-0): POST
// /api/harnesses/{id}/boxes/{boxId}/run. El run NO es conversacional — es el loop
// determinista del BoxConductor (orquestacion-determinista-entre-cajas); el turno
// humano del Dock vive aparte en /sessions/{id}/turn.

// precondicionBody es la respuesta 409 de RF-120: qué insumos faltan, accionable.
type precondicionBody struct {
	Error        string   `json:"error"`
	Precondicion string   `json:"precondicion"`
	Faltantes    []string `json:"faltantes"`
}

// startRunBody es la respuesta 202 (deuda BACKLOG «run async», 2026-07-23): el run_id
// para pollear GET .../runs/{runId} o escuchar /events (event: run).
type startRunBody struct {
	RunID string `json:"run_id"`
}

// runBox arranca la caja ASÍNCRONO (StartRun valida síncrono — 404/422/409 le llegan al
// caller en el POST mismo, cero tokens quemados en un run que ni arrancó — y el loop del
// conductor, lo que de verdad tarda, corre en background): 202 + run_id de inmediato; el
// desenlace final se consulta con GET .../runs/{runId}, el progreso vivo sigue viajando
// por /events (event: run), sin cambios respecto al contrato previo.
func runBox(runs *usecase.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID, err := runs.StartRun(r.Context(), r.PathValue("id"), r.PathValue("boxId"))
		if err != nil {
			var pre *usecase.PrecondicionError
			switch {
			case errors.Is(err, usecase.ErrArnesNoIndexado), errors.Is(err, usecase.ErrNodoNoExiste):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrNoEsCaja):
				writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: err.Error()})
			case errors.As(err, &pre):
				// RF-120: el run NO arrancó (cero tokens) — 409 con la lista de faltantes.
				writeJSON(w, http.StatusConflict, precondicionBody{
					Error: err.Error(), Precondicion: "incumplida", Faltantes: pre.Faltantes,
				})
			default:
				// Incluye el fallo de resolver permisos/workdir: el error real le
				// llega al caller, jamás un 501 genérico.
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusAccepted, startRunBody{RunID: runID})
	}
}

// getRun devuelve el desenlace de un run ya arrancado (GET .../boxes/{boxId}/runs/{runId}
// — deuda BACKLOG «run async», 2026-07-23). 404 si el run_id no existe en el registro
// (nunca arrancó, o el daemon reinició — el registro es in-memory, no persiste).
func getRun(runs *usecase.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		st, ok := runs.GetRun(r.PathValue("runId"))
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "run no encontrado: " + r.PathValue("runId")})
			return
		}
		writeJSON(w, http.StatusOK, st)
	}
}
