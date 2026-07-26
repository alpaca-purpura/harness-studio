import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, within } from "storybook/test"
import { DETECTORES_MVP, PUNTO_B1, PUNTO_B3, PUNTO_P1 } from "@/entities/telemetria"
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

// H-2 — **nunca una sección vacía**. Listar los seis detectores que corrieron es lo único que
// distingue «no encontramos nada» de «no buscamos», que son conclusiones opuestas.
export const VaciaConDatos: Story = {
  args: { puntos: [], detectores: DETECTORES_MVP },
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
    await expect(canvasElement.querySelectorAll(".mej-vacia-detectores li")).toHaveLength(6)
    await expect(canvasElement.querySelector(".mejora")).toBeNull()
    await expect(canvasElement.querySelector(".mej-vacia")).not.toBeNull()
  },
}

// design §5.4 — el skeleton tiene la ALTURA de una tarjeta: un placeholder de 20 px seguido de
// dos tarjetas de 148 saltaría el layout entero al resolver.
export const Cargando: Story = {
  args: { estado: "cargando" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll("[data-skeleton='mejora']")).toHaveLength(2)
    await expect(c.getByRole("status", { name: "Buscando puntos de mejora" })).toBeInTheDocument()
  },
}

// design §5.4 — el error trae el motivo REAL y el canvas y la franja siguen funcionando.
export const ErrorDeConsulta: Story = {
  args: { estado: "error", error: "el almacén de telemetría no está disponible" },
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
