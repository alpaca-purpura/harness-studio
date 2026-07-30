# PARIDAD — volverlo de arnesia y publicar (B1 + B2)

> Evidencia recogida 2026-07-30. **Gate 🧑‍⚖️: PENDIENTE del operador** (fila «En vivo
> laptop»). Código en main: B1 `9c91449` · B2 `f50a24a`.

## B1 — sello válido

| Qué (RF/AC) | Evidencia | Estado |
|---|---|---|
| Sello escrito PASA `graph.l0.schema.json` real | `TestIdentificarSelloValidaContraSchemaReal` (SchemaSet real; la forma vieja `{id,nombre,version}` FALLA — bug pinchado) | ✅ |
| Incompleto ⇒ 400, NADA escrito | `TestIdentificarRechazaSelloIncompleto` + E2E sandbox: 400 «sello incompleto — rol, proceso y ≥1 empresa» | ✅ |
| `version` fuera del sello (B-D2) | assert invertido en `TestIdentificarSellaYRekey`; E2E: sello en disco sin `version` | ✅ |
| plugin.json mínimo solo donde es seguro | `TestIdentificarNoGeneraPluginJSONEnProyectoInstalado`; E2E: proyecto-instalado ⇒ sin plugin.json + aviso «sin plugin.json: no publicable en esta forma» | ✅ |
| Marketplace ⇒ re-key con Home | E2E sandbox: `sin-home~~c1dbb2fb98fb` → `(github.com/alpacapurpura/prenter-marketplace, mi-arnes)`; CLI (2º proceso) ve la entrada | ✅ |
| Form FE 4 campos + disabled honesto | 3 stories drawer (verify exit 0) | ✅ |
| Hoja `versionado-arnes.md` | v1.0 `proposed` → v1.1 con enforcers reales (B2) | ✅ |
| **AC-e: E2E en laptop del operador** | Portafolio → dir crudo → ✦ Identificar → sello | ⬜ **gate** |

## B2 — publicar

| Qué | Evidencia | Estado |
|---|---|---|
| Publica: árbol `plugins/<id>/<v>/` + marketplace.json round-trip + catalogo con `sello` + tag | tests adapter contra bares git reales desde fixture prenter | ✅ |
| Idempotencia: republicar ⇒ 409, nada tocado | test working-tree-limpio + remoto sin commits nuevos | ✅ |
| Push rechazado ⇒ 409 «reintentá» (sin force) | shim determinista (rival pushea antes) | ✅ |
| Gate conformance rojo BLOQUEA con reporte | `TestPublicarConformanceRojoBloquea` + body `{error, conformance}` | ✅ |
| Guardas 400/503 (sin-canónico/home/no-propio/semver/auth) | 12 tests usecase + tabla HTTP | ✅ |
| Botón FE vivo + checks FAIL listados | 4 stories drawer + fix asserts motivo-viejo | ✅ |
| CLI `arnesia publish <clave>` | recableado al usecase real (no se recortó) | ✅ |
| **AC-e: publicar contra repo GitHub REAL desde laptop** (re-publicar ⇒ 409 · `gh auth logout` ⇒ 503) | exige credenciales del operador | ⬜ **gate** |

## Desviaciones declaradas (aceptar u objetar al firmar)

1. Symlinks del canónico NO viajan al estante (se omiten con aviso visible).
2. El RMW puede normalizar el ORDEN de claves de marketplace.json (contenido intacto).
3. Changelog-de-arnés y mutación de canales = Fase 2 (B-D6).
4. Celda Publicar del catálogo pinta habilitada sin onClick (la puerta es el drawer).

## Firma

- [ ] 🧑‍⚖️ B1 en vivo (laptop) — fecha:
- [ ] 🧑‍⚖️ B2 en vivo (repo real) — fecha:
