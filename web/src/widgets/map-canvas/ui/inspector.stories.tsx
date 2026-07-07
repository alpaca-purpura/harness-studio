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

// A non-caja node has no contract → honest empty state.
export const SinContrato: Story = {
  args: {
    box: { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" },
    onClose: fn(),
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText(/Sin contrato/)).toBeInTheDocument()
  },
}
