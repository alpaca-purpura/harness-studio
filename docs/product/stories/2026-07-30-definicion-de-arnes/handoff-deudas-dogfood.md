# HANDOFF — resolver las deudas de dogfood D · E · F (+ gotcha peek) de este paquete

> vig: 2026-08-01 · para retomar en conversación NUEVA. Contexto mínimo suficiente:
> este archivo + `PARIDAD.md` + `informe-forja-conversacional.md` + `decisiones.md`
> §HALLAZGOS. Todo lo de abajo está VERIFICADO (código leído + reproducido en vivo),
> no es hipótesis. main @ `5a0f11d`, suite Go 30/30, stories 472/472.

## Estado vivo al momento del handoff

- **Daemon instalado** `~/.local/bin/arnesia` sello `0.6.0.2608010346` (seam de
  actividades + FE N0/N1 embebidos). Estaba corriendo backgrounded en `:4200` —
  en sesión nueva relanzar: `ARNESIA_REPO=~/Proyectos/harness-studio ~/.local/bin/arnesia serve`.
- **`developer-vitalia` 0.1.0**: publicado en `~/Proyectos/prenter-marketplace`
  (tag `developer-vitalia/v0.1.0`, commits `e4cfb9a`+`2de8426`+`80923cd`), clone
  de CC pulleado, checkout canónico en `~/.arnesia/checkouts/prenter-marketplace/
  developer-vitalia`, entrada de Portafolio `al-hilo`, indexado, Mapa con chips vivo.
- **Sesiones del dock que YA existen** (no recrear): la de la forja bloqueada
  sobre `sin-home~~github-com-alpacapurpura-vitalia-app` (transcript del bloqueo:
  `~/.claude/projects/-home-chalreme-Proyectos-vitalia-app/dc1a7e2b-*.jsonl`) y
  una sesión propia de `developer-vitalia`.
- Backup del saneo MA-T2: `~/.arnesia/arneses.json.bak-saneo-MA-T2`.
- Gate 🧑‍⚖️ de PARIDAD del carril D sigue pendiente del operador (independiente
  de estas deudas).

## Orden de ataque sugerido (chico→grande)

**F (FE puro, ~horas) → peek (FE chico) → D (diseño+backend) → E (diseño+backend+UI).**
D y E llevan decisión del operador ANTES de codear (preguntas al final).

---

## Deuda F · una faceta inválida tira el lienzo ENTERO

**Repro real:** publiqué `arquetipo: guiado` (no existe; enum = `pipeline ·
excepcion · abierto · no-arnesar`, `internal/domain/box.go:64-79`) → el canvas
completo cayó al ErrorBoundary: «El lienzo no pudo renderizar este arnés —
Cannot read properties of undefined (reading 'label')».

**Punto exacto:** `ARQUETIPO_MARK[arquetipo].label` en el bloque `.node-meta` de
`web/src/entities/arnes/ui/arnes-node.tsx` (el mapa vive en
`entities/arnes/model/node-view.ts`). Mismo riesgo en los lookups hermanos:
`GATE_TONE[gate.tipo]`, perfil, y cualquier `MAP[facet].algo` sin guard.

**Dirección de fix (§4.5 reconciliación honesta — degradar el NODO, jamás el
mapa, jamás inventar):** lookup con fallback visible — p. ej. marca `?` con
`title="arquetipo no reconocido: guiado"` en tono `--warn`. Auditar TODOS los
lookups por facet en `arnes-node.tsx`/`node-view.ts`/`inspector.tsx`.

**Verificación:** story nueva tipo `NodoMalformado` pero con `arquetipo`/`gate`
fuera del enum y assert de que el canvas SIGUE VIVO (`.lane` > 0) + el nodo
muestra la marca degradada; gate a11y entero; las 13+5+8 vigentes intactas.
Bonus dato-real: apuntar el loader a un fixture con facet inválida.

---

## Gotcha peek · el foco de actividad es PARCIAL en «vista previa»

**Verificado en vivo** (misma build, mismo arnés):
- Peek («Abrir en Mapa» desde Portafolio, `viewedId ≠ arnesId`): dim ✓, badge ×N ✓,
  pero `{seqPaths:0, circles:0, dimlane:0, crumb:false}`.
- Sesión propia: `{circles:4, texts:[1,2,3,4], dimlane:3, crumb:true}` ✓.

**Dónde mirar:** `web/src/pages/shell/ui/workspace-stage.tsx` — es el dueño del
estado `actividadFoco/actividadPre` (patrón capa/artefactos). El implementador
gateó el CTA E7 a `viewedId === arnesId` (CH-D6, correcto); sospecha: el gating
alcanzó también props del foco (secuencia/dimlane/breadcrumb) o la rama peek
monta MapBar/EdgeLayer sin esas props. Confirmar leyendo el wiring de peek.

**Decisión previa (operador):** ¿el foco COMPLETO debe funcionar en peek?
Recomendación: sí — es lectura pura (MA-L6), no toca el guardrail CH-D6 (que
solo aplica al CTA de chat). Si se acepta: pasar las props también en peek +
story de regresión peek-con-foco.

---

## Deuda D · sesión sobre arnés SIN SELLO = read-only sin HITL

**Causa verificada en código:** `internal/adapters/agent/claudecode/conductor.go`
→ `permissionArgs()` — con `PermissionSet` de valor cero devuelve `nil`: NO emite
`--permission-mode default` NI `--permission-prompt-tool stdio` (comentario
literal: «A zero-value set emits NO flags (Dock unchanged)» — fue deliberado).
Sesión sobre arnés sin sello → sin rol → set vacío → el conductor corre headless
SIN canal de permisos → CC auto-niega Write («requested permissions … but you
haven't granted it yet») y el prompt **jamás llega al panel**.

**Segundo bloqueo, SEPARADO (ojo, no mezclar):** el validador interno de CC negó
`mkdir` con «may only create directories in the allowed working directories:
'/home/chalreme/Proyectos/vitalia-app', '/home/chalreme/.arnesia/knowhow'» —
**aun con targets DENTRO de esos dirs** (hasta un probe dentro de knowhow falló).
Huele a bug/quirk de la versión de CC del operador con `--setting-sources
project,local` o el modo headless sin prompt-tool. Reproducir aislado ANTES de
diseñar encima (puede desaparecer al cablear el prompt-tool; si no, es issue de
CC aguas arriba).

**Dónde se puebla el PermissionSet:** seguir `SpawnOpts.Permisos` desde
`internal/usecase/session_service.go` (~línea 626, `Cwd`) hacia atrás — quién
deriva el rol del sello y qué pasa cuando `Arnes == nil`. El sistema de roles/TTL
es de Fase E (HS-08/HS-11): `internal/adapters/permission/provisioner.go`.

**Opciones de diseño (decisión del operador ANTES de codear):**
- **(a) Rol de forja**: perfil explícito para sesiones sobre material sin sello
  (allow lectura amplia + escritura SOLO bajo un subdir de forja declarado, todo
  vía HITL). Más diseño, más seguro.
- **(b) Canal siempre cableado**: `permissionArgs` emite `--permission-mode
  default --permission-prompt-tool stdio` SIEMPRE (set vacío = todo pasa por el
  panel, nada auto-allow). Cambio de 5 líneas + fitness: OJO, los tests de
  `docs/architecture/fitness` asertan `SpawnArgs` («the flags ARE the enforcement
  surface») y el boundary `permisos-gui` — hay que actualizar el contrato as-code
  en el MISMO commit, no silenciarlo.
- Recomendación: (b) ahora (destraba la forja conversacional con el humano en el
  loop), (a) como evolución con el modelo de roles.

**Verificación (la de verdad):** repetir la forja E2E — sesión del dock sobre
material crudo, pedir UN Write → **la tarjeta de permiso APARECE en el panel**,
aprobar, el archivo aterriza en disco. Usar la skill
`verificando-binario-instalado` para el lado daemon (build+serve+HTTP real, HOME
sandbox). El criterio de cierre escrito en el informe: «la próxima forja debería
poder hacerse ENTERA desde el chat».

---

## Deuda E · el ALTA de un plugin nuevo no tiene camino a canónico

**El círculo, verificado:**
1. Canónico = dir dentro de un checkout: `esCanonico = contieneEn(checkouts,
   h.Dir)` — `internal/usecase/portafolio.go:111` (RN-IDENT-4).
2. El wizard NO escanea `~/.arnesia`: `POST /api/portafolio/escaneos` → 400
   «root se superpone con una ubicación protegida» (reproducido).
3. B2 `Publicar` exige canónico previo: guard `SinCanonico` en
   `MarketplaceService.Publicar` (RF-B2.3; spec en
   `stories/2026-07-30-volverlo-de-arnesia-y-publicar/spec.md` §B2).

**E-bis (sincronización):** el «↻ Refrescar» del catálogo NO hace fetch — el
`LectorLocal` lee `m.InstallLocation` = el clone que administra CC
(`~/.claude/plugins/marketplaces/prenter-marketplace`), ver
`internal/adapters/marketplace/catalogo_local.go:29`. Hoy sincronizar = `git
pull` manual en ese clone. (El `LectorRemoto` vía `gh api` existe como fallback
— catalogo_remoto.go — pero el local gana si el installLocation existe.)

**E-ter (formato):** `catalogo.json` del marketplace es mono-kit (canales/
versiones de UN plugin, «instancia 0»). El alta de un segundo plugin no puede
anotar versiones ahí sin decidir el esquema multi-plugin. En el alta manual lo
dejé INTACTO a propósito; `developer-vitalia` versiona solo por `plugin.json` y
el lector lo tolera (probado: catálogo mostró 3 entradas).

**Opciones de diseño (decisión del operador):**
- **(a) «Adoptar como canónico»**: acción del Portafolio que toma un dir local
  (p. ej. el forjado en el proyecto), lo copia al checkout del marketplace-home
  declarado en su sello y sella la entrada con canónico. Luego B2 publica normal
  (la situación pasa a «mi copia adelantada»). Rompe el círculo SIN tocar B2.
- **(b) B2-alta**: extender `Publicar` para la situación «existe local, falta en
  marketplace» (el paquete AG ya la dibujó: 6 situaciones con acción Publicar).
  Toca guards + RMW marketplace.json para fila nueva (el publisher YA hace RMW).
- **(c) E-bis**: «Refrescar» = fetch/pull (del clone CC o de un checkout git
  propio) — decisión aparte, chica.
- Recomendación: (a) + (c) primero (desbloquean el ciclo entero con lo ya
  construido), (b) después; E-ter se decide cuando haya 2ª versión de
  developer-vitalia.

**Verificación:** ciclo completo SIN mis manos — forjar (con D resuelta) →
adoptar canónico → Publicar (B2 real, guards + idempotencia) → Refrescar →
aparece versión nueva → Traer/Actualizar → Mapa. El `check_marketplace_shape`
del repo del marketplace debe seguir verde.

---

## Disciplina para la sesión que retome (no negociable, §10)

- Decisiones nuevas → `decisiones.md` de ESTE paquete, mismo turno (prefijo
  sugerido: DD-1, DD-2… «deudas dogfood»).
- Todo cambio de código → capability nueva/modificada (R2 muerde en pre-commit;
  seguir numeración desde CAP-153) + `cap_doctor.py --index`.
- Changelog en el mismo turno (`scripts/changelog.py add Corregido|Agregado …`);
  publicar con `make bump-*`, jamás a mano. Tras build del daemon: `make dev-sync`.
- D y E se verifican contra el BINARIO INSTALADO (skill
  `verificando-binario-instalado`), no con lectura de código.
- F y peek: stories con gate a11y entero; las 13+5+8 vigentes intactas.
- Al cerrar: fila de estado en `INDEX.md` + evidencia en `PARIDAD.md` (sección
  nueva «Deudas D/E/F») + BACKLOG tachado.

## Preguntas al operador (hacerlas ANTES de codear D y E — AskUserQuestion)

1. **D**: ¿canal de permisos siempre cableado (b, recomendada) o rol de forja (a)?
2. **E**: ¿«Adoptar como canónico» (a, recomendada) o B2-alta (b)? ¿Refrescar
   con fetch (c) entra en este corte?
3. **peek**: ¿foco completo en vista previa? (recomendado: sí, es read-only).
