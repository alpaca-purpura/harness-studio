import { useEffect, useId, useRef, useState } from "react"
import type { Conversacion } from "@/shared"
import { ConversacionFila } from "./conversacion-fila"

// ConversacionesPanel (RF-317…RF-325 · CV-D2/CV-D4 · design.md §3.4/§4.3/§8.2) — el organismo:
// buscador + rótulo de alcance + lista + pie.
//
// Abre EN SITIO y toma el área del transcript: cero `<dialog>`, cero backdrop, cero portal. Es
// el gesto que `NewSessionPicker` ya usa en el rail (TS-D6/D10), y el header, la fila de
// conversación y el composer siguen visibles — nunca se pierde de vista qué sesión es.
//
// Props PURAS: cero transporte. El cableado vive en `ChatDock`.

export type PanelEstado = "cargando" | "error" | "datos"

export interface ConversacionesPanelProps {
  id: string
  estado: PanelEstado
  error?: string | undefined
  conversaciones: Conversacion[]
  /** el total de la SESIÓN — el denominador de «N de M coinciden» (RF-318). */
  total: number
  /** el Frente de la SESIÓN: el rótulo lo NOMBRA, el alcance se dice, no se deduce (CV-D4). */
  frenteSesion: string
  busqueda: string
  /** `streaming`|`await` ⇒ filas deshabilitadas con motivo; el buscador NO (RF-312, E-28). */
  bloqueadoMotivo?: string | undefined
  /** foco inicial: "buscador" si entró por 🔍, "filas" si entró por ▶ (C-10, RF-355). */
  focoInicial: "filas" | "buscador"
  /** referencia para las fechas relativas; inyectable para las stories. */
  ahora?: number | undefined
  onBusqueda: (q: string) => void
  onElegir: (id: string) => void
  onCancelar: () => void
  onReintentar: () => void
}

export function ConversacionesPanel({
  id,
  estado,
  error,
  conversaciones,
  total,
  frenteSesion,
  busqueda,
  bloqueadoMotivo,
  focoInicial,
  ahora,
  onBusqueda,
  onElegir,
  onCancelar,
  onReintentar,
}: ConversacionesPanelProps) {
  const rotuloId = useId()
  const listaRef = useRef<HTMLDivElement>(null)
  const buscadorRef = useRef<HTMLInputElement>(null)
  const [cursor, setCursor] = useState(0)

  // Ordenar acá y no en el daemon: `ultima_interaccion` desc, y las de 0 turnos AL FINAL por
  // `creada_en` desc (RF-320 CA-2). La activa NO se fija arriba — es una más de la lista.
  const filas = [...conversaciones].sort(ordenar)
  const marcada = filas[Math.min(cursor, filas.length - 1)]

  // Foco inicial: dos controles, dos destinos (C-10). Entrar por ▶ y que el cursor caiga en el
  // buscador obligaría a un Tab para hacer lo que se vino a hacer.
  //
  // A-4 — se enfoca cuando el destino EXISTE, no al montar. Al abrir, el store deja el panel en
  // `cargando` con `total` 0, así que el primer frame es el esqueleto: no hay listbox (va detrás
  // del estado) ni input (va detrás de `total > 1`). Los dos refs valían `null`, el `?.focus()`
  // era un no-op silencioso y el efecto de montaje no volvía a correr: el camino de teclado del
  // panel no arrancaba nunca. El `ya` conserva la intención del `biome-ignore` que había acá —
  // enfocar UNA vez y no pelear con el operador en cada render— pero contando desde el foco que
  // de verdad ocurrió, no desde el montaje.
  const ya = useRef(false)
  useEffect(() => {
    if (ya.current) return
    const destino =
      focoInicial === "buscador" ? (buscadorRef.current ?? listaRef.current) : listaRef.current
    if (!destino) return // todavía no hay adónde: se reintenta en el próximo render.
    destino.focus()
    ya.current = true
  })

  // El cursor arranca en la activa: es de donde el operador viene.
  // biome-ignore lint/correctness/useExhaustiveDependencies: sólo al cambiar el conjunto de filas; moverlo en cada render pisaría las flechas.
  useEffect(() => {
    const i = filas.findIndex((c) => c.activa)
    setCursor(i < 0 ? 0 : i)
  }, [conversaciones])

  // A-5 — el cursor tiene que VERSE. El contenedor del listbox es `overflow-y-auto` y el cursor
  // se movía sólo por estado: con la lista más larga que el panel, `aria-activedescendant`
  // avanzaba y la vista no, así que a partir de la primera fila fuera del viewport se navegaba
  // a ciegas. `block: "nearest"` scrollea lo mínimo — no recentra la lista en cada flecha.
  const marcadaId = marcada?.id
  useEffect(() => {
    if (!marcadaId) return
    document.getElementById(`${id}-${marcadaId}`)?.scrollIntoView({ block: "nearest" })
  }, [id, marcadaId])

  const cuerpo = () => {
    if (estado === "cargando") return <Esqueleto />
    if (estado === "error") return <ErrorDeLista motivo={error} onReintentar={onReintentar} />
    if (filas.length === 0) {
      return (
        <p className="m-auto max-w-[90%] p-4 text-center text-xs text-muted-foreground">
          Ninguna conversación de esta sesión menciona{" "}
          <b className="text-foreground">«{busqueda}»</b>.
          <br />
          <span className="text-[10px]">Se buscó en el título y en el texto de las {total}.</span>
          <span className="mt-2 block">
            <BotonFantasma onClick={() => onBusqueda("")}>Limpiar búsqueda</BotonFantasma>
          </span>
        </p>
      )
    }
    return (
      <>
        {/* aria-live: al filtrar, «N de M coinciden» se ANUNCIA (RF-356). */}
        <p
          id={rotuloId}
          aria-live="polite"
          className="px-3 pb-1.5 text-[9.5px] text-muted-foreground"
        >
          {busqueda ? (
            `${filas.length} de ${total} coinciden`
          ) : (
            <>
              {total} {total === 1 ? "conversación" : "conversaciones"} de la sesión{" "}
              <b className="font-semibold text-foreground">«{frenteSesion}»</b>
            </>
          )}
        </p>
        <div
          ref={listaRef}
          tabIndex={0}
          role="listbox"
          aria-label="Conversaciones"
          aria-describedby={rotuloId}
          aria-activedescendant={marcada ? `${id}-${marcada.id}` : undefined}
          onKeyDown={(e) => {
            if (e.key === "ArrowDown" || e.key === "ArrowUp") {
              e.preventDefault()
              const paso = e.key === "ArrowDown" ? 1 : -1
              setCursor((v) => Math.min(Math.max(v + paso, 0), filas.length - 1))
              return
            }
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault()
              if (marcada && !bloqueadoMotivo) onElegir(marcada.id)
            }
          }}
          className="flex flex-col gap-[3px] overflow-y-auto px-2 pb-2 focus:outline-none"
        >
          {filas.map((c) => (
            <ConversacionFila
              key={c.id}
              id={`${id}-${c.id}`}
              conversacion={c}
              deshabilitadaMotivo={bloqueadoMotivo}
              termino={busqueda}
              marcada={marcada?.id === c.id}
              ahora={ahora}
              onElegir={() => onElegir(c.id)}
            />
          ))}
        </div>
        {/* Con UNA sola conversación el vacío explica qué va a pasar cuando haya otra
            (mockup:672-675). No es relleno: es la única superficie donde el operador se
            entera de que ＋ no pierde lo que tiene. */}
        {total === 1 && !busqueda && (
          <p className="px-4 pb-2 text-center text-xs text-muted-foreground">
            Esta sesión recién arranca. Cuando abras otra con <b className="text-foreground">＋</b>,
            esta queda acá y la podés retomar.
          </p>
        )}
      </>
    )
  }

  return (
    <section
      id={id}
      aria-label={`Conversaciones de la sesión «${frenteSesion}»`}
      // A-6 — Escape cierra el panel desde CUALQUIER parte de él, no sólo desde el buscador.
      // El panel abre en sitio y tapa el transcript (C-4: sin `<dialog>`, sin backdrop), así
      // que sin Escape la única salida era volver al mouse o tabular hasta el ▶. Va acá y no
      // en el listbox porque el buscador no siempre se dibuja (`total > 1`) y el error y el
      // vacío tampoco tienen listbox: el contenedor es lo único que está siempre.
      onKeyDown={(e) => {
        if (e.key !== "Escape") return
        e.stopPropagation()
        onCancelar()
      }}
      className="flex min-h-24 flex-1 flex-col overflow-hidden"
    >
      {/* El buscador NO se dibuja con una sola conversación: no hay nada que filtrar
          (RF-325 CA-2). Y sigue habilitado con un turno en vuelo — leer lo que ya se dijo no
          compite con el turno (E-28). */}
      {total > 1 && (
        <div className="relative flex-none px-3 pt-2 pb-1.5">
          <span
            aria-hidden
            className="-translate-y-1/2 pointer-events-none absolute top-1/2 left-5 text-[11px] text-muted-foreground"
          >
            🔍
          </span>
          <input
            ref={buscadorRef}
            type="search"
            value={busqueda}
            aria-label="Buscar en estas conversaciones"
            placeholder="Buscar en estas conversaciones…"
            onChange={(e) => onBusqueda(e.target.value)}
            // Sin `onKeyDown` propio: el Escape lo maneja el contenedor (A-6) y desde acá
            // burbujea hasta él. Duplicarlo llamaría a `onCancelar` dos veces por tecla.
            className="w-full rounded-md border border-border bg-secondary py-1.5 pr-2.5 pl-6 text-xs text-foreground placeholder:text-muted-foreground focus:border-primary focus:outline-none"
          />
        </div>
      )}

      {cuerpo()}

      {/* C-9: el pie lleva SÓLO `Cancelar`. Seleccionar ES la acción (CV-D11) — un «Retomar»
          sería un segundo paso para algo que ya pasó. */}
      <div className="mt-auto flex flex-none justify-end gap-2 border-t border-border border-dashed px-3 py-2">
        <BotonFantasma onClick={onCancelar}>Cancelar</BotonFantasma>
      </div>
    </section>
  )
}

function BotonFantasma({ onClick, children }: { onClick: () => void; children: React.ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className="rounded-md border border-border px-2.5 py-1 text-[10.5px] text-muted-foreground hover:bg-secondary hover:text-foreground focus:outline-2 focus:outline-primary"
    >
      {children}
    </button>
  )
}

// Esqueleto — calca el CONTRATO ARIA de `PickerSkeleton` (`new-session-picker.tsx:309-326`) en
// Tailwind, sin las clases `pf-*`: reusarlas arrastraría el scope CSS del Portafolio a un
// widget que hoy es 100 % utilidades (C-4).
function Esqueleto() {
  return (
    <div
      role="status"
      aria-live="polite"
      aria-label="Cargando conversaciones"
      className="flex flex-col gap-1.5 px-2 py-1"
    >
      {[1, 0.7, 0.4].map((o) => (
        <span
          key={o}
          aria-hidden="true"
          style={{ opacity: o }}
          className="h-8 animate-pulse rounded-md border border-border bg-secondary"
        />
      ))}
    </div>
  )
}

// ErrorDeLista — el motivo TAL CUAL y un reintento. Jamás «0 conversaciones»: no poder
// preguntar y no tener ninguna son cosas distintas (BR-CV-10, RF-347).
//
// C-3: el texto va en `--foreground` (`--warn` da 3,76:1 sobre `--card` y rompe el gate); el
// tono de alarma lo da un borde-izquierdo en `--warn`, que es no textual y sí cumple 3:1.
function ErrorDeLista({
  motivo,
  onReintentar,
}: {
  motivo?: string | undefined
  onReintentar: () => void
}) {
  return (
    <div className="m-auto max-w-[92%] p-3 text-center text-xs">
      <p role="alert" className="border-warn border-l-[3px] pl-2 text-left text-foreground">
        No pude leer las conversaciones: {motivo ?? "motivo desconocido"}
      </p>
      <span className="mt-2 block">
        <BotonFantasma onClick={onReintentar}>Reintentar</BotonFantasma>
      </span>
    </div>
  )
}

// ordenar: última interacción desc; las de 0 turnos al final, por `creada_en` desc
// (RF-320 CA-2). Una conversación sin turnos no tiene «última interacción» que comparar, y
// ponerla arriba con una fecha inventada sería inventarla.
function ordenar(a: Conversacion, b: Conversacion): number {
  const aVacia = a.turnos === 0
  const bVacia = b.turnos === 0
  if (aVacia !== bVacia) return aVacia ? 1 : -1
  if (aVacia) return b.creada_en.localeCompare(a.creada_en)
  return (b.ultima_interaccion ?? "").localeCompare(a.ultima_interaccion ?? "")
}
