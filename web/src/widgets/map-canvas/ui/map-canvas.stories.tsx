import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { expect, within } from "storybook/test"
import { devFullCycle, luanaFeatureCycle } from "@/entities/arnes"
import { MapCanvas } from "./map-canvas"

// Story = test: the full map surface fed by the recorded dogfood arnés. This is the Fase B SSOT —
// the mockup is derived from what renders here.
const meta = {
  title: "widgets/map-canvas/MapCanvas",
  component: MapCanvas,
  parameters: { layout: "fullscreen" },
  decorators: [
    (Story) => (
      <div style={{ height: 580 }}>
        <Story />
      </div>
    ),
  ],
  args: { graph: devFullCycle },
} satisfies Meta<typeof MapCanvas>

export default meta
type Story = StoryObj<typeof meta>

export const Dogfood: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Phase-band boxes render.
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    await expect(c.getByText("construir contra el spec")).toBeInTheDocument()
    // Base-band rule renders.
    await expect(c.getByText("estándar de spec")).toBeInTheDocument()
    // Guardia is empty for this arnés.
    await expect(c.getByText("— sin nodos —")).toBeInTheDocument()
    // Estructura is the active (only enabled) layer.
    await expect(c.getByRole("tab", { name: "Estructura" })).toHaveAttribute(
      "aria-selected",
      "true",
    )
  },
}

function Selectable() {
  const [sel, setSel] = useState<string>()
  return <MapCanvas graph={devFullCycle} selectedId={sel} onSelect={setSel} />
}

export const WithSelection: Story = {
  render: () => <Selectable />,
}

// A COMPLETE arnés (Luana): all bands populated, the 10 clases, invoca/lee/escribe edges —
// how the map reads with full density.
export const LuanaCompleto: Story = {
  render: () => (
    <div style={{ height: 760 }}>
      <MapCanvas graph={luanaFeatureCycle} />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("destilar la necesidad")).toBeInTheDocument()
    await expect(c.getByText("guardia PII/compliance")).toBeInTheDocument()
    await expect(c.getByText("acceso a la API")).toBeInTheDocument()
    await expect(c.getByText("kit luana")).toBeInTheDocument()
  },
}
