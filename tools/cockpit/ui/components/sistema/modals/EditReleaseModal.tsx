'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import { updateRelease } from '@/lib/api-client';
import { useSistema } from '@/components/providers/SistemaProvider';
import type { Release } from '@/lib/types';

interface EditReleaseModalProps {
  open: boolean;
  onClose: () => void;
  release: Release | null;
  onUpdated?: () => void;
}

/**
 * Edita nombre / descripción / target_date de un release por venir.
 * Los releases `shipped` son inmutables — este modal no se abre para ellos
 * (el botón "Editar" está oculto en el ReleaseCard) y el server lo rechaza igual.
 */
export function EditReleaseModal({ open, onClose, release, onUpdated }: EditReleaseModalProps) {
  const { sistema } = useSistema();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [targetDate, setTargetDate] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (release) {
      setName(release.name ?? '');
      setDescription(release.description ?? '');
      setTargetDate(release.target_date ?? '');
    }
  }, [release]);

  async function handleSubmit() {
    if (!release) return;
    setSubmitting(true);
    try {
      await updateRelease(release.release_id, sistema, {
        name: name.trim(),
        description: description.trim(),
        target_date: targetDate || null,
      });
      toast.success(`Release ${release.release_id} actualizado`);
      onClose();
      onUpdated?.();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  const valid = name.trim().length > 0;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={release ? `Editar release ${release.release_id}` : 'Editar release'}
      size="md"
    >
      <div className="space-y-3">
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Nombre descriptivo
          </label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="ej: Camila reputation + tail marketing"
            autoFocus
          />
        </div>
        <div>
          <label className="text-xs text-[var(--color-muted)] mb-1 block">
            Descripción
          </label>
          <Textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
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
        Solo se editan releases por venir. Un release{' '}
        <span className="font-mono">shipped</span> es inmutable (base entregada).
      </p>
      <div className="flex justify-end gap-2 mt-4">
        <Button onClick={onClose} disabled={submitting}>
          Cancelar
        </Button>
        <Button variant="primary" onClick={handleSubmit} disabled={!valid || submitting}>
          Guardar cambios
        </Button>
      </div>
    </Modal>
  );
}
