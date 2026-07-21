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

// MapBar is the top bar of the map surface (galaxia `.mapbar`): the META strip (empresa ·
// reporta a · marketplace) and the layer tablist. It is CHROME the shell mounts ABOVE the pure
// canvas. The layer switcher lives here (signed doctrine: "conmutador de capas en la barra del
// mapa").
//
// TS-D20 — el picker de arnés (RF-72) deja de ser un `<select>` de uso libre: "1 sesión = 1
// arnés" tiene que ser cierto en la UI, no solo en el modelo. Solo se vuelve interactivo cuando
// `showPicker` es true — el motivo REAL original de RF-72 ("sigue alcanzable incluso cuando el
// arnés actual falla al cargar"), no un switcher casual. En cualquier otro caso es texto de solo
// lectura, igual que el Topbar (TS-D2). Si estás viendo un arnés distinto al de tu sesión sin
// error (peek de "Abrir en Mapa", GAP-1) se lo indica explícito con un link para volver — nunca un
// dropdown genérico.
interface MapBarProps {
  arnes?: Arnes | undefined
  capa: Capa
  onCapa: (c: Capa) => void
  // Arnés picker (RF-72), SOLO interactivo en el escape-hatch real (arnés no cargó / no indexado).
  harnesses?: readonly { id: string; label: string }[] | undefined
  activeId?: string | undefined
  onPick?: ((id: string) => void) | undefined
  showPicker?: boolean | undefined
  // Arnés fijo de la sesión (TS-D20): si difiere de `activeId` sin `showPicker`, estás en un
  // peek — se ofrece volver.
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
  harnesses,
  activeId,
  onPick,
  showPicker,
  ownId,
  artefactos,
  onArtefactos,
}: MapBarProps) {
  // Keep the current arnés selectable even when it is not in the index (a session may point at an
  // arnés the index has not seeded) — otherwise the <select> value would not match any option.
  const inList = !activeId || (harnesses?.some((h) => h.id === activeId) ?? false)
  const options =
    harnesses && activeId && !inList
      ? [{ id: activeId, label: `${activeId} · no indexado` }, ...harnesses]
      : harnesses
  const isPeek = Boolean(!showPicker && ownId && activeId && ownId !== activeId)

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-border bg-card px-4 py-2.5">
      {showPicker && options && options.length > 0 ? (
        <label
          className="flex items-center gap-1.5 text-[10px] uppercase tracking-wide text-muted-foreground"
          title="Tu arnés no cargó — elegí otro indexado para ver algo mientras lo resolvés"
        >
          ver otro
          <select
            name="arnes-picker"
            aria-label="Elegir otro arnés indexado (el tuyo no cargó)"
            value={activeId ?? ""}
            onChange={(e) => onPick?.(e.target.value)}
            className="rounded-md border border-border bg-secondary px-2 py-1 font-mono text-xs normal-case tracking-normal text-foreground"
          >
            {options.map((h) => (
              <option key={h.id} value={h.id}>
                {h.label}
              </option>
            ))}
          </select>
        </label>
      ) : (
        <span className="font-mono text-xs text-muted-foreground">
          {arnes?.id ?? activeId ?? "—"}
        </span>
      )}
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
            title={l.disabled ? "Necesita telemetría (indexer JSONL)" : undefined}
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
    </div>
  )
}
