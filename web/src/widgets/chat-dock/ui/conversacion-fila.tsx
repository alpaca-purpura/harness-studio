import type { ReactNode } from "react"
import type { Conversacion } from "@/shared"
import { cn } from "@/shared/lib/cn"

// ConversacionFila (RF-319/RF-322/RF-323 · CV-D13 · design.md §3.5/§4.4) — una fila de la
// lista: título · última interacción · nº de turnos · ctx final. **Cuatro datos y ninguno
// más**: la fila NO es una tarjeta.
//
// Es una OPCIÓN de un listbox y no un `<button>` suelto (C-8): un `<button><span>` no le dice
// a un lector de pantalla «N opciones, una elegida». El elemento es un `<div role="option">` y
// no el `<li>` que esboza `design.md` §8.2 por una razón de lint que es también de fondo:
// `noNoninteractiveElementToInteractiveRole` prohíbe darle un rol interactivo a un `<li>`, y
// `useSemanticElements` está OFF en `biome.json` justamente porque el repo prefiere el rol
// explícito. El contrato ARIA es idéntico; el aspecto, también.
//
// `tabIndex={-1}`: el foco del teclado vive en el listbox y viaja por `aria-activedescendant`
// (design.md §8.2). Cada opción es alcanzable por programa, no por Tab — con N opciones en un
// dock de 300 px, tabular una por una sería peor que las flechas.

export interface ConversacionFilaProps {
  id: string
  conversacion: Conversacion
  /** `aria-disabled` + `title`; nunca un control apagado y mudo (C-1). */
  deshabilitadaMotivo?: string | undefined
  /** para el `<mark>`: el daemon manda el fragmento en texto plano, el FE resalta (RF-322). */
  termino?: string | undefined
  /** la fila que el teclado tiene marcada (`aria-activedescendant` del listbox). */
  marcada?: boolean
  /** instante de referencia de la fecha relativa. Inyectable para que las stories no dependan del reloj. */
  ahora?: number | undefined
  onElegir: () => void
}

export function ConversacionFila({
  id,
  conversacion: c,
  deshabilitadaMotivo,
  termino,
  marcada,
  ahora,
  onElegir,
}: ConversacionFilaProps) {
  const inhabilitada = deshabilitadaMotivo !== undefined
  return (
    // biome-ignore lint/a11y/useKeyWithClickEvents: el teclado lo maneja el listbox padre vía aria-activedescendant (design.md §8.2) — un handler por fila duplicaría el modelo de foco.
    <div
      id={id}
      role="option"
      tabIndex={-1}
      aria-selected={c.activa}
      aria-disabled={inhabilitada || undefined}
      title={deshabilitadaMotivo}
      onClick={() => {
        if (!inhabilitada) onElegir()
      }}
      className={cn(
        "flex w-full cursor-pointer items-start gap-2 rounded-md border px-2 py-1.5 text-left",
        c.activa ? "border-primary bg-accent-soft" : "border-transparent hover:bg-secondary",
        marcada && "outline-2 outline-primary",
        inhabilitada && "cursor-not-allowed opacity-50",
      )}
    >
      {/* El radio relleno es la mitad no-cromática de «esta es la activa»: el borde
          `--primary` mide 2,51:1 sobre `--card` en claro y no llega al 3:1 no textual, así
          que la distinción NUNCA depende sólo del color (RF-356). */}
      <span
        aria-hidden
        className={cn(
          "mt-[3px] grid size-[11px] flex-none place-items-center rounded-full border-[1.5px]",
          c.activa ? "border-primary" : "border-input",
        )}
      >
        {c.activa && <i className="block size-[5px] rounded-full bg-primary" />}
      </span>

      <span className="block min-w-0 flex-1">
        <span className="block truncate text-xs font-semibold">{c.titulo}</span>
        <span className="mt-px block truncate font-mono text-[9.5px] text-muted-foreground">
          {meta(c, ahora)}
        </span>
        {/* Sin fragmento no hay tercera línea: una coincidencia sólo en el título no tiene
            nada que explicar, y un renglón vacío con sangría sería peor (RF-322 CA-3). */}
        {c.fragmento && (
          <span className="mt-0.5 block text-[10px] leading-[1.45] text-muted-foreground">
            {resaltar(c.fragmento, termino)}
          </span>
        )}
      </span>

      {/* «activa» en `--foreground`: en `--primary` mediría 2,20:1 sobre `--accent-soft`. */}
      {c.activa && (
        <span className="mt-0.5 flex-none text-[9px] font-bold tracking-[0.04em] text-foreground uppercase">
          activa
        </span>
      )}
    </div>
  )
}

// meta arma la línea de cuatro datos. Lo que NO hace: inventar. Sin turnos dice «sin turnos
// todavía»; con turnos pero sin fecha (una conversación migrada, H-4) dice «sin fecha». Un
// `—` o una fecha derivada del `creada_en` serían dos formas de mentir sobre lo mismo.
function meta(c: Conversacion, ahora?: number): string {
  const ctx = `ctx ${c.ctx_pct} %`
  if (c.turnos === 0) return `sin turnos todavía · ${ctx}`
  const turnos = `${c.turnos} ${c.turnos === 1 ? "turno" : "turnos"}`
  const cuando = c.ultima_interaccion ? fecha(c.ultima_interaccion, ahora) : "sin fecha"
  return `${cuando} · ${turnos} · ${ctx}`
}

const MESES = ["ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"]

// fecha: relativa cerca, absoluta lejos — los cuatro formatos del dibujo (`hace 4 min` ·
// `ayer 18:02` · `24 jul` · `23 jul`). Tabla de meses propia y no `Intl`: el formato tiene que
// ser el mismo en la máquina del operador y en el runner de CI.
function fecha(iso: string, ahora = Date.now()): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return "sin fecha"
  const mins = Math.floor((ahora - d.getTime()) / 60000)
  if (mins < 1) return "recién"
  if (mins < 60) return `hace ${mins} min`
  const hoy = new Date(ahora)
  const mismoDia = (a: Date, b: Date) => a.toDateString() === b.toDateString()
  if (mismoDia(d, hoy)) return `hace ${Math.floor(mins / 60)} h`
  const ayer = new Date(ahora - 86400000)
  const hhmm = `${String(d.getHours()).padStart(2, "0")}:${String(d.getMinutes()).padStart(2, "0")}`
  if (mismoDia(d, ayer)) return `ayer ${hhmm}`
  const dm = `${d.getDate()} ${MESES[d.getMonth()]}`
  return d.getFullYear() === hoy.getFullYear() ? dm : `${dm} ${d.getFullYear()}`
}

// resaltar envuelve la coincidencia en un `<mark>` REAL con sus clases propias: el elemento
// porque es el que significa «esto coincidió», las clases porque el aspecto no puede depender
// del reset del navegador (design.md §9).
//
// Compara sin acentos ni mayúsculas —el daemon busca así (`normalizar`, session_conversaciones.go)—
// y recorta sobre el texto ORIGINAL. El plegado es 1:1 por punto de código justamente para que
// los índices de los dos lados se correspondan: si cambiara el largo, el recorte saldría movido.
function resaltar(texto: string, termino?: string): ReactNode {
  const t = (termino ?? "").trim()
  if (!t) return texto
  const crudo = Array.from(texto)
  const i = plegar(crudo).indexOf(plegar(Array.from(t)))
  if (i < 0) return texto
  const n = Array.from(t).length
  return (
    <>
      {crudo.slice(0, i).join("")}
      <mark className="rounded-[2px] bg-accent-soft px-0.5 text-foreground">
        {crudo.slice(i, i + n).join("")}
      </mark>
      {crudo.slice(i + n).join("")}
    </>
  )
}

const DIACRITICOS = "ÀÁÂÃÄÅàáâãäåÇçÈÉÊËèéêëÌÍÎÏìíîïÑñÒÓÔÕÖØòóôõöøÙÚÛÜùúûüÝýÿŠšŽžĀāĒēĪīŌōŪū"
const PLANOS = "AAAAAAaaaaaaCcEEEEeeeeIIIIiiiiNnOOOOOOooooooUUUUuuuuYyySsZzAaEeIiOoUu"

function plegar(runas: string[]): string {
  return runas
    .map((r) => {
      const i = DIACRITICOS.indexOf(r)
      return (i < 0 ? r : (PLANOS[i] ?? r)).toLowerCase()
    })
    .join("")
}
