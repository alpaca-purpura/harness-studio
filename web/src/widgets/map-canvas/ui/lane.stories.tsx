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
