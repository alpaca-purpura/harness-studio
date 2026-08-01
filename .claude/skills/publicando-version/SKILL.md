---
name: publicando-version
description: "Publica una versión de ArnesIA COMPLETA: changelog → bump → instalador (.deb/.rpm/.AppImage con FE+BE frescos) → override dev-sync → tag → verificación contra el binario vivo. Usar SIEMPRE que haya que publicar versión, bumpear, generar/regenerar instalador o cerrar un release — nunca correr bump-* o bundle.sh sueltos de memoria. Cubre también la reparación «versión bumpeada que quedó sin instalador» (installer-actual)."
contract:
  caja: false
  rol: apoyo-release
---

# Publicando una versión (skill de proceso)

> Nace del incidente 2026-08-01 (deudas de dogfood, DD-4): se publicó v0.7.0 con
> `make bump-minor` + `make dev-sync` y la versión quedó **sin instalador** — y ya no se
> podía generar, porque `make installer` re-bumpeaba patch y el bump falla con
> `[Sin publicar]` vacío. El Makefile se desacopló ese día; esta skill es el PROCESO que
> hace imposible repetirlo.

## Regla de oro

**UN comando por release.** Bump y empaquetado jamás se separan a mano:

| Qué cambió | Comando único |
|---|---|
| Fix compatible | `make installer` (bump Z + bundle + `instaladores/vX.Y.Z/` + dev-sync) |
| Superficie nueva compatible | `make installer-minor` |
| Rompe algo que el usuario usaba | `make installer-major` |
| **Reparación**: versión YA bumpeada sin instalador | `make installer-actual` (no bumpea; falla si `vX.Y.Z/` ya tiene archivos) |

Criterio fino de qué bump: `docs/architecture/conventions/versionado.md` §changelog-y-bump.

## Secuencia completa

1. **Changelog primero** — ya debió anotarse EN EL TURNO de cada cambio
   (`python3 scripts/changelog.py add …`). Verificar: `make changelog` (imprime
   `[Sin publicar]`). Vacío ⇒ el bump va a fallar y ESO ES CORRECTO — no se publica una
   versión que no dice qué trae. Jamás inventar una entrada de relleno para destrabar.
2. **El comando único** de la tabla. Qué hace por dentro (`_installer-build`):
   - `scripts/bundle.sh` COMPLETO: paso 1 = `pnpm run build` (vite → `web/dist`, **el FE
     SIEMPRE se recompila** — no existe instalador con FE stale) → paso 2 = daemon Go con
     SPA+doctrina+kit **embebidos** + sello `X.Y.Z.AAMMDDHHMM` → paso 3 = shell Tauri.
   - Copia `.deb/.rpm/.AppImage` a `instaladores/vX.Y.Z/` + `checksums.txt`. **Nunca pisa**
     una generación publicada.
   - Si existe `~/.local/bin/arnesia` (override de self-update): lo reemplaza atómico con el
     MISMO binario del instalador (mismo sello) y mata el daemon viejo — cierra el gotcha
     2026-07-25 («instalé el .deb y no cambió nada») sin depender de memoria.
3. **Git**: commitear bump (3 manifiestos) + `CHANGELOG.md` + `instaladores/vX.Y.Z/` +
   docs del turno; `git tag vX.Y.Z`; push con `--tags`. Los manifiestos JAMÁS se editan a
   mano (regla dura de CLAUDE.md).
4. **Verificar, no asumir** (honestidad del repo):
   - relanzar daemon → `curl -s localhost:4200/api/version` → `version` = `X.Y.Z.<sello>`
     del build recién hecho y `huella` = el commit pusheado, `sucio:false`;
   - `instaladores/vX.Y.Z/` tiene los 3 artefactos + checksums;
   - si el cambio era de FE: abrir la app/`:4200` y ver la superficie nueva (el sello del
     daemon NO prueba qué SPA quedó dentro — la prueba es mirarla).

## Gotchas que ya mordieron

- `pkill -f "arnesia serve"` **matchea tu propio shell** si el patrón viaja en el comando →
  siempre el truco del corchete: `pkill -f "[a]rnesia serve"`.
- `CURRENT_VERSION` del Makefile se evalúa AL PARSEAR: tras un bump, el valor fresco solo
  existe en una invocación nueva de make (por eso `_installer-build` corre en sub-make).
- El shell Tauri instalado PREFIERE `~/.local/bin/arnesia` sobre el sidecar del paquete —
  un `.deb` nuevo sin sincronizar el override es invisible en esta máquina (auto-resuelto
  por el paso 2, pero saberlo explica «instalé y no pasó nada» en máquinas de dev).
- `make installer` en un árbol donde ya se corrió `bump-minor` a mano NO repara nada:
  fabricaría un patch hueco. La reparación es `make installer-actual`.
