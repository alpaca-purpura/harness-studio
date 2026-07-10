#!/usr/bin/env python3
# capabilities_to_yaml.py — migración CAPABILITIES.md → docs/product/capabilities/{module}/{slug}.yaml
# (homologación 2026-07-09). GENERADO, no tecleado: parsea el SSoT actual y emite un YAML por-cap
# en el formato del plugin (schema docs/product/_templates/capability.template.yaml).
#
# Idempotente: reescribe el árbol capabilities/ desde CAPABILITIES.md. Mientras CAPABILITIES.md
# siga siendo el enforcer R1/R2 (capability_trace_test.go), este árbol es la VISTA homologada;
# tras el flip de autoridad, el YAML pasa a ser fuente y el converter deja de correr.
#
# Uso: python3 scripts/capabilities_to_yaml.py   (desde la raíz del repo)

from __future__ import annotations

import re
import unicodedata
from pathlib import Path

SRC = Path("CAPABILITIES.md")
OUT = Path("docs/product/capabilities")

# grupo (letra del header `## X · Título`) → slug de módulo (carpeta)
GROUP_MODULE = {
    "A": "cli-daemon", "B": "dominio-l0", "C": "loader", "D": "indice-persistencia",
    "E": "conformance", "F": "conductor", "G": "provisioning", "H": "handoff",
    "I": "http-sse", "J": "usecases", "K": "self-update", "L": "fe-mapa",
    "M": "fe-chat", "N": "fe-shell", "O": "tauri",
}
STATUS_MAP = {"vivo": "vivo", "nc": "vivo·nc", "stub": "stub", "parcial": "parcial"}
CODE_EXTS = (".go", ".ts", ".tsx", ".rs")

RE_GROUP = re.compile(r"^##\s+([A-O])\s+·")
RE_CAP = re.compile(r"^-\s+\*\*CAP-(\d+)\s+·\s+(.+?)\*\*\s*(.*)$")
RE_BT = re.compile(r"`([^`]+)`")


def slugify(s: str) -> str:
    s = re.sub(r"\(.*?\)", "", s)  # tirar (paréntesis)
    s = unicodedata.normalize("NFKD", s).encode("ascii", "ignore").decode()
    s = re.sub(r"[^a-zA-Z0-9]+", "-", s).strip("-").lower()
    return re.sub(r"-+", "-", s)[:48].strip("-")


def clean_pointer(tok: str) -> list[str]:
    """`path#A:12,#B:34` → ['path#A','path#B'] · `path.rs:52` → ['path.rs'] · descarta _test.go."""
    file = tok.split("#", 1)[0].strip().rstrip(",").strip()
    file = re.sub(r":\d+$", "", file)  # despojar `:línea` (caps Rust usan file.rs:NN sin símbolo)
    if not file.endswith(CODE_EXTS) or file.endswith("_test.go"):
        return []
    syms = re.findall(r"#([A-Za-z0-9_.]+)", tok)
    return [f"{file}#{s}" for s in syms] if syms else [file]


def parse_line(num: str, name: str, rest: str) -> dict:
    # estado = primer backtick de rest (`verb·status`)
    m = RE_BT.search(rest)
    verb, status = "infra", "vivo·nc"
    if m and "·" in m.group(1):
        v, _, st = m.group(1).partition("·")
        verb, status = v.strip(), STATUS_MAP.get(st.strip().lower(), "vivo·nc")
    # separar sección de punteros (antes de `valida:`) de la de tests
    ptr_part, _, val_part = rest.partition("valida:")
    pointers: list[str] = []
    for bt in RE_BT.findall(ptr_part):
        pointers.extend(clean_pointer(bt))
    # dedup preservando orden
    seen: dict[str, None] = {}
    for p in pointers:
        seen.setdefault(p, None)
    pointers = list(seen)
    valida = [t.strip().rstrip(".") for bt in RE_BT.findall(val_part) for t in bt.split(",") if t.strip()]
    return {"num": int(num), "name": name.strip(), "verb": verb, "status": status,
            "pointers": pointers, "valida": valida}


def yaml_list(items: list[str], indent: str = "  ") -> str:
    if not items:
        return " []"
    return "\n" + "\n".join(f'{indent}- "{i}"' for i in items)


def emit(cap: dict, module: str) -> str:
    slug = slugify(cap["name"]) or f"cap-{cap['num']:02d}"
    cid = f"arnesia.{module}.{slug}"
    return f"""---
# GENERADO desde CAPABILITIES.md por scripts/capabilities_to_yaml.py — no editar a mano
# (mientras CAPABILITIES.md sea el enforcer). Schema: docs/product/_templates/capability.template.yaml
capability_id: {cid}
cap_num: CAP-{cap['num']:02d}
slug: {slug}
name: "{cap['name'].replace('"', "'")}"
status: {cap['status']}          # GENERADO (R4) — no teclear
module: {module}
verb: {cap['verb']}
pointers:{yaml_list(cap['pointers'])}
valida:{yaml_list(cap['valida'])}
scenarios: []
business_rules: []
related_capabilities:
  depends_on: []
  enables: []
source_ref: "CAPABILITIES.md#CAP-{cap['num']:02d}"
---

# {cap['name']}

Capability derivada del código real (barrido 7-subagentes, CAPABILITIES.md HS-18).
""", slug


def main() -> int:
    text = SRC.read_text(encoding="utf-8")
    module = None
    caps: list[tuple[dict, str]] = []
    incov = False
    coverage: list[str] = []
    for ln in text.splitlines():
        if "<!--coverage-->" in ln:
            incov = True
            continue
        if "<!--/coverage-->" in ln:
            incov = False
            continue
        if incov:
            for bt in RE_BT.findall(ln):
                f = bt.split("#", 1)[0].strip()
                if f.endswith(CODE_EXTS):
                    coverage.append(f)
            continue
        g = RE_GROUP.match(ln)
        if g:
            module = GROUP_MODULE[g.group(1)]
            continue
        c = RE_CAP.match(ln.strip())
        if c and module:
            caps.append((parse_line(*c.groups()), module))

    # limpiar árbol previo (idempotencia) + escribir
    if OUT.exists():
        for old in OUT.glob("*/*.yaml"):
            old.unlink()
    written = 0
    slugs_by_mod: dict[str, set] = {}
    for cap, module in caps:
        body, slug = emit(cap, module)
        s = slugs_by_mod.setdefault(module, set())
        if slug in s:  # colisión → sufijo cap-num
            slug = f"{slug}-{cap['num']:02d}"
            body = body.replace(f"slug: ", f"slug: ", 1)  # slug ya embebido; regenerar mínimo
            body = re.sub(r"^slug: .*$", f"slug: {slug}", body, count=1, flags=re.M)
            body = re.sub(r"capability_id: arnesia\.%s\..*$" % re.escape(module),
                          f"capability_id: arnesia.{module}.{slug}", body, count=1, flags=re.M)
        s.add(slug)
        d = OUT / module
        d.mkdir(parents=True, exist_ok=True)
        (d / f"{slug}.yaml").write_text(body, encoding="utf-8")
        written += 1

    # _coverage.yaml (soporte reclamado para R2)
    cov_uniq = list(dict.fromkeys(coverage))
    cov_yaml = "# GENERADO desde el bloque <!--coverage--> de CAPABILITIES.md — soporte para R2.\n"
    cov_yaml += "support_files:\n" + "\n".join(f'  - "{f}"' for f in cov_uniq) + "\n"
    (OUT / "_coverage.yaml").write_text(cov_yaml, encoding="utf-8")

    print(f"✓ {written} capabilities → {OUT}/  ({len(slugs_by_mod)} módulos) · coverage: {len(cov_uniq)} archivos")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
