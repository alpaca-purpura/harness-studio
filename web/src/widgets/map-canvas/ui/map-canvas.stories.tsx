import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { expect, within } from "storybook/test"
import { devFullCycle, luanaFeatureCycle } from "@/entities/arnes"
import { MapCanvas } from "./map-canvas"

// Story = test: the full map surface. This is the fitness fixture of record — the signed mockup
// is derived from what renders here. a11y `todo`: intentional low-contrast micro-labels (badges/
// chips of the signed palette); the DOM assertions still gate CI.
const meta = {
  title: "widgets/map-canvas/MapCanvas",
  component: MapCanvas,
  parameters: { layout: "fullscreen", a11y: { test: "todo" } },
  decorators: [
    (Story) => (
      <div style={{ height: 620 }}>
        <Story />
      </div>
    ),
  ],
  args: { graph: devFullCycle },
} satisfies Meta<typeof MapCanvas>

export default meta
type Story = StoryObj<typeof meta>

// The dogfood arnés (dev-full-cycle) — what the real mount serves (shot7).
export const Dogfood: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Phase cajas render with their transition.
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    await expect(c.getByText("construir contra el spec")).toBeInTheDocument()
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
    // Base rule renders (inside the collapsed Reglas subband — present, just hidden).
    await expect(c.getByText("estándar de spec")).toBeInTheDocument()
    // Guardia is empty for this arnés.
    await expect(c.getByText("— sin hooks —")).toBeInTheDocument()
  },
}

function Selectable() {
  const [sel, setSel] = useState<string>()
  return <MapCanvas graph={devFullCycle} selectedId={sel} onSelect={setSel} />
}
export const WithSelection: Story = { render: () => <Selectable /> }

// A COMPLETE arnés (Luana) with the example/picker toggle — the dense parity view (shot1).
export const LuanaCompleto: Story = {
  args: {
    graph: luanaFeatureCycle,
    harnesses: [
      { id: "luana-feature-cycle", label: "Luana" },
      { id: "dev-full-cycle", label: "dev-full-cycle" },
    ],
    activeId: "luana-feature-cycle",
  },
  decorators: [
    (Story) => (
      <div style={{ height: 820 }}>
        <Story />
      </div>
    ),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("destilar necesidad")).toBeInTheDocument()
    await expect(c.getByText("guardia PII")).toBeInTheDocument()
    await expect(c.getByText("acceso a la API")).toBeInTheDocument()
    await expect(c.getByText("kit luana")).toBeInTheDocument()
    // The active picker button is pressed (RF-55).
    await expect(c.getByRole("button", { name: "Luana" })).toHaveAttribute("aria-pressed", "true")
  },
}
