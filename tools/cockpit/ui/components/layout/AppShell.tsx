'use client';

import { Toaster } from 'react-hot-toast';
import { type ReactNode } from 'react';
import { useSistema } from '@/components/providers/SistemaProvider';
import { DrawerProvider } from '@/components/providers/DrawerProvider';
import { FileWatchProvider } from '@/components/providers/FileWatchProvider';
import { ProcesoProvider } from '@/components/providers/ProcesoProvider';
import { Sidebar } from './Sidebar';
import { Header } from './Header';
import { StoryDrawer } from '@/components/sistema/story-drawer/StoryDrawer';
import { CapDrawer } from '@/components/sistema/cap-drawer/CapDrawer';

// Stage 4 (CK-07): el shell ya no pinta un overview de "gap no-instrumentado"
// (sistemas compartido/externo/terciarizado sin board) — ese concepto es
// negocio de Cockpit (Vista Negocio, binario propio), no de DevHub. La lista
// de /api/sistemas que alimenta a DevHub solo trae sistemas reales.
function ShellInner({ children }: { children: ReactNode }) {
  const { sistema, mode } = useSistema();
  return (
    <div className="flex min-h-screen h-screen overflow-hidden">
      <Sidebar sistema={sistema} mode={mode} />
      <main className="flex-1 flex flex-col min-w-0">
        <Header />
        <div className="flex-1 overflow-auto">{children}</div>
      </main>
      <StoryDrawer />
      <CapDrawer />
    </div>
  );
}

export function AppShell({ children }: { children: ReactNode }) {
  return (
    <FileWatchProvider>
      <ProcesoProvider>
      <DrawerProvider>
        <ShellInner>{children}</ShellInner>
        <Toaster
          position="bottom-right"
          toastOptions={{
            style: {
              background: 'var(--color-panel2)',
              color: 'var(--color-text)',
              border: '1px solid var(--color-border)',
              fontSize: '12px',
            },
          }}
        />
      </DrawerProvider>
      </ProcesoProvider>
    </FileWatchProvider>
  );
}
