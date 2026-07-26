// Dictado por voz en el composer (paquete 2026-07-25-spike-voz-dictado, RF-216..228).
// Dibujo: mockups/arnesia-voz-dictado.html.
//
// Dos piezas, a propósito:
//   · <DictadoButton> — el botón, que entra como UN BOTÓN MÁS a la izquierda del enviar.
//     No cambia el layout de la fila del composer ni toca nada de lo vigente.
//   · <VoiceBar> — la franja de estado, que SOLO existe mientras hay un dictado en curso.
//
// Regla que atraviesa todo el archivo: **ningún estado sin nombre**. El peor modo de falla
// que encontró el spike fue una promesa que nunca resuelve, y un spinner anónimo es
// exactamente lo que la haría invisible (RF-228).

import { useEffect } from "react"
import { AVISO_MS, ETAPA_LABEL, mmss, type Resultado, TOPE_MS, useDictado } from "@/shared"
import { cn } from "@/shared/lib/cn"

// MicIcon — el glifo del micrófono. `tachado` lo cruza con una barra para los estados
// no-disponibles: un ícono idéntico en gris no distingue «apagado» de «prohibido».
function MicIcon({ tachado = false }: { tachado?: boolean }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      className="size-4"
      role="img"
      aria-label={tachado ? "micrófono no disponible" : "micrófono"}
    >
      {/* El nombre accesible va en aria-label del svg y en el del botón; el estado REAL —por
          qué no se puede dictar— se escribe bajo el composer, no en un tooltip. */}
      <rect x="9" y="2" width="6" height="11" rx="3" />
      <path d="M5 11a7 7 0 0 0 14 0M12 18v4" />
      {tachado && <path d="M3 3l18 18" />}
    </svg>
  )
}

// Nivel es el medidor de entrada del micrófono.
//
// No es decoración: sin él, un micrófono MUDO se ve exactamente igual que uno vivo, y el
// operador se entera de que no grabó nada recién cuando vuelve el transcripto vacío.
function Nivel({ nivel }: { nivel: number }) {
  // Barras nombradas, no índices: la lista es fija y cada barra es una identidad estable.
  const barras = { izq: 0.35, medioIzq: 0.7, centro: 1, medioDer: 0.6, der: 0.4 }
  return (
    <span className="flex h-3 items-end gap-px" aria-hidden>
      {Object.entries(barras).map(([nombre, peso]) => (
        <span
          key={nombre}
          className="w-[2.5px] rounded-[1px] bg-destructive transition-[height] duration-75"
          style={{ height: `${Math.max(12, nivel * peso * 100)}%` }}
        />
      ))}
    </span>
  )
}

// DictadoButton is the mic toggle. `onTexto` recibe el resultado listo para el composer;
// este componente NUNCA envía el turno (RF-226: el operador revisa y aprieta enviar).
export function DictadoButton({
  sesionId,
  onTexto,
  disabled = false,
}: {
  sesionId: string
  onTexto: (r: Resultado) => void
  disabled?: boolean
}) {
  const etapa = useDictado((s) => s.etapa)
  const disp = useDictado((s) => s.disponibilidad)
  const consultar = useDictado((s) => s.consultarDisponibilidad)
  const empezar = useDictado((s) => s.empezar)
  const cortar = useDictado((s) => s.cortar)

  // Se pregunta al montar: saber que falta el motor DESPUÉS de hablar tres minutos sería
  // el peor momento posible para enterarse (RF-223).
  useEffect(() => {
    if (disp === undefined) void consultar()
  }, [disp, consultar])

  const grabando = etapa === "escuchando"
  const trabajando = etapa === "transcribiendo" || etapa === "ordenando"
  // `disp === undefined` = todavía no sabemos: no se ofrece un botón que capaz no anda.
  const noDisponible = disp === undefined || !disp.disponible
  const off = disabled || trabajando || noDisponible

  if (grabando) {
    return (
      <button
        type="button"
        onClick={cortar}
        title="Cortar el dictado"
        aria-label="Cortar el dictado"
        className="grid size-9 flex-none place-items-center rounded-lg bg-destructive text-xs text-destructive-foreground"
      >
        ■
      </button>
    )
  }
  return (
    <button
      type="button"
      onClick={() => void empezar(sesionId, onTexto)}
      disabled={off}
      // El title repite el motivo, pero NO es donde vive: el motivo se pinta bajo el
      // composer (<DictadoAviso>) porque un botón gris sin explicación es un gap escondido.
      title={disp?.motivo ?? "Dictar (click para empezar, click para cortar)"}
      aria-label="Dictar"
      className="grid size-9 flex-none place-items-center rounded-lg border border-border bg-secondary text-muted-foreground hover:text-foreground disabled:opacity-40 disabled:hover:text-muted-foreground"
    >
      <MicIcon tachado={noDisponible} />
    </button>
  )
}

// VoiceBar es la franja de estado del dictado. Devuelve null en reposo: la franja aparece
// SOLO mientras hay algo en curso, para no comerle alto permanente al dock.
export function VoiceBar() {
  const etapa = useDictado((s) => s.etapa)
  const transcurrido = useDictado((s) => s.transcurrido)
  const nivel = useDictado((s) => s.nivel)
  const cancelar = useDictado((s) => s.cancelar)
  const usarCrudo = useDictado((s) => s.usarCrudo)

  if (etapa === "inactivo") return null

  const grabando = etapa === "escuchando"
  const cerca = grabando && TOPE_MS - transcurrido <= AVISO_MS

  return (
    <div
      className={cn(
        "flex flex-none items-center gap-2 border-t border-border px-3 py-1.5 text-[10.5px]",
        grabando ? "bg-crit-soft" : "bg-accent-soft",
        cerca && "bg-warn-soft",
      )}
      // El cambio de etapa se anuncia: quien no mira la franja igual se entera.
      aria-live="polite"
    >
      {grabando && (
        <span
          className="size-[7px] flex-none animate-pulse rounded-full bg-destructive"
          aria-hidden
        />
      )}
      <span className="font-semibold">{ETAPA_LABEL[etapa]}</span>
      {grabando && <Nivel nivel={nivel} />}

      <span className="ml-auto flex items-center gap-1.5">
        {grabando && (
          <span
            className={cn(
              "font-mono text-[10px] text-muted-foreground",
              cerca && "font-bold text-warn",
            )}
          >
            {mmss(transcurrido)} / {mmss(TOPE_MS)}
          </span>
        )}
        {etapa === "ordenando" ? (
          // El escape a crudo de V-D4, EN VIVO: a los 15 s el operador ya sabe si quiere
          // seguir esperando el paso que cuesta el 87 % del tiempo.
          <button
            type="button"
            onClick={usarCrudo}
            className="text-muted-foreground underline hover:text-foreground"
          >
            usar el crudo
          </button>
        ) : (
          // Cancelar ≠ cortar (RF-219): cortar sigue el flujo, cancelar tira todo.
          <button
            type="button"
            onClick={cancelar}
            className="text-muted-foreground underline hover:text-foreground"
          >
            cancelar
          </button>
        )}
      </span>
    </div>
  )
}

// DictadoAviso es el motivo bajo el composer: por qué no se puede dictar, por qué el texto
// quedó crudo, o qué etapa falló.
//
// Vive en la superficie y no en un `title` a propósito. `crudoMotivo` lo pasa el composer
// porque el crudo NO es un fallo del store —es un resultado válido y marcado— y el store no
// tiene por qué recordar el desenlace de un dictado que ya entregó.
export function DictadoAviso({ crudoMotivo }: { crudoMotivo?: string | undefined }) {
  const fallo = useDictado((s) => s.fallo)
  const cortadoPorTope = useDictado((s) => s.cortadoPorTope)
  const disp = useDictado((s) => s.disponibilidad)
  const etapa = useDictado((s) => s.etapa)

  // Prioridad: lo que acaba de pasar gana sobre la condición de fondo.
  if (fallo) {
    return (
      <Aviso tono="bad" ico="✕">
        Falló <b>{ETAPA_LABEL[fallo.etapa].replace("…", "").toLowerCase()}</b>: {fallo.mensaje} El
        composer no se tocó.
      </Aviso>
    )
  }
  if (cortadoPorTope) {
    return (
      <Aviso tono="warn" ico="▲">
        Se cortó por el tope de {mmss(TOPE_MS)}. <b>Sigue el flujo con lo grabado</b> — no se
        descarta nada.
      </Aviso>
    )
  }
  if (crudoMotivo) {
    return (
      <Aviso tono="warn" ico="▲">
        Quedó <b>sin ordenar</b> — {crudoMotivo}. El dictado no se perdió.
      </Aviso>
    )
  }
  // La condición de fondo solo se muestra en reposo: durante un dictado la franja ya habla.
  if (etapa === "inactivo" && disp && !disp.disponible && disp.motivo) {
    return (
      <Aviso tono="plain" ico="✕">
        {disp.motivo}
        {disp.instalar?.length ? (
          <>
            {" "}
            Instalá{" "}
            {disp.instalar.map((b, i) => (
              <span key={b}>
                {i > 0 && " o "}
                <code className="rounded-sm bg-secondary px-1 font-mono text-[10px]">{b}</code>
              </span>
            ))}{" "}
            y reabrí la app.
          </>
        ) : null}
      </Aviso>
    )
  }
  return null
}

function Aviso({
  tono,
  ico,
  children,
}: {
  tono: "bad" | "warn" | "plain"
  ico: string
  children: React.ReactNode
}) {
  return (
    <div
      className={cn(
        "flex flex-none items-start gap-1.5 px-3 pb-2 text-[10.5px]",
        tono === "bad" && "text-crit",
        tono === "warn" && "text-warn",
        tono === "plain" && "text-muted-foreground",
      )}
    >
      <span aria-hidden className="flex-none leading-[1.35]">
        {ico}
      </span>
      <span>{children}</span>
    </div>
  )
}
