# PARIDAD — Identidad de build en Ajustes (RF-231)

> Paquete `2026-07-26-identidad-de-build`. Cada RF ⇒ código real + su verificación; cada
> verificación ⇒ corrida de verdad, con su salida. **Fecha: 2026-07-26.**
> Firma 🧑‍⚖️ del gate: **PENDIENTE**.
> **Refinado 2026-07-26** (2ª pasada): se re-corrieron las suites Go, se agregó la fila del punto
> único de inyección (RN-7), la § Límites conocidos y la § Re-verificación. Cero cambios de código.

## Fila por fila

| Qué exige el RF | Código | Verificación | Estado |
|---|---|---|---|
| identificador completo del build en Ajustes | `internal/adapters/selfupdate/identidad.go#VersionCompleta` · `updater.go#Version` · `features/self-update/ui/update-card.tsx` | daemon real: `GET /api/version` | ✅ `{"version":"0.2.21.2607261006","compilado":"2026-07-26 10:06"}` |
| el sello se inyecta al compilar | `scripts/bundle.sh` (`-ldflags -X`) | `bash scripts/bundle.sh --daemon-only` | ✅ `identidad: 0.2.21.2607261006 (2026-07-26 10:06)` |
| **un solo punto de inyección** (RN-7 · B-D2) | `scripts/bundle.sh:40-48` | lectura directa: `updater.go:231` corre ese script con `--daemon-only`; `Makefile:56` (`installer`) y `Makefile:76` (`dev-sync`) también. No hay otro `go build` que produzca un binario distribuible | ✅ verificado en el refinamiento |
| dos builds del mismo commit se distinguen | idem | los dos bundles de v0.2.21 del 2026-07-26 | ✅ mismo commit `a5dc5f3`, sellos `2607260225` vs `2607261006` |
| más grande = más nuevo | `identidad.go#selladoEn` | `TestSelloMasGrandeEsMasNuevo` | ✅ pass |
| avisa si el instalado es más nuevo | `identidad.go#avisoDeBuildViejo` | daemon real con el exe tocado a futuro | ✅ `"el binario instalado es más nuevo (2026-07-26 12:07) que el que está corriendo — cerrá y reabrí la app"` |
| avisa si el repo tiene un build sin instalar | idem | daemon real con `bin/arnesia` tocado a futuro | ✅ `"hay un build más nuevo sin instalar (2026-07-26 11:06 en …/bin/arnesia) — corré \`make dev-sync\`"` |
| si corre el último, NO avisa | idem | `GET /api/version` del build recién compilado | ✅ `aviso_build` ausente (omitempty) |
| el propio build no se denuncia a sí mismo | `identidad.go` (margen de 5 min) | `TestElPropioBuildNoSeDenunciaASiMismo` | ✅ pass |
| sin identidad inyectada dice `dev`, no inventa | `identidad.go#VersionCompleta` | `TestVersionCompletaSinIdentidadDiceDev` · `TestSinSelloNoAvisaNada` | ✅ pass |
| la ruta borrada del self-update no rompe | `identidad.go#masNuevoQueEsteBuild` | `TestRutaBorradaNoRompeNiAvisa` | ✅ pass |
| la superficie lo pinta | `update-card.tsx` | stories `IdentidadDelBuild` · `BuildViejoCorriendo` · `BuildSinSellar` | ✅ **15 passed** en `update-card.stories.tsx` |
| el contrato lo declara | `docs/architecture/contracts/api/openapi.yaml:955-975` (`VersionInfo`) | lectura: `version` · `compilado` · `aviso_build` documentados en el schema de `GET /api/version` | ✅ |

La cobertura completa **bifurcación × regla × criterio → escenario → verificación** está en
[`spec.md` § Matriz de cobertura](./spec.md). Huecos: ninguno salvo AC-9 (el gate humano).

## Suites

| Suite | Resultado |
|---|---|
| `go build ./...` · `go test ./...` | verde |
| `go test ./internal/adapters/selfupdate/` | **12 tests nuevos de identidad**, todos pass |
| `go test ./docs/architecture/fitness/` | verde (R1/R2/R3/R4: `identidad.go` reclamado por CAP-60) |
| `pnpm run verify` | verde |
| stories de la tarjeta | **15 passed** (eran 12) |

### Re-verificación del refinamiento (2026-07-26, 2ª pasada)

| Comando | Salida |
|---|---|
| `go test ./...` | **exit 0**, cero fallos |
| `go test ./internal/adapters/selfupdate/ ./docs/architecture/fitness/` | `ok … selfupdate` · `ok … fitness (23.1s)` |
| `grep -c '^func Test' internal/adapters/selfupdate/identidad_test.go` | **12** — coincide con lo declarado |
| `grep -c '^export const' update-card.stories.tsx` | **15** — coincide con lo declarado |

Las suites del FE **no se re-corrieron** en esta pasada: el refinamiento no tocó una línea de código
(solo documentación), y `pnpm run verify` no corre en background — Chromium no arranca headless.

## § Verificación EN VIVO contra la app real (2026-07-26, 2ª pasada)

Levantada de verdad, no transcrita: `bash scripts/bundle.sh --daemon-only` → sello
**`0.2.21.2607261032`** → copia del binario a un directorio aparte, corriendo con estado aislado
(`--sessions/--arneses/--index` al scratchpad) para **no tocar `~/.arnesia` ni
`~/.local/bin/arnesia`** → SPA embebido abierto en el navegador a 1440×900 → los 4 estados forzados
con `touch` de mtimes. Capturas en [`shots/`](./shots/).

| Estado | Cómo se forzó | Lo que se VIO en la tarjeta | Captura |
|---|---|---|---|
| al día | build recién sellado | `daemon · arnesia v0.2.21.2607261032` · `compilado 2026-07-26 10:32 · commit e758d13+sucio · 2026-07-26` · **sin aviso** | `shots/01-al-dia-sin-aviso.png` |
| build sin instalar | `touch -d 12:40 bin/arnesia` | caja warn `▲ hay un build más nuevo sin instalar (2026-07-26 12:40 en …/bin/arnesia) — corré \`make dev-sync\`` | `shots/02-aviso-build-sin-instalar.png` |
| el instalado es más nuevo | `touch -d 13:10` sobre el ejecutable que corre | caja warn `▲ el binario instalado es más nuevo (2026-07-26 13:10) que el que está corriendo — cerrá y reabrí la app` · **gana sobre el del repo**, que también estaba a futuro (orden de urgencia de B-D4) | `shots/03-aviso-cerra-y-reabri.png` |
| build sin sellar | `go build` pelado, sin `-ldflags` | `daemon · arnesia · build sin sellar (dev)`, la fila 2 muta a `commit`, y **ningún aviso** pese a que en el repo había un binario más nuevo | `shots/04-build-sin-sellar-dev.png` |

**Corrección de lo que este mismo archivo afirmaba.** El límite «se evalúa por request» se midió y es
**menos malo de lo que estaba escrito**: la tarjeta refetchea al **montarse**, así que navegar
`Portafolio → Ajustes` con la app abierta trae el aviso al instante (verificado: se tocó el binario
con la app ya cargada y el aviso apareció al volver a Ajustes, sin recargar). Lo que NO hay es
polling estando ya parado en Ajustes.

**Hallazgo ajeno, destapado de paso:** el SPA embebido tiene el origen del daemon **hardcodeado a
`127.0.0.1:4200`**. Correr el daemon en otro puerto deja la UI entera en `Failed to fetch`
(`GET /api/version falló`) aunque el daemon responda perfecto por `curl`. No es de RF-231 — al
BACKLOG, sin arreglarlo al voleo.

**Estado tras la prueba:** daemons apagados, puerto 4200 libre, `bin/arnesia` con su mtime real
restaurado, `~/.local/bin/arnesia` y `~/.arnesia` **sin tocar**.

## § Límites conocidos (declarados, no descubiertos después)

| Límite | Efecto real | Decisión |
|---|---|---|
| Ventana ciega de 5 min | un rebuild dentro del margen no dispara aviso | B-D6 — es el precio de no tener falsos positivos |
| El aviso se evalúa al MONTAR la tarjeta | entrar a Ajustes lo trae al instante; parado ahí no hay polling | B-D6 — **medido en vivo**, no supuesto |
| CI compila sin sello | los binarios de `ci.yml` dicen `dev` — nunca se distribuyen | B-D6 |
| Sin `arnesia --version` | la identidad solo se ve por HTTP/Ajustes | B-D6 — deuda en BACKLOG |
| El aviso sin `aria-live` | no se anuncia si algún día cambia en vivo | `design.md` §7 |

## ⛔ Lo que NO se probó

Tras la verificación en vivo, la lista se acortó a tres cosas — y ninguna es la superficie:

- **La ventana Tauri INSTALADA** (AC-9b). Lo que se vio es la superficie real servida por el binario
  sellado, en el navegador. Falta el `.deb` instalado abriendo su ventana nativa: `make installer`
  necesita rust/Tauri y la instalación pide **sudo**, así que es del operador. La tarjeta no cambia
  —es el mismo SPA embebido—, pero la doctrina dice que no se firma sin verlo.
- **El sello a través del botón «Actualizar»**, de punta a punta. El self-update corre
  `bundle.sh --daemon-only`, camino ya verificado a mano, pero no se disparó el botón.
- **El aviso con un `.deb` recién instalado de verdad** (dos binarios, dos dueños — ver
  `docs/architecture/conventions/versionado.md`): se forzó con `touch` de mtimes, que es el mismo
  hecho que el código observa, pero no reinstalando el paquete.
