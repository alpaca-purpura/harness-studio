/**
 * CodeFilesSection · 📁 Archivos código asociados
 * Lee del CodeIndexReport y muestra files asociados a la cap.
 */

import toast from 'react-hot-toast';
import { Button } from '@/components/ui/Button';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { openInEditor } from '@/lib/api-client';
import { ExternalLink } from 'lucide-react';
import type { CodeIndexReport } from '@/lib/types';

interface CodeFilesSectionProps {
  capId: string;
  codeIndex: CodeIndexReport | null;
  /** Hint mostrado cuando el reporte no existe */
  hint?: string | null;
  /** Cuántos archivos top mostrar antes de truncar (default 8) */
  topN?: number;
}

export function CodeFilesSection({
  capId,
  codeIndex,
  hint,
  topN = 8,
}: CodeFilesSectionProps) {
  if (!codeIndex) {
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.code_files} variant="header">
            📁 Archivos código asociados
          </Tooltip>
        </h3>
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Sin reporte de code-index disponible.
          {hint && (
            <div className="mt-1 font-mono text-[10px] text-[var(--color-text)]">
              {hint}
            </div>
          )}
        </div>
      </section>
    );
  }

  const files = codeIndex.cap_to_files[capId] ?? [];

  if (files.length === 0) {
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.code_files} variant="header">
            📁 Archivos código asociados
          </Tooltip>
        </h3>
        <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
          Ningún archivo declara{' '}
          <Tooltip content={TOOLTIPS.cap_header}>
            <span className="font-mono">{`# cap: ${capId}`}</span>
          </Tooltip>{' '}
          en su header.
        </div>
      </section>
    );
  }

  async function handleOpen(filePath: string) {
    try {
      await openInEditor(filePath);
      toast.success('abierto');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  const visible = files.slice(0, topN);
  const rest = files.length - topN;

  return (
    <section>
      <h3 className="text-sm font-semibold mb-2">
        <Tooltip content={TOOLTIPS.code_files} variant="header">
          📁 Archivos código asociados ({files.length})
        </Tooltip>
      </h3>
      <ul className="space-y-1">
        {visible.map((f) => (
          <li
            key={f}
            className="flex items-center gap-1.5 text-[10px] group"
          >
            <span className="font-mono text-[var(--color-muted)] truncate flex-1">
              {f}
            </span>
            <Button
              size="sm"
              variant="ghost"
              onClick={() => handleOpen(f)}
              className="opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <ExternalLink className="w-3 h-3" />
            </Button>
          </li>
        ))}
        {rest > 0 && (
          <li className="text-[10px] text-[var(--color-muted)] italic pl-1">
            … y {rest} más
          </li>
        )}
      </ul>
    </section>
  );
}
