import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { BuscadorFiltro } from "./buscador-filtro"
import { FiltroDisclosure } from "./filtro-disclosure"

// Story = test (fe-visual-fitness) — molécula BuscadorFiltro (AG-D3).
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const meta = {
  title: "shared/ui/BuscadorFiltro",
  component: BuscadorFiltro,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    busqueda: "",
    onBusqueda: fn(),
    placeholder: "Buscar en el catálogo…",
    ariaLabel: "Buscar en el catálogo",
  },
} satisfies Meta<typeof BuscadorFiltro>

export default meta
type Story = StoryObj<typeof meta>

// El input es un `search` rotulado por `aria-label` inyectado; teclear propaga tal cual (la
// molécula no filtra: los selectores son del dominio y viven en entities/*/model).
export const SoloBuscador: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const input = c.getByRole("searchbox", { name: "Buscar en el catálogo" })
    await expect(input).toHaveAttribute("placeholder", "Buscar en el catálogo…")
    await userEvent.type(input, "harness")
    await expect(args.onBusqueda).toHaveBeenCalled()
  },
}

// Composición: el caller arma los `FiltroDisclosure` y los pasa por `filtros` — así el
// vocabulario de dominio nunca entra a `shared/ui` (`shared-no-upward`, severidad `error`).
export const ConFiltros: Story = {
  args: {
    filtros: (
      <FiltroDisclosure<"al-hilo" | "en-deriva">
        etiqueta="Situación"
        abierto={false}
        onToggleAbierto={fn()}
        panelId="demo-situacion"
        valores={["al-hilo", "en-deriva"]}
        labelDe={(v) => v}
        seleccion={new Set()}
        onCambiar={fn()}
      />
    ),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("searchbox", { name: "Buscar en el catálogo" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Situación" })).toBeInTheDocument()
  },
}
