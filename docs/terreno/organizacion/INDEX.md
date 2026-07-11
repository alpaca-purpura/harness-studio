# Territorio · Organización (Endeavour · definición)

> vig: activo · cómo se organiza el trabajo. Derivado de [`arnes.yaml`](../../product/stories/2026-07-10-terreno-conocimiento/arnes.yaml) `territorio == organizacion`.

## Dimensiones (4 canónicas)

| dimensión | Zachman | label (dev) | activación | salud | hoja |
|---|---|---|---|---|---|
| `proceso` | When/How | Ciclo dev (spines por tipo-de-paquete) | always | ⬜ vacío | (pendiente · paso 2) |
| `equipo` | Who | Roles dev (RACI + permisos) | demand | ⬜ vacío | (pendiente) |
| **`forma-trabajo`** | How | code-style · PARIDAD | always | 🟩 **lleno** | [forma-trabajo/](./forma-trabajo/INDEX.md) |
| `gestion-trabajo` | How | Backlog / flujo de historias | always | ⬜ vacío | (pendiente · paso 2) |

> **Deslinde (D19):** `proceso` = la SECUENCIA (fases→cajas→rol→plantilla). `gestion-trabajo` = el
> CONTENEDOR/lifecycle (tipos-de-paquete · estados · jerarquía · cierre=ratifica capability). La máquina de
> estados de `gestion-trabajo` se ALINEA con las fases de `proceso`, no las duplica.
> `forma-trabajo` = dimensión piloto del Slice 1a (fixture del forjador).
