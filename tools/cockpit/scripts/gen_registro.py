#!/usr/bin/env python3
"""gen_registro.py — adapter de la banda RUNTIME del registro de la torre (DH-04/D2,
espejo del corte 3 bandas de OBS-15 · RN-04..RN-07).

El join: CURADO (prenter/sistemas/sistemas.yaml — la EMPRESA, I-39; o el fixture de la
fábrica mientras el gate de prenter siga rojo, RN-07) × config del OPERADOR
(~/.config/prenter/devhub.yaml → torre.workspaces, las rutas locales que el curado jamás
lleva) → ~/.cockpit/cockpit.yaml en el wire format ACTUAL (registry.go no cambia para
leerlo; `nombre` y `gate_check` viajan como campos aditivos — RN-06, red:
TestRegistryWireFormatRoundTrip).

Reemplaza la FUENTE del registro vivo (el huérfano pre-I-39
chris-corp/gen_cockpit_registry.py queda obsoleto de facto — RN-08, retiro aparte).
El archivo anterior se respalda en cockpit.yaml.bak antes de sobrescribir.

El adapter SE NIEGA a emitir lo que el gate rechazaría (importa validate() del check —
patrón SC-30). Sistema sin workspace en la config NO desaparece: se emite con path vacío
y active:false — la torre lo muestra como fila `no-medido` (RN-17, el registro dice que
debería existir y eso ES señal).

Uso:    python3 products/devhub/scripts/gen_registro.py
Config: ~/.config/prenter/devhub.yaml → bloque `torre:` {registro, workspaces} (operador-only;
el gate NO corre esto — check_registro.py valida el fixture siempre y el curado si existe).
"""
import os
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
from check_registro import validate  # un solo validador, dos llamadores (SC-30)

CONFIG_OPERADOR = os.path.expanduser("~/.config/prenter/devhub.yaml")
REGISTRO_VIVO = os.path.expanduser("~/.cockpit/cockpit.yaml")

CABECERA = """\
# GENERADO por prenter-harness/products/devhub/scripts/gen_registro.py — NO editar a mano.
# Fuente: curado de sistemas (torre.registro) × ~/.config/prenter/devhub.yaml (torre.workspaces).
# SPEC: products/devhub/specs/torre-read-only.md §2 (RN-04..RN-07 · DH-04/D2).
"""


def cargar_config():
    import yaml
    if not os.path.isfile(CONFIG_OPERADOR):
        sys.exit(f"✗ falta la config del operador ({CONFIG_OPERADOR}) — el join es operador-only; "
                 "declara torre: {registro, workspaces} (RN-04)")
    cfg = (yaml.safe_load(open(CONFIG_OPERADOR, encoding="utf-8")) or {}).get("torre") or {}
    for req in ("registro", "workspaces"):
        if req not in cfg:
            sys.exit(f"✗ la config no declara torre.{req}")
    return cfg


def cargar_curado(path):
    import yaml
    p = os.path.expanduser(path)
    if not os.path.isfile(p):
        sys.exit(f"✗ registro curado ausente: {p} — la banda CURADA vive en la EMPRESA (I-39); "
                 "mientras el gate de prenter siga rojo, apunta torre.registro al fixture de la fábrica (RN-07)")
    return yaml.safe_load(open(p, encoding="utf-8")) or {}


def construir_registry(curado, workspaces):
    """El join: cada sistema del curado → un project del wire format actual."""
    projects = []
    for s in curado.get("sistemas") or []:
        slug = s["slug"]
        ws = workspaces.get(slug)
        ws = os.path.expanduser(ws) if ws else ""
        if not ws:
            print(f"  ⚠ sistema '{slug}' sin workspace en la config — la torre lo mostrará no-medido (RN-17)")
        proj = {"name": slug, "path": ws, "active": bool(ws)}
        proj["nombre"] = s["nombre"]            # aditivo (RN-06) — display de la fila
        proj["kind"] = s["tipo"]                # reusa el campo kind existente del wire
        if s.get("gate_check"):
            proj["gate_check"] = s["gate_check"]  # aditivo (RN-06) — comando del gate de fábrica
        projects.append(proj)

    for slug in sorted(set(workspaces) - {s.get("slug") for s in curado.get("sistemas") or []}):
        print(f"  ⚠ workspace '{slug}' en la config sin sistema en el curado — no se emite (curado = SSoT)")
    return {"projects": projects}


def main():
    import yaml
    cfg = cargar_config()
    curado = cargar_curado(cfg["registro"])

    # SC-30: el adapter se niega a emitir lo que el gate rechazaría (mismo validate del check)
    errs = validate(curado, "curado")
    if errs:
        for m in errs:
            print(f"  ERR   {m}")
        sys.exit("✗ el registro NO se emite — corrige el curado")

    reg = construir_registry(curado, cfg.get("workspaces") or {})
    body = CABECERA + yaml.safe_dump(reg, allow_unicode=True, sort_keys=False, width=1000)

    os.makedirs(os.path.dirname(REGISTRO_VIVO), exist_ok=True)
    if os.path.isfile(REGISTRO_VIVO):
        actual = open(REGISTRO_VIVO, encoding="utf-8").read()
        if actual != body:
            open(REGISTRO_VIVO + ".bak", "w", encoding="utf-8").write(actual)
            print(f"     (registro anterior respaldado en {REGISTRO_VIVO}.bak)")
    open(REGISTRO_VIVO, "w", encoding="utf-8").write(body)

    n = len(reg["projects"])
    con_gate = sum(1 for p in reg["projects"] if p.get("gate_check"))
    print(f"ok · registro emitido → {REGISTRO_VIVO}")
    print(f"     {n} sistema(s) · {con_gate} con gate de fábrica declarado · fuente: {os.path.expanduser(cfg['registro'])}")


if __name__ == "__main__":
    main()
