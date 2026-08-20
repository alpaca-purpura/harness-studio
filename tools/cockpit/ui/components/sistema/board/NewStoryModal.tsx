'use client';

/**
 * NewStoryModal — "+ Nueva story": captura una idea DESDE CERO al vuelo.
 * Scaffoldea checkpoint.md + operator-input.md juntos (R4) en el estado INICIAL
 * del descriptor (F6: derivado, no literal).
 * La cap se declara recién al refinar (el cap-gate lo exige así en idea-stage).
 */

import { useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Textarea } from '@/components/ui/Textarea';
import { postStoryNew } from '@/lib/api-client';
import { useProceso } from '@/components/providers/ProcesoProvider';
import { estadoInicial } from '@/lib/proceso';
import type { Release } from '@/lib/types';

const SLUG_RE = /^[a-z0-9][a-z0-9-]*$/;

export function NewStoryModal({
  open,
  sistema,
  releases,
  onClose,
  onCreated,
}: {
  open: boolean;
  sistema: string;
  releases: Release[];
  onClose: () => void;
  onCreated: (storyId: string) => void;
}) {
  const proceso = useProceso();
  const inicial = (proceso && estadoInicial(proceso)?.id) ?? 'inicial';
  const [slug, setSlug] = useState('');
  const [goal, setGoal] = useState('');
  const [release, setRelease] = useState('');
  const [type, setType] = useState<'feature' | 'bugfix'>('feature');
  const [submitting, setSubmitting] = useState(false);

  const slugOk = SLUG_RE.test(slug) && slug.length >= 3;
  const goalOk = goal.trim().length >= 10;

  async function crear() {
    setSubmitting(true);
    try {
      const r = await postStoryNew({
        sistema,
        slug,
        goal: goal.trim(),
        release: release || undefined,
        type,
      });
      toast.success(`Story ${r.storyId} creada (${inicial}).`);
      setSlug('');
      setGoal('');
      setRelease('');
      onCreated(r.storyId);
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={`Nueva story (${inicial})`} size="md">
      <div className="space-y-3 text-xs">
        <p className="text-[11px] text-[var(--color-muted)]">
          Captura la idea al vuelo: nace en <strong>{inicial}</strong> con
          checkpoint + operator-input. El refinamiento (spec, cap, diseño)
          viene después con el arnés dueño del proceso.
        </p>
        <div>
          <label className="block text-[10px] text-[var(--color-muted)] mb-1">
            story_id (kebab-case)
          </label>
          <Input
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="ej. lisa-consent-tracker"
            autoFocus
          />
          {slug && !slugOk && (
            <p className="text-[10px] text-red-400 mt-0.5">
              kebab-case, ≥3 caracteres (a-z, 0-9, guiones)
            </p>
          )}
        </div>
        <div>
          <label className="block text-[10px] text-[var(--color-muted)] mb-1">
            ¿Qué busca lograr? (goal)
          </label>
          <Textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            placeholder="Una frase: el problema o la oportunidad…"
            rows={2}
          />
        </div>
        <div className="flex gap-3">
          <div className="flex-1">
            <label className="block text-[10px] text-[var(--color-muted)] mb-1">
              release (opcional)
            </label>
            <Select value={release} onChange={(e) => setRelease(e.target.value)}>
              <option value="">— sin asignar —</option>
              {releases.map((r) => (
                <option key={r.release_id} value={r.release_id}>
                  {r.release_id} · {r.name}
                </option>
              ))}
            </Select>
          </div>
          <div className="flex-1">
            <label className="block text-[10px] text-[var(--color-muted)] mb-1">
              tipo
            </label>
            <Select
              value={type}
              onChange={(e) => setType(e.target.value as 'feature' | 'bugfix')}
            >
              <option value="feature">feature</option>
              <option value="bugfix">bugfix (repro-first)</option>
            </Select>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <Button onClick={onClose} disabled={submitting}>
            Cancelar
          </Button>
          <Button
            variant="primary"
            disabled={!slugOk || !goalOk || submitting}
            onClick={crear}
          >
            {submitting ? 'Creando…' : 'Crear (idea)'}
          </Button>
        </div>
      </div>
    </Modal>
  );
}
