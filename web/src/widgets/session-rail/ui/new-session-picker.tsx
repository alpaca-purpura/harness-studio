import { useId, useMemo, useState } from "react"
import {
  DerivaChip,
  DotSaludPortafolio,
  EmblemaInicial,
  type EntradaPortafolio,
  type EstadoDeriva,
  identificadorDe,
  idsColisionados,
  saludDe,
  type TipoInstalacion,
  TipoInstalacionChip,
} from "@/entities/portafolio"
import { cn } from "@/shared/lib/cn"
import { useConversaciones } from "../model/conversaciones-store"
import type { PickerEstado } from "../model/portafolio-picker-store"

// NewSessionPicker — el selector de arnés al abrir sesión (RF-5..17, TS-D6..D16). Props PURAS:
// CERO transporte (el fetch/refetch vive en portafolio-picker-store.ts, cableado por SessionRail
// — mismo patrón que PortafolioList↔PortafolioView). La lista sale del Portafolio REAL del
// operador; cada fila reusa los componentes presentacionales de entities/portafolio (EmblemaInicial
// /DerivaChip/TipoInstalacionChip/DotSaludPortafolio) tal cual — por eso la raíz lleva la clase
// `arnesia-portafolio`: es el scope bajo el que viven los estilos `.pf-*` de esos chips (cero CSS
// nuevo, cero drift contra la Lista del Portafolio).

// ArnesElegido — lo que el picker emite al crear: los datos REALES de la EntradaPortafolio +
// copia elegidas (TS-D16). `empresa` va SOLO cuando hay alguna (join N:M) — nunca un "—" de
// relleno; `puesto` no se escribe (el Portafolio no tiene ese concepto).
export interface ArnesElegido {
  arnes: string
  empresa?: string
  path: string
  // reparacion (RF-191, ley A4): la copia elegida es una INSTALACIÓN → la sesión abre como
  // sesión de reparación (edición legal con deriva visible). Ausente para canónico/única.
  reparacion?: boolean
}

export interface NewSessionPickerProps {
  estado: PickerEstado
  error?: string | undefined
  entradas: EntradaPortafolio[]
  onCrear: (elegido: ArnesElegido) => void
  onCancelar: () => void
  onReintentar: () => void
  onIrPortafolio: () => void
}

// Copia — una ruta seleccionable de una identidad (TS-D15): canónico (0-1) + N instalaciones. El
// `path` confina el cwd del conductor (S2). `tipo` es la etiqueta de la sub-lista.
interface Copia {
  path: string
  esCanonico: boolean
  tipoInstalacion?: TipoInstalacion | ""
  deriva?: EstadoDeriva
}

function copiasDe(e: EntradaPortafolio): Copia[] {
  const cs: Copia[] = []
  if (e.canonico) cs.push({ path: e.canonico.path, esCanonico: true })
  for (const i of e.instalaciones ?? []) {
    cs.push({ path: i.install_path, esCanonico: false, tipoInstalacion: i.tipo, deriva: i.deriva })
  }
  return cs
}

// coincide — buscador en vivo (RF-6): substring case-insensitive sobre id/nombre/empresas. Difiere
// de filtrarEntradas (que cubre id/nombre/descripción) — este flujo busca por EMPRESA, no descr,
// por eso es un matcher local del picker, no el selector de la Lista.
function coincide(e: EntradaPortafolio, q: string): boolean {
  const s = q.trim().toLowerCase()
  if (!s) return true
  const campos = [identificadorDe(e.identidad), e.nombre, ...(e.empresas ?? [])]
  return campos.some((c) => c?.toLowerCase().includes(s))
}

// colisionEtiqueta — home ?? scope truncado, para diferenciar dos filas con el mismo id (RF-9/
// TS-D12). No bloquea la selección (la fila selecciona por `clave`, única) — solo desambigua.
function colisionEtiqueta(e: EntradaPortafolio): string {
  return e.identidad.home ?? e.identidad.scope ?? ""
}

export function NewSessionPicker({
  estado,
  error,
  entradas,
  onCrear,
  onCancelar,
  onReintentar,
  onIrPortafolio,
}: NewSessionPickerProps) {
  const [busqueda, setBusqueda] = useState("")
  const [seleccionada, setSeleccionada] = useState<string | null>(null)
  const [copiaPath, setCopiaPath] = useState<string | null>(null)
  const buscadorId = useId()

  const colisiones = useMemo(() => idsColisionados(entradas), [entradas])
  const filtradas = useMemo(
    () => entradas.filter((e) => coincide(e, busqueda)),
    [entradas, busqueda],
  )

  const elegirFila = (e: EntradaPortafolio) => {
    const copias = copiasDe(e)
    if (copias.length === 0) return // 0 copias: fila deshabilitada, no seleccionable (TS-D15).
    setSeleccionada(e.clave)
    // 1 copia → ruta resuelta ya; 2+ → hay que elegir cuál (sub-lista) antes de habilitar Crear.
    setCopiaPath(copias.length === 1 ? (copias[0]?.path ?? null) : null)
    // Historial B2 (RF-203): las conversaciones pasadas de este arnés, junto al selector.
    // Clave calificada (deuda BACKLOG «re-key», cerrada 2026-07-23) — las sesiones NUEVAS se
    // crean bajo `e.clave` (ver crear()), así que buscar por ahí es lo que las encuentra.
    void useConversaciones.getState().cargar(e.clave)
  }

  const crear = () => {
    if (!seleccionada || !copiaPath) return
    const e = entradas.find((x) => x.clave === seleccionada)
    if (!e) return
    const empresas = e.empresas ?? []
    // RF-191: elegir una instalación (no el canónico) abre la sesión como REPARACIÓN.
    const copia = copiasDe(e).find((c) => c.path === copiaPath)
    onCrear({
      // clave calificada (home,id,scope), NUNCA el id pelado — deuda BACKLOG «re-key»,
      // cerrada 2026-07-23: cierra el círculo Register→loadArnesDir→Upsert→Session.Arnes
      // bajo la MISMA llave sin colisión, dos arneses del mismo id ya no se pisan.
      arnes: e.clave,
      path: copiaPath,
      ...(empresas.length > 0 ? { empresa: empresas.join(" · ") } : {}),
      ...(copia && !copia.esCanonico ? { reparacion: true } : {}),
    })
  }

  return (
    <div className="arnesia-portafolio flex min-h-0 flex-1 flex-col gap-2.5 p-3">
      <div>
        <div className="text-[13px] font-bold text-foreground">Elegí el arnés de esta sesión</div>
        <p className="mt-0.5 text-[11px] text-muted-foreground">
          No se cambia después — otra sesión, para otro arnés.
        </p>
      </div>

      {estado === "cargando" && <PickerSkeleton />}

      {estado === "error" && (
        <div className="flex flex-col items-start gap-2 py-4">
          <p role="alert" className="text-[11.5px] text-muted-foreground">
            No se pudo cargar el portafolio — {error ?? "motivo desconocido"}
          </p>
          <button
            type="button"
            onClick={onReintentar}
            className="rounded-md border border-primary bg-primary px-3 py-1 text-[11.5px] font-semibold text-primary-foreground"
          >
            Reintentar
          </button>
        </div>
      )}

      {estado === "datos" && entradas.length === 0 && (
        <div className="flex flex-col items-center gap-2 py-6 text-center">
          <p className="text-[12px] text-foreground">Tu portafolio está vacío.</p>
          <p className="text-[11px] text-muted-foreground">
            Agregá un arnés desde la vista Portafolio para poder abrir una sesión con él.
          </p>
          <button
            type="button"
            onClick={onIrPortafolio}
            className="rounded-md border border-primary bg-primary px-3 py-1 text-[11.5px] font-semibold text-primary-foreground"
          >
            Ir a Portafolio
          </button>
        </div>
      )}

      {estado === "datos" && entradas.length > 0 && (
        <>
          <div className="relative flex-none">
            <span
              aria-hidden="true"
              className="-translate-y-1/2 pointer-events-none absolute top-1/2 left-2.5 text-[12px] text-muted-foreground"
            >
              🔍
            </span>
            <input
              // Foco inicial al abrir el picker (design §A11y, mockup search.focus()). El rule
              // noAutofocus está OFF en biome.json — uso legítimo, no accidental.
              autoFocus
              id={buscadorId}
              type="search"
              aria-label="Buscar arnés"
              placeholder="Buscar por id, rol o empresa…"
              value={busqueda}
              onChange={(e) => setBusqueda(e.target.value)}
              className="w-full rounded-md border border-input bg-background py-1.5 pr-2.5 pl-7 text-[12px] text-foreground placeholder:text-muted-foreground focus:outline focus:outline-2 focus:outline-primary"
            />
          </div>

          {filtradas.length === 0 ? (
            <div className="flex flex-col items-center gap-1.5 py-5 text-center">
              <p className="text-[11.5px] text-muted-foreground">
                Ningún arnés coincide con la búsqueda.
              </p>
              <button
                type="button"
                onClick={() => setBusqueda("")}
                className="rounded-md border border-border px-3 py-1 text-[11px] text-foreground hover:bg-secondary"
              >
                Limpiar búsqueda
              </button>
            </div>
          ) : (
            <ul className="flex min-h-0 flex-1 list-none flex-col gap-1.5 overflow-y-auto">
              {filtradas.map((e) => (
                <FilaPicker
                  key={e.clave}
                  entrada={e}
                  seleccionada={seleccionada === e.clave}
                  copiaPath={copiaPath}
                  colisiona={colisiones.has(e.identidad.id)}
                  onElegirFila={() => elegirFila(e)}
                  onElegirCopia={setCopiaPath}
                />
              ))}
            </ul>
          )}

          {copiaPath && (
            <p className="flex-none text-[10.5px] text-muted-foreground">
              usará{" "}
              <code className="rounded bg-secondary px-1.5 py-px font-mono text-foreground">
                {copiaPath}
              </code>
            </p>
          )}

          {seleccionada && <ConversacionesDelArnes />}

          <div className="mt-auto flex flex-none justify-end gap-2 pt-1.5">
            <button
              type="button"
              onClick={onCancelar}
              className="rounded-md border border-border px-3 py-1.5 text-[11.5px] text-muted-foreground hover:bg-secondary"
            >
              Cancelar
            </button>
            <button
              type="button"
              onClick={crear}
              disabled={!seleccionada || !copiaPath}
              className="rounded-md border border-primary bg-primary px-3 py-1.5 text-[11.5px] font-bold text-primary-foreground disabled:cursor-not-allowed disabled:opacity-40"
            >
              Crear sesión
            </button>
          </div>
        </>
      )}
    </div>
  )
}

// ConversacionesDelArnes (RF-203, historial B2): vivas + cerradas del arnés seleccionado.
// Una cerrada se expande a sus turnos reconstruidos desde la JSONL nativa; las JSONL ya
// ausentes se DICEN (`faltantes`), jamás se finge un historial completo.
function ConversacionesDelArnes() {
  const { vivas, cerradas, error, historialDe, turnos, faltantes, abrirHistorial } =
    useConversaciones()
  if (error === undefined && vivas.length === 0 && cerradas.length === 0) return null
  return (
    <div className="flex max-h-44 flex-none flex-col gap-1 overflow-y-auto rounded-md border border-border p-2 text-[10px]">
      <span className="font-bold text-muted-foreground">
        Conversaciones: {vivas.length} abierta{vivas.length === 1 ? "" : "s"} · {cerradas.length}{" "}
        cerrada{cerradas.length === 1 ? "" : "s"}
      </span>
      {error !== undefined && (
        // C-3 (design.md, paquete conversaciones-del-panel): `--warn` como color de TEXTO da
        // 3,76:1 sobre `--card` y rompe el gate axe. El motivo va en `--foreground` (18,74:1) y
        // la señal de alarma la da un borde izquierdo — no textual, 3,76 ≥ 3 ✓.
        <span className="border-warn border-l-[3px] pl-1.5 text-foreground">
          historial de cerradas: {error}
        </span>
      )}
      {cerradas.map((c) => (
        <div key={c.id} className="flex flex-col">
          <button
            type="button"
            onClick={() => void abrirHistorial(c.id)}
            className="flex items-center gap-1.5 rounded px-1 py-px text-left font-mono text-muted-foreground hover:bg-secondary"
          >
            <span className="truncate">{c.frente}</span>
            {c.cerrada_en && <span className="flex-none">{c.cerrada_en.slice(0, 10)}</span>}
            {typeof c.turnos === "number" && c.turnos > 0 && (
              <span className="flex-none">· {c.turnos} turnos</span>
            )}
          </button>
          {historialDe === c.id && (
            <div className="ml-2 flex max-h-24 flex-col gap-0.5 overflow-y-auto border-l border-border pl-2">
              {faltantes.length > 0 && (
                // C-3: mismo criterio — texto en `--foreground`, alarma por borde.
                <span className="border-warn border-l-[3px] pl-1.5 text-foreground">
                  {faltantes.length} tramo(s) ya no están en disco — historial parcial
                </span>
              )}
              {turnos.length === 0 && faltantes.length === 0 && <span>cargando…</span>}
              {turnos.map((t, i) => (
                // biome-ignore lint/suspicious/noArrayIndexKey: transcript estático, orden estable.
                <span key={i} className="truncate">
                  <b>[{t.rol}]</b> {t.text}
                </span>
              ))}
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

function PickerSkeleton() {
  return (
    <div
      role="status"
      aria-live="polite"
      aria-label="Cargando portafolio"
      className="flex flex-col gap-1.5 py-1"
    >
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          aria-hidden="true"
          className="h-11 animate-pulse rounded-md border border-border bg-secondary"
        />
      ))}
    </div>
  )
}

function FilaPicker({
  entrada: e,
  seleccionada,
  copiaPath,
  colisiona,
  onElegirFila,
  onElegirCopia,
}: {
  entrada: EntradaPortafolio
  seleccionada: boolean
  copiaPath: string | null
  colisiona: boolean
  onElegirFila: () => void
  onElegirCopia: (path: string) => void
}) {
  const copias = copiasDe(e)
  const sinCopia = copias.length === 0
  const ambigua = copias.length >= 2
  const instalaciones = e.instalaciones ?? []
  const enDeriva = instalaciones.find((i) => i.deriva === "en-deriva")
  const idMostrado = identificadorDe(e.identidad)

  const presencia: string[] = []
  if (e.canonico) presencia.push(`◆ canónico v${e.canonico.version ?? "?"}`)
  if (instalaciones.length > 0) presencia.push(`▣ ${instalaciones.length} instalac.`)

  return (
    <li
      className={cn("rounded-md border bg-card", seleccionada ? "border-primary" : "border-border")}
    >
      <button
        type="button"
        aria-pressed={seleccionada}
        aria-disabled={sinCopia}
        disabled={sinCopia}
        title={sinCopia ? "sin copia local registrada" : undefined}
        onClick={onElegirFila}
        className={cn(
          "flex w-full items-start gap-2.5 rounded-md p-2 text-left",
          sinCopia ? "cursor-not-allowed opacity-50" : "hover:bg-secondary",
          seleccionada && "bg-accent-soft",
        )}
      >
        <span
          aria-hidden="true"
          className={cn(
            "mt-0.5 grid size-[13px] flex-none place-items-center rounded-full border-[1.5px]",
            seleccionada ? "border-primary" : "border-input",
          )}
        >
          {seleccionada && <span className="size-[7px] rounded-full bg-primary" />}
        </span>
        <EmblemaInicial texto={e.identidad.id} />
        <span className="min-w-0 flex-1">
          <span className="flex min-w-0 items-baseline gap-1.5">
            <span className="flex-none whitespace-nowrap font-mono text-[11.5px] font-bold">
              {idMostrado}
            </span>
            {e.nombre && (
              <span className="min-w-0 flex-1 truncate text-[10.5px] text-muted-foreground">
                {e.nombre}
              </span>
            )}
          </span>
          <span className="mt-0.5 flex flex-wrap items-center gap-1">
            {(e.empresas ?? []).map((emp) => (
              <span
                key={emp}
                className="whitespace-nowrap rounded-sm border border-border px-1.5 py-px font-mono text-[9px] text-muted-foreground"
              >
                {emp}
              </span>
            ))}
            {presencia.map((p) => (
              <span
                key={p}
                className="whitespace-nowrap rounded-sm border border-dashed border-border px-1.5 py-px font-mono text-[9px] text-muted-foreground"
              >
                {p}
              </span>
            ))}
            {enDeriva && <DerivaChip estado="en-deriva" detalle={enDeriva.deriva_detalle} />}
            {colisiona && colisionEtiqueta(e) && (
              <span
                title={colisionEtiqueta(e)}
                className="max-w-[110px] truncate rounded-sm border border-input bg-secondary px-1.5 py-px font-mono text-[9px] text-muted-foreground"
              >
                {colisionEtiqueta(e)}
              </span>
            )}
          </span>
        </span>
        <DotSaludPortafolio salud={saludDe(e)} />
      </button>

      {seleccionada && ambigua && (
        // <div role="group">, no <ul> — con role="group" el <ul> pierde su rol de lista y sus
        // <li> quedan huérfanos (axe listitem). Los botones son los items interactivos directos.
        <div
          className="flex flex-col gap-1 border-border border-t border-dashed py-1.5 pr-2.5 pb-2 pl-[41px]"
          role="group"
          aria-label={`Copias de ${idMostrado}`}
        >
          {copias.map((c) => (
            <button
              key={c.path}
              type="button"
              aria-pressed={copiaPath === c.path}
              onClick={() => onElegirCopia(c.path)}
              className={cn(
                "flex w-full items-center gap-2 rounded-sm border p-1.5 text-left text-[10px]",
                copiaPath === c.path
                  ? "border-primary bg-accent-soft"
                  : "border-border hover:bg-secondary",
              )}
            >
              <span
                aria-hidden="true"
                className={cn(
                  "grid size-2.5 flex-none place-items-center rounded-full border-[1.5px]",
                  copiaPath === c.path ? "border-primary" : "border-input",
                )}
              >
                {copiaPath === c.path && <span className="size-[5px] rounded-full bg-primary" />}
              </span>
              {c.esCanonico ? (
                <span className="flex-none font-mono font-bold">canónico</span>
              ) : (
                <>
                  <TipoInstalacionChip tipo={c.tipoInstalacion ?? ""} />
                  <span
                    title="Elegir esta copia abre una sesión de REPARACIÓN: edición legal in situ, con la deriva visible en el Portafolio (ley A4)"
                    className="flex-none text-[9px] text-muted-foreground"
                  >
                    → reparación
                  </span>
                </>
              )}
              <span className="min-w-0 flex-1 truncate font-mono text-muted-foreground">
                {c.path}
              </span>
              {c.deriva === "en-deriva" && <DerivaChip estado="en-deriva" />}
            </button>
          ))}
        </div>
      )}
    </li>
  )
}
