package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// telemetria.go son los handlers de `/api/telemetria/*`.
//
// **Contrato transversal, y es el que un implementador «prolijo» rompería:** `null` y `0`
// significan cosas distintas en TODA esta superficie. Convertir un `null` en `0` es el pass
// fabricado que la doctrina prohíbe. Por eso los DTO usan punteros y por eso las ausencias
// viajan explícitas en vez de omitirse.

// ventanaDeQuery lee `desde`/`hasta` en RFC3339. Un valor ilegible NO se ignora en silencio:
// se responde 400. Ignorarlo devolvería una ventana distinta de la pedida y el usuario leería
// cifras de otro rango creyendo que son del suyo.
func ventanaDeQuery(r *http.Request) (ports.ConsultaTelemetria, error) {
	q := ports.ConsultaTelemetria{
		ArnesID:       r.URL.Query().Get("arnes"),
		InstalacionID: r.URL.Query().Get("instalacion"),
		SesionID:      r.URL.Query().Get("sesion"),
	}
	if v := r.URL.Query().Get("desde"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return q, errors.New("desde: se espera RFC3339 (ej. 2026-07-26T00:00:00Z)")
		}
		q.Desde = t
	}
	if v := r.URL.Query().Get("hasta"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return q, errors.New("hasta: se espera RFC3339 (ej. 2026-07-26T23:59:59Z)")
		}
		q.Hasta = t
	}
	if v := r.URL.Query().Get("limite"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return q, errors.New("limite: se espera un entero no negativo")
		}
		q.Limite = n
	}
	return q, nil
}

// getResumen sirve la franja de la barra: total con su cobertura y su confianza.
func getResumen(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := ventanaDeQuery(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		out, err := svc.Resumen(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// getCajas devuelve el gasto por caja. **Incluye las cajas SIN dato**, con `atribuible:false`,
// `motivo` no vacío y `costo_micros: null` — omitirlas obligaría al FE a inventar por qué
// faltan, y un 0 diría que corrieron gratis.
func getCajas(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := ventanaDeQuery(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		q.ArnesID = r.PathValue("clave")
		cajas, err := svc.PorCaja(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		if cajas == nil {
			cajas = []domain.GastoCaja{} // lista vacía, no `null`: el FE no tiene que distinguir.
		}
		writeJSON(w, http.StatusOK, cajas)
	}
}

// getDetalleCaja es la 4ª tab del inspector.
func getDetalleCaja(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := ventanaDeQuery(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		q.ArnesID = r.PathValue("clave")
		q.CajaID = r.PathValue("cajaId")
		out, err := svc.DetalleCaja(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// getMejoras devuelve las TRES listas siempre (A16).
func getMejoras(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := ventanaDeQuery(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		q.ArnesID = r.PathValue("clave")
		out, err := svc.Mejoras(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// getPortafolio devuelve una fila por (arnés, instalación), con `puesto: null` cuando el
// arnés no declara rol (D20).
func getPortafolio(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := ventanaDeQuery(r)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		filas, err := svc.Portafolio(r.Context(), q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		if filas == nil {
			filas = []domain.FilaPortafolio{}
		}
		writeJSON(w, http.StatusOK, filas)
	}
}

// borradoResponse es lo que devuelve el botón de borrado.
type borradoResponse struct {
	Borrados int64 `json:"borrados"`
	// Ventana dice si el borrado fue acotado. Sin este campo, «borrados: 61» no distingue
	// «borré los 61 de la ventana» de «borré 61, que era todo lo que había».
	Ventana bool `json:"ventana"`
}

// deleteTelemetriaArnes borra la telemetría de un arnés — detalle **y** agregado.
//
// Acepta `desde`/`hasta` (D26.5 · A-4): **el borrado se acota a la ventana que la confirmación
// declara**. Sin ellos borra todo el historial, que es el comportamiento anterior y sigue
// siendo válido — pero ahora es una elección del llamador, no la única opción.
func deleteTelemetriaArnes(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave := r.PathValue("clave")
		if clave == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "falta la clave del arnés"})
			return
		}
		q, qerr := ventanaDeQuery(r)
		if qerr != nil {
			// Una ventana ilegible en un borrado **no se ignora**: ignorarla borraría todo
			// cuando el usuario pidió una parte.
			writeJSON(w, http.StatusBadRequest, errorBody{Error: qerr.Error()})
			return
		}
		n, err := svc.BorrarArnes(r.Context(), clave, q.Desde, q.Hasta)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, borradoResponse{Borrados: n, Ventana: !q.Desde.IsZero() || !q.Hasta.IsZero()})
	}
}

// getSalud reporta el estado del módulo, con el TTL **rotulado como propuesto** mientras el
// número no esté firmado (J-6).
func getSalud(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, err := svc.Salud(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// postProceso es la ingesta del hook de S2. El cuerpo es un evento **ya proyectado**; el
// daemon **lo re-valida igual** (A13: la allowlist se aplica dos veces, y no se confía en que
// el emisor haya filtrado aunque el emisor sea nuestro propio binario).
func postProceso(svc *usecase.TelemetriaService, revalidar func(domain.EventoTelemetria) domain.EventoTelemetria) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ev domain.EventoTelemetria
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10)).Decode(&ev); err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: domain.ErrPayloadInvalido.Error()})
			return
		}
		if ev.SesionID == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: domain.ErrEventoSinSesion.Error()})
			return
		}
		if revalidar != nil {
			ev = revalidar(ev)
		}
		if _, err := svc.Ingerir(r.Context(), []domain.EventoTelemetria{ev}); err != nil {
			// Ni siquiera acá se devuelve 5xx: el hook no reintenta y un error del servidor
			// solo lo haría esperar. El descarte ya quedó contado en la salud.
			writeJSON(w, http.StatusAccepted, borradoResponse{})
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

// descarteResponse confirma la decisión. Lleva el estado resultante y no solo un 200: la UI
// tiene que poder pintar el botón de vuelta atrás sin adivinar en qué quedó la cosa.
type descarteResponse struct {
	Punto      string `json:"punto"`
	Descartado bool   `json:"descartado"`
}

// postDescartarPunto guarda que el operador no quiere volver a ver un punto (D26.4).
//
// Hasta acá el botón nacía `disabled` y lo decía —honesto, pero no era la afordancia (A-1)—
// porque **no existía dónde guardar el descarte**. Ahora existe, y persiste: un descarte que
// se pierde al reiniciar no es un descarte.
func postDescartarPunto(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave, punto := r.PathValue("clave"), r.PathValue("puntoId")
		if clave == "" || punto == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "falta el arnés o el punto"})
			return
		}
		if err := svc.DescartarPunto(r.Context(), clave, punto); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, descarteResponse{Punto: punto, Descartado: true})
	}
}

// deleteDescartarPunto deshace el descarte. Es la vuelta atrás que la tarjeta promete.
func deleteDescartarPunto(svc *usecase.TelemetriaService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clave, punto := r.PathValue("clave"), r.PathValue("puntoId")
		if clave == "" || punto == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "falta el arnés o el punto"})
			return
		}
		if err := svc.RecuperarPunto(r.Context(), clave, punto); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, descarteResponse{Punto: punto, Descartado: false})
	}
}
