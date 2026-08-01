# Informe honesto — qué hizo ArnesIA y qué hizo el constructor en la forja de `developer-vitalia`

> vig: 2026-08-01 · pedido del operador: «cómo usaste los prompts para que el mismo
> arnesia corrija todo, por qué no detectaba el arnés inicialmente, y cómo creaste
> el canónico y sincronizaste — ¿o lo hiciste tú mismo sin usar arnesia?»
> Respuesta sin maquillaje: **ArnesIA hizo una parte; otra parte la hice yo a mano
> porque el producto HOY no tiene el camino.** Cada paso manual quedó registrado
> como deuda (D/E en `BACKLOG.md`). Este archivo es la traza completa.

## 0 · TL;DR

| Qué | ¿Quién lo hizo? |
|---|---|
| Agregar vitalia-app al Portafolio (escanear → elegir → agregar) | **ArnesIA** (wizard real de la app) |
| Verificar el inventario del material crudo (42/10/10/45/4) y el corte | **ArnesIA** (sesión del dock, el conductor leyó el disco) |
| Detectar 2 errores de mi plan (no existía `.mcp.json`; el hook `Stop` era del proyecto) | **ArnesIA** (el conductor, conversando) |
| **Escribir** los archivos del plugin (skills, sello, arnes.yaml, contracts) | **YO, a mano** — el gate de permisos bloqueó al conductor (deuda D) |
| Publicar 0.1.0 al marketplace (commit+push+tag) | **YO, a mano** en formato B2 — el write-side no cubre el alta de un plugin nuevo (deuda E) |
| Sincronizar el clone del marketplace que lee la app (`git pull`) | **YO, a mano** — el «Refrescar» de la app lee un clone que CC actualiza, no lo actualiza ella (deuda E-bis, abajo) |
| Crear la entrada CANÓNICA (checkout + `canonico.path` + deriva `al-hilo`) | **ArnesIA** (botón «↧ Traer canónico») |
| Indexarlo y servir el Mapa con `actividades[]` | **ArnesIA** (Observar en Mapa → loader → índice → wire) |

La frase justa: **el criterio y la verificación fueron conversacionales; la
materialización fue mía porque el producto me negó las manos del conductor.**

## 1 · Los prompts exactos al dock (sesión sobre vitalia-app)

**Prompt 1 — el plan de forja completo.** Sesión nueva sobre la entrada
`sin-home~~github-com-alpacapurpura-vitalia-app`, y un solo mensaje con el plan
EXACTO (sin margen de invento): la lista cerrada de los 27 skills DENTRO y los 15
FUERA, los 10 agents, 10 commands, 43 rules, los 4 hooks con su `hooks.json`
derivado de `settings.json`, el contenido literal de `plugin.json`, del sello
`arnes.l0.json` (spine de 5 estados con categorías I-77), del `arnes.yaml` v3
(4 actividades con `pasos` y `caja:`), y los 4 bloques `contract:` de las cajas
(architect · dev-team · auditor · commit-push). Cierre: «listá el árbol y reportá
conteos; NO hagas git commit».

*Resultado:* el conductor **verificó el inventario contra el disco real**
(confirmó que las 27 existen, los conteos exactos), cazó que **`.mcp.json` no
existe** (mi dato venía stale de una exploración anterior — él tenía razón) y que
el hook **`Stop` apunta a `scripts/` del proyecto**, no a `.claude/hooks/` (su
criterio: es maquinaria del proyecto, no del arnés — moverla violaría la frontera
de cuerpos). Después **se bloqueó**: mkdir/Write/cp negados por el gate.

**Prompt 2 — confirmación de criterios + retry mínimo.**
> «Sí a tus dos criterios (sin .mcp.json — mi dato era stale, bien cazado; y el
> hook Stop queda con su ruta ${CLAUDE_PROJECT_DIR}, es maquinaria del proyecto).
> Ahora reintentá SOLO este primer paso y esperá mi aprobación en el panel:
> mkdir -p developer-vitalia/{skills,agents,commands,rules,hooks,.claude-plugin}»

*Resultado:* volvió a fallar y el conductor lo diagnosticó él mismo: «**es un
bloqueo duro del harness, no un permiso que puedas aprobar… nunca llega al
panel**». Propuso: «o corré vos: mkdir -p …».

**Prompt 3 — aislar la causa con un solo Write.** Corrí yo el mkdir (como él
pidió) y le pedí escribir SOLO `plugin.json` y frenar con el error exacto.

*Resultado:* error literal `Claude requested permissions to write to …
plugin.json, but you haven't granted it yet` — y su tabla distinguiendo los DOS
mecanismos: el bloqueo duro de mkdir (validador interno de CC) vs el Write en
«permiso pendiente» que **nunca aparece en el panel**. No escribió nada.

## 2 · Por qué NO detectaba el arnés inicialmente (cadena de causas, en orden)

1. **vitalia-app jamás había sido escaneado.** El índice tenía otra cosa: el
   overlay `vitalia` de `~/Proyectos/luana-vitalia/vitalia` (más una clave legacy
   duplicada, residuo del re-key, que saneamos). vitalia-app solo existía en
   prosa de docs. *Fix:* wizard real «＋ Agregar → Proyecto → escanear».
2. **Escaneado, era material crudo SIN manifiesto** → identidad provisional
   `sin-home` (S1-D27, correcto y honesto): sin `arnes.l0.json` ni `plugin.json`
   no hay identidad que sellar. Por eso el daemon logueó al crear la sesión:
   `«árbol registrado no indexable aún — grafo sin manifiesto (arnes.id)»`.
   No era un bug: **el detector detecta la FORMA física, la identidad la da el
   sello** — y no había sello.
3. **El plugin recién forjado dentro del checkout no era escaneable:** el wizard
   rechaza rutas bajo `~/.arnesia` («ubicación protegida»). Correcto como
   protección, pero deja al alta de un plugin nuevo sin puerta (deuda E).
4. **Publicado al marketplace, la app seguía viendo 2 entradas:** el lector
   LOCAL de catálogo lee `installLocation` del marketplace = **el clone que
   administra Claude Code** (`~/.claude/plugins/marketplaces/prenter-marketplace`),
   que estaba stale respecto de mi push. El «↻ Refrescar» de la app relee ese
   clone, **no hace fetch**. *Fix manual:* `git pull` en ese clone → refrescar →
   apareció `developer-vitalia v0.1.0 · no lo tengo`.
5. **Detalle final:** el primer render del Mapa crasheó por MI dato inválido
   (`arquetipo: guiado` no existe en el enum `pipeline·excepcion·abierto·
   no-arnesar` — vocabulario de un mockup viejo). El ErrorBoundary lo degradó
   visible; corregí el dato (`80923cd`) y renderizó. Que una faceta mala tire el
   lienzo ENTERO quedó como deuda F.

## 3 · Cómo nació el CANÓNICO y quién sincronizó qué

El modelo (RN-IDENT-4): **canónico = la copia que vive dentro de un checkout de
marketplace**; todo lo demás son instalaciones espejo. Por eso no podía
«declararlo» canónico donde estaba — tenía que ENTRAR por el marketplace.

Secuencia real:

1. **(YO)** Materialicé el árbol del plugin siguiendo el plan conversacional
   EXACTO del prompt 1 (copias + sello + arnes.yaml + 4 contracts + hooks.json
   con los criterios que el conductor fijó).
2. **(YO)** Intenté el camino de producto (escanear el dir → canónico): wizard
   bloquea `~/.arnesia` → sin camino (deuda E).
3. **(YO)** Publiqué al repo REAL del marketplace (`~/Proyectos/
   prenter-marketplace`) replicando el formato del write-side B2 a mano:
   `plugins/developer-vitalia/0.1.0/` + fila en `.claude-plugin/marketplace.json`
   + commit + push + tag `developer-vitalia/v0.1.0`. `catalogo.json` quedó
   intacto (es SSoT del kit, esquema mono-plugin). El **pre-commit del propio
   marketplace** (`check_marketplace_shape`) validó verde las dos veces — más el
   fix de los 12 `clerk-*` que eran symlinks a `.agents/` del proyecto (rotos
   fuera de vitalia-app; los desreferencié).
4. **(YO)** `git pull` del clone de CC (la sincronización del punto 2.4).
5. **(ARNESIA)** «↻ Refrescar» → catálogo con 3 entradas → **«↧ Traer canónico»**:
   la app materializó `~/.arnesia/checkouts/prenter-marketplace/developer-vitalia`,
   creó la entrada del Portafolio con identidad calificada
   `(github.com/alpacapurpura/prenter-marketplace, developer-vitalia)`,
   `canonico.path` al checkout, versión 0.1.0 y **deriva `al-hilo`** (el hash
   coincide con el home — la ley anti-drift midiendo de verdad).
6. **(ARNESIA)** «Abrir en Mapa» / Observar → el loader leyó el checkout,
   **derivó `arnes.actividades[]` + facetas AL INDEXAR** (el seam MA-T1b) y el
   wire lo sirvió: historia 4 pasos · bugfix 4 · spike 2 (uno sin caja, E13) ·
   revisar-capability 0 (E7) · `dev-team`×2 · `commit-push`×2.

Nota: antes del paso 3 hubo un intento intermedio (mover el árbol directo al
checkout) que descarté al confirmar que el checkout NO es un git repo y que el
escaneo estaba bloqueado — por eso el camino terminó siendo repo→pull→traer, que
además deja la deriva EVALUABLE (checkout contra home real).

## 4 · Por qué el conductor no pudo escribir (la causa técnica, verificada en código)

`internal/adapters/agent/claudecode/conductor.go` → `permissionArgs()`: si la
sesión tiene `PermissionSet` **de valor cero** (sin rol, sin allow/ask/deny, sin
TTL), **no emite NINGÚN flag** — ni `--permission-mode default` ni
`--permission-prompt-tool stdio`. Una sesión sobre un arnés **sin sello** no
tiene rol del cual derivar permisos → set vacío → el conductor corre headless
**sin canal de permisos** → Claude Code auto-niega toda escritura (y el mkdir lo
corta su validador interno de directorios). Por eso «nunca llega al panel»: no
hay tubería conectada, no es que nadie apruebe.

Consecuencia de producto: **una «sesión de forja» sobre material crudo hoy es
read-only de facto.** El arreglo (deuda D) es diseño, no parche: un rol/perfil de
forja con su PermissionSet, o el canal de permisos SIEMPRE cableado aunque el
arnés no esté sellado.

## 5 · Las deudas que este dogfood dejó anotadas (BACKLOG)

- **D** · sesión sin sello = read-only sin HITL (el permiso jamás llega al panel).
- **E** · el ALTA de un plugin nuevo no tiene camino a canónico (checkout
  protegido + B2 exige canónico previo = círculo). *E-bis implícita:* «Refrescar»
  no hace fetch del clone de CC — sincronizar es manual hoy.
- **F** · una faceta inválida (`arquetipo` fuera del enum) tira el lienzo entero
  en vez de degradar el nodo (§4.5).

Con D y E resueltas, la próxima forja debería poder hacerse **entera** desde el
chat: ese es el criterio de cierre real de esta deuda.
