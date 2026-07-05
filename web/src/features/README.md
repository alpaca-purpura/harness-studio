# features/ (FSD)

Interacciones de usuario con valor de negocio: **conversar con el arnés** (dock ⌘K → Claude Code
headless por el daemon), **crear/editar caja**, **evaluar A/B**, **actualizar estándar**, **publicar
al marketplace**. Cada feature = carpeta con Public API (`index.ts`), con segmentos `ui/model/api`.

Vacío por ahora. Regla FSD: importa `entities/shared`, nunca `pages/widgets/app` ni otra `feature`.
El transporte (SSE/HTTP al daemon) vive en `shared/api`; las features lo consumen, no lo abren.
