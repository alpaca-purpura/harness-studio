import type { Meta, StoryObj } from "@storybook/react-vite";
import { expect, within } from "storybook/test";
import { Button } from "./button";

// Story = test (arch/boundaries/fe-visual-fitness.md). El `play` corre en CI vía addon-vitest.
const meta = {
  title: "shared/ui/Button",
  component: Button,
  parameters: { layout: "centered" },
  args: { children: "Conversar" },
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    await expect(canvas.getByRole("button", { name: "Conversar" })).toBeInTheDocument();
  },
};

export const Secondary: Story = { args: { variant: "secondary" } };
export const Outline: Story = { args: { variant: "outline" } };
export const Ghost: Story = { args: { variant: "ghost" } };
export const Destructive: Story = { args: { variant: "destructive", children: "Descartar" } };
