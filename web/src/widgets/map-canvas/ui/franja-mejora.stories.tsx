import type { Meta, StoryObj } from "@storybook/react-vite"
import { expect, fn, userEvent, within } from "storybook/test"
import { CATALOGO_MEDIDO, COBERTURA_ILUSTRATIVA, RESUMEN_ILUSTRATIVO } from "@/entities/telemetria"
import { FranjaMejora } from "./franja-mejora"

// Story = test (fe-visual-fitness). La franja es dueña de los estados 1, 1b, 2, 3, 3b, 4, 5 y
// del resumen del 7 (design §5.2 · §7.7). Los ids `S <nombre>` de `escenarios.md` mapean acá.
//
// Gate a11y en `error` para TODAS. `ReposoDark` es obligatoria (D21 ítem 3): `preview.ts`
// renderiza en `light` por default, así que sin ella el gate mira **la mitad** de los temas.

const meta = {
  title: "widgets/map-canvas/FranjaMejora",
  component: FranjaMejora,
  parameters: { layout: "fullscreen" },
  args: {
    ventana: "7d",
    resumen: RESUMEN_ILUSTRATIVO,
    retencionDias: 90,
    onVentana: fn(),
    onPolitica: fn(),
    onReintentar: fn(),
  },
} satisfies Meta<typeof FranjaMejora>

export default meta
type Story = StoryObj<typeof meta>

// RF-234…237 — el ORDEN de lectura del dato: la ventana precede al total. Si va después, el
// número se lee antes de saber sobre qué período habla.
export const Reposo: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const sel = c.getByLabelText("ventana")
    await expect(sel).toHaveValue("7d")
    const total = canvasElement.querySelector(".fm-total") as HTMLElement
    await expect(sel.compareDocumentPosition(total) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy()
    await expect(c.getByText("4,82")).toBeInTheDocument()
    await expect(c.getByText("estimado por el runtime, no es facturación")).toBeVisible()
    await expect(c.getByText("cobertura")).toBeVisible()
  },
}

// D21 ítem 3 — el espejo en oscuro. Sin esta story el gate a11y solo ve el tema claro, y los
// dos contrastes que D21 corrigió se miden distinto por tema.
export const ReposoDark: Story = {
  globals: { theme: "dark" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("estimado por el runtime, no es facturación")).toBeVisible()
    await expect(document.documentElement.dataset["theme"]).toBe("dark")
  },
}

// RF-234 — las tres opciones existen y cambiar la ventana avisa a la página (el estado vive
// arriba: `fe-transporte-independiente`).
export const VentanaCambia: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    const opts = c.getAllByRole("option")
    await expect(opts).toHaveLength(3)
    await expect(opts.map((o) => o.textContent)).toEqual(["7 días", "30 días", "todo"])
    await userEvent.selectOptions(c.getByLabelText("ventana"), "30d")
    await expect(args.onVentana).toHaveBeenCalledWith("30d")
  },
}

// RF-235 — el total NUNCA viaja solo: su denominador dice de cuántas corridas salió, cuántos
// turnos se pudieron atribuir, cuántas sesiones y cuántas cajas.
//
// ⚠️ **Desviación del literal de design §7.2, por A-7**: el backend NO cuenta lo mismo en las dos
// mitades — `corridas` es `COUNT(DISTINCT corrida_id)` y la cobertura cuenta TURNOS, con filtros
// distintos (`consultas.go:95` vs `:143`). El copy firmado («de 61 corridas, 58 con atribución»)
// promete que cierran, y en vivo se vio `de 340 corridas, 58 con atribución` al lado de `sobre 61
// corridas`. Se nombran las DOS unidades en vez de fingir la coherencia.
export const TotalConDenominador: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("4,82")).toBeInTheDocument()
    await expect(
      c.getByText("de 61 corridas · 58 de 61 turnos con atribución · 12 sesiones · 4 cajas"),
    ).toBeInTheDocument()
  },
}

// RF-236 — el disclaimer se lee SIN hover. Un `title` no es superficie: se pierde en teclado,
// en touch y en cualquier lector que no lo anuncie.
export const DisclaimerEnSuperficie: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const d = c.getByText("estimado por el runtime, no es facturación")
    await expect(d).toBeVisible()
    await expect(d.tagName).not.toBe("TITLE")
    await expect(d.closest("[title]")).toBeNull()
    await expect(d.closest("[hidden]")).toBeNull()
    await expect(d.closest("details:not([open])")).toBeNull()
  },
}

// RF-237 — la cobertura se integra con sus cuatro niveles y su rótulo visible.
export const CoberturaCuatroNiveles: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.querySelectorAll(".cov-seg")).toHaveLength(4)
    await expect(c.getByText("cobertura")).toBeVisible()
  },
}

// RF-269 · A8 — **nunca un tablero en cero**. El assert de `not.toContain("0,00")` es el que
// impide que alguien «rellene» la franja con ceros para que no se vea vacía.
export const Estado1SinDatos: Story = {
  args: { resumen: null },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Este arnés nunca corrió con telemetría.")).toBeInTheDocument()
    await expect(
      c.getByText("Abrí una sesión desde ArnesIA y la medición arranca sola."),
    ).toBeInTheDocument()
    await expect(c.getByText("— —")).toBeInTheDocument()
    await expect(canvasElement.textContent).not.toContain("0,00")
    await expect(canvasElement.querySelector(".cov-bar")).toBeNull()
  },
}

// H-9 · T-15 — «nunca corrió» ≠ «no corrió en estos 7 días». La primera pide instrumentar; la
// segunda pide ampliar la ventana, y por eso trae el botón que la amplía.
export const Estado1bSinCorridasEnVentana: Story = {
  args: { resumen: null, ultimaCorridaFuera: "12/07" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("Sin corridas en los últimos 7 días. La última fue el 12/07."),
    ).toBeInTheDocument()
    await expect(c.queryByText(/nunca corrió/)).toBeNull()
    await userEvent.click(c.getByRole("button", { name: "Ver todo" }))
    await expect(args.onVentana).toHaveBeenCalledWith("todo")
  },
}

// RF-270 — cobertura parcial: el total dice de cuántas salió. D21: el contado resalta en
// `--foreground` NEGRITA, no en `--warn` (3,76:1 sobre `--card`, deuda ya abierta que este
// paquete no puede agravar).
export const Estado2CoberturaParcial: Story = {
  args: {
    resumen: {
      ...RESUMEN_ILUSTRATIVO,
      costo_reportado_micros: 3_100_000,
      corridas: 5,
      cobertura: {
        esperados: 5,
        exacta: 3,
        por_hash: 0,
        por_proceso: 0,
        sin_dato: 2,
        no_llegaron: 0,
      },
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(canvasElement.textContent).toContain("3,10 — de 5 corridas, 3")
    await expect(c.getByText("2 corridas quedaron sin atribución.")).toBeInTheDocument()
    const n = canvasElement.querySelector(".fm-parcial-n") as HTMLElement
    await expect(n.textContent).toBe("3")
    await expect(getComputedStyle(n).fontWeight).toBe("700")
  },
}

// RF-271 · J-10 — S2 degradado: corrió afuera y SIN instrumentar. El rótulo es un string
// distinto del de 3b, y la story asserta la desigualdad para que nadie los funda en uno.
export const Estado3S2SinInstrumentar: Story = {
  args: { escenario: "s2-degradado" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("corrió fuera de ArnesIA, sin instrumentar")).toBeInTheDocument()
    await expect(c.queryByText("corrió fuera de ArnesIA, instrumentado")).toBeNull()
    await expect(c.getByText(/sin el result del stream-json/)).toBeInTheDocument()
  },
}

// RF-271 · ANEXO H9 — S2 INSTRUMENTADO: el bloque `env` de un `settings.json` enciende OTel, así
// que llega la misma señal que dentro de ArnesIA. **Las cifras de dinero se muestran igual.**
export const Estado3bS2Instrumentado: Story = {
  args: { escenario: "s2-instrumentado" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const rotulo = c.getByText("corrió fuera de ArnesIA, instrumentado")
    await expect(rotulo).toBeInTheDocument()
    await expect(c.getByText("Llega la misma señal que dentro de ArnesIA.")).toBeInTheDocument()
    await expect(c.getByText("USD")).toBeInTheDocument()
    await expect(c.getByText("4,82")).toBeInTheDocument()
    // Assert cruzado: los dos rótulos de S2 son strings DISTINTOS (J-10).
    await expect(rotulo.textContent).not.toBe("corrió fuera de ArnesIA, sin instrumentar")
  },
}

// RF-272 — otro runtime: el número **sí se muestra**, rotulado como calculado con el catálogo.
// Ocultarlo por no venir reportado sería perder la única cifra disponible.
export const Estado4OtroRuntime: Story = {
  args: {
    resumen: {
      ...RESUMEN_ILUSTRATIVO,
      costo_reportado_micros: null,
      costo_calculado_micros: 740_000,
    },
    costoDelCatalogo: true,
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("0,74")).toBeInTheDocument()
    await expect(
      c.getByText("calculado con el catálogo v2026-07-20 — este runtime no reporta costo"),
    ).toBeInTheDocument()
  },
}

// RF-273 — catálogo viejo: se avisa EN LA SUPERFICIE, no en un `title`. Funciona sin internet y
// eso es una decisión, no una falla — pero el usuario tiene que saber que está pasando.
export const Estado5CatalogoViejo: Story = {
  args: { catalogoSinRefrescar: "20/07" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const t = c.getByText("⚠ Precios del release, sin refrescar desde el 20/07.")
    await expect(t).toBeVisible()
    await expect(t.closest("[title]")).toBeNull()
    await expect(
      c.getByText("Funciona sin internet; te avisamos que está funcionando así."),
    ).toBeInTheDocument()
  },
}

// A5 — un runtime que todavía no medimos se DICE, con lo que implica: las corridas viejas no se
// recuperan. Un `0,00` ahí sería afirmar que corrió gratis.
export const RuntimeNoSoportado: Story = {
  args: { resumen: null, runtimeNoSoportado: "codex" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText(/todavía no medimos ese runtime/)).toBeInTheDocument()
    await expect(c.queryByText("0,00")).toBeNull()
  },
}

// RF-275 · H-5 — el resumen de privacidad es lo que se ve SIN abrir nada, y el enlace lleva al
// diálogo. El `90` sale de la config, no de la UI (A-2), y **ya no lleva rótulo**: el número está
// firmado (D26.3) y un «(propuesto)» sobre algo decidido entrena a ignorar los rótulos que sí
// importan. El assert de ausencia es el candado de esa decisión.
export const ResumenQueGuardamos: Story = {
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("Nada de tu cuenta. Nada de la conversación.")).toBeInTheDocument()
    await expect(canvasElement.textContent).toContain("Retención 90 días")
    await expect(c.queryByText("(propuesto)")).toBeNull()
    await userEvent.click(c.getByRole("button", { name: "qué guardamos" }))
    await expect(args.onPolitica).toHaveBeenCalledTimes(1)
  },
}

// H-6 · design §5.2 — cargando: el `select` de ventana SIGUE USABLE. Deshabilitarlo obligaría a
// esperar una consulta para poder pedir otra.
export const Cargando: Story = {
  args: { estado: "cargando" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    const status = c.getByRole("status", { name: "midiendo…" })
    await expect(status).toHaveAttribute("aria-live", "polite")
    await expect(c.getByText("midiendo…")).toBeInTheDocument()
    await expect(c.getByLabelText("ventana")).not.toBeDisabled()
    await expect(canvasElement.querySelector(".cov-bar")).toBeNull()
  },
}

// H-6 — error: la franja **no desaparece**. Desaparecer se leería como «no hay capa», que es
// una conclusión distinta de «no pude preguntar».
export const ErrorDeConsulta: Story = {
  args: { estado: "error", error: "ECONNREFUSED" },
  play: async ({ canvasElement, args }) => {
    const c = within(canvasElement)
    await expect(c.getByText("No se pudo leer la telemetría — ECONNREFUSED.")).toBeInTheDocument()
    await expect(c.getByLabelText("ventana")).toBeInTheDocument()
    const btn = c.getByRole("button", { name: "Reintentar" })
    await expect(btn).toBeEnabled()
    await userEvent.click(btn)
    await expect(args.onReintentar).toHaveBeenCalledTimes(1)
  },
}

// H-6 — daemon caído: las cifras previas SE CONSERVAN, marcadas como posiblemente viejas. Es
// distinguible de «no hay datos», que es otra situación con otra acción.
export const DaemonCaido: Story = {
  args: { estado: "daemon-caido", fechaUltimaMedicion: "25/07" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(
      c.getByText("El daemon no respondió. Lo que ves es la última medición, del 25/07."),
    ).toBeInTheDocument()
    await expect(c.getByText("4,82")).toBeInTheDocument()
    await expect(c.queryByText(/nunca corrió/)).toBeNull()
  },
}

// H-7 · D13 — el reenvío externo encendido es un estado peligroso: se declara, y en reposo el
// chip NO existe (el assert espejo es lo que impide que se pinte «apagado» y se confunda).
export const ReenvioEncendido: Story = {
  args: { forwardDestino: "langfuse.local" },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("reenvío externo encendido → langfuse.local")).toBeInTheDocument()
  },
}

export const ReenvioApagadoNoSeDibuja: Story = {
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.queryByText(/reenvío externo/)).toBeNull()
    await expect(canvasElement.querySelector(".fm-forward")).toBeNull()
  },
}

// H-11 · design §6.1 — a 1024 px la franja ENVUELVE y el rótulo `cobertura` es lo que impide
// que la barra quede huérfana. El `<body>` jamás scrollea horizontal.
export const Angosta1024: Story = {
  decorators: [(Story) => <div style={{ width: 1024 }}>{Story()}</div>],
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("cobertura")).toBeVisible()
    await expect(document.body.scrollWidth).toBeLessThanOrEqual(document.body.clientWidth)
  },
}

// RF-237 — con cobertura completa la franja lo dice en positivo, sin enumerar tres ceros.
export const CoberturaCompletaEnLaFranja: Story = {
  args: {
    resumen: {
      ...RESUMEN_ILUSTRATIVO,
      cobertura: { ...COBERTURA_ILUSTRATIVA, exacta: 61, por_hash: 0, por_proceso: 0, sin_dato: 0 },
      catalogo: CATALOGO_MEDIDO,
    },
  },
  play: async ({ canvasElement }) => {
    const c = within(canvasElement)
    await expect(c.getByText("atribución exacta en las 61 corridas")).toBeInTheDocument()
    await expect(canvasElement.querySelectorAll(".cov-seg")).toHaveLength(1)
  },
}
