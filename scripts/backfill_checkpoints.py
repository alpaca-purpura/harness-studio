#!/usr/bin/env python3
"""backfill_checkpoints.py — siembra el `checkpoint.md` que le falta a los paquetes viejos.

POR QUÉ: el board del cockpit (`tools/cockpit`) lee UN archivo por story —
`docs/product/stories/<pkg>/checkpoint.md` — y toma el estado de su frontmatter. Los
paquetes anteriores a la homologación no lo tienen, así que el board los ignora: 4 de 42
visibles. Esto los incorpora SIN inventar historia.

QUÉ NO HACE (regla dura, METODOLOGIA §4 — gris ≠ verde):

  * No fabrica firmas. `chris_verify.signoff` sale siempre en `false`: la firma 🧑‍⚖️ la
    pone un humano en el PARIDAD, y deducirla desde un grep sería exactamente el «pass
    fabricado» que la doctrina prohíbe.
  * No pisa un checkpoint existente.
  * No escribe nada sin `--write`. El default es el dry-run, para que el estado inferido
    se RATIFIQUE antes de tocar disco.

CÓMO INFIERE: cada estado sale de una señal concreta del paquete, y la señal se IMPRIME
al lado del estado. Si una fila está mal, se corrige a mano después del write — la tabla
del dry-run existe para verlo antes.

Uso:
    python scripts/backfill_checkpoints.py              # dry-run: tabla + evidencia
    python scripts/backfill_checkpoints.py --write      # escribe los que faltan
"""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

import yaml

# La consola de Windows usa cp1252 y revienta al imprimir el 🧑‍⚖️ del gate
# (UnicodeEncodeError) — stdout/stderr a UTF-8 SIEMPRE (no-op en unix).
# Mismo remedio que scripts/cap_doctor.py.
for _stream in (sys.stdout, sys.stderr):
    if hasattr(_stream, "reconfigure"):
        _stream.reconfigure(encoding="utf-8")

RAIZ = Path(__file__).resolve().parent.parent
STORIES = RAIZ / "docs" / "product" / "stories"
CHECKPOINT_RAIZ = RAIZ / "docs" / "product" / "checkpoint.md"

# Los 10 estados del descriptor de proceso que consume el cockpit
# (tools/cockpit/go/process/sdd-default.yaml). Un estado fuera de esta lista lo rechaza
# el board, así que se valida acá antes de escribirlo.
ESTADOS = {
    "idea", "refining", "refined", "ready", "developing",
    "developed", "reviewing", "done", "parked", "dropped",
}

JUEZ = "\N{ADULT}‍\N{SCALES}"  # 🧑‍⚖️ — el emoji del gate humano


def _lineas_del_gate(texto: str) -> list[str]:
    """Las líneas del PARIDAD que hablan del gate humano, la más pertinente primero.

    Un PARIDAD menciona el 🧑‍⚖️ varias veces (el mockup que se firmó, la checklist, el
    gate final). Las que además dicen «gate»/«firma» van adelante para que la evidencia
    que se imprime sea la que sostiene el veredicto y no una mención suelta.
    """
    lineas = [ln.strip() for ln in texto.splitlines() if JUEZ in ln]
    pertinente = re.compile(r"\bgate\b|\bfirma\b", re.IGNORECASE)
    return sorted(lineas, key=lambda ln: 0 if pertinente.search(ln) else 1)


def _estado_por_paridad(paridad: Path) -> tuple[str, str]:
    """Lee el gate 🧑‍⚖️ de un PARIDAD. Devuelve (estado, evidencia)."""
    texto = paridad.read_text(encoding="utf-8", errors="replace")
    lineas = _lineas_del_gate(texto)

    # PENDIENTE gana sobre FIRMADO: un paquete con varios gates no está cerrado hasta
    # que no queda ninguno pendiente. Preferir el estado MENOS avanzado ante la duda es
    # la lectura honesta — inflar el estado es el error caro.
    for ln in lineas:
        if re.search(r"PENDIENTE|SIN FIRMAR|\[ \]", ln, re.IGNORECASE):
            return "developed", f"PARIDAD con gate {JUEZ} pendiente: «{_corta(ln)}»"

    for ln in lineas:
        if re.search(r"FIRMAD[OA]", ln, re.IGNORECASE):
            return "done", f"PARIDAD con gate {JUEZ} firmado: «{_corta(ln)}»"

    return "developed", "PARIDAD presente, sin línea de gate reconocible"


def _corta(s: str, n: int = 60) -> str:
    s = re.sub(r"\s+", " ", s).strip("> -*[]x ")
    return s if len(s) <= n else s[: n - 1] + "…"


def _menciones_del_checkpoint_raiz() -> dict[str, str]:
    """Slugs que el checkpoint raíz declara ACTIVOS o PAUSADOS.

    Dos formas conviven en el archivo y las dos cuentan:
      * por SECCIÓN — todo lo enlazado bajo `## Paquete de trabajo activo` está activo
        aunque su línea no diga la palabra (es el caso de ingesta-escenario-b);
      * por LÍNEA — un `ACTIVO`/`PAUSADA` explícito en la misma línea del enlace.
    """
    if not CHECKPOINT_RAIZ.exists():
        return {}
    fuera: dict[str, str] = {}
    en_seccion = False
    bullet = 0
    for ln in CHECKPOINT_RAIZ.read_text(encoding="utf-8", errors="replace").splitlines():
        if ln.startswith("#"):
            en_seccion = bool(re.search(r"paquete.*activo", ln, re.IGNORECASE))
            bullet = 0
        elif en_seccion and ln.startswith("- "):
            bullet += 1
        # La sección «Paquete de trabajo activo» son ~240 líneas de relato: incluye el
        # «Retomar aquí», los pausados y el historial. Solo su PRIMER bullet es el
        # paquete en curso — el resto necesita decir ACTIVO/PAUSADA en su propia línea.
        # Sin este corte, la sección entera marcaba 18 paquetes como activos.
        en_bullet_activo = en_seccion and bullet == 1
        for slug in re.findall(r"stories/([0-9]{4}-[0-9]{2}-[0-9]{2}-[a-z0-9-]+)", ln):
            if re.search(r"\bPAUSAD[AO]\b", ln):
                fuera[slug] = "pausado"
            elif en_bullet_activo or re.search(r"\bACTIVO\b|\bACTIVA\b|\bEN CURSO\b", ln):
                # No pisar un «pausado» ya registrado en otra línea: pausa gana.
                fuera.setdefault(slug, "activo")
    return fuera


def inferir(pkg: Path, raiz_dice: dict[str, str]) -> tuple[str, str]:
    """(estado, evidencia) para UN paquete. Precedencia explícita, de más fuerte a más débil."""
    slug = pkg.name
    archivos = {p.name for p in pkg.iterdir()}

    # 1. Lo que el checkpoint raíz declara gana sobre cualquier inferencia de archivos:
    #    es el estado que el operador mantiene a mano.
    if raiz_dice.get(slug) == "activo":
        return "developing", "checkpoint raíz lo declara ACTIVO / EN CURSO"
    if raiz_dice.get(slug) == "pausado":
        return "parked", "checkpoint raíz lo declara PAUSADO"

    # 2. El PARIDAD es el gate de cierre: si existe, su firma manda.
    for nombre in ("PARIDAD.md", "paridad.md"):
        if nombre in archivos:
            return _estado_por_paridad(pkg / nombre)

    # 3. Sin PARIDAD, la etapa se lee por el artefacto más avanzado que produjo.
    if {"spec.md", "spec-funcional.md", "00-story.md"} & archivos:
        return "developing", "tiene spec pero no llegó a PARIDAD"
    if "informe.md" in archivos:
        return "done", "paquete de auditoría: su entregable es el informe"
    if "decisiones.md" in archivos:
        return "refined", "tiene decisiones firmadas, sin spec"
    return "idea", "solo INDEX — no arrancó"


def _modulo(pkg: Path) -> str:
    """Módulo aproximado desde el slug (el board agrupa por módulo)."""
    slug = re.sub(r"^\d{4}-\d{2}-\d{2}-", "", pkg.name)
    return slug.split("-")[0]


def render(slug: str, estado: str, modulo: str, evidencia: str) -> str:
    return f"""---
story_id: {slug}
state: {estado}
module: {modulo}
# Sembrado por scripts/backfill_checkpoints.py — el paquete es anterior a la disciplina
# de checkpoint (METODOLOGIA §10) y sin este archivo el board no lo ve.
# Estado inferido de: {evidencia}
# Corregilo a mano si la inferencia erró: este archivo es la fuente, no el script.
backfilled: true
chris_verify:
  # NUNCA se deduce una firma desde un grep (§4: gris ≠ verde). Si el gate 🧑‍⚖️ del
  # PARIDAD está firmado, ponelo en true a mano citando la línea que lo firma.
  signoff: false
---

# {slug}

Checkpoint sembrado para dar visibilidad al paquete en el board. El contenido real del
paquete vive en su `INDEX.md` y en los artefactos hermanos.
"""


# ── sincronización release → story ──────────────────────────────────────────
#
# El roadmap del cockpit NO agrupa por la lista `stories[]` del release: agrupa por el
# campo `release:` del frontmatter de cada story (handlers_stories.go::readCheckpoint).
# Sin ese campo el release se ve con «0 stories» aunque las liste todas. Esta pasada
# copia la pertenencia desde el release (la fuente que la declara) a cada checkpoint.

RELEASES = RAIZ / "docs" / "product" / "releases"


def _stories_por_release() -> dict[str, str]:
    """story_id → release_id, leído de docs/product/releases/*.yaml."""
    fuera: dict[str, str] = {}
    if not RELEASES.is_dir():
        return fuera
    for yml in sorted(RELEASES.glob("*.yaml")):
        texto = yml.read_text(encoding="utf-8", errors="replace")
        if texto.lstrip().startswith("---"):
            cuerpo = texto.lstrip()[3:]
            fin = cuerpo.find("\n---")
            texto = cuerpo[:fin] if fin != -1 else cuerpo
        datos = yaml.safe_load(texto)
        if not isinstance(datos, dict):
            continue
        rid = datos.get("release_id") or yml.stem
        for sid in datos.get("stories") or []:
            fuera[str(sid)] = str(rid)
    return fuera


def sync_releases(escribir: bool) -> int:
    pertenencia = _stories_por_release()
    if not pertenencia:
        print("no hay releases con stories declaradas — nada que sincronizar")
        return 0

    cambios = 0
    print(f"{'STORY':<48} {'RELEASE':<26} ACCIÓN")
    print("-" * 100)
    for sid, rid in sorted(pertenencia.items()):
        ckpt = STORIES / sid / "checkpoint.md"
        if not ckpt.exists():
            print(f"{sid:<48} {rid:<26} SIN checkpoint (¿story futura?) — se omite")
            continue
        texto = ckpt.read_text(encoding="utf-8")
        actual = re.search(r"^release:\s*(.+)$", texto, re.MULTILINE)
        if actual and actual.group(1).strip() == rid:
            continue
        if actual:
            nuevo = re.sub(r"^release:.*$", f"release: {rid}", texto, count=1, flags=re.MULTILINE)
            accion = f"actualiza (era {actual.group(1).strip()})"
        else:
            # Se inserta justo después de `state:` para que quede junto a los campos que
            # el board lee, no al final entre los comentarios del backfill.
            nuevo = re.sub(r"^(state:.*)$", rf"\1\nrelease: {rid}", texto, count=1, flags=re.MULTILINE)
            accion = "agrega"
        print(f"{sid:<48} {rid:<26} {accion}")
        cambios += 1
        if escribir:
            ckpt.write_text(nuevo, encoding="utf-8")

    print("-" * 100)
    print(f"{cambios} checkpoints a tocar de {len(pertenencia)} pertenencias declaradas")
    if not escribir:
        print("\nDRY-RUN — no se escribió nada.")
    elif cambios:
        print(f"\nOK — {cambios} checkpoints sincronizados.")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--write", action="store_true", help="escribe (sin esto: dry-run)")
    ap.add_argument("--sync-releases", action="store_true",
                    help="copia la pertenencia a release desde docs/product/releases/*.yaml al frontmatter de cada story")
    args = ap.parse_args()

    if args.sync_releases:
        return sync_releases(args.write)

    if not STORIES.is_dir():
        print(f"no existe {STORIES}", file=sys.stderr)
        return 1

    raiz_dice = _menciones_del_checkpoint_raiz()
    faltantes = sorted(p for p in STORIES.iterdir() if p.is_dir() and not (p / "checkpoint.md").exists())
    ya = sum(1 for p in STORIES.iterdir() if p.is_dir() and (p / "checkpoint.md").exists())

    if not faltantes:
        print(f"nada que sembrar — los {ya} paquetes ya tienen checkpoint.md")
        return 0

    print(f"{'PAQUETE':<48} {'ESTADO':<12} EVIDENCIA")
    print("-" * 118)
    escritos = 0
    for pkg in faltantes:
        estado, evidencia = inferir(pkg, raiz_dice)
        if estado not in ESTADOS:  # cinturón: un estado inválido rompe el board
            print(f"estado inferido inválido {estado!r} para {pkg.name}", file=sys.stderr)
            return 1
        print(f"{pkg.name:<48} {estado:<12} {evidencia}")

        if args.write:
            contenido = render(pkg.name, estado, _modulo(pkg), evidencia)
            # El archivo que escribimos tiene que ser legible por el MISMO parser que lo
            # va a leer: se valida antes de tocar disco, no después.
            fm = contenido.split("---")[1]
            datos = yaml.safe_load(fm)
            if not isinstance(datos, dict) or datos.get("state") not in ESTADOS:
                print(f"frontmatter generado inválido para {pkg.name}", file=sys.stderr)
                return 1
            (pkg / "checkpoint.md").write_text(contenido, encoding="utf-8")
            escritos += 1

    print("-" * 118)
    resumen: dict[str, int] = {}
    for pkg in faltantes:
        e, _ = inferir(pkg, raiz_dice)
        resumen[e] = resumen.get(e, 0) + 1
    print(f"{len(faltantes)} paquetes sin checkpoint · {ya} ya tenían · reparto: {resumen}")

    if args.write:
        print(f"\nOK — {escritos} checkpoint.md escritos.")
    else:
        print("\nDRY-RUN — no se escribió nada. Ratificá la tabla y corré con --write.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
