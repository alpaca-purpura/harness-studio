'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { Select } from '@/components/ui/Select';
import { Card } from '@/components/ui/Card';
import { cn } from '@/lib/cn';
import { postExtendCap, listReleases } from '@/lib/api-client';
import { useEffect } from 'react';
import type { Release } from '@/lib/types';

type ExtendCapType = 'fix' | 'extend' | 'derive';

interface ExtendCapModalProps {
  open: boolean;
  onClose: () => void;
  sistema: string;
  parentCap: { module: string; slug: string };
  onCreated?: (storyId: string) => void;
}

const TYPE_CARDS: Array<{
  key: ExtendCapType;
  emoji: string;
  title: string;
  desc: string;
}> = [
  {
    key: 'fix',
    emoji: '🔧',
    title: 'Fix',
    desc: 'corrige bug · no agrega scenarios nuevos · mismo cap',
  },
  {
    key: 'extend',
    emoji: '➕',
    title: 'Extend',
    desc: 'agrega scenario nuevo al cap actual · misma identidad',
  },
  {
    key: 'derive',
    emoji: '🌱',
    title: 'Derive',
    desc: 'crea cap hijo distinto · misma raíz · identidad nueva',
  },
];

export function ExtendCapModal({
  open,
  onClose,
  sistema,
  parentCap,
  onCreated,
}: ExtendCapModalProps) {
  const [type, setType] = useState<ExtendCapType>('fix');
  const [slug, setSlug] = useState('');
  const [goal, setGoal] = useState('');
  const [release, setRelease] = useState('');
  const [derivedName, setDerivedName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [releases, setReleases] = useState<Release[]>([]);

  useEffect(() => {
    if (!open) return;
    listReleases(sistema)
      .then(setReleases)
      .catch(() => setReleases([]));
  }, [open, sistema]);

  function reset() {
    setType('fix');
    setSlug('');
    setGoal('');
    setRelease('');
    setDerivedName('');
  }

  async function handleSubmit() {
    setSubmitting(true);
    try {
      const result = await postExtendCap({
        sistema,
        parentCap,
        capChangeType: type,
        newStorySlug: slug.trim(),
        goal: goal.trim(),
        release: release.trim(),
        derivedName: type === 'derive' ? derivedName.trim() : undefined,
      });
      toast.success(`Story creada: ${result.storyId}`);
      reset();
      onClose();
      onCreated?.(result.storyId);
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  const validBase =
    slug.match(/^[a-z0-9][a-z0-9-]*$/) && goal.length >= 10 && release.length > 0;
  const valid =
    type === 'derive' ? validBase && derivedName.match(/^[a-z0-9][a-z0-9-]*$/) : validBase;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Extender capability · ${parentCap.module}/${parentCap.slug}`}
      size="lg"
    >
      <div className="space-y-4">
        <div>
          <div className="text-xs text-[var(--color-muted)] mb-2 font-medium">
            ¿Qué tipo de cambio quieres sobre este capability?
          </div>
          <div className="grid grid-cols-3 gap-2">
            {TYPE_CARDS.map((c) => (
              <button
                key={c.key}
                type="button"
                onClick={() => setType(c.key)}
                className={cn(
                  'text-left p-3 rounded border transition-all',
                  type === c.key
                    ? 'bg-[var(--color-accent)]/10 border-[var(--color-accent)]'
                    : 'bg-[var(--color-panel2)] border-[var(--color-border)] hover:border-[#3a4358]'
                )}
              >
                <div className="text-base mb-1">
                  {c.emoji} <span className="font-semibold text-sm">{c.title}</span>
                </div>
                <div className="text-[11px] text-[var(--color-muted)]">{c.desc}</div>
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Slug de la nueva story (kebab-case)
          </label>
          <Input
            value={slug}
            onChange={(e) => setSlug(e.target.value)}
            placeholder="ej: appointments-add-multi-doctor"
          />
        </div>

        {type === 'derive' && (
          <div>
            <label className="text-xs text-[var(--color-muted)] mb-1 block">
              Slug del cap derivado a materializar al merge
            </label>
            <Input
              value={derivedName}
              onChange={(e) => setDerivedName(e.target.value)}
              placeholder="ej: appointments-multi-doctor"
            />
          </div>
        )}

        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Goal (≥10 caracteres)
          </label>
          <Textarea
            value={goal}
            onChange={(e) => setGoal(e.target.value)}
            rows={2}
            placeholder="Resumen del objetivo de esta story…"
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

        <div className="flex justify-end gap-2 pt-2">
          <Button onClick={onClose} disabled={submitting}>
            Cancelar
          </Button>
          <Button
            variant="primary"
            disabled={!valid || submitting}
            onClick={handleSubmit}
          >
            Crear story
          </Button>
        </div>
      </div>
    </Modal>
  );
}
