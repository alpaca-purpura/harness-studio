---
regla: editor
version: 1.0
updated: 2026-07-05
status: proposed
ledger: HS-05
sources:
  - url: https://editorconfig.org/
    autoridad: estándar
    revisado: 2026-07-05
enforced_by:
  - /.editorconfig
severity: medium
---

# Indentación y EOL únicos vía EditorConfig

## L1 · Principio (estándar de industria)

**EditorConfig** = fuente única de indentación/EOL/charset para editores + formatters, respetada por la
mayoría de IDEs sin plugin y por Biome (`formatter.useEditorconfig: true`) y gofmt. *(estándar:
EditorConfig)*

## L2 · Realización (este repo Go+TS+Rust)

Config = **`/.editorconfig`** en la raíz: `end_of_line = lf`, `charset = utf-8`, `insert_final_newline =
true`, `trim_trailing_whitespace = true`; `indent_style = space`/`indent_size = 2` para TS/JSON/CSS/MD;
`[*.go]` → `indent_style = tab` (gofmt manda). Biome lo lee; gofmt fija Go a tabs → sin conflicto. ⇐ L1:
fuente única.

## Checklist evaluable

| id | qué chequea | severidad | señal | enforcer |
|----|-------------|-----------|-------|----------|
| editorconfig-existe | `/.editorconfig` define EOL/charset/final-newline/trailing-ws | warn | «sin fuente única de indentación» | /.editorconfig |
| go-tabs | `[*.go]` usa `indent_style = tab` (coherente con gofmt) | warn | «indent Go no-tab (choca con gofmt)» | /.editorconfig |

## Changelog

- 2026-07-05 · v1.0 · Nodo fundacional (HS-05). L1 = EditorConfig fuente única. L2 = `/.editorconfig`
  (lf/utf-8/final-newline; 2-space TS, tab Go); Biome y gofmt lo respetan. 2 checks.
