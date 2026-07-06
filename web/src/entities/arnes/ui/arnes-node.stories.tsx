import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { ArnesNode } from "./arnes-node"

const meta = {
  title: "entities/arnes/ArnesNode",
  component: ArnesNode,
  parameters: { layout: "centered" },
  args: {
    box: { id: "spec-writer", clase: "skill", nombre: "escribir el spec", banda: "fase" },
  },
} satisfies Meta<typeof ArnesNode>

export default meta
type Story = StoryObj<typeof meta>

export const Skill: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("escribir el spec")).toBeInTheDocument()
    await expect(within(canvasElement).getByText("Skill")).toBeInTheDocument()
  },
}

export const Rule: Story = {
  args: { box: { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" } },
}

export const Selected: Story = { args: { selected: true } }
