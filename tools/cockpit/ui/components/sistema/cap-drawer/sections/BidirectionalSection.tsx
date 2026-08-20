/**
 * BidirectionalSection · 🔍 Validación bidireccional
 * Lee del BidirectionalValidationReport y muestra drift counts relevantes al cap.
 */

import { cn } from '@/lib/cn';
import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import type { BidirectionalValidationReport } from '@/lib/types';

interface BidirectionalSectionProps {
  capId: string;
  report: BidirectionalValidationReport | null;
  hint?: string | null;
}

const VERDICT_CLS = {
  CLEAN: 'bg-[#14532d] text-[#86efac]',
  SOFT_DRIFT: 'bg-[#713f12] text-[#fbbf24]',
  HARD_FAIL: 'bg-[#450a0a] text-[#fca5a5]',
} as const;

export function BidirectionalSection({
  capId,
  report,
  hint,
}: BidirectionalSectionProps) {
  if (!report) {
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.bidirectional_validation} variant="header">
            🔍 Validación bidireccional
          </Tooltip>
        </h3>
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Sin reporte bidireccional disponible.
          {hint && (
            <div className="mt-1 font-mono text-[10px] text-[var(--color-text)]">
              {hint}
            </div>
          )}
        </div>
      </section>
    );
  }

  // cross_check_1/2 (atomics↔headers) eliminados 2026-05-28 (atomics killed · ver lifecycle.md).
  // Solo sobreviven cross_check_3 (scenario→e2e, HARD) y cross_check_4 (access roles).
  const checks = [
    { num: 3, label: 'scenario → e2e test', result: report.cross_check_3, tip: TOOLTIPS.cross_check_3 },
    { num: 4, label: 'access role → @decorator', result: report.cross_check_4, tip: TOOLTIPS.cross_check_4 },
  ].filter((c) => c.result != null);

  const verdictCls = VERDICT_CLS[report.summary.verdict];
  const isHard = (n: number) => report.hard_checks.includes(n);

  // Try to find this cap in the cross-check details to surface drift specific to it
  const capRelatedDrift = checks.flatMap((c) =>
    (c.result.details ?? [])
      .filter((d) => {
        const capField = (d.cap ?? d.cap_id ?? '') as string;
        return capField === capId;
      })
      .map((d) => ({ check: c.num, label: c.label, detail: d }))
  );

  return (
    <section>
      <h3 className="text-sm font-semibold mb-2 flex items-center gap-2">
        <Tooltip content={TOOLTIPS.bidirectional_validation} variant="header">
          🔍 Validación bidireccional
        </Tooltip>
        <Tooltip
          content={
            report.summary.verdict === 'CLEAN'
              ? TOOLTIPS.clean
              : report.summary.verdict === 'SOFT_DRIFT'
              ? TOOLTIPS.soft_drift
              : TOOLTIPS.hard_fail
          }
          variant="badge"
        >
          <Pill className={cn(verdictCls, 'text-[9px] py-0')}>
            {report.summary.verdict}
          </Pill>
        </Tooltip>
      </h3>
      <dl className="grid grid-cols-[1fr_auto] gap-x-3 gap-y-1 text-[11px]">
        {checks.map((c) => {
          const driftCls =
            c.result.drift > 0
              ? isHard(c.num)
                ? 'text-red-400 font-medium'
                : 'text-amber-400'
              : 'text-[#86efac]';
          return (
            <div key={c.num} className="contents">
              <dt className="text-[var(--color-muted)] text-[10px] flex items-center gap-1">
                <Tooltip content={c.tip}>
                  <span>cross_check_{c.num}</span>
                </Tooltip>{' '}
                <span className="text-[9px] opacity-70">· {c.label}</span>
                {isHard(c.num) && (
                  <Pill className="bg-[#450a0a] text-[#fca5a5] text-[8px] py-0">HARD</Pill>
                )}
              </dt>
              <dd className={`text-[10px] font-mono ${driftCls}`}>
                {c.result.pass}/{c.result.total} pass · {c.result.drift} drift
              </dd>
            </div>
          );
        })}
      </dl>

      {capRelatedDrift.length > 0 && (
        <div className="mt-2 pt-2 border-t border-[var(--color-border)]">
          <div className="text-[10px] text-[var(--color-muted)] font-semibold mb-1">
            Drift específico a esta cap:
          </div>
          <ul className="space-y-0.5">
            {capRelatedDrift.slice(0, 5).map((d, i) => (
              <li key={i} className="text-[10px] font-mono text-amber-400">
                [check {d.check}] {JSON.stringify(d.detail)}
              </li>
            ))}
            {capRelatedDrift.length > 5 && (
              <li className="text-[10px] text-[var(--color-muted)] italic">
                … y {capRelatedDrift.length - 5} más
              </li>
            )}
          </ul>
        </div>
      )}
    </section>
  );
}
