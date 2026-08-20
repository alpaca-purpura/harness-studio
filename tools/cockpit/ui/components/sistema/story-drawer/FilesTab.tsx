'use client';

import { useEffect, useState } from 'react';
import toast from 'react-hot-toast';
import { ExternalLink, FileText, Folder } from 'lucide-react';
import { Spinner, EmptyState, ErrorBanner } from '@/components/ui/Spinner';
import { Button } from '@/components/ui/Button';
import { openInEditor } from '@/lib/api-client';

interface FilesTabProps {
  /** Path absoluto al directorio de la story */
  storyPath: string;
}

interface FileEntry {
  name: string;
  type: 'file' | 'dir';
  rel: string;
}

/**
 * Lista archivos del directorio de la story. Por ahora hardcodea
 * los artifacts canónicos esperados — un endpoint /api/file/list
 * sería ideal a futuro (Phase 6).
 */
const KNOWN_ARTIFACTS = [
  '01-spec.md',
  '02-design-agentic.md',
  '02-design-ui.md',
  '03-arch.md',
  '03-arch-be.md',
  '03-arch-fe.md',
  '03-arch-agentic.md',
  '04-validators.yaml',
  '05-guidelines.md',
  '06-tickets.yaml',
  '06-audit/gherkin-matrix.md',
  '07-merge.md',
  'checkpoint.md',
  'operator-input.md',
  'chris-input.md', // legacy F-1 (filename del template del kit)
  'CHECKPOINTS.md',
  'dispatch-plan.md',
  'mockups/',
  'observed-bugs/',
];

export function FilesTab({ storyPath }: FilesTabProps) {
  const [entries] = useState<FileEntry[]>(() =>
    KNOWN_ARTIFACTS.map((name) => ({
      name: name.replace(/\/$/, ''),
      type: name.endsWith('/') ? 'dir' : 'file',
      rel: name,
    }))
  );
  const [loading] = useState(false);
  const [error] = useState<string | null>(null);

  async function handleOpen(rel: string) {
    try {
      // Construir abs path desde storyPath + rel
      // openInEditor acepta abs paths dentro del workspace
      await openInEditor(`${storyPath}/${rel.replace(/\/$/, '')}`);
      toast.success('archivo abierto en editor');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  if (loading) return <Spinner />;
  if (error) return <ErrorBanner message={error} />;
  if (entries.length === 0) return <EmptyState>Sin archivos.</EmptyState>;

  return (
    <div className="space-y-2">
      <div className="text-[11px] text-[var(--color-muted)] mb-3">
        Archivos canónicos esperados en{' '}
        <span className="font-mono">{storyPath}</span>. Click ↗ para abrir en
        editor externo (xed).
      </div>
      {entries.map((e) => (
        <div
          key={e.rel}
          className="flex items-center gap-2 px-3 py-2 bg-[var(--color-panel2)] border border-[var(--color-border)] rounded text-xs hover:border-[#3a4358]"
        >
          {e.type === 'dir' ? (
            <Folder className="w-3.5 h-3.5 text-[var(--color-muted)]" />
          ) : (
            <FileText className="w-3.5 h-3.5 text-[var(--color-muted)]" />
          )}
          <span className="font-mono flex-1 truncate">{e.name}</span>
          <Button
            size="sm"
            variant="ghost"
            onClick={() => handleOpen(e.rel)}
            aria-label={`Abrir ${e.name}`}
          >
            <ExternalLink className="w-3 h-3" />
          </Button>
        </div>
      ))}
    </div>
  );
}
