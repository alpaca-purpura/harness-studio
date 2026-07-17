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

// Paso1Fuente (G3/S1-D23) — radio "Carpeta local" activo con input+Escanear; radio "Repositorio
// GitHub" disabled+title; tab Marketplace disabled+title y NINGÚN "✓ marketplace válido" en
// TODO el DOM (G3, el fix central del ticket); modo "Escribir ruta" activo por default, "Elegir
// carpeta" disabled+TOOLTIP_WEB (S1-D23: sin `onElegirCarpeta`, cubre el caso "browser plano" —
// jamás oculto en silencio, mismo criterio que GitHub/Marketplace). "Escanear" arranca disabled
// SIN ruta cargada (S1-D23) y se habilita recién al tipear. El caso "con picker" vive en
// `Paso1FuenteConElegirCarpeta` (abajo) — dos stories, no un assert condicional.
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

    // S1-D23 — modo: "Escribir ruta" activo por default; "Elegir carpeta" disabled+tooltip de
    // plataforma (distinto de "próximo · S2": esto NO es roadmap, es límite permanente fuera de
    // Tauri) sin `onElegirCarpeta` (default de esta story).
    const radioEscribir = c.getByRole("radio", { name: "Escribir ruta" })
    await expect(radioEscribir).toBeChecked()
    const radioElegir = c.getByRole("radio", { name: "Elegir carpeta" })
    await expect(radioElegir).toBeDisabled()
    await expect(radioElegir).toHaveAttribute("title", "solo disponible en la app de escritorio")

    // Sin onElegirCarpeta (default de esta story) — el botón trigger no existe (solo aparece en
    // modo "elegir", que acá es inalcanzable).
    await expect(c.queryByRole("button", { name: "Elegir carpeta…" })).toBeNull()

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toBeEnabled()

    // S1-D23 — Escanear arranca disabled: sin ruta cargada, no hay nada que escanear.
    const escanear = c.getByRole("button", { name: "Escanear" })
    await expect(escanear).toBeDisabled()

    await userEvent.type(input, "~/Proyectos/mi-arnes")
    await expect(escanear).toBeEnabled()
    await userEvent.click(escanear)
    await expect(args.onEscanear).toHaveBeenCalledWith("~/Proyectos/mi-arnes")
  },
}

// Paso1FuenteConElegirCarpeta (S1-D23, supersede S1-D9) — CON `onElegirCarpeta` (patrón Tauri de
// RF-110/AjustesView) el radio "Elegir carpeta" queda habilitado; seleccionarlo revela el input
// en modo solo-lectura + el botón trigger — recién el CLICK del botón abre el picker (nunca la
// sola selección del radio, ver decisiones.md S1-D23 sobre por qué). Si resuelve con un path, lo
// muestra arriba y habilita Escanear.
export const Paso1FuenteConElegirCarpeta: Story = {
  args: {
    onElegirCarpeta: fn(async () => "~/Proyectos/elegida-por-el-picker"),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    const radioElegir = c.getByRole("radio", { name: "Elegir carpeta" })
    await expect(radioElegir).toBeEnabled()
    await userEvent.click(radioElegir)

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toHaveAttribute("readonly")
    await expect(input).toHaveAttribute("placeholder", "ninguna carpeta elegida todavía")

    const escanear = c.getByRole("button", { name: "Escanear" })
    await expect(escanear).toBeDisabled()

    const elegir = c.getByRole("button", { name: "Elegir carpeta…" })
    await expect(elegir).toBeEnabled()
    await userEvent.click(elegir)

    await waitFor(() => expect(input).toHaveValue("~/Proyectos/elegida-por-el-picker"))
    await expect(escanear).toBeEnabled()
  },
}

// Escaneando (G5) — spinner + Cancelar habilitado (dispara onCancelarEscaneo); input+Escanear+
// radios de modo bloqueados (el usuario no puede disparar un segundo escaneo en paralelo).
export const Escaneando: Story = {
  args: { estado: "escaneando" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.queryByText(/✓ marketplace válido/)).toBeNull()
    await expect(c.getByRole("status")).toHaveTextContent("Escaneando…")

    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toBeDisabled()
    const radioEscribir = c.getByRole("radio", { name: "Escribir ruta" })
    await expect(radioEscribir).toBeDisabled()
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

// ── Fixtures del caso monorepo REAL (S1-D26 — calcado del escaneo vivo de luana-vitalia que
// destapó la observación del operador, 2026-07-16): N hallazgos `proyecto-instalado` SIN
// manifiesto (id vacío — ANTES: N tarjetas «(sin id)» indistinguibles) + un aviso sin dir
// físico cuyo id SÍ se conoce (enabledPlugins lo declara; el scanner ahora lo conserva vía
// IDConocido). Shape real del wire post-fix. ──
const RAIZ_MONOREPO = "/home/u/Proyectos/luana-vitalia"

function candidatoSinManifiesto(scope: string, installPath: string): Candidato {
  return {
    clave: `sin-home~~${scope.replace(/[^a-z0-9-_]+/g, "-")}`,
    identidad: { id: "", scope },
    instalacion: {
      proyecto_path: RAIZ_MONOREPO,
      install_path: installPath,
      tipo: "proyecto-instalado",
      origen: {},
      deriva: "deriva-no-evaluable",
      deriva_detalle: "sin home ni registry accesible",
    },
  }
}

const candidatosMonorepo: Candidato[] = [
  candidatoSinManifiesto("github.com/alpacapurpura/luana-platform", RAIZ_MONOREPO),
  {
    clave: "sin-home~commit-commands~",
    identidad: { id: "commit-commands" },
    instalacion: {
      proyecto_path: RAIZ_MONOREPO,
      install_path: "",
      tipo: "",
      origen: {},
      deriva: "deriva-no-evaluable",
      deriva_detalle: "sin dir físico resoluble",
      aviso: `enabledPlugins declara "commit-commands@claude-plugins-official" pero sin record de instalación para ${RAIZ_MONOREPO}`,
    },
  },
  candidatoSinManifiesto("comunify", `${RAIZ_MONOREPO}/comunify`),
  candidatoSinManifiesto("fitflow", `${RAIZ_MONOREPO}/fitflow`),
]

// CandidatosMonorepoAgrupados (S1-D26) — el fix de la observación del operador: cada tarjeta
// sin manifiesto se identifica por su scope (ruta relativa / remote del proyecto) + chip «sin
// manifiesto»; los hallazgos se agrupan por subcarpeta (raíz primero); el aviso sin record
// muestra el id que enabledPlugins declara; «(sin id)» ya no existe cuando hay CUALQUIER dato
// mejor; el hallazgo sin forma física no pinta un chip de tipo vacío.
export const CandidatosMonorepoAgrupados: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosMonorepo,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    // Agrupación: raíz primero, una sección por subcarpeta (orden de aparición del walker).
    const grupos = canvasElement.querySelectorAll(".pf-wizard-grupo .pf-faceta-label")
    await expect(Array.from(grupos).map((g) => g.textContent)).toEqual([
      "proyecto (raíz)",
      "comunify/",
      "fitflow/",
    ])

    // Identificación sin manifiesto: scope visible + chip «sin manifiesto» (3 tarjetas sin id).
    await expect(c.getByText("github.com/alpacapurpura/luana-platform")).toBeInTheDocument()
    await expect(c.getByText("comunify")).toBeInTheDocument()
    await expect(c.getByText("fitflow")).toBeInTheDocument()
    await expect(c.getAllByText("sin manifiesto")).toHaveLength(3)

    // El aviso sin record YA NO es anónimo: el id que enabledPlugins declara se muestra.
    const filaAviso = c.getByText("commit-commands").closest(".pf-wizard-candidato") as HTMLElement
    await expect(filaAviso).not.toBeNull()
    await expect(within(filaAviso).getByText(/sin record de instalación/)).toBeInTheDocument()
    // …y sin forma física no hay chip de tipo (nada, no un chip vacío).
    await expect(filaAviso.querySelector(".pf-chip-tipo")).toBeNull()

    // «(sin id)» solo queda como último recurso — acá ninguna tarjeta lo necesita.
    await expect(c.queryByText("(sin id)")).toBeNull()
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
// Carpeta local · radio Escribir ruta · input) — Marketplace/GitHub/Elegir-carpeta quedan FUERA
// del trap por `disabled`, y Escanear TAMBIÉN queda fuera (S1-D23: arranca disabled sin ruta
// cargada, el input pasa a ser el último focuseable en vez de Escanear) — mismo criterio que
// `focusablesEn`, shared/lib/focus-trap.ts.
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
    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(c.getByRole("button", { name: "Escanear" })).toBeDisabled()

    // Foco inicial: dentro del wizard, en Cerrar.
    await waitFor(() => expect(closeBtn).toHaveFocus())

    // Trap hacia atrás: Shift+Tab desde el primer focuseable (Cerrar) va al último — el input
    // (S1-D23: Escanear arranca disabled sin ruta, sale del loop hasta que haya texto).
    await userEvent.tab({ shift: true })
    await expect(input).toHaveFocus()

    // Trap hacia adelante: Tab desde el último vuelve al primero.
    await userEvent.tab()
    await expect(closeBtn).toHaveFocus()

    // Escape llama onClose (fuera de "agregando" — ver story Agregando para el caso bloqueado).
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}
