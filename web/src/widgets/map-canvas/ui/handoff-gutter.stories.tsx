import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import {
  CAP_GUTTER,
  type ChipArtefacto,
  cobranzaProveedores,
  luanaFeatureCycle,
  planGutter,
  selectArtefactos,
  selectRefsEntrada,
} from "@/entities/arnes"
import { HandoffGutter } from "./handoff-gutter"

// Story = test: el gutter de hand-off (D2 firmada, RF-141/144) con el plan derivado de
// entities — tope de densidad D11c («+N más»/«− plegar»), prioridad de relacionados, y
// el panel de entrada ↖ (D11b, refs de largo alcance — índice navegable, jamás chip).

// El gutter saturado del mockup v2: la caja «pagar» de Cobranza (5 necesita).
const chipsCobranza = selectArtefactos(cobranzaProveedores)
const gutterPago = chipsCobranza.filter((c) => c.after === "registro" || c.before === "pago")

const meta = {
  title: "widgets/map-canvas/HandoffGutter",
  component: HandoffGutter,
  parameters: { a11y: { test: "todo" } },
  decorators: [
    (Story) => (
      <div className="arnesia-map">
        <div style={{ display: "flex" }}>{Story()}</div>
      </div>
    ),
  ],
  args: {
    plan: planGutter(gutterPago, { mode: "todos" }),
    refs: [],
    expanded: false,
    onToggleExpand: fn(),
  },
} satisfies Meta<typeof HandoffGutter>

export default meta
type Story = StoryObj<typeof meta>

// Tope D11c: el gutter de «pagar» trae >3 elegibles → CAP chips + botón «+N más».
export const TopeMasN: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const visibles = canvasElement.querySelectorAll(".artchip")
    await expect(visibles.length).toBe(CAP_GUTTER)
    await expect(gutterPago.length).toBeGreaterThan(CAP_GUTTER)
    await expect(c.getByText(`+${gutterPago.length - CAP_GUTTER} más`)).toBeInTheDocument()
  },
}

// Expandido: todos los elegibles a la vista + «− plegar».
export const Expandido: Story = {
  args: {
    plan: planGutter(gutterPago, { mode: "todos", expanded: true }),
    expanded: true,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".artchip").length).toBe(gutterPago.length)
    await expect(c.getByText("− plegar")).toBeInTheDocument()
  },
}

// Prioridad D11c: los relacionados con la selección SALTAN el tope — con «pagar»
// seleccionada, sus chips entran aunque el gutter esté al tope.
export const RelacionadosSaltanTope: Story = {
  args: {
    plan: planGutter(gutterPago, { mode: "todos", selectedId: "pagar" }),
  },
  play: async ({ canvasElement }) => {
    const relacionados = gutterPago.filter((c: ChipArtefacto) =>
      c.consumidores.some((k) => k.id === "pagar"),
    )
    const shown = [...canvasElement.querySelectorAll(".artchip")].map((el) =>
      el.getAttribute("data-node-id"),
    )
    for (const r of relacionados) {
      await expect(shown).toContain(r.id)
    }
  },
}

// Panel de entrada ↖ (D11b): el fan-in de Luana — «promover a prod» seleccionada muestra
// las refs de largo alcance (spec/diseño/necesidad, opcionales) SIN duplicar chips.
export const PanelDeEntrada: Story = {
  args: {
    plan: planGutter([], { mode: "auto", selectedId: "releaser" }),
    refs: selectRefsEntrada(luanaFeatureCycle, "releaser", "review"),
    onSelect: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const c = within(canvasElement)
    // veredicto viene de la fase adyacente (review) → NO es ref; los 3 lejanos sí.
    const refs = canvasElement.querySelectorAll(".ref")
    await expect(refs.length).toBe(3)
    await expect(c.getByText("spec.md")).toBeInTheDocument()
    await expect(c.getByText("necesidad.md")).toBeInTheDocument()
    // opcional (D11d): tag en la referencia.
    await expect(c.getAllByText("opcional").length).toBe(3)
    // click en la ref navega al productor (RF-144): la primera ref es spec.md → spec-writer.
    canvasElement.querySelector<HTMLButtonElement>(".ref")?.click()
    await expect(args.onSelect).toHaveBeenCalledWith("spec-writer")
  },
}
