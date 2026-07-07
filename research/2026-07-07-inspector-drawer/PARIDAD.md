# PARIDAD — mockup v6 ↔ app (se llena fila por fila al implementar)

> Contrato de «lo que ves en el mockup ES lo que hace la app». Gate final del paquete:
> todas las filas ✅ + click-through lado a lado (app real vs artifact) con consola
> limpia y screenshots revisados.

| RF | Elemento (mockup v6) | Componente real | Story=test | Estado |
|---|---|---|---|---|
| RF-80 | ⤢ ampliar (524) | inspector.tsx (`dw-expand`, aria-pressed) | ExpandeColapsaCierra | 🔶 |
| RF-81 | ✕ cierra ≠ ⤡ colapsa (552-568) | inspector.tsx + página (drawer persistente) | ExpandeColapsaCierra | 🔶 |
| RF-82 | tabs 3-up (527-531) | inspector.tsx (tablist ARIA) | Tabs | 🔶 |
| RF-83 | expandido: sticky+rejilla+920 (152-167) | inspector.css (`.expanded`) | ExpandeColapsaCierra (clase) | 🔶 |
| RF-84 | estado vacío (541-550) | inspector.tsx (`box` opcional) + página | Vacio | 🔶 |
| RF-85 | «i» secciones + SEC_TIP (309-326) | entities/arnes/model/doctrina.ts + Section | TooltipsDoctrinales | 🔶 |
| RF-86 | campos punteados + DEF_CAMPO/VALOR (327-362) | doctrina.ts (tipDe) + Field | TooltipsDoctrinales | 🔶 |
| RF-88 | a11y tooltips/contraste | ambos + axe (a11y addon `error`) | toda la suite (axe) | 🔶 |
| RF-89 | Viene de (364-371) | selectors.ts (selectVieneDe) + inspector.tsx | VieneDeChips | 🔶 |
| RF-90 | chips navegables + cond (373-378) | inspector.tsx (Chip + onSelect) | VieneDeChips | 🔶 |
| RF-91 | Hallazgos (379-388) | selectors.ts (selectHallazgosConformance) + Hallazgos + api.getConformance | HallazgoGateNone · HallazgosBotonera | 🔶 |
| RF-92 | botonera staged (501-508) | inspector.tsx (BotoneraStaged) | HallazgosBotonera | 🔶 |
| RF-93 | tab Contenido + endpoint fuente (410-433) | Contenido/SrcView + ports.FuenteReader + FuenteService + GET …/fuente | ContenidoFuenteReal · ContenidoErrorHonesto + go test (usecase/adapter) | 🔶 |
| RF-94 | no-reconocido RAW | mismo viewer (el endpoint sirve el artefacto tal cual) | ContenidoFuenteReal (universal) | 🔶 |
| RF-95 | acciones staged Contenido | inspector.tsx (acciones disabled rotuladas) | ContenidoFuenteReal | 🔶 |
| RF-96 | tab Corridas honesta (434-443) | inspector.tsx (Corridas) | CorridasHonesta | 🔶 |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar · ✅ verificado en paridad.

## Desviaciones registradas (a consultar en el gate final)

1. **AA sobre el mockup (RF-88 manda):** `.cond` (badge condición de ruta) — el mockup
   pinta texto `--warn` a 10px sobre fondo claro = 3.76:1 en axe → la app usa
   `--warn-soft` de fondo + texto `--foreground` (borde dashed warn se conserva).
   Ídem `.chip-versiona`: texto `--primary` a 9px = 3.44:1 → texto `--foreground`
   (borde/fondo primary se conservan). El a11y addon corre con `test: "error"`; el
   propio design.md fija «contraste AA en ambos temas» — el mockup pierde.
2. **Interpretación RF-81 ↔ RF-84:** «✕ cierra del todo (mapa sin drawer)» convive con
   «sin selección → affordance, no ausencia» así: ✕ quita el overlay expandido Y el
   drawer del NODO; queda el drawer VACÍO (affordance 340px). Única lectura que
   mantiene vivos ambos RF firmados.
3. **Tabs/header sticky también en drawer normal** (el mockup solo lo definía en
   expandido — sus tarjetas de galería no scrolleaban): RF-82 pide «conmutan panes sin
   perder el header»; sticky siempre lo garantiza.
4. **Chips inertes** = `disabled` real (mockup: `data-off`): mismo estilo (opacity .85,
   cursor default), semántica nativa.
5. **Showcase sin disco:** `content-studio-full` es fixture embebido → la tab Contenido
   responde el 404 honesto del endpoint («arnés sin directorio registrado»); el mockup
   lo suplía con fuente RECONSTRUIDA rotulada. La app jamás reconstruye (decisión #4:
   «lo no reconstruible muestra su estado»).
