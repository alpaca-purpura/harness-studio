# PARIDAD — Portafolio · Slice 0 «Cimientos» (spec ↔ implementación ↔ verificación)

> Construido por Sonnet 5 (constructor), plan de Fable 5 (arquitecto). Ejecutados T1→T9 en orden,
> un commit Conventional por ticket, gate local (§M del plan) verde antes de cada commit — ver
> `git log --oneline` (`feat(portafolio): T1..T9 — …`). **Firma 🧑‍⚖️ del operador PENDIENTE** — esta
> hoja NO la simula.
>
> **Verificación final (T9, 2026-07-13):** `go build/vet/test -race` ✓ (0 fail, 1 flake
> pre-existente `TestResumeAutoSana` confirmado ajeno — timing de resume de sesión, no toca
> Portafolio) · `golangci-lint run` 0 issues · `go-arch-lint check` OK · `conformance --todo`
> **253 checks · pass 42 · fail 0 · error 0 · deferred 211** (sin regresión: baseline pre-Slice-0
> era 249·42·0·207 — los +4 checks/+4 deferred son el boundary nuevo, honestamente no-mecánico) ·
> `conformance --arnes dogfood/dev-full-cycle.graph.json` **21·20·1** (warn `art-es-path`
> preexistente, sin regresión) · `pnpm run verify` (web) verde · **E2E vivo contra la máquina real**
> (ver §E2E abajo) · `cap_doctor.py` 87 capabilities válidas · `estado.sh --check` sin drift.

## E2E vivo (máquina real, read-only, 2026-07-13)

Corrido con el binario compilado (`go build -o arnesia-e2e ./cmd/arnesia`), `$HOME` real —
**dos bugs reales encontrados y corregidos en el momento** (S0-D14/S0-D15 en `decisiones.md`):

1. **`installed_plugins.json` real** anida bajo una clave `plugins` (`{"version":2,"plugins":{…}}`),
   no en el nivel raíz como S0-D1 había asumido — `leerInstalledPlugins` corregido; sin este fix,
   TODO hallazgo `referenciada-cc` real salía `no-legible` en silencio (el `json.Unmarshal` fallaba
   parseando cada `id@mkt` como si fuera directamente un array).
2. **Deriva con `home` vacío** (el caso más común real: un plugin CC sin `arnes.l0.json`, C-N-14) no
   evaluaba NUNCA aunque hubiera checkout local — `candidatoDe` ahora usa `origen.Registry`
   canonicalizado como fallback de `home` para la llamada a `EvaluarDeriva` (la identidad sigue
   honestamente provisional; solo la deriva gana la fuente).

**`arnesia portafolio escanear ~/Proyectos/harness-studio`** (este mismo repo):
- `proyecto-instalado` (root, `.claude/` real) — identidad provisional (sin `arnes.l0.json` en la
  raíz del repo, correcto: este repo no es-un-arnés, es la fábrica) + eslabón `git-proyecto`
  `https://github.com/alpacapurpura/harness-studio.git` (RN-GIT-1: hogar del proyecto, no del arnés).
- `referenciada-cc` para `harness@prenter-marketplace` (`.claude/settings.json#enabledPlugins`) →
  `Aviso: "sin record de instalación para …"` — HONESTO: cotejado a mano contra
  `~/.claude/plugins/installed_plugins.json`, esta ruta exacta genuinamente no tiene entry ahí (el
  kit se habilitó por edición directa del settings, no por `/plugin add`) — el fallback ni fabrica
  ni omite, avisa.

**`arnesia portafolio escanear ~/Proyectos/luana-vitalia`** (monorepo real, más rico que lo previsto
en el plan): 14 hallazgos — `proyecto-instalado` del root + de cada subpaquete anidado
(`comunify/`, `fitflow/`, `vitalia/`, …, C-N-13 monorepo) + 2 `referenciada-cc` con `Aviso`
sin-record + **`harness@prenter-marketplace` con datos REALES**: `registry:
alpacapurpura/prenter-marketplace · version: 0.5.2` (cotejado a mano, coincide con
`installed_plugins.json`) y **`deriva: en-deriva` REAL** — hash de contenido distinto entre
`~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2/` y el checkout del marketplace en
`~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.2/` (el layout por-versión
SÍ existe ahí — evaluación real, no `deriva-no-evaluable`).

**Ciclo `agregar` → `listar` → corromper a mano → `listar` → `serve` → `desvincular`:**
- `agregar` persistió 2 candidatos reales (harness-studio + el `harness` de luana-vitalia) en
  `~/.arnesia/portafolio.json`; `listar` reprodujo exactamente lo mismo; **inspección a mano del
  archivo** confirmó el JSON en disco byte-a-byte igual al `listar`.
- **Corrompí una entrada a mano** (`python3` agregó `{"identidad":"esto rompe el shape…"}` al
  array `entradas`): `listar` siguió mostrando las 2 sanas + la corrupta visible aparte con
  `motivo` real (`json: cannot unmarshal string into Go struct field…`) — degradación honesta.
- **`arnesia serve --addr 127.0.0.1:4277` booteó igual** con el archivo corrupto presente (log:
  `INFO arnesia serve addr=127.0.0.1:4277`, sin error) — `GET /healthz` → `{"status":"ok"}`.
- `curl GET /api/portafolio` → mismo contenido que el CLI `listar` (entradas + corruptas).
- `curl POST /api/portafolio/escaneos {"path":"~/Proyectos/devhub"}` → candidato real, NO persistió
  (verificado: un segundo GET /api/portafolio antes del POST /proyectos no lo mostraba).
- `curl POST /api/portafolio/proyectos {"path":…, "elegidos":[clave]}` → persistió.
- `curl DELETE /api/portafolio/arneses/{clave}` → `{"desvinculado":true}`; un segundo DELETE de la
  misma clave → **404** (verificado con `-o /dev/null -w "%{http_code}"`).
- **Estado final**: `~/.arnesia/portafolio.json` restaurado a como estaba antes de este E2E (no
  existía) — el E2E no dejó basura en la máquina real, tal como manda la instrucción read-only.

## Spec §9 (Slice 0 — entra) ↔ realidad

| Item spec §9 | Estado | Evidencia |
|---|---|---|
| Store de portafolio nuevo (separado, degrada honesto) | ✅ | `portafolio.Store` (T6); `TestStoreDegradaHonesto`, `TestStoreArchivoTotalmenteIlegible` + E2E corrompido a mano |
| Walker `.claude/plugins/<id>/` multi-arnés | ✅ | `portafolio.Scanner` (T4); `TestScannerMaterializada` + E2E monorepo `luana-vitalia` (14 hallazgos reales) |
| Fallback `plugin.json` en el loader | ✅ | `loader.leerManifiesto` (T3, C-N-14); `TestLoaderFallbackPluginJSON` + E2E (`harness` cargado sin `arnes.l0.json`) |
| Campo `version` en el dominio | ✅ | `domain.Arnes.Version` (T1); `TestLoaderVersionDesdePluginJSON` |
| `empresas[]` en `domain.Arnes` + reconciliar seed | ✅ | `Empresas []string` + `UnmarshalJSON` tolerante (T1); seed `index/store.go` migrado (T1) |
| `registries[]` (colección) | ✅ (redirigido, S0-D3) | Vive en `domain.EntradaPortafolio.Registries`, NO en `domain.Arnes` — la procedencia de la copia no viaja en el manifiesto (S0-D3, decisión ya cerrada antes del build) |
| Identidad `(home,id)` + canonicalización + clave provisional | ✅ | `IdentidadArnes.Clave()`, `CanonicalizarRepo`, `ResolverIdentidad` (T1); 19 casos de test + E2E |
| Cadena de origen collect-all (eslabones 1-2-3-5) | ✅ | `ResolverOrigen` (T1) + `Scanner` los recolecta (T4); eslabón 4 (CC) investigado S0-D1 y CORREGIDO en T9 (S0-D14) |
| Deriva = hash vs `home/plugins/<id>/<versión>/` | ✅ | `HashFormaPlugin`/`EvaluarDeriva` (T5); `TestDerivaNuncaSemver` + E2E real (`en-deriva` genuino) |
| Detector 3° del lock `.devstudio/arneses.yaml` | ✅ | `Scanner.escanearLock` (T4); `TestScannerLockDevstudio` (ningún proyecto real lo tiene aún, S0-D1 — cubierto por fixture) |
| Ubicación del canónico fuera de `~/.arnesia` | ✅ (forma; clone en Slice 2) | `domain.Canonico{Path,Version}` + clasificación RN-IDENT-4 (T1/T6); el `git clone` real es Slice 2 (S0-D10, fuera de alcance) |
| Capabilities nuevas (5) | ✅ | `docs/product/capabilities/portafolio/*.yaml`, status `vivo`, 27 tests en `valida:` (T1→T8) |
| Reusa/extiende `loader.leer-manifiesto` (+fallback) | ✅ | CAP-17 `change_log: extend` (T8) |
| Reusa/extiende `loader.reconocer-forma-fisica` (+plugin.json) | ⚠ no aplicado | `detectarElementos` (CAP-15) no cambió — el fallback plugin.json vive en `leerManifiesto` (CAP-17), no en la detección de forma física; la nota original del spec era imprecisa, no se le fabrica un `change_log` |
| Registro (nuevo store, NO extiende el de cwd) | ✅ | `~/.arnesia/portafolio.json` separado de `arneses.json` (A1, S0-D5) |

## Reglas de negocio (BR-1..11) aplicables a Slice 0

| BR | Estado | Evidencia |
|---|---|---|
| BR-1 unidad = identidad `(home,id)` | ✅ | `IdentidadArnes`, `TestStoreUpsertMergePorIdentidad` |
| BR-2 solo canónico editable; paths ∅ | ⚠ parcial | Sin editor en Slice 0 (nada que editar todavía) — la disjunción práctica la da RN-IDENT-4 (checkout≠instalación); una validación explícita de disjunción de paths queda para cuando el editor exista (Slice 1+) |
| BR-3 origen honesto y trazable | ✅ | `ResolverOrigen` + `Eslabones[]`; `TestResolverOrigen` |
| BR-4 deriva por hash, nunca semver | ✅ | `TestDerivaNuncaSemver` + E2E real |
| BR-5 remote proyecto ≠ home/registry del arnés | ✅ | eslabón `git-proyecto` campo `proyecto-remote` nunca alimenta `Registry` (estructural, no por chequeo) |
| BR-6 reparar (overwrite dir privado, merge compartidas) | ⏸ diferido | Slice 5, fuera de alcance (P4 del plan) |
| BR-7 publicar (pull→conformance→bump→push→tag) | ⏸ diferido | Slice 3, fuera de alcance |
| BR-8 lo no construido = deshabilitado+tooltip | ⏸ diferido | Concern de FE, Slice 1 |
| BR-9 duplicado por remote/path canonicalizado | ✅ | `mergeInstalaciones` dedup por `InstallPath`; `TestStoreUpsertMergePorIdentidad` (re-upsert no duplica) |
| BR-10 GitHub = conductor (`gh`/PAT), nunca proxy | ⏸ diferido | Sin red en Slice 0 (P4); Slice 2+ |
| BR-11 store degrada honesto | ✅ | `TestStoreDegradaHonesto` + `TestStoreArchivoTotalmenteIlegible` + E2E corrompido a mano + `serve` bootea |

## Deltas de dominio (D-DOM-1..7)

Los 7 implementados: D-DOM-1 `Version` · D-DOM-2 `Empresas[]`+`Marketplace` crudo (S0-D3 aterrizó
el alcance exacto) · D-DOM-3 walker multi-arnés · D-DOM-4 loader fallback plugin.json · D-DOM-5
store separado · D-DOM-6 detector lock · D-DOM-7 canonicalización. Ver tabla spec §9 arriba para
el pointer de cada uno.

## Decisiones S0-D1..D15 — todas resueltas

S0-D1..D11 son del arquitecto (plan, `decisiones.md`), cerradas antes del build. S0-D12..D15 las
agregó el constructor durante la ejecución (desviaciones anti-drift documentadas EN el mismo turno
que ocurrieron, por instrucción del prompt de arranque):

- **S0-D12** — colisión de nombre `Origen` (→ `domain.OrigenPortafolio`) + 2 fixes de compilación
  forzados a T1 (`router.go` `harnessSummary.Empresas[]`, `index/store.go` seed).
- **S0-D13** — puerto `ports.DerivaEvaluator` nuevo (el plan dejaba "`refs …`" sin cerrar cómo el
  usecase invoca `EvaluarDeriva` sin importar el adapter).
- **S0-D14** — `installed_plugins.json` real anida bajo `plugins` (bug real, corregido en T9).
- **S0-D15** — deriva usa `origen.Registry` como fallback de `home` (gap real, corregido en T9;
  sin este fix la deriva NUNCA se evaluaba en el caso más común).

Ninguna decisión quedó abierta. El detalle completo de cada una vive en `decisiones.md`.

## Desviaciones registradas (a firmar en gate humano 🧑‍⚖️)

1. **`TestLoaderSinManifiesto` cambió de fixture** (T3): el test verificaba el comportamiento VIEJO
   que C-N-14 mandaba corregir (plugin.json sin arnes.l0.json → `Arnes` nil); ahora prueba la forma
   genuinamente sin-ningún-manifiesto (`.claude/` pelado). El comportamiento nuevo (fallback puebla
   `Arnes`) tiene su propio test dedicado (`TestLoaderFallbackPluginJSON`).
2. **go-arch-lint: componente `portafolio` registrado en T4**, no en T8 como preveía S0-D11 — el
   propio gate de T4 lo exigía (el linter rechaza archivos sin componente asignado). T8 completó
   los `deps` finales.
3. **`docs/product/capabilities/loader/reconocer-forma-fisica.yaml` (CAP-15) sin `change_log`
   nuevo** — la nota del spec §9 ("+plugin.json") apuntaba a la detección de forma física, pero el
   fallback real vive en la LECTURA del manifiesto (CAP-17); no se le fabrica un log a código que
   no cambió.
4. **Store del portafolio: una corrupción que rompe la sintaxis JSON del ARCHIVO ENTERO no
   sobrevive a un `save` posterior** (a diferencia de una fila individual corrupta pero
   sintácticamente válida, que SÍ sobrevive — `json.RawMessage` exige JSON válido para
   re-serializar; no hay forma honesta de re-insertar bytes no-JSON en un documento que debe
   serlo). Documentado en el docstring de `saveLocked` + cubierto por
   `TestStoreArchivoTotalmenteIlegible`.
5. **`empresas` legacy vs nueva conviven en el schema** (`anyOf`) — los `arnes.l0.json` reales del
   repo migraron a la forma nueva (T2), pero el schema sigue aceptando el escalar legacy
   indefinidamente (no hay fecha de deprecación fijada — queda en BACKLOG si hace falta una).

## Firma

- [ ] 🧑‍⚖️ **Gate humano** — **PENDIENTE**. El operador revisa: goal §0 del plan (8 puntos) contra
  esta hoja + el E2E vivo de arriba, y las 5 desviaciones registradas. Al firmar, actualizar este
  checkbox + `checkpoint.md` (paquete activo → cerrado) + `BACKLOG.md` (item 0 → construido y
  firmado).
