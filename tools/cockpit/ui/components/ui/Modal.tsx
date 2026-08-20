'use client';

import { useEffect, type ReactNode } from 'react';
import { X } from 'lucide-react';
import { cn } from '@/lib/cn';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: ReactNode;
  children: ReactNode;
  /** ancho máximo: sm=400 md=600 lg=900 xl=1200 full=95vw */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  /** ocultar botón cerrar header (e.g., para modales con CTA dedicados) */
  hideCloseButton?: boolean;
}

const SIZE: Record<NonNullable<ModalProps['size']>, string> = {
  sm: 'max-w-md',
  md: 'max-w-2xl',
  lg: 'max-w-4xl',
  xl: 'max-w-6xl',
  full: 'max-w-[95vw]',
};

export function Modal({
  open,
  onClose,
  title,
  children,
  size = 'md',
  hideCloseButton = false,
}: ModalProps) {
  useEffect(() => {
    if (!open) return;
    function handleEsc(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', handleEsc);
    return () => window.removeEventListener('keydown', handleEsc);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 bg-black/60 z-50 flex items-center justify-center p-4 overflow-y-auto"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      role="dialog"
      aria-modal="true"
    >
      <div
        className={cn(
          'w-full bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg shadow-2xl my-8',
          SIZE[size]
        )}
      >
        {(title !== undefined || !hideCloseButton) && (
          <header className="flex items-center justify-between px-5 py-3 border-b border-[var(--color-border)]">
            <div className="text-sm font-semibold">{title}</div>
            {!hideCloseButton && (
              <button
                type="button"
                onClick={onClose}
                className="p-1 rounded hover:bg-[var(--color-panel2)] text-[var(--color-muted)] hover:text-[var(--color-text)]"
                aria-label="Cerrar"
              >
                <X className="w-4 h-4" />
              </button>
            )}
          </header>
        )}
        <div className="px-5 py-4">{children}</div>
      </div>
    </div>
  );
}
