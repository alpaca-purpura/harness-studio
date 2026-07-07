import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { Glyph } from "./glyph"

// Story = test (arch/boundaries/fe-visual-fitness). Every shape rendered for visual regression.
const meta = {
  title: "shared/canvas/Glyph",
  component: Glyph,
  parameters: { layout: "centered" },
  args: { color: "var(--c-skill)", char: "S", shape: "square", size: 22 },
} satisfies Meta<typeof Glyph>

export default meta
type Story = StoryObj<typeof meta>

export const Square: Story = {
  play: async ({ canvasElement }) => {
    // The glyph is aria-hidden; assert it rendered its char.
    await expect(within(canvasElement).getByText("S")).toBeInTheDocument()
  },
}
export const Circle: Story = { args: { shape: "circle", char: "A", color: "var(--c-agent)" } }
export const Diamond: Story = { args: { shape: "diamond", char: "H", color: "var(--c-hook)" } }
export const Hexagon: Story = { args: { shape: "hexagon", char: "M", color: "var(--c-mcp)" } }
export const Shield: Story = { args: { shape: "shield", char: "R", color: "var(--c-rule)" } }
// The five config classes share the `rounded` shape; the char + own --c-* color carry the
// sub-distinction (gap 6→10 closed by tokens:build, PR-B).
export const Command: Story = { args: { shape: "rounded", char: "/", color: "var(--c-command)" } }
export const Plugin: Story = { args: { shape: "rounded", char: "P", color: "var(--c-plugin)" } }
export const Settings: Story = { args: { shape: "rounded", char: "⚙", color: "var(--c-settings)" } }
export const OutputStyle: Story = {
  args: { shape: "rounded", char: "◐", color: "var(--c-output-style)" },
}
export const Statusline: Story = {
  args: { shape: "rounded", char: "▭", color: "var(--c-statusline)" },
}
