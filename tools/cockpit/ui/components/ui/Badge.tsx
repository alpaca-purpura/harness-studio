'use client';

import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/cn';
import type { ProcesoCategoria, StoryState } from '@/lib/types';
import { estadoDe } from '@/lib/proceso';
import { useProceso } from '@/components/providers/ProcesoProvider';

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  children: ReactNode;
}

export function Badge({ children, className, ...rest }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-block px-2 py-0.5 rounded text-[11px] font-medium',
        className
      )}
      {...rest}
    >
      {children}
    </span>
  );
}

// Theme de PRESENTACIÓN por id nativo (paleta histórica del board — RN-51): NO gobierna
// el ciclo. Estado fuera del theme → fallback por CATEGORÍA (contrato), abajo.
const STATE_THEME: Partial<Record<string, string>> = {
  idea: 'bg-[#1f2937] text-[#9ca3af]',
  refining: 'bg-[#1e3a8a] text-[#93c5fd]',
  refined: 'bg-[#5b21b6] text-[#ddd6fe]',
  ready: 'bg-[#3b3568] text-[#c4b5fd]',
  developing: 'bg-[#713f12] text-[#fbbf24]',
  developed: 'bg-[#365314] text-[#a3e635]',
  reviewing: 'bg-[#7c2d12] text-[#fb923c]',
  done: 'bg-[#14532d] text-[#86efac]',
  parked: 'bg-[#3f3f46] text-[#d4d4d8]',
  dropped: 'bg-[#450a0a] text-[#fca5a5]',
};

// Categorías FIJAS del contrato L0 → estilo (misma paleta que /proceso y la torre).
const CATEGORIA_CLASSES: Record<ProcesoCategoria, string> = {
  propuesto: 'bg-[#1f2937] text-[#9ca3af]',
  'en-progreso': 'bg-[#713f12] text-[#fbbf24]',
  completado: 'bg-[#14532d] text-[#86efac]',
  descartado: 'bg-[#450a0a] text-[#fca5a5]',
  pausado: 'bg-[#3f3f46] text-[#d4d4d8]',
};

export function StateBadge({ state, className }: { state: StoryState; className?: string }) {
  const proceso = useProceso();
  const categoria = proceso ? estadoDe(proceso, state)?.categoria : undefined;
  const cls =
    STATE_THEME[state] ??
    (categoria ? CATEGORIA_CLASSES[categoria] : undefined) ??
    'bg-[#27272a] text-[#a1a1aa]';
  return <Badge className={cn(cls, className)}>{state}</Badge>;
}

const RELEASE_STATUS_CLASSES: Record<string, string> = {
  shipped: 'bg-[#14532d] text-[#86efac]',
  in_progress: 'bg-[#1e3a8a] text-[#93c5fd]',
  ready_to_merge: 'bg-[#5b21b6] text-[#ddd6fe]',
  planning: 'bg-[#1f2937] text-[#9ca3af]',
  backlog: 'bg-[#3f3f46] text-[#d4d4d8]',
};

export function ReleaseStatusBadge({ status, className }: { status: string; className?: string }) {
  return (
    <Badge className={cn(RELEASE_STATUS_CLASSES[status] ?? RELEASE_STATUS_CLASSES.planning, className)}>
      {status.replace(/_/g, ' ')}
    </Badge>
  );
}

export function Pill({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        'inline-block px-1.5 py-0.5 rounded-sm text-[10px] font-medium tracking-wide',
        className
      )}
    >
      {children}
    </span>
  );
}
