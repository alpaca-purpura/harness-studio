import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { type EntradaCorrupta, entradasDemo } from "@/entities/portafolio"
import { PortafolioList } from "./portafolio-list"

// Story = test (fe-visual-fitness) — widget Lista del Portafolio (plan §2.6/§3 T4,
// G1/G2/G4/G5/G6/G8/G9). Frame reproduce el scope real (.arnesia-portafolio, portafolio.css)
// para que tokens/clases resuelvan igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

// corruptaEjemplo — sintética-con-shape-real (T4): entities/portafolio/testing/entradas.ts
// (T2) solo cubre EntradaPortafolio, no EntradaCorrupta. El texto replica el formato REAL del
// store (internal/adapters/portafolio/store.go:71, `derr.Error()` de un envelope que no
// matchea domain.EntradaPortafolio) — un error de decode Go real, no un motivo inventado.
const corruptaEjemplo: EntradaCorrupta = {
  motivo:
    "json: cannot unmarshal string into Go struct field EntradaPortafolio.instalaciones of type []domain.Instalacion",
}

const meta = {
  title: "widgets/portafolio/PortafolioList",
  component: PortafolioList,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    estado: "datos",
    entradas: [],
    corruptas: [],
    lente: "empresa",
    busqueda: "",
    onLente: fn(),
    onBusqueda: fn(),
    onAbrir: fn(),
    onAgregar: fn(),
    onReintentar: fn(),
  },
} satisfies Meta<typeof PortafolioList>

export default meta
type Story = StoryObj<typeof meta>

// G5 — portafolio genuinamente vacío (no confundir con FiltroSinResultados): mensaje honesto +
// el ＋ Agregar de la topbar (SIEMPRE presente) es la CTA que abre el wizard.
export const Vacia: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Tu portafolio está vacío.")).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "＋ Agregar" })
    await expect(btn).toBeEnabled()
    await userEvent.click(btn)
    await expect(args.onAgregar).toHaveBeenCalled()
  },
}

// G5 — cargando: skeleton honesto, CERO filas fantasma (assert estructural, no solo visual).
export const Cargando: Story = {
  args: { estado: "cargando" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status", { name: "Cargando portafolio" })).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".pf-fila").length).toBe(0)
  },
}

// G5 — error de carga: motivo textual REAL + Reintentar habilitado que dispara onReintentar.
export const ErrorDeCarga: Story = {
  args: {
    estado: "error",
    error: "GET /api/portafolio: conexión rechazada (ECONNREFUSED)",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/conexión rechazada/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Reintentar" })
    await expect(btn).toBeEnabled()
    await userEvent.click(btn)
    await expect(args.onReintentar).toHaveBeenCalled()
  },
}

// El caso central: lente empresa (grupo `alpacapurpura` + «sin empresa» al final, G4) · chip
// `en-deriva` real (G1) SIN ningún flag de update (G2) · dots con su aria-label (G9) · contadores
// reales · click/Enter en fila llaman onAbrir(clave) (G6/G8) · lentes/filtros diferidos
// disabled+tooltip (BR-8/S1-D8).
export const ConDatos: Story = {
  args: { entradas: entradasDemo },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    // contadores REALES de la topbar (3 entradas, 1 empresa declarada: alpacapurpura)
    await expect(c.getByText(/3 arneses/)).toBeInTheDocument()
    await expect(c.getByText(/1 empresas/)).toBeInTheDocument()
    await expect(c.getByText(/lente: empresa/)).toBeInTheDocument()

    // G4 — grupos N:M: "alpacapurpura" (acme-cli) + "sin empresa" (harness · proyecto-instalado)
    await expect(c.getByText("alpacapurpura")).toBeInTheDocument()
    await expect(c.getByText("sin empresa")).toBeInTheDocument()

    // G1 — el chip de deriva SOLO sale del dato real (la fila del harness, en-deriva de verdad)
    await expect(c.getByText("en-deriva")).toBeInTheDocument()

    // G2 — muere el flag de update fabricado: ausencia explícita en TODO el DOM de la fila
    await expect(c.queryByText(/⬆/)).toBeNull()

    // G9 — dot con SU aria-label: 2 filas en "atención" (en-deriva · aviso+discrepancias),
    // 1 fila "sin señal" (deriva-no-evaluable, identidad provisional).
    await expect(c.getAllByRole("img", { name: "salud: atención" })).toHaveLength(2)
    await expect(c.getByRole("img", { name: "salud: sin señal" })).toBeInTheDocument()

    // S1-D26 — la entrada provisional (sin manifiesto) se identifica por su scope, ya no
    // «(sin id)»: la cadena id → scope → «(sin id)» rige en TODAS las superficies.
    await expect(c.getByText("github.com/alpacapurpura/harness-studio")).toBeInTheDocument()
    await expect(c.queryByText("(sin id)")).toBeNull()

    // G6/G8 — la fila es un <button> nativo: click Y Enter llaman onAbrir(clave) EXACTA.
    // `/^harness\b/`: el nombre accesible de la fila provisional ahora ARRANCA con su scope
    // (contiene «harness-studio») — anclar al inicio desambigua (S1-D26).
    const filaHarness = c.getByRole("button", { name: /^harness\b/ })
    await userEvent.click(filaHarness)
    await expect(args.onAbrir).toHaveBeenCalledWith(
      "sin-home~harness~github-com-alpacapurpura-luana-vitalia",
    )
    filaHarness.focus()
    await userEvent.keyboard("{Enter}")
    await expect(args.onAbrir).toHaveBeenCalledTimes(2)

    // S1-D8 (cerrada 2026-07-23) — las 4 lentes están vivas; los FILTROS (distinta afordancia,
    // sin slice aún) siguen disabled + tooltip "próximo", jamás simulados (BR-8).
    const lenteProyecto = c.getByRole("button", { name: "Proyecto" })
    await expect(lenteProyecto).not.toBeDisabled()
    await expect(lenteProyecto).toHaveAttribute("aria-pressed", "false")
    await userEvent.click(lenteProyecto)
    await expect(args.onLente).toHaveBeenCalledWith("proyecto")
    const filtroEstado = c.getByRole("button", { name: "Estado" })
    await expect(filtroEstado).toBeDisabled()
    await expect(filtroEstado).toHaveAttribute("title", "próximo")
  },
}

// Lente plano: mismas 3 filas, sin headers de grupo (S1-D8).
export const LentePlano: Story = {
  args: { entradas: entradasDemo, lente: "plano" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("alpacapurpura")).toBeNull()
    await expect(c.queryByText("sin empresa")).toBeNull()
    await expect(canvasElement.querySelectorAll(".pf-fila").length).toBe(3)
  },
}

// Lente proyecto (S1-D8, cerrada 2026-07-23): N:M vía instalaciones[].proyecto_path — acme-cli
// (2 instalaciones) aparece en 2 grupos de proyecto distintos.
export const LenteProyecto: Story = {
  args: { entradas: entradasDemo, lente: "proyecto" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("~/Proyectos/luana-vitalia")).toBeInTheDocument()
    await expect(c.getByText("~/Proyectos/harness-studio")).toBeInTheDocument()
    await expect(c.getByText("~/Proyectos/acme-app")).toBeInTheDocument()
    await expect(c.getByText("~/Proyectos/otro-app")).toBeInTheDocument()
    await expect(c.queryByText("sin proyecto instalado")).toBeNull()
  },
}

// Lente marketplace (S1-D8, cerrada 2026-07-23): N:M vía registriesDe — la entrada sin registry
// (proyecto-instalado provisional) cae en «origen desconocido» al final.
export const LenteMarketplace: Story = {
  args: { entradas: entradasDemo, lente: "marketplace" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("github.com/alpacapurpura/prenter-marketplace")).toBeInTheDocument()
    await expect(c.getByText("github.com/acme/acme-cli")).toBeInTheDocument()
    await expect(c.getByText("github.com/acme-fork/acme-cli")).toBeInTheDocument()
    await expect(c.getByText("origen desconocido")).toBeInTheDocument()
  },
}

// BR-11 — corruptas visibles: banner con el motivo REAL, el resto de la lista sigue viva.
export const CorruptasVisibles: Story = {
  args: { entradas: entradasDemo, corruptas: [corruptaEjemplo] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const banner = c.getByRole("alert")
    await expect(banner).toHaveTextContent("1 entrada(s) corrupta(s) en el registro")
    await expect(banner).toHaveTextContent(/cannot unmarshal string/)
    // la lista sigue viva pese a la corrupta (no es un estado excluyente)
    await expect(c.getByText("alpacapurpura")).toBeInTheDocument()
  },
}

// G5 — búsqueda sin match sobre datos reales (≠ Vacia: el portafolio SÍ tiene entradas).
export const FiltroSinResultados: Story = {
  args: { entradas: entradasDemo, busqueda: "zzz-no-existe-en-ninguna-fixture" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Ningún arnés coincide/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Limpiar búsqueda" })
    await userEvent.click(btn)
    await expect(args.onBusqueda).toHaveBeenCalledWith("")
  },
}
