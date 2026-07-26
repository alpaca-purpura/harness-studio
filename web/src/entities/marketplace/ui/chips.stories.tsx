import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import {
  catPrenter,
  mkNoLeido,
  mkPrenterPropio,
  mkSinAcceso,
  mkUrlNoResuelve,
} from "../testing/marketplaces"
import { CanalChip, ClaseChip, EstadoLecturaChip, SituacionChip, ViaChip } from "./chips"

// Story = test (fe-visual-fitness) — chips de dominio de Marketplace (design.md §8.4). Frame
// reproduce el scope real (.arnesia-portafolio, portafolio.css) para que tokens/clases resuelvan
// igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

// AHORA fijo: los textos de lectura son deterministas (el chip recibe `ahora`, no lo lee del
// reloj) — sin esto la story sería flaky por diseño.
const AHORA = new Date("2026-07-25T14:11:33Z")

const meta = {
  title: "entities/marketplace/Chips",
  component: ClaseChip,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { clase: "propio" },
} satisfies Meta<typeof ClaseChip>

export default meta
type Story = StoryObj<typeof meta>

// AG-D8 decisión 1 — las dos clases NO son simétricas y se distinguen ANTES de entrar al
// catálogo. Descriptivas (sin tono de salud: la clase no es sano/enfermo, es qué se puede hacer).
export const ClasesDosNoSimetricas: Story = {
  render: () => (
    <div className="pf-fila-chips">
      <ClaseChip clase="propio" />
      <ClaseChip clase="referencia" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("propio")).toBeInTheDocument()
    await expect(c.getByText("de referencia")).toBeInTheDocument()
    // clases estructurales distintas (los tests asertan CLASE, no color — G9).
    await expect(canvasElement.querySelector(".pf-chip-clase.propio")).not.toBeNull()
    await expect(canvasElement.querySelector(".pf-chip-clase.referencia")).not.toBeNull()
  },
}

// AG-D8 decisión 4 — los 4 estados de lectura, cada uno con su clase propia y su MOTIVO
// completo. Reglas duras de token verificadas por clase: `no-leido` usa `sin-senal` (jamás la de
// `ok`) y `sin-acceso`/`url-no-resuelve` NO usan `crit`.
export const LecturaCuatroEstados: Story = {
  render: () => (
    <div className="pf-fila-chips">
      <EstadoLecturaChip lectura={mkPrenterPropio.lectura} ahora={AHORA} />
      <EstadoLecturaChip lectura={mkNoLeido.lectura} ahora={AHORA} />
      <EstadoLecturaChip lectura={mkSinAcceso.lectura} ahora={AHORA} />
      <EstadoLecturaChip lectura={mkUrlNoResuelve.lectura} ahora={AHORA} />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("leído hace 4 min")).toBeInTheDocument()
    await expect(c.getByText("no leído aún")).toBeInTheDocument()
    await expect(c.getByText(/^sin acceso — gh: HTTP 404/)).toBeInTheDocument()
    await expect(c.getByText(/^url no resuelve — marketplace\.json inválido/)).toBeInTheDocument()

    // «no sé» ≠ «sano»: la clase de no-leido es sin-senal, y NINGÚN chip de lectura es `ok`.
    await expect(canvasElement.querySelector(".pf-chip-lectura.no-leido")).not.toBeNull()
    await expect(canvasElement.querySelector(".pf-chip-lectura.ok")).toBeNull()
    // `--crit` está reservado a fallo: un marketplace sin acceso es recuperable (autenticar).
    await expect(canvasElement.querySelector(".pf-chip-lectura.crit")).toBeNull()
  },
}

// El degradado con caché viejo dice LAS DOS cosas (BR-3 + BR-4 juntas): cuándo se leyó bien la
// última vez Y qué pasa ahora. Nunca se afirma frescura que no hubo.
export const LecturaDegradadaConCacheViejo: Story = {
  render: () => (
    <EstadoLecturaChip
      lectura={{
        tipo: "sin-acceso",
        cuando: "2026-07-23T14:07:33Z",
        entradas: 1,
        motivo: "dial tcp: lookup github.com: no such host",
      }}
      ahora={AHORA}
    />
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(
        "leído hace 2 días · ahora sin acceso — dial tcp: lookup github.com: no such host",
      ),
    ).toBeInTheDocument()
  },
}

// Las 6 ramas de spec §4.3 con su tono. `no-comparable` y `no-lo-tengo` van en `sin-senal`:
// NUNCA la clase de `ok` (regla dura de spec §1, misma que `DotSaludPortafolio`).
export const SituacionSeisRamas: Story = {
  render: () => (
    <div className="pf-fila-chips">
      <SituacionChip situacion={{ tipo: "no-lo-tengo" }} />
      <SituacionChip situacion={{ tipo: "al-hilo" }} />
      <SituacionChip situacion={{ tipo: "mi-copia-adelantada", mia: "0.5.4", estante: "0.5.3" }} />
      <SituacionChip situacion={{ tipo: "estante-adelantado", mia: "0.2.0", estante: "0.3.1" }} />
      <SituacionChip situacion={{ tipo: "instalaciones-en-deriva", cuantas: 2 }} />
      <SituacionChip situacion={{ tipo: "no-comparable", motivo: "sin versión derivable" }} />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("no lo tengo")).toBeInTheDocument()
    await expect(c.getByText("lo tengo · al hilo")).toBeInTheDocument()
    await expect(
      c.getByText("mi copia adelantada — estante v0.5.3 · tu canónico v0.5.4"),
    ).toBeInTheDocument()
    await expect(c.getByText("2 instalaciones en deriva")).toBeInTheDocument()
    await expect(c.getByText("no comparable")).toBeInTheDocument()

    await expect(canvasElement.querySelectorAll(".pf-chip-situacion.ok")).toHaveLength(1)
    await expect(canvasElement.querySelectorAll(".pf-chip-situacion.atencion")).toHaveLength(3)
    await expect(canvasElement.querySelectorAll(".pf-chip-situacion.sin-senal")).toHaveLength(2)
  },
}

// AG-D11 FIRMADA — dos entradas del índice con el MISMO `source` son canales del mismo
// contenido. El chip existe para que nadie crea que son dos arneses distintos. Sin dato, no
// renderiza nada (mismo criterio que AvisoChip).
export const CanalMismoContenido: Story = {
  render: () => (
    <div className="pf-fila-chips">
      <CanalChip otros={catPrenter.entradas?.[0]?.comparte_source_con} />
      <CanalChip otros={[]} />
      <CanalChip />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/mismo contenido que/)).toBeInTheDocument()
    await expect(c.getByText("harness-beta")).toBeInTheDocument()
    // dos de los tres no tienen dato ⇒ un solo chip en el DOM.
    await expect(canvasElement.querySelectorAll(".pf-chip-canal")).toHaveLength(1)
  },
}

// BR-3 — un cruce DÉBIL de identidad se declara como tal; una identidad resuelta
// (`home-declarado`) no pinta nada (no hay nada que advertir).
export const ViaCrucesDebiles: Story = {
  render: () => (
    <div className="pf-fila-chips">
      <ViaChip via="faceta-registry" nombre="harness" />
      <ViaChip via="rename" nombre="convex" nombreAnterior={["convex-backend"]} />
      <ViaChip via="home-declarado" nombre="harness" />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("origen no declarado — cruzado por el registry de la copia"),
    ).toBeInTheDocument()
    await expect(c.getByText("convex-backend")).toBeInTheDocument()
    await expect(c.getByText("convex")).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".pf-chip-via")).toHaveLength(2)
  },
}
