import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { useState } from "react"
import { expect, fn, userEvent, waitFor, within } from "storybook/test"
import { candidatosOrigenDemo } from "@/entities/marketplace"
import { ResolverOrigenDialog } from "./resolver-origen-dialog"

// Story = test (fe-visual-fitness) — diálogo S7 «Resolver origen» (AG-D8 decisión 7 +
// PENDIENTE-01 §3). Frame reproduce el scope real (.arnesia-portafolio, portafolio.css).
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const meta = {
  title: "widgets/portafolio/ResolverOrigenDialog",
  component: ResolverOrigenDialog,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    abierto: true,
    arnesId: "legal-administrativo",
    candidatos: candidatosOrigenDemo,
    eleccion: undefined,
    confirmando: false,
    onElegir: fn(),
    onConfirmar: fn(),
    onClose: fn(),
  },
} satisfies Meta<typeof ResolverOrigenDialog>

export default meta
type Story = StoryObj<typeof meta>

// E-25 — **nada premarcado** y `Confirmar origen` nace `disabled` con el tooltip literal del
// mockup. Al elegir cualquier opción se habilita. El orden lo decidió el backend por señales
// blandas; el widget no reordena ni auto-decide.
export const ResolverOrigenConfirma: Story = {
  render: (args) => {
    const [eleccion, setEleccion] = useState<string | null | undefined>(undefined)
    return <ResolverOrigenDialog {...args} eleccion={eleccion} onElegir={setEleccion} />
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    // nada premarcado: ninguno de los radios (4 candidatos + «ninguno») viene checked.
    const radios = c.getAllByRole("radio")
    await expect(radios).toHaveLength(5)
    for (const r of radios) await expect(r).not.toBeChecked()

    const confirmar = c.getByRole("button", { name: "Confirmar origen" })
    await expect(confirmar).toBeDisabled()
    await expect(confirmar).toHaveAttribute("title", "elegí un destino primero")

    // el orden es el del backend, tal cual llegó (el widget NO ordena).
    const valores = Array.from(canvasElement.querySelectorAll(".pf-resolver-op .mono")).map(
      (e) => e.textContent,
    )
    await expect(valores).toEqual([
      "github.com/vitalia/arneses",
      "github.com/alpacapurpura/prenter-marketplace",
      "github.com/nordia/plugins-rrhh",
      "github.com/anthropics/claude-plugins-official",
    ])

    await userEvent.click(c.getByRole("radio", { name: /github\.com\/vitalia\/arneses/ }))
    await expect(confirmar).toBeEnabled()
    await userEvent.click(confirmar)
    await expect(args.onConfirmar).toHaveBeenCalledTimes(1)
  },
}

// E-26 — «ninguno — dejarlo sin origen» es la ÚLTIMA opción, lleva la señal `honesto, no se
// inventa un home`, y **habilita** Confirmar: es una decisión válida, no la ausencia de decisión.
export const ResolverOrigenNinguno: Story = {
  render: (args) => {
    const [eleccion, setEleccion] = useState<string | null | undefined>(undefined)
    return <ResolverOrigenDialog {...args} eleccion={eleccion} onElegir={setEleccion} />
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    const opciones = Array.from(canvasElement.querySelectorAll(".pf-resolver-op"))
    const ultima = opciones[opciones.length - 1] as HTMLElement
    await expect(ultima).toHaveTextContent("ninguno — dejarlo sin origen")
    await expect(ultima).toHaveTextContent("honesto, no se inventa un home")

    const confirmar = c.getByRole("button", { name: "Confirmar origen" })
    await expect(confirmar).toBeDisabled()
    await userEvent.click(within(ultima).getByRole("radio"))
    await expect(confirmar).toBeEnabled()
    await userEvent.click(confirmar)
    await expect(args.onConfirmar).toHaveBeenCalledTimes(1)
  },
}

// Cada candidato muestra POR QUÉ está sugerido: la señal la ARMA el backend (`senal`, testeable
// en el usecase), no el widget. El operador ve el criterio en vez de confiar en una lista muda.
export const ResolverOrigenConSenales: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("el registry de tu copia ya apunta acá")).toBeInTheDocument()
    await expect(c.getByText("propio · no leído aún")).toBeInTheDocument()
    await expect(
      c.getByText("de referencia · resuelve procedencia de 1 arnés tuyo"),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".pf-resolver-senal")).toHaveLength(5)
  },
}

// E-69 — colisión de identidad (409): se EXPLICA, no se resuelve sola. El diálogo sigue abierto
// para elegir otro destino y las dos entradas quedan intactas (fusionar es otra operación).
export const ResolverOrigenColision: Story = {
  args: {
    eleccion: "github.com/vitalia/arneses",
    error:
      "POST /api/portafolio/arneses/…/origen: 409 portafolio: ya existe una entrada con esa identidad (fusionar es otra operación)",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("ya existe una entrada con esa identidad")
    await expect(c.getByRole("alert")).toHaveTextContent("fusionar es otra operación")
    // el diálogo sigue vivo y la elección se conserva: se puede elegir otro destino.
    await expect(c.getByRole("button", { name: "Confirmar origen" })).toBeEnabled()
    await expect(c.getByRole("dialog")).toBeInTheDocument()
  },
}

// Cancelar = CERO efectos: `onConfirmar` nunca se llama, ni por click ni por Esc.
export const ResolverOrigenCancela: Story = {
  args: { eleccion: "github.com/vitalia/arneses" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Cancelar" }))
    await expect(args.onClose).toHaveBeenCalledTimes(1)
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(2)
    await expect(args.onConfirmar).not.toHaveBeenCalled()
  },
}

// Con una escritura en vuelo, cerrar a mitad es peor que esperar un instante (mismo criterio que
// S1-D19): botones bloqueados, `aria-busy`, y Esc no cierra.
export const ResolverOrigenConfirmando: Story = {
  args: { eleccion: "github.com/vitalia/arneses", confirmando: true },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Confirmando…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")
    await expect(c.getByRole("button", { name: "Cancelar" })).toBeDisabled()
    await expect(c.getByRole("button", { name: "Cerrar" })).toBeDisabled()
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).not.toHaveBeenCalled()
  },
}

// Un arnés que YA tiene home declarado: el diálogo lo dice y avisa que confirmar lo reemplaza —
// no se pisa nada en silencio.
export const ResolverOrigenConActual: Story = {
  args: { actual: "github.com/alpacapurpura/prenter-marketplace" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/ya tiene origen declarado/)).toBeInTheDocument()
    await expect(c.getByText(/confirmar lo reemplaza/)).toBeInTheDocument()
  },
}

// E-30 — a11y del diálogo (gate axe en `error`: cualquier violación ROMPE el test sin aserción
// extra). `role=dialog` + `aria-modal` + `aria-labelledby` al `<h2>`; foco inicial dentro; Tab
// cicla en las dos direcciones (`trapTabKeyDown`); Esc cierra.
export const ResolverOrigenA11y: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const dialog = c.getByRole("dialog")
    await expect(dialog).toHaveAttribute("aria-modal", "true")
    const labelledby = dialog.getAttribute("aria-labelledby")
    await expect(labelledby).toBeTruthy()
    const titleEl = labelledby ? document.getElementById(labelledby) : null
    await expect(titleEl).toHaveTextContent("Resolver origen — legal-administrativo")

    // el radiogroup está rotulado (los radios no quedan huérfanos para un lector de pantalla).
    await expect(c.getByRole("radiogroup", { name: "Marketplace de origen" })).toBeInTheDocument()

    const cerrar = c.getByRole("button", { name: "Cerrar" })
    await waitFor(() => expect(cerrar).toHaveFocus())

    // Trap hacia atrás desde el primer focuseable: va al último. Confirmar arranca `disabled`
    // (nada elegido) ⇒ el último focuseable es `Cancelar`.
    await userEvent.tab({ shift: true })
    await expect(c.getByRole("button", { name: "Cancelar" })).toHaveFocus()
    // …y hacia adelante vuelve al primero.
    await userEvent.tab()
    await expect(cerrar).toHaveFocus()

    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}
