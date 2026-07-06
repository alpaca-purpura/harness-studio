import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { Band } from "./band"

const meta = {
  title: "widgets/map-canvas/Band",
  component: Band,
  parameters: { layout: "padded" },
  args: {
    label: "Base",
    sublabel: "siempre en contexto — reglas · knowledge · mcp",
    nodes: [{ id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" }],
  },
} satisfies Meta<typeof Band>

export default meta
type Story = StoryObj<typeof meta>

export const Base: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("estándar de spec")).toBeInTheDocument()
  },
}

export const Empty: Story = {
  args: { label: "Guardia", sublabel: "hooks transversales", nodes: [] },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("— sin nodos —")).toBeInTheDocument()
  },
}
