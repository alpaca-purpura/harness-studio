'use client';

/**
 * MapStructure — la capa "arquitectura" fusionada DENTRO del mapa (2026-06-11,
 * pedido del operador: una sola vista con drill-down progresivo, no dos tabs
 * que leen el mismo SYSTEM-MAP).
 *
 *  · BoxProgress       — barra "áreas live / total" por caja (nivel 0)
 *  · BoxDetailDrawer   — click en caja → descripción + áreas + flujos que la
 *                        tocan + data entities que posee/consume (nivel 1)
 *  · CollapsibleSection + FlowCard + DataOwnershipTable — vistas transversales
 *                        globales al pie del mapa, colapsadas por default (nivel 2)
 *
 * FlowCard y DataOwnershipTable portados de ArchitectureView (tab retirado).
 */

import { useState, type ReactNode } from 'react';
import { cn } from '@/lib/cn';
import { Drawer } from '@/components/ui/Drawer';
import { Pill } from '@/components/ui/Badge';
import { EmptyState } from '@/components/ui/Spinner';
import type {
  SystemMap,
  CrossAgentFlow,
  AreaStatus,
} from '@/lib/types';
import type { BoxNode } from '@/lib/map-zones';

const STATUS_BADGES: Record<AreaStatus, { label: string; cls: string }> = {
  live: { label: 'live', cls: 'bg-[#14532d] text-[#86efac]' },
  beta: { label: 'beta', cls: 'bg-[#713f12] text-[#fbbf24]' },
  planned: { label: 'planned', cls: 'bg-[#1f2937] text-[#94a3b8]' },
  deprecated: { label: 'deprecated', cls: 'bg-[#450a0a] text-[#fca5a5]' },
};

// ────────────────────────────────────────────────────────────────────────────
// BoxProgress — "lo previsto que se va construyendo" visible de un vistazo
// ────────────────────────────────────────────────────────────────────────────

export function BoxProgress({ node }: { node: BoxNode }) {
  const total = node.areas.length;
  if (total === 0) return null;
  const live = node.areas.filter((a) => a.area.status === 'live').length;
  const pct = Math.round((live / total) * 100);
  return (
    <div
      className="flex items-center gap-1.5 mt-1"
      title={`${live} de ${total} áreas previstas están live (${pct}%). Las planned muestran su release destino.`}
    >
      <div className="flex-1 h-1 rounded-full bg-[var(--color-panel2)] overflow-hidden">
        <div
          className={cn('h-full rounded-full', pct === 100 ? 'bg-[#22c55e]' : 'bg-[var(--color-accent)]')}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-[9px] text-[var(--color-muted)] shrink-0">
        {live}/{total} áreas
      </span>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Flujos cross-agent — helpers + card (portado de ArchitectureView)
// ────────────────────────────────────────────────────────────────────────────

/** ¿El flujo toca esta caja? (trigger o cualquier action) */
export function flowTouchesBox(flow: CrossAgentFlow, boxId: string): boolean {
  if (flow.trigger.agent === boxId) return true;
  return flow.actions.some((a) => a.agent === boxId);
}

export function FlowCard({ flow }: { flow: CrossAgentFlow }) {
  const badge = STATUS_BADGES[flow.status];
  return (
    <div className="border border-[var(--color-border)] rounded p-3 bg-[var(--color-panel)] text-xs">
      <div className="flex items-center gap-2 mb-2 flex-wrap">
        <span className="font-mono text-[10px]">{flow.id}</span>
        <Pill className={`${badge.cls} text-[9px] py-0`}>{badge.label}</Pill>
        <Pill className="bg-[var(--color-panel2)] border border-[var(--color-border)] text-[9px] py-0">
          {flow.mechanism}
        </Pill>
        {flow.event_name && (
          <span className="text-[10px] text-[var(--color-muted)] font-mono">
            event: {flow.event_name}
          </span>
        )}
        {flow.endpoint && (
          <span className="text-[10px] text-[var(--color-muted)] font-mono">{flow.endpoint}</span>
        )}
        {flow.table && (
          <span className="text-[10px] text-[var(--color-muted)] font-mono">table: {flow.table}</span>
        )}
        {flow.target_release && (
          <span className="text-[9px] text-[var(--color-muted)]">→ {flow.target_release}</span>
        )}
      </div>
      <div className="text-[11px] space-y-1">
        <div>
          <strong>Trigger:</strong>{' '}
          <span className="font-mono text-[10px]">
            {flow.trigger.agent}.{flow.trigger.area}
          </span>{' '}
          · {flow.trigger.condition}
        </div>
        <ul className="list-disc list-inside text-[var(--color-muted)] space-y-0.5 mt-1">
          {flow.actions.map((action, i) => (
            <li key={i}>
              <strong className="text-[var(--color-text)] font-mono text-[10px]">
                {action.agent}.{action.area}:
              </strong>{' '}
              {action.what}
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// Data ownership — tabla (portada de ArchitectureView)
// ────────────────────────────────────────────────────────────────────────────

export function DataOwnershipTable({
  ownership,
}: {
  ownership: SystemMap['data_ownership'];
}) {
  const entries = Object.entries(ownership ?? {});
  if (entries.length === 0) return <EmptyState>Sin entities declaradas.</EmptyState>;
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-[11px]">
        <thead className="text-[10px] text-[var(--color-muted)] uppercase tracking-wide">
          <tr className="border-b border-[var(--color-border)]">
            <th className="text-left p-2">Entity</th>
            <th className="text-left p-2">Module</th>
            <th className="text-left p-2">Owner</th>
            <th className="text-left p-2">Consumida por</th>
            <th className="text-left p-2">PHI</th>
            <th className="text-left p-2">Descripción</th>
          </tr>
        </thead>
        <tbody>
          {entries.map(([name, ent]) => (
            <tr key={name} className="border-b border-[var(--color-border)] hover:bg-[var(--color-panel)]">
              <td className="p-2 font-medium">{name}</td>
              <td className="p-2 font-mono text-[10px]">{ent.owner_module}</td>
              <td className="p-2">{ent.owner_agent}</td>
              <td className="p-2 text-[10px] text-[var(--color-muted)]">
                {ent.consumed_by.join(', ')}
              </td>
              <td className="p-2">{ent.phi ? '🔒 PHI' : '—'}</td>
              <td className="p-2 text-[10px] text-[var(--color-muted)] max-w-xs leading-relaxed">
                {ent.description}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// CollapsibleSection — nivel 2: vistas transversales, cerradas por default
// ────────────────────────────────────────────────────────────────────────────

export function CollapsibleSection({
  title,
  subtitle,
  count,
  children,
}: {
  title: string;
  subtitle?: string;
  count?: number;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);
  return (
    <section className="border border-[var(--color-border)] rounded-lg bg-[var(--color-panel)]">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="w-full flex items-center gap-2 px-4 py-3 text-left"
        aria-expanded={open}
      >
        <span className="text-xs text-[var(--color-muted)]">{open ? '▾' : '▸'}</span>
        <span className="text-sm font-semibold">{title}</span>
        {count != null && (
          <span className="text-[10px] text-[var(--color-muted)]">({count})</span>
        )}
        {subtitle && (
          <span className="text-[10px] text-[var(--color-muted)] italic ml-2 hidden sm:inline">
            {subtitle}
          </span>
        )}
      </button>
      {open && <div className="px-4 pb-4">{children}</div>}
    </section>
  );
}

// ────────────────────────────────────────────────────────────────────────────
// BoxDetailDrawer — nivel 1: la caja en profundidad (estructura + conexiones)
// ────────────────────────────────────────────────────────────────────────────

export function BoxDetailDrawer({
  node,
  systemMap,
  zoneName,
  onClose,
  onCapClick,
}: {
  node: BoxNode | null;
  systemMap: SystemMap | null;
  zoneName?: string;
  onClose: () => void;
  /** Firma de useDrawer().openCap — abre el cap drawer (nivel 2 del drill-down). */
  onCapClick?: (module: string, slug: string) => void;
}) {
  const box = node?.box;
  const flows = (systemMap?.cross_agent_flows ?? []).filter(
    (f) => box && flowTouchesBox(f, box.id)
  );
  const owned = Object.entries(systemMap?.data_ownership ?? {}).filter(
    ([, e]) => box && e.owner_agent === box.id
  );
  const consumed = Object.entries(systemMap?.data_ownership ?? {}).filter(
    ([, e]) => box && e.owner_agent !== box.id && e.consumed_by.includes(box.id as never)
  );

  return (
    <Drawer
      open={node !== null}
      onClose={onClose}
      width={700}
      title={
        box ? (
          <div className="flex items-center gap-2 min-w-0">
            <span aria-hidden="true" className="text-lg">{box.emoji}</span>
            <span className="text-sm font-semibold truncate">{box.name}</span>
            {zoneName && (
              <span className="text-[10px] text-[var(--color-muted)] shrink-0">
                zona {zoneName}
              </span>
            )}
          </div>
        ) : null
      }
    >
      {node && box && (
        <div className="space-y-4 text-xs">
          {box.description && (
            <p className="text-[11px] text-[var(--color-muted)] leading-relaxed">
              {box.description}
            </p>
          )}

          <BoxProgress node={node} />

          {/* Áreas: lo previsto y su estado */}
          <div>
            <h3 className="text-xs font-semibold mb-2">
              Áreas funcionales{' '}
              <span className="text-[var(--color-muted)] font-normal">
                (lo previsto y cómo va)
              </span>
            </h3>
            <div className="space-y-1.5">
              {node.areas.map(({ area, fullId, caps }) => {
                const badge = STATUS_BADGES[area.status];
                return (
                  <div
                    key={fullId}
                    className="border border-[var(--color-border)] rounded px-2.5 py-1.5 bg-[var(--color-panel)]"
                  >
                    <div className="flex items-center gap-2 flex-wrap">
                      <span className="font-medium text-[11px]">{area.name}</span>
                      <span className={`text-[9px] px-1.5 py-0.5 rounded font-medium ${badge.cls}`}>
                        {badge.label}
                      </span>
                      {area.status === 'planned' && (
                        <span className="text-[9px] text-[var(--color-muted)]">
                          → {area.target_release ?? 'TBD'}
                        </span>
                      )}
                      <span className="text-[9px] text-[var(--color-muted)] ml-auto">
                        {caps.length} cap{caps.length !== 1 ? 's' : ''}
                      </span>
                    </div>
                    {area.description && (
                      <p className="text-[10px] text-[var(--color-muted)] italic mt-0.5 leading-relaxed">
                        {area.description}
                      </p>
                    )}
                    {caps.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1.5">
                        {caps.map((c) => (
                          <button
                            key={c.slug}
                            type="button"
                            onClick={() => onCapClick?.(c.module, c.slug)}
                            className="text-[9px] px-1.5 py-0.5 rounded border border-[var(--color-border)] bg-[var(--color-panel2)] hover:border-[var(--color-accent)] transition-colors font-mono"
                            title={c.user_facing_description ?? undefined}
                          >
                            {c.slug}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>

          {/* Flujos que tocan esta caja */}
          <div>
            <h3 className="text-xs font-semibold mb-2">
              Flujos cross-agent que la tocan{' '}
              <span className="text-[var(--color-muted)] font-normal">({flows.length})</span>
            </h3>
            {flows.length === 0 ? (
              <p className="text-[10px] text-[var(--color-muted)] italic">
                Ningún flujo declarado pasa por esta caja.
              </p>
            ) : (
              <div className="space-y-2">
                {flows.map((f) => (
                  <FlowCard key={f.id} flow={f} />
                ))}
              </div>
            )}
          </div>

          {/* Data entities */}
          {(owned.length > 0 || consumed.length > 0) && (
            <div>
              <h3 className="text-xs font-semibold mb-2">Data entities</h3>
              {owned.length > 0 && (
                <div className="mb-2">
                  <div className="text-[10px] text-[var(--color-muted)] uppercase tracking-wide mb-1">
                    Posee ({owned.length})
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {owned.map(([name, e]) => (
                      <span
                        key={name}
                        className="text-[10px] px-2 py-0.5 rounded border border-[var(--color-border)] bg-[var(--color-panel)] font-mono"
                        title={e.description}
                      >
                        {name} {e.phi ? '🔒' : ''}
                      </span>
                    ))}
                  </div>
                </div>
              )}
              {consumed.length > 0 && (
                <div>
                  <div className="text-[10px] text-[var(--color-muted)] uppercase tracking-wide mb-1">
                    Consume ({consumed.length})
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {consumed.map(([name, e]) => (
                      <span
                        key={name}
                        className="text-[10px] px-2 py-0.5 rounded border border-dashed border-[var(--color-border)] text-[var(--color-muted)] font-mono"
                        title={`owner: ${e.owner_agent} · ${e.description}`}
                      >
                        {name}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </Drawer>
  );
}
