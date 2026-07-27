package httpapi

// Las cinco rutas del panel de conversaciones (T18 · RF-340…RF-345 · arquitectura §5.1).
//
// Los handlers de `sessions.go` no se tocan salvo `listSessions`, que deja de ser bimorfo.
// El registro en el mux va ANTES de `GET /api/sessions/{id}` para que el patrón más
// específico gane — sin eso, `/api/sessions/{id}` se traga `/api/sessions/{id}/conversaciones`
// según el orden en que el mux resuelva.

import (
	"errors"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// errorDeConversacion traduce un error del usecase a su código HTTP. Está en UN solo lugar
// porque las cuatro rutas que mutan comparten exactamente las mismas cinco salidas, y tener
// la tabla repetida cuatro veces sería tener cuatro tablas que se van a separar.
//
//   - turno en vuelo        → 409, el MISMO criterio que `POST /turn` usa hoy: no es un
//     fallo, es un conflicto con algo que ya está pasando.
//   - conversación ajena    → 404. Nunca se busca fuera de la sesión (BR-CV-2).
//   - sesión inexistente    → 404.
//   - título vacío          → 400, y el título anterior no cambia.
//   - registro solo-lectura → 503. No es culpa del pedido: el daemon no puede escribir.
//   - cualquier otra cosa   → 500 con el motivo del filesystem, que es lo que el operador
//     necesita leer para saber si se le llenó el disco.
func errorDeConversacion(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, usecase.ErrBusy):
		writeJSON(w, http.StatusConflict, errorBody{Error: err.Error()})
	case errors.Is(err, domain.ErrConvNoEncontrada), errors.Is(err, usecase.ErrSesionNoEncontrada):
		writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
	case errors.Is(err, domain.ErrTituloVacio):
		writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
	case errors.Is(err, usecase.ErrSoloLectura):
		writeJSON(w, http.StatusServiceUnavailable, errorBody{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
	}
}

// listarConversaciones (RF-340/RF-341) — la lista del panel, sólo de ESTA sesión (CV-D4).
// Nunca trae los turnos. Con `?q=` sólo viajan las que coinciden, cada una con su fragmento,
// y `total` sigue siendo el total de la sesión: es el denominador del «N de M coinciden».
func listarConversaciones(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		convs, total, err := svc.Conversaciones(r.PathValue("id"), r.URL.Query().Get("q"))
		if err != nil {
			errorDeConversacion(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"conversaciones": convs, "total": total})
	}
}

// crearConversacion (RF-342) — UNA operación que desactiva la anterior y activa la nueva.
// `desactivada` viaja explícito (y `null` cuando no había ninguna): el FE necesita nombrarla
// en el vacío del transcript nuevo, y deducirla de la lista anterior sería adivinarla.
func crearConversacion(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nueva, desactivada, err := svc.CrearConversacion(r.PathValue("id"))
		if err != nil {
			errorDeConversacion(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{"nueva": nueva, "desactivada": idONulo(desactivada)})
	}
}

// activarConversacion (RF-343) — retomar. NO spawnea: el conductor arranca perezoso, en el
// turno siguiente, que es donde el `--resume` ocurre de verdad.
func activarConversacion(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		activada, desactivada, err := svc.ActivarConversacion(r.PathValue("id"), r.PathValue("cid"))
		if err != nil {
			errorDeConversacion(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"activada": activada, "desactivada": idONulo(desactivada)})
	}
}

// renombrarBody es el cuerpo del PATCH. `titulo` es un puntero para distinguir «no lo mandé»
// de «lo mandé vacío»: lo segundo es un 400 con el título intacto, lo primero también, pero
// por otro motivo, y el mensaje tiene que poder decir cuál de los dos fue.
type renombrarBody struct {
	Titulo *string `json:"titulo"`
}

// renombrarConversacion (RF-344). Los títulos duplicados son legales: no hay unicidad que
// validar, y dos hilos sobre el mismo tema con el mismo nombre es una situación real.
func renombrarConversacion(svc *usecase.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body renombrarBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if body.Titulo == nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "falta `titulo` en el cuerpo"})
			return
		}
		c, err := svc.RenombrarConversacion(r.PathValue("id"), r.PathValue("cid"), *body.Titulo)
		if err != nil {
			errorDeConversacion(w, err)
			return
		}
		writeJSON(w, http.StatusOK, c)
	}
}

// idONulo devuelve nil para el id vacío. `""` y «no había ninguna» son cosas distintas y el
// wire tiene que poder decirlas distinto (boundary no-aplica-no-es-cero).
func idONulo(id string) any {
	if id == "" {
		return nil
	}
	return id
}
