// toggleEnSet — helper puro de UI (no dominio): agrega/saca un valor de un Set inmutable.
// PROMOVIDO desde `widgets/portafolio/ui/portafolio-list.tsx` (donde nació privado) por el
// paquete 2026-07-23-portafolio-agregar-marketplace §8.2: lo necesitan `FiltroDisclosure`
// (shared/ui, que ya no puede verlo dentro del widget) y los filtros del catálogo. Movimiento
// literal, cero cambio de conducta — `shared/lib` ya está en el allowlist de R2.
export function toggleEnSet<T>(set: ReadonlySet<T>, v: T): Set<T> {
  const next = new Set(set)
  if (next.has(v)) next.delete(v)
  else next.add(v)
  return next
}
