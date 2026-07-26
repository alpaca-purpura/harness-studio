import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import {
  FILA_SIN_DATO,
  FILAS_PORTAFOLIO_ILUSTRATIVAS,
  type FilaPortafolio,
} from "@/entities/telemetria"
import { TablaMejoraPortafolio } from "./tabla-mejora-portafolio"

// Story = test (fe-visual-fitness). RF-265…268 · H-12 · D20 · D21.
//
// 🔴 **Este archivo tiene el gate axe en `error`** (no hereda ningún `todo`), así que es acá
// donde el fix de contraste de D21 es load-bearing: el chip de fuga grave con `--crit` sobre
// `--crit-soft` daría 4,04:1 en tema claro y **rompería el build**.

const meta = {
  title: "widgets/portafolio/TablaMejoraPortafolio",
  component: TablaMejoraPortafolio,
  parameters: { layout: "padded" },
  args: { filas: FILAS_PORTAFOLIO_ILUSTRATIVAS },
} satisfies Meta<typeof TablaMejoraPortafolio>

export default meta
type Story = StoryObj<typeof meta>

// RF-265 — la celda dice `0,31` **sin prefijo**: el `USD` vive en el encabezado, que es lo que
// permite alinear la columna sin repetir tres letras en cada fila.
export const ConDato: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("columnheader", { name: "USD/corrida" })).toBeInTheDocument()
    await expect(c.getByRole("columnheader", { name: "tendencia" })).toBeInTheDocument()
    await expect(c.getByRole("columnheader", { name: "punto de mejora" })).toBeInTheDocument()
    const celda = canvasElement.querySelector(".pf-mej-usd") as HTMLElement
    await expect(celda.textContent).toBe("0,31")
    await expect(celda.textContent).not.toContain("USD")
    await expect(getComputedStyle(celda).fontVariantNumeric).toContain("tabular-nums")
  },
}

// D21 ítem 3 — el espejo en oscuro (story NUEVA respecto de plan-storybook: sin ella el gate
// mira solo el tema claro, y `--crit`/`--crit-soft` en oscuro pasa raspando a 4,66:1).
export const ConDatoDark: Story = {
  globals: { theme: "dark" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/re-warm de cache/)).toBeInTheDocument()
    await expect(document.documentElement.dataset["theme"]).toBe("dark")
  },
}

// RF-265 · D20 — **un arnés en dos puestos se cumple POR INSTALACIÓN**, no por puesto: dos
// instalaciones del mismo arnés dan dos filas, cada una con su propio costo.
export const UnArnesDosPuestos: Story = {
  play: async ({ canvasElement }) => {
    const filas = [...canvasElement.querySelectorAll(".pf-mej-fila")].filter((f) =>
      f.textContent?.includes("acme-cli"),
    )
    await expect(filas).toHaveLength(2)
    const costos = filas.map((f) => f.querySelector(".pf-mej-usd")?.textContent)
    await expect(costos[0]).not.toBe(costos[1])
    const inst = filas.map((f) => f.getAttribute("data-instalacion"))
    await expect(new Set(inst).size).toBe(2)
  },
}

// RF-265 · D20 — **nadie fabrica un puesto.** Es el caso NORMAL de hoy: ningún arnés del dogfood
// declara `rol`, así que la cláusula «en este puesto» se renderiza así en el 100 % de los casos.
export const SinPuestoDeclarado: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByText("puesto sin declarar").length).toBe(3)
    await expect(c.queryByText(/Product Owner/)).toBeNull()
    await expect(c.queryByText(/Tech Lead/)).toBeNull()
  },
}

// RF-267 · RF-280 — «sin fugas detectadas» es **distinguible de «sin dato»**: son conclusiones
// opuestas (una midió y no encontró; la otra no midió), y se distinguen por clase, no por tono.
export const SinFugas: Story = {
  args: { filas: [...FILAS_PORTAFOLIO_ILUSTRATIVAS, FILA_SIN_DATO] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const ok = c.getAllByText("sin fugas detectadas")[0]?.closest(".pf-mej-chip") as HTMLElement
    const nada = c.getByText("nunca corrió con telemetría")
    await expect(ok).toHaveClass("pf-mej-chip-ok")
    await expect(nada).toHaveClass("pf-mej-chip-neutro")
    await expect(ok.textContent).not.toBe(nada.textContent)
  },
}

// RF-268 — **la fila SIGUE en la tabla**. Sacarla haría desaparecer del inventario a un arnés
// que existe; ponerle 0 diría que corrió gratis.
export const SinDato: Story = {
  args: { filas: [FILA_SIN_DATO] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const fila = canvasElement.querySelector(".pf-mej-fila") as HTMLElement
    await expect(fila).not.toBeNull()
    await expect(fila.querySelector(".pf-mej-usd")?.textContent).toBe("sin dato")
    await expect(within(fila).queryByText("0,00")).toBeNull()
    await expect(
      within(fila).getByLabelText("sin tendencia: nunca corrió con telemetría"),
    ).toBeInTheDocument()
    await expect(c.getByText("nunca corrió con telemetría")).toBeInTheDocument()
  },
}

// RF-266 — con una sola corrida no hay tendencia que afirmar, y se DICE. Una línea plana diría
// «estable», que es una afirmación que nadie midió.
export const PocasCorridas: Story = {
  args: {
    filas: [
      { ...(FILAS_PORTAFOLIO_ILUSTRATIVAS[0] as FilaPortafolio), corridas: 1, serie: [240_000] },
    ],
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("pocas corridas para una tendencia")).toBeInTheDocument()
    await expect(canvasElement.querySelector(".spark")).toBeNull()
  },
}

// RF-268 — al ordenar por costo, los «sin dato» **se agrupan al final con separador rotulado**.
// Nunca intercalados como si valieran 0: un arnés que nunca corrió no es el más barato.
export const OrdenadaSinDatoAlFinal: Story = {
  args: {
    filas: [
      ...FILAS_PORTAFOLIO_ILUSTRATIVAS,
      FILA_SIN_DATO,
      { ...FILA_SIN_DATO, arnes_id: "otro", clave: "x~otro", instalacion_id: "inst-otro" },
    ],
    ordenarPorCosto: true,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const filas = [...canvasElement.querySelectorAll(".pf-mej-fila")]
    await expect(filas).toHaveLength(5)
    const valores = filas
      .map((f) => f.querySelector(".pf-mej-usd")?.textContent ?? "")
      .map((t) => (t === "sin dato" ? null : Number.parseFloat(t.replace(",", "."))))
    const conDato = valores.filter((v): v is number => v !== null)
    await expect(conDato).toHaveLength(3)
    for (let i = 1; i < conDato.length; i += 1) {
      await expect(conDato[i - 1] as number).toBeGreaterThanOrEqual(conDato[i] as number)
    }
    await expect(c.getByText("Sin datos de telemetría")).toBeInTheDocument()
    const idxUltimoConDato = valores.findLastIndex((v) => v !== null)
    const idxPrimeroSinDato = valores.indexOf(null)
    await expect(idxPrimeroSinDato).toBeGreaterThan(idxUltimoConDato)
  },
}

// H-12 — el pie de tabla, y la MISMA marca de duda que el canvas en la fila `por-hash`: la
// pieza es la misma (`MarcaConfianza` de `entities/telemetria`), no una copia.
export const PieConDisclaimerDeEstimacion: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(
        "Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no aparecen en cero.",
      ),
    ).toBeInTheDocument()
    await expect(c.getByText("por huella")).toBeInTheDocument()
    await expect(c.getByText("por proceso")).toBeInTheDocument()
  },
}

// design §6.3 — con ids largos y montos grandes, la columna elástica se encoge primero y **el
// `<body>` no scrollea horizontal**: el scroll vive en `.pf-mej-scroll`, alcanzable por teclado.
export const NombresLargosNumerosGrandes: Story = {
  args: {
    filas: [
      {
        ...(FILAS_PORTAFOLIO_ILUSTRATIVAS[0] as FilaPortafolio),
        arnes_id: "arnes-de-ciclo-completo-para-el-equipo-de-plataforma-de-alpacapurpura",
        costo_por_corrida: 12_345_670_000,
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const scroll = canvasElement.querySelector(".pf-mej-scroll") as HTMLElement
    await expect(scroll).toHaveAttribute("tabindex", "0")
    await expect(scroll).toHaveAttribute("role", "region")
    await expect(scroll).toHaveAttribute("aria-label")
    await expect(document.body.scrollWidth).toBeLessThanOrEqual(document.body.clientWidth)
  },
}

// H-3 — el puente al Mapa: la fila abre el Mapa de ese arnés (la página le agrega la capa
// Mejora activa y la caja seleccionada). Reusa «Abrir en Mapa», que ya existe (GAP-1).
export const AbreElMapaDeEsaInstalacion: Story = {
  args: { onAbrirEnMapa: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await c.getAllByRole("button")[0]?.click()
    await expect(args.onAbrirEnMapa).toHaveBeenCalledWith("alpacapurpura~vitalia", "inst-vitalia-1")
  },
}
