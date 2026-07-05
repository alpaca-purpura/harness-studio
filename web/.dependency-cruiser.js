// dependency-cruiser — grafo de imports de la SPA (CI-gate, espeja arch/fitness/.go-arch-lint.yml).
// enforced_by de arch/boundaries/fe-topologia-fsd.md · fe-taxonomia-componentes.md · fe-transporte-independiente.md.
// Correr (cwd = web/): `depcruise src --config .dependency-cruiser.js`.
//
// NOTA (honestidad, arch/CADENCE): la SPA aún no existe (fase 5). Esto DECLARA el grafo objetivo;
// los globs asumen la disposición FSD `src/{app,pages,widgets,features,entities,shared}/**`.
/** @type {import('dependency-cruiser').IConfiguration} */
module.exports = {
  forbidden: [
    // ---- FSD: dirección de capa (fe-topologia-fsd) ----
    {
      name: "layer-direction",
      comment: "FSD: una capa solo importa capas estrictamente inferiores",
      severity: "error",
      from: { path: "^src/entities/" },
      to: { path: "^src/(features|widgets|pages|app)/" },
    },
    {
      name: "features-no-upward",
      comment: "FSD: features no importa widgets/pages/app (solo capas inferiores)",
      severity: "error",
      from: { path: "^src/features/" },
      to: { path: "^src/(widgets|pages|app)/" },
    },
    {
      name: "widgets-no-upward",
      comment: "FSD: widgets no importa pages/app (solo capas inferiores)",
      severity: "error",
      from: { path: "^src/widgets/" },
      to: { path: "^src/(pages|app)/" },
    },
    {
      name: "pages-no-upward",
      comment: "FSD: pages no importa app (solo capas inferiores)",
      severity: "error",
      from: { path: "^src/pages/" },
      to: { path: "^src/app/" },
    },
    {
      name: "shared-no-upward",
      comment: "shared es la base: no importa entities/features/widgets/pages/app",
      severity: "error",
      from: { path: "^src/shared/" },
      to: { path: "^src/(entities|features|widgets|pages|app)/" },
    },
    {
      name: "no-sibling-feature-imports",
      comment: "una feature no importa otra feature (cross-slice prohibido)",
      severity: "error",
      from: { path: "^src/features/([^/]+)/.+" },
      to: { path: "^src/features/([^/]+)/.+", pathNot: "^src/features/$1/.+" },
    },
    {
      name: "no-deep-import",
      comment: "importa la slice por su public API (index.ts), no por deep-import a internos",
      severity: "warn",
      from: { pathNot: "^src/(app)/" },
      to: {
        path: "^src/(entities|features|widgets)/[^/]+/(?!index).+/.+",
        pathNot: "\\.(css|svg|png)$",
      },
    },

    // ---- Taxonomía de componentes (fe-taxonomia-componentes) ----
    {
      name: "canvas-not-chrome",
      comment: "el canvas (map-canvas, shared/canvas) no importa chrome (rail, dock, app/shell)",
      severity: "error",
      from: { path: "^src/(shared/canvas|widgets/map-canvas)/" },
      to: { path: "^src/(widgets/(command-rail|dock)|app/shell)/" },
    },
    {
      name: "chrome-not-canvas-internals",
      comment: "el chrome no hace deep-import a internos del canvas (solo props/store)",
      severity: "error",
      from: { path: "^src/(widgets/(command-rail|dock)|app/shell)/" },
      to: { path: "^src/(shared/canvas|widgets/map-canvas)/(?!index).+" },
    },
    {
      name: "ui-not-domain",
      comment: "shared/ui (primitivos/moléculas) no importa el dominio ni el store",
      severity: "error",
      from: { path: "^src/shared/ui/" },
      to: { path: "^src/(entities|features)/[^/]+/model/" },
    },

    // ---- Transporte ⊥ dominio + SSE singleton (fe-transporte-independiente) ----
    {
      name: "domain-not-transport",
      comment: "el dominio (entities/*/model) no importa el transporte (shared/api)",
      severity: "error",
      from: { path: "^src/entities/[^/]+/model/" },
      to: { path: "^src/shared/api/" },
    },
    {
      name: "sse-singleton",
      comment: "solo app/realtime importa la conexión SSE cruda; ninguna feature/widget/entity",
      severity: "error",
      from: { pathNot: "^src/app/realtime/" },
      to: { path: "^src/shared/api/sse(/|$)" },
    },

    // ---- Higiene general ----
    { name: "no-circular", severity: "error", from: {}, to: { circular: true } },
    { name: "no-orphans", severity: "warn", from: { orphan: true, pathNot: "\\.d\\.ts$" }, to: {} },
    { name: "not-to-unresolvable", severity: "error", from: {}, to: { couldNotResolve: true } },
  ],
  options: {
    doNotFollow: { path: "node_modules" },
    tsConfig: { fileName: "tsconfig.json" },
    tsPreCompilationDeps: true,
    reporterOptions: { dot: { collapsePattern: "node_modules/(@[^/]+/[^/]+|[^/]+)" } },
  },
};
