import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { Lane } from "./lane"

const meta = {
  title: "widgets/map-canvas/Lane",
  component: Lane,
  parameters: { layout: "padded" },
  args: {
    fase: "spec",
    nodes: [
      {
        id: "spec-writer",
        clase: "skill",
        nombre: "escribir el spec",
        banda: "fase",
        fase: "spec",
      },
    ],
  },
} satisfies Meta<typeof Lane>

export default meta
type Story = StoryObj<typeof meta>

export const Spec: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    await expect(c.getByText("1")).toBeInTheDocument()
  },
}

export const Empty: Story = {
  args: { fase: "release", nodes: [] },
}
