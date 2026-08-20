# Learnings — carril L2 del CIL

Cada aprendizaje vive en **su propio archivo** `.md`. Este README no cuenta como uno: el
cockpit lo ignora a propósito (`handlers_learnings.go`), así que sirve de índice sin
contaminar la lista.

## Dónde va cada cosa

| Tipo | Path |
|---|---|
| Producto / arquitectura / doctrina | `docs/learnings/AAAA-MM-DD-<slug>.md` |
| Herramientas y workspace | `docs/learnings/tooling/<slug>.md` |

Lo que es fricción de PROCESO va al carril L1
([`docs/process/harness-backlog.md`](../process/harness-backlog.md)); lo que es deuda de
CÓDIGO va al L3 ([`tech-debt.md`](../process/tech-debt.md)). Un aprendizaje explica **por
qué pasó algo**; un ítem de backlog pide **que se haga algo**. Si no explica nada, no es un
learning: es un ticket.

## Frontmatter

El cockpit lee estas claves (`listLearningsV2`):

```yaml
title: "…"            # el título que se pinta
date: 2026-08-19
type: doctrina | arquitectura | tooling | proceso
promotable: si | no   # `no` = referencia, no espera decisión
applied: pending | applied | promoted | wont-apply
tags: [.., ..]
```

**`applied` no es decorativo.** Capturado NO es estado final: un learning en `pending`
está esperando que alguien decida si se promueve a regla, a check o a nada. La métrica del
carril es que ese número BAJE — no que suba el total.
