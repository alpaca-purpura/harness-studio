// Package index implements ports.IndexPort over modernc.org/sqlite (pure Go, no CGO —
// boundary indice-desechable-jsonl-es-verdad.md, check sin-cgo). The index is a
// disposable projection of the arnés tree the daemon knows about via
// ports.ArnesRegistry — NOT the ~/.claude conversation JSONL, a separate corpus read by
// internal/adapters/history: if the .db is deleted or its schema_version drifts, it is
// rebuilt from scratch, never migrated.
package index

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/alpacapurpura/arnesia/dogfood"
	"github.com/alpacapurpura/arnesia/internal/domain"
	"github.com/alpacapurpura/arnesia/internal/ports"
	_ "modernc.org/sqlite" // driver registration only ("sqlite"); pure Go, sin CGO.
)

// ErrNotFound is returned when a harness id is unknown to the index.
var ErrNotFound = errors.New("harness not found")

// errSchemaMismatch is the internal sentinel New uses to trigger a wipe-and-recreate
// (RF-208) — it never escapes New.
var errSchemaMismatch = errors.New("index: schema_version mismatch")

// driverName is the name modernc.org/sqlite registers itself under via database/sql.
const driverName = "sqlite"

// schemaVersion identifies the shape of the `graphs`/`schema_meta` tables. Bumping it
// is how a shape change ships: no ALTER TABLE, ever — an on-disk .db whose
// schema_meta.version disagrees is deleted whole and rebuilt (RF-208).
const schemaVersion = 1

const (
	createMetaTable = `CREATE TABLE IF NOT EXISTS schema_meta (version INTEGER NOT NULL)`
	// graphs is the single JSON-blob table (decisiones.md D4): every IndexPort method
	// operates on a whole domain.Graph, no caller filters by node/edge, so there is
	// nothing to normalize yet.
	createGraphsTable = `CREATE TABLE IF NOT EXISTS graphs (
		clave TEXT PRIMARY KEY,
		graph_json TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`
	upsertGraphSQL = `INSERT INTO graphs (clave, graph_json, updated_at) VALUES (?, ?, ?)
		ON CONFLICT(clave) DO UPDATE SET graph_json = excluded.graph_json, updated_at = excluded.updated_at`
)

// Store is an on-disk SQLite ports.IndexPort. Two handles share the file (boundary
// L2): writer has SetMaxOpenConns(1) so database/sql's own pool serializes every
// write — no SQLITE_BUSY between goroutines of this process — while reader is pooled
// for concurrent reads under WAL. Safe for concurrent use.
type Store struct {
	writer *sql.DB
	reader *sql.DB
	// reg + load son la fuente que Rebuild reconstruye (RF-207): el árbol de arneses
	// conocido por ArnesRegistry, cargado por el loader real. Ambos pueden ser nil —
	// un Store que nunca llama Rebuild (p.ej. tests que solo ejercitan Query/List/
	// Upsert) no los necesita; Rebuild sobre un Store así simplemente vacía el índice.
	reg  ports.ArnesRegistry
	load func(dir string) (domain.Graph, error)
}

var _ ports.IndexPort = (*Store)(nil)

// New opens (or creates) the SQLite index at path, wiring the ports.ArnesRegistry +
// loader that Rebuild reads from (RF-207). An empty path defaults to
// ~/.arnesia/index.db (mirrors store.NewRegistry/NewArnesRegistry). reg/load may be
// nil for callers that never call Rebuild.
//
// New also seeds two demo/dogfood graphs so a fresh index has something real to serve
// before the first Rebuild runs (unchanged from the pre-SQLite behavior) — but Rebuild
// itself always replaces the whole table from reg, so no seed survives a Rebuild call
// (RF-207: "ningún dato demo/seed hardcodeado sobrevive en el índice").
func New(path string, reg ports.ArnesRegistry, load func(dir string) (domain.Graph, error)) (*Store, error) {
	path, err := resolvePath(path)
	if err != nil {
		return nil, err
	}
	if dir := filepath.Dir(path); dir != "." {
		if merr := os.MkdirAll(dir, 0o750); merr != nil {
			return nil, fmt.Errorf("index: mkdir %s: %w", dir, merr)
		}
	}
	writer, reader, err := openHandles(path)
	if err != nil {
		return nil, err
	}
	if verr := ensureSchema(context.Background(), writer); verr != nil {
		if !errors.Is(verr, errSchemaMismatch) {
			_ = writer.Close()
			_ = reader.Close()
			return nil, verr
		}
		// RF-208: el índice es desechable, nunca se migra — un schema_version que no
		// coincide borra el .db entero (sin ALTER TABLE) y arranca de cero.
		_ = writer.Close()
		_ = reader.Close()
		if werr := wipeFile(path); werr != nil {
			return nil, werr
		}
		if writer, reader, err = openHandles(path); err != nil {
			return nil, err
		}
		if verr := ensureSchema(context.Background(), writer); verr != nil {
			_ = writer.Close()
			_ = reader.Close()
			return nil, fmt.Errorf("index: schema tras wipe: %w", verr)
		}
	}
	s := &Store{writer: writer, reader: reader, reg: reg, load: load}
	s.seed()
	return s, nil
}

// Close releases both handles. Not part of ports.IndexPort — callers that want a
// clean shutdown may call it explicitly; the index is disposable (WAL is crash-safe
// by design), so skipping it is never a correctness problem, only a leaked fd.
func (s *Store) Close() error {
	werr := s.writer.Close()
	rerr := s.reader.Close()
	if werr != nil {
		return werr
	}
	return rerr
}

// Rebuild reconstructs the whole index from ports.ArnesRegistry (RF-207, decisiones.md
// D6 — NOT the Portafolio, see that decision for why): every prior row (seed data
// included) is cleared inside one transaction, then each registered arnés is reloaded
// and upserted. An arnés whose tree fails to load is indexed Degradado:true, never
// dropped silently (mirrors usecase.NewTurnReindexer's degrade-on-failure pattern).
func (s *Store) Rebuild(ctx context.Context) error {
	var entries []ports.ArnesPath
	if s.reg != nil {
		entries = s.reg.List()
	}
	tx, err := s.writer.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("index: rebuild begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once Commit succeeds.

	if _, err := tx.ExecContext(ctx, `DELETE FROM graphs`); err != nil {
		return fmt.Errorf("index: rebuild clear: %w", err)
	}
	for _, ap := range entries {
		if ap.Arnes == "" {
			continue // ArnesRegistry.Register ya rechaza ids vacíos; guarda defensiva.
		}
		g, lerr := s.loadGraph(ap.Path)
		if lerr != nil {
			slog.Warn("rebuild: arnés no cargable — se indexa degradado", "arnes", ap.Arnes, "path", ap.Path, "err", lerr)
			g = domain.Graph{Nodes: []domain.Box{}, Degradado: true}
		}
		if g.Arnes == nil {
			g.Arnes = &domain.Arnes{ID: ap.Arnes, Nombre: ap.Arnes}
			g.Degradado = true
		}
		raw, merr := json.Marshal(g)
		if merr != nil {
			return fmt.Errorf("index: rebuild encode %s: %w", ap.Arnes, merr)
		}
		if _, eerr := tx.ExecContext(ctx, upsertGraphSQL, ap.Arnes, string(raw), nowStamp()); eerr != nil {
			return fmt.Errorf("index: rebuild upsert %s: %w", ap.Arnes, eerr)
		}
	}
	return tx.Commit()
}

// loadGraph calls the configured loader, or fails honestly if none was wired.
func (s *Store) loadGraph(path string) (domain.Graph, error) {
	if s.load == nil {
		return domain.Graph{}, errors.New("index: sin loader configurado")
	}
	return s.load(path)
}

// Query returns the graph of one harness, or ErrNotFound.
func (s *Store) Query(ctx context.Context, harnessID string) (domain.Graph, error) {
	var raw string
	err := s.reader.QueryRowContext(ctx, `SELECT graph_json FROM graphs WHERE clave = ?`, harnessID).Scan(&raw)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Graph{}, ErrNotFound
	case err != nil:
		return domain.Graph{}, fmt.Errorf("index: query %s: %w", harnessID, err)
	}
	return decodeGraph([]byte(raw))
}

// List returns every indexed harness graph, ordered by clave for a stable portfolio (S1).
func (s *Store) List(ctx context.Context) ([]domain.Graph, error) {
	rows, err := s.reader.QueryContext(ctx, `SELECT graph_json FROM graphs ORDER BY clave`)
	if err != nil {
		return nil, fmt.Errorf("index: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []domain.Graph{}
	for rows.Next() {
		var raw string
		if serr := rows.Scan(&raw); serr != nil {
			return nil, fmt.Errorf("index: list scan: %w", serr)
		}
		g, derr := decodeGraph([]byte(raw))
		if derr != nil {
			return nil, derr
		}
		out = append(out, g)
	}
	if rerr := rows.Err(); rerr != nil {
		return nil, fmt.Errorf("index: list rows: %w", rerr)
	}
	return out, nil
}

// Upsert inserts or replaces one harness graph bajo `clave` — la llave AUTORITATIVA que el
// caller decide (deuda BACKLOG «re-key (home,id,scope)», 2026-07-23: nunca se re-deriva del
// propio `g.Arnes.ID`, que dos arneses distintos pueden compartir). Un grafo sin manifiesto
// no es indexable (no hay nada que mostrar) — error explícito, jamás un grafo inventado.
func (s *Store) Upsert(ctx context.Context, clave string, g domain.Graph) error {
	if clave == "" {
		return errors.New("index: clave vacía — no indexable")
	}
	if g.Arnes == nil {
		return errors.New("index: grafo sin manifiesto (arnes.id) — no indexable")
	}
	raw, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("index: encode graph %s: %w", clave, err)
	}
	if _, err := s.writer.ExecContext(ctx, upsertGraphSQL, clave, string(raw), nowStamp()); err != nil {
		return fmt.Errorf("index: upsert %s: %w", clave, err)
	}
	return nil
}

// seed loads a minimal demo harness plus the two embedded dogfood graphs so a fresh
// index has something real to serve immediately (unchanged intent from the pre-SQLite
// in-memory Store — see New's doc comment for why Rebuild does not preserve this).
func (s *Store) seed() {
	demo := domain.Graph{
		Arnes: &domain.Arnes{
			ID:       "demo",
			Rol:      "backend",
			Proceso:  "desarrollo",
			Empresas: []string{"alpacapurpura"},
			ReportaA: nil, // raíz (emite null, válido contra graph.l0).
			Canal:    domain.CanalBeta,
			Fases:    []domain.Fase{"spec"},
			Spine: &domain.Spine{
				Inicial:      "grill",
				Terminales:   []string{"spec"},
				Estados:      []string{"grill", "spec"},
				Transiciones: []domain.Transicion{{De: "grill", A: "spec"}},
			},
		},
		Nodes: []domain.Box{
			{
				ID:     "guard-format",
				Clase:  domain.ClaseHook,
				Nombre: "gofmt guard",
				Banda:  domain.BandaGuardia,
			},
			{
				ID:     "spec",
				Clase:  domain.ClaseSkill,
				Nombre: "escribir spec",
				Banda:  domain.BandaFase,
				Fase:   "spec",
				Estado: "grill -> spec",
				Contract: &domain.Contract{
					Why:       "convertir la conversación en un spec ejecutable",
					Clase:     domain.ClaseSkill,
					Arquetipo: domain.ArqExcepcion,
					Perfil:    domain.PerfilT2,
					Caja:      true,
					Fase:      "spec",
					Estado:    "grill -> spec",
					Gate:      &domain.Gate{Tipo: domain.GateManual, Detalle: "revisión humana del spec"},
				},
			},
			{
				ID:     "std-go",
				Clase:  domain.ClaseRule,
				Nombre: "estándar Go",
				Banda:  domain.BandaBase,
			},
		},
		Edges: []domain.Edge{
			{De: "spec", A: "std-go", Tipo: domain.EdgeLee},
		},
	}
	if err := s.Upsert(context.Background(), demo.Arnes.ID, demo); err != nil {
		slog.Error("index: seed demo graph", "err", err)
	}
	// Sembrar los grafos de arneses embebidos: dev-full-cycle (el dogfood REAL, honesto)
	// + content-studio-full (el showcase kitchen-sink, HS-09 Hito 2). Un decode fallido
	// de un asset embebido y testeado es un error del programador, pero se loggea y
	// sigue para que el demo no muera entero por eso.
	for name, raw := range map[string][]byte{
		"dev-full-cycle":      dogfood.DevFullCycleJSON,
		"content-studio-full": dogfood.ContentStudioFullJSON,
	} {
		g, err := decodeGraph(raw)
		if err != nil {
			slog.Error("index: seed embedded graph", "arnes", name, "err", err)
			continue
		}
		if g.Arnes == nil {
			continue
		}
		if uerr := s.Upsert(context.Background(), g.Arnes.ID, g); uerr != nil {
			slog.Error("index: seed embedded graph upsert", "arnes", name, "err", uerr)
		}
	}
}

// decodeGraph unmarshals an embedded/stored L0 graph. The JSON keys mirror the domain
// tags exactly, so it decodes straight into a Graph with no field mapping.
func decodeGraph(raw []byte) (domain.Graph, error) {
	var g domain.Graph
	if err := json.Unmarshal(raw, &g); err != nil {
		return domain.Graph{}, fmt.Errorf("decode graph: %w", err)
	}
	return g, nil
}

// resolvePath defaults an empty path to ~/.arnesia/index.db, mirroring
// store.NewRegistry/NewArnesRegistry's convention for the daemon's other file-backed
// adapters.
func resolvePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("index: resolve home: %w", err)
	}
	return filepath.Join(home, ".arnesia", "index.db"), nil
}

// openHandles opens the writer (single-conn, serializes writes) and reader (pooled)
// *sql.DB pair over the same file (boundary L2 · WAL + busy_timeout=5000 +
// synchronous=NORMAL + BEGIN IMMEDIATE on the writer's transactions).
func openHandles(path string) (writer, reader *sql.DB, err error) {
	writerDSN := "file:" + path +
		"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate"
	writer, err = sql.Open(driverName, writerDSN)
	if err != nil {
		return nil, nil, fmt.Errorf("index: open writer: %w", err)
	}
	writer.SetMaxOpenConns(1) // RF-209: un solo escritor — database/sql serializa todo Exec sobre él.

	readerDSN := "file:" + path + "?_pragma=busy_timeout(5000)"
	reader, err = sql.Open(driverName, readerDSN)
	if err != nil {
		_ = writer.Close()
		return nil, nil, fmt.Errorf("index: open reader: %w", err)
	}
	return writer, reader, nil
}

// ensureSchema creates schema_meta/graphs if missing and checks schema_meta.version:
// a fresh file gets the current schemaVersion stamped; an existing one that disagrees
// returns errSchemaMismatch (caller wipes and recreates — RF-208, never migrated).
func ensureSchema(ctx context.Context, db *sql.DB) error {
	if _, cerr := db.ExecContext(ctx, createMetaTable); cerr != nil {
		return fmt.Errorf("index: create schema_meta: %w", cerr)
	}
	if _, cerr := db.ExecContext(ctx, createGraphsTable); cerr != nil {
		return fmt.Errorf("index: create graphs: %w", cerr)
	}
	var version int
	qerr := db.QueryRowContext(ctx, `SELECT version FROM schema_meta LIMIT 1`).Scan(&version)
	switch {
	case errors.Is(qerr, sql.ErrNoRows):
		if _, ierr := db.ExecContext(ctx, `INSERT INTO schema_meta (version) VALUES (?)`, schemaVersion); ierr != nil {
			return fmt.Errorf("index: seed schema_meta: %w", ierr)
		}
		return nil
	case qerr != nil:
		return fmt.Errorf("index: read schema_meta: %w", qerr)
	case version != schemaVersion:
		return errSchemaMismatch
	default:
		return nil
	}
}

// wipeFile removes the .db plus its WAL/SHM sidecar files (RF-208 — a schema mismatch
// discards the whole index, never migrates it). Missing sidecars are not an error.
func wipeFile(path string) error {
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if rerr := os.Remove(p); rerr != nil && !os.IsNotExist(rerr) {
			return fmt.Errorf("index: wipe %s: %w", p, rerr)
		}
	}
	return nil
}

// nowStamp is the updated_at value written on every insert/replace.
func nowStamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
