import { useState } from "react"
import type { Session } from "@/shared"
import { GLOBAL_VIEWS, selectAttention, useAppStore, useSessions, VIEWS } from "@/shared"
import { cn } from "@/shared/lib/cn"
import { HealthDot, Pip } from "@/shared/ui/indicators"
import { usePortafolioPicker } from "../model/portafolio-picker-store"
import { type ArnesElegido, NewSessionPicker } from "./new-session-picker"

const viewGlyph = (v: string) => VIEWS.find((x) => x[0] === v)?.[1] ?? "•"

// SessionRail is the left rail (mockup it.14): the parallel work-fronts as WARP-style
// tabs, collapsible to a gutter. Session = a live Claude Code conversation N:1 with an
// arnés. Persists across restarts.
export function SessionRail() {
  const sessions = useSessions((s) => s.sessions)
  const activeId = useSessions((s) => s.activeId)
  const collapsed = useSessions((s) => s.railCollapsed)
  const attention = useSessions(selectAttention)
  const switchTo = useSessions((s) => s.switchTo)
  const toggleRail = useSessions((s) => s.toggleRail)
  const create = useSessions((s) => s.create)
  const setGlobal = useAppStore((s) => s.setView)
  const theme = useAppStore((s) => s.theme)
  const toggleTheme = useAppStore((s) => s.toggleTheme)

  // Picker de nueva sesión (TS-D6/D10): el propio <aside> crece a 360px y esconde lista+pie —
  // 100% inline, cero popover/backdrop. El fetch vive en un store propio del widget (TS-D17);
  // el picker es props-puras. `showCollapsed` sólo colapsa cuando el picker NO está abierto (abrir
  // el picker con el rail colapsado lo expande para dar contexto).
  const [pickerOpen, setPickerOpen] = useState(false)
  const pickerEstado = usePortafolioPicker((s) => s.estado)
  const pickerError = usePortafolioPicker((s) => s.error)
  const pickerEntradas = usePortafolioPicker((s) => s.entradas)
  const showCollapsed = collapsed && !pickerOpen

  const arnCount = (arnes: string) => sessions.filter((s) => s.arnes === arnes).length

  // RF-15/TS-D14: refetch en CADA apertura — nunca una lista cacheada de una apertura anterior.
  const abrirPicker = () => {
    void usePortafolioPicker.getState().cargar()
    setPickerOpen(true)
  }
  // RF-12/RF-17: cerrar descarta búsqueda/selección y restaura el ancho — cero efectos.
  const cerrarPicker = () => {
    setPickerOpen(false)
    usePortafolioPicker.getState().reset()
  }
  // RF-16: crea con datos REALES de la EntradaPortafolio+copia elegidas (arnes/empresa/path);
  // sin `puesto` (el Portafolio no tiene ese concepto). salud/view ciertos para una sesión naciente.
  const crearSesion = (elegido: ArnesElegido) => {
    cerrarPicker()
    setGlobal("") // dejar cualquier vista global para que el canvas+dock de la sesión nueva salgan.
    void create({ ...elegido, salud: "info", view: "Mapa" })
  }
  // RF-13/TS-D11: portafolio vacío → derivar a la vista Portafolio (el picker no tiene wizard).
  const irPortafolio = () => {
    cerrarPicker()
    setGlobal("portafolio")
  }

  return (
    <aside
      className={cn(
        "flex flex-none flex-col border-r border-border bg-card transition-[width] duration-150",
        pickerOpen ? "w-[360px]" : collapsed ? "w-[52px]" : "w-[224px]",
      )}
    >
      {/* header */}
      <div
        className={cn(
          "flex items-center gap-2 px-2.5 pb-2 pt-2.5",
          showCollapsed && "flex-col gap-1.5 px-0",
        )}
      >
        <div className="grid size-7 flex-none place-items-center rounded-lg bg-primary text-sm font-extrabold text-primary-foreground">
          A
        </div>
        {!showCollapsed && (
          <span className="text-[13px] font-bold tracking-tight text-foreground">ArnesIA</span>
        )}
        {!showCollapsed && attention > 0 && (
          <span className="rounded-full border border-warn bg-warn-soft px-1.5 py-0.5 font-mono text-[9px] font-bold text-warn">
            {attention} ◐
          </span>
        )}
        {!pickerOpen && (
          <button
            type="button"
            onClick={toggleRail}
            title={collapsed ? "Expandir rail" : "Colapsar a gutter"}
            className={cn(
              "rounded-md px-1.5 py-0.5 text-muted-foreground hover:bg-secondary hover:text-foreground",
              !collapsed && "ml-auto",
            )}
          >
            {collapsed ? "»" : "«"}
          </button>
        )}
      </div>

      {pickerOpen ? (
        <NewSessionPicker
          estado={pickerEstado}
          error={pickerError}
          entradas={pickerEntradas}
          onCrear={crearSesion}
          onCancelar={cerrarPicker}
          onReintentar={() => void usePortafolioPicker.getState().cargar()}
          onIrPortafolio={irPortafolio}
        />
      ) : (
        <>
          {/* section label */}
          {!collapsed && (
            <div className="flex items-center gap-1.5 px-3 pb-1 pt-1.5 font-mono text-[9px] uppercase tracking-wider text-muted-foreground">
              Sesiones
              <span className="ml-auto text-[8.5px] normal-case tracking-normal text-ok">
                ⭯ persisten
              </span>
            </div>
          )}

          {/* session cards */}
          <div
            className={cn(
              "flex flex-1 flex-col gap-1 overflow-y-auto p-2",
              collapsed && "items-center p-1",
            )}
          >
            {sessions.map((s) => (
              <SessionCard
                key={s.id}
                session={s}
                active={s.id === activeId}
                collapsed={collapsed}
                multi={arnCount(s.arnes) > 1}
                onClick={() => {
                  switchTo(s.id)
                  setGlobal("")
                }}
              />
            ))}
          </div>

          {/* footer */}
          <div className="flex flex-none flex-col gap-1 border-t border-border p-2">
            <NewSessionButton collapsed={collapsed} onOpen={abrirPicker} />
            <div className={cn("flex gap-0.5", collapsed && "flex-col")}>
              {GLOBAL_VIEWS.map(([key, glyph, label]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => setGlobal(key)}
                  title={label}
                  className="flex flex-1 flex-col items-center gap-0.5 rounded-md py-1.5 text-[8.5px] text-muted-foreground hover:bg-secondary hover:text-foreground"
                >
                  <span className="text-sm">{glyph}</span>
                  {!collapsed && <span>{label}</span>}
                </button>
              ))}
              <button
                type="button"
                onClick={toggleTheme}
                title={theme === "light" ? "Modo oscuro" : "Modo claro"}
                className="flex flex-1 flex-col items-center gap-0.5 rounded-md py-1.5 text-[8.5px] text-muted-foreground hover:bg-secondary hover:text-foreground"
              >
                <span className="text-sm">{theme === "light" ? "☾" : "☀"}</span>
                {!collapsed && <span>{theme === "light" ? "Oscuro" : "Claro"}</span>}
              </button>
            </div>
          </div>
        </>
      )}
    </aside>
  )
}

function SessionCard({
  session: s,
  active,
  collapsed,
  multi,
  onClick,
}: {
  session: Session
  active: boolean
  collapsed: boolean
  multi: boolean
  onClick: () => void
}) {
  const rename = useSessions((st) => st.rename)
  const closeSession = useSessions((st) => st.closeSession)
  const canClose = useSessions((st) => st.sessions.length > 1)
  const [editing, setEditing] = useState(false)
  const [value, setValue] = useState(s.frente)

  const commit = () => {
    setEditing(false)
    if (value.trim() && value.trim() !== s.frente) void rename(s.id, value.trim())
    else setValue(s.frente)
  }

  if (collapsed) {
    return (
      <button
        type="button"
        onClick={onClick}
        title={`${s.arnes} · ${s.frente}`}
        aria-current={active}
        className={cn(
          "flex w-[38px] flex-col items-center gap-1 rounded-[10px] border py-1.5",
          active ? "border-primary bg-accent-soft" : "border-border hover:bg-secondary",
        )}
      >
        <Pip status={s.status} />
        <HealthDot salud={s.salud} />
      </button>
    )
  }

  return (
    <div
      role="button"
      tabIndex={0}
      onClick={onClick}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault()
          onClick()
        }
      }}
      aria-current={active}
      title={`${s.arnes} · ${s.frente}`}
      className={cn(
        "group relative flex cursor-pointer flex-col gap-1 rounded-[9px] border p-2 text-left transition-colors",
        active
          ? "border-primary bg-accent-soft shadow-[inset_3px_0_0_var(--primary)]"
          : "border-border bg-card hover:border-input hover:bg-secondary",
      )}
    >
      {canClose && (
        <button
          type="button"
          title="Cerrar sesión"
          onClick={(e) => {
            e.stopPropagation()
            void closeSession(s.id)
          }}
          className="absolute right-1.5 top-1.5 hidden size-4 items-center justify-center rounded text-muted-foreground hover:bg-border hover:text-foreground group-hover:flex"
        >
          ✕
        </button>
      )}

      <div className="flex items-center gap-1.5">
        <Pip status={s.status} />
        {editing ? (
          <input
            autoFocus
            value={value}
            onClick={(e) => e.stopPropagation()}
            onChange={(e) => setValue(e.target.value)}
            onBlur={commit}
            onKeyDown={(e) => {
              e.stopPropagation()
              if (e.key === "Enter") commit()
              if (e.key === "Escape") {
                setValue(s.frente)
                setEditing(false)
              }
            }}
            className="min-w-0 flex-1 rounded border border-primary bg-card px-1.5 py-px text-xs font-semibold text-foreground"
          />
        ) : (
          <span className="flex-1 truncate text-xs font-semibold">{s.frente}</span>
        )}
        <button
          type="button"
          title="Renombrar frente"
          onClick={(e) => {
            e.stopPropagation()
            setValue(s.frente)
            setEditing(true)
          }}
          className="hidden size-[15px] items-center justify-center rounded text-[10px] text-muted-foreground hover:bg-border hover:text-foreground group-hover:flex"
        >
          ✎
        </button>
      </div>

      <div className="flex items-center gap-1.5 font-mono text-[9.5px] text-muted-foreground">
        <span className="text-primary/90">{s.arnes}</span>
        {s.reparacion && (
          <span
            title="Sesión de reparación: edita una instalación (ley A4) — deriva visible, backport al canónico según causa"
            className="rounded-[5px] border border-border bg-secondary px-1 py-px text-[8.5px]"
          >
            reparación
          </span>
        )}
        {multi && <span title="este arnés tiene 2+ frentes">·2 frentes</span>}
        <HealthDot salud={s.salud} />
      </div>

      <div className="flex items-center gap-1.5 font-mono text-[9px] text-muted-foreground">
        <span className="rounded-[5px] border border-border bg-secondary px-1.5 py-px">
          {viewGlyph(s.view)} {s.view}
        </span>
        {s.status === "await" ? (
          <span className="text-[9.5px] font-bold text-warn">◐ te necesita</span>
        ) : (
          s.parked && <span className="truncate">{s.parked}</span>
        )}
      </div>
    </div>
  )
}

// NewSessionButton — ya NO crea la sesión ni pregunta la ruta por window.prompt (TS-D5/RF-5):
// abre el picker inline, donde el arnés se elige del Portafolio real y la ruta sale de la copia.
function NewSessionButton({ collapsed, onOpen }: { collapsed: boolean; onOpen: () => void }) {
  return (
    <button
      type="button"
      onClick={onOpen}
      title="Nueva sesión — elegí el arnés de tu Portafolio"
      className={cn(
        "flex items-center justify-center gap-1.5 rounded-lg border border-dashed border-input py-2 text-xs font-semibold text-muted-foreground hover:border-primary hover:bg-accent-soft hover:text-primary",
        collapsed && "px-0",
      )}
    >
      ＋{!collapsed && <span>Nueva sesión</span>}
    </button>
  )
}
