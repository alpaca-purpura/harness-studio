# PARIDAD — mockup v6 ↔ app (verificada fila por fila, 2026-07-07)

> Contrato de «lo que ves en el mockup ES lo que hace la app». Verificación FASE 3:
> daemon real :4200 (dogfood dev-full-cycle REGISTRADO en disco + showcase embebido) +
> vite :5173 + Chrome DevTools MCP; lado a lado contra el mockup v6 abierto en el mismo
> browser; consola limpia (0 errores/warnings) al cierre del recorrido; screenshots en
> el scratchpad de la sesión (`rf84-vacio…`, `rf-resumen-specwriter…`, `rf93-contenido-
> builder-real…`, `rf83-expandido-{resumen-rejilla,contenido-920}…`, `rf86-tooltip-tipo-
> none-hover…`, `rf93-404-honesto-showcase…`, `rf94-noreconocido-raw…`,
> `rf-drawer-dark-theme…`, `app-reviewcaja-ladoAlado…`, `mockup-v6-case16…`).

| RF | Elemento (mockup v6) | Componente real | Story=test | Estado |
|---|---|---|---|---|
| RF-80 | ⤢ ampliar (524) | inspector.tsx (`dw-expand`, aria-pressed) | ExpandeColapsaCierra | ✅ vivo: ⤢→cubre el área del mapa (rect 270,177→1599,901), botón conmuta a ⤡ |
| RF-81 | ✕ cierra ≠ ⤡ colapsa (552-568) | inspector.tsx + página (drawer persistente) | ExpandeColapsaCierra | ✅ vivo: Esc colapsa a 340px CONSERVANDO la tab activa; ✕ expandido → sin overlay, queda affordance (desviación #2) |
| RF-82 | tabs 3-up (527-531) | inspector.tsx (tablist ARIA) | Tabs | ✅ vivo: 3 tabs, header nunca se pierde (sticky) |
| RF-83 | expandido: sticky+rejilla+920 (152-167) | inspector.css (`.expanded`) | ExpandeColapsaCierra | ✅ vivo medido: header sticky/0 · tabs sticky/57px compactas flex-start · Resumen grid 4 col minmax(280) max 1280 · botonera grid-column 1/-1 fila · Contenido max-width 920px |
| RF-84 | estado vacío (541-550) | inspector.tsx (`box` opcional) + página | Vacio | ✅ vivo: al cargar el Mapa y tras ✕ — affordance, nunca ausencia |
| RF-85 | «i» secciones + SEC_TIP (309-326) | entities/arnes/model/doctrina.ts + Section | TooltipsDoctrinales | ✅ vivo: «i» botón focusable por sección; tip de Gate = SEC_TIP exacto |
| RF-86 | campos punteados + DEF_CAMPO/VALOR (327-362) | doctrina.ts (tipDe) + Field | TooltipsDoctrinales | ✅ vivo: hover sobre `tipo: none` pinta «Tipo de eval del gate (A4) — none — SIN eval formal…»; Gherkin literal: `procedencia: estimado` (draft-caja showcase) → «se dibuja atenuado (gris ≠ verde)» |
| RF-88 | a11y tooltips/contraste | ambos + axe (a11y addon `error`) | toda la suite (axe) | ✅ tabindex=0 verificado vivo; axe verde en 62/62 (cazó 2 violaciones AA del mockup → desviación #1) |
| RF-89 | Viene de (364-371) | selectors.ts (selectVieneDe) + inspector.tsx | VieneDeChips | ✅ vivo: builder ← «spec-writer invoca» (dogfood) · style-guide ← «draft-caja lee» (showcase); nodos sin entradas no muestran la sección |
| RF-90 | chips navegables + cond (373-378) | inspector.tsx (Chip + onSelect) | VieneDeChips | ✅ vivo (Gherkin E2E): click chip «edited.md ← caja: edit-caja» en review-caja → el mapa marca edit-caja `.selected` y el drawer lo muestra; «usuario»/«humano» inertes rotulados; badge «si cambios mayores solicitados» |
| RF-91 | Hallazgos (379-388) | selectors.ts (selectHallazgosConformance) + Hallazgos + api.getConformance | HallazgoGateNone · HallazgosBotonera | ✅ vivo: releaser (dogfood REAL, gate none) y draft-caja (showcase) → hallazgo crit A4; spec-writer → «Sin hallazgos abiertos»; conformance del daemon consumido (sin nota de indisponible) |
| RF-92 | botonera staged (501-508) | inspector.tsx (BotoneraStaged) | HallazgosBotonera | ✅ vivo: 4 acciones disabled con title de la fase que las cablea + nota staged |
| RF-93 | tab Contenido + endpoint fuente (410-433) | Contenido/SrcView + ports.FuenteReader + FuenteService + GET …/fuente | ContenidoFuenteReal · ContenidoErrorHonesto + go test (usecase/adapter) | ✅ vivo: builder → 54 líneas del SKILL.md REAL en disco (X-Arnesia-Fuente-Path); showcase brief-caja → estado honesto «sin directorio registrado»; draft-caja sin fuente_path → «pendiente del reconocedor»; curl directo 200/404/403 verificados |
| RF-94 | no-reconocido RAW | mismo viewer (el endpoint sirve el artefacto tal cual) | ContenidoFuenteReal (universal) | ✅ vivo con loader REAL: artefacto suelto `skills/misterio.txt` → nodo no-reconocido (Reconciliación warn + hallazgo D-c) y Contenido = RAW tal cual (limpiado tras la prueba) |
| RF-95 | acciones staged Contenido | inspector.tsx (acciones disabled rotuladas) | ContenidoFuenteReal | ✅ vivo: Editar fuente (primaria) + Editar conversando (dock) disabled + nota del flujo diff→beta→tren |
| RF-96 | tab Corridas honesta (434-443) | inspector.tsx (Corridas) | CorridasHonesta | ✅ vivo: estado honesto + nota de caja `POST …/boxes/builder/run` + «Ver todas» disabled |

Leyenda: ⬜ pendiente · 🔶 implementado sin verificar · ✅ verificado en paridad.

Extra verificado: **dark theme** — drawer completo legible con tokens dark (screenshot);
**consola limpia** al cierre del recorrido (solo debug de vite + info de React DevTools).

## Hallazgos de la validación (corregidos en `6cff1a0`)

1. El fetch de fuente moría en vuelo con red real (dep `fuente?.id` re-disparaba el
   effect y su cleanup lo abortaba) — invisible para la story (fake resuelve en
   microtask), cazado SOLO en la app real. La FASE 3 pagó su costo.
2. El 404 honesto del showcase ensuciaba la consola («Failed to load resource» de
   Chrome, insuprimible). La página ahora conoce el registro S2 (`api.listArneses`) y
   rechaza LOCAL con el MISMO mensaje del daemon — estado idéntico, red limpia.

## Desviaciones registradas (a consultar con el operador)

1. **AA sobre el mockup (RF-88 manda):** `.cond` (badge condición) — texto `--warn` a
   10px sobre claro = 3.76:1 en axe → warn-soft + texto foreground (borde dashed warn
   se conserva). Ídem `.chip-versiona`: `--primary` a 9px = 3.44:1 → texto foreground.
   El a11y addon corre en `error` y design.md fija «contraste AA en ambos temas».
2. **Interpretación RF-81 ↔ RF-84:** «✕ cierra del todo (mapa sin drawer)» convive con
   «sin selección → affordance, no ausencia» así: ✕ quita el overlay expandido Y el
   drawer del NODO; queda el drawer VACÍO (affordance 340px). Única lectura que
   mantiene vivos ambos RF firmados.
3. **Tabs/header sticky también en drawer normal** (el mockup solo lo definía en
   expandido — sus tarjetas de galería no scrolleaban): RF-82 pide «conmutan panes sin
   perder el header»; sticky siempre lo garantiza.
4. **Chips inertes** = `disabled` real (mockup: `data-off`): mismo estilo, semántica nativa.
5. **Showcase sin disco:** `content-studio-full` es fixture embebido → estado honesto
   («arnés sin directorio registrado»), sin round-trip 404; el mockup lo suplía con
   fuente RECONSTRUIDA rotulada. La app jamás reconstruye (decisión #4).
