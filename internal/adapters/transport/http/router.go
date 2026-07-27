// Package httpapi is the HTTP delivery adapter: it mounts the daemon's routes
// (docs/architecture/contracts/api/openapi.yaml) onto a net/http mux and translates requests to
// use-case calls. It holds no business rules (see
// docs/architecture/boundaries/dominio-independiente-de-transporte.md). The SSE stream is injected
// as a plain http.Handler so this package need not know the concrete broker.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/alpacapurpura/arnesia/internal/ports"
	"github.com/alpacapurpura/arnesia/internal/usecase"
)

// errorBody is the JSON envelope for error responses.
type errorBody struct {
	Error string `json:"error"`
}

// NewHandler builds the daemon's router. maps serves the Map/graph endpoints; sessions
// drives the multisesión Dock; runs is the T3 conductor entry (D2); arneses is the
// arnés→path registry (per-session workdir confinement); portafolio es el Portafolio de
// arneses (Slice 0, S0-D9 — superficie observable sin FE); events is the SSE broker
// mounted at /events. auth confines the whole surface (Host+Origin+token, boundary
// superficie-local-confinada).
func NewHandler(maps *usecase.MapService, sessions *usecase.SessionService, runs *usecase.RunService, fuentes *usecase.FuenteService, arneses ports.ArnesRegistry, conf ports.ConformancePort, confBase func(id string) string, onArnesRegistered func(id, path string) error, updates *usecase.SelfUpdateService, portafolio *usecase.PortafolioService, marketplaces *usecase.MarketplaceService, dictado *usecase.DictadoService, telemetria *usecase.TelemetriaService, otlp http.Handler, ui http.Handler, events http.Handler, auth AuthConfig) http.Handler {
	mux := http.NewServeMux()

	// Telemetría embebida (paquete 2026-07-24-telemetria-embebida-otel). El receptor OTLP
	// entra **inyectado como http.Handler**, igual que el broker SSE: este paquete NO importa
	// `telemetria/otlp` — go-arch-lint lo prohíbe y no hace falta.
	//
	// `/v1/logs` y `/v1/metrics` viven fuera de `/api` porque la ruta la fija la spec del
	// protocolo, no nosotros. Quedan igual bajo los gates (isAPIPath los cubre).
	if otlp != nil {
		mux.Handle("POST /v1/logs", otlp)
		mux.Handle("POST /v1/metrics", otlp)
	}
	if telemetria != nil {
		mux.HandleFunc("GET /api/telemetria/resumen", getResumen(telemetria))
		mux.HandleFunc("GET /api/telemetria/salud", getSalud(telemetria))
		mux.HandleFunc("GET /api/telemetria/portafolio", getPortafolio(telemetria))
		mux.HandleFunc("GET /api/telemetria/arneses/{clave}/cajas", getCajas(telemetria))
		mux.HandleFunc("GET /api/telemetria/arneses/{clave}/cajas/{cajaId}", getDetalleCaja(telemetria))
		mux.HandleFunc("GET /api/telemetria/arneses/{clave}/mejoras", getMejoras(telemetria))
		mux.HandleFunc("DELETE /api/telemetria/arneses/{clave}", deleteTelemetriaArnes(telemetria))
		// D26.4 — el descarte de un punto de mejora, con su vuelta atrás. Son DOS rutas
		// porque son dos decisiones distintas del operador, y la segunda existe para que un
		// clic distraído no borre para siempre un hallazgo que costó dinero producir.
		mux.HandleFunc("POST /api/telemetria/arneses/{clave}/mejoras/{puntoId}/descartar",
			postDescartarPunto(telemetria))
		mux.HandleFunc("DELETE /api/telemetria/arneses/{clave}/mejoras/{puntoId}/descartar",
			deleteDescartarPunto(telemetria))
		mux.HandleFunc("POST /api/telemetria/proceso", postProceso(telemetria, nil))
	}

	// UI embebida (HS-11): el daemon sirve la SPA en "/" cuando el build la trae
	// (scripts/bundle.sh compila web/dist ANTES del daemon). Queda DENTRO de withAuth:
	// gates Host+Origin aplican; el token solo protege /api|/events (isAPIPath — la
	// SPA debe cargar antes de tener token). Un build de dev sin dist responde honesto.
	if ui != nil {
		mux.Handle("/", ui)
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "arnesia: este build no embebe la UI (scripts/bundle.sh la incluye); la API vive en /api", http.StatusNotFound)
		})
	}

	// Liveness + the multiplexed SSE stream (the two endpoints the shell polls first).
	mux.HandleFunc("GET /healthz", healthz)
	mux.Handle("GET /events", events)
	mux.Handle("GET /api/events", events) // OpenAPI server base is /api.

	// Self-update sin sudo (paquete boton-actualizar, RF-104..107). BAJO withAuth como
	// todo /api/* (RF-106); el POST no acepta parámetros — cero rutas del request.
	mux.HandleFunc("GET /api/version", getVersion(updates))
	mux.HandleFunc("POST /api/self-update", postSelfUpdate(updates))
	// PUT /api/self-update/repo (RF-108/109, bugfix fix-repo-self-update): ÚNICO
	// endpoint que acepta un path — configura, no dispara; POST /api/self-update arriba
	// sigue sin parámetros.
	mux.HandleFunc("PUT /api/self-update/repo", putSelfUpdateRepo(updates))

	// Portfolio / Map / Inspector / Runs (S1–S3, S8).
	mux.HandleFunc("GET /api/harnesses", listHarnesses(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/graph", getHarnessGraph(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/nodes/{nodeId}", getNode(maps))
	mux.HandleFunc("GET /api/harnesses/{id}/nodes/{nodeId}/fuente", getNodeFuente(fuentes))
	mux.HandleFunc("GET /api/harnesses/{id}/runs", listRuns)
	mux.HandleFunc("GET /api/harnesses/{id}/conformance", getConformance(maps, conf, confBase))

	// Conductor T3 (Fase E, D2): correr una caja = recurso propio, NO un turno del Dock.
	// Async (deuda BACKLOG «run async», 2026-07-23): 202+run_id inmediato; el desenlace
	// se consulta en runs/{runId} (progreso vivo sigue en /events, event: run).
	mux.HandleFunc("POST /api/harnesses/{id}/boxes/{boxId}/run", runBox(runs))
	mux.HandleFunc("GET /api/harnesses/{id}/boxes/{boxId}/runs/{runId}", getRun(runs))

	// Arnés registry (S2) — maps an arnés to the working dir its sessions run claude in.
	mux.HandleFunc("GET /api/arneses", listArneses(arneses))
	mux.HandleFunc("PUT /api/arneses/{id}", registerArnes(arneses, onArnesRegistered))

	// Portafolio de arneses (Slice 0, S0-D9): superficie observable sin FE — Slice 1 la consume.
	mux.HandleFunc("GET /api/portafolio", listPortafolio(portafolio))
	mux.HandleFunc("POST /api/portafolio/escaneos", postEscanear(portafolio))
	mux.HandleFunc("POST /api/portafolio/proyectos", postAgregar(portafolio))
	mux.HandleFunc("DELETE /api/portafolio/arneses/{clave}", deleteDesvincular(portafolio))
	// Observar en Mapa (Slice 1, S1-D1, cierra GAP-1): publica una presencia YA
	// PERSISTIDA al índice del Mapa, read-only — NO registra cwd.
	mux.HandleFunc("POST /api/portafolio/arneses/{clave}/mapa", postObservarEnMapa(portafolio))
	mux.HandleFunc("POST /api/portafolio/arneses/{clave}/identificar", postIdentificar(portafolio))

	// Plano Marketplaces + catálogo (paquete 2026-07-23-portafolio-agregar-marketplace, AG-D8):
	// el estante de lo que vendemos y el espejo de si el cliente coincide. Lectura de catálogo
	// CACHEADA con refresco explícito (BR-3); `entradas: null` cuando no se pudo leer (BR-4).
	mux.HandleFunc("GET /api/marketplaces", listMarketplaces(marketplaces))
	mux.HandleFunc("POST /api/marketplaces", postRegistrarMarketplace(marketplaces))
	mux.HandleFunc("POST /api/marketplaces/validaciones", postValidarMarketplace(marketplaces))
	mux.HandleFunc("DELETE /api/marketplaces/{nombre}", deleteOlvidarMarketplace(marketplaces))
	mux.HandleFunc("GET /api/marketplaces/{nombre}/catalogo", getCatalogoMarketplace(marketplaces))
	mux.HandleFunc("POST /api/marketplaces/{nombre}/lecturas", postLeerCatalogo(marketplaces))
	// `↧ Traer canónico` (AG-D17): materializa una fila del catálogo como canónico editable en
	// `~/.arnesia/checkouts/`. Nunca escribe en `~/.claude` (BR-13).
	mux.HandleFunc("POST /api/marketplaces/{nombre}/traidos", postTraerCanonico(marketplaces))
	// Reconciliación de origen (S7): actúa sobre un ARNÉS, por eso vive bajo /portafolio.
	mux.HandleFunc("GET /api/portafolio/arneses/{clave}/origen/candidatos", getCandidatosOrigen(marketplaces))
	mux.HandleFunc("POST /api/portafolio/arneses/{clave}/origen", postAsignarOrigen(portafolio))

	// Multisesión + Dock (S4). Every conductor turn streams back over /events.
	mux.HandleFunc("GET /api/sessions", listSessions(sessions))
	mux.HandleFunc("POST /api/sessions", createSession(sessions, arneses, onArnesRegistered))
	// El panel de conversaciones (RF-340…RF-344). Van ANTES de `GET /api/sessions/{id}`:
	// el patrón más específico tiene que quedar registrado primero para que el mux no
	// ambigüe. `GET /api/sessions/cerradas/{id}/historial` se RETIRA acá (RF-345 CA-2):
	// bajo el modelo nuevo el transcript de una conversación archivada viaja en su propio
	// registro y no hay nada que reconstruir. `HistorialCerrada` y el lector del corpus
	// nativo se CONSERVAN como capacidad del dominio, sin ruta: siguen siendo el único
	// fallback para lo archivado ANTES de la migración, que no tiene transcript propio.
	mux.HandleFunc("GET /api/sessions/{id}/conversaciones", listarConversaciones(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/conversaciones", crearConversacion(sessions))
	mux.HandleFunc("PATCH /api/sessions/{id}/conversaciones/{cid}", renombrarConversacion(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/conversaciones/{cid}/activar", activarConversacion(sessions))
	mux.HandleFunc("GET /api/sessions/{id}", getSession(sessions))
	mux.HandleFunc("PATCH /api/sessions/{id}", patchSession(sessions))
	mux.HandleFunc("DELETE /api/sessions/{id}", deleteSession(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/turn", sessionTurn(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/permission", resolvePermission(sessions))
	mux.HandleFunc("POST /api/sessions/{id}/interrupt", sessionInterrupt(sessions))

	// Dictado por voz (paquete 2026-07-25-spike-voz-dictado, RF-222/RF-223). El
	// `disponibilidad` NO cuelga de una sesión: el FE lo consulta al montar el composer
	// para saber si ofrece el botón, antes de que haya nada que dictar.
	if dictado != nil {
		mux.HandleFunc("POST /api/sessions/{id}/dictado", postDictado(dictado))
		mux.HandleFunc("GET /api/dictado/disponibilidad", getDisponibilidad(dictado))
	}

	// Diagnóstico del FE (RF-230). NO va condicionado a ningún servicio: es la vía por la que
	// el WebView —que no tiene devtools ni escribe a disco— deja rastro de sus fallos, y el
	// primero que hay que poder diagnosticar es el arranque, cuando todavía no hay nada más.
	mux.HandleFunc("POST /api/diagnostico", postDiagnostico())

	return withAuth(auth, mux)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// harnessSummary is one row of GET /api/harnesses — the lightweight portfolio entry the Map
// picker (RF-72) consumes; the full graph is a separate call (getHarnessGraph).
type harnessSummary struct {
	ID       string   `json:"id"`
	Rol      string   `json:"rol,omitempty"`
	Proceso  string   `json:"proceso,omitempty"`
	Empresas []string `json:"empresas,omitempty"` // S0-D3: facet N:M, ya no escalar (portafolio/T1).
}

// listHarnesses (S1) — portfolio, from the index (RF-72).
func listHarnesses(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gs, err := maps.Harnesses(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorBody{Error: err.Error()})
			return
		}
		out := make([]harnessSummary, 0, len(gs))
		for _, g := range gs {
			if g.Arnes == nil {
				continue
			}
			out = append(out, harnessSummary{
				ID:       g.Arnes.ID,
				Rol:      g.Arnes.Rol,
				Proceso:  g.Arnes.Proceso,
				Empresas: g.Arnes.Empresas,
			})
		}
		writeJSON(w, http.StatusOK, out)
	}
}

// getHarnessGraph (S2) — the agnostic graph of one harness (real, from the index).
func getHarnessGraph(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		g, err := maps.Graph(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, g)
	}
}

// getNode (S3) — inspector: one node (Box + fused Contract) of a harness (RF-71).
func getNode(maps *usecase.MapService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		box, ok, err := maps.Node(r.Context(), r.PathValue("id"), r.PathValue("nodeId"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorBody{Error: err.Error()})
			return
		}
		if !ok {
			writeJSON(w, http.StatusNotFound, errorBody{Error: "node not found"})
			return
		}
		writeJSON(w, http.StatusOK, box)
	}
}

// listRuns (S8) — runs derived from the JSONL. Stub.
func listRuns(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w)
}

func notImplemented(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotImplemented, errorBody{Error: "not implemented (fase 5)"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("http: encode response", "err", err)
	}
}
