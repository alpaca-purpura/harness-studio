import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  marketplacesDemo,
  mkConDiscrepancia,
  mkNoLeido,
  mkOficialReferencia,
  mkPrenterPropio,
  mkSinAcceso,
} from "@/entities/marketplace"
import { MarketplaceList } from "./marketplace-list"

// Story = test (fe-visual-fitness) — widget del plano Marketplaces (S2, AG-D8 FIRMADA 🧑‍⚖️).
// Frame reproduce el scope real (.arnesia-portafolio, portafolio.css) para que tokens/clases
// resuelvan igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

// AHORA fijo: «leído hace 4 min» tiene que ser determinista (el widget recibe `ahora`, no lee el
// reloj) — sin esto la story sería flaky por diseño.
const AHORA = new Date("2026-07-25T14:11:33Z")

const meta = {
  title: "widgets/marketplace/MarketplaceList",
  component: MarketplaceList,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    estado: "datos",
    marketplaces: marketplacesDemo,
    corruptas: [],
    sinOrigenResuelto: 0,
    ahora: AHORA,
    onAbrirCatalogo: fn(),
    onLeerCatalogo: fn(),
    onAgregar: fn(),
    onReintentar: fn(),
    onVerSinOrigen: fn(),
  },
} satisfies Meta<typeof MarketplaceList>

export default meta
type Story = StoryObj<typeof meta>

// E-01 — el plano NACE POBLADO por lo que Claude Code ya conoce: no hay estado vacío
// «registrá tu primer marketplace». Cada fila trae repo · clase · nº de entradas · estado de
// lectura, y el orden es propios legibles → propios no legibles → referencia.
export const MarketplacesConDatos: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    // contadores REALES (5 filas de la fixture, 1 propio legible) — ninguna cifra tecleada.
    await expect(c.getByText(/5 conocidos · 1 propio legible/)).toBeInTheDocument()
    await expect(c.queryByText(/registrá tu primer marketplace/)).toBeNull()

    // orden de `ordenarMarketplaces`: el repo de cada fila, de arriba a abajo.
    const urls = Array.from(canvasElement.querySelectorAll(".mk-url")).map((e) => e.textContent)
    await expect(urls).toEqual([
      "github.com/alpacapurpura/prenter-marketplace",
      "github.com/nordia/plugins-rrhh",
      "github.com/vitalia/arneses",
      "github.com/juliusbrussee/caveman",
      "github.com/anthropics/claude-plugins-official",
    ])

    // la fila legible dice la cuenta REAL (2 de prenter, NO las «8» del mockup) y CUÁNDO se leyó.
    await expect(c.getByText("2 entradas de catálogo")).toBeInTheDocument()
    await expect(c.getByText("leído hace 4 min")).toBeInTheDocument()
    await expect(c.getByText("273 entradas de catálogo")).toBeInTheDocument()

    // la fila legible ES la acción (click al catálogo).
    await userEvent.click(c.getByRole("button", { name: /prenter-marketplace/ }))
    await expect(args.onAbrirCatalogo).toHaveBeenCalledWith("prenter-marketplace")

    // el ＋ Agregar de la topbar (compartido con el plano Arneses) abre el wizard.
    await userEvent.click(c.getByRole("button", { name: "＋ Agregar" }))
    await expect(args.onAgregar).toHaveBeenCalled()
  },
}

// BR-4 — una fila `sin acceso` NO navega: no hay ningún `<button class="mk-fila">` que lleve a un
// catálogo vacío (que se leería como «este marketplace no tiene arneses»). Muestra el motivo
// REAL y ofrece `Reintentar`. El copy dice explícitamente que NO es «no tiene arneses».
export const MarketplaceSinAcceso: Story = {
  args: { marketplaces: [mkSinAcceso] },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(canvasElement.querySelectorAll("button.mk-fila")).toHaveLength(0)
    await expect(canvasElement.querySelectorAll(".mk-fila.inerte")).toHaveLength(1)
    await expect(
      c.getByText(/^sin acceso — gh: HTTP 404 — repo inexistente o sin acceso/),
    ).toBeInTheDocument()
    await expect(c.getByText(/catálogo no legible/)).toBeInTheDocument()
    await expect(c.getByText(/no tiene arneses/)).toBeInTheDocument()

    // `--warn` (recuperable: autenticar), JAMÁS `--crit` (reservado a fallo).
    await expect(canvasElement.querySelector(".pf-chip-lectura.sin-acceso")).not.toBeNull()
    await expect(canvasElement.querySelector(".pf-chip-lectura.crit")).toBeNull()

    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onLeerCatalogo).toHaveBeenCalledWith("vitalia-arneses")
  },
}

// BR-4 + regla de token — `no leído aún` tampoco navega, ofrece `Leer catálogo`, y usa la clase
// `sin-senal` (transparente + dashed sobre `--muted-foreground`): «no sé» NUNCA reusa el color de
// `ok`. Mismo criterio que `DotSaludPortafolio`, aserción por CLASE (no por color).
export const MarketplaceNoLeido: Story = {
  args: { marketplaces: [mkNoLeido] },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(canvasElement.querySelectorAll("button.mk-fila")).toHaveLength(0)
    await expect(c.getByText("no leído aún")).toBeInTheDocument()
    await expect(
      c.getByText(/registrado, sin lectura — nada que afirmar todavía/),
    ).toBeInTheDocument()

    await expect(canvasElement.querySelector(".pf-chip-lectura.no-leido")).not.toBeNull()
    await expect(canvasElement.querySelector(".pf-chip-lectura.ok")).toBeNull()

    await userEvent.click(c.getByRole("button", { name: "Leer catálogo" }))
    await expect(args.onLeerCatalogo).toHaveBeenCalledWith("caveman")
  },
}

// E-07 / AG-D8 decisión 1 — las dos clases se distinguen A LA VISTA antes de entrar, y la
// asimetría está explicada en la superficie: el de referencia navega a un catálogo READ-ONLY y su
// fila lo dice.
export const MarketplacesDosClases: Story = {
  args: { marketplaces: [mkPrenterPropio, mkOficialReferencia] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    // «propio»/«de referencia» salen en el ClaseChip de la fila Y en la nota que explica la
    // asimetría: se scopea a las filas para asertar el CHIP, no la prosa.
    const filas = within(canvasElement.querySelector(".mk-lista") as HTMLElement)
    await expect(filas.getByText("propio")).toBeInTheDocument()
    await expect(filas.getByText("de referencia")).toBeInTheDocument()
    await expect(c.getByText("ver catálogo →")).toBeInTheDocument()
    await expect(c.getByText("ver catálogo (read-only) →")).toBeInTheDocument()

    // la nota que explica que NO son simétricas está en la superficie, no en un tooltip.
    await expect(c.getByText(/sin Traer/)).toBeInTheDocument()

    // los dos navegan (los dos son filas legibles).
    await expect(canvasElement.querySelectorAll("button.mk-fila")).toHaveLength(2)
  },
}

// AG-D8 decisión 7 — el contador cruzado es la SEGUNDA puerta a la reconciliación: solo conmuta
// al plano Arneses ya filtrado (la acción se ejecuta en la ficha del arnés, un solo lugar). La
// cifra viene del wire (`sin_origen_resuelto`), nunca se teclea.
export const ContadorCruzadoNavega: Story = {
  args: { sinOrigenResuelto: 3 },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("3 arneses sin origen resuelto.")).toBeInTheDocument()
    await expect(
      c.getByText(/Reparar y Actualizar no tienen contra qué comparar/),
    ).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Resolver origen" }))
    await expect(args.onVerSinOrigen).toHaveBeenCalledTimes(1)
  },
}

// Sin pendientes no hay puerta: la cifra 0 no pinta un banner vacío (mismo criterio que AvisoChip).
export const SinPendientesNoHayPuerta: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByRole("button", { name: "Resolver origen" })).toBeNull()
  },
}

// E-68/E-75 · BR-11 — registro corrupto Y detector ilegible a la vez: banner VISIBLE (no modal),
// las dos señales, y el resto de las filas sigue usable. Nunca un 500 ni una lista vacía muda.
export const MarketplacesConCorruptas: Story = {
  args: {
    corruptas: [{ motivo: "fila sin nombre: no hay clave de merge" }],
    avisoDetector:
      "known_marketplaces.json ilegible: invalid character '}' looking for beginning of object key string",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const banner = c.getByRole("alert")
    await expect(banner).toHaveTextContent("1 fila(s) corrupta(s) en el registro de marketplaces")
    await expect(banner).toHaveTextContent("fila sin nombre: no hay clave de merge")
    await expect(banner).toHaveTextContent("no pude leer lo que Claude Code conoce")
    // ninguna de las dos oculta a la otra ni tapa la lista.
    await expect(canvasElement.querySelectorAll(".mk-fila")).toHaveLength(5)
  },
}

// E-24 · BR-8 — el mismo nombre con repo distinto entre eslabones: UNA fila, gana el declarado, y
// la divergencia queda visible con LOS DOS valores crudos. **No hay ningún botón de «elegir
// una»**: el sistema no resuelve discrepancias solo.
export const MarketplaceConDiscrepancia: Story = {
  args: { marketplaces: [mkConDiscrepancia] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(canvasElement.querySelectorAll(".mk-fila")).toHaveLength(1)
    await expect(c.getByText("github.com/alpacapurpura/ponytail")).toBeInTheDocument()
    const aviso = c.getByText(/repo distinto entre eslabones/)
    await expect(aviso).toHaveTextContent("github.com/dietrichgebert/ponytail")
    await expect(aviso).toHaveTextContent("github.com/alpacapurpura/ponytail")

    // los dos eslabones existen y no se pide elegir.
    await expect(c.queryByRole("button", { name: /elegir/i })).toBeNull()
  },
}

// G5 — cargando: skeleton honesto, CERO filas fantasma.
export const Cargando: Story = {
  args: { estado: "cargando", marketplaces: [] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status", { name: "Cargando marketplaces" })).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".mk-fila")).toHaveLength(0)
  },
}

// G5 — error del GET: motivo textual REAL + Reintentar. No se confunde con «no hay marketplaces».
export const ErrorDeCarga: Story = {
  args: {
    estado: "error",
    marketplaces: [],
    error: "GET /api/marketplaces: conexión rechazada (ECONNREFUSED)",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/conexión rechazada/)).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onReintentar).toHaveBeenCalled()
  },
}

// El vacío HONESTO (distinto del error): cero marketplaces es una afirmación verificable, y el
// copy dice DÓNDE se miró para poder afirmarla.
export const NingunoConocido: Story = {
  args: { marketplaces: [] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("No conozco ningún marketplace todavía.")).toBeInTheDocument()
    await expect(c.getByText(/known_marketplaces\.json/)).toBeInTheDocument()
  },
}

// Una lectura en vuelo bloquea su propio botón (sin dos disparos) y lo dice con `aria-busy`.
export const LecturaEnVuelo: Story = {
  args: { marketplaces: [mkNoLeido], leyendo: new Set(["caveman"]) },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Leyendo…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")
  },
}
