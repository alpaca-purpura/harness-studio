# CADENCE — cómo vive el árbol de arquitectura

> **La arquitectura as code no se revisa por calendario; se revisa al cambiar.** A diferencia de
> [`../docs/architecture/knowledge/CADENCE.md`](knowledge/CADENCE.md) (barrido semanal del ecosistema externo),
> `arch/` refleja **nuestras** decisiones estructurales: se toca cuando una decisión de
> arquitectura nace, muta o se supera. Norte: [`../product/vision.md`](../product/vision.md) y el
> [`../LEDGER.md`](../product/LEDGER.md) (el «por qué» firmado). Este doc define el mecanismo: cómo entra
> un boundary, cómo se versiona, cómo se mantiene la disciplina L1↔L2, y cómo se vuelve el
> ruleset que `arnesia conformance` corre contra el propio código de la app.

## Por qué existe (y por qué NO es un doc pesado)

Un documento de arquitectura clásico (50 páginas de prosa + diagramas Visio) se pudre en
semanas: el código deriva, nadie re-dibuja, y a los tres sprints la doc miente. La regla de la
casa (operador, HS-04): **la arquitectura vive como código versionado, parte ejecutable, parte
renderizada, con la mínima prosa a mano posible.** Tres separaciones estrictas:

- **El «por qué»** → una ficha `HS-NN` del [`../LEDGER.md`](../product/LEDGER.md). Nunca se duplica aquí.
- **La «prueba»** → una check en [`fitness/`](./fitness/) que **falla CI** si el código viola la
  regla. La arquitectura que no se puede romper en CI es un deseo, no una restricción.
- **El «dibujo»** → [`model/`](./model/), renderizado on-demand desde texto (D2/Mermaid), que
  **diffea en git** y nunca se sincroniza a mano.

Lo único a mano en un boundary node es la regla viva (L1) y su mapeo al código (L2) — mínimo.

## Anatomía de un boundary node (`boundaries/<regla>.md`)

```markdown
---
regla: <slug-de-la-regla>
version: <n.m>              # sube con cada pasada que cambia el nodo
updated: <YYYY-MM-DD>
status: proposed | enforced | advisory | superseded
ledger: HS-NN              # la ficha firmada = el «por qué» (sin duplicar)
sources:                   # L1: el principio como estándar de industria, fechado
  - url: <...>
    autoridad: oficial | estándar | experto
    revisado: <YYYY-MM-DD>
enforced_by:               # link 1:1 a la check ejecutable que guarda la regla
  - fitness/.go-arch-lint.yml#<componente>
  - fitness/arch_test.go:<TestNombre>
severity: critical | high | medium
---

## L1 · Principio (estándar de industria)
El patrón como lo recomienda la industria, dated + con fuente. Evidencia, no opinión.

## L2 · Realización (este árbol Go+React)
Qué paquetes SON core/shell/dominio/transporte/adaptador; cómo la regla se encarna aquí.
Cada afirmación de mapeo cita el punto de L1 del que deriva (`⇐ L1: <ancla>`).

## Checklist evaluable
| id | qué chequea | severidad | señal en el mapa | enforcer |
Cada check es objetivamente verificable desde el árbol de imports, el schema o un test.

## Changelog
YYYY-MM-DD · vX.Y · qué cambió · qué ficha lo disparó
```

## El ritual (al cambiar, no por calendario)

1. **Decisión estructural.** Nace en una conversación/ficha `HS-NN` (o la muta). Ejemplo: «el
   dock habla AG-UI», «el conductor es subproceso, no SDK-sidecar».
2. **¿Toca un boundary?** Si crea o cambia una regla estructural → nodo aquí (nuevo o bump). Si es
   una decisión sin regla enforçable (p.ej. «usamos Zustand») → basta la ficha + el doc de historias;
   no todo va a `arch/`, solo lo que se puede **romper en CI**.
3. **L1 + L2.** Escribir el principio con fuente y el mapeo al código. Si L2 diverge de L1 →
   marcar `⚠ divergencia` + justificar, nunca silenciosa (regla heredada de knowledge).
4. **Check + enforcer.** Traducir la regla a una check en `fitness/` y ponerla en `enforced_by:`.
   Si la superficie que la enforça aún no existe, `status: proposed` y la check queda declarada.
5. **Diagrama, si cambió la topología.** Actualizar `model/` (el texto; el render es on-demand).
6. **Bump + changelog + propagar.** Subir `version`/`updated`, registrar el anillo; si toca specs
   (fase 4) o la UX, apuntarlo.

## Reglas del árbol (no re-negociar)

- **El LEDGER es el «por qué» canónico.** `arch/` no forkea el diario; lo proyecta (`ledger:`).
- **Toda regla que llega a `boundaries/` DEBE tener un `enforced_by:`** — si no se puede romper en
  CI, no es un boundary, es una nota (va al doc de historias o la ficha).
- **Superar, no borrar** (aditivo, principio 4): un boundary superado va a `status: superseded`
  con puntero al que lo reemplaza; no se elimina.
- **L2 deriva de L1.** Divergencia marcada y justificada.
- **Honestidad (METODOLOGIA §4):** lo no verificado dice «unverified»; `status: proposed` mientras
  no haya código que enforce; jamás se pinta «enforced» sin check corriendo.
- **Schema-first:** el dominio (`meta.clase` L0, `contract:` de caja) vive en
  [`contracts/schema/`](./contracts/schema/) como single source; los tipos Go/TS se **generan**,
  no se escriben a mano (evita el drift doc↔código).

## Hacia dónde va (el árbol ES enforcement)

El código ya aterrizó (HS-06/HS-08: daemon Go `internal/`+`cmd/`, SPA `web/`, shell Tauri
`web/src-tauri/`), y el enforcement dejó de ser futuro: `fitness/.go-arch-lint.yml` **corre en CI
desde HS-10** (`go-arch-lint check --project-path . --arch-file docs/architecture/fitness/.go-arch-lint.yml`)
y rompe el build si el core importa el shell o el dominio importa `net/http`;
`contracts/schema/*.json` valida cada `contract:` de caja (`TestBoxContractValidatesAgainstSchema`,
enforced en HS-08); y el motor **`arnesia conformance`** (construido en HS-08) unifica los
checks de arquitectura con los 138 de `docs/architecture/knowledge/` — **247 checks a datos** (`--todo`, 2026-07-09) — en un solo reporte
severidad+señal. La misma disciplina que aplicamos a los arneses que fabricamos, aplicada a la
fábrica misma (dogfood). Por eso cada check mal puesto es un falso positivo en CI.
