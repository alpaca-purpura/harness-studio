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

// ContentStudioFullJSON is the showcase arnés (content-studio-full): a deliberately
// maximal, kitchen-sink graph that exercises every L0 casuistic (10 clases · 7 bandas ·
// 4 canal · 5 procedencia · 4 arquetipo · 3 perfil · 4 gate.tipo · both origen · rework
// spine). Its rol is Editorial — NON-engineering, on purpose — so the Map proves the
// agnostic-to-rubro principle (VISION P7). Conformance-valid (arnesia conformance --arnes).
// Design record: research/2026-07-06-plan-hito2-doctrina-edicion-showcase/. HS-09 Hito 2.
//
//go:embed content-studio-full.graph.json
var ContentStudioFullJSON []byte
