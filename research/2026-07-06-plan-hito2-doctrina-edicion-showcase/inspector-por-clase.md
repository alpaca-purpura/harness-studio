# Inspector por clase — qué ve el usuario al click, por tipo de elemento

> Ficha HS-09 (fase 5) · plan Hito 2 · addendum Fase 2 · 2026-07-07
> Origen: pedido del operador («al click en un elemento, ¿lo que veo queda así?») — el backlog
> no tenía paso explícito por clase; este doc lo crea. Complementa
> [`ui-doctrina-visible.md`](./ui-doctrina-visible.md) (que es contrato-céntrico: cajas).

## Problema

El inspector (`web/src/widgets/map-canvas/ui/inspector.tsx`) es un layout ÚNICO para las 10
clases con dos caminos: caja → contrato fusionado completo; no-caja → Clasificación + el texto
genérico «Sin contrato — este nodo no es una caja de proceso». Para regla/hook/mcp/settings/…
eso es pobre Y engañoso (no tener contrato es su estado LEGAL, no un faltante). Gate 2 exige
«cada casuística del showcase legible» — el genérico no lo da para no-cajas.

Hallazgo colateral (D-c firmada, violada en FE): el loader emite `no-reconocido` como nodo
visible, pero el `Clase` FE (`entities/arnes/model/types.ts`) NO lo incluye → `KIND[clase]`
devuelve `undefined` y el canvas ENTERO cae al ErrorBoundary. D-c pide nodo visible con warn,
no mapa muerto.

## Realidad del dato (3 niveles, honesto)

1. **Nodo L0 hoy** (`graph.l0.schema.json $defs.nodo`): `id · clase · nombre · banda · fase ·
   canal · fuente_path · contract · procedencia · origen`. CERO campos per-class.
2. **Loader v1** (`internal/adapters/loader`): reconoce solo `skill` + `rule`. Los otros 8
   reconocedores de nomenclatura §3 son TODO honesto (se añaden con el primer arnés real que
   los use).
3. **Showcase** (`showcase.graph.json`): trae las 10 clases como dato pero sin campos
   específicos (hooks sin evento, mcp sin transporte, subagents sin tools).

## Spec — tabla clase → qué DEBE ver el usuario en el inspector

Celda canónica = `arch/contracts/nomenclatura-arnes.md` §3. «Hoy» = renderizable con el dato
existente; lo demás es Tier B/C.

| clase | handle | campos específicos (meta) | celda canónica (fuente) | hoy |
|---|---|---|---|---|
| `skill` (caja) | `/id` | contrato fusionado completo (ya está) | `skills/<id>/SKILL.md` frontmatter | ✅ |
| `skill` (apoyo) | `/id` | description (disparo) · allowed-tools | ídem | rol + fuente |
| `subagent` | `@id` | description · tools · model | `agents/<id>.md` frontmatter | rol + fuente (+contract si caja) |
| `command` | `id` | description · argument-hint · allowed-tools | `commands/<id>.md` frontmatter | rol + fuente |
| `hook` | `evento` | **evento** (PreToolUse/…) · matcher · comando · timeout | `hooks/hooks.json` (una entrada = un nodo) | rol + fuente |
| `rule` | always-on/condicional | **activación** (siempre vs `paths:` condicional) · secciones/tamaño | `CLAUDE.md` | activación (PROPUESTA `alwFor`) + fuente |
| `mcp` | `id` | **transporte** (stdio/http) · command/url · tools expuestas (sin env secretos) | `.mcp.json` (un server = un nodo) | rol + fuente |
| `plugin` | `id` | versión · marketplace de origen · conteo de componentes por clase | `.claude-plugin/plugin.json` | rol + fuente |
| `settings` | `id` | claves relevantes: permisos allow/deny (conteos) · env keys · scope | `settings.json` | rol + fuente |
| `output-style` | `id` | description · cuándo aplica | `output-styles/<id>.md` frontmatter | rol + fuente |
| `statusline` | `id` | tipo/comando | entrada en settings | rol + fuente |
| `no-reconocido` | — | warn + fuente_path + qué esperaba el reconocedor | — | ✅ (tras fix FE) |

Común a TODAS las clases (sección **Fuente**): `fuente_path` (mono) · `canal` · `origen`
(estandar/del-puesto; PROPUESTA hasta que el provisioner estampe).

## Tiers de ejecución

### Tier A — dato existente (SIN decisión nueva; entra a Fase 2 antes de Gate 2)

1. `no-reconocido` al `Clase` FE + entrada `KIND` (warn, char `?`) — cumple D-c en FE; nodo
   visible, mapa vivo.
2. Inspector: sección **Fuente** (fuente_path · canal · origen) para todas las clases.
3. Inspector: **encuadre por clase** en vez del genérico «Sin contrato…» — una línea de rol
   doctrinal por clase (hook→Guardia, rule→Base, mcp→server externo, etc.) + los campos
   per-class aún no extraídos listados como «pendiente del reconocedor» (honestidad: se dice
   qué falta, no se inventa).
4. Regla: sección **Activación** (siempre-en-contexto / condicional / desconocida) desde
   `alwFor` — rotulada PROPUESTA hasta que el loader derive de `paths:`.
5. Fix: el header del inspector pasa `alw` a `handleFor` (hoy toda regla muestra `always-on`
   aunque sea condicional — el nodo del mapa SÍ lo pasa).
6. Story = test por casuística nueva (regla condicional · hook · mcp · no-reconocido).

### Tier B — extracción per-class (DECISIÓN A FIRMAR + loader)

- **Decisión de schema L0 (operador):** cómo viaja el dato per-class. Propuesta: objeto
  opcional `meta` por nodo (bolsa tipada por clase, p.ej. `meta.evento`/`meta.matcher` para
  hook, `meta.transporte` para mcp) — additive, `additionalProperties:false` por clase.
  Alternativa: campos planos opcionales. NO se implementa sin firma.
- Reconocedores pendientes de nomenclatura §3 (hook · mcp · command · subagent · settings ·
  output-style · statusline · plugin) extraen la tabla de arriba. Regla doctrinal intacta:
  se construyen con el primer arnés real que los use — el showcase como grafo servido NO
  los destraba (es dato, no directorio); el dogfood plugin-form SÍ cuando gane esos elementos.
- `alwFor`/`isDelPuesto` pasan de fixture a dato (ya registrado en Fase 2: «cablear
  origen/split-reglas desde el dato»).

### Tier C — uso real (telemetría, Hito 3)

Desempeño/Tokens por nodo (JSONL → indexer). Fuera de este addendum.

## Verificación

- Tier A entra al click-through de **Gate 2**: click en UNA casuística de cada clase del
  showcase → encuadre legible + Fuente presente + regla condicional distinguible + nodo
  no-reconocido visible sin tumbar el canvas. Stories verdes + gates FE verdes.
- Tier B tiene su propia mini-firma (schema) y se verifica con round-trip
  dir→grafo→inspector sobre el dogfood cuando gane hooks/mcp reales.
