import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import {
  AvisoChip,
  colorDeterminista,
  DerivaChip,
  DotSaludPortafolio,
  EmblemaInicial,
  TipoInstalacionChip,
} from "./chips"

// Story = test (fe-visual-fitness): los chips/dot de dominio del Portafolio (T3, G1/G8/G9).
// Frame reproduce el scope real (.arnesia-portafolio, portafolio.css) para que los tokens y
// las clases modificadoras resuelvan igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const meta = {
  title: "entities/portafolio/Chips",
  component: DerivaChip,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { estado: "al-hilo" },
} satisfies Meta<typeof DerivaChip>

export default meta
type Story = StoryObj<typeof meta>

// G1 — DerivaChip pinta el literal EXACTO de EstadoDeriva, las 3 ramas, sin parafrasear
// (S1-D14: "al-hilo"/"en-deriva"/"deriva-no-evaluable" tal cual).
export const TresEstadosDeriva: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <DerivaChip estado="al-hilo" />
      <DerivaChip estado="en-deriva" />
      <DerivaChip estado="deriva-no-evaluable" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("al-hilo")).toBeInTheDocument()
    await expect(c.getByText("en-deriva")).toBeInTheDocument()
    await expect(c.getByText("deriva-no-evaluable")).toBeInTheDocument()
  },
}

// G1 — `deriva_detalle` es tooltip nativo (title), nunca texto inventado en el chip.
export const DerivaConDetalle: Story = {
  args: {
    estado: "en-deriva",
    detalle: "hash de contenido distinto de la referencia del marketplace",
  },
  play: async ({ canvasElement }) => {
    const chip = canvasElement.querySelector(".pf-chip-deriva")
    await expect(chip).toHaveAttribute(
      "title",
      "hash de contenido distinto de la referencia del marketplace",
    )
    await expect(chip).toHaveClass("en-deriva")
  },
}

// TipoInstalacionChip — badge con el literal de TipoInstalacion (S0-D2), las 3 formas.
export const TresTiposInstalacion: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <TipoInstalacionChip tipo="materializada" />
      <TipoInstalacionChip tipo="referenciada-cc" />
      <TipoInstalacionChip tipo="proyecto-instalado" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("materializada")).toBeInTheDocument()
    await expect(c.getByText("referenciada-cc")).toBeInTheDocument()
    await expect(c.getByText("proyecto-instalado")).toBeInTheDocument()
  },
}

// AvisoChip — SOLO existe si hay `aviso`; sin dato no renderiza nada (plan §3 T3).
export const AvisoPresente: Story = {
  render: () => <AvisoChip aviso="declarada en lock (acme-cli@2.1.0), dir ausente en el caché" />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/declarada en lock/)).toBeInTheDocument()
  },
}

export const AvisoAusente: Story = {
  render: () => (
    <div data-testid="wrap">
      <AvisoChip aviso={undefined} />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const wrap = canvasElement.querySelector('[data-testid="wrap"]')
    await expect(wrap).toBeEmptyDOMElement()
  },
}

// G9 — DotSaludPortafolio: las 3 ramas de S1-D4, cada una con SU aria-label; «sin-senal» es
// un hueco/muted que JAMÁS comparte clase (ni color) con «ok» — assert explícito, no visual.
export const DotSaludTresRamas: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <DotSaludPortafolio salud="ok" />
      <DotSaludPortafolio salud="atencion" />
      <DotSaludPortafolio salud="sin-senal" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const dotOk = c.getByRole("img", { name: "salud: ok" })
    const dotAtencion = c.getByRole("img", { name: "salud: atención" })
    const dotSinSenal = c.getByRole("img", { name: "salud: sin señal" })
    await expect(dotOk).toBeInTheDocument()
    await expect(dotAtencion).toBeInTheDocument()
    await expect(dotSinSenal).toBeInTheDocument()

    // G9 «no sé» ≠ «sano»: sin-senal NO usa la clase (ergo ni el color) de ok.
    await expect(dotOk).toHaveClass("ok")
    await expect(dotSinSenal).not.toHaveClass("ok")
    await expect(dotSinSenal).toHaveClass("sin-senal")
    await expect(dotOk.className).not.toBe(dotSinSenal.className)

    // Y el color computado (no solo la clase) difiere — assert explícito sobre el CSSOM real.
    const bgOk = getComputedStyle(dotOk).backgroundColor
    const bgSinSenal = getComputedStyle(dotSinSenal).backgroundColor
    await expect(bgOk).not.toBe(bgSinSenal)
  },
}

// EmblemaInicial — inicial + color determinista (función pura `colorDeterminista`, no
// Math.random): el MISMO texto da SIEMPRE el mismo color; textos distintos pueden diferir.
export const EmblemaColorDeterminista: Story = {
  render: () => (
    <div style={{ display: "flex", gap: 8 }}>
      <EmblemaInicial texto="harness" />
      <EmblemaInicial texto="harness" />
      <EmblemaInicial texto="acme-cli" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    await expect(colorDeterminista("harness")).toBe(colorDeterminista("harness"))
    const spans = canvasElement.querySelectorAll("span[aria-hidden]")
    await expect(spans).toHaveLength(3)
    // Las dos primeras (mismo texto "harness") pintan EXACTAMENTE el mismo background.
    const bg0 = (spans[0] as HTMLElement).style.background
    const bg1 = (spans[1] as HTMLElement).style.background
    await expect(bg0).toBe(bg1)
    await expect((spans[0] as HTMLElement).textContent).toBe("H")
  },
}

// G6 — identidad provisional real (fixture b: `identidad.id === ""`): el emblema pinta "?"
// honesto, jamás un placeholder inventado ni un crash.
export const EmblemaSinTexto: Story = {
  render: () => <EmblemaInicial texto="" />,
  play: async ({ canvasElement }) => {
    const span = canvasElement.querySelector("span[aria-hidden]")
    await expect(span).toHaveTextContent("?")
  },
}
