# Arquitectura de inyección de know-how — cómo ArnesIA POSEE su doctrina y la IMPLEMENTA vía Claude Code

> **Estado:** propuesta arquitectónica detallada — insumo para **HS-07** (specs). 2026-07-05.
> No firmada; baja a detalle para que la siguiente sesión se enfoque en la **doctrina**
> (`knowledge/` — los 122 checks + L2) y no en re-derivar el mecanismo.
>
> **⇑ SUPERADO: FIRMADA (operador, 2026-07-07 — HS-10).** La estrategia **(c) embed+plugin**
> (3 cuerpos: `go:embed` del ruleset para conformance portable + kit materializado en
> `~/.arnesia/` inyectado por flags al spawn, ② jamás se escribe en ③) queda **vigente**.
> Firmada junto con `arch/contracts/nomenclatura-arnes.md` v1 (D-a/D-b/D-c). Los conteos de
> abajo (122/89 checks, 11 elementos) son la foto de 2026-07-05 — hoy 138+97=235 y 12 nodos.
> Implementación = los 3 puentes (loader por nomenclatura · inyección al conductor ·
> conformance embebido + endpoint), ficha siguiente.
> **Norte:** [`VISION.md`](../VISION.md) (11 principios + A1–A7) · [`METODOLOGIA.md`](../METODOLOGIA.md)
> (§0 dueño de crear+mantener · §5 sensor=telemetry-emit del kit · §6 conformación) ·
> [`knowledge/`](../knowledge/INDEX.md) (11 elementos · 122 checks) · [`arch/`](../arch/INDEX.md)
> (14 boundaries · 89 checks).
> **Verificación externa:** investigación 4-subagentes live contra docs oficiales `code.claude.com/docs`
> (CC v2.1.x, 2026-07-05); ver §Fuentes.

## 0. El problema

ArnesIA tiene el know-how de cómo se hace el mejor arnés (los 11 elementos, 122 checks, la
metodología de cajas, el contrato de caja). Ese know-how debe **llegar a Claude Code** para que el
conductor CREE, EVALÚE, MONITOREE y MEJORE arneses conforme a nuestra doctrina — y todo esto bajo
una restricción de UX dura:

> El usuario instala Claude Code (con suscripción activa), instala ArnesIA, y **usa ArnesIA**.
> **Nunca entra a Claude Code a instalar nada.** Todo lo que ArnesIA necesita viaja con su
> instalador y se aplica por detrás; los updates los aplica ArnesIA.

## 1. Los tres cuerpos (la separación que resuelve todo)

Todo el diseño cuelga de no confundir tres cosas:

| Cuerpo | Qué es | Dónde vive | Quién lo lee |
|---|---|---|---|
| **① Doctrina** (el know-how) | 11 elementos · 122 checks · plantillas · metodología · contrato de caja | `knowledge/`+`arch/` en el repo → **compilada y `go:embed`** en el binario | el runner Go **Y** el meta-arnés |
| **② Maquinaria** (meta-arnés) | el KIT `harness@prenter-marketplace` + skills/subagents/system-prompt que ENCARNAN la doctrina para que CC *produzca* conforme | `~/.arnesia/kit/` (materializado del embed al abrir) | el `claude` local, **inyectado al spawn** |
| **③ Arnés-producto** | lo vendible: los `.claude/` reales (`luana-platform`, …) | **su ruta local** (`arnes_registry`) → publicado como **plugin a marketplace git** | CC **nativo por cwd** |

**Regla dura (candidata a boundary):** ② nunca se escribe en el árbol de ③; ③ nunca contiene la
maquinaria de arnesia. Esa línea es lo que hace verdad el «cero config del usuario».

## 2. Dónde vive físicamente cada cosa

```
REPO fábrica (fuente de verdad, versionada)
  knowledge/elements/*.md   → L2 checklists = los 122 checks
  arch/, plantillas/, kit/  → doctrina + meta-arnés
        │  build step (go:generate / `arnesia build-kit`) compila:
        │    · ruleset.json     (los 122 checks COMO DATA)
        │    · kit/             (meta-arnés: skills+subagents+hooks+mcp)
        │    · doctrine.md      (metodología como system-prompt overlay)
        ▼
BINARIO arnesia  ──go:embed──> lleva doctrina+kit DENTRO (viaja con el instalador)
        │  primer `serve` / al abrir la app — POR DETRÁS, idempotente
        ▼
~/.arnesia/                (provisión invisible; el usuario nunca la toca)
  kit/                     ← meta-arnés materializado (plugin LOCAL)
  doctrine.md              ← system-prompt de la metodología
  knowhow/                 ← los 122 checks legibles (referencia read-only del conductor)
  ruleset.json             ← para `arnesia conformance` (runner Go)
  mcp.json                 ← MCP de arnesia (sensor · langfuse-espejo, luego)
  arneses.json·sessions.json·index.sqlite  ← estado (ya existe)
        │
        ├──cwd──> <ruta-del-arnés>/.claude/   ← ③ el arnés-producto REAL (arnesia lo autora, CC lo lee)
        │
        └──publish──> marketplace git (plugin.json + marketplace.json + semver/SHA)
```

## 3. El mecanismo estrella: `--plugin-dir` + flags, SIN `--bare`

Docs oficiales (CC v2.1.x, verificado 2026-07-05): **`--plugin-dir <path>` carga un plugin local
para ESA sesión** — sin `/plugin install`, sin marketplace, sin escribir el árbol del usuario, y
**sin requerir `--bare`** → la suscripción OAuth del `claude` local del usuario queda intacta.

`conductor.go` crece para spawnear (aditivo sobre lo que ya hace):

```
claude -p --input-format stream-json --output-format stream-json --include-partial-messages --verbose
  [--resume ID] [--model] [--max-turns N]          # ← lo actual (HS-06)
  --cwd <ruta-del-arnés>                            # ③ editar el arnés-producto (ya: cmd.Dir)
  --plugin-dir ~/.arnesia/kit                       # ② meta-arnés (skills+subagents+hooks+mcp) SESSION-SCOPED
  --append-system-prompt-file ~/.arnesia/doctrine.md   # ① doctrina como overlay
  --mcp-config ~/.arnesia/mcp.json                  # sensor + langfuse-espejo
  --add-dir ~/.arnesia/knowhow                      # ① los 122 checks como referencia read-only
  --permission-mode <según modo>                    # HS-07 spike control_request
```

En una sola sesión convergen los tres cuerpos: el conductor **ve** el arnés-producto (cwd, para
editarlo) **+ carga** la maquinaria+doctrina de arnesia (por flags, invisible al repo del cliente)
**+ mantiene** la suscripción del usuario (sin `--bare`, sin API key). El usuario no configura nada.

> **Trade-off cementado:** `--bare` y suscripción-propia son **excluyentes** (bare salta OAuth/keychain,
> no lee `CLAUDE_CODE_OAUTH_TOKEN` → exige `ANTHROPIC_API_KEY`). Elegimos **suscripción-propia**: legítimo
> para una app local que usa el CC del propio usuario. La zona prohibida por terms es **hostear login
> claude.ai / poolear rate-limits** de terceros — que NO es nuestro caso local-first. ⇒ nada de `--bare`.

## 4. El know-how tiene DOS representaciones (desde una fuente)

Lo que vuelve real el principio 10 («nada sin eval») y el §6 (conformación):

- **Como DATA** → `ruleset.json` (122 checks compilados) → **`arnesia conformance`**, runner **Go,
  determinista, sin LLM**. Barato, exacto = el **gate estático** (cargar→chequear→reporte de
  brechas→corregir). Corre sobre el `.claude/` de ③.
- **Como PRIMITIVAS CC** → el meta-arnés `~/.arnesia/kit/` (skills tipo *authoring-a-skill-caja*,
  subagents de forja) → el conductor **genera/corrige conforme**. LLM, para crear y mejorar.

Ambas derivan del **mismo** `knowledge/elements/*.md`. El build las mantiene sincronizadas → un solo
lugar donde vive la verdad. (Aterriza el «runner `arnesia conformance` corre ambos [knowledge+arch]»
que CLAUDE.md ya prometía.)

## 5. El ciclo completo → mecanismo exacto

| Verbo | Mecanismo | Cuerpos |
|---|---|---|
| **Crear** | conductor con `--plugin-dir kit` en la ruta del arnés (nueva); grill→spec→build materializa `.claude/` desde plantilla embebida; **nace con `telemetry-emit`** (principio 9) | ①②③ |
| **Evaluar** | `arnesia conformance` (Go, `ruleset.json`) = fitness estático + gate · evals runtime = corridas headless del conductor · **gate en promote** (A4) | ①③ |
| **Monitorear** | ③ **ya emite** OTLP (telemetry-emit del kit) → collector que arnesia embebe → índice JSONL/SQLite (JSONL = verdad) → mapa pinta salud desde los checks | ③ |
| **Mejorar** | hallazgos (conformance+monitor) → puntos en el mapa → turno al conductor (doctrina inyectada) → corrige `.claude/` → re-eval → `arnesia publish` → marketplace git | ①②③ |

## 6. Telemetría de nacimiento / Langfuse (el «meter langfuse»)

No se instrumenta de cero — **se opera la primitiva del kit** (METODOLOGIA §5, VISION principio 9):

1. Al **crear**, arnesia materializa en el `.claude/settings.json` de ③ el hook **`telemetry-emit`**
   del kit (KIT-03, OTLP GenAI-semconv). Todo arnés **nace** conectado; no es opt-in.
2. Egress OTLP → **collector que arnesia embebe/opera** (local-first) → índice SQLite; los **JSONL de
   `~/.claude` son la verdad** (daemon caído = cero pérdida).
3. **Langfuse = espejo opcional**, provisionado por detrás vía `--mcp-config`/exporter que arnesia
   configura — jamás dependencia dura (VISION «Lo que NO es»). Solo cruzan métricas·scores·hashes (I-53).

## 7. Instalación y updates (el constraint, cumplido)

- **Prereq:** `claude` instalado + suscripción activa. Arnesia lo **detecta** al abrir (`claude --version`
  / `/status`); si falta, **guía** — no lo instala por el usuario.
- **Install de arnesia** = un binario (GoReleaser → brew/scoop/deb) dentro del shell Tauri. Al abrir:
  provisiona `~/.arnesia/` desde el `go:embed`. **Cero pasos en CC.**
- **Updates** = update del binario → re-materializa `kit/`+`doctrine.md`+`ruleset.json`+`knowhow/`. La
  doctrina evoluciona semanal (`knowledge/CADENCE.md`) → cada release trae el know-how nuevo. Arneses ya
  publicados no se rompen (versionado semver/SHA en el marketplace).

## 8. Cambios concretos al código (sobre el mapa multisesión HS-06)

Aditivo; respeta los boundaries vigentes.

1. **Nuevo puerto `KitProvisioner`** (`internal/ports/kit.go`):
   `Provision(mode) → Injection{PluginDir, SystemPromptFile, MCPConfig, AddDirs, PermissionMode}`.
   `mode ∈ {build, eval, improve}`. El composition root cablea el provisioner que materializa el embed;
   `SessionService.spawnLocked` le pide la inyección y la pasa en `SpawnOpts`. Mantiene el dominio limpio
   (boundary `adaptadores-de-agente-intercambiables` + `dominio-independiente-de-transporte`).
2. **`SpawnOpts` crece** (`internal/ports/agent.go`): `+PluginDir, +SystemPromptFile, +MCPConfig, +AddDirs
   []string, +PermissionMode`. El conductor los mapea a flags. **Prohibido `--bare`** en la ruta suscripción.
3. **`go:embed` del know-how compilado** + subcomando/paso `arnesia build-kit` (go:generate) que extrae
   `knowledge/` → `ruleset.json` + `kit/` + `doctrine.md`.
4. **`arnesia conformance`** — nuevo subcomando/servicio: runner Go del ruleset sobre un `.claude/` (hoy
   `runIndex` es stub in-memory; este es su hermano determinista).
5. **Materialización de ③** — hoy `arnes_registry` solo **mapea** dirs existentes; sumar
   `scaffold(arnesID, clase)` que escribe el `.claude/` inicial desde plantilla embebida (el arnés-de-crear).
6. **`publish` real** (`main.go:runPublish` stub) — `.claude/` materializado → empaqueta `plugin.json` +
   entrada `marketplace.json` → push git semver/SHA.
7. **Collector OTLP** embebido + materialización del hook `telemetry-emit` al crear.

## 9. Tres boundaries nuevos (draft para HS-07)

| regla | L1 (principio) | L2 (realización) | severidad |
|---|---|---|---|
| `maquinaria-no-contamina-arnes` | la config del agente-fábrica no debe filtrarse al artefacto que edita | ② se inyecta SOLO por flags (`--plugin-dir`/`--append-system-prompt-file`); jamás se escribe en `<ruta-arnés>/.claude/` | error |
| `doctrina-una-fuente-dos-targets` | un ruleset ejecutable y su guía generativa no deben divergir | `ruleset.json` (Go) y `kit/` (CC) se derivan ambos de `knowledge/` en build; drift = check rojo | error |
| `telemetria-de-nacimiento` | todo arnés nace observable (VISION p9) | `scaffold` materializa `telemetry-emit` en el `.claude/settings.json` de todo arnés creado | error |

## 10. Decisiones

- **Auth** → **resuelta**: local-first suscripción-propia, sin `--bare`. ✅
- **Formato ②** → **`--plugin-dir` (plugin local session-scoped)**, no install persistente: da
  skills+subagents+hooks de un golpe, cero acción del usuario, suscripción intacta.
- **Abiertas (HS-07):** (a) doctrina como `--append-system-prompt-file` (transversal) vs skill invocable
  del kit — voto: ambas, prompt para lo always-on + skills para lo invocable. (b) aislamiento fino build
  vs eval (¿`--disallowedTools` para no correr hooks de ③ en corridas de eval? — el split
  estático-Go/runtime-CC lo resuelve casi todo). (c) sandbox / `auto` mode para corridas desatendidas
  (**cwd NO es frontera de seguridad** — Read amplio, Bash llega a todo; hallazgo de esta investigación). (d)
  cómo DevHub/apps-de-rol cargan ③ publicado sin `/plugin install` del trabajador (probable: mismo
  `--plugin-dir`).

## Fuentes (investigación 4-subagentes, live 2026-07-05)

- Empaquetado/distribución: `code.claude.com/docs/en/{plugins,plugin-marketplaces,discover-plugins,settings,claude-directory,headless,cli-reference}.md`.
- Embedding/SDK/auth: `code.claude.com/docs/en/{agent-sdk/overview,headless,cli-reference,permission-modes,sandboxing,authentication}` · `claude.com/blog/building-agents-with-the-claude-agent-sdk` (2025-09-29) · issue anthropics/claude-agent-sdk-python#498 (Go SDK, abierto). **No hay Go Agent SDK oficial → conductor subprocess = vía sancionada.**
- Mercado/competencia: Superpowers/Prime Radiant · ClaudeFast · blencorp/claude-code-kit · Cloudflare `cloudflare/skills` · Anthropic **Managed Agents** (beta abr-2026, descartado por HS-04) · Devin Playbooks · Gemini CLI extensions. Estándares neutros: **SKILL.md** (agentskills.io) · **AGENTS.md** (Linux Foundation; CC no lo lee nativo). Memoria: `hs-research-provisioning-auth-2026`.
