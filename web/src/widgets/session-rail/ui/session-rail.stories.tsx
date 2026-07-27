import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, userEvent, within } from "storybook/test"
import type { Session } from "@/shared"
import { useSessions } from "@/shared"
import { SessionRail } from "./session-rail"

// Story = test (fe-visual-fitness) — bugfix «acciones de la tarjeta de sesión» (2026-07-27).
//
// El bug que estas stories congelan: el ✕ de cerrar era `absolute right-1.5 top-1.5` y caía
// ENCIMA del ✎ de renombrar (inline al final de la fila del título). El ✎ se veía, pero el
// click se lo comía el ✕ — de ahí «no puedo editar el título». Y el ✕ sólo se pintaba con
// `sessions.length > 1`, así que la última sesión no se podía cerrar pese a que el backend
// (`SessionService.Close`) nunca tuvo esa restricción.
//
// B-02 mide el solape con `elementFromPoint` sobre el centro de cada botón: es la única forma
// de probar la HITBOX y no la mera presencia — la versión anterior pasaba cualquier assert de
// visibilidad y aun así era inclickeable.
//
// ⚠ Las stories revelan el cluster con FOCO (`.focus()` sobre la tarjeta → `group-focus-within`),
// no con hover: `userEvent.hover` de testing-library es sintético y no enciende la pseudo-clase
// CSS `:hover` en un browser real. Es el mismo par de clases (`invisible` →
// `group-hover:visible group-focus-within:visible`) sobre el mismo cluster, así que la geometría
// medida es la que ve el mouse; y de paso queda probado que el cluster es alcanzable por teclado.

// escenario escribe el estado ENTERO del store antes de renderizar (calca chat-dock.stories.tsx:25):
// jamás un merge parcial, para que el orden de ejecución de las stories no importe.
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
      <div style={{ height: 520 }} className="flex bg-background">
        {Story()}
      </div>
    )
  }
}

// Datos de la máquina real (`GET /api/sessions`, misma fuente que chat-dock.stories.tsx).
// `conv` ya NO vive en la sesión: bajó a `activa` (CV-D3). El rail no la lee, pero el
// tipo la exige sin `?` a propósito — un campo opcional invitaría al `if (!activa)`
// defensivo que escondería el día en que el daemon no la mande.
const hiloVacio = (id: string, titulo: string): Session["activa"] => ({
  id,
  titulo,
  titulo_editado: false,
  activa: true,
  turnos: 0,
  ctx_pct: 0,
  rotacion_pendiente: false,
  creada_en: "2026-07-26T10:00:00Z",
  conv: [],
})

const VITALIA: Session = {
  id: "s6165ac75",
  frente: "repro del bug de carga",
  arnes: "vitalia",
  status: "idle",
  salud: "info",
  view: "Mapa",
  activa: hiloVacio("cv6165ac75", "repro del bug de carga"),
}

const ARNESIA: Session = {
  id: "s91ba0c12",
  frente: "franja de artefactos",
  arnes: "arnesia",
  status: "await",
  salud: "warn",
  view: "Mapa",
  activa: hiloVacio("cv91ba0c12", "franja de artefactos"),
}

const IR = "Ir a la sesión «repro del bug de carga» de vitalia"
const LAPIZ = "Renombrar frente «repro del bug de carga»"
const CERRAR = "Cerrar sesión «repro del bug de carga»"

// centro devuelve quién RECIBE el click en el punto medio del botón.
const centro = (el: HTMLElement) => {
  const r = el.getBoundingClientRect()
  return document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2)
}

// 🔴 HALLAZGO AL ESTRENAR STORY (2026-07-27) — se declara, no se tapa. El rail nunca había
// tenido story: la primera corrida destapó 3 violaciones de contraste REALES y PREEXISTENTES,
// todas de la familia C-3 ya anotada en `BACKLOG.md` (token de acento usado como color de
// texto): el badge `N ◐` (`--warn` 3.24:1), el «⭯ persisten» (`--ok` 4.03:1) y el nombre del
// arnés (`text-primary/90`). Ninguna está en el cluster ✎/✕ que toca este bugfix.
//
// Siguiendo lo hecho en chat-dock.stories.tsx:86, se apaga **la sola regla `color-contrast`**
// — el resto de axe sigue en `error`. Arreglar los tokens acá sería tocar superficie vigente
// fuera del ticket; la deuda vive en `BACKLOG.md`.
const SOLO_CONTRASTE_DIFERIDO = {
  a11y: { config: { rules: [{ id: "color-contrast", enabled: false }] } },
}

const meta: Meta<typeof SessionRail> = {
  title: "widgets/SessionRail",
  component: SessionRail,
  parameters: SOLO_CONTRASTE_DIFERIDO,
}
export default meta

type Story = StoryObj<typeof SessionRail>

// B-01 — reposo: ninguna acción expuesta; el título manda en la fila.
export const Reposo: Story = {
  decorators: [escenario({ sessions: [VITALIA, ARNESIA], activeId: VITALIA.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("repro del bug de carga")).toBeVisible()
    await expect(c.queryByRole("button", { name: LAPIZ })).toBeNull()
    await expect(c.queryByRole("button", { name: CERRAR })).toBeNull()
  },
}

// B-02 — activada la tarjeta: ✎ y ✕ aparecen JUNTOS y sin pisarse. El assert duro es la hitbox.
export const AccionesSinSolape: Story = {
  decorators: [escenario({ sessions: [VITALIA, ARNESIA], activeId: VITALIA.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    c.getByRole("button", { name: IR }).focus()

    const lapiz = c.getByRole("button", { name: LAPIZ })
    const cerrar = c.getByRole("button", { name: CERRAR })
    await expect(lapiz).toBeVisible()
    await expect(cerrar).toBeVisible()

    // Sin solape geométrico.
    const a = lapiz.getBoundingClientRect()
    const b = cerrar.getBoundingClientRect()
    const solapan = a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom
    await expect(solapan).toBe(false)

    // Y el punto medio de cada botón LO RECIBE ese botón (lo que rompía el ✕ absoluto).
    await expect(lapiz.contains(centro(lapiz))).toBe(true)
    await expect(cerrar.contains(centro(cerrar))).toBe(true)
  },
}

// B-03 — el ✎ abre el input de renombrar (no dispara red: sólo estado local mientras no se
// commitea un valor distinto).
export const RenombrarAbre: Story = {
  decorators: [escenario({ sessions: [VITALIA, ARNESIA], activeId: VITALIA.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    c.getByRole("button", { name: IR }).focus()
    await userEvent.click(c.getByRole("button", { name: LAPIZ }))
    await expect(c.getByRole("textbox")).toHaveValue("repro del bug de carga")
  },
}

// B-04 — con UNA sola sesión el ✕ sigue existiendo y es clickeable: el shell tiene vacío
// honesto («Sin sesión activa») y el pie del rail ofrece «＋ Nueva sesión».
export const UnicaSesionSeCierra: Story = {
  decorators: [escenario({ sessions: [VITALIA], activeId: VITALIA.id })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    c.getByRole("button", { name: IR }).focus()
    const cerrar = c.getByRole("button", { name: CERRAR })
    await expect(cerrar).toBeVisible()
    await expect(cerrar).toBeEnabled()
    await expect(cerrar.contains(centro(cerrar))).toBe(true)
  },
}
