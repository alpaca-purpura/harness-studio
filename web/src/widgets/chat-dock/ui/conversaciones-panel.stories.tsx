import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import type { Conversacion } from "@/shared"
import { ConversacionesPanel } from "./conversaciones-panel"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel, T28.
// D-01…D-14 de `plan-storybook.md` §2.4. Ninguna baja `a11y` a "todo" (RF-353 CA-4).

const AHORA = new Date(2026, 6, 26, 20, 0).getTime()
const iso = (d: number, h: number, m: number) => new Date(2026, 6, d, h, m).toISOString()

const CUATRO: Conversacion[] = [
  {
    id: "c-mapa",
    titulo: "por qué el mapa sale vacío",
    titulo_editado: false,
    activa: true,
    turnos: 90,
    ctx_pct: 68,
    rotacion_pendiente: false,
    ultima_interaccion: new Date(AHORA - 4 * 60000).toISOString(),
    creada_en: iso(24, 10, 0),
  },
  {
    id: "c-sellar",
    titulo: "sellar el arnés con arnes.l0.json",
    titulo_editado: false,
    activa: false,
    turnos: 41,
    ctx_pct: 34,
    rotacion_pendiente: false,
    ultima_interaccion: iso(25, 18, 2),
    creada_en: iso(25, 9, 0),
  },
  {
    id: "c-dictado",
    titulo: "prueba de dictado",
    titulo_editado: false,
    activa: false,
    turnos: 6,
    ctx_pct: 4,
    rotacion_pendiente: false,
    ultima_interaccion: iso(24, 11, 0),
    creada_en: iso(24, 11, 0),
  },
  {
    id: "c-nuevo",
    titulo: "nuevo frente",
    titulo_editado: false,
    activa: false,
    turnos: 0,
    ctx_pct: 0,
    rotacion_pendiente: false,
    creada_en: iso(23, 12, 0),
  },
]

const meta = {
  title: "widgets/chat-dock/ConversacionesPanel",
  component: ConversacionesPanel,
  parameters: { layout: "centered" },
  args: {
    id: "cv-panel",
    estado: "datos",
    conversaciones: CUATRO,
    total: 4,
    frenteSesion: "repro del bug de carga",
    busqueda: "",
    focoInicial: "filas",
    ahora: AHORA,
    onBusqueda: fn(),
    onElegir: fn(),
    onCancelar: fn(),
    onReintentar: fn(),
  },
  decorators: [
    (Story) => (
      <div
        style={{ width: 340, height: 380 }}
        className="flex flex-col border border-border bg-card"
      >
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ConversacionesPanel>

export default meta
type Story = StoryObj<typeof meta>

// D-01 · cargando: el esqueleto anuncia que está cargando y NO hay ninguna opción. Calca el
// contrato ARIA de `PickerSkeleton` sin arrastrar sus clases `pf-*` (C-4).
export const Cargando: Story = {
  args: { estado: "cargando", conversaciones: [], total: 0 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status")).toHaveAttribute("aria-label", "Cargando conversaciones")
    await expect(c.queryAllByRole("option")).toHaveLength(0)
  },
}

// D-02 · error: el motivo TAL CUAL y un reintento. Jamás «0 conversaciones» — no poder
// preguntar y no tener ninguna son cosas distintas (BR-CV-10, RF-347).
export const ErrorDeLectura: Story = {
  args: { estado: "error", error: "daemon no responde", conversaciones: [], total: 0 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("daemon no responde")
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeInTheDocument()
    await expect(c.queryByText(/0 conversaciones/)).not.toBeInTheDocument()
  },
}

// D-03 · reintentar vuelve a preguntar.
export const ErrorReintentar: Story = {
  args: { estado: "error", error: "daemon no responde", conversaciones: [], total: 0 },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onReintentar).toHaveBeenCalledTimes(1)
  },
}

// D-04 · cuatro conversaciones: el rótulo NOMBRA la sesión (CV-D4 — el alcance se dice, no se
// deduce), hay 4 opciones y el buscador existe.
export const Cuatro: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/4 conversaciones de la sesión/)).toBeInTheDocument()
    await expect(c.getByText("«repro del bug de carga»")).toBeInTheDocument()
    await expect(c.getAllByRole("option")).toHaveLength(4)
    await expect(c.getByRole("searchbox")).toBeInTheDocument()
    // Orden: última interacción desc, y la de 0 turnos AL FINAL (RF-320 CA-2).
    const titulos = c.getAllByRole("option").map((o) => o.textContent ?? "")
    await expect(titulos[0]).toContain("por qué el mapa sale vacío")
    await expect(titulos[3]).toContain("nuevo frente")
  },
}

// D-05 · una sola: SIN buscador (nada que filtrar, RF-325 CA-2) y con el mensaje que explica
// qué va a pasar cuando haya otra (E-02, E-21).
export const UnaSola: Story = {
  args: { conversaciones: [CUATRO[0] as Conversacion], total: 1 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByRole("searchbox")).not.toBeInTheDocument()
    await expect(c.getByText(/1 conversación de la sesión/)).toBeInTheDocument()
    await expect(c.getByText(/Esta sesión recién arranca/)).toBeInTheDocument()
  },
}

// D-06 · buscando: el rótulo dice «N de M» —M es el total de la SESIÓN, no el de
// coincidencias— y cada fila muestra POR QUÉ apareció (RF-321, RF-322).
export const BuscandoConCoincidencias: Story = {
  args: {
    busqueda: "manifiesto",
    total: 4,
    conversaciones: [
      { ...(CUATRO[0] as Conversacion), fragmento: "…el manifiesto declara 0 elementos…" },
      { ...(CUATRO[1] as Conversacion), fragmento: "…escribí el manifiesto en la raíz…" },
    ],
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("2 de 4 coinciden")).toBeInTheDocument()
    await expect(c.getAllByRole("option")).toHaveLength(2)
    await expect(canvasElement.querySelectorAll("mark")).toHaveLength(2)
  },
}

// D-07 · la activa participa de la búsqueda como cualquier otra y conserva sus marcas (E-24).
export const BuscandoConLaActiva: Story = {
  args: {
    busqueda: "manifiesto",
    total: 4,
    conversaciones: [
      { ...(CUATRO[0] as Conversacion), fragmento: "…el manifiesto declara 0 elementos…" },
      { ...(CUATRO[1] as Conversacion), fragmento: "…escribí el manifiesto…" },
    ],
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const activa = c.getAllByRole("option")[0]
    await expect(activa).toHaveAttribute("aria-selected", "true")
    await expect(within(activa as HTMLElement).getByText("activa")).toBeInTheDocument()
  },
}

// D-08 · sin coincidencias: el vacío dice QUÉ se buscó, DÓNDE y CUÁNTAS se miraron. Un «sin
// resultados» pelado dejaría al operador con la misma duda que originó el paquete (E-22).
export const SinCoincidencias: Story = {
  args: { busqueda: "telemetría", conversaciones: [], total: 4 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("«telemetría»")).toBeInTheDocument()
    await expect(c.getByText(/Se buscó en el título y en el texto de las 4\./)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Limpiar búsqueda" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Cancelar" })).toBeInTheDocument()
  },
}

// D-09 · limpiar la búsqueda vuelve a la lista entera.
export const LimpiarBusqueda: Story = {
  args: { busqueda: "telemetría", conversaciones: [], total: 4 },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Limpiar búsqueda" }))
    await expect(args.onBusqueda).toHaveBeenCalledWith("")
  },
}

// D-10 · turno en vuelo: las filas quedan `aria-disabled` con motivo, pero el BUSCADOR sigue
// activo — leer lo que ya se dijo no compite con el turno (E-11, E-28).
export const Bloqueado: Story = {
  args: { bloqueadoMotivo: "esperá tu decisión de permiso" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    for (const o of c.getAllByRole("option")) {
      await expect(o).toHaveAttribute("aria-disabled", "true")
    }
    await expect(c.getByRole("searchbox")).toBeEnabled()
    await userEvent.click(c.getAllByRole("option")[1] as HTMLElement)
    await expect(args.onElegir).not.toHaveBeenCalled()
  },
}

// D-11 · elegir una inactiva la retoma, con SU id exacto (RF-310, E-10).
export const ElegirInactiva: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getAllByRole("option")[1] as HTMLElement)
    await expect(args.onElegir).toHaveBeenCalledWith("c-sellar")
  },
}

// D-12 · click en la ACTIVA: emite igual. El no-op lo resuelve el dominio (RF-316/E-15), no
// una rama del FE que adivine si hace falta.
export const ClickEnLaActiva: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getAllByRole("option")[0] as HTMLElement)
    await expect(args.onElegir).toHaveBeenCalledWith("c-mapa")
  },
}

// D-13 · teclado: ↓ ↓ Enter elige la TERCERA fila, y el `aria-activedescendant` fue contando
// (RF-354). El cursor arranca en la activa, que es de donde el operador viene.
export const TecladoFlechas: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const lista = c.getByRole("listbox")
    await expect(lista).toHaveAttribute("aria-activedescendant", "cv-panel-c-mapa")
    lista.focus()
    await userEvent.keyboard("{ArrowDown}{ArrowDown}")
    await expect(lista).toHaveAttribute("aria-activedescendant", "cv-panel-c-dictado")
    await userEvent.keyboard("{Enter}")
    await expect(args.onElegir).toHaveBeenCalledWith("c-dictado")
  },
}

// D-14 · Escape en el buscador cierra la lista (RF-317, RF-354).
export const EscapeCierra: Story = {
  args: { focoInicial: "buscador" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("searchbox")).toHaveFocus()
    await userEvent.keyboard("{Escape}")
    await expect(args.onCancelar).toHaveBeenCalledTimes(1)
  },
}
