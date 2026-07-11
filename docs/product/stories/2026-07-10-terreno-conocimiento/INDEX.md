# Paquete · Terreno · Dimensiones · Conocimiento (co-diseño)

> Realineación core ciclo-arnés (2026-07-10). **Co-diseño en curso · NO firmado.**
> Estado del UI: SSoT = Storybook; baseline = `mockups/arnesia-mapa-baseline.html`.

## Retomar aquí
> 📌 **LEER PRIMERO: [`decisiones.md`](./decisiones.md) (D0-D20)** — fuente de verdad del modelo. `HANDOFF.md`
> da contexto pero su §1-§7 quedó stale (banner adentro apunta a D18-D20). Visual al día = `estructura-terreno.html` (v4).
> ✅ Observaciones del user cocheadas → **D18** (regla de las 3 caras) · **D19** (11 canónicas + 2 overlays) ·
> **D20** (paquete vs artefacto + multi-pipeline por tipo). ✅ `arnes.yaml` (dogfood dev) escrito. **Nada firmado ni commiteado.**

Definiendo el **modelo de dimensiones del landscape** (lo que un arnés exige/construye en un
proyecto) con base metodológica, antes de tocar el product-map.

**Hecho:**
- Distinción **Terreno (todo) vs Dimensión (eje) vs Conocimiento (implementa)** — `decisiones.md` D0-D1.
- Wave 1 research (5 subagentes, metodologías + marcos + naming) → `research-terrenos.md` + `research-marcos-naming.md`.
- Decidido: naming triada (D1) · macro Essence+Zachman (D2) · calidad=overlay con Seguridad/Testing/Compliance-con-carpeta (D3) · set de dimensiones propuesto (D4) · conocimiento co-locado (D5) · nav anti-token (D6).

**Hecho (cont.):**
- Wave 2 research (arneses-as-code) → `research-arneses-as-code.md`; mecanismos en `decisiones.md` D7-D9.
- Carpetas = `docs/terreno/<dimensión>/` (absorbe `docs/architecture/`, dogfood; `capabilities/` queda).
- **Modelo del motor cerrado (D10-D17):** ArnesIA = **FORJA** (no marketplace ajeno; nosotros creamos los
  arneses, los sabemos leer) · forjador determinista+gates+dogfood · **10 dimensiones canónicas** + calidad = gate.
- **4 TERRITORIOS (D14-D17):** **Propósito · Producto · Organización · WIP** + Calidad (overlay). Rename
  per-arnés OBLIGATORIO. **WIP** = 4º territorio de naturaleza INSTANCIA (trabajo vivo con estados → `done/`).
  **capability** = detalle implementado + CIERRE del paquete; **Producto = Σ capabilities func+no-func**
  (Calidad produce las no-funcionales). Fix-épica (outcome-limpio · release=tag ortogonal · snapshot≠bitácora).
- **Visual actualizado a v3** (`estructura-terreno.html`): 4 territorios + toggle rubro (Canónico/Dev/Finanzas)
  + WIP con estados/done + flujo necesidad→capability→Producto.

**Siguiente:**
1. ✅ **`arnes.yaml`** (linchpin) escrito como instancia del arnés-dev — 11 canónicas + 2 overlays + proceso multi-pipeline + gestion-trabajo + gate.
2. **Refresco del mapa-UI DESTINO** (`mockups/arnesia-mapa-destino.html`) como superset ESTRICTO del baseline (SSoT=Storybook), con el modelo D20.
3. **Scaffold real** `docs/terreno/` + `docs/wip/` (migrar `docs/architecture/`) derivado del `arnes.yaml`.
4. **Dogfood:** forjar el arnés-dev end-to-end → validar 11 canónicas + tipos-de-paquete + gate.
5. **Firmar** el modelo → registrar en `mockups/INDEX.md` → repriorizar BACKLOG Fase-1/Fase-2.

## Archivos
- `decisiones.md` — decisiones vivas (**D0-D20** + abiertas). ← fuente de verdad del modelo.
- `arnes.yaml` — **linchpin as-code**: instancia del arnés-dev (dogfood). 11 canónicas + overlays + proceso multi-pipeline + gestion-trabajo + gate.
- `estructura-terreno.html` — **mapa visual v4** (D20: 3 caras · 11 canónicas · paquete⊃artefactos · multi-pipeline · overlays calidad+economía). ← abrir.
- `research-terrenos.md` — Wave 1a: metodologías dev por dimensión.
- `research-marcos-naming.md` — Wave 1b: marcos (42010·SWEBOK/Essence·Zachman·ISO25010) + naming.
- `research-arneses-as-code.md` — Wave 2: BMAD·Kiro·OpenSpec·plugins CC·cruft/copier/projen·aider.
- `research-proyecto-ejecucion.md` — taxonomía PM + Vitalia local (el 4º nivel WIP, fix-épica).
