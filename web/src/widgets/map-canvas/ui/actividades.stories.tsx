import type { Meta, StoryObj } from "@storybook/react-vite"
import { useState } from "react"
import { expect, fn, waitFor, within } from "storybook/test"
import { developerVitaliaActividades, devFullCycle, type Graph } from "@/entities/arnes"
import { ActividadChips } from "./actividad-chips"
import { Inspector } from "./inspector"
import { MapBar } from "./map-bar"
import { MapCanvas } from "./map-canvas"

// Story = test del Mapa multi-actividad (MA-T7, spec 2026-07-30-definicion-de-arnes §3/§4).
// ARCHIVO NUEVO a propósito: las 13 stories de map-canvas.stories.tsx (+5 de map-bar) son la
// geografía FIRMADA y no se tocan (MA-L3 medible — superset estricto). Fixture =
// developer-vitalia calcado de la escena A del mockup firmado (mockup-mapa-actividades.html).
// a11y `todo`: mismo gate que el archivo vigente (micro-etiquetas de bajo contraste de la
// paleta firmada); las aserciones de DOM son las que gatean CI.

const meta = {
  title: "widgets/map-canvas/Actividades",
  parameters: { layout: "fullscreen", a11y: { test: "todo" } },
} satisfies Meta

export default meta
type Story = StoryObj<typeof meta>

const forjarSpy = fn()

// La composición REAL: MapBar + fila de chips + canvas, con el estado del foco levantado al
// dueño — exactamente el patrón prop+callback de workspace-stage (capa/artefactos).
function MapaMultiActividad({
  graph,
  inicial,
  conCta,
}: {
  graph: Graph
  inicial?: string | undefined
  conCta?: boolean | undefined
}) {
  const [foco, setFoco] = useState<string | null>(inicial ?? null)
  const [pre, setPre] = useState<string | null>(null)
  const [sel, setSel] = useState<string>()
  return (
    <div style={{ height: 680, display: "flex", flexDirection: "column" }}>
      <MapBar
        arnes={graph.arnes}
        capa="estructura"
        onCapa={fn()}
        actividadFoco={foco ?? undefined}
        onVerTodo={() => setFoco(null)}
      />
      <ActividadChips
        graph={graph}
        actividadFoco={foco ?? undefined}
        onActividadFoco={setFoco}
        onActividadPre={setPre}
        onForjarConversando={conCta ? forjarSpy : undefined}
      />
      <div style={{ flex: 1, minHeight: 0 }}>
        <MapCanvas
          graph={graph}
          selectedId={sel}
          onSelect={setSel}
          actividadFoco={foco ?? undefined}
          actividadPre={pre ?? undefined}
        />
      </div>
    </div>
  )
}

// ── 1 · N0: la fila de chips — un chip por tipo + sin-actividad, salud honesta ────────────
export const ChipsN0: Story = {
  render: () => <MapaMultiActividad graph={developerVitaliaActividades} conCta />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // La fila existe, con rol/aria-label de grupo (a11y del corte).
    const fila = c.getByRole("group", { name: "Actividades del arnés" })
    // 4 actividades + chip sin-actividad = 5 chips.
    await expect(fila.querySelectorAll(".achip")).toHaveLength(5)
    // Conteos de cajas DISTINTAS por procedimiento.
    await expect(c.getByRole("button", { name: /historia\s*4 cajas/ })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /bugfix\s*4 cajas/ })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: /spike\s*1 caja$/ })).toBeInTheDocument()
    // Salud worst-of SIN dato nuevo: historia crit (test-all gate:none) · bugfix ok ·
    // spike warn (paso sin caja E13).
    const dotDe = (nombre: RegExp) =>
      within(canvasElement)
        .getByRole("button", { name: nombre })
        .querySelector(".dot")
        ?.getAttribute("style")
    expect(dotDe(/historia/)).toContain("var(--crit)")
    expect(dotDe(/bugfix/)).toContain("var(--ok)")
    expect(dotDe(/spike/)).toContain("var(--warn)")
    // E6 — el grupo sin-actividad es visible, focable y con su salud (crit: gate:none).
    const sinAct = c.getByRole("button", { name: /sin-actividad\s*2 cajas/ })
    await expect(sinAct).toHaveAttribute("aria-pressed", "false")
    expect(sinAct.querySelector(".dot")?.getAttribute("style")).toContain("var(--crit)")
    // Sin foco: ni breadcrumb ni atenuación.
    await expect(canvasElement.querySelector("[data-testid='crumb-actividad']")).toBeNull()
    await expect(canvasElement.querySelectorAll(".dimlane")).toHaveLength(0)
  },
}

// ── 2 · N1: foco «historia» — dim por faceta, secuencia numerada, breadcrumb ──────────────
export const FocoHistoria: Story = {
  render: () => <MapaMultiActividad graph={developerVitaliaActividades} inicial="historia" />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const nodo = (id: string) =>
      canvasElement.querySelector(`[data-node-id="${id}"]`) as HTMLElement
    // Cajas de fase FUERA de historia → dim; las de historia, plenas.
    await expect(nodo("gate-runner")).toHaveClass("dim")
    await expect(nodo("chrome-devtools-verify")).toHaveClass("dim")
    await expect(nodo("dev-team")).not.toHaveClass("dim")
    await expect(nodo("commit-push")).not.toHaveClass("dim")
    // MA-L2 — Guardia y Base SIEMPRE plenas.
    await expect(nodo("contract-guard")).not.toHaveClass("dim")
    await expect(nodo("vitalia-design-system")).not.toHaveClass("dim")
    // La secuencia del procedimiento: 4 pasos → 4 números + 3 trazos sólidos --primary
    // (medida asíncrona, mismo patrón bezier que los edges).
    await waitFor(() => {
      expect(canvasElement.querySelectorAll(".seq-trazo")).toHaveLength(3)
      expect(canvasElement.querySelectorAll(".seq-num")).toHaveLength(4)
    })
    // Historia toca las 4 fases → cero carriles atenuados.
    await expect(canvasElement.querySelectorAll(".dimlane")).toHaveLength(0)
    // Breadcrumb en la barra: arnés ▸ actividad + «ver todo» + disclaimer §3.
    const crumb = canvasElement.querySelector("[data-testid='crumb-actividad']") as HTMLElement
    await expect(crumb).not.toBeNull()
    await expect(within(crumb).getByText("historia")).toBeInTheDocument()
    await expect(within(crumb).getByText("cifras del arnés completo")).toBeInTheDocument()
    // «ver todo» vuelve al panorama y desarma secuencia + dim (MA-L1).
    await within(crumb).getByRole("button", { name: "ver todo" }).click()
    await waitFor(() => {
      expect(canvasElement.querySelectorAll(".seq-trazo")).toHaveLength(0)
      expect(nodo("gate-runner")).not.toHaveClass("dim")
    })
    await expect(c.getByRole("button", { name: /historia/ })).toHaveAttribute(
      "aria-pressed",
      "false",
    )
  },
}

// ── 3 · E5 + E13: foco «spike» — el spike TERMINA ANTES y el paso sin caja se VE ──────────
export const FocoSpikeTerminaAntes: Story = {
  render: () => <MapaMultiActividad graph={developerVitaliaActividades} inicial="spike" />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // E5 — spike solo pisa preparación: los otros 3 carriles se atenúan ENTEROS.
    await expect(canvasElement.querySelectorAll(".lane")).toHaveLength(4)
    await expect(canvasElement.querySelectorAll(".lane.dimlane")).toHaveLength(3)
    // E13 — el paso «decidir» sin caja es un ghost EN la secuencia, hueco visible.
    const ghost = canvasElement.querySelector(".node.ghost") as HTMLElement
    await expect(ghost).not.toBeNull()
    await expect(c.getByText("decidir — paso sin caja aún")).toBeInTheDocument()
    await expect(c.getByText("entrega: decision.md")).toBeInTheDocument()
    // El ghost NO es un botón (no clickeable como nodo real).
    await expect(ghost.tagName).toBe("DIV")
    // La secuencia lo incluye con su número: 2 pasos → 2 números + 1 trazo.
    await waitFor(() => {
      expect(canvasElement.querySelectorAll(".seq-num")).toHaveLength(2)
      expect(canvasElement.querySelectorAll(".seq-trazo")).toHaveLength(1)
    })
  },
}

// ── 4 · E7: actividad sin procedimiento — degradado honesto, NO focable, CTA al chat ──────
export const ChipDegradadoE7: Story = {
  render: () => <MapaMultiActividad graph={developerVitaliaActividades} conCta />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByRole("button", { name: /revisar-capability/ })
    await expect(chip).toHaveClass("degradado")
    await expect(chip).toBeDisabled()
    await expect(within(chip).getByText("sin procedimiento aún (E7)")).toBeInTheDocument()
    // El CTA vive porque la composición pasó el handler; click lo invoca.
    const cta = c.getByRole("button", { name: "⌨ forjar el procedimiento conversando" })
    await cta.click()
    await expect(forjarSpy).toHaveBeenCalled()
  },
}

// ── 5 · E1/E8: arnés SIN tipos declarados — el Mapa EXACTO de hoy, cero DOM nuevo ─────────
export const SinTiposRegresion: Story = {
  render: () => <MapaMultiActividad graph={devFullCycle} conCta />,
  play: async ({ canvasElement }) => {
    // La fila de chips NO se renderiza (MA-L5): ni grupo, ni chips, ni CTA.
    await expect(canvasElement.querySelector(".arnesia-actividades")).toBeNull()
    await expect(canvasElement.querySelectorAll(".achip")).toHaveLength(0)
    // Ni breadcrumb ni ninguna marca nueva en el canvas: DOM idéntico al vigente.
    await expect(canvasElement.querySelector("[data-testid='crumb-actividad']")).toBeNull()
    await expect(canvasElement.querySelectorAll(".dimlane")).toHaveLength(0)
    await expect(canvasElement.querySelectorAll(".xn")).toHaveLength(0)
    await expect(canvasElement.querySelectorAll(".node.ghost")).toHaveLength(0)
    await expect(canvasElement.querySelectorAll(".seq-trazo")).toHaveLength(0)
    await expect(canvasElement.querySelectorAll(".seq-num")).toHaveLength(0)
    // La geografía vigente, intacta: mismos nodos, ninguno atenuado.
    await expect(canvasElement.querySelectorAll(".arnesia-map .node")).toHaveLength(
      devFullCycle.nodos.length,
    )
    await expect(canvasElement.querySelectorAll(".arnesia-map .node.dim")).toHaveLength(0)
  },
}

// ── 6 · MA-L4: caja compartida ×N — el badge declara la pertenencia ────────────────────────
export const BadgeCompartidaXn: Story = {
  render: () => <MapaMultiActividad graph={developerVitaliaActividades} />,
  play: async ({ canvasElement }) => {
    const devTeam = canvasElement.querySelector('[data-node-id="dev-team"]') as HTMLElement
    const badge = devTeam.querySelector(".xn") as HTMLElement
    await expect(badge).not.toBeNull()
    await expect(badge.textContent).toBe("×3")
    await expect(badge.getAttribute("title")).toContain(
      "usada por 3 actividades: historia · bugfix · spike",
    )
    // Faceta de 1 sola actividad ⇒ SIN badge (no hay nada compartido que declarar).
    const testAll = canvasElement.querySelector('[data-node-id="test-all"]') as HTMLElement
    await expect(testAll.querySelector(".xn")).toBeNull()
  },
}

// ── 7 · Inspector: fila «Actividades» entre Clasificación y Fuente + radio de impacto ─────
function Frame({ children }: { children: React.ReactNode }) {
  return <div style={{ position: "relative", width: 360, height: 560 }}>{children}</div>
}

const focarSpy = fn()
const devTeamBox = developerVitaliaActividades.nodos.find((n) => n.id === "dev-team")
const sinActividadBox = developerVitaliaActividades.nodos.find(
  (n) => n.id === "chrome-devtools-verify",
)
if (!devTeamBox || !sinActividadBox) throw new Error("fixture sin las cajas esperadas")

export const InspectorFilaActividades: Story = {
  render: () => (
    <Frame>
      <Inspector
        box={devTeamBox}
        onClose={fn()}
        graph={developerVitaliaActividades}
        onFocarActividad={focarSpy}
      />
    </Frame>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    // La sección existe, ENTRE Clasificación y Fuente.
    const titulos = [...canvasElement.querySelectorAll(".sec > h4")].map(
      (h) => h.textContent?.replace(/i$/, "").trim() ?? "",
    )
    const iClas = titulos.indexOf("Clasificación")
    const iAct = titulos.indexOf("Actividades")
    const iFuente = titulos.indexOf("Fuente")
    expect(iAct).toBeGreaterThan(iClas)
    if (iFuente >= 0) expect(iAct).toBeLessThan(iFuente)
    // Chips de la faceta; click → foca la actividad en el Mapa.
    await c.getByRole("button", { name: "historia" }).click()
    await expect(focarSpy).toHaveBeenCalledWith("historia")
    // Radio de impacto (MA-L4) en sección y botonera.
    await expect(
      c.getByText(/usada por 3 actividades — editarla avisa el radio de impacto/),
    ).toBeInTheDocument()
    await expect(
      c.getByText(/⚠ usada por 3 actividades \(historia · bugfix · spike\)/),
    ).toBeInTheDocument()
  },
}

export const InspectorSinActividadE6: Story = {
  render: () => (
    <Frame>
      <Inspector box={sinActividadBox} onClose={fn()} graph={developerVitaliaActividades} />
    </Frame>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("sin-actividad — ningún procedimiento la referencia (E6)"),
    ).toBeInTheDocument()
  },
}
