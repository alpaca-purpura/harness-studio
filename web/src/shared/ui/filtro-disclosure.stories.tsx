import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { useState } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { FiltroDisclosure } from "./filtro-disclosure"

// Story = test (fe-visual-fitness) — molécula FiltroDisclosure, PROMOVIDA de
// `widgets/portafolio/ui/portafolio-list.tsx` (§8.2). Refactor de movimiento: la conducta ya
// estaba firmada y testeada por `FiltroEstadoAcota`/`FiltroMarketplaceAcota` (que siguen
// pasando SIN tocarlas — ése es el criterio de éxito). Estas stories cubren la molécula sola.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

type Valor = "uno" | "dos" | "tres"

const VALORES: readonly Valor[] = ["uno", "dos", "tres"]

const meta = {
  title: "shared/ui/FiltroDisclosure",
  component: FiltroDisclosure<Valor>,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    etiqueta: "Estado",
    abierto: false,
    onToggleAbierto: fn(),
    panelId: "demo-panel",
    valores: VALORES,
    labelDe: (v: Valor): string => v,
    seleccion: new Set<Valor>(),
    onCambiar: fn(),
  },
} satisfies Meta<typeof FiltroDisclosure<Valor>>

export default meta
type Story = StoryObj<typeof meta>

// Cerrado: `aria-expanded="false"` + panel ausente del DOM (no oculto por CSS).
export const Cerrado: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Estado" })
    await expect(btn).toHaveAttribute("aria-expanded", "false")
    await expect(btn).toHaveAttribute("aria-controls", "demo-panel")
    await expect(c.queryByRole("group")).toBeNull()
    await userEvent.click(btn)
    await expect(args.onToggleAbierto).toHaveBeenCalledTimes(1)
  },
}

// Abierto: panel con `role="group"` rotulado + una chip por valor; el contador del botón sigue
// a la selección; el vocabulario entra por `labelDe` (la molécula no conoce ningún dominio).
export const AbiertoConSeleccion: Story = {
  args: {
    abierto: true,
    seleccion: new Set<Valor>(["dos"]),
    labelDe: (v: Valor) => `valor ${v}`,
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Estado (1)" })).toBeInTheDocument()
    await expect(c.getByRole("group", { name: "Filtrar por estado" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "valor dos" })).toHaveAttribute(
      "aria-pressed",
      "true",
    )
    await userEvent.click(c.getByRole("button", { name: "valor uno" }))
    await expect(args.onCambiar).toHaveBeenCalledWith(new Set<Valor>(["dos", "uno"]))
  },
}

// Multi-select real: `toggleEnSet` (promovido a shared/lib) agrega y saca sin mutar el Set
// entrante. Se ejercita con estado local para ver las dos direcciones.
export const MultiSelectToggle: Story = {
  render: (args) => {
    const [seleccion, setSeleccion] = useState<ReadonlySet<Valor>>(new Set<Valor>())
    return (
      <FiltroDisclosure<Valor> {...args} abierto seleccion={seleccion} onCambiar={setSeleccion} />
    )
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const uno = c.getByRole("button", { name: "uno" })
    await userEvent.click(uno)
    await expect(uno).toHaveAttribute("aria-pressed", "true")
    await expect(c.getByRole("button", { name: "Estado (1)" })).toBeInTheDocument()
    await userEvent.click(uno)
    await expect(uno).toHaveAttribute("aria-pressed", "false")
    await expect(c.getByRole("button", { name: "Estado" })).toBeInTheDocument()
  },
}

// Sin valores: el copy `vacio` lo pone el caller (la molécula no inventa un texto de dominio).
export const SinValores: Story = {
  args: { abierto: true, valores: [], vacio: "sin marketplaces en los datos actuales" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin marketplaces en los datos actuales")).toBeInTheDocument()
  },
}
