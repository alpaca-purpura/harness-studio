#!/usr/bin/env python3
"""changelog.py — el gate del changelog (paquete 2026-07-26-versionado-y-changelog-metodologicos).

Por qué existe: hasta 2026-07-26 el repo tenía 20 releases en `instaladores/` y CERO registro de
qué traía cada uno. El pedido del operador fue explícito: que actualizar la versión Y decir qué
se agrega / corrige / elimina sea **metodológico**, no un recordatorio. Una norma escrita se
olvida; un gate se choca. Este script es el gate.

Contrato (VC-D2): el bump de versión NO es «tocar 3 manifiestos», es un acto atómico que
PROMUEVE la sección `[Sin publicar]` a una sección versionada. Si `[Sin publicar]` está vacía,
`release` falla y no escribe nada — el Makefile aborta antes de tocar un solo manifiesto.

Subcomandos:
  check [--exige-entradas]   valida formato/coherencia; con la flag exige `[Sin publicar]` no vacía
  add <categoria> <texto>    agrega una entrada a `[Sin publicar]` (idempotente: no duplica)
  release <X.Y.Z> [--fecha]  promueve `[Sin publicar]` → `## [X.Y.Z] — AAAA-MM-DD`
  sin-publicar               imprime lo que hoy iría en la próxima versión

Formato: Keep a Changelog 1.1.0 + SemVer 2.0.0, categorías en español (VC-D1).
Historia previa a `convencion-desde` NO se valida ni se inventa (VC-D4).
"""

from __future__ import annotations

import argparse
import datetime
import pathlib
import re
import sys

# Windows: la consola cp1252 no mapea «✓/·/—» y el print del gate crashearía con
# UnicodeEncodeError — abortando el bump por un carácter de adorno. stdout/stderr a UTF-8
# SIEMPRE (no-op en unix). Mismo fix que scripts/cap_doctor.py.
for _stream in (sys.stdout, sys.stderr):
    if hasattr(_stream, "reconfigure"):
        _stream.reconfigure(encoding="utf-8")

ROOT = pathlib.Path(__file__).resolve().parent.parent
CHANGELOG = ROOT / "CHANGELOG.md"

# Las 6 canónicas de Keep a Changelog, traducidas. Cerrado a propósito: una categoría inventada
# ("Notas", "Varios") es exactamente el agujero por el que se cuela el changelog que no dice nada.
CATEGORIAS = ["Agregado", "Cambiado", "Deprecado", "Eliminado", "Corregido", "Seguridad"]
ALIAS = {
    "added": "Agregado", "agregado": "Agregado", "add": "Agregado", "nuevo": "Agregado",
    "changed": "Cambiado", "cambiado": "Cambiado", "cambio": "Cambiado",
    "deprecated": "Deprecado", "deprecado": "Deprecado",
    "removed": "Eliminado", "eliminado": "Eliminado", "borrado": "Eliminado",
    "fixed": "Corregido", "corregido": "Corregido", "fix": "Corregido", "arreglado": "Corregido",
    "security": "Seguridad", "seguridad": "Seguridad",
}

SIN_PUBLICAR = "Sin publicar"
RE_SECCION = re.compile(r"^## \[([^\]]+)\](?:\s*[—-]\s*(\S+))?\s*$")
RE_CATEGORIA = re.compile(r"^### (.+?)\s*$")
RE_ENTRADA = re.compile(r"^- (.+?)\s*$")
RE_CONVENCION = re.compile(r"<!--\s*convencion-desde:\s*(\d+\.\d+\.\d+)\s*-->")
RE_SEMVER = re.compile(r"^\d+\.\d+\.\d+$")
# Relleno que hace que un changelog exista sin decir nada. Se rechaza explícito.
PLACEHOLDERS = {"tbd", "todo", "pendiente", "n/a", "-", "...", "wip", "por definir"}


class ChangelogError(Exception):
    """Falla de validación: se imprime tal cual y el proceso sale con 1."""


def semver(v: str) -> tuple[int, int, int]:
    return tuple(int(p) for p in v.split("."))  # type: ignore[return-value]


def leer() -> str:
    if not CHANGELOG.exists():
        raise ChangelogError(
            f"no existe {CHANGELOG.relative_to(ROOT)} — es obligatorio "
            "(docs/architecture/conventions/versionado.md §changelog)"
        )
    return CHANGELOG.read_text(encoding="utf-8")


def convencion_desde(texto: str) -> str:
    m = RE_CONVENCION.search(texto)
    if not m:
        raise ChangelogError(
            "falta el marcador `<!-- convencion-desde: X.Y.Z -->`: sin él no se sabe desde qué "
            "versión el changelog es exigible y qué historia quedó sin reconstruir (VC-D4)"
        )
    return m.group(1)


def parsear(texto: str) -> list[dict]:
    """Devuelve las secciones en orden de aparición: {version, fecha, linea, categorias}."""
    secciones: list[dict] = []
    actual: dict | None = None
    cat: str | None = None
    for n, linea in enumerate(texto.splitlines(), start=1):
        if m := RE_SECCION.match(linea):
            actual = {"version": m.group(1), "fecha": m.group(2), "linea": n, "categorias": {}}
            secciones.append(actual)
            cat = None
            continue
        if actual is None:
            continue
        if m := RE_CATEGORIA.match(linea):
            cat = m.group(1)
            actual["categorias"].setdefault(cat, [])
            continue
        if cat and (m := RE_ENTRADA.match(linea)):
            actual["categorias"][cat].append((n, m.group(1)))
    return secciones


def validar(texto: str, exige_entradas: bool = False) -> list[str]:
    """Todas las fallas juntas — arreglar de a una por corrida es un castigo innecesario."""
    fallas: list[str] = []
    desde = convencion_desde(texto)
    secciones = parsear(texto)

    nombres = [s["version"] for s in secciones]
    if SIN_PUBLICAR not in nombres:
        fallas.append(f"falta la sección `## [{SIN_PUBLICAR}]` — es donde se acumula lo que va a la próxima versión")
    if nombres.count(SIN_PUBLICAR) > 1:
        fallas.append(f"hay {nombres.count(SIN_PUBLICAR)} secciones `[{SIN_PUBLICAR}]`; debe haber exactamente una")

    publicadas = [s for s in secciones if s["version"] != SIN_PUBLICAR]
    vistas: dict[str, int] = {}
    previa: tuple[int, int, int] | None = None
    for s in publicadas:
        v = s["version"]
        if not RE_SEMVER.match(v):
            # Se tolera SOLO el bloque de historia no reconstruida, que se rotula aparte.
            if "anteriores" not in v:
                fallas.append(f"línea {s['linea']}: `[{v}]` no es semver plano X.Y.Z")
            continue
        if v in vistas:
            fallas.append(f"línea {s['linea']}: versión [{v}] duplicada (ya estaba en la línea {vistas[v]})")
        vistas[v] = s["linea"]
        if not s["fecha"]:
            fallas.append(f"línea {s['linea']}: [{v}] sin fecha — el formato es `## [{v}] — AAAA-MM-DD`")
        elif not re.match(r"^\d{4}-\d{2}-\d{2}$", s["fecha"]):
            fallas.append(f"línea {s['linea']}: fecha `{s['fecha']}` no es AAAA-MM-DD")
        actual = semver(v)
        if previa is not None and actual >= previa:
            fallas.append(f"línea {s['linea']}: [{v}] rompe el orden descendente (la más nueva va arriba)")
        previa = actual
        if semver(v) >= semver(desde) and not any(s["categorias"].values()):
            fallas.append(f"línea {s['linea']}: [{v}] no tiene ni una entrada — una versión sin cambios no se publica")

    for s in secciones:
        for cat, entradas in s["categorias"].items():
            if cat not in CATEGORIAS:
                fallas.append(
                    f"línea {s['linea']}: categoría `{cat}` en [{s['version']}] no es una de las 6 "
                    f"canónicas ({' · '.join(CATEGORIAS)})"
                )
            for n, txt in entradas:
                if txt.strip().lower().rstrip(".") in PLACEHOLDERS:
                    fallas.append(f"línea {n}: entrada de relleno `{txt}` — un changelog que no dice nada miente")

    if exige_entradas:
        sp = next((s for s in secciones if s["version"] == SIN_PUBLICAR), None)
        if sp is None or not any(sp["categorias"].values()):
            fallas.append(
                f"`[{SIN_PUBLICAR}]` está vacía: no se puede versionar sin decir qué cambió (VC-D2).\n"
                f"    agregá lo que hiciste:  python3 scripts/changelog.py add Agregado \"lo que hiciste\"\n"
                f"    categorías: {' · '.join(CATEGORIAS)}"
            )
    return fallas


def bloque_sin_publicar(texto: str) -> str:
    lineas = texto.splitlines()
    secciones = parsear(texto)
    sp = next((s for s in secciones if s["version"] == SIN_PUBLICAR), None)
    if sp is None:
        return ""
    inicio = sp["linea"]  # 1-based: la línea del encabezado
    fin = len(lineas)
    for s in secciones:
        if s["linea"] > inicio:
            fin = s["linea"] - 1
            break
    return "\n".join(lineas[inicio:fin]).strip("\n")


def plantilla_vacia() -> str:
    return "\n".join(f"### {c}\n" for c in CATEGORIAS).rstrip("\n")


def cmd_check(args: argparse.Namespace) -> int:
    texto = leer()
    fallas = validar(texto, exige_entradas=args.exige_entradas)
    if fallas:
        print("CHANGELOG.md — no pasa:", file=sys.stderr)
        for f in fallas:
            print(f"  ✗ {f}", file=sys.stderr)
        return 1
    print(f"changelog ✓ (convención desde {convencion_desde(texto)})")
    return 0


def cmd_add(args: argparse.Namespace) -> int:
    cat = ALIAS.get(args.categoria.strip().lower(), args.categoria.strip().capitalize())
    if cat not in CATEGORIAS:
        print(f"categoría `{args.categoria}` inválida. Usá una de: {' · '.join(CATEGORIAS)}", file=sys.stderr)
        return 1
    texto = leer()
    entrada = " ".join(args.texto).strip()
    if not entrada or entrada.lower() in PLACEHOLDERS:
        print("la entrada no puede estar vacía ni ser relleno (TBD/pendiente/…)", file=sys.stderr)
        return 1

    lineas = texto.splitlines()
    secciones = parsear(texto)
    sp = next((s for s in secciones if s["version"] == SIN_PUBLICAR), None)
    if sp is None:
        print(f"no encontré `## [{SIN_PUBLICAR}]`", file=sys.stderr)
        return 1
    if any(entrada == t for _, t in sp["categorias"].get(cat, [])):
        print(f"ya estaba en {cat} — no se duplica")
        return 0

    fin_sp = len(lineas)
    for s in secciones:
        if s["linea"] > sp["linea"]:
            fin_sp = s["linea"] - 1
            break

    idx_cat = None
    for i in range(sp["linea"], fin_sp):
        if (m := RE_CATEGORIA.match(lineas[i])) and m.group(1) == cat:
            idx_cat = i
            break
    if idx_cat is None:  # la categoría no estaba: se crea al final del bloque, en orden canónico
        insert = fin_sp
        lineas[insert:insert] = [f"### {cat}", f"- {entrada}", ""]
    else:
        insert = idx_cat + 1
        while insert < fin_sp and RE_ENTRADA.match(lineas[insert]):
            insert += 1
        lineas.insert(insert, f"- {entrada}")

    CHANGELOG.write_text("\n".join(lineas).rstrip("\n") + "\n", encoding="utf-8")
    print(f"{cat}: {entrada}")
    return 0


def cmd_release(args: argparse.Namespace) -> int:
    version = args.version.lstrip("v")
    if not RE_SEMVER.match(version):
        print(f"`{version}` no es semver plano X.Y.Z (sin prefijo v — lo exige Keygen)", file=sys.stderr)
        return 1
    texto = leer()
    fallas = validar(texto, exige_entradas=True)
    if fallas:
        print(f"no se puede versionar a {version} — el changelog no pasa:", file=sys.stderr)
        for f in fallas:
            print(f"  ✗ {f}", file=sys.stderr)
        return 1
    if any(s["version"] == version for s in parsear(texto)):
        print(f"[{version}] ya existe en el changelog — ¿bump repetido?", file=sys.stderr)
        return 1

    fecha = args.fecha or datetime.date.today().isoformat()
    cuerpo = bloque_sin_publicar(texto)
    lineas = texto.splitlines()
    sp = next(s for s in parsear(texto) if s["version"] == SIN_PUBLICAR)
    fin_sp = len(lineas)
    for s in parsear(texto):
        if s["linea"] > sp["linea"]:
            fin_sp = s["linea"] - 1
            break

    nuevo = (
        lineas[: sp["linea"]]                        # todo hasta el encabezado [Sin publicar]
        + ["", plantilla_vacia(), "", f"## [{version}] — {fecha}", "", cuerpo, ""]
        + lineas[fin_sp:]
    )
    CHANGELOG.write_text("\n".join(nuevo).rstrip("\n") + "\n", encoding="utf-8")
    print(f"changelog: [{SIN_PUBLICAR}] → [{version}] — {fecha}")
    return 0


def cmd_sin_publicar(_: argparse.Namespace) -> int:
    cuerpo = bloque_sin_publicar(leer())
    entradas = [l for l in cuerpo.splitlines() if l.startswith("- ")]
    if not entradas:
        print("(vacío — no hay nada que publicar todavía)")
        return 0
    print(cuerpo)
    return 0


def main() -> int:
    p = argparse.ArgumentParser(prog="changelog.py", description=__doc__.splitlines()[0])
    sub = p.add_subparsers(dest="cmd", required=True)

    c = sub.add_parser("check", help="valida formato y coherencia")
    c.add_argument("--exige-entradas", action="store_true", help="además, [Sin publicar] no puede estar vacía")
    c.set_defaults(func=cmd_check)

    a = sub.add_parser("add", help="agrega una entrada a [Sin publicar]")
    a.add_argument("categoria", help=" · ".join(CATEGORIAS))
    a.add_argument("texto", nargs="+")
    a.set_defaults(func=cmd_add)

    r = sub.add_parser("release", help="promueve [Sin publicar] a una versión")
    r.add_argument("version")
    r.add_argument("--fecha", default=None, help="AAAA-MM-DD (default: hoy)")
    r.set_defaults(func=cmd_release)

    s = sub.add_parser("sin-publicar", help="imprime lo que iría en la próxima versión")
    s.set_defaults(func=cmd_sin_publicar)

    args = p.parse_args()
    try:
        return args.func(args)
    except ChangelogError as e:
        print(f"CHANGELOG.md: {e}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
