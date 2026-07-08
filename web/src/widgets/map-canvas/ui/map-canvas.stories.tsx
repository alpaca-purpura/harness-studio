import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { expect, within } from "storybook/test"
import {
  type Banda,
  type Clase,
  cobranzaProveedores,
  devFullCycle,
  type Graph,
  luanaFeatureCycle,
  selectArtefactos,
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

// ═══ Franja Artefactos (D1–D11, RF-140..145) — asserts del click-through v2 ═══

// RF-140 · derivación pura: dogfood 5 chips (4 entregas + 1 externo) · Luana fan-out
// spec.md → 3 consumidores · Cobranza refina ↻v2 · dead-end y salida-del-proceso derivados.
export const ArtefactosDerivacion: Story = {
  args: { graph: devFullCycle, artefactos: "todos" },
  play: async ({ canvasElement }) => {
    // dogfood: 5 chips (spec.md · código+tests · veredicto · release@version + idea externa).
    const dogfood = selectArtefactos(devFullCycle)
    expect(dogfood.length).toBe(5)
    expect(dogfood.filter((c) => c.externo).length).toBe(1)

    // Luana: fan-out de spec.md → 3 consumidores (ux-designer, builder req; releaser opt).
    const luana = selectArtefactos(luanaFeatureCycle)
    const specChip = luana.find((c) => c.id === "art-spec.md-spec-writer")
    expect(specChip?.consumidores.length).toBe(3)
    expect(specChip?.consumidores.find((k) => k.id === "releaser")?.requerido).toBe(false)
    // dead-end real (C16): notas de build sin consumidor en caja NO terminal.
    expect(luana.find((c) => c.art === "notas de build")?.dead).toBe(true)
    // salida del proceso (C15): reporte de operación — terminalidad DERIVADA del spine.
    expect(luana.find((c) => c.art === "reporte de operación")?.final).toBe(true)

    // Cobranza: la cadena refina deriva v2 (D9c: jamás declarada).
    const cobranza = selectArtefactos(cobranzaProveedores)
    const v2 = cobranza.find((c) => c.id === "art-factura.pdf-validar-factura")
    expect(v2?.version).toBe(2)
    expect(v2?.opaco).toBe(true)

    // Y el canvas en «todos» pinta los chips del dogfood (RF-141).
    const chips = canvasElement.querySelectorAll(".artchip")
    expect(chips.length).toBeGreaterThan(0)
  },
}

// RF-143 · reposo en «auto» = mapa actual idéntico: CERO chips/gutters en el DOM sin selección.
export const ArtefactosAutoReposo: Story = {
  args: { graph: devFullCycle, artefactos: "auto" },
  play: async ({ canvasElement }) => {
    expect(canvasElement.querySelectorAll(".artchip").length).toBe(0)
    expect(canvasElement.querySelectorAll(".ref").length).toBe(0)
  },
}

// RF-143 · «off» = cero DOM extra (ni siquiera gutters vacíos).
export const ArtefactosOff: Story = {
  args: { graph: devFullCycle, artefactos: "off" },
  play: async ({ canvasElement }) => {
    expect(canvasElement.querySelectorAll(".gutter").length).toBe(0)
  },
}

// RF-144/145 · Cobranza con «pagar» seleccionada (el caso denso del mockup v2): chips
// relacionados visibles, refs ↖ de largo alcance, y click en chip navega al productor.
function CobranzaSeleccionable({ inicial }: { inicial?: string }) {
  const [sel, setSel] = useState<string | undefined>(inicial)
  return (
    <MapCanvas graph={cobranzaProveedores} selectedId={sel} onSelect={setSel} artefactos="auto" />
  )
}
export const ArtefactosCobranzaPagar: Story = {
  render: () => <CobranzaSeleccionable inicial="pagar" />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // Chips de entrada de «pagar» iluminados donde VIVEN (D11a — jamás duplicados):
    // asiento.md nace tras registro; factura.pdf v2 (refina) vive tras validación.
    await expect(c.getByText("asiento.md")).toBeInTheDocument()
    await expect(c.getByText("↻ v2")).toBeInTheDocument()
    // externo de usuario en el gutter de entrada de su consumidor (D11e).
    await expect(c.getByText("aprobación de gerencia")).toBeInTheDocument()
    // ref ↖ de largo alcance: factura.pdf (validación NO es la fase previa de pago).
    const refs = canvasElement.querySelectorAll(".ref")
    await expect(refs.length).toBeGreaterThan(0)
    // el chip real de factura.pdf NO se duplica: vive UNA vez (gutter de registro).
    const facturas = [...canvasElement.querySelectorAll(".artchip")].filter((el) =>
      el.getAttribute("data-node-id")?.startsWith("art-factura.pdf"),
    )
    await expect(facturas.length).toBe(1)
  },
}
