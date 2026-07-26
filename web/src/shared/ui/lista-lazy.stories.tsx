import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, userEvent, within } from "storybook/test"
import { ListaLazy } from "./lista-lazy"

// Story = test (fe-visual-fitness) — molécula ListaLazy (AG-D3 + AG-D16). Frame reproduce el
// scope real (.arnesia-portafolio, portafolio.css) para que las clases del pie/scroll resuelvan
// igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

interface FilaDemo {
  clave: string
  texto: string
}

// 273 filas — la cifra REAL del catálogo de `claude-plugins-official` (AG-D16, verificada
// 2026-07-25 sobre el archivo de 159 KB de esta máquina). Se genera en vez de fixturearse
// porque lo que se prueba es el conteo, no el contenido de cada fila (plan-pruebas §1 E-28 lo
// dice explícito: «273 filas generadas»).
const filas273: FilaDemo[] = Array.from({ length: 273 }, (_, i) => ({
  clave: `fila-${i}`,
  texto: `entrada ${i}`,
}))

const meta = {
  title: "shared/ui/ListaLazy",
  component: ListaLazy<FilaDemo>,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    items: filas273,
    claveDe: (f: FilaDemo) => f.clave,
    render: (f: FilaDemo) => <li className="cat-fila">{f.texto}</li>,
    claseLista: "cat-lista",
  },
} satisfies Meta<typeof ListaLazy<FilaDemo>>

export default meta
type Story = StoryObj<typeof meta>

// E-28 — el corazón: 273 filas no se montan de una, y el faltante SE DICE con el número EXACTO.
// El botón de fallback existe porque `IntersectionObserver` puede no disparar (contenedor sin
// scroll) y en jsdom no existe: sin él, un faltante sería un recorte mudo.
export const ListaLazyNoTruncaEnSilencio: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    // montaje inicial acotado al paso (default 20), NO las 273.
    await expect(c.getAllByRole("listitem")).toHaveLength(20)

    // el pie dice el número EXACTO de faltantes — jamás «y algunas más».
    await expect(c.getByText("… 253 más (se cargan al bajar)")).toBeInTheDocument()

    // el botón de fallback monta el paso siguiente.
    const cargar = c.getByRole("button", { name: "Cargar más" })
    await userEvent.click(cargar)
    await expect(c.getAllByRole("listitem")).toHaveLength(40)
    await expect(c.getByText("… 233 más (se cargan al bajar)")).toBeInTheDocument()

    // tras N pasos se montan LAS 273 y el pie desaparece (ya no hay nada que declarar).
    for (let i = 0; i < 12; i++) {
      await userEvent.click(c.getByRole("button", { name: "Cargar más" }))
    }
    await expect(c.getAllByRole("listitem")).toHaveLength(273)
    await expect(c.queryByRole("button", { name: "Cargar más" })).toBeNull()
    await expect(c.queryByText(/más \(se cargan al bajar\)/)).toBeNull()
  },
}

// Cabe entera: sin faltantes no hay pie ni botón — la molécula no agrega ruido cuando no hay
// nada que declarar (mismo criterio que AvisoChip: sin dato, no renderiza).
export const ListaLazyCabeEntera: Story = {
  args: { items: filas273.slice(0, 4) },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByRole("listitem")).toHaveLength(4)
    await expect(c.queryByRole("button", { name: "Cargar más" })).toBeNull()
  },
}

// Paso configurable + etiqueta inyectada: el vocabulario de dominio entra por prop, la molécula
// no conoce ninguno (`shared-no-upward`).
export const ListaLazyPasoPropio: Story = {
  args: {
    items: filas273.slice(0, 10),
    paso: 3,
    etiquetaRestantes: (n: number) => `${n} entradas de catálogo sin cargar`,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getAllByRole("listitem")).toHaveLength(3)
    await expect(c.getByText("7 entradas de catálogo sin cargar")).toBeInTheDocument()
  },
}
