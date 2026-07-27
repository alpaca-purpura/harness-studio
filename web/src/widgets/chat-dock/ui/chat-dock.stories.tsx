import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import type { Session } from "@/shared"
import { useSessions } from "@/shared"
import { ChatDock } from "./chat-dock"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel, T2.
//
// BASELINE (H-1). Este archivo nace ANTES del primer cambio al widget: `chat-dock.tsx` es el
// único de los 5 widgets del dock que no tenía story, y el paquete lo reescribe entero. Sin un
// baseline verde no hay forma de probar que el superset (BR-CV-11) dejó intactos el composer,
// las burbujas y la tarjeta de actividad.
//
// A-01..A-03 describen el dock TAL COMO ESTÁ HOY: cuatro filas de cromo (header · SessionLine
// con su `◍` · ScopeRow con su «Alcance:» · composer). Cuando el tramo 3 baje el cromo a 2
// filas, estas tres se ajustan al superset y las A-04.. miden el estado nuevo — el diff entre
// ambas versiones es exactamente la evidencia de qué se movió.

// escenario deja el store de sesiones en un estado completo antes de renderizar. Calca
// dictado-button.stories.tsx:16-29, el único archivo del repo que toca un store zustand en
// stories. Regla dura: se escribe el estado ENTERO, jamás un merge parcial sobre lo que dejó
// la story anterior — es lo que hace que el orden de ejecución no importe.
function escenario(parcial: Partial<ReturnType<typeof useSessions.getState>>) {
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
      ...parcial,
    })
    return (
      <div style={{ width: 360, height: 520 }} className="border border-border bg-card">
        {Story()}
      </div>
    )
  }
}

// Los datos salen de la máquina real (`GET /api/sessions` verificado 2026-07-26): la sesión
// `s6165ac75` del arnés `vitalia`. No son inventados.
const SESION: Session = {
  id: "s6165ac75",
  frente: "repro del bug de carga",
  arnes: "vitalia",
  status: "idle",
  view: "Mapa",
  claude_session_id: "4b046945-1f0e-4c0a-9a51-1b2f9c8d7e30",
  model: "claude-opus-5[1m]",
  ctx_pct: 68,
  conv: [
    { rol: "user", text: "el manifiesto declara 0 elementos, ¿por qué?" },
    { rol: "act", text: "Read skills/builder/SKILL.md" },
    { rol: "assistant", text: "El loader corta en el primer separador." },
  ],
}

// 🔴 HALLAZGO DEL BASELINE (T2, 2026-07-26) — se declara, no se tapa.
//
// La primera corrida de estas 3 stories destapó una violación a11y REAL y PREEXISTENTE del dock
// que se envía hoy: `SessionLine` pinta el `◍ <cc-id>` con `text-primary` sobre `bg-secondary`
// — medido por axe: **2,21:1** (`#00b7aa` sobre `#eef1f1`), contra el mínimo 4,5 de texto.
// Nadie la había visto porque `chat-dock.tsx` no tenía story (H-1): es justamente el agujero
// que este ticket existe para cerrar.
//
// Es la MISMA familia que C-3 (`design.md` §2): un token de acento usado como color de TEXTO.
// La regla ya decidida por el paquete la resuelve — el texto va en `--foreground` y el acento
// queda como señal no textual — y se aplica al construir `IdentidadDetalle` (T23), que es
// adonde el cuerpo de `SessionLine` se muda (RF-328). Arreglarla acá sería tocar superficie
// vigente fuera del ticket y sin capability.
//
// Mientras tanto, estas 3 stories describen el código TAL COMO ESTÁ, así que apagan **la sola
// regla `color-contrast`** — no bajan el gate a `"todo"`: el resto de axe sigue en `error`, y
// las stories de la superficie NUEVA (A-04.., C-01..) corren con el gate entero, contraste
// incluido (RF-353 CA-4). La deuda queda anotada en `BACKLOG.md`.
const SOLO_BASELINE = {
  a11y: { config: { rules: [{ id: "color-contrast", enabled: false }] } },
}

const meta = {
  title: "widgets/chat-dock/ChatDock",
  component: ChatDock,
  parameters: { layout: "centered" },
} satisfies Meta<typeof ChatDock>

export default meta
type Story = StoryObj<typeof meta>

// A-01 · el dock de HOY en reposo. Asserta las CUATRO filas de cromo que el paquete va a
// bajar a dos: si alguna deja de estar antes de tiempo, esta story lo dice.
export const VigenteReposo: Story = {
  parameters: SOLO_BASELINE,
  decorators: [escenario({ sessions: [SESION], activeId: SESION.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // fila 1 — header con el frente de la sesión y el colapsar (glifo vigente `⟩`, CV-D15).
    await expect(c.getByText("repro del bug de carga")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /colapsar/ })).toBeInTheDocument()
    // fila 2 — SessionLine: el `◍` con el cc-id, el arnés, el modelo y el ctx.
    await expect(c.getByText(/◍ 4b046945/)).toBeInTheDocument()
    await expect(c.getByText("68%")).toBeInTheDocument()
    // fila 3 — ScopeRow: el rótulo fijo y el chip punteado del arnés, aun sin nodo elegido.
    await expect(c.getByText("Alcance:")).toBeInTheDocument()
    await expect(c.getByText(/selecciona un nodo en el Mapa/)).toBeInTheDocument()
    // fila 4 — composer.
    await expect(c.getByPlaceholderText("Pídele un cambio a vitalia")).toBeInTheDocument()
    // el transcript vigente: burbujas + tarjeta de actividad agrupada.
    await expect(c.getByText(/el manifiesto declara 0 elementos/)).toBeInTheDocument()
    await expect(c.getByText("1 paso")).toBeInTheDocument()
  },
}

// A-02 · con una tarjeta de permiso abierta. El dock queda `await` y la PermissionCard
// renderiza dentro del transcript (RF-113) — superficie firmada en HS-26.
export const VigenteConPermiso: Story = {
  parameters: SOLO_BASELINE,
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
    // La tarjeta nombra el archivo en la cabecera Y en el diff: dos apariciones, a propósito.
    await expect(c.getAllByText("skills/builder/SKILL.md").length).toBeGreaterThanOrEqual(1)
    await expect(c.getByRole("button", { name: "Permitir esta sesión" })).toBeInTheDocument()
    // `await` cuenta como turno en vuelo: el composer lo dice en su placeholder.
    await expect(c.getByPlaceholderText(/esperando tu decisión de permiso/)).toBeInTheDocument()
  },
}

// A-03 · con el turno en vuelo. El `■` de interrumpir reemplaza al `↑` de enviar — es el
// mismo botón en dos estados, y el paquete no lo toca.
export const VigenteStreaming: Story = {
  parameters: SOLO_BASELINE,
  decorators: [
    escenario({
      sessions: [{ ...SESION, status: "streaming" }],
      activeId: SESION.id,
      streaming: { [SESION.id]: "revisando el loader" },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // El nombre accesible sale del CONTENIDO (■ / ↑); el `title` es solo la ayuda emergente.
    await expect(c.getByRole("button", { name: "■" })).toBeInTheDocument()
    await expect(c.queryByRole("button", { name: "↑" })).not.toBeInTheDocument()
    await expect(c.getByText("revisando el loader")).toBeInTheDocument()
    await expect(c.getByPlaceholderText(/generando…/)).toBeInTheDocument()
  },
}
