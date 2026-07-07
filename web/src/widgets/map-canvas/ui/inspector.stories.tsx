import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, within } from "storybook/test"
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
    await expect(c.getByText("dogfood/dev-full-cycle/CLAUDE.md")).toBeInTheDocument()
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
    await expect(c.getByText("dogfood/dev-full-cycle/skills/misterio")).toBeInTheDocument()
  },
}
