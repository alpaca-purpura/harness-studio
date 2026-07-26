import type { Arnes, ArtefactosMode } from "@/entities/arnes"
import { cn } from "@/shared/lib/cn"
import { type Capa, LAYERS } from "../model/layers"

// Los tres estados del toggle Artefactos (RF-143): off = mapa actual idéntico ·
// auto = chips solo al seleccionar una caja · todos = siempre visibles.
const ARTEFACTOS: readonly { id: ArtefactosMode; label: string; title: string }[] = [
  { id: "off", label: "off", title: "Sin artefactos — el Mapa tal como está hoy" },
  { id: "auto", label: "auto", title: "Chips solo al seleccionar una caja" },
  { id: "todos", label: "todos", title: "Chips de hand-off siempre visibles (D2)" },
]

// MapBar is the top bar of the map surface (galaxia `.mapbar`): el arnés visto (etiqueta de solo
// lectura), la META strip (empresa · reporta a · marketplace) y el layer tablist. Es CHROME que el
// shell monta ARRIBA del canvas puro. El layer switcher vive acá (doctrina firmada: "conmutador de
// capas en la barra del mapa").
//
// TS-D21 — el picker de arnés (RF-72) SE SACA por completo, ni siquiera como escape-hatch de
// error: "1 sesión = 1 arnés" no tiene excepción — para ver otro arnés se abre otra sesión (TS-D5),
// también cuando el actual falla al cargar. El arnés visto es SIEMPRE texto de solo lectura, igual
// que el Topbar (TS-D2). Única excepción real: al espiar otro arnés vía "Abrir en Mapa" del
// Portafolio (GAP-1, `viewedId` programático, no un dropdown) se ofrece un link para volver al
// arnés fijo de la sesión — nunca un selector de uso general.
interface MapBarProps {
  arnes?: Arnes | undefined
  capa: Capa
  onCapa: (c: Capa) => void
  activeId?: string | undefined
  // Vuelve al arnés fijo de la sesión tras un peek (TS-D21) — no es un picker general.
  onPick?: ((id: string) => void) | undefined
  // Arnés fijo de la sesión: si difiere de `activeId`, estás en un peek — se ofrece volver.
  ownId?: string | undefined
  // Toggle de la franja Artefactos (RF-143). Optional: sin callback no se pinta.
  artefactos?: ArtefactosMode | undefined
  onArtefactos?: ((m: ArtefactosMode) => void) | undefined
}

function MetaChip({ k, v }: { k: string; v?: string | null | undefined }) {
  if (!v) return null
  return (
    <span className="rounded border border-border bg-secondary px-1.5 py-0.5 font-mono text-xs text-muted-foreground">
      {k} <b className="font-semibold text-foreground">{v}</b>
    </span>
  )
}

export function MapBar({
  arnes,
  capa,
  onCapa,
  activeId,
  onPick,
  ownId,
  artefactos,
  onArtefactos,
}: MapBarProps) {
  const isPeek = Boolean(ownId && activeId && ownId !== activeId)

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2.5">
      <span className="font-mono text-xs text-muted-foreground">
        {arnes?.id ?? activeId ?? "—"}
      </span>
      {isPeek && ownId && (
        <button
          type="button"
          onClick={() => onPick?.(ownId)}
          title="Estás viendo otro arnés en este Mapa — tu sesión sigue en el suyo"
          className="rounded-md border border-border px-2 py-1 font-mono text-[11px] text-muted-foreground hover:bg-secondary"
        >
          vista previa · volver
        </button>
      )}
      <div className="flex flex-wrap gap-1.5">
        <MetaChip
          k="empresa"
          v={arnes?.empresas?.length ? arnes.empresas.join(" · ") : undefined}
        />
        <MetaChip k="reporta a" v={arnes?.reporta_a ?? "—"} />
        <MetaChip k="⬡" v={arnes?.marketplace} />
      </div>
      {onArtefactos && (
        <div
          role="group"
          aria-label="Artefactos"
          className="ml-auto flex items-center gap-0.5 rounded-lg border border-border bg-secondary p-0.5"
        >
          <span className="px-1.5 font-mono text-[10px] uppercase tracking-wide text-muted-foreground">
            artefactos
          </span>
          {ARTEFACTOS.map((m) => (
            <button
              key={m.id}
              type="button"
              aria-pressed={artefactos === m.id}
              title={m.title}
              onClick={() => onArtefactos(m.id)}
              className={cn(
                "rounded-md px-2 py-1 text-xs text-muted-foreground",
                artefactos === m.id && "bg-card font-semibold text-foreground shadow-sm",
              )}
            >
              {m.label}
            </button>
          ))}
        </div>
      )}
      <div
        role="tablist"
        aria-label="Capas del mapa"
        className={cn(
          "flex gap-0.5 rounded-lg border border-border bg-secondary p-0.5",
          !onArtefactos && "ml-auto",
        )}
      >
        {LAYERS.map((l) => (
          <button
            key={l.id}
            type="button"
            role="tab"
            aria-selected={capa === l.id}
            disabled={l.disabled}
            // RF-233 — el motivo sale de `l.motivo`, no de un literal fijo acá. El que estaba
            // hardcodeado («Necesita telemetría (indexer JSONL)») YA NO ERA VERDAD: la señal
            // llega (V1), lo que falta es el diseño.
            title={l.motivo}
            // RF-276 — el motivo NO puede vivir solo en `title`: un lector de pantalla puede no
            // anunciarlo. Va también en un `sr-only` referenciado por `aria-describedby`, con el
            // MISMO texto (la story asserta la igualdad para que no driften).
            aria-describedby={l.motivo ? `capa-motivo-${l.id}` : undefined}
            onClick={() => onCapa(l.id)}
            className={cn(
              "rounded-md px-2.5 py-1 text-xs text-muted-foreground disabled:opacity-40",
              capa === l.id && "bg-card font-semibold text-foreground shadow-sm",
            )}
          >
            {l.label}
          </button>
        ))}
      </div>
      {/* Los motivos viven FUERA del tablist a propósito, por dos razones duras:
          · el nombre accesible de la tab tiene que seguir siendo su etiqueta y nada más
            (la story firmada `Default` busca la tab por `name: "Desempeño"`);
          · `role="tablist"` exige que sus hijos sean `tab` (axe `aria-required-children`).
          `aria-describedby` cruza el documento sin problema. */}
      {LAYERS.filter((l) => l.motivo).map((l) => (
        <span key={l.id} id={`capa-motivo-${l.id}`} className="sr-only">
          {l.motivo}
        </span>
      ))}
    </div>
  )
}
