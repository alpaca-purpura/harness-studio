// handlers_regen.go — regen on-demand de los índices cap↔código (2026-06-11).
//
// El tab Drift/Functionality lee `_code-index.json` + `_bidirectional-validation.json`
// PRE-GENERADOS (los regenera el pre-commit al tocar caps/headers). Entre edición
// y commit el cockpit muestra el índice viejo. Este endpoint cierra esa ventana:
// shellea los scripts python CANÓNICOS del workspace (un solo SSoT — NUNCA
// reimplementar el resolver en Go: un mirror driftearía).
//
// Si el workspace no trae los scripts (proyecto cliente sin ese tooling), el
// endpoint responde hint honesto en vez de fallar.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var regenMu sync.Mutex

// pythonBin: venv del workspace primero (resolución de deps del repo), intérprete del
// PATH como fallback.
//
// El venv tiene layout distinto por OS (`Scripts\python.exe` en Windows, `bin/python` en
// POSIX) — se sondean ambos. Y el fallback NO puede ser `python3` a secas en Windows: el
// `python3` de WindowsApps es un stub que existe en el PATH y no ejecuta nada (mismo
// problema que ya resuelve scripts/installer.ps1 en el repo host), así que el candidato
// se valida EJECUTÁNDOLO antes de devolverlo.
func pythonBin(root string) string {
	for _, rel := range [][]string{{".venv", "Scripts", "python.exe"}, {".venv", "bin", "python"}} {
		venv := filepath.Join(append([]string{root}, rel...)...)
		if st, err := os.Stat(venv); err == nil && !st.IsDir() {
			return venv
		}
	}
	for _, cand := range []string{"python3", "python"} {
		if exec.Command(cand, "-c", "pass").Run() == nil {
			return cand
		}
	}
	return "python3" // ninguno respondió: fallar con el nombre canónico y mensaje claro.
}

func handleCapRegen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Sistema string `json:"sistema"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if !containsStr(writableSistemas(), body.Sistema) {
		writeError(w, 400, "sistema desconocido: "+body.Sistema, nil)
		return
	}

	root, err := getWorkspaceRoot()
	if err != nil {
		writeError(w, 500, "workspace root no resuelto: "+err.Error(), nil)
		return
	}

	scripts := []string{
		filepath.Join(root, "scripts", "generate_code_to_cap_index.py"),
		filepath.Join(root, "scripts", "validate_code_cap_bidirectional.py"),
		// computed_status (verified-live/partial/stub por cap) NO corre en pre-commit
		// — sin esto el tab Drift muestra una foto vieja sin que se note.
		filepath.Join(root, "scripts", "compute_capability_status.py"),
	}
	present := scripts[:0]
	for _, sc := range scripts {
		if _, statErr := os.Stat(sc); statErr == nil {
			present = append(present, sc)
		}
	}
	if len(present) == 0 {
		// Hint honesto y ACCIONABLE: si el workspace trae otro doctor de caps (ArnesIA usa
		// `cap_doctor.py`, no el trío de vitalia), se nombra ése en vez de un genérico que
		// deja al operador sin próximo paso. Lo que NO se hace nunca es reimplementar el
		// resolver en Go: un mirror driftearía del SSoT.
		hint := "Este workspace no trae los scripts de índices cap↔código — se regeneran solo vía pre-commit (si el kit lo instala)."
		if doctor := filepath.Join(root, "scripts", "cap_doctor.py"); fileExists(doctor) {
			hint = "Este workspace no usa el trío de índices cap↔código; su doctor de capabilities es otro. " +
				"Corré: python scripts/cap_doctor.py --index"
		}
		writeJSON(w, 200, map[string]any{"ok": false, "hint": hint})
		return
	}
	scripts = present

	// Serializar regens (los scripts escriben los mismos JSONs).
	regenMu.Lock()
	defer regenMu.Unlock()

	py := pythonBin(root)
	var outputs []string
	start := time.Now()
	for _, sc := range scripts {
		cmd := exec.Command(py, sc, "--sistema", body.Sistema)
		cmd.Dir = root
		out, runErr := cmd.CombinedOutput()
		tail := string(out)
		if len(tail) > 1200 {
			tail = "…" + tail[len(tail)-1200:]
		}
		outputs = append(outputs, filepath.Base(sc)+":\n"+tail)
		if runErr != nil {
			// El validador sale ≠0 si hay drift — eso NO es error del regen:
			// el reporte se escribió igual y el cockpit lo va a mostrar.
			if strings.Contains(filepath.Base(sc), "validate_code_cap_bidirectional") {
				continue
			}
			writeError(w, 500, "regen falló en "+filepath.Base(sc), map[string]any{
				"output": tail,
			})
			return
		}
	}

	writeJSON(w, 200, map[string]any{
		"ok":          true,
		"sistema":       body.Sistema,
		"duration_ms": time.Since(start).Milliseconds(),
		"output_tail": strings.Join(outputs, "\n---\n"),
	})
}
