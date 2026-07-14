// Fixtures honestas del Portafolio (plan §2.8, S1-D del paquete Slice 1) — calcadas de las
// salidas REALES del E2E de Slice 0 (`docs/product/stories/2026-07-13-portafolio-slice0-cimientos/
// paridad.md` §E2E vivo). PROHIBIDO inventar campos que el wire no tiene (`drift: 2`,
// `update: "1.2.0"`, `dirty: false` — los vicios G1/G2 del mockup que este slice mata). Cada
// const documenta de qué salida real sale, o por qué es sintética-con-shape-real.

import type { EntradaPortafolio } from "../model/types"

// (a) `harness@prenter-marketplace` — REAL (paridad.md §E2E, `arnesia portafolio escanear
// ~/Proyectos/luana-vitalia`): identidad PROVISIONAL (home="", el caso más común real — un
// plugin CC sin `arnes.l0.json`, S0-D14/D15); origen.registry crudo "alpacapurpura/
// prenter-marketplace" y version "0.5.2" cotejados a mano contra `installed_plugins.json`;
// deriva "en-deriva" REAL — hash de contenido distinto entre el cache
// (~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2/) y el checkout del marketplace
// (~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.2/, layout
// por-versión confirmado real); tipo `referenciada-cc` (la forma de instalación MÁS COMÚN
// real, GAP-1). `registries` de la entrada = CanonicalizarRepo(origen.registry) por S1-D3
// ("github.com/…", con prefijo — domain.CanonicalizarRepo asume github.com para "owner/repo"
// corto). `identidad.scope` es REPRESENTATIVO: RN-IDENT-2 exige uno con home="" (scopeRemote =
// el remote canonicalizado del proyecto contenedor), pero el remote exacto de luana-vitalia no
// quedó citado literal en paridad.md — seguimos la convención `github.com/alpacapurpura/*` que
// SÍ está confirmada para este mismo org en el hallazgo hermano de harness-studio (fixture b).
export const entradaHarnessEnDeriva: EntradaPortafolio = {
  clave: "sin-home~harness~github-com-alpacapurpura-luana-vitalia",
  identidad: { id: "harness", scope: "github.com/alpacapurpura/luana-vitalia" },
  registries: ["github.com/alpacapurpura/prenter-marketplace"],
  instalaciones: [
    {
      proyecto_path: "~/Proyectos/luana-vitalia",
      install_path: "~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2",
      tipo: "referenciada-cc",
      origen: {
        registry: "alpacapurpura/prenter-marketplace",
        version: "0.5.2",
      },
      deriva: "en-deriva",
      deriva_detalle:
        "hash de contenido distinto de la referencia " +
        "~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.2/",
    },
  ],
  agregado: "2026-07-13T18:42:00Z",
}

// (b) `proyecto-instalado` de harness-studio consigo mismo — REAL (paridad.md §E2E, `arnesia
// portafolio escanear ~/Proyectos/harness-studio`): "identidad provisional (sin
// `arnes.l0.json` en la raíz del repo, correcto: este repo no es-un-arnés, es la fábrica) +
// eslabón git-proyecto https://github.com/alpacapurpura/harness-studio.git". Sin manifiesto
// (`arnes.l0.json` NI `plugin.json`) el loader devuelve `Arnes: nil` (loader.go: "Ausente →
// modo degradado honesto") y el hallazgo `proyecto-instalado` no trae `IDConocido` (solo el
// detector de lock lo puebla) — `ResolverIdentidad` con `a==nil` e `idFallback==""` deja
// `identidad.id` LITERALMENTE VACÍO: dato real de la máquina, no un placeholder (el hueco
// mismo es la instancia G6 «identidad provisional VISIBLE» que este slice expone, no oculta).
// `scope` = scopeRemote (RN-IDENT-2 — gana sobre scopeLocal), el remote real citado arriba,
// canonicalizado. Sin `origen.registry` (git-proyecto NUNCA alimenta Registry, BR-5) → deriva
// "sin version instalada conocida" (adapters/portafolio/deriva.go:111, motivo real cuando
// `version==""`).
export const entradaProyectoInstaladoProvisional: EntradaPortafolio = {
  clave: "sin-home~~github-com-alpacapurpura-harness-studio",
  identidad: { id: "", scope: "github.com/alpacapurpura/harness-studio" },
  instalaciones: [
    {
      proyecto_path: "~/Proyectos/harness-studio",
      install_path: "~/Proyectos/harness-studio",
      tipo: "proyecto-instalado",
      origen: {},
      deriva: "deriva-no-evaluable",
      deriva_detalle: "sin version instalada conocida",
    },
  ],
  agregado: "2026-07-13T18:42:00Z",
}

// (c) `acme-cli` — SINTÉTICA con shape real (plan §2.8: "para los estados restantes" que el
// E2E vivo de Slice 0 no exhibió). Identidad RESUELTA (home≠"", scope ausente — mismo shape
// que `ResolverIdentidad` cuando el home resuelve). `empresas`/`canonico` poblados para las
// stories de zona canónico (T5). Dos instalaciones: la primera es el caso limpio real-shape
// («al-hilo» sin aviso ni discrepancias — la fila que exhibe la rama `ok` de `saludDe`,
// S1-D4); la segunda combina discrepancias (C-OR-6, texto EXACTO del formato real de
// `domain.ResolverOrigen`: "home declarado (%s) ≠ registry de adquisición (%s): normal en el
// modelo N:M, visible por trazabilidad") + aviso C-P-14 (texto EXACTO del formato real de
// `scanner.go`: "declarada en lock (%s@%s), dir ausente en el caché: %s") + `eslabones[]`
// crudos para la story `TrazabilidadEslabones` (BR-3). Por S1-D4, la instalación #2 empuja la
// entrada COMPLETA a `atencion` — la rama "ok" de saludDe se ejercita con esta MISMA
// instalación #1 aislada en el unit test, no con la entrada entera.
export const entradaCanonicaCompleta: EntradaPortafolio = {
  clave: "github-com-acme-acme-cli~acme-cli~",
  identidad: { home: "github.com/acme/acme-cli", id: "acme-cli" },
  nombre: "Acme CLI",
  descripcion: "Arnés de línea de comandos de Acme (fixture sintética, shape real).",
  empresas: ["alpacapurpura"],
  registries: ["github.com/acme/acme-cli"],
  canonico: { path: "~/dev/acme-cli", version: "2.1.0" },
  instalaciones: [
    {
      proyecto_path: "~/Proyectos/acme-app",
      install_path: "~/Proyectos/acme-app/.claude/plugins/acme-cli",
      tipo: "materializada",
      origen: { registry: "github.com/acme/acme-cli", version: "2.1.0" },
      deriva: "al-hilo",
    },
    {
      proyecto_path: "~/Proyectos/otro-app",
      install_path: "~/Proyectos/otro-app/.claude/plugins/acme-cli",
      tipo: "materializada",
      origen: {
        registry: "github.com/acme-fork/acme-cli",
        version: "2.0.0",
        eslabones: [
          { fuente: "manifiesto", campo: "home", valor: "github.com/acme/acme-cli" },
          { fuente: "cc-plugins", campo: "registry", valor: "acme-fork/acme-cli" },
          { fuente: "cc-plugins", campo: "version", valor: "2.0.0" },
        ],
        discrepancias: [
          "home declarado (github.com/acme/acme-cli) ≠ registry de adquisición " +
            "(github.com/acme-fork/acme-cli): normal en el modelo N:M, visible por trazabilidad",
        ],
      },
      deriva: "deriva-no-evaluable",
      deriva_detalle: "sin dir físico resoluble",
      aviso:
        "declarada en lock (acme-cli@2.1.0), dir ausente en el caché: ~/.arnesia/cache/acme-cli/2.1.0",
    },
  ],
  agregado: "2026-07-10T09:15:00Z",
}

// entradasDemo — las 3 fixtures juntas, orden estable (a→b→c). Conveniencia para stories/tests
// que necesitan una lista (lente empresa: `entradaCanonicaCompleta` cae en "alpacapurpura",
// (a) y (b) caen en «sin empresa» — grupo que S1-D8 exige SIEMPRE al final).
export const entradasDemo: EntradaPortafolio[] = [
  entradaHarnessEnDeriva,
  entradaProyectoInstaladoProvisional,
  entradaCanonicaCompleta,
]
