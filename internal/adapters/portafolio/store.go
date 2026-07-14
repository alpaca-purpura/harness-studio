package portafolio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// envelope es la forma versionada de `~/.arnesia/portafolio.json` (S0-D5). Cada entrada
// se decodifica INDEPENDIENTE (`[]json.RawMessage`, no un `[]domain.EntradaPortafolio`
// de un tirón): una fila corrupta no le pega al resto ni al boot del daemon (BR-11/C-N-4).
type envelope struct {
	Version  int               `json:"version"`
	Entradas []json.RawMessage `json:"entradas"`
}

// Store es el registro persistente del Portafolio — separado de `arneses.json` (A1, S0-D5):
// ese es el cwd de sesión, este es la identidad-a-largo-plazo del Portafolio.
type Store struct {
	path string

	mu        sync.Mutex
	entradas  map[string]domain.EntradaPortafolio // key = Identidad.Clave()
	corruptas []domain.EntradaCorrupta
}

var _ ports.PortafolioStore = (*Store)(nil)

// NewStore abre (o crea) el store en path (default `~/.arnesia/portafolio.json`). NUNCA
// falla por contenido corrupto del archivo — eso degrada a `corruptas` (C-N-4); solo un
// error de I/O real (permisos, etc., DISTINTO de "no existe") aborta la apertura.
func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("portafolio store: resolve home: %w", err)
		}
		path = filepath.Join(home, ".arnesia", "portafolio.json")
	}
	s := &Store{path: path, entradas: map[string]domain.EntradaPortafolio{}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load lee el archivo entry-wise. Un envelope top-level ilegible (archivo editado a mano
// hasta romperlo) degrada a UNA corrupta con el blob entero — jamás impide abrir el store.
func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("portafolio store: leer %s: %w", s.path, err)
	}
	var env envelope
	if uerr := json.Unmarshal(b, &env); uerr != nil {
		s.corruptas = append(s.corruptas, domain.EntradaCorrupta{Raw: b, Motivo: fmt.Sprintf("envelope ilegible: %v", uerr)})
		return nil
	}
	for _, raw := range env.Entradas {
		var e domain.EntradaPortafolio
		if derr := json.Unmarshal(raw, &e); derr != nil {
			s.corruptas = append(s.corruptas, domain.EntradaCorrupta{Raw: raw, Motivo: derr.Error()})
			continue
		}
		s.entradas[e.Identidad.Clave()] = e
	}
	return nil
}

// Listar devuelve las entradas sanas (orden estable por clave) + las corruptas visibles
// aparte (BR-11) — nunca mezcladas, nunca ocultas.
func (s *Store) Listar() ([]domain.EntradaPortafolio, []domain.EntradaCorrupta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.EntradaPortafolio, 0, len(s.entradas))
	for _, e := range s.entradas {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Identidad.Clave() < out[j].Identidad.Clave() })
	corr := make([]domain.EntradaCorrupta, len(s.corruptas))
	copy(corr, s.corruptas)
	return out, corr
}

// Upsert inserta o fusiona e por su Identidad.Clave(): instalaciones dedup por
// InstallPath (fresco pisa viejo); dos canónicos DISTINTOS para la misma identidad es un
// error explícito (C-N-5), nunca una elección silenciosa. Una identidad provisional y una
// resuelta JAMÁS colisionan de clave — Clave() ya las diferencia estructuralmente
// (prefijo `sin-home~` vs el home canonicalizado) — así que Upsert nunca las fusiona
// (C-ID-2) sin necesitar lógica extra.
func (s *Store) Upsert(e domain.EntradaPortafolio) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clave := e.Identidad.Clave()
	existing, ok := s.entradas[clave]
	if !ok {
		s.entradas[clave] = e
		return s.saveLocked()
	}

	merged := e
	merged.Instalaciones = mergeInstalaciones(existing.Instalaciones, e.Instalaciones)
	// S1-D3 (cierra GAP-3): Empresas/Registries son facetas N:M — se UNEN, jamás se
	// reemplazan; un re-agregar con menos datos (p.ej. un candidato sin registry
	// resuelto) NO borra lo que ya estaba persistido.
	merged.Empresas = unionDedup(existing.Empresas, e.Empresas)
	merged.Registries = unionDedup(existing.Registries, e.Registries)
	switch {
	case e.Canonico != nil && existing.Canonico != nil && e.Canonico.Path != existing.Canonico.Path:
		return fmt.Errorf("portafolio store: dos canónicos distintos para %q: %q vs %q (C-N-5)",
			clave, existing.Canonico.Path, e.Canonico.Path)
	case e.Canonico == nil:
		merged.Canonico = existing.Canonico
	}
	if merged.Agregado == "" {
		merged.Agregado = existing.Agregado
	}
	s.entradas[clave] = merged
	return s.saveLocked()
}

// mergeInstalaciones dedupea por InstallPath (C-P-8/C-N-3): la entrada nueva pisa la
// vieja del mismo path; el resto se conserva.
func mergeInstalaciones(oldList, newList []domain.Instalacion) []domain.Instalacion {
	byPath := make(map[string]domain.Instalacion, len(oldList)+len(newList))
	for _, i := range oldList {
		byPath[i.InstallPath] = i
	}
	for _, i := range newList {
		byPath[i.InstallPath] = i
	}
	out := make([]domain.Instalacion, 0, len(byPath))
	for _, i := range byPath {
		out = append(out, i)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].InstallPath < out[b].InstallPath })
	return out
}

// unionDedup une existing∪nueva preservando el orden de aparición (lo existente primero,
// luego lo nuevo que no estaba) sin duplicados ni strings vacíos (S1-D3): una faceta vacía
// en `nueva` NUNCA borra lo existente — unionDedup(existing, nil) == existing. Devuelve
// nil (no un slice vacío) cuando ambas entradas están vacías, para que `omitempty` siga
// funcionando en el wire.
func unionDedup(existing, nueva []string) []string {
	if len(existing) == 0 && len(nueva) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(existing)+len(nueva))
	out := make([]string, 0, len(existing)+len(nueva))
	for _, v := range existing {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	for _, v := range nueva {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

// Desvincular quita clave del registro (bool=true si existía). NO toca ningún archivo de
// proyecto ni el clon canónico (C-UNL-3) — el store es lo único que se edita.
func (s *Store) Desvincular(clave string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.entradas[clave]; !ok {
		return false, nil
	}
	delete(s.entradas, clave)
	if err := s.saveLocked(); err != nil {
		return false, err
	}
	return true, nil
}

// Checkouts devuelve los paths de canónicos conocidos (RN-IDENT-4).
func (s *Store) Checkouts() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, e := range s.entradas {
		if e.Canonico != nil && e.Canonico.Path != "" {
			out = append(out, e.Canonico.Path)
		}
	}
	sort.Strings(out)
	return out
}

// saveLocked escribe el envelope atómico (temp+rename, mismo patrón que
// arnes_registry.go#saveLocked). Las corruptas se RE-SERIALIZAN crudas junto a las sanas
// (decisión explícita: se conservan, jamás se descartan en silencio — sobreviven a un
// ciclo load→save→load) SIEMPRE QUE su Raw sea JSON válido a nivel sintáctico (una fila
// individual de forma incorrecta, p.ej. `identidad` con el tipo equivocado). Un archivo
// TOTALMENTE ilegible (ni siquiera tokeniza como array — editado a mano hasta romper la
// sintaxis) no puede re-insertarse como elemento de un array JSON nuevo por definición;
// esa corrupta se reporta en el Listar() de esta apertura pero no sobrevive a un write
// posterior — no hay forma honesta de "conservar" bytes que no son JSON dentro de un
// documento que sí debe serlo. Caller sostiene s.mu.
func (s *Store) saveLocked() error {
	claves := make([]string, 0, len(s.entradas))
	for k := range s.entradas {
		claves = append(claves, k)
	}
	sort.Strings(claves)

	entradas := make([]json.RawMessage, 0, len(claves)+len(s.corruptas))
	for _, k := range claves {
		b, err := json.Marshal(s.entradas[k])
		if err != nil {
			return fmt.Errorf("portafolio store: encode %s: %w", k, err)
		}
		entradas = append(entradas, b)
	}
	for _, c := range s.corruptas {
		if !json.Valid(c.Raw) {
			continue // no hay forma honesta de re-insertar bytes no-JSON en el envelope.
		}
		entradas = append(entradas, c.Raw)
	}

	env := envelope{Version: 1, Entradas: entradas}
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("portafolio store: encode envelope: %w", err)
	}

	dir := filepath.Dir(s.path)
	if merr := os.MkdirAll(dir, 0o750); merr != nil {
		return fmt.Errorf("portafolio store: mkdir %s: %w", dir, merr)
	}
	tmp, err := os.CreateTemp(dir, ".portafolio-*.json")
	if err != nil {
		return fmt.Errorf("portafolio store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op tras un rename OK.
	if _, werr := tmp.Write(b); werr != nil {
		_ = tmp.Close()
		return fmt.Errorf("portafolio store: write temp: %w", werr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("portafolio store: close temp: %w", cerr)
	}
	if rerr := os.Rename(tmpName, s.path); rerr != nil {
		return fmt.Errorf("portafolio store: rename %s: %w", s.path, rerr)
	}
	return nil
}
