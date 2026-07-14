# fix: repo configurable para self-update (paquete de trabajo)

> Origen: el operador reporta que una instalación empaquetada (Tauri/.deb) no puede usar
> el botón «Actualizar» — el sidecar arranca el daemon sin `--repo`/`ARNESIA_REPO` y no
> hay forma de fijarlo desde la app. Investigación confirmó: gap real, scoped-out
> honesto en el paquete `2026-07-07-boton-actualizar` (RF-103, KIT-06), no un bug ciego.
> Disciplina METODOLOGIA §10: decisiones → spec → 🧑‍⚖️ → código → PARIDAD.

## Rumbo firmado (2026-07-13)

Separar CONFIGURAR (nuevo, persistido, validado, gated a picker nativo Tauri) de
DISPARAR (`POST /api/self-update` sin cambios — RF-106 intacto). Detalle en
`decisiones.md` (#1–#5, FIRMADAS).

## Flujo y gates

1. **Decisiones** (`decisiones.md`) — FIRMADAS 2026-07-13.
2. **Spec** (`spec.md`, RF-108..111) — FIRMADO 2026-07-13 (mismo turno de las
   decisiones — bugfix, no arrastra mockup propio; reusa el mockup/UI ya firmados del
   paquete boton-actualizar, solo agrega el botón «Elegir carpeta…» al estado sinRepo).
3. **Implementación** — Go primero (store + puerto + adapter + usecase + endpoint +
   wiring), luego FE (client + UpdateCard + AjustesView) + Tauri (plugin dialog).
4. **Paridad** (`PARIDAD.md`) → gate final. El lado Go se verifica E2E (daemon real,
   headless). El lado Tauri (picker nativo) requiere sesión con display — gate humano
   explícito, no fingido headless.

## Estado

- [x] decisiones FIRMADAS (#1–#5)
- [x] spec.md (RF-108..111) FIRMADO
- [x] implementación Go (store + puerto + adapter + usecase + endpoint + wiring + tests) — `go test ./...` verde, E2E vivo (daemon real: sin repo · PUT inválido/válido · persistencia tras reinicio · precedencia flag>persistido)
- [x] implementación FE (client + UpdateCard + AjustesView + stories) — typecheck/lint/fsd/depcruise/stylelint/vitest (93/93) verdes
- [x] implementación Tauri (plugin-dialog + capability + Cargo.toml) — `cargo check` verde, permiso `dialog:allow-open` validado al build
- [x] capabilities actualizados (self-update-sin-sudo.yaml + boton-actualizar.yaml) — `cap_doctor.py` + `TestCapabilityCoverage` verdes
- [x] PARIDAD.md llenado (RF-108/109/111 ✅ en vivo; RF-110 🔶 código+stories, falta interacción real del picker)
- [ ] 🧑‍⚖️ gate humano final — interacción real del diálogo nativo (requiere sesión con display)

## Retomar aquí

Todo lo automatizable (Go E2E headless + FE typecheck/lint/tests + Rust cargo check)
quedó VERIFICADO EN VIVO esta sesión (2026-07-13) — ver `PARIDAD.md` para la evidencia
exacta. Única desviación real: el picker nativo Tauri no se click-through porque esta
sesión corrió sin display. Próximo paso: abrir la app en una sesión con display
(`pnpm --dir web tauri dev`, o el `.deb`/`.AppImage`) y firmar el gate humano
(instrucciones en `PARIDAD.md` §Firma).
