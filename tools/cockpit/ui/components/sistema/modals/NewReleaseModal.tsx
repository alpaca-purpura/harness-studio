'use client';

import { useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { createRelease } from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';

interface NewReleaseModalProps {
  open: boolean;
  onClose: () => void;
  onCreated?: () => void;
}

export function NewReleaseModal({ open, onClose, onCreated }: NewReleaseModalProps) {
  const { sistema } = useSistema();
  const [releaseId, setReleaseId] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [targetDate, setTargetDate] = useState('');
  const [submitting, setSubmitting] = useState(false);

  function reset() {
    setReleaseId('');
    setName('');
    setDescription('');
    setTargetDate('');
  }

  async function handleSubmit() {
    setSubmitting(true);
    try {
      await createRelease({
        release_id: releaseId.trim(),
        sistema,
        name: name.trim(),
        description: description.trim(),
        target_date: targetDate || null,
      });
      toast.success(`Release ${releaseId} creado`);
      reset();
      onClose();
      onCreated?.();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  const valid =
    releaseId.match(/^[A-Za-z0-9_-]+$/) && name.trim().length > 0;

  return (
    <Modal open={open} onClose={onClose} title="Nuevo Release" size="md">
      <div className="space-y-3">
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            ID (corto · ej F9, INFRA-Q2, MILESTONE-1)
          </label>
          <Input
            value={releaseId}
            onChange={(e) => setReleaseId(e.target.value)}
            placeholder="F9"
            autoFocus
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Nombre descriptivo
          </label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="ej: Camila reputation + tail marketing"
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Descripción
          </label>
          <Textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={2}
            placeholder="Objetivo de negocio o técnico del release"
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Target date (opcional)
          </label>
          <Input
            type="date"
            value={targetDate}
            onChange={(e) => setTargetDate(e.target.value)}
          />
        </div>
      </div>
      <p className="text-[10px] text-[var(--color-muted)] mt-3">
        Al confirmar crea{' '}
        <span className="font-mono">
          {sistema}/docs/product/releases/{releaseId || '{id}'}.yaml
        </span>
      </p>
      <div className="flex justify-end gap-2 mt-4">
        <Button onClick={onClose} disabled={submitting}>
          Cancelar
        </Button>
        <Button
          variant="primary"
          onClick={handleSubmit}
          disabled={!valid || submitting}
        >
          + crear release
        </Button>
      </div>
    </Modal>
  );
}
