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
3. **Decisiones de forma ING-D11..D15 FIRMADAS 🧑‍⚖️** (2026-08-01, «Sigo todas tus propuestas»):
   dos gates con ingesta primero · ingesta = tabla de decisión, no wizard de 6 pasos · primer
   gesto del editor = «esto es un paso del proceso» · Base colapsada por autoría · escenario con
   cifras medidas de vitalia-app.
4. **ING-D16..D19 FIRMADAS 🧑‍⚖️** (2026-08-01, elección sobre opciones renderizadas): escritura
   **híbrida** desde el Mapa (determinista→Go · redactar→conductor) · el muro se ataca **en la
   ingesta** primero · el eje de ING-D3 se llama **`autoría`** (`del-plugin`·`propia`·`suelta`·
   `derivada`·`de-referencia`) · **T0 sale ya**, solo el fallback de lectura al canónico.
5. **Próxima etapa: MOCKUP 1 · INGESTA** (gate 🧑‍⚖️ antes de spec) — **DESBLOQUEADO**. Siguen
   abiertas Q5 (facetas) · Q6 (`CLAUDE.md` no reconocido) · Q7 (pasos que no son cajas, es del
   mockup 2); el mockup 1 dibuja Q5/Q6 como propuesta rotulada. Superficies, en orden por ING-D11:
   - **(1) Ingesta**: apuntar a carpeta cruda → inventario → autoría → facetas → curaduría. Hoy el
     wizard es callejón sin salida (`portafolio-wizard.tsx:308-327`).
   - **(2) Mapa editor**: crear/cablear/reapuntar manual con clicks → escribe archivos. Superset
     ESTRICTO de la línea base — **arrancar leyendo [`mockups/INDEX.md`](../../../../mockups/INDEX.md)**
     (SSoT UI = Storybook; no pisar vocabulario L0 — ver Q4).
6. **T0 · FIRMADO y desacoplado (ING-D19)**: el loader cae al `arnes.yaml` del **canónico** vía
   Portafolio cuando la instalación no lo trae. Cura las instalaciones que YA existen (incluida
   la que se arregló a mano en vitalia-app) sin escribir una línea en disco ajeno. Sale como
   bugfix propio ANTES del mockup.

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
| Auditoría previa (verificación independiente + medición del loader) | ✅ escrita — 5 correcciones al informe, Q1-Q8 abiertas |
| Decisiones de forma del mockup (ING-D11..D15) | ✅ FIRMADAS 🧑‍⚖️ 2026-08-01 |
| T0 quick-win `arnes.yaml`→instalaciones | ⬜ diseño chico pendiente (§PENDIENTES P1 · Q8) |
| **Mockup 1 · ingesta** (ING-D11 partió el gate en dos) | ⬜ **próxima** — gate 🧑‍⚖️ · **bloqueada por Q1-Q2** |
| Mockup 2 · Mapa editor | ⬜ tras mockup 1 firmado — gate 🧑‍⚖️ |
| Spec | ⬜ tras mockups firmados |
| Implementación (slices) | ⬜ |
| PARIDAD | ⬜ |
