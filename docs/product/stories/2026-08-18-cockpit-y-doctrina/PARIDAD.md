# PARIDAD — cockpit y doctrina

> Paquete `2026-08-18-cockpit-y-doctrina`. Cada requisito ⇒ código real + su verificación;
> cada verificación ⇒ corrida de verdad, con su salida. **Fecha: 2026-08-19.**
> **Firma 🧑‍⚖️ del gate: PENDIENTE.**

## Fila por fila

| Qué se pidió | Código | Verificación corrida | Resultado |
|---|---|---|---|
| El cockpit lee las capabilities de este repo | `tools/cockpit/go/handlers_caps.go` | `curl :4300/api/capabilities?sistema=platform` | ✅ **157** capabilities · 19 módulos · 0 sin `capability_id` · 0 sin nombre legible |
| Nombre legible en vez del slug | `readCapability` coalesce `user_facing_name ← name` | mismo curl | ✅ `abrir-shell` → «Abrir shell (`open`)» |
| Los status propios sobreviven | — (el reader pasa el string) | mismo curl | ✅ `{vivo: 108, vivo·nc: 31, parcial: 17, stub: 1}` — sin `live` fabricado |
| El board muestra los paquetes | `scripts/backfill_checkpoints.py` | `curl :4300/api/stories` | ✅ **41** stories · 0 `parse_error` · 6 estados poblados |
| El board ESCRIBE | `writableSistemas()` + `escribibles[]` + `BoardView` | `POST /api/transition` idea→refining sobre `2026-07-24-ctx-derivado-etiquetado`, y revertido | ✅ `{"newState":"refining","ok":true,"verbo":"spec.started"}` · `state:` cambió en disco y volvió a `idea` |
| El roadmap muestra releases | `docs/product/releases/*.yaml` + `--sync-releases` | navegación real a `:4300/roadmap` | ✅ `v0.8-cockpit-y-doctrina` (planning, 5 stories con sus estados) + `v0.7-mvp-instalable` en Historial |
| Solo se muestran tabs con datos | `cockpit.config.yaml` | `curl :4300/api/nav-config` + captura | ✅ `[board, roadmap, learnings, harness, proceso]` — las 5 con fuente real (ver filas del CIL abajo); `map`/`drift`/`evolucion`/`delivery`/`torre` siguen fuera y con su motivo escrito |
| El carril L1 del CIL existe y se cuenta | `docs/process/harness-backlog.md` | `curl :4300/api/harness` + navegación real a `/harness` | ✅ **10 ítems** · `{reported 5, triaged 2, ratified 1, applied 1, deferred 1}` · la vista los pinta en columnas por estado |
| El carril L3 existe y se cuenta | `docs/process/tech-debt.md` | `curl :4300/api/cil` | ✅ **9 ítems**, 8 abiertos · indexan `BACKLOG.md § Deuda viva` sin duplicar la explicación |
| El carril L2 existe y se lee | `docs/learnings/` + `tooling/` | `curl :4300/api/learnings` + navegación real a `/learnings` | ✅ **3 aprendizajes** (2 transversales + 1 tooling), frontmatter parseado, `README.md` correctamente excluido de la lista |
| Los 4 carriles se ven juntos | — | vista `/harness` | ✅ L1 9 abiertos · L2 4 · L3 8 abiertos · L4 → Drift. **Discrepancia detectada y registrada:** el contador L2 dice 4 y la lista muestra 3 (el CIL cuenta el README que la lista excluye) → `HB-10` |
| Los diffs del vendored quedan fijados | `tools/cockpit/go/vendored_arnesia_test.go` | `go test ./...` en `tools/cockpit/go` | ✅ ok · 10 tests nuevos, todos ejercitados (`-v` confirma RUN+PASS) |
| La UI no rompe | — | `npx tsc --noEmit` · `npx vitest run` | ✅ tsc limpio · **172 passed**, 2 skipped, 19 files |
| Existe skill para crear un arnés | `kit/skills/forjar-arnes/SKILL.md` | `go test ./internal/adapters/provision/...` | ✅ ok — la skill se materializa a `~/.arnesia/kit/skills/forjar-arnes/SKILL.md` en el spawn |
| El knowhow se abre ANTES de escribir | `forjar-caja` paso 1 · `auditar-arnes` paso 2 · `doctrine.md` § regla dura | lectura directa de los 3 archivos | ✅ el `Read knowhow/*` pasó del paso 8 (verificar) al paso 1 (antes de escribir) y exige DECLARAR qué se abrió |
| El presupuesto del binario aguanta | `embed_doctrina.go` (`all:kit`) | `git archive HEAD` + mis cambios del kit → `go build ./cmd/arnesia` | ✅ **23,62 MB** · techo 25 MB (94,5 %) · delta +0,58 sobre la base 23,04 (tope +1,50) |
| Instalador Windows de la versión nueva | `scripts/bump.sh 0.8.0` + `scripts/installer.ps1` | los dos comandos, corridos | ✅ `instaladores/v0.8.0/` — `ArnesIA_0.8.0_x64-setup.exe` (8,63 MB, NSIS) + `ArnesIA_0.8.0_x64_en-US.msi` (11,1 MB, WiX) + `checksums.txt` |
| El daemon sellado arranca y se identifica | `bundle.py` (`-ldflags -X`) | `arnesia.exe serve -addr 127.0.0.1:4299` → `GET /api/version` y `/healthz` | ✅ `{"version":"0.8.0.2608190026"}` · healthz **200** |
| El instalador coincide con la fuente | `scripts/bump.sh 0.8.1` + `installer.ps1` | rebuild tras las correcciones del eval + grep de los marcadores en el binario | ✅ `instaladores/v0.8.1/` (setup.exe 8,62 MB + msi 11,1 MB + checksums) · daemon `0.8.1.2608201514` · el binario contiene el esqueleto del manifiesto, `additionalProperties: false`, «lista de STRINGS» y `hooks/hooks.json` |

> **Por qué hubo un 0.8.1 el mismo día:** v0.8.0 se empaquetó ANTES de que el eval
> encontrara los dos defectos de `forjar-arnes`. Como `kit/` viaja dentro del binario
> (`go:embed all:kit`), ese `.msi` forja manifiestos inválidos. Una generación publicada
> nunca se pisa —el propio `installer.ps1` lo impide—, así que la corrección es una versión
> nueva. `instaladores/v0.8.0/` queda como está, y con este motivo escrito.

## Lo que NO se verificó (gris honesto)

- **El gate 🧑‍⚖️ de esta hoja.** Sin firmar: requiere que el operador ejerza el cockpit y
  una sesión de `forjar-arnes` de verdad.
- ~~**`forjar-arnes` end-to-end.**~~ **CERRADO 2026-08-19** con el eval de
  `evals/forjar-arnes/` — ver la sección § Eval de abajo.
- **`TestPresupuestoDeBinario` no completó** en esta máquina: agota su timeout de 600 s
  compilando en un dir sin caché. La medición de arriba replica su método a mano
  (`git archive HEAD` + build) y da el mismo número que el repo ya registraba. El test en sí
  sigue siendo el juez, y corre en CI (Linux).
- **La suite Go completa en Windows.** Deuda preexistente inventariada en
  `stories/2026-08-13-compilacion-windows/decisiones.md § CW-D5`; este paquete no la toca.

## Eval de `forjar-arnes` — el gap cerrado con evidencia

`python evals/forjar-arnes/run.py` · **4/4 en verde** (2026-08-19).

```
── crear-desde-cero ──────────────────────────────────────────────
   ✓ invoca `forjar-arnes`            sí
   ✓ abre nodos de knowhow/           harness-profile.md, hooks.md, rules.md, skills.md
   ✓ declara qué nodos abrió          nombró: harness-profile, rules, hooks, skills
   ✓ arnes.l0.json válido             válido · 3 fases · 5 estados
   ✓ spine sin estados inventados     5 estados, sin inventados

── deriva-a-forjar-caja ──────────────────────────────────────────
   ✓ invoca `forjar-caja`             sí
   ✓ NO invoca `forjar-arnes`         invocó: arnesia-kit:forjar-caja

── no-arnesar ────────────────────────────────────────────────────
   ✓ invoca `forjar-arnes`            sí
   ✓ menciona «no-arnesar»            sí

── baseline-sin-kit  (sin kit — baseline) ────────────────────────
   · skills invocadas       ninguna
   · nodos del estándar     ninguno (no se inyectó)
   · arnes.l0.json          sin arnes.l0.json · escribió: README.md, manifest.yaml,
                            proceso/estados.yaml, rol/soporte-tecnico-n1.md
```

**El delta, en una línea:** con el kit la sesión abre el estándar, lo declara y deja un
`arnes.l0.json` que valida. Sin el kit inventa una estructura entera —`manifest.yaml`,
`proceso/estados.yaml`— que ningún reconocedor de ArnesIA puede leer. La nomenclatura no se
deduce: se enseña.

### Lo que el eval ENCONTRÓ (y se arregló)

La primera corrida dio 3 fallos. Ninguno era lo que parecía, y ninguno se resolvió aflojando
un assert:

| Fallo | Qué era en realidad | Arreglo |
|---|---|---|
| `crear-desde-cero` sin manifiesto | La skill hizo lo correcto: abrió los nodos, propuso el spine y **se frenó a pedir `reporta_a`**, campo requerido que la doctrina le prohíbe inventar. El prompt se lo ocultaba | El caso entrega todo cerrado — una caja T2 no se mide con un turno único (`HB-11`) |
| 9 errores de schema en el manifiesto | **La skill no decía la FORMA**: `fases` como objetos, `marketplace` como objeto, un `fase` extra en cada transición | Esqueleto JSON exacto en el paso 3, validado contra el schema (`HB-12`) |
| La forja se frenó pidiendo permiso | La skill no decía que un arnés nace en **forma-plugin**; el modelo fue a `.claude/settings.json`, ruta protegida | Layout explícito + «no escribas en `.claude/`» (`HB-13`) |
| «Ya leí los tres nodos requeridos» | Decir que leíste no es declarar qué leíste | La skill exige NOMBRAR cada archivo; el assert cuenta nombres, no la palabra «knowhow» |
| `baseline-sin-kit` no arrancaba | **Bug del runner**: `--add-dir` es variádico y se comía el prompt cuando no había `--plugin-dir` detrás | El prompt va por stdin |
| El baseline invocaba una skill inexistente | **Baseline sucio**: quitaba `--plugin-dir` pero dejaba `doctrine.md`, que la nombra | Sin kit = sesión pelada (los 3 flags fuera), y pasa a `modo: medicion` |

Dos de esos —el esqueleto del manifiesto y la forma-plugin— son mejoras de la skill que
**ninguna lectura del código habría encontrado**: hacía falta correrla.

## Capabilities

Ninguna nueva: el gate R2 escanea `cmd`, `internal`, `web/src` y `web/src-tauri/src`
(`docs/architecture/fitness/capability_trace_test.go:367`). Lo que cambió vive en `tools/`,
`kit/`, `scripts/` y `docs/` — fuera de ese alcance. El único archivo tocado dentro es
`internal/adapters/provision/provisioner_test.go`, y es un test.
