package httpapi

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// maxAudio caps the request body of a dictation.
//
// La cota FUNCIONAL es el tope de 3 min de grabación (RF-218), pero vive en el FE — y el
// endpoint no puede confiar en que el cliente la respetó. 24 MB deja pasar de sobra 3 min de
// `audio/mp4` (que ronda 1-2 MB/min) y le pone techo a lo que un POST cualquiera contra la
// superficie local puede hacerle crecer a la memoria del daemon.
const maxAudio = 24 << 20

// mimeDefault is what we assume when the client sent no Content-Type.
//
// Desde RF-229 el FE arma el WAV él mismo (`MediaRecorder` de WebKitGTK dice soportar
// `audio/mp4` y entrega 0 bytes — ver `web/src/shared/lib/wav.ts`), así que el formato real
// que sube la app es `audio/wav`. El endpoint SIGUE aceptando cualquier mime que el motor
// soporte: el default es para un cliente que no lo declara, no una restricción.
const mimeDefault = "audio/wav"

// postDictado (RF-222) recibe el audio grabado de una sesión y devuelve el texto listo para
// poblar el composer.
//
// El cuerpo es el audio crudo (no multipart, no base64): es un blob de una sola pieza que va
// derecho al motor, y envolverlo solo agregaría una copia y una decodificación.
//
// La respuesta SIEMPRE dice si el texto quedó `limpio` o `crudo`.
func postDictado(svc *usecase.DictadoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		// Se acota ANTES de leer: MaxBytesReader corta el stream en vez de tragarse el
		// cuerpo entero y medirlo después.
		r.Body = http.MaxBytesReader(w, r.Body, maxAudio)
		audio, err := io.ReadAll(r.Body)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				writeJSON(w, http.StatusRequestEntityTooLarge, errorBody{
					Error: "la grabación es demasiado grande",
				})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "no pude leer la grabación: " + err.Error()})
			return
		}
		if len(audio) == 0 {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "la grabación vino vacía"})
			return
		}

		mime := strings.TrimSpace(r.Header.Get("Content-Type"))
		if mime == "" {
			mime = mimeDefault
		}

		// Se cronometra el POST entero (RF-230): la cadena mide 2.2 s de STT contra 13-17 s
		// de limpieza (§1.8), así que el número es lo que distingue «tardó lo que tenía que
		// tardar» de «algo se colgó» sin tener que reproducir el dictado.
		arranque := time.Now()
		dic, err := svc.Dictar(r.Context(), id, audio, mime)
		if err != nil {
			// El detalle va acá y no solo en la respuesta: el operador ve una frase, el log
			// tiene que tener con qué arreglarlo (qué formato, cuántos bytes, cuánto tardó).
			slog.Error("dictado",
				"sesion", id, "mime", mime, "bytes", len(audio),
				"ms", time.Since(arranque).Milliseconds(), "err", err)
			switch {
			case errors.Is(err, usecase.ErrSesionDesconocida):
				writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrDictadoVacio):
				// No se entendió nada. NO es un 500: el pipeline anduvo, el audio no
				// tenía voz. El FE avisa y no toca el composer (RF-227).
				writeJSON(w, http.StatusUnprocessableEntity, errorBody{Error: err.Error()})
			case errors.Is(err, usecase.ErrDictadoNoDisponible):
				// 503, no 500: falta una capacidad (un motor STT), no se rompió nada.
				// Devolver 500 mandaría al operador a buscar un bug donde falta un
				// `apt install`.
				writeJSON(w, http.StatusServiceUnavailable, errorBody{Error: err.Error()})
			default:
				writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			}
			return
		}
		// También el camino feliz: sin la línea de éxito no hay con qué comparar cuando algo
		// empieza a fallar, y `estado=crudo` es una degradación que conviene ver acumularse.
		slog.Info("dictado",
			"sesion", id, "mime", mime, "bytes", len(audio),
			"ms", time.Since(arranque).Milliseconds(),
			"estado", dic.Estado, "motor", dic.Motor, "chars", len(dic.Texto), "motivo", dic.Motivo)
		writeJSON(w, http.StatusOK, dic)
	}
}

// getDisponibilidad (RF-223/RF-227) dice si se puede dictar y, si no, POR QUÉ y qué instalar.
//
// El FE la consulta al montar el composer para decidir si ofrece el botón. Sin esto la única
// forma de enterarse de que falta el motor sería grabar tres minutos y fallar al final.
func getDisponibilidad(svc *usecase.DictadoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, svc.Disponibilidad(r.Context()))
	}
}
