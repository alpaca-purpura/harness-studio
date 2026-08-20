'use client';

import { useEffect, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/cn';

interface DrawerProps {
  open: boolean;
  onClose: () => void;
  title?: ReactNode;
  children: ReactNode;
  /** Ancho en px (default 800) */
  width?: number;
}

export function Drawer({ open, onClose, title, children, width = 800 }: DrawerProps) {
  useEffect(() => {
    if (!open) return;
    function handleEsc(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [open, onClose]);

  return (
    <>
      {open && (
        <div
          className="fixed inset-0 bg-black/50 z-40"
          onClick={onClose}
          aria-hidden="true"
        />
      )}
      <aside
        role="dialog"
        aria-modal="true"
        style={{ width: `${width}px`, maxWidth: '95vw' }}
        className={cn(
          'fixed top-0 right-0 h-screen bg-[var(--color-panel)] border-l border-[var(--color-border)] shadow-2xl z-50 transition-transform duration-200 ease-out overflow-y-auto',
          open ? 'translate-x-0' : 'translate-x-full'
        )}
      >
        {title !== undefined && (
          <header className="sticky top-0 z-10 bg-[var(--color-panel)] border-b border-[var(--color-border)] px-4 py-3 flex items-center justify-between gap-2">
            <div className="min-w-0 flex-1">{title}</div>
            <button
              type="button"
              onClick={onClose}
              className="p-1 rounded hover:bg-[var(--color-panel2)] text-[var(--color-muted)] hover:text-[var(--color-text)]"
              aria-label="Cerrar"
            >
              <X className="w-4 h-4" />
            </button>
          </header>
        )}
        <div className="p-4">{children}</div>
      </aside>
    </>
  );
}
