'use client';

import { MarkdownView } from './MarkdownView';
import { YamlEditor } from '@/components/ui/YamlEditor';
import { storyArtifactRel } from '@/lib/story-paths';
import { useSistema } from '@/components/providers/SistemaProvider';
import type { StoryState } from '@/lib/types';

interface ArtifactTabProps {
  storyPath: string;
  storyState: StoryState;
  isArchived: boolean;
  /** Lista ordenada de artifacts candidatos · primero encontrado se renderiza */
  candidates: string[];
  missingMessage?: string;
}

export function ArtifactTab({
  storyPath,
  storyState,
  isArchived,
  candidates,
  missingMessage,
}: ArtifactTabProps) {
  const { sistema } = useSistema();
  const rel = storyArtifactRel(storyPath, candidates[0], sistema);
  if (!rel) {
    return (
      <div className="text-xs text-[var(--color-muted)] italic">
        path inválido (story fuera del workspace).
      </div>
    );
  }

  // .yaml/.yml → YamlEditor con sintaxis + validación
  if (/\.(ya?ml)$/i.test(rel)) {
    return (
      <YamlEditor
        relPath={rel}
        storyState={storyState}
        isArchived={isArchived}
        missingMessage={missingMessage}
      />
    );
  }

  return (
    <MarkdownView
      relPath={rel}
      storyState={storyState}
      isArchived={isArchived}
      missingMessage={missingMessage}
    />
  );
}
