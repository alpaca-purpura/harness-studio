# RF-166 — Gate de calidad (round-trip, sin pérdida de entrega propia)

Verificación final tras RF-160..165, contra el código YA commiteado (973cfdc, 55cc8db,
21a33f5, 648a863). Todo corrido en vivo el 2026-07-08, sin nada simulado.

## 1 · Suites automatizadas

| comando | resultado |
|---|---|
| `go build ./...` | limpio |
| `go test ./... -count=1` | verde, todos los paquetes |
| `go test ./... -race -count=1` | verde, todos los paquetes (sin data races) |
| `golangci-lint run ./...` | `0 issues` (repo completo, incluidas rutas ajenas a HS-17) |
| `pnpm run verify` (web/: typecheck+biome+depcruise+fsd+stylelint) | limpio (1 info de deprecación de config Biome, preexistente, no relacionado a HS-17) |

## 2 · `arnesia conformance --arnes` (dogfood real) — sin regresión

Antes de HS-17 (documentado en CLAUDE.md/LEDGER.md) y después de RF-161..165, idéntico:

```
conformance arnes:dev-full-cycle
  21 checks · pass 20 · fail 1 · error 0 · deferred 0 · n/a 0
```

El único fail es `art-es-path` (warn, honesto, preexistente — 3 cajas sin path de
etiqueta, nada que ver con este paquete).

## 3 · `arnesia conformance superficie-local-confinada` — checks nuevos reales

```
9 checks · pass 8 · fail 0 · error 0 · deferred 1 · n/a 0
  PASS  error  mcp-config-siempre         TestConfigSourceMCPAislado pasa
  PASS  error  setting-sources-siempre    TestConfigSourceSettingSourcesExcludeUser pasa
```

## 4 · Round-trip funcional — kit propio, doctrina, knowhow, permisos intactos

`SpawnArgs` es el ÚNICO punto de armado de argv para AMBOS consumidores — sesión Dock
(chat) y corrida T3 (`BoxConductor`) llaman al mismo `ports.AgentPort.Spawn` →
`Conductor.Spawn` → `SpawnArgs`. No hay una segunda ruta que pudiera divergir. Verificado
en vivo contra el argv de PRODUCCIÓN (RF-161+RF-162 aplicados, `post-settings.md` +
`experimento-safe-mode.md` de este paquete):

- **Kit propio (②):** `arnesia-kit:auditar-arnes` / `arnesia-kit:forjar-caja` presentes,
  invocables — idéntico al baseline pre-cambio.
- **Overlay de doctrina (①, `--append-system-prompt-file`):** cita textual exacta
  verificada (`# Doctrina ArnesIA — overlay del conductor...`).
- **Knowhow (①, `--add-dir`):** listado real de `~/.arnesia/knowhow/` confirmado legible.
- **CLAUDE.md propio del arnés:** cita textual exacta de `dogfood/dev-full-cycle/CLAUDE.md`
  (regla `std-spec`), sobrevive `--setting-sources project,local`.
- **Permisos derivados del rol** (`permisos-derivan-del-rol`, boundary aparte, eje de
  invocación no de config-carga): `permissionArgs` no se tocó en este paquete —
  `TestPermissionArgsMaterialization` sigue verde sin cambios (parte de la suite §1).

No se levantó el shell Tauri + navegador (sin superficie UI en este paquete — backend
puro, ver INDEX.md «Adaptación explícita»); el round-trip real se hizo al nivel del
conductor/CLI, que es donde vive el cambio.

## 5 · Medición de contexto — la baja cuantificada

Mismo cwd (`dogfood/dev-full-cycle`), mismo turno-sonda, mismos flags de inyección ①②;
única variable: `--mcp-config`+`--strict-mcp-config`+`--setting-sources project,local`
presentes o ausentes. `result.usage` del frame final de cada corrida
(`medicion-contexto-baseline.md` / `medicion-contexto-post.md`):

| métrica | baseline (sin D2/D3) | post (D2+D3) | Δ |
|---|---|---|---|
| `input_tokens` (fresco, no cacheado — la señal más limpia de superficie real) | 12 959 | 3 193 | **−75.4 %** |
| `cache_creation_input_tokens` | 19 832 | 16 186 | −18.4 % |
| `cache_read_input_tokens` | 23 323 | 15 206 | −34.8 % |
| total | 56 114 | 34 585 | −38.4 % |

`input_tokens` es la métrica más honesta acá (no depende de qué quedó cacheado de
corridas previas): el spawn deja de pagar contexto fresco por el MCP de cuenta y los
`enabledPlugins`/hooks personales del operador. Coincide en dirección y orden de magnitud
con el disparador original de HS-17 (62.2k tok de MCP ajenos vistos en `/context`).

## Veredicto

Sin regresión en ningún eje medido. El kit propio, la doctrina, el knowhow, el CLAUDE.md
del arnés y los permisos por rol quedan intactos; el costo de contexto ajeno bajó
cuantificado. RF-166 PASA — habilita RF-167 (cierre de ficha).
