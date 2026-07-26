import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import {
  BUCKETS_ILUSTRATIVOS,
  BUCKETS_MEDIDOS,
  DETECTORES_MVP,
  DETECTORES_NO_MEDIDOS,
  MOTIVO_SIN_DATO,
  PARIDAD_COINCIDEN,
  PARIDAD_DIVERGEN,
} from "@/entities/telemetria"
import { InspectorMejora } from "./inspector-mejora"

// Story = test (fe-visual-fitness). RF-258…264 · RF-277. El drawer real mide **340 px**
// (`inspector.css:13`), no los 380 de la iteración 1 del mockup: el frame lo reproduce para que
// el desbordamiento que se mida sea el real.
function Frame({ children }: { children: ReactNode }) {
  return (
    <div className="arnesia-inspector" style={{ width: 340 }}>
      {children}
    </div>
  )
}

const join = {
  corridas: 14,
  rechazadas: 3,
  costo_rechazadas_micros: 270_000,
  rotaciones: 2,
}

const meta = {
  title: "widgets/map-canvas/InspectorMejora",
  component: InspectorMejora,
  parameters: { layout: "padded" },
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    esCaja: true,
    ventanaLabel: "7 días",
    corridas: 14,
    buckets: BUCKETS_ILUSTRATIVOS,
    totalMicros: 1_920_000,
    paridad: PARIDAD_COINCIDEN,
    join,
    detectores: DETECTORES_MVP,
    noMedidos: [],
    onReintentar: () => {},
  },
} satisfies Meta<typeof InspectorMejora>

export default meta
type Story = StoryObj<typeof meta>

/** Lee los USD de las 6 filas de bucket y los suma en micros. Assert ARITMÉTICO, no de texto. */
function sumaBuckets(canvasElement: HTMLElement): number {
  let total = 0
  for (const tr of canvasElement.querySelectorAll("tbody tr[data-bucket]")) {
    const monto = tr.querySelector(".num-celda:last-child .num-monto")?.textContent
    if (!monto) continue
    total += Math.round(Number.parseFloat(monto.replace(/\s/gu, "").replace(",", ".")) * 1_000_000)
  }
  return total
}

// RF-259 — el desglose completo, y **la suma cierra**: si las seis filas no suman el total de
// la caja, el número de arriba es una afirmación sin respaldo.
export const Completo: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Tokens · 7 días (14 corridas)")).toBeInTheDocument()
    const filas = canvasElement.querySelectorAll("tbody tr[data-bucket]")
    await expect(filas).toHaveLength(6)
    await expect([...filas].map((f) => f.querySelector("th")?.textContent)).toEqual([
      "entrada",
      "salida",
      "cache · lectura",
      "cache · escritura 5 m",
      "cache · escritura 1 h",
      "razonamiento",
    ])
    // **La suma cierra**: las seis filas suman el total de la caja, 1,92. Assert ARITMÉTICO
    // sobre los valores mostrados, no de texto: si el desglose no cierra con el total, el
    // número de arriba es una afirmación sin respaldo.
    await expect(sumaBuckets(canvasElement)).toBe(1_920_000)
    // Las cuatro secciones, por su título.
    for (const t of [
      "Tokens · 7 días (14 corridas)",
      "Costo — reportado vs. calculado",
      "El join — corridas de esta caja",
      "Detectores",
    ]) {
      await expect(c.getByText(t)).toBeInTheDocument()
    }
  },
}

// RF-259 · plan-storybook §3.2 — la corrida REAL del 2026-07-26, con sus 18 473 micros. Existe
// para que el inspector se pruebe contra el payload MEDIDO y no solo contra el juego ilustrativo:
// a dos decimales cada bucket redondea, y por eso el assert acá es de FORMA (seis filas, el cero
// y la ausencia en su lugar), no de suma.
export const CorridaRealMedida: Story = {
  args: { buckets: BUCKETS_MEDIDOS, totalMicros: 18_473, corridas: 1 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll("tbody tr[data-bucket]")).toHaveLength(6)
    await expect(c.getByText("17 536")).toBeInTheDocument()
    await expect(c.getByText("8 257")).toBeInTheDocument()
    const cero = canvasElement.querySelector("tr[data-bucket='cache_escritura_5m']") as HTMLElement
    await expect(within(cero).getByText("0")).toBeInTheDocument()
  },
}

// RF-260 · RF-286 — «no aplica» ocupa las DOS columnas numéricas. Un guion en cada celda se
// leería como cero, y un cero donde el concepto no existe sería mentira.
export const NoAplicaNoEsCero: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const fila = canvasElement.querySelector("tr[data-bucket='razonamiento']") as HTMLElement
    const celda = fila.querySelector("td") as HTMLTableCellElement
    await expect(celda.colSpan).toBe(2)
    await expect(celda.textContent).toBe("no aplica en este runtime")
    await expect(within(fila).queryByText("0")).toBeNull()
    await expect(
      c.getByText("«No aplica» no es 0. Un cero donde el concepto no existe sería mentira."),
    ).toBeInTheDocument()
  },
}

// RF-260 · RF-286 — el otro lado de la moneda: **`0` es un dato válido**. El fixture usa el cero
// REAL medido (`ephemeral_5m_input_tokens: 0`), y su fila tiene estructura DOM distinta de la de
// «no aplica»: dos celdas, sin `colspan`.
export const CeroLegitimo: Story = {
  play: async ({ canvasElement }) => {
    const cero = canvasElement.querySelector("tr[data-bucket='cache_escritura_5m']") as HTMLElement
    const noAplica = canvasElement.querySelector("tr[data-bucket='razonamiento']") as HTMLElement
    await expect(cero.querySelectorAll("td")).toHaveLength(2)
    await expect(within(cero).getByText("0")).toBeInTheDocument()
    await expect(within(cero).getByText("0,00")).toBeInTheDocument()
    await expect((cero.querySelector("td") as HTMLTableCellElement).colSpan).toBe(1)
    // Las dos ausencias se distinguen por ESTRUCTURA, no solo por texto.
    await expect(cero.querySelectorAll("td").length).not.toBe(
      noAplica.querySelectorAll("td").length,
    )
  },
}

// RF-261 · RF-280 — el veredicto es TEXTO. El ✓ es decorativo y redundante con la palabra.
export const ParidadCoinciden: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const paridad = canvasElement.querySelector(".mej-paridad") as HTMLElement
    await expect(paridad.textContent).toContain("runtime")
    await expect(paridad.textContent).toContain("nuestro catálogo")
    // Los dos números, los dos visibles.
    await expect(paridad.textContent?.match(/1,92/gu)).toHaveLength(2)
    await expect(canvasElement.querySelector(".mej-veredicto")?.textContent).toContain("coinciden")
    await expect(c.queryByText(/difieren/)).toBeNull()
  },
}

// RF-261 — **los dos números siguen visibles y la UI NO ELIGE.** Si difieren, o el catálogo está
// viejo o el runtime cambió su tarifa: ninguna de las dos es «la correcta», y el DOM no lo dice.
export const ParidadDivergen: Story = {
  args: { paridad: PARIDAD_DIVERGEN },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".mej-veredicto")?.textContent).toContain(
      "difieren en USD 0,18",
    )
    const paridad = canvasElement.querySelector(".mej-paridad") as HTMLElement
    await expect(paridad.textContent).toContain("1,92")
    await expect(paridad.textContent).toContain("2,10")
    await expect(
      c.getByText(
        "Guardamos los dos. Si difieren, o nuestro catálogo está viejo o el runtime cambió su tarifa.",
      ),
    ).toBeInTheDocument()
    await expect(c.queryByText(/correcto|válido/i)).toBeNull()
  },
}

// RF-261 · RF-272 — un runtime que no reporta costo: el número calculado se muestra IGUAL,
// rotulado como calculado. Ocultarlo perdería la única cifra disponible.
export const SinCostoDelRuntime: Story = {
  args: {
    paridad: { ...PARIDAD_COINCIDEN, reportado_micros: null, divergencia_pct: null },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("este runtime no reporta costo — calculado con el catálogo v2026-07-20"),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelector(".mej-paridad")?.textContent).toContain("1,92")
  },
}

// design §3.4 — **el estado REAL de hoy**: el catálogo embebido todavía no se construyó. Se
// dice; el costo reportado se muestra igual, y no se fabrica un `0,00` de relleno.
export const CatalogoSinConstruir: Story = {
  args: {
    paridad: {
      reportado_micros: 1_920_000,
      calculado_micros: null,
      divergencia_pct: null,
      completo: false,
      catalogo_sin_construir: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("catálogo sin construir")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".mej-paridad")?.textContent).toContain("1,92")
    await expect(canvasElement.querySelector(".mej-paridad")?.textContent).not.toContain("0,00")
  },
}

// RF-262 — el join: cuántas corridas, cuántas se rechazaron, cuánto costaron las rechazadas y
// cuántas rotaciones. Es la mitad de proceso de la frase del producto.
export const JoinCompleto: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const filas = [...canvasElement.querySelectorAll("tbody tr")].filter(
      (t) => !t.hasAttribute("data-bucket"),
    )
    await expect(filas).toHaveLength(4)
    await expect(c.getByText("corridas")).toBeInTheDocument()
    await expect(c.getByText("rechazadas en el gate")).toBeInTheDocument()
    await expect(c.getByText("rotaciones de contexto")).toBeInTheDocument()
    await expect(canvasElement.textContent).toMatch(/USD\s*0,27/)
  },
}

// RF-262 — sin señal de gate, la fila lo DICE. **`0` diría «ninguna se rechazó»**, que es una
// afirmación sobre el proceso que nadie midió. Las filas de dinero se muestran igual.
export const SinSenalDeGate: Story = {
  args: {
    join: { corridas: 14, rechazadas: null, costo_rechazadas_micros: null, rotaciones: 2 },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByText("sin señal de gate en estas corridas").length).toBeGreaterThan(0)
    const fila = canvasElement.querySelector("tr[data-fila='rechazadas']") as HTMLElement
    await expect(within(fila).queryByText("0")).toBeNull()
    await expect(canvasElement.querySelectorAll("tbody tr[data-bucket]")).toHaveLength(6)
  },
}

// RF-263 · RF-280 — cada detector tiene su ESTADO EN TEXTO; el punto de color es refuerzo y
// está `aria-hidden`. Tres formas posibles: activo, sin hallazgos, o no disponible con motivo.
export const SeisDetectoresConEstado: Story = {
  play: async ({ canvasElement }) => {
    const items = canvasElement.querySelectorAll(".mej-detectores li")
    await expect(items).toHaveLength(6)
    for (const li of items) {
      const linea = li.querySelector(".det-linea")?.textContent ?? ""
      const motivo = li.querySelector(".det-motivo")?.textContent ?? ""
      const ok =
        /· activo$/.test(linea) || /· sin hallazgos$/.test(linea) || /^no disponible: /.test(motivo)
      await expect(ok).toBe(true)
    }
    await expect(canvasElement.querySelectorAll(".det-dot[aria-hidden='true']")).toHaveLength(6)
  },
}

// RF-271 · T-09 — **el detector NO se esconde** cuando no puede correr: aparece con su motivo.
// Esconderlo sería el gap invisible que este paquete existe para no crear.
export const DetectorB1ApagadoEnS2: Story = {
  args: {
    detectores: DETECTORES_MVP.map((d) =>
      d.detector === "b1-rewarm-por-ttl"
        ? {
            ...d,
            aplica: false,
            hallazgos: 0,
            motivo: "sin el result del stream-json — este arnés corrió fuera de ArnesIA",
          }
        : d,
    ),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/re-warm/)).toBeInTheDocument()
    await expect(
      c.getByText(
        "no disponible: sin el result del stream-json — este arnés corrió fuera de ArnesIA",
      ),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".mej-detectores li")).toHaveLength(6)
    await expect(canvasElement.querySelectorAll(".det-motivo")).toHaveLength(1)
  },
}

// RF-263 — un detector que corre PERO NO VE TODO viaja con su matiz explícito, jamás como un ✅
// liso: «vi 1 de 3 rotaciones» y «vi todas» son dos cosas distintas.
export const DetectorB2Parcial: Story = {
  args: {
    detectores: DETECTORES_MVP.map((d) =>
      d.detector === "b2-costo-de-la-rotacion"
        ? {
            ...d,
            aplica: false,
            cobertura_parcial: true,
            motivo: "2 de 3 rotaciones ocurrieron fuera de ArnesIA",
          }
        : d,
    ),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("no disponible: 2 de 3 rotaciones ocurrieron fuera de ArnesIA"),
    ).toBeInTheDocument()
  },
}

// RF-263 — los siete de fuera del MVP dicen `no medido todavía`. **`sin hallazgos` está
// prohibido ahí**: «no lo miramos» y «lo miramos y está limpio» son conclusiones opuestas.
export const SieteNoMedidos: Story = {
  args: { noMedidos: DETECTORES_NO_MEDIDOS },
  play: async ({ canvasElement }) => {
    const bloque = canvasElement.querySelector(".mej-nomedidos") as HTMLElement
    const items = bloque.querySelectorAll("li")
    await expect(items).toHaveLength(7)
    for (const li of items) await expect(li.textContent).toContain("no medido todavía")
    await expect(within(bloque).queryByText(/sin hallazgos/)).toBeNull()
  },
}

// RF-246 — la contraparte de `DescartaSinContrafactual`: el hallazgo sin fix **no se esconde**,
// el inspector lo lista. Filtrarlo de la lista Y no listarlo acá sería perderlo.
export const DetectorSinFixPropuesto: Story = {
  args: {
    detectores: DETECTORES_MVP.map((d) =>
      d.detector === "b6-sesion-abandonada" ? { ...d, hallazgos: 1, sin_fix: true } : d,
    ),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin fix propuesto")).toBeInTheDocument()
    await expect(c.getByText(/sesión abandonada · activo/)).toBeInTheDocument()
  },
}

// RF-264 · J-3 — la ventana es **la de la capa**, heredada. Las «14 corridas» son el
// DENOMINADOR, no la ventana: el string `últimas 14 corridas` sería una ventana inventada.
export const VentanaHeredada: Story = {
  args: { ventanaLabel: "30 días" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Tokens · 30 días (14 corridas)")).toBeInTheDocument()
    await expect(canvasElement.textContent).not.toContain("últimas 14 corridas")
  },
}

// RF-258 — un nodo que no es caja recibe UNA sección con el motivo, el MISMO que el canvas ya
// le mostró. Cuatro tablas vacías serían cuatro invitaciones a buscar un dato que no existe.
export const NodoNoCaja: Story = {
  args: { esCaja: false, motivoNoCaja: MOTIVO_SIN_DATO["mcp"] as string },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/^Esta capa mide cajas\./)).toBeInTheDocument()
    await expect(canvasElement.querySelector(".mej-nocaja")?.textContent).toBe(
      `Esta capa mide cajas. ${MOTIVO_SIN_DATO["mcp"]}`,
    )
    await expect(canvasElement.querySelectorAll("table")).toHaveLength(0)
    await expect(c.queryByText("0")).toBeNull()
  },
}

// design §5.5 — cargando: cada sección con su skeleton y **los títulos ya visibles**. Un
// skeleton sin títulos no dice qué está por llegar.
export const Cargando: Story = {
  args: { estado: "cargando" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Detectores")).toBeInTheDocument()
    await expect(c.getByText("El join — corridas de esta caja")).toBeInTheDocument()
    await expect(
      canvasElement.querySelectorAll("[data-skeleton='inspector']").length,
    ).toBeGreaterThan(0)
  },
}

// design §5.5 — el error vive DENTRO de la tab; las otras tres siguen intactas.
export const ErrorDeConsulta: Story = {
  args: { estado: "error", error: "el almacén de telemetría no está disponible" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/el almacén de telemetría no está disponible/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeEnabled()
  },
}

// 🔴 **C-3 · un GET que falla NO se pinta como dato.** La página se tragaba el error del detalle
// y nunca pasaba `estado="error"`, así que los defaults hacían el resto: la 4ª tab afirmaba
// **ocho cosas falsas** —«0 corridas», seis «no aplica en este runtime» y «catálogo sin
// construir»— con la nota «"No aplica" no es 0» desplegada EN DEFENSA de la mentira, mientras el
// nodo de al lado mostraba USD 1,08 y /salud devolvía 694 modelos.
export const ErrorNoSePintaComoNoAplica: Story = {
  args: {
    estado: "error",
    error: "Failed to fetch",
    corridas: 0,
    buckets: BUCKETS_ILUSTRATIVOS.map((b) => ({ ...b, tokens: null, costo_micros: null })),
    paridad: {
      reportado_micros: null,
      calculado_micros: null,
      divergencia_pct: null,
      completo: false,
      catalogo_sin_construir: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // El motivo real, y el reintento.
    await expect(c.getByText(/Failed to fetch/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeEnabled()
    // Y NINGUNA de las ocho afirmaciones falsas.
    await expect(c.queryByText("no aplica en este runtime")).toBeNull()
    await expect(c.queryByText("catálogo sin construir")).toBeNull()
    await expect(canvasElement.textContent).not.toContain("0 corridas")
    await expect(canvasElement.textContent).not.toContain("«No aplica» no es 0")
  },
}

// C-3 (bonus) · sin dos números que comparar no hay veredicto, y **sin veredicto no hay tono**.
// Antes la caja se pintaba ámbar «difieren» sin una palabra que lo explicara.
export const ParidadSinVeredictoNoSePintaComoDivergencia: Story = {
  args: {
    paridad: {
      reportado_micros: 1_920_000,
      calculado_micros: null,
      divergencia_pct: null,
      completo: false,
      catalogo_sin_construir: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const caja = canvasElement.querySelector(".mej-paridad") as HTMLElement
    await expect(caja).toHaveClass("sin-veredicto")
    await expect(caja).not.toHaveClass("difieren")
    await expect(canvasElement.querySelector(".mej-veredicto")).toBeNull()
    await expect(c.queryByText(/difieren/)).toBeNull()
  },
}

// M-7 · `catálogo v` colgante: sin versión no se promete una.
export const SinVersionDeCatalogoNoQuedaColgando: Story = {
  args: {
    paridad: {
      reportado_micros: null,
      calculado_micros: 740_000,
      divergencia_pct: null,
      completo: true,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("este runtime no reporta costo — y no llegó la versión del catálogo"),
    ).toBeInTheDocument()
    await expect(canvasElement.textContent).not.toMatch(/catálogo v(?![0-9])/)
  },
}
