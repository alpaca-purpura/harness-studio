package marketplace

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/alpacapurpura/arnesia/internal/adapters/filelock"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
)

// esquemaActual es la versión de forma que este binario sabe escribir/leer — mismo chequeo
// y mismos motivos que su store hermano (portafolio/store.go#esquemaActual, D1).
const esquemaActual = 1

// store.go persiste el lado DECLARADO del registro en `~/.arnesia/marketplaces.json`.
//
// Por qué acá y no en el índice SQLite (design.md §4.1): el `.db` es DESECHABLE por doctrina
// (`indice-desechable-jsonl-es-verdad` v1.3 hace `wipeFile` en mismatch de esquema y re-indexa
// desde `ports.ArnesRegistry`), y la clase `propio`/`referencia` es una DECLARACIÓN del operador
// que no vive en ningún otro lado: ponerla ahí sería garantizar su pérdida silenciosa en el
// próximo bump. Calco exacto de `internal/adapters/portafolio/store.go`, el dato hermano.

// envelopeMkt es la forma versionada del archivo. Cada fila se decodifica INDEPENDIENTE
// (`[]json.RawMessage`): una fila corrupta no le pega al resto ni al boot del daemon (BR-11).
type envelopeMkt struct {
	Version      int               `json:"version"`
	Marketplaces []json.RawMessage `json:"marketplaces"`
}

// filaPersistida es lo que el REGISTRO posee, y solo eso (design.md §4.1). No se serializa
// `domain.MarketplaceConocido` entero a propósito: sus campos `eslabones` y `lectura` son
// DERIVADOS (los eslabones salen del collect-all y la lectura del caché), y guardarlos en blanco
// dejaba un `"eslabones": null` + `"lectura": {"tipo": ""}` de ruido en un archivo que el operador
// tiene que poder abrir y entender a mano.
type filaPersistida struct {
	Nombre     string                  `json:"nombre"`
	Repo       string                  `json:"repo,omitempty"`
	Clase      domain.ClaseMarketplace `json:"clase"`
	Registrado string                  `json:"registrado,omitempty"`
}

// Store es el registro persistente de los marketplaces DECLARADOS por el operador.
type Store struct {
	path string

	mu        sync.Mutex
	filas     map[string]domain.MarketplaceConocido // key = Nombre (la clave de merge, AG-D9)
	corruptas []domain.EntradaCorrupta

	// bloqueo explica por qué este store NO se puede escribir ("" = se puede) — mismo
	// mecanismo que portafolio.Store.bloqueo (D1).
	bloqueo string
}

var _ ports.MarketplaceStore = (*Store)(nil)

// NewStore abre (o crea) el store en path (default `~/.arnesia/marketplaces.json`). NUNCA falla
// por contenido corrupto — eso degrada a `corruptas`; solo un error de I/O real (permisos, etc.,
// DISTINTO de «no existe») aborta la apertura.
func NewStore(path string) (*Store, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("marketplace store: resolve home: %w", err)
		}
		path = filepath.Join(home, ".arnesia", "marketplaces.json")
	}
	s := &Store{path: path, filas: map[string]domain.MarketplaceConocido{}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

// load lee el archivo entry-wise. Un envelope top-level ilegible se pone en CUARENTENA (D1,
// calco de portafolio/store.go#load) — los bytes originales quedan intactos en
// `<ruta>.corrupto-<sello>`, jamás impide abrir el store (E-44).
func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("marketplace store: leer %s: %w", s.path, err)
	}
	var env envelopeMkt
	if uerr := json.Unmarshal(b, &env); uerr != nil {
		destino, qerr := encuarentenar(s.path)
		if qerr != nil {
			s.corruptas = append(s.corruptas, domain.EntradaCorrupta{
				Raw: b, Motivo: fmt.Sprintf("envelope ilegible: %v (cuarentena falló: %v)", uerr, qerr),
			})
			return nil
		}
		slog.Warn("marketplace store: envelope ilegible, puesto en cuarentena",
			"archivo", s.path, "cuarentena", destino, "motivo", uerr)
		return nil
	}
	if env.Version > esquemaActual {
		s.bloqueo = fmt.Sprintf("%s lo escribió un binario más nuevo (esquema %d, este entiende %d)",
			s.path, env.Version, esquemaActual)
		slog.Warn("marketplace store: esquema futuro, store en solo-lectura", "archivo", s.path, "bloqueo", s.bloqueo)
		return nil
	}
	for _, raw := range env.Marketplaces {
		var m domain.MarketplaceConocido
		if derr := json.Unmarshal(raw, &m); derr != nil {
			s.corruptas = append(s.corruptas, domain.EntradaCorrupta{Raw: raw, Motivo: derr.Error()})
			continue
		}
		if m.Nombre == "" {
			// NO se le inventa un nombre derivándolo del repo (E-46): sin nombre no hay clave.
			s.corruptas = append(s.corruptas, domain.EntradaCorrupta{
				Raw: raw, Motivo: "fila sin nombre: no hay clave de merge",
			})
			continue
		}
		s.filas[m.Nombre] = m
	}
	return nil
}

// reloadLocked descarta el estado en memoria y vuelve a leer `s.path` desde cero — mismo
// mecanismo y mismos motivos que `portafolio.Store.reloadLocked` (Fase 1, D2/D3). Caller
// sostiene s.mu.
func (s *Store) reloadLocked() error {
	s.filas = map[string]domain.MarketplaceConocido{}
	s.corruptas = nil
	s.bloqueo = ""
	return s.load()
}

// Listar devuelve las filas sanas (orden estable por nombre) + las corruptas visibles aparte
// (BR-11) — nunca mezcladas, nunca ocultas. Cada fila sale con su eslabón `declarado-por-operador`.
func (s *Store) Listar() ([]domain.MarketplaceConocido, []domain.EntradaCorrupta) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]domain.MarketplaceConocido, 0, len(s.filas))
	for _, m := range s.filas {
		m.Eslabones = []domain.EslabonMarketplace{domain.EslabonDeclarado}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nombre < out[j].Nombre })
	corr := make([]domain.EntradaCorrupta, len(s.corruptas))
	copy(corr, s.corruptas)
	return out, corr
}

// Upsert inserta o reemplaza la fila declarada de m.Nombre. Es el mecanismo; la regla BR-7 («no
// duplica ni pisa») la enforça el usecase, que consulta antes de llamar — separar mecanismo de
// política es lo que permite que `Olvidar`+`Registrar` funcione sin un flag de «forzar».
//
// Corre bajo `filelock.Guard` (Fase 1, D2/D4): recarga desde disco antes de mutar, mismo
// mecanismo que `portafolio.Store.Upsert` — uniforme entre los dos stores hermanos aunque hoy
// solo el daemon escriba `marketplaces.json` (D3: sin CLI standalone confirmado todavía).
func (s *Store) Upsert(m domain.MarketplaceConocido) error {
	if m.Nombre == "" {
		return errors.New("marketplace store: upsert sin nombre (no hay clave de merge)")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return filelock.Guard(s.path, func() error {
		if err := s.reloadLocked(); err != nil {
			return err
		}
		s.filas[m.Nombre] = m
		return s.saveLocked()
	})
}

// Olvidar quita la fila DECLARADA (bool=true si existía). Si Claude Code igual lo conoce, la
// fila sigue apareciendo en el plano con eslabón `cc-known-marketplaces` — no podemos hacer que
// CC deje de conocerlo, y lo decimos (E-71).
func (s *Store) Olvidar(nombre string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var borrada bool
	err := filelock.Guard(s.path, func() error {
		if rerr := s.reloadLocked(); rerr != nil {
			return rerr
		}
		if _, ok := s.filas[nombre]; !ok {
			return nil
		}
		delete(s.filas, nombre)
		borrada = true
		return s.saveLocked()
	})
	if err != nil {
		return false, err
	}
	return borrada, nil
}

// saveLocked escribe el envelope atómico (temp+rename). Las corruptas se RE-SERIALIZAN crudas
// junto a las sanas siempre que su Raw sea JSON válido — misma decisión y misma limitación
// honesta que `portafolio.Store.saveLocked` (bytes que no son JSON no pueden re-insertarse como
// elemento de un array JSON). Caller sostiene s.mu.
func (s *Store) saveLocked() error {
	if s.bloqueo != "" {
		return fmt.Errorf("marketplace store: %s", s.bloqueo)
	}
	nombres := make([]string, 0, len(s.filas))
	for n := range s.filas {
		nombres = append(nombres, n)
	}
	sort.Strings(nombres)

	filas := make([]json.RawMessage, 0, len(nombres)+len(s.corruptas))
	for _, n := range nombres {
		m := s.filas[n]
		fila := filaPersistida{Nombre: m.Nombre, Repo: m.Repo, Clase: m.Clase, Registrado: m.Registrado}
		b, err := json.Marshal(fila)
		if err != nil {
			return fmt.Errorf("marketplace store: encode %s: %w", n, err)
		}
		filas = append(filas, b)
	}
	for _, c := range s.corruptas {
		if !json.Valid(c.Raw) {
			continue
		}
		filas = append(filas, c.Raw)
	}

	b, err := json.MarshalIndent(envelopeMkt{Version: esquemaActual, Marketplaces: filas}, "", "  ")
	if err != nil {
		return fmt.Errorf("marketplace store: encode envelope: %w", err)
	}
	return escribirAtomico(s.path, ".marketplaces-*.json", b)
}

// escribirAtomico escribe b en path con temp+rename dentro del MISMO dir (el rename es atómico
// a nivel POSIX solo dentro del mismo filesystem). Compartido por Store y Cache.
func escribirAtomico(path, patron string, b []byte) error {
	dir := filepath.Dir(path)
	if merr := os.MkdirAll(dir, 0o750); merr != nil {
		return fmt.Errorf("marketplace store: mkdir %s: %w", dir, merr)
	}
	tmp, err := os.CreateTemp(dir, patron)
	if err != nil {
		return fmt.Errorf("marketplace store: temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op tras un rename OK.
	if _, werr := tmp.Write(b); werr != nil {
		_ = tmp.Close()
		return fmt.Errorf("marketplace store: write temp: %w", werr)
	}
	if cerr := tmp.Close(); cerr != nil {
		return fmt.Errorf("marketplace store: close temp: %w", cerr)
	}
	if rerr := os.Rename(tmpName, path); rerr != nil {
		return fmt.Errorf("marketplace store: rename %s: %w", path, rerr)
	}
	return nil
}

// encuarentenar mueve un archivo ilegible a `<ruta>.corrupto-<sello>` — calco exacto de
// portafolio/store.go#encuarentenar (D1), el mecanismo hermano.
func encuarentenar(ruta string) (string, error) {
	sello := time.Now().UTC().Format("0601021504")
	destino := fmt.Sprintf("%s.corrupto-%s", ruta, sello)
	for i := 2; ; i++ {
		if _, err := os.Stat(destino); os.IsNotExist(err) {
			break
		}
		destino = fmt.Sprintf("%s.corrupto-%s-%d", ruta, sello, i)
	}
	if err := os.Rename(ruta, destino); err != nil {
		return "", fmt.Errorf("marketplace store: cuarentena de %s: %w", ruta, err)
	}
	return destino, nil
}
