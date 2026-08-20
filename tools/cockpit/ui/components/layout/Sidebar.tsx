'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Map,
  ClipboardList,
  Compass,
  BookOpen,
  AlertTriangle,
  Wrench,
  History,
  TowerControl,
  Workflow,
  Rocket,
  type LucideIcon,
} from 'lucide-react';
import { cn } from '@/lib/cn';
import { viewAppliesTo, isPlatform, PLATFORM_LABEL, type CockpitView } from '@/lib/platform-context';
import { listCilBoard, getNavConfig } from '@/lib/api-client';
import { resolveSidebarNav } from '@/lib/nav';

interface NavItem {
  /** id estable para cockpit.config.yaml (nav configurable por workspace). */
  id: string;
  href: string;
  label: string;
  Icon: LucideIcon;
  /** Vista a la que mapea · usado para atenuar las que no aplican a platform. */
  view?: CockpitView;
}

// Vistas sistema-scoped (dependen de el sistema activo).
const NAV_ITEMS: NavItem[] = [
  { id: 'roadmap', href: '/roadmap', label: 'Roadmap', Icon: Map, view: 'roadmap' },
  { id: 'evolucion', href: '/evolucion', label: 'Evolución', Icon: History, view: 'evolucion' },
  { id: 'board', href: '/board', label: 'Backlog Board', Icon: ClipboardList, view: 'board' },
  { id: 'delivery', href: '/delivery', label: 'Delivery', Icon: Rocket }, // cockpit de delivery (F5 · DH-08)
  { id: 'map', href: '/map', label: 'Mapa del producto', Icon: Compass, view: 'map' },
  { id: 'drift', href: '/drift', label: 'Drift', Icon: AlertTriangle, view: 'drift' },
  { id: 'learnings', href: '/learnings', label: 'Learnings', Icon: BookOpen, view: 'learnings' },
];

// Vistas transversales (no dependen del sistema). Vista Negocio (CK-06/CK-07)
// dejó de ser un item de este menú — vive en el binario propio de Cockpit
// (Stage 4), no como tab dentro de DevHub. `harness` es global (core/CIL).
const CORE_NAV_ITEMS: NavItem[] = [
  { id: 'torre', href: '/torre', label: 'Torre de control', Icon: TowerControl },
  { id: 'proceso', href: '/proceso', label: 'Proceso', Icon: Workflow }, // descriptor de proceso (F4 · I-77)
  { id: 'harness', href: '/harness', label: 'Harness · CIL', Icon: Wrench },
];

function NavLink({
  item,
  active,
  dimmed,
  badge,
}: {
  item: NavItem;
  active: boolean;
  dimmed?: boolean;
  badge?: number;
}) {
  const { href, label, Icon } = item;
  return (
    <Link
      href={href}
      title={dimmed ? 'No aplica para Platform' : undefined}
      className={cn(
        'flex items-center gap-2 px-4 py-2 text-xs transition-colors',
        active
          ? 'bg-[var(--color-panel2)] text-[var(--color-text)] border-l-2 border-[var(--color-accent)]'
          : 'text-[var(--color-muted)] hover:bg-[var(--color-panel2)] hover:text-[var(--color-text)]',
        dimmed && 'opacity-40'
      )}
    >
      <Icon className="w-4 h-4" />
      <span>{label}</span>
      {badge != null && badge > 0 && (
        <span
          className="ml-auto text-[10px] font-semibold px-1.5 py-0.5 rounded-full bg-amber-900/40 text-amber-300 border border-amber-700/50"
          title={`${badge} items abiertos en el CIL (L1 harness + L3 deuda)`}
        >
          {badge}
        </span>
      )}
    </Link>
  );
}

export function Sidebar({ sistema, mode }: { sistema: string; mode?: string }) {
  const pathname = usePathname();
  // Badge en vivo del nav /harness = items abiertos del CIL (L1 + L3). Best-effort:
  // si el fetch falla (worktree sin archivos), el badge simplemente no aparece.
  const [cilOpen, setCilOpen] = useState<number | null>(null);
  // Nav configurable per-board (cockpit.config.yaml::nav del board activo). null = default.
  const [navIds, setNavIds] = useState<string[] | null>(null);
  useEffect(() => {
    let alive = true;
    listCilBoard()
      .then((b) => alive && setCilOpen(b.l1.open + b.l3.open))
      .catch(() => alive && setCilOpen(null));
    return () => {
      alive = false;
    };
  }, []);
  // Per-board (I-51): re-fetch al cambiar de board — cada uno trae el nav de SU repo.
  useEffect(() => {
    let alive = true;
    getNavConfig(sistema)
      .then((c) => alive && setNavIds(c.nav))
      .catch(() => alive && setNavIds(null));
    return () => {
      alive = false;
    };
  }, [sistema]);

  const { nav: navItems, core: coreNavItems } = resolveSidebarNav(
    NAV_ITEMS,
    CORE_NAV_ITEMS,
    navIds,
    mode
  );

  return (
    <aside className="bg-[var(--color-panel)] border-r border-[var(--color-border)] w-56 flex flex-col shrink-0">
      <div className="px-4 py-4 border-b border-[var(--color-border)]">
        <h1 className="text-base font-semibold flex items-center gap-2">
          <span aria-hidden="true">🧭</span> Cockpit · SDD
        </h1>
        <p className="text-[10px] text-[var(--color-muted)] mt-1">
          v0.6 · Spec-Driven Development
        </p>
      </div>
      <nav className="flex-1 py-3 flex flex-col">
        {navItems.map((item) => (
          <NavLink
            key={item.href}
            item={item}
            active={pathname === item.href}
            dimmed={item.view ? !viewAppliesTo(sistema, item.view) : false}
          />
        ))}
        {coreNavItems.length > 0 && (
          <div className="px-4 pt-4 pb-1 text-[9px] uppercase tracking-wider text-[var(--color-muted)]">
            Transversal · core
          </div>
        )}
        {coreNavItems.map((item) => (
          <NavLink
            key={item.href}
            item={item}
            active={pathname === item.href}
            badge={item.href === '/harness' ? cilOpen ?? undefined : undefined}
          />
        ))}
      </nav>
      <div className="px-4 py-3 border-t border-[var(--color-border)] text-[11px] text-[var(--color-muted)]">
        {isPlatform(sistema) ? (
          <span>
            Contexto:{' '}
            <span className="text-[var(--color-text)] font-medium">⬡ {PLATFORM_LABEL}</span>
            <span className="block text-[10px] mt-0.5 italic">
              stories transversales (owner /pm) — Board + Learnings
            </span>
          </span>
        ) : (
          <span>
            Sistema activo: <span className="text-[var(--color-text)] font-medium">{sistema}</span>
            <span className="block text-[10px] mt-0.5 italic">
              ⬡ stories platform/core → elegí «Platform · core» en el selector
            </span>
          </span>
        )}
      </div>
    </aside>
  );
}
