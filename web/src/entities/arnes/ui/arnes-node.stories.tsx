import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import type { Box } from "../model/types"
import { ArnesNode } from "./arnes-node"

// Story = test (fe-visual-fitness). The node styles live in map.css scoped under
// `.arnesia-map`, so every story is wrapped in that root + a 232px lane-cell to see the real
// look. Variants mirror design.md §5-6 (caja · support · compact · puesto · propuesto · dim).
function Cell({ children }: { children: ReactNode }) {
  return (
    <div className="arnesia-map">
      <div style={{ width: 232 }}>{children}</div>
    </div>
  )
}

const cajaBox: Box = {
  id: "spec-writer",
  clase: "skill",
  nombre: "escribir el spec",
  banda: "fase",
  fase: "spec",
  estado: "idea -> spec",
  canal: "beta",
  contract: { caja: true },
}

const meta = {
  title: "entities/arnes/ArnesNode",
  component: ArnesNode,
  // a11y `todo`: the node carries intentional low-contrast micro-UI of the signed design — the
  // dim state (opacity .4, RF-33), the caja/propuesto badges and the transition tag (9px, tinted
  // over color-mix). These are contract, not regressions; surfaced as todo. The DOM `play`
  // assertions still gate CI.
  parameters: { a11y: { test: "todo" } },
  decorators: [(Story) => <Cell>{Story()}</Cell>],
  args: {
    box: { id: "reviewer", clase: "skill", nombre: "revisar el build", banda: "fase" },
  },
} satisfies Meta<typeof ArnesNode>

export default meta
type Story = StoryObj<typeof meta>

export const Skill: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("revisar el build")).toBeInTheDocument()
    // The type is carried by the derived handle (/id), not a text label.
    await expect(within(canvasElement).getByText("/reviewer")).toBeInTheDocument()
  },
}

export const Caja: Story = {
  args: { box: cajaBox },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("caja")).toBeInTheDocument()
    // Transition normalized from the real estado "idea -> spec".
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
  },
}

export const Rule: Story = {
  args: { box: { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" } },
  play: async ({ canvasElement }) => {
    // rule with unknown alw → always-on handle (mockup faithful).
    await expect(within(canvasElement).getByText("always-on")).toBeInTheDocument()
  },
}

export const Support: Story = {
  args: {
    box: { id: "test-author", clase: "subagent", nombre: "escribir pruebas", banda: "fase" },
    support: true,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("@test-author")).toBeInTheDocument()
  },
}

export const Compact: Story = {
  args: {
    box: { id: "guard-format", clase: "hook", nombre: "formato pre-write", banda: "guardia" },
    compact: true,
  },
}

export const Puesto: Story = {
  args: {
    box: { id: "domain-glossary", clase: "rule", nombre: "glosario de dominio", banda: "base" },
  },
}

export const Propuesto: Story = {
  args: {
    box: {
      id: "indice-semantico",
      clase: "mcp",
      nombre: "índice semántico",
      banda: "base",
      canal: "propuesto",
    },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("propuesto")).toBeInTheDocument()
  },
}

export const Dim: Story = { args: { dim: true } }
export const Selected: Story = { args: { selected: true } }
