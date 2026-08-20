// Resolución del nav del sidebar (cockpit.config.yaml::nav).
//
// Per-board (I-51): en multi cada board declara su propio `nav:` en la raíz de su
// repo; el Go lo sirve por ?sistema= (ver handlers_navconfig.go). Acá vive la regla
// pura de qué tabs muestra el sidebar, separada del componente para testearla.
// Genérica sobre `{id}` → sin acoplar a lucide/NavItem.

/**
 * Aplica una lista `nav` (si existe): filtra + ordena por ella. IDs desconocidos se
 * ignoran (forward-compat para tabs futuros). null = sin config → set completo.
 */
export function applyNavConfig<T extends { id: string }>(
  items: T[],
  nav: string[] | null
): T[] {
  if (!nav) return items;
  const byId: Record<string, T> = {};
  for (const i of items) byId[i.id] = i;
  return nav.map((id) => byId[id]).filter((i): i is T => Boolean(i));
}

/**
 * Qué tabs muestra el sidebar (I-51 · per-board nav):
 * - sistema-scoped (navItems): el nav per-board los filtra/ordena en todo modo.
 * - transversales (coreItems): en MULTI son always-on — no cuelgan de un board
 *   (Negocio es de la empresa, Harness es global), así que un board no puede tacharlas.
 *   En SINGLE el nav los gobierna igual (caso cliente: un repo cura todos sus tabs).
 */
export function resolveSidebarNav<T extends { id: string }>(
  navItems: T[],
  coreItems: T[],
  navIds: string[] | null,
  mode?: string
): { nav: T[]; core: T[] } {
  return {
    nav: applyNavConfig(navItems, navIds),
    core: mode === 'multi' ? coreItems : applyNavConfig(coreItems, navIds),
  };
}
