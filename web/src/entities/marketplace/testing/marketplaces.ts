// Fixtures honestas de Marketplace (design.md §8.5) — **copiadas de la máquina, no inventadas**:
// cada const dice de qué archivo real sale. Los archivos: `~/.claude/plugins/known_marketplaces.json`
// (5 marketplaces) · `.../marketplaces/prenter-marketplace/.claude-plugin/marketplace.json`
// (2 entradas, MISMO `source`) + su `catalogo.json` · `.../claude-plugins-official/…` (273
// entradas, 4 formas de `source`, `renames`) · `.../caveman/…` (`source: "./"`, `owner` con
// `url`). Verificado 2026-07-25.
//
// PROHIBIDO inventar campos que el wire no tiene (los vicios G1/G2 que el Slice 1 mató). Las
// cifras que la UI muestra salen SIEMPRE de acá o del wire — el mockup tiene cifras ilustrativas
// inconsistentes a propósito (dice «41» y «8» donde el real es 273 y 2, design.md §2 C14) y
// NINGUNA se hardcodea «para que quede igual al dibujo».

import type {
  Catalogo,
  EntradaCatalogo,
  MarketplaceConocido,
  SituacionCatalogo,
} from "../model/types"

// El literal EXACTO del mockup firmado — la única fila del dibujo que trae el par correcto
// `disabled` + `title` (design.md §2 C5). Vive en el dominio Go (`AccionDeSituacion`, §6.3) y
// llega por el wire; acá se repite SOLO porque las fixtures espejan el wire.
const NO_APLICA = "no aplica: solo arneses propios"

// ── Filas del plano (S2) ────────────────────────────────────────────────────────────────────

// mkPrenterPropio — REAL: `known_marketplaces.json#prenter-marketplace` (repo
// `alpacapurpura/prenter-marketplace`, installLocation y lastUpdated literales) + el registro
// propio del operador ⇒ los DOS eslabones (collect-all AG-D9). `lectura.entradas: 2` es la
// cuenta REAL de `plugins[]` de ese marketplace (NO las «8» del mockup, design.md §2 C14).
export const mkPrenterPropio: MarketplaceConocido = {
  nombre: "prenter-marketplace",
  repo: "github.com/alpacapurpura/prenter-marketplace",
  clase: "propio",
  eslabones: ["cc-known-marketplaces", "declarado-por-operador"],
  install_location: "/home/chalreme/.claude/plugins/marketplaces/prenter-marketplace",
  cc_actualizado: "2026-07-10T00:36:43.459Z",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:07:33Z", entradas: 2, fuente: "local" },
  registrado: "2026-07-25T14:02:11Z",
}

// mkOficialReferencia — REAL: `known_marketplaces.json#claude-plugins-official`. `entradas: 273`
// es la cuenta REAL del archivo de 159 KB (AG-D16); `catalogo.json` NO existe en ese checkout
// (degradado sin ruido, E-06). Clase `referencia`: no publicamos ahí, solo resuelve procedencia.
export const mkOficialReferencia: MarketplaceConocido = {
  nombre: "claude-plugins-official",
  repo: "github.com/anthropics/claude-plugins-official",
  clase: "referencia",
  eslabones: ["cc-known-marketplaces"],
  install_location: "/home/chalreme/.claude/plugins/marketplaces/claude-plugins-official",
  cc_actualizado: "2026-07-24T17:15:06.920Z",
  lectura: { tipo: "leido", cuando: "2026-07-23T09:12:00Z", entradas: 273, fuente: "local" },
}

// mkSinAcceso — SINTÉTICA con shape real: un marketplace propio declarado por el operador que
// todavía no tiene checkout de CC y cuyo remoto no responde. `motivo` con el formato REAL del
// stderr de `gh api`. **Nunca se convierte en catálogo vacío** (BR-4): la fila NO navega.
export const mkSinAcceso: MarketplaceConocido = {
  nombre: "vitalia-arneses",
  repo: "github.com/vitalia/arneses",
  clase: "propio",
  eslabones: ["declarado-por-operador"],
  lectura: {
    tipo: "sin-acceso",
    motivo: "gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual",
  },
  registrado: "2026-07-25T13:40:02Z",
}

// mkNoLeido — REAL: `known_marketplaces.json#caveman` (repo `JuliusBrussee/caveman`,
// installLocation y lastUpdated literales). Nunca lo leímos ⇒ `no-leido`, que usa la clase
// `sin-senal` (transparente + dashed): «no sé» ≠ «sano», JAMÁS el color de `ok`.
export const mkNoLeido: MarketplaceConocido = {
  nombre: "caveman",
  repo: "github.com/juliusbrussee/caveman",
  clase: "referencia",
  eslabones: ["cc-known-marketplaces"],
  install_location: "/home/chalreme/.claude/plugins/marketplaces/caveman",
  cc_actualizado: "2026-06-02T21:24:11.366Z",
  lectura: { tipo: "no-leido" },
}

// mkUrlNoResuelve — SINTÉTICA con shape real: la 4ta rama de AG-D8 decisión 4. Distinta de
// `sin-acceso` (que es «no puedo mirar»): acá el archivo se alcanzó y NO parsea.
export const mkUrlNoResuelve: MarketplaceConocido = {
  nombre: "nordia-plugins-rrhh",
  repo: "github.com/nordia/plugins-rrhh",
  clase: "propio",
  eslabones: ["declarado-por-operador"],
  lectura: {
    tipo: "url-no-resuelve",
    motivo: "marketplace.json inválido: unexpected end of JSON input (offset 4096)",
  },
  registrado: "2026-07-24T18:05:00Z",
}

// mkConDiscrepancia — BR-8: el mismo `nombre` con `repo` DISTINTO entre eslabones. Gana el
// declarado (el operador es la autoridad de lo que él registró) y la divergencia queda VISIBLE
// con LOS DOS valores crudos — jamás se elige en silencio. Texto con el formato de
// `MergeMarketplaces` (design.md §3.1.7 regla 2).
export const mkConDiscrepancia: MarketplaceConocido = {
  nombre: "ponytail",
  repo: "github.com/alpacapurpura/ponytail",
  clase: "propio",
  eslabones: ["cc-known-marketplaces", "declarado-por-operador"],
  install_location: "/home/chalreme/.claude/plugins/marketplaces/ponytail",
  cc_actualizado: "2026-07-22T20:33:42.715Z",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:00:00Z", entradas: 1, fuente: "local" },
  discrepancias: [
    "repo distinto entre eslabones: cc-known-marketplaces dice " +
      "github.com/dietrichgebert/ponytail y declarado-por-operador dice " +
      "github.com/alpacapurpura/ponytail: se usa el declarado",
  ],
  registrado: "2026-07-25T14:00:00Z",
}

// marketplacesDemo — las 5 filas juntas, orden de wire (sin ordenar): las stories ejercitan
// `ordenarMarketplaces` sobre esto.
export const marketplacesDemo: MarketplaceConocido[] = [
  mkOficialReferencia,
  mkSinAcceso,
  mkPrenterPropio,
  mkNoLeido,
  mkUrlNoResuelve,
]

// ── Catálogos ───────────────────────────────────────────────────────────────────────────────

// catPrenter — REAL: las 2 entradas literales de `prenter-marketplace` (`harness` y
// `harness-beta`, **mismo `source` `./plugins/harness/0.5.3`** ⇒ `comparte_source_con` cruzado,
// AG-D11) + el enriquecimiento REAL de su `catalogo.json` (`canales` estable/beta = 0.5.3 y las
// 4 `versiones`, con 0.5.1 `deprecada`). `version` DERIVADA del último segmento de la ruta
// (`derivada-de-source`, BR-2): ese archivo no declara el campo `version`.
//
// Situaciones: `harness` cruza con la entrada REAL del Portafolio (canónico 0.5.2 vs estante
// 0.5.3 ⇒ `estante-adelantado`) por `faceta-registry` (la identidad de esa copia es provisional:
// el caso dominante real). `harness-beta` no está ⇒ `no-lo-tengo`, la ÚNICA celda de la tabla de
// §6.3 que AG-D17 habilita.
export const catPrenter: Catalogo = {
  marketplace: "prenter-marketplace",
  clase: "propio",
  repo: "github.com/alpacapurpura/prenter-marketplace",
  owner_nombre: "Prenter",
  owner_email: "hola@alpacapurpura.lat",
  descripcion:
    "Private Prenter marketplace (instancia 0 — KIT-06/I-72). Access = your git credentials " +
    "(client code). Channels are index entries; catalogo.json is the version-state SSoT.",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:07:33Z", entradas: 2, fuente: "local" },
  entradas: [
    {
      nombre: "harness",
      source: {
        tipo: "ruta-relativa",
        crudo: "./plugins/harness/0.5.3",
        ruta: "./plugins/harness/0.5.3",
      },
      version: "0.5.3",
      version_de: "derivada-de-source",
      descripcion:
        "Canal ESTABLE — tech/sistema-agnostic dev-process kit (SDD spine, gates, roles, " +
        "embedded telemetry). Reads the seam (project.config.yaml).",
      comparte_source_con: ["harness-beta"],
      estado_canal: "habilitada",
      situacion: {
        tipo: "estante-adelantado",
        mia: "0.5.2",
        estante: "0.5.3",
        clave_portafolio: "sin-home~harness~github-com-alpacapurpura-luana-vitalia",
        via: "faceta-registry",
      },
      accion: {
        verbo: "actualizar-mi-copia",
        habilitada: false,
        motivo: "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)",
      },
    },
    {
      nombre: "harness-beta",
      source: {
        tipo: "ruta-relativa",
        crudo: "./plugins/harness/0.5.3",
        ruta: "./plugins/harness/0.5.3",
      },
      version: "0.5.3",
      version_de: "derivada-de-source",
      descripcion:
        "Canal BETA (beta-testing) — same kit, pre-promotion. One channel per machine: " +
        "switching = uninstall + install.",
      comparte_source_con: ["harness"],
      estado_canal: "habilitada",
      situacion: { tipo: "no-lo-tengo" },
      accion: { verbo: "traer-canonico", habilitada: true },
    },
  ],
  canales: { estable: "0.5.3", beta: "0.5.3" },
  versiones: [
    {
      version: "0.5.0",
      estado: "habilitada",
      fuente: "prenter-harness@v0.5.0",
      fecha: "2026-07-02",
    },
    {
      version: "0.5.1",
      estado: "deprecada",
      fuente: "prenter-harness@v0.5.1",
      fecha: "2026-07-02",
    },
    {
      version: "0.5.2",
      estado: "habilitada",
      fuente: "prenter-harness@v0.5.2",
      fecha: "2026-07-02",
    },
    {
      version: "0.5.3",
      estado: "habilitada",
      fuente: "prenter-harness@v0.5.3",
      fecha: "2026-07-03",
    },
  ],
}

// entradasOficialMuestra — MUESTRA de 25 filas LITERALES de
// `claude-plugins-official/.claude-plugin/marketplace.json` (273 filas / 159 KB). Cubre las
// **4 formas reales de `source`** (ruta relativa · git-subdir · url · github), `version` ausente
// (259/273) y `version` semver real (14/273, p.ej. `clangd-lsp`), y los 6 destinos de `renames`
// con su `nombre_anterior`. Por qué 25 y no 273: el archivo real infla el árbol sin agregar una
// rama de comportamiento (plan-pruebas §0 lo registra como decisión); las 273 se ejercitan en la
// capa E (vivo) y en la story de lazy con un generador.
//
// Todas las acciones vienen `habilitada: false` con el literal `NO_APLICA`: es clase
// `referencia`, y BR-1 no tiene excepciones. `frontend-design` es la fila que SÍ cruza con el
// Portafolio real y, sin versión declarable, sale `no-comparable` con motivo (BR-9).
export const entradasOficialMuestra: EntradaCatalogo[] = [
  {
    nombre: "agent-sdk-dev",
    source: {
      tipo: "ruta-relativa",
      crudo: "./plugins/agent-sdk-dev",
      ruta: "./plugins/agent-sdk-dev",
    },
    descripcion: "Development kit for working with the Claude Agent SDK",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "frontend-design",
    source: {
      tipo: "ruta-relativa",
      crudo: "./plugins/frontend-design",
      ruta: "./plugins/frontend-design",
    },
    descripcion:
      "Create distinctive, production-grade frontend interfaces with high design quality. Generates creative, polished code that avoids generic AI aesthetics.",
    situacion: {
      tipo: "no-comparable",
      motivo: "el catálogo no declara versión de esta entrada ni se puede derivar de su source",
      clave_portafolio: "sin-home~frontend-design~github-com-alpacapurpura-luana-platform",
      via: "faceta-registry",
    },
    accion: {
      verbo: "",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "asana",
    source: {
      tipo: "ruta-relativa",
      crudo: "./external_plugins/asana",
      ruta: "./external_plugins/asana",
    },
    descripcion:
      "Asana project management integration. Create and manage tasks, search projects, update assignments, track progress, and integrate your development workflow with Asana's work management platform.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "context7",
    source: {
      tipo: "ruta-relativa",
      crudo: "./external_plugins/context7",
      ruta: "./external_plugins/context7",
    },
    descripcion:
      "Upstash Context7 MCP server for up-to-date documentation lookup. Pull version-specific documentation and code examples directly from source repositories into your LLM context.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "security-guidance",
    source: {
      tipo: "ruta-relativa",
      crudo: "./plugins/security-guidance",
      ruta: "./plugins/security-guidance",
    },
    version: "2.0.6",
    version_de: "campo-version",
    descripcion:
      "Security review for Claude-generated code. Pattern-based warnings on edits, LLM-powered diff review on Stop, and an agentic commit reviewer that catches injection, XSS, SSRF, hardcoded secrets, and 25+ other vulnerability classes.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "clangd-lsp",
    source: {
      tipo: "ruta-relativa",
      crudo: "./plugins/clangd-lsp",
      ruta: "./plugins/clangd-lsp",
    },
    version: "1.0.0",
    version_de: "campo-version",
    descripcion: "C/C++ language server (clangd) for code intelligence",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "42crunch-api-security-testing",
    source: {
      tipo: "git-subdir",
      crudo:
        "git-subdir:https://github.com/42Crunch-AI/claude-plugins.git#plugins/api-security-testing@v1.5.5",
      ruta: "plugins/api-security-testing",
      url: "https://github.com/42Crunch-AI/claude-plugins.git",
      ref: "v1.5.5",
      sha: "30287f5e3f122a646d1ac5ca3ab96e130c52a3ad",
    },
    descripcion:
      "Automate API security directly in Claude Code with 42Crunch - automatically audit OpenAPI specs, detect vulnerabilities aligned with OWASP API Security risks (including BOLA/BFLA), and apply AI-powered fixes. Designed for AI-assisted development workflows, it provides continuous guardrails through an audit->scan->remediate->validate loop, ensuring APIs meet enterprise security standards before deployment.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "airwallex-agentos",
    source: {
      tipo: "git-subdir",
      crudo:
        "git-subdir:https://github.com/airwallex/airwallex-marketplace.git#plugins/airwallex-agentos@master",
      ruta: "plugins/airwallex-agentos",
      url: "https://github.com/airwallex/airwallex-marketplace.git",
      ref: "master",
      sha: "b0bd2c3d65da47e39db8c779501119376d91c431",
    },
    descripcion:
      "Bring Airwallex's global financial infrastructure to Claude. Orchestrate actions across your account in plain language, e.g., set up invoices from a PO, onboard suppliers from invoices, and check current cash position across currencies. AgentOS bundles pre-built finance Skills with MCP servers. A public CLI connects your agent to Airwallex's capabilities.",
    nombre_anterior: ["airwallex"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "valtown",
    source: {
      tipo: "git-subdir",
      crudo: "git-subdir:https://github.com/val-town/plugins.git#plugin@main",
      ruta: "plugin",
      url: "https://github.com/val-town/plugins.git",
      ref: "main",
      sha: "1bd1c3f93161d88908dc0838aa81980ff1b2e4f7",
    },
    descripcion:
      "Build and deploy on Val Town. Bundles the Val Town MCP server and platform skills (HTTP vals, cron/intervals, SQLite, email, OAuth, React UI, third-party integrations, templates).",
    nombre_anterior: ["vals"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "datadog",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/datadog-labs/claude-code-plugin.git#@",
      url: "https://github.com/datadog-labs/claude-code-plugin.git",
      sha: "c5c062abba0df33f6bfc2c0fd0f8d17857e3fa2c",
    },
    descripcion:
      "Use Datadog directly in Claude Code through a preconfigured Datadog MCP server. Query logs, metrics, traces, dashboards, and more through natural conversation. This plugin is in preview.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "honeycomb",
    source: {
      tipo: "git-subdir",
      crudo: "git-subdir:https://github.com/honeycombio/agent-skill.git#honeycomb@main",
      ruta: "honeycomb",
      url: "https://github.com/honeycombio/agent-skill.git",
      ref: "main",
      sha: "189553c8a879cfaaf206ba9cc68c1b2117d12a76",
    },
    descripcion:
      "Skills, agents, and workflows for Honeycomb observability — query patterns, production investigations, SLOs, OpenTelemetry instrumentation, and Beeline migration. Designed to complement the Honeycomb MCP server.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "sentry",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/getsentry/plugin-claude.git#@",
      url: "https://github.com/getsentry/plugin-claude.git",
      sha: "8ed4f25563d90806e8a50c7fb78170ede9ca2f5e",
    },
    descripcion:
      "Sentry error monitoring integration. Access error reports, analyze stack traces, search issues by fingerprint, and debug production errors directly from your development environment.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "agentforce-adlc",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/SalesforceAIResearch/agentforce-adlc.git#@",
      url: "https://github.com/SalesforceAIResearch/agentforce-adlc.git",
      sha: "f2433e68c8213f3eee0358093552ddc4de40602f",
    },
    descripcion:
      "Agentforce Agent Development Life Cycle — author, discover, scaffold, deploy, test, and optimize .agent files",
    nombre_anterior: ["adlc"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "convex",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/get-convex/convex-backend-skill.git#@",
      url: "https://github.com/get-convex/convex-backend-skill.git",
      sha: "8b01557dc3dd2e390ad090025c67f5e022bd21ad",
    },
    descripcion:
      "Official Convex plugin for Claude Code with bundled Convex skills, the convex-expert subagent for code-writing, a runtime-error monitor, and MCP access for backend development, schema design, real-time features, auth, file storage, scheduled jobs, and AI agents.",
    nombre_anterior: ["convex-backend"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "build-with-wordpress",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/Automattic/claude-code-wordpress.com.git#@",
      url: "https://github.com/Automattic/claude-code-wordpress.com.git",
      sha: "052ca970df2c577d7c651e784935186ff93e6779",
    },
    descripcion:
      "Craft production-grade WordPress sites and applications. Everything from themes and plugins to commerce and deployment.",
    nombre_anterior: ["wordpress.com"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "qodo",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/qodo-ai/qodo-skills.git#@",
      url: "https://github.com/qodo-ai/qodo-skills.git",
      sha: "ad848afd32eff172ce9073d9b1b1efe462ea0af2",
    },
    descripcion:
      "Qodo Skills provides a curated library of reusable AI agent capabilities that extend Claude's functionality for software development workflows. Each skill is designed to integrate seamlessly into your development process, enabling tasks like code quality checks, automated testing, security scanning, and compliance validation. Skills operate across your entire SDLC—from IDE to CI/CD—ensuring consistent standards and catching issues early.",
    nombre_anterior: ["qodo-skills"],
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "linear",
    source: {
      tipo: "ruta-relativa",
      crudo: "./external_plugins/linear",
      ruta: "./external_plugins/linear",
    },
    descripcion:
      "Linear issue tracking integration. Create issues, manage projects, update statuses, search across workspaces, and streamline your software development workflow with Linear's modern issue tracker.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "notion",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/makenotion/claude-code-notion-plugin.git#@",
      url: "https://github.com/makenotion/claude-code-notion-plugin.git",
      sha: "9847f2aa1a15f25df35ed1fb7b4557dbb60cd651",
    },
    descripcion:
      "Notion workspace integration. Search pages, create and update documents, manage databases, and access your team's knowledge base directly from Claude Code for seamless documentation workflows.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "stripe",
    source: {
      tipo: "git-subdir",
      crudo: "git-subdir:https://github.com/stripe/ai.git#providers/claude/plugin@main",
      ruta: "providers/claude/plugin",
      url: "https://github.com/stripe/ai.git",
      ref: "main",
      sha: "84c364c683c58ba8dfbc003d812e6a35f4c97b47",
    },
    descripcion: "Stripe development plugin for Claude",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "fullstory",
    source: {
      tipo: "github",
      crudo: "github:fullstorydev/fullstory-skills#@",
      repo: "fullstorydev/fullstory-skills",
      sha: "b20614e2d08d7a7c70775bb62b5af640f60b024b",
    },
    descripcion:
      "Connect Claude to Fullstory to query behavioral analytics, session replays, and customer experience insights.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "adobe-for-creativity",
    source: {
      tipo: "git-subdir",
      crudo:
        "git-subdir:https://github.com/adobe/skills.git#plugins/creative-cloud/adobe-for-creativity@main",
      ruta: "plugins/creative-cloud/adobe-for-creativity",
      url: "https://github.com/adobe/skills.git",
      ref: "main",
      sha: "17ef6fb53d2eb23158dec11823ff569258b7a26e",
    },
    descripcion:
      "Harness Adobe's creative AI-powered tools to edit images, automate design workflows, and bring creative visions to life — from background removal to vectorization and professional retouching.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "ai-plugins",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/endorlabs/ai-plugins.git#@",
      url: "https://github.com/endorlabs/ai-plugins.git",
      sha: "a6737fcf72336399e212e45cd25a250c2df3b7b4",
    },
    descripcion:
      "Set up endorctl and use Endor Labs to scan, prioritize, and fix security risks across your software supply chain",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "aikido",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/AikidoSec/aikido-claude-plugin.git#@",
      url: "https://github.com/AikidoSec/aikido-claude-plugin.git",
      sha: "17b7c113b50aa57c43856612ce98ce9d96bc37f8",
    },
    descripcion:
      "Aikido Security scanning for Claude Code — SAST, secrets, and IaC vulnerability detection powered by the Aikido MCP server.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "airtable",
    source: {
      tipo: "git-subdir",
      crudo: "git-subdir:https://github.com/Airtable/skills.git#plugins/airtable@main",
      ruta: "plugins/airtable",
      url: "https://github.com/Airtable/skills.git",
      ref: "main",
      sha: "812ee67f1fd3d76fb45ff8df40afaa0448602ba8",
    },
    descripcion:
      "Airtable is the database and operations layer for your agents — whether running product, marketing, sales, ops, HR, or a custom business app. It combines structured data with multiplayer visual surfaces (grid, kanban, calendar, gallery, timeline) humans and agents share — plus sync integrations to Jira, Salesforce, Zendesk, Google Drive, Databricks, and the rest of your stack, all backed by enterprise governance. This plugin makes Claude fluent in Airtable: creating bases and schema, working with records, and sharing UI for collaboration. Bundles the official Airtable MCP server.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
  {
    nombre: "alloydb",
    source: {
      tipo: "url",
      crudo: "url:https://github.com/gemini-cli-extensions/alloydb.git#@",
      url: "https://github.com/gemini-cli-extensions/alloydb.git",
      sha: "3ce1fe143644c31c0ffeb904b31634f8cea06e0c",
    },
    descripcion: "Create, connect, and interact with an AlloyDB for PostgreSQL database and data.",
    situacion: {
      tipo: "no-lo-tengo",
    },
    accion: {
      verbo: "traer-canonico",
      habilitada: false,
      motivo: "no aplica: solo arneses propios",
    },
  },
]

// catOficialMuestra — el catálogo de referencia con las 25 filas de arriba. `canales`/`versiones`
// AUSENTES: ese checkout no tiene `catalogo.json` y el degradado es SIN ruido (E-06,
// `lectura.motivo` vacío). `lectura.entradas: 273` es la cuenta REAL del archivo — la muestra es
// de las filas, no de la cifra (E-58: un recorte se declararía en `truncado`, y acá no hay
// recorte del backend: hay una muestra de fixture, que es otra cosa y está documentada arriba).
export const catOficialMuestra: Catalogo = {
  marketplace: "claude-plugins-official",
  clase: "referencia",
  repo: "github.com/anthropics/claude-plugins-official",
  owner_nombre: "Anthropic",
  owner_email: "support@anthropic.com",
  descripcion:
    "Directory of popular Claude Code extensions including development tools, productivity " +
    "plugins, and MCP integrations",
  lectura: { tipo: "leido", cuando: "2026-07-23T09:12:00Z", entradas: 273, fuente: "local" },
  entradas: entradasOficialMuestra,
}

// catVacio — `entradas: []` con `lectura` EXITOSA: «leí, hace 4 min, y este marketplace no
// declara ningún arnés». Es una AFIRMACIÓN evidenciada, distinta de `catNull` (E-32 vs E-31).
export const catVacio: Catalogo = {
  marketplace: "caveman",
  clase: "referencia",
  repo: "github.com/juliusbrussee/caveman",
  owner_nombre: "Julius Brussee",
  owner_url: "https://github.com/JuliusBrussee",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:03:00Z", entradas: 0, fuente: "local" },
  entradas: [],
}

// catNull — **el fixture que prueba que la UI no dice «no tiene arneses»**: `entradas: null` =
// «no sé» (design.md §2 C4, BR-4, normativo en el cable). La superficie muestra el MOTIVO y no
// afirma nada sobre cuántos arneses hay.
export const catNull: Catalogo = {
  marketplace: "vitalia-arneses",
  clase: "propio",
  repo: "github.com/vitalia/arneses",
  lectura: {
    tipo: "sin-acceso",
    motivo: "gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual",
  },
  entradas: null,
}

// catNullConCacheViejo — BR-3 + BR-4 juntas: hubo una lectura buena hace 2 días y AHORA no se
// puede leer. `cuando` se conserva ⇒ la fila dice «leído hace 2 días · ahora sin acceso — …».
// Nunca se afirma frescura que no hubo, y nunca se calla el motivo.
export const catDegradadoConCache: Catalogo = {
  marketplace: "vitalia-arneses",
  clase: "propio",
  repo: "github.com/vitalia/arneses",
  lectura: {
    tipo: "sin-acceso",
    cuando: "2026-07-23T14:07:33Z",
    entradas: 1,
    fuente: "remoto",
    motivo: "dial tcp: lookup github.com: no such host",
  },
  entradas: [
    {
      nombre: "reclutamiento-seleccion",
      source: {
        tipo: "ruta-relativa",
        crudo: "./plugins/reclutamiento-seleccion/1.1.0",
        ruta: "./plugins/reclutamiento-seleccion/1.1.0",
      },
      version: "1.1.0",
      version_de: "derivada-de-source",
      situacion: { tipo: "al-hilo", mia: "1.1.0", estante: "1.1.0", via: "home-declarado" },
      accion: { verbo: "", habilitada: false },
    },
  ],
}

// ── Las 6 situaciones (una fila por rama) ───────────────────────────────────────────────────

// Los 4 tooltips LITERALES de `AccionDeSituacion` para clase `propio` (design.md §6.3). Viven en
// el dominio Go; acá se espejan porque la fixture ES el wire.
const MOTIVO_PUBLICAR = "Publicar se construye en su propio paquete (ítem 3 del outcome)"
const MOTIVO_ACTUALIZAR =
  "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)"
const MOTIVO_REPARAR = "Reparar se construye en su propio paquete (ítem 5 del outcome)"
const MOTIVO_NO_COMPARABLE =
  "el catálogo no declara versión de esta entrada ni se puede derivar de su source"

function filaSituacion(
  nombre: string,
  situacion: SituacionCatalogo,
  verbo: EntradaCatalogo["accion"]["verbo"],
  motivo: string | undefined,
  habilitada = false,
): EntradaCatalogo {
  return {
    nombre,
    source: {
      tipo: "ruta-relativa",
      crudo: `./plugins/${nombre}/0.5.3`,
      ruta: `./plugins/${nombre}/0.5.3`,
    },
    version: "0.5.3",
    version_de: "derivada-de-source",
    situacion,
    accion: { verbo, habilitada, ...(motivo ? { motivo } : {}) },
  }
}

// situacionesLas6 — una fila por rama de spec §4.3, clase `propio`, con la acción EXACTA que
// `AccionDeSituacion` devuelve (design.md §6.3 con la corrección de §13.11): **7 de 8 celdas
// siguen `habilitada: false`**; la única habilitada es `propio × no-lo-tengo`, que AG-D17 abrió.
export const situacionesLas6: EntradaCatalogo[] = [
  filaSituacion("dev-full-cycle", { tipo: "no-lo-tengo" }, "traer-canonico", undefined, true),
  filaSituacion(
    "code-review-base",
    {
      tipo: "al-hilo",
      mia: "0.5.3",
      estante: "0.5.3",
      clave_portafolio: "github-com-alpacapurpura-prenter-marketplace~code-review-base~",
      via: "home-declarado",
    },
    "",
    undefined,
  ),
  filaSituacion(
    "harness",
    {
      tipo: "mi-copia-adelantada",
      mia: "0.5.4",
      estante: "0.5.3",
      clave_portafolio: "github-com-alpacapurpura-prenter-marketplace~harness~",
      via: "home-declarado",
    },
    "publicar",
    MOTIVO_PUBLICAR,
  ),
  filaSituacion(
    "ux-nordia",
    {
      tipo: "estante-adelantado",
      mia: "0.2.0",
      estante: "0.5.3",
      clave_portafolio: "github-com-alpacapurpura-prenter-marketplace~ux-nordia~",
      via: "home-declarado",
    },
    "actualizar-mi-copia",
    MOTIVO_ACTUALIZAR,
  ),
  filaSituacion(
    "reclutamiento-seleccion",
    {
      tipo: "instalaciones-en-deriva",
      cuantas: 1,
      mia: "0.5.2",
      estante: "0.5.3",
      clave_portafolio: "github-com-alpacapurpura-prenter-marketplace~reclutamiento-seleccion~",
      via: "home-declarado",
    },
    "reparar",
    MOTIVO_REPARAR,
  ),
  {
    nombre: "legacy-onboarding",
    source: {
      tipo: "ruta-relativa",
      crudo: "./plugins/legacy-onboarding/latest",
      ruta: "./plugins/legacy-onboarding/latest",
    },
    situacion: {
      tipo: "no-comparable",
      motivo: MOTIVO_NO_COMPARABLE,
      clave_portafolio: "github-com-alpacapurpura-prenter-marketplace~legacy-onboarding~",
      via: "home-declarado",
    },
    accion: { verbo: "", habilitada: false, motivo: MOTIVO_NO_COMPARABLE },
  },
]

// catLas6Situaciones — el catálogo propio con las 6 ramas representadas.
export const catLas6Situaciones: Catalogo = {
  marketplace: "prenter-marketplace",
  clase: "propio",
  repo: "github.com/alpacapurpura/prenter-marketplace",
  owner_nombre: "Prenter",
  owner_email: "hola@alpacapurpura.lat",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:07:33Z", entradas: 6, fuente: "local" },
  entradas: situacionesLas6,
}

// situacionesLas6Referencia — LAS MISMAS 6 ramas con clase `referencia`: `AccionDeSituacion`
// devuelve `habilitada: false` en TODA rama, sin excepción (BR-1 + boundary
// marketplace-referencia-es-solo-procedencia). `no-lo-tengo` conserva el verbo (para poder
// mostrar el botón apagado con su motivo); las otras 5 no ofrecen verbo alguno.
export const situacionesLas6Referencia: EntradaCatalogo[] = situacionesLas6.map((e) => ({
  ...e,
  accion: {
    verbo: e.situacion.tipo === "no-lo-tengo" ? ("traer-canonico" as const) : ("" as const),
    habilitada: false,
    motivo: NO_APLICA,
  },
}))

// catReferenciaLas6 — el catálogo read-only con las 6 ramas: E-07 asserta que NINGÚN
// Publicar/Actualizar/Reparar existe en el DOM y que Traer está `disabled` con el literal.
export const catReferenciaLas6: Catalogo = {
  marketplace: "claude-plugins-official",
  clase: "referencia",
  repo: "github.com/anthropics/claude-plugins-official",
  owner_nombre: "Anthropic",
  owner_email: "support@anthropic.com",
  lectura: { tipo: "leido", cuando: "2026-07-23T09:12:00Z", entradas: 273, fuente: "local" },
  entradas: situacionesLas6Referencia,
}

// catCruceDebil — E-66/E-38: los dos cruces DÉBILES de identidad, cada uno declarado como tal.
// `faceta-registry` = la identidad de la copia es provisional pero su `registry` apunta acá (el
// caso REAL dominante). `rename` = el catálogo renombró la entrada y el Portafolio la tiene
// keyeada con el nombre viejo (`renames` real del oficial: `convex-backend` → `convex`).
export const catCruceDebil: Catalogo = {
  marketplace: "prenter-marketplace",
  clase: "propio",
  repo: "github.com/alpacapurpura/prenter-marketplace",
  lectura: { tipo: "leido", cuando: "2026-07-25T14:07:33Z", entradas: 2, fuente: "local" },
  entradas: [
    {
      nombre: "harness",
      source: {
        tipo: "ruta-relativa",
        crudo: "./plugins/harness/0.5.3",
        ruta: "./plugins/harness/0.5.3",
      },
      version: "0.5.3",
      version_de: "derivada-de-source",
      situacion: {
        tipo: "estante-adelantado",
        mia: "0.5.2",
        estante: "0.5.3",
        clave_portafolio: "sin-home~harness~github-com-alpacapurpura-luana-vitalia",
        via: "faceta-registry",
      },
      accion: { verbo: "actualizar-mi-copia", habilitada: false, motivo: MOTIVO_ACTUALIZAR },
    },
    {
      nombre: "convex",
      source: {
        tipo: "url",
        crudo: "url:https://github.com/get-convex/convex-claude-plugin.git#@",
        url: "https://github.com/get-convex/convex-claude-plugin.git",
        sha: "0d17e3a5c2b1f4e8a9c7d6b5a4938271605f4e3d",
      },
      nombre_anterior: ["convex-backend"],
      situacion: {
        tipo: "no-comparable",
        motivo: MOTIVO_NO_COMPARABLE,
        clave_portafolio: "github-com-get-convex-convex~convex-backend~",
        via: "rename",
      },
      accion: { verbo: "", habilitada: false, motivo: MOTIVO_NO_COMPARABLE },
    },
  ],
}

// catTruncado — E-58: el backend recortó por techo de seguridad y lo DECLARA. `truncado: 1000`
// es un dato visible, jamás un recorte silencioso.
export const catTruncado: Catalogo = {
  marketplace: "claude-plugins-official",
  clase: "referencia",
  repo: "github.com/anthropics/claude-plugins-official",
  lectura: { tipo: "leido", cuando: "2026-07-23T09:12:00Z", entradas: 5000, fuente: "remoto" },
  entradas: entradasOficialMuestra,
  truncado: 1000,
}

// catalogo273 — el catálogo de 273 filas (la cifra REAL del oficial, AG-D16) para la story de
// lazy. Se GENERA en vez de fixturearse: lo que se prueba es el conteo y el faltante declarado,
// no el contenido de cada fila (plan-pruebas E-28 lo dice explícito).
export function catalogo273(): Catalogo {
  const base = entradasOficialMuestra
  const entradas: EntradaCatalogo[] = Array.from({ length: 273 }, (_, i) => {
    const plantilla = base[i % base.length] as EntradaCatalogo
    return { ...plantilla, nombre: `${plantilla.nombre}-${i}` }
  })
  return {
    marketplace: "claude-plugins-official",
    clase: "referencia",
    repo: "github.com/anthropics/claude-plugins-official",
    owner_nombre: "Anthropic",
    lectura: { tipo: "leido", cuando: "2026-07-23T09:12:00Z", entradas: 273, fuente: "local" },
    entradas,
  }
}

// ── S6: validación (BR-5/G3) ────────────────────────────────────────────────────────────────

// validacionPrenter — lo que `POST /api/marketplaces/validaciones` devuelve para el repo REAL:
// `name` + `owner` + la cuenta REAL de `plugins[]` (2) LEÍDOS del archivo, `fuente: "local"`
// porque CC ya tiene ese marketplace clonado (cero red). Sin un 200 así no hay ✓ posible.
export const validacionPrenter = {
  url_canonica: "github.com/alpacapurpura/prenter-marketplace",
  nombre: "prenter-marketplace",
  owner_nombre: "Prenter",
  owner_email: "hola@alpacapurpura.lat",
  descripcion:
    "Private Prenter marketplace (instancia 0 — KIT-06/I-72). Access = your git credentials " +
    "(client code). Channels are index entries; catalogo.json is the version-state SSoT.",
  entradas: 2,
  fuente: "local",
  ya_registrado: false,
}

// validacionCaveman — el otro shape REAL de `owner` (AG-D15): `{name, url}` sin email. Ningún
// campo se inventa: si no hay email, no se muestra un email.
export const validacionCaveman = {
  url_canonica: "github.com/juliusbrussee/caveman",
  nombre: "caveman",
  owner_nombre: "Julius Brussee",
  owner_url: "https://github.com/JuliusBrussee",
  descripcion:
    "Ultra-compressed communication mode for Claude Code. Cuts ~75% of tokens while keeping " +
    "full technical accuracy.",
  entradas: 1,
  fuente: "remoto",
  ya_registrado: false,
}

// ── S7: candidatos de origen ────────────────────────────────────────────────────────────────

// candidatosOrigenDemo — el orden lo decide el BACKEND por señales blandas (§3.5) y el widget NO
// reordena. `senal` se arma en el usecase para que sea testeable. Literales del mockup firmado.
export const candidatosOrigenDemo = [
  {
    nombre: "vitalia-arneses",
    repo: "github.com/vitalia/arneses",
    clase: "propio" as const,
    senal: "el registry de tu copia ya apunta acá",
  },
  {
    nombre: "prenter-marketplace",
    repo: "github.com/alpacapurpura/prenter-marketplace",
    clase: "propio" as const,
    senal: "propio",
  },
  {
    nombre: "nordia-plugins-rrhh",
    repo: "github.com/nordia/plugins-rrhh",
    clase: "propio" as const,
    senal: "propio · no leído aún",
  },
  {
    nombre: "claude-plugins-official",
    repo: "github.com/anthropics/claude-plugins-official",
    clase: "referencia" as const,
    senal: "de referencia · resuelve procedencia de 1 arnés tuyo",
  },
]
