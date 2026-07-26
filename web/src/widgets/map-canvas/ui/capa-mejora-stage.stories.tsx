import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import { devFullCycle } from "@/entities/arnes"
import {
  CAJAS_ILUSTRATIVAS,
  PUNTO_B1,
  PUNTO_B3,
  RESUMEN_ILUSTRATIVO,
  type ResumenTelemetria,
} from "@/entities/telemetria"
import { CapaMejoraStage } from "./capa-mejora-stage"

// 🔒 **EL CANDADO DE COMPOSICIÓN, sobre las CINCO superficies que hablan del mismo hecho**:
// franja · carril · nodo · lista (+ el inspector, que entra como slot).
//
// El candado anterior (`capa-mejora-coherencia.stories.tsx`) montaba **dos de los cinco**, y por
// eso no vio ninguno de los cuatro críticos de `auditoria-tramo-b.md`. Éste monta el widget REAL
// que la página renderiza, así que lo que se prueba acá es lo que el operador ve.
//
// Además cubre lo que ninguna story de componente puede ver: **el layout de la composición**
// (C-1), porque el defecto no estaba en ningún componente sino en cómo convivían.

const FRASES_DE_QUE_HUBO_BUSQUEDA = [/Hay datos/, /detectores corrieron/]

/** El wire de `s2-instrumentado`: `corridas: 0` por construcción, todo lo demás poblado. */
const S2_INSTRUMENTADO: ResumenTelemetria = {
  ...RESUMEN_ILUSTRATIVO,
  escenario: "s2-instrumentado",
  corridas: 0,
  turnos: 58,
  costo_reportado_micros: 1_920_000,
  cobertura: {
    esperados: null,
    exacta: 58,
    por_hash: 0,
    por_proceso: 0,
    sin_dato: 3,
    no_llegaron: null,
  },
}

const NUNCA_CORRIO: ResumenTelemetria = {
  ...RESUMEN_ILUSTRATIVO,
  corridas: 0,
  turnos: 0,
  sesiones: 0,
  cajas: 0,
  costo_reportado_micros: null,
  costo_calculado_micros: null,
  cobertura: {
    esperados: null,
    exacta: 0,
    por_hash: 0,
    por_proceso: 0,
    sin_dato: 0,
    no_llegaron: null,
  },
}

/** Marco con altura FIJA: sin esto no se puede medir el reparto entre canvas y lista. */
function Marco({ children }: { children: React.ReactNode }) {
  return <div style={{ height: 620, display: "flex", flexDirection: "column" }}>{children}</div>
}

const meta = {
  title: "widgets/map-canvas/CapaMejoraStage",
  component: CapaMejoraStage,
  parameters: { layout: "fullscreen", a11y: { test: "todo" } },
  decorators: [(Story) => <Marco>{Story()}</Marco>],
  args: {
    capa: "mejora",
    graph: devFullCycle,
    resumen: RESUMEN_ILUSTRATIVO,
    cajas: CAJAS_ILUSTRATIVAS,
    puntos: [PUNTO_B1, PUNTO_B3],
    noAplican: [],
    ventana: "7d",
    retencionDias: 90,
    retencionPropuesta: true,
    onVentana: fn(),
    onPolitica: fn(),
    onReintentar: fn(),
    onDescartar: fn(),
    onProponer: fn(),
  },
} satisfies Meta<typeof CapaMejoraStage>

export default meta
type Story = StoryObj<typeof meta>

const altoDelCanvas = (el: HTMLElement) =>
  (el.querySelector(".mapa-canvas-slot") as HTMLElement).getBoundingClientRect().height

// 🔴 **C-1 · encender la capa no puede destruir el Mapa.** Medido antes del fix: el canvas pasaba
// de 739 px a **257 px con una tarjeta** y a **0 px con cuatro**. La vista Mapa dejaba de mostrar
// mapa — ni un carril, ni una caja.
export const CapaEncendidaConservaElMapa: Story = {
  play: async ({ canvasElement }) => {
    const alto = altoDelCanvas(canvasElement)
    await expect(alto).toBeGreaterThanOrEqual(240)
    // Y el mapa está REALMENTE ahí, no solo la caja que lo contendría.
    await expect(canvasElement.querySelectorAll(".lane").length).toBeGreaterThan(0)
    await expect(canvasElement.querySelectorAll(".node").length).toBeGreaterThan(0)
  },
}

// C-1 · el caso que lo llevaba a 0 px. Cuatro tarjetas no es un borde: el mockup firmado dibuja
// dos y `MuchasTarjetas` existe como story.
export const CuatroTarjetasNoAplastanElMapa: Story = {
  args: {
    puntos: Array.from({ length: 4 }, (_, i) => ({
      ...PUNTO_B1,
      id: `b1-${i}`,
      diferencia_micros: 530_000 - i * 1000,
    })),
  },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelectorAll(".mejora")).toHaveLength(4)
    await expect(altoDelCanvas(canvasElement)).toBeGreaterThanOrEqual(240)
    // La lista se acota y scrollea DENTRO de sí misma en vez de empujar al canvas afuera.
    const lista = canvasElement.querySelector(".mapa-lista-slot") as HTMLElement
    await expect(lista.getBoundingClientRect().height).toBeLessThanOrEqual(420)
    await expect(lista.scrollHeight).toBeGreaterThan(lista.clientHeight)
  },
}

// C-1 · a 200 % de zoom rompía con UNA sola tarjeta. El piso del canvas cede un poco en
// pantallas cortas, pero nunca desaparece.
export const ZoomAltoConservaElMapa: Story = {
  decorators: [
    (Story) => (
      <div style={{ height: 450, display: "flex", flexDirection: "column" }}>{Story()}</div>
    ),
  ],
  args: { puntos: [PUNTO_B1] },
  play: async ({ canvasElement }) => {
    await expect(altoDelCanvas(canvasElement)).toBeGreaterThanOrEqual(180)
    await expect(canvasElement.querySelectorAll(".node").length).toBeGreaterThan(0)
  },
}

// 🔴 **C-2 · el escenario mayoritario.** Fuera de S1 el wire manda `corridas: 0` con el dinero,
// la cobertura y las cajas poblados. Antes: la franja decía «nunca corrió», el carril cobraba
// USD 1,08 y la lista se escondía. Los tres bloques, ahora, dicen lo mismo.
export const S2InstrumentadoNoDiceNuncaCorrio: Story = {
  args: { resumen: S2_INSTRUMENTADO },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(S2_INSTRUMENTADO.corridas).toBe(0)
    // 1 · la franja NO dice «nunca corrió» y sí alcanza su rama de s2 (antes inalcanzable).
    await expect(c.queryByText(/nunca corrió con telemetría/)).toBeNull()
    await expect(c.getByText("corrió fuera de ArnesIA, instrumentado")).toBeInTheDocument()
    // 2 · el canvas sigue mostrando el dinero, que es verdad.
    await expect(canvasElement.querySelectorAll(".mej-cifra").length).toBeGreaterThan(0)
    // 3 · y la lista de puntos —el entregable del paquete— NO se esconde.
    await expect(canvasElement.querySelectorAll(".mejora").length).toBeGreaterThan(0)
    // 4 · el denominador del vacío no puede decir «0 corridas».
    await expect(canvasElement.textContent).not.toContain("sobre 0 corridas")
  },
}

// 🔴 **C-2 (la otra mitad) · si el estado 1 manda, el canvas NO puede seguir cobrando.** Éste es
// el defecto de D24 llevado a las cinco superficies: antes, con el resumen vacío, el nodo y el
// carril seguían pintando cifras porque se alimentaban de otra consulta.
export const SinDatosNingunBloqueAfirmaNada: Story = {
  args: { resumen: NUNCA_CORRIO, puntos: [] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Este arnés nunca corrió con telemetría.")).toBeInTheDocument()
    for (const frase of FRASES_DE_QUE_HUBO_BUSQUEDA) {
      await expect(canvasElement.textContent).not.toMatch(frase)
    }
    // Canvas, carril y nodo: ninguna cifra, ningún total, ninguna marca de fuga.
    await expect(canvasElement.querySelector(".mej-cifra")).toBeNull()
    await expect(canvasElement.querySelector(".mej-fuga")).toBeNull()
    await expect(canvasElement.querySelector(".lane-usd")).toBeNull()
    // Lista: ausente.
    await expect(canvasElement.querySelector(".mej-lista")).toBeNull()
    // Y el mapa sigue siendo un mapa.
    await expect(canvasElement.querySelectorAll(".node").length).toBeGreaterThan(0)
  },
}

// ✅ **Control positivo.** Con datos, los cinco bloques SÍ afirman — si no, «esconder todo
// siempre» pasaría los asserts de arriba y nadie se enteraría (§4 del plan).
export const ConDatosLosCincoBloquesAfirman: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".fm")).not.toBeNull()
    await expect(canvasElement.querySelectorAll(".mej-cifra").length).toBe(4)
    await expect(canvasElement.querySelector(".lane-usd")).not.toBeNull()
    await expect(canvasElement.querySelectorAll(".mejora").length).toBe(2)
    await expect(c.queryByText(/nunca corrió/)).toBeNull()
  },
}

// La capa apagada no deja rastro en ninguno de los cinco bloques (RF-245 · BR-M16), y el canvas
// recupera toda la altura.
export const CapaApagadaNoDejaRastro: Story = {
  args: { capa: "estructura" },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelector(".fm")).toBeNull()
    await expect(canvasElement.querySelector(".mej-lista")).toBeNull()
    await expect(canvasElement.querySelector(".mej-cifra")).toBeNull()
    await expect(canvasElement.querySelector(".lane-usd")).toBeNull()
    await expect(altoDelCanvas(canvasElement)).toBeGreaterThan(500)
  },
}

// A-2 · el chip de reenvío externo se dibuja cuando `/salud` dice que está encendido. Antes NO
// se dibujaba nunca, porque nadie consumía ese endpoint: una promesa de privacidad repetida en
// pantalla sin poder saber si los datos se estaban reenviando afuera (D13 · H-7).
export const ReenvioExternoSeDeclara: Story = {
  args: { forwardDestino: "otlp.datadoghq.com" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("reenvío externo encendido → otlp.datadoghq.com")).toBeInTheDocument()
  },
}

// 🔴 **C-3 · un GET de detalle que falla NO se pinta como dato.** El test vive acá y no en las
// stories de `InspectorMejora` a propósito: el componente ya tenía su rama de error y su story
// **y aun así el defecto existía**, porque la decisión —«¿esto es un error o son datos?»— la
// tomaba la página. Reverti el cableado y la story del componente seguía verde; ésta no.
export const DetalleRotoNoAfirmaSobreElRuntime: Story = {
  args: {
    cajaSeleccionada: { esCaja: true, motivoNoCaja: "" },
    detalle: null,
    detalleError: "Failed to fetch",
    inspector: (cuerpo) => <div data-testid="tab-mejora">{cuerpo}</div>,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const tab = canvasElement.querySelector("[data-testid='tab-mejora']") as HTMLElement
    await expect(tab).not.toBeNull()
    // El motivo REAL y el reintento.
    await expect(within(tab).getByText(/Failed to fetch/)).toBeInTheDocument()
    await expect(within(tab).getByRole("button", { name: "Reintentar" })).toBeEnabled()
    // Y ninguna de las ocho afirmaciones que inventaba: seis «no aplica», el catálogo y las 0
    // corridas. `detalle == null` ≠ «este runtime no tiene estos conceptos».
    await expect(within(tab).queryByText("no aplica en este runtime")).toBeNull()
    await expect(within(tab).queryByText("catálogo sin construir")).toBeNull()
    await expect(tab.textContent).not.toContain("0 corridas")
    await expect(c.queryByText(/«No aplica» no es 0/)).toBeNull()
  },
}

// ✅ Control positivo del anterior: con detalle bueno, la 4ª tab SÍ afirma.
export const DetalleBuenoSiAfirma: Story = {
  args: {
    cajaSeleccionada: { esCaja: true, motivoNoCaja: "" },
    detalle: {
      turnos_totales: 14,
      tokens: { entrada: 12_400, salida: 8_900 },
      paridad: {
        reportado_micros: 1_920_000,
        calculado_micros: 1_920_000,
        divergencia_pct: 0,
        completo: true,
        catalogo_version: "2026-07-20",
      },
      detectores: [],
    },
    inspector: (cuerpo) => <div data-testid="tab-mejora">{cuerpo}</div>,
  },
  play: async ({ canvasElement }) => {
    const tab = canvasElement.querySelector("[data-testid='tab-mejora']") as HTMLElement
    await expect(within(tab).getByText("Tokens · 7 días (14 corridas)")).toBeInTheDocument()
    await expect(within(tab).queryByText(/Reintentar/)).toBeNull()
  },
}
