import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import { Region } from "./region"

function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-map">{children}</div>
}

// Story = test (fe-visual-fitness). The three tinted territories (design §3.1).
const meta = {
  title: "widgets/map-canvas/Region",
  component: Region,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    kind: "proceso",
    title: "Proceso · carriles por fase",
    children: <div style={{ padding: 8 }}>contenido</div>,
  },
} satisfies Meta<typeof Region>

export default meta
type Story = StoryObj<typeof meta>

export const Proceso: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("Proceso · carriles por fase")).toBeInTheDocument()
  },
}
export const Guardia: Story = {
  args: { kind: "guardia", title: "Guardia · hooks transversales" },
}
export const Base: Story = {
  args: { kind: "soporte", title: "Base · conocimiento, reglas y soporte del arnés" },
}
