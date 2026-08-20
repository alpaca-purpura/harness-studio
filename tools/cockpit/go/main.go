// cockpit-go — Cockpit SDD como binario standalone (cero deps npm en runtime).
//
// Sirve la UI Next.js (static export embebido vía go:embed) + implementa las
// API routes del cockpit Next.js original (paridad JSON, filesystem-as-DB).
//
// Uso:
//   cockpit -workspace ~/Proyectos/demo-environment -port 4000
//   WORKSPACE_ROOT=~/Proyectos/demo-environment ./cockpit
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed all:ui
var uiFS embed.FS

// version del binario · inyectada en build por build.sh vía
// -ldflags "-X main.version=$KIT_VERSION" (lee core-harness/VERSION).
var version = "dev"

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string, extra map[string]any) {
	payload := map[string]any{"error": message}
	for k, v := range extra {
		payload[k] = v
	}
	writeJSON(w, status, payload)
}

// uiHandler sirve el static export con fallback estilo trailingSlash
// (/board → board/index.html · /board/ → ídem).
func uiHandler() http.Handler {
	sub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		log.Fatalf("ui embebida no disponible: %v", err)
	}
	fileServer := http.FileServer(http.FS(sub))
	serveHTML := func(w http.ResponseWriter, name string, status int) bool {
		data, err := fs.ReadFile(sub, name)
		if err != nil {
			return false
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(status)
		_, _ = w.Write(data)
		return true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.Trim(strings.TrimPrefix(r.URL.Path, "/"), "/")
		if p == "" {
			p = "index.html"
		}
		// Página: / · /board · /board/ → {p}/index.html servido directo
		// (FileServer redirige index.html en 301-loop con trailingSlash).
		if strings.HasSuffix(p, ".html") {
			if serveHTML(w, p, http.StatusOK) {
				return
			}
		}
		if st, err := fs.Stat(sub, p); err == nil && st.IsDir() {
			if serveHTML(w, p+"/index.html", http.StatusOK) {
				return
			}
		}
		if _, err := fs.Stat(sub, p); err != nil {
			if serveHTML(w, p+"/index.html", http.StatusOK) {
				return
			}
			if serveHTML(w, "404.html", http.StatusNotFound) {
				return
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	runCLI()
}

// serve arranca el server HTTP. workspace != "" o WORKSPACE_ROOT → modo single;
// si no, registry con proyectos activos → modo multi-workspace (sistemas proyecto/sistema);
// fallback: auto-detección single (git toplevel).
func serve(port int, workspace string) {
	if env := os.Getenv("PORT"); env != "" && port == 4000 {
		fmt.Sscanf(env, "%d", &port)
	}

	var modeDesc string
	if workspace != "" {
		os.Setenv("WORKSPACE_ROOT", workspace)
	}
	if workspace != "" || os.Getenv("WORKSPACE_ROOT") != "" {
		root, err := getWorkspaceRoot()
		if err != nil {
			log.Fatalf("✗ %v", err)
		}
		modeDesc = "workspace=" + root
	} else if reg := loadRegistry(); len(reg.activeProjects()) > 0 {
		active := reg.activeProjects()
		enableMultiMode(active)
		enableSystemWorkspaces(reg.Projects) // Fase 3: reenvío de sistemas cross-repo (TODOS, incl. inactivos)
		names := make([]string, len(active))
		for i, p := range active {
			names[i] = p.Name
		}
		modeDesc = fmt.Sprintf("multi-workspace=%v", names)
	} else {
		root, err := getWorkspaceRoot()
		if err != nil {
			log.Fatalf("✗ %v (o registrá proyectos: cockpit add PATH)", err)
		}
		modeDesc = "workspace=" + root
	}

	mux := http.NewServeMux()

	// Lectura
	mux.HandleFunc("GET /api/sistemas", handleSistemas)
	mux.HandleFunc("GET /api/stories", handleStoriesList)
	mux.HandleFunc("GET /api/stories/{id}", handleStorySingle)
	mux.HandleFunc("PATCH /api/stories/{id}", handleStorySingle)
	mux.HandleFunc("GET /api/capabilities", handleCapabilitiesList)
	mux.HandleFunc("GET /api/capabilities/status", capArtifactHandler("_status-computed.json", "status",
		"Ejecuta: python3 scripts/compute_capability_status.py --sistema {sistema}"))
	mux.HandleFunc("GET /api/capabilities/bidirectional", capArtifactHandler("_bidirectional-validation.json", "validation",
		"Ejecuta: python3 scripts/validate_code_cap_bidirectional.py --sistema {sistema}"))
	mux.HandleFunc("GET /api/capabilities/code-index", capArtifactHandler("_code-index.json", "index",
		"Ejecuta: python3 scripts/generate_code_to_cap_index.py --sistema {sistema}"))
	mux.HandleFunc("GET /api/capabilities/doctor", handleCapDoctor)
	mux.HandleFunc("GET /api/capabilities/{module}/{cap}", handleCapabilitySingle)
	mux.HandleFunc("PATCH /api/capabilities/{module}/{cap}", handleCapabilitySingle)
	mux.HandleFunc("/api/releases", handleReleases)
	mux.HandleFunc("/api/learnings", handleLearningsV2)
	mux.HandleFunc("GET /api/system-map", handleSystemMap)
	mux.HandleFunc("GET /api/ledger", handleLedger)
	mux.HandleFunc("GET /api/value-stream", handleValueStream)
	mux.HandleFunc("GET /api/sessions", handleSessions)
	mux.HandleFunc("GET /api/harness", handleHarness)
	mux.HandleFunc("GET /api/nav-config", handleNavConfig)
	mux.HandleFunc("GET /api/cil", handleCIL)
	mux.HandleFunc("GET /api/watch", handleWatch)
	mux.HandleFunc("GET /api/operator-input/{storyId}", handleOperatorInput)
	mux.HandleFunc("PATCH /api/operator-input/{storyId}", handleOperatorInput)
	mux.HandleFunc("GET /api/torre", handleTorre)
	mux.HandleFunc("GET /api/proceso", handleProceso) // descriptor de proceso (I-77 · RN-32)
	mux.HandleFunc("/api/file", handleFile)

	// Cockpit de delivery (F5 · DH-08 · RN-39/RN-40): plantilla Go-nativa + proxy al sidecar
	mux.HandleFunc("GET /api/delivery/salud", handleDeliverySalud)
	mux.HandleFunc("GET /api/delivery/plantilla", handleDeliveryPlantilla)
	mux.HandleFunc("GET /api/delivery/sesiones", handleDeliverySesionesList)
	mux.HandleFunc("GET /api/delivery/sesiones/{id}", handleDeliverySesionUna)
	mux.HandleFunc("POST /api/delivery/sesiones", handleDeliverySesionCrear)

	// Escritura / acciones
	mux.HandleFunc("POST /api/transition", handleTransition)
	mux.HandleFunc("POST /api/operator-verify", handleOperatorVerify)
	mux.HandleFunc("POST /api/merge-release", handleMergeRelease)
	mux.HandleFunc("POST /api/open", handleOpen)
	mux.HandleFunc("POST /api/extend-cap", handleExtendCap)
	mux.HandleFunc("POST /api/from-done", handleFromDone)
	mux.HandleFunc("POST /api/story/new", handleStoryNew)
	mux.HandleFunc("POST /api/capabilities/regen", handleCapRegen)
	mux.HandleFunc("GET /api/gherkin-status", handleGherkinStatus)
	mux.HandleFunc("POST /api/refs/upload", handleRefsUpload)

	// UI estática (catch-all)
	mux.Handle("/", uiHandler())

	// Torre de control: el gate de fábrica se mide al BOOT + por ?medir=gate_fabrica,
	// jamás en un loop (RN-16). Async: los checks declarados son caros por diseño.
	if reg := loadRegistry(); len(reg.Projects) > 0 {
		go torreBootGates(reg)
	}

	// Sidecar de delivery (RN-35): un solo comando sirve TODO — UI + API + sesiones.
	go deliveryBoot()

	addr := fmt.Sprintf(":%d", port)
	log.Printf("🧭 cockpit-go · %s · http://localhost%s", modeDesc, addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
