import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import { Band } from "./band"

// Story = test (fe-visual-fitness). Band styles live in map.css scoped under `.arnesia-map`.
// a11y `todo`: the signed design uses intentional low-contrast micro-labels (9px chips); DOM
// assertions still gate CI.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-map">{children}</div>
}

const meta = {
  title: "widgets/map-canvas/Band",
  component: Band,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    band: { id: "libreria-expertos", label: "Librería de expertos", act: "demanda" },
    nodes: [
      { id: "react-expert", clase: "skill", nombre: "experto React", banda: "libreria-expertos" },
      {
        id: "a11y-expert",
        clase: "skill",
        nombre: "experto accesibilidad",
        banda: "libreria-expertos",
      },
    ],
  },
} satisfies Meta<typeof Band>

export default meta
type Story = StoryObj<typeof meta>

export const Libreria: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("experto React")).toBeInTheDocument()
    await expect(c.getByText("bajo demanda")).toBeInTheDocument()
  },
}

export const Dormida: Story = {
  args: {
    band: { id: "marcas-dormidas", label: "Marcas dormidas", act: "dormida" },
    nodes: [
      { id: "legacy-mcp", clase: "mcp", nombre: "servicio dormido", banda: "marcas-dormidas" },
    ],
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("dormida · 0 corridas")).toBeInTheDocument()
  },
}
