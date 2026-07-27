import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, userEvent, within } from "storybook/test"
import type { Conversacion, ConversacionActiva, Session } from "@/shared"
import { useSessions } from "@/shared"
import { useConversaciones } from "../model/conversaciones-store"
import { ChatDock } from "./chat-dock"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel.
//
// A-01…A-03 nacieron en T2 como BASELINE del dock vigente (H-1): el widget que este paquete
// reescribe entero no tenía ni una story. Sobreviven acá **ajustadas al superset**: ya no
// miden las 4 filas de cromo (el tramo 3 las bajó a 2) sino lo que NO se podía tocar — el
// composer, las burbujas, la tarjeta de actividad y la tarjeta de permiso. Que sigan verdes
// ES la prueba de que la mudanza no rompió superficie firmada de HS-26.
//
// 🟢 N-1 CERRADO. El baseline de T2 tuvo que apagar la regla `color-contrast` de axe porque
// `SessionLine` pintaba el `◍ <cc-id>` con `text-primary` sobre `bg-secondary` — 2,21:1
// medido, contra el 4,5 de texto. Esa fila se retiró y su cuerpo se mudó a `IdentidadDetalle`
// con el texto en `--foreground` (T23). **Ninguna story de este archivo apaga ya ninguna
// regla**: el gate a11y corre entero sobre todo el dock, contraste incluido (RF-353 CA-4).

const CWD = "~/Proyectos/luana-vitalia/vitalia"

// Los datos salen de la máquina real (`GET /api/sessions` verificado 2026-07-26): la sesión
// `s6165ac75` del arnés `vitalia`, sus 4 conversaciones. No son inventados.
const ACTIVA: ConversacionActiva = {
  id: "c-mapa",
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
  conv: [
    { rol: "user", text: "el manifiesto declara 0 elementos, ¿por qué?" },
    { rol: "act", text: "Read skills/builder/SKILL.md" },
    { rol: "assistant", text: "El loader corta en el primer separador." },
  ],
}

const OTRAS: Conversacion[] = [
  { ...ACTIVA, fragmento: undefined },
  {
    id: "c-sellar",
    titulo: "sellar el arnés con arnes.l0.json",
    titulo_editado: false,
    activa: false,
    turnos: 41,
    ctx_pct: 34,
    rotacion_pendiente: false,
    ultima_interaccion: "2026-07-25T18:02:00Z",
    creada_en: "2026-07-25T09:00:00Z",
    claude_session_id: "d303a93f-2b1c-4f0a-8e51-0c2f9c8d7e31",
  },
  {
    id: "c-dictado",
    titulo: "prueba de dictado",
    titulo_editado: false,
    activa: false,
    turnos: 6,
    ctx_pct: 4,
    rotacion_pendiente: false,
    ultima_interaccion: "2026-07-24T11:00:00Z",
    creada_en: "2026-07-24T11:00:00Z",
  },
  {
    id: "c-nuevo",
    titulo: "nuevo frente",
    titulo_editado: false,
    activa: false,
    turnos: 0,
    ctx_pct: 0,
    rotacion_pendiente: false,
    creada_en: "2026-07-23T12:00:00Z",
  },
]

const SESION: Session = {
  id: "s6165ac75",
  frente: "repro del bug de carga",
  arnes: "vitalia",
  status: "idle",
  view: "Mapa",
  cwd: CWD,
  activa: ACTIVA,
}

type EstadoSesiones = ReturnType<typeof useSessions.getState>
type EstadoPanel = ReturnType<typeof useConversaciones.getState>

// escenario deja los DOS stores en un estado completo antes de renderizar. Calca
// dictado-button.stories.tsx:16-29. Regla dura: se escribe el estado ENTERO, jamás un merge
// parcial sobre lo que dejó la story anterior — es lo que hace que el orden no importe.
function escenario(parcial: Partial<EstadoSesiones>, panel: Partial<EstadoPanel> = {}) {
  return (Story: () => ReactNode) => {
    useSessions.setState({
      sessions: [],
      activeId: null,
      chatOpen: true,
      railCollapsed: false,
      connected: true,
      loaded: true,
      streaming: {},
      msgFlushed: {},
      finalizedRun: {},
      pendingPerms: {},
      scope: {},
      wroteInRun: {},
      convRev: {},
      detalleForzado: {},
      desactivadaTitulo: {},
      ...parcial,
    })
    useConversaciones.setState({
      sesionId: null,
      estado: "datos",
      error: undefined,
      conversaciones: [],
      total: 0,
      busqueda: "",
      abierta: false,
      focoInicial: "filas",
      retomando: null,
      fallo: undefined,
      ...panel,
    })
    return (
      <div style={{ width: 360, height: 520 }} className="border border-border bg-card">
        {Story()}
      </div>
    )
  }
}

const conPanel: Partial<EstadoPanel> = {
  sesionId: SESION.id,
  abierta: true,
  estado: "datos",
  conversaciones: OTRAS,
  total: 4,
}

const meta = {
  title: "widgets/chat-dock/ChatDock",
  component: ChatDock,
  parameters: { layout: "centered" },
} satisfies Meta<typeof ChatDock>

export default meta
type Story = StoryObj<typeof meta>

// A-01 · lo que el superset NO podía tocar: el composer con su placeholder, las burbujas y la
// tarjeta de actividad agrupada. Antes medía además las 4 filas de cromo; ahora mide 2.
export const VigenteReposo: Story = {
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("repro del bug de carga")).toBeInTheDocument()
    await expect(c.getByPlaceholderText("Pídele un cambio a vitalia")).toBeInTheDocument()
    await expect(c.getByText(/el manifiesto declara 0 elementos/)).toBeInTheDocument()
    await expect(c.getByText("1 paso")).toBeInTheDocument()
  },
}

// A-02 · la tarjeta de permiso sigue renderizando dentro del transcript (RF-113, HS-26).
export const VigenteConPermiso: Story = {
  decorators: [
    escenario({
      sessions: [{ ...SESION, status: "await" }],
      activeId: SESION.id,
      pendingPerms: {
        [SESION.id]: [
          {
            request_id: "cr-1",
            tool: "Edit",
            input: { file_path: "skills/builder/SKILL.md", old_string: "a", new_string: "b" },
          },
        ],
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByText("skills/builder/SKILL.md").length).toBeGreaterThanOrEqual(1)
    await expect(c.getByRole("button", { name: "Permitir esta sesión" })).toBeInTheDocument()
    await expect(c.getByPlaceholderText(/esperando tu decisión de permiso/)).toBeInTheDocument()
  },
}

// A-03 · el `■` de interrumpir reemplaza al `↑`. Es el mismo botón en dos estados y el
// paquete no lo toca.
export const VigenteStreaming: Story = {
  decorators: [
    escenario({
      sessions: [{ ...SESION, status: "streaming" }],
      activeId: SESION.id,
      streaming: { [SESION.id]: "revisando el loader" },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "■" })).toBeInTheDocument()
    await expect(c.queryByRole("button", { name: "↑" })).not.toBeInTheDocument()
    await expect(c.getByText("revisando el loader")).toBeInTheDocument()
  },
}

// A-04 · EL assert del tramo: en reposo y sin nodo, EXACTAMENTE 2 filas de cromo. Se mide por
// ausencia de lo que se retiró (el «Alcance:» y el `◍` de la SessionLine) y presencia de lo
// que llegó (el chip de ctx y el glifo nuevo).
export const Reposo: Story = {
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("Alcance:")).not.toBeInTheDocument()
    await expect(c.queryByText(/selecciona un nodo en el Mapa/)).not.toBeInTheDocument()
    await expect(c.queryByText(/◍/)).not.toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 68/ })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "» colapsar" })).toBeInTheDocument()
    await expect(c.queryByRole("button", { name: "⟩ colapsar" })).not.toBeInTheDocument()
    await expect(c.getByRole("button", { name: "por qué el mapa sale vacío" })).toBeInTheDocument()
  },
}

// A-05 · el mismo assert estructural en OSCURO (RF-353).
export const ReposoDark: Story = {
  globals: { theme: "dark" },
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("Alcance:")).not.toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 68/ })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "» colapsar" })).toBeInTheDocument()
  },
}

// A-06 · con un nodo elegido la fila 3 aparece y se gana el lugar: cambia qué le estás
// pidiendo. El chip es LITERAL el vigente, ✕ incluido; lo que no vuelve es el rótulo.
export const ConNodoEnAlcance: Story = {
  decorators: [
    escenario({
      sessions: [SESION],
      activeId: SESION.id,
      scope: {
        [SESION.id]: { nodeId: "hipaa-check", clase: "caja", fuentePath: "/hipaa-check" },
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("hipaa-check")).toBeInTheDocument()
    await expect(c.getByText("/hipaa-check")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "✕" })).toHaveAttribute(
      "title",
      "quitar del alcance",
    )
    await expect(c.queryByText("Alcance:")).not.toBeInTheDocument()
  },
}

// A-07 · el detalle desplegado trae los 4 datos de la SessionLine retirada MÁS el cwd, que
// hasta hoy no se veía en ninguna superficie (RF-328).
export const DetalleDesplegado: Story = {
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 68/ })
    await userEvent.click(chip)
    await expect(chip).toHaveAttribute("aria-expanded", "true")
    await expect(c.getByText(/◍ 4b046945/)).toBeInTheDocument()
    await expect(c.getByText(CWD)).toBeInTheDocument()
    await expect(c.getByText("claude-opus-5[1m]")).toBeInTheDocument()
  },
}

// A-08 · la lista abre EN SITIO y toma el área del transcript — el composer y el cromo siguen
// visibles. Cero `<dialog>`, cero backdrop, cero portal (RF-317).
export const ListaAbierta: Story = {
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id }, conPanel)],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("listbox")).toBeInTheDocument()
    await expect(c.getAllByRole("option")).toHaveLength(4)
    await expect(c.getByPlaceholderText("Pídele un cambio a vitalia")).toBeInTheDocument()
    await expect(c.getByText("repro del bug de carga")).toBeInTheDocument()
    await expect(canvasElement.querySelector("dialog")).toBeNull()
  },
}

// A-09 · recién creada: el vacío NOMBRA la que se desactivó y dice cómo volver (RF-307). El
// copy vigente («Esta conversación ES la sesión Claude Code del frente…») dejó de ser verdad
// bajo CV-D3 y no aparece. El chip a 0 % se pinta igual (BR-CV-9).
export const ConversacionRecienCreada: Story = {
  decorators: [
    escenario({
      sessions: [
        {
          ...SESION,
          activa: {
            ...ACTIVA,
            id: "c-nueva",
            titulo: "nueva conversación",
            turnos: 0,
            ctx_pct: 0,
            claude_session_id: undefined,
            model: undefined,
            ultima_interaccion: undefined,
            conv: [],
          },
        },
      ],
      activeId: SESION.id,
      desactivadaTitulo: { [SESION.id]: "por qué el mapa sale vacío" },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/quedó guardada — la retomás desde ▶/)).toBeInTheDocument()
    await expect(c.queryByText(/Esta conversación ES la sesión/)).not.toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 0%/ })).toBeInTheDocument()
  },
}

// A-10 · la PRIMERA conversación de una sesión: mismo copy SIN la segunda oración. No hay
// anterior que nombrar y no se inventa una.
export const PrimeraConversacionDeLaSesion: Story = {
  decorators: [
    escenario({
      sessions: [
        {
          ...SESION,
          activa: { ...ACTIVA, id: "c-1a", titulo: "nueva conversación", turnos: 0, conv: [] },
        },
      ],
      activeId: SESION.id,
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Pídele un cambio a/)).toBeInTheDocument()
    await expect(c.queryByText(/quedó guardada/)).not.toBeInTheDocument()
  },
}

// A-11 · retomando: la franja efímera con el `--resume` real, y el detalle abierto solo — al
// retomar, el cc-id pasa a ser el de ESTA conversación y eso hay que verlo sin buscarlo.
export const Retomando: Story = {
  decorators: [
    escenario(
      { sessions: [SESION], activeId: SESION.id, detalleForzado: { [SESION.id]: true } },
      { ...conPanel, retomando: "c-sellar" },
    ),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status")).toHaveTextContent("Retomando la conversación…")
    await expect(c.getByText("--resume d303a93f")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 68/ })).toHaveAttribute(
      "aria-expanded",
      "true",
    )
  },
}

// A-12 · retomar una de 0 turnos: no hay proceso viejo, así que el `--resume` NO se dibuja
// (E-16). La franja sí — el operador pidió algo y algo está pasando.
export const RetomandoSinResume: Story = {
  decorators: [
    escenario({ sessions: [SESION], activeId: SESION.id }, { ...conPanel, retomando: "c-nuevo" }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status")).toHaveTextContent("Retomando la conversación…")
    await expect(c.queryByText(/--resume/)).not.toBeInTheDocument()
  },
}

// A-13 · la retoma falló: `role="alert"` con el motivo TAL CUAL, sobre `--crit-soft` con el
// texto en `--foreground` (C-11 — en `--warn` daría 3,19:1).
export const RetomaFallida: Story = {
  decorators: [
    escenario(
      { sessions: [SESION], activeId: SESION.id },
      { ...conPanel, fallo: "la sesión de Claude Code ya no existe — reinicié el hilo" },
    ),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent(
      "la sesión de Claude Code ya no existe — reinicié el hilo",
    )
  },
}

// A-14 · rotación: UNA marca inline, no un corte de hilo (CV-D10). El texto es el del código
// —el que ya está persistido en los transcripts vivos—, no el del dibujo (C-5). El ctx bajó y
// el detalle se abrió solo, porque el cc-id es otro.
export const TranscriptConRotacion: Story = {
  decorators: [
    escenario({
      sessions: [
        {
          ...SESION,
          activa: {
            ...ACTIVA,
            ctx_pct: 12,
            claude_session_id: "7ae10c42-9d3e-4b0a-8c51-1f2a9c8d7e33",
            conv: [
              { rol: "assistant", text: "…entonces el índice quedó vacío por eso." },
              { rol: "sys", text: "— contexto rotado, seguimos —" },
              { rol: "user", text: "seguí con el sellado" },
            ],
          },
        },
      ],
      activeId: SESION.id,
      detalleForzado: { [SESION.id]: true },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("— contexto rotado, seguimos —")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 12/ })).toHaveAttribute(
      "aria-expanded",
      "true",
    )
    await expect(c.getByText(/◍ 7ae10c42/)).toBeInTheDocument()
  },
}

// A-15 · turno en vuelo: el ＋ queda deshabilitado CON MOTIVO y el 🔍 sigue habilitado —
// leer lo que ya se dijo no compite con el turno (E-07, E-28).
export const TurnoEnVuelo: Story = {
  decorators: [
    escenario({
      sessions: [{ ...SESION, status: "streaming" }],
      activeId: SESION.id,
      streaming: { [SESION.id]: "revisando" },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const mas = c.getByRole("button", { name: "Nueva conversación — desactiva la actual" })
    await expect(mas).toBeDisabled()
    await expect(mas).toHaveAttribute("title", "esperá a que termine el turno en vuelo")
    await expect(
      c.getByRole("button", { name: "Buscar en las conversaciones de esta sesión" }),
    ).toBeEnabled()
  },
}

// A-16 · esperando una decisión de permiso: mismo bloqueo, OTRO motivo. La tarjeta sigue.
export const PermisoPendiente: Story = {
  decorators: [
    escenario({
      sessions: [{ ...SESION, status: "await" }],
      activeId: SESION.id,
      pendingPerms: { [SESION.id]: [{ request_id: "cr-9", tool: "Bash", input: {} }] },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const mas = c.getByRole("button", { name: "Nueva conversación — desactiva la actual" })
    await expect(mas).toBeDisabled()
    await expect(mas).toHaveAttribute("title", "esperá tu decisión de permiso")
    await expect(c.getByRole("button", { name: "Permitir esta sesión" })).toBeInTheDocument()
  },
}

// A-17 · la lista abierta en OSCURO (RF-353).
export const ListaAbiertaDark: Story = {
  globals: { theme: "dark" },
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id }, conPanel)],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("listbox")).toBeInTheDocument()
    await expect(c.getAllByRole("option")).toHaveLength(4)
    await expect(c.getByPlaceholderText("Pídele un cambio a vitalia")).toBeInTheDocument()
  },
}
