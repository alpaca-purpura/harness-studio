# Instalador público (`curl | sh`) + licencias por organización

> `tipo: idea` (research hecho, sin firmar). Nace de la pregunta «cuál es el comando para
> instalar la última versión de arnesia» (2026-07-14) — no había ningún release publicado.

## Pregunta

¿Cómo distribuimos `arnesia` con un instalador público, y cómo controlamos qué organizaciones
internas de alpacapurpura (y sus usuarios) pueden activarlo — sin convertir ArnesIA misma en un
producto vendido a terceros (eso sigue fuera de `vision.md`, ver IL-D0)?

## Hallazgo — estado de la distribución hoy

- ❌ Cero releases publicados: `.goreleaser.yaml` configurado, nunca taggeado.
- ❌ Única vía de instalación hoy: build local (`scripts/bundle.sh`), exige toolchain de dev.
- ✅ Investigado y verificado en vivo: Keygen.sh self-hosted CE cubre el caso (SDK Go sin cgo,
  cross-platform, $0 de licencia para uso comercial) — ver comparación completa en
  `research.md`.

## Retomar aquí

**Sigue en `idea`.** Antes de promover a `refining`: resolver 2 decisiones abiertas en
`decisiones.md` — **IL-D7** (confirmar en código real, no solo docs, que `Groups` de Keygen
vive en CE) e **IL-D6** (módulo destino en el seam: extender `self-update` o crear
`distribucion`). Alcance ya acotado con el operador: interno, no venta externa (IL-D0).

## Archivos

- `research.md` — research de mercado: problema, 8 opciones de licensing comparadas y
  verificadas en vivo, viabilidad técnica, costo, opciones de solución A/B/C.
- `decisiones.md` — decisiones IL-D0…IL-D8 (alcance, punto de gate, proveedor, hallazgos
  verificados, y las 2 aberturas que bloquean `refining`).
- `00-story.md` — JTBD, por qué importa, outcome esperado, antecedentes, out-of-scope, riesgos.
- `checkpoint.md` — estado de tracking (`state: idea`), leído por el story-closure-gate scan
  de `/pm`.
