import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { Sparkline } from "./sparkline"

// Story = test (fe-visual-fitness). RF-266 · RF-279: la tendencia SIEMPRE lleva su texto
// equivalente — cinco barras de 5 px no son un canal accesible.

const meta = {
  title: "entities/telemetria/Sparkline",
  component: Sparkline,
  args: { puntos: [240_000, 260_000, 255_000, 290_000, 310_000] },
} satisfies Meta<typeof Sparkline>

export default meta
type Story = StoryObj<typeof meta>

// RF-266 · RF-279 — la dirección va en palabras en el `aria-label`, con el conteo: «5 corridas»
// no es «siempre». La última barra se distingue porque es la corrida más reciente.
export const EnAlza: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByRole("img", { name: "Tendencia en alza en las últimas 5 corridas." }),
    ).toBeInTheDocument()
    const barras = canvasElement.querySelectorAll(".spark-bar")
    await expect(barras).toHaveLength(5)
    await expect(canvasElement.querySelectorAll("[data-ultima='true']")).toHaveLength(1)
    await expect(barras[4]).toHaveAttribute("data-ultima", "true")
  },
}

export const Estable: Story = {
  args: { puntos: [100_000, 101_000, 99_000, 100_500, 100_000] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("img", { name: /estable/ })).toBeInTheDocument()
  },
}

export const HaciaLaBaja: Story = {
  args: { puntos: [310_000, 290_000, 255_000, 260_000, 240_000] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("img", { name: /a la baja/ })).toBeInTheDocument()
  },
}

// RF-266 — con un solo punto **no se dibuja nada** y se dice por qué. Una línea plana de un
// punto diría «estable», que es una afirmación que nadie midió.
export const PocasCorridas: Story = {
  args: { puntos: [240_000] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".spark")).toBeNull()
    await expect(c.getByText("pocas corridas para una tendencia")).toBeInTheDocument()
  },
}
