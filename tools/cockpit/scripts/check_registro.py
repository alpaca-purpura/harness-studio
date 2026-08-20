#!/usr/bin/env python3
"""check_registro.py — valida el registro CURADO de sistemas de la torre contra su contrato
(SPEC torre-read-only §2, RN-04..RN-07 · DH-04/D2 — espejo del corte 3 bandas de OBS-15).

Tres cosas, en orden de peligro:
  1. RN-04 (I-39): el curado lleva slugs/refs, JAMÁS rutas — llaves path/workspace/ruta
     prohibidas; ningún valor string puede ser ruta absoluta ni empezar con `~`.
  2. Forma: version int ≥ 1 · sistemas no vacío · slug kebab único · nombre · tipo ∈ enum ·
     ref requerido si tipo=engagement · llaves desconocidas prohibidas.
  3. RN-05: gate_check (si existe) es string no vacío — el comando es relativo al workspace.

Gate oportunista (espejo RN-23 de flota): el FIXTURE commiteado se valida SIEMPRE (el
validador se ejercita en todo clone); el curado REAL (torre.registro de la config del
operador) SOLO si existe — su ausencia jamás rompe el gate. Corre en gen_all.py --check.

gen_registro.py IMPORTA validate() de aquí (un solo validador, dos llamadores: el adapter
se niega a emitir lo que el gate rechazaría — patrón SC-30).
"""
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
FIXTURE = os.path.join(HERE, "testdata", "sistemas-fixture.yaml")
CONFIG_OPERADOR = os.path.expanduser("~/.config/prenter/devhub.yaml")

TIPOS = {"fabrica", "empresa", "engagement"}
LLAVES = {"slug", "nombre", "tipo", "gate_check", "ref"}
SLUG_RE = re.compile(r"^[a-z0-9][a-z0-9-]*$")


def _es_ruta(v):
    return isinstance(v, str) and (v.startswith("/") or v.startswith("~"))


def validate(curado, etiqueta="curado"):
    """Valida un registro curado de sistemas. Devuelve lista de errores (vacía = conforme)."""
    errs = []
    e = lambda m: errs.append(f"{etiqueta}: {m}")

    if not isinstance(curado, dict):
        e("el curado no es un mapa YAML")
        return errs
    if not isinstance(curado.get("version"), int) or curado["version"] < 1:
        e("version ausente o inválida (int ≥ 1 requerido)")
    sistemas = curado.get("sistemas")
    if not isinstance(sistemas, list) or not sistemas:
        e("sistemas ausente o vacío (requerido)")
        return errs

    slugs = set()
    for i, s in enumerate(sistemas):
        if not isinstance(s, dict):
            e(f"sistemas[{i}] no es un mapa")
            continue
        slug = s.get("slug")
        quien = slug or f"sistemas[{i}]"
        if not slug or not SLUG_RE.match(str(slug)):
            e(f"{quien}: slug ausente o no-kebab (patrón ^[a-z0-9][a-z0-9-]*$)")
        elif slug in slugs:
            e(f"{quien}: slug duplicado (LA LLAVE del registro)")
        slugs.add(slug)
        if not s.get("nombre"):
            e(f"{quien}: nombre ausente (requerido)")
        if s.get("tipo") not in TIPOS:
            e(f"{quien}: tipo '{s.get('tipo')}' fuera del enum {sorted(TIPOS)}")
        if s.get("tipo") == "engagement" and not s.get("ref"):
            e(f"{quien}: engagement sin ref al CRM (requerido — I-39)")
        gc = s.get("gate_check")
        if gc is not None and (not isinstance(gc, str) or not gc.strip()):
            e(f"{quien}: gate_check presente pero vacío (o declara el comando o quita la llave → sin-gate)")
        for k in set(s) - LLAVES:
            e(f"{quien}: llave desconocida '{k}' — el curado lleva slugs/refs, jamás rutas (RN-04)")
        for k, v in s.items():
            if _es_ruta(v):
                e(f"{quien}: {k} parece una ruta ('{v}') — las rutas viven en la config del operador (RN-04)")
    return errs


def _curado_real_path():
    """Ruta del curado real (torre.registro de la config del operador) — None si no hay config."""
    try:
        import yaml
        if not os.path.isfile(CONFIG_OPERADOR):
            return None
        cfg = (yaml.safe_load(open(CONFIG_OPERADOR, encoding="utf-8")) or {}).get("torre") or {}
        reg = cfg.get("registro")
        return os.path.expanduser(reg) if reg else None
    except Exception:
        return None


def main():
    import yaml
    errors = []

    # el fixture — SIEMPRE (ejercita el validador en todo clone)
    if not os.path.isfile(FIXTURE):
        errors.append("testdata/sistemas-fixture.yaml ausente — el gate perdería los dientes (RN-07)")
    else:
        fx = yaml.safe_load(open(FIXTURE, encoding="utf-8"))
        errors += validate(fx, "fixture")

    # el curado real — SOLO si existe (su ausencia jamás rompe el gate; gate de prenter aparte)
    partes = ["fixture"]
    rp = _curado_real_path()
    if rp and os.path.isfile(rp):
        real = yaml.safe_load(open(rp, encoding="utf-8"))
        errors += validate(real, "curado-real")
        partes.append(f"curado real ({len((real or {}).get('sistemas') or [])} sistemas)")

    if errors:
        for m in errors:
            print(f"  ERR   registro-sistemas: {m}")
        return 1
    print(f"ok · registro de sistemas conforme (RN-04 + forma + RN-05) — validado: {' + '.join(partes)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
