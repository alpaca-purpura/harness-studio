# Identidad de build en Ajustes (paquete de trabajo)

> `tipo: bugfix/mejora chica` · **Vig: activo · construido y refinado, falta firma 🧑‍⚖️.**
> Detonante: *«¿puedes hacer que en ajustes aparezca la última versión incluyendo un último número
> que sea el build, para así saber si estoy usando la última versión compilada e instalada o una
> anterior?»* (operador, 2026-07-26).

## El agujero

Ajustes mostraba `arnesia · huella 1c7443f · 2026-07-07`, y eso **no responde** la pregunta:

1. **El semver no estaba en el binario** — `0.2.21` vivía solo en `Cargo.toml`/`package.json`.
2. **La huella es el COMMIT, no el BUILD** — ese mismo día se bundleó v0.2.21 **dos veces** desde el
   mismo árbol: dos binarios distintos, huella idéntica.

## Lo que quedó

**`arnesia v0.2.21.2607260225`** — semver + sello `AAMMDDHHMM`. Cambia siempre, es monótono (más
grande = más nuevo) y no guarda estado en el repo. Debajo, `compilado 2026-07-26 02:25 · commit
a5dc5f3`. Y si en disco hay un build más nuevo, un aviso en la superficie que dice qué hacer:
cerrar y reabrir la app, o correr `make dev-sync`.

## Estado

- [x] Decisiones B-D1..B-D6 → [`decisiones.md`](./decisiones.md) (B-D1 y B-D4 las eligió el
      operador; B-D5/B-D6 salieron del refinamiento y documentan lo ya construido)
- [x] [`spec.md`](./spec.md) — **RF-231**: mapa funcional (7 bifurcaciones · RN-1..8 · AC-1..9) +
      Gherkin `SC-1..SC-7` + matriz de cobertura + límites aceptados
- [x] [`design.md`](./design.md) — el UI al pixel (filas, tokens, microcopy, a11y)
- [x] Implementado: sello por `-ldflags` en `bundle.sh` (un solo punto: cubre build a mano,
      `dev-sync` **y self-update**) · `GET /api/version` extendido · tarjeta de Ajustes
- [x] 12 tests Go de identidad + 3 stories nuevas (15 en total en la tarjeta)
- [x] [`PARIDAD.md`](./PARIDAD.md) — fila por fila, con las salidas reales + límites conocidos
- [x] Snapshot derivado re-derivado en su bloque de versión + fila en `mockups/INDEX.md` (DoD §5)
- [x] Gate pendiente registrado en [`BACKLOG.md`](../../BACKLOG.md)
- [x] **AC-9a — los 4 estados VISTOS en la app real** (daemon sellado + SPA embebido, 1440×900):
      al día · build sin instalar · «cerrá y reabrí» · sin sellar → [`shots/`](./shots/)
- [ ] 🧑‍⚖️ **Gate de PARIDAD (AC-9b)** — falta la **ventana Tauri instalada** (`make installer` +
      sudo: lo corre el operador)

## Archivos

- `decisiones.md` — B-D1 (sello vs contador) · B-D2 (por qué en `bundle.sh` y no en el Makefile) ·
  B-D3 (aditivo a la huella) · B-D4 (el aviso y su margen) · B-D5 (sin mockup nuevo, a propósito) ·
  B-D6 (los límites del aviso, declarados).
- `spec.md` — RF-231 + mapa funcional + matriz de cobertura.
- `design.md` — la tarjeta al pixel.
- `PARIDAD.md` — verificación con salidas reales.

## Retomar aquí

- **Último hecho (2026-07-26, refinamiento en 2 tramos):**
  1. *Documental.* Se auditó el paquete contra §10 y contra el código real (`go test ./...` exit 0 ·
     fitness verde · 12 tests y 15 stories contados, coinciden) y se cerraron cinco huecos:
     faltaban `design.md`, la capa humana del spec (mapa funcional + matriz), la justificación de no
     llevar mockup, los límites del aviso, y el registro del gate en el BACKLOG.
  2. *En vivo, con los ojos.* Se compiló (`bundle.sh --daemon-only`, sello `0.2.21.2607261032`), se
     levantó el daemon con estado aislado —sin tocar `~/.arnesia` ni `~/.local/bin`— y se forzaron
     los **4 estados** en el navegador: al día · build sin instalar · «cerrá y reabrí» · sin sellar.
     Capturas en [`shots/`](./shots/). Salieron dos correcciones: el límite «se evalúa por request»
     era peor de lo real (**refetchea al entrar a Ajustes**, sin recargar), y un hallazgo ajeno —
     el SPA tiene `127.0.0.1:4200` **hardcodeado**, en otro puerto la UI entera dice
     `Failed to fetch` (al BACKLOG, no se arregló acá).
- **Próximo paso concreto (AC-9b, lo único abierto):** `make installer` → instalar el `.deb` (pide
  sudo) → **`make dev-sync` porque `~/.local/bin/arnesia` existe** (si no, el shell instalado sigue
  con el binario viejo: `docs/architecture/conventions/versionado.md`) → abrir Ajustes en la ventana
  nativa y confirmar el identificador. Lo demás ya está visto.
- **Firmas pendientes:** una sola — 🧑‍⚖️ PARIDAD (AC-9b). B-D1..B-D6 quedan decididas a la espera de
  esa misma firma.
- **Nada de este paquete está commiteado todavía**: `git status` lo muestra untracked junto a los
  paquetes de dictado y marketplace. Al firmar, commitear (§10 regla 3).
