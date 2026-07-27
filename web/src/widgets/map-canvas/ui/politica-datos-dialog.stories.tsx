import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { CAMPOS_PERSISTIDOS_HOOK } from "@/entities/telemetria"
import { PoliticaDatosDialog } from "./politica-datos-dialog"

// Story = test (fe-visual-fitness). RF-275 · RF-278 · RF-283 · H-5 · H-14.
//
// La lista de campos que estas stories assertan es la allowlist REAL del canal de hooks
// (`ANEXO-hooks.md` H4), importada del bloque MEDIDO de las fixtures — no una lista escrita
// para el test. Si la ingesta cambia su allowlist y la fixture no, la story se cae.

const meta = {
  title: "widgets/map-canvas/PoliticaDatosDialog",
  component: PoliticaDatosDialog,
  parameters: { layout: "padded" },
  args: {
    arnes: "vitalia",
    camposPersistidos: CAMPOS_PERSISTIDOS_HOOK,
    retencionDias: 90,
    corridasPorBorrar: 1284,
    alcance: "los últimos 7 días",
    onBorrar: fn(),
    onCerrar: fn(),
  },
} satisfies Meta<typeof PoliticaDatosDialog>

export default meta
type Story = StoryObj<typeof meta>

// RF-275 · H-14 — los tres bloques, con el copy literal. La segunda línea de «qué NO se guarda»
// es H-14: dice qué pasa con lo que SÍ llega por el canal de hooks. Callarlo dejaría al lector
// suponiendo una promesa distinta de la real.
export const Reposo: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const d = c.getByRole("dialog")
    await expect(d).toHaveAttribute("aria-modal", "true")
    const tituloId = d.getAttribute("aria-labelledby") as string
    await expect(canvasElement.querySelector(`#${CSS.escape(tituloId)}`)?.textContent).toBe(
      "Qué guardamos de la telemetría",
    )
    await expect(
      c.getByText(
        "No guardamos nada de tu cuenta: ni email, ni identificadores de usuario, ni de organización.",
      ),
    ).toBeInTheDocument()
    await expect(
      c.getByText(
        "Y no guardamos nada del contenido: ni tu prompt, ni la respuesta, ni lo que leyó o escribió una herramienta. Eso llega por el canal de hooks y se descarta antes de escribirse en disco.",
      ),
    ).toBeInTheDocument()
    await expect(c.getByText("Qué NO se guarda")).toBeInTheDocument()
    await expect(c.getByText("Qué SÍ se guarda")).toBeInTheDocument()
    await expect(c.getByText("Retención")).toBeInTheDocument()
  },
}

// RF-275 · RF-282 — **la lista de campos no la inventa la UI**: viene por prop y el DOM refleja
// exactamente ese array, ni uno más ni uno menos. Si el componente la escribiera, la promesa de
// privacidad dejaría de ser contrastable, que es lo único que la hace valer algo.
export const CamposDesplegados: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const b = c.getByRole("button", { name: "ver los campos exactos" })
    await expect(b).toHaveAttribute("aria-expanded", "false")
    await userEvent.click(b)
    await expect(b).toHaveAttribute("aria-expanded", "true")
    const items = [...canvasElement.querySelectorAll(".pol-campos li")].map((li) => li.textContent)
    await expect(items).toEqual([...args.camposPersistidos])
    for (const campo of [
      "session_id",
      "prompt_id",
      "hook_event_name",
      "tool_name",
      "duration_ms",
      "cwd",
    ]) {
      await expect(items).toContain(campo)
    }
  },
}

// RF-283 · J-6 — el `{N}` sale de la CONFIG, **no** de una constante de la UI. El número está
// firmado en 90 (D26.3) y sigue siendo configurable: firmarlo no lo clava. Esta story monta 45
// justamente para probar que el 90 no está hardcodeado en ningún lado — y que el rótulo
// «propuesto» ya no existe.
export const RetencionDesdeConfig: Story = {
  args: { retencionDias: 45 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Se borra solo a los 45 días.")).toBeInTheDocument()
    await expect(c.queryByText(/90 días/)).toBeNull()
    await expect(c.queryByText(/propuesto/i)).toBeNull()
  },
}

// RF-275 — la confirmación dice **cuánto** se borra antes de borrarlo, y el botón destructivo
// nace `disabled` hasta que el operador confirma que entendió el alcance.
export const Confirmacion: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Borrar la telemetría de este arnés" }))
    await expect(c.getByText("Borrar la telemetría de «vitalia»")).toBeInTheDocument()
    // D26.5 · A-4 — la frase NOMBRA el alcance. Hasta D26.5 el conteo era el de la ventana
    // activa y el `DELETE` era total: la confirmación subdeclaraba la destrucción.
    await expect(
      c.getByText(
        /Se borran 1 284 corridas medidas de los últimos 7 días y los puntos de mejora que salieron de ellas\. No se puede deshacer\./,
      ),
    ).toBeInTheDocument()
    const borrar = c.getByRole("button", { name: "Borrar" })
    await expect(borrar).toBeDisabled()
    await userEvent.click(c.getByRole("checkbox"))
    await expect(borrar).toBeEnabled()
  },
}

// D26.5 · A-4 — el otro alcance, para que el candado no pase con un literal hardcodeado. Si la
// frase no leyera la prop, esta story diría «los últimos 7 días» sobre un borrado total.
export const ConfirmacionDeTodoElHistorial: Story = {
  args: { alcance: "todo el historial", corridasPorBorrar: 47_912 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Borrar la telemetría de este arnés" }))
    await expect(
      c.getByText(/Se borran 47 912 corridas medidas de todo el historial/),
    ).toBeInTheDocument()
    await expect(c.queryByText(/últimos 7 días/)).toBeNull()
  },
}

// design §5.7 · S1-D19 — **mientras borra, el diálogo no se puede cerrar**: cerrarlo dejaría una
// operación en vuelo sin dónde reportar su resultado, y el operador no sabría si borró o no.
export const Borrando: Story = {
  args: { estado: "borrando" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const borrar = c.getByRole("button", { name: "Borrando…" })
    await expect(borrar).toHaveAttribute("aria-busy", "true")
    await expect(c.getByRole("button", { name: "Cancelar" })).toBeDisabled()
    await userEvent.keyboard("{Escape}")
    await expect(args.onCerrar).not.toHaveBeenCalled()
  },
}

// design §5.7 — el error deja el diálogo ABIERTO y el dato INTACTO, y lo dice con esas palabras.
export const ErrorDeBorrado: Story = {
  args: { estado: "error", error: "permiso denegado" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("No se pudo borrar — permiso denegado. No se borró nada."),
    ).toBeInTheDocument()
    await expect(c.getByRole("dialog")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Cancelar" })).toBeEnabled()
  },
}

// RF-275 — el éxito se anuncia por `aria-live`: el operador que no ve la pantalla también tiene
// que enterarse de que el borrado terminó.
export const Exito: Story = {
  args: { estado: "exito" },
  play: async ({ canvasElement }) => {
    const live = canvasElement.querySelector("[aria-live='polite']") as HTMLElement
    await expect(live.textContent).toBe("Listo. Este arnés vuelve a estar sin datos de telemetría.")
  },
}

// RF-278 — teclado: foco dentro al abrir, Tab que da la vuelta, Esc que cierra EN REPOSO, y
// anillo visible en cada control. Un modal sin trap deja el teclado detrás del diálogo.
export const FocoAtrapado: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const d = c.getByRole("dialog")
    await expect(d.contains(document.activeElement)).toBe(true)
    const controles = [...d.querySelectorAll<HTMLElement>("button, input")]
    await expect(controles.length).toBeGreaterThan(1)
    // Recorrer el diálogo entero con Tab: el foco NUNCA se escapa, y cada control enfocado
    // por teclado muestra su anillo (`:focus-visible` real, no un `focus()` programático).
    for (let i = 0; i <= controles.length; i += 1) {
      await userEvent.tab()
      const activo = document.activeElement as HTMLElement
      await expect(d.contains(activo)).toBe(true)
      await expect(getComputedStyle(activo).outlineWidth).not.toBe("0px")
    }
    // Esc en reposo cierra.
    await userEvent.keyboard("{Escape}")
    await expect(args.onCerrar).toHaveBeenCalled()
  },
}
