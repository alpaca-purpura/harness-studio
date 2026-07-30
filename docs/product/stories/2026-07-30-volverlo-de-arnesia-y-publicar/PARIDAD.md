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
| AC-e: E2E contra el binario INSTALABLE | **Auto-verificado 2026-07-30 con el payload del `.deb` v0.6.0** (`dpkg-deb -x`, daemon sandbox): escanear crudo → 400 incompleto → 200 → sello en disco `[empresas,id,nombre,proceso,reporta_a,rol]` (sin version) | ✅ |
| **Residuo humano: mismo flujo en la app instalada (sudo) + firma** | | ⬜ **gate** |

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
| AC-e: publicar contra repo GitHub **REAL** | **Auto-verificado 2026-07-30** (daemon del `.deb`, repo descartable `arnesia-e2e-publicar-20260730`, borrado al cierre): registrar propio → catálogo leído REMOTO vía gh → traer canónico → bump 0.2.0 → **POST publicaciones 200** → verificado EN GITHUB: tag `demo-pub/v0.2.0` + árbol `plugins/demo-pub/0.2.0/` + fila catalogo `sello: 2607301251` → **re-publicar 409** («una versión publicada no se pisa») | ✅ |
| Sin-auth honesto | daemon con HOME aislado sin config de gh: «gh está instalado pero no autenticado, y ARNESIA_GH_TOKEN está vacío» (vía de registro; el clasificador del publisher queda cubierto por unit) | ✅ |
| **Residuo humano: mismo flujo desde la app instalada + firma** | | ⬜ **gate** |

## Desviaciones declaradas (aceptar u objetar al firmar)

1. Symlinks del canónico NO viajan al estante (se omiten con aviso visible).
2. El RMW puede normalizar el ORDEN de claves de marketplace.json (contenido intacto).
3. Changelog-de-arnés y mutación de canales = Fase 2 (B-D6).
4. Celda Publicar del catálogo pinta habilitada sin onClick (la puerta es el drawer).

## Firma

- [ ] 🧑‍⚖️ B1 en vivo (laptop) — fecha:
- [ ] 🧑‍⚖️ B2 en vivo (repo real) — fecha:
