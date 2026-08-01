import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, within } from "storybook/test"
import type { Arquetipo, Box, GateTipo } from "../model/types"
import { ArnesNode } from "./arnes-node"

// Story = test (fe-visual-fitness). The node styles live in map.css scoped under
// `.arnesia-map`, so every story is wrapped in that root + a 232px lane-cell to see the real
// look. Variants mirror design.md §5-6 (caja · support · compact · puesto · propuesto · dim).
function Cell({ children }: { children: ReactNode }) {
  return (
    <div className="arnesia-map">
      <div style={{ width: 232 }}>{children}</div>
    </div>
  )
}

const cajaBox: Box = {
  id: "spec-writer",
  clase: "skill",
  nombre: "escribir el spec",
  banda: "fase",
  fase: "spec",
  estado: "idea -> spec",
  canal: "beta",
  procedencia: "medido",
  // Full classification so the doctrinal meta-row (arquetipo · perfil · gate) renders (Fase 2).
  contract: { caja: true, arquetipo: "excepcion", perfil_harness: "T2", gate: { tipo: "manual" } },
}

const meta = {
  title: "entities/arnes/ArnesNode",
  component: ArnesNode,
  // a11y `todo`: the node carries intentional low-contrast micro-UI of the signed design — the
  // dim state (opacity .4, RF-33), the caja/propuesto badges and the transition tag (9px, tinted
  // over color-mix). These are contract, not regressions; surfaced as todo. The DOM `play`
  // assertions still gate CI.
  parameters: { a11y: { test: "todo" } },
  decorators: [(Story) => <Cell>{Story()}</Cell>],
  args: {
    box: { id: "reviewer", clase: "skill", nombre: "revisar el build", banda: "fase" },
  },
} satisfies Meta<typeof ArnesNode>

export default meta
type Story = StoryObj<typeof meta>

export const Skill: Story = {
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("revisar el build")).toBeInTheDocument()
    // The type is carried by the derived handle (/id), not a text label.
    await expect(within(canvasElement).getByText("/reviewer")).toBeInTheDocument()
  },
}

export const Caja: Story = {
  args: { box: cajaBox },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("caja")).toBeInTheDocument()
    // Transition normalized from the real estado "idea -> spec".
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
    // Doctrinal meta-row (Fase 2): the 3 classification marks are on the card, not just inspector.
    await expect(c.getByText("T2")).toBeInTheDocument() // perfil_harness pill
    await expect(c.getByText("≈")).toBeInTheDocument() // arquetipo excepcion char-badge
    await expect(c.getByLabelText(/gate manual/)).toBeInTheDocument() // gate dot (skill-blue)
  },
}

// The eval-gate honesty (A4): a caja whose gate is `none` must SURFACE the hole (hollow crit dot
// + a "hallazgo" label), never hide it. Also exercises arquetipo=abierto/perfil=T3/procedencia
// dashed border. This is the strongest doctrinal mark on the card.
export const CajaGateNone: Story = {
  args: {
    box: {
      id: "draft-caja",
      clase: "skill",
      nombre: "redactar el borrador",
      banda: "fase",
      fase: "draft",
      estado: "investigado -> borrador",
      procedencia: "estimado",
      contract: { caja: true, arquetipo: "abierto", perfil_harness: "T3", gate: { tipo: "none" } },
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByLabelText(/gate none — SIN eval/)).toBeInTheDocument()
    await expect(c.getByText("T3")).toBeInTheDocument()
    await expect(c.getByText("✳")).toBeInTheDocument() // abierto
  },
}

// The determinista end: pipeline · T1 · auto — the other extreme of the classification axes.
export const CajaPipelineAuto: Story = {
  args: {
    box: {
      id: "publish-caja",
      clase: "skill",
      nombre: "publicar",
      banda: "fase",
      fase: "publish",
      estado: "aprobado -> publicado",
      procedencia: "medido",
      contract: { caja: true, arquetipo: "pipeline", perfil_harness: "T1", gate: { tipo: "auto" } },
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("T1")).toBeInTheDocument()
    await expect(c.getByText("═")).toBeInTheDocument() // pipeline
    await expect(c.getByLabelText(/gate auto/)).toBeInTheDocument()
  },
}

export const Rule: Story = {
  args: { box: { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" } },
  play: async ({ canvasElement }) => {
    // rule with unknown alw → always-on handle (mockup faithful).
    await expect(within(canvasElement).getByText("always-on")).toBeInTheDocument()
  },
}

export const Support: Story = {
  args: {
    box: { id: "test-author", clase: "subagent", nombre: "escribir pruebas", banda: "fase" },
    support: true,
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("@test-author")).toBeInTheDocument()
  },
}

export const Compact: Story = {
  args: {
    box: { id: "guard-format", clase: "hook", nombre: "formato pre-write", banda: "guardia" },
    compact: true,
  },
}

export const Puesto: Story = {
  args: {
    box: { id: "domain-glossary", clase: "rule", nombre: "glosario de dominio", banda: "base" },
  },
}

export const Propuesto: Story = {
  args: {
    box: {
      id: "indice-semantico",
      clase: "mcp",
      nombre: "índice semántico",
      banda: "base",
      canal: "propuesto",
    },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("propuesto")).toBeInTheDocument()
  },
}

export const Dim: Story = { args: { dim: true } }
export const Selected: Story = { args: { selected: true } }

// A no-reconocido node stays VISIBLE with the warn KIND (D-c, nomenclatura §4.5) — before
// the KIND entry existed this crashed the whole canvas into the ErrorBoundary.
export const NoReconocido: Story = {
  args: {
    box: { id: "misterio", clase: "no-reconocido", nombre: "misterio (no reconocido)" },
  },
  play: async ({ canvasElement }) => {
    await expect(within(canvasElement).getByText("misterio (no reconocido)")).toBeInTheDocument()
  },
}

// Deuda F (§4.5) — facetas FUERA del enum degradan SU marca, jamás el nodo ni el lienzo. Los
// grafos llegan como JSON en runtime; los casts simulan ese dato sucio. Repro real del dogfood:
// `arquetipo: guiado` (forja developer-vitalia) tumbaba el canvas ENTERO al ErrorBoundary.
// arquetipo desconocido → «?» warn que NOMBRA el valor · gate desconocido → anillo warn hollow
// (≠ parcial warn lleno, ≠ none crit hollow: no puede disfrazarse de ninguno).
export const CajaFacetasNoReconocidas: Story = {
  args: {
    box: {
      id: "forjada",
      clase: "skill",
      nombre: "caja forjada a mano",
      banda: "fase",
      fase: "spec",
      contract: {
        caja: true,
        arquetipo: "guiado" as unknown as Arquetipo,
        perfil_harness: "T2",
        gate: { tipo: "quimera" as unknown as GateTipo },
      },
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // El nodo VIVE — nada de ErrorBoundary.
    await expect(c.getByText("caja forjada a mano")).toBeInTheDocument()
    const arq = c.getByLabelText("arquetipo no reconocido: guiado")
    await expect(arq).toHaveTextContent("?")
    await expect(arq).toHaveClass("nr")
    await expect(c.getByLabelText("gate no reconocido: quimera")).toHaveClass("hollow")
    // La marca sana del medio sigue intacta.
    await expect(c.getByText("T2")).toBeInTheDocument()
  },
}

// ══ Capa «Mejora» (paquete 2026-07-24, T31) ═══════════════════════════════════════════════
//
// Las props son PRIMITIVAS (D18): `entities/arnes` no importa `entities/telemetria`, nunca. El
// widget `map-canvas` —que sí puede importar las dos— compone `CifraCaja → props`. El copy de
// confianza NO se duplica: llega por prop, y la story `CopyConfianzaEsUnaSola` del widget
// asserta la igualdad literal contra `entities/telemetria`.
//
// Este archivo hereda `a11y: { test: "todo" }` del `meta` y **no se amplía** (regla vigente):
// las marcas nuevas se pintan con `--foreground`, no con `--warn`. Como ahí axe no corre,
// `MejoraConMarcaDeFuga` lleva un assert COMPUTADO del `backgroundColor` — es la única forma de
// que el fix de D21 no se pierda en un archivo sin gate.

const CONF_HASH =
  "Identificado por la huella del arnés: el runtime redacta su nombre. Corrió fuera de ArnesIA."
const CONF_PROC =
  "Deducido por el directorio donde corrió. Si ahí corre más de un arnés, este número los mezcla."

// RF-238 · RF-239 · RF-240 · RF-242 — el caso completo: cifra, participación, barra con su
// aria-label, y NINGUNA marca de duda porque la atribución es exacta (la ausencia ES la señal).
export const MejoraCifraExacta: Story = {
  args: {
    box: cajaBox,
    cifraUsd: "1,92",
    participacionPct: 40,
    confianza: "exacta",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("USD")).toBeInTheDocument()
    await expect(c.getByText("1,92")).toBeInTheDocument()
    await expect(c.getByText("40 %")).toBeInTheDocument()
    const barra = c.getByRole("img", { name: "40 % del gasto del arnés en esta ventana." })
    await expect(barra).toBeInTheDocument()
    await expect((barra.querySelector(".mej-share-fill") as HTMLElement).style.width).toBe("40%")
    // exacta ⇒ cero marca de duda.
    await expect(c.queryByText("por huella")).toBeNull()
    await expect(canvasElement.querySelector("[data-confianza]")).toBeNull()
    // Las 5 marcas de hoy siguen intactas (superset estricto, BR-M16).
    await expect(c.getByText("caja")).toBeInTheDocument()
    await expect(c.getByText("idea → spec")).toBeInTheDocument()
    await expect(c.getByText("T2")).toBeInTheDocument()
    await expect(c.getByText("≈")).toBeInTheDocument()
    await expect(c.getByLabelText(/gate/)).toBeInTheDocument()
  },
}

// RF-242 · RF-274 — la cifra por huella SE SUMA al total igual: es un número real, solo que
// identificado por el hash del plugin en vez de por el nombre que el runtime redacta.
export const MejoraPorHuella: Story = {
  args: {
    box: cajaBox,
    cifraUsd: "1,92",
    participacionPct: 40,
    confianza: "por-hash",
    etiquetaConfianza: "por huella",
    tituloConfianza: CONF_HASH,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("por huella")).toHaveAttribute("title", CONF_HASH)
    await expect(c.getByText("1,92")).toBeInTheDocument()
  },
}

// RF-242 · H-13 — «por proceso» advierte la MEZCLA. No es lo mismo que «por huella» y el chip
// lo dice con otro texto y otro title.
export const MejoraPorProceso: Story = {
  args: {
    box: cajaBox,
    cifraUsd: "0,47",
    participacionPct: 10,
    confianza: "por-proceso",
    etiquetaConfianza: "por proceso",
    tituloConfianza: CONF_PROC,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByText("por proceso")
    await expect(chip).toHaveAttribute("title", expect.stringContaining("los mezcla"))
    await expect(c.queryByText("por huella")).toBeNull()
  },
}

// RF-240 · RF-243 — una caja sin corridas en la ventana NO vale 0: no tiene cifra, no tiene
// barra, y dice por qué. Un `USD 0,00` afirmaría que corrió y no gastó.
export const MejoraCajaSinCorridas: Story = {
  args: { box: cajaBox, motivoSinDato: "sin corridas en esta ventana" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin corridas en esta ventana")).toBeInTheDocument()
    await expect(c.queryByText(/USD/)).toBeNull()
    await expect(c.queryByText("0,00")).toBeNull()
    await expect(canvasElement.querySelector(".mej-share")).toBeNull()
  },
}

// RF-241 · D21 — el detector va NOMBRADO (un ⚠ sin nombre obliga a adivinar) y el ⚠ es
// `aria-hidden`. El assert COMPUTADO del fondo es el que sostiene el fix de contraste en un
// archivo donde axe no corre: `--card`, no `--crit-soft`.
export const MejoraConMarcaDeFuga: Story = {
  args: {
    box: cajaBox,
    cifraUsd: "1,92",
    participacionPct: 40,
    confianza: "exacta",
    marcaFuga: "re-warm TTL",
    marcaFugaGrave: true,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("re-warm TTL")).toBeInTheDocument()
    const warn = canvasElement.querySelector(".mej-fuga [aria-hidden='true']")
    await expect(warn).not.toBeNull()
    await expect(warn?.textContent).toBe("⚠")
    // D21: fondo --card (#ffffff en claro), NO --crit-soft (rgba(201,69,69,.12)).
    const fuga = canvasElement.querySelector(".mej-fuga.grave") as HTMLElement
    const fondo = getComputedStyle(fuga).backgroundColor
    const card = getComputedStyle(document.documentElement).getPropertyValue("--card").trim()
    await expect(fondo).toBe("rgb(255, 255, 255)")
    await expect(card).toBe("#ffffff")
    await expect(fondo).not.toContain("201, 69, 69")
  },
}

// RF-241 — con dos detectores se pinta UNA sola marca: la de mayor ahorro (el wire ya viene
// ordenado). Dos ⚠ en un nodo de 232 px no se leen, se acumulan.
export const MejoraDosDetectores: Story = {
  args: {
    box: cajaBox,
    cifraUsd: "1,92",
    participacionPct: 40,
    confianza: "exacta",
    marcaFuga: "re-warm TTL",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".mej-fuga")).toHaveLength(1)
    await expect(c.getByText("re-warm TTL")).toBeInTheDocument()
    await expect(c.queryByText("modelo cambiado")).toBeNull()
  },
}

// ── RF-243 · BR-M2 — los CINCO motivos de «sin dato atribuible». Son cinco y no seis porque
// D19 borró el de «conocimiento»: en este árbol el conocimiento es una BANDA (`selectBase`) y un
// token (`--c-knowledge`), no una clase de nodo — `Clase` no tiene `knowledge` (box.go:27, y el
// valor legacy pliega a `rule`). El copy de un nodo de la Base es el de `rule`.
const SIN_DATO_ASSERTS = async (canvasElement: HTMLElement, texto: string) => {
  const c = within(canvasElement)
  await expect(c.getByText(texto)).toBeInTheDocument()
  await expect(c.queryByText(/USD/)).toBeNull()
  await expect(c.queryByText("0")).toBeNull()
  await expect(canvasElement.querySelector(".node")).toHaveClass("sindato")
}

export const SinDatoSubagente: Story = {
  args: {
    box: { id: "test-author", clase: "subagent", nombre: "escribir pruebas", banda: "fase" },
    motivoSinDato: "sin dato atribuible — el subagente no se distingue en el turno",
  },
  play: async ({ canvasElement }) =>
    SIN_DATO_ASSERTS(
      canvasElement,
      "sin dato atribuible — el subagente no se distingue en el turno",
    ),
}

export const SinDatoRegla: Story = {
  args: {
    box: { id: "std-spec", clase: "rule", nombre: "estándar de spec", banda: "base" },
    motivoSinDato: "sin dato atribuible — una regla no consume por sí misma",
  },
  play: async ({ canvasElement }) =>
    SIN_DATO_ASSERTS(canvasElement, "sin dato atribuible — una regla no consume por sí misma"),
}

// H-10 — el literal CORREGIDO de la iteración 2: dice además que todavía no se desglosa. La
// iteración 1 decía solo «va en la caja que lo llama», que se leía como «ya está resuelto».
export const SinDatoMcp: Story = {
  args: {
    box: { id: "api-mcp", clase: "mcp", nombre: "acceso a la API", banda: "base" },
    motivoSinDato:
      "sin dato atribuible — el costo del MCP está incluido en la caja que lo llama; todavía no se desglosa",
  },
  play: async ({ canvasElement }) =>
    SIN_DATO_ASSERTS(
      canvasElement,
      "sin dato atribuible — el costo del MCP está incluido en la caja que lo llama; todavía no se desglosa",
    ),
}

export const SinDatoHook: Story = {
  args: {
    box: { id: "hook-stop", clase: "hook", nombre: "cierre de turno", banda: "guardia" },
    motivoSinDato: "sin dato atribuible — un hook informa, no consume",
  },
  play: async ({ canvasElement }) =>
    SIN_DATO_ASSERTS(canvasElement, "sin dato atribuible — un hook informa, no consume"),
}

// D19 — el motivo genérico cubre el hueco que dejó borrar la fila de «conocimiento».
export const SinDatoResto: Story = {
  args: {
    box: { id: "spec-review", clase: "command", nombre: "/spec-review", banda: "fase" },
    motivoSinDato: "sin dato atribuible — esta capa mide cajas",
  },
  play: async ({ canvasElement }) =>
    SIN_DATO_ASSERTS(canvasElement, "sin dato atribuible — esta capa mide cajas"),
}

// design §6.3 — nada desborda la celda de 232 px, y el motivo ENVUELVE: `text-overflow:
// ellipsis` está prohibido porque un motivo truncado es peor que no tenerlo.
export const MejoraDesbordamiento: Story = {
  args: {
    box: {
      ...cajaBox,
      nombre:
        "escribir el spec ejecutable del paquete de trabajo con todos sus criterios de aceptación",
    },
    cifraUsd: "123 456,78",
    participacionPct: 100,
    confianza: "por-hash",
    etiquetaConfianza: "por huella",
    tituloConfianza: CONF_HASH,
    marcaFuga: "re-warm TTL",
    motivoSinDato:
      "sin dato atribuible — el costo del MCP está incluido en la caja que lo llama; todavía no se desglosa",
  },
  play: async ({ canvasElement }) => {
    for (const sel of [".mej-cifra", ".mej-fuga", ".mej-motivo"]) {
      const el = canvasElement.querySelector(sel) as HTMLElement
      await expect(el).not.toBeNull()
      await expect(el.scrollWidth).toBeLessThanOrEqual(el.clientWidth + 1)
    }
    const motivo = canvasElement.querySelector(".mej-motivo") as HTMLElement
    await expect(getComputedStyle(motivo).textOverflow).not.toBe("ellipsis")
  },
}

// RF-245 · BR-M16 — **el guardián del superset**: sin ninguna prop de mejora el DOM es el de
// hoy, carácter por carácter. La capa apagada no deja rastro.
export const EstructuraIntacta: Story = {
  args: { box: cajaBox },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelector(".mej-cifra")).toBeNull()
    await expect(canvasElement.querySelector(".mej-fuga")).toBeNull()
    await expect(canvasElement.querySelector(".mej-share")).toBeNull()
    await expect(canvasElement.querySelector(".mej-motivo")).toBeNull()
    await expect(canvasElement.querySelector(".node")).not.toHaveClass("conmejora")
    await expect(canvasElement.querySelector(".node")).not.toHaveClass("sindato")
    await expect(c.queryByText(/sin dato atribuible/)).toBeNull()
  },
}
