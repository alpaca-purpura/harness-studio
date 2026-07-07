import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, userEvent, within } from "storybook/test"
import { luanaFeatureCycle } from "@/entities/arnes"
import { BaseBand } from "./base-band"

function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-map">{children}</div>
}

// The real Luana base nodes: 10 rules (5 always-on + 5 conditional) + 3 mcp knowledge services.
const baseNodes = luanaFeatureCycle.nodos.filter((n) => n.banda === "base")

// Story = test (fe-visual-fitness). The collapsible Reglas split (shot4/shot5) + static Knowledge.
const meta = {
  title: "widgets/map-canvas/BaseBand",
  component: BaseBand,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { nodes: baseNodes },
} satisfies Meta<typeof BaseBand>

export default meta
type Story = StoryObj<typeof meta>

// Collapsed by default (shot4): header shows the counts, body hidden.
export const Colapsada: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("5 siempre")).toBeInTheDocument()
    await expect(c.getByText("5 condicional")).toBeInTheDocument()
    await expect(c.getByText("leído por skills")).toBeInTheDocument()
    // The Reglas button is collapsed (aria-expanded=false).
    await expect(c.getByRole("button", { name: /Reglas/ })).toHaveAttribute(
      "aria-expanded",
      "false",
    )
  },
}

// Expanding shows the two rule groups (shot5).
export const Expandida: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: /Reglas/ }))
    await expect(c.getByText("siempre en contexto (CLAUDE.md)")).toBeVisible()
    await expect(c.getByText("carga condicional (paths:)")).toBeVisible()
  },
}
