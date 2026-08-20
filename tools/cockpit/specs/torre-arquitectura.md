# SPEC — Torre de Control · lente ARQUITECTURA (v1) — CONGELADA

> **Congelada 2026-07-03** (épica Torre de Control, F3 · ficha `DH-06`). La implementación
> ejecuta ESTO sin retocar diseño; cambio de diseño = nueva ficha DH-NN + edición explícita aquí.
> Extiende `torre-read-only.md` (RN-01..RN-19) — su §7 dejó esta lente prevista en el eje
> PROYECTO. Decisiones D5-D7 de esta SPEC: checkpoint expirado → recomendadas aplicadas
> (mecánica de la casa), registradas en `../epicas/torre-de-control/ESTADO.md` y en `DH-06`.

## 0 · Qué es

La lente de **arquitectura por sistema** de la torre: cada proyecto del eje PROYECTO gana un
objeto `arquitectura` que rinde el modelo curado de la célula (`arquitectura.yaml`,
arquitectura-como-dato — I-73) con **renderizador propio de DevHub** discriminado por
`meta.clase` (I-75 — el Atlas murió; el modelo del ECOSISTEMA jamás viene a DevHub, frontera #4
del NORTE). Read-only, como toda la torre.

## 1 · Fuente y contrato de lectura (D7: filesystem directo)

- **RN-20** · Fuente = `products/<slug>/arquitectura.yaml` **del workspace monitoreado**, solo
  para proyectos `tipo: celula` (mismo dirname que descubrió RN-03). Proyectos `tipo: sdd` →
  `sin-arquitectura` en v1 (los engagements aún no curan arquitectura — extensión v2).
- **RN-21** · **El YAML gated ES el contrato cross-célula**: schema L0 (`harness-descriptor`,
  enums) + gate de fábrica = contrato de datos; la torre lo lee read-only por convención —
  mismo patrón que `ledger.yaml` (RN-09), cero imports de código (frontera #3 del NORTE se
  cumple). CK-08 (Pull API con envelope) queda como precedente para cross-BINARIO en runtime;
  NO aplica a artefactos curados de un workspace que la torre ya monitorea por path.
- **RN-22** · `meta.clase` = **discriminador de render** (I-75). Ausente → default `modelo`:
  el archivo CURADO es arquitectura-como-dato por naturaleza (I-73) — ojo, distinto del default
  `arnes` del schema, que aplica al descriptor EMITIDO por adapters. Declaración explícita gana.
  Clase sin renderizador conocido → render degradado (lista plana), la fila jamás rompe.

## 2 · DevHub gana su arquitectura.yaml (D5: sí, con gate)

- **RN-23** · Nace `products/devhub/arquitectura.yaml`: curado a eventos de decisión (cada
  elemento con `fichas:` que RESUELVEN — DH-NN/I-NN/CK-NN), mismo formato que el de cockpit
  (CK-07). Declara `meta.clase: modelo` explícito (dogfood del discriminador RN-22).
  **SIN data.js**: la torre ES el renderizador de DevHub (I-75 — cada producto el suyo);
  emitir al shell de P4 sería resucitar el Atlas.
- **RN-24** · Check genérico de célula en fábrica: `tooling/scripts/check_arquitectura_celula.py`
  descubre `products/*/arquitectura.yaml` y valida (fichas resuelven en su ledger ·
  relaciones joinean from/to · `tipo` de relación ∈ `enums.relacion_tipo` · `estado` ∈
  `enums.estado` · `plano`/`tipo` de componente declarados · rutas existen · `meta.clase` ∈
  `enums.clase` si declarada). **Check-only, cero emisión.** Registrado en `gen_all.py`
  (patrón DH-04/RN-12: fábrica al servicio de la célula; elevarlo a governance de ecosistema =
  ficha I-NN del operador). Cockpit queda doble-cubierto (su gen valida+emite; este check
  valida) — mismas reglas espejadas, sin conflicto.

## 3 · Veredicto (D4 heredado: categoría fija + nativo)

- **RN-25** · Campo **ADITIVO** `arquitectura` en cada elemento de `proyectos[]` (el shape
  RN-15 no cambia para consumidores actuales). Categorías FIJAS:

  | Categoría | Cuándo | Dato nativo al lado |
  |---|---|---|
  | `con-arquitectura` | el YAML existe y parsea | `clase` · `id` · `nombre` · `version` · `proposito` · `total_componentes` · `total_relaciones` · `planos[]` · `tipos[]` · `componentes[]` · `relaciones[]` · `path` |
  | `sin-arquitectura` | archivo ausente (RN-02: empty-state honesto, jamás inventa) | — |
  | `no-medido` | YAML ilegible / error de lectura (RN-13: jamás se disfraza de otra cosa) | `motivo` |

  El modelo viaja **inline** en `GET /api/torre` (hecho barato: lectura YAML, cache ≤5s vía
  `torreFacts` — RN-16; tamaño acotado por ser modelo curado). Todo veredicto con `medido_en`
  (RN-17).

## 4 · UI — renderizador por clase (D6: tabla por planos)

- **RN-26** · Celda `arquitectura` en la fila de proyecto; click → detalle (RN-18: objeto del
  dominio operable). Render POR CLASE: `clase: modelo` → **tabla de componentes agrupada por
  PLANO** (las bandas del modelo), cabecera con id·clase·version·propósito; click en un
  componente → su detalle (propósito · fichas · ruta · estado · relaciones from/to con tipo).
  Clase desconocida → lista plana degradada (RN-22). Grafo = v2; nada más del visual se
  congela (herencia RN-19).

## 5 · Fuera de alcance (v1)

Grafo de relaciones · data.js/entry wiki para DevHub (la torre es el render) · arquitectura de
proyectos `tipo: sdd`/engagements (v2) · edición o escritura del modelo · veredicto de
"modelo desactualizado vs código" (eso es del gate de fábrica, no de la torre) · render del
modelo del ECOSISTEMA (`tooling/strategy/arquitectura.yaml` — frontera #4, jamás en DevHub) ·
refactor del gen de cockpit para compartir validador (cruce de célula sin ficha CK).
