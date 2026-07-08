# Doctrina visible — revisión UI + diseño de tarjeta e inspector

> Ficha HS-09 (fase 5) · plan Hito 2 · 2026-07-06
> Revisión hecha con ojo-UI propio sobre el mockup firmado `mockups/arnesia-mapa-mvp.html` (ejemplo Luana, renderizado 1600px). Screenshot en scratchpad.

## Qué lee BIEN hoy (mantener)

El Mapa ya es fiel a VISION §184-207. Lo que funciona y no se toca:

- **Geografía doctrinal de 3 territorios** — Guardia (magenta) · Proceso (carriles por fase) · Base (teal). El ojo ubica la banda en <1s.
- **Spine horizontal de cajas** — las cajas alineadas en la fila 1 de cada carril forman el backbone de la fábrica. Es la lectura más fuerte del mapa: se ve el flujo idea→released de un vistazo.
- **Glyph = canal redundante** (forma + char + color) — colorblind-safe, distingue las 10 clases.
- **Distinción caja** (borde 5px + tint + badge «caja») + **tag de transición `◇`** — la doctrina «una caja = una transición» ya es visible.
- **Marcas PROPUESTA** — `origen:del-puesto` (borde punteado), `canal:propuesto` (badge). Bien.
- **Activación por banda** (chips: siempre / condicional / bajo demanda / leído / dormida) + reglas colapsables (46 reglas → sub-banda). Anti-ruido correcto.
- **Edges anti-spaghetti** — solo el backbone `invoca` siempre visible; `lee`/`escribe` revelados al hover. Correcto.

## Hueco central: la CLASIFICACIÓN doctrinal NO se ve en la tarjeta

METODOLOGIA §8 define la clasificación como **tres ejes perpendiculares**: `clase ⊥ arquetipo ⊥ perfil_harness`. Hoy la tarjeta pinta **solo `clase`** (glyph). Los otros dos ejes — el corazón de «autonomía ≠ automatización» — viven solo como dato y en el inspector. **Somos el promotor de esta metodología: debe verse en el mapa mismo.**

### Faltantes confirmados (dato existe en `Contract`/`Box`, no se dibuja)

En la **tarjeta del nodo**:
1. **`arquetipo`** (pipeline · excepcion · abierto · no-arnesar) — la FORMA del trabajo. Sin marca.
2. **`perfil_harness`** (T1 · T2 · T3) — CÓMO ejecuta. Sin marca.
3. **`gate.tipo` + honestidad** — `none` = hallazgo (A4). Hoy el gate no aparece en ninguna parte de la tarjeta.
4. **`procedencia`** — medido≠estimado≠declarado (gris≠verde). Sin marca en tarjeta.

En el **inspector**:
5. **`constraints`** y **`non_goals`** — parte de INTENCIÓN, no renderizados.
6. **`gate.aceptacion`** (Gherkin given/when/then) — **el eval ejecutable**, lo más doctrinal (METODOLOGIA §3). No renderizado.
7. **`gate.evidencia`** (telemetría de nacimiento) — no renderizado.
8. **El spine del arnés como máquina de estados** — existe (`arnes.spine`), no hay vista; se infiere solo de los tags alineados.

## Diseño propuesto — cerrar el hueco (respeta tokens DTCG + canvas⊥chrome)

### Tarjeta del nodo (`entities/arnes/ui/arnes-node.tsx`)

Añadir una **fila de metadatos densa y discreta** al pie de la tarjeta-caja (solo cuando `caja:true`), tipografía mono `text-xs`, sin romper la limpieza:

```
┌──────────────────────────────┐
│ ▣ destilar el brief    [caja] │
│   /brief-writer               │
│   ◇ idea → brief              │
│   ⬡pipeline  ◈T2  ●manual     │  ← NUEVA fila: arquetipo · perfil · gate
└──────────────────────────────┘
```

- **`arquetipo`** → un char-badge mono (`pipeline`=`═` recto · `excepcion`=`≈` · `abierto`=`✳` generativo · `no-arnesar`=`○` fuera-de-caja). Tooltip con el nombre. NO necesita token de color nuevo — usa `--muted-foreground` + forma.
- **`perfil_harness`** → pill `T1`/`T2`/`T3` con opacidad creciente (T1 tenue → T3 fuerte) = «más autonomía, más presencia». Reusa `--foreground`.
- **`gate.tipo`** → punto de estado con token `color.health`: `auto`=`--ok` · `parcial`=`--warn` · `manual`=`--c-skill` (humano) · **`none`=`--crit` anillo punteado** (el hueco grita, no se esconde). **Esta es la marca de honestidad — la más importante.**
- **`procedencia`** → el borde/relleno ya existente puede modularse: `medido`=sólido pleno · `estimado`/`inferido`=tint atenuado · `no-declarado`=hairline gris. «gris ≠ verde» hecho pixel.

Todos los valores = `var(--…)` (disciplina stylelint anti-magic-value). Cada marca = una **story = test** (fitness visual): una story por arquetipo, por perfil, por gate.tipo (incl. `none`).

### Inspector (`widgets/map-canvas/ui/inspector.tsx`)

Completar el contrato de 4 ejes (INTENCIÓN · CLASIFICACIÓN · CABLEADO · ACEPTACIÓN):

- **INTENCIÓN** — añadir `constraints` (lista con `⊘`) y `non_goals` (lista con `✕`) bajo `why`+`capabilities`.
- **ACEPTACIÓN** — sección **Gherkin** renderizada como bloque (given/when/then con sangría), + `evidencia` como pie citado. Si `gate.tipo:none` → banner honesto «sin eval formal — hallazgo».
- **CLASIFICACIÓN** — hacer explícitos los 3 ejes con etiqueta (`clase · arquetipo · perfil_harness`) en vez de listarlos sueltos.
- **Spine mini-mapa** — un strip de estados `idea→…→medido` con el estado de ESTA caja resaltado (la caja se ubica en la máquina de estados del arnés).

## Sobre el mojibake del mockup

El screenshot mostró `DiseÃ±o`, `â·`, `TelemetrÃ­a`. Es artefacto de **servir el mockup sin `charset=utf-8`** (python http.server), NO un bug de producto: la app Vite emite `<meta charset>` y `map.css` es port verbatim. **Verificación a agendar en Gate 2**: confirmar en la app real corriendo que los glyphs (`◇ ⬡ ◈ ✳`) y acentos renderizan limpios.

## Checkpoints de verificación humana (UI)

- **Gate 2 (doctrina visible)** — click-through en la app real: cada casuística del showcase legible; las 4 marcas nuevas de tarjeta + las 4 secciones nuevas de inspector presentes; `gate:none` inconfundible; consola limpia; screenshots revisados (disciplina UX.md: mismo artifact, asserts, consola limpia).
- Comparar lado-a-lado showcase vs `dev-full-cycle` (el honesto pobre) — el contraste prueba que las marcas escalan.
