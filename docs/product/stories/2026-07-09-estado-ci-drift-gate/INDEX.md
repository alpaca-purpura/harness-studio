# estado.sh → CI drift-gate — INDEX

**Slug:** `2026-07-09-estado-ci-drift-gate` · **Estado global:** 🟡 `refining` · **Ledger:** HS-21

## Marco

Eje **N2 «honestidad automática»** (elegido 2026-07-09), **Paquete A** (el quick-win de mayor
palanca). Cablear `scripts/estado.sh` a CI: si las cifras vivas de `docs/product/checkpoint.md`
divergen del estado REAL del repo, **el merge se rompe**. Mata el drift «cifras tecleadas» para
siempre. Patrón a copiar: el job `tokens-sync` de `.github/workflows/ci.yml` ya hace
`build && git diff --exit-code`.

## Por qué (prior-art)

- `estado.sh` **existe** y genera todo el bloque `<!--stats-->` (RF-178 + HS-20). Falta SOLO el
  enforcement: nadie lo corre en CI ni en lefthook.
- Doctrina honestidad (METODOLOGIA §4): «cifras se GENERAN, no se teclean». Hoy la generación es
  MANUAL — la disciplina depende de que un humano se acuerde de correr el script. Esto lo automatiza.

## Flujo de gates 🧑‍⚖️

1. Investigación de `estado.sh` (determinismo · deps · gotcha date-stamp) — ⏳ en curso
2. Decisiones D1..Dn firmadas — ☐
3. Spec del gate — ☐
4. Implementar (CI + posible `--check` mode) + capability — ☐
5. Validación propia (drift real atrapado · sin falsos positivos) — ☐
6. Auditoría independiente (subagente) — ☐
7. PARIDAD / gate humano 🧑‍⚖️ — ☐

## Estado

- [x] Paquete nacido (checkpoint + INDEX + decisiones)
- [ ] Investigación cerrada
- [ ] Decisiones firmadas
- [ ] Implementado + capability ratificado
- [ ] Validación propia verde
- [ ] Auditoría
- [ ] Gate humano
