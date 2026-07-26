import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import { TITULO_CONFIANZA } from "../model/selectors"
import { MarcaConfianza } from "./marca-confianza"

// Story = test (fe-visual-fitness). RF-242 · RF-274 · H-13: los cuatro casos de
// `atribucion_confianza` se distinguen por TEXTO, no por tono.

const meta = {
  title: "entities/telemetria/MarcaConfianza",
  component: MarcaConfianza,
  args: { confianza: "exacta" },
} satisfies Meta<typeof MarcaConfianza>

export default meta
type Story = StoryObj<typeof meta>

// RF-242 — el caso que se cae del olvido: **exacta no renderiza NADA**. La ausencia de marca
// *es* la señal (design §5.3). Si pintáramos «exacta» al lado de cada número, el ruido enterraría
// las tres marcas que sí importan.
export const Exacta: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector("[data-confianza]")).toBeNull()
    await expect(c.queryByText(/por huella|por proceso|aproximad/i)).toBeNull()
  },
}

// RF-242 · RF-274 — el runtime redacta el nombre del arnés; se identifica por la huella del
// plugin. El `title` largo dice de dónde salió y que corrió afuera.
export const PorHash: Story = {
  args: { confianza: "por-hash" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByText("por huella")
    await expect(chip).toHaveAttribute("title", TITULO_CONFIANZA["por-hash"] as string)
    await expect(chip).toHaveAttribute(
      "title",
      "Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA.",
    )
    // El subrayado punteado es la afordancia de «hay más que leer acá».
    await expect(getComputedStyle(chip).textDecorationStyle).toBe("dotted")
  },
}

// RF-242 · H-13 — «por proceso» NO es lo mismo que «por huella»: una identifica, la otra
// deduce y puede estar MEZCLANDO dos arneses del mismo directorio. El title lo advierte, y el
// assert de desigualdad impide que alguien los unifique «para simplificar».
export const PorProceso: Story = {
  args: { confianza: "por-proceso" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByText("por proceso")
    await expect(chip).toHaveAttribute(
      "title",
      "Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla.",
    )
    await expect(TITULO_CONFIANZA["por-proceso"]).not.toBe(TITULO_CONFIANZA["por-hash"])
    await expect(c.queryByText("por huella")).toBeNull()
  },
}

// RF-242 — sin dato NO hay número que marcar. La marca existe igual, con su motivo: la caja
// suma a la cobertura, no al total.
export const SinDato: Story = {
  args: { confianza: "sin-dato" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin dato atribuible")).toBeInTheDocument()
    await expect(c.queryByText(/USD/)).toBeNull()
    await expect(c.queryByText("0,00")).toBeNull()
    await expect(c.getByText("sin dato atribuible")).toHaveAttribute(
      "title",
      TITULO_CONFIANZA["sin-dato"] as string,
    )
  },
}

// RF-242 · H-13 — el candado del vocabulario: los cuatro juntos, y **ninguno dice
// «aproximada»**. Esa palabra taparía tres cosas distintas bajo una. Tres `title` distintos
// (exacta no tiene) prueba que las tres marcas visibles dicen cosas diferentes.
export const LosCuatroJuntos: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <MarcaConfianza confianza="exacta" />
      <MarcaConfianza confianza="por-hash" />
      <MarcaConfianza confianza="por-proceso" />
      <MarcaConfianza confianza="sin-dato" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByText("por huella")).toHaveLength(1)
    await expect(c.getAllByText("por proceso")).toHaveLength(1)
    await expect(c.queryByText(/aproximad/i)).toBeNull()
    const titles = [...canvasElement.querySelectorAll("[data-confianza]")].map((e) =>
      e.getAttribute("title"),
    )
    await expect(titles).toHaveLength(3)
    await expect(new Set(titles).size).toBe(3)
  },
}
