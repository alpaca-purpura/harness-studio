import type { Meta, StoryObj } from "@storybook/react-vite"
import type React from "react"
import type { ReactNode } from "react"
import { useState } from "react"
import { expect, fn, userEvent, waitFor, within } from "storybook/test"
import type { Validacion } from "@/entities/marketplace"
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
    // §9.2 del paquete 2026-07-23 — la tab Marketplace DEJA de estar `disabled` y su
    // «próximo · S2» se borra: la rama se construyó (AG-D8 decisión 8). El tooltip de
    // «Repositorio GitHub», más abajo, SÍ se queda (eso sigue fuera de alcance).
    const tabMarketplace = c.getByRole("tab", { name: "Marketplace" })
    await expect(tabMarketplace).toBeEnabled()
    await expect(tabMarketplace).not.toHaveAttribute("title")
    await expect(tabMarketplace).toHaveAttribute("aria-selected", "false")

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

    // AG-D6 (paquete 2026-07-23) — ahora hay DOS salidas donde antes había una, y son distintas:
    // `← Atrás` aborta el fetch y vuelve al paso 1 (eso es exactamente lo que `onCancelarEscaneo`
    // ya hacía en el Slice 1: se reusa, no se inventa un callback), y `Cancelar` aborta y cierra
    // el flujo entero (reusa `onClose`, que ya hace abort + cero efectos).
    const atras = c.getByRole("button", { name: "← Atrás" })
    await expect(atras).toBeEnabled()
    await userEvent.click(atras)
    await expect(args.onCancelarEscaneo).toHaveBeenCalledTimes(1)

    const cancelar = c.getByRole("button", { name: "Cancelar" })
    await expect(cancelar).toBeEnabled()
    await userEvent.click(cancelar)
    await expect(args.onClose).toHaveBeenCalledTimes(1)
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
    await expect(c.getByText(/No encontré arneses instalados aquí/)).toBeInTheDocument()
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

    // AG-D6 — el pie tampoco deja salir con un POST en vuelo, y dice por qué.
    const atras = c.getByRole("button", { name: "← Atrás" })
    await expect(atras).toBeDisabled()
    await expect(atras).toHaveAttribute("title", "agregando en curso — esperá a que termine")
    const cancelar = c.getByRole("button", { name: "Cancelar" })
    await expect(cancelar).toBeDisabled()
    await expect(cancelar).toHaveAttribute("title", "agregando en curso — esperá a que termine")

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
    await expect(c.getByRole("button", { name: "Escanear" })).toBeDisabled()

    // Foco inicial: dentro del wizard, en Cerrar.
    await waitFor(() => expect(closeBtn).toHaveFocus())

    // Trap hacia atrás: Shift+Tab desde el primer focuseable (Cerrar) va al último. Con el pie de
    // AG-D6 el último focuseable pasa a ser `Cancelar` (`← Atrás` queda FUERA del loop por
    // `disabled` en el primer paso, y `Escanear` también mientras no haya ruta — S1-D23).
    await userEvent.tab({ shift: true })
    await expect(c.getByRole("button", { name: "Cancelar" })).toHaveFocus()

    // Trap hacia adelante: Tab desde el último vuelve al primero.
    await userEvent.tab()
    await expect(closeBtn).toHaveFocus()

    // Escape llama onClose (fuera de "agregando" — ver story Agregando para el caso bloqueado).
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}

// ══════════════════════════════════════════════════════════════════════════════════════════
// Paquete 2026-07-23-portafolio-agregar-marketplace — AG-D6 (Atrás/Cancelar, W3/W4) y AG-D8
// decisión 8 (rama Marketplace, S6). Todo lo de abajo es SUPERSET: nada de arriba se quitó.
// ══════════════════════════════════════════════════════════════════════════════════════════

const RUTA_ESCANEADA = "/home/chalreme/Proyectos/luana-platform"

// WizardConEstado — wrapper con estado local para las stories que necesitan ejercitar una
// TRANSICIÓN real (ir a `fuente` y volver). Las props del widget son puras: el dueño de `estado`
// es la página, así que acá lo simulamos con `useState` — es lo que hace la página.
function WizardConEstado(props: React.ComponentProps<typeof PortafolioWizard>) {
  const [estado, setEstado] = useState<React.ComponentProps<typeof PortafolioWizard>["estado"]>(
    props.estado,
  )
  return (
    <PortafolioWizard
      {...props}
      estado={estado}
      onAtras={() => {
        props.onAtras?.()
        setEstado("fuente")
      }}
    />
  )
}

// E-11 — `← Atrás` desde `candidatos`: llama `onAtras` UNA vez, y al volver al paso 1 el input
// conserva la ruta EXACTA. Volver a paso 1 para corregir un typo no debe castigar al operador
// con re-tildar 30 filas: el set de elegidos SOBREVIVE (el widget lo posee y no desmonta).
export const WizardAtrasDesdeCandidatos: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosDemo,
    pathEscaneado: RUTA_ESCANEADA,
    onAtras: fn(),
  },
  render: (args) => <WizardConEstado {...args} />,
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    // se tildan 2 (AG-D7: nada premarcado, el operador tilda).
    const filas = canvasElement.querySelectorAll(".pf-wizard-candidato")
    await userEvent.click(within(filas[0] as HTMLElement).getByRole("checkbox"))
    await userEvent.click(within(filas[2] as HTMLElement).getByRole("checkbox"))
    await expect(c.getByRole("button", { name: "Agregar 2 al portafolio" })).toBeEnabled()

    // W4 — el contador de HALLAZGOS (4) es otra cifra que la de elegidos (2), a propósito.
    await expect(c.getByText(/encontrados 4 arneses/)).toBeInTheDocument()
    // W3 — la ruta escaneada es VISIBLE.
    await expect(c.getByText(RUTA_ESCANEADA)).toBeInTheDocument()

    await userEvent.click(c.getByRole("button", { name: "← Atrás" }))
    await expect(args.onAtras).toHaveBeenCalledTimes(1)

    // ya en `fuente`: el input conserva la ruta exacta.
    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toHaveValue(RUTA_ESCANEADA)
  },
}

// E-12 — `cambiar ruta` es LA MISMA transición que `← Atrás`, con otra puerta: un solo handler,
// dos botones. Esta story asserta el mismo resultado que E-11 a propósito.
export const WizardCambiarRuta: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosDemo,
    pathEscaneado: RUTA_ESCANEADA,
    onAtras: fn(),
  },
  render: (args) => <WizardConEstado {...args} />,
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "cambiar ruta" }))
    await expect(args.onAtras).toHaveBeenCalledTimes(1)
    const input = c.getByRole("textbox", { name: "Ruta del proyecto" })
    await expect(input).toHaveValue(RUTA_ESCANEADA)
  },
}

// E-13 — Cancelar = CERO efectos: `onClose` se llama y NI `onAgregar` NI `onEscanear` se tocan.
// El widget no persiste nada por su cuenta (S1-D9).
export const WizardCancelarCeroEfectos: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosDemo,
    pathEscaneado: RUTA_ESCANEADA,
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const filas = canvasElement.querySelectorAll(".pf-wizard-candidato")
    await userEvent.click(within(filas[0] as HTMLElement).getByRole("checkbox"))
    await userEvent.click(within(filas[1] as HTMLElement).getByRole("checkbox"))

    await userEvent.click(c.getByRole("button", { name: "Cancelar" }))
    await expect(args.onClose).toHaveBeenCalledTimes(1)
    await expect(args.onAgregar).not.toHaveBeenCalled()
    await expect(args.onEscanear).not.toHaveBeenCalled()
  },
}

// E-14 — con un POST en vuelo el cierre está BLOQUEADO por las TRES puertas (✕, Atrás, Cancelar)
// y Esc no cierra; cada una dice por qué (S1-D19). Complementa `Agregando` sin reemplazarla.
export const WizardCierreBloqueadoAgregando: Story = {
  args: { estado: "agregando" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    for (const nombre of ["Cerrar", "← Atrás", "Cancelar"]) {
      const btn = c.getByRole("button", { name: nombre })
      await expect(btn).toBeDisabled()
      await expect(btn).toHaveAttribute("title", "agregando en curso — esperá a que termine")
    }
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).not.toHaveBeenCalled()
  },
}

// E-08 — 0 hallazgos HONESTO: el mensaje completo + `PasoFuente` re-montado (la forma de «volver
// a paso 1» sin inventar una prop) + CERO filas de candidato. Complementa `SinHallazgos`.
export const WizardCeroHallazgos: Story = {
  args: { estado: "candidatos", candidatos: [], pathEscaneado: "/home/chalreme/Proyectos/vitalia" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/No encontré arneses instalados aquí/)).toBeInTheDocument()
    await expect(c.getByLabelText("Ruta del proyecto")).toBeVisible()
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(0)
    // sin hallazgos no hay nada que agregar: el botón contador no existe.
    await expect(c.queryByRole("button", { name: /Agregar \d+ al portafolio/ })).toBeNull()
  },
}

// AG-D6 — el primer paso muestra `← Atrás` **`disabled` y VISIBLE** con el literal del mockup:
// nunca oculto (mismo criterio que el resto del Portafolio con lo que no aplica).
export const WizardAtrasDisabledEnPrimerPaso: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const atras = c.getByRole("button", { name: "← Atrás" })
    await expect(atras).toBeVisible()
    await expect(atras).toBeDisabled()
    await expect(atras).toHaveAttribute("title", "ya estás en el primer paso")
  },
}

// AG-D3 — buscador + filtros + lazy sobre los hallazgos, obligatorio *porque* AG-D7 no premarca.
// Los filtros reusan vocabulario YA firmado (`EstadoDeriva` literal, `origen.registry` real):
// cero palabra nueva.
export const WizardFiltraHallazgos: Story = {
  args: {
    estado: "candidatos",
    candidatos: candidatosDemo,
    pathEscaneado: RUTA_ESCANEADA,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const buscador = c.getByRole("searchbox", { name: "Buscar entre los hallazgos" })
    await userEvent.type(buscador, "harness")
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(1)

    await userEvent.clear(buscador)
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(4)

    const filtros = within(c.getByRole("group", { name: "Filtros de los hallazgos" }))
    await userEvent.click(filtros.getByRole("button", { name: "Deriva" }))
    await userEvent.click(c.getByRole("button", { name: "en-deriva" }))
    await expect(canvasElement.querySelectorAll(".pf-wizard-candidato")).toHaveLength(1)
  },
}

// ── Rama Marketplace (S6) ──────────────────────────────────────────────────────────────────

const validacionDemo: Validacion = {
  url_canonica: "github.com/alpacapurpura/prenter-marketplace",
  nombre: "prenter-marketplace",
  owner_nombre: "Prenter",
  owner_email: "hola@alpacapurpura.lat",
  entradas: 2,
  fuente: "local",
  ya_registrado: false,
}

// BR-5 / criterio G3 — **no existe `Registrar` sin haber validado**: no está en el DOM, y al
// estado `validado` solo se llega con un 200 real. BR-5 vive en la máquina de estados, no en un
// `disabled` que se pueda saltear. Y el ✓ no aparece en ninguna parte.
// biome-ignore lint/style/useNamingConvention: nombre de story EXIGIDO literal por el capability wizard-registrar-marketplace (scenario wiz-mkt-sin-validado-no-hay-registrar)
export const WizardMarketplacePasoURL: Story = {
  args: { rama: "marketplace", mkEstado: "url", onRama: fn(), onValidarMarketplace: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.getByRole("tab", { name: "Marketplace" })).toHaveAttribute(
      "aria-selected",
      "true",
    )
    await expect(c.queryByRole("button", { name: /Registrar/ })).toBeNull()
    // Cero ✓ de AFIRMACIÓN. (El copy firmado de la nota menciona el carácter «✓» justamente para
    // decir que sin respuesta real no se pinta ninguno — eso no es un resultado fingido.)
    await expect(c.queryByText(/✓ leí/)).toBeNull()
    await expect(canvasElement.querySelector(".pf-wizard-validado")).toBeNull()

    const input = c.getByRole("textbox", { name: "URL del marketplace" })
    const validar = c.getByRole("button", { name: "Validar" })
    await expect(validar).toBeDisabled()
    await userEvent.type(input, "https://github.com/alpacapurpura/prenter-marketplace")
    await expect(validar).toBeEnabled()
    await userEvent.click(validar)
    await expect(args.onValidarMarketplace).toHaveBeenCalledWith(
      "https://github.com/alpacapurpura/prenter-marketplace",
    )

    // AG-D6/§9.2 — desde `url`, `← Atrás` vuelve a la tab Proyecto (el literal «un solo paso» del
    // mockup se descarta: ahora hay 4 estados).
    await userEvent.click(c.getByRole("button", { name: "← Atrás" }))
    await expect(args.onRama).toHaveBeenCalledWith("proyecto")
  },
}

// E-15 / C6 — el ✓ muestra lo que se LEYÓ del archivo real: `name`, `owner.name` (+ email),
// «N arneses en el catálogo» y la fuente. Nada inferido. La clase arranca en `propio` (caso
// dominante) y la nota explica la diferencia A LA VISTA: no es un default silencioso.
export const WizardMarketplaceValidado: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "validado",
    mkValidacion: validacionDemo,
    onRama: fn(),
    onRegistrarMarketplace: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    const validado = c.getByRole("group", { name: "Marketplace validado" })
    await expect(validado).toHaveTextContent("github.com/alpacapurpura/prenter-marketplace")
    await expect(validado).toHaveTextContent("prenter-marketplace")
    await expect(validado).toHaveTextContent("Prenter")
    await expect(validado).toHaveTextContent("hola@alpacapurpura.lat")
    await expect(validado).toHaveTextContent("2 arneses en el catálogo")
    await expect(validado).toHaveTextContent("fuente: local")

    await expect(c.getByRole("radio", { name: "Propio" })).toBeChecked()
    await expect(c.getByRole("radio", { name: "De referencia" })).not.toBeChecked()
    await expect(c.getByText(/publicamos ahí; habilita traer/)).toBeInTheDocument()
    await expect(c.getByText(/solo resuelve procedencia de arneses ajenos/)).toBeInTheDocument()

    // la url ya no se puede editar sin volver atrás: lo validado es lo que se registra.
    await expect(c.getByRole("textbox", { name: "URL del marketplace" })).toBeDisabled()
    await expect(c.queryByRole("button", { name: "Validar" })).toBeNull()
  },
}

// WizardMkConEstado — wrapper que simula lo que hace la PÁGINA en la rama Marketplace: pasar de
// `url` a `validado` cuando el backend devuelve un 200. Sin esto no se puede asertar que
// `Registrar` manda la url REAL que el operador tipeó (el widget la posee, igual que `path`).
function WizardMkConEstado(props: React.ComponentProps<typeof PortafolioWizard>) {
  const [mkEstado, setMkEstado] =
    useState<React.ComponentProps<typeof PortafolioWizard>["mkEstado"]>("url")
  return (
    <PortafolioWizard
      {...props}
      rama="marketplace"
      mkEstado={mkEstado}
      mkValidacion={mkEstado === "url" ? undefined : validacionDemo}
      onValidarMarketplace={(url) => {
        props.onValidarMarketplace?.(url)
        setMkEstado("validado")
      }}
    />
  )
}

const URL_MK = "https://github.com/alpacapurpura/prenter-marketplace"

// E-19 — el camino completo `url → validado → registrando`: `Registrar y ver catálogo` manda los
// valores EXACTOS (la url que el operador tipeó + la clase elegida, `propio` por default). Y
// **antes de validar el botón no existe**: BR-5 vive en la máquina de estados.
// biome-ignore lint/style/useNamingConvention: nombre de story EXIGIDO literal por plan-pruebas.md E-19
export const WizardRegistrarYAterriza: Story = {
  args: {
    rama: "marketplace",
    onRama: fn(),
    onValidarMarketplace: fn(),
    onRegistrarMarketplace: fn(),
  },
  render: (args) => <WizardMkConEstado {...args} />,
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.queryByRole("button", { name: /Registrar/ })).toBeNull()
    await userEvent.type(c.getByRole("textbox", { name: "URL del marketplace" }), URL_MK)
    await userEvent.click(c.getByRole("button", { name: "Validar" }))
    await expect(args.onValidarMarketplace).toHaveBeenCalledWith(URL_MK)

    // recién en `validado` aparece el primario, y la clase arranca en `propio` (caso dominante).
    const registrar = c.getByRole("button", { name: "Registrar y ver catálogo" })
    await expect(c.getByRole("radio", { name: "Propio" })).toBeChecked()
    await userEvent.click(registrar)
    await expect(args.onRegistrarMarketplace).toHaveBeenCalledWith(URL_MK, "propio")
  },
}

// …y la clase es una ELECCIÓN, no un default silencioso: cambiarla cambia lo que se registra.
export const WizardRegistrarDeReferencia: Story = {
  args: {
    rama: "marketplace",
    onRama: fn(),
    onValidarMarketplace: fn(),
    onRegistrarMarketplace: fn(),
  },
  render: (args) => <WizardMkConEstado {...args} />,
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.type(c.getByRole("textbox", { name: "URL del marketplace" }), URL_MK)
    await userEvent.click(c.getByRole("button", { name: "Validar" }))
    await userEvent.click(c.getByRole("radio", { name: "De referencia" }))
    await userEvent.click(c.getByRole("button", { name: "Registrar y ver catálogo" }))
    await expect(args.onRegistrarMarketplace).toHaveBeenCalledWith(URL_MK, "referencia")
  },
}

// E-18 / BR-7 — registrar un duplicado NO pisa ni duplica: se informa y se ofrece `ir a él`.
export const WizardRegistrarDuplicado: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "validado",
    mkValidacion: validacionDemo,
    mkYaRegistrado: { nombre: "prenter-marketplace" },
    mkError:
      "arnesia POST /api/marketplaces: 409 marketplace: ya registrado (no se duplica ni se pisa)",
    onRama: fn(),
    // biome-ignore lint/style/useNamingConvention: nombre de prop EXIGIDO literal por design.md §9.1
    onIrAMarketplace: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/ya está registrado/)).toBeInTheDocument()
    await expect(c.getByText(/no se duplica ni se pisa lo que ya declaraste/)).toBeInTheDocument()
    await expect(c.getByRole("alert")).toHaveTextContent("409")
    await userEvent.click(c.getByRole("button", { name: "ir a él" }))
    await expect(args.onIrAMarketplace).toHaveBeenCalledWith("prenter-marketplace")
  },
}

// E-16 — url inexistente (400): motivo REAL del backend y **CERO ✓** en todo el DOM.
export const WizardMarketplaceErrorNoExiste: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "url",
    mkError:
      "arnesia POST /api/marketplaces/validaciones: 400 gh: HTTP 404 — repo inexistente o sin acceso con la credencial actual",
    onRama: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("404")
    await expect(c.queryByText(/✓ leí/)).toBeNull()
    await expect(c.queryByRole("button", { name: /Registrar/ })).toBeNull()
  },
}

// E-17 — repo que existe pero NO es marketplace (400): mensaje DISTINGUIBLE del 404 de E-16.
export const WizardMarketplaceErrorNoEsMarketplace: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "url",
    mkError:
      "arnesia POST /api/marketplaces/validaciones: 400 marketplace: el repo no expone .claude-plugin/marketplace.json legible",
    onRama: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent(
      "no expone .claude-plugin/marketplace.json legible",
    )
    await expect(c.getByRole("alert")).not.toHaveTextContent("404")
    await expect(c.queryByText(/✓ leí/)).toBeNull()
  },
}

// E-54 — sin `gh` ni PAT (503): «NO PUEDO MIRAR» nunca se pinta como «tu url está mal». Tercer
// mensaje distinto de los dos 400 de arriba.
export const WizardMarketplaceSinViaDeLectura: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "url",
    mkError:
      "arnesia POST /api/marketplaces/validaciones: 503 marketplace: sin vía de lectura (ni checkout local ni gh/PAT)",
    onRama: fn(),
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const alerta = c.getByRole("alert")
    await expect(alerta).toHaveTextContent("503")
    await expect(alerta).toHaveTextContent("sin vía de lectura")
    await expect(alerta).not.toHaveTextContent("404")
    await expect(alerta).not.toHaveTextContent("no expone")
    await expect(c.queryByText(/✓ leí/)).toBeNull()
  },
}

// Validando: spinner honesto y la url bloqueada (no se dispara una segunda validación en paralelo).
export const WizardMarketplaceValidando: Story = {
  args: { rama: "marketplace", mkEstado: "validando", onRama: fn() },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("status")).toHaveTextContent("Validando…")
    await expect(c.getByRole("textbox", { name: "URL del marketplace" })).toBeDisabled()
    await expect(c.queryByText(/✓ leí/)).toBeNull()
  },
}

// Registrando: mismo criterio que `agregando` (S1-D19) — el cierre está bloqueado por las tres
// puertas y cada una dice por qué; Esc no cierra.
export const WizardMarketplaceRegistrando: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "registrando",
    mkValidacion: validacionDemo,
    onRama: fn(),
    onRegistrarMarketplace: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Registrando…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")

    for (const nombre of ["Cerrar", "← Atrás", "Cancelar"]) {
      const b = c.getByRole("button", { name: nombre })
      await expect(b).toBeDisabled()
      await expect(b).toHaveAttribute("title", "registrando en curso — esperá a que termine")
    }
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).not.toHaveBeenCalled()
  },
}

// E-30 — a11y de la rama Marketplace (gate axe en `error`: cualquier violación ROMPE el test).
// `role=dialog` + `aria-modal` + `aria-labelledby` al `<h2>`; foco inicial dentro; Tab cicla en
// las dos direcciones; Esc cierra.
export const WizardMarketplaceA11y: Story = {
  args: {
    rama: "marketplace",
    mkEstado: "validado",
    mkValidacion: validacionDemo,
    onRama: fn(),
    onRegistrarMarketplace: fn(),
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const dialog = c.getByRole("dialog")
    await expect(dialog).toHaveAttribute("aria-modal", "true")
    const labelledby = dialog.getAttribute("aria-labelledby")
    await expect(labelledby).toBeTruthy()
    const titleEl = labelledby ? document.getElementById(labelledby) : null
    await expect(titleEl).toHaveTextContent("Agregar al portafolio")

    // el radiogroup de clase está rotulado (los radios no quedan huérfanos).
    await expect(c.getByRole("radiogroup", { name: "Clase de marketplace" })).toBeInTheDocument()

    const cerrar = c.getByRole("button", { name: "Cerrar" })
    await waitFor(() => expect(cerrar).toHaveFocus())
    await userEvent.tab({ shift: true })
    await expect(c.getByRole("button", { name: "Cancelar" })).toHaveFocus()
    await userEvent.tab()
    await expect(cerrar).toHaveFocus()

    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}
