package httpapi

import (
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// forja.go expone la siembra `.arnesia/` (contrato semilla-arnesia.md; paquete
// 2026-07-30-arnesia-en-el-proyecto, A-T3). El CLI (`arnesia init`) es la vía de
// verificación E2E; este HTTP es lo que consumirá el FE. El campo del body es `path`
// — el MISMO nombre que POST /api/portafolio/escaneos, no un sinónimo nuevo por ruta.

// postSemillaBody is the payload of both /api/forja/semillas endpoints.
type postSemillaBody struct {
	Path string `json:"path"`
}

// semillaSembradaResponse: el informe de la siembra + la salud posterior — la salud viaja
// SIEMPRE, también insana (el criterio de «no avanzamos» es del caller; HTTP no tiene
// exit code y un 200 con estado incompleta es más honesto que un error opaco).
type semillaSembradaResponse struct {
	Informe domain.InformeSemilla `json:"informe"`
	Salud   domain.SaludSemilla   `json:"salud"`
}

// postSembrarSemilla — POST /api/forja/semillas: siembra idempotente (lo existente JAMÁS
// se pisa) + doctor al final. Path inválido/protegido → 400 con motivo, nada se escribe.
func postSembrarSemilla(svc *usecase.ForjaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postSemillaBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		informe, err := svc.Sembrar(r.Context(), body.Path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		salud, err := svc.Chequear(r.Context(), body.Path)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, semillaSembradaResponse{Informe: informe, Salud: salud})
	}
}

// postChequearSemilla — POST /api/forja/semillas/chequeos: solo el doctor v0 (presencia,
// sin hashes — §4). El veredicto se emite TAL CUAL salga, jamás fabricado.
func postChequearSemilla(svc *usecase.ForjaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postSemillaBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		salud, err := svc.Chequear(r.Context(), body.Path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, salud)
	}
}
