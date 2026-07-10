// Package mechanism implements the conformance mechanism adapters (ports.MechanismAdapter,
// one per domain.Mecanismo) plus a reusable JSON-Schema validator. Each adapter runs one
// check against a target; a mechanism that cannot execute deterministically here returns
// VeredictoDiferido — honest, never a fabricated pass (VISION p6/p10).
package mechanism

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

// SchemaSet loads and caches JSON Schemas from an fs.FS, resolving sibling $refs (e.g.
// graph.l0.schema.json → box.contract.schema.json) from the same root. It is the shared
// schema-validation engine used by the arnés conformance path and the schema adapter.
// El FS es el disco del repo en dev y el go:embed del binario en cliente (portabilidad
// firmada HS-10) — misma lógica, cero divergencia.
type SchemaSet struct {
	fsys     fs.FS
	mu       sync.Mutex
	resolved map[string]*jsonschema.Resolved
}

// NewSchemaSet returns a SchemaSet rooted at the on-disk schema directory
// (docs/architecture/contracts/schema under the repo root).
func NewSchemaSet(schemaDir string) *SchemaSet { return NewSchemaSetFS(os.DirFS(schemaDir)) }

// NewSchemaSetFS returns a SchemaSet over any fs.FS whose root contains the schema
// files (p.ej. fs.Sub(doctrina.Files, "docs/architecture/contracts/schema")).
func NewSchemaSetFS(fsys fs.FS) *SchemaSet {
	return &SchemaSet{fsys: fsys, resolved: map[string]*jsonschema.Resolved{}}
}

// resolvedFor loads+resolves a schema file (by basename), caching the result. The Loader
// resolves any remote $ref by basename against the same directory so the box-contract
// $ref inside graph.l0 works without network.
func (s *SchemaSet) resolvedFor(file string) (*jsonschema.Resolved, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.resolved[file]; ok {
		return r, nil
	}
	root, err := s.loadSchema(file)
	if err != nil {
		return nil, err
	}
	res, err := root.Resolve(&jsonschema.ResolveOptions{Loader: s.siblingLoader})
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", file, err)
	}
	s.resolved[file] = res
	return res, nil
}

func (s *SchemaSet) loadSchema(file string) (*jsonschema.Schema, error) {
	raw, err := fs.ReadFile(s.fsys, path.Base(file))
	if err != nil {
		return nil, err
	}
	var sch jsonschema.Schema
	if err := json.Unmarshal(raw, &sch); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}
	return &sch, nil
}

// siblingLoader resolves a remote $ref URI to a schema file in the same directory,
// matched by the URI's basename.
func (s *SchemaSet) siblingLoader(uri *url.URL) (*jsonschema.Schema, error) {
	return s.loadSchema(path.Base(uri.Path))
}

// Validate checks a JSON-decoded instance against the named schema file. A nil error
// means the instance conforms.
func (s *SchemaSet) Validate(schemaFile string, instance any) error {
	res, err := s.resolvedFor(schemaFile)
	if err != nil {
		return err
	}
	return res.Validate(instance)
}

// ValidateJSON marshals v to JSON, decodes it to the generic shape the validator wants,
// and validates it against schemaFile. Handy for validating domain structs.
func (s *SchemaSet) ValidateJSON(schemaFile string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var inst any
	if err := json.Unmarshal(b, &inst); err != nil {
		return err
	}
	return s.Validate(schemaFile, inst)
}
