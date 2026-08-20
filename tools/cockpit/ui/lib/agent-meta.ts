/**
 * Metadata visual por agente para pintar las cards del board de un vistazo:
 * franja de color + emoji + nombre.
 *
 * Genérico (F-4): el cockpit NO hardcodea rosters de producto. La metadata se
 * deriva del `agent_roster` del seam (`project.config.yaml`) cuando hay acceso
 * server-side; este módulo es CLIENT-SAFE y provee el fallback determinístico:
 *   - color: hash estable del slug → paleta fija (contraste OK sobre dark theme)
 *   - emoji: por rol (builder/auditor/supervisor) o genérico
 *   - name:  slug capitalizado
 *
 * Si un componente recibe roster real (via props/API), puede pasarlo como
 * override a `agentMetaOf` — el fallback solo aplica cuando no hay seam data.
 */

import type { Story } from './types';

/** Slug libre del agente — lo declara el adopter en su seam/SYSTEM-MAP. */
export type AgentId = string;

export interface AgentMeta {
  emoji: string;
  name: string;
  /** Color hex (contraste OK sobre dark theme) — franja + acentos. */
  color: string;
}

/** Paleta estable (contraste OK sobre dark theme). El hash del slug indexa acá. */
const COLOR_PALETTE = [
  '#10b981', // emerald
  '#a855f7', // purple
  '#3b82f6', // blue
  '#f59e0b', // amber
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#84cc16', // lime
  '#f97316', // orange
  '#8b5cf6', // violet
  '#14b8a6', // teal
  '#eab308', // yellow
  '#ef4444', // red
];

/** Emojis genéricos por rol del roster (builder/auditor/supervisor/otro). */
const ROLE_EMOJI: Record<string, string> = {
  builder: '🛠',
  auditor: '🔍',
  supervisor: '🧭',
  config: '⚙',
  infra: '🔧',
};

/** Hash determinístico (djb2) → índice estable en la paleta. */
function hashSlug(slug: string): number {
  let h = 5381;
  for (let i = 0; i < slug.length; i++) {
    h = (h * 33) ^ slug.charCodeAt(i);
  }
  return Math.abs(h);
}

/** Color determinístico por slug (mismo slug → mismo color, sin roster). */
export function colorOf(slug: string): string {
  return COLOR_PALETTE[hashSlug(slug.toLowerCase()) % COLOR_PALETTE.length];
}

function capitalize(slug: string): string {
  return slug.length > 0 ? slug[0].toUpperCase() + slug.slice(1) : slug;
}

/**
 * Metadata visual de un agente. Fallback determinístico client-safe; si el
 * caller tiene el `agent_roster` del seam puede pasar `override` (emoji/name/
 * color reales) y este helper solo completa lo que falte.
 */
export function agentMetaOf(
  slug: string,
  override?: Partial<AgentMeta> & { role?: string }
): AgentMeta {
  const lower = slug.toLowerCase();
  return {
    emoji: override?.emoji ?? ROLE_EMOJI[override?.role ?? lower] ?? '🤖',
    name: override?.name ?? capitalize(slug),
    color: override?.color ?? colorOf(slug),
  };
}

function asAgent(v: unknown): AgentId | null {
  return typeof v === 'string' && /^[a-z0-9][a-z0-9_-]*$/i.test(v) ? v : null;
}

/**
 * Infiere el agente dueño de una story, en orden de confianza:
 * 1. `agent_owner` frontmatter · 2. `owner` · 3. prefijo de `cap_target` (`{agent}.{area}`)
 * (el viejo fallback por regex sobre `story_id` requería roster hardcodeado — retirado F-4).
 */
export function agentOf(
  story: Partial<Story> & { agent_owner?: string | null }
): AgentId | null {
  return (
    asAgent(story.agent_owner) ??
    asAgent(story.owner) ??
    asAgent(story.cap_target?.includes('.') ? story.cap_target.split('.')[0] : null)
  );
}

/** Hue determinístico para teñir el badge de release (F0..F8 → tonos distintos estables). */
export function releaseHue(release: string | null | undefined): number | null {
  if (!release) return null;
  const m = release.match(/(\d+)/);
  const n = m ? Number.parseInt(m[1], 10) : 0;
  return (n * 47 + 200) % 360;
}

/** Color de la prioridad (data real: critical/high/medium/low). */
export function priorityColor(priority: string | null | undefined): string | null {
  if (!priority) return null;
  const p = priority.toLowerCase();
  if (p.includes('critical') || p === 'p0') return '#ef4444'; // red
  if (p.includes('high') || p === 'p1') return '#f59e0b'; // amber
  if (p.includes('medium') || p === 'p2') return '#3b82f6'; // blue
  if (p.includes('low') || p === 'p3') return '#64748b'; // slate
  return '#64748b';
}

/** Ícono + label por naturaleza de la story (data real: ui-story, service-story, design-story…). */
export interface TypeMeta {
  icon: string;
  label: string;
}

const TYPE_META: Record<string, TypeMeta> = {
  agentic: { icon: '🤖', label: 'Agentic' },
  service: { icon: '🔌', label: 'Service' },
  design: { icon: '🎨', label: 'Design' },
  tech: { icon: '🛠', label: 'Tech' },
  func: { icon: '📦', label: 'Func' },
  ui: { icon: '🖥', label: 'UI' },
};

/** Normaliza `type` (ej. "ui-story-followup" → ui) y devuelve su metadata visual. */
export function typeMetaOf(type: string | null | undefined): TypeMeta | null {
  if (!type) return null;
  const t = type.toLowerCase();
  // Orden importa: agentic/service/design antes que el fallback ui.
  for (const key of ['agentic', 'service', 'design', 'tech', 'func', 'ui']) {
    if (t.includes(key)) return TYPE_META[key];
  }
  return null;
}
