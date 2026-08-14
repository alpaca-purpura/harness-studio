#!/usr/bin/env python3
"""bundle.py — espejo PORTABLE de scripts/bundle.sh (paquete 2026-08-13-compilacion-windows).

Es el camino de build en Windows (donde el `bash` del PATH de sistema suele ser el relay
roto de WSL) y funciona igual en unix. `bundle.sh` sigue siendo el camino canónico Linux
(el self-update unix lo invoca tal cual); ESTE script es el que invoca el self-update
Windows (`Build()` de os_windows.go → `python scripts/bundle.py --solo-daemon`).

Regla RF-231 (idéntica a bundle.sh): el sello de identidad vive EN EL SCRIPT DE BUNDLE,
nunca en el Makefile — el self-update corre este mismo script, así que un binario
producido por el botón «Actualizar» sella igual que uno hecho a mano. Las tres variables
(`Version` desde Cargo.toml · `Build` AAMMDDHHMM hora local · `Compilado` legible) van
por `-ldflags -X` al paquete selfupdate.

Pasos (mismos que bundle.sh):
  1/3  SPA: pnpm install --frozen-lockfile && pnpm run build   (web/ → web/dist,
       que el daemon EMBEBE — por eso corre también en modo solo-daemon)
  2/3  daemon: go build -trimpath -ldflags "<sello>" -o bin/arnesia[.exe]
       (CGO_ENABLED=0 — modernc/sqlite es puro Go; binario estático en ambos OS)
  3/3  sidecar + tauri build (SOLO modo completo; requiere rustc — si no está, error
       honesto con el prerreq, jamás un bundle a medias)

Flags: --solo-daemon (alias --daemon-only, paridad con bundle.sh) corta tras el paso 2.
"""

from __future__ import annotations

import datetime
import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
IDENT = "github.com/alpacapurpura/arnesia/internal/adapters/selfupdate"
ES_WINDOWS = os.name == "nt"
BIN = ROOT / "bin" / ("arnesia.exe" if ES_WINDOWS else "arnesia")


def fallar(msg: str) -> "None":
    print(f"bundle.py: {msg}", file=sys.stderr)
    raise SystemExit(1)


def herramienta(nombre: str) -> str:
    """Resuelve un ejecutable en PATH (shutil.which maneja PATHEXT en Windows)."""
    ruta = shutil.which(nombre)
    if not ruta:
        fallar(f"{nombre!r} no está en PATH — prerreq del bundle (ver README § Desarrollo en Windows)")
    return ruta


def correr(cmd: list[str], cwd: Path, env: dict | None = None) -> None:
    print(f"$ {' '.join(cmd)}  (cwd={cwd})")
    subprocess.run(cmd, cwd=str(cwd), env=env, check=True)


def version_de_cargo() -> str:
    cargo = ROOT / "web" / "src-tauri" / "Cargo.toml"
    m = re.search(r'^version = "(.*)"', cargo.read_text(encoding="utf-8"), re.MULTILINE)
    if not m:
        fallar(f"no pude leer version de {cargo}")
    return m.group(1)


def triple_de_rustc() -> str:
    rustc = shutil.which("rustc")
    if not rustc:
        fallar(
            "Tauri no disponible: falta rustc (toolchain "
            + ("x86_64-pc-windows-msvc + VS Build Tools" if ES_WINDOWS else "de rust")
            + "). El modo --solo-daemon no lo necesita."
        )
    out = subprocess.run([rustc, "-vV"], capture_output=True, text=True, check=True).stdout
    m = re.search(r"^host: (.+)$", out, re.MULTILINE)
    if not m:
        fallar("rustc -vV no reportó host triple")
    return m.group(1).strip()


def main() -> None:
    solo_daemon = any(a in ("--solo-daemon", "--daemon-only") for a in sys.argv[1:])
    pnpm = herramienta("pnpm")
    go = herramienta("go")

    # ── 1/3 · SPA (el daemon la embebe: corre SIEMPRE) ─────────────────────────────
    correr([pnpm, "install", "--frozen-lockfile"], cwd=ROOT / "web")
    correr([pnpm, "run", "build"], cwd=ROOT / "web")

    # ── 2/3 · daemon con sello de identidad (RF-231) ───────────────────────────────
    version = version_de_cargo()
    ahora = datetime.datetime.now()  # hora LOCAL a propósito (ver identidad.go)
    build = ahora.strftime("%y%m%d%H%M")
    compilado = ahora.strftime("%Y-%m-%d %H:%M")
    BIN.parent.mkdir(parents=True, exist_ok=True)
    ldflags = (
        f"-s -w -X '{IDENT}.Version={version}' "
        f"-X '{IDENT}.Build={build}' -X '{IDENT}.Compilado={compilado}'"
    )
    env = dict(os.environ, CGO_ENABLED="0")
    correr([go, "build", "-trimpath", "-ldflags", ldflags, "-o", str(BIN), "./cmd/arnesia"], cwd=ROOT, env=env)
    print(f"OK — daemon {version}.{build} → {BIN}")

    if solo_daemon:
        return

    # ── 3/3 · sidecar + tauri build ────────────────────────────────────────────────
    triple = triple_de_rustc()
    sidecar = ROOT / "web" / "src-tauri" / "binaries" / (
        f"arnesia-daemon-{triple}.exe" if ES_WINDOWS else f"arnesia-daemon-{triple}"
    )
    sidecar.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(BIN, sidecar)
    print(f"OK — sidecar → {sidecar}")
    correr([pnpm, "exec", "tauri", "build"], cwd=ROOT / "web")


if __name__ == "__main__":
    main()
