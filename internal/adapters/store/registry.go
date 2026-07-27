// Package store implements ports.SessionStore as a single JSON file under the user's
// home (~/.arnesia/sessions.json). It is deliberately trivial: the session *registry*
// is small and low-churn, and the disposable SQLite index (fase 5) is a separate
// concern. Writes are atomic (temp file + rename) so a crash mid-save never corrupts
// the registry.
package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/domain"
)

// Registry is a file-backed session store. Safe for concurrent use.
type Registry struct {
	path string
	// sello de build del binario que escribe, estampado en el sobre. Vacío cuando el
	// Registry se construyó con NewRegistry a secas: el sobre lo omite en vez de mentir.
	sello string

	// bloqueo explica por qué este registro NO se puede escribir; "" = se puede.
	// Lo pone AbrirRegistro cuando el archivo en disco lo escribió un binario más nuevo:
	// degradar a «registro vacío» y después persistir destruiría el archivo nuevo con el
	// binario viejo, que es el único modo de fallo que no se puede deshacer.
	bloqueo string

	// corrupto marca que el archivo se puso en cuarentena. El registro arranca vacío, y
	// eso NO es lo mismo que un primer arranque: la semilla ilustrativa no puede tapar
	// una corrupción haciéndola parecer una instalación nueva.
	corrupto bool

	mu sync.Mutex
}

// SoloLectura reporta si el registro está bloqueado y por qué. Toda mutación consulta
// esto ANTES de tocar memoria.
func (r *Registry) SoloLectura() (bool, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.bloqueo != "", r.bloqueo
}

// Sembrable reporta si un registro vacío puede rellenarse con las sesiones de ejemplo.
// Falso cuando el vacío NO significa «primer arranque»: un archivo en cuarentena o un
// esquema futuro también cargan vacío, y sembrarlos ahí taparía el problema con datos
// inventados justo cuando el operador necesita ver que algo pasó.
func (r *Registry) Sembrable() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return !r.corrupto && r.bloqueo == ""
}

var _ = (interface {
	Load(context.Context) ([]domain.Session, error)
	Save(context.Context, []domain.Session) error
})((*Registry)(nil))

// NewRegistry returns a Registry writing to path. If path is empty it defaults to
// ~/.arnesia/sessions.json. The parent directory is created on first Save.
func NewRegistry(path string) (*Registry, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("store: resolve home: %w", err)
		}
		path = filepath.Join(home, ".arnesia", "sessions.json")
	}
	return &Registry{path: path}, nil
}

// Load reads the persisted sessions. A missing file is not an error — it yields an
// empty registry (first run). Lee las dos formas: el sobre versionado que este binario
// escribe y el array desnudo que se escribía antes, porque un registro sin migrar
// (el de sesiones archivadas, p. ej.) sigue siendo legible.
func (r *Registry) Load(_ context.Context) ([]domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return []domain.Session{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("store: read %s: %w", r.path, err)
	}
	version, err := detectarVersion(b)
	if err != nil {
		return nil, fmt.Errorf("store: %s: %w", r.path, err)
	}
	if version > EsquemaActual {
		return nil, fmt.Errorf(
			"store: %s lo escribió un binario más nuevo (esquema %d, este entiende %d)",
			r.path, version, EsquemaActual)
	}
	payload, err := payloadDe(b, version)
	if err != nil {
		return nil, fmt.Errorf("store: %s: %w", r.path, err)
	}
	// Un archivo de una versión anterior se MIGRA en memoria antes de leerlo. Sin esto,
	// decodificar la forma vieja con el tipo de hoy tira en silencio todo lo que cambió de
	// lugar — es el «unmarshal tolerante que se lleva lo que entre» que el boundary
	// prohíbe, y le costó 90 turnos al primer registro real que lo cruzó. Leer no escribe:
	// acá no hay respaldo porque no hay nada que respaldar.
	if payload, err = migrarEnMemoria(payload, version, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("store: %s: %w", r.path, err)
	}
	var sessions []domain.Session
	if err := json.Unmarshal(payload, &sessions); err != nil {
		return nil, fmt.Errorf("store: decode %s: %w", r.path, err)
	}
	return sessions, nil
}

// respaldarSiEsViejoLocked copia el archivo en disco si es de un esquema anterior al que
// se va a escribir. Caller holds r.mu.
func (r *Registry) respaldarSiEsViejoLocked() error {
	b, err := os.ReadFile(r.path)
	if os.IsNotExist(err) {
		return nil // no hay nada que respaldar.
	}
	if err != nil {
		return fmt.Errorf("store: read %s: %w", r.path, err)
	}
	// Un archivo ilegible no se respalda acá: la cuarentena de AbrirRegistro ya lo pone a
	// salvo entero, y hacerlo dos veces dejaría dos copias del mismo problema.
	version, derr := detectarVersion(b)
	if derr != nil {
		return nil //nolint:nilerr // ilegible ⇒ lo maneja la cuarentena, no el respaldo.
	}
	if version >= EsquemaActual {
		return nil // ya al día: no hay versión anterior que preservar.
	}
	if _, berr := respaldar(r.path, version, r.sello); berr != nil {
		return berr
	}
	return nil
}

// Save atomically replaces the file with sessions (temp file in the same directory,
// then rename — an atomic swap on the same filesystem).
func (r *Registry) Save(_ context.Context, sessions []domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.bloqueo != "" {
		return fmt.Errorf("%w: %s", ErrSoloLectura, r.bloqueo)
	}
	// Si en disco hay un archivo de una versión anterior, se respalda ANTES de pisarlo.
	// El respaldo va acá y no en el llamador para que no haya un camino de escritura que
	// se lo saltee — el que se lo saltea es el que borra el archivo del operador.
	if err := r.respaldarSiEsViejoLocked(); err != nil {
		return err
	}

	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("store: mkdir %s: %w", dir, err)
	}
	payload, err := json.Marshal(sessions)
	if err != nil {
		return fmt.Errorf("store: encode: %w", err)
	}
	b, err := envolver(payload, r.sello)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".sessions-*.json")
	if err != nil {
		return fmt.Errorf("store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup: a no-op after a successful rename, and on the error paths the
	// write/close error below is the one worth reporting, not the leftover-temp removal.
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(b); err != nil {
		_ = tmp.Close() // the write error is the root cause; Close only releases the fd.
		return fmt.Errorf("store: write temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("store: close temp: %w", err)
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		return fmt.Errorf("store: rename %s: %w", r.path, err)
	}
	return nil
}
