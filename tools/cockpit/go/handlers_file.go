// handlers_file.go — el file-API (/api/file GET+PUT, /api/open) y su guarda de
// seguridad (whitelist + defensa contra path-traversal). Aislado del god-file
// handlers_misc.go porque es código sensible a seguridad: shell-out a editores y
// lectura/escritura de archivos arbitrarios del workspace. El núcleo de la defensa
// (isTraversalSafe, segmentWhitelisted) es puro y testeable → handlers_file_test.go.
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ── núcleo de seguridad (puro, testeable) ───────────────────────────────────

// isTraversalSafe limpia relPath y rechaza absolutos + parent-traversal (`../`).
// Devuelve el path limpio (SIEMPRE en forma slash) y ok=false si es inseguro. Núcleo
// de resolveSafePath.
//
// El path viaja por JSON en forma slash y así se compara aguas abajo
// (segmentWhitelisted parte por "/"), así que la normalización es a slash — nunca a
// separador nativo. Dos razones concretas, ambas verificadas en Windows:
//
//  1. `filepath.Clean` traduce a `\` en Windows: el clean de "docs/x.md" salía
//     `docs\x.md`, segmentWhitelisted no encontraba el segmento "docs" y el file-API
//     rechazaba TODO path. Falla cerrada, pero el visor y el editor quedaban muertos.
//     Además `Contains(clean, "/../")` jamás matcheaba. Por eso `path.Clean`, no
//     `filepath.Clean`.
//  2. `filepath.IsAbs("/etc/passwd")` es FALSE en Windows (allá absoluto = `C:\…` o
//     UNC), así que un absoluto POSIX entraba como relativo. Se chequea la forma slash
//     y el volumen aparte para que el veredicto sea el MISMO en los dos sistemas.
func isTraversalSafe(relPath string) (clean string, ok bool) {
	if relPath == "" {
		return "", false
	}
	slash := filepath.ToSlash(relPath)
	// Absoluto en cualquiera de las dos convenciones: POSIX (`/x`, y en Windows también
	// el `\x` rooted que ToSlash acaba de normalizar) o con volumen (`C:\x`, `C:x`, UNC).
	if filepath.IsAbs(relPath) || strings.HasPrefix(slash, "/") || filepath.VolumeName(relPath) != "" {
		return "", false
	}
	clean = strings.TrimPrefix(path.Clean(slash), "./")
	if strings.HasPrefix(clean, "..") || strings.Contains(clean, "/../") {
		return "", false
	}
	return clean, true
}

// segmentWhitelisted decide si `clean` pasa el whitelist del file-API dado el set de
// sistemas válidos en su root. Puro: la pertenencia de sistema entra como argumento.
// Permitido: `{sistema}/docs/...` · `.claude/...` · `docs/...`.
func segmentWhitelisted(clean string, sistemas []string) bool {
	segments := strings.Split(clean, "/")
	first := segments[0]
	if containsStr(sistemas, first) {
		return len(segments) > 1 && segments[1] == "docs"
	}
	return first == ".claude" || first == "docs"
}

// ── resolución contra el workspace (lee disco vía sistemasIn/multiProjects) ────

// resolveSafeIn valida un rel path contra UN root: whitelist + Rel sin escape.
func resolveSafeIn(root, clean string) string {
	if !segmentWhitelisted(clean, sistemasIn(root)) {
		return ""
	}
	abs := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, abs)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}
	return abs
}

func resolveSafePath(relPath string) string {
	clean, ok := isTraversalSafe(relPath)
	if !ok {
		return ""
	}
	if isMultiMode() {
		// Forma "{proyecto}/{resto}" primero; fallback: probar cada root.
		segments := strings.SplitN(clean, "/", 2)
		if len(segments) == 2 {
			for _, p := range multiProjects {
				if p.Name == segments[0] {
					if abs := resolveSafeIn(p.Path, segments[1]); abs != "" {
						return abs
					}
				}
			}
		}
		for _, p := range multiProjects {
			if abs := resolveSafeIn(p.Path, clean); abs != "" {
				return abs
			}
		}
		return ""
	}
	root, err := getWorkspaceRoot()
	if err != nil {
		return ""
	}
	return resolveSafeIn(root, clean)
}

// ── /api/file GET + PUT (whitelist) ─────────────────────────────────────────

func handleFile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		relPath := r.URL.Query().Get("path")
		if relPath == "" {
			writeError(w, 400, "query param \"path\" requerido", nil)
			return
		}
		abs := resolveSafePath(relPath)
		if abs == "" {
			writeError(w, 403, "path fuera del whitelist o inválido", map[string]any{"path": relPath})
			return
		}
		raw, err := os.ReadFile(abs)
		if err != nil {
			if os.IsNotExist(err) {
				writeError(w, 404, "archivo no existe", map[string]any{"path": relPath})
				return
			}
			writeError(w, 500, "error leyendo archivo", map[string]any{"detail": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"path": relPath, "content": string(raw)})

	case http.MethodPut:
		var body struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, 400, "body JSON inválido", nil)
			return
		}
		if body.Path == "" {
			writeError(w, 400, "body inválido", nil)
			return
		}
		abs := resolveSafePath(body.Path)
		if abs == "" {
			writeError(w, 403, "path fuera del whitelist o inválido", map[string]any{"path": body.Path})
			return
		}
		if err := writeFileAtomic(abs, body.Content); err != nil {
			writeError(w, 500, "error escribiendo archivo", map[string]any{"detail": err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true})

	default:
		writeError(w, 405, "method not allowed", nil)
	}
}

// ── POST /api/open (abrir archivo en editor local) ──────────────────────────

// defaultEditorChain: la cadena de editores a probar, por OS.
//
// La cadena upstream (`xdg-open, code, xed, gnome-text-editor, nano`) es Linux pura: en
// Windows ninguno de los cuatro no-`code` existe, así que "Abrir" fallaba las 5 veces y
// devolvía 500 sin que el operador supiera por qué. En Windows el último recurso es el
// `start` del shell, que respeta la asociación de archivos del usuario — el análogo real
// de `xdg-open`.
func defaultEditorChain() []string {
	if runtime.GOOS == "windows" {
		return []string{"code", "start"}
	}
	return []string{"xdg-open", "code", "xed", "gnome-text-editor", "nano"}
}

// editorCommand arma el comando para un binario de la cadena.
//
// `start` no es un ejecutable sino un builtin de cmd.exe, y su primer argumento entre
// comillas se interpreta como TÍTULO de ventana — por eso va el `""` vacío antes del
// path (sin él, un path con espacios abre una consola titulada en vez del archivo).
func editorCommand(bin, abs string) *exec.Cmd {
	if runtime.GOOS == "windows" && strings.EqualFold(bin, "start") {
		return exec.Command("cmd", "/c", "start", "", abs)
	}
	return exec.Command(bin, abs)
}

func handleOpen(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "body JSON inválido", nil)
		return
	}
	if body.Path == "" {
		writeError(w, 400, "body inválido", nil)
		return
	}

	roots := workspaceRoots()
	if len(roots) == 0 {
		writeError(w, 500, "sin workspace", nil)
		return
	}
	var abs string
	if filepath.IsAbs(body.Path) {
		abs = filepath.Clean(body.Path)
	} else {
		abs = filepath.Join(roots[0], filepath.Clean(body.Path))
	}
	inside := false
	for _, root := range roots {
		rel, err := filepath.Rel(root, abs)
		if err == nil && !strings.HasPrefix(rel, "..") {
			inside = true
			break
		}
	}
	if !inside {
		writeError(w, 403, "path fuera del workspace o inválido", map[string]any{"path": body.Path})
		return
	}
	if !fileExists(abs) {
		writeError(w, 404, "path no existe en el filesystem", map[string]any{"path": body.Path})
		return
	}

	chain := defaultEditorChain()
	if env := os.Getenv("EDITOR_BIN"); env != "" {
		chain = strings.Split(env, ",")
	}

	tried := []map[string]any{}
	for _, bin := range chain {
		bin = strings.TrimSpace(bin)
		cmd := editorCommand(bin, abs)
		if err := cmd.Start(); err != nil {
			tried = append(tried, map[string]any{"bin": bin, "ok": false, "error": err.Error()})
			continue
		}
		go func() { _ = cmd.Wait() }()
		time.Sleep(100 * time.Millisecond)
		writeJSON(w, 200, map[string]any{"ok": true, "editor": bin, "attempts_count": len(tried) + 1})
		return
	}
	writeJSON(w, 500, map[string]any{
		"error": "no se pudo abrir el archivo en ningún editor de la cadena",
		"tried": tried,
		"hint":  "Setea EDITOR_BIN en el .env.local del cockpit con tu editor preferido. Ejemplos: EDITOR_BIN=code · EDITOR_BIN=cursor · EDITOR_BIN=xdg-open · EDITOR_BIN=code,xed,xdg-open (fallback chain).",
		"path":  body.Path,
	})
}
