# Portafolio · Slice 0 «Cimientos» — paquete de build

> `tipo: build` (implementación del modelo FIRMADO HS-22). Deriva del spike
> [`../2026-07-10-spike-carga-arneses/`](../2026-07-10-spike-carga-arneses/INDEX.md) (specs) +
> épica «Programa Portafolio» item 0 en [`../../BACKLOG.md`](../../BACKLOG.md).
> Roles: **Fable 5 = arquitecto** (plan + decisiones técnicas, 2026-07-13) · **Sonnet 5 = constructor**
> (ejecuta tickets T1-T9) · **operador = firma** (gate PARIDAD).

## Qué se construye (1 línea)

Los cimientos de dominio/backend del Portafolio (SIN FE): identidad `(home,id)` · walker 3-formas ·
fallback `plugin.json` · cadena de origen collect-all (con eslabón CC resuelto) · `deriva` por hash ·
store separado que degrada honesto · HTTP+CLI mínimos · capabilities + boundary as-code.

## Archivos

- [`plan-implementacion.md`](./plan-implementacion.md) — **EL plan**: goal verificable (§0), diseño técnico
  con contratos exactos (§2), tickets T1-T9 con tests nombrados (§3), gate local (§M), prohibiciones (§P).
- [`decisiones.md`](./decisiones.md) — S0-D1..D11: decisiones del arquitecto que cierran lo que S-D11
  difirió al build (eslabón CC investigado con evidencia real · 3 tipos de instalación · empresas[] ·
  layout hexagonal · store · deriva local-only · E3 en papel · boundary nuevo).
- [`paridad.md`](./paridad.md) — tabla spec↔realidad + evidencia E2E vivo contra la máquina real
  (2 bugs reales encontrados y corregidos: S0-D14/S0-D15). **Firma 🧑‍⚖️ PENDIENTE.**

## Retomar aquí

> **Estado (2026-07-13): T1→T9 CONSTRUIDOS Y COMMITEADOS a `main`.** Los 9 tickets del plan
> ejecutados en orden (`git log --oneline --grep="feat(portafolio): T"`), gate local (§M) verde
> antes de cada commit. `paridad.md` escrita con el goal §0 del plan verificado punto por punto +
> evidencia E2E real. **Siguiente paso: gate humano 🧑‍⚖️ del operador** — revisar `paridad.md`
> (8 puntos del goal + 5 desviaciones registradas) y firmar. Tras la firma: cerrar el paquete en
> `checkpoint.md`/`BACKLOG.md` (item 0 → construido y firmado) y decidir el recorte de Slice 1 (FE).
