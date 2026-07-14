import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, waitFor, within } from "storybook/test"
import type { Candidato } from "@/entities/portafolio"
import { PortafolioWizard } from "./portafolio-wizard"

// Story = test (fe-visual-fitness) — widget Wizard del Portafolio (plan §2.6/§3 T6,
// G3/G5/G8). Frame reproduce el scope real (.arnesia-portafolio, portafolio.css) para que
// tokens/clases resuelvan igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

// ── Fixtures LOCALES sintéticas-con-shape-real (T6): entities/portafolio/testing/entradas.ts
// (T2) solo fixturea `EntradaPortafolio`, no `Candidato` (resultado NO persistido de un
// escaneo — plan §2.8 lo dice explícito: "T2 solo cubre EntradaPortafolio") — mismo patrón que
// T4 (`corruptaEjemplo`) y T5 (`entradaResueltaSinCanonico`/`entradaSinCanonicoSinInstalaciones`)
// documentaron inline. Shape real de `usecase.Candidato` (plan §2.1/§2.3): `{clave, identidad,
// nombre?, descripcion?, empresas?, instalacion, es_canonico?}`. ──

// candidatoHarness — la instalación REAL del E2E de Slice 0 (paridad.md §E2E, misma fuente que
// `entradaHarnessEnDeriva` de T2), en forma de Candidato (aún NO agregado): en-deriva real
// (G1), tipo `referenciada-cc` (GAP-1), registry+version resueltos.
const candidatoHarness: Candidato = {
  clave: "sin-home~harness~github-com-alpacapurpura-luana-vitalia",
  identidad: { id: "harness", scope: "github.com/alpacapurpura/luana-vitalia" },
  instalacion: {
    proyecto_path: "~/Proyectos/luana-vitalia",
    install_path: "~/.claude/plugins/cache/prenter-marketplace/harness/0.5.2",
    tipo: "referenciada-cc",
    origen: { registry: "alpacapurpura/prenter-marketplace", version: "0.5.2" },
    deriva: "en-deriva",
    deriva_detalle:
      "hash de contenido distinto de la referencia " +
      "~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.2/",
  },
}

// candidatoSinVersionSinRegistry — sintético (T6, shape real): la rama honesta `v?`/«origen
// desconocido» — un hallazgo `proyecto-instalado` sin `origen.registry` ni `version` (mismo
// espíritu que la fixture (b) de T2, S0-D14/D15), en forma de Candidato.
const candidatoSinVersionSinRegistry: Candidato = {
  clave: "sin-home~bar-cli~github-com-acme-bar-app",
  identidad: { id: "bar-cli", scope: "github.com/acme/bar-app" },
  instalacion: {
    proyecto_path: "~/Proyectos/bar-app",
    install_path: "~/Proyectos/bar-app/.claude/plugins/bar-cli",
    tipo: "proyecto-instalado",
    origen: {},
    deriva: "deriva-no-evaluable",
    deriva_detalle: "sin version instalada conocida",
  },
}

// candidatoYaPresente — misma `clave` que una entrada YA en el portafolio (S1-D10): badge «ya
// en el portafolio», checkbox sigue HABILITADO (re-agregar = re-escanear idempotente, C-P-8 —
// actualiza instalaciones, no duplica).
const candidatoYaPresente: Candidato = {
  clave: "github-com-acme-foo-cli~foo-cli~",
  identidad: { home: "github.com/acme/foo-cli", id: "foo-cli" },
  nombre: "Foo CLI",
  instalacion: {
    proyecto_path: "~/Proyectos/foo-app",
    install_path: "~/Proyectos/foo-app/.claude/plugins/foo-cli",
    tipo: "materializada",
    origen: { registry: "github.com/acme/foo-cli-mirror", version: "1.0.0" },
    deriva: "al-hilo",
  },
}

// candidatoCanonico — `es_canonico: true` (RN-IDENT-4): el checkout editable detectado en el
// propio path escaneado — se pinta con copy PROPIO, jamás como espejo/instalación.
const candidatoCanonico: Candidato = {
  clave: "github-com-acme-acme-cli~acme-cli~",
  identidad: { home: "github.com/acme/acme-cli", id: "acme-cli" },
  nombre: "Acme CLI",
  instalacion: {
    proyecto_path: "~/dev/acme-cli",
    install_path: "~/dev/acme-cli",
    tipo: "materializada",
    origen: { registry: "github.com/acme/acme-cli", version: "2.1.0" },
    deriva: "al-hilo",
  },
  es_canonico: true,
}

const candidatosDemo: Candidato[] = [
  candidatoHarness,
  candidatoSinVersionSinRegistry,
  candidatoYaPresente,
  candidatoCanonico,
]

const meta = {
  title: "widgets/portafolio/PortafolioWizard",
  component: PortafolioWizard,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    abierto: true,
    estado: "fuente",
    clavesExistentes: new Set<string>(),
    onClose: fn(),
    onEscanear: fn(),
    onCancelarEscaneo: fn(),
    onAgregar: fn(),
  },
} satisfies Meta<typeof PortafolioWizard>

export default meta
type Story = StoryObj<typeof meta>

// Paso1Fuente (G3) — radio "Carpeta local" activo con input+Escanear; radio "Repositorio
// GitHub" disabled+title; tab Marketplace disabled+title y NINGÚN "✓ marketplace válido" en
// TODO el DOM (G3, el fix central del ticket); sin `onElegirCarpeta` el botón "Elegir
// carpeta…" NO se renderiza (cubre el caso "browser plano", S1-D9). El caso "con picker" vive
// en `Paso1FuenteConElegirCarpeta` (abajo) — dos stories, no un assert condicional.
export const Paso1Fuente: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    // G3 — el fix central: jamás un resultado fingido para la rama no construida.
    await expect(c.queryByText(/✓ marketplace válido/)).toBeNull()

    const tabProyecto = c.getByRole("tab", { name: "Proyecto" })
    await expect(tabProyecto).toHaveAttribute("aria-selected", "true")
    const tabMarketplace = c.getByRole("tab", { name: "Marketplace" })
    await expect(tabMarketplace).toBeDisabled()
    await expect(tabMarketplace).toHaveAttribute("title", "próximo · S2")

    const radioLocal = c.getByRole("radio", { name: "Carpeta local" })
    await expect(radioLocal).toBeEnabled()
    await expect(radioLocal).toBeChecked()
    const radioGithub = c.getByRole("radio", { name: "Repositorio GitHub" })
    await expect(radioGithub).toBeDisabled()
    await expect(radioGithub).toHaveAttribute("title", "próximo · S2")

    // Sin onElegirCarpeta (default de esta story) — el botón no existe (S1-D9).
    await expect(c.queryByRole("button", { name: "Elegir carpeta…" })).toBeNull()

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toBeEnabled()
    await userEvent.type(input, "~/Proyectos/mi-arnes")

    const escanear = c.getByRole("button", { name: "Escanear" })
    await expect(escanear).toBeEnabled()
    await userEvent.click(escanear)
    await expect(args.onEscanear).toHaveBeenCalledWith("~/Proyectos/mi-arnes")
  },
}

// Paso1FuenteConElegirCarpeta (S1-D9) — la otra mitad de la cobertura: CON `onElegirCarpeta`
// (patrón Tauri de RF-110/AjustesView), el botón "Elegir carpeta…" SÍ se renderiza, lo llama, y
// si resuelve con un path lo pone en el input.
export const Paso1FuenteConElegirCarpeta: Story = {
  args: {
    onElegirCarpeta: fn(async () => "~/Proyectos/elegida-por-el-picker"),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const elegir = c.getByRole("button", { name: "Elegir carpeta…" })
    await expect(elegir).toBeEnabled()
    await userEvent.click(elegir)

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await waitFor(() => expect(input).toHaveValue("~/Proyectos/elegida-por-el-picker"))
  },
}

// Escaneando (G5) — spinner + Cancelar habilitado (dispara onCancelarEscaneo); input+Escanear
// bloqueados (el usuario no puede disparar un segundo escaneo en paralelo).
export const Escaneando: Story = {
  args: { estado: "escaneando" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.queryByText(/✓ marketplace válido/)).toBeNull()
    await expect(c.getByRole("status")).toHaveTextContent("Escaneando…")

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toBeDisabled()
    const escanear = c.getByRole("button", { name: "Escanear" })
    await expect(escanear).toBeDisabled()

    const cancelar = c.getByRole("button", { name: "Cancelar" })
    await expect(cancelar).toBeEnabled()
    await userEvent.click(cancelar)
    await expect(args.onCancelarEscaneo).toHaveBeenCalledTimes(1)
  },
}

// CandidatosReales (S1-D10, G1) — checkbox por candidato; v?/«origen desconocido» honestos;
// chip deriva honesto; badge «ya en el portafolio» con checkbox HABILITADO; es_canonico con su
// copy propio (NO pintado como espejo); contador del botón sigue a los checks; nota
// «espejos read-only» visible.
export const CandidatosReales: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosDemo,
    clavesExistentes: new Set([candidatoYaPresente.clave]),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const filas = canvasElement.querySelectorAll(".pf-wizard-candidato")
    await expect(filas).toHaveLength(4)

    // candidatoHarness — deriva honesta (G1) + tipo referenciada-cc.
    const filaHarness = within(filas[0] as HTMLElement)
    await expect(filaHarness.getByText("en-deriva")).toBeInTheDocument()
    await expect(filaHarness.getByText("referenciada-cc")).toBeInTheDocument()
    await expect(filaHarness.getByText("v0.5.2")).toBeInTheDocument()

    // candidatoSinVersionSinRegistry — v? y "origen desconocido", nada inventado.
    const filaSinDatos = within(filas[1] as HTMLElement)
    await expect(filaSinDatos.getByText("v?")).toBeInTheDocument()
    await expect(filaSinDatos.getByText("origen desconocido")).toBeInTheDocument()

    // candidatoYaPresente — badge visible, checkbox HABILITADO (S1-D10).
    const filaYaPresente = within(filas[2] as HTMLElement)
    await expect(filaYaPresente.getByText("ya en el portafolio")).toBeInTheDocument()
    const checkYaPresente = filaYaPresente.getByRole("checkbox")
    await expect(checkYaPresente).toBeEnabled()

    // candidatoCanonico — copy propio, JAMÁS pintado como espejo (sin chip de tipo).
    const filaCanonico = within(filas[3] as HTMLElement)
    await expect(
      filaCanonico.getByText("checkout editable — se registra como canónico, no como instalación"),
    ).toBeInTheDocument()
    await expect(filaCanonico.queryByText("materializada")).toBeNull()

    await expect(c.getByText(/espejos read-only/)).toBeInTheDocument()

    // Contador sigue a los checks: 0 → 2, botón pasa de disabled a enabled.
    const btnAgregar0 = c.getByRole("button", { name: "Agregar 0 al portafolio" })
    await expect(btnAgregar0).toBeDisabled()

    const checkHarness = filaHarness.getByRole("checkbox")
    await userEvent.click(checkHarness)
    await userEvent.click(checkYaPresente)

    const btnAgregar2 = c.getByRole("button", { name: "Agregar 2 al portafolio" })
    await expect(btnAgregar2).toBeEnabled()
    await userEvent.click(btnAgregar2)
    await expect(args.onAgregar).toHaveBeenCalledWith([
      candidatoHarness.clave,
      candidatoYaPresente.clave,
    ])
  },
}

// SinHallazgos (G5, C-P-4) — copy honesto de 0 hallazgos + el formulario de fuente reaparece
// para "elegir otra carpeta" (no hay callback dedicado de "volver" en el contrato — reusar el
// mismo formulario ES la forma de "vuelve a paso 1", ver comentario en portafolio-wizard.tsx).
export const SinHallazgos: Story = {
  args: { estado: "candidatos", candidatos: [] },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("No encontré arneses instalados aquí.")).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(0)

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await userEvent.type(input, "~/Proyectos/otra-carpeta")
    await userEvent.click(c.getByRole("button", { name: "Escanear" }))
    await expect(args.onEscanear).toHaveBeenCalledWith("~/Proyectos/otra-carpeta")
  },
}

// ErrorDePath (G5, C-P-3) — el motivo 400 REAL del backend, textual; CERO candidato fantasma.
export const ErrorDePath: Story = {
  args: {
    estado: "candidatos",
    error: "POST /api/portafolio/escaneos: 400 la ruta no existe: /tmp/no-existe-de-verdad",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("POST /api/portafolio/escaneos: 400 la ruta no existe: /tmp/no-existe-de-verdad"),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(0)
  },
}

// Agregando (S1-D19) — botón bloqueado "Agregando…", SIN checklist inventada (ningún
// checkbox/progreso-por-ítem fabricado); Cerrar queda disabled+tooltip y Esc NO cierra (el
// POST /proyectos está en vuelo — criterio delegado por el ticket, documentado en
// decisiones.md S1-D19).
export const Agregando: Story = {
  args: { estado: "agregando" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Agregando…" })
    await expect(btn).toBeDisabled()
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(0)

    const cerrar = c.getByRole("button", { name: "Cerrar" })
    await expect(cerrar).toBeDisabled()
    await expect(cerrar).toHaveAttribute("title", "agregando en curso — esperá a que termine")

    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).not.toHaveBeenCalled()
  },
}

// A11yModal (G8, CRÍTICO) — role="dialog"+aria-modal+aria-labelledby; tabs role="tab"/
// aria-selected; foco inicial dentro del wizard (Cerrar); trap de Tab en ambas direcciones;
// Escape llama onClose. Paso "fuente" sin picker: 5 focuseables (Cerrar · tab Proyecto · radio
// Carpeta local · input · Escanear) — Marketplace/GitHub quedan FUERA del trap por `disabled`
// (mismo criterio que `focusablesEn`, shared/lib/focus-trap.ts).
export const A11yModal: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const dialog = c.getByRole("dialog")
    await expect(dialog).toHaveAttribute("aria-modal", "true")
    const labelledby = dialog.getAttribute("aria-labelledby")
    await expect(labelledby).toBeTruthy()
    const titleEl = labelledby ? document.getElementById(labelledby) : null
    await expect(titleEl).toHaveTextContent("Agregar al portafolio")

    const tabProyecto = c.getByRole("tab", { name: "Proyecto" })
    await expect(tabProyecto).toHaveAttribute("aria-selected", "true")
    const tabMarketplace = c.getByRole("tab", { name: "Marketplace" })
    await expect(tabMarketplace).toHaveAttribute("aria-selected", "false")

    const closeBtn = c.getByRole("button", { name: "Cerrar" })
    const escanearBtn = c.getByRole("button", { name: "Escanear" })

    // Foco inicial: dentro del wizard, en Cerrar.
    await waitFor(() => expect(closeBtn).toHaveFocus())

    // Trap hacia atrás: Shift+Tab desde el primer focuseable (Cerrar) va al último (Escanear) —
    // los disabled (tab Marketplace, radio GitHub) quedan fuera del loop.
    await userEvent.tab({ shift: true })
    await expect(escanearBtn).toHaveFocus()

    // Trap hacia adelante: Tab desde el último vuelve al primero.
    await userEvent.tab()
    await expect(closeBtn).toHaveFocus()

    // Escape llama onClose (fuera de "agregando" — ver story Agregando para el caso bloqueado).
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}
