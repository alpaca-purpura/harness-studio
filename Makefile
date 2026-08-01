# Makefile — build local de instaladores de escritorio ArnesIA (Tauri).
#
# ⚠ EL BUMP EXIGE CHANGELOG (2026-07-26, paquete versionado-y-changelog-metodologicos, VC-D2):
# `make bump-patch|bump-minor|bump-major` → scripts/bump.sh → valida CHANGELOG.md ANTES de tocar
# ningún manifiesto y, al bumpear, promueve `[Sin publicar]` a `## [X.Y.Z] — fecha`. Con
# `[Sin publicar]` vacía el bump FALLA: no se publica una versión que no dice qué cambió.
# Entrada nueva, en el mismo turno en que construís:
#   python3 scripts/changelog.py add Agregado|Cambiado|Corregido|Eliminado|... "qué cambió"
#
# `make installer` bumpea el patch (Z de X.Y.Z) en la fuente de verdad de versión
# (web/src-tauri/Cargo.toml — Tauri cae a este valor si tauri.conf.json lo omite, pero
# igual lo sincronizamos explícito porque así lo recomienda la doc oficial de Tauri 2),
# sincroniza tauri.conf.json + web/package.json para que no haya drift entre los 3, corre
# el bundle completo (scripts/bundle.sh, el mismo que ya arma .deb/.AppImage/.rpm) y copia
# los instaladores generados a instaladores/vX.Y.Z/ — versionado, nunca pisa una
# generación anterior.
#
# Convención de versión (coherente con docs/product/stories/2026-07-15-instalador-
# publico-licencias-org/decisiones.md IL-D9): semver plano "X.Y.Z" en los manifiestos, SIN
# prefijo "v" — formato exacto que exige Keygen para `Release.version`. La carpeta de
# salida sí lleva prefijo "v" (mismo estilo que los tags de `.goreleaser.yaml`). Cuando
# exista un canal real (dev/beta/rc, alineado a los 5 canales de Keygen) se agrega como
# sufijo de prerelease (X.Y.Z-dev.N) recién en ese momento — no antes, no hay canal hoy.
#
# No toca git: el bump queda en el working tree para que el operador lo revise/commitee.
#
# ⚠ GOTCHA (2026-07-25, incidente real): `make installer` arma el .deb/.rpm/.AppImage con un
# sidecar fresco, pero el shell Tauri instalado (web/src-tauri/src/lib.rs, override_local)
# PREFIERE SIEMPRE ~/.local/bin/arnesia por sobre ese sidecar si ese archivo existe (así el
# self-update sin sudo puede pisarse a sí mismo tras la instalación inicial). Si ya migraste a
# self-update alguna vez, instalar un .deb nuevo NO cambia nada visible hasta que también
# corras `make dev-sync` — ver docs/architecture/conventions/versionado.md.

SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c

ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CARGO_TOML := $(ROOT)/web/src-tauri/Cargo.toml
TAURI_CONF := $(ROOT)/web/src-tauri/tauri.conf.json
WEB_PKG := $(ROOT)/web/package.json
BUNDLE_DIR := $(ROOT)/web/src-tauri/target/release/bundle
INSTALL_DIR := $(ROOT)/instaladores
DEV_DAEMON := $(HOME)/.local/bin/arnesia

CURRENT_VERSION := $(shell sed -n 's/^version = "\(.*\)"/\1/p' $(CARGO_TOML))
NEXT_VERSION := $(shell echo '$(CURRENT_VERSION)' | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}')
NEXT_MINOR := $(shell echo '$(CURRENT_VERSION)' | awk -F. '{printf "%d.%d.0", $$1, $$2+1}')
NEXT_MAJOR := $(shell echo '$(CURRENT_VERSION)' | awk -F. '{printf "%d.0.0", $$1+1}')

.PHONY: version bump-patch bump-minor bump-major changelog installer installer-minor installer-major installer-actual _installer-build dev-sync

version: ## imprime la version actual (fuente de verdad: Cargo.toml)
	@echo "$(CURRENT_VERSION)"

changelog: ## imprime lo que hoy iría en la próxima versión ([Sin publicar])
	@python3 "$(ROOT)/scripts/changelog.py" sin-publicar

# Los 3 bumps pasan por scripts/bump.sh — punto único, con el changelog como CONDICIÓN: si
# [Sin publicar] está vacía el bump aborta sin tocar ningún manifiesto (VC-D2). Criterio de cuál
# usar: docs/architecture/conventions/versionado.md §changelog-y-bump.
bump-patch: ## fix compatible: bumpea Z + promueve el changelog (falla si no hay entradas)
	@bash "$(ROOT)/scripts/bump.sh" "$(NEXT_VERSION)"

bump-minor: ## superficie nueva compatible: bumpea Y, resetea Z + promueve el changelog
	@bash "$(ROOT)/scripts/bump.sh" "$(NEXT_MINOR)"

bump-major: ## rompe algo que el usuario ya usaba: bumpea X + promueve el changelog
	@bash "$(ROOT)/scripts/bump.sh" "$(NEXT_MAJOR)"

# ── Instalador desacoplado del bump (DD-4, 2026-08-01) ────────────────────────────────────
#
# El acople viejo (`installer: bump-patch` como ÚNICA vía) tenía un agujero real: publicar con
# `make bump-minor` dejaba el changelog promovido y VACÍO, y entonces `make installer` (que
# re-bumpea patch) FALLABA — la versión recién publicada se quedaba SIN instalador, o peor,
# fabricaba un X.Y.Z+1 hueco. Ahora: el empaquetado es un paso propio (`_installer-build`,
# versión ACTUAL de los manifiestos) y CADA sabor de bump tiene su target de release completo.
# Regla: UN comando por release — bump y empaquetado nunca se separan a mano.
#
#   make installer         → fix compatible:      bump-patch + bundle + instaladores/vX.Y.Z/
#   make installer-minor   → superficie nueva:    bump-minor + bundle + instaladores/vX.Y.0/
#   make installer-major   → breaking:            bump-major + bundle + instaladores/vX.0.0/
#   make installer-actual  → la versión YA bumpeada (reparación: un release que quedó sin
#                            instalador). No bumpea nada; falla si vX.Y.Z/ ya tiene archivos.
#
# `_installer-build` corre en sub-make a propósito: CURRENT_VERSION se evalúa al parsear, y
# tras un bump el valor fresco solo existe en una invocación nueva.

installer: bump-patch ## release patch completo: bump Z + bundle + instaladores/vX.Y.Z/ + dev-sync
	@$(MAKE) _installer-build

installer-minor: bump-minor ## release minor completo: bump Y + bundle + instaladores/vX.Y.0/ + dev-sync
	@$(MAKE) _installer-build

installer-major: bump-major ## release major completo: bump X + bundle + instaladores/vX.0.0/ + dev-sync
	@$(MAKE) _installer-build

installer-actual: ## empaqueta la versión ACTUAL ya bumpeada (release que quedó sin instalador)
	@$(MAKE) _installer-build

_installer-build: # interno — bundle completo (SPA vite + daemon con SPA/doctrina/kit embebidos + shell Tauri) para CURRENT_VERSION
	@test -z "$$(ls -A "$(INSTALL_DIR)/v$(CURRENT_VERSION)" 2>/dev/null)" || { echo "instaladores/v$(CURRENT_VERSION)/ ya tiene archivos — una generación publicada NUNCA se pisa (bumpeá primero)" >&2; exit 1; }
	@echo "== build instalador v$(CURRENT_VERSION) =="
	@rm -rf "$(BUNDLE_DIR)"
	@bash "$(ROOT)/scripts/bundle.sh"
	@mkdir -p "$(INSTALL_DIR)/v$(CURRENT_VERSION)"
	@find "$(BUNDLE_DIR)" -type f \( -name '*.deb' -o -name '*.rpm' -o -name '*.AppImage' -o -name '*.dmg' -o -name '*.msi' -o -name '*.exe' \) -exec cp {} "$(INSTALL_DIR)/v$(CURRENT_VERSION)/" \;
	@test -n "$$(ls -A "$(INSTALL_DIR)/v$(CURRENT_VERSION)" 2>/dev/null)" || { echo "bundle no genero ningun instalador reconocido en $(BUNDLE_DIR) — revisar tauri.conf.json bundle.targets" >&2; exit 1; }
	@(cd "$(INSTALL_DIR)/v$(CURRENT_VERSION)" && sha256sum * > checksums.txt)
	@echo "OK — instalador v$(CURRENT_VERSION) en $(INSTALL_DIR)/v$(CURRENT_VERSION)/"
	@echo "recorda commitear el bump de version (Cargo.toml/tauri.conf.json/package.json) + CHANGELOG.md + instaladores/"
	@# Cierre del GOTCHA 2026-07-25: el shell instalado prefiere SIEMPRE ~/.local/bin/arnesia
	@# (override de self-update) sobre el sidecar del .deb nuevo. El bundle de arriba YA dejó
	@# ese MISMO daemon fresco en bin/arnesia (mismo sello) — se sincroniza acá, automático,
	@# en vez de confiar en que alguien recuerde `make dev-sync`.
	@if [ -f "$(DEV_DAEMON)" ]; then \
		install -m755 "$(ROOT)/bin/arnesia" "$(DEV_DAEMON).new"; \
		mv -f "$(DEV_DAEMON).new" "$(DEV_DAEMON)"; \
		pkill -f "$(shell dirname $(DEV_DAEMON))/[a]rnesia serve" 2>/dev/null || true; \
		echo "OK — override local $(DEV_DAEMON) sincronizado con el MISMO build del instalador (daemon corriendo detenido; el shell lo relanza)"; \
	fi

dev-sync: ## sincroniza ~/.local/bin/arnesia (override local de self-update) con un build fresco del daemon
	@if [ ! -f "$(DEV_DAEMON)" ]; then \
		echo "no hay override local en $(DEV_DAEMON) — nada que sincronizar (el shell Tauri usa el sidecar empaquetado normal)"; \
		exit 0; \
	fi
	@echo "== dev-sync: build daemon-only + reemplazo atómico de $(DEV_DAEMON) =="
	@bash "$(ROOT)/scripts/bundle.sh" --daemon-only
	@install -m755 "$(ROOT)/bin/arnesia" "$(DEV_DAEMON).new"
	@mv -f "$(DEV_DAEMON).new" "$(DEV_DAEMON)"
	@# `pkill -f` matchea líneas de comando COMPLETAS — incluida la del shell que corre esta
	@# misma receta, que contiene el patrón literal. Se auto-mataba: make reportaba
	@# "Terminado" en cada corrida aunque el sync hubiera terminado bien (y el `|| true` no
	@# ayuda: el shell recibe SIGTERM, no sale con código). El truco del corchete evita el
	@# self-match: `[a]rnesia` matchea "arnesia" pero el cmdline propio dice "[a]rnesia".
	@pkill -f "$(shell dirname $(DEV_DAEMON))/[a]rnesia serve" 2>/dev/null || true
	@echo "OK — $(DEV_DAEMON) actualizado. Si la app ya estaba abierta, cerrala y volvé a abrirla (el shell la relanza sola)."
