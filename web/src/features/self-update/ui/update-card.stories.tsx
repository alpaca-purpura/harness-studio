import type { Meta, StoryObj } from "@storybook/react-vite"
import type { ReactNode } from "react"
import { expect, fn, within } from "storybook/test"
import type { SelfUpdateReport, VersionInfo } from "../model/types"
import { UpdateCard } from "./update-card"

// La tarjeta vive en la vista Ajustes (scoped .arnesia-ajustes) — el frame replica ese scope.
function Frame({ children }: { children: ReactNode }) {
  return <div className="arnesia-ajustes">{children}</div>
}

// Identidad OBJETIVO (mockup caso 01): instalado en ~/.local/bin, repo configurado.
const versionOk: VersionInfo = {
  huella: "1c7443f",
  fecha: "2026-07-07",
  instalado_en: "/home/chalreme/.local/bin/arnesia",
  escribible: true,
  repo: "/home/chalreme/Proyectos/harness-studio",
  sucio: false,
  version: "0.2.21.2607260225",
  compilado: "2026-07-26 02:25",
}

// Estado ACTUAL real del operador (mockup caso 05): /usr/bin root + sin repo.
const versionHoy: VersionInfo = {
  huella: "dev",
  fecha: "",
  instalado_en: "/usr/bin/arnesia",
  escribible: false,
  repo: "",
  sucio: false,
  version: "dev",
}

// RF-231 — el caso que motivó el sello: `make dev-sync` reemplazó el binario mientras la app
// seguía abierta. El archivo en disco ya es nuevo; el proceso que responde es el viejo.
const versionBuildViejo: VersionInfo = {
  ...versionOk,
  aviso_build:
    "el binario instalado es más nuevo (2026-07-26 03:18) que el que está corriendo — cerrá y reabrí la app",
}

const reporteExito: SelfUpdateReport = {
  resultado: "actualizado",
  huella_nueva: "a1b2c3d",
  pasos: [
    { paso: "verificar", estado: "ok", detalle: "repo OK · go/pnpm/bash presentes" },
    { paso: "build", estado: "ok", detalle: "scripts/bundle.sh --daemon-only OK (SPA + go build)" },
    { paso: "verificar binario", estado: "ok", detalle: "bin/arnesia · huella a1b2c3d" },
    {
      paso: "instalar",
      estado: "ok",
      detalle: "instalado en ~/.local/bin/arnesia (rename atómico, sin sudo)",
    },
    {
      paso: "reiniciar",
      estado: "agendado",
      detalle: "re-exec en ≤1s — la UI reconecta por /api/version",
    },
  ],
}

const reporteFalloBuild: SelfUpdateReport = {
  resultado: "fallo",
  pasos: [
    { paso: "verificar", estado: "ok", detalle: "repo OK · go/pnpm/bash presentes" },
    {
      paso: "build",
      estado: "fallo",
      detalle: "bundle.sh --daemon-only: exit status 1\n./web: pnpm build: error TS2345 …stderr…",
    },
    { paso: "verificar binario", estado: "no-corrido" },
    { paso: "instalar", estado: "no-corrido" },
    { paso: "reiniciar", estado: "no-corrido" },
  ],
}

const meta = {
  title: "features/self-update/UpdateCard",
  component: UpdateCard,
  decorators: [(Story) => <Frame>{Story()}</Frame>],
  args: { version: versionOk, estado: "idle", onUpdate: fn() },
} satisfies Meta<typeof UpdateCard>

export default meta
type Story = StoryObj<typeof meta>

// RF-101 — identidad honesta: huella·fecha·ruta·pill·repo reales; click dispara onUpdate.
export const Idle: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/1c7443f/)).toBeInTheDocument()
    await expect(c.getByText(/2026-07-07/)).toBeInTheDocument()
    await expect(c.getByText(/\.local\/bin\/arnesia/)).toBeInTheDocument()
    await expect(c.getByText("espacio de usuario · sin sudo")).toBeInTheDocument()
    await expect(c.getByText(/Proyectos\/harness-studio/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Actualizar desde el repo" })
    await expect(btn).toBeEnabled()
    await btn.click()
    await expect(args.onUpdate).toHaveBeenCalled()
  },
}

// RF-231 — la identidad del BUILD, que es lo que responde «¿corro lo último que compilé?».
// La huella sola NO lo responde: dos compilaciones del mismo árbol la comparten (pasó de
// verdad el 2026-07-26, bundleando v0.2.21 dos veces).
export const IdentidadDelBuild: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("arnesia v0.2.21.2607260225")).toBeInTheDocument()
    await expect(c.getByText(/2026-07-26 02:25/)).toBeInTheDocument()
    // El commit sigue estando: es otro dato, no el mismo.
    await expect(c.getByText(/1c7443f/)).toBeInTheDocument()
    // Al día = NINGÚN aviso. Un aviso que aparece siempre no se lee nunca.
    await expect(c.queryByText(/más nuevo/)).not.toBeInTheDocument()
  },
}

// RF-231 — el caso que lo motivó: `make dev-sync` reemplazó el binario con la app abierta.
export const BuildViejoCorriendo: Story = {
  args: { version: versionBuildViejo },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/el binario instalado es más nuevo/)).toBeInTheDocument()
    await expect(c.getByText(/cerrá y reabrí la app/)).toBeInTheDocument()
  },
}

// RF-231 — un build sin sellar (CI, `go run`) lo DICE en vez de inventar un número.
export const BuildSinSellar: Story = {
  args: { version: versionHoy },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/build sin sellar \(dev\)/)).toBeInTheDocument()
  },
}

// RF-104 — actualizando: botón bloqueado y SIN checklist inventada (honestidad del spec).
export const Actualizando: Story = {
  args: { estado: "actualizando" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Actualizando…" })
    await expect(btn).toBeDisabled()
    // Sin reporte no hay pasos: la checklist solo pinta veredictos REALES.
    await expect(c.queryByRole("list")).not.toBeInTheDocument()
    await expect(c.getByText(/jamás un spinner mudo/)).toBeInTheDocument()
  },
}

// RF-104 — error de build: checklist real (fallo + no-corridos), stderr visible,
// binario intacto dicho, Reintentar habilitado.
export const ErrorDeBuild: Story = {
  args: { estado: "error", reporte: reporteFalloBuild },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const items = c.getAllByRole("listitem")
    await expect(items).toHaveLength(5)
    await expect(items[1]).toHaveClass("fail")
    await expect(items[2]).not.toHaveClass("fail")
    // El stderr vive en el paso Y en la critbox (ambos reales): al menos una vez.
    await expect(c.getAllByText(/TS2345/)).not.toHaveLength(0)
    await expect(c.getByText(/NO se tocó/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Reintentar" })).toBeEnabled()
  },
}

// RF-102 — no-escribible (estado ACTUAL real): pill warn + warnbox con la migración
// única + botón disabled con el porqué en el title.
export const NoEscribible: Story = {
  args: { version: versionHoy },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("root · actualizar pide sudo")).toBeInTheDocument()
    await expect(c.getByText(/install -m755/)).toBeInTheDocument()
    await expect(c.getByText(/versión no embebida/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Actualizar desde el repo" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("title", "Migra a ~/.local/bin para actualizar sin sudo")
  },
}

// RF-103 — sin repo: estado honesto + disabled con el porqué.
export const SinRepo: Story = {
  args: { version: { ...versionOk, repo: "" } },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/repo no configurado/)).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Actualizar desde el repo" })
    await expect(btn).toBeDisabled()
    await expect(btn).toHaveAttribute("title", "Configura --repo / ARNESIA_REPO al daemon")
  },
}

// RF-110 (bugfix fix-repo-self-update) — sin repo DENTRO de Tauri: gana el botón
// «Elegir carpeta…» (la página solo pasa onElegirRepo cuando isTauri()).
export const SinRepoConSelectorNativo: Story = {
  args: { version: { ...versionOk, repo: "" }, onElegirRepo: fn() },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const btn = c.getByRole("button", { name: "Elegir carpeta…" })
    await expect(btn).toBeEnabled()
    await btn.click()
    await expect(args.onElegirRepo).toHaveBeenCalled()
  },
}

// RF-110 — configurando (PUT en vuelo): botón del picker bloqueado, texto «Configurando…».
export const ConfigurandoRepo: Story = {
  args: { version: { ...versionOk, repo: "" }, onElegirRepo: fn(), configurandoRepo: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Configurando…" })).toBeDisabled()
  },
}

// RF-109 — candidato inválido: el motivo exacto del servidor, nunca un fallo mudo.
export const RepoConfigInvalido: Story = {
  args: {
    version: { ...versionOk, repo: "" },
    onElegirRepo: fn(),
    repoConfigError:
      "el repo configurado no es un árbol Go legible: open /tmp/go.mod: no such file or directory",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/árbol Go legible/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Elegir carpeta…" })).toBeEnabled()
  },
}

// RF-105 — reiniciando: checklist real con reiniciar «en curso», botón bloqueado,
// nota del polling.
export const Reiniciando: Story = {
  args: { estado: "reiniciando", reporte: reporteExito },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const items = c.getAllByRole("listitem")
    await expect(items[4]).toHaveClass("doing")
    await expect(c.getByRole("button", { name: "Reiniciando…" })).toBeDisabled()
    await expect(c.getByText(/reconecta sola por \/api\/version/)).toBeInTheDocument()
  },
}

// RF-105 — éxito: 5 pasos consumados + okbox con la huella NUEVA confirmada.
export const Exito: Story = {
  args: {
    estado: "exito",
    reporte: reporteExito,
    version: { ...versionOk, huella: "a1b2c3d" },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    for (const item of c.getAllByRole("listitem")) {
      await expect(item).toHaveClass("done")
    }
    await expect(c.getByText(/daemon reiniciado, UI reconectada/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Actualizar desde el repo" })).toBeEnabled()
  },
}

// RF-104 ③ — ya al día: éxito sin reinstalar ni reiniciar (instalar/reiniciar no-corridos).
export const YaAlDia: Story = {
  args: {
    estado: "ya-al-dia",
    reporte: {
      resultado: "ya-al-dia",
      huella_nueva: "1c7443f",
      pasos: [
        { paso: "verificar", estado: "ok", detalle: "repo OK" },
        { paso: "build", estado: "ok", detalle: "bundle OK" },
        { paso: "verificar binario", estado: "ok", detalle: "bin/arnesia · huella 1c7443f" },
        {
          paso: "instalar",
          estado: "no-corrido",
          detalle: "misma huella que el binario corriendo — nada que instalar",
        },
        { paso: "reiniciar", estado: "no-corrido", detalle: "sin reinicio" },
      ],
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Ya estás al día/)).toBeInTheDocument()
    const items = c.getAllByRole("listitem")
    await expect(items[3]).not.toHaveClass("done")
    await expect(c.getByRole("button", { name: "Actualizar desde el repo" })).toBeEnabled()
  },
}

// Honestidad extra — GET /api/version caído: la tarjeta lo dice, jamás inventa identidad.
export const VersionNoDisponible: Story = {
  args: { version: null, versionError: "arnesia GET /api/version: fetch failed" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/GET \/api\/version falló/)).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Actualizar desde el repo" })).toBeDisabled()
  },
}
