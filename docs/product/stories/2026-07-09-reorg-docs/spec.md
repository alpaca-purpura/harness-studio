# Spec — reorg de docs (árbol de 3 ejes + hojas atómicas)

> vig: activo · revisar: 2026-08-01 · estado: PROPUESTA (espera firma del árbol)

## 1. Problema (auditoría, datos duros)

| Archivo | tok | Se carga | Anti-patrón |
|---|---|---|---|
| `CLAUDE.md` | ~4.5k | cada sesión | narrativa histórica, no puntero; ~70% duplica LEDGER; cifras a mano → drift |
| `LEDGER.md` | ~26k | al buscar pendientes | fichas-prosa **+** tabla Log = dato 2 veces; `Siguiente`/`Deuda` enterrados en historia |
| `UX.md` | ~15k | — | backlog + inventario + log iteración mezclados |
| 7× `INDEX.md` | — | — | «Estado» (hecho) + «Retomar aquí» (por venir) misma hoja |
| `memory/` | — | auto | **ya es el patrón objetivo** (índice + hojas atómicas + TTL) |

Los 3 males que nombró el operador, confirmados:
- **E1** «lo-que-viene en la misma hoja de lo-hecho» → backlog requiere cargar 26k de LEDGER.
- **E2** CLAUDE.md = repositorio, no puntero → 4.5k/sesión desperdiciados.
- **E3** sin hoja atómica de backlog → pendientes dispersos en 4 sitios, sin un grep.

## 2. Árbol propuesto (RF-170)

Regla madre: **3 ejes TEMPORALES (pasado/presente/futuro) + 1 eje de ESTADO (ortogonal).**

```
CLAUDE.md          ROUTER puro (~1.2k). identidad 1-línea + tabla «necesito X → leé Y»
                   + reglas duras (git, disciplina §10). CERO historia/cifras/narrativa.
BACKLOG.md         [FUTURO] LO-QUE-VIENE. hoja atómica grep-able. solo items ABIERTOS.
                   item cerrado se BORRA (su cierre ya vive en historia).
ESTADO.md          [PRESENTE] AHORA. fase activa · paquete activo→RETOMAR · cifras GENERADAS.
CAPABILITIES.md    [ESTADO] SALDO. qué sabe hacer el sistema HOY, validado (§8). eje ORTOGONAL
                   al tiempo: no es «cuándo», es «qué es». estado GENERADO por checks.
LEDGER.md          [PASADO] ÍNDICE de historia = solo la tabla 1-línea/ficha (hoy 997-1020).
ledger/HS-NN.md    hoja atómica por ficha (la prosa gorda de hoy). on-demand. TTL=archivada.
historias/<pkg>/
  RETOMAR.md       NOW del paquete, 1 pantalla (el «Retomar aquí» de hoy, aislado del Estado).
  spec/design/…    detalle on-demand (sin cambio).
```

Los 4 ejes, sin solape:
`LEDGER`=transacciones («hicimos X») · `BACKLOG`=deltas futuros · `ESTADO`=ahora ·
`CAPABILITIES`=estado de cuenta («el sistema HOY sabe hacer X, validado»).

- **RF-170** El árbol vive con esos 4 nodos raíz (`CLAUDE.md`/`BACKLOG.md`/`ESTADO.md`/
  `LEDGER.md`) + carpeta `ledger/`. Cada uno cubre UN eje temporal, sin solape.
- **RF-171** `CLAUDE.md` es puntero puro: ninguna línea narra qué se hizo ni cita una cifra.
  Toda cifra o historia se referencia por puntero, no se copia. Máx ~80 líneas.
- **RF-172** `BACKLOG.md` contiene SOLO items abiertos; formato por línea
  `[origen] item · tag`; cerrar un item = borrar la línea (no marcar done).
- **RF-173** `ESTADO.md` = fase del gran plan + paquete activo (puntero a su RETOMAR) +
  bloque de cifras generado (§5). Nada de backlog ni historia.
- **RF-174** `LEDGER.md` queda como la tabla-índice; el cuerpo prosa de cada ficha migra a
  `ledger/HS-NN.md`. La tabla es el grep-target: `grep HS-11 LEDGER.md` → puntero → hoja.
- **RF-175** Cada `historias/<pkg>/INDEX.md` parte su «Retomar aquí» a un `RETOMAR.md` de 1
  pantalla; el checklist «Estado» (lo-hecho) queda en INDEX. (Opcional; barato.)

## 3. Migración de `Siguiente`/`Deuda` (RF-176)

- **RF-176** Los campos `Siguiente:`/`Deuda:` se ELIMINAN de las fichas (`ledger/HS-NN.md`) y
  se consolidan en `BACKLOG.md`. La historia queda inmutable; el futuro, en una sola hoja.
  Detalle del round-trip en `migracion.md`. Draft real de destino en `BACKLOG.md` (este pkg).

## 4. TTL / ciclo de vida de la hoja (RF-177)

- **RF-177** Toda hoja lleva cabecera de cita: `> vig: activo | archivado · revisar: AAAA-MM-DD`.
  - `activo` = candidata a cargarse; `archivado` = solo por grep explícito.
  - `revisar` vencido → la hoja se re-valida o se archiva (mecanismo en `docs/architecture/CADENCE.md`,
    igual que docs/architecture/knowledge/arch). Las fichas `ledger/HS-NN.md` nacen `archivado` (son historia).

## 5. Cifras generadas (RF-178) — D2 firmada

- **RF-178** El bloque de cifras de `ESTADO.md` se puebla con la cola de
  `arnesia conformance --todo` (línea ya emitida, `cmd/arnesia/conformance.go:142`). Nadie
  teclea `240 checks · 28 pass · …` a mano. Opciones de cableado (elige el operador al
  ejecutar): (a) target `make estado` / script que reescribe el bloque entre marcadores
  `<!--stats-->…<!--/stats-->`; (b) hook lefthook pre-push que falla si el bloque quedó
  stale. Sin flag nuevo; a lo sumo un modo `--stats` quiet opcional más adelante.

## 6. Ganancia de tokens (objetivo medible)

Pregunta común «¿qué está abierto / en qué fase estamos?»:
- Hoy: `CLAUDE.md` 4.5k + escaneo `LEDGER` 26k ≈ **30k tok**.
- Después: router 1.2k + `BACKLOG` ~1k + `ESTADO` ~0.8k ≈ **3k tok** (~**90% menos**).
- Historia solo al grep-ear una ficha puntual (`ledger/HS-NN.md`, ~1-1.5k c/u).

Métrica de aceptación: medir `wc -c` de las 3 hojas de orientación post-migración y
confirmar ≤ 4k tok combinados; comparar contra el baseline 30k documentado aquí.

## 7. Eje CAPABILITIES — Business Capability Map + Living Documentation (RF-179..182)

Nombre metodológico: **TOGAF Business Capability Map** (qué sabe hacer, por rol×proceso) ·
mecanismo: **BDD Living Documentation** (as-code, linkeado a checks — no prosa que se pudre).
Modelo mental SAFe *Solution Intent* como paraguas si se quiere framing corporativo.
Draft real (20 capabilities derivados de HS-01..HS-17) en `CAPABILITIES.md`.

- **RF-179** Existe `CAPABILITIES.md` = registro del estado validado del sistema actual. NO es
  historia (LEDGER) ni backlog. Organizado por los 4 verbos de VISION (crear·mapear·observar·
  mejorar) + no-funcional.
- **RF-180** Unidad = capability de negocio nombrado (`CAP-NN`), agrupando 1..N RFs (D4). Cada
  capability declara: `rol×proceso` · `origen` (RFs + ficha) · `valida` (check/test) · `estado`.
- **RF-181** El `estado` de cada capability (`vivo`/`degradado`/`planificado`/`sin-check`) se
  DERIVA de su(s) check(s), no se teclea (D2/BDD): `vivo` ⟺ check verde. `arnesia conformance`
  es la fuente; capability sin check linkeado = `sin-check` (visible, no se oculta — honestidad
  del repo, mismo criterio que el warn-fail del dogfood).
- **RF-182** Cada RF de cada paquete declara qué capability **agrega / edita / borra** → la
  historia de usuario queda como *delta* trazable al saldo. Cierra el loop story→capability.

## 8. Fuera de alcance (este paquete)

- Reorg de `UX.md`/`METODOLOGIA.md` (mismo mal, pero se ataca en paquete aparte para no
  mezclar ejes). Se deja anotado en `BACKLOG.md`.
- Cambiar el formato de `memory/` (ya es correcto; es el modelo, no el problema).
