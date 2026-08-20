/**
 * AccessSection · 🔑 Cómo accedo
 * Extraído de FunctionalityView.tsx (2026-05-28 refactor).
 * Reusado en CapDrawer.
 */

import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import type { Capability, AccessEntryPoint } from '@/lib/types';

export function AccessSection({ access }: { access: NonNullable<Capability['access']> }) {
  const eps = access.entry_points ?? [];
  return (
    <section>
      <h3 className="text-sm font-semibold mb-2">
        <Tooltip content={TOOLTIPS.access_entry_points} variant="header">
          🔑 Cómo accedo
        </Tooltip>
      </h3>
      {eps.length === 0 ? (
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Sin entry_points declarados.
        </div>
      ) : (
        <div className="space-y-1.5">
          {eps.map((ep, i) => (
            <AccessEntryRow key={i} entry={ep} />
          ))}
        </div>
      )}
      {access.forbidden_roles && access.forbidden_roles.length > 0 && (
        <div className="mt-2 text-[11px] text-red-400">
          ❌{' '}
          <Tooltip content={TOOLTIPS.forbidden_roles}>
            <span>Roles prohibidos</span>
          </Tooltip>
          : {access.forbidden_roles.join(', ')}
        </div>
      )}
    </section>
  );
}

function AccessEntryRow({ entry }: { entry: AccessEntryPoint }) {
  return (
    <div className="text-[11px] text-[var(--color-text)] pl-2 border-l-2 border-[var(--color-accent)]">
      <div className="font-mono">{entry.path}</div>
      {entry.navigation && (
        <div className="text-[10px] text-[var(--color-muted)] italic mt-0.5">
          {entry.navigation}
        </div>
      )}
      {entry.requires_role && entry.requires_role.length > 0 && (
        <div className="text-[10px] mt-0.5 flex items-center gap-1 flex-wrap">
          <Tooltip content={TOOLTIPS.requires_role}>
            <span>Roles requeridos</span>
          </Tooltip>
          :{' '}
          {entry.requires_role.map((r) => (
            <Pill
              key={r}
              className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[9px] py-0"
            >
              {r}
            </Pill>
          ))}
        </div>
      )}
      {entry.requires_clinic_scope && (
        <div className="text-[9px] text-yellow-400 mt-0.5">
          ⚠{' '}
          <Tooltip content={TOOLTIPS.requires_clinic_scope}>
            <span>requiere clinic_scope (HIPAA dual filter)</span>
          </Tooltip>
        </div>
      )}
      {entry.entry_type && (
        <div className="text-[9px] text-[var(--color-muted)] mt-0.5 font-mono">
          type: {entry.entry_type}
        </div>
      )}
    </div>
  );
}
