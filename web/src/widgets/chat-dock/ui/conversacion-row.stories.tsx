import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import type { Conversacion } from "@/shared"
import { ConversacionRow } from "./conversacion-row"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel, T24.
// B-01…B-11 de `plan-storybook.md` §2.2. Ninguna baja `a11y` a "todo" (RF-353 CA-4).
//
// El gesto de renombrado se mide contra el del rail (`session-rail.tsx:256-273`): B-08…B-11
// son literalmente los cuatro caminos de ese gesto — confirmar, descartar por vacío, descartar
// por Escape, confirmar por blur. Divergir del rail sería tener dos gestos para lo mismo.

const ACTIVA: Conversacion = {
  id: "c1",
  titulo: "por qué el mapa sale vacío",
  titulo_editado: false,
  activa: true,
  turnos: 90,
  ctx_pct: 68,
  rotacion_pendiente: false,
  ultima_interaccion: "2026-07-26T18:02:00Z",
  creada_en: "2026-07-24T10:00:00Z",
  claude_session_id: "4b046945-1f0e-4c0a-9a51-1b2f9c8d7e30",
  model: "claude-opus-5[1m]",
}

const meta = {
  title: "widgets/chat-dock/ConversacionRow",
  component: ConversacionRow,
  parameters: { layout: "centered" },
  args: {
    activa: ACTIVA,
    arnes: "vitalia",
    cwd: "~/Proyectos/luana-vitalia/vitalia",
    status: "idle",
    listaAbierta: false,
    detalleForzado: false,
    panelId: "cv-panel",
    onToggleLista: fn(),
    onNueva: fn(),
    onRenombrar: fn(),
  },
  decorators: [
    (Story) => (
      <div style={{ width: 320 }} className="border border-border bg-card">
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof ConversacionRow>

export default meta
type Story = StoryObj<typeof meta>

// B-01 · reposo: los cinco controles con su nombre accesible. El ✎ es superset declarado —
// el dibujo describe el gesto pero no dibuja su disparador (PARIDAD §2).
export const Reposo: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: ACTIVA.titulo })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Renombrar la conversación" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 68/ })).toBeInTheDocument()
    await expect(
      c.getByRole("button", { name: "Buscar en las conversaciones de esta sesión" }),
    ).toBeInTheDocument()
    await expect(
      c.getByRole("button", { name: "Nueva conversación — desactiva la actual" }),
    ).toBeEnabled()
  },
}

// B-02 · con la lista desplegada el botón del título lo DECLARA (no sólo rota el chevron).
//
// La story monta un contenedor con el `id` del panel, y no es decorado: axe exige que un
// `aria-controls` con `aria-expanded="true"` apunte a un elemento que EXISTA
// (`aria-valid-attr-value`). Sin el stub, esta story medía una composición imposible. Lo
// cazó el gate en la primera corrida.
export const ListaAbierta: Story = {
  args: { listaAbierta: true },
  decorators: [
    (Story) => (
      <div style={{ width: 320 }} className="border border-border bg-card">
        <Story />
        <div id="cv-panel" />
      </div>
    ),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const titulo = c.getByRole("button", { name: ACTIVA.titulo })
    await expect(titulo).toHaveAttribute("aria-expanded", "true")
    await expect(titulo).toHaveAttribute("aria-controls", "cv-panel")
  },
}

// B-03 · título largo: trunca en pantalla, completo en el `title` nativo.
export const TituloLargo: Story = {
  args: {
    activa: {
      ...ACTIVA,
      titulo:
        "por qué el manifiesto declara cero elementos cuando el árbol tiene doce y el loader corta en el primer separador",
    },
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const titulo = c.getByRole("button", { name: args.activa.titulo })
    await expect(titulo).toHaveAttribute("title", args.activa.titulo)
    const texto = titulo.querySelector(".truncate") as HTMLElement
    await expect(texto.scrollWidth).toBeGreaterThan(texto.clientWidth)
  },
}

// B-04 · abrir por el chevron ⇒ el foco va a las FILAS (C-10, RF-355).
export const AbrirPorChevron: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: ACTIVA.titulo }))
    await expect(args.onToggleLista).toHaveBeenCalledWith("filas")
  },
}

// B-05 · abrir por la lupa ⇒ el foco va al BUSCADOR. Dos controles, dos intenciones.
export const AbrirPorLupa: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(
      c.getByRole("button", { name: "Buscar en las conversaciones de esta sesión" }),
    )
    await expect(args.onToggleLista).toHaveBeenCalledWith("buscador")
  },
}

// B-06 · ＋ crea (E-06). Una sola llamada: la transición es una, no dos encadenadas.
export const CrearNueva: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(
      c.getByRole("button", { name: "Nueva conversación — desactiva la actual" }),
    )
    await expect(args.onNueva).toHaveBeenCalledTimes(1)
  },
}

// B-07 · con un turno en vuelo el ＋ queda deshabilitado CON MOTIVO (C-1, E-07). Nunca un
// control apagado y mudo: el servidor responde 409 y un click que se veía legal es peor UX.
export const Bloqueada: Story = {
  args: { status: "streaming" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const mas = c.getByRole("button", { name: "Nueva conversación — desactiva la actual" })
    await expect(mas).toBeDisabled()
    await expect(mas).toHaveAttribute("title", "esperá a que termine el turno en vuelo")
    await userEvent.click(mas, { pointerEventsCheck: 0 })
    await expect(args.onNueva).not.toHaveBeenCalled()
    // El buscador NO se bloquea: leer lo que ya se dijo no compite con el turno (E-28).
    await expect(
      c.getByRole("button", { name: "Buscar en las conversaciones de esta sesión" }),
    ).toBeEnabled()
  },
}

// B-08 · renombrar y confirmar con Enter. Y la otra mitad del criterio: mientras se edita, el
// input toma la fila ENTERA — el ctx y las acciones no compiten con el cursor (RF-331 CA-2).
export const RenombrarConfirma: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Renombrar la conversación" }))
    const input = c.getByRole("textbox", { name: "Título de la conversación" })
    await expect(c.queryByRole("button", { name: /contexto/ })).not.toBeInTheDocument()
    await expect(c.queryByRole("button", { name: /Nueva conversación/ })).not.toBeInTheDocument()
    await userEvent.clear(input)
    await userEvent.type(input, "el bug del índice")
    await userEvent.keyboard("{Enter}")
    await expect(args.onRenombrar).toHaveBeenCalledWith("el bug del índice")
  },
}

// B-09 · vacío DESCARTA, no borra (E-30, RF-331 CA-3). Es lo que ya hace el rail.
export const RenombrarVacio: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Renombrar la conversación" }))
    await userEvent.clear(c.getByRole("textbox", { name: "Título de la conversación" }))
    await userEvent.keyboard("{Enter}")
    await expect(args.onRenombrar).not.toHaveBeenCalled()
    await expect(c.getByRole("button", { name: ACTIVA.titulo })).toBeInTheDocument()
  },
}

// B-10 · Escape descarta y DEVUELVE EL FOCO al botón del título (E-32, design.md §8.3).
export const RenombrarEscape: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Renombrar la conversación" }))
    await userEvent.type(c.getByRole("textbox", { name: "Título de la conversación" }), " y algo")
    await userEvent.keyboard("{Escape}")
    await expect(args.onRenombrar).not.toHaveBeenCalled()
    const titulo = c.getByRole("button", { name: ACTIVA.titulo })
    await expect(titulo).toBeInTheDocument()
    await expect(titulo).toHaveFocus()
  },
}

// B-11 · blur CONFIRMA (E-34). Divergir acá sería que el mismo gesto signifique cosas
// distintas en el rail y en el dock.
export const RenombrarBlur: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Renombrar la conversación" }))
    const input = c.getByRole("textbox", { name: "Título de la conversación" })
    await userEvent.clear(input)
    await userEvent.type(input, "sellar el arnés")
    await userEvent.tab()
    await expect(args.onRenombrar).toHaveBeenCalledWith("sellar el arnés")
  },
}
