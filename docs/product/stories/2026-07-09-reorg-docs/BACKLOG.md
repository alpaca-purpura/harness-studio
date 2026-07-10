# BACKLOG — ArnesIA (DRAFT del paquete reorg-docs)

> vig: activo · revisar: 2026-08-01
> Hoja atómica de LO-QUE-VIENE. Solo items ABIERTOS. Item cerrado se BORRA (su cierre vive
> en `ledger/`). Formato: `[origen] item · tag`. Tags: `gate` (firma humana) · `deuda` ·
> `bloqueo` (depende de otra cosa) · `verificar` (posible stale, confirmar antes de migrar).
> DRAFT: consolidado de LEDGER `Siguiente`/`Deuda` + los 7 `INDEX.md`. Al ejecutar la
> migración se reconcilian los `verificar`.

## Gates humanos pendientes (código listo, falta firma 🧑‍⚖️ PARIDAD)

- [chat-cc-funcional] firmar 7 desviaciones `PARIDAD.md` + mockup v1 · `gate`
- [franja-artefactos] gate final lado a lado, 7 desviaciones (HS-13 F5) · `gate`
- [boton-actualizar] gate final, 6 desviaciones `PARIDAD.md` · `gate`
- [inspector-drawer] gate final, 5 desviaciones registradas · `gate`

## Deuda viva (registrada, no bloquea la línea principal)

- [HS-09/11] telemetría JSONL → indexer real ⇒ desbloquea capas Tokens/Desempeño/Proceso del Mapa · `bloqueo`
- [HS-11/chat] spike `control_response` vs claude real (confirmar en papel/e2e) · `deuda`
- [HS-11] run async del `/boxes/{id}/run` + gate post-run · `deuda`
- [HS-11] 3 boundaries de research → materializar en `arch/` · `deuda`
- [chat] fase presentación: assistant-ui + CodeMirror merge + widgets ricos (decisión #5) · `deuda`
- [HS-09] 212 checks `deferred` → correr en CI (hoy solo la ruta `--arnes`) · `deuda`
- [HS-12] loader detector 3°: leer lock `.devstudio/arneses.yaml` y resolver multi-arnés · `deuda`
- [HS-16] loader reconocedores `deferred`: subagent · plugin-nodo-raíz · edges-de-librería (necesitan diseño) · `deuda`
- [HS-16 Grupo A] 6 checks composición `deferred` (rediseño de motor; bloqueado por SQLite fase5 / OTel / modo-por-fase) · `bloqueo`
- [HS-16 Grupo C] `gate-honesto`: necesita diseño previo · `deuda`
- [HS-14] deep-link `arnesia://` en callback single-instance (ojo bug tauri#12726) · `deuda`

## A reconciliar (posible stale — confirmar antes de dar por abierto)

- [plan-hito2] «Fase E backend» + «go-arch-lint binario»: HS-11 cerró Fase E y HS-10 puso
  go-arch-lint VIVO en CI — el INDEX del plan puede estar stale · `verificar`
- [interop-devstudio] entregar `respuesta-devstudio.md` al agente DevStudio: ¿hecho? · `verificar`

## Reorg de docs (este paquete)

- [reorg-docs] 🧑‍⚖️ firmar el árbol de 3 ejes → ejecutar migración (spec §2) · `gate`
- [reorg-docs] cablear cifras generadas de `conformance` a `ESTADO.md` (RF-178) · `deuda`

## Fuera de alcance ahora (anotado para no perderlo)

- [reorg-docs] `UX.md` (15k) y `METODOLOGIA.md` (8.5k) tienen el mismo mal (backlog+historia
  mezclados) — atacar en paquete aparte, no mezclar ejes · `deuda`
