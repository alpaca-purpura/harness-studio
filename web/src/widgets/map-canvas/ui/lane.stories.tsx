import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import type { Box } from "@/entities/arnes"
import { Lane } from "./lane"

// Story = test (fe-visual-fitness). Lane styles live in map.css scoped under `.arnesia-map`.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-map">{children}</div>
}

// A caja (skill-front, contract.caja) + a subagent apoyo → caja-first order + hairline + support.
const specNodes: Box[] = [
  {
    id: "spec-writer",
    clase: "skill",
    nombre: "escribir el spec",
    banda: "fase",
    fase: "spec",
    estado: "idea -> spec",
    contract: { caja: true },
  },
  { id: "spec-review", clase: "command", nombre: "/spec-review", banda: "fase", fase: "spec" },
]

const meta = {
  title: "widgets/map-canvas/Lane",
  component: Lane,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { fase: "spec", nodes: specNodes },
} satisfies Meta<typeof Lane>

export default meta
type Story = StoryObj<typeof meta>

export const Spec: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    // caja badge + transition on the caja node; the count pill shows 2.
    await expect(c.getByText("caja")).toBeInTheDocument()
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
    await expect(c.getByText("2")).toBeInTheDocument()
  },
}

export const Empty: Story = { args: { fase: "release", nodes: [] } }

// ══ Capa «Mejora» (paquete 2026-07-24, T31) ═══════════════════════════════════════════════

// RF-244 · J-8 — el `lane-hd` pasa a TRES hijos y **el `.count` se conserva**: el conteo de
// nodos y el gasto de la fase son dos cosas distintas y las dos se leen. Sustituir uno por el
// otro habría sido perder información firmada para ganar 40 px.
export const LaneConTotalMejora: Story = {
  args: { totalUsd: "1,92" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const hd = canvasElement.querySelector(".lane-hd") as HTMLElement
    await expect(hd.children).toHaveLength(3)
    await expect(hd.querySelector("h3")?.textContent).toBe("spec")
    await expect(hd.querySelector(".count")?.textContent).toBe("2")
    await expect(c.getByText("1,92")).toBeInTheDocument()
    await expect(c.getByText("USD")).toBeInTheDocument()
  },
}

// RF-244 — un carril cuyas cajas no tienen costo atribuido dice `sin dato`. **`USD 0,00`
// afirmaría que la fase corrió y no gastó**, que es otra cosa completamente distinta.
export const LaneSinDato: Story = {
  args: { totalUsd: null },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin dato")).toBeInTheDocument()
    await expect(c.queryByText("USD 0,00")).toBeNull()
    await expect(c.queryByText("0,00")).toBeNull()
    await expect(canvasElement.querySelector(".lane-hd .count")).not.toBeNull()
  },
}
