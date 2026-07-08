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

// runBox corre la caja SÍNCRONO (el loop está acotado por repair-cap × --max-turns) y
// devuelve el desenlace; el progreso vivo viaja por /events (event: run).
func runBox(runs *usecase.RunService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := runs.RunBox(r.Context(), r.PathValue("id"), r.PathValue("boxId"))
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
				// Incluye el fallo de spawn/permisos: el error real del conductor le
				// llega al caller, jamás un 501 genérico.
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}
