// watch.go — SSE /api/watch con fsnotify (paridad con chokidar-watcher.ts).
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var watchIgnored = []string{".git", "node_modules", ".next", "__pycache__", ".venv", "dist", "build"}

func watchPathIgnored(p string) bool {
	if strings.HasSuffix(p, ".tmp") {
		return true
	}
	for _, seg := range strings.Split(p, string(filepath.Separator)) {
		for _, ig := range watchIgnored {
			if seg == ig {
				return true
			}
		}
	}
	return false
}

func docTypeFor(p string) string {
	base := filepath.Base(p)
	// ToSlash: en Windows fsnotify entrega `docs\product\capabilities\...` y los Contains
	// con separador `/` nunca matchean — TODO cambio caía en "other" y la UI no refrescaba
	// la tarjeta correcta. Normalizar acá y no en el caller: este es el único punto que
	// compara el path como texto.
	slash := filepath.ToSlash(p)
	switch {
	case base == "checkpoint.md":
		return "checkpoint"
	case base == "operator-input.md" || base == "chris-input.md":
		return "operator-input"
	case strings.Contains(slash, "/capabilities/") && strings.HasSuffix(slash, ".yaml"):
		return "capability"
	case strings.Contains(slash, "/releases/") && strings.HasSuffix(slash, ".yaml"):
		return "release"
	case strings.Contains(slash, "/learnings/") && strings.HasSuffix(slash, ".md"):
		return "learning"
	}
	return "other"
}

func sistemaForPath(p string) string {
	if isMultiMode() {
		for _, proj := range multiProjects {
			rel, err := filepath.Rel(proj.Path, p)
			if err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
			first := strings.SplitN(rel, string(filepath.Separator), 2)[0]
			if containsStr(sistemasIn(proj.Path), first) {
				return proj.Name + "/" + first
			}
			if first == "docs" {
				return proj.Name + "/" + platformSlug
			}
		}
		return ""
	}
	root, err := getWorkspaceRoot()
	if err != nil {
		return ""
	}
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return ""
	}
	first := strings.SplitN(rel, string(filepath.Separator), 2)[0]
	if containsStr(getSistemas(), first) {
		return first
	}
	if first == "docs" {
		return platformSlug
	}
	return ""
}

// addWatchRecursive registra dir + todos los subdirs (fsnotify no es recursivo).
func addWatchRecursive(watcher *fsnotify.Watcher, dir string) {
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && !watchPathIgnored(p) {
			_ = watcher.Add(p)
		}
		return nil
	})
}

func handleWatch(w http.ResponseWriter, r *http.Request) {
	sistemaParam := r.URL.Query().Get("sistema")
	validSistemas := getSelectableSistemas()

	var sistemasToWatch []string
	if sistemaParam != "" {
		if !containsStr(validSistemas, sistemaParam) {
			writeError(w, 400, "sistema desconocido: "+sistemaParam, map[string]any{"valid_sistemas": validSistemas})
			return
		}
		sistemasToWatch = []string{sistemaParam}
	} else {
		sistemasToWatch = validSistemas
	}


	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		writeError(w, 500, "error creando watcher", map[string]any{"detail": err.Error()})
		return
	}
	defer watcher.Close()

	for _, sistema := range sistemasToWatch {
		base := sistemaDocsRoot(sistema)
		for _, sub := range []string{"product", "archive", "learnings"} {
			dir := filepath.Join(base, sub)
			if dirExists(dir) {
				addWatchRecursive(watcher, dir)
			}
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, 500, "streaming no soportado", nil)
		return
	}

	var mu sync.Mutex
	send := func(payload any) bool {
		data, err := json.Marshal(payload)
		if err != nil {
			return true
		}
		mu.Lock()
		defer mu.Unlock()
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	send(map[string]any{"type": "connected", "sistemas": sistemasToWatch})

	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()

	// Debounce por path (200ms, como chokidar)
	pending := map[string]fsnotify.Event{}
	var pendingMu sync.Mutex
	debounce := time.NewTicker(200 * time.Millisecond)
	defer debounce.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeat.C:
			if !send(map[string]any{"type": "heartbeat", "ts": time.Now().UnixMilli()}) {
				return
			}
		case ev, okCh := <-watcher.Events:
			if !okCh {
				return
			}
			if watchPathIgnored(ev.Name) {
				continue
			}
			// dirs nuevos → registrar para watch recursivo
			if ev.Op.Has(fsnotify.Create) && dirExists(ev.Name) {
				addWatchRecursive(watcher, ev.Name)
				continue
			}
			pendingMu.Lock()
			pending[ev.Name] = ev
			pendingMu.Unlock()
		case <-debounce.C:
			pendingMu.Lock()
			batch := pending
			pending = map[string]fsnotify.Event{}
			pendingMu.Unlock()
			for p, ev := range batch {
				action := "change"
				if ev.Op.Has(fsnotify.Create) {
					action = "add"
				} else if ev.Op.Has(fsnotify.Remove) || ev.Op.Has(fsnotify.Rename) {
					action = "unlink"
				}
				payload := map[string]any{"path": p, "action": action, "docType": docTypeFor(p)}
				if b := sistemaForPath(p); b != "" {
					payload["sistema"] = b
				}
				if !send(payload) {
					return
				}
			}
		case _, okCh := <-watcher.Errors:
			if !okCh {
				return
			}
		}
	}
}
