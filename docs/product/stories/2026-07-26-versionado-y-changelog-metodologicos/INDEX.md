# Versionado y changelog metodológicos (paquete de trabajo)

> `tipo: proceso as-code` · **Vig: activo · construido y probado E2E, falta firma 🧑‍⚖️.**
> Orden del operador (2026-07-26): *«que ya sea metodológico esto, no debo andar diciéndote nada»*
> — actualizar la versión **y** registrar qué se agrega / corrige / elimina no puede depender de
> que alguien se acuerde.

## El agujero (medido, no supuesto)

| Qué | Antes |
|---|---|
| Sello de build `AAMMDDHHMM` | ✅ automático en todo `bundle.sh` (RF-231, el día anterior) |
| Bump de semver | ⚠️ solo en `make installer`, y solo PATCH |
| Minor / major | ❌ a mano, sin target ni criterio escrito |
| **Qué trae cada versión** | ❌ **no existía**: 20 releases en `instaladores/` y `ls CHANGELOG*` vacío |
| `versionado.md` v1.1 (`enforced`) | ⚠️ ni una mención del sello de build (`grep` vacío) — describía un mundo de solo-semver |

## Lo que quedó

`CHANGELOG.md` (Keep a Changelog, 6 categorías cerradas) + **el bump como acto atómico**: valida,
sincroniza los 3 manifiestos y promueve `[Sin publicar]` a una sección fechada. **Con el changelog
vacío el bump falla y no toca ningún archivo.** Tres capas de enforcement (gate del bump · fitness
en CI · hook local) y la regla escrita donde una sesión nueva la lee sola.

```bash
python3 scripts/changelog.py add Agregado "lo que hiciste"   # en el turno en que lo hacés
make bump-patch | bump-minor | bump-major                    # publica: valida + promueve
```

## Estado

- [x] Decisiones VC-D1..VC-D7 → [`decisiones.md`](./decisiones.md)
- [x] [`spec.md`](./spec.md) — **RF-232**: mapa funcional (6 bifurcaciones · RN-1..8 · AC-1..9) +
      `SC-1..SC-7` + matriz de cobertura + límites
- [x] `CHANGELOG.md` sembrado sin inventar historia (`convencion-desde: 0.2.22`)
- [x] `scripts/changelog.py` (check · add · release · sin-publicar) + `scripts/bump.sh` (punto único)
- [x] `make bump-patch|bump-minor|bump-major` + `make changelog`
- [x] `docs/architecture/fitness/changelog_test.go` — 3 tests
- [x] `lefthook.yml` job `changelog` (glob: los 3 manifiestos)
- [x] `versionado.md` **v1.2**: §changelog-y-bump + §identidad de build ≠ versión de release + 4 checks
- [x] `CLAUDE.md` (regla dura + fila del router) · `metodologia.md` §10 regla 1
- [x] [`PARIDAD.md`](./PARIDAD.md) — ciclo E2E corrido de verdad y revertido
- [ ] 🧑‍⚖️ **Gate de PARIDAD (AC-9)** — el operador corre un ciclo real y firma

## Retomar aquí

- **Último hecho (2026-07-26):** construido completo y probado E2E sobre el repo real: bump
  0.2.21→0.2.22 con promoción del changelog, segundo bump **abortado** por changelog vacío sin
  tocar archivos, `bump-minor` 0.2.22→0.3.0, alias/idempotencia de `add`, rechazo de categoría
  inventada. Todo **revertido** desde backup: los manifiestos volvieron a `0.2.21`.
- **Próximo paso concreto (AC-9):** correr un ciclo real cuando toque publicar —
  `python3 scripts/changelog.py sin-publicar` para ver qué hay acumulado → `make bump-minor`
  (esta tanda trae superficie nueva: identidad de build, dictado, marketplaces) → `make installer`
  → firmar.
- **Firmas pendientes:** una — 🧑‍⚖️ PARIDAD (AC-9). VC-D1..VC-D7 quedan decididas a la espera de
  esa misma firma.
- **Nada commiteado.** Al commitear, el job `changelog` de lefthook se dispara solo si tocás los 3
  manifiestos; el resto del árbol no lo activa.
