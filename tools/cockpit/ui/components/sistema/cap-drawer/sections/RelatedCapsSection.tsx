/**
 * RelatedCapsSection · 🔗 Capabilities relacionadas
 * Pills clickeables que abren OTRO CapDrawer (mismo provider, capRef se reusa).
 */

import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { useDrawer } from '@/components/providers/DrawerProvider';
import type { CapRelated } from '@/lib/types';

interface RelatedCapsSectionProps {
  related: CapRelated;
}

const GROUPS: Array<{ key: keyof CapRelated; label: string; cls: string }> = [
  { key: 'depends_on', label: 'depende de', cls: 'bg-[#1e3a5f] text-[#93c5fd]' },
  { key: 'enables', label: 'habilita', cls: 'bg-[#14532d] text-[#86efac]' },
  { key: 'similar', label: 'similar a', cls: 'bg-[#27272a] text-[#a1a1aa]' },
  { key: 'obsoletes', label: 'obsoleta', cls: 'bg-[#450a0a] text-[#fca5a5]' },
];

export function RelatedCapsSection({ related }: RelatedCapsSectionProps) {
  const hasAny = GROUPS.some((g) => (related[g.key]?.length ?? 0) > 0);

  if (!hasAny) {
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.related_capabilities} variant="header">
            🔗 Capabilities relacionadas
          </Tooltip>
        </h3>
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Sin relaciones declaradas.
        </div>
      </section>
    );
  }

  return (
    <section>
      <h3 className="text-sm font-semibold mb-2">
        <Tooltip content={TOOLTIPS.related_capabilities} variant="header">
          🔗 Capabilities relacionadas
        </Tooltip>
      </h3>
      <div className="space-y-2">
        {GROUPS.map((g) => {
          const items = related[g.key] ?? [];
          if (items.length === 0) return null;
          return (
            <div key={g.key} className="text-[11px]">
              <span className="text-[var(--color-muted)] mr-1.5">{g.label}:</span>
              <span className="inline-flex gap-1 flex-wrap">
                {items.map((capRef) => (
                  <RelatedCapPill key={capRef} capRef={capRef} cls={g.cls} />
                ))}
              </span>
            </div>
          );
        })}
      </div>
    </section>
  );
}

/**
 * Pill clickeable que abre otro cap drawer.
 *
 * Formato esperado: "{module}.{slug}" o "{module}/{slug}".
 * Si no parsea, renderiza pill no-clickeable.
 */
function RelatedCapPill({ capRef, cls }: { capRef: string; cls: string }) {
  const { openCap } = useDrawer();
  const parts = capRef.includes('/') ? capRef.split('/', 2) : capRef.split('.', 2);
  const canOpen = parts.length === 2 && parts[0] && parts[1];

  if (!canOpen) {
    return (
      <Pill className={`${cls} text-[10px] py-0 opacity-60`}>{capRef}</Pill>
    );
  }

  return (
    <button
      type="button"
      onClick={() => openCap(parts[0], parts[1])}
      className="inline-flex"
    >
      <Pill
        className={`${cls} text-[10px] py-0 cursor-pointer hover:opacity-80 transition-opacity`}
      >
        {capRef}
      </Pill>
    </button>
  );
}
