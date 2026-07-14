package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// selfupdate.go expone el self-update sin sudo (paquete boton-actualizar):
// GET /api/version (RF-107, identidad honesta) y POST /api/self-update (RF-104).
// Ambos viven BAJO withAuth (RF-106, superficie-local-confinada); el POST IGNORA el
// cuerpo por completo — cero parámetros: repo y destino los conoce SOLO el daemon.

// reexecGracia es la espera post-respuesta antes del re-exec (RF-105: «≤1s»): da
// tiempo a que el 200 se transmita por loopback antes de que exec mate el proceso.
const reexecGracia = 600 * time.Millisecond

// versionBody es el wire-format de GET /api/version (RF-107; `sucio` aditivo por la
// decisión #7). El transporte posee el JSON — el puerto no sabe de tags.
type versionBody struct {
	Huella      string `json:"huella"`
	Fecha       string `json:"fecha"`
	InstaladoEn string `json:"instalado_en"`
	Escribible  bool   `json:"escribible"`
	Repo        string `json:"repo"`
	Sucio       bool   `json:"sucio"`
}

// repoConfigBody es el wire-format de PUT /api/self-update/repo (RF-109, bugfix
// fix-repo-self-update) — ÚNICO endpoint que acepta un path del request; el disparador
// POST /api/self-update sigue con cero parámetros (RF-106 intacto).
type repoConfigBody struct {
	Path string `json:"path"`
}

// putSelfUpdateRepo — PUT /api/self-update/repo: valida path (mismas reglas del paso
// Verificar) y, solo si pasa, lo fija en caliente + persiste (RF-108/109). Un path
// inválido responde 400 con el motivo exacto — no persiste, no toca el repo activo.
func putSelfUpdateRepo(updates *usecase.SelfUpdateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body repoConfigBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "body inválido: " + err.Error()})
			return
		}
		if body.Path == "" {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: "path requerido"})
			return
		}
		detalle, err := updates.ConfigurarRepo(r.Context(), body.Path)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"detalle": detalle})
	}
}

// getVersion — GET /api/version: la fuente de RF-101 y del polling del reinicio
// (RF-105). La UI jamás inventa estos datos.
func getVersion(updates *usecase.SelfUpdateService) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		v := updates.Version()
		writeJSON(w, http.StatusOK, versionBody{
			Huella:      v.Huella,
			Fecha:       v.Fecha,
			InstaladoEn: v.InstaladoEn,
			Escribible:  v.Escribible,
			Repo:        v.Repo,
			Sucio:       v.Sucio,
		})
	}
}

// postSelfUpdate — POST /api/self-update: corre el flujo SÍNCRONO (la respuesta ES el
// reporte de pasos, RF-104) y, SOLO tras responder «actualizado», agenda el re-exec
// (RF-105). Mapea sentinels: 409 en vuelo · 503 no-actualizable con el motivo.
func postSelfUpdate(updates *usecase.SelfUpdateService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rep, err := updates.Actualizar(r.Context())
		if err != nil {
			status := http.StatusInternalServerError
			switch {
			case errors.Is(err, usecase.ErrActualizacionEnCurso):
				status = http.StatusConflict
			case errors.Is(err, usecase.ErrNoActualizable):
				status = http.StatusServiceUnavailable
			}
			writeJSON(w, status, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, rep)
		if rep.Resultado != usecase.ResultadoActualizado {
			return
		}
		// Respuesta escrita y flusheada — recién AHORA se agenda el re-exec, para que
		// el 200 jamás muera con el proceso (RF-105: responder → flush → agendar).
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		go func() {
			time.Sleep(reexecGracia)
			slog.Info("self-update: re-exec del daemon (binario nuevo instalado)")
			if err := updates.Reiniciar(); err != nil {
				slog.Error("self-update: el re-exec falló — el binario nuevo corre recién al próximo arranque", "err", err)
			}
		}()
	}
}
