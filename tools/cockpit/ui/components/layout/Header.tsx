'use client';

import { SistemaSwitcher } from './SistemaSwitcher';
import { WatchingIndicator } from './WatchingIndicator';

export function Header() {
  return (
    <header className="bg-[var(--color-panel)] border-b border-[var(--color-border)] px-6 py-3 flex items-center justify-between gap-4 shrink-0">
      <div className="flex items-center gap-3">
        <SistemaSwitcher />
      </div>
      <div className="flex items-center gap-3 text-[11px] text-[var(--color-muted)]">
        <WatchingIndicator />
        <span className="hidden sm:inline">workspace</span>
        <span>operador</span>
      </div>
    </header>
  );
}
