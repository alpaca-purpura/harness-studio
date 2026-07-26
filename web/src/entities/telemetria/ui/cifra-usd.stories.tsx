import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { CifraUsd } from "./cifra-usd"

// Story = test (fe-visual-fitness). RF-281: UN formateador de dinero para las cinco superficies.
// D23 — estas 4 stories cubren el CONTRATO completo de la pieza, no solo el caso del Mapa.

const meta = {
  title: "entities/telemetria/CifraUsd",
  component: CifraUsd,
  args: { micros: 1_920_000 },
} satisfies Meta<typeof CifraUsd>

export default meta
type Story = StoryObj<typeof meta>

// RF-281 — el caso normal: `USD` atenuado en su span, el monto en mono con `tabular-nums` para
// que las columnas de una tabla alineen sin depender del ancho de cada dígito.
export const Estandar: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("USD")).toBeInTheDocument()
    await expect(c.getByText("1,92")).toBeInTheDocument()
    const raiz = canvasElement.querySelector(".num") as HTMLElement
    await expect(raiz).not.toBeNull()
    await expect(getComputedStyle(raiz).fontVariantNumeric).toContain("tabular-nums")
  },
}

// RF-281 — separador de millar de espacio fino y coma decimal. **Jamás el formato EN**: un
// `1,234.56` en una app en español se lee como mil doscientos treinta y cuatro coma cincuenta y seis.
export const SeparadorDosDecimales: Story = {
  args: { micros: 1_234_560_000 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/1[\s ]234,56/)).toBeInTheDocument()
    await expect(c.queryByText(/1,234\.56/)).toBeNull()
  },
}

// RF-281 — el caso que da nombre a la regla: **`0,004` no se redondea a `0,00`**. Un monto que
// existe y se muestra como cero es la mentira barata que este paquete existe para no decir.
export const MenorAlCentavo: Story = {
  args: { micros: 4000 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("0,00")).toBeNull()
    await expect(canvasElement.textContent).toContain("0,004")
  },
}

// RF-281 · BR-M2 — sin monto se DICE. Ni `0,00` (sería falso), ni `USD` huérfano, ni `—` a
// secas (un guion no distingue «falta el dato» de «no aplica» de «es cero»).
export const Ausente: Story = {
  args: { micros: null },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin dato")).toBeInTheDocument()
    await expect(c.queryByText("USD")).toBeNull()
    await expect(c.queryByText("0,00")).toBeNull()
    await expect(c.queryByText("—")).toBeNull()
  },
}
