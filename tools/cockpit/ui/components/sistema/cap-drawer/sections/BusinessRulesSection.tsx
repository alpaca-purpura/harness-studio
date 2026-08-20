/**
 * BusinessRulesSection · 📋 Reglas de negocio
 * Extraído de FunctionalityView.tsx (2026-05-28 refactor).
 */

import { cn } from '@/lib/cn';
import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import type { Capability } from '@/lib/types';

const SEVERITY_CLS: Record<string, string> = {
  critical: 'bg-[#450a0a] text-[#fca5a5]',
  high: 'bg-[#713f12] text-[#fbbf24]',
  medium: 'bg-[#1e3a5f] text-[#93c5fd]',
  low: 'bg-[#27272a] text-[#a1a1aa]',
};

export function BusinessRulesSection({
  rules,
}: {
  rules: NonNullable<Capability['business_rules']>;
}) {
  if (rules.length === 0) {
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.business_rules} variant="header">
            📋 Reglas de negocio
          </Tooltip>
        </h3>
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Sin business_rules declaradas todavía.
        </div>
      </section>
    );
  }

  return (
    <section>
      <h3 className="text-sm font-semibold mb-2">
        <Tooltip content={TOOLTIPS.business_rules} variant="header">
          📋 Reglas de negocio ({rules.length})
        </Tooltip>
      </h3>
      <div className="space-y-1.5">
        {rules.map((r) => (
          <div key={r.id} className="text-[11px] pl-2 border-l-2 border-yellow-700">
            <div className="flex items-baseline gap-1.5 flex-wrap">
              <Tooltip content={TOOLTIPS.severity} variant="badge">
                <Pill className={cn(SEVERITY_CLS[r.severity] ?? SEVERITY_CLS['medium'], 'text-[9px] py-0')}>
                  {r.severity}
                </Pill>
              </Tooltip>
              <span>{r.rule}</span>
              {r.audit_trail && (
                <Tooltip content={TOOLTIPS.audit_trail} variant="badge">
                  <Pill className="bg-[#1e3a5f] text-[#93c5fd] text-[9px] py-0">audit</Pill>
                </Tooltip>
              )}
              {/* Badge de verdad N2: ¿la regla está implementada (code_ref) o es papel? */}
              {r.code_ref ? (
                <Tooltip content={`Implementada en: ${r.code_ref}`} variant="badge">
                  <Pill className="bg-[#064e3b] text-[#6ee7b7] text-[9px] py-0">🟢 enforced</Pill>
                </Tooltip>
              ) : (
                <Tooltip
                  content="Regla declarada SIN code_ref → no se ve dónde está implementada (posible regla de papel)."
                  variant="badge"
                >
                  <Pill className="bg-[#450a0a] text-[#fca5a5] text-[9px] py-0">🔴 sin enforcement</Pill>
                </Tooltip>
              )}
            </div>
            {r.enforcement && r.enforcement.length > 0 && (
              <div className="text-[9px] text-[var(--color-muted)] mt-0.5 font-mono">
                ↗{' '}
                <Tooltip content={TOOLTIPS.enforcement}>
                  <span>enforcement</span>
                </Tooltip>
                : {r.enforcement.join(' · ')}
              </div>
            )}
            {r.code_ref && (
              <div className="text-[9px] text-[var(--color-muted)] mt-0.5 font-mono">
                code: {r.code_ref}
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  );
}
