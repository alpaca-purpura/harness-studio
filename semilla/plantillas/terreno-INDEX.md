---
arnes: {{.ArnesID}}
sembrado: "{{.Fecha}}"
---

# Terreno — {{.ProyectoNombre}}

> El mapa de lo que este proyecto DEBE definir. Cada dimensión es un eje que se puebla con
> decisiones; su hoja llega en Fase 2 — hoy el hueco queda VISIBLE como `pendiente`, jamás
> fabricado.

## Territorios (4)

| Territorio | Naturaleza | Hoja |
|---|---|---|
| Propósito | definición | [proposito/INDEX.md](./proposito/INDEX.md) |
| Producto | definición | [producto/INDEX.md](./producto/INDEX.md) |
| Organización | definición | [organizacion/INDEX.md](./organizacion/INDEX.md) |
| WIP | instancia | [../wip/INDEX.md](../wip/INDEX.md) — trabajo vivo, vive fuera de `terreno/` |

## Dimensiones canónicas (11) — todas pendientes

| Dimensión | Territorio | Zachman | Estado |
|---|---|---|---|
| encargo | proposito | Why | pendiente |
| stakeholders | proposito | Who | pendiente |
| necesidad | proposito | What | pendiente |
| requisitos | producto | What | pendiente |
| producto | producto | How/What | pendiente |
| datos | producto | What | pendiente |
| contexto | producto | Where | pendiente |
| proceso | organizacion | When/How | pendiente |
| equipo | organizacion | Who | pendiente |
| forma-trabajo | organizacion | How | pendiente |
| gestion-trabajo | organizacion | How | pendiente |

Overlays transversales (calidad · economia): proyectan sobre las dimensiones; sin carpeta en v0.
