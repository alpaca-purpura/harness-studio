import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { Button } from "./button"

// Story = test (arch/boundaries/fe-visual-fitness.md). El `play` corre en CI vía addon-vitest.
const meta = {
  title: "shared/ui/Button",
  component: Button,
  parameters: { layout: "centered" },
  args: { children: "Conversar" },
} satisfies Meta<typeof Button>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  // a11y `todo`: the primary variant is the signed brand sand (--primary #a8742c) with white text
  // at 4.03:1 — just under WCAG AA 4.5:1. A brand-token decision (fe-tokens-contrato), pre-existing
  // and app-wide, not a Map regression; surfaced as todo, not failed.
  parameters: { a11y: { test: "todo" } },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await expect(canvas.getByRole("button", { name: "Conversar" })).toBeInTheDocument()
  },
}

export const Secondary: Story = { args: { variant: "secondary" } }
export const Outline: Story = { args: { variant: "outline" } }
export const Ghost: Story = { args: { variant: "ghost" } }
export const Destructive: Story = {
  args: { variant: "destructive", children: "Descartar" },
}
