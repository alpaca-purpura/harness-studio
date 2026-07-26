import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, within } from "storybook/test"
import type { Cobertura } from "../model/types"
import { BarraCobertura } from "./barra-cobertura"

// Story = test (fe-visual-fitness). RF-237 · RF-279 · H-8 · H-11.
//
// El fixture de estas 5 stories (12/3/2/1 sobre 18) es DEL COMPONENTE, no del producto: prueba
// el contrato de formateo con cuatro segmentos no nulos. El juego coherente de la iteración 2
// del mockup (44/9/5/3 sobre 61) vive en `FranjaMejora`, que es donde tiene que cerrar con el
// denominador del total. Los dos textos los produce ESTE componente a partir de sus props, así
// que no hay copy duplicado — solo dos fixtures.

const cobertura: Cobertura = {
  esperados: 18,
  exacta: 12,
  por_hash: 3,
  por_proceso: 2,
  sin_dato: 1,
  no_llegaron: 0,
}

const meta = {
  title: "entities/telemetria/BarraCobertura",
  component: BarraCobertura,
  args: { cobertura },
} satisfies Meta<typeof BarraCobertura>

export default meta
type Story = StoryObj<typeof meta>

// RF-237 · H-8 — cuatro niveles, anchos proporcionales, y **el assert de verdad es el texto**:
// axe no mira contraste de no-texto, así que un test de colores dejaría pasar una barra
// ilegible. El `aria-label` declara la unidad («corridas»), que es lo que H-8 pedía cerrar.
export const CuatroSegmentos: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const segs = canvasElement.querySelectorAll(".cov-seg")
    await expect(segs).toHaveLength(4)
    const anchos = [...segs].map((s) => Number.parseFloat((s as HTMLElement).style.width))
    for (const [i, esperado] of [12, 3, 2, 1].entries()) {
      await expect(Math.abs((anchos[i] as number) - (esperado / 18) * 100)).toBeLessThan(1)
    }
    await expect(
      c.getByText("12 exactas · 3 por huella · 2 por proceso · 1 sin dato — sobre 18 corridas"),
    ).toBeInTheDocument()
    await expect(
      c.getByRole("img", {
        name: "Cobertura de la atribución: 12 corridas exactas, 3 por huella, 2 por proceso, 1 sin dato, sobre 18 corridas.",
      }),
    ).toBeInTheDocument()
  },
}

// RF-237 — una categoría en cero **no ocupa lugar ni se nombra**: un segmento de ancho 0 con su
// etiqueta al lado es ruido que el ojo cuenta igual.
export const CategoriaEnCeroNoOcupaLugar: Story = {
  args: { cobertura: { ...cobertura, por_proceso: 0 } },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".cov-seg")).toHaveLength(3)
    await expect(c.queryByText(/por proceso/)).toBeNull()
  },
}

// RF-237 — cobertura perfecta: un solo segmento y una frase que lo dice en positivo, no una
// enumeración de tres ceros.
export const CoberturaCompleta: Story = {
  args: {
    cobertura: {
      esperados: 18,
      exacta: 18,
      por_hash: 0,
      por_proceso: 0,
      sin_dato: 0,
      no_llegaron: 0,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".cov-seg")).toHaveLength(1)
    await expect(c.getByText("atribución exacta en las 18 corridas")).toBeInTheDocument()
  },
}

// RF-237 · RF-269 — con 0 corridas **la barra no se dibuja**. No es un 0 %: es que no hay
// denominador, y quien manda es el estado 1 de la franja.
export const SinCorridas: Story = {
  args: {
    cobertura: {
      esperados: 0,
      exacta: 0,
      por_hash: 0,
      por_proceso: 0,
      sin_dato: 0,
      no_llegaron: 0,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".cov-bar")).toBeNull()
    await expect(c.getByText("sin corridas que atribuir en esta ventana")).toBeInTheDocument()
    await expect(c.queryByText("0")).toBeNull()
  },
}

// H-11 · RF-279 — el rótulo `cobertura` es visible SIEMPRE, aun a 320 px: al envolver, la barra
// no puede quedar huérfana de su nombre.
export const RotuloVisibleSiempre: Story = {
  decorators: [(Story) => <div style={{ width: 320 }}>{Story()}</div>],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("cobertura")).toBeVisible()
  },
}
