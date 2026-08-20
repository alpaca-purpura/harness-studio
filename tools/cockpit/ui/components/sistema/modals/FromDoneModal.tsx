'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Select } from '@/components/ui/Select';
import { listReleases, postFromDone } from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';
import type { Release } from '@/lib/types';

interface FromDoneModalProps {
  open: boolean;
  onClose: () => void;
  parentStoryId: string;
  onCreated?: (storyId: string) => void;
}

export function FromDoneModal({
  open,
  onClose,
  parentStoryId,
  onCreated,
}: FromDoneModalProps) {
  const { sistema } = useSistema();
  const [slug, setSlug] = useState('');
  const [goal, setGoal] = useState('');
  const [release, setRelease] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [releases, setReleases] = useState<Release[]>([]);

  useEffect(() => {
    if (!open) return;
    listReleases(sistema)
      .then(setReleases)
      .catch(() => setReleases([]));
  }, [open, sistema]);

  async function handleSubmit() {
    setSubmitting(true);
    try {
      const result = await postFromDone({
        sistema,
        parentStoryId,
        newStorySlug: slug.trim(),
        goal: goal.trim(),
        release: release.trim(),
      });
      toast.success(`Story creada: ${result.storyId}`);
      onClose();
      onCreated?.(result.storyId);
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  const valid =
    slug.match(/^[a-z0-9][a-z0-9-]*$/) && goal.length >= 10 && release.length > 0;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Nueva story basada en ${parentStoryId}`}
      size="md"
    >
      <p className="text-xs text-[var(--color-muted)] mb-3">
        Crea una story nueva que hereda parent · auto-fill cap_target del parent
        + appendea ref al operator-input de la nueva.
      </p>
      <div className="space-y-3">
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Slug nueva story
          </label>
          <Input
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="ej: mi-feature-evolucion"
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Goal (≥10 caracteres)
          </label>
          <Textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            rows={2}
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Release destino
          </label>
          <Select value={release} onChange={(e) => setRelease(e.target.value)}>
            <option value="">— seleccionar release —</option>
            {releases.map((r) => (
              <option key={r.release_id} value={r.release_id}>
                {r.release_id} · {r.name}
              </option>
            ))}
          </Select>
        </div>
      </div>
      <div className="flex justify-end gap-2 mt-4">
        <Button onClick={onClose} disabled={submitting}>
          Cancelar
        </Button>
        <Button
          variant="primary"
          onClick={handleSubmit}
          disabled={!valid || submitting}
        >
          Crear story
        </Button>
      </div>
    </Modal>
  );
}
