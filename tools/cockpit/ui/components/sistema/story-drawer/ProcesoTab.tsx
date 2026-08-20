'use client';

/**
 * ProcesoTab — la story dentro del ciclo SDD v5, sin jerga:
 *   · stepper idea→done con los gates G·R visibles antes de reviewing
 *   · panel Gate G (signoff del operador — form cuando la story está en developed)
 *   · panel DoD live-verify (rule #37) con la evidencia ejercida
 *   · panel Reconciliación (gate R)
 * Diseñado para que un dev nuevo entienda QUÉ pasa y QUÉ le toca sin capacitación.
 */

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import toast from 'react-hot-toast';
import { cn } from '@/lib/cn';
import type { StoryWithArchive } from '@/lib/api-client';
import { postOperatorVerify } from '@/lib/api-client';
import type { ProcesoCategoria, ProcesoResponse, StoryState } from '@/lib/types';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Textarea } from '@/components/ui/Textarea';
import { Tooltip } from '@/components/ui/Tooltip';
import { OperatorTransitions } from './OperatorTransitions';
import { useDrawer } from '@/components/providers/DrawerProvider';
import { useProceso } from '@/components/providers/ProcesoProvider';
import {
  duenoRender,
  esTerminal,
  estadoDe,
  gateOperadorEstado,
  happyPath,
} from '@/lib/proceso';

// F6/RN-52: el stepper (flujo, quién/qué, estilos, gate G) se deriva del descriptor —
// murieron MAIN_FLOW · STEP_META · STEP_CURRENT_CLS y el literal del gate. Los paneles
// chris_verify/reconciled quedan: son narrativa de ENTIDAD del SDD v5, no topología.

const RESULT_OPTIONS = [
  { value: 'SATISFIED', label: 'SATISFIED — todo OK' },
  { value: 'SATISFIED_WITH_FOLLOWUPS', label: 'SATISFIED_WITH_FOLLOWUPS — OK con pendientes menores' },
  { value: 'REJECTED', label: 'REJECTED — vuelve a dev' },
] as const;

// Theme de PRESENTACIÓN del step activo por id nativo + fallback por categoría (RN-51).
const STEP_THEME: Partial<Record<string, string>> = {
  idea: 'bg-[#374151] text-white',
  refining: 'bg-[#1d4ed8] text-white',
  refined: 'bg-[#6d28d9] text-white',
  ready: 'bg-[#0e7490] text-white',
  developing: 'bg-[#b45309] text-white',
  developed: 'bg-[#15803d] text-white',
  reviewing: 'bg-[#a21caf] text-white',
  done: 'bg-[#166534] text-white',
};

const STEP_CATEGORIA_CLS: Record<ProcesoCategoria, string> = {
  propuesto: 'bg-[#374151] text-white',
  'en-progreso': 'bg-[#b45309] text-white',
  completado: 'bg-[#166534] text-white',
  descartado: 'bg-[#7f1d1d] text-white',
  pausado: 'bg-[#3f3f46] text-white',
};

function Step({
  proceso,
  state,
  current,
  done,
}: {
  proceso: ProcesoResponse;
  state: string;
  current: boolean;
  done: boolean;
}) {
  const e = estadoDe(proceso, state);
  const who = duenoRender(proceso, state) ?? (e?.inicial ? 'operador' : 'proceso');
  const what = e?.descripcion ?? state;
  const currentCls =
    STEP_THEME[state] ?? (e ? STEP_CATEGORIA_CLS[e.categoria] : 'bg-[#374151] text-white');
  return (
    <Tooltip content={`${who} — ${what}`} variant="badge">
      <span
        className={cn(
          'text-[10px] px-2 py-1 rounded-full border whitespace-nowrap cursor-help',
          current
            ? cn('font-bold border-transparent', currentCls)
            : done
              ? 'text-[var(--color-text)] border-[var(--color-border)] bg-[var(--color-panel2)]'
              : 'text-[var(--color-muted)] border-[var(--color-border)] opacity-60'
        )}
      >
        {done && !current ? '✓ ' : ''}
        {state}
      </span>
    </Tooltip>
  );
}

function PanelTitle({ children, tip }: { children: React.ReactNode; tip: string }) {
  return (
    <div className="text-xs font-semibold mb-2 flex items-center gap-1.5">
      <Tooltip content={tip} variant="header">
        <span className="cursor-help">{children}</span>
      </Tooltip>
    </div>
  );
}

function KV({ k, children }: { k: string; children: React.ReactNode }) {
  return (
    <div className="flex gap-2 text-[11px] py-0.5">
      <span className="text-[var(--color-muted)] min-w-[130px] shrink-0">{k}</span>
      <span className="min-w-0">{children}</span>
    </div>
  );
}

export function ProcesoTab({
  story,
  onUpdated,
}: {
  story: StoryWithArchive;
  onUpdated: () => void;
}) {
  const router = useRouter();
  const proceso = useProceso();
  const { closeStory } = useDrawer();
  const [result, setResult] = useState<string>('SATISFIED');
  const [notes, setNotes] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!proceso) return null;

  const mainFlow = happyPath(proceso).map((e) => e.id);
  const gateEstado = gateOperadorEstado(proceso); // el "gate G": momento-estado con autoridad operador
  const curIdx = mainFlow.indexOf(story.state);
  const offFlow = curIdx < 0; // categorías pausado/descartado (fuera del ciclo feliz)
  const estadoActual = estadoDe(proceso, story.state);
  const lanzable =
    !esTerminal(proceso, story.state) && estadoActual?.categoria !== 'pausado';
  const signoff = story.chris_verify?.signoff ?? null;
  const signed = Boolean(signoff?.result);
  const signedBy =
    (signoff as { by?: string; signed_by?: string } | null)?.by ??
    (signoff as { by?: string; signed_by?: string } | null)?.signed_by ??
    null;
  const awaiting = story.phase === 'AWAIT_CHRIS_VERIFY';
  const evidence = story.dod_evidence ?? [];

  async function firmar() {
    if (notes.trim().length < 5) {
      toast.error('Contá qué ejerciste y qué viste (≥5 caracteres).');
      return;
    }
    setSubmitting(true);
    try {
      await postOperatorVerify({
        sistema: story.sistema,
        story_id: story.story_id,
        result: result as 'SATISFIED' | 'SATISFIED_WITH_FOLLOWUPS' | 'REJECTED',
        notes: notes.trim(),
      });
      toast.success('Signoff del gate G firmado.');
      setNotes('');
      onUpdated();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="space-y-3">
      {/* ── Aviso accionable: la story te espera ── */}
      {awaiting && !signed && (
        <div className="border border-amber-700 bg-[#1f1300] text-amber-300 rounded px-3 py-2 text-[11px]">
          ⏳ <strong>Esta story te está esperando (gate G).</strong> Ejercé el
          kit en la app corriendo (la acción real, no solo mirar) y firmá el
          signoff acá abajo. Recién después pasa al auditor.
        </div>
      )}

      {/* ── Stepper ── */}
      {offFlow ? (
        <Card className="!p-3">
          <div className="text-xs">
            Estado fuera del ciclo: <strong>{story.state}</strong>
            {story.parked_reason && (
              <span className="text-[var(--color-muted)]"> — {story.parked_reason}</span>
            )}
            {story.dropped_reason && (
              <span className="text-[var(--color-muted)]"> — {story.dropped_reason}</span>
            )}
          </div>
          <div className="text-[10px] text-[var(--color-muted)] mt-1">
            {estadoActual?.descripcion ??
              'pausada o descartada — fuera del ciclo feliz del proceso.'}
          </div>
        </Card>
      ) : (
        <Card className="!p-3">
          <div className="flex items-center gap-1 flex-wrap">
            {mainFlow.map((s, i) => (
              <span key={s} className="flex items-center gap-1">
                {i > 0 && <span className="text-[var(--color-border)] text-[10px]">→</span>}
                {gateEstado && i > 0 && mainFlow[i - 1] === gateEstado && (
                  <>
                    <Tooltip
                      variant="badge"
                      content="Gate G: el operador ejerce el kit live y firma el signoff. Gate R: /pm reconcilia los docs con la realidad construida ANTES del auditor."
                    >
                      <span className="text-[9px] px-1.5 py-0.5 rounded border border-dashed border-[var(--color-border)] text-amber-400 cursor-help">
                        G·R
                      </span>
                    </Tooltip>
                    <span className="text-[var(--color-border)] text-[10px]">→</span>
                  </>
                )}
                <Step proceso={proceso} state={s} current={i === curIdx} done={i < curIdx} />
              </span>
            ))}
          </div>
          <div className="text-[10px] text-[var(--color-muted)] mt-2">
            Tú solo decidés las transiciones de operador y el gate del proceso;
            el resto lo mueven los arneses dueños de cada estado.
          </div>
        </Card>
      )}

      {/* ── Gate G · signoff ── */}
      <Card className="!p-3">
        <PanelTitle tip="Proceso v5: con la story en developed, el operador ejerce el kit live ANTES del auditor y firma el resultado. Un solo signoff (chris_verify.signoff).">
          🖊 Gate G · verificación del operador
        </PanelTitle>
        {signed ? (
          <div>
            <KV k="resultado">
              <span
                className={cn(
                  'font-semibold',
                  signoff?.result === 'REJECTED' ? 'text-red-400' : 'text-[#86efac]'
                )}
              >
                {signoff?.result}
              </span>
            </KV>
            <KV k="firmado por / fecha">
              {signedBy ?? '—'} · {signoff?.date ?? '—'}
            </KV>
            {signoff?.notes && <KV k="notas">{signoff.notes}</KV>}
            {(signoff?.open_items?.length ?? 0) > 0 && (
              <KV k="pendientes">
                <ul className="list-disc list-inside">
                  {signoff!.open_items!.map((it, i) => (
                    <li key={i}>{String(it)}</li>
                  ))}
                </ul>
              </KV>
            )}
          </div>
        ) : story.state === gateEstado ? (
          <div className="space-y-2">
            <select
              value={result}
              onChange={(e) => setResult(e.target.value)}
              className="w-full bg-[var(--color-panel2)] border border-[var(--color-border)] rounded px-2 py-1.5 text-[11px]"
            >
              {RESULT_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>
            <Textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="¿Qué ejerciste en la app y qué viste? (ej: creé un lead, apareció en el board, el backend logueó 201)"
              rows={2}
            />
            <Button onClick={firmar} disabled={submitting}>
              {submitting ? 'Firmando…' : 'Firmar signoff'}
            </Button>
          </div>
        ) : (
          <div className="text-[11px] text-[var(--color-muted)]">
            Sin firmar — el form aparece cuando la story llega a{' '}
            <strong>{gateEstado ?? 'el estado del gate'}</strong> (ahí te toca probarla live).
          </div>
        )}
      </Card>

      {/* ── DoD live-verify ── */}
      <Card className="!p-3">
        <PanelTitle tip="Rule #37: ninguna story es done sin ejercer la acción real contra el stack dev + leer logs + confirmar el efecto. Una suite verde o un GET 200 NO cuentan.">
          ✅ DoD · verificación live
        </PanelTitle>
        <KV k="verificada live">
          {story.dod_live_verified ? (
            <span className="text-[#86efac] font-semibold">✓ sí</span>
          ) : (
            <span className="text-amber-400">pendiente</span>
          )}
        </KV>
        {story.dod_env && <KV k="entorno">{story.dod_env}</KV>}
        {story.verified_at && <KV k="fecha">{String(story.verified_at)}</KV>}
        {evidence.length > 0 && (
          <div className="mt-2 space-y-1.5">
            {evidence.map((ev, i) => (
              <div
                key={i}
                className="border-l-2 border-[#15803d] bg-[var(--color-panel2)] rounded-r px-2.5 py-1.5 text-[10px] space-y-0.5"
              >
                {ev.action && (
                  <div>
                    <span className="text-[var(--color-muted)]">acción:</span> {ev.action}
                  </div>
                )}
                {ev.observed && (
                  <div>
                    <span className="text-[var(--color-muted)]">observado:</span> {ev.observed}
                  </div>
                )}
                {ev.backend_log && (
                  <div className="font-mono">
                    <span className="text-[var(--color-muted)] font-sans">log:</span>{' '}
                    {ev.backend_log}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* ── Gate R · reconciliación ── */}
      <Card className="!p-3">
        <PanelTitle tip="Gate R: /pm alinea spec/arch/validators/cap con lo realmente construido y ratificado. El auditor NO empieza sin reconciled: true (leería un spec stale).">
          🔄 Gate R · docs reconciliados
        </PanelTitle>
        <KV k="reconciled">
          {story.reconciled ? (
            <span className="text-[#86efac] font-semibold">✓ sí</span>
          ) : (
            <span className="text-[var(--color-muted)]">todavía no (lo hace /pm después del gate G)</span>
          )}
        </KV>
      </Card>

      {/* ── Acciones del operador ── */}
      <Card className="!p-3">
        <OperatorTransitions story={story} onUpdated={onUpdated} />
        {lanzable && (
          <div className="mt-3 pt-3 border-t border-[var(--color-border)]">
            <Button
              size="sm"
              variant="primary"
              onClick={() => {
                closeStory();
                router.push(`/delivery?story=${encodeURIComponent(story.story_id)}`);
              }}
              title="Abre el cockpit de delivery con esta story preseleccionada (el descriptor deriva tramo, arnés y gate)"
            >
              ⚡ Lanzar sesión de delivery
            </Button>
          </div>
        )}
      </Card>
    </div>
  );
}
