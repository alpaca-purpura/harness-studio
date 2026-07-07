import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, userEvent, within } from "storybook/test"
import { ErrorBoundary } from "./error-boundary"

// Story = test (arch/boundaries/fe-visual-fitness.md): the render-error net. A child that
// throws must surface the HONEST fallback panel (nomenclatura-arnes §4.5 — visible, never a
// white screen); children that don't throw render untouched, and «Reintentar» recovers.

const meta = {
  title: "shared/ui/ErrorBoundary",
  component: ErrorBoundary,
  parameters: { layout: "centered" },
  args: { children: null },
} satisfies Meta<typeof ErrorBoundary>

export default meta
type Story = StoryObj<typeof meta>

// Module-level (not nested in render) so Biome/React see a stable component identity.
function Bomba(): never {
  throw new Error("clase desconocida: quimera")
}

// FallaUnaVez throws only on its first render after arm() — the Reintentar story proves the
// full cycle: capture → honest panel → reset → recovered children.
let pendiente = true
function armarFallaUnaVez(): void {
  pendiente = true
}
function FallaUnaVez() {
  if (pendiente) {
    pendiente = false
    throw new Error("fallo transitorio")
  }
  return <div>contenido recuperado</div>
}

export const Normal: Story = {
  args: { children: <div>contenido sano</div> },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("contenido sano")).toBeInTheDocument()
    await expect(c.queryByRole("alert")).not.toBeInTheDocument()
  },
}

export const HijoQueLanza: Story = {
  args: { label: "dev-full-cycle", children: <Bomba /> },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toBeInTheDocument()
    await expect(c.getByText("El lienzo no pudo renderizar este arnés")).toBeInTheDocument()
    // The raw error.message is visible (honesty: the failure is named, not decorated away).
    await expect(c.getByText("clase desconocida: quimera")).toBeInTheDocument()
    await expect(c.getByText("dev-full-cycle")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeInTheDocument()
  },
}

export const Reintentar: Story = {
  render: () => {
    armarFallaUnaVez()
    return (
      <ErrorBoundary>
        <FallaUnaVez />
      </ErrorBoundary>
    )
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("El lienzo no pudo renderizar este arnés")).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(c.getByText("contenido recuperado")).toBeInTheDocument()
    await expect(c.queryByRole("alert")).not.toBeInTheDocument()
  },
}
