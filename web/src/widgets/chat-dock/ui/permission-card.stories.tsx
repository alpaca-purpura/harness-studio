import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { PermissionCard } from "./permission-card"

// Story = test (fe-visual-fitness) — RF-113: la tarjeta de permiso del Dock pinta el
// input crudo del control_request por herramienta (Edit=diff · Write=contenido ·
// Bash=comando · resto=JSON) y expone las 3 decisiones.
const meta = {
  title: "widgets/chat-dock/PermissionCard",
  component: PermissionCard,
  parameters: { layout: "padded", a11y: { test: "todo" } },
  decorators: [(Story) => <div style={{ maxWidth: 420 }}>{Story()}</div>],
  args: { onResolve: fn() },
} satisfies Meta<typeof PermissionCard>

export default meta
type Story = StoryObj<typeof meta>

export const EditConDiff: Story = {
  args: {
    ask: {
      request_id: "cr-1",
      tool: "Edit",
      input: {
        file_path: "skills/builder/SKILL.md",
        old_string: "  gate:\n    tipo: manual",
        new_string: "  gate:\n    tipo: auto\n    eval: pnpm test --run",
      },
    },
  },
  play: async ({ args, canvasElement }) => {
    const c = within(canvasElement)
    // El nombre del tool aparece en cabecera Y en la explicación del grant — ambas a propósito.
    await expect(c.getAllByText("Edit").length).toBeGreaterThanOrEqual(2)
    await expect(c.getByText("skills/builder/SKILL.md")).toBeInTheDocument()
    // El diff separa − y +: la línea vieja y la nueva son visibles a la vez.
    await expect(c.getByText(/tipo: manual/)).toBeInTheDocument()
    await expect(c.getByText(/tipo: auto/)).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Permitir esta sesión" }))
    await expect(args.onResolve).toHaveBeenCalledWith("allow")
    await userEvent.click(c.getByRole("button", { name: "Permitir una vez" }))
    await expect(args.onResolve).toHaveBeenCalledWith("allow", true)
    await userEvent.click(c.getByRole("button", { name: "Denegar" }))
    await expect(args.onResolve).toHaveBeenCalledWith("deny")
  },
}

export const WriteConContenido: Story = {
  args: {
    ask: {
      request_id: "cr-2",
      tool: "Write",
      input: { file_path: "skills/nueva-caja/SKILL.md", content: "---\nnombre: nueva-caja\n---" },
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("skills/nueva-caja/SKILL.md")).toBeInTheDocument()
    await expect(c.getByText("archivo completo")).toBeInTheDocument()
    await expect(c.getByText(/nombre: nueva-caja/)).toBeInTheDocument()
  },
}

export const BashComando: Story = {
  args: {
    ask: { request_id: "cr-3", tool: "Bash", input: { command: "pnpm test --run" } },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByText("Bash").length).toBeGreaterThanOrEqual(2)
    await expect(c.getByText(/pnpm test --run/)).toBeInTheDocument()
  },
}

export const ToolDesconocidoJsonCrudo: Story = {
  args: {
    ask: { request_id: "cr-4", tool: "WebFetch", input: { url: "https://example.com" } },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Shape no reconocido ⇒ JSON crudo, jamás una vista inventada (honesto).
    await expect(c.getByText(/"url"/)).toBeInTheDocument()
  },
}
