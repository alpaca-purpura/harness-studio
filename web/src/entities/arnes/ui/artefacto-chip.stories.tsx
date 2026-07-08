import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, within } from "storybook/test"
import type { ChipArtefacto } from "../model/artefactos"
import { ArtefactoChip } from "./artefacto-chip"

// Story = test (fe-visual-fitness): cada MARCA del chip del mockup v2 (RF-141,
// mockup:122-139,490-508). Los estilos viven en map.css bajo `.arnesia-map`; el
// wrapper reproduce el ancho del gutter (150px, mockup:142).
function Cell({ children }: { children: ReactNode }) {
  return (
    <div className="arnesia-map">
      <div style={{ width: 150, display: "flex", flexDirection: "column", gap: 8 }}>{children}</div>
    </div>
  )
}

const documento: ChipArtefacto = {
  id: "art-spec.md-spec-writer",
  art: "spec.md",
  externo: false,
  productor: "spec-writer",
  consumidores: [{ id: "builder", requerido: true }],
  path: true,
  plantilla: true,
  opaco: false,
  refina: null,
  final: false,
  dead: false,
  after: "spec",
  version: 1,
}

const meta = {
  title: "entities/arnes/ArtefactoChip",
  component: ArtefactoChip,
  // a11y `todo`: micro-tags de 9px del design firmado (contraste intencional del mockup);
  // las aserciones DOM del play siguen gateando CI.
  parameters: { a11y: { test: "todo" } },
  decorators: [(Story) => <Cell>{Story()}</Cell>],
  args: { chip: documento },
} satisfies Meta<typeof ArtefactoChip>

export default meta
type Story = StoryObj<typeof meta>

// Documento con path + plantilla (C17): nombre mono recto, tag «plantilla».
export const DocumentoConPlantilla: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("spec.md")).toBeInTheDocument()
    await expect(c.getByText("plantilla")).toBeInTheDocument()
    const chip = canvasElement.querySelector(".artchip")
    await expect(chip).not.toHaveClass("etiqueta")
    await expect(chip).toHaveAttribute("data-node-id", "art-spec.md-spec-writer")
  },
}

// Etiqueta sin path (C19/D3): cursiva, sin tag plantilla.
export const Etiqueta: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-codigo-tests-builder",
      art: "código + tests",
      productor: "builder",
      path: false,
      plantilla: false,
    },
  },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector(".artchip")).toHaveClass("etiqueta")
  },
}

// Opaco (C20/D10): papel relleno; título lo declara.
export const Opaco: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-factura.pdf-recibir",
      art: "factura.pdf",
      productor: "recibir-factura",
      plantilla: false,
      opaco: true,
    },
  },
  play: async ({ canvasElement }) => {
    const chip = canvasElement.querySelector(".artchip")
    await expect(chip).toHaveClass("opaco")
    await expect(chip).toHaveAttribute("title", expect.stringContaining("opaco (D10)"))
  },
}

// Externo (C3/C4/D10): borde punteado + tag «externo»; click → primer consumidor.
export const Externo: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-ext-aprobacion-pagar",
      art: "aprobación de gerencia",
      externo: true,
      origen: "usuario",
      productor: null,
      consumidores: [{ id: "pagar", requerido: true }],
      path: false,
      plantilla: false,
    },
    onSelect: fn(),
  },
  play: async ({ args, canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("externo")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".artchip")).toHaveClass("externo")
    canvasElement.querySelector<HTMLButtonElement>(".artchip")?.click()
    await expect(args.onSelect).toHaveBeenCalledWith("pagar")
  },
}

// Revisión refina (C11/D9): tag «↻ v2» — versión DERIVADA de la cadena.
export const RefinaV2: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-factura.pdf-validar",
      art: "factura.pdf",
      productor: "validar-factura",
      plantilla: false,
      opaco: true,
      refina: "factura.pdf",
      version: 2,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("↻ v2")).toBeInTheDocument()
  },
}

// Dead-end (C16/D8): warn «sin consumidor».
export const DeadEnd: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-notas-builder",
      art: "notas de build",
      productor: "builder",
      consumidores: [],
      plantilla: false,
      dead: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin consumidor")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".artchip")).toHaveClass("dead")
  },
}

// Entrega terminal (C15): badge verde «salida del proceso» — terminalidad derivada.
export const SalidaDelProceso: Story = {
  args: {
    chip: {
      ...documento,
      id: "art-comprobante-pagar",
      art: "comprobante de pago.pdf",
      productor: "pagar",
      consumidores: [],
      plantilla: false,
      opaco: true,
      final: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("salida del proceso")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".artchip")).toHaveClass("final")
  },
}

// Click → productor (RF-145).
export const ClickVaAlProductor: Story = {
  args: { onSelect: fn() },
  play: async ({ args, canvasElement }) => {
    canvasElement.querySelector<HTMLButtonElement>(".artchip")?.click()
    await expect(args.onSelect).toHaveBeenCalledWith("spec-writer")
  },
}
