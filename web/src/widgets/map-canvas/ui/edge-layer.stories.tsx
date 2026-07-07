import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect } from "storybook/test"
import type { EdgePath } from "../model/use-edge-paths"
import { EdgeLayer } from "./edge-layer"

// The svg is absolute inset-0, so it needs a positioned, sized parent (like `.content`).
function Frame({ children }: { children: ReactNode }) {
  return (
    <div className="arnesia-map">
      <div style={{ position: "relative", width: 320, height: 160 }}>{children}</div>
    </div>
  )
}

// One of each relation type (design §9.1): invoca (crit, arrow) · escribe (ok, dash, arrow) ·
// lee (warn, dash, no arrow).
const paths: EdgePath[] = [
  {
    key: "a->b",
    d: "M 10,30 C 90,30 150,30 230,30",
    stroke: "var(--crit)",
    opacity: 0.75,
    width: 1.6,
    marker: true,
  },
  {
    key: "a->c",
    d: "M 10,80 C 90,80 150,80 230,80",
    stroke: "var(--ok)",
    opacity: 0.7,
    width: 1.6,
    dash: "3 3",
    marker: true,
  },
  {
    key: "a->d",
    d: "M 10,130 C 90,130 150,130 230,130",
    stroke: "var(--warn)",
    opacity: 0.65,
    width: 1.6,
    dash: "4 4",
    marker: false,
  },
]

const meta = {
  title: "widgets/map-canvas/EdgeLayer",
  component: EdgeLayer,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { paths },
} satisfies Meta<typeof EdgeLayer>

export default meta
type Story = StoryObj<typeof meta>

export const ThreeTypes: Story = {
  play: async ({ canvasElement }) => {
    // Three connectors drawn; lee (warn) carries no arrow marker.
    const svgPaths = canvasElement.querySelectorAll("svg.edges path[stroke]")
    await expect(svgPaths.length).toBe(3)
    const lee = canvasElement.querySelector("svg.edges path[stroke='var(--warn)']")
    await expect(lee).not.toBeNull()
    await expect(lee?.getAttribute("marker-end")).toBeNull()
  },
}
