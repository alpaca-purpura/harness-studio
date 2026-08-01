# Auditoría previa al mockup — verificación independiente (2026-08-01)

> Segunda pasada sobre [`informe-estado-real.md`](./informe-estado-real.md), hecha SIN dar por
> cierto nada del informe: cada cita `archivo:línea` se abrió, y el loader real se corrió contra
> los dos árboles (canónico + instalación). Resultado: **el informe es fiel en lo estructural**
> (los 15 gaps de su tabla §2 se confirman uno por uno), con **5 correcciones de dato** y **una
> medición que no estaba y que cambia el diseño del mockup**.

## 1. La medición que faltaba

Corrido: `arnesia index <dir>` sobre los dos árboles. Cifras GENERADAS, no tecleadas.

| | canónico publicado<br>`~/.arnesia/checkouts/prenter-marketplace/developer-vitalia` | instalación cruda<br>`~/Proyectos/vitalia-app` |
|---|---|---|
| nodos totales | **83** | **103** |
| por clase | 27 skill · 43 rule · 10 command · 3 hook | 30 skill · 45 rule · 10 command · 4 hook · 1 settings · **13 no-reconocido** |
| **cajas** (banda=fase) | **5** | **1** |
| fases declaradas | 4 | **0** (l0 degradado) |
| actividades | 4 | 4 (del `arnes.yaml` copiado a mano) |
| **nodos en Base / sin banda** | **75 / 83 = 90 %** | **98 / 103 = 95 %** |
| aristas | 0 | 1 |
| agents emitidos | 0 (de 0) | **0 (de 10 en disco)** |

### 1.1 · Qué separa realmente «proyecto crudo» de «arnés»

El canónico y la instalación son **el mismo material** (el canónico se forjó copiando de
vitalia-app). Toda la diferencia de estructura son **5 bloques de frontmatter** (`contract.caja:
true` + `fase:` en 5 skills) más **2 archivos de raíz** (`arnes.l0.json` completo + `arnes.yaml`).
Nada más. Eso es lo que el operador puso a mano en el carril D («la materialización fue mía»).

**Consecuencia de diseño:** la ingesta NO es primariamente un motor de clasificación de 169
archivos. Es **decidir qué N piezas son pasos del proceso y estamparles caja+fase**, más escribir
el `arnes.yaml`. El inventario total (ING-D3) y la clasificación por facetas (ING-D4) son el
contexto que hace posible esa decisión — no el entregable.

### 1.2 · El muro de la banda Base (problema de UI no cubierto por el baseline)

90-95 % de los nodos caen en **una sola banda**. `selectSoporte` los pliega todos a
`FALLBACK_BAND = "base"` (`web/src/entities/arnes/model/selectors.ts:60-67`) — correcto por
honestidad (§4.5, invisible es mentira), pero significa que el Mapa de un proyecto real es
**1 caja arriba y 98 nodos apilados abajo**.

El baseline no cubre esta escala: el fixture más grande de las stories es `luanaFeatureCycle`
≈ 38 nodos; `devFullCycle` = 11. **El mockup tiene que resolver 100+ nodos en Base**, o el Mapa
editor nace ilegible sobre el caso que ING-D1 declara como el 90 % del uso.

### 1.3 · Inconsistencia en el arnés YA publicado

Los `pasos[].caja` del canónico apuntan a nodos que en 3 casos **no son cajas**:

| actividad | pasos → estado del target |
|---|---|
| `historia` | architect ✅ · dev-team ✅ · auditor ✅ · commit-push ✅ |
| `bugfix` | **chrome-devtools-verify → nodo-sin-fase** · dev-team ✅ · **test-all → nodo-sin-fase** · commit-push ✅ |
| `spike` | **explore-module → nodo-sin-fase** |
| `revisar-capability` | **(cero pasos)** |

Ninguno es «fantasma» (todos existen como nodo), pero tampoco son cajas: renderizan en Base, no
en la secuencia. **Este estado no está en MA-E1..E13**: `E13` es «paso SIN caja» (ghost); esto es
«paso CON caja que apunta a algo que no es caja». Gap de spec, no de datos.

## 2. Correcciones al `informe-estado-real.md`

| # | El informe dice | Verificado |
|---|---|---|
| C1 | «11 agents» | **10** agents en `.claude/agents/`. El gap (0 nodos emitidos) se confirma igual |
| C2 | «51 rules» | **45** en `.claude/rules/` (y el loader emite 45 nodos `rule`) |
| C3 | «12 skills Clerk materializadas en `.agents/skills/` **FUERA de** `.claude/`» | Están en **AMBOS** lados. Los 12 `clerk-*` viven DENTRO de `.claude/skills/` — por eso el loader los emite como `no-reconocido` con `fuente_path` = `.claude/skills/clerk*`. La copia de `.agents/` es adicional |
| C4 | «30 skills» | **42 dirs** en `.claude/skills/` = 30 reconocidos + 12 clerk no-reconocidos |
| C5 | «13 no-reconocidos» (implícito: los 12 clerk) | El 13° es **`claude-md` → el `CLAUDE.md` de la raíz** (ver §3) |

Todo lo demás de la tabla §2 del informe se abrió y **se confirma**: wizard callejón sin salida
(`portafolio-wizard.tsx:308-327`) · Mapa read-only del grafo (`router.go:86-97`, cero POST/PUT
sobre nodos/aristas) · «Editar fuente» disabled hardcodeado (`inspector.tsx:744-751`) ·
reconocedor de agents + contenedor plugin como TODO declarado (`loader.go:20-24`) · gate D19 sin
lector (cero ocurrencias de `gate` en `internal/adapters/forja/`) · clasificación cero código
(no existe tipo `faceta`/`knowledge` en `internal/domain/`) · `arnes.yaml` no llega a
instalaciones · publisher copia el árbol completo (`publisher.go:310-352`).

## 3. Hallazgo nuevo — el reconocedor de reglas exige la convención de ArnesIA

`reconocerRegla` (`internal/adapters/loader/regla.go:24-46`) sólo entiende el `CLAUDE.md` raíz si
trae frontmatter con `id:` **o** un heading `# <id> — <nombre>` (em dash con espacios). El de
vitalia-app abre con `# CLAUDE.md` ⇒ **`no-reconocido`**.

Un proyecto crudo, por definición, no cumple una convención de ArnesIA que nunca conoció. En el
escenario B esto no es un caso borde: es **el caso**. El nodo más central del proyecto (su
CLAUDE.md) entra al Mapa como pieza incomprendida.

## 4. Hallazgo nuevo — la ruta de escritura YA existe (y no es el Mapa)

El informe marca «manos del conductor ✅ resuelta v0.7.0» en una fila, y «Mapa read-only ❌» en
otra, sin cruzarlas. Cruzadas dicen algo que decide arquitectura:

- El conductor cablea **siempre** `--permission-mode default --permission-prompt-tool stdio`
  (`conductor.go:141-160`, DD-1) y `Write/Edit/MultiEdit/NotebookEdit` **jamás** viajan en
  `--allowedTools` (`conductor.go:113-117`): cada escritura pasa por la tarjeta del Dock.
- O sea: **hoy existe exactamente una ruta viva de escritura de archivos as-code — el chat, con
  aprobación humana por diff.** Los 11 `os.WriteFile` de Go escriben sellos, manifiestos,
  scaffolds y provisión; ninguno edita el as-code de un arnés a pedido del usuario.

Esto abre la pregunta central de arquitectura del mockup (Q1 abajo): un click en el Mapa ¿escribe
por Go, o despacha un turno del conductor?

## 5. Verificación de higiene

- **7 ramas/worktrees vivos** (`audit-`, `fix-`, `tramo2/3/5-conversaciones`, `worktree-agent-*`,
  `hooktmp`): **0 commits fuera de `main`** en todas. Nada de trabajo apilado sin mergear.
- `git status` limpio en `main` @`4854b31`.

---

## Preguntas abiertas para el operador (previas al mockup)

Numeradas para responder por número. Q1-Q3 bloquean el mockup; Q4-Q8 lo moldean.

**Q1 · ¿Quién escribe cuando hago click en el Mapa?**
(a) Go escribe el archivo directo (endpoints nuevos `POST /nodes`, `PATCH /nodes/{id}`), diff de
confirmación construido por el FE; (b) el click arma un prompt y lo despacha al conductor, que
escribe y pasa por la tarjeta de permiso que YA existe; (c) híbrido: gestos deterministas
(estampar `caja:true`+fase, cablear un edge) por Go; gestos que exigen redactar prosa (crear una
skill nueva desde cero) por conductor.
→ (a) es predecible y testeable pero duplica la superficie de escritura y el diff; (b) reusa todo
lo construido pero hace que un click tarde ~10 s y pueda fallar por criterio del modelo; (c) es
la que recomiendo, y el corte «determinista vs redactado» es exactamente el que ING-D2 dibuja
(«lo conversacional AYUDA pero no reemplaza»).

**Q2 · ¿El mockup dibuja los 100 nodos, o la ingesta los reduce antes de llegar al Mapa?**
El muro de §1.2 se puede atacar por dos lados: **(a) en el Mapa** (Base colapsada por subbandas
con contador, expandir bajo demanda, buscador) o **(b) en la ingesta** (la curaduría marca la
mayoría como `de-referencia`/`no-propio` y sólo lo curado entra al lienzo). Son compatibles, pero
la que elijas primero define si el mockup arranca por la superficie de ingesta o por el Mapa.

**Q3 · La escala del mockup: ¿una superficie o dos?**
El INDEX pide dos (ingesta + Mapa editor) en el mismo gate. Son ~10-14 pantallas juntas. ¿Un
mockup con gate único, o dos gates secuenciales (ingesta primero, que es la que desbloquea
material real para el Mapa editor)?

**Q4 · Procedencia (ING-D3) necesita nombre propio.**
`mockups/INDEX.md` regla dura 4 lo advierte: `procedencia` YA significa honestidad-del-dato
(medido/estimado/declarado/inferido/no-declarado) y `origen` YA significa estandar/del-puesto. El
eje nuevo (provisto-por-plugin · creado-por-usuario · suelto · drifteado · de-referencia)
**choca de frente**. Propongo **`autoría`** (o `procedencia-de-archivo`). ¿Cuál preferís?

**Q5 · Las 5 facetas de ING-D4 (P2 sin firmar): ¿eje nuevo o proyección del terreno?**
`knowledge · knowledge-as-code · docs-as-code · process-as-code · WIP-home` vs las 11 canónicas
D19 + regla de las 3 caras D18. Mi lectura: las facetas son **la cara** (D18) y las canónicas son
**el tema** — o sea son ejes ortogonales y `WIP-home` es el territorio de instancia de D19, no
una faceta par de las otras 4. Si eso es correcto, el vocabulario final es 4 facetas + el puntero
a WIP-home aparte. ¿Confirmás la lectura o querés las 5 planas?

**Q6 · El `CLAUDE.md` que no se reconoce (§3): ¿se afloja el reconocedor o se repara en la ingesta?**
(a) aflojar `regla.go` para aceptar cualquier `CLAUDE.md` (id = `claude-md`, nombre = primer
heading); (b) dejarlo estricto y que la ingesta OFREZCA estampar el heading canónico. (a) es 5
líneas y arregla a todo proyecto crudo del mundo; (b) mantiene la convención pero exige escribir
en disco ajeno en el primer segundo de la ingesta. Recomiendo (a).

**Q7 · Los 3 `pasos[].caja` que no son cajas (§1.3): ¿se repara el dato o se dibuja el estado?**
El arnés `developer-vitalia` YA está publicado con eso. ¿El mockup dibuja el estado degradado
(«paso apunta a una pieza que no es caja — [convertir en caja]», que además sería un gesto
perfecto del Mapa editor), o primero se saneó el dato y el mockup asume consistencia?

**Q8 · T0 (`arnes.yaml` → instalaciones): ¿sale ya, desacoplado?**
El INDEX lo declara desacoplable. Mi recomendación coincide con la de `decisiones.md` §P1 pero
con un matiz: **(c) fallback de lectura** resuelve el síntoma sin tocar disco ajeno y es ~20
líneas en el loader; **(a) escribir al sellar** sólo aplica a instalaciones que se identifiquen
DESPUÉS del fix — no cura las existentes. Si el objetivo es «que nunca más se vean chips
vacíos», (c) es la que hay que hacer primero y sola. ¿Sale como bugfix propio antes del mockup?

## Propuestas (no firmadas — insumo para el gate)

**P-A · Cortar el mockup en dos gates, ingesta primero.**
La ingesta produce el material que el Mapa editor edita. Diseñar el editor antes de saber qué
sale del inventario es diseñar sobre supuestos. Además la ingesta tiene un caso de prueba real y
medido (vitalia-app, §1) y el editor todavía no.

**P-B · La ingesta como «tabla de decisión», no como wizard de 6 pasos.**
El inventario ya es automático (el loader emite los 103 nodos hoy). Lo que falta es humano y es
UNO: por cada pieza, ¿propio / de-referencia / suelto? y ¿es paso del proceso? Una tabla densa
con multi-select, agrupada por autoría (Q4) y con acciones en lote, cubre 103 filas mejor que un
flujo por pasos. El wizard actual se conserva como puerta (escenario A intacto) y gana una tercera
rama: «no encontré arnés — ¿inventariar igual?» (ING-D5).

**P-C · El primer gesto del Mapa editor = «esto es un paso del proceso».**
Es el gesto de mayor palanca medido en §1.1: convierte un nodo de Base en caja estampando
`caja:true` + `fase:` en el frontmatter. Es determinista (Go, sin conductor), reversible,
diffeable, y es literalmente la diferencia entre el crudo y el arnés. Si el mockup sólo dibujara
un gesto de escritura, tiene que ser ese.

**P-D · Base colapsable por autoría antes que por clase.**
En un proyecto crudo, «45 rules» no ayuda a decidir; «38 provistos por el plugin · 5 propios · 12
de terceros» sí — lo propio es lo que se empaqueta. Subbandas plegadas con contador, expandibles,
y el buscador de la map-bar como escape.

**P-E · Escenario del mockup = vitalia-app real, cifras medidas.**
Los números de §1 son reales y reproducibles con un comando. Mismo criterio que el mockup MA-T6
(que ya usó inventario real de vitalia-app y por eso funcionó como gate).
