import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import type { Conversacion } from "@/shared"
import { ConversacionFila } from "./conversacion-fila"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel, T27.
// E-01s…E-09s de `plan-storybook.md` §2.5. Ninguna baja `a11y` a "todo" (RF-353 CA-4).
//
// Las fechas se construyen en hora LOCAL y el instante de referencia se INYECTA: así los
// cuatro formatos del dibujo dan lo mismo en la máquina del operador y en el runner de CI,
// que es lo que un `Date.now()` implícito no garantiza.

const AHORA = new Date(2026, 6, 26, 20, 0).getTime()
const HACE_4_MIN = new Date(AHORA - 4 * 60000).toISOString()
const AYER_18_02 = new Date(2026, 6, 25, 18, 2).toISOString()

const BASE: Conversacion = {
  id: "c-sellar",
  titulo: "sellar el arnés con arnes.l0.json",
  titulo_editado: false,
  activa: false,
  turnos: 41,
  ctx_pct: 34,
  rotacion_pendiente: false,
  ultima_interaccion: AYER_18_02,
  creada_en: new Date(2026, 6, 25, 9, 0).toISOString(),
}

const meta = {
  title: "widgets/chat-dock/ConversacionFila",
  component: ConversacionFila,
  parameters: { layout: "centered" },
  args: { id: "cv-1", conversacion: BASE, ahora: AHORA, onElegir: fn() },
  // La fila ES un `role="option"`: fuera de un listbox axe la rechaza
  // (`aria-required-parent`), y con razón — una opción suelta no es una opción de nada.
  decorators: [
    (Story) => (
      <div role="listbox" aria-label="Conversaciones" style={{ width: 320 }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ConversacionFila>

export default meta
type Story = StoryObj<typeof meta>

// E-01s · inactiva: los CUATRO datos de CV-D13 y ninguno más. Sin rótulo `activa`.
export const Inactiva: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(BASE.titulo)).toBeInTheDocument()
    await expect(c.getByText(/^ayer \d\d:\d\d · 41 turnos · ctx 34 %$/)).toBeInTheDocument()
    await expect(c.queryByText("activa")).not.toBeInTheDocument()
    await expect(c.getByRole("option")).toHaveAttribute("aria-selected", "false")
  },
}

// E-02s · activa: el rótulo, el `aria-selected` Y el radio relleno. Tres señales, ninguna
// cromática sola — el borde `--primary` no llega a 3:1 en claro (RF-356).
export const Activa: Story = {
  args: { conversacion: { ...BASE, activa: true, ultima_interaccion: HACE_4_MIN, turnos: 90 } },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("activa")).toBeInTheDocument()
    await expect(c.getByRole("option")).toHaveAttribute("aria-selected", "true")
    await expect(c.getByText(/hace 4 min · 90 turnos/)).toBeInTheDocument()
    await expect(canvasElement.querySelector(".bg-primary")).toBeInTheDocument()
  },
}

// E-03s · 0 turnos: se DICE que no hubo ninguno. Cero no se esconde y la fecha no se inventa
// (BR-CV-9, BR-CV-14, E-04).
export const SinTurnos: Story = {
  args: {
    conversacion: {
      ...BASE,
      titulo: "nuevo frente",
      turnos: 0,
      ctx_pct: 0,
      ultima_interaccion: undefined,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin turnos todavía · ctx 0 %")).toBeInTheDocument()
    await expect(c.queryByText(/jul|ayer|hace/)).not.toBeInTheDocument()
  },
}

// E-04s · migrada (H-4): tiene turnos pero el registro viejo no guardaba la última
// interacción. Dice «sin fecha» — derivarla del `creada_en` sería inventarla.
export const SinFecha: Story = {
  args: { conversacion: { ...BASE, ultima_interaccion: undefined } },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin fecha · 41 turnos · ctx 34 %")).toBeInTheDocument()
  },
}

// E-05s · el `<mark>` envuelve EXACTAMENTE el término, ni una letra más.
export const ConFragmento: Story = {
  args: {
    conversacion: {
      ...BASE,
      fragmento: "…escribí el manifiesto en la raíz y volvé a indexar…",
    },
    termino: "manifiesto",
  },
  play: async ({ canvasElement }) => {
    const marca = canvasElement.querySelector("mark")
    await expect(marca).toBeInTheDocument()
    await expect(marca?.textContent).toBe("manifiesto")
  },
}

// E-06s · el resaltado NO depende de igualdad literal: el daemon busca sin acentos ni
// mayúsculas y el FE pliega con la misma tabla. Buscar «vacio» resalta «vacío» (E-23).
export const ConFragmentoAcentuado: Story = {
  args: {
    conversacion: { ...BASE, fragmento: "…el índice quedó vacío después del reindex…" },
    termino: "vacio",
  },
  play: async ({ canvasElement }) => {
    const marca = canvasElement.querySelector("mark")
    await expect(marca?.textContent).toBe("vacío")
  },
}

// E-07s · coincidencia sólo en el título ⇒ ni `<mark>` ni tercera línea. No hay nada que
// explicar y un renglón vacío con sangría sería peor (RF-322 CA-3, E-27).
export const SinFragmento: Story = {
  args: { termino: "sellar" },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector("mark")).toBeNull()
    const c = within(canvasElement)
    await expect(c.getByText(BASE.titulo)).toBeInTheDocument()
  },
}

// E-08s · deshabilitada: `aria-disabled` (no `disabled`) para que el lector de pantalla la
// alcance y pueda LEER el motivo. Precedente `new-session-picker.tsx:362-363`.
export const Deshabilitada: Story = {
  args: { deshabilitadaMotivo: "esperá a que termine el turno en vuelo" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const fila = c.getByRole("option")
    await expect(fila).toHaveAttribute("aria-disabled", "true")
    await expect(fila).toHaveAttribute("title", "esperá a que termine el turno en vuelo")
    await userEvent.click(fila)
    await expect(args.onElegir).not.toHaveBeenCalled()
  },
}

// E-09s · la activa en OSCURO (RF-353).
export const ActivaDark: Story = {
  globals: { theme: "dark" },
  args: { conversacion: { ...BASE, activa: true } },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("activa")).toBeInTheDocument()
    await expect(c.getByRole("option")).toHaveAttribute("aria-selected", "true")
  },
}
