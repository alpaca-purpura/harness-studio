# SPEC — Cockpit de delivery (v1) — CONGELADA

> **Congelada 2026-07-03** (épica Torre de Control, F5 · ficha `DH-08`). La implementación
> ejecuta ESTO sin retocar diseño; cambio de diseño = nueva ficha DH-NN + edición explícita
> aquí. Continúa la numeración de la épica (RN-01..RN-34 en `torre-read-only.md`,
> `torre-arquitectura.md`, `proceso-descriptor.md`). Decisiones D11-D14: AskUserQuestion
> respondida (recomendadas confirmadas), registradas en
> `../epicas/torre-de-control/ESTADO.md` y `DH-08`.

## 0 · Qué es

La capa 3 del norte v3 conectada a las otras dos: el **descriptor de proceso** (capa 1, F4)
PARAMETRIZA una **sesión de agente** (capa 3) alimentada por **contexto as-code** (capa 2).
El developer no redacta prompts de cero: la pantalla precarga el prompt desde el descriptor
(estado → tramo siguiente → gate/checklist → arnés dueño) + el contexto as-code; el humano
ENCAUSA (edita/aprueba) y lanza. Diferencial contra Conductor: orquestamos ENTREGA, no
sesiones — la sesión nace del proceso, no de la cabeza del dev.

**Contrato intra-célula**: el sidecar es superficie interna de P2 (Go ↔ Node por loopback).
No cruza células → NO pide ficha I-NN (regla del dominio products/).

## 1 · Decisiones D11-D14

- **D11 · Sidecar supervisado por el binario Go**: el binario spawnea el sidecar Node al
  boot y lo mata al salir; contrato HTTP loopback. Un solo comando sirve TODO = embrión del
  instalable (DH-03). Sin Node/sidecar → degradación honesta (RN-40), el resto del binario
  vive.
- **D12 · Primera sesión parametrizada = ESPECIFICACIÓN**: story en estados tempranos del
  ciclo; la sesión produce la spec (texto, barato, minutos). BUILD (`developing`, worktree,
  gates caros) = fase posterior.
- **D13 · Pantalla = vista nueva `/delivery`** (nav item propio, patrón `/proceso`). Board
  intacto: las 8 defs duplicadas de la UI NO migran en F5 (D10 sigue); lanzar desde el
  story-drawer = F6.
- **D14 · Contexto as-code v1 = story + arquitectura + visión**: `checkpoint.md` de la
  story + `arquitectura.yaml` de la célula (gated — F3/D7) + `VISION.md` de la célula.
  Capabilities (principio 4, sobre Backstage) NO entra v1 — el catálogo-como-entidad no
  tiene sobre definido; forzarlo = inventar formato sin ficha.

## 2 · Sidecar (Node · Agent SDK TS)

- **RN-35** · **Proceso y contrato**: `products/devhub/sidecar/` (TypeScript,
  `@anthropic-ai/claude-agent-sdk`, cero framework — `node:http`). El binario Go lo lanza al
  boot (`node <sidecar>/dist/index.js`), escucha en `127.0.0.1:0` (puerto efímero) y hace
  handshake imprimiendo `SIDECAR_PORT=<n>` en stdout. Muere con el binario (Pdeathsig en
  unix; best-effort en windows). v1 NO reinicia un sidecar caído: proxy responde honesto
  (RN-40). Descubrimiento del sidecar: flag `-sidecar <dir>` > env `DEVHUB_SIDECAR` >
  `{exeDir}/sidecar` > `{exeDir}/../sidecar` (layout dev: `go/` y `sidecar/` hermanos).
  **Branding propio** (frontera #6): el proceso se llama `devhub-sidecar`; la UI dice
  "sesión de agente"/"powered by Claude" — jamás se llama ni se parece a "Claude Code".
- **RN-36** · **Auth**: contrato de producto = `ANTHROPIC_API_KEY` en el env del sidecar
  (o Bedrock/Vertex vía sus env estándar). PROHIBIDO shipear pass-through de login
  claude.ai (ToS 2026). En el control plane del operador (dogfood), sin API key el runtime
  del SDK resuelve la credencial local del CLI — mismo precedente que P4/I-76
  (`claude -p` desde el binario `studio`); esto es aceptable SOLO operador-local, jamás en
  un instalable a cliente. `GET /salud` reporta `auth: api-key | credencial-local |
  ninguna`; lanzar con `ninguna` → 503, la UI lo muestra sin eufemismo.
- **RN-37** · **API del sidecar** (loopback, JSON): `GET /salud` →
  `{ok, auth, sesiones}` · `POST /sesiones` `{prompt, contexto: [{nombre, ruta_abs}],
  meta: {sistema, story, tramo}}` → `201 {id}` · `GET /sesiones` → lista resumida ·
  `GET /sesiones/{id}` → `{id, meta, estado: creada|corriendo|terminada|error, eventos[],
  workspace, resultado?}`. Eventos = proyección chica de los mensajes del SDK:
  `{ts, tipo: init|texto|herramienta|resultado, detalle}` — nunca el mensaje crudo entero.
- **RN-38** · **Workspace aislado por sesión** (doctrina I-76): el sidecar crea
  `~/.prenter/devhub/sesiones/<id>/` con `PROMPT.md` (el prompt lanzado, verbatim),
  `contexto/` (COPIAS de los archivos as-code — la sesión jamás lee el repo),
  `salida/` (el entregable) y `eventos.jsonl` (persistencia del stream). La sesión corre
  con `cwd` = ese workspace, `permissionMode: acceptEdits`,
  `allowedTools: [Read, Write, Edit, Glob, Grep]`, sin Bash/Web, `maxTurns` acotado (25),
  sin `settingSources` (no carga CLAUDE.md del operador). **RN-31 intacto**: la sesión NO
  transiciona stories ni escribe el repo — el ciclo sigue siendo de los skills; el humano
  revisa `salida/` y aplica. Resultado en disco = la evidencia.

## 3 · Plantilla parametrizada (el corazón — capa 1 → capa 3)

- **RN-39** · `GET /api/delivery/plantilla?sistema&story[&contexto=<id>]` (Go-nativo, motor
  del descriptor): deriva el **tramo** de la story SOLO del dato — cero literales de estado:
  - `siguienteTramoRol(S)`: primera transición `ejecutor: rol` alcanzable desde S recorriendo
    `transiciones[]` en su orden (profundidad ≤ 2 — cubre el estado inicial, que solo tiene
    salidas de operador). Estado en categoría terminal o `pausado` → `422 no-lanzable`.
  - El prompt precargado nombra: descriptor (id·nombre·versión) · story (id·estado·categoría·
    descripcción del estado) · tramo (estado destino + su descripción) · dueño del destino
    (bindings `duenos[]`: rol + arnés + nota) · gates cuyo `momento` toca el tramo (estado
    destino o transición `de→a`) con su checklist verbatim · el mandato de trabajar SOLO en
    el workspace de la sesión y entregar en `salida/`.
  - **Contexto as-code por convención** (D14): candidatos = `{sistemaRoot}/arquitectura.yaml`
    y `{sistemaRoot}/*/arquitectura.yaml` (el MISMO glob que el gate D7), cada uno con su
    `VISION.md` hermano si existe. La respuesta lista `contextos_disponibles[]`; `?contexto=`
    elige (default: el primero). 0 hallazgos → contexto honesto vacío (la plantilla lo dice).
  - Respuesta: `{prompt, tramo, gates[], contexto: {id, fuentes: [{nombre, ruta_abs}]},
    contextos_disponibles[]}`. El `checkpoint.md` de la story SIEMPRE es fuente primera.
- El prompt es EDITABLE en la pantalla antes de lanzar — la plantilla propone, el humano
  encausa (tesis del norte). Lo lanzado queda verbatim en `PROMPT.md` (RN-38).

## 4 · Superficie Go + vista

- **RN-40** · **Proxy honesto**: `/api/delivery/*` (salud·sesiones) reenvía al sidecar.
  Sidecar caído/ausente → `503 {disponible: false, motivo}` — nunca datos inventados ni
  spinner eterno; la UI banner "delivery no disponible" y el resto de la consola vive.
  `plantilla` es Go-nativo y funciona AUN sin sidecar (se puede previsualizar el prompt).
- **RN-41** · **Vista `/delivery`** (nav item propio): selector de story (de
  `GET /api/stories`) → panel de plantilla (tramo + dueño + gate + fuentes as-code +
  selector de contexto) → prompt editable → lanzar → feed de eventos de la sesión + estado.
  Cero literal de estado/categoría en el componente: todo viene de `plantilla`/`proceso`.
- **RN-42** · **Eventos por polling** (fork técnico menor, decidido aquí): la UI hace poll
  de `GET /api/delivery/sesiones/{id}` cada ~1.5s mientras `corriendo`. Streaming
  SSE/token-a-token = extensión (no cambia el contrato: los eventos ya son dato).
- **RN-43** · **Embrión del instalable** (DH-03/entregable 5): `cockpit start -port N -d`
  sirve UI + API + sidecar arrancado — UN comando. El instalable real (bundle de Node,
  firma, updater) = fuera de alcance; el corte de v1 es que nada más que ese comando haga
  falta en una máquina con Node.

## 5 · Fuera de alcance (v1)

- **RN-44** · No entra: sesiones BUILD (worktree + gates) · N sesiones paralelas como
  producto (el sidecar no lo impide; la UI muestra una) · aprobaciones interactivas
  (`canUseTool` render en UI — la sesión v1 corre no-interactiva con tools acotados) ·
  lanzar desde el board/story-drawer (F6, con la migración de las 8 defs a `/api/proceso`) ·
  transicionar estados desde delivery (RN-31) · telemetría/spans (frontera #1: si la sesión
  emite, la analiza P4, no esta consola) · `resume`/`fork` de sesiones · descriptor por
  workspace (`?sistema=` en `/api/proceso`) · editor del descriptor.
