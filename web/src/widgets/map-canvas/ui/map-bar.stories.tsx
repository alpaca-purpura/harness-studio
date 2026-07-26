import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import { devFullCycle } from "@/entities/arnes"
import { MapBar } from "./map-bar"

const meta = {
  title: "widgets/map-canvas/MapBar",
  component: MapBar,
  parameters: { layout: "fullscreen" },
  args: { arnes: devFullCycle.arnes, capa: "estructura", onCapa: fn() },
} satisfies Meta<typeof MapBar>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("tab", { name: "Estructura" })).toHaveAttribute(
      "aria-selected",
      "true",
    )
    // Layers needing telemetry are disabled (honest, not hidden).
    await expect(c.getByRole("tab", { name: "Desempeño" })).toBeDisabled()
  },
}

// RF-232 — el conmutador ofrece «Mejora» y ya NO ofrece «Tokens». Los cuatro slots se
// conservan, en el mismo orden: el renombre no le quita una posición al conmutador.
// 🔴 Desviación declarada de un baseline firmado (mockups/INDEX.md regla 3 · D17.1).
export const CapaMejoraDisponible: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const tabs = c.getAllByRole("tab")
    await expect(tabs).toHaveLength(4)
    await expect(tabs.map((t) => t.textContent)).toEqual([
      "Estructura",
      "Mejora",
      "Desempeño",
      "Proceso",
    ])
    await expect(c.queryByRole("tab", { name: "Tokens" })).toBeNull()
    await expect(c.getByRole("tab", { name: "Mejora" })).not.toBeDisabled()
  },
}

// RF-232 · RF-276 — con la capa activa, `aria-selected` la marca y el resto queda en false;
// volver a Estructura sigue siendo un click que avisa a la página (el estado vive arriba).
export const CapaMejoraActiva: Story = {
  args: { capa: "mejora" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("tab", { name: "Mejora" })).toHaveAttribute("aria-selected", "true")
    await expect(c.getByRole("tab", { name: "Estructura" })).toHaveAttribute(
      "aria-selected",
      "false",
    )
    await c.getByRole("tab", { name: "Estructura" }).click()
    await expect(args.onCapa).toHaveBeenCalledWith("estructura")
  },
}

// RF-233 — los dos slots apagados dicen la VERDAD DE HOY. El literal viejo («Necesita
// telemetría (indexer JSONL)») mandaba a construir un indexer que ya existe: la señal llega
// (V1), lo que falta es el diseño. El assert de ausencia es el que impide que vuelva.
export const MotivosHonestos: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("tab", { name: "Desempeño" })).toHaveAttribute(
      "title",
      "La señal ya llega —duración por request y por herramienta—. Falta decidir qué es «desempeño» a nivel Mapa.",
    )
    await expect(c.getByRole("tab", { name: "Proceso" })).toHaveAttribute(
      "title",
      "Entra en parte por la capa Mejora. Falta mapear todos los eventos a fases del arnés.",
    )
    await expect(c.queryByText(/Necesita telemetría/)).toBeNull()
    const titles = [...canvasElement.querySelectorAll("[title]")].map((e) =>
      e.getAttribute("title"),
    )
    for (const t of titles) await expect(t).not.toMatch(/Necesita telemetría/)
  },
}

// RF-276 — el motivo no puede vivir SOLO en `title`: un lector de pantalla puede no anunciarlo.
// El `sr-only` referenciado por `aria-describedby` lleva el MISMO texto, y el assert de igualdad
// es lo que impide que los dos driften con el tiempo.
export const MotivoAccesible: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    for (const nombre of ["Desempeño", "Proceso"]) {
      const tab = c.getByRole("tab", { name: nombre })
      const id = tab.getAttribute("aria-describedby")
      await expect(id).toBeTruthy()
      const desc = canvasElement.querySelector(`#${id}`) as HTMLElement
      await expect(desc).not.toBeNull()
      await expect(desc).toHaveClass("sr-only")
      await expect(desc.textContent).toBe(tab.getAttribute("title"))
    }
    // Los slots encendidos no describen nada: no hay motivo que dar.
    await expect(c.getByRole("tab", { name: "Mejora" })).not.toHaveAttribute("aria-describedby")
  },
}
