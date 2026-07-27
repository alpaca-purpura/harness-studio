import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { CtxChip, IdentidadDetalle } from "./ctx-chip"

// Story = test (fe-visual-fitness) — paquete 2026-07-26-conversaciones-del-panel, T23.
// C-01…C-10 de `plan-storybook.md` §2.3. Ninguna baja `a11y` a "todo" (RF-353 CA-4): el gate
// corre entero, contraste incluido, sobre TODA la superficie nueva.

const CWD = "~/Proyectos/luana-vitalia/vitalia"

// rgbDe resuelve un token DTCG al `rgb(...)` que el navegador computa, en el tema vigente.
// Compararlo así y no contra un hex tecleado es lo que hace que el assert valga en los DOS
// temas: `--warn` es `#c96a2e` en claro y otro valor en oscuro.
function rgbDe(raiz: HTMLElement, token: string): string {
  const sonda = document.createElement("span")
  sonda.style.color = `var(${token})`
  raiz.appendChild(sonda)
  const v = getComputedStyle(sonda).color
  sonda.remove()
  return v
}

const meta = {
  title: "widgets/chat-dock/CtxChip",
  component: CtxChip,
  parameters: { layout: "centered" },
  args: {
    ctxPct: 68,
    arnes: "vitalia",
    claudeSessionId: "4b046945-1f0e-4c0a-9a51-1b2f9c8d7e30",
    model: "claude-opus-5[1m]",
    cwd: CWD,
    caliente: false,
    abierto: false,
    detalleId: "cv-detalle",
    onToggle: fn(),
  },
  // El detalle es una fila HERMANA del chip (así lo dibuja el mockup: `.ctxchip` en la
  // `.convrow`, `.detalle` debajo). La story compone las dos piezas como lo hace el widget.
  render: (args) => (
    <div style={{ width: 300 }}>
      <div className="flex items-center justify-end border-border border-b px-2 py-1">
        <CtxChip {...args} />
      </div>
      {args.abierto && (
        <IdentidadDetalle
          id={args.detalleId}
          claudeSessionId={args.claudeSessionId}
          arnes={args.arnes}
          model={args.model}
          cwd={args.cwd}
        />
      )}
    </div>
  ),
} satisfies Meta<typeof CtxChip>

export default meta
type Story = StoryObj<typeof meta>

// C-01 · cerrado y frío. La cifra vive en el NOMBRE del botón, no sólo en la barra — la
// barra es decorativa y está `aria-hidden` (design.md §8.1).
export const Frio: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 68/ })
    await expect(chip).toHaveAttribute("aria-expanded", "false")
    await expect(canvasElement.querySelector("[data-barra-ctx]")?.parentElement).toHaveAttribute(
      "aria-hidden",
    )
    await expect(chip).not.toHaveAttribute("data-caliente")
  },
}

// C-02 · caliente. EL assert de estilo del plan, y el único: el NÚMERO no se tiñe de
// `--warn` (3,76:1 sobre `--card`, bajo el 4,5 de texto) — la señal va por la barra, que es
// no textual y sí cumple 3:1. Es la realización de C-3.
export const Caliente: Story = {
  args: { caliente: true, ctxPct: 92 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 92/ })
    const warn = rgbDe(canvasElement, "--warn")
    await expect(getComputedStyle(chip).color).not.toBe(warn)
    await expect(getComputedStyle(chip).color).toBe(rgbDe(canvasElement, "--foreground"))
    const barra = canvasElement.querySelector("[data-barra-ctx]") as HTMLElement
    await expect(getComputedStyle(barra).backgroundColor).toBe(warn)
  },
}

// C-03 · a 0 % el chip se pinta IGUAL. Cero es un dato, no un vacío (BR-CV-9, E-04): esconderlo
// haría indistinguible «recién nace» de «no sé cuánto».
export const Cero: Story = {
  args: { ctxPct: 0, claudeSessionId: undefined, model: undefined },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: /contexto 0%/ })).toBeInTheDocument()
    await expect(c.getByText("0%")).toBeInTheDocument()
  },
}

// C-04 · abierto: el detalle existe y el botón lo declara con aria-expanded/aria-controls.
export const Abierto: Story = {
  args: { abierto: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 68/ })
    await expect(chip).toHaveAttribute("aria-expanded", "true")
    await expect(chip).toHaveAttribute("aria-controls", "cv-detalle")
    await expect(canvasElement.querySelector("#cv-detalle")).toBeInTheDocument()
  },
}

// C-05 · dos clics, dos toggles. El chip es CONTROLADO: no guarda el estado, lo informa.
export const Toggle: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 68/ })
    await userEvent.click(chip)
    await userEvent.click(chip)
    await expect(args.onToggle).toHaveBeenCalledTimes(2)
  },
}

// C-06 · los CUATRO datos de la SessionLine vigente + el cwd (RF-328). Es la prueba del
// superset: la fila que se retira no perdió nada.
export const DetalleCompleto: Story = {
  args: { abierto: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/◍ 4b046945/)).toBeInTheDocument()
    await expect(c.getByText("vitalia")).toBeInTheDocument()
    await expect(c.getByText("claude-opus-5[1m]")).toBeInTheDocument()
    await expect(c.getByText(CWD)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /contexto 68/ })).toBeInTheDocument()
  },
}

// C-07 · sin sesión de Claude Code todavía (E-05). El literal vigente sobrevive tal cual
// (`chat-dock.tsx:82`), adentro del detalle.
export const DetalleSinSesionCc: Story = {
  args: { abierto: true, claudeSessionId: undefined },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("◍ sin sesión CC")).toBeInTheDocument()
  },
}

// C-08 · sin cwd la línea NO se dibuja. Un «—» de relleno diría «no hay»; lo que pasa es que
// el daemon no lo registró, y eso se calla, no se inventa.
export const DetalleSinCwd: Story = {
  args: { abierto: true, cwd: undefined },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("cwd")).not.toBeInTheDocument()
    await expect(c.getByText(/◍ 4b046945/)).toBeInTheDocument()
  },
}

// C-09 · cwd larguísimo: trunca en pantalla y el completo queda en el `title`.
export const DetalleCwdLargo: Story = {
  args: {
    abierto: true,
    cwd: "~/Proyectos/alpacapurpura/clientes/luana-vitalia/instalaciones/produccion/vitalia-core/arnes",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const nodo = c.getByTitle(args.cwd ?? "")
    await expect(nodo).toBeInTheDocument()
    await expect(nodo.scrollWidth).toBeGreaterThan(nodo.clientWidth)
  },
}

// C-10 · el mismo assert de color en OSCURO. `--warn` cambia de valor entre temas: por eso la
// sonda resuelve el token en vivo en vez de comparar contra un hex tecleado (RF-353).
export const CalienteDark: Story = {
  args: { caliente: true, ctxPct: 92 },
  globals: { theme: "dark" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /contexto 92/ })
    const warn = rgbDe(canvasElement, "--warn")
    await expect(getComputedStyle(chip).color).not.toBe(warn)
    const barra = canvasElement.querySelector("[data-barra-ctx]") as HTMLElement
    await expect(getComputedStyle(barra).backgroundColor).toBe(warn)
  },
}
