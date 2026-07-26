import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import {
  COBERTURA_ILUSTRATIVA,
  type EstadoDetector,
  PUNTO_B1,
  RESUMEN_ILUSTRATIVO,
  type ResumenTelemetria,
  type Ventana,
} from "@/entities/telemetria"
import { coberturaEsParcial, hayDatosAtribuibles } from "../model/capa-mejora"
import { FranjaMejora } from "./franja-mejora"
import { PuntosMejoraList } from "./puntos-mejora-list"

// 🔒 **CANDADO DE COHERENCIA entre la franja y la lista.**
//
// Nació de un defecto REAL, encontrado mirando la app instalada contra los datos del operador
// (`verificacion-2026-07-26/evidencia/instalada-v0222-capa-mejora.png`), no razonando:
//
//   franja  →  «— —  Este arnés nunca corrió con telemetría.»
//   lista   →  «Hay datos y ningún punto de mejora que pase el corte.
//               Los seis detectores corrieron sobre 0 corridas.»
//
// **Cada bloque era defendible por separado; juntos afirmaban algo falso.** Ninguna story lo vio
// porque cada una renderiza UN componente aislado, y el que compone los dos —el
// composition-root— no tiene stories en este repo.
//
// Este archivo cierra ese agujero con el mismo patrón que `CopyConfianzaEsUnaSola` (candado de
// D18): **los dos bloques en la misma story, derivados de UN solo `resumen`**, con las mismas
// funciones (`hayDatosAtribuibles`) que usa la página. Si alguien cambia la condición en un lado,
// acá se cae.

/** Compone los dos bloques exactamente como lo hace `workspace-stage.tsx`. */
function Vista({
  resumen,
  puntos = [],
  noAplican,
  ventana = "7d",
  ultimaCorridaFuera,
}: {
  resumen: ResumenTelemetria | null
  puntos?: (typeof PUNTO_B1)[]
  noAplican?: EstadoDetector[]
  ventana?: Ventana
  ultimaCorridaFuera?: string
}) {
  return (
    <div>
      <FranjaMejora
        ventana={ventana}
        onVentana={fn()}
        resumen={resumen !== null && resumen.corridas > 0 ? resumen : null}
        ultimaCorridaFuera={ultimaCorridaFuera}
        escenario={resumen?.escenario}
        retencionDias={90}
        retencionPropuesta
        onPolitica={fn()}
        onReintentar={fn()}
      />
      <PuntosMejoraList
        puntos={puntos}
        hayDatos={hayDatosAtribuibles(resumen)}
        noAplican={noAplican}
        corridas={resumen?.corridas ?? 0}
        onDescartar={fn()}
        onProponer={fn()}
        onReintentar={fn()}
      />
    </div>
  )
}

/** Las dos frases que NUNCA pueden convivir con un estado vacío de la franja. */
const FRASES_DE_QUE_HUBO_BUSQUEDA = [/Hay datos/, /detectores corrieron/, /sin fugas/]

const meta = {
  title: "widgets/map-canvas/CapaMejoraCoherencia",
  component: Vista,
  parameters: { layout: "padded" },
} satisfies Meta<typeof Vista>

export default meta
type Story = StoryObj<typeof meta>

// 🔴 **EL DEFECTO, cementado.** Arnés que nunca corrió: la franja dice el estado 1 y la lista
// **no se dibuja**. El DOM completo no puede contener ninguna afirmación de que hubo búsqueda.
export const SinDatosLaListaNoContradiceLaFranja: Story = {
  args: { resumen: null },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Este arnés nunca corrió con telemetría.")).toBeInTheDocument()
    // Las tres afirmaciones de búsqueda, ausentes del DOM ENTERO — no del componente aislado.
    for (const frase of FRASES_DE_QUE_HUBO_BUSQUEDA) {
      await expect(canvasElement.textContent).not.toMatch(frase)
    }
    // Y la sección entera no existe: ni su encabezado, ni su contador en 0, ni su nota al pie.
    await expect(canvasElement.querySelector(".mej-lista")).toBeNull()
    await expect(c.queryByRole("heading", { name: "Puntos de mejora" })).toBeNull()
    await expect(canvasElement.textContent).not.toContain("sobre 0 corridas")
  },
}

// H-9 — «no corrió en estos 7 días» tiene el MISMO problema de composición: hubo corridas
// alguna vez, pero no en esta ventana, así que en esta ventana no hubo búsqueda.
export const VentanaSinCorridasTampocoAfirmaBusqueda: Story = {
  args: { resumen: null, ultimaCorridaFuera: "12/07" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("Sin corridas en los últimos 7 días. La última fue el 12/07."),
    ).toBeInTheDocument()
    for (const frase of FRASES_DE_QUE_HUBO_BUSQUEDA) {
      await expect(canvasElement.textContent).not.toMatch(frase)
    }
    await expect(canvasElement.querySelector(".mej-lista")).toBeNull()
  },
}

// El caso límite que la condición ingenua se comería: **corridas > 0 pero NINGUNA atribuible.**
// Los detectores no pudieron encontrar nada porque no había a qué caja colgarlo.
export const CorridasSinAtribucionNoCuentanComoBusqueda: Story = {
  args: {
    resumen: {
      ...RESUMEN_ILUSTRATIVO,
      corridas: 4,
      cobertura: {
        esperados: 4,
        exacta: 0,
        por_hash: 0,
        por_proceso: 0,
        sin_dato: 4,
        no_llegaron: 0,
      },
    },
  },
  play: async ({ canvasElement }) => {
    // La franja SÍ se muestra (hubo corridas, y eso es un dato)…
    await expect(canvasElement.querySelector(".fm")).not.toBeNull()
    // …pero la lista no afirma que se buscó nada.
    for (const frase of FRASES_DE_QUE_HUBO_BUSQUEDA) {
      await expect(canvasElement.textContent).not.toMatch(frase)
    }
    await expect(canvasElement.querySelector(".mej-lista")).toBeNull()
  },
}

// ✅ **El control positivo.** Con datos atribuibles la sección SÍ se dibuja y SÍ dice que
// corrieron los seis: sin esta story, «esconder siempre la lista» pasaría los tres asserts de
// arriba y nadie se enteraría (§4 del plan, la regla del control positivo).
export const ConDatosLaListaSiAfirmaBusqueda: Story = {
  args: { resumen: RESUMEN_ILUSTRATIVO },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".mej-lista")).not.toBeNull()
    await expect(c.getByText("Hay datos y ningún punto de mejora que pase el corte.")).toBeVisible()
    await expect(canvasElement.textContent).toContain("Los seis detectores corrieron sobre 61")
    await expect(c.queryByText(/nunca corrió/)).toBeNull()
  },
}

// Estado 2 — **la cobertura parcial SÍ tiene datos**: es el caso que una condición basada en
// `resumen.confianza === "sin-dato"` habría escondido por error, porque la confianza del agregado
// es la PEOR de la ventana (`telemetria_service.go:334`). Una sola corrida sin atribuir la deja
// en `sin-dato` con 60 perfectas al lado.
export const CoberturaParcialConservaLaLista: Story = {
  args: {
    resumen: {
      ...RESUMEN_ILUSTRATIVO,
      confianza: "sin-dato",
      corridas: 5,
      costo_reportado_micros: 3_100_000,
      cobertura: {
        esperados: 5,
        exacta: 3,
        por_hash: 0,
        por_proceso: 0,
        sin_dato: 2,
        no_llegaron: 0,
      },
    },
    puntos: [PUNTO_B1],
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(hayDatosAtribuibles(RESUMEN_ILUSTRATIVO)).toBe(true)
    await expect(canvasElement.querySelector(".mejora")).not.toBeNull()
    await expect(c.getByText("2 corridas quedaron sin atribución.")).toBeInTheDocument()
    await expect(c.queryByText(/nunca corrió/)).toBeNull()
  },
}

// El umbral del estado 2 es una decisión de PRODUCTO declarada (más de un tercio sin atribuir),
// no un `corridas <= 5` escondido en el JSX. Acá se asserta el borde por los dos lados.
export const UmbralDeCoberturaParcialEsDeclarado: Story = {
  args: { resumen: RESUMEN_ILUSTRATIVO },
  play: async () => {
    // El juego coherente del mockup: 3 de 61 sin atribuir (5 %) ⇒ alcanza el denominador.
    await expect(coberturaEsParcial(RESUMEN_ILUSTRATIVO)).toBe(false)
    // 2 de 5 (40 %) ⇒ el total necesita su advertencia propia.
    await expect(
      coberturaEsParcial({
        ...RESUMEN_ILUSTRATIVO,
        corridas: 5,
        cobertura: {
          ...COBERTURA_ILUSTRATIVA,
          exacta: 3,
          por_hash: 0,
          por_proceso: 0,
          sin_dato: 2,
        },
      }),
    ).toBe(true)
  },
}
