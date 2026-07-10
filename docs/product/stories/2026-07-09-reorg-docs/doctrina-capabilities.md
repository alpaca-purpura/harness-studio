# Doctrina — CAPABILITIES como SSoT funcional (docs-as-code) + enforcement

> vig: activo · revisar: 2026-08-01 · estado: PROPUESTA (espera firma antes de que el gate BLOQUEE)
> Mandato del operador (2026-07-09): «los capabilities son el SSoT de lo funcional y es nuestra
> documentación as-code»; un hook debe validar que **NADA crea código si no construye o modifica
> un capability y lo actualiza». Rigor máximo.

## 1. Principio

`CAPABILITIES.md` (o `capabilities/`) es la **única fuente de verdad de lo que el sistema hace**.
Derivado del CÓDIGO REAL (no de fichas ni comentarios). Cada capability apunta al código
autoritativo que lo implementa y al check que lo valida. Es Living Documentation (BDD): un
capability está `vivo` ⟺ su check pasa; si el código se mueve y el puntero se rompe → falla.

Corolario duro (el mandato): **todo cambio de código fuente traza a ≥1 capability.** Si tocás
código sin tocar el registro, el gate te frena. No hay código huérfano de capability.

## 2. Reglas (as-code, cada una con enforcer)

- **R1 · Integridad hacia adelante (sin docs colgantes).** Cada `punteros:` de cada capability
  RESUELVE: el archivo existe y el símbolo (`file#Func`/`file#Tipo`) existe. Puntero roto = FAIL.
  Determinista, sin LLM.
- **R2 · Cobertura hacia atrás (sin código huérfano).** Todo archivo fuente bajo `cmd/`,
  `internal/`, `web/src/`, `web/src-tauri/src/` está reclamado por ≥1 capability. Archivo sin
  dueño = FAIL (salvo allowlist: generados, `*_test.go`, `embed_*.go`, `main.rs` boilerplate).
- **R3 · Gate de commit.** Un commit que toca código fuente DEBE tocar `CAPABILITIES.md` en el
  mismo commit. (lefthook pre-commit; trunk-based = commit directo, así que es el gate real.)
- **R4 · Estado generado.** El `estado` (`vivo`/`degradado`/`planificado`/`sin-check`) lo da el
  check linkeado vía `arnesia conformance`, nunca a mano (D2/RF-181). Capability `sin-check`
  se muestra VISIBLE (honestidad, no se oculta — mismo criterio que el warn-fail del dogfood).

## 3. Contrato de puntero (anti-drift)

Los números de línea SE PUDREN. Preferencia estricta de puntero, de más estable a menos:
1. `paquete/` o `carpeta-widget/` — granularidad paquete (capability = subsistema entero).
2. `archivo.go` / `Componente.tsx` — granularidad archivo.
3. `archivo.go#Símbolo` (función/tipo/handler exportado) — granularidad función.
4. `archivo:línea` — SOLO cuando no hay símbolo estable (evitar; el validador avisa si abusás).

R1 valida contra símbolo/archivo, no contra línea → el registro no se rompe al reformatear.

## 4. Capas de enforcement (defensa en profundidad)

| Capa | Mecanismo | Momento | Fuerza |
|---|---|---|---|
| **Dev (CC)** | hook PostToolUse en Edit/Write de fuente → recuerda actualizar el registro | al editar | nudge |
| **Commit** | `lefthook` pre-commit: R3 (registro staged) + R1 (punteros resuelven) + R2 (sin huérfano nuevo) | al commitear | **bloquea** |
| **CI** | boundary `codigo-traza-a-capability` corrido por `arnesia conformance` | en push/PR | **bloquea** |

El validador de R1/R2 es un binario/adapter determinista (parse `CAPABILITIES.md` → set de paths
reclamados; walk del árbol fuente → set de paths; huérfanos = fuente − reclamados − allowlist;
colgantes = reclamados que no existen). Cero LLM, cero costo de contexto.

## 5. As-code (dónde vive la doctrina)

- **Nuevo boundary** `docs/architecture/boundaries/codigo-traza-a-capability.md` (L1 Living Documentation/BDD
  ↔ L2 realización aquí) con checks `enforced_by:` reales (R1/R2/R3).
- **Validador** Go: `internal/adapters/conformance/mechanism/` nuevo mecanismo `capability-trace`
  O `cmd/arnesia` subcomando `capabilities check`. Reusa el parser de rules del ruleset.
- **lefthook.yml**: job `capabilities` en pre-commit (llama al validador sobre staged).
- **hook CC**: en el kit (`kit/`) PostToolUse que nudge-ea al editar fuente sin tocar registro.
- **knowledge**: si aplica, nota en el nodo pertinente que el registro es SSoT funcional.

## 6. Rollout (no romper todo el día 1)

R2 (cobertura) falla en TODO commit hasta que el registro cubra el 100% del código. Por eso:

1. **Fase A — poblar** el registro real desde el código (los 7 subagentes → `CAPABILITIES.md`
   v1 con cobertura medida). Mientras, el gate corre en `warn`.
2. **Fase B — cerrar huecos**: cada archivo huérfano se reclama o se allowlist-ea con razón.
   Medir cobertura = archivos-reclamados / archivos-fuente. Meta: 100% (o allowlist explícito).
3. **Fase C — graduar a `block`**: cuando cobertura=100%, R1/R2/R3 pasan de `warn` a `enforced`.
   Recién ahí el hook «nada de código sin capability» es LEY. → requiere firma del operador.

## 7. Trazado historia→capability (cierra el loop)

Cada RF de cada paquete de trabajo declara `agrega|edita|borra CAP-NN` (RF-182). Así la historia
de usuario queda como *delta* auditable contra el saldo. El PARIDAD.md de cada feature suma una
columna `capability`. Una feature nueva sin capability declarado no pasa su gate de PARIDAD.

## 8. Riesgos

- **Cobertura 100% es dura** (48k LOC rust son build-artifacts; el shell real son 160 LOC). El
  allowlist debe ser explícito y con razón, no un cajón de sastre — el validador lista qué hay
  dentro. Revisión del allowlist en `docs/architecture/CADENCE.md`.
- **Granularidad inconsistente** entre subagentes → la síntesis normaliza (§3) antes de v1.
- **Bloqueo prematuro** mata la velocidad → rollout warn→block (§6) es obligatorio, no opcional.
