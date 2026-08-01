import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, userEvent, waitFor, within } from "storybook/test"
import {
  type EntradaPortafolio,
  entradaCanonicaCompleta,
  entradaHarnessEnDeriva,
  entradaProyectoInstaladoProvisional,
} from "@/entities/portafolio"
import { PortafolioDrawer } from "./portafolio-drawer"

// Story = test (fe-visual-fitness) — widget Drawer del Portafolio (plan §2.6/§3 T5,
// G1/G2/G4/G5/G6/G7/G8). Frame reproduce el scope real (.arnesia-portafolio, portafolio.css)
// para que tokens/clases resuelvan igual que en la app.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-portafolio">{children}</div>
}

// ── Fixtures LOCALES sintéticas-con-shape-real (T5): entities/portafolio/testing/entradas.ts
// (T2) no cubre estas dos combinaciones — mismo patrón que T4 hizo con `corruptaEjemplo`
// (documentado inline, NO se toca entradas.ts de T2). ──

// entradaResueltaSinCanonico — story IdentidadResuelta: (a)/(b) de T2 son identidades
// PROVISIONALES (home=""), (c) trae `canonico` poblado. Falta el cruce "home resuelto SIN
// canónico local" (el caso más común de una instalación observada de un arnés cuyo autor sos
// vos, pero todavía no clonaste el canónico acá). `registries` usa un mirror distinto del
// `home` a propósito (home = declarado en el manifiesto, registry = de dónde se adquirió esta
// copia — pueden diferir, BR-3) para que ambos strings sean distinguibles en los asserts.
const entradaResueltaSinCanonico: EntradaPortafolio = {
  clave: "github-com-acme-foo-cli~foo-cli~",
  identidad: { home: "github.com/acme/foo-cli", id: "foo-cli" },
  nombre: "Foo CLI",
  empresas: ["acme"],
  registries: ["github.com/acme/foo-cli-mirror"],
  instalaciones: [
    {
      proyecto_path: "~/Proyectos/foo-app",
      install_path: "~/Proyectos/foo-app/.claude/plugins/foo-cli",
      tipo: "materializada",
      origen: { registry: "github.com/acme/foo-cli-mirror", version: "1.0.0" },
      deriva: "al-hilo",
    },
  ],
  agregado: "2026-07-12T10:00:00Z",
}

// entradaSinCanonicoSinInstalaciones — stories SinInstalaciones + FocoYTeclado: ninguna
// fixture de T2 tiene 0 instalaciones (G5 exige el estado honesto "sin instalaciones
// registradas", nunca una lista vacía silenciosa). Deliberadamente mínima para FocoYTeclado:
// sin canónico (CTA disabled, fuera del trap) y sin instalaciones (sin botones Reparar/
// Backport/Observar) deja SOLO 2 focuseables (cerrar · Desvincular) — el trap de Tab se
// testea limpio, sin contar botones de más.
const entradaSinCanonicoSinInstalaciones: EntradaPortafolio = {
  clave: "sin-home~bar-cli~github-com-acme-bar-app",
  identidad: { id: "bar-cli", scope: "github.com/acme/bar-app" },
  instalaciones: [],
  agregado: "2026-07-11T08:30:00Z",
}

// entradaSinSello — story IdentificarSinSello (S1-D28): una presencia SIN manifiesto
// (`identidad.id` vacío) escaneada en su raíz → se ofrece «Identificar» para sellar in-situ.
const entradaSinSello: EntradaPortafolio = {
  clave: "sin-home~~abc123def456",
  identidad: { id: "", scope: "." },
  instalaciones: [
    {
      proyecto_path: "~/Proyectos/cruda",
      install_path: "~/Proyectos/cruda",
      tipo: "proyecto-instalado",
      origen: {},
      deriva: "deriva-no-evaluable",
    },
  ],
  agregado: "2026-07-17T10:00:00Z",
}

const meta = {
  title: "widgets/portafolio/PortafolioDrawer",
  component: PortafolioDrawer,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: {
    entrada: entradaHarnessEnDeriva,
    onClose: fn(),
    onDesvincular: fn(),
  },
} satisfies Meta<typeof PortafolioDrawer>

export default meta
type Story = StoryObj<typeof meta>

// IdentidadResuelta — home visible; facetas (empresa+marketplace); zona canónico AUSENTE con
// CTA disabled+tooltip S2; acciones staged Reparar/Backport disabled+"próximo · S5" (S1-D11);
// línea de frescura visible (S1-D15).
export const IdentidadResuelta: Story = {
  args: { entrada: entradaResueltaSinCanonico },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(c.getByText("github.com/acme/foo-cli")).toBeInTheDocument()
    await expect(c.getByText("acme")).toBeInTheDocument()
    // "github.com/acme/foo-cli-mirror" aparece 2 veces (facet marketplaces + origen-de-la-copia
    // de la instalación) — se scopea al facet para un assert sin ambigüedad.
    const facetas = canvasElement.querySelector(".pf-drawer-facetas") as HTMLElement
    await expect(within(facetas).getByText("github.com/acme/foo-cli-mirror")).toBeInTheDocument()

    const traer = c.getByRole("button", { name: "↧ Traer canónico" })
    await expect(traer).toBeDisabled()
    await expect(traer).toHaveAttribute("title", "próximo · S2")

    const reparar = c.getByRole("button", { name: "⚒ Reparar" })
    await expect(reparar).toBeDisabled()
    await expect(reparar).toHaveAttribute("title", "próximo · S5")
    const backport = c.getByRole("button", { name: "↩ Backport" })
    await expect(backport).toBeDisabled()
    await expect(backport).toHaveAttribute("title", "próximo · S5")

    await expect(c.getByText(/datos del último escaneo — agregado 2026-07-12/)).toBeInTheDocument()
  },
}

// IdentidadProvisional — "identidad provisional (sin home)" + scope visible (G6); facet
// marketplaces "desconocido"; empresa "desconocida" (G4). Fixture (b) real de T2: `identidad.id`
// literalmente vacío (sin manifiesto), scope = remote canonicalizado.
export const IdentidadProvisional: Story = {
  args: { entrada: entradaProyectoInstaladoProvisional },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    await expect(
      c.getByText(
        /identidad provisional \(sin home\) · scope: github\.com\/alpacapurpura\/harness-studio/,
      ),
    ).toBeInTheDocument()

    // Scopeado a .pf-drawer-facetas: "desconocido" también aparece en el origen-de-la-copia de
    // la instalación (misma fixture, sin registry) — se distingue del facet por contenedor.
    const facetas = canvasElement.querySelector(".pf-drawer-facetas") as HTMLElement
    const cf = within(facetas)
    await expect(cf.getByText("desconocida")).toBeInTheDocument()
    await expect(cf.getByText("desconocido")).toBeInTheDocument()
  },
}

// IdentificarSinSello — S1-D28 + B1: una presencia sin manifiesto muestra la marca roja
// `manifiesto-ausente` y ofrece «Identificar»; el form pide TAMBIÉN rol/proceso/empresa
// (B-D1: graph.l0 los exige, jamás se inventan) y sellar dispara onIdentificar con el objeto
// DatosIdentificar completo — el sello se escribe in-situ.
export const IdentificarSinSello: Story = {
  args: { entrada: entradaSinSello, onIdentificar: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await expect(c.getByText("manifiesto-ausente")).toBeInTheDocument()

    const abrir = c.getByRole("button", { name: "✦ Identificar" })
    await expect(abrir).toBeEnabled()
    await userEvent.click(abrir)

    await userEvent.type(c.getByRole("textbox", { name: "id" }), "mi-cruda")
    // El input de rol lleva `list` (datalist) ⇒ su role ARIA es combobox, no textbox.
    await userEvent.type(c.getByRole("combobox", { name: "rol" }), "dev-full-cycle")
    await userEvent.type(c.getByRole("textbox", { name: "proceso" }), "delivery")
    await userEvent.type(c.getByRole("textbox", { name: "empresa" }), "vitalia")

    await userEvent.click(c.getByRole("button", { name: "Sellar in-situ" }))
    await waitFor(() =>
      expect(args.onIdentificar).toHaveBeenCalledWith({
        installPath: "~/Proyectos/cruda",
        id: "mi-cruda",
        nombre: "",
        rol: "dev-full-cycle",
        proceso: "delivery",
        empresas: ["vitalia"],
        marketplace: "",
      }),
    )
  },
}

// IdentificarFormIncompletoDisabled — B-D1: sin rol/proceso/empresa el submit queda disabled
// con el motivo en title (el 400 del backend es el enforcement; el botón es la puerta
// honesta) y onIdentificar NO se dispara.
export const IdentificarFormIncompletoDisabled: Story = {
  args: { entrada: entradaSinSello, onIdentificar: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await userEvent.click(c.getByRole("button", { name: "✦ Identificar" }))
    // Solo el id: rol/proceso/empresa vacíos (la fixture no trae empresas que prellenar).
    await userEvent.type(c.getByRole("textbox", { name: "id" }), "mi-cruda")

    const sellar = c.getByRole("button", { name: "Sellar in-situ" })
    await expect(sellar).toBeDisabled()
    await expect(sellar).toHaveAttribute(
      "title",
      "rol, proceso y empresa son obligatorios — el sello debe pasar el contrato graph.l0",
    )
    await expect(args.onIdentificar).not.toHaveBeenCalled()
  },
}

// IdentificarFormCompletoValido — B-D1 lado verde: empresa PRELLENADA desde
// entrada.empresas[0]; el datalist de rol sugiere los `rolesConocidos` que llegan por props
// (widget puro — el fetch de GET /api/harnesses vive en la página); con rol/proceso
// completados el submit se habilita y entrega la empresa prellenada + el marketplace opcional.
export const IdentificarFormCompletoValido: Story = {
  args: {
    entrada: { ...entradaSinSello, empresas: ["vitalia"] },
    onIdentificar: fn(),
    rolesConocidos: ["dev-full-cycle", "legal-administrativo"],
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)

    await userEvent.click(c.getByRole("button", { name: "✦ Identificar" }))

    // Empresa prellenada de la faceta ya conocida (entrada.empresas[0]).
    await expect(c.getByRole("textbox", { name: "empresa" })).toHaveValue("vitalia")

    // El input de rol es texto libre CON datalist (sugerencia, no catálogo).
    const rolInput = c.getByRole("combobox", { name: "rol" })
    const listId = rolInput.getAttribute("list")
    await expect(listId).toBeTruthy()
    const opciones = listId
      ? Array.from(document.getElementById(listId)?.querySelectorAll("option") ?? [])
      : []
    await expect(opciones.map((o) => o.getAttribute("value"))).toEqual([
      "dev-full-cycle",
      "legal-administrativo",
    ])

    const sellar = c.getByRole("button", { name: "Sellar in-situ" })
    await expect(sellar).toBeDisabled() // rol/proceso aún vacíos

    await userEvent.type(rolInput, "dev-full-cycle")
    await userEvent.type(c.getByRole("textbox", { name: "proceso" }), "delivery")
    await userEvent.type(
      c.getByRole("textbox", { name: "marketplace (opcional)" }),
      "vitalia/arneses",
    )
    await expect(sellar).toBeEnabled()

    await userEvent.click(sellar)
    await waitFor(() =>
      expect(args.onIdentificar).toHaveBeenCalledWith({
        installPath: "~/Proyectos/cruda",
        id: "",
        nombre: "",
        rol: "dev-full-cycle",
        proceso: "delivery",
        empresas: ["vitalia"],
        marketplace: "vitalia/arneses",
      }),
    )
  },
}

// InstalacionEnDerivaReal — chip en-deriva + detalle en title (G1); tipo referenciada-cc;
// origen de la copia con registry+version reales (fixture (a) de T2, harness@0.5.2 en-deriva).
export const InstalacionEnDerivaReal: Story = {
  args: { entrada: entradaHarnessEnDeriva },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)

    const chip = canvasElement.querySelector(".pf-chip-deriva")
    await expect(chip).toHaveClass("en-deriva")
    await expect(chip).toHaveAttribute(
      "title",
      "hash de contenido distinto de la referencia " +
        "~/.claude/plugins/marketplaces/prenter-marketplace/plugins/harness/0.5.2/",
    )

    await expect(c.getByText("referenciada-cc")).toBeInTheDocument()
    await expect(c.getByText("v0.5.2")).toBeInTheDocument()
    // "alpacapurpura/prenter-marketplace" aparece 2 veces (facet marketplaces, vía
    // `registriesDe` que lo suma crudo desde la instalación, + origen-de-la-copia) — se scopea a
    // la zona de instalaciones, que es lo que esta story verifica (BR-3/S1-D14).
    const origenCopia = canvasElement.querySelector(".pf-origen-copia") as HTMLElement
    await expect(
      within(origenCopia).getByText("alpacapurpura/prenter-marketplace"),
    ).toBeInTheDocument()
  },
}

// UpdateNoVerificado (G2/S1-D5) — el texto EXACTO "update: no-verificado" + tooltip S4;
// AUSENCIA total de cualquier flag "⬆" fabricado en todo el drawer.
export const UpdateNoVerificado: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const linea = c.getByText("update: no-verificado")
    await expect(linea).toBeInTheDocument()
    await expect(linea).toHaveAttribute("title", "update-check llega en Slice 4")
    await expect(c.queryByText(/⬆/)).toBeNull()
  },
}

// ConDiscrepancias (C-OR-6) — las discrepancias del origen visibles como texto plano, jamás
// resueltas en silencio (fixture (c), instalación #2).
export const ConDiscrepancias: Story = {
  args: { entrada: entradaCanonicaCompleta },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(/home declarado \(github\.com\/acme\/acme-cli\) ≠ registry de adquisición/),
    ).toBeInTheDocument()
  },
}

// TrazabilidadEslabones (BR-3, S1-D14) — el <details> arranca colapsado, se expande al click y
// lista fuente·campo·valor crudos (fixture (c), instalación #2, 3 eslabones).
export const TrazabilidadEslabones: Story = {
  args: { entrada: entradaCanonicaCompleta },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const details = canvasElement.querySelector(".pf-trazabilidad") as HTMLDetailsElement
    await expect(details).not.toBeNull()
    await expect(details.open).toBe(false)

    await userEvent.click(c.getByText("trazabilidad"))
    await expect(details.open).toBe(true)

    await expect(c.getByText("manifiesto · home · github.com/acme/acme-cli")).toBeInTheDocument()
    await expect(c.getByText("cc-plugins · registry · acme-fork/acme-cli")).toBeInTheDocument()
    await expect(c.getByText("cc-plugins · version · 2.0.0")).toBeInTheDocument()
  },
}

// ConCanonico — fixture con canónico poblado (path+version): "estado del checkout: no evaluado
// en este slice" visible, "Abrir en Mapa" HABILITADO dispara onObservar(canonico.path) con el
// path exacto (fixture (c): ~/dev/acme-cli).
export const ConCanonico: Story = {
  args: { entrada: entradaCanonicaCompleta, onObservar: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("~/dev/acme-cli")).toBeInTheDocument()
    await expect(
      c.getByText("v2.1.0 · estado del checkout: no evaluado en este slice"),
    ).toBeInTheDocument()

    const abrir = c.getByRole("button", { name: "◉ Abrir en Mapa" })
    await expect(abrir).toBeEnabled()
    await userEvent.click(abrir)
    await expect(args.onObservar).toHaveBeenCalledWith("~/dev/acme-cli")
  },
}

// SinInstalaciones (G5) — copy honesto "sin instalaciones registradas", nunca una lista vacía
// silenciosa.
export const SinInstalaciones: Story = {
  args: { entrada: entradaSinCanonicoSinInstalaciones },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("sin instalaciones registradas")).toBeInTheDocument()
  },
}

// ObservarSinSesion (S1-D13) — `onObservar` undefined → TODOS los botones de observar (zona
// Canónico Y cada instalación) quedan disabled + title = observarDisabledMotivo.
export const ObservarSinSesion: Story = {
  args: {
    entrada: entradaCanonicaCompleta,
    onObservar: undefined,
    observarDisabledMotivo: "necesita una sesión abierta — el Mapa vive en el stage de sesión",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const motivo = "necesita una sesión abierta — el Mapa vive en el stage de sesión"

    const abrir = c.getByRole("button", { name: "◉ Abrir en Mapa" })
    await expect(abrir).toBeDisabled()
    await expect(abrir).toHaveAttribute("title", motivo)

    const observarBtns = c.getAllByRole("button", { name: "◉ Observar en Mapa" })
    await expect(observarBtns.length).toBeGreaterThan(0)
    for (const btn of observarBtns) {
      await expect(btn).toBeDisabled()
      await expect(btn).toHaveAttribute("title", motivo)
    }
  },
}

// ConfirmarDesvincular (G7) — click Desvincular → dialog con copy EXACTO, checkbox
// "…y borrar clon local" disabled+tooltip S2; Cancelar cierra SIN llamar onDesvincular;
// Confirmar SÍ lo llama; `desvincularError` visible cuando se pasa.
export const ConfirmarDesvincular: Story = {
  args: {
    entrada: entradaHarnessEnDeriva,
    desvincularError: "DELETE /api/portafolio/arneses/…: 404 no encontrado (ya desvinculado)",
  },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const abrirConfirm = () => c.getByRole("button", { name: "⊘ Desvincular" })

    await userEvent.click(abrirConfirm())

    await expect(
      c.getByText(/solo lo saca del portafolio — no desinstala ni borra nada del disco/),
    ).toBeInTheDocument()

    const checkbox = c.getByRole("checkbox", { name: "…y borrar clon local" })
    await expect(checkbox).toBeDisabled()
    await expect(checkbox).toHaveAttribute("title", "próximo · S2")

    await expect(c.getByText(/404 no encontrado \(ya desvinculado\)/)).toBeInTheDocument()

    // Cancelar: cierra el confirm SIN llamar onDesvincular.
    await userEvent.click(c.getByRole("button", { name: "Cancelar" }))
    await expect(args.onDesvincular).not.toHaveBeenCalled()
    await expect(c.queryByText(/solo lo saca del portafolio/)).toBeNull()

    // Reabrir y Confirmar: SÍ llama onDesvincular.
    await userEvent.click(abrirConfirm())
    await userEvent.click(c.getByRole("button", { name: "Confirmar" }))
    await expect(args.onDesvincular).toHaveBeenCalledTimes(1)
  },
}

// FocoYTeclado (G8, CRÍTICO) — role="dialog" + aria-modal + aria-labelledby apuntando al
// título; foco inicial DENTRO del drawer (el botón cerrar); trap de Tab en ambas direcciones;
// Escape llama onClose. Fixture mínima (2 focuseables) para un trap sin ruido.
// biome-ignore lint/style/useNamingConvention: nombre de story literal del plan (T5 §3) — "YT" no es un acrónimo, es la conjunción "y".
export const FocoYTeclado: Story = {
  args: { entrada: entradaSinCanonicoSinInstalaciones },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const dialog = c.getByRole("dialog")
    await expect(dialog).toHaveAttribute("aria-modal", "true")

    const labelledby = dialog.getAttribute("aria-labelledby")
    await expect(labelledby).toBeTruthy()
    const titleEl = labelledby ? document.getElementById(labelledby) : null
    await expect(titleEl).not.toBeNull()
    await expect(titleEl).toHaveTextContent("bar-cli")

    const closeBtn = c.getByRole("button", { name: "Cerrar" })
    const desvincularBtn = c.getByRole("button", { name: "⊘ Desvincular" })

    // Foco inicial: cae en el botón cerrar al montar.
    await waitFor(() => expect(closeBtn).toHaveFocus())

    // Trap hacia adelante: Tab desde el último focuseable vuelve al primero.
    await userEvent.tab()
    await expect(desvincularBtn).toHaveFocus()
    await userEvent.tab()
    await expect(closeBtn).toHaveFocus()

    // Trap hacia atrás: Shift+Tab desde el primero va al último.
    await userEvent.tab({ shift: true })
    await expect(desvincularBtn).toHaveFocus()

    // Escape llama onClose.
    await userEvent.keyboard("{Escape}")
    await expect(args.onClose).toHaveBeenCalledTimes(1)
  },
}

// ══════════════════════════════════════════════════════════════════════════════════════════
// Paquete 2026-07-23-portafolio-agregar-marketplace — S8 (`↧ Traer canónico`, AG-D17) y S7
// (segunda puerta a la reconciliación). SUPERSET: sin estos callbacks el drawer rinde EXACTAMENTE
// como en el Slice 1 (`IdentidadResuelta` sigue asertando `disabled` + «próximo · S2»).
// ══════════════════════════════════════════════════════════════════════════════════════════

// E-107 / AG-D17 — con `onTraerCanonico` el botón que ya existía se HABILITA (dos puertas, un
// acto: la fila del catálogo llama al mismo callback) y **deja de llevar `TOOLTIP_S2`**.
export const DrawerTraerHabilitado: Story = {
  args: { entrada: entradaResueltaSinCanonico, onTraerCanonico: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const traer = c.getByRole("button", { name: "↧ Traer canónico" })
    await expect(traer).toBeEnabled()
    await expect(traer).not.toHaveAttribute("title")
    await userEvent.click(traer)
    await expect(args.onTraerCanonico).toHaveBeenCalledTimes(1)
  },
}

// E-107 — `trayendo`: botón bloqueado + `aria-busy`; no se puede disparar dos veces el mismo.
export const DrawerTrayendo: Story = {
  args: { entrada: entradaResueltaSinCanonico, onTraerCanonico: fn(), trayendo: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Trayendo…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")
    await expect(c.queryByRole("button", { name: "↧ Traer canónico" })).toBeNull()
  },
}

// E-107 — `falló`: el motivo textual del backend, LITERAL (incluido el rechazo por clase
// `referencia`, que es el dominio hablando: BR-1 no se re-implementa en el FE).
export const DrawerTraerFallo: Story = {
  args: {
    entrada: entradaResueltaSinCanonico,
    onTraerCanonico: fn(),
    traerError:
      "arnesia POST /api/marketplaces/claude-plugins-official/traidos: 400 traer: no aplica: solo arneses propios",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("no aplica: solo arneses propios")
    await expect(c.getByRole("button", { name: "↧ Traer canónico" })).toBeEnabled()
  },
}

// DD-2/E-a (deudas de dogfood 2026-07-30) — «⌂ Adoptar como canónico», el REVERSO de Traer:
// la instalación local forjada sube al checkout del marketplace-home. Solo existe cuando la
// página pasa `onAdoptar` (entrada sin canónico + instalación local); sin la prop el DOM es
// EXACTAMENTE el de antes (superset estricto — asertado por las stories firmadas vigentes).
export const DrawerAdoptarHabilitado: Story = {
  args: { entrada: entradaResueltaSinCanonico, onAdoptar: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const adoptar = c.getByRole("button", { name: "⌂ Adoptar como canónico" })
    await expect(adoptar).toBeEnabled()
    await userEvent.click(adoptar)
    await expect(args.onAdoptar).toHaveBeenCalledTimes(1)
  },
}

// DD-2/E-a — adoptando: bloqueado + aria-busy; y el fallo muestra el motivo LITERAL del
// dominio (p. ej. sello sin marketplace-home) — el FE no re-implementa las guardas.
export const DrawerAdoptarFallo: Story = {
  args: {
    entrada: entradaResueltaSinCanonico,
    onAdoptar: fn(),
    adoptarError:
      "arnesia POST /api/portafolio/adopciones: 400 adoptar: el sello no declara marketplace-home y el request no nombra uno",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("no declara marketplace-home")
    await expect(c.getByRole("button", { name: "⌂ Adoptar como canónico" })).toBeEnabled()
  },
}

// DD-2/E-a — sin `onAdoptar` el botón NO existe (regresión del superset estricto).
export const DrawerSinAdoptar: Story = {
  args: { entrada: entradaResueltaSinCanonico },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByRole("button", { name: "⌂ Adoptar como canónico" })).toBeNull()
  },
}

// AG-D8 decisión 7 — la reconciliación se ejecuta ACÁ, in-situ sobre la ficha (mismo patrón que
// «Identificar»). Solo se ofrece con identidad PROVISIONAL: un home ya declarado no necesita
// resolverse.
export const DrawerResolverOrigen: Story = {
  args: { entrada: entradaProyectoInstaladoProvisional, onResolverOrigen: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/no puede decir de qué marketplace viene/)).toBeInTheDocument()
    await userEvent.click(c.getByRole("button", { name: "Resolver origen" }))
    await expect(args.onResolverOrigen).toHaveBeenCalledTimes(1)
  },
}

// …y con home declarado NO se ofrece (nada que resolver): la puerta no aparece.
export const DrawerSinPuertaConHome: Story = {
  args: { entrada: entradaResueltaSinCanonico, onResolverOrigen: fn() },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByRole("button", { name: "Resolver origen" })).toBeNull()
  },
}

// ══════════════════════════════════════════════════════════════════════════════════════════
// Paquete 2026-07-30-volverlo-de-arnesia-y-publicar — B2 (`▲ Publicar`, B-D5). SUPERSET
// estricto: sin `onPublicar` el botón sigue `disabled` + «próximo · S3» exactamente como antes
// (las stories firmadas que no pasan el callback no cambian). Estas stories SON el mockup
// (B-D4: cero html nuevo).
// ══════════════════════════════════════════════════════════════════════════════════════════

// B2 — con `onPublicar` el botón que ya existía se HABILITA y pierde el tooltip de diferido.
// Vive en la zona Canónico: solo se publica la única copia editable (ley anti-drift).
export const DrawerPublicarHabilitado: Story = {
  args: { entrada: entradaCanonicaCompleta, onPublicar: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const publicar = c.getByRole("button", { name: "▲ Publicar" })
    await expect(publicar).toBeEnabled()
    await expect(publicar).not.toHaveAttribute("title")
    await userEvent.click(publicar)
    await expect(args.onPublicar).toHaveBeenCalledTimes(1)
  },
}

// B2 — `publicando`: botón bloqueado + `aria-busy`; no se dispara dos veces el mismo push.
export const DrawerPublicando: Story = {
  args: { entrada: entradaCanonicaCompleta, onPublicar: fn(), publicando: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Publicando…" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("aria-busy", "true")
    await expect(c.queryByRole("button", { name: "▲ Publicar" })).toBeNull()
  },
}

// B2 — `falló`: el motivo LITERAL del backend en `role="alert"`, incluidos los checks FAIL que
// la página extrajo del 409 del gate de conformance (RF-B2.6) — el FE presenta, no re-calcula.
export const DrawerPublicarError: Story = {
  args: {
    entrada: entradaCanonicaCompleta,
    onPublicar: fn(),
    publicarError:
      "arnesia POST /api/portafolio/arneses/github-com-acme-acme-cli~acme-cli~/publicaciones: 409 " +
      "publicar: el gate de conformance está en rojo · checks en rojo: escritor-unico",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("alert")).toHaveTextContent("checks en rojo: escritor-unico")
    // El botón queda re-disparable: un gate rojo se arregla y se reintenta.
    await expect(c.getByRole("button", { name: "▲ Publicar" })).toBeEnabled()
  },
}

// Superset: SIN `onPublicar` la zona Canónico queda EXACTAMENTE como el Slice 1 la firmó —
// `disabled` + «próximo · S3».
export const DrawerPublicarSinCallback: Story = {
  args: { entrada: entradaCanonicaCompleta },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const publicar = c.getByRole("button", { name: "▲ Publicar" })
    await expect(publicar).toBeDisabled()
    await expect(publicar).toHaveAttribute("title", "próximo · S3")
  },
}
