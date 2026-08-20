'use client';

import { useCallback, useEffect, useState } from 'react';
import {
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import toast from 'react-hot-toast';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Panel } from '@/components/ui/Card';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { ReleaseCard } from './ReleaseCard';
import { NewReleaseModal } from '@/components/sistema/modals/NewReleaseModal';
import { EditReleaseModal } from '@/components/sistema/modals/EditReleaseModal';
import { MergeReleaseModal } from '@/components/sistema/modals/MergeReleaseModal';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchEvents } from '@/components/providers/FileWatchProvider';
import { viewAppliesTo } from '@/lib/platform-context';
import { NotApplicableForPlatform } from '@/components/platform/NotApplicableForPlatform';
import {
  listReleases,
  listStories,
  updateStory,
  type StoryWithArchive,
} from '@/lib/api-client';
import type { Release } from '@/lib/types';

export function RoadmapView() {
  const { sistema } = useSistema();
  const [releases, setReleases] = useState<Release[]>([]);
  const [stories, setStories] = useState<StoryWithArchive[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tab, setTab] = useState<'active' | 'shipped'>('active');
  const [newReleaseOpen, setNewReleaseOpen] = useState(false);
  const [editRelease, setEditRelease] = useState<Release | null>(null);
  const [mergeReleaseId, setMergeReleaseId] = useState<string | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } })
  );

  const load = useCallback(() => {
    if (!viewAppliesTo(sistema, 'roadmap')) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    Promise.all([listReleases(sistema), listStories(sistema)])
      .then(([r, s]) => {
        setReleases(r);
        setStories(s);
      })
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  // Live reload cuando archivos del SSoT cambien
  useFileWatchEvents((event) => {
    if (event.sistema && event.sistema !== sistema) return;
    if (event.docType === 'checkpoint' || event.docType === 'release') {
      load();
    }
  });

  async function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    if (!over) return;
    const storyId = active.id as string;
    const overId = over.id as string;
    if (!overId.startsWith('release:')) return;

    const targetReleaseId = overId.replace('release:', '');
    const story = stories.find((s) => s.story_id === storyId);
    if (!story) return;
    if (story.release === targetReleaseId) return;

    // Optimistic update
    setStories((prev) =>
      prev.map((s) =>
        s.story_id === storyId ? { ...s, release: targetReleaseId } : s
      )
    );

    try {
      await updateStory(storyId, sistema, { release: targetReleaseId });
      toast.success(`${storyId} → ${targetReleaseId}`);
    } catch (err) {
      toast.error(`No se pudo mover: ${(err as Error).message}`);
      load(); // revertir desde server
    }
  }

  if (!viewAppliesTo(sistema, 'roadmap')) return <NotApplicableForPlatform view="Roadmap" />;

  if (loading) {
    return (
      <div className="p-6 flex items-center gap-2 text-sm text-[var(--color-muted)]">
        <Spinner /> Cargando roadmap…
      </div>
    );
  }
  if (error) {
    return (
      <div className="p-6">
        <ErrorBanner message={error} />
      </div>
    );
  }

  const activeCount = releases.filter((r) => r.status !== 'shipped').length;
  const shippedCount = releases.filter((r) => r.status === 'shipped').length;

  const releasesShown = releases.filter((r) =>
    tab === 'shipped' ? r.status === 'shipped' : r.status !== 'shipped'
  );

  // En curso: orden cronológico (F0 → Fn). Historial: más reciente arriba.
  const sortedReleases = [...releasesShown].sort((a, b) =>
    tab === 'shipped' ? b.order - a.order : a.order - b.order
  );

  return (
    <div className="p-6">
      <header className="flex items-start justify-between mb-4 gap-4 flex-wrap">
        <div>
          <h1 className="text-lg font-semibold">
            <Tooltip content={TOOLTIPS.roadmap} variant="header">
              Roadmap · planeación temporal
            </Tooltip>
          </h1>
          <p className="text-[11px] text-[var(--color-muted)] mt-0.5 max-w-3xl">
            Arrastra stories entre{' '}
            <Tooltip content={TOOLTIPS.release_concept}>
              <span>releases</span>
            </Tooltip>
            {' '}mientras estén en estado{' '}
            <b>idea</b>, <b>refining</b> o <b>refined</b>. Una vez en{' '}
            <code>ready</code> o más, queda anclada (solo Claude la mueve). Al
            cerrar todas las stories de un release, se habilita merge.
          </p>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <div
            className="inline-flex rounded-md border border-[var(--color-border)] overflow-hidden"
            role="tablist"
            aria-label="Filtro de releases"
          >
            <button
              type="button"
              role="tab"
              aria-selected={tab === 'active'}
              onClick={() => setTab('active')}
              className={cn(
                'px-3 py-1 text-xs font-medium transition-colors',
                tab === 'active'
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'text-[var(--color-muted)] hover:text-[var(--color-text)]'
              )}
            >
              En curso · {activeCount}
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={tab === 'shipped'}
              onClick={() => setTab('shipped')}
              className={cn(
                'px-3 py-1 text-xs font-medium transition-colors border-l border-[var(--color-border)]',
                tab === 'shipped'
                  ? 'bg-[var(--color-accent)] text-white'
                  : 'text-[var(--color-muted)] hover:text-[var(--color-text)]'
              )}
            >
              Historial · {shippedCount}
            </button>
          </div>
          <Button variant="primary" onClick={() => setNewReleaseOpen(true)}>
            <Plus className="w-3 h-3" />
            Nuevo release
          </Button>
        </div>
      </header>

      {sortedReleases.length === 0 ? (
        <Panel className="p-8">
          <EmptyState>
            {tab === 'shipped'
              ? `Sin releases en el historial todavía para ${sistema}.`
              : `Sin releases en curso para ${sistema}. Crea el primero con el botón arriba.`}
          </EmptyState>
        </Panel>
      ) : (
        <DndContext sensors={sensors} onDragEnd={handleDragEnd}>
          <div className="space-y-3">
            {sortedReleases.map((r) => (
              <ReleaseCard
                key={r.release_id}
                release={r}
                stories={stories.filter((s) => s.release === r.release_id)}
                onMergeRequested={setMergeReleaseId}
                onEditRequested={setEditRelease}
              />
            ))}
          </div>
        </DndContext>
      )}

      <NewReleaseModal
        open={newReleaseOpen}
        onClose={() => setNewReleaseOpen(false)}
        onCreated={load}
      />
      <EditReleaseModal
        open={editRelease !== null}
        onClose={() => setEditRelease(null)}
        release={editRelease}
        onUpdated={load}
      />
      {mergeReleaseId && (
        <MergeReleaseModal
          open={mergeReleaseId !== null}
          onClose={() => setMergeReleaseId(null)}
          releaseId={mergeReleaseId}
          onMerged={load}
        />
      )}
    </div>
  );
}
