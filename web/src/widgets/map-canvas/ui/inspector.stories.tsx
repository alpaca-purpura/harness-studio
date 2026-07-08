import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import type { Box, Graph } from "@/entities/arnes"
import { Inspector } from "./inspector"

// Inspector is a right-docked, full-height overlay → give it a positioned, sized frame.
function Frame({ children }: { children: ReactNode }) {
  return <div style={{ position: "relative", width: 360, height: 560 }}>{children}</div>
}

// A caja with a full fused contract (the inspector's rich case, S3).
const cajaBox: Box = {
  id: "spec-writer",
  clase: "skill",
  nombre: "escribir el spec",
  banda: "fase",
  fase: "spec",
  estado: "idea -> spec",
  canal: "beta",
  procedencia: "declarado",
  contract: {
    why: "convertir una idea conversada en un spec ejecutable que blinde la deriva",
    capabilities: [
      {
        id: "CAP-01",
        what: "destilar la idea en capacidades",
        success: "cada capability verificable",
      },
    ],
    arquetipo: "excepcion",
    perfil_harness: "T2",
    caja: true,
    necesita: [{ art: "idea del usuario", de: "usuario", requerido: true }],
    entrega: [{ art: "spec.md", escritor_unico: true }],
    ruta: [{ a: "builder", si: "gate del spec verde" }],
    gate: { tipo: "manual", detalle: "revisión humana del spec" },
    handoff: { cuando: "no converge en 3 vueltas", a: "humano" },
  },
}

const meta = {
  title: "widgets/map-canvas/Inspector",
  component: Inspector,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { box: cajaBox, onClose: fn() },
} satisfies Meta<typeof Inspector>

export default meta
type Story = StoryObj<typeof meta>

export const Caja: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    await expect(c.getByText(/convertir una idea conversada/)).toBeInTheDocument()
    await expect(c.getByText("spec.md")).toBeInTheDocument()
    // Close is wired.
    await c.getByRole("button", { name: "Cerrar inspector" }).click()
    await expect(args.onClose).toHaveBeenCalled()
  },
}

// RF-80/81 — ⤢ expande (header/tabs sticky, cubre el mapa), Esc COLAPSA, ✕ CIERRA
// incluso expandido (cerrar ≠ colapsar, decisión #5e).
export const ExpandeColapsaCierra: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const aside = canvasElement.querySelector("aside.arnesia-inspector") as HTMLElement
    const expand = c.getByRole("button", { name: "Ampliar inspector" })
    await expect(expand).toHaveAttribute("aria-pressed", "false")
    await expand.click()
    await expect(aside).toHaveClass("expanded")
    const collapse = c.getByRole("button", { name: "Colapsar al drawer normal" })
    await expect(collapse).toHaveAttribute("aria-pressed", "true")
    // Esc colapsa al drawer normal (no cierra).
    await userEvent.keyboard("{Escape}")
    await expect(aside).not.toHaveClass("expanded")
    await expect(args.onClose).not.toHaveBeenCalled()
    // ✕ cierra del todo, también estando expandido.
    await c.getByRole("button", { name: "Ampliar inspector" }).click()
    await c.getByRole("button", { name: "Cerrar inspector" }).click()
    await expect(args.onClose).toHaveBeenCalled()
  },
}

// RF-82 — tres tabs con roles ARIA; conmutan panes sin perder el header.
export const Tabs: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const tabs = c.getAllByRole("tab")
    await expect(tabs).toHaveLength(3)
    await expect(c.getByRole("tab", { name: "Resumen" })).toHaveAttribute("aria-selected", "true")
    await c.getByRole("tab", { name: "Contenido" }).click()
    await expect(c.getByText("versiona con el arnés")).toBeInTheDocument()
    await c.getByRole("tab", { name: "Corridas" }).click()
    await expect(c.getByText(/Sin corridas indexadas/)).toBeInTheDocument()
    // El header (identidad) nunca se pierde al conmutar.
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
  },
}

// RF-96 — Corridas honesta: estado + nota de caja (qué listará primero) + acción staged.
export const CorridasHonesta: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await c.getByRole("tab", { name: "Corridas" }).click()
    await expect(c.getByText(/indexer JSONL/)).toBeInTheDocument()
    await expect(c.getByText(/boxes\/spec-writer\/run/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /Ver todas las corridas/ })).toBeDisabled()
  },
}

// Grafo mínimo alrededor de cajaBox para RF-89/90: builder existe (chip navegable),
// std-spec lee (edge inverso), y la ruta apunta a builder.
const miniGraph: Graph = {
  arnes: { reporta_a: null },
  nodos: [
    cajaBox,
    { id: "builder", clase: "skill", nombre: "construir", banda: "fase", fase: "build" },
    { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" },
  ],
  edges: [
    { de: "builder", a: "spec-writer", tipo: "invoca" },
    { de: "spec-writer", a: "std-spec", tipo: "lee" },
  ],
}

// RF-89/90 — Viene de (edges inversos con tipo) + chips navegables en Necesita/Ruta:
// click en chip cuyo destino existe → onSelect; destino ausente → chip inerte rotulado.
export const VieneDeChips: Story = {
  args: { box: cajaBox, graph: miniGraph, onSelect: fn(), onClose: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    // Viene de: builder invoca a spec-writer (derivado de graph.edges).
    await expect(c.getByText("Viene de")).toBeInTheDocument()
    const inverso = c.getByRole("button", { name: /builder invoca/ })
    await inverso.click()
    await expect(args.onSelect).toHaveBeenCalledWith("builder")
    // Ruta: chip navegable a builder + condición como badge punteado.
    const rutaChip = c.getAllByRole("button", { name: /^builder$/ })[0]
    await expect(rutaChip).toBeEnabled()
    await expect(c.getByText("si gate del spec verde")).toBeInTheDocument()
    // Necesita: "usuario" no tiene id de nodo → chip inerte rotulado.
    const inerte = c.getByRole("button", { name: /idea del usuario/ })
    await expect(inerte).toBeDisabled()
    await expect(inerte).toHaveAttribute("title", "nodo fuera del grafo cargado")
  },
}

// RF-91 — hallazgo determinista gate:none (crit-soft, A4) — nunca «sin hallazgos».
export const HallazgoGateNone: Story = {
  args: {
    box: {
      ...cajaBox,
      id: "draft-caja",
      nombre: "redactar el borrador",
      contract: {
        ...cajaBox.contract,
        caja: true,
        gate: { tipo: "none", detalle: "caja generativa sin eval formal aún" },
      },
    },
    conformance: [],
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("gate:none")).toBeInTheDocument()
    await expect(c.getByText(/caja sin eval formal/)).toBeInTheDocument()
    await expect(c.queryByText("Sin hallazgos abiertos.")).not.toBeInTheDocument()
  },
}

// RF-91/92 — sin hallazgos como estado + checks rojos de conformance filtrados por nodo
// + botonera staged (disabled, rotulada — jamás finge funcionar).
export const HallazgosBotonera: Story = {
  args: {
    box: cajaBox,
    conformance: [
      {
        check: { id: "estado-en-spine-declarado", severidad: "error" },
        veredicto: "fail",
        detalle: "spec-writer (estado mal formado)",
      },
      {
        check: { id: "escritor-unico", severidad: "error" },
        veredicto: "fail",
        detalle: "otra-caja escribe spec.md",
      },
      { check: { id: "spine-auto-consistente", severidad: "error" }, veredicto: "pass" },
    ],
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Solo el check rojo que MENCIONA a este nodo (word-boundary sobre detalle).
    await expect(c.getByText("estado-en-spine-declarado")).toBeInTheDocument()
    await expect(c.queryByText("escritor-unico")).not.toBeInTheDocument()
    await expect(c.queryByText("Sin hallazgos abiertos.")).not.toBeInTheDocument()
    // Botonera staged: primaria + 3, todas disabled y rotuladas.
    await expect(c.getByRole("button", { name: "Editar conversando" })).toBeDisabled()
    await expect(c.getByRole("button", { name: "Promover a estable" })).toBeDisabled()
    await expect(c.getByText(/se cablea en Fase 3\/4/)).toBeInTheDocument()
  },
}

// RF-93/95 — tab Contenido con la fuente REAL: viewer con números de línea, chip
// «versiona con el arnés», acciones staged disabled. loadFuente lo inyecta la página.
export const ContenidoFuenteReal: Story = {
  args: {
    box: { ...cajaBox, fuente_path: "skills/spec-writer/SKILL.md" },
    loadFuente: fn(async () => "---\nname: spec-writer\n---\n# spec-writer"),
    onClose: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await c.getByRole("tab", { name: "Contenido" }).click()
    await expect(args.loadFuente).toHaveBeenCalledWith("spec-writer")
    // Viewer con números de línea y el texto TAL CUAL.
    await expect(await c.findByText("# spec-writer")).toBeInTheDocument()
    await expect(c.getByText("name: spec-writer")).toBeInTheDocument()
    await expect(c.getByText("versiona con el arnés")).toBeInTheDocument()
    // Acciones staged: disabled y rotuladas (RF-95).
    await expect(c.getByRole("button", { name: "Editar fuente" })).toBeDisabled()
    await expect(c.getByRole("button", { name: "Editar conversando (dock)" })).toBeDisabled()
  },
}

// RF-93 (Gherkin «nodo sin fuente») — sin fuente_path la tab dice el estado honesto.
export const ContenidoSinFuente: Story = {
  args: {
    box: { id: "pii-guard", clase: "hook", nombre: "guardia PII", banda: "guardia" },
    loadFuente: fn(async () => ""),
    onClose: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await c.getByRole("tab", { name: "Contenido" }).click()
    // «pendiente del reconocedor» también vive en el Rol del Resumen (pane oculto) → getAll.
    await expect(c.getAllByText(/pendiente del reconocedor/).length).toBeGreaterThan(0)
    await expect(c.getByText(/el loader aún no estampa/)).toBeInTheDocument()
    await expect(args.loadFuente).not.toHaveBeenCalled()
  },
}

// RF-93 — el daemon niega la lectura (404 honesto: arnés sin dir registrado, archivo
// ausente…): la tab muestra el motivo, jamás inventa contenido.
export const ContenidoErrorHonesto: Story = {
  args: {
    box: { ...cajaBox, fuente_path: "skills/spec-writer/SKILL.md" },
    loadFuente: fn(async () => {
      throw new Error("404 el arnés no tiene directorio registrado — carga la carpeta")
    }),
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await c.getByRole("tab", { name: "Contenido" }).click()
    await expect(await c.findByText(/No se pudo leer la fuente/)).toBeInTheDocument()
    await expect(c.getByText(/no tiene directorio registrado/)).toBeInTheDocument()
  },
}

// RF-85/86/88 — tooltips doctrinales: «i» por sección (qué agrupa) + campo punteado
// (definición del campo Y del valor concreto), ambos operables por teclado.
export const TooltipsDoctrinales: Story = {
  args: {
    box: {
      ...cajaBox,
      procedencia: "estimado",
    },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // «i» de sección: botón focusable con el texto del diccionario SEC_TIP (RF-85).
    const info = c.getByRole("button", { name: "Qué agrupa Clasificación" })
    await expect(info).toHaveAttribute("data-tip", expect.stringContaining("clase ⊥ arquetipo"))
    // Campo con valor explicado (RF-86, Gherkin): procedencia "estimado" → atenuado, gris ≠ verde.
    const kProc = c.getByText("procedencia", { selector: ".k" })
    await expect(kProc).toHaveClass("tip")
    await expect(kProc).toHaveAttribute("tabindex", "0")
    await expect(kProc).toHaveAttribute("data-tip", expect.stringContaining("se dibuja atenuado"))
    // Foco con teclado dispara el tooltip (:focus-visible) — el ancla es tabulable.
    kProc.focus()
    await expect(kProc).toHaveFocus()
  },
}

// A contract-less rule → per-class framing (inspector-por-clase.md Tier A): doctrinal role,
// Activación (unknown here — not in the proposal sets) and Fuente, never a generic "lacks".
export const Regla: Story = {
  args: {
    box: {
      id: "std-spec",
      clase: "rule",
      nombre: "estándar de spec",
      banda: "base",
      fuente_path: "dogfood/dev-full-cycle/CLAUDE.md",
    },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Regla de la Base/)).toBeInTheDocument()
    await expect(c.getByText("desconocida")).toBeInTheDocument()
    // fuente_path aparece en Resumen y en la tab Contenido (pane oculto) → getAll.
    await expect(c.getAllByText("dogfood/dev-full-cycle/CLAUDE.md")[0]).toBeInTheDocument()
  },
}

// A conditional rule (PROPOSED_CONDITIONAL) → the header handle AND Activación say so.
export const ReglaCondicional: Story = {
  args: {
    box: { id: "code-style", clase: "rule", nombre: "estilo de código", banda: "base" },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/condicional \(paths:\)/)).toBeInTheDocument()
    // The header handle span reads exactly "condicional" (alw=false → handleFor).
    await expect(c.getByText("condicional")).toBeInTheDocument()
  },
}

// A hook → Guardia framing + the pending per-class fields named honestly (Tier B).
export const Hook: Story = {
  args: {
    box: { id: "pii-guard", clase: "hook", nombre: "guardia PII", banda: "guardia" },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Hook de la Guardia/)).toBeInTheDocument()
    await expect(c.getByText(/evento · matcher/)).toBeInTheDocument()
  },
}

// A no-reconocido node (D-c) → the inspector renders (no crash) with the warn framing.
export const NoReconocido: Story = {
  args: {
    box: {
      id: "misterio",
      clase: "no-reconocido",
      nombre: "misterio (no reconocido)",
      fuente_path: "dogfood/dev-full-cycle/skills/misterio",
    },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/reconocedor no entendió/)).toBeInTheDocument()
    // fuente_path aparece en Resumen y en la tab Contenido (pane oculto) → getAll.
    await expect(c.getAllByText("dogfood/dev-full-cycle/skills/misterio")[0]).toBeInTheDocument()
  },
}
