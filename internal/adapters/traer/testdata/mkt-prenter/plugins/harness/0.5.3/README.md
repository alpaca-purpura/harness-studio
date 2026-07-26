# core-harness/ — the extractable dev-process kit (SSoT)

> **Status: POPULATED (W7 move executed + W8 extraction-test PASSING, 2026-06-09).** The proxy-clean `tier:core` files were moved here per `docs/process/harness-refactor-w7/W7-EXEC-OUTPUT.md` (21 rules, 6 process-docs, 5 templates, 6 git-scripts + `harness_config.py`, 3 pre-commit checks, 1 agent, the bootstrap skill + plugin shape). It is the **extractable kit**: the tech/sistema-agnostic software-development PROCESS (lifecycle, gates, roles, doctrine, safety, learning-capture, DoD) that ports to ANY product. The dependency-grep over this dir = **0 project tokens** (only grep-bot's generic build-artifact skip-dirs). W8 proved an empty-config drop + a fictional product (Go/single-tenant) runs `--doctor` → fill → resolve without editing any core file.

## What lives here (target)

```
core-harness/
├── rules/        ← proxy-clean always-on process invariants (21 core rules + rules-detail)
├── agents/       ← generic worker contracts (grep-bot today; role-skeletons = W8 lift-kit)
├── skills/       ← harness-bootstrap (+ generic role skeletons = W8)
├── hooks/        ← session-start-rules injector + (later) the core enforcement skeleton
├── templates/    ← 6 core spec/arch/validators/tickets skeletons
├── process/      ← 7 core P1 doctrine docs + the kit's own charter/PROCESS-MODEL
├── scripts/      ← harness_config.py (the seam loader) + ~6 core git coordination scripts
└── .claude-plugin/  ← marketplace.json + plugin.json (the distribution shape)
```

**NOT here:** anything tech/sistema-specific (your stack's web framework / UI lib / auth / linter, your engine package paths, your sistema/market names). Those are the PROJECT layer (the adopter's `.claude/`, `scripts/`, `docs/`) and a new product writes its own. The split is by the `tier:` tag recorded in the 6 W-OUTPUT manifests. The **seam** `project.config.yaml` (repo root) is what the core reads instead of naming a tech/sistema.

## How it re-exposes to Claude Code (the ratified mechanism — Option C, Chris 2026-06-09)

Claude Code only auto-loads rules from `.claude/rules/`, discovers agents from `.claude/agents/`, skills from `.claude/skills/` (gate-zero research: `docs/process/harness-refactor-w7/GATE-ZERO-{RESEARCH,ADVERSARIAL}.md`). So the kit's real files live HERE, and re-expose per surface by its **doc-blessed** channel:

| Surface | Mechanism | Why |
|---|---|---|
| **rules** | symlink `.claude/rules/X.md → core-harness/rules/X.md` | symlink is DOCUMENTED-SAFE for `.claude/rules/` (memory doc); recursive, circular-safe |
| **agents** | copy into `.claude/agents/` (symlink only if the empirical session-restart test passes) | agents-symlink is UNDOCUMENTED (#14836) |
| **skills** | plugin `skills/` (namespaced) OR copy into `.claude/skills/` | skills-symlink undocumented + buggy |
| **hooks** | `settings.json` / plugin `hooks.json` points straight at `core-harness/hooks/*` | hooks are path-referenced |
| **templates / process-docs / scripts** | symlink-back at the old path (read-by-path → OS resolves) | path-stable → zero consumer repoint |

**Path-stability is the point:** every `.claude/…` / `docs/…` / `scripts/…` consumer path stays valid (symlink resolves) → the machinery validator's 29 hardcoded `.claude/` paths + all script/template refs are unaffected.

## Extraction / bootstrap procedure (the program DoD)

A new product:
1. Drops `core-harness/` into a fresh repo + a `project.config.yaml` with all slots `__FILL_ME__`.
2. Runs `python core-harness/scripts/harness_config.py --doctor` → it lists every unfilled slot (toolchain, sistemas, locale, engine_prefix, live_verify_infra, design_system_ref, domain_modules, …).
3. Fills the slots with their product's info (vision/tech/design) + runs `/harness:bootstrap` (the skill below) → it symlinks/copies the core rules into `.claude/rules/`, wires the hooks, and re-runs the doctor.
4. The `idea→done` cycle runs **without editing the CORE**. (Multisistema is just `sistemas.active` with one entry = single-sistema.)

## Distribution (cross-project / cross-account)

`core-harness/` doubles as a **Claude Code plugin marketplace** (`.claude-plugin/marketplace.json`): one "harness" plugin shipping skills+agents+hooks (incl. the embedded telemetry, KIT-03), **versioned explicitly**: `plugin.json.version == KIT_VERSION` (`VERSION`), factory-gated (`check_plugin_shape.py`) — an update reaches installs only on a deliberate semver bump (KIT-04/D-5; reproducible reinstalls beat push-equals-update). The always-on rules can't ship via a plugin (no plugin rules channel) → the plugin's **SessionStart hook injects the slim rule core (≤10k)** + `/harness:bootstrap` copies the full corpus. Host the repo private (git creds, NOT the Claude account, gate access; share to a 2nd account via collaborator/org) or public. It lives in its own repo (`prenter-harness`), consumed by adopters via versioned install (`installer/new-project.sh`, pinned tag) — not git subtree (retired).

## Pointers

- `docs/process/harness-refactor-charter-2026-06-08.md` — the constitution (3-layer model, seam, DoD).
- `docs/process/harness-refactor-w7/W7-OUTPUT.md` — the gate-zero verdict + move manifest + §9 execution playbook.
- `docs/process/harness-refactor-w7/PLUGIN-DISTRIBUTION-RESEARCH.md` — the distribution model citations.
- `project.config.yaml` (repo root) — the seam the core reads (filled by the adopter).
