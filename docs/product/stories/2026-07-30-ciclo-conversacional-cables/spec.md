# Spec-lite — ciclo conversacional, cables del Dock

> Sancionada por la firma del corte (decisiones.md).

**RF-C.1 (C-T1a)** `web/src/pages/shell/ui/workspace-stage.tsx` (~:476): `onProponer`
suma `useSessions.getState().openChat()` + `setScope({nodeId, clase, fuentePath})` de
la caja del punto (`caja_id` → lookup en `graph.nodos`; patrón :372-379). Solo corre
cuando `viewedId===arnesId` (ya garantizado por `proponerDeshabilitado`). Composer
prellenado con foco, nada se envía. Descartar NO se toca (ya cableado); el DELETE
«volver a mostrar» sigue como deuda declarada en la tarjeta.
**RF-C.2 (C-T1b)** `web/src/widgets/map-canvas/ui/inspector.tsx`: prop
`onEditarConversando?: () => void` en `InspectorProps` y `Contenido` (:663-684);
habilitado ⇒ onClick; disabled ⇒ title honesto nuevo. `workspace-stage.tsx` pasa la
prop SOLO si `viewedId===arnesId`. Stories del inspector con/sin prop (gate a11y).
**RF-C.3 (stretch C-T2)** watcher instalaciones per C-D3 + frame
`{install_path, deriva}` + 4º callback `onPortafolio` en `sse.ts` + refetch de
`portafolio-view`. Riesgo ráfagas: goroutine ya serializa; debounce declarado, no
construido.

Capabilities: change_log en `fe-mapa/capa-mejora.yaml` · `fe-chat/acotar-alcance.yaml`
· `fe-mapa` inspector (tab contenido) — sin capability nueva (cableo de existentes);
C-T2 tocaría `portafolio/evaluar-deriva.yaml` + `http-sse` + `fe-portafolio`.

AC: (a) click «Proponerlo en el chat» ⇒ Dock abre, chip de alcance = caja del punto,
composer prellenado, foco al final, NADA enviado; segundo click no repone propuesta
consumida; (b) «Editar conversando» vivo solo con arnés-de-la-sesión visto; disabled
honesto en caso contrario; (c) stories verdes con a11y; (d) verificación en navegador
contra daemon real + laptop.
