import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import { GrupoControl } from "./grupo-control"

// Story = test (fe-visual-fitness) — molécula GrupoControl (AG-D4, cierra L1/L3 de la auditoría).
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const meta = {
  title: "shared/ui/GrupoControl",
  component: GrupoControl,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    rotulo: "Ver por",
    ariaLabel: "Lente del Portafolio",
    children: (
      <button type="button" className="pf-lente-btn" aria-pressed="true">
        Empresa
      </button>
    ),
  },
} satisfies Meta<typeof GrupoControl>

export default meta
type Story = StoryObj<typeof meta>

// AG-D4 — el rótulo es TEXTO VISIBLE en el DOM (no un `aria-label` invisible) y el
// `role="group"`+`aria-label` existente se CONSERVA: se agrega afordancia visual sin degradar
// a11y. Las mayúsculas las pone el CSS, así que el DOM dice «Ver por» tal cual.
export const RotuloVisibleConAriaIntacto: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Ver por")).toBeInTheDocument()
    await expect(c.getByRole("group", { name: "Lente del Portafolio" })).toBeInTheDocument()
  },
}

// L3 de la auditoría: el hint del mockup existía y el código lo había perdido. Es opcional —
// sin dato, no se renderiza nada (mismo criterio que AvisoChip).
export const ConHint: Story = {
  args: { hint: "lente para encontrar un arnés cuando hay muchos" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("lente para encontrar un arnés cuando hay muchos")).toBeInTheDocument()
  },
}
