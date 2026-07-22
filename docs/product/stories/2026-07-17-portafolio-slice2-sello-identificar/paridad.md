# PARIDAD — Portafolio · Slice 2 «El sello + Identificar»

> Paquete `2026-07-17-portafolio-slice2-sello-identificar`. Base: `spike-spec.md` §7 (decisiones
> operador) + §8 (progreso T1-T4) + §9 (verificación). Esta hoja no existía al construir — se escribe
> junto con la firma, el paquete no tenía `paridad.md` propio, solo la nota «Retomar aquí» de `INDEX.md`.

## Corrección de estado — "sin commitear" quedó STALE

`INDEX.md`/`spike-spec.md` decían "TODO CONSTRUIDO Y VERIFICADO (2026-07-17, sin commitear)". Verificado
contra `git log` en esta sesión (2026-07-22): el código (T1-T4: `HuellaPath`/`Disc`, `Graph.Degradado`,
`Identificar`, capabilities `portafolio/identificar` + extensiones) SÍ está en `main` — commit
`824acd8 feat(portafolio, session-rail): drawer + picker selector de arnés · decisiones + specs`. La nota
de "sin commitear" nunca se actualizó después de ese commit. `git status` confirma tree limpio.

## Estado de la verificación (spike-spec.md §9)

- **Go:** `go build ./...` limpio · `go test ./internal/...` verde (5 tests nuevos: `TestClaveDesempataPorPath`,
  `TestLoaderSinManifiesto` ext., `TestObservarEnMapaDegradadoSintetiza`, `TestIdentificarSellaYRekey`,
  `TestIdentificarNoPisaSelloExistente`) · `go vet` limpio.
- **Conformance:** `257 checks · pass 48 · fail 0 · error 0` · R1/R2 PASS.
- **FE:** `pnpm run verify` limpio · `pnpm test` **156/156** (incl. story `IdentificarSinSello`).
- **Capabilities:** `portafolio/identificar.yaml` (nueva, `vivo`) + `observar-en-mapa`/`registrar-identidad`
  extendidos. Cifras regeneradas: 93 caps · 50 vivo · cobertura 100 %.
- **E2E contra el daemon real** (`scratchpad/e2e_sello.sh`, HOME temporal, sin fakes): crudo → escanear
  (`sin-home~~<huella>`) → agregar → **observar degradado HTTP 200** (antes 400, GAP-1 cerrado en este
  slice también para el caso sin-sello) → identificar (sella + re-key a `miproj`) → sello en disco →
  re-escanear (`sin-home~miproj~`) → observar sellado (arnés completo). Circuito completo VERDE.

## Decisiones del operador (spike-spec.md §7, ya firmadas de palabra 2026-07-17)

1. Entrada degradada PERSISTE en el Portafolio.
2. «Identificar» V1 sella IN-SITU sin clon (evita duplicidad; git ya da seguridad ante cambios).
3. Nombre del paquete → Slice 2.

## Firma

- [x] 🧑‍⚖️ **Gate humano — FIRMADO 2026-07-22** (operador): *"ya lo vi, firma vos nomás"* — confirma
  haber revisado en vivo el circuito crudo→degradado→identificar→sellado descrito en spike-spec.md §9
  antes de esta sesión. Respaldo: verificación Go+conformance+FE arriba + el E2E real. Sin desviaciones
  registradas más allá de la nota "sin commitear" ya corregida (era documental, no de código). Cierre:
  `checkpoint.md`/`BACKLOG.md` actualizados, `ledger/HS-25.md`.
