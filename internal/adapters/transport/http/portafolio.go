package httpapi

import (
	"errors"
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

// postObservarBody is the POST /api/portafolio/arneses/{clave}/mapa payload.
type postObservarBody struct {
	InstallPath string `json:"install_path"`
}

// postObservarEnMapa — POST /api/portafolio/arneses/{clave}/mapa: publica una presencia
// YA PERSISTIDA del Portafolio al índice del Mapa (S1-D1) — read-only, jamás registra cwd
// (NO toca ArnesRegistry/arneses.json). 404 clave desconocida · 400 install_path ajeno o
// no cargable · 500 sin índice cableado (subcomando CLI).
func postObservarEnMapa(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave := r.PathValue("clave")
		var body postObservarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		id, err := svc.ObservarEnMapa(r.Context(), clave, body.InstallPath)
		if err != nil {
			switch {
			case errors.Is(err, usecase.ErrObservarClaveNoEncontrada):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrObservarSinIndice):
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			default:
				// ErrObservarInstallPathAjeno o un dir que el loader no pudo cargar: 400-style
				// — el motivo real viaja en el body. Un dir sin manifiesto ya NO cae acá: se
				// observa en modo degradado (S1-D27), jamás un grafo inventado.
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "indexed": true})
	}
}

// postIdentificarBody is the POST /api/portafolio/arneses/{clave}/identificar payload.
// id/nombre opcionales — el usecase cae a defaults (basename del install-path) si vienen "".
// rol/proceso/empresas los declara el operador (B-D1: graph.l0 los exige y el schema no se
// relaja; empresas vacías caen a las de la entrada). marketplace = home autor-declarado
// CRUDO, opcional (S0-D3).
type postIdentificarBody struct {
	InstallPath string   `json:"install_path"`
	ID          string   `json:"id,omitempty"`
	Nombre      string   `json:"nombre,omitempty"`
	Rol         string   `json:"rol,omitempty"`
	Proceso     string   `json:"proceso,omitempty"`
	Empresas    []string `json:"empresas,omitempty"`
	Marketplace string   `json:"marketplace,omitempty"`
}

// postIdentificar — POST /api/portafolio/arneses/{clave}/identificar: escribe el sello
// `arnes.l0.json` IN-SITU en la instalación y re-keya la entrada (S1-D28 + B1). 404 clave
// desconocida · 409 ya sellado · 400 install_path ajeno / path protegido / sello incompleto
// o inválido contra graph.l0 (nada se escribe) / otro. Devuelve la entrada re-keyed (su
// nueva clave viaja en identidad → el FE re-apunta el drawer).
func postIdentificar(svc *usecase.PortafolioService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave := r.PathValue("clave")
		var body postIdentificarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		entrada, err := svc.Identificar(r.Context(), clave, usecase.SolicitudIdentificar{
			InstallPath: body.InstallPath,
			ID:          body.ID,
			Nombre:      body.Nombre,
			Rol:         body.Rol,
			Proceso:     body.Proceso,
			Empresas:    body.Empresas,
			Marketplace: body.Marketplace,
		})
		if err != nil {
			switch {
			case errors.Is(err, usecase.ErrObservarClaveNoEncontrada):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrIdentificarYaSellado):
				writeJSON(w, http.StatusConflict, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrIdentificarSelloIncompleto), errors.Is(err, usecase.ErrIdentificarSelloInvalido):
				// B-D1/RF-B1.2: 400 con el motivo real — el sello jamás se escribe
				// incompleto ni inválido contra graph.l0.
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			default:
				// install_path ajeno, path protegido, o fallo de escritura/re-escaneo — el
				// motivo real viaja en el body.
				writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			}
			return
		}
		writeJSON(w, http.StatusOK, entrada)
	}
}
