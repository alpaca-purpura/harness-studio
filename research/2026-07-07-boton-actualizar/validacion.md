# Validación REAL — FASE 3 (2026-07-07/08)

> Este archivo nace como el «commit trivial» del guion de validación (PROMPT §FASE 3):
> el binario instalado corre la huella del commit anterior; ESTE commit crea la delta
> que el botón «Actualizar» debe compilar, instalar y confirmar tras el reinicio.
> Se completa con los resultados reales al cierre de la fase.

## Guion

1. `scripts/bundle.sh --daemon-only` → `install -m755 bin/arnesia ~/.local/bin/` (migración única) ✔
2. `~/.local/bin/arnesia serve --repo <repo>` — daemon INSTALADO corriendo ✔
3. este commit → click «Actualizar desde el repo» → checklist real → re-exec → huella NUEVA
4. contracasos: build roto · ya-al-día · 409 doble click · no-escribible · sin --repo
5. gates curl: Host/Origin 403 ✔ (RF-106)

(resultados al cierre)
