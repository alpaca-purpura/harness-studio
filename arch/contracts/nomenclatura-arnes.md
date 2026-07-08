# Nomenclatura de reconocimiento — directorio de arnés ⟷ grafo L0

> **Estado: FIRMADA v1.1 (operador, 2026-07-07 — HS-12 interop DevStudio; aditiva sobre
> v1 = HS-10; D-a/D-b/D-c resueltas abajo).** Este
> contrato define cómo ArnesIA RECONOCE un arnés leyendo sus archivos y lo convierte en un
> `graph.l0` — el algoritmo que el loader real (Hito 3, reemplaza los fixtures `go:embed`)
> implementará y que hasta la auditoría 2026-07-07 no estaba escrito en ningún documento
> (los grafos dogfood están armados A MANO; `fuente_path` era «puntero, no contrato»).
>
> Norte: [`../../VISION.md`](../../VISION.md) (A1–A7) · [`../../METODOLOGIA.md`](../../METODOLOGIA.md)
> §3 (contrato fusionado) · [`schema/graph.l0.schema.json`](./schema/graph.l0.schema.json) ·
> los L1 de [`../../knowledge/elements/`](../../knowledge/INDEX.md) (ubicación oficial CC de
> cada elemento — esta tabla los ANCLA, no los duplica).

## 1. La unidad reconocible

Un arnés se reconoce en DOS formas físicas (ambas ciudadanas de primera):

| Forma | Qué es | Cuándo |
|---|---|---|
| **Plugin CC** | repo con `.claude-plugin/plugin.json` (layout oficial de plugin) | el arnés como se distribuye por marketplace git — la forma canónica de «Cargar» para modificar |
| **Arnés instalado** | proyecto con `.claude/` poblado (skills/agents/commands/settings) | el arnés desplegado en un proyecto de cliente — auditoría/observación in situ |
| **Roster de app de rol (lock)** | proyecto con `.devstudio/arneses.yaml` (roster: arnés@versión@canal + registry de origen) | proyecto de cliente gestionado por DevStudio: **N arneses complementarios**, cada uno en forma-plugin resuelta vía caché local (`~/.dev-studio/arneses/`) o rehidratación del marketplace — v1.1, HS-12 |

El detector: si hay `.claude-plugin/plugin.json` → plugin; si hay `.claude/` → instalado;
si hay ambos, plugin manda. Ninguno → no es arnés (error honesto, jamás grafo vacío).

**Multi-arnés (v1.1, HS-12):** el lock `.devstudio/arneses.yaml` NO es un arnés — es un
**puntero de descubrimiento read-only** que enumera N arneses. ArnesIA lo lee y carga cada
entrada como **forma-plugin** (la unidad reconocible no cambia); entrada no resoluble en
caché/marketplace → check rojo visible por-arnés (reconciliación honesta §4.5, jamás se
omite en silencio). El lock lo escribe y posee DevStudio (sus campos estables:
`id`·`versión`·`canal`·`registry` — pedido recíproco fichado en HS-12); coexiste con un
`.claude/` propio del proyecto sin pisarlo (superficies independientes).

## 2. El manifiesto: `arnes.l0.json`

En la raíz del arnés vive **`arnes.l0.json`**: la parte del grafo que NO es derivable de
archivos — `id`, `nombre`, `fases[]`, `spine` (estados + transiciones), `meta` de enganche
(rol · proceso · reporta_a · empresa), `marketplace`. Es el subconjunto arnés-level de
`graph.l0.schema.json`; se valida contra él (gate G1).

- Lo escribe ArnesIA al crear el arnés; lo mantiene ArnesIA al modificar. Un arnés sin
  manifiesto se carga en modo **degradado honesto**: nodos sin carriles de fase ni spine,
  con check `manifiesto-ausente` rojo — visible, nunca inventado.
- **`nombre`/`descripcion` (v1.1, HS-12):** `nombre` es el campo CANÓNICO para pintar el
  arnés en una vista (Roles, picker, Organigrama). Cadena de fallback bendecida:
  `arnes.l0.nombre` → `plugin.json` `name` → `id` (ídem `descripcion` → `plugin.json`
  `description`). Opcionales — el schema los admite desde HS-12 (reparación: §2 los
  nombraba pero `graph.l0.schema.json` los rechazaba por `additionalProperties:false`).
- **`spine.categorias` (v1.1, HS-12):** mapa opcional estado→categoría semántica FIJA
  (enum de 5 idéntico al contrato de ecosistema I-77 RN-28: propuesto · en-progreso ·
  completado · descartado · pausado). Los estados siguen siendo dato per-arnés; la
  categoría es la capa semántica del producto. Checks warn: `categoria-estado-existe` ·
  `terminal-categoria-coherente` (terminalidad derivada: categoría ∈
  {completado, descartado}).
- El `.graph.json` completo (como los dogfood) queda como **formato de export/intercambio**;
  la fuente de verdad en un arnés real es `arnes.l0.json` + los archivos.

## 3. Mapeo clase → ubicación (los 10 reconocedores)

La celda canónica de cada clase vive en el L1 de su nodo `knowledge/elements/<clase>.md`;
esta tabla fija QUÉ escanea el loader y qué nodo emite:

| `clase` | En plugin | En arnés instalado | Nodo emitido |
|---|---|---|---|
| `skill` | `skills/<id>/SKILL.md` | `.claude/skills/<id>/SKILL.md` | caja (si `contract.caja`) o soporte |
| `subagent` | `agents/<id>.md` | `.claude/agents/<id>.md` | soporte de su caja |
| `command` | `commands/<id>.md` | `.claude/commands/<id>.md` | soporte |
| `hook` | `hooks/hooks.json` (una entrada = un nodo) | `.claude/settings.json#hooks` | banda **Guardia** |
| `mcp` | `.mcp.json` (un server = un nodo) | `.mcp.json` | soporte |
| `rule` | `CLAUDE.md` + reglas declaradas | `CLAUDE.md` | banda **Base** |
| `settings` | `settings.json` del plugin | `.claude/settings.json` | soporte |
| `output-style` | `output-styles/<id>.md` | `.claude/output-styles/<id>.md` | soporte |
| `statusline` | entrada en settings | entrada en settings | soporte |
| `plugin` | el contenedor mismo (`.claude-plugin/plugin.json`) | entrada de plugin en config | nodo raíz del arnés |

## 4. Reglas de derivación (archivo → grafo)

1. **Nodos**: scan según la tabla §3. El loader ESTAMPA `fuente_path` (deja de ser manual)
   y `clase` según el reconocedor que disparó.
2. **Cajas**: el frontmatter fusionado (METODOLOGIA §3) de cada `SKILL.md` es el
   `contract:` del nodo — `caja`, `fase`, `estado` (la transición del spine), 3 ejes.
   Frontmatter inválido → nodo visible + check G1 rojo (no se descarta el nodo).
3. **Edges**: derivados, jamás declarados sueltos — `necesita.de: caja:<id>` → edge
   `de→a`; `hooks` matchers → edges de Guardia; `entrega` compartida → hand-off.
4. **Bandas**: `hook`→Guardia · `rule`→Base · cajas→su carril de fase · resto→soporte.
5. **Reconciliación honesta** (regla estrella, espeja «gris ≠ verde»):
   - archivo presente que ningún reconocedor entiende → **nodo `no-reconocido` VISIBLE**
     (check warn) — nunca invisible, nunca crash (obliga fallback en FE);
   - declarado en manifiesto sin archivo en disco → check G1 **rojo**;
   - elemento fuera del enum de 10 → nodo visible con clase `no-reconocido`.

## 5. Consecuencias inmediatas (al firmar)

- **Migrar `dogfood/skills/spec-writer.SKILL.md`** → `dogfood/skills/spec-writer/SKILL.md`
  (hoy el arnés insignia viola el layout que skills.md L1 predica).
- FE: fallback `no-reconocido` en `arnes-node.tsx` + ErrorBoundary (hoy: `TypeError` con
  clase fuera de enum; banda desconocida = nodo invisible — ambos violan §4.5).
- `graph.l0.schema.json`: `fuente_path` pasa de «puntero, no contrato» a «estampado por el
  loader según nomenclatura-arnes.md».

## Decisiones FIRMADAS (operador, 2026-07-07)

| # | Decisión | Resolución firmada |
|---|---|---|
| D-a | Nombre/lugar del manifiesto | **`arnes.l0.json` en la raíz del arnés** (visible, versionable, no escondido en `.claude-plugin/`) |
| D-b | ¿Arnés instalado (.claude/ sin plugin.json) es ciudadano de primera? | **SÍ** — es la forma en que se audita in situ; sin él no hay «instalar en proyecto existente» |
| D-c | Elemento no reconocido | **nodo `no-reconocido` visible con warn** (honestidad > limpieza) |

Firmada junto con la **estrategia de empaquetado (c)** del diseño 3-cuerpos
(`research/2026-07-05-arquitectura-inyeccion-knowhow.md`): `go:embed` del ruleset (conformance
portable, cero contexto LLM) + doctrina como plugin CC propio inyectado por flags al spawn
(progressive disclosure). La implementación de ambos = los puentes (ficha siguiente).

## Changelog

- 2026-07-07 · v1.1 — **FIRMADA** (HS-12, interop DevStudio): detector 3° «roster de app de
  rol» (`.devstudio/arneses.yaml` = puntero de descubrimiento multi-arnés, read-only) ·
  `nombre`/`descripcion` reparados en el schema + cadena de fallback canónica ·
  `spine.categorias` opcional (enum fijo 5 = I-77 RN-28) con 2 checks warn. Todo aditivo.
- 2026-07-07 · v1 — **FIRMADA** (HS-10): D-a/D-b/D-c resueltas según recomendación; se firma en el
  mismo acto la estrategia de empaquetado (c) embed+plugin.
- 2026-07-07 · v0 — draft inicial (HS-10), sale de la auditoría 4-frentes: el hueco «nomenclatura
  no escrita» era el mayor hallazgo doctrinal.
