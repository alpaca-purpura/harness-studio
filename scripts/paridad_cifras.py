#!/usr/bin/env python3
"""Cuenta los veredictos de las tablas de un PARIDAD.md — A-12.

Las dos cifras que el operador lee para decidir si firma estaban TECLEADAS, y las dos
estaban mal (§2 decía 30 ✅ · 5 ⚠️ y eran 31 · 4; §2.1 decía 17 ✅ y eran 18). La causa era
benigna —una fila pasó de ⚠️ a ✅ al cerrarse N-1 y nadie recontó el resumen— pero la regla
del repo es dura: **las cifras se GENERAN, no se teclean**. Un gate humano no debería
apoyarse en un conteo a mano.

Cómo se regenera (desde la raíz del repo):

    python3 scripts/paridad_cifras.py \\
        docs/product/stories/2026-07-26-conversaciones-del-panel/PARIDAD.md

    # y para que falle si el texto declara algo distinto de lo contado:
    python3 scripts/paridad_cifras.py <ruta> --check

Qué cuenta y qué no: sólo la **última celda** de cada fila de tabla markdown (la columna
`veredicto`), y sólo dentro de las secciones que declaran una. El veredicto es el PRIMER
símbolo de esa celda — así una fila `⚠️ …porque ✅ tal cosa` cuenta como ⚠️ y no como las
dos. Las filas de encabezado y las separadoras (`|---|`) se descartan.
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

# El orden importa: se busca el PRIMER símbolo que aparezca en la celda.
SIMBOLOS = ("✅", "⚠️", "❌")

# Una sección arranca en un encabezado `## N` o `### N.N` y termina en el siguiente.
RE_ENCABEZADO = re.compile(r"^(#{2,4})\s+(.*)$")
# «**Resumen: 30 ✅ · 5 ⚠️ · 0 ❌.**» / «**Resumen §2.1: 17 ✅ · 1 ⚠️ · 2 ❌.**»
RE_RESUMEN = re.compile(r"Resumen[^:]*:\s*\*{0,2}\s*(\d+)\s*✅\s*·\s*(\d+)\s*⚠️\s*·\s*(\d+)\s*❌")


def veredicto_de(celda: str) -> str | None:
    """Devuelve el símbolo de veredicto de una celda, o None si no tiene ninguno."""
    posiciones = [(celda.find(s), s) for s in SIMBOLOS if s in celda]
    if not posiciones:
        return None
    return min(posiciones)[1]


def contar(texto: str) -> dict[str, dict[str, int]]:
    """Cuenta veredictos por sección. Clave = título del encabezado más cercano."""
    conteos: dict[str, dict[str, int]] = {}
    seccion = "(sin sección)"
    for linea in texto.splitlines():
        if m := RE_ENCABEZADO.match(linea):
            seccion = m.group(2).strip()
            continue
        s = linea.strip()
        if not s.startswith("|") or not s.endswith("|"):
            continue
        celdas = [c.strip() for c in s.strip("|").split("|")]
        if len(celdas) < 2 or all(set(c) <= set("-: ") for c in celdas):
            continue  # separadora
        if (v := veredicto_de(celdas[-1])) is None:
            continue  # encabezado, o fila sin veredicto
        conteos.setdefault(seccion, dict.fromkeys(SIMBOLOS, 0))[v] += 1
    return conteos


def declarados(texto: str) -> dict[str, tuple[int, int, int]]:
    """Lee los resúmenes TECLEADOS, para poder contrastarlos."""
    out: dict[str, tuple[int, int, int]] = {}
    seccion = "(sin sección)"
    for linea in texto.splitlines():
        if m := RE_ENCABEZADO.match(linea):
            seccion = m.group(2).strip()
            continue
        if m := RE_RESUMEN.search(linea):
            out[seccion] = (int(m.group(1)), int(m.group(2)), int(m.group(3)))
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("ruta", type=Path, help="el PARIDAD.md a contar")
    ap.add_argument("--check", action="store_true", help="sale 1 si un resumen tecleado no coincide")
    args = ap.parse_args()

    texto = args.ruta.read_text(encoding="utf-8")
    conteos, dichos = contar(texto), declarados(texto)

    print(f"{args.ruta}\n")
    print(f"{'sección':<52} {'✅':>4} {'⚠️':>4} {'❌':>4}   declarado")
    print("─" * 92)
    malas = 0
    for seccion, c in conteos.items():
        contado = (c["✅"], c["⚠️"], c["❌"])
        d = dichos.get(seccion)
        if d is None:
            nota = "—"
        elif d == contado:
            nota = f"{d[0]} ✅ · {d[1]} ⚠️ · {d[2]} ❌  ✓"
        else:
            nota = f"{d[0]} ✅ · {d[1]} ⚠️ · {d[2]} ❌  ✗ TECLEADO MAL"
            malas += 1
        print(f"{seccion[:52]:<52} {contado[0]:>4} {contado[1]:>4} {contado[2]:>4}   {nota}")

    total = tuple(sum(c[s] for c in conteos.values()) for s in SIMBOLOS)
    print("─" * 92)
    print(f"{'TOTAL de filas con veredicto':<52} {total[0]:>4} {total[1]:>4} {total[2]:>4}")

    if args.check and malas:
        print(f"\n✗ {malas} resumen(es) tecleado(s) no coinciden con lo contado.", file=sys.stderr)
        return 1
    if args.check:
        print("\n✓ los resúmenes declarados coinciden con lo contado.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
