import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { ErrorBody, Skeleton } from "./estado-carga"

// Story = test (fe-visual-fitness, check `story-es-test`: cada componente de shared/ui tiene
// story). La promoción de T28 crea dos componentes en `shared/ui` que hoy no tenían story
// propia: acá se cementan su CONTRATO completo (D23), no solo el caso que el Portafolio usa.
//
// Frame reproduce el scope real del Portafolio para que las clases `pf-*` resuelvan con sus
// tokens — el refactor es de movimiento y el criterio de éxito es que nada cambie de aspecto.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const meta = {
  title: "shared/ui/EstadoCarga",
  component: Skeleton,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { label: "Cargando portafolio" },
} satisfies Meta<typeof Skeleton>

export default meta
type Story = StoryObj<typeof meta>

// H-6 · G5 — skeleton honesto: lo único anunciado es el label; las barras son decorativas y no
// se pueden confundir con filas de datos.
export const SkeletonHonesto: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const status = c.getByRole("status", { name: "Cargando portafolio" })
    await expect(status).toBeInTheDocument()
    await expect(status).toHaveAttribute("aria-live", "polite")
    // Cero filas fantasma: todas las barras son aria-hidden.
    const barras = canvasElement.querySelectorAll(".pf-skeleton-fila")
    await expect(barras).toHaveLength(3)
    for (const b of barras) await expect(b).toHaveAttribute("aria-hidden", "true")
    // …y ninguna aporta texto que pueda leerse como dato.
    await expect(status.textContent).toBe("")
  },
}

// D23 — el contrato completo, no solo el caso del Portafolio: `filas`, `altura` y `data` son
// las tres perillas que la capa Mejora necesita (2 tarjetas con la altura de una tarjeta).
export const SkeletonConMedidaPropia: Story = {
  args: { label: "midiendo…", filas: 2, altura: 96, data: "mejora" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status", { name: "midiendo…" })).toBeInTheDocument()
    const barras = canvasElement.querySelectorAll("[data-skeleton='mejora']")
    await expect(barras).toHaveLength(2)
    await expect((barras[0] as HTMLElement).style.height).toBe("96px")
  },
}

// H-6 — error con MOTIVO real (nunca «Error al obtener los datos») + reintento operable.
export const ErrorConMotivoReintentable: StoryObj<typeof ErrorBody> = {
  render: (args) => <ErrorBody {...args} />,
  args: {
    motivo: "No se pudo leer la telemetría — ECONNREFUSED.",
    onReintentar: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const alerta = c.getByRole("alert")
    // El motivo es el INYECTADO, no uno genérico del componente.
    await expect(alerta).toHaveTextContent("No se pudo leer la telemetría — ECONNREFUSED.")
    const btn = c.getByRole("button", { name: "Reintentar" })
    await expect(btn).toBeEnabled()
    await userEvent.click(btn)
    await expect(args.onReintentar).toHaveBeenCalledTimes(1)
  },
}
