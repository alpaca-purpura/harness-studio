'use client';

import { Info } from 'lucide-react';

/**
 * Empty-state para las vistas que NO aplican a la pseudo-sistema "platform"
 * (roadmap/map/arquitectura/drift). Platform es solo-trazabilidad: tiene stories
 * + learnings platform-level, pero no releases/SYSTEM-MAP/capabilities. En vez de
 * ocultar la vista o fallar un fetch, explicamos por qué está vacía y a dónde ir.
 */
export function NotApplicableForPlatform({ view }: { view: string }) {
  return (
    <div className="p-6">
      <div className="max-w-lg mx-auto mt-12 rounded border border-[var(--color-border)] bg-[var(--color-panel)] border-l-2 border-l-amber-500/70 p-5 text-sm">
        <h2 className="font-semibold flex items-center gap-2 mb-2">
          <Info className="w-4 h-4 text-amber-400" />
          {view} no aplica para Platform
        </h2>
        <p className="text-[var(--color-muted)] leading-relaxed">
          <span className="text-amber-400 font-medium">Platform · core</span> es un
          contexto <b>solo de trazabilidad</b>: agrupa las stories y learnings de nivel
          plataforma (owner <span className="font-mono">/pm</span>), que no tienen
          releases, SYSTEM-MAP ni capabilities propias.
        </p>
        <p className="text-[var(--color-muted)] leading-relaxed mt-2">
          Usá el <b>Backlog Board</b> para ver sus stories (ej.{' '}
          <span className="font-mono">empleados-ia-auto-extension</span>), o cambiá a
          un sistema real para esta vista.
        </p>
      </div>
    </div>
  );
}
