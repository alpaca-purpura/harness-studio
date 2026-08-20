'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Textarea } from '@/components/ui/Textarea';
import { Spinner, ErrorBanner } from '@/components/ui/Spinner';
import { useSistema } from '@/components/providers/SistemaProvider';
import { postMergeRelease, type MergeReleasePlan } from '@/lib/api-client';

interface MergeReleaseModalProps {
  open: boolean;
  onClose: () => void;
  releaseId: string;
  onMerged?: () => void;
}

/**
 * Cierra un release → shipped. Gate de comportamiento (release-protocol.md § 5):
 * todas las stories done/dropped (validado server-side) + el operador confirma que
 * corrió la prueba de integración + E2E smoke y dio verde (no rompe lo anterior).
 */
export function MergeReleaseModal({
  open,
  onClose,
  releaseId,
  onMerged,
}: MergeReleaseModalProps) {
  const { sistema } = useSistema();
  const [plan, setPlan] = useState<MergeReleasePlan | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [verified, setVerified] = useState(false);
  const [note, setNote] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Prompt exacto que el operador pega en Claude Code para correr la verificación.
  // Comandos/puertos reales: slot `toolchain` + `sistemas[].ports` del seam
  // (project.config.yaml) — el cockpit no los hardcodea (F-4).
  const testPrompt = `Corré toda la suite de integración + comportamiento para verificar el release ${releaseId} de ${sistema} antes de marcarlo shipped: 1) la suite completa del proyecto (lint + typecheck + tests + coverage — comandos exactos en project.config.yaml → toolchain), 2) el smoke E2E de ${sistema} si el proyecto lo tiene. Reportá si TODO pasa en verde y si NO hay regresión sobre lo anterior. Si algo falla, frená y mostrame el error — no marco shipped hasta que esté verde.`;

  useEffect(() => {
    if (!open) return;
    setLoading(true);
    setError(null);
    setVerified(false);
    setNote('');
    postMergeRelease(sistema, releaseId, false)
      .then((p) => setPlan(p))
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [open, sistema, releaseId]);

  async function handleConfirm() {
    setSubmitting(true);
    try {
      const result = await postMergeRelease(sistema, releaseId, true, {
        verified: true,
        verificationNote: note.trim() || undefined,
      });
      if (result.executed) {
        toast.success(`${releaseId} marcado shipped ✓`);
      } else if (result.note) {
        toast(result.note, { icon: 'ℹ️' });
      }
      onClose();
      onMerged?.();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  function copyPrompt() {
    navigator.clipboard?.writeText(testPrompt).then(
      () => toast.success('Prompt copiado'),
      () => toast.error('No se pudo copiar')
    );
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Cerrar release ${releaseId} → shipped`}
      size="lg"
    >
      {loading && (
        <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
          <Spinner /> Construyendo preview…
        </div>
      )}
      {error && <ErrorBanner message={error} />}
      {plan && (
        <>
          <p className="text-xs text-[var(--color-muted)] mb-3">
            Todas las stories del release están cerradas. Para marcarlo{' '}
            <b>shipped</b> (la base sólida sobre la que todo funciona) confirmá que
            corriste la prueba de comportamiento y dio verde.
          </p>

          {/* Paso 1 — prueba de comportamiento (prompt para Claude Code) */}
          <div className="bg-[var(--color-panel2)] border border-[var(--color-border)] rounded p-3 mb-4">
            <div className="flex items-center justify-between mb-1.5">
              <span className="text-[11px] font-semibold">
                1 · Corré la prueba de integración (pegá esto en Claude Code)
              </span>
              <button
                type="button"
                onClick={copyPrompt}
                className="text-[10px] px-2 py-0.5 rounded border border-[var(--color-border)] hover:bg-[var(--color-panel)] transition-colors"
              >
                Copiar prompt
              </button>
            </div>
            <pre className="text-[10px] whitespace-pre-wrap leading-relaxed text-[var(--color-muted)] font-mono">
{testPrompt}
            </pre>
          </div>

          {/* Operaciones planeadas */}
          <div className="text-[11px] font-semibold mb-1.5">
            2 · Al confirmar, el cockpit marca el release shipped:
          </div>
          <ol className="space-y-2 mb-4">
            {plan.plan.map((op, i) => (
              <li
                key={i}
                className="bg-[var(--color-panel2)] border border-[var(--color-border)] rounded p-3 text-xs"
              >
                <div className="font-mono text-[10px] text-[var(--color-accent)] mb-1">
                  [{i + 1}] {op.op}
                </div>
                <div>{op.description}</div>
              </li>
            ))}
          </ol>
          <p className="text-[10px] text-[var(--color-muted)] mb-4">
            El cockpit escribe <span className="font-mono">status=shipped</span> +
            sellos de verificación en el release.yaml. El{' '}
            <span className="font-mono">git mv</span> de stories al archive (si falta)
            lo haces manual o vía <span className="font-mono">/pm-{sistema}</span>. El
            pase a producción es otro paso (futuro).
          </p>

          {/* Nota de verificación (opcional) */}
          <label className="text-[11px] text-[var(--color-muted)] mb-1 block">
            Nota de verificación (opcional · qué corriste, resultado)
          </label>
          <Textarea
            value={note}
            onChange={(e) => setNote(e.target.value)}
            rows={2}
            placeholder="ej: suite full verde + smoke E2E 12/12 pass · sin regresión"
            className="mb-3"
          />

          {/* Checkbox gate */}
          <label className="flex items-start gap-2 text-xs mb-4">
            <input
              type="checkbox"
              className="!w-auto mt-0.5"
              checked={verified}
              onChange={(e) => setVerified(e.target.checked)}
            />
            <span>
              Confirmo que corrí la prueba de integración + E2E smoke y dio{' '}
              <b>verde</b>: todas las historias del release funcionan y{' '}
              <b>no rompen lo anterior</b>.
            </span>
          </label>

          <div className="flex justify-end gap-2">
            <Button onClick={onClose} disabled={submitting}>
              Cancelar
            </Button>
            <Button
              variant="primary"
              disabled={!verified || submitting}
              onClick={handleConfirm}
            >
              Marcar shipped
            </Button>
          </div>
        </>
      )}
    </Modal>
  );
}
