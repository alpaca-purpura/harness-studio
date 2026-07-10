# Decisiones — reorg de docs

> vig: activo · revisar: 2026-08-01
> Regla anti-pérdida (METODOLOGIA §10): cada decisión conversada se escribe aquí EN EL MISMO
> TURNO. FIRMADA = el operador la aprobó; PROPUESTA = espera firma.

## FIRMADAS (2026-07-09, vía AskUserQuestion)

### D1 · Alcance = diseño primero — FIRMADA
Se redacta la propuesta completa como paquete de trabajo (este) ANTES de tocar docs
firmados (LEDGER/CLAUDE/VISION). La migración se ejecuta en un turno posterior, tras firma
del árbol. Motivo: LEDGER/CLAUDE son artefactos firmados; reestructurarlos sin firma viola
la disciplina. Reversible por git, pero se firma igual.

### D2 · Cifras vivas = generadas por conformance — FIRMADA
Las cuentas (checks · pass · deferred · boundaries) NO se teclean en prosa nunca más. Salen
de `arnesia conformance`. Motivo: el drift de cifras es recurrente (HS-16 y HS-17 tuvieron
que corregir «cifra stale» en CLAUDE.md). La app ya conoce el número real.
- **Hecho técnico habilitante:** `cmd/arnesia/conformance.go:142` ya imprime
  `%d checks · pass %d · fail %d · error %d · deferred %d · n/a %d`. No hace falta flag
  nuevo — basta capturar la cola de `arnesia conformance --todo` en `ESTADO.md`.

### D3 · 4º eje = CAPABILITIES (Business Capability Map + Living Documentation) — FIRMADA
El árbol suma un 4º eje ORTOGONAL a los 3 temporales: el **estado validado** del sistema.
`LEDGER`=transacciones · `BACKLOG`=deltas futuros · `ESTADO`=ahora · **`CAPABILITIES`=saldo**
(«estado de cuenta»: qué sabe hacer el sistema HOY, validado). Motivo: las historias de
usuario son efímeras (mutan capabilities); faltaba el registro consolidado del estado actual.
Nombre = TOGAF *Business Capability Map* · mecanismo = BDD *Living Documentation* (as-code,
linkeado a checks — no prosa que se pudre). Meta-dogfood: la app que mapea capabilities de
arneses mapea los suyos igual. Draft real en `CAPABILITIES.md`.

### D4 · Unidad = capability de negocio (rol×proceso), agrupa 1..N RFs — FIRMADA
La unidad atómica del registro es el **capability nombrado** (ej. «cargar arnés», «aislar
config al spawn»), no el RF suelto. El RF es el detalle interno que lo crea/edita. Coincide
con el producto (Mapa por rol×proceso) y con TOGAF. Cada capability linkea sus RFs + sus
check(s) validadores; su estado (`vivo`/`degradado`/`planificado`/`sin-check`) lo da el check
(D2). Cada RF de cada paquete declara qué capability agrega/edita/borra → cierra el loop.

### D5 · Registro DERIVADO del código (no de fichas) + doctrina de enforcement — FIRMADA
Mandato del operador (2026-07-09): CAPABILITIES = SSoT funcional, docs-as-code; barrer TODO
el código con subagentes y estructurar los capabilities con punteros archivo/función a la
granularidad correcta; hook que valide que NADA crea código sin construir/modificar un
capability y actualizarlo. Ejecutado: 7 subagentes barrieron cmd/internal/web/src-tauri →
`CAPABILITIES.md` v1 = **82 capabilities con punteros `file#Símbolo` reales**. Doctrina de
enforcement (R1 integridad · R2 cobertura · R3 gate de commit · R4 estado generado; rollout
warn→block) en `doctrina-capabilities.md`. Forma final = `capabilities/<area>.md` atómicas.

## HALLAZGOS del barrido (a resolver al migrar)

- **DRIFT CLAUDE.md ↔ código (honestidad):** CLAUDE.md dice «SQLite puro-Go (modernc, WAL)».
  Realidad del código: índice = **map in-memory** (`adapters/index/store.go`), persistencia =
  **archivos JSON atómicos** en `~/.arnesia/` (`adapters/store/*.go`). SQLite es fase 5 futura
  (comentarios `store.go:1-5`, `registry.go:3-4`). 2 subagentes independientes lo cazaron.
  → corregir el stack en el router/ESTADO; NO es un fix de código, es un fix de doc stale.
- **FE sin tests:** `web/` solo tiene `.stories.tsx`, cero `*.test.*`/`e2e/`. Muchas
  capabilities FE reales quedan `nc` (sin-check). → deuda de validación a BACKLOG.
- **Dead-code candidatos:** `domain/graph.go#UnidadDeTrabajo:119` · enums Banda/Canal/
  Procedencia/Origen sin comportamiento · `app-store.ts#toggleTheme` latente. → reclamar o borrar.
- **6 STUB** confirmados (publish, open, watcher, seed-parcial, listRuns 501, mint_token nc).

## PROPUESTAS (pendientes de firma — spec.md las detalla)

### P1 · Árbol de 3 ejes temporales
`CLAUDE.md` router puro · `BACKLOG.md` (lo-que-viene) · `ESTADO.md` (ahora) · `LEDGER.md`
índice + `ledger/HS-NN.md` hojas (lo-hecho). Regla madre: **pasado ≠ presente ≠ futuro
nunca comparten hoja**. → spec §2.

### P2 · Nombres de hojas
`BACKLOG.md` · `ESTADO.md` · carpeta `ledger/`. Alternativas si el operador prefiere:
`PENDIENTES.md`/`ABIERTO.md` · `STATUS.md`/`AHORA.md`. Sin preferencia declarada aún.

### P3 · Cabecera TTL por hoja
Formato espejo de `memory/` frontmatter, en una línea de cita:
`> vig: activo | archivado · revisar: AAAA-MM-DD`. Hoja `archivado` = no se carga salvo
grep explícito. → spec §4.

### P4 · Destino de `Siguiente`/`Deuda` de las fichas
Salen de cada `ledger/HS-NN.md` (historia inmutable) → consolidados en `BACKLOG.md` (futuro
vivo). Item cerrado se BORRA de BACKLOG (no se marca done: desaparece; su cierre ya vive en
la historia). → spec §3.
