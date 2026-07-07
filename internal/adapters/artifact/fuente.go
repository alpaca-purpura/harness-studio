package artifact

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alpacapurpura/arnesia/internal/ports"
)

// FuenteReader implements ports.FuenteReader: it serves the RAW bytes of a node's
// fuente_path for the inspector's Contenido tab (RF-93), with the same confinement
// discipline as Reader (document-as-cache status): nothing outside the arnés dir.
// Unlike Reader.Status, an absolute ref is legal here — the loader stamps absolute
// fuente_path when the arnés was registered by absolute path — as long as it still
// falls INSIDE dir.
type FuenteReader struct{}

// NewFuenteReader returns the confined raw-source reader.
func NewFuenteReader() FuenteReader { return FuenteReader{} }

// LeerConfinado reads ref confined to dir. See ports.FuenteReader for the contract.
func (FuenteReader) LeerConfinado(dir, ref string) ([]byte, error) {
	if dir == "" {
		return nil, errors.New("fuente: sin directorio de arnés — la lectura debe estar confinada")
	}
	if ref == "" {
		return nil, errors.New("fuente: referencia vacía")
	}
	path := ref
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	path = filepath.Clean(path)
	rel, err := filepath.Rel(dir, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("%w: %q fuera de %q", ports.ErrFueraDelArnes, ref, dir)
	}
	b, rerr := os.ReadFile(path)
	if rerr != nil {
		return nil, fmt.Errorf("fuente: leer %s: %w", ref, rerr)
	}
	return b, nil
}
