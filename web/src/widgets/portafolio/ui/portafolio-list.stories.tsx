import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, within } from "storybook/test"
import { type EntradaCorrupta, entradasDemo } from "@/entities/portafolio"
import {
  FILA_SIN_DATO,
  FILAS_PORTAFOLIO_ILUSTRATIVAS,
  type FilaPortafolio,
} from "@/entities/telemetria"
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
    filtroSalud: new Set(),
    filtroMarketplace: new Set(),
    onLente: fn(),
    onBusqueda: fn(),
    onFiltroSalud: fn(),
    onFiltroMarketplace: fn(),
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
// reales · click/Enter en fila llaman onAbrir(clave) (G6/G8) · filtros Estado/Marketplace
// disclosure (S1-D8, cerrada 2026-07-24).
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

    // S1-D8 (cerrada 2026-07-23) — las 4 lentes están vivas.
    const lenteProyecto = c.getByRole("button", { name: "Proyecto" })
    await expect(lenteProyecto).not.toBeDisabled()
    await expect(lenteProyecto).toHaveAttribute("aria-pressed", "false")
    await userEvent.click(lenteProyecto)
    await expect(args.onLente).toHaveBeenCalledWith("proyecto")

    // S1-D8 (cerrada 2026-07-24) — filtros Estado/Marketplace: disclosure cerrado por defecto;
    // abrir pinta chips reales (Estado = 3 valores fijos, Marketplace = registries presentes en
    // los datos). Afordancia DISTINTA de la lente: seleccionar una chip ACOTA, no reagrupa.
    // Scope al grupo "Filtros…" — "Marketplace" es también el nombre de una lente (grupo
    // distinto), getByRole sin scope sería ambiguo.
    const filtros = within(c.getByRole("group", { name: "Filtros del Portafolio" }))
    const filtroEstadoBtn = filtros.getByRole("button", { name: "Estado" })
    await expect(filtroEstadoBtn).toHaveAttribute("aria-expanded", "false")
    await userEvent.click(filtroEstadoBtn)
    await expect(filtroEstadoBtn).toHaveAttribute("aria-expanded", "true")
    const chipAtencion = c.getByRole("button", { name: "atención" })
    await userEvent.click(chipAtencion)
    await expect(args.onFiltroSalud).toHaveBeenCalledWith(new Set(["atencion"]))

    // un solo panel abierto a la vez: abrir Marketplace cierra Estado.
    const filtroMktBtn = filtros.getByRole("button", { name: "Marketplace" })
    await userEvent.click(filtroMktBtn)
    await expect(filtroEstadoBtn).toHaveAttribute("aria-expanded", "false")
    const chipRegistry = c.getByRole("button", {
      name: "github.com/alpacapurpura/prenter-marketplace",
    })
    await userEvent.click(chipRegistry)
    await expect(args.onFiltroMarketplace).toHaveBeenCalledWith(
      new Set(["github.com/alpacapurpura/prenter-marketplace"]),
    )
  },
}

// Filtro Estado acota SIN reagrupar (distinta afordancia de la lente): 1 de las 3 filas queda
// (la provisional, «sin señal»); "atención" desaparece del DOM entero, no solo se oculta.
export const FiltroEstadoAcota: Story = {
  args: { entradas: entradasDemo, filtroSalud: new Set(["sin-senal"]) },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".pf-fila").length).toBe(1)
    await expect(c.getByRole("img", { name: "salud: sin señal" })).toBeInTheDocument()
    await expect(c.queryByRole("img", { name: "salud: atención" })).toBeNull()
  },
}

// Filtro Marketplace acota por registriesDe(e) — mismo dato que la lente marketplace, distinta
// afordancia: acá "github.com/acme/acme-cli" aísla la fila de acme-cli sola.
export const FiltroMarketplaceAcota: Story = {
  args: { entradas: entradasDemo, filtroMarketplace: new Set(["github.com/acme/acme-cli"]) },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".pf-fila").length).toBe(1)
    await expect(c.getByText("Acme CLI")).toBeInTheDocument()
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
// «Limpiar» borra búsqueda Y ambos filtros a la vez — un solo botón, un solo estado vacío
// (no distinguimos SI la búsqueda o un filtro causó el cero: cualquiera de los dos disparó lo
// mismo, así que limpiar todo junto es lo honesto y lo simple).
export const FiltroSinResultados: Story = {
  args: { entradas: entradasDemo, busqueda: "zzz-no-existe-en-ninguna-fixture" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Ningún arnés coincide/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Limpiar búsqueda y filtros" })
    await userEvent.click(btn)
    await expect(args.onBusqueda).toHaveBeenCalledWith("")
    await expect(args.onFiltroSalud).toHaveBeenCalledWith(new Set())
    await expect(args.onFiltroMarketplace).toHaveBeenCalledWith(new Set())
  },
}

// E-27 / AG-D4 — cierra el defecto **L1** de la auditoría: los rótulos de grupo son TEXTO
// VISIBLE en el DOM (no solo un `aria-label` que un ojo no ve), y los `role="group"` con sus
// `aria-label` existentes SIGUEN presentes: se agrega afordancia visual sin degradar a11y.
// Las dos «Marketplace» (lente y filtro) quedan dentro de grupos rotulados DISTINTOS, así que se
// leen sin ambigüedad — y ninguna se renombra (eso rompería vocabulario firmado en el Slice 1).
export const ToolbarRotulosVisibles: Story = {
  args: { entradas: entradasDemo },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    // VISIBLES en el DOM (el CSS los pone en mayúsculas; el dato dice «Ver por»).
    await expect(c.getByText("Ver por")).toBeVisible()
    await expect(c.getByText("Filtros")).toBeVisible()

    // los `role="group"` + `aria-label` del Slice 1 se CONSERVAN (cero degradación a11y).
    const lentes = c.getByRole("group", { name: "Lente del Portafolio" })
    const filtros = c.getByRole("group", { name: "Filtros del Portafolio" })
    await expect(lentes).toBeInTheDocument()
    await expect(filtros).toBeInTheDocument()

    // las DOS «Marketplace» existen, cada una en su grupo rotulado: ya no son 6 pills iguales.
    await expect(within(lentes).getByRole("button", { name: "Marketplace" })).toBeInTheDocument()
    await expect(within(filtros).getByRole("button", { name: "Marketplace" })).toBeInTheDocument()

    // L3 — el hint del mockup que el código había perdido.
    await expect(c.getByText("lente para encontrar un arnés cuando hay muchos")).toBeInTheDocument()
  },
}

// AG-D8 decisión 7 — el filtro «sin origen resuelto» que prende el contador cruzado del plano
// Marketplaces se DECLARA en la superficie: un filtro invisible que acota la lista es
// indistinguible de «no tengo arneses». Y salir de él es un click.
export const FiltroSinOrigenSeDeclara: Story = {
  args: { entradas: entradasDemo, filtroSinOrigen: true, onFiltroSinOrigen: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/solo los arneses sin origen resuelto/)).toBeInTheDocument()
    // entradasDemo: 2 de las 3 son provisionales (home vacío) ⇒ sobreviven 2.
    await expect(canvasElement.querySelectorAll(".pf-fila")).toHaveLength(2)
    await userEvent.click(c.getByRole("button", { name: "Ver todos" }))
    await expect(args.onFiltroSinOrigen).toHaveBeenCalledWith(false)
  },
}

// ══ Capa «Mejora» en la fila (T36 · reconstruido tras C-4) ════════════════════════════════
//
// 🔴 Estas stories reemplazan a las de `TablaMejoraPortafolio`, que era **código muerto**: la
// tabla no se montaba en ningún lado y la superficie real era una segunda implementación de las
// mismas celdas, escrita en la página y sin el guard de D24.4. Ahora hay UNA implementación, y
// las stories corren sobre la que el operador ve.
//
// Las 3 celdas son OPCIONALES: las 13 stories firmadas del Slice 1 pasan sin tocarlas, y
// `SinMejoraDomIntacto` es el guardián.

const filaConDato: FilaPortafolio = {
  ...(FILAS_PORTAFOLIO_ILUSTRATIVAS[0] as FilaPortafolio),
  clave: entradasDemo[0]?.clave ?? "",
}
const mejoraDemo = new Map<string, FilaPortafolio>([[filaConDato.clave, filaConDato]])

// RF-265 · BR-M16 — las 3 celdas se insertan **entre chips y dot de salud**: el dot sigue
// CERRANDO la fila, que es el ancla visual que la PARIDAD del Slice 1 firmó. Y el comportamiento
// de la fila no cambia: sigue siendo un `<button>` que llama `onAbrir(clave)`.
export const ConMejora: Story = {
  args: { entradas: entradasDemo, lente: "plano", mejora: mejoraDemo },
  play: async ({ canvasElement, args }) => {
    const fila = canvasElement.querySelector(".pf-fila") as HTMLElement
    const hijos = [...fila.children]
    const iChips = hijos.findIndex((h) => h.classList.contains("pf-fila-chips"))
    const iUsd = hijos.findIndex((h) => h.classList.contains("pf-mej-usd"))
    const iDot = hijos.findIndex((h) => h.classList.contains("pf-dot-salud"))
    await expect(iChips).toBeGreaterThanOrEqual(0)
    await expect(iUsd).toBeGreaterThan(iChips)
    await expect(iDot).toBeGreaterThan(iUsd)
    await expect(iDot).toBe(hijos.length - 1)
    await userEvent.click(fila)
    await expect(args.onAbrir).toHaveBeenCalledWith(entradasDemo[0]?.clave)
  },
}

// A-6 — la celda carga su UNIDAD. Sin encabezado de columna, un `0,31` pelado entre chips no
// dice ni moneda, ni período, ni que es por corrida; y el pie H-12 impide que se lea como
// facturación. Los dos se habían perdido al no montar la tabla.
export const ConMejoraDeclaraUnidadPie: Story = {
  args: {
    entradas: entradasDemo,
    lente: "plano",
    // `por-hash` a propósito: es el caso que H-12 nombra. Con `exacta` la marca NO se dibuja —
    // la ausencia ES la señal (RF-242)— y el assert de abajo no probaría nada.
    mejora: new Map([[filaConDato.clave, { ...filaConDato, confianza: "por-hash" as const }]]),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("0,31")).toBeInTheDocument()
    await expect(c.getByText("USD/corrida")).toBeInTheDocument()
    await expect(
      c.getByText(
        "Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no aparecen en cero.",
      ),
    ).toBeInTheDocument()
    // D23 · H-12 — la MISMA marca de duda que el canvas, que la superficie real no tenía.
    await expect(canvasElement.querySelector(".pf-fila [data-confianza]")).not.toBeNull()
  },
}

// 🔴 **C-4 · el ✓ mentiroso.** Con `puntos_de_mejora: 3` y sin punto principal, la superficie
// real pintaba `✓ sin fugas detectadas` en verde como el elemento más saliente del renglón —
// mintiendo «acá no hay nada que mirar», la dirección que D24 nombra como la más cara.
export const ConHallazgosNoPintaElTilde: Story = {
  args: {
    entradas: entradasDemo,
    lente: "plano",
    mejora: new Map([
      [filaConDato.clave, { ...filaConDato, puntos_de_mejora: 3, punto: undefined }],
    ]),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/3 punto\(s\) de mejora/)).toBeInTheDocument()
    await expect(c.queryByText("sin fugas detectadas")).toBeNull()
    await expect(canvasElement.querySelector(".pf-mej-chip-ok")).toBeNull()
  },
}

// RF-267 — el ✓ SÍ aparece cuando el backend afirma que midió y no encontró nada. Control
// positivo: sin esto, «nunca pintar el ✓» pasaría la story de arriba.
export const SinFugasPintaElTilde: Story = {
  args: {
    entradas: entradasDemo,
    lente: "plano",
    mejora: new Map([
      [filaConDato.clave, { ...filaConDato, puntos_de_mejora: 0, punto: undefined }],
    ]),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin fugas detectadas")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".pf-mej-chip-ok")).not.toBeNull()
  },
}

// 🔴 **A-5 · «nunca corrió» sobre un arnés con 47 corridas.** `costo_por_corrida` es null fuera
// de S1 por construcción, así que derivarlo de ahí afirmaba lo contrario del payload. `corridas`
// está en el wire y es el campo que responde la pregunta.
export const CorridasSinCostoNoEsNuncaCorrio: Story = {
  args: {
    entradas: entradasDemo,
    lente: "plano",
    mejora: new Map([
      [filaConDato.clave, { ...filaConDato, corridas: 47, costo_por_corrida: null }],
    ]),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText("nunca corrió con telemetría")).toBeNull()
    await expect(c.getByText("sin dato")).toBeInTheDocument()
  },
}

// RF-268 — el que de verdad nunca corrió sí lo dice, y la celda de tendencia va marcada.
export const NuncaCorrioLoDice: Story = {
  args: {
    entradas: entradasDemo,
    lente: "plano",
    mejora: new Map([[filaConDato.clave, { ...FILA_SIN_DATO, clave: filaConDato.clave }]]),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("nunca corrió con telemetría")).toBeInTheDocument()
    await expect(
      c.getByLabelText("sin tendencia: este arnés no registra corridas medidas"),
    ).toBeInTheDocument()
    await expect(c.queryByText("0,00")).toBeNull()
  },
}

// BR-M16 — **el guardián del superset**: sin las props nuevas la fila vuelve a sus 5 celdas
// exactas y no queda ni una clase `.pf-mej-*` en el DOM.
export const SinMejoraDomIntacto: Story = {
  args: { entradas: entradasDemo, lente: "plano" },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector(".pf-mej-usd")).toBeNull()
    await expect(canvasElement.querySelector(".pf-mej-tend")).toBeNull()
    await expect(canvasElement.querySelector(".pf-mej-punto")).toBeNull()
    await expect(canvasElement.querySelector(".pf-mej-pie")).toBeNull()
    const fila = canvasElement.querySelector(".pf-fila") as HTMLElement
    await expect(fila.children).toHaveLength(5)
    await expect(fila).not.toHaveClass("con-mejora")
  },
}
