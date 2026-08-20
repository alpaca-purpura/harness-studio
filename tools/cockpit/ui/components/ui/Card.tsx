'use client';

import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/cn';

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export function Card({ children, className, ...rest }: CardProps) {
  return (
    <div
      className={cn(
        'bg-[var(--color-panel2)] border border-[var(--color-border)] rounded-md p-3 transition-all duration-100',
        className
      )}
      {...rest}
    >
      {children}
    </div>
  );
}

export function Panel({ children, className, ...rest }: CardProps) {
  return (
    <div
      className={cn(
        'bg-[var(--color-panel)] border border-[var(--color-border)] rounded-lg',
        className
      )}
      {...rest}
    >
      {children}
    </div>
  );
}
