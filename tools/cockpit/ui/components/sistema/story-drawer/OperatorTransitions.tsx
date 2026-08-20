'use client';

/**
 * OperatorTransitions — las acciones manuales que el operador puede ejecutar
 * sobre la story desde el cockpit. Vive en el tab Proceso: es la respuesta a
 * "¿qué puedo hacer YO acá?".
 *
 * F6/RN-51: la whitelist, los labels, los dueños y la razón mínima se derivan
 * del descriptor de proceso (useProceso) — cero literales del ciclo acá.
 */

import { useState } from 'react';
import toast from 'react-hot-toast';
import { postTransition } from '@/lib/api-client';
import type { StoryState } from '@/lib/types';
import type { StoryWithArchive } from '@/lib/api-client';
import { Button } from '@/components/ui/Button';
import { StateBadge } from '@/components/ui/Badge';
import { Textarea } from '@/components/ui/Textarea';
import { Modal } from '@/components/ui/Modal';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useProceso } from '@/components/providers/ProcesoProvider';
import {
  ejecutorSalida,
  esTerminal,
  labelTransicion,
  transicionesOperador,
} from '@/lib/proceso';

export function OperatorTransitions({
  story,
  onUpdated,
}: {
  story: StoryWithArchive;
  onUpdated: () => void;
}) {
  const { sistema } = useSistema();
  const proceso = useProceso();
  const [pendingTarget, setPendingTarget] = useState<StoryState | null>(null);
  const [reason, setReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!proceso) return null;

  const allowedTransitions = transicionesOperador(proceso).filter(
    (t) => t.de === story.state
  );
  const isTerminal = esTerminal(proceso, story.state);
  const razonMinima = proceso.parametros.razon_minima;

  async function executeTransition(target: StoryState, reasonText: string) {
    setSubmitting(true);
    try {
      await postTransition(sistema, story.story_id, target, reasonText || undefined);
      toast.success(`Estado: ${story.state} → ${target}`);
      setPendingTarget(null);
      setReason('');
      onUpdated();
    } catch (err) {
      toast.error(`Error: ${(err as Error).message}`);
    } finally {
      setSubmitting(false);
    }
  }

  function handleTransitionClick(target: StoryState) {
    const transition = allowedTransitions.find((t) => t.a === target);
    if (!transition) return;
    if (transition.requiere_razon) {
      setPendingTarget(target);
    } else {
      executeTransition(target, '');
    }
  }

  return (
    <div>
      <div className="text-xs text-[var(--color-muted)] mb-2 font-medium">
        ¿Qué puedo hacer yo acá?
      </div>
      {allowedTransitions.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {allowedTransitions.map((t) => (
            <Button
              key={t.a}
              variant={
                esTerminal(proceso, t.a) && t.requiere_razon ? 'danger' : 'default'
              }
              size="sm"
              onClick={() => handleTransitionClick(t.a as StoryState)}
              disabled={submitting}
              title={t.verbo ? `evento: ${t.verbo}` : undefined}
            >
              {labelTransicion(t)} → <StateBadge state={t.a as StoryState} className="ml-1" />
            </Button>
          ))}
        </div>
      ) : (
        <div className="text-xs text-[var(--color-muted)] italic">
          {isTerminal
            ? 'Estado terminal · no admite transiciones desde el cockpit.'
            : `La transición desde ${story.state} la ejecuta ${
                ejecutorSalida(proceso, story.state) ?? 'una skill de Claude'
              }. Invócala desde Claude Code.`}
        </div>
      )}

      <Modal
        open={pendingTarget !== null}
        onClose={() => {
          setPendingTarget(null);
          setReason('');
        }}
        title={`Razón obligatoria → ${pendingTarget}`}
        size="md"
      >
        <p className="text-xs text-[var(--color-muted)] mb-3">
          Esta razón queda registrada en checkpoint.md y sirve como contexto
          futuro (al reactivar una pausada o auditar una descartada).
        </p>
        <Textarea
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder={`Por qué tomás esta decisión… (mínimo ${razonMinima} caracteres)`}
          rows={4}
          autoFocus
        />
        <div className="flex justify-end gap-2 mt-4">
          <Button
            onClick={() => {
              setPendingTarget(null);
              setReason('');
            }}
            disabled={submitting}
          >
            Cancelar
          </Button>
          <Button
            variant={
              pendingTarget && esTerminal(proceso, pendingTarget) ? 'danger' : 'primary'
            }
            disabled={reason.trim().length < razonMinima || submitting}
            onClick={() => pendingTarget && executeTransition(pendingTarget, reason)}
          >
            Confirmar
          </Button>
        </div>
      </Modal>
    </div>
  );
}
