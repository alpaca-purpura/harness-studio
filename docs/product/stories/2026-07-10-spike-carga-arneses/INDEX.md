# Spike · Mecanismo de CARGA de arneses y proyectos-con-arneses en la app

> `tipo: spike` (D19/D20 gestion_trabajo) — investigación acotada. Estados: abierto · investigando · cerrado.
> **Cierre = decisión documentada** (→ LEDGER), + thin proof opcional. Vig: activo · NO firmado.
> Detonante: el usuario quiere **agencia** — poder elegir desde la app qué arnés/proyecto cargar y mejorar,
> en vez de que entren solo por config de boot.

## Pregunta del spike

¿Cuál es el mecanismo para que el usuario **cargue** en la app (a) un arnés (carpeta única) y (b) un
**proyecto que CONTIENE arneses instalados** (caso real: Vitalia), y que aparezca en el picker+Mapa listo
para chatear/mejorar? ¿Qué existe, qué falta, y cuál es el slice fino que da agencia YA?

## Hallazgo (investigar) — estado del backend de carga

- ✅ `PUT /api/arneses/{id}` {path}: valida (abs · existe · no-protegido) → registra (arnesID→path) →
  `onRegistered`(loader) carga al índice → visible en Mapa. Respuesta honesta `{indexed, detail}` (indexa
  o no según si el dir es arnés reconocible; jamás grafo inventado). `internal/adapters/transport/http/arneses.go`.
- ✅ loader: `LoadArnes(path)` reconoce forma física (plugin|instalado, CAP-15-20) para **un arnés**.
- ✅ FE client: `api.registerArnes(id, path)` + `listArneses()` en `web/src/shared/api/client.ts`.
- ❌ **Falta FE:** dialog para elegir carpeta. Diferido: `session-rail.tsx:258` («Full dialog UX = HS-07»);
  `workspace-stage.tsx:167` = empty-state que pide «carga la carpeta» sin cómo.
- ❌ **Falta detección proyecto-multi:** un repo con VARIOS arneses instalados → leer lock
  `.devstudio/arneses.yaml` y resolver multi-arnés. Diferido (HS-12, deuda BACKLOG).

## Dos caras
- **A · Cargar arnés (carpeta única):** backend listo; falta dialog FE. ← slice fino (agencia hoy).
- **B · Cargar proyecto con arneses:** detectar arneses instalados dentro de un repo real. Más pesado
  (detector multi-arnés HS-12). El spike lo **decide en papel**; build posterior.

## Retomar aquí
> Slice fino en curso: **dialog FE "Cargar carpeta"** → input path → `PUT /api/arneses/{id}` existente →
> arnés en picker+Mapa → sesión → chat. Decisiones de diseño → `decisiones.md`.

## Archivos
- `decisiones.md` — hallazgos + decisiones de diseño del mecanismo (S-D0…).
