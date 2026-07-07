package httpapi

import (
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// arneses.go exposes the arnés→working-directory registry (S2). Registering a path is what
// confines a session's conductor to its arnés's tree instead of a shared cwd
// (boundary permisos-gui `sesion-aislada-por-cwd`). The registry validates the path
// (absolute, existing dir, not a protected location) before storing it.

// listArneses (S2) — the registered arnés→path entries.
func listArneses(reg ports.ArnesRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, reg.List())
	}
}

// registerArnesBody is the PUT /api/arneses/{id} payload.
type registerArnesBody struct {
	Path string `json:"path"`
}

// registerArnes records (or replaces) the working directory for an arnés. A rejected path
// (relative, missing, or protected) returns 400 with the validation reason. onRegistered
// (inyectado por el composition root — el transporte no importa el loader) carga el
// directorio al índice según la nomenclatura: «Cargar» una carpeta = verla en el Mapa.
func registerArnes(reg ports.ArnesRegistry, onRegistered func(id, path string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var body registerArnesBody
		if err := decodeJSON(w, r, &body); err != nil {
			return
		}
		if err := reg.Register(id, body.Path); err != nil {
			writeJSON(w, http.StatusBadRequest, errorBody{Error: err.Error()})
			return
		}
		indexed := true
		var detail string
		if onRegistered != nil {
			if err := onRegistered(id, body.Path); err != nil {
				// Registrar SÍ, indexar NO (p.ej. el dir aún no es un arnés reconocible):
				// honesto en la respuesta, jamás un grafo inventado.
				indexed, detail = false, err.Error()
			}
		}
		writeJSON(w, http.StatusOK, registerArnesResponse{
			ArnesPath: ports.ArnesPath{Arnes: id, Path: body.Path},
			Indexed:   indexed,
			Detail:    detail,
		})
	}
}

// registerArnesResponse es la respuesta del PUT: el registro + si el directorio se
// reconoció e indexó como arnés (nomenclatura-arnes.md).
type registerArnesResponse struct {
	ports.ArnesPath
	Indexed bool   `json:"indexed"`
	Detail  string `json:"detail,omitempty"`
}
