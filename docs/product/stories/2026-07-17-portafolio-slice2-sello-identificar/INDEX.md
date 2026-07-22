# Portafolio · Slice 2 — «El sello + Identificar» (jalar/sellar/degradado)

> `tipo: paquete-de-trabajo` · abierto 2026-07-17 · deriva de las decisiones **S1-D27..D29**
> (`../2026-07-13-portafolio-slice1-fe/decisiones.md`) y de la visión firmada
> [[hs-vision-sello-jalar-arneses]]. Etapa: **spike de spec** (resolver forks técnicos → firma
> del enfoque → recién ahí código). El gate 🧑‍⚖️ del enfoque está ABIERTO.

## Qué resuelve

El operador quedó atrapado: el Portafolio pinta como arnés algo que el Mapa niega (vitalia, sin
manifiesto → `400`). Este slice cierra la tensión con tres piezas coherentes:

1. **Modo degradado en el Mapa** (S1-D27, Opción A) — honra el contrato `nomenclatura-arnes.md`
   §2: una presencia sin sello se OBSERVA igual, grafo esqueleto + check rojo `manifiesto-ausente`,
   nunca un `400` crudo. Es el *preview antes de sellar*.
2. **Acción «Identificar»** (S1-D28) — genera el sello `arnes.l0.json`; el verbo que faltaba entre
   *Reparar* y *Traer canónico*. Es el *acto de sellar*. Primera pieza del inicializador universal.
3. **Fix llave degenerada `sin-home~~`** (S1-D29) — huella de path para que dos proyectos crudos no
   se pisen; la MISMA huella sirve de llave sintética para el degradado (cierra el fallback de GAP-2).

## Documentos

- [`spike-spec.md`](./spike-spec.md) — el spike: contexto verificado contra código, los 3 forks
  técnicos resueltos con diseño recomendado, superficie de capabilities, plan de tickets, y las
  decisiones que faltan del operador.

## Retomar aquí

> **Estado (2026-07-22): SLICE 2 CONSTRUIDO Y VERIFICADO PUNTA A PUNTA · gate 🧑‍⚖️ FIRMADO** — ver
> [`paridad.md`](./paridad.md). T1-T4 hechos (ver `spike-spec.md` §8): huella de path (D29) · degradado
> en Mapa (D27, el 400 murió) · Identificar in-situ + re-key (D28) · capabilities + cifras (93 caps,
> cobertura 100%). Verificación (§9): Go verde · conformance `257·pass 48·fail 0` R1/R2 PASS · FE
> `pnpm test` **156/156** · **E2E contra el daemon real** (crudo→degradado 200→identificar→sellado→arnés
> completo) VERDE. Corrección: la nota "sin commitear" quedó stale — el código está en `main` desde
> `824acd8`, confirmado por `git log` al cerrar (ver `paridad.md`). El operador reinstala el binario
> nuevo si quiere verlo en la app instalada (el daemon vivo podía correr uno viejo).
