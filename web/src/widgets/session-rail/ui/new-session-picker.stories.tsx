import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { type EntradaPortafolio, entradasDemo } from "@/entities/portafolio"
import { NewSessionPicker } from "./new-session-picker"

// Story = test (fe-visual-fitness) — el picker de «＋ Nueva sesión» (RF-5..17, TS-D6..D16). El
// componente ya lleva la clase `arnesia-portafolio` en su raíz, así que los chips reusados de
// entities/portafolio resuelven sus estilos `.pf-*` igual que en la app. El Frame sólo le da el
// ancho (360px, el rail ensanchado) y una altura fija para que el layout flex (min-h-0 flex-1 +
// acciones con mt-auto) se comporte como en el <aside>.
function Frame({ children }: { children: ReactNode }) {
  return (
    <div className="flex h-[460px] w-[360px] flex-col border border-border bg-card">{children}</div>
  )
}

// Fuente = el Portafolio REAL (fixtures honestas de entities/portafolio/testing): (a) harness =
// 1 copia (caso simple, en-deriva), (b) provisional = 1 copia, (c) acme-cli = canónico + 2
// instalaciones (caso AMBIGUO real). No se inventan arneses de catálogo.

const HARNESS_PATH = "~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2"

const meta = {
  title: "widgets/session-rail/NewSessionPicker",
  component: NewSessionPicker,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    estado: "datos",
    entradas: entradasDemo,
    onCrear: fn(),
    onCancelar: fn(),
    onReintentar: fn(),
    onIrPortafolio: fn(),
  },
} satisfies Meta<typeof NewSessionPicker>

export default meta
type Story = StoryObj<typeof meta>

// RF-10/11/16 caso simple: una identidad con 1 sola copia → la fila resuelve el path sola (sin
// sub-lista) y Crear emite el payload REAL (arnes/path; sin empresa porque harness no declara).
export const CasoSimple: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const crear = c.getByRole("button", { name: "Crear sesión" })
    await expect(crear).toBeDisabled()

    await userEvent.click(c.getByText("harness"))
    await expect(c.getByText(HARNESS_PATH)).toBeInTheDocument() // "usará <path>"
    await expect(crear).toBeEnabled()

    await userEvent.click(crear)
    await expect(args.onCrear).toHaveBeenCalledWith({ arnes: "harness", path: HARNESS_PATH })
  },
}

// RF-10 caso ambiguo: canónico + 2 instalaciones → sub-lista, Crear deshabilitado hasta elegir la
// copia; empresa = empresas.join (N:M), path = la copia elegida (acá el canónico).
export const CasoAmbiguo: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const crear = c.getByRole("button", { name: "Crear sesión" })

    await userEvent.click(c.getByText("acme-cli"))
    // sub-lista visible + Crear sigue deshabilitado (hay que elegir CUÁL copia)
    await expect(c.getByRole("group", { name: "Copias de acme-cli" })).toBeInTheDocument()
    await expect(crear).toBeDisabled()

    await userEvent.click(c.getByText("~/dev/acme-cli")) // la copia canónica
    await expect(crear).toBeEnabled()

    await userEvent.click(crear)
    await expect(args.onCrear).toHaveBeenCalledWith({
      arnes: "acme-cli",
      empresa: "alpacapurpura",
      path: "~/dev/acme-cli",
    })
  },
}

// RF-13/TS-D11 vacío: 0 entradas → copy + «Ir a Portafolio» (deriva a la vista), sin buscador.
export const PortafolioVacio: Story = {
  args: { entradas: [] },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Tu portafolio está vacío.")).toBeInTheDocument()
    await expect(c.queryByLabelText("Buscar arnés")).toBeNull()
    await userEvent.click(c.getByRole("button", { name: "Ir a Portafolio" }))
    await expect(args.onIrPortafolio).toHaveBeenCalled()
  },
}

// RF-14/TS-D13 error: motivo textual REAL + Reintentar que dispara onReintentar.
export const ErrorDeCarga: Story = {
  args: {
    estado: "error",
    error: "GET /api/portafolio: conexión rechazada (ECONNREFUSED)",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent(/conexión rechazada/)
    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onReintentar).toHaveBeenCalled()
  },
}

// RF-9/TS-D12 colisión: dos filas con el MISMO id se distinguen por un chip home/scope — no
// bloquea la selección (la fila selecciona por `clave`, única).
const colisionA: EntradaPortafolio = {
  clave: "github-com-acme-dup~dup~",
  identidad: { home: "github.com/acme/dup", id: "dup" },
  empresas: ["acme"],
  canonico: { path: "~/dev/acme-dup", version: "1.0.0" },
}
const colisionB: EntradaPortafolio = {
  clave: "github-com-beta-dup~dup~",
  identidad: { home: "github.com/beta/dup", id: "dup" },
  empresas: ["beta"],
  canonico: { path: "~/dev/beta-dup", version: "2.0.0" },
}

export const Colision: Story = {
  args: { entradas: [colisionA, colisionB] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // ambas filas visibles con su chip distintivo (home)
    await expect(c.getByText("github.com/acme/dup")).toBeInTheDocument()
    await expect(c.getByText("github.com/beta/dup")).toBeInTheDocument()
    // seleccionar una NO se bloquea: 1 copia (canónico) → resuelve y habilita Crear
    const filas = c.getAllByText("dup")
    await expect(filas).toHaveLength(2)
    const primera = filas[0]
    if (primera) await userEvent.click(primera)
    await expect(c.getByRole("button", { name: "Crear sesión" })).toBeEnabled()
  },
}

// RF-6 buscador en vivo (id/nombre/empresa) + sin-resultados + limpiar.
export const Buscar: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const buscar = c.getByLabelText("Buscar arnés")

    await userEvent.type(buscar, "acme")
    await expect(c.getByText("acme-cli")).toBeInTheDocument()
    await expect(c.queryByText("harness")).toBeNull() // filtrada fuera

    await userEvent.clear(buscar)
    await userEvent.type(buscar, "zzz-no-existe")
    await expect(c.getByText("Ningún arnés coincide con la búsqueda.")).toBeInTheDocument()

    await userEvent.click(c.getByRole("button", { name: "Limpiar búsqueda" }))
    await expect(c.getByText("acme-cli")).toBeInTheDocument() // vuelven todas
  },
}

// RF-12 cancelar: dispara onCancelar (el reset de búsqueda/selección ocurre por unmount en la app,
// donde SessionRail desmonta el picker + resetea el store — RF-17).
export const Cancelar: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByText("harness"))
    await userEvent.click(c.getByRole("button", { name: "Cancelar" }))
    await expect(args.onCancelar).toHaveBeenCalled()
  },
}
