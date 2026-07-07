import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { expect, within } from "storybook/test"
import {
  type Banda,
  type Clase,
  devFullCycle,
  type Graph,
  luanaFeatureCycle,
  selectBandaDesconocida,
} from "@/entities/arnes"
import { MapCanvas } from "./map-canvas"

// Story = test: the full map surface. This is the fitness fixture of record — the signed mockup
// is derived from what renders here. a11y `todo`: intentional low-contrast micro-labels (badges/
// chips of the signed palette); the DOM assertions still gate CI.
const meta = {
  title: "widgets/map-canvas/MapCanvas",
  component: MapCanvas,
  parameters: { layout: "fullscreen", a11y: { test: "todo" } },
  decorators: [
    (Story) => (
      <div style={{ height: 620 }}>
        <Story />
      </div>
    ),
  ],
  args: { graph: devFullCycle },
} satisfies Meta<typeof MapCanvas>

export default meta
type Story = StoryObj<typeof meta>

// The dogfood arnés (dev-full-cycle) — what the real mount serves (shot7).
export const Dogfood: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Phase cajas render with their transition.
    await expect(c.getByText("escribir el spec")).toBeInTheDocument()
    await expect(c.getByText("construir contra el spec")).toBeInTheDocument()
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
    // Base rule renders (inside the collapsed Reglas subband — present, just hidden).
    await expect(c.getByText("estándar de spec")).toBeInTheDocument()
    // Guardia is empty for this arnés.
    await expect(c.getByText("— sin hooks —")).toBeInTheDocument()
  },
}

function Selectable() {
  const [sel, setSel] = useState<string>()
  return <MapCanvas graph={devFullCycle} selectedId={sel} onSelect={setSel} />
}
export const WithSelection: Story = { render: () => <Selectable /> }

// A COMPLETE arnés (Luana) with the example/picker toggle — the dense parity view (shot1).
export const LuanaCompleto: Story = {
  args: {
    graph: luanaFeatureCycle,
    harnesses: [
      { id: "luana-feature-cycle", label: "Luana" },
      { id: "dev-full-cycle", label: "dev-full-cycle" },
    ],
    activeId: "luana-feature-cycle",
  },
  decorators: [
    (Story) => (
      <div style={{ height: 820 }}>
        <Story />
      </div>
    ),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("destilar necesidad")).toBeInTheDocument()
    await expect(c.getByText("guardia PII")).toBeInTheDocument()
    await expect(c.getByText("acceso a la API")).toBeInTheDocument()
    await expect(c.getByText("kit luana")).toBeInTheDocument()
    // The active picker button is pressed (RF-55).
    await expect(c.getByRole("button", { name: "Luana" })).toHaveAttribute("aria-pressed", "true")
  },
}

// Reconciliación honesta (nomenclatura-arnes §4.5) — a node whose banda is absent or outside
// the contract enum does NOT vanish: selectSoporte folds it into the `base` band of the Base
// region, and selectBandaDesconocida marks it detectably. Graphs arrive as JSON at runtime,
// so the `as unknown as Banda` cast simulates exactly that dirty data.
const conBandaDesconocida: Graph = {
  ...devFullCycle,
  nodos: [
    ...devFullCycle.nodos,
    { id: "sin-banda", clase: "mcp", nombre: "nodo sin banda" },
    {
      id: "banda-rara",
      clase: "skill",
      nombre: "nodo banda rara",
      banda: "quimera" as unknown as Banda,
    },
  ],
}

export const BandaDesconocida: Story = {
  args: { graph: conBandaDesconocida },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Both misfits render, visible in the Base region (never silently dropped).
    await expect(c.getByText("nodo sin banda")).toBeInTheDocument()
    await expect(c.getByText("nodo banda rara")).toBeInTheDocument()
    // And the selector marks them detectably (the visual badge lands with the Hito 2 WIP).
    expect([...selectBandaDesconocida(conBandaDesconocida)]).toEqual(["sin-banda", "banda-rara"])
  },
}

// A clase outside the enum crashes the node render (KIND[clase] → TypeError). The canvas
// boundary catches it and paints the honest fallback — the failure is visible and CONTAINED
// to the canvas, never a white screen for the whole app (§4.5).
const conClaseRota: Graph = {
  ...devFullCycle,
  nodos: [
    ...devFullCycle.nodos,
    {
      id: "alien",
      clase: "quimera" as unknown as Clase,
      nombre: "nodo alien",
      banda: "base",
    },
  ],
}

export const NodoMalformado: Story = {
  args: { graph: conClaseRota },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("El lienzo no pudo renderizar este arnés")).toBeInTheDocument()
    // The panel names the failing arnés and offers recovery.
    await expect(c.getByText("dev-full-cycle")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeInTheDocument()
  },
}
