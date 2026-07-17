---
story_id: 2026-07-15-instalador-publico-licencias-org
state: idea               # idea→refining→refined→ready→developing→developed→reviewing→[done]
module: self-update       # TENTATIVO — ver IL-D6 en decisiones.md (puede exigir módulo nuevo `distribucion`)
cap_target: ninguno (no existe capability aún — nace al construir)
chris_verify:
  signoff: false
defer_audit: false
last_audit: 2026-07-15
ledger: HS-24              # IL-D9 firmada (versionado instaladores); resto del paquete sigue sin firmar
---

# checkpoint — Instalador público + licencias por organización (presente)

## Estado

`idea` — research de mercado hecho y verificado en vivo (comparación de 8 opciones de
licensing-as-a-service + DIY), alcance acotado por el operador (interno, NO venta externa — no
toca `vision.md`), pero **sin spec funcional, sin mockup, sin decisión firmada**. `/po` (owner
formal de `refining`) no está forjado en este repo todavía — el operador conduce esta etapa a
mano vía `/pm`.

## dod_evidence

- Prior-art scan: sin capability, story ni research previos sobre instalación pública o
  licencias (`capabilities/`, `stories/`, `product/research/` — 0 resultados).
- Research de mercado con 2 rondas de verificación en vivo (WebFetch contra docs/pricing
  reales, no memoria de training): comparación 8 opciones → `research.md`.
- Confirmado en vivo: Keygen CE self-hosted es gratis para uso comercial (README del repo,
  no solo interpretación de pricing page) + SDK Go sin cgo cross-compila a los mismos targets
  que ya usa `.goreleaser.yaml` (`darwin`/`windows`/`linux` × `amd64`/`arm64`).
- Gate de alcance corrido con el operador (`AskUserQuestion`): **interno acotado**, no pivot de
  negocio — no requiere enmendar `vision.md` línea 156.

## Retomar aquí

**Sigue en `idea`.** Próximo paso legal: decidir si se promueve a `refining` (mapa funcional +
spec) o se parquea. Ver `decisiones.md` por las 3 decisiones abiertas (IL-D6, IL-D7, IL-D8) que
bloquean ese salto — la más dura es confirmar en código real (no solo docs) que `Groups` de
Keygen vive en CE, no en EE. Archivos del paquete: `research.md` (comparación de mercado) ·
`decisiones.md` (IL-D0…IL-D8) · `00-story.md` (JTBD + outcome + out-of-scope).
