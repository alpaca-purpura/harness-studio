# Paquete · Forja · Ciclo de vida del arnés VIVO (Fase 1)

> Realineación core ciclo-arnés (2026-07-10). **En curso · NO firmado.**
> Modelo BASE ya FIRMADO 🧑‍⚖️ en [`../2026-07-10-terreno-conocimiento/`](../2026-07-10-terreno-conocimiento/INDEX.md)
> (D0-D20 + `arnes.yaml`). Este paquete lo **EJECUTA**: que ArnesIA forje, muestre y edite arneses de verdad.
> UI: SSoT = Storybook; baseline = `mockups/arnesia-mapa-baseline.html` (anti-drift §10/ux).

## Objetivo (outcome)

Loop end-to-end **USABLE**: `chat → arnes.yaml → gate de completitud → scaffold determinista → Mapa`.
Dogfood: ArnesIA se forja con ArnesIA. Agnóstico al rubro (dev = un ejemplo, el motor NO asume software).
Appetite Shape-Up: **slice fino primero** (una dimensión real end-to-end antes que las 11), luego expandir.

## Retomar aquí
> 📌 **LEER PRIMERO: [`decisiones.md`](./decisiones.md) (F-D0…)** — decisiones de ejecución de esta fase.
> El modelo (qué se construye) está FIRMADO en terreno-conocimiento; acá se decide CÓMO se ejecuta.

**Secuencia (cada etapa cierra con capability + firma 🧑‍⚖️):**
1. **DOGFOOD SCAFFOLD** — `docs/terreno/{proposito,producto,organizacion}/` + `docs/wip/` derivados del `arnes.yaml`
   (INDEX por dimensión + hojas atómicas + `knowledge/`), migrando `docs/architecture/` → `terreno/producto/`.
   Reconciliar 11 dimensiones ↔ 82 capabilities SIN romper R1-R4 (P6). Es el ejemplo VIVO del modelo.
2. **FORJADOR MÍNIMO + GATE** — motor que LEE `arnes.yaml`, corre el GATE DE COMPLETITUD (D19, extiende
   `arnesia conformance`) y hace el scaffold del paso 1 **determinísticamente** (rellena PLANTILLAS, nunca de cero).
3. **CHAT FORJA/EDITA** (el corazón) — cablear el chat CC in-app: (a) CREAR arnés por conversación → `arnes.yaml`
   → gate → scaffold; (b) EDITAR uno existente vía init/doctor/loop-forward (D8: sello/deriva/cosecha-back).
4. **MAPA DESTINO** — `mockups/arnesia-mapa-destino.html` como SUPERSET ESTRICTO del baseline, renderizando el
   terreno REAL (salud vacío/parcial/lleno · WIP con estados→done · overlays calidad+economía · sello/deriva ·
   receta en el inspector). Luego portar a Storybook stories.
5. **DESPUÉS** — forjar arneses 2° orden (`/po /architect /dev-team /auditor`) + upstream del schema `arnes.yaml`
   + terreno al kit `harness@prenter-marketplace`.

## Estado de la secuencia — ⏸ PAUSADA (F-D6)

> **Historia pausada** para meter primero el **spike de CARGA de arneses** (agencia: el user elige qué
> cargar/mejorar desde la app). Paquete del spike: [`../2026-07-10-spike-carga-arneses/`](../2026-07-10-spike-carga-arneses/INDEX.md).
> Retomar ESTA historia = cuando el spike cierre su decisión.

- **Slice 1a — CONSTRUIDO, pendiente firma 🧑‍⚖️:** golden scaffold `docs/terreno/` (raíz + Propósito/Producto stub
  + Organización con `forma-trabajo` end-to-end: INDEX·hoja D9·knowledge) + esqueleto `docs/wip/`. Verificado:
  `estado.sh --check` ✓ · 82 caps R1 ✓ (F-D4). **NO revertir** (migrará, F-D5). Committeado como WIP checkpoint.
- **Aclaración clave (F-D5):** la expertise de forja se ENCARNA en el **kit ② embebido** (cuerpos ①②, HS-10),
  no en `docs/terreno/` top-level; el output de forjar = un arnés ③ en su propio tree. El paso 2 (forjador) se
  reencuadra por esto. `docs/terreno` firmado-absorbe-architecture queda bajo revisión (P9 en terreno-conocimiento).
- **Retomar aquí (tras el spike):** firmar Slice 1a → **paso 2 (forjador + gate D19)** reencuadrado por F-D5, o
  extender Slice 1b. Contexto imprescindible: F-D5 + F-D6 en `decisiones.md`.

## Archivos
- `decisiones.md` — decisiones de ejecución (F-D0…). ← fuente de verdad de esta fase.
- (spec/paridad se agregan al avanzar cada etapa)
