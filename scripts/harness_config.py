#!/usr/bin/env python3
# PROJECT-LAYER COPY del kit harness@prenter-marketplace v0.5.3 (ADOPTING §1 "cp -r").
# NO editar aquí — fixes van UPSTREAM al kit (backflow, doctrina CLAUDE.md). Copiado 2026-07-09.
# voseo-allowed: loader interno de maquinaria (no user-facing)
"""harness_config.py — the harness DIP-seam loader (W5b · 2026-06-09).

Reads the single PROJECT/SISTEMA/TECH store ``project.config.yaml`` (repo root) for the
harness CORE. The CORE names ZERO concrete tech/sistema/locale/engine/live-env; it reads
abstract SLOTS through this one module (charter §0.5 north-star · §3 seam).

ONE module, four consumer classes (RESEARCH-loader-mechanism.md, date-aware verified):
  · class A  bash / git-hooks : ``"$VENV/bin/python" "$WS/scripts/harness_config.py" <dotted.slot> [pluck]``
  · class B  python scripts   : ``sys.path.insert(0, "$WS/scripts"); from harness_config import load, get``
  · class C  cockpit (Next TS): separate ``lib/project-config.ts`` (same store, `yaml` pkg)
  · class D  markdown         : ``{slot}`` CONVENTION (the model Reads the yaml) — no loader call

CLI:
  python harness_config.py <dotted.slot> [pluck_key]   → value to stdout, exit 0
  python harness_config.py --doctor                    → lists every __FILL_ME__ slot

Exit codes: 0 ok · 2 config not found · 3 slot/tree UNFILLED (``__FILL_ME__``) · 4 slot not found.

``--doctor`` is the extraction-test gate (W8): drop core-harness/ + an empty project.config.yaml
into a fresh repo → ``--doctor`` exits 3 and names the slots to fill. Never trust ``yaml.load`` —
``safe_load`` only (a config file must not be able to exec arbitrary code).
"""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

import yaml

CONFIG_NAME = "project.config.yaml"
FILL_SENTINEL = "__FILL_ME__"

# Exit codes (pinned by tests; bash call-sites switch on these).
EXIT_OK = 0
EXIT_NO_CONFIG = 2
EXIT_UNFILLED = 3
EXIT_NOT_FOUND = 4

_cache: dict[str, Any] | None = None


def find_config(start: Path | None = None) -> Path:
    """Parent-walk from this file (or ``start``) to the repo root holding project.config.yaml.

    Walking from ``__file__`` (not cwd) means the loader resolves correctly regardless of the
    caller's working directory — and inside a git worktree it finds THAT worktree's config.
    """
    base = (start or Path(__file__)).resolve()
    for d in (base, *base.parents):
        cfg = d / CONFIG_NAME
        if cfg.is_file():
            return cfg
    raise FileNotFoundError(f"{CONFIG_NAME} not found walking up from {base}")


def load(refresh: bool = False) -> dict[str, Any]:
    """Parse + cache the seam config (``safe_load`` only)."""
    global _cache
    if _cache is None or refresh:
        _cache = yaml.safe_load(find_config().read_text(encoding="utf-8")) or {}
    return _cache


class SlotNotFound(KeyError):
    """Dotted slot path does not resolve in the config tree."""


# Back-compat seam aliases (I-52): the kit now speaks `sistemas`; seams authored before the
# rename declare `brands`. The alias is BIDIRECTIONAL so the rename can land incrementally —
# a new kit asking `sistemas.*` resolves an old `brands:` seam, AND an un-migrated script still
# asking `brands.*` resolves a migrated `sistemas:` seam. Either side moves first, never breaks
# (the cockpit reader does the same, trying `sistemas` then `brands`).
_SEAM_ALIASES = {"sistemas": "brands", "brands": "sistemas"}


def _navigate(cfg: Any, dotted: str) -> Any:
    """Walk a dotted path; a digit segment indexes a list, else a dict key.

    An absent dict segment falls back to its legacy alias (``_SEAM_ALIASES``) so
    ``sistemas.*`` resolves against a pre-rename ``brands:`` seam.
    """
    node = cfg
    for seg in dotted.split("."):
        if isinstance(node, list) and seg.lstrip("-").isdigit():
            idx = int(seg)
            if not -len(node) <= idx < len(node):
                raise SlotNotFound(dotted)
            node = node[idx]
        elif isinstance(node, dict):
            if seg in node:
                node = node[seg]
            elif seg in _SEAM_ALIASES and _SEAM_ALIASES[seg] in node:
                node = node[_SEAM_ALIASES[seg]]
            else:
                raise SlotNotFound(dotted)
        else:
            raise SlotNotFound(dotted)
    return node


def get(
    dotted: str,
    pluck: str | None = None,
    where: tuple[str, str] | None = None,
    cfg: dict[str, Any] | None = None,
) -> Any:
    """Resolve a dotted slot.

    ``where=(field, value)`` filters a list-of-dicts to the items whose ``field`` equals
    ``value`` (string-compared) BEFORE plucking — e.g. ``get("sistemas.active", pluck="slug",
    where=("cap_gate", "hard"))`` → the slugs whose cap-gate is hard. ``pluck`` then extracts
    one key from each surviving dict.

    Raises ``SlotNotFound`` if the path (or ``pluck``/``where`` shape) does not resolve.
    """
    value = _navigate(cfg if cfg is not None else load(), dotted)
    if where is not None:
        if not isinstance(value, list):
            raise SlotNotFound(f"{dotted} (where={where[0]}: not a list)")
        field, wanted = where
        value = [el for el in value if isinstance(el, dict) and str(el.get(field)) == wanted]
    if pluck is not None:
        if not isinstance(value, list):
            raise SlotNotFound(f"{dotted} (pluck={pluck}: not a list)")
        return [el[pluck] for el in value if isinstance(el, dict) and pluck in el]
    return value


def find_unfilled(cfg: Any = None, prefix: str = "") -> list[str]:
    """Deep-walk the tree; return the dotted path of every ``__FILL_ME__`` leaf."""
    node = cfg if cfg is not None else load()
    out: list[str] = []
    if isinstance(node, dict):
        for k, v in node.items():
            out += find_unfilled(v, f"{prefix}{k}.")
    elif isinstance(node, list):
        for i, v in enumerate(node):
            out += find_unfilled(v, f"{prefix}{i}.")
    elif node == FILL_SENTINEL:
        out.append(prefix.rstrip("."))
    return out


def _format(value: Any) -> str:
    """Render a resolved value for stdout: scalars plain, scalar-lists newline-joined,
    complex (dict / list-of-dict) as compact JSON (a python/jq consumer can parse)."""
    if isinstance(value, bool):
        return "true" if value else "false"
    if value is None:
        return ""
    if isinstance(value, (str, int, float)):
        return str(value)
    if isinstance(value, list) and all(isinstance(el, (str, int, float)) for el in value):
        return "\n".join(_format(el) for el in value)
    return json.dumps(value, ensure_ascii=False)


def _doctor() -> int:
    try:
        cfg = load()
    except FileNotFoundError as e:
        print(str(e), file=sys.stderr)
        return EXIT_NO_CONFIG
    unfilled = find_unfilled(cfg)
    product = (cfg.get("meta") or {}).get("product", "?")
    print(f"harness-doctor · project.config.yaml · product={product}")
    if not unfilled:
        print("  ✓ all slots filled — ready (idea→done runs without editing the CORE)")
        return EXIT_OK
    print(f"  ✗ {len(unfilled)} slot(s) UNFILLED ({FILL_SENTINEL}) — declare these:")
    for slot in unfilled:
        print(f"      {slot}")
    return EXIT_UNFILLED


def main(argv: list[str]) -> int:
    if not argv:
        print(
            "usage: harness_config.py <dotted.slot> [pluck_key] [--where field=value] | --doctor",
            file=sys.stderr,
        )
        return EXIT_NOT_FOUND
    if argv[0] in ("--doctor", "-d"):
        return _doctor()
    # Extract --where field=value (filter a list-of-dicts before pluck); rest = positionals.
    where: tuple[str, str] | None = None
    pos: list[str] = []
    i = 0
    while i < len(argv):
        if argv[i] == "--where" and i + 1 < len(argv) and "=" in argv[i + 1]:
            field, _, val = argv[i + 1].partition("=")
            where = (field, val)
            i += 2
            continue
        pos.append(argv[i])
        i += 1
    dotted = pos[0]
    pluck = pos[1] if len(pos) > 1 else None
    try:
        cfg = load()
    except FileNotFoundError as e:
        print(str(e), file=sys.stderr)
        return EXIT_NO_CONFIG
    try:
        value = get(dotted, pluck=pluck, where=where, cfg=cfg)
    except SlotNotFound:
        print(f"slot not found: {dotted}" + (f" (pluck={pluck})" if pluck else ""), file=sys.stderr)
        return EXIT_NOT_FOUND
    if value == FILL_SENTINEL:
        print(f"slot unfilled ({FILL_SENTINEL}): {dotted}", file=sys.stderr)
        print(FILL_SENTINEL)  # also stdout so a bash `$(...)` capture is non-empty + greppable
        return EXIT_UNFILLED
    print(_format(value))
    return EXIT_OK


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
