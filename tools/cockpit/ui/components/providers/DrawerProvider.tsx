'use client';

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

interface DrawerState {
  openStory: (storyId: string) => void;
  closeStory: () => void;
  storyId: string | null;
  openCap: (module: string, slug: string) => void;
  closeCap: () => void;
  capRef: { module: string; slug: string } | null;
}

const Ctx = createContext<DrawerState | null>(null);

export function DrawerProvider({ children }: { children: ReactNode }) {
  const [storyId, setStoryId] = useState<string | null>(null);
  const [capRef, setCapRef] = useState<{ module: string; slug: string } | null>(null);

  const openStory = useCallback((id: string) => setStoryId(id), []);
  const closeStory = useCallback(() => setStoryId(null), []);
  const openCap = useCallback((module: string, slug: string) => setCapRef({ module, slug }), []);
  const closeCap = useCallback(() => setCapRef(null), []);

  const value = useMemo(
    () => ({ storyId, openStory, closeStory, capRef, openCap, closeCap }),
    [storyId, openStory, closeStory, capRef, openCap, closeCap]
  );

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useDrawer(): DrawerState {
  const ctx = useContext(Ctx);
  if (!ctx) throw new Error('useDrawer fuera de DrawerProvider');
  return ctx;
}
