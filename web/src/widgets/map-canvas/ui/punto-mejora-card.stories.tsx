import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { PUNTO_B1, PUNTO_B3, PUNTO_P1 } from "@/entities/telemetria"
import {
  MOTIVO_DESCARTE_SIN_CABLEAR,
  MOTIVO_PROPONER_SIN_CABLEAR,
  PuntoMejoraCard,
} from "./punto-mejora-card"

// Story = test (fe-visual-fitness). La tarjeta es el corazón del entregable: RF-246…257.
// Gate a11y en `error`, con `CompletaB1AtencionDark` obligatoria (D21 ítem 3).

const meta = {
  title: "widgets/map-canvas/PuntoMejoraCard",
  component: PuntoMejoraCard,
  parameters: { layout: "padded" },
  decorators: [(Story) => <div style={{ maxWidth: 720 }}>{Story()}</div>],
  args: { punto: PUNTO_B1, onDescartar: fn(), onProponer: fn() },
} satisfies Meta<typeof PuntoMejoraCard>

export default meta
type Story = StoryObj<typeof meta>

// RF-247…254 — **la tarjeta insignia**, completa. El orden de las filas ES el argumento: qué
// pasa → cuánto → cuánto se ahorra → por qué lo creemos → contra qué → qué hacer.
export const CompletaB1Atencion: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(/La caja «escribir el spec» reescribe el cache en cada corrida/),
    ).toBeInTheDocument()
    await expect(
      c.getByText(
        "Gasta USD 0,84 de los 1,92 de la caja (44 %) escribiendo cache que se vence antes de volver a leerse.",
      ),
    ).toBeInTheDocument()
    // Los cuatro `dt` presentes, en orden.
    const dts = [...canvasElement.querySelectorAll("dt")].map((d) => d.textContent)
    await expect(dts).toEqual(["Contrafactual", "Umbral", "Confianza", "Sesgo"])
    // RF-249 — el contrafactual declara su UNIDAD: «0,53» sobre 14 corridas se leyó de dos
    // maneras distintas en la iteración 1 del mockup.
    await expect(c.getByText(/por corrida/)).toBeInTheDocument()
    await expect(
      c.getByText("relectura 61 % > break-even (2−1,25)/(2−0,1) = 39,47 %"),
    ).toBeInTheDocument()
    await expect(c.getByText("exacta · 14 de 14 corridas con atribución")).toBeInTheDocument()
    // El ajuste concreto va en `<code>`: es lo que se copia, no prosa.
    await expect(canvasElement.querySelector(".mej-fix code")).not.toBeNull()
    await expect(canvasElement.querySelector(".mej-fix code")?.textContent).toBe("cache_ttl: 1h")
    // El score se lee SIN hover.
    await expect(c.getByText("score v1")).toBeVisible()
    await expect(c.getByRole("button", { name: "Descartar" })).toBeInTheDocument()
    await expect(c.getByRole("button", { name: "Proponerlo en el chat" })).toBeInTheDocument()
  },
}

// D21 ítem 3 — el espejo en oscuro. Sin esta story el gate a11y no mira los chips de severidad
// ni el fondo teñido en el tema donde `--crit`/`--crit-soft` pasa raspando (4,66:1).
export const CompletaB1AtencionDark: Story = {
  globals: { theme: "dark" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("atención")).toBeVisible()
    await expect(document.documentElement.dataset["theme"]).toBe("dark")
  },
}

// RF-250 · RF-257 — P1 **no se decide por un umbral**, se decide por un patrón de motivos de
// rechazo. Pintarle una desigualdad inventada sería fabricar el rigor que no tiene.
export const CriticaP1SinUmbral: Story = {
  args: { punto: PUNTO_P1 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(/La caja «revisar el build» se rechaza en el gate 3 de cada 4 veces/),
    ).toBeInTheDocument()
    await expect(c.queryByText("Umbral")).toBeNull()
    await expect(c.getByText("Patrón")).toBeInTheDocument()
    await expect(
      c.getByText("Los 3 rechazos citan el mismo motivo: «el veredicto no lista hallazgos»."),
    ).toBeInTheDocument()
    await expect(c.getByText("crítico")).toBeInTheDocument()
  },
}

// RF-252 — el sesgo va EN CONTRA de la recomendación, y su DIRECCIÓN va en negrita: es la
// palabra que decide si el número es un piso o un techo.
export const SesgoSubestima: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/Asume 0 lecturas fuera de la ventana de 7 días/)).toBeInTheDocument()
    await expect(canvasElement.querySelector("dd strong")?.textContent).toContain("subestima")
  },
}

export const SesgoSobreestima: Story = {
  args: { punto: PUNTO_P1 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText(/No descuenta lo que la revisión aporta aunque rechace/),
    ).toBeInTheDocument()
    await expect(canvasElement.querySelector("dd strong")?.textContent).toContain("sobreestima")
  },
}

// RF-252 — **la fila `Sesgo` NUNCA se omite**. Una fila ausente se lee como «no hay sesgo», que
// es una afirmación distinta de «no encontramos ninguno».
export const SinSesgoIdentificado: Story = {
  args: { punto: PUNTO_B3 },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Sesgo")).toBeInTheDocument()
    await expect(
      c.getByText("No se identificó ningún supuesto que sesgue este cálculo."),
    ).toBeInTheDocument()
  },
}

// J-2 — sin este chip, en una instalación mayormente S2 la tarjeta insignia desaparecería sin
// explicación. Tono NEUTRO: es información sobre el alcance, no una señal de problema.
export const S1Only: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const chip = c.getByText("solo con telemetría de ArnesIA")
    await expect(chip).toHaveClass("mej-chip-neutro")
    await expect(chip).not.toHaveClass("sev-warn")
    await expect(chip).not.toHaveClass("sev-crit")
  },
}

// H-4 — «ver el cálculo» nace CERRADO, con su contrato ARIA completo.
export const CalculoCerrado: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const b = c.getByRole("button", { name: "ver el cálculo" })
    await expect(b).toHaveAttribute("aria-expanded", "false")
    const id = b.getAttribute("aria-controls") as string
    const bloque = canvasElement.querySelector(`#${CSS.escape(id)}`) as HTMLElement
    await expect(bloque).not.toBeNull()
    await expect(bloque).toHaveAttribute("hidden")
  },
}

// H-4 — y **no es un modal**: un modal para auditar un número te saca de la tarjeta que
// estabas leyendo. Es un despliegue en línea.
export const CalculoAbierto: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "ver el cálculo" }))
    const b = c.getByRole("button", { name: "ocultar el cálculo" })
    await expect(b).toHaveAttribute("aria-expanded", "true")
    // La fórmula resuelta, la ventana y las corridas contadas — dentro del bloque desplegado,
    // no en cualquier parte de la tarjeta (el contrafactual también nombra las 14 corridas).
    const calc = canvasElement.querySelector(".mej-calc") as HTMLElement
    await expect(calc).not.toHaveAttribute("hidden")
    await expect(calc.textContent).toMatch(/0,75\s*\/\s*1,9/)
    await expect(calc.textContent).toContain("14 corridas")
    await expect(calc.textContent).toContain("7 días")
    await expect(c.queryByRole("dialog")).toBeNull()
  },
}

// design §5.4 · RF-280 — la relación con el canvas se dice con TEXTO. Si dependiera del grosor
// del borde, en escala de grises o con poca visión la relación desaparece.
export const Resaltada: Story = {
  args: { resaltada: true },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("↔ caja seleccionada")).toBeInTheDocument()
  },
}

// RF-256 — descartar avisa con `aria-live` **y dice cómo revertirlo**: un descarte sin vuelta
// atrás es una pérdida de dato disfrazada de limpieza. El payload es solo el id: la cifra no
// viaja en el handler.
export const DescartarLlamaHandler: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Descartar" }))
    await expect(args.onDescartar).toHaveBeenCalledWith("b1-spec-writer")
    await expect(args.onDescartar).toHaveBeenCalledTimes(1)
    const live = canvasElement.querySelector("[aria-live='polite']") as HTMLElement
    // A-1 · el copy firmado prometía una afordancia que no existe; se asserta la que sí.
    await expect(live.textContent).toBe("Descartado. Vuelve a aparecer al recargar la ventana.")
  },
}

// RF-255 · BR-M12 · D17.3 — **proponer NO escribe**. El assert por ausencia de props de
// escritura es el candado: si alguien agrega `onAplicar`, esta story se pone roja.
export const ProponerAbreChatNoEscribe: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await userEvent.click(c.getByRole("button", { name: "Proponerlo en el chat" }))
    await expect(args.onProponer).toHaveBeenCalledTimes(1)
    await expect(args.onProponer).toHaveBeenCalledWith({
      puntoId: "b1-spec-writer",
      textoPropuesto: "Fijar cache_ttl: 1h en esta caja",
    })
    await expect(Object.keys(args)).not.toContain("onAplicar")
    await expect(Object.keys(args)).not.toContain("onEscribir")
  },
}

// RF-255 — el guardrail del chat embebido: el motivo va en `title` **y en texto visible**. Un
// botón muerto sin explicación se lee como un bug del producto.
export const ProponerDeshabilitadoFueraDeAlcance: Story = {
  args: { proponerDeshabilitado: "Este arnés está fuera del alcance del chat embebido." },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const b = c.getByRole("button", { name: "Proponerlo en el chat" })
    await expect(b).toBeDisabled()
    await expect(b).toHaveAttribute("title", "Este arnés está fuera del alcance del chat embebido.")
    await expect(c.getByText(/fuera del alcance del chat embebido/)).toBeVisible()
  },
}

// RF-257 · RF-280 — las dos severidades se distinguen por TEXTO, no por el borde. El ⚠ es
// decorativo y está `aria-hidden`.
export const SeveridadSinColor: Story = {
  render: (args) => (
    <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
      <PuntoMejoraCard {...args} punto={PUNTO_B1} />
      <PuntoMejoraCard {...args} punto={PUNTO_P1} />
    </div>
  ),
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("atención")).toBeInTheDocument()
    await expect(c.getByText("crítico")).toBeInTheDocument()
    const chips = [...canvasElement.querySelectorAll(".mej-sev")].map((e) => e.textContent)
    await expect(new Set(chips).size).toBe(2)
    for (const w of canvasElement.querySelectorAll(".mej-titular span")) {
      await expect(w).toHaveAttribute("aria-hidden", "true")
    }
  },
}

// RF-247 — el titular habla del PRODUCTO, no del motor: ni el id del detector ni la jerga del
// runtime. «B1» no significa nada para quien tiene que decidir si arreglarlo.
export const TitularSinJerga: Story = {
  play: async ({ canvasElement }) => {
    const titular = canvasElement.querySelector(".mej-titular") as HTMLElement
    await expect(titular.textContent).not.toMatch(/\bB1\b/)
    await expect(titular.textContent).not.toMatch(/ephemeral|cache_creation|TTL_/)
  },
}

// 🔴 **A-1 · sin handler, los botones NO fingen funcionar.** La página no puede cablearlos: no
// existe endpoint de descarte y esta superficie no abre el chat. Pasarles un `() => refetch()`
// los hacía parecer vivos —el refetch remontaba la tarjeta y el anuncio quedaba vacío— así que
// ahora nacen deshabilitados y dicen QUÉ falta, que es el patrón `BotoneraStaged` de este repo.
export const AccionesSinCablearSeDeclaran: Story = {
  // `render` en vez de `args`: con `exactOptionalPropertyTypes` no se puede pasar `undefined`
  // explícito sobre un arg tipado por `fn()`. Omitir la prop ES el caso que se prueba.
  render: (args) => <PuntoMejoraCard punto={args.punto} />,
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByRole("button", { name: "Descartar" })).toBeDisabled()
    await expect(c.getByRole("button", { name: "Proponerlo en el chat" })).toBeDisabled()
    // El motivo, en TEXTO visible: un `title` no llega por teclado ni por touch.
    await expect(c.getByText(MOTIVO_DESCARTE_SIN_CABLEAR)).toBeVisible()
    await expect(c.getByText(MOTIVO_PROPONER_SIN_CABLEAR)).toBeVisible()
  },
}
