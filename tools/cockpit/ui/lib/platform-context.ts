/**
 * Pseudo-sistema "platform" (Vía A · HB-27) — CLIENT-SAFE (sin imports de node).
 * ────────────────────────────────────────────────────────────────────────────
 * Platform NO es un sistema real: es un contexto solo-trazabilidad que el cockpit
 * muestra junto a los sistemas, apuntando al `docs/` RAÍZ del workspace (stories +
 * learnings platform-level, owner /pm).
 *
 * Read-only: las stories platform SIGUEN el SDD de producto (10 estados) y se
 * transicionan vía /pm — el cockpit solo las HACE VISIBLES, no las edita.
 *
 * Este módulo vive separado de `workspace.ts` (que usa node:fs/child_process,
 * server-only) para que los componentes cliente puedan importar el slug + las
 * capacidades de vista sin arrastrar APIs de Node al bundle.
 */

export const PLATFORM_SLUG = 'platform';
export const PLATFORM_LABEL = 'Platform · core';

/** Vistas sistema-scoped del cockpit (harness es transversal, no entra acá). */
export type CockpitView = 'board' | 'learnings' | 'roadmap' | 'map' | 'arquitectura' | 'drift' | 'evolucion';

/**
 * Qué vistas aplican a platform. Tiene Board (docs/product/stories), Learnings
 * (docs/learnings) y Roadmap (docs/product/releases) en el root; NO tiene SYSTEM-MAP
 * ni los índices de drift → map/arquitectura/drift NO aplican (el cockpit las corta
 * antes de fetch).
 *
 * `roadmap` estaba en false porque en el workspace de origen platform era un contexto
 * de solo-trazabilidad sin releases propias. En un workspace donde platform ES el
 * proyecto entero (harness-studio: su docs/product vive en la raíz), las releases
 * existen — `docs/product/releases/*.yaml` — y dejarlo en false apagaba una vista con
 * datos reales detrás. Qué tabs se muestran finalmente lo decide `cockpit.config.yaml`;
 * esto solo declara cuáles TIENEN sentido.
 */
const PLATFORM_VIEWS: Record<CockpitView, boolean> = {
  board: true,
  learnings: true,
  roadmap: true,
  map: false,
  arquitectura: false,
  drift: false,
  evolucion: false,
};

export function isPlatform(sistema: string): boolean {
  return sistema === PLATFORM_SLUG;
}

/**
 * ¿La vista aplica para este sistema? Sistemas reales → todo aplica.
 * Platform → según PLATFORM_VIEWS (solo board + learnings).
 */
export function viewAppliesTo(sistema: string, view: CockpitView): boolean {
  if (sistema !== PLATFORM_SLUG) return true;
  return PLATFORM_VIEWS[view];
}
