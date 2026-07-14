#!/usr/bin/env python3
# cap_doctor.py — doctor local de capabilities-as-code (homologación 2026-07-09).
#
# Espeja el patrón de harness_config.py --doctor: valida cada
# docs/product/capabilities/{module}/{slug}.yaml contra el schema homologado
# (docs/product/_templates/capability.template.yaml) y reporta lo que falta.
#
# Es el doctor RÁPIDO local (pre-commit / feedback). El enforcer DURO de R1/R2
# (punteros que resuelven a archivo real · cobertura sin huérfanos) vive en Go:
# docs/architecture/fitness/capability_trace_test.go. cap_doctor valida FORMA +
# existencia de archivo del puntero + enum de status. OJO (honestidad, auditoría
# 2026-07-14): NADIE resuelve la parte `#Símbolo` hoy — ni este doctor ni go test
# (ambos stat-ean solo el archivo). R1 a nivel símbolo = deuda BACKLOG.
#
# Uso:
#   python3 scripts/cap_doctor.py            # valida todo el árbol
#   python3 scripts/cap_doctor.py --module loader
# Exit: 0 ok · 2 no hay árbol de capabilities · 3 hay errores de validación.

from __future__ import annotations

import sys
from pathlib import Path

import yaml

CAP_DIR = Path("docs/product/capabilities")
STATUS_ENUM = {"vivo", "vivo·nc", "parcial", "stub"}
NATURE_ENUM = {"feature", "scaffold", "extension-point"}
REQUIRED = ("capability_id", "cap_num", "slug", "name", "module", "status", "pointers")

EXIT_OK, EXIT_NO_DIR, EXIT_ERRORS = 0, 2, 3


def _repo_root(start: Path | None = None) -> Path:
    base = (start or Path(__file__)).resolve()
    for d in (base, *base.parents):
        if (d / "go.mod").is_file() or (d / "project.config.yaml").is_file():
            return d
    return Path.cwd()


def _load(path: Path) -> tuple[dict | None, str | None]:
    try:
        text = path.read_text(encoding="utf-8")
    except OSError as e:  # noqa: BLE001
        return None, f"no se pudo leer: {e}"
    # el YAML es el frontmatter (--- ... ---) seguido de markdown opcional
    if text.lstrip().startswith("---"):
        body = text.lstrip()[3:]
        end = body.find("\n---")
        text = body[:end] if end != -1 else body
    try:
        data = yaml.safe_load(text)
    except yaml.YAMLError as e:  # noqa: BLE001
        return None, f"YAML inválido: {e}"
    if not isinstance(data, dict):
        return None, "el frontmatter no es un mapping"
    return data, None


def _check_cap(root: Path, path: Path, data: dict) -> list[str]:
    errs: list[str] = []
    for field in REQUIRED:
        if field not in data or data[field] in (None, "", []):
            errs.append(f"falta campo obligatorio `{field}`")
    status = data.get("status")
    if status is not None and status not in STATUS_ENUM:
        errs.append(f"status `{status}` fuera de enum {sorted(STATUS_ENUM)}")
    nature = data.get("nature")
    if nature is not None and nature not in NATURE_ENUM:
        errs.append(f"nature `{nature}` fuera de enum {sorted(NATURE_ENUM)}")
    # slug del archivo == slug del YAML
    if data.get("slug") and path.stem != data["slug"]:
        errs.append(f"slug `{data['slug']}` != nombre de archivo `{path.stem}`")
    # módulo del path == módulo del YAML
    if data.get("module") and path.parent.name != data["module"]:
        errs.append(f"module `{data['module']}` != carpeta `{path.parent.name}`")
    # punteros: archivo (parte antes de #) existe (R1 — nadie resuelve el símbolo aún, deuda)
    for ptr in data.get("pointers") or []:
        rel = str(ptr).split("#", 1)[0].strip()
        if rel and not (root / rel).exists():
            errs.append(f"puntero a archivo inexistente: `{rel}` (de `{ptr}`)")
    # status con evidencia: vivo exige al menos un test en `valida`
    if status == "vivo" and not (data.get("valida") or []):
        errs.append("status `vivo` sin `valida` (test) — ¿debería ser `vivo·nc`?")
    return errs


GROUP_ORDER = [
    "cli-daemon", "dominio-l0", "loader", "indice-persistencia", "conformance",
    "conductor", "provisioning", "handoff", "http-sse", "usecases", "self-update",
    "portafolio", "fe-mapa", "fe-chat", "fe-shell", "fe-portafolio", "tauri",
]
IDX_BEGIN, IDX_END = "<!--caps:begin-->", "<!--caps:end-->"


def _index() -> int:
    """Regenera la tabla navegable de INDEX.md entre marcadores DESDE los YAML (no teclear)."""
    root = _repo_root()
    cap_root = root / CAP_DIR
    idx = cap_root / "INDEX.md"
    by_mod: dict[str, list[dict]] = {}
    for cap in sorted(cap_root.glob("*/*.yaml")):
        data, err = _load(cap)
        if err or not data:
            continue
        by_mod.setdefault(cap.parent.name, []).append(data)
    lines = [f"_Generado por `cap_doctor.py --index` desde las hojas `*.yaml` (SSoT). No editar a mano._\n"]
    total = 0
    mods = [m for m in GROUP_ORDER if m in by_mod] + [m for m in sorted(by_mod) if m not in GROUP_ORDER]
    for mod in mods:
        caps = sorted(by_mod[mod], key=lambda d: d.get("cap_num", ""))
        lines.append(f"\n### `{mod}` ({len(caps)})\n")
        for d in caps:
            total += 1
            ptr = (d.get("pointers") or ["—"])[0]
            lines.append(f"- **{d.get('cap_num')} · {d.get('name')}** `{d.get('status')}` · "
                         f"{mod}/{d.get('slug')}.yaml — `{ptr}`")
    table = "\n".join(lines) + "\n"
    if idx.exists():
        t = idx.read_text(encoding="utf-8")
        if IDX_BEGIN in t and IDX_END in t:
            pre = t.split(IDX_BEGIN)[0]
            post = t.split(IDX_END)[1]
            idx.write_text(f"{pre}{IDX_BEGIN}\n{table}{IDX_END}{post}", encoding="utf-8")
        else:
            idx.write_text(f"{t}\n{IDX_BEGIN}\n{table}{IDX_END}\n", encoding="utf-8")
    else:
        idx.write_text(f"# CAPABILITIES — índice\n\n{IDX_BEGIN}\n{table}{IDX_END}\n", encoding="utf-8")
    print(f"✓ INDEX.md regenerado · {total} capabilities en {len(mods)} módulos")
    return EXIT_OK


def main(argv: list[str]) -> int:
    if argv and argv[0] == "--index":
        return _index()
    only_module = None
    if "--module" in argv:
        i = argv.index("--module")
        only_module = argv[i + 1] if i + 1 < len(argv) else None

    root = _repo_root()
    cap_root = root / CAP_DIR
    print(f"cap-doctor · {CAP_DIR}" + (f" · module={only_module}" if only_module else ""))
    if not cap_root.is_dir():
        print(f"  ✗ no existe {CAP_DIR} (¿migración pendiente?)", file=sys.stderr)
        return EXIT_NO_DIR

    caps = sorted(cap_root.glob("*/*.yaml"))
    if only_module:
        caps = [c for c in caps if c.parent.name == only_module]
    if not caps:
        print("  ⓘ 0 capabilities encontradas (árbol vacío — nada que validar todavía)")
        return EXIT_OK

    total_errs = 0
    seen_ids: dict[str, Path] = {}
    for cap in caps:
        data, load_err = _load(cap)
        rel = cap.relative_to(root)
        if load_err:
            print(f"  ✗ {rel}: {load_err}")
            total_errs += 1
            continue
        errs = _check_cap(root, cap, data)
        cid = data.get("capability_id")
        if cid and cid in seen_ids:
            errs.append(f"capability_id duplicado (también en {seen_ids[cid].relative_to(root)})")
        elif cid:
            seen_ids[cid] = cap
        if errs:
            total_errs += len(errs)
            print(f"  ✗ {rel}")
            for e in errs:
                print(f"      – {e}")

    n = len(caps)
    if total_errs == 0:
        print(f"  ✓ {n} capabilities válidas — schema OK (R1 archivo-existe · enum · unicidad)")
        return EXIT_OK
    print(f"  ✗ {total_errs} error(es) en {n} capabilities")
    return EXIT_ERRORS


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
