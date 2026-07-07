// Package artifact implements ports.ArtifactReader: it reads ONLY the `status:` of the
// YAML frontmatter of a box's output artifact (document-as-cache). That status — never
// the chat text — is the machine signal the T3 conductor advances on (boundary
// orquestacion-determinista-entre-cajas: continue/stop/block comes from `result` +
// `status`, jamás de scrapear la conversación).
//
// Confinement: the artifact path resolves UNDER the arnés working directory passed per
// call; a ref that escapes it (../, absolute) is rejected — the reader never looks
// outside the tree the run is confined to.
package artifact

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Reader is the concrete frontmatter reader. Stateless: the confinement dir comes per
// call from the conductor (it is the run's cwd, resolved by the WorkdirResolver).
type Reader struct{}

// NewReader returns the frontmatter status reader.
func NewReader() Reader { return Reader{} }

// Status reads the `status:` of the artifact's YAML frontmatter. Contract:
//   - ref escapes dir (../, absolute, empty dir) → error (confinement violated);
//   - artifact does not exist yet → ("", false, nil) — the box has not produced it;
//   - artifact without frontmatter, or frontmatter without `status:` → ("", true, nil)
//     — an honest "no status declared", not an error;
//   - frontmatter opened but never closed → error (a broken document must surface,
//     never read as "working forever").
func (Reader) Status(_ context.Context, dir, ref string) (string, bool, error) {
	path, err := confinedPath(dir, ref)
	if err != nil {
		return "", false, err
	}
	b, err := os.ReadFile(path) //nolint:gosec // G304: path is verified by confinedPath to stay under the arnés dir.
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("artifact: leer %s: %w", ref, err)
	}
	status, err := frontmatterStatus(string(b))
	if err != nil {
		return "", true, fmt.Errorf("artifact: %s: %w", ref, err)
	}
	return status, true, nil
}

// confinedPath joins ref under dir and verifies the result cannot escape it.
func confinedPath(dir, ref string) (string, error) {
	if dir == "" {
		return "", errors.New("artifact: sin directorio de arnés — la lectura debe estar confinada")
	}
	if ref == "" {
		return "", errors.New("artifact: referencia de artefacto vacía")
	}
	if filepath.IsAbs(ref) {
		return "", fmt.Errorf("artifact: la referencia %q es absoluta — debe ser relativa al arnés", ref)
	}
	joined := filepath.Clean(filepath.Join(dir, ref))
	rel, err := filepath.Rel(dir, joined)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("artifact: la referencia %q escapa del directorio del arnés", ref)
	}
	return joined, nil
}

// frontmatterStatus extracts the top-level `status:` value of a YAML frontmatter block.
// Deliberately a line scan, not a YAML decoder: the contract is ONE scalar key — pulling
// a YAML dependency here would invite reading more of the document than the boundary
// allows (the reader knows `status:`, nothing else).
func frontmatterStatus(doc string) (string, error) {
	doc = strings.TrimPrefix(doc, "\ufeff") // BOM defensivo.
	if !strings.HasPrefix(doc, "---\n") && !strings.HasPrefix(doc, "---\r\n") {
		return "", nil // no frontmatter — no status declared.
	}
	rest := doc[strings.IndexByte(doc, '\n')+1:]
	end := frontmatterEnd(rest)
	if end < 0 {
		return "", errors.New("frontmatter sin cerrar (falta el segundo ---)")
	}
	for _, line := range strings.Split(rest[:end], "\n") {
		if !strings.HasPrefix(line, "status:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(line, "status:"))
		val = strings.Trim(val, `"'`)
		return val, nil
	}
	return "", nil // frontmatter present, no status: declared.
}

// frontmatterEnd returns the offset in rest where the closing `---` line starts, or -1.
func frontmatterEnd(rest string) int {
	offset := 0
	for _, line := range strings.Split(rest, "\n") {
		if strings.TrimRight(line, "\r") == "---" {
			return offset
		}
		offset += len(line) + 1
	}
	return -1
}
