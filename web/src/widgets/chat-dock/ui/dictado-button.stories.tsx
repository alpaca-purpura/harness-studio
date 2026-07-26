import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { TOPE_MS, useDictado } from "@/shared"
import { DictadoAviso, DictadoButton, VoiceBar } from "./dictado-button"

// Story = test (fe-visual-fitness) — paquete 2026-07-25-spike-voz-dictado, RF-216..228.
// Dibujo de referencia: mockups/arnesia-voz-dictado.html.
//
// El store del dictado es global (Zustand), así que cada story fija el estado que quiere
// mostrar con `setState`. Es a propósito: los estados intermedios (escuchando · ordenando ·
// crudo · sin motor) dependen de hardware y de un daemon con motor instalado, y esperar a
// eso para poder MIRARLOS sería exactamente el drift entre Storybook y la app que la casa
// prohíbe. Lo que estas stories prueban es la SUPERFICIE; el loop real es T6, abierto.

// escenario deja el store en un estado concreto antes de renderizar.
function escenario(estado: Partial<ReturnType<typeof useDictado.getState>>) {
  return (Story: () => React.ReactNode) => {
    useDictado.setState({
      etapa: "inactivo",
      transcurrido: 0,
      nivel: 0,
      cortadoPorTope: false,
      fallo: undefined,
      disponibilidad: { disponible: true, motor: "whisper-cli" },
      ...estado,
    })
    return <div style={{ maxWidth: 420 }}>{Story()}</div>
  }
}

// Composer calca la fila vigente de chat-dock.tsx#Composer para que el botón se vea EN su
// contexto: el superset se juzga con lo vigente al lado, no en aislamiento.
function Composer({ texto = "", crudo }: { texto?: string; crudo?: string }) {
  return (
    <div className="flex flex-col rounded-lg border border-border bg-card">
      <VoiceBar />
      <div className="flex flex-none items-end gap-2 border-t border-border p-3">
        <div
          className={`max-h-[66px] min-h-[36px] flex-1 rounded-lg border bg-secondary px-2.5 py-2 text-xs ${
            crudo ? "border-warn" : "border-border"
          } ${texto ? "text-foreground" : "text-muted-foreground"}`}
        >
          {texto || "Pídele un cambio a vitalia-api"}
        </div>
        <DictadoButton sesionId="s1" onTexto={fn()} />
        <button
          type="button"
          disabled={!texto}
          className="grid size-9 flex-none place-items-center rounded-lg bg-primary text-sm text-primary-foreground disabled:opacity-40"
        >
          ↑
        </button>
      </div>
      <DictadoAviso crudoMotivo={crudo} />
    </div>
  )
}

const meta = {
  title: "widgets/chat-dock/Dictado",
  component: Composer,
  parameters: { layout: "padded", a11y: { test: "todo" } },
} satisfies Meta<typeof Composer>

export default meta
type Story = StoryObj<typeof meta>

// Reposo — el botón entra como UN BOTÓN MÁS, sin tocar el layout de la fila (RF-216).
export const Reposo: Story = {
  decorators: [escenario({})],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Dictar" })).toBeEnabled()
    // Nada de franja de estado en reposo: no le come alto permanente al dock.
    await expect(c.queryByText(/Escuchando/)).not.toBeInTheDocument()
  },
}

// Escuchando — el toggle muestra ■ y la franja nombra la etapa + el contador (RF-217/218).
export const Escuchando: Story = {
  decorators: [escenario({ etapa: "escuchando", transcurrido: 24_000, nivel: 0.7 })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Escuchando…")).toBeInTheDocument()
    await expect(c.getByText("0:24 / 3:00")).toBeInTheDocument()
    // El botón de dictar se volvió el de cortar: es el MISMO botón (toggle, V-D3).
    await expect(c.getByRole("button", { name: "Cortar el dictado" })).toBeInTheDocument()
    await expect(c.queryByRole("button", { name: "Dictar" })).not.toBeInTheDocument()
    // Cancelar existe y es distinto de cortar (RF-219).
    await expect(c.getByRole("button", { name: "cancelar" })).toBeInTheDocument()
  },
}

// CercaDelTope — el contador avisa antes de que el corte ocurra (RF-218).
export const CercaDelTope: Story = {
  decorators: [escenario({ etapa: "escuchando", transcurrido: TOPE_MS - 13_000, nivel: 0.5 })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("2:47 / 3:00")).toHaveClass(/text-warn/)
  },
}

// Transcribiendo — etapa CON NOMBRE. Ningún spinner anónimo (RF-228).
export const Transcribiendo: Story = {
  decorators: [escenario({ etapa: "transcribiendo" })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Transcribiendo…")).toBeInTheDocument()
    // Durante el trabajo el mic queda deshabilitado: no se encima un dictado sobre otro.
    await expect(c.getByRole("button", { name: "Dictar" })).toBeDisabled()
  },
}

// Ordenando — el paso caro (13-17 s, el 87 % del tiempo) tiene su propio escalón visible, y
// el escape a crudo de V-D4 está disponible EN VIVO.
export const Ordenando: Story = {
  decorators: [escenario({ etapa: "ordenando" })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Ordenando lo dictado…")).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "usar el crudo" })).toBeInTheDocument()
  },
}

// Limpio — el composer poblado y editable. La franja YA no está: el dictado terminó.
export const Limpio: Story = {
  decorators: [escenario({})],
  args: { texto: "Cambiá el color del pip de la PermissionCard a --warn" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/PermissionCard/)).toBeInTheDocument()
    // El enviar se habilitó, pero lo aprieta el operador: NUNCA se auto-envía (RF-226).
    await expect(c.getByRole("button", { name: "↑" })).toBeEnabled()
    await expect(c.queryByText(/sin ordenar/)).not.toBeInTheDocument()
  },
}

// Crudo — el transcripto sin ordenar se DISTINGUE del limpio (RF-227). Pasarlo por limpio
// sería el pass fabricado que la doctrina prohíbe.
export const Crudo: Story = {
  decorators: [escenario({})],
  args: {
    texto: "este eh… el pip de la tarjeta esa de permisos ponele el color de warning",
    crudo: "falló el paso de limpieza",
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/sin ordenar/)).toBeInTheDocument()
    await expect(c.getByText(/El dictado no se perdió/)).toBeInTheDocument()
  },
}

// CortadoPorTope — para solo, CONSERVA lo grabado, y avisa. Las tres cosas (RF-218).
export const CortadoPorTope: Story = {
  decorators: [escenario({ etapa: "transcribiendo", cortadoPorTope: true })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Se cortó por el tope/)).toBeInTheDocument()
    await expect(c.getByText(/no se descarta nada/)).toBeInTheDocument()
  },
}

// FalloDeEtapa — se nombra LA ETAPA que falló y se promete lo que importa: el composer no
// se tocó (RF-227 + RF-228).
export const FalloDeEtapa: Story = {
  decorators: [
    escenario({ fallo: { etapa: "transcribiendo", mensaje: "el motor no respondió." } }),
  ],
  args: { texto: "lo que ya tenía tecleado, intacto" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("transcribiendo")).toBeInTheDocument()
    await expect(c.getByText(/El\s+composer no se tocó/)).toBeInTheDocument()
    await expect(c.getByText("lo que ya tenía tecleado, intacto")).toBeInTheDocument()
  },
}

// SinMotorDeTranscripcion — la cara visible de la firma V-D1 (adaptador por PATH): el motivo se
// escribe EN LA SUPERFICIE y dice qué instalar, no solo que algo falta.
export const SinMotorDeTranscripcion: Story = {
  decorators: [
    escenario({
      disponibilidad: {
        disponible: false,
        motivo: "Falta el motor de transcripción.",
        instalar: ["whisper-cli", "faster-whisper"],
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Dictar" })).toBeDisabled()
    await expect(c.getByText(/Falta el motor de transcripción/)).toBeInTheDocument()
    await expect(c.getByText("whisper-cli")).toBeInTheDocument()
    // El ícono cambia, no solo el color: gris-idéntico no distingue apagado de prohibido.
    await expect(c.getByLabelText("micrófono no disponible")).toBeInTheDocument()
  },
}

// PermisoDenegado — se arregla distinto que «sin dispositivo», así que dice otra cosa.
export const PermisoDenegado: Story = {
  decorators: [
    escenario({
      disponibilidad: { disponible: true, motor: "whisper-cli" },
      fallo: {
        etapa: "escuchando",
        mensaje: "Micrófono denegado. Habilitalo en el sistema y reabrí la app.",
      },
    }),
  ],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Micrófono denegado/)).toBeInTheDocument()
  },
}

// NoSeSabeTodavia — mientras la sonda no contestó, el botón NO se ofrece habilitado: mejor
// esperar medio segundo que abrir un dictado que capaz no se puede transcribir.
export const NoSeSabeTodavia: Story = {
  decorators: [escenario({ disponibilidad: undefined })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Dictar" })).toBeDisabled()
  },
}

// CortarDisparaElFlujo — el toggle de V-D3, ejercido de verdad.
//
// La story fija DOS cosas: que el botón ES el toggle (su click llega al store) y —desde
// RF-229— que cortar sin captura viva **sale del estado** nombrando el motivo, en vez de
// quedarse en «Escuchando…» para siempre. Esa era la falla real de v0.2.20: el recorder de
// WebKitGTK entregaba 0 bytes y el operador solo veía «No se grabó nada», sin causa.
export const CortarDisparaElFlujo: Story = {
  decorators: [escenario({ etapa: "escuchando", transcurrido: 5_000 })],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Cortar el dictado" }))
    // Salió de escuchando: el botón volvió a ser «Dictar», no quedó colgado en ■.
    await expect(c.getByRole("button", { name: "Dictar" })).toBeInTheDocument()
    // Y el motivo está en la superficie, nombrando la etapa (RF-227/228).
    await expect(c.getByText(/no entregó ni un bloque de audio/)).toBeInTheDocument()
  },
}
