'use client';

/**
 * DoctorGates — los 9 gates canónicos cap↔código (G1-G9), mismo SSoT que el
 * pre-commit (`_bidirectional-validation.json` del validador bidireccional).
 * Verde = sin drift; un gate en rojo expande el detalle con qué arreglar.
 */

import { useCallback, useEffect, useState } from 'react';
import { cn } from '@/lib/cn';
import { Card } from '@/components/ui/Card';
import { Tooltip } from '@/components/ui/Tooltip';
import toast from 'react-hot-toast';
import { getCapDoctor, postCapRegen, type CapDoctorReport } from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchEvents } from '@/components/providers/FileWatchProvider';

const GATE_TIPS: Record<string, string> = {
  G1: 'Cada header `# cap:` del código debe resolver a una capability existente.',
  G2: 'Cada área live del SYSTEM-MAP debe tener al menos una cap.',
  G3: 'Cada cap debe tener hogar (functional_area presente en el SYSTEM-MAP).',
  G4: 'Los paths de código que la cap declara deben existir en disco.',
  G5: 'Las supersesiones (cap A reemplaza a B) deben apuntar a caps reales.',
  G6: 'Una cap live debe ser visible en el mapa del producto.',
  G7: 'El YAML de la cap debe ser legible por el cockpit (sin keys duplicadas).',
  G8: 'Caps live+visibles necesitan user_facing_description.',
  G9: 'Caps live+visibles necesitan scenarios.',
};

function detailLine(d: unknown): string {
  if (typeof d === 'string') return d;
  if (d && typeof d === 'object') {
    const o = d as Record<string, unknown>;
    return String(o.drift_reason ?? o.reason ?? JSON.stringify(d));
  }
  return String(d);
}

export function DoctorGates({ onRegenerated }: { onRegenerated?: () => void } = {}) {
  const { sistema } = useSistema();
  const [report, setReport] = useState<CapDoctorReport | null>(null);
  const [hint, setHint] = useState<string | null>(null);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [regenning, setRegenning] = useState(false);

  const load = useCallback(() => {
    getCapDoctor(sistema)
      .then((data) => {
        setReport(data.doctor);
        setHint(data.hint ?? null);
      })
      .catch(() => setReport(null));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  useFileWatchEvents((event) => {
    if (event.sistema && event.sistema !== sistema) return;
    if (event.docType === 'capability') load();
  });

  async function regen() {
    setRegenning(true);
    try {
      const r = await postCapRegen(sistema);
      if (r.ok) {
        toast.success(`Índices regenerados (${Math.round((r.duration_ms ?? 0) / 1000)}s).`);
        load();
        onRegenerated?.();
      } else {
        toast(r.hint ?? 'Regen no disponible en este workspace.', { icon: 'ℹ️' });
      }
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setRegenning(false);
    }
  }

  if (!report) {
    return hint ? (
      <Card className="!p-3 text-[10px] text-[var(--color-muted)]">
        Gates G1-G9 sin reporte. {hint}
      </Card>
    ) : null;
  }

  return (
    <Card className="!p-3">
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <Tooltip
          variant="header"
          content="Los 9 gates canónicos que el pre-commit también corre: si acá está verde, cap ↔ código están en sync."
        >
          <span className="text-xs font-semibold cursor-help">
            🩺 Gates canónicos G1-G9
          </span>
        </Tooltip>
        <span
          className={cn(
            'text-[10px] px-2 py-0.5 rounded font-semibold',
            report.healthy
              ? 'bg-[#052e16] text-[#86efac]'
              : 'bg-[#450a0a] text-[#fca5a5]'
          )}
        >
          {report.healthy ? '✓ healthy' : `${report.total_drift} drift`}
        </span>
        <span className="ml-auto flex items-center gap-2">
          {report.validated_at && (
            <span className="text-[9px] text-[var(--color-muted)]">
              validado: {String(report.validated_at).slice(0, 19).replace('T', ' ')}
            </span>
          )}
          <button
            type="button"
            onClick={regen}
            disabled={regenning}
            title="Regenera los índices cap↔código AHORA (shellea los scripts canónicos del workspace — útil si editaste código y todavía no commiteaste)"
            className="text-[10px] px-2 py-0.5 rounded border border-[var(--color-border)] text-[var(--color-muted)] hover:text-[var(--color-text)] hover:border-[var(--color-accent)] transition-colors disabled:opacity-50"
          >
            {regenning ? '⏳ regenerando…' : '🔄 regenerar'}
          </button>
        </span>
      </div>
      <div className="grid grid-cols-3 sm:grid-cols-9 gap-1">
        {report.gates.map((g) => {
          const bad = g.drift > 0;
          return (
            <button
              key={g.id}
              type="button"
              onClick={() => setExpanded(expanded === g.id ? null : bad ? g.id : null)}
              title={`${g.label} — ${GATE_TIPS[g.id] ?? ''}`}
              className={cn(
                'rounded border px-1 py-1.5 text-center transition-colors',
                bad
                  ? 'border-red-800 bg-[#1f0000] text-[#fca5a5] cursor-pointer hover:bg-[#2d0000]'
                  : 'border-[var(--color-border)] bg-[var(--color-panel2)] text-[#86efac] cursor-help'
              )}
            >
              <div className="text-[10px] font-bold">{g.id}</div>
              <div className="text-[9px]">
                {bad ? `${g.drift}✗` : `${g.total - g.drift}/${g.total}`}
              </div>
            </button>
          );
        })}
      </div>
      {expanded && (
        <div className="mt-2 border-t border-[var(--color-border)] pt-2 space-y-1">
          {report.gates
            .filter((g) => g.id === expanded)
            .flatMap((g) =>
              g.details.map((d, i) => (
                <div key={i} className="text-[10px] text-[#fca5a5]">
                  • {detailLine(d)}
                </div>
              ))
            )}
        </div>
      )}
    </Card>
  );
}
