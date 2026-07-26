import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import {
  catalogo273,
  catCruceDebil,
  catLas6Situaciones,
  catNull,
  catPrenter,
  catReferenciaLas6,
  catTruncado,
  catVacio,
} from "@/entities/marketplace"
import { MarketplaceCatalogo } from "./marketplace-catalogo"

// Story = test (fe-visual-fitness) — widget del catálogo de un marketplace (S3 propio / S4
// referencia read-only). Frame reproduce el scope real (.arnesia-portafolio, portafolio.css).
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

const AHORA = new Date("2026-07-25T14:11:33Z")

const meta = {
  title: "widgets/marketplace/MarketplaceCatalogo",
  component: MarketplaceCatalogo,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    estado: "datos",
    catalogo: catPrenter,
    ahora: AHORA,
    busqueda: "",
    filtroSituacion: new Set(),
    onBusqueda: fn(),
    onFiltroSituacion: fn(),
    onVolver: fn(),
    onRefrescar: fn(),
    onVerFicha: fn(),
    onTraerCanonico: fn(),
    onAbrirCanonico: fn(),
  },
} satisfies Meta<typeof MarketplaceCatalogo>

export default meta
type Story = StoryObj<typeof meta>

// E-20/E-21/E-22/E-23 — cada fila muestra SU situación real y la acción que le corresponde
// (AG-D8 decisión 6): no hay un «Agregar» genérico por fila. Y BR-10 con el literal EXACTO del
// dominio: Publicar/Actualizar/Reparar van `disabled` + `title` — el widget no inventa ese texto
// ni puede habilitarlos (cierra C5: el mockup los tenía habilitados; el spec gana).
export const CatalogoLas6Situaciones: Story = {
  args: { catalogo: catLas6Situaciones },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(c.getAllByRole("listitem")).toHaveLength(6)

    // las 6 situaciones, con las versiones que el WIRE mandó (ninguna cifra tecleada).
    await expect(c.getByText("no lo tengo")).toBeInTheDocument()
    await expect(c.getByText("lo tengo · al hilo")).toBeInTheDocument()
    await expect(
      c.getByText("mi copia adelantada — estante v0.5.3 · tu canónico v0.5.4"),
    ).toBeInTheDocument()
    await expect(
      c.getByText("el estante adelantado — estante v0.5.3 · tu canónico v0.2.0"),
    ).toBeInTheDocument()
    await expect(c.getByText("1 instalación en deriva")).toBeInTheDocument()
    await expect(c.getByText("no comparable")).toBeInTheDocument()

    // BR-10 — los 3 verbos fuera de alcance: `disabled` + el tooltip LITERAL del dominio.
    const publicar = c.getByRole("button", { name: "Publicar" })
    await expect(publicar).toBeDisabled()
    await expect(publicar).toHaveAttribute(
      "title",
      "Publicar se construye en su propio paquete (ítem 3 del outcome)",
    )
    const actualizar = c.getByRole("button", { name: "Actualizar mi copia" })
    await expect(actualizar).toBeDisabled()
    await expect(actualizar).toHaveAttribute(
      "title",
      "Actualizar mi copia se construye en su propio paquete (ítem 4 del outcome)",
    )
    const reparar = c.getByRole("button", { name: "Reparar" })
    await expect(reparar).toBeDisabled()
    await expect(reparar).toHaveAttribute(
      "title",
      "Reparar se construye en su propio paquete (ítem 5 del outcome)",
    )

    // AG-D17 — la ÚNICA celda habilitada de la tabla de §6.3: propio × no-lo-tengo.
    const traer = c.getByRole("button", { name: "↧ Traer canónico" })
    await expect(traer).toBeEnabled()

    // Las dos filas SIN verbo (`al-hilo` y `no-comparable`) dicen «nada que hacer · ver ficha»
    // en vez de un botón genérico apagado: la acción que no aplica NO se pinta.
    await expect(c.getAllByText(/nada que hacer/)).toHaveLength(2)
  },
}

// AG-D17 — la fila habilitada dispara EL MISMO acto que el botón del drawer (dos puertas, un
// acto): `onTraerCanonico` con el nombre EXACTO de la entrada.
export const CatalogoTraerHabilitadoDispara: Story = {
  args: { catalogo: catLas6Situaciones },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "↧ Traer canónico" }))
    await expect(args.onTraerCanonico).toHaveBeenCalledWith("dev-full-cycle")
  },
}

// E-07 / BR-1 — catálogo de clase `referencia`: `↧ Traer canónico` aparece `disabled` con el
// literal EXACTO del mockup firmado, y **NO existe ningún Publicar/Actualizar/Reparar en el DOM**.
// Habilitarlos sería operar arneses de terceros, y eso está muerto por visión.
export const CatalogoReferenciaReadOnly: Story = {
  args: { catalogo: catReferenciaLas6 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    const traer = c.getByRole("button", { name: "↧ Traer canónico" })
    await expect(traer).toBeDisabled()
    await expect(traer).toHaveAttribute("title", "no aplica: solo arneses propios")

    await expect(c.queryByRole("button", { name: /Publicar|Actualizar|Reparar/ })).toBeNull()
    await expect(c.getByText(/Read-only por diseño/)).toBeInTheDocument()
    await expect(c.getByText("de referencia")).toBeInTheDocument()
  },
}

// E-23 / BR-9 — `no-comparable` es rama de PRIMERA CLASE: dot `sin-senal` (nunca el de `ok`) + el
// motivo textual COMPLETO, sin truncar (mismo criterio que AvisoChip).
export const CatalogoNoComparable: Story = {
  args: { catalogo: catLas6Situaciones, filtroSituacion: new Set(["no-comparable"]) },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(c.getAllByRole("listitem")).toHaveLength(1)
    await expect(c.getByText("no comparable")).toBeInTheDocument()
    await expect(
      c.getByText(
        "el catálogo no declara versión de esta entrada ni se puede derivar de su source",
      ),
    ).toBeInTheDocument()

    // el dot es `sin-senal` estructuralmente, jamás la clase de `ok`.
    await expect(c.getByRole("img", { name: "salud: sin señal" })).toBeInTheDocument()
    await expect(c.queryByRole("img", { name: "salud: ok" })).toBeNull()

    // sin versión declarable, la fila lo DICE (no inventa un «v?» que parezca un dato).
    await expect(c.getByText(/sin versión declarada/)).toBeInTheDocument()
  },
}

// E-03 / AG-D11 FIRMADA — las 2 entradas REALES de prenter comparten `source`
// (`./plugins/harness/0.5.3`): son CANALES, no dos arneses. Las dos filas existen (la fila es la
// entrada de índice, que es lo instalable) y cada una lleva el chip que lo aclara.
export const CatalogoCanalesMismoSource: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(c.getAllByRole("listitem")).toHaveLength(2)
    const chips = canvasElement.querySelectorAll(".pf-chip-canal")
    await expect(chips).toHaveLength(2)
    await expect(chips[0]).toHaveTextContent("mismo contenido que harness-beta")
    await expect(chips[1]).toHaveTextContent("mismo contenido que harness")

    // el enriquecimiento REAL de `catalogo.json` (estado del canal) se muestra cuando existe.
    await expect(c.getAllByText("habilitada")).toHaveLength(2)
  },
}

// E-28 / AG-D16 — 273 filas (la cifra REAL del catálogo oficial) se montan de a poco y el faltante
// SE DICE con el número exacto. `ListaLazy` no es confort: es requisito.
// biome-ignore lint/style/useNamingConvention: nombre de story EXIGIDO literal por plan-pruebas.md E-28
export const CatalogoDoscientasSetentaYTres: Story = {
  args: { catalogo: catalogo273() },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByRole("listitem")).toHaveLength(20)
    await expect(c.getByText("… 253 más (se cargan al bajar)")).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Cargar más" }))
    await expect(c.getAllByRole("listitem")).toHaveLength(40)
    // la cabecera dice la cuenta REAL del wire, no la del montaje parcial.
    await expect(c.getByText(/273 entradas de catálogo/)).toBeInTheDocument()
  },
}

// E-31 / BR-4 — `entradas: null` es «no pude leer». La UI muestra el MOTIVO y dice explícitamente
// que eso NO significa que no tenga arneses. Es el mecanismo anti-pass-fabricado del paquete.
export const CatalogoNoLegible: Story = {
  args: { catalogo: catNull },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.getByText("No pude leer el catálogo de este marketplace.")).toBeInTheDocument()
    // el motivo aparece en la cabecera (estado de lectura) y en el cuerpo: el MISMO texto del
    // wire en los dos lados, nunca dos versiones distintas del mismo hecho.
    await expect(c.getAllByText(/gh: HTTP 404 — repo inexistente/).length).toBeGreaterThan(0)
    await expect(c.getByText(/no tenga arneses/)).toBeInTheDocument()

    // cero filas Y cero afirmación sobre cuántos arneses hay.
    await expect(c.queryAllByRole("listitem")).toHaveLength(0)
    await expect(c.queryByText(/entradas de catálogo/)).toBeNull()
    await expect(c.queryByText(/no declara ningún arnés/)).toBeNull()

    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onRefrescar).toHaveBeenCalled()
  },
}

// E-32 / BR-4 — `entradas: []` es una AFIRMACIÓN evidenciada: «leí, y no declara ninguno», con su
// fecha. Distinguible a la vista de `CatalogoNoLegible`.
export const CatalogoVacioReal: Story = {
  args: { catalogo: catVacio },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Este marketplace no declara ningún arnés.")).toBeInTheDocument()
    await expect(c.getAllByText(/leído hace/).length).toBeGreaterThan(0)
    await expect(c.getByText(/plugins\[\]/)).toBeInTheDocument()
    await expect(c.queryByText("No pude leer el catálogo de este marketplace.")).toBeNull()
  },
}

// E-66/E-38 — un cruce DÉBIL de identidad se declara COMO TAL: la UI no presenta
// «cruzado por el registry de la copia» como si el origen estuviera declarado, y un rename dice
// de qué nombre viene.
export const CatalogoCruceDebil: Story = {
  args: { catalogo: catCruceDebil },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("origen no declarado — cruzado por el registry de la copia"),
    ).toBeInTheDocument()
    const rename = canvasElement.querySelectorAll(".pf-chip-via")[1] as HTMLElement
    await expect(rename).toHaveTextContent("renombrado: convex-backend → convex")
  },
}

// E-58 — un recorte del backend es VISIBLE con su número exacto, jamás silencioso.
export const CatalogoTruncadoVisible: Story = {
  args: { catalogo: catTruncado },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent(
      "1000 entradas no cargadas (techo de seguridad)",
    )
  },
}

// AG-D3 — buscar acota el catálogo del lado del cliente; sin match, el estado honesto con un solo
// botón que limpia búsqueda Y filtros.
export const CatalogoBuscaAcota: Story = {
  args: { catalogo: catLas6Situaciones, busqueda: "harness" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getAllByRole("listitem")).toHaveLength(1)
    const buscador = c.getByRole("searchbox", { name: "Buscar en el catálogo" })
    await expect(buscador).toHaveValue("harness")
    await userEvent.type(buscador, "-zzz")
    await expect(args.onBusqueda).toHaveBeenCalled()
  },
}

export const CatalogoSinResultados: Story = {
  args: { catalogo: catLas6Situaciones, busqueda: "zzz-no-existe" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Ninguna entrada del catálogo coincide/)).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Limpiar búsqueda y filtros" }))
    await expect(args.onBusqueda).toHaveBeenCalledWith("")
    await expect(args.onFiltroSituacion).toHaveBeenCalledWith(new Set())
  },
}

// El filtro de situación solo ofrece las ramas PRESENTES, con la misma palabra que la fila.
export const CatalogoFiltroSituacion: Story = {
  args: { catalogo: catPrenter },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const filtros = within(c.getByRole("group", { name: "Filtros del catálogo" }))
    const btn = filtros.getByRole("button", { name: "Situación" })
    await expect(btn).toHaveAttribute("aria-expanded", "false")
    await userEvent.click(btn)
    await expect(btn).toHaveAttribute("aria-expanded", "true")
    // catPrenter tiene 2 ramas: estante-adelantado y no-lo-tengo (orden canónico).
    const panel = within(c.getByRole("group", { name: "Filtrar por situación" }))
    await expect(panel.getByRole("button", { name: "no lo tengo" })).toBeInTheDocument()
    await userEvent.click(panel.getByRole("button", { name: "el estante adelantado" }))
    await expect(args.onFiltroSituacion).toHaveBeenCalledWith(new Set(["estante-adelantado"]))
  },
}

// E-107 — estado `trayendo` (§13.10): botón `disabled` + `aria-busy`, la fila entera marcada como
// ocupada. No se puede disparar dos veces la misma.
export const CatalogoTrayendo: Story = {
  args: {
    catalogo: catLas6Situaciones,
    traer: { "dev-full-cycle": { fase: "trayendo" } },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Trayendo…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")
    await expect(c.queryByRole("button", { name: "↧ Traer canónico" })).toBeNull()
    // la fila entera queda no-interactiva y lo declara.
    const fila = btn.closest(".cat-fila") as HTMLElement
    await expect(fila).toHaveAttribute("aria-busy", "true")
  },
}

// E-107 · BR-17 — un Traer exitoso muestra el veredicto de deriva TAL CUAL salga: un
// `en-deriva` recién traído es un dato honesto, no un fallo de la operación.
export const CatalogoTraido: Story = {
  args: {
    catalogo: catLas6Situaciones,
    traer: {
      "dev-full-cycle": {
        fase: "traido",
        destino: "/home/chalreme/.arnesia/checkouts/prenter-marketplace/dev-full-cycle",
        camino: "local",
        deriva: "en-deriva",
        derivaDetalle: "hash de contenido distinto de la referencia del marketplace",
        clavePortafolio: "github-com-alpacapurpura-prenter-marketplace~dev-full-cycle~",
      },
    },
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/deriva: en-deriva/)).toBeInTheDocument()
    await expect(
      c.getByText("/home/chalreme/.arnesia/checkouts/prenter-marketplace/dev-full-cycle"),
    ).toBeInTheDocument()
    await expect(c.getByText(/hash de contenido distinto/)).toBeInTheDocument()
    // «ver ficha» también existe en la fila `al-hilo` (nada que hacer · ver ficha): se scopea a
    // la fila traída, que es la que este caso ejercita.
    const filaTraida = within(c.getByText(/deriva: en-deriva/).closest(".cat-fila") as HTMLElement)
    await userEvent.click(filaTraida.getByRole("button", { name: "ver ficha" }))
    await expect(args.onVerFicha).toHaveBeenCalledWith(
      "github-com-alpacapurpura-prenter-marketplace~dev-full-cycle~",
    )
  },
}

// E-107 — estado `falló`: el motivo textual REAL del backend + `Reintentar`; y en el 409 (BR-14,
// destino ya poblado) además el botón para abrir el canónico que ya tenés.
export const CatalogoTraerFallo: Story = {
  args: {
    catalogo: catLas6Situaciones,
    traer: {
      "dev-full-cycle": {
        fase: "fallo",
        motivo:
          "POST /api/marketplaces/prenter-marketplace/traidos: 409 traer: el destino ya tiene contenido",
        destinoExistente: "/home/chalreme/.arnesia/checkouts/prenter-marketplace/dev-full-cycle",
      },
    },
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("409 traer: el destino ya tiene contenido")
    await userEvent.click(c.getByRole("button", { name: "Reintentar" }))
    await expect(args.onTraerCanonico).toHaveBeenCalledWith("dev-full-cycle")
    await userEvent.click(c.getByRole("button", { name: "abrir el canónico que ya tenés" }))
    await expect(args.onAbrirCanonico).toHaveBeenCalledWith(
      "/home/chalreme/.arnesia/checkouts/prenter-marketplace/dev-full-cycle",
    )
  },
}

// G5 — cargando / error del GET del catálogo (distinto de `lectura.motivo`, que viaja DENTRO de
// un 200): siempre queda la puerta de vuelta al plano.
export const Cargando: Story = {
  args: { estado: "cargando", catalogo: undefined },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status", { name: "Cargando catálogo" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "← Marketplaces" })).toBeInTheDocument()
  },
}

export const ErrorDelPedido: Story = {
  args: {
    estado: "error",
    catalogo: undefined,
    error: "GET /api/marketplaces/x/catalogo: 404 marketplace: nombre no conocido",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/404 marketplace: nombre no conocido/)).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "← Marketplaces" }))
    await expect(args.onVolver).toHaveBeenCalled()
  },
}
