'use client';

/**
 * GherkinScenariosTab — discovery enriquecido de escenarios (decisión operador:
 * el cockpit NO ejecuta tests; muestra cada escenario del spec con su estado de
 * última verificación del gherkin-matrix del auditor + comando listo para copiar).
 */

import { useCallback, useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { cn } from '@/lib/cn';
import { Card } from '@/components/ui/Card';
import { Spinner, EmptyState } from '@/components/ui/Spinner';
import { getGherkinStatus, type GherkinScenario } from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';

function statusBadge(status: string): { label: string; cls: string } {
  if (status.includes('✅')) return { label: status, cls: 'bg-[#052e16] text-[#86efac]' };
  if (status.includes('⚠️')) return { label: status, cls: 'bg-[#1f1300] text-amber-400' };
  if (status.includes('❌')) return { label: status, cls: 'bg-[#450a0a] text-[#fca5a5]' };
  return { label: 'sin auditar', cls: 'bg-[var(--color-panel2)] text-[var(--color-muted)]' };
}

/** Comando ejecutable según el tipo de grader (e2e playwright vs vitest vs pytest). */
function commandFor(graderPath: string): string {
  // path estilo "{sistema}/frontend/e2e/..." → correr desde {sistema}/frontend
  const feMatch = graderPath.match(/^(.*?frontend)\/(e2e\/.*\.spec\.ts)$/);
  if (feMatch) return `cd ${feMatch[1]} && npx playwright test ${feMatch[2]}`;
  const vitestMatch = graderPath.match(/^(.*?frontend)\/(src\/.*\.test\.tsx?)$/);
  if (vitestMatch) return `cd ${vitestMatch[1]} && npx vitest run ${vitestMatch[2]}`;
  const beMatch = graderPath.match(/^(.*?backend)\/(tests\/.*\.py)$/);
  if (beMatch) return `cd ${beMatch[1]} && pytest ${beMatch[2]} -v`;
  return graderPath;
}

function ScenarioRow({ sc }: { sc: GherkinScenario }) {
  const [expanded, setExpanded] = useState(false);
  const badge = statusBadge(sc.status ?? '');

  async function copy(cmd: string) {
    try {
      await navigator.clipboard.writeText(cmd);
      toast.success('Comando copiado.');
    } catch {
      toast.error('No se pudo copiar.');
    }
  }

  return (
    <div className="border border-[var(--color-border)] rounded bg-[var(--color-panel)]">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="w-full flex items-center gap-2 px-3 py-2 text-left"
        aria-expanded={expanded}
      >
        <span className="text-[10px] text-[var(--color-muted)]">{expanded ? '▾' : '▸'}</span>
        <span className="font-mono text-[10px] font-semibold shrink-0">{sc.id}</span>
        <span className="text-[11px] truncate flex-1">{sc.title}</span>
        <span
          className={cn('text-[9px] px-1.5 py-0.5 rounded font-medium shrink-0', badge.cls)}
          title={sc.notes || undefined}
        >
          {badge.label.length > 40 ? badge.label.slice(0, 40) + '…' : badge.label}
        </span>
      </button>
      {expanded && (
        <div className="border-t border-[var(--color-border)] px-3 py-2 space-y-2">
          {sc.gherkin && (
            <pre className="text-[10px] bg-[var(--color-panel2)] border border-[var(--color-border)] rounded p-2 overflow-x-auto whitespace-pre-wrap leading-relaxed">
              {sc.gherkin}
            </pre>
          )}
          {sc.notes && (
            <p className="text-[10px] text-[var(--color-muted)] italic">{sc.notes}</p>
          )}
          {(sc.graders ?? []).length > 0 && (
            <div className="space-y-1">
              <div className="text-[9px] text-[var(--color-muted)] uppercase tracking-wide">
                Tests declarados — copiá y corré en tu terminal
              </div>
              {sc.graders.map((g) => {
                const cmd = commandFor(g);
                return (
                  <div key={g} className="flex items-center gap-1.5">
                    <code className="text-[9px] font-mono text-[var(--color-muted)] truncate flex-1">
                      {cmd}
                    </code>
                    <button
                      type="button"
                      onClick={() => copy(cmd)}
                      className="text-[9px] px-1.5 py-0.5 rounded border border-[var(--color-border)] text-[var(--color-muted)] hover:text-[var(--color-text)] hover:border-[var(--color-accent)] shrink-0"
                      title="Copiar comando al portapapeles"
                    >
                      📋 copiar
                    </button>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export function GherkinScenariosTab({ storyId }: { storyId: string }) {
  const { sistema } = useSistema();
  const [scenarios, setScenarios] = useState<GherkinScenario[] | null>(null);
  const [matrixExists, setMatrixExists] = useState(false);
  const [specExists, setSpecExists] = useState(true);
  const [loading, setLoading] = useState(true);

  const load = useCallback(() => {
    setLoading(true);
    getGherkinStatus(sistema, storyId)
      .then((d) => {
        setScenarios(d.scenarios);
        setMatrixExists(d.matrix_exists);
        setSpecExists(d.spec_exists);
      })
      .catch(() => setScenarios([]))
      .finally(() => setLoading(false));
  }, [sistema, storyId]);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
        <Spinner /> Cargando escenarios…
      </div>
    );
  }

  if (!specExists) {
    return (
      <EmptyState>
        01-spec.md aún no existe — los escenarios nacen en refining (/po-ux o /po).
      </EmptyState>
    );
  }

  if (!scenarios || scenarios.length === 0) {
    return (
      <EmptyState>
        El spec no declara escenarios SC-N todavía (sección Gherkin del 01-spec.md).
      </EmptyState>
    );
  }

  const verified = scenarios.filter((s) => (s.status ?? '').includes('✅')).length;

  return (
    <div className="space-y-2">
      <Card className="!p-2.5 flex items-center gap-3 text-[11px]">
        <span>
          <strong>{scenarios.length}</strong> escenarios en el spec
        </span>
        <span className="text-[var(--color-muted)]">·</span>
        <span className={verified === scenarios.length ? 'text-[#86efac]' : ''}>
          <strong>{verified}</strong> verificados por el auditor
        </span>
        {!matrixExists && (
          <span
            className="text-[10px] text-[var(--color-muted)] italic ml-auto"
            title="El gherkin-matrix.md lo genera /auditor en Phase D — hasta entonces los escenarios figuran sin auditar."
          >
            sin gherkin-matrix todavía
          </span>
        )}
      </Card>
      {scenarios.map((sc) => (
        <ScenarioRow key={sc.id} sc={sc} />
      ))}
      <p className="text-[10px] text-[var(--color-muted)] italic">
        El cockpit no ejecuta tests (la verificación real vive en el proceso:
        auditor + live-verify). Copiá el comando y corrélo en tu terminal.
      </p>
    </div>
  );
}
