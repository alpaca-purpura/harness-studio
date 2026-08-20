'use client';

import { useCallback, useEffect, useState } from 'react';
import { Drawer } from '@/components/ui/Drawer';
import { Tabs, type TabItem } from '@/components/ui/Tabs';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { StateBadge } from '@/components/ui/Badge';
import { CheckpointTab } from './CheckpointTab';
import { ProcesoTab } from './ProcesoTab';
import { OperatorInputTab } from './OperatorInputTab';
import { ArtifactTab } from './ArtifactTab';
import { FilesTab } from './FilesTab';
import { GherkinScenariosTab } from './GherkinScenariosTab';
import { useDrawer } from '@/components/providers/DrawerProvider';
import { useSistema } from '@/components/providers/SistemaProvider';
import { getStory, type StoryWithArchive } from '@/lib/api-client';

export function StoryDrawer() {
  const { storyId, closeStory } = useDrawer();
  const { sistema } = useSistema();
  const [story, setStory] = useState<StoryWithArchive | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    if (!storyId) return;
    setLoading(true);
    setError(null);
    getStory(storyId, sistema)
      .then((s) => setStory(s as StoryWithArchive))
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [storyId, sistema]);

  useEffect(() => {
    if (storyId) {
      load();
    } else {
      setStory(null);
      setError(null);
    }
  }, [storyId, load]);

  // 5 tabs en forma de proceso (simplificación 2026-06-11, pedido del operador):
  // Proceso (qué pasa + qué me toca) · Definición (refining) · Técnico
  // (ready-package + checkpoint + files, agrupado) · Audit · Operador.
  const artifactTab = (candidates: string[], missingMessage: string) => (
    <ArtifactTab
      storyPath={story!.path}
      storyState={story!.state}
      isArchived={story!.is_archived ?? false}
      candidates={candidates}
      missingMessage={missingMessage}
    />
  );

  const tabs: TabItem[] = story
    ? [
        {
          id: 'proceso',
          label: <span title="Dónde está la story en el ciclo idea→done, los gates G/R, la DoD y tus acciones">🧭 Proceso</span>,
          content: <ProcesoTab story={story} onUpdated={load} />,
        },
        {
          id: 'definicion',
          label: <span title="Lo que define QUÉ se construye: spec funcional + diseño (se escriben en refining)">📐 Definición</span>,
          content: (
            <Tabs
              variant="pills"
              tabs={[
                {
                  id: 'spec',
                  label: <span title="01-spec.md — mapa funcional + gherkin (lo escribe /po-ux o /po)">📝 Spec</span>,
                  content: artifactTab(
                    ['01-spec.md'],
                    '01-spec.md aún no existe. Se crea cuando la story entra en refining via /po-ux o /po.'
                  ),
                },
                {
                  id: 'design',
                  label: <span title="02-design — diseño UI o conversacional (agentic)">🎨 Diseño</span>,
                  content: artifactTab(
                    ['02-design-ui.md', '02-design-agentic.md'],
                    '02-design-*.md aún no existe. Se crea cuando la story entra en refining (UI o agéntica).'
                  ),
                },
                {
                  id: 'escenarios',
                  label: <span title="Escenarios SC-N del spec con su estado de verificación (gherkin-matrix del auditor) + comando para correr cada test">🥒 Escenarios</span>,
                  content: <GherkinScenariosTab storyId={story.story_id} />,
                },
              ]}
            />
          ),
        },
        {
          id: 'tecnico',
          label: <span title="Lo que define CÓMO se construye: arquitectura, validators, tickets, checkpoint y archivos (lo produce /architect)">🔧 Técnico</span>,
          content: (
            <Tabs
              variant="pills"
              tabs={[
                {
                  id: 'arch',
                  label: <span title="03-arch.md — contratos + integration design">🏗 Arquitectura</span>,
                  content: artifactTab(
                    ['03-arch.md'],
                    '03-arch.md aún no existe. Se crea cuando /architect cierra el ready package.'
                  ),
                },
                {
                  id: 'validators',
                  label: <span title="04-validators.yaml — gates por naturaleza (técnica/funcional) + test plan">✓ Validators</span>,
                  content: artifactTab(
                    ['04-validators.yaml'],
                    '04-validators.yaml aún no existe. Se crea cuando /architect cierra el ready package.'
                  ),
                },
                {
                  id: 'tickets',
                  label: <span title="06-tickets.yaml — tickets con agente asignado explícito">🎟 Tickets</span>,
                  content: artifactTab(
                    ['06-tickets.yaml'],
                    '06-tickets.yaml aún no existe. Se crea cuando /architect cierra el ready package.'
                  ),
                },
                {
                  id: 'checkpoint',
                  label: <span title="checkpoint.md — el frontmatter completo + editor raw (escape hatch)">📋 Checkpoint</span>,
                  content: <CheckpointTab story={story} onUpdated={load} />,
                },
                {
                  id: 'files',
                  label: <span title="Todos los archivos del folder de la story">📂 Files</span>,
                  content: <FilesTab storyPath={story.path} />,
                },
              ]}
            />
          ),
        },
        {
          id: 'audit',
          label: <span title="gherkin-matrix del auditor — se genera cuando /auditor cierra Phase D">✅ Audit</span>,
          content: artifactTab(
            ['06-audit/gherkin-matrix.md'],
            'audit aún no se ejecutó. Se genera cuando /auditor cierra Phase D.'
          ),
        },
        {
          id: 'operator-input',
          label: <span title="Tu log de feedback/ratificaciones por ronda (operator-input.md)">💭 Operador</span>,
          content: <OperatorInputTab storyId={story.story_id} />,
        },
      ]
    : [];

  return (
    <Drawer
      open={storyId !== null}
      onClose={closeStory}
      title={
        story ? (
          <div className="flex items-center gap-2 min-w-0">
            <StateBadge state={story.state} />
            <span className="font-mono text-sm truncate">{story.story_id}</span>
            {story.phase === 'AWAIT_CHRIS_VERIFY' && !story.chris_verify?.signoff?.result && (
              <span
                className="text-[10px] px-1.5 py-0.5 rounded bg-[#1f1300] text-amber-400 border border-amber-800 shrink-0"
                title="Gate G: esperando la verificación live del operador"
              >
                ⏳ te espera
              </span>
            )}
            {story.dod_live_verified && (
              <span
                className="text-[10px] px-1.5 py-0.5 rounded bg-[#052e16] text-[#86efac] border border-green-900 shrink-0"
                title="DoD live-verify cumplida (acción real ejercida + efecto observado)"
              >
                ✓ live
              </span>
            )}
            {story.is_archived && (
              <span className="text-[10px] text-[var(--color-muted)] uppercase">
                archive
              </span>
            )}
          </div>
        ) : (
          <span className="text-sm text-[var(--color-muted)]">Story…</span>
        )
      }
      width={900}
    >
      {loading && (
        <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
          <Spinner /> Cargando story…
        </div>
      )}
      {error && <ErrorBanner message={error} />}
      {!loading && !error && story && (
        <>
          {story.goal && (
            <div className="mb-3 text-xs">
              <span className="text-[var(--color-muted)]">Goal:</span>{' '}
              {story.goal}
            </div>
          )}
          <Tabs tabs={tabs} />
        </>
      )}
      {!loading && !error && !story && storyId && (
        <EmptyState>Story sin datos.</EmptyState>
      )}
    </Drawer>
  );
}
