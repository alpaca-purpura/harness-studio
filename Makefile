# Makefile — build local de instaladores de escritorio ArnesIA (Tauri).
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

SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c

ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
CARGO_TOML := $(ROOT)/web/src-tauri/Cargo.toml
TAURI_CONF := $(ROOT)/web/src-tauri/tauri.conf.json
WEB_PKG := $(ROOT)/web/package.json
BUNDLE_DIR := $(ROOT)/web/src-tauri/target/release/bundle
INSTALL_DIR := $(ROOT)/instaladores

CURRENT_VERSION := $(shell sed -n 's/^version = "\(.*\)"/\1/p' $(CARGO_TOML))
NEXT_VERSION := $(shell echo '$(CURRENT_VERSION)' | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}')

.PHONY: version bump-patch installer

version: ## imprime la version actual (fuente de verdad: Cargo.toml)
	@echo "$(CURRENT_VERSION)"

bump-patch: ## bumpea el patch en Cargo.toml + tauri.conf.json + package.json (sync los 3)
	@test -n "$(CURRENT_VERSION)" || { echo "no pude leer version de $(CARGO_TOML)" >&2; exit 1; }
	@echo "version: $(CURRENT_VERSION) -> $(NEXT_VERSION)"
	@sed -i 's/^version = "$(CURRENT_VERSION)"/version = "$(NEXT_VERSION)"/' "$(CARGO_TOML)"
	@sed -i '0,/"version":/s/"version": "[^"]*"/"version": "$(NEXT_VERSION)"/' "$(TAURI_CONF)"
	@sed -i '0,/"version":/s/"version": "[^"]*"/"version": "$(NEXT_VERSION)"/' "$(WEB_PKG)"

installer: bump-patch ## genera instalador (.deb/.rpm/.AppImage) versionado en instaladores/vX.Y.Z/
	@echo "== build instalador v$(NEXT_VERSION) =="
	@rm -rf "$(BUNDLE_DIR)"
	@bash "$(ROOT)/scripts/bundle.sh"
	@mkdir -p "$(INSTALL_DIR)/v$(NEXT_VERSION)"
	@find "$(BUNDLE_DIR)" -type f \( -name '*.deb' -o -name '*.rpm' -o -name '*.AppImage' -o -name '*.dmg' -o -name '*.msi' -o -name '*.exe' \) -exec cp {} "$(INSTALL_DIR)/v$(NEXT_VERSION)/" \;
	@test -n "$$(ls -A "$(INSTALL_DIR)/v$(NEXT_VERSION)" 2>/dev/null)" || { echo "bundle no genero ningun instalador reconocido en $(BUNDLE_DIR) — revisar tauri.conf.json bundle.targets" >&2; exit 1; }
	@(cd "$(INSTALL_DIR)/v$(NEXT_VERSION)" && sha256sum * > checksums.txt)
	@echo "OK — instalador v$(NEXT_VERSION) en $(INSTALL_DIR)/v$(NEXT_VERSION)/"
	@echo "recorda commitear el bump de version (Cargo.toml/tauri.conf.json/package.json)"
