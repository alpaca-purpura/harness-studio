// Package dogfood embeds the recorded dogfood arnés (dev-full-cycle) so the in-memory
// index can seed a REAL arnés — the one the Map renders in HS-09 Hito 1 — before the
// JSONL indexer lands. The bytes are the source of record until the live loader (Hito 3)
// walks ~/.claude. go:embed cannot reach files above its package dir, so this tiny
// package lives beside the JSON and re-exports it to internal/adapters/index.
package dogfood

import _ "embed"

// DevFullCycleJSON is the L0 graph of the dogfood arnés, verbatim. Its JSON keys mirror
// the domain tags exactly (nodos/clase/banda/fase/contract/de/a/tipo), so it unmarshals
// straight into domain.Graph with no field mapping.
//
//go:embed dev-full-cycle.graph.json
var DevFullCycleJSON []byte
