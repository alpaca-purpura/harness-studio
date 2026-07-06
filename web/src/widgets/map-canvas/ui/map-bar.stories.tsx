import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import { devFullCycle } from "@/entities/arnes"
import { MapBar } from "./map-bar"

const meta = {
  title: "widgets/map-canvas/MapBar",
  component: MapBar,
  parameters: { layout: "fullscreen" },
  args: { arnes: devFullCycle.arnes, capa: "estructura", onCapa: fn() },
} satisfies Meta<typeof MapBar>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("tab", { name: "Estructura" })).toHaveAttribute(
      "aria-selected",
      "true",
    )
    // Layers needing telemetry are disabled (honest, not hidden).
    await expect(c.getByRole("tab", { name: "Desempeño" })).toBeDisabled()
  },
}
