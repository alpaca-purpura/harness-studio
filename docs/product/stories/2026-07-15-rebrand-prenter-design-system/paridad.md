# PARIDAD — Rebrand sistema de diseño PRENTER

> Paquete `2026-07-15-rebrand-prenter-design-system`. Base: `propuesta-tokens.md` + `decisiones.md`
> D1-D9. Esta hoja no existía al momento de implementar (`INDEX.md` lo señalaba como deuda del cierre)
> — se escribe ahora, junto con la firma, consolidando la evidencia que ya estaba documentada en vivo
> en `INDEX.md` §Retomar aquí.

## Estado de la verificación automática (ya corrida al implementar, 2026-07-15)

| Suite | Resultado |
|---|---|
| `pnpm run verify` | ✅ verde |
| `vitest --project=storybook` | ✅ 127/127 (incl. 24 fallos de contraste reales cazados y corregidos, D9) |
| `vitest --project=unit` | ✅ 21/21 |
| Storybook `:6006` en vivo | ✅ screenshot real (Portafolio Lista + Wizard, light y dark), Jost confirmado vía `getComputedStyle` |
| `color.kind`/`health`/`heat` (paleta funcional del Mapa) | ✅ intactos, verificado en `theme.css` generado |

## Desviaciones (a firmar en el gate)

1. **D8/D3 — botones/CTA por-módulo (`.pf-btn-*` etc) siguen en `font.sans`**, sin QA visual por
   superficie. Deuda explícita, no bloquea la firma (decisión ya tomada en `decisiones.md`).
2. **`mockups/arnesia-shell-A-galaxia.html` (🔒 firmado) NO se re-derivó** al momento de esta firma —
   sigue reflejando los tokens VIEJOS (pre-PRENTER). El acto de cierre real (re-derivar el mockup +
   re-estampar `mockups/INDEX.md`) queda como ticket propio en `BACKLOG.md` (no se simula "ya
   re-derivado" — es trabajo de diseño real, no de paperwork).

## Firma

- [x] 🧑‍⚖️ **Gate humano — FIRMADO 2026-07-22** (operador): *"ya lo vi, firma vos nomás"* — confirma
  haber corrido la app/Storybook en vivo y revisado el resultado (dark-first, paleta del Mapa intacta,
  fuentes self-hosted) antes de esta sesión. Respaldo: tabla de verificación automática arriba + las 3
  decisiones D1-D3 + D9 (correcciones WCAG). Desviación #2 (mockup sin re-derivar) queda como deuda
  visible en `BACKLOG.md`, no bloquea esta firma. Cierre: `checkpoint.md`/`BACKLOG.md` actualizados,
  `ledger/HS-25.md`.
