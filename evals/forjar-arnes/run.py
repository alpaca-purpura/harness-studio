#!/usr/bin/env python3
"""run.py — eval de la skill `forjar-arnes`.

Mide si la skill CAMBIA LA CONDUCTA de la sesión: que abra los nodos del estándar antes de
escribir, que lo declare, y que el manifiesto que deje sea válido. Los tests del repo ya
cubren que el archivo existe y se materializa; esto cubre la otra mitad.

Cómo: replica la inyección que hace `provisioner.go::materialize` en un temp y lanza
`claude -p` con los flags de `SpawnArgs`. Las divergencias respecto del spawn real (permisos
sin HITL, sin MCP) están tabuladas en ../README.md — son deliberadas y son la única forma de
correr desatendido.

Uso:
    python evals/forjar-arnes/run.py
    python evals/forjar-arnes/run.py --caso crear-desde-cero
    python evals/forjar-arnes/run.py --conservar
"""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

import yaml
from jsonschema import Draft202012Validator

# La consola de Windows usa cp1252 y revienta con los emojis del veredicto.
# `line_buffering` no es cosmético: cada caso son minutos de sesión real, y con el buffer
# por bloques (el default cuando la salida es un pipe, p.ej. en CI o `| tee`) no se ve NADA
# hasta que termina todo. El avance tiene que verse mientras corre.
for _s in (sys.stdout, sys.stderr):
    if hasattr(_s, "reconfigure"):
        _s.reconfigure(encoding="utf-8", line_buffering=True)

AQUI = Path(__file__).resolve().parent
RAIZ = AQUI.parent.parent
CASOS = AQUI / "casos.yaml"
SCHEMA = RAIZ / "docs" / "architecture" / "contracts" / "schema" / "graph.l0.schema.json"

# Los nodos que `forjar-arnes` manda abrir en su paso 1. Que aparezca alguno en el texto
# final es lo que cuenta como «declaró qué abrió».
NODOS_ESPERADOS = ("harness-profile", "rules", "hooks", "skills")

TIMEOUT_POR_CASO = 600  # 10 min: una forja completa escribe varios archivos.


# ── la inyección (espejo de provisioner.go::materialize) ─────────────────────

def armar_inyeccion(destino: Path) -> dict[str, Path]:
    """Materializa kit/ + knowhow/ + doctrine.md igual que el provisioner.

    El nombre `knowhow/` NO es decorativo: las skills referencian `knowhow/<clase>.md` por
    esa ruta exacta. Si el directorio se llamara distinto, las referencias no resolverían y
    el eval mediría otra cosa.
    """
    kit = destino / "kit"
    knowhow = destino / "knowhow"
    shutil.copytree(RAIZ / "kit", kit)
    shutil.copytree(RAIZ / "docs" / "architecture" / "knowledge" / "elements", knowhow)
    doctrine = destino / "doctrine.md"
    # Suelto, no dentro del plugin: --append-system-prompt-file no lee de plugins.
    shutil.copy(RAIZ / "kit" / "doctrine.md", doctrine)
    return {"kit": kit, "knowhow": knowhow, "doctrine": doctrine}


def flags_spawn(inj: dict[str, Path], con_kit: bool) -> list[str]:
    """Los flags de conductor.go::SpawnArgs, con las divergencias del README.

    `con_kit: False` quita los TRES flags de inyección, no solo `--plugin-dir`. Quitar solo
    el plugin dejaba un baseline sucio: `doctrine.md` nombra a `forjar-arnes`, así que la
    sesión intentaba invocar una skill que no estaba cargada. Sin kit = sesión pelada.
    """
    args = [
        "-p",
        "--output-format", "stream-json",
        "--verbose",
        "--setting-sources", "project,local",
        "--max-turns", "40",
        # Divergencia declarada: el spawn real usa `default` + --permission-prompt-tool
        # stdio, que manda cada escritura al Dock a esperar a un humano. Headless eso
        # cuelga la forja.
        "--permission-mode", "acceptEdits",
    ]
    if con_kit:
        args += [
            "--append-system-prompt-file", str(inj["doctrine"]),
            "--add-dir", str(inj["knowhow"]),
            "--plugin-dir", str(inj["kit"]),
        ]
    return args


# ── correr una sesión y leer qué hizo ───────────────────────────────────────

def correr(prompt: str, cwd: Path, args: list[str]) -> dict:
    """Lanza claude y devuelve lo observado: tool_uses + texto + crudo.

    El prompt va por STDIN, no como argumento posicional. `--add-dir` es variádico
    (`<directories...>`), así que un prompt detrás se lo come como si fuera otro
    directorio: el caso baseline —el único sin `--plugin-dir` cerrando la lista— moría con
    «Input must be provided either through stdin or as a prompt argument». Por stdin el
    orden de los flags deja de importar.
    """
    cmd = ["claude", *args]
    try:
        proc = subprocess.run(
            cmd, cwd=cwd, input=prompt, capture_output=True, text=True,
            encoding="utf-8", errors="replace", timeout=TIMEOUT_POR_CASO,
        )
    except subprocess.TimeoutExpired:
        return {"error": f"timeout tras {TIMEOUT_POR_CASO}s", "tool_uses": [], "texto": "", "crudo": ""}
    except FileNotFoundError:
        return {"error": "no se encontró el binario `claude` en el PATH", "tool_uses": [], "texto": "", "crudo": ""}

    tool_uses: list[dict] = []
    textos: list[str] = []
    for linea in proc.stdout.splitlines():
        linea = linea.strip()
        if not linea.startswith("{"):
            continue
        try:
            frame = json.loads(linea)
        except json.JSONDecodeError:
            continue
        if frame.get("type") == "assistant":
            for bloque in frame.get("message", {}).get("content", []) or []:
                if bloque.get("type") == "tool_use":
                    tool_uses.append({"name": bloque.get("name", ""), "input": bloque.get("input", {})})
                elif bloque.get("type") == "text":
                    textos.append(bloque.get("text", ""))
        elif frame.get("type") == "result" and isinstance(frame.get("result"), str):
            textos.append(frame["result"])

    salida = {"tool_uses": tool_uses, "texto": "\n".join(textos), "crudo": proc.stdout}
    if proc.returncode != 0 and not tool_uses and not textos:
        salida["error"] = f"claude salió con {proc.returncode}: {proc.stderr.strip()[:300]}"
    return salida


# ── los asserts ─────────────────────────────────────────────────────────────

def skills_invocadas(obs: dict) -> list[str]:
    """Las skills que la sesión INVOCÓ, no las que mencionó.

    Solo cuentan los tool_use de la herramienta `Skill`. Buscar el slug en cualquier
    tool_use daba falsos positivos evidentes: un `Read` de
    `kit/skills/forjar-arnes/SKILL.md` (que es exactamente lo que hace `forjar-caja` para
    orientarse) se contaba como haber invocado la skill.
    """
    fuera: list[str] = []
    for tu in obs["tool_uses"]:
        if tu["name"] != "Skill":
            continue
        # El nombre puede venir namespaced por el plugin (`arnesia-kit:forjar-arnes`).
        for clave in ("skill", "command", "name"):
            valor = tu["input"].get(clave)
            if isinstance(valor, str) and valor:
                fuera.append(valor)
                break
    return fuera


def invoco_skill(obs: dict, slug: str) -> bool:
    return any(slug in s for s in skills_invocadas(obs))


def leyo_knowhow(obs: dict) -> list[str]:
    """Los nodos de knowhow/ que la sesión abrió de verdad (via Read)."""
    leidos = []
    for tu in obs["tool_uses"]:
        ruta = str(tu["input"].get("file_path", "")).replace("\\", "/")
        if "/knowhow/" in ruta:
            leidos.append(Path(ruta).name)
    return sorted(set(leidos))


def declaro_nodos(obs: dict) -> list[str]:
    """Los nodos que la sesión NOMBRÓ en su texto.

    La skill pide declarar «NOMBRANDO cada archivo»: «ya leí los tres nodos requeridos» no
    dice cuáles y no prueba nada. Se cuentan los nombres concretos, no la palabra
    «knowhow» — nombrar los archivos ES la declaración; decir «knowhow» es incidental.
    """
    t = obs["texto"].lower()
    return [n for n in NODOS_ESPERADOS if n in t]


def validar_manifiesto(workdir: Path) -> tuple[bool, str]:
    """El arnes.l0.json contra el SUBESQUEMA arnés-level de graph.l0.

    No contra la raíz: el manifiesto es el subconjunto arnés-level (contrato
    nomenclatura-arnes.md §2) y la raíz exige `nodos`, que solo tiene el grafo exportado.
    Validar contra la raíz daría un rojo falso.
    """
    ruta = workdir / "arnes.l0.json"
    if not ruta.exists():
        # Qué SÍ escribió importa para el diagnóstico: «no está el manifiesto» y «no
        # escribió nada» son fallos distintos, y el segundo suele ser un problema del eval
        # (permisos, cwd) más que de la skill.
        creados = sorted(p.relative_to(workdir).as_posix() for p in workdir.rglob("*") if p.is_file())
        detalle = ", ".join(creados[:6]) + ("…" if len(creados) > 6 else "") if creados else "el dir quedó vacío"
        return False, f"sin arnes.l0.json en la raíz · escribió: {detalle}"
    try:
        datos = json.loads(ruta.read_text(encoding="utf-8"))
    except json.JSONDecodeError as e:
        return False, f"arnes.l0.json no es JSON válido: {e}"

    sub = json.loads(SCHEMA.read_text(encoding="utf-8"))["properties"]["arnes"]
    errores = sorted(Draft202012Validator(sub).iter_errors(datos), key=lambda e: list(e.path))
    if errores:
        e = errores[0]
        campo = "/".join(str(p) for p in e.path) or "(raíz)"
        return False, f"{len(errores)} error(es) de schema · primero en `{campo}`: {e.message[:110]}"
    return True, f"válido · {len(datos.get('fases', []))} fases · {len(datos.get('spine', {}).get('estados', []))} estados"


def spine_coherente(workdir: Path) -> tuple[bool, str]:
    """Ningún estado inventado: transiciones, terminales e inicial ⊆ estados declarados."""
    ruta = workdir / "arnes.l0.json"
    if not ruta.exists():
        return False, "sin manifiesto"
    try:
        spine = json.loads(ruta.read_text(encoding="utf-8")).get("spine") or {}
    except json.JSONDecodeError:
        return False, "manifiesto ilegible"
    estados = set(spine.get("estados") or [])
    if not estados:
        return False, "el spine no declara estados"
    fuera = set()
    for t in spine.get("transiciones") or []:
        fuera |= {t.get("de"), t.get("a")} - estados
    fuera |= set(spine.get("terminales") or []) - estados
    if spine.get("inicial") and spine["inicial"] not in estados:
        fuera.add(spine["inicial"])
    fuera.discard(None)
    if fuera:
        return False, f"estados fuera del spine declarado: {sorted(fuera)}"
    return True, f"{len(estados)} estados, sin inventados"


def evaluar(caso: dict, obs: dict, workdir: Path) -> list[tuple[str, bool, str]]:
    """(assert, pasa, detalle) por cada expectativa del caso."""
    esp = caso.get("espera", {})
    res: list[tuple[str, bool, str]] = []

    # Se listan siempre: un fallo de ruteo se diagnostica viendo QUÉ invocó, no solo que
    # no invocó lo esperado.
    invocadas = skills_invocadas(obs) or ["(ninguna)"]

    if "invoca" in esp:
        slug = esp["invoca"]
        ok = invoco_skill(obs, slug)
        res.append((f"invoca `{slug}`", ok, "sí" if ok else f"invocó: {', '.join(invocadas)}"))

    if "no_invoca" in esp:
        slug = esp["no_invoca"]
        ok = not invoco_skill(obs, slug)
        res.append((f"NO invoca `{slug}`", ok, f"invocó: {', '.join(invocadas)}"))

    if "lee_knowhow" in esp:
        leidos = leyo_knowhow(obs)
        ok = bool(leidos) if esp["lee_knowhow"] else not leidos
        detalle = ", ".join(leidos) if leidos else "ningún nodo abierto"
        res.append(("abre nodos de knowhow/" if esp["lee_knowhow"] else "NO abre knowhow/", ok, detalle))

    if esp.get("declara_nodos"):
        nombrados = declaro_nodos(obs)
        # Dos o más: nombrar uno solo de pasada no es haber declarado la lectura del paso 1.
        ok = len(nombrados) >= 2
        res.append((
            "declara qué nodos abrió", ok,
            f"nombró: {', '.join(nombrados)}" if nombrados else "no nombró ninguno («leí los nodos requeridos» no cuenta)",
        ))

    if esp.get("escribe_manifiesto"):
        ok, detalle = validar_manifiesto(workdir)
        res.append(("arnes.l0.json válido", ok, detalle))

    if esp.get("spine_coherente"):
        ok, detalle = spine_coherente(workdir)
        res.append(("spine sin estados inventados", ok, detalle))

    for termino in esp.get("menciona", []):
        ok = termino.lower() in obs["texto"].lower()
        res.append((f"menciona «{termino}»", ok, "sí" if ok else "no aparece en la respuesta"))

    return res


# ── orquestación ────────────────────────────────────────────────────────────

def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--caso", help="corre solo ese id")
    ap.add_argument("--conservar", action="store_true", help="no borra los temps (para inspeccionarlos)")
    args = ap.parse_args()

    casos = yaml.safe_load(CASOS.read_text(encoding="utf-8"))["casos"]
    if args.caso:
        casos = [c for c in casos if c["id"] == args.caso]
        if not casos:
            print(f"no existe el caso {args.caso!r}", file=sys.stderr)
            return 2

    base = Path(tempfile.mkdtemp(prefix="eval-forjar-arnes-"))
    inj = armar_inyeccion(base / "inyeccion")
    print(f"inyección armada en {base / 'inyeccion'}")
    print(f"  kit: {len(list((inj['kit'] / 'skills').iterdir()))} skills · knowhow: {len(list(inj['knowhow'].iterdir()))} nodos\n")

    fallados: list[str] = []
    for caso in casos:
        workdir = base / caso["id"]
        workdir.mkdir(parents=True)
        for nombre, contenido in (caso.get("semilla") or {}).items():
            (workdir / nombre).write_text(contenido, encoding="utf-8")

        etiqueta = caso["id"] + ("" if caso.get("con_kit", True) else "  (sin kit — baseline)")
        print(f"── {etiqueta} " + "─" * max(0, 62 - len(etiqueta)))
        obs = correr(caso["prompt"], workdir, flags_spawn(inj, caso.get("con_kit", True)))

        # El transcripto va FUERA del workdir: adentro contaminaría el listado de archivos
        # que inspecciona `validar_manifiesto` y haría que un dir vacío nunca lo parezca.
        # Sin él, un assert en rojo no se puede diagnosticar sin gastar otra sesión.
        logs = base / f"{caso['id']}__logs"
        logs.mkdir(exist_ok=True)
        (logs / "transcripto.txt").write_text(obs.get("texto", ""), encoding="utf-8")
        (logs / "stream.jsonl").write_text(obs.get("crudo", ""), encoding="utf-8")

        if obs.get("error"):
            print(f"   ✗ la sesión no corrió: {obs['error']}\n")
            fallados.append(caso["id"])
            continue

        if caso.get("modo") == "medicion":
            # El baseline MIDE, no juzga. Con la sesión pelada los asserts serían
            # tautológicos (no puede invocar una skill que no está cargada ni leer un dir
            # que no se inyectó): lo que aporta es el registro de qué hace el modelo solo,
            # que es contra lo que se compara el caso con kit. Nunca rompe la suite.
            manifiesto_ok, detalle_m = validar_manifiesto(workdir)
            print(f"   · skills invocadas       {', '.join(skills_invocadas(obs)) or 'ninguna'}")
            print(f"   · nodos del estándar     {', '.join(leyo_knowhow(obs)) or 'ninguno (no se inyectó)'}")
            print(f"   · arnes.l0.json          {'válido — ' if manifiesto_ok else ''}{detalle_m}")
            print("   (medición, no assert: es la vara contra la que se lee `crear-desde-cero`)\n")
            continue

        resultados = evaluar(caso, obs, workdir)
        for nombre, ok, detalle in resultados:
            print(f"   {'✓' if ok else '✗'} {nombre:<32} {detalle}")
        if not all(ok for _, ok, _ in resultados):
            fallados.append(caso["id"])
        print()

    print("═" * 72)
    if fallados:
        print(f"FALLARON {len(fallados)}/{len(casos)}: {', '.join(fallados)}")
    else:
        print(f"OK — {len(casos)}/{len(casos)} casos en verde")

    if args.conservar:
        print(f"\ntemps conservados en {base}")
    else:
        shutil.rmtree(base, ignore_errors=True)

    return 1 if fallados else 0


if __name__ == "__main__":
    raise SystemExit(main())
