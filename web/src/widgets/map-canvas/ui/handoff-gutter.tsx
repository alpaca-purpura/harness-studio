import {
  ArtefactoChip,
  type ChipArtefacto,
  type GutterPlan,
  type RefEntrada,
} from "@/entities/arnes"

// HandoffGutter — el gutter entre carriles (D2 firmada, RF-141/144): los chips viven
// donde el hand-off OCURRE, a la altura del spine. Contiene: chips visibles (plan con
// tope D11c) · referencias ↖ del panel de entrada de la caja seleccionada (D11b — índice
// navegable, JAMÁS un segundo chip) · botón «+N más / − plegar». Widget tonto: el plan
// llega derivado de entities (canvas ⊥ chrome); pinta y reporta clicks.

interface HandoffGutterProps {
  plan: GutterPlan
  refs: RefEntrada[]
  expanded: boolean
  onToggleExpand: () => void
  // Chip click → su productor (RF-145); ref click → el productor del insumo (D11b).
  onSelect?: ((id: string) => void) | undefined
  dimmed?: ((id: string) => boolean) | undefined
}

export function HandoffGutter({
  plan,
  refs,
  expanded,
  onToggleExpand,
  onSelect,
  dimmed,
}: HandoffGutterProps) {
  const showMore = plan.ocultos > 0 || (expanded && plan.expandible)
  if (plan.visibles.length === 0 && refs.length === 0 && !showMore) return null
  return (
    <div className="gutter">
      {plan.visibles.map((c: ChipArtefacto) => (
        <ArtefactoChip key={c.id} chip={c} dim={dimmed?.(c.id)} onSelect={onSelect} />
      ))}
      {refs.map((r) => (
        <button
          key={`${r.productor}::${r.art}`}
          type="button"
          className="ref on"
          title={`entrada de largo alcance: la produce ${r.productor} — click para ir`}
          onClick={() => onSelect?.(r.productor)}
        >
          <span className="r-arr">↖</span>
          <span>{r.art}</span>
          {r.opcional && <span className="r-opc">opcional</span>}
        </button>
      ))}
      {showMore && (
        <button type="button" className="gutter-more on" onClick={onToggleExpand}>
          {expanded ? "− plegar" : `+${plan.ocultos} más`}
        </button>
      )}
    </div>
  )
}
