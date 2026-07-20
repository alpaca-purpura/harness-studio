# Spike de spec — El sello + Identificar + Mapa degradado

> `tipo: spec` (spike de enfoque) · paquete `2026-07-17-portafolio-slice2-sello-identificar` ·
> arquitecto (Fable 5, 2026-07-17). Cierra los 3 forks técnicos que las decisiones **S1-D27..D29**
> dejaron abiertos, ANTES de código. Todo verificado contra el código REAL en `main`
> (`internal/{domain,usecase,adapters}/…`, `web/src/…`). El ejecutor NO relitiga; si un fork resulta
> inviable en código, documenta el porqué aquí y elige la alternativa más cercana al espíritu.

## 0. Norte

Visión firmada [[hs-vision-sello-jalar-arneses]]: `arnes.l0.json` = **el sello de la fábrica**; la
etapa es **jalar** arneses que ya funcionan y **sellarlos** para mejorarlos acá. Este slice construye
el circuito mínimo: *ver sin sello (degradado) → sellar (Identificar) → mapear como arnés propio*.

## 1. Estado real (verificado)

- **Reconocer ≠ sellar.** `nomenclatura-arnes.md` §1: un `.claude/` poblado ES un arnés reconocible
  (forma-instalada). §2: sin manifiesto se carga **degradado honesto** («nodos sin fases ni spine,
  check `manifiesto-ausente` rojo — visible, nunca inventado»). Ese degradado **el código nunca lo
  implementó**.
- **Loader** (`internal/adapters/loader/loader.go`): `leerManifiesto` sin `arnes.l0.json` ni
  `plugin.json` devuelve `(nil, "", nil)` — silencioso, SIN señal de degradado. Igual carga los nodos
  que encuentra (rules/skills/hooks); solo `g.Arnes` queda `nil`.
- **ObservarEnMapa** (`internal/usecase/portafolio.go:261`): exige `g.Arnes != nil`; si es nil →
  `400 "no resolvió un arnés cargable"` (la pared donde el operador chocó). Nunca sintetiza degradado.
- **Índice del Mapa** (GAP-2, S0-D6/S1-D2): `index.Store.Upsert` keyea por `g.Arnes.ID` pelado — dos
  identidades con el mismo `id` se pisan; un degradado sin id no tiene llave.
- **Llave del Portafolio** (`domain.IdentidadArnes.Clave()`, `portafolio.go:30`):
  `slug(home)~slug(id)~slug(scope)`. Carpeta cruda escaneada en su raíz → `home=""`, `id=""`,
  `scope="."→""` → clave **`sin-home~~`**, idéntica para cualquier proyecto crudo → corrupción
  silenciosa por `Upsert`.
- **HTTP** (`adapters/transport/http/portafolio.go:125`): mapea el nil-Arnes al `400` default (sin
  sentinela propio). **FE** (`web/src/widgets/portafolio/ui/portafolio-drawer.tsx`): botón «Observar»
  solo gateado por sesión activa (S1-D13), no por cargabilidad → clic garantiza el `400`.

## 2. Fork A — Huella de path unificada (S1-D29 + GAP-2, una sola solución)

**Problema doble:** (a) la clave del Portafolio degenera a `sin-home~~`; (b) el índice del Mapa no
tiene llave para un degradado sin id. **Ambos** necesitan un discriminador estable cuando no hay
identidad — y es el MISMO: la ruta física.

**Diseño recomendado.** Un `Disc` = hash corto y estable de la ruta canónica de instalación
(`EvalSymlinks`+`Clean` → `sha256` → primeros N hex). Un solo concepto, dos usos:

- **Portafolio:** campo nuevo opcional `IdentidadArnes.Disc string json:"disc,omitempty"`, poblado en
  la resolución SOLO cuando `Home+ID+Scope` quedarían todos vacíos. `Clave()` lo incorpora como último
  segmento únicamente en ese caso degenerado (las claves ya sanas NO cambian → cero migración de lo
  existente). Resultado: dos roots crudos distintos → dos claves distintas.
- **Mapa:** `ObservarEnMapa` pasa una llave estable al índice; `index.Store.Upsert` keyea por
  `g.Arnes.ID` si es no-vacío, si no por la llave sintética recibida. Cierra el **fallback** de GAP-2
  (no el re-key total por `(home,id)` — eso queda deuda mayor, fuera de este slice).

**Alternativa descartada:** re-key completo del índice por compuesto `(home,id)`. Más correcto, mucho
más grande; se difiere. El fallback por huella basta para el circuito degradado.

## 3. Fork B — Modo degradado en el Mapa (S1-D27, honra contrato §2)

**Diseño recomendado.**

1. **Loader emite la señal** (fuente única de verdad de «esto está degradado»): cuando no hay ningún
   manifiesto, marcar el grafo — p.ej. `domain.Graph.Degradado bool` + un check `manifiesto-ausente`
   (rojo). Hoy la ausencia es silenciosa; que el loader la NOMBRE evita que cada consumidor la
   reinvente.
2. **ObservarEnMapa deja de fallar:** con `g.Arnes == nil`, sintetiza un `Arnes` mínimo
   `{ ID: <llave sintética Fork A>, Nombre: <entrada.Nombre | último segmento del path>, Degradado:
   true }` y lo indexa bajo la llave sintética. Los nodos que el loader SÍ encontró (rules/skills/
   hooks) se pintan igual — el grafo es fino, no inventado.
3. **HTTP/FE:** el botón «Observar» se habilita para entradas sin sello; el Mapa pinta la marca roja
   `manifiesto-ausente`; el drawer muestra el estado degradado + el CTA «Identificar» (Fork C).

Esto **implementa doctrina ya firmada** (§2), no política nueva.

## 4. Fork C — «Identificar» V1 (S1-D28, el verbo que faltaba)

**Diseño recomendado (V1 mínimo = el scaffold que se hizo a mano para vitalia).**

- **Usecase:** `Identificar(ctx, clave, installPath, {id?, nombre?})` escribe `arnes.l0.json` en la
  ruta canónica de la entrada. Defaults: `id = slug(último segmento del path)`, `nombre = id`,
  `empresas = entrada.empresas`, `reporta_a: null`, `version: "0.1.0"`. Sin asistente elaborado.
- **SEGURIDAD (crítico — primera escritura del Portafolio fuera de `~/.arnesia`, dentro de un
  proyecto del usuario):** (1) pasa por la política de path protegido (rechaza `~/.claude`, etc.,
  misma que `ArnesRegistry.checkProtected`); (2) **NO pisa** un `arnes.l0.json` existente (error
  explícito o confirmación); (3) el FE confirma antes de escribir. El sello se escribe en la FUENTE —
  ese es el punto de «jalar y sellar».
- **Post-escritura:** re-escanea → la entrada resuelve con id real → **re-key** de la entrada en el
  store (la vieja clave degenerada migra/se reemplaza, no se duplica).
- **Endpoint:** `POST /api/portafolio/arneses/{clave}/identificar`.
- **FE:** botón «Identificar» en entradas degradadas + form mínimo (`id`+`nombre` prefill, editables)
  + confirm. Al sellar, la entrada deja de ser degradada.

**FUERA DE ALCANCE V1 (flag al operador, alinea con la visión S1-D27):** «extraer los skills de
alguien → plugin *nuestro* reutilizable entre orgs» es el camino grande de clonar/extraer-a-
marketplace = **Publicar (S3)** + inicializador universal. V1 solo sella IN-SITU («jalar y sellar
donde está»). La extracción-a-plugin-reutilizable es un slice posterior.

## 5. Superficie de capabilities (SSoT, R2 — el ejecutor ancla los slugs reales)

- `portafolio/identificar.yaml` — sellar in-situ (nuevo).
- `portafolio/observar-mapa` — modificar: aceptar degradado (o `observar-mapa-degradado.yaml` nuevo).
- `portafolio/identidad` — modificar: huella de path (o `identidad-huella-path.yaml`).
- `mapa/index-llave-fallback.yaml` — fallback de llave sintética (GAP-2 parcial).

## 6. Plan de tickets (orden por costo/impacto: D29 → D27 → D28)

- **T1 (S1-D29):** huella de path en `IdentidadArnes` + `Clave()` + resolución de identidad; test de
  no-colisión de dos roots crudos distintos. *Seguridad de datos, chico.*
- **T2 (S1-D27):** señal `manifiesto-ausente` en loader + `Graph.Degradado`; `ObservarEnMapa`
  sintetiza degradado + llave fallback en el índice; HTTP deja de `400` en manifest-less; FE marca
  roja + botón habilitado. *Reabre y cierra (fallback) GAP-2.*
- **T3 (S1-D28):** usecase `Identificar` + endpoint + writer con las 3 guardas + re-key; FE botón +
  form mínimo. *La feature más grande.*
- **T4:** PARIDAD (mockup/stories = SSoT) + capabilities graduadas + cifras regeneradas
  (`estado.sh`/`conformance --todo`). Firma 🧑‍⚖️ por ticket, como Slice 0/1.

## 7. Decisiones del operador — RESUELTAS (firmadas de palabra, 2026-07-17)

1. **Entrada degradada → PERSISTE** en el Portafolio (se puede agregar sin sello; coherente con
   Opción A / «mostrar todo honestamente»).
2. **«Identificar» V1 → sella IN-SITU**, sin clon. Fundamento del operador: clonar a un dir «canónico»
   crearía justo la **duplicidad de información** que confunde al usuario inexperto; y la **seguridad
   ante cambios ya la da git** del propio proyecto — ArnesIA no necesita una 2ª copia para eso en V1.
   Nota para el ejecutor: sellar in-situ un dir SIN identidad previa **no viola la ley anti-drift** —
   no se edita una instalación downstream de un canónico, se CREA la identidad (el dir pasa a ser un
   arnés, candidato a su propio canónico). La extracción-a-plugin-reutilizable-entre-orgs sigue fuera
   de V1 (Publicar/S3).
3. **Nombre → Slice 2** (confirmado).

## 8. Progreso — TODO CONSTRUIDO Y VERIFICADO (2026-07-17, sin commitear)

- **T1 (S1-D29) — HECHO.** `HuellaPath` + `IdentidadArnes.Disc` + `Clave()` desempata el caso
  anónimo + `ResolverIdentidad(installPath)` + caller. Test `TestClaveDesempataPorPath`. Prueba viva:
  dos roots crudos → `sin-home~~d5013737b2d1` vs `…8121a288967c` (antes ambos `sin-home~~`).
- **T2 (S1-D27) — HECHO.** `Graph.Degradado` + el loader lo marca (`manifiesto-ausente`) +
  `ObservarEnMapa` sintetiza arnés mínimo con la huella como llave + HTTP deja de 400. FE: sección
  roja «sin sello» en el drawer. Tests `TestLoaderSinManifiesto` (ext.) · `TestObservarEnMapaDegradadoSintetiza`.
- **T3 (S1-D28) — HECHO.** `Identificar` (writer in-situ + 3 guardas + re-key vía `entradaDeCandidato`
  compartido) + `POST …/identificar` + `domain.Slug`. FE: botón «✦ Identificar» + form (id·nombre) +
  cliente `identificarArnes` + wiring de página. Tests `TestIdentificarSellaYRekey` ·
  `TestIdentificarNoPisaSelloExistente` + story `IdentificarSinSello`.
- **T4 — HECHO.** Capability nueva `portafolio/identificar` + `observar-en-mapa` extendido (degradado);
  `registrar-identidad` extendido (T1). Cifras regeneradas (`estado.sh`): 93 caps · 50 vivo · cobertura
  100%. Conformance `257 · pass 48 · fail 0`, `cap-ptr-resuelve`/`cap-sin-huerfano` verdes.

## 9. Verificación (mis propios mecanismos)

- **Go:** `go build ./...` limpio · `go test ./internal/...` verde (incl. los 5 tests nuevos) · `go vet` limpio.
- **Conformance:** `257 checks · pass 48 · fail 0 · error 0` · R1/R2 (integridad+cobertura) PASS.
- **FE:** `pnpm run verify` (typecheck·lint·depcruise·fsd·stylelint) limpio · `pnpm test` **156/156**
  (incl. story `IdentificarSinSello` y `IdentidadProvisional` con la sección nueva).
- **E2E contra el daemon REAL** (`scratchpad/e2e_sello.sh`, HOME temporal, loader real, no fakes):
  crudo → escanear (`sin-home~~<huella>`) → agregar → **observar degradado HTTP 200** (antes 400) →
  identificar (sella + re-key a `miproj`) → sello en disco → re-escanear (`sin-home~miproj~`) →
  observar sellado (arnés completo). **Circuito completo VERDE.**

Falta SOLO la firma 🧑‍⚖️ del operador (gate PARIDAD) + el commit — decisión suya.
