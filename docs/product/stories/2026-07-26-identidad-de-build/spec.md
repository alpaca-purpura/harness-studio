# Spec — Identidad de build en Ajustes (RF-231)

> Paquete `2026-07-26-identidad-de-build`. **El QUÉ**, no el cómo.
> Decisiones: [`decisiones.md`](./decisiones.md) (B-D1..B-D6).
> Numeración: continúa desde RF-230.
> **Refinado 2026-07-26** (2ª pasada): se agregaron § Mapa funcional · § Matriz de cobertura ·
> § Límites aceptados. El Gherkin no cambió de contenido, solo se le pusieron IDs `SC-N` para
> que la matriz pueda apuntarle.

### RF-231 — Ajustes dice QUÉ BUILD está corriendo, y avisa si hay uno más nuevo

Extiende RF-107 («identidad honesta»), que reportaba solo la huella del commit. La huella **no**
distingue dos compilaciones del mismo árbol — el caso real que dejó al operador sin poder saber si
corría lo último que había compilado.

## § Mapa funcional (la capa humana — se lee sin Gherkin)

### 1. Happy path

1. El operador compila (a mano, `make dev-sync`, `make installer` o el botón **Actualizar**).
   `scripts/bundle.sh` sella el binario con **semver + `AAMMDDHHMM`** vía `-ldflags`.
2. Abre **Ajustes**. El FE pide `GET /api/version`.
3. El daemon responde con su identidad completa **y** con el resultado de mirar el disco: ¿hay
   algún binario escrito después de que YO me sellara?
4. La tarjeta encabeza con **`arnesia v0.2.21.2607260225`**, debajo **`compilado 2026-07-26 02:25 ·
   commit a5dc5f3`**, y **ningún aviso** — porque el que corre es el último.
5. El operador compara ese número contra el que imprimió su última compilación y cierra la
   pregunta de un vistazo.

### 2. Bifurcaciones (árbol)

```
¿el binario trae identidad inyectada?
├─ NO (go build pelado: CI, `go run`)
│   └─ Bif-1 · dice «build sin sellar (dev)», NO inventa número · y NO evalúa avisos → [SC-6]
└─ SÍ
    ├─ ¿hay un archivo más nuevo que mi sello en la ruta de MI ejecutable?
    │   ├─ SÍ (dev-sync/self-update reemplazó el binario con la app abierta)
    │   │   └─ Bif-2 · avisa «el binario instalado es más nuevo … cerrá y reabrí la app» → [SC-3]
    │   └─ NO ↓
    ├─ ¿hay repo configurado y `<repo>/bin/arnesia` es más nuevo que mi sello?
    │   ├─ SÍ (compilaste y no instalaste)
    │   │   └─ Bif-3 · avisa «hay un build más nuevo sin instalar … corré `make dev-sync`» → [SC-4]
    │   └─ NO ↓
    └─ Bif-4 · sin aviso: el que corre ES el último → [SC-5]

casos de borde que NO son bifurcación de negocio pero sí de robustez:
├─ Bif-5 · la ruta del ejecutable quedó «(deleted)» tras el rename del self-update → no stat, no aviso
├─ Bif-6 · la ruta no existe / es un directorio → no rompe, no avisa
└─ Bif-7 · el sello es ilegible (largo ≠ 10, no parsea) → se trata como «sin sello» (Bif-1)
```

### 3. Reglas de negocio

| ID | Regla | Ancla |
|---|---|---|
| **RN-1** | El identificador de build **cambia siempre** — dos compilaciones del mismo commit dan números distintos | B-D1 |
| **RN-2** | El identificador es **monótono**: número más grande = build más nuevo, sin excepciones | B-D1 |
| **RN-3** | Sin identidad inyectada se dice `dev`; **jamás** se fabrica un número ni un aviso | B-D1 · honestidad §4 |
| **RN-4** | El sello es **aditivo** a la huella del commit: son dos preguntas y las dos se contestan | B-D3 |
| **RN-5** | El aviso **no puede aparecer cuando el que corre ES el último** — un aviso que aparece siempre no se lee nunca | B-D4 |
| **RN-6** | El aviso dice **qué hacer**, no solo que algo pasa (cerrá y reabrí / `make dev-sync`) | B-D4 |
| **RN-7** | El sello se inyecta en **un solo punto** (`bundle.sh`), el mismo por el que pasa el self-update | B-D2 |
| **RN-8** | La comparación es **mtime del otro archivo contra el sello propio**, nunca sello contra sello | B-D4 |

### 4. Criterios de aceptación

- [x] **AC-1** — Ajustes muestra semver + sello juntos, como un solo identificador copiable.
- [x] **AC-2** — Muestra la fecha-hora de compilación legible.
- [x] **AC-3** — El commit sigue visible, subordinado al build.
- [x] **AC-4** — Dos bundles del mismo árbol producen identificadores distintos y ordenables.
- [x] **AC-5** — El aviso de build viejo aparece en la **superficie** (caja `warn`), no en un `title`.
- [x] **AC-6** — Estar al día no dibuja nada.
- [x] **AC-7** — Un binario sin sellar lo dice con todas las letras.
- [x] **AC-8** — El binario recién instalado **no se denuncia a sí mismo** (margen B-D4).
- [x] **AC-9a** — Los 4 estados vistos con los ojos en la **app real** (SPA embebido servido por el
      binario sellado, 1440×900): al día · build sin instalar · «cerrá y reabrí» · sin sellar.
      Capturas en [`shots/`](./shots/), detalle en [`PARIDAD.md`](./PARIDAD.md) §Verificación en vivo.
- [ ] **AC-9b** — Visto en la **ventana Tauri instalada** (`.deb`). ⛔ **lo único abierto** → gate
      🧑‍⚖️. Pide `make installer` + sudo: lo corre el operador.

## § Gherkin

```gherkin
Escenario: SC-1 · el operador abre Ajustes
  Cuando mira la tarjeta de versión
  Entonces ve el identificador completo del build: semver + sello de compilación
  Y ve cuándo se compiló, en hora legible
  Y sigue viendo el commit, que es otro dato y no el mismo

Escenario: SC-2 · dos compilaciones del mismo commit
  Dado un binario compilado y otro compilado después desde el MISMO árbol
  Cuando se comparan sus identificadores
  Entonces son distintos
  Y el más grande es el más nuevo

Escenario: SC-3 · el binario se reemplazó con la app abierta
  Dado un daemon corriendo
  Cuando en su misma ruta aparece un binario compilado después
  Entonces Ajustes avisa que el instalado es más nuevo y que hay que cerrar y reabrir la app

Escenario: SC-4 · se compiló pero no se instaló
  Dado un daemon corriendo y un repo configurado
  Cuando en el repo hay un binario compilado después que el que corre
  Entonces Ajustes avisa que hay un build sin instalar, y con qué comando instalarlo

Escenario: SC-5 · el que corre ES el último
  Cuando no hay ningún binario más nuevo en disco
  Entonces NO se muestra ningún aviso

Escenario: SC-6 · un build sin sellar
  Dado un binario compilado sin inyectar identidad (CI, `go run`)
  Cuando el operador mira Ajustes
  Entonces dice que el build no está sellado
  Y NO inventa un número ni un aviso

Escenario: SC-7 · el propio build recién instalado
  Dado un binario cuyo mtime es POSTERIOR a su propio sello (siempre lo es: el linker escribe después)
  Cuando el daemon evalúa si hay algo más nuevo
  Entonces NO se avisa a sí mismo
```

## § Matriz de cobertura

Cada bifurcación y cada regla tiene ≥1 escenario y **una verificación que corrió de verdad**.

| ID | Escenario | Verificación real | Estado |
|---|---|---|---|
| Bif-1 · RN-3 | SC-6 | `identidad_test.go#TestVersionCompletaSinIdentidadDiceDev` · `#TestSinSelloNoAvisaNada` · story `BuildSinSellar` | ✅ |
| Bif-2 · RN-6 | SC-3 | daemon real con el exe tocado a futuro · `#TestAvisaCuandoElInstaladoEsMasNuevo` · story `BuildViejoCorriendo` | ✅ |
| Bif-3 · RN-6 | SC-4 | daemon real con `bin/arnesia` tocado a futuro · `#TestAvisaCuandoElRepoTieneUnBuildSinInstalar` | ✅ |
| Bif-4 · RN-5 | SC-5 | `GET /api/version` del build recién compilado (`aviso_build` ausente) · `#TestSinNadaMasNuevoNoAvisa` · story `IdentidadDelBuild` | ✅ |
| Bif-5 | — (robustez) | `#TestRutaBorradaNoRompeNiAvisa` | ✅ |
| Bif-6 | — (robustez) | `#TestRutaInexistenteNoRompe` | ✅ |
| Bif-7 · RN-3 | SC-6 | `#TestSelloIlegibleNoRompe` · `#TestVersionCompletaTolerantesAMediaIdentidad` | ✅ |
| RN-1 | SC-2 | los dos bundles reales de v0.2.21 del 2026-07-26 (`2607260225` vs `2607261006`, mismo commit `a5dc5f3`) | ✅ |
| RN-2 | SC-2 | `#TestSelloMasGrandeEsMasNuevo` | ✅ |
| RN-4 | SC-1 | story `IdentidadDelBuild` (asserta build **y** commit) · `versionBody` con `version`+`huella` | ✅ |
| RN-5 (anti-falso-positivo) | SC-7 | `#TestElPropioBuildNoSeDenunciaASiMismo` | ✅ |
| RN-7 | — | `scripts/bundle.sh` es el único `go build` de release; `updater.go:231` invoca ese mismo script con `--daemon-only`; `Makefile:56` (`installer`) y `Makefile:76` (`dev-sync`) también | ✅ |
| RN-8 | SC-3 · SC-4 | `identidad.go#masNuevoQueEsteBuild` (un `os.Stat`, cero ejecución del otro binario) | ✅ |
| AC-1..AC-8 | SC-1..SC-7 | ver [`PARIDAD.md`](./PARIDAD.md) fila por fila | ✅ |
| AC-9a | SC-1 · SC-3 · SC-4 · SC-5 · SC-6 | **los 4 estados vistos en la app real** — `shots/01..04` | ✅ |
| AC-9b | — | click-through en la **ventana Tauri instalada** | ⛔ **abierto** |

**Huecos detectados:** ninguno sin verificación, salvo AC-9b (gate humano por definición).
**SC huérfanos:** ninguno.

## § Límites aceptados (no son bugs — son el precio elegido)

| Límite | Por qué se acepta | Dónde vive |
|---|---|---|
| **Ventana ciega de 5 min**: un rebuild que ocurre *antes* de `sello + 5 min` del build corriendo no dispara aviso | Es el mismo margen que evita el falso positivo de RN-5, y el falso negativo dura lo que tarda el operador en abrir Ajustes; el siguiente rebuild sí avisa | `identidad.go#margenDeSello` · B-D6 |
| **El aviso se evalúa al montar la tarjeta**, no se empuja | Entrar a Ajustes lo trae al instante (medido en vivo: se tocó el binario con la app abierta y el aviso apareció al volver a Ajustes, sin recargar). Estando ya parado en Ajustes no hay polling — un push exigiría watcher sobre dos rutas para un dato que se mira una vez por sesión | B-D6 |
| **CI compila sin sello** (`ci.yml:34,88` = `go build` pelado) | Esos binarios **nunca se distribuyen**: `ci.yml` solo valida (no hay workflow de release), los instaladores salen de `make installer` → `bundle.sh`. Y si alguno corriera, dice `dev` (RN-3) | B-D6 |
| **No hay `arnesia --version` en la CLI** | La identidad se pidió para Ajustes; la CLI queda como deuda registrada, no como promesa incumplida | B-D6 · BACKLOG |
| **Cambio de huso / fin de horario de verano** puede repetir una hora una vez al año | Elegido a cambio de que el número no esté corrido 5 h del reloj que el operador mira | B-D1 |

## Trazabilidad

- **Capabilities:** `self-update/self-update-sin-sudo` (CAP-60 — el wire de `GET /api/version`,
  `identidad.go#VersionCompleta` y `#avisoDeBuildViejo` en sus `pointers`) ·
  `fe-shell/boton-actualizar` (la tarjeta de Ajustes). Ninguna nueva: el RF extiende superficie
  existente.
- **Sin mockup nuevo, a propósito (B-D5).** La superficie ya existe (`UpdateCard`) y el SSoT del UI
  es Storybook: el QUÉ se especifica contra las stories, y el snapshot derivado
  `stories/2026-07-07-boton-actualizar/mockup-actualizar.html` se re-deriva en el bloque de versión
  (DoD de [`mockups/INDEX.md`](../../../../mockups/INDEX.md) §5).
- **El UI al pixel:** [`design.md`](./design.md).
- **PARIDAD:** [`PARIDAD.md`](./PARIDAD.md).
