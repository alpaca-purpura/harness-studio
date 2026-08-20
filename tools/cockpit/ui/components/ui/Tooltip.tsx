'use client';

/**
 * Tooltip híbrido — 2026-05-28 · portal-based 2026-05-28.
 *
 * Variants:
 *   - inline (default): envuelve texto con border-dotted subtle
 *   - header: icon ⓘ adjacent al children (para section headers densos)
 *   - badge: el children ES el tooltipped element (sin envoltura visual extra)
 *
 * Accesible: aria-describedby + role="tooltip" + keyboard focus.
 *
 * Positioning: el bubble se renderiza en un PORTAL a document.body con
 * position:fixed, calculando coords desde el getBoundingClientRect() del
 * trigger. Esto lo hace FLOTAR encima de todo, escapando de cualquier
 * ancestro con overflow (ej. el Drawer con overflow-y-auto que antes lo
 * recortaba) o stacking context. Reposiciona en scroll/resize para seguir
 * al trigger, y hace clamp al viewport para no salirse de pantalla.
 */

import {
  useId,
  useState,
  useRef,
  useLayoutEffect,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react';
import { createPortal } from 'react-dom';
import { Info } from 'lucide-react';
import { cn } from '@/lib/cn';

type Position = NonNullable<TooltipProps['position']>;

interface TooltipProps {
  /** El texto que mostrará el tooltip al hover/focus */
  content: ReactNode;
  /** Contenido envuelto. En 'badge' mode, el children es el target del tooltip directamente. */
  children: ReactNode;
  /** Variante visual */
  variant?: 'inline' | 'header' | 'badge';
  /** Posición del bubble */
  position?: 'top' | 'bottom' | 'left' | 'right';
  /** Class opcional para el wrapper */
  className?: string;
}

/** Separación en px entre el trigger y el bubble */
const GAP = 6;
/** Margen mínimo al borde del viewport al hacer clamp */
const VIEWPORT_MARGIN = 8;

interface Coords {
  left: number;
  top: number;
  transform: string;
}

function computeCoords(rect: DOMRect, position: Position): Coords {
  const cx = rect.left + rect.width / 2;
  const cy = rect.top + rect.height / 2;
  switch (position) {
    case 'bottom':
      return { left: cx, top: rect.bottom + GAP, transform: 'translate(-50%, 0)' };
    case 'left':
      return { left: rect.left - GAP, top: cy, transform: 'translate(-100%, -50%)' };
    case 'right':
      return { left: rect.right + GAP, top: cy, transform: 'translate(0, -50%)' };
    case 'top':
    default:
      return { left: cx, top: rect.top - GAP, transform: 'translate(-50%, -100%)' };
  }
}

export function Tooltip({
  content,
  children,
  variant = 'inline',
  position = 'top',
  className,
}: TooltipProps) {
  const [show, setShow] = useState(false);
  const [coords, setCoords] = useState<Coords | null>(null);
  const [mounted, setMounted] = useState(false);
  const triggerRef = useRef<HTMLSpanElement>(null);
  const bubbleRef = useRef<HTMLSpanElement>(null);
  const id = useId();

  // Portal sólo tras montar en cliente (evita SSR mismatch)
  useEffect(() => {
    setMounted(true);
  }, []);

  const updatePosition = useCallback(() => {
    const el = triggerRef.current;
    if (!el) return;
    setCoords(computeCoords(el.getBoundingClientRect(), position));
  }, [position]);

  // Calcular posición al mostrar + seguir al trigger en scroll/resize.
  // scroll con capture=true para captar el scroll del Drawer interno (el
  // trigger se mueve aunque el bubble sea fixed → reposicionamos).
  useLayoutEffect(() => {
    if (!show) return;
    updatePosition();
    window.addEventListener('scroll', updatePosition, true);
    window.addEventListener('resize', updatePosition);
    return () => {
      window.removeEventListener('scroll', updatePosition, true);
      window.removeEventListener('resize', updatePosition);
    };
  }, [show, updatePosition]);

  // Clamp al viewport: si el bubble se sale por algún borde, lo corremos.
  // Converge en 1-2 pasadas (dx/dy → 0).
  useLayoutEffect(() => {
    if (!show || !coords || !bubbleRef.current) return;
    const b = bubbleRef.current.getBoundingClientRect();
    let dx = 0;
    let dy = 0;
    if (b.left < VIEWPORT_MARGIN) dx = VIEWPORT_MARGIN - b.left;
    else if (b.right > window.innerWidth - VIEWPORT_MARGIN)
      dx = window.innerWidth - VIEWPORT_MARGIN - b.right;
    if (b.top < VIEWPORT_MARGIN) dy = VIEWPORT_MARGIN - b.top;
    else if (b.bottom > window.innerHeight - VIEWPORT_MARGIN)
      dy = window.innerHeight - VIEWPORT_MARGIN - b.bottom;
    if (dx !== 0 || dy !== 0) {
      setCoords((c) => (c ? { ...c, left: c.left + dx, top: c.top + dy } : c));
    }
  }, [show, coords]);

  const handlers = {
    onMouseEnter: () => setShow(true),
    onMouseLeave: () => setShow(false),
    onFocus: () => setShow(true),
    onBlur: () => setShow(false),
  };

  const bubble =
    show && mounted && coords
      ? createPortal(
          <span
            ref={bubbleRef}
            role="tooltip"
            id={id}
            style={{ left: coords.left, top: coords.top, transform: coords.transform }}
            className={cn(
              'fixed z-[9999] px-2 py-1.5 rounded shadow-lg',
              'bg-[#1f2937] text-[#e5e7eb] border border-[#374151]',
              'text-[11px] leading-snug font-normal',
              'max-w-[300px] w-max whitespace-normal pointer-events-none'
            )}
          >
            {content}
          </span>,
          document.body
        )
      : null;

  if (variant === 'badge') {
    return (
      <span
        ref={triggerRef}
        className={cn('relative inline-block', className)}
        aria-describedby={show ? id : undefined}
        {...handlers}
      >
        {children}
        {bubble}
      </span>
    );
  }

  if (variant === 'header') {
    return (
      <span
        ref={triggerRef}
        className={cn('relative inline-flex items-center gap-1', className)}
      >
        {children}
        <button
          type="button"
          aria-label="Más info"
          aria-describedby={show ? id : undefined}
          className="inline-flex items-center justify-center text-[var(--color-muted)] hover:text-[var(--color-text)] focus:text-[var(--color-text)] focus:outline-none rounded"
          {...handlers}
        >
          <Info className="w-3 h-3" aria-hidden="true" />
        </button>
        {bubble}
      </span>
    );
  }

  // inline (default)
  return (
    <span
      ref={triggerRef}
      className={cn(
        'relative inline border-b border-dotted border-[var(--color-muted)] cursor-help',
        className
      )}
      tabIndex={0}
      aria-describedby={show ? id : undefined}
      {...handlers}
    >
      {children}
      {bubble}
    </span>
  );
}
