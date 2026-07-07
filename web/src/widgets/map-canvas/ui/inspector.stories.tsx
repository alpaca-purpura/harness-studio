import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import type { Box } from "@/entities/arnes"
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

// RF-84 — estado vacío: línea de affordance, no un panel en blanco ni ausencia.
export const Vacio: Story = {
  args: { box: undefined, onClose: fn() },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Clic en un nodo del mapa/)).toBeInTheDocument()
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
