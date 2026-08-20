'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core';
import toast from 'react-hot-toast';
import { Search } from 'lucide-react';
import { Spinner, ErrorBanner } from '@/components/ui/Spinner';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { Select } from '@/components/ui/Select';
import { Badge } from '@/components/ui/Badge';
import { BoardColumn } from './BoardColumn';
import { NewStoryModal } from './NewStoryModal';
import { useDrawer } from '@/components/providers/DrawerProvider';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchEvents } from '@/components/providers/FileWatchProvider';
import {
  listReleases,
  listSessions,
  listStories,
  postTransition,
  type StoryWithArchive,
} from '@/lib/api-client';
import type { ActiveSession, Release, StoryState } from '@/lib/types';
import { useProceso } from '@/components/providers/ProcesoProvider';
import { isOperatorAllowed, labelTransicion, statesOrder, wipDe } from '@/lib/proceso';

// F6/RN-51: el ciclo ya no vive acá — orden de columnas, WIP caps, whitelist de drag
// y labels se derivan del descriptor de proceso (`useProceso` + lib/proceso.ts).

export function BoardView() {
  const { sistema, escribibles } = useSistema();
  // La autoridad sobre si se puede escribir es el SERVIDOR (GET /api/sistemas →
  // `escribibles`), no el nombre del sistema. Upstream lo deducía con
  // `isPlatform(sistema)` hardcodeado, lo que dejaba el board de harness-studio en
  // solo-lectura para siempre: ahí TODO el workspace es el pseudo-sistema platform.
  const soloLectura = !escribibles.includes(sistema);
  const proceso = useProceso();
  const [stories, setStories] = useState<StoryWithArchive[]>([]);
  const [releases, setReleases] = useState<Release[]>([]);
  const [sessionsByStory, setSessionsByStory] = useState<
    Record<string, ActiveSession>
  >({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [filterRelease, setFilterRelease] = useState('');
  const [filterType, setFilterType] = useState('');
  const [search, setSearch] = useState('');
  const [newStoryOpen, setNewStoryOpen] = useState(false);
  const { openStory } = useDrawer();
  const [showParked, setShowParked] = useState(false);

  const [draggingState, setDraggingState] = useState<StoryState | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } })
  );

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    Promise.all([listStories(sistema), listReleases(sistema)])
      .then(([s, r]) => {
        setStories(s);
        setReleases(r);
      })
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  // Build-claims vivos (ADR-009): `.session-locks/` es runtime gitignored → el
  // file-watcher de docs no los cubre. Poll liviano cada 5s + carga inicial.
  useEffect(() => {
    let cancelled = false;
    const refresh = () =>
      listSessions()
        .then((r) => {
          if (!cancelled) setSessionsByStory(r.by_story);
        })
        .catch(() => {
          /* sin .session-locks/ → mapa vacío, no es error */
        });
    refresh();
    const id = setInterval(refresh, 5000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, []);

  // Live reload cuando un checkpoint o release del sistema cambia
  useFileWatchEvents((event) => {
    if (event.sistema && event.sistema !== sistema) return;
    if (event.docType === 'checkpoint' || event.docType === 'release') {
      load();
    }
  });

  const filtered = useMemo(() => {
    return stories.filter((s) => {
      if (filterRelease && s.release !== filterRelease) return false;
      if (filterType && s.type !== filterType) return false;
      if (search) {
        const q = search.toLowerCase();
        const haystack = [
          s.story_id,
          s.goal ?? '',
          s.cap_target ?? '',
          s.module ?? '',
        ]
          .join(' ')
          .toLowerCase();
        if (!haystack.includes(q)) return false;
      }
      return true;
    });
  }, [stories, filterRelease, filterType, search]);

  const visibleStates = useMemo(() => {
    if (!proceso) return [];
    if (showParked) return statesOrder(proceso);
    // ocultar pausa/descarte POR CATEGORÍA (jamás por nombre de estado — D4).
    return proceso.estados
      .filter((e) => e.categoria !== 'pausado' && e.categoria !== 'descartado')
      .map((e) => e.id);
  }, [proceso, showParked]);

  function handleDragStart(event: DragStartEvent) {
    const story = event.active.data.current?.story as StoryWithArchive | undefined;
    setDraggingState(story?.state ?? null);
  }

  async function handleDragEnd(event: DragEndEvent) {
    setDraggingState(null);
    if (soloLectura) {
      toast('Este sistema es solo lectura · las transiciones las hace /pm.', {
        icon: '🔒',
      });
      return;
    }
    const { active, over } = event;
    if (!over) return;
    const overId = over.id as string;
    if (!overId.startsWith('state:')) return;

    const targetState = overId.replace('state:', '') as StoryState;
    const story = filtered.find((s) => s.story_id === active.id);
    if (!story) return;
    if (story.state === targetState) return;

    const allowed = proceso && isOperatorAllowed(proceso, story.state, targetState);

    if (!allowed) {
      toast.error(
        `Transición ${story.state} → ${targetState} la hace Claude vía skill.`
      );
      return;
    }

    if (allowed.requiere_razon) {
      toast(
        `${labelTransicion(allowed)} requiere razón. Abre la story para confirmar.`,
        { icon: '✋' }
      );
      return;
    }

    // Optimistic update
    setStories((prev) =>
      prev.map((s) =>
        s.story_id === story.story_id ? { ...s, state: targetState } : s
      )
    );

    try {
      await postTransition(sistema, story.story_id, targetState);
      toast.success(`${story.story_id}: ${story.state} → ${targetState}`);
    } catch (err) {
      toast.error(`No se pudo: ${(err as Error).message}`);
      load();
    }
  }

  if (loading || !proceso) {
    return (
      <div className="p-6 flex items-center gap-2 text-sm text-[var(--color-muted)]">
        <Spinner /> Cargando board…
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

  // Chips de WIP derivados: uno por estado con límite declarado (RN-48).
  const wipChips = proceso.estados
    .filter((e) => wipDe(proceso, e.id) !== undefined)
    .map((e) => ({
      id: e.id,
      cap: wipDe(proceso, e.id)!,
      count: filtered.filter((s) => s.state === e.id).length,
    }));

  return (
    <div className="p-6">
      <header className="mb-4">
        <h1 className="text-lg font-semibold flex items-center gap-2">
          Backlog Board · {proceso.estados.length} estados · {proceso.descriptor.nombre}
          {soloLectura && (
            <Badge className="bg-amber-900/30 text-amber-300 border border-amber-700/50">
              CORE · solo lectura
            </Badge>
          )}
        </h1>
        <p className="text-[11px] text-[var(--color-muted)] italic mt-1">
          {soloLectura ? (
            <>
              Stories platform-level (owner <span className="font-mono">/pm</span>).
              Solo lectura: las transiciones se hacen vía skill, no acá.
            </>
          ) : (
            <>
              Arrastra solo las transiciones del operador (el descriptor las
              declara). El resto las mueven los arneses dueños del proceso.
            </>
          )}
        </p>
      </header>

      <div className="flex items-center gap-2 mb-3 flex-wrap">
        <Select
          value={filterRelease}
          onChange={(e) => setFilterRelease(e.target.value)}
          className="!w-auto !py-1"
        >
          <option value="">release: todos</option>
          {releases.map((r) => (
            <option key={r.release_id} value={r.release_id}>
              {r.release_id} · {r.name}
            </option>
          ))}
        </Select>
        <Select
          value={filterType}
          onChange={(e) => setFilterType(e.target.value)}
          className="!w-auto !py-1"
        >
          <option value="">tipo: todos</option>
          <option value="ui">ui</option>
          <option value="service">service</option>
          <option value="agentic">agentic</option>
          <option value="tech">tech</option>
          <option value="func">func</option>
        </Select>
        <label className="flex items-center gap-1.5 text-xs">
          <input
            type="checkbox"
            className="!w-auto"
            checked={showParked}
            onChange={(e) => setShowParked(e.target.checked)}
          />
          mostrar pausadas + descartadas
        </label>
        <div className="flex items-center gap-1.5 flex-1 max-w-xs">
          <Search className="w-3.5 h-3.5 text-[var(--color-muted)]" />
          <Input
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Buscar…"
            className="!py-1"
          />
        </div>
        <Button
          variant="primary"
          size="sm"
          onClick={() => setNewStoryOpen(true)}
          title="Capturar una idea desde cero: nace en estado idea con checkpoint + operator-input"
        >
          + Nueva story
        </Button>
        <div className="flex items-center gap-2 text-xs">
          {wipChips.map((c) => (
            <Badge
              key={c.id}
              className={
                c.count > c.cap
                  ? 'bg-orange-900/40 text-orange-300'
                  : 'bg-[var(--color-panel2)] text-[var(--color-muted)]'
              }
            >
              WIP {c.id} {c.count}/{c.cap}
            </Badge>
          ))}
        </div>
      </div>

      {Object.keys(sessionsByStory).length > 0 && (
        <div className="flex items-center gap-2 mb-3 flex-wrap text-[11px]">
          <span className="text-[var(--color-muted)]">🔨 Construyendo ahora:</span>
          {Object.values(sessionsByStory).map((s) => (
            <Badge
              key={s.storyId}
              className="bg-amber-900/30 text-amber-300 border border-amber-700/50 font-mono"
            >
              {s.lane ?? `pid${s.pid}`} → {s.storyId}
            </Badge>
          ))}
        </div>
      )}

      <DndContext
        sensors={sensors}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <div className="flex gap-3 overflow-x-auto pb-4">
          {visibleStates.map((state) => (
            <BoardColumn
              key={state}
              state={state as StoryState}
              stories={filtered.filter((s) => s.state === state)}
              draggingFromState={draggingState}
              wipCap={wipDe(proceso, state)}
              sessions={sessionsByStory}
            />
          ))}
        </div>
      </DndContext>

      <NewStoryModal
        open={newStoryOpen}
        sistema={sistema}
        releases={releases}
        onClose={() => setNewStoryOpen(false)}
        onCreated={(storyId) => {
          setNewStoryOpen(false);
          load();
          openStory(storyId);
        }}
      />
    </div>
  );
}
