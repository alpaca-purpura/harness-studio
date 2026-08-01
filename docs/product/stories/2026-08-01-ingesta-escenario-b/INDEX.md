# Ingesta escenario B — jalar proyectos crudos + Mapa editor

> **Paquete de trabajo** (metodología §10) · abierto 2026-08-01 · estado: **decisiones FIRMADAS,
> próxima etapa = mockup**. Origen: conversación del operador post-fix actividades (2026-08-01).

## Retomar aquí

1. **Decisiones ING-D1..D10 FIRMADAS 🧑‍⚖️** (aprobación conversacional 2026-08-01, transcrita en
   [`decisiones.md`](./decisiones.md)) — la dirección está cerrada, NO relitigar.
2. **Estado real auditado** con evidencia archivo:línea → [`informe-estado-real.md`](./informe-estado-real.md)
   (2 exploraciones: anatomía de `~/Proyectos/vitalia-app` como espécimen de ingesta + maquinaria
   existente en arnesia). Leer ANTES de diseñar — dice qué existe y qué no, sin filtro.
   ⚠ **Leer JUNTO con [`auditoria-previa-mockup.md`](./auditoria-previa-mockup.md)** (2026-08-01):
   verificación independiente que confirma los 15 gaps, **corrige 5 datos del informe** (C1-C5) y
   agrega la medición que faltaba — el loader corrido de verdad contra los dos árboles: la
   distancia crudo→arnés son **5 bloques de frontmatter + 2 archivos**, y **90-95 % de los nodos
   caen en una sola banda** (muro que el baseline no cubre: su fixture mayor tiene 38 nodos).
   Ahí viven las **8 preguntas abiertas (Q1-Q8)** y las **5 propuestas (P-A..P-E)** previas al gate.
3. **Próxima etapa: MOCKUP** (gate 🧑‍⚖️ antes de spec) — **bloqueado por Q1-Q3** de la auditoría
   (quién escribe al hacer click · cómo se ataca el muro de 100 nodos · uno o dos gates). Dos superficies:
   - **Ingesta**: wizard/flujo al apuntar a carpeta cruda (inventario → procedencia → clasificación
     → curaduría). Hoy el wizard es callejón sin salida (`portafolio-wizard.tsx:308`).
   - **Mapa editor**: crear/cablear/reapuntar manual con clicks → escribe archivos. Superset
     ESTRICTO de la línea base — **arrancar leyendo [`mockups/INDEX.md`](../../../../mockups/INDEX.md)**
     (SSoT UI = Storybook; no pisar vocabulario L0).
4. **Quick-win desacoplado (T0)**: propagar `arnes.yaml` a instalaciones — hoy Identificar solo
   escribe el l0 y las instalaciones quedan ciegas de actividades (bug vivido 2026-08-01, fix
   manual hecho en vitalia-app). Diseño chico en `decisiones.md` §PENDIENTES → puede salir como
   bugfix propio antes del mockup grande.

## Qué es (1 párrafo)

Dos puertas de entrada al Portafolio: **A)** proyecto con arnés formal ya instalado (lo que arnesia
hace hoy) y **B)** proyecto crudo que hace las cosas bien → inventariar TODO su as-code (skills ·
rules · agents · hooks · commands, por-plugin/propios/sueltos/drifteados), leer su grafo de
punteros real, clasificar (knowledge · knowledge-as-code · docs-as-code · process-as-code ·
WIP-home), curar conversacional Y manualmente desde el Mapa, sellar, empaquetar, publicar al
marketplace propio, **reinstalar formal** → recién ahí telemetría y mejora. **B = 90 % del uso
real.** El Mapa deja de ser visor: es EDITOR (clicks crean/cablean → archivos reales); el chat
asiste, no reemplaza.

## Relación con paquetes existentes

- **Absorbe/reencuadra** [`stories/2026-07-10-forja-ciclo-vivo/`](../2026-07-10-forja-ciclo-vivo/INDEX.md)
  (⏸ pausada): forjar-de-cero pasa a ser SUBCASO de la ingesta; el gate D19 y el scaffolder
  siguen siendo piezas del motor. El Slice 1a construido (golden scaffold) no se tira.
- **Extiende** «Identificar» (S1-D28, Slice 2 firmado): la spec de definición-de-arnés ya nombró
  «Identificar actividades» como futuro (E8) — este paquete ES ese futuro, generalizado.
- **Se apoya en** el modelo de terreno D0-D20 firmado (regla de las 3 caras, 11 canónicas) y en
  DEF-D2 v3 (arnés = paquete de UN puesto; actividades = pipelines internos).
- **Espécimen de dogfood**: `~/Proyectos/vitalia-app` (169 archivos as-code, mitad plugin/mitad
  cadena de proceso propia, grafo `DOCS-GRAPH.md` autogenerado, 3 desalineaciones sello↔realidad).

## Etapas (disciplina §10)

| Etapa | Estado |
|---|---|
| Decisiones de dirección (ING-D1..D10) | ✅ FIRMADAS 🧑‍⚖️ 2026-08-01 |
| Informe de estado real (research) | ✅ escrito |
| T0 quick-win `arnes.yaml`→instalaciones | ⬜ diseño chico pendiente (§PENDIENTES) |
| Mockup ingesta + Mapa editor | ⬜ **próxima** — gate 🧑‍⚖️ |
| Spec | ⬜ tras mockup firmado |
| Implementación (slices) | ⬜ |
| PARIDAD | ⬜ |
