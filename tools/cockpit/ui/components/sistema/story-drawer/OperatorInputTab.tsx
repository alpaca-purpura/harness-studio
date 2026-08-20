'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Textarea } from '@/components/ui/Textarea';
import { Input } from '@/components/ui/Input';
import { Select } from '@/components/ui/Select';
import { Spinner, ErrorBanner } from '@/components/ui/Spinner';
import { useSistema } from '@/components/providers/SistemaProvider';
import {
  appendOperatorInputConversation,
  appendOperatorInputNote,
  appendOperatorInputRef,
  getOperatorInput,
} from '@/lib/api-client';
import {
  REF_TYPE_TO_EMOJI,
  VERDICT_TO_LABEL,
  type OperatorInput,
  type RefType,
} from '@/lib/types';

const REF_TYPE_OPTIONS: RefType[] = [
  'link',
  'img',
  'text',
  'story-ref',
  'learning-ref',
  'doc',
];

interface OperatorInputTabProps {
  storyId: string;
}

export function OperatorInputTab({ storyId }: OperatorInputTabProps) {
  const { sistema } = useSistema();
  const [data, setData] = useState<OperatorInput | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [missing, setMissing] = useState(false);

  const [noteDraft, setNoteDraft] = useState('');
  const [convDraft, setConvDraft] = useState('');
  const [refDraft, setRefDraft] = useState<{
    type: RefType;
    value: string;
    comment: string;
  }>({ type: 'link', value: '', comment: '' });

  function load() {
    setLoading(true);
    setError(null);
    setMissing(false);
    getOperatorInput(storyId, sistema)
      .then((ci) => setData(ci))
      .catch((err: Error & { status?: number }) => {
        if (err.status === 404) setMissing(true);
        else setError(err.message);
      })
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [storyId, sistema]);

  async function handleAppendNote() {
    if (noteDraft.trim().length === 0) return;
    try {
      const updated = await appendOperatorInputNote(storyId, sistema, noteDraft.trim());
      setData(updated);
      setNoteDraft('');
      toast.success('Nota agregada');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  async function handleAppendRef() {
    if (refDraft.value.trim().length === 0) return;
    try {
      const updated = await appendOperatorInputRef(storyId, sistema, {
        type: refDraft.type,
        value: refDraft.value.trim(),
        comment: refDraft.comment.trim() || undefined,
      });
      setData(updated);
      setRefDraft({ type: 'link', value: '', comment: '' });
      toast.success('Referencia agregada');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  async function handleAppendConv() {
    if (convDraft.trim().length === 0) return;
    try {
      const updated = await appendOperatorInputConversation(
        storyId,
        sistema,
        convDraft.trim()
      );
      setData(updated);
      setConvDraft('');
      toast.success('Mensaje enviado');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
        <Spinner /> Cargando operator-input…
      </div>
    );
  }
  if (error) return <ErrorBanner message={error} />;
  if (missing || !data) {
    return (
      <div className="text-xs text-[var(--color-muted)] italic">
        operator-input todavía no existe para esta story. Se crea cuando la story
        avanza a refining. Mientras tanto, contacto directo con Claude.
      </div>
    );
  }

  return (
    <div className="space-y-5">
      {/* Notas */}
      <section>
        <header className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-semibold">💭 Notas</h3>
          <span className="text-[10px] text-[var(--color-muted)]">
            {data.notes.length} entradas
          </span>
        </header>
        <div className="space-y-2 max-h-72 overflow-y-auto pr-1 mb-3">
          {data.notes.length === 0 && (
            <p className="text-xs text-[var(--color-muted)] italic">
              Sin notas todavía.
            </p>
          )}
          {data.notes.map((n, i) => (
            <Card key={i} className="!p-2">
              <div className="text-[10px] text-[var(--color-muted)] font-mono mb-1">
                {n.timestamp}
              </div>
              <div className="text-xs whitespace-pre-wrap">{n.text}</div>
            </Card>
          ))}
        </div>
        <div className="flex gap-2">
          <Textarea
            rows={2}
            value={noteDraft}
            onChange={(e) => setNoteDraft(e.target.value)}
            placeholder="Nueva nota…"
          />
          <Button
            variant="primary"
            onClick={handleAppendNote}
            disabled={!noteDraft.trim()}
          >
            Agregar
          </Button>
        </div>
      </section>

      {/* Refs */}
      <section>
        <header className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-semibold">📎 Referencias</h3>
          <span className="text-[10px] text-[var(--color-muted)]">
            {data.refs.length} entradas
          </span>
        </header>
        <div className="space-y-1.5 max-h-60 overflow-y-auto pr-1 mb-3">
          {data.refs.length === 0 && (
            <p className="text-xs text-[var(--color-muted)] italic">
              Sin referencias todavía.
            </p>
          )}
          {data.refs.map((r, i) => (
            <Card key={i} className="!p-2">
              <div className="flex items-center gap-2 text-xs">
                <span>{REF_TYPE_TO_EMOJI[r.type]}</span>
                <span className="font-mono text-[11px] truncate flex-1">
                  {r.value}
                </span>
              </div>
              {r.comment && (
                <div className="text-[11px] text-[var(--color-muted)] mt-1 italic">
                  {r.comment}
                </div>
              )}
            </Card>
          ))}
        </div>
        <div className="space-y-2">
          <div className="flex gap-2">
            <Select
              value={refDraft.type}
              onChange={(e) =>
                setRefDraft((s) => ({ ...s, type: e.target.value as RefType }))
              }
              className="!w-32"
            >
              {REF_TYPE_OPTIONS.map((t) => (
                <option key={t} value={t}>
                  {REF_TYPE_TO_EMOJI[t]} {t}
                </option>
              ))}
            </Select>
            <Input
              value={refDraft.value}
              onChange={(e) =>
                setRefDraft((s) => ({ ...s, value: e.target.value }))
              }
              placeholder="URL · path · slug…"
            />
          </div>
          <Input
            value={refDraft.comment}
            onChange={(e) =>
              setRefDraft((s) => ({ ...s, comment: e.target.value }))
            }
            placeholder="Comentario opcional"
          />
          <div className="flex justify-end">
            <Button
              variant="primary"
              size="sm"
              onClick={handleAppendRef}
              disabled={!refDraft.value.trim()}
            >
              + Agregar ref
            </Button>
          </div>
        </div>
      </section>

      {/* Conversación */}
      <section>
        <header className="flex items-center justify-between mb-2">
          <h3 className="text-sm font-semibold">🗣 Conversación</h3>
          <span className="text-[10px] text-[var(--color-muted)]">
            {data.conversation.length} entradas
          </span>
        </header>
        <div className="space-y-2 max-h-80 overflow-y-auto pr-1 mb-3">
          {data.conversation.length === 0 && (
            <p className="text-xs text-[var(--color-muted)] italic">
              Sin conversación todavía.
            </p>
          )}
          {data.conversation.map((c, i) => (
            <Card
              key={i}
              className={
                c.author !== 'claude'
                  ? '!p-2 border-l-2 border-l-[var(--color-accent)]'
                  : '!p-2 border-l-2 border-l-blue-400'
              }
            >
              <div className="flex items-center gap-2 text-[10px] text-[var(--color-muted)] mb-1">
                <span className="font-semibold">{c.author}</span>
                {c.skill && <span className="font-mono">{c.skill}</span>}
                {c.verdict && (
                  <span>
                    {VERDICT_TO_LABEL[c.verdict].emoji}{' '}
                    {VERDICT_TO_LABEL[c.verdict].label}
                  </span>
                )}
                <span className="ml-auto font-mono">{c.timestamp}</span>
              </div>
              <div className="text-xs whitespace-pre-wrap">{c.text}</div>
            </Card>
          ))}
        </div>
        <div className="flex gap-2">
          <Textarea
            rows={2}
            value={convDraft}
            onChange={(e) => setConvDraft(e.target.value)}
            placeholder="Respondele a Claude…"
          />
          <Button
            variant="primary"
            onClick={handleAppendConv}
            disabled={!convDraft.trim()}
          >
            Enviar
          </Button>
        </div>
      </section>
    </div>
  );
}
