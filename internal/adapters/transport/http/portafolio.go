package httpapi

import (
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// portafolio.go expone el Portafolio de arneses (Slice 0, S0-D9): superficie observable
// mínima sin FE — el CLI (cmd/arnesia) es la vía de verificación E2E, este HTTP es lo que
// consumirá Slice 1. Respuestas SIEMPRE honestas: corruptas visibles aparte, errores 400
// con motivo (nunca un 200 que oculte un path inválido).

// entradaWire agrega `clave` (Identidad.Clave() ya resuelto) a domain.EntradaPortafolio
// para el listado — el cliente lo necesita para DELETE /api/portafolio/arneses/{clave}
// sin reimplementar el slug.
type entradaWire struct {
	Clave string `json:"clave"`
	domain.EntradaPortafolio
}

// corruptaWire expone solo el motivo — el blob crudo no viaja al wire.
type corruptaWire struct {
	Motivo string `json:"motivo"`
}

type portafolioListResponse struct {
	Entradas  []entradaWire  `json:"entradas"`
	Corruptas []corruptaWire `json:"corruptas,omitempty"`
}

// listPortafolio — GET /api/portafolio: entradas + corruptas visibles (BR-11).
func listPortafolio(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sanas, corruptas, err := svc.Listar(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		out := portafolioListResponse{Entradas: make([]entradaWire, 0, len(sanas))}
		for _, e := range sanas {
			out.Entradas = append(out.Entradas, entradaWire{Clave: e.Identidad.Clave(), EntradaPortafolio: e})
		}
		for _, c := range corruptas {
			out.Corruptas = append(out.Corruptas, corruptaWire{Motivo: c.Motivo})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// postEscanearBody is the POST /api/portafolio/escaneos payload.
type postEscanearBody struct {
	Path string `json:"path"`
}

// postEscanear — POST /api/portafolio/escaneos: escanea root y devuelve candidatos, NO
// persiste nada (spec §7.1: el usuario elige). Path inválido/no-escaneable → 400 con motivo.
func postEscanear(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postEscanearBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		candidatos, err := svc.Escanear(r.Context(), body.Path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, candidatos)
	}
}

// postAgregarBody is the POST /api/portafolio/proyectos payload.
type postAgregarBody struct {
	Path     string   `json:"path"`
	Elegidos []string `json:"elegidos"`
}

// postAgregar — POST /api/portafolio/proyectos: re-escanea root y persiste SOLO los
// candidatos cuya Clave está en elegidos (idempotente, C-P-8).
func postAgregar(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body postAgregarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		persistidas, err := svc.AgregarProyecto(r.Context(), body.Path, body.Elegidos)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, persistidas)
	}
}

// deleteDesvincular — DELETE /api/portafolio/arneses/{clave}: quita del registro; NO
// borra nada de disco (C-UNL-3). clave desconocida → 404.
func deleteDesvincular(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave := r.PathValue("clave")
		ok, err := svc.Desvincular(r.Context(), clave)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "clave no encontrada en el Portafolio"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"desvinculado": true})
	}
}
