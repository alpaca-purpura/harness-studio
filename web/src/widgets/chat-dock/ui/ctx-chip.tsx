import { cn } from "@/shared/lib/cn"

// CtxChip + IdentidadDetalle (RF-327/RF-328/RF-330 · CV-D14 · design.md §3.3/§4.2).
//
// El ctx deja de ser una fila fija y pasa a ser un chip-disclosure: la única cifra accionable
// queda a la vista (dispara la rotación) y la identidad técnica —cc-id · arnés · modelo ·
// cwd— vive detrás de un clic. Nada se pierde: los 4 datos de la `SessionLine` vigente siguen
// estando, más el `cwd`, que hoy no se ve en ninguna superficie y es el confinamiento real
// del conductor (BR-CV-11, superset estricto).
//
// Props PURAS: cero transporte, cero stores. El cableado vive en `ChatDock`.

export interface CtxChipProps {
  /** 0 cuando el hilo recién nace. 0 es DATO, no ausencia: el chip no se esconde (BR-CV-9). */
  ctxPct: number
  /** ausente ⇒ «sin sesión CC» — literal vigente (`chat-dock.tsx:82`). */
  claudeSessionId?: string | undefined
  arnes: string
  model?: string | undefined
  cwd?: string | undefined
  /**
   * variante caliente: la BARRA en `--warn`, el número NO (C-3).
   *
   * Viene por prop y no se calcula: el umbral vive en el daemon (`SetUmbralRotacion`,
   * default 40) y el FE no lo conoce. Se deriva de `rotacion_pendiente` del wire, jamás de
   * un `>= 40` tecleado acá — dos umbrales serían dos umbrales que se separan.
   */
  caliente: boolean
  abierto: boolean
  onToggle: () => void
  /** el `id` del detalle que este chip controla (`aria-controls`). */
  detalleId: string
}

export function CtxChip({ ctxPct, caliente, abierto, onToggle, detalleId }: CtxChipProps) {
  return (
    <button
      type="button"
      onClick={onToggle}
      aria-expanded={abierto}
      aria-controls={detalleId}
      // `data-caliente` es el selector de la realización de C-3 (design.md §5.3) y el
      // asidero del único assert de estilo del plan: que el NÚMERO no se tiña de `--warn`.
      data-caliente={caliente || undefined}
      // El `%` va EN EL NOMBRE del botón, no sólo en la barra: la barra es `aria-hidden`
      // (design.md §8.1) y sin esto la cifra no existiría para un lector de pantalla.
      aria-label={`contexto ${ctxPct}%, ver identidad de la conversación`}
      title={`ctx ${ctxPct} % — ver identidad de la conversación`}
      className={cn(
        "flex flex-none items-center gap-1.5 rounded-full border px-1.5 py-0.5 font-mono text-[9.5px]",
        "hover:border-border hover:bg-secondary hover:text-foreground",
        "focus:outline-2 focus:outline-primary",
        // C-3 realizado: el NÚMERO nunca en `--warn` (3,76:1 sobre `--card`, bajo el 4,5 de
        // texto). Caliente cambia el borde y el fondo, que son señales no textuales.
        caliente
          ? "border-warn bg-warn-soft text-foreground"
          : abierto
            ? "border-border bg-secondary text-foreground"
            : "border-transparent text-muted-foreground",
      )}
    >
      <span
        aria-hidden
        className="h-[5px] w-11 flex-none overflow-hidden rounded-full border border-border bg-card"
      >
        <span
          data-barra-ctx
          className={cn("block h-full", caliente ? "bg-warn" : "bg-primary")}
          style={{ width: `${ctxPct}%` }}
        />
      </span>
      {ctxPct}%
    </button>
  )
}

// IdentidadDetalle — el cuerpo desplegado. Es una FILA hermana de la de conversación (como
// en el dibujo), no un popover: el dock no tiene capas flotantes y este paquete no las
// introduce.
export function IdentidadDetalle({
  id,
  claudeSessionId,
  arnes,
  model,
  cwd,
}: {
  id: string
  claudeSessionId?: string | undefined
  arnes: string
  model?: string | undefined
  cwd?: string | undefined
}) {
  return (
    <div
      id={id}
      className="flex flex-none flex-col gap-0.5 border-b border-border bg-secondary px-3 py-1.5 font-mono text-[9.5px] text-muted-foreground"
    >
      <div className="flex flex-wrap items-center gap-1.5">
        {/* N-1 · el `◍ <cc-id>` va en `--foreground`, NO en `--primary`. El vigente lo pinta
            con `text-primary` sobre `bg-secondary`: 2,21:1 medido por axe, contra el 4,5 de
            texto. Es la misma familia que C-3 y se corrige acá, que es adonde esta línea se
            muda (RF-328). El acento sigue existiendo — en el glifo, no en la legibilidad. */}
        <span className="text-foreground">
          ◍ {claudeSessionId ? claudeSessionId.slice(0, 8) : "sin sesión CC"}
        </span>
        <span>·</span>
        <span className="text-foreground">{arnes}</span>
        {model && (
          <>
            <span>·</span>
            <span>{model}</span>
          </>
        )}
      </div>
      {/* Sin cwd la línea NO se dibuja: un «—» de relleno diría «no hay», y lo que pasa es
          que el daemon no lo registró (no-aplica-no-es-cero). */}
      {cwd && (
        <div className="flex min-w-0 items-center gap-1.5">
          <span className="flex-none">cwd</span>
          <span className="truncate" title={cwd}>
            {cwd}
          </span>
        </div>
      )}
    </div>
  )
}
