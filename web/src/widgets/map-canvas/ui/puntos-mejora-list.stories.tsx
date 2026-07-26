import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import { PUNTO_B1, PUNTO_B3, PUNTO_P1 } from "@/entities/telemetria"
import { PuntosMejoraList } from "./puntos-mejora-list"

// Story = test (fe-visual-fitness). La lista resuelve H-1 (dónde vive la tarjeta) y H-2 (cómo
// se ve el vacío honesto). Gate a11y en `error`.

const meta = {
  title: "widgets/map-canvas/PuntosMejoraList",
  component: PuntosMejoraList,
  parameters: { layout: "padded" },
  decorators: [(Story) => <div style={{ maxWidth: 860 }}>{Story()}</div>],
  args: {
    puntos: [PUNTO_B1, PUNTO_B3],
    hayDatos: true,
    corridas: 61,
    onDescartar: fn(),
    onProponer: fn(),
    onReintentar: fn(),
  },
} satisfies Meta<typeof PuntosMejoraList>

export default meta
type Story = StoryObj<typeof meta>

// H-1 — la lista vive DEBAJO del canvas, ordenada por ahorro descendente. Lo de más plata
// primero es lo único que ordena sin discutir.
export const DosTarjetasOrdenadas: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const cards = canvasElement.querySelectorAll(".mejora")
    await expect(cards).toHaveLength(2)
    await expect(cards[0]?.textContent).toContain("reescribe el cache")
    await expect(c.getByRole("heading", { name: "Puntos de mejora" })).toBeInTheDocument()
    await expect(canvasElement.querySelector(".mej-lista-n")?.textContent).toBe("2")
  },
}

// RF-246 · A4 — **una tarjeta sin contrafactual no existe**. No se pinta degradada: la lista la
// filtra. Pintarla a medias sería violar la regla que el paquete entero defiende; esconderla
// sin más sería el gap invisible, y por eso el inspector la lista igual (T34).
export const DescartaSinContrafactual: Story = {
  args: {
    hayDatos: true,
    puntos: [
      PUNTO_B1,
      PUNTO_P1,
      {
        ...PUNTO_B3,
        id: "b6-sin-contrafactual",
        titulo: "La sesión de «publicar» quedó abierta sin cerrar",
        contrafactual: null,
      },
    ],
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".mejora")).toHaveLength(2)
    await expect(c.queryByText(/quedó abierta sin cerrar/)).toBeNull()
  },
}

// H-2 — **nunca una sección vacía**: decir cuántos detectores corrieron y sobre cuántas
// corridas es lo único que distingue «no encontramos nada» de «no buscamos», que son
// conclusiones opuestas. Con los seis corriendo, el copy es el literal firmado.
//
// ⚠️ Este vacío SOLO es legal con datos atribuibles (`hayDatos`). Sobre un arnés que nunca
// corrió afirmaba «Hay datos… los seis detectores corrieron sobre 0 corridas» — el defecto que
// la verificación en la app instalada cazó. El candado está en `capa-mejora-coherencia.stories`.
export const VaciaConDatos: Story = {
  args: { puntos: [] },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("Hay datos y ningún punto de mejora que pase el corte."),
    ).toBeInTheDocument()
    await expect(
      c.getByText(
        "Los seis detectores corrieron sobre 61 corridas. Ninguno encontró una fuga que se pueda cotizar y arreglar.",
      ),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelector(".mejora")).toBeNull()
    await expect(canvasElement.querySelector(".mej-vacia")).not.toBeNull()
    // Con los seis corriendo no hay nada que enumerar: la lista de «no pudieron» está vacía.
    await expect(canvasElement.querySelectorAll(".mej-vacia-detectores li")).toHaveLength(0)
  },
}

// design §5.4 — el skeleton tiene la ALTURA de una tarjeta: un placeholder de 20 px seguido de
// dos tarjetas de 148 saltaría el layout entero al resolver.
export const Cargando: Story = {
  args: { estado: "cargando", hayDatos: false },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll("[data-skeleton='mejora']")).toHaveLength(2)
    await expect(c.getByRole("status", { name: "Buscando puntos de mejora" })).toBeInTheDocument()
  },
}

// design §5.4 — el error trae el motivo REAL y el canvas y la franja siguen funcionando.
export const ErrorDeConsulta: Story = {
  args: { estado: "error", hayDatos: false, error: "el almacén de telemetría no está disponible" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/el almacén de telemetría no está disponible/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Reintentar" })
    await expect(btn).toBeEnabled()
    await btn.click()
    await expect(args.onReintentar).toHaveBeenCalledTimes(1)
  },
}

// RF-255 · BR-M12 — la nota al pie dice qué hace el botón ANTES de que lo aprete. Es parte del
// contrato con el operador, no relleno legal.
export const NotaAlPie: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(
        "«Proponerlo en el chat» no escribe archivos: abre el chat con el cambio propuesto, y se aplica por el camino de siempre, con sus permisos y su gate.",
      ),
    ).toBeInTheDocument()
  },
}

// design §6 — con 12 tarjetas la página no scrollea horizontal y ninguna desborda su columna.
export const MuchasTarjetas: Story = {
  args: {
    hayDatos: true,
    puntos: Array.from({ length: 12 }, (_, i) => ({
      ...PUNTO_B1,
      id: `b1-${i}`,
      diferencia_micros: 530_000 - i * 10_000,
    })),
  },
  play: async ({ canvasElement }) => {
    await expect(canvasElement.querySelectorAll(".mejora")).toHaveLength(12)
    await expect(document.body.scrollWidth).toBeLessThanOrEqual(document.body.clientWidth)
    for (const card of canvasElement.querySelectorAll(".mejora")) {
      await expect(card.scrollWidth).toBeLessThanOrEqual(card.clientWidth + 1)
    }
  },
}
