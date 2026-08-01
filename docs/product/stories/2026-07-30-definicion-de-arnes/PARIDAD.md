# PARIDAD — crear · visualizar · mantener arneses multi-actividad (carril D)

> vig: 2026-08-01 · spec v2 FIRMADA (DEF-D8) · evidencia por ticket, verificada
> por el constructor con sus propios mecanismos (directiva del /goal). Gate 🧑‍⚖️
> final del operador: casillas al pie.

## Por ticket

| Ticket | Estado | Evidencia |
|---|---|---|
| MA-T1a `caja:` en PasoSpine | ✅ | `forja/parser.go` + `TestParseSemillaCajaOpcionalEnPaso` + golden `TestParseSemillaRealSinCajas` (la semilla NO declara cajas — aviso si cambia). Commit `1acb730` |
| MA-T1b seam al indexar + enmienda schema + espejo TS | ✅ | `loader/actividades.go` (CAP-151): derivación AL INDEXAR (patrón edges); 6 tests incl. **grafo nuevo Y degradado validan contra `graph.l0.schema.json`** (enmienda aditiva + `degradado` regularizado); espejo `types.ts`. Suite Go 30/30 ok. Commit `1acb730` |
| MA-T2 saneo vitalia | ✅ | Clave legacy `vitalia` (residuo re-key) fuera de `arneses.json` (backup `.bak-saneo-MA-T2`); el `Rebuild` del boot limpió el índice SOLO (cero cirugía de DB). `vitalia-app` agregado por el **wizard real** → identidad provisional sin-home. `verificacion-e2e/portafolio-antes.png` |
| MA-T3 forja `developer-vitalia` | ✅ con desviación | Conversacional: sesión real del dock verificó inventario + resolvió 2 desvíos (sin `.mcp.json`; hook `Stop` = del proyecto). **Materialización mecánica por el constructor** (el gate de permisos bloqueó la escritura del conductor — hallazgo D, deuda en BACKLOG). Publicado: commits `e4cfb9a`+`2de8426`+`80923cd` del marketplace, tag `developer-vitalia/v0.1.0`; `check_marketplace_shape` verde. «Traer canónico» desde la app → checkout `al-hilo`. **Wire E2E:** `arnes.actividades[]` (historia 4 · bugfix 4 · spike 2 E13 · revisar-capability 0 E7) + facetas (`dev-team`×2 · `commit-push`×2) servidos por el daemon instalado `0.6.0.2608010346` |
| MA-T4 chips N0 | ✅ | `actividad-chips.tsx` (CAP-152) + `verificacion-e2e/e2e-n0-chips-developer-vitalia.png`: 4 chips + `sin-actividad · 1` (= `hipaa-check`, E6 real) + E7 degradado; salud worst-of derivada |
| MA-T5 foco N1 + inspector | ✅ | `verificacion-e2e/e2e-n1-bugfix-secuencia.png`: secuencia ①→④ sólida `--primary`, 3 dimlanes, breadcrumb+«ver todo»+disclaimer; `e2e-spike-e13-ghost.png`: ghost «② decidir — paso sin caja aún» + 5 dimlanes (spike no toca NINGÚN carril — E5 extremo y honesto); badges ×2 en `dev-team`/`commit-push` (MA-L4). Fila «Actividades» en inspector (story con play) |
| MA-T6 mockup | ✅ firmado | Gate 🧑‍⚖️ DEF-D8; `mockup-mapa-actividades.html` + 6 capturas en `verificacion-mockup/` |
| MA-T7 stories + a11y | ✅ | 8 stories nuevas (`actividades.stories.tsx`) con gate a11y ENTERO; **13 de map-canvas + 5 de map-bar intactas y verdes** (26/26 re-corridas por el constructor, 4.5 s headless); suite completa 472/472; unit 162/162 (13 nuevos); tsc/biome/stylelint/depcruise/steiger limpios |

## §8 de la spec — checklist

- [x] 13 stories del canvas (+5 map-bar) verdes SIN tocar — MA-L3 medible.
- [x] Enmienda del schema: grafo vigente Y grafo nuevo (actividades+degradado) validan; `VerificarSpine` intacto (suite conformance verde).
- [x] Stories nuevas N0/N1/E5/E7/E13 con a11y entero.
- [x] Dogfood doble: semilla propia (historia+spike SIN cajas, golden) + `developer-vitalia` REAL servido por el daemon instalado.
- [x] E2E: publicado → Portafolio canónico → Mapa → foco `bugfix` muestra su procedimiento → spike termina antes.

## Desviaciones y hallazgos (visibles, jamás silenciosos)

1. **La forja conversacional no pudo ESCRIBIR** (hallazgo D → BACKLOG): sesión
   sobre arnés sin sello = `PermissionSet` vacío = sin canal HITL = CC auto-niega.
   El plan se acordó conversando; la materialización fue del constructor.
2. **Alta de plugin nuevo sin camino a canónico** (hallazgo E → BACKLOG): push
   manual formato-B2 (B2 solo re-publica canónicos existentes).
3. **Faceta inválida tira el lienzo** (hallazgo F → BACKLOG): mi `arquetipo:
   guiado` inválido crasheó el canvas vigente; corregido en dato (`80923cd`).
4. **Foco en modo peek («vista previa») es parcial**: dim + ×N sí; secuencia,
   dimlanes y breadcrumb solo en sesión propia del arnés. Anotado; decidir si
   es limitación aceptada o fix.
5. Desviaciones FE↔mockup declaradas por el implementador (7, menores — conteo
   por cajas distintas, E7 `disabled` real, texto E6 al `title`, artchips sin
   dim bajo foco, `.xn` oculto en compact, enums corregidos) — detalle en el
   reporte del agente, reproducido en `decisiones.md` si el operador lo pide.
6. La entrada del Portafolio de `developer-vitalia` quedó bajo «sin empresa»
   (el traer no propaga `empresas` del sello al listado) — menor, visible.

## Gate 🧑‍⚖️ (operador)

- [ ] Chips N0 + foco N1 en la app instalada (sesión propia de `developer-vitalia`)
- [ ] `developer-vitalia` 0.1.0 en el marketplace propio (repo + tag)
- [ ] Deudas D/E/F aceptadas como backlog
