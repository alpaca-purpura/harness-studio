'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { cn } from '@/lib/cn';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Card } from '@/components/ui/Card';
import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { useDrawer } from '@/components/providers/DrawerProvider';
import { useSistema } from '@/components/providers/SistemaProvider';
import { useFileWatchEvents } from '@/components/providers/FileWatchProvider';
import { viewAppliesTo } from '@/lib/platform-context';
import { NotApplicableForPlatform } from '@/components/platform/NotApplicableForPlatform';
import { ProductHealthBanner } from './ProductHealthBanner';
import { agentMetaOf } from '@/lib/agent-meta';
import {
  listCapabilities,
  getSystemMap,
  openInEditor,
  getCapabilityStatus,
  getValueStreamStages,
} from '@/lib/api-client';
import type {
  Capability,
  SystemMap,
  SystemMapZone,
  AreaStatus,
  ComputedStatusReport,
  CapStatusComputed,
} from '@/lib/types';
import { getStatusBadge } from '@/lib/types';
import {
  BoxProgress,
  BoxDetailDrawer,
  CollapsibleSection,
  FlowCard,
  DataOwnershipTable,
} from './MapStructure';
import {
  buildZoneTree,
  buildProcessLens,
  capsByFunctionalArea,
  findOrphanCaps,
  findSupervisor,
  type BoxNode,
  type ZoneNode,
  type ProcessStageNode,
  type ValueStreamStage,
} from '@/lib/map-zones';

/** Lentes del mapa: por trabajador/superficie (zonas) o por value-stream (proceso). */
type MapLens = 'trabajadores' | 'proceso';

const STATUS_CLASSES: Record<string, string> = {
  live: 'bg-[#14532d] text-[#86efac]',
  beta: 'bg-[#713f12] text-[#fbbf24]',
  deprecated: 'bg-[#3f3f46] text-[#d4d4d8]',
  sunset: 'bg-[#450a0a] text-[#fca5a5]',
};

const STATUS_BADGES: Record<AreaStatus, { label: string; cls: string }> = {
  live: { label: 'live', cls: 'bg-[#14532d] text-[#86efac]' },
  beta: { label: 'beta', cls: 'bg-[#713f12] text-[#fbbf24]' },
  planned: { label: 'planned', cls: 'bg-[#1f2937] text-[#94a3b8]' },
  deprecated: { label: 'deprecated', cls: 'bg-[#450a0a] text-[#fca5a5]' },
};

export function MapView() {
  const { sistema } = useSistema();
  const { openCap } = useDrawer();
  // Nivel 1 del drill-down: caja seleccionada → drawer con estructura + conexiones.
  const [selectedBox, setSelectedBox] = useState<{ node: BoxNode; zoneName?: string } | null>(null);
  const [caps, setCaps] = useState<Capability[]>([]);
  const [systemMap, setSystemMap] = useState<SystemMap | null>(null);
  const [statusReport, setStatusReport] = useState<ComputedStatusReport | null>(null);
  const [statusHint, setStatusHint] = useState<string | null>(null);
  const [seamStages, setSeamStages] = useState<ValueStreamStage[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showLive, setShowLive] = useState(true);
  const [showDraft, setShowDraft] = useState(false);
  const [showInfra, setShowInfra] = useState(false);
  const [showPlanned, setShowPlanned] = useState(true);
  const [lens, setLens] = useState<MapLens>('trabajadores');

  // R3.2 · filtros nuevos (search natural + onboarding rol + solo poblados v3.2)
  const [searchTerm, setSearchTerm] = useState('');
  const [roleFilter, setRoleFilter] = useState<string>('all');
  const [showOnlyPopulated, setShowOnlyPopulated] = useState(false);

  const load = useCallback(() => {
    if (!viewAppliesTo(sistema, 'map')) {
      setLoading(false);
      return;
    }
    setLoading(true);
    setError(null);
    Promise.all([
      listCapabilities(sistema),
      getSystemMap(sistema),
      getCapabilityStatus(sistema),
      getValueStreamStages(sistema),
    ])
      .then(([capsData, mapData, statusData, stagesData]) => {
        setCaps(capsData);
        setSystemMap(mapData);
        setStatusReport(statusData.status);
        setStatusHint(statusData.hint ?? null);
        setSeamStages(stagesData);
      })
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [sistema]);

  useEffect(() => {
    load();
  }, [load]);

  // Live reload cuando un capability YAML cambia
  // (system_map no está en el enum de docType del watcher, pero capability reload también refresca el mapa)
  useFileWatchEvents((event) => {
    if (event.sistema && event.sistema !== sistema) return;
    if (event.docType === 'capability') {
      load();
    }
  });

  const filtered = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();
    return caps.filter((c) => {
      if (c.status === 'live' && !showLive) return false;
      if (c.status === 'beta' && !showDraft) return false;
      if (c.status === 'deprecated' || c.status === 'sunset') return false;
      // NOTA: las caps infra (user_visible: false) NO se filtran acá — `showInfra`
      // controla el COLAPSO VISUAL de la zona Infraestructura (ZoneBlock), no el
      // filtrado de datos. Así el teaser colapsado muestra el conteo REAL de caps
      // (antes mostraba "0 caps" engañoso porque se filtraban antes de contar).

      // R3.2 · "solo poblados v3.2"
      if (showOnlyPopulated && !(c.scenarios && c.scenarios.length > 0)) return false;

      // R3.2 · filtro por rol (onboarding)
      if (roleFilter !== 'all') {
        const eps = c.access?.entry_points ?? [];
        const hasRole = eps.some((ep) =>
          (ep.requires_role ?? []).includes(roleFilter)
        );
        if (!hasRole) return false;
      }

      // R3.2 · search natural
      if (term) {
        const haystacks: string[] = [
          c.user_facing_name ?? '',
          c.user_facing_description ?? '',
          c.module ?? '',
          c.slug ?? '',
          c.functional_area ?? '',
          ...(c.scenarios ?? []).flatMap((s) => [
            s.name ?? '',
            s.given ?? '',
            s.when ?? '',
            s.then ?? '',
            ...(s.edge_cases ?? []),
          ]),
          ...(c.business_rules ?? []).map((r) => r.rule ?? ''),
        ];
        if (!haystacks.some((h) => h.toLowerCase().includes(term))) return false;
      }

      return true;
    });
  }, [caps, showLive, showDraft, showOnlyPopulated, roleFilter, searchTerm]);

  // Árbol zona → caja → área (SYSTEM-MAP v2.0). null si el map no trae `zones`.
  const zoneTree = useMemo<ZoneNode[] | null>(() => {
    if (!systemMap?.zones || systemMap.zones.length === 0) return null;
    return buildZoneTree(systemMap, capsByFunctionalArea(filtered));
  }, [systemMap, filtered]);

  // Lente "proceso": cajas de la zona Agentes ordenadas por el value-stream del
  // seam (`project.config.yaml`); fallback genérico si el slot está vacío.
  const processLens = useMemo<ProcessStageNode[] | null>(
    () => (zoneTree ? buildProcessLens(zoneTree, seamStages ?? undefined) : null),
    [zoneTree, seamStages]
  );

  // Roles disponibles para el filtro "onboarding por rol": derivados del data
  // presente (access.entry_points[].requires_role de las caps) — sin lista
  // hardcodeada (F-4). Si no hay roles en el data, el select no se muestra.
  const availableRoles = useMemo<string[]>(() => {
    const roles = new Set<string>();
    for (const c of caps) {
      for (const ep of c.access?.entry_points ?? []) {
        for (const r of ep.requires_role ?? []) roles.add(r);
      }
    }
    return Array.from(roles).sort();
  }, [caps]);

  // Supervisora (Valeria) — sidebar, no caja de valor.
  const supervisor = useMemo(() => findSupervisor(systemMap?.agents), [systemMap]);

  // Caps sin caja/área conocida (huérfanas · anti-isla)
  const orphans = useMemo(
    () => (zoneTree ? findOrphanCaps(filtered, zoneTree) : []),
    [zoneTree, filtered]
  );

  if (!viewAppliesTo(sistema, 'map')) return <NotApplicableForPlatform view="Mapa Implementado" />;

  if (loading) {
    return (
      <div className="p-6 flex items-center gap-2 text-sm text-[var(--color-muted)]">
        <Spinner /> Cargando capabilities…
      </div>
    );
  }
  if (error) {
    return (
      <div className="p-6">
        <ErrorBanner message={error} />
      </div>
    );
  }

  return (
    <div className="p-6">
      {/* Salud de Producto · vista honesta del bosque (lee summary del JSON live) */}
      <ProductHealthBanner report={statusReport} />

      <header className="mb-4 space-y-2">
        <div className="flex items-start justify-between gap-4 flex-wrap">
          <div>
            <h1 className="text-lg font-semibold">
              <Tooltip content={TOOLTIPS.mapa_implementado} variant="header">
                Mapa del producto
              </Tooltip>
            </h1>
            <p className="text-[11px] text-[var(--color-muted)] italic mt-1">
              La estructura prevista y lo construido sobre ella. Click en una caja
              para ver su detalle (áreas, flujos, datos). Las developing viven en el Backlog Board.
            </p>
            {systemMap && (
              <div className="text-[11px] text-[var(--color-muted)] mt-1">
                Lee skeleton de{' '}
                <Tooltip content={TOOLTIPS.system_map}>
                  <button
                    onClick={() => openInEditor(systemMap._path ?? '')}
                    className="font-mono text-[var(--color-accent)] hover:underline"
                  >
                    SYSTEM-MAP.yaml
                  </button>
                </Tooltip>
                {' · '}
                {systemMap.metadata.total_functional_areas} áreas · {systemMap.metadata.total_cross_agent_flows} flujos cross-agent · {Object.keys(systemMap.data_ownership ?? {}).length} entities
              </div>
            )}
          </div>
          <div className="text-[var(--color-muted)] text-xs shrink-0">
            total: {caps.length} · mostrando: {filtered.length}
          </div>
        </div>

        {/* Filtros R3.2 · search + onboarding rol + solo poblados + status checkboxes */}
        <div className="flex flex-wrap gap-3 items-center text-xs">
          <input
            type="text"
            placeholder="Buscar por scenarios, reglas, descripción…"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="px-3 py-1.5 rounded border border-[var(--color-border)] bg-[var(--color-panel)] text-[var(--color-text)] flex-1 min-w-[200px] max-w-[400px]"
            aria-label="Buscar capabilities"
          />
          {availableRoles.length > 0 && (
            <select
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value)}
              className="px-3 py-1.5 rounded border border-[var(--color-border)] bg-[var(--color-panel)] text-[var(--color-text)]"
              aria-label="Modo onboarding por rol"
            >
              <option value="all">Todos los roles</option>
              {availableRoles.map((r) => (
                <option key={r} value={r}>
                  Modo onboarding: {r}
                </option>
              ))}
            </select>
          )}
          <label className="flex items-center gap-1.5">
            <input
              type="checkbox"
              className="!w-auto"
              checked={showOnlyPopulated}
              onChange={(e) => setShowOnlyPopulated(e.target.checked)}
            />
            <Tooltip content={TOOLTIPS.v3_2_badge}>
              <span>solo poblados v3.2</span>
            </Tooltip>
          </label>
        </div>
        <div className="flex items-center gap-3 text-xs flex-wrap">
          <label className="flex items-center gap-1.5">
            <input
              type="checkbox"
              className="!w-auto"
              checked={showLive}
              onChange={(e) => setShowLive(e.target.checked)}
            />
            live ✓ ({filtered.filter((c) => c.status === 'live').length})
          </label>
          <label className="flex items-center gap-1.5">
            <input
              type="checkbox"
              className="!w-auto"
              checked={showDraft}
              onChange={(e) => setShowDraft(e.target.checked)}
            />
            beta ⚪ ({caps.filter((c) => c.status === 'beta').length})
          </label>
          <label className="flex items-center gap-1.5">
            <input
              type="checkbox"
              className="!w-auto"
              checked={showInfra}
              onChange={(e) => setShowInfra(e.target.checked)}
            />
            <Tooltip content={TOOLTIPS.infra_role}>
              <span>infra 🔧</span>
            </Tooltip>
            {' '}
            ({caps.filter((c) => c.user_visible === false).length})
          </label>
          <label className="flex items-center gap-1.5">
            <input
              type="checkbox"
              className="!w-auto"
              checked={showPlanned}
              onChange={(e) => setShowPlanned(e.target.checked)}
            />
            <Tooltip content={TOOLTIPS.planned_status}>
              <span>planned 📋</span>
            </Tooltip>
          </label>
        </div>

        {/* Lente del mapa: trabajadores (zonas) · proceso (value-stream) */}
        {zoneTree && (
          <div className="flex items-center gap-2 text-xs flex-wrap">
            <span className="text-[var(--color-muted)]">Lente:</span>
            <div className="inline-flex rounded border border-[var(--color-border)] overflow-hidden">
              {(
                [
                  { id: 'trabajadores', label: '👥 Trabajadores', hint: 'Por zona y caja (quién opera qué)' },
                  { id: 'proceso', label: '🔄 Proceso', hint: 'Por etapas del value-stream (slot value_stream del seam)' },
                ] as const
              ).map((opt) => (
                <button
                  key={opt.id}
                  type="button"
                  title={opt.hint}
                  aria-pressed={lens === opt.id}
                  onClick={() => setLens(opt.id)}
                  className={cn(
                    'px-3 py-1 transition-colors',
                    lens === opt.id
                      ? 'bg-[var(--color-accent)] text-black font-medium'
                      : 'bg-[var(--color-panel)] text-[var(--color-muted)] hover:text-[var(--color-fg)]'
                  )}
                >
                  {opt.label}
                </button>
              ))}
            </div>
            {supervisor && (
              <span className="text-[11px] text-[var(--color-muted)] italic ml-1">
                {supervisor.emoji} {supervisor.name} = supervisor (sidebar · orquesta, no es caja de valor)
              </span>
            )}
          </div>
        )}
      </header>

      {!statusReport && statusHint && (
        <div className="mb-3 text-[11px] text-amber-400 px-2 py-1 border border-amber-700 rounded flex items-center gap-1.5">
          <span aria-hidden="true">⚠️</span>
          <span>
            Sin datos de verificación.{' '}
            <span className="font-mono">{statusHint}</span>
          </span>
        </div>
      )}

      {zoneTree ? (
        lens === 'proceso' && processLens ? (
          <ProcessLensView
            stages={processLens}
            zoneTree={zoneTree}
            showPlanned={showPlanned}
            showInfra={showInfra}
            statusReport={statusReport}
            onBoxOpen={(node, zoneName) => setSelectedBox({ node, zoneName })}
          />
        ) : (
          <ZoneLensView
            zoneTree={zoneTree}
            showPlanned={showPlanned}
            showInfra={showInfra}
            statusReport={statusReport}
            onBoxOpen={(node, zoneName) => setSelectedBox({ node, zoneName })}
          />
        )
      ) : (
        // Fallback si SYSTEM-MAP no cargó o no trae `zones`: vista legacy por agent_owner
        <LegacyFallbackView filtered={filtered} caps={caps} showInfra={showInfra} statusReport={statusReport} />
      )}

      {orphans.length > 0 && (
        <Card className="!p-4 mt-4 border-red-700">
          <header className="flex items-baseline gap-2 mb-3">
            <span aria-hidden="true">⚠️</span>
            <h2 className="text-sm font-semibold text-red-400">
              Capabilities sin agent_owner declarado
            </h2>
            <span className="text-[10px] text-[var(--color-muted)]">
              ({orphans.length} caps · v3 schema incompleto)
            </span>
          </header>
          <div className="text-[11px] text-[var(--color-muted)] mb-2">
            Estos caps necesitan refining para declarar `agent_owner` + `functional_area` (schema v3).
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
            {orphans.map((c, i) => (
              <CapItem key={capKey(c, i)} cap={c} statusReport={statusReport} />
            ))}
          </div>
        </Card>
      )}

      {/* ── Nivel 2 · estructura transversal (ex-tab Arquitectura, fusionado) ── */}
      {systemMap && (
        <div className="space-y-3 mt-6">
          <CollapsibleSection
            title="🔄 Flujos cross-agent"
            count={systemMap.cross_agent_flows?.length ?? 0}
            subtitle="cómo se conectan las cajas entre sí (triggers → acciones)"
          >
            {(systemMap.cross_agent_flows ?? []).length === 0 ? (
              <EmptyState>Sin flujos declarados.</EmptyState>
            ) : (
              <div className="space-y-3">
                {systemMap.cross_agent_flows.map((flow) => (
                  <FlowCard key={flow.id} flow={flow} />
                ))}
              </div>
            )}
          </CollapsibleSection>
          <CollapsibleSection
            title="📦 Data ownership"
            count={Object.keys(systemMap.data_ownership ?? {}).length}
            subtitle="qué caja posee cada entity de datos y quién la consume"
          >
            <DataOwnershipTable ownership={systemMap.data_ownership} />
          </CollapsibleSection>
        </div>
      )}

      {/* ── Nivel 1 · drawer de caja ── */}
      <BoxDetailDrawer
        node={selectedBox?.node ?? null}
        zoneName={selectedBox?.zoneName}
        systemMap={systemMap}
        onClose={() => setSelectedBox(null)}
        onCapClick={(module, slug) => {
          setSelectedBox(null);
          openCap(module, slug);
        }}
      />
    </div>
  );
}

// ── Lente "trabajadores": render por zona → caja → área ─────────────────────

const ZONE_TIER_BADGE: Record<string, { label: string; cls: string }> = {
  core: { label: 'valor', cls: 'bg-[#14532d] text-[#86efac]' },
  supporting: { label: 'transversal', cls: 'bg-[#1e3a5f] text-[#93c5fd]' },
  enabling: { label: 'no-funcional', cls: 'bg-[#3f3f46] text-[#d4d4d8]' },
};

function ZoneHeader({ zone, totalCaps }: { zone: SystemMapZone; totalCaps: number }) {
  const tier = ZONE_TIER_BADGE[zone.tier] ?? null;
  return (
    <div className="mb-3 border-b border-[var(--color-border)] pb-2">
      <div className="flex items-baseline gap-2 flex-wrap">
        <h2 className="text-base font-semibold">{zone.name}</h2>
        {tier && (
          <span className={`text-[9px] px-1.5 py-0.5 rounded font-medium ${tier.cls}`}>
            {tier.label}
          </span>
        )}
        <span className="text-[10px] text-[var(--color-muted)]">
          {zone.boxes?.length ?? 0} cajas · {totalCaps} caps
        </span>
      </div>
      {zone.description && (
        <p className="text-[11px] text-[var(--color-muted)] italic mt-0.5 leading-relaxed">
          {zone.description}
        </p>
      )}
    </div>
  );
}

function ZoneBlock({
  zoneNode,
  showPlanned,
  showInfra,
  statusReport = null,
  onBoxOpen,
}: {
  zoneNode: ZoneNode;
  showPlanned: boolean;
  showInfra: boolean;
  statusReport?: ComputedStatusReport | null;
  onBoxOpen?: (node: BoxNode, zoneName?: string) => void;
}) {
  const { zone, boxes, totalCaps } = zoneNode;
  const isEnabling = zone.user_visible === false;

  // Zona no-funcional (Infraestructura) oculta salvo toggle "infra 🔧"
  if (isEnabling && !showInfra) {
    return (
      <section aria-label={zone.name}>
        <div className="text-[11px] text-[var(--color-muted)] px-2 py-1.5 border border-dashed border-[var(--color-border)] rounded">
          🔧 <b>{zone.name}</b> · {zone.boxes?.length ?? 0} cajas · {totalCaps} caps ·{' '}
          activá <span className="font-mono">infra 🔧</span> arriba para ver
        </div>
      </section>
    );
  }

  return (
    <section aria-label={zone.name}>
      <ZoneHeader zone={zone} totalCaps={totalCaps} />
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {boxes.map((node) => (
          <BoxSection
            key={node.box.id}
            node={node}
            showPlanned={showPlanned}
            statusReport={statusReport}
            onBoxOpen={onBoxOpen ? (n) => onBoxOpen(n, zone.name) : undefined}
          />
        ))}
      </div>
    </section>
  );
}

function ZoneLensView({
  zoneTree,
  showPlanned,
  showInfra,
  statusReport = null,
  onBoxOpen,
}: {
  zoneTree: ZoneNode[];
  showPlanned: boolean;
  showInfra: boolean;
  statusReport?: ComputedStatusReport | null;
  onBoxOpen?: (node: BoxNode, zoneName?: string) => void;
}) {
  return (
    <div className="space-y-6 mb-4">
      {zoneTree.map((zoneNode) => (
        <ZoneBlock
          key={zoneNode.zone.id}
          zoneNode={zoneNode}
          showPlanned={showPlanned}
          showInfra={showInfra}
          statusReport={statusReport}
          onBoxOpen={onBoxOpen}
        />
      ))}
    </div>
  );
}

// ── Lente "proceso": value-stream del GTM (zona Agentes) + capas habilitadoras ─

function StageColumn({
  node,
  showPlanned,
  statusReport = null,
  onBoxOpen,
}: {
  node: ProcessStageNode;
  showPlanned: boolean;
  statusReport?: ComputedStatusReport | null;
  onBoxOpen?: (node: BoxNode, zoneName?: string) => void;
}) {
  return (
    <div className="space-y-2">
      <div className="border-b border-[var(--color-accent)] pb-1">
        <h3 className="text-sm font-semibold">{node.stage.name}</h3>
        <p className="text-[10px] text-[var(--color-muted)] leading-snug">
          {node.stage.description}
        </p>
        <span className="text-[10px] text-[var(--color-muted)]">{node.totalCaps} caps</span>
      </div>
      {node.boxes.length === 0 ? (
        <EmptyState>Sin cajas.</EmptyState>
      ) : (
        node.boxes.map((b) => (
          <BoxSection
            key={b.box.id}
            node={b}
            showPlanned={showPlanned}
            statusReport={statusReport}
            onBoxOpen={onBoxOpen ? (n) => onBoxOpen(n, node.stage.name) : undefined}
          />
        ))
      )}
    </div>
  );
}

function ProcessLensView({
  stages,
  zoneTree,
  showPlanned,
  showInfra,
  statusReport = null,
  onBoxOpen,
}: {
  stages: ProcessStageNode[];
  zoneTree: ZoneNode[];
  showPlanned: boolean;
  showInfra: boolean;
  statusReport?: ComputedStatusReport | null;
  onBoxOpen?: (node: BoxNode, zoneName?: string) => void;
}) {
  // Capas habilitadoras: todo lo que no es la zona Agentes (Plataforma + Infraestructura)
  const enablingZones = zoneTree.filter((z) => z.zone.id !== 'agentes');

  return (
    <div className="space-y-6 mb-4">
      <section aria-label="Value-stream Agentes">
        <div className="mb-3 border-b border-[var(--color-border)] pb-2">
          <h2 className="text-base font-semibold">Value-stream · Agentes</h2>
          <p className="text-[11px] text-[var(--color-muted)] italic mt-0.5">
            El recorrido del usuario operado por los trabajadores:{' '}
            {stages.map((s) => s.stage.name).join(' → ')}.
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 items-start">
          {stages.map((node) => (
            <StageColumn
              key={node.stage.id}
              node={node}
              showPlanned={showPlanned}
              statusReport={statusReport}
              onBoxOpen={onBoxOpen}
            />
          ))}
        </div>
      </section>

      {enablingZones.length > 0 && (
        <section aria-label="Capas habilitadoras">
          <div className="mb-3 border-b border-[var(--color-border)] pb-2">
            <h2 className="text-base font-semibold">Capas habilitadoras</h2>
            <p className="text-[11px] text-[var(--color-muted)] italic mt-0.5">
              Plataforma (el usuario atraviesa) + Infraestructura (no-funcional) que sostienen el value-stream.
            </p>
          </div>
          <div className="space-y-6">
            {enablingZones.map((zoneNode) => (
              <ZoneBlock
                key={zoneNode.zone.id}
                zoneNode={zoneNode}
                showPlanned={showPlanned}
                showInfra={showInfra}
                statusReport={statusReport}
                onBoxOpen={onBoxOpen}
              />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}

function BoxSection({
  node,
  showPlanned,
  fullWidth = false,
  statusReport = null,
  onBoxOpen,
}: {
  node: BoxNode;
  showPlanned: boolean;
  fullWidth?: boolean;
  statusReport?: ComputedStatusReport | null;
  onBoxOpen?: (node: BoxNode) => void;
}) {
  const { box, areas, totalCaps } = node;

  // Filtrar áreas: si !showPlanned, ocultar áreas planned sin caps
  const visibleAreas = areas.filter((a) => {
    if (!showPlanned && a.area.status === 'planned' && a.caps.length === 0) return false;
    return true;
  });

  return (
    <Card className={`!p-4 h-full ${fullWidth ? 'col-span-full' : ''}`}>
      <header className="mb-3 border-b border-[var(--color-border)] pb-2">
        <div
          className={cn('flex items-baseline gap-2', onBoxOpen && 'cursor-pointer group')}
          onClick={onBoxOpen ? () => onBoxOpen(node) : undefined}
          title={onBoxOpen ? 'Click: detalle de la caja (áreas, flujos, datos)' : undefined}
          role={onBoxOpen ? 'button' : undefined}
        >
          <span aria-hidden="true" className="text-xl">
            {box.emoji}
          </span>
          <div className="flex-1 min-w-0">
            <h2 className={cn('text-sm font-semibold', onBoxOpen && 'group-hover:text-[var(--color-accent)] transition-colors')}>
              {box.name}
              {onBoxOpen && (
                <span className="text-[9px] text-[var(--color-muted)] ml-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                  ver detalle ›
                </span>
              )}
            </h2>
            {box.subtitle && (
              <p className="text-[10px] text-[var(--color-muted)]">{box.subtitle}</p>
            )}
          </div>
          <span className="text-[10px] text-[var(--color-muted)] shrink-0">
            {totalCaps}
          </span>
        </div>
        <BoxProgress node={node} />
      </header>
      {visibleAreas.length === 0 ? (
        <EmptyState>Sin áreas visibles.</EmptyState>
      ) : (
        <div className="space-y-3">
          {visibleAreas.map(({ area, fullId, caps }) => (
            <div key={fullId}>
              <div className="flex items-center gap-1.5 mb-1">
                <div className="text-[10px] text-[var(--color-muted)] font-mono uppercase tracking-wide flex-1">
                  {area.name}
                </div>
                <AreaStatusBadge status={area.status} />
              </div>
              {area.description && (
                <div className="text-[10px] text-[var(--color-muted)] italic mb-1 leading-relaxed" title={area.description}>
                  {area.description.length > 80 ? `${area.description.slice(0, 80)}…` : area.description}
                </div>
              )}
              <div className={fullWidth ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2' : 'space-y-1.5'}>
                {caps.length === 0 ? (
                  area.status === 'planned' ? (
                    <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
                      📋 sin caps · estimado release {area.target_release ?? 'TBD'}
                    </div>
                  ) : (
                    <EmptyState>Sin capabilities shipped.</EmptyState>
                  )
                ) : (
                  caps.map((c, i) => (
                    <CapItem key={capKey(c, i)} cap={c} statusReport={statusReport} />
                  ))
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </Card>
  );
}

function AreaStatusBadge({ status }: { status: AreaStatus }) {
  const badge = STATUS_BADGES[status];
  return (
    <span className={`text-[9px] px-1.5 py-0.5 rounded font-medium ${badge.cls}`}>
      {badge.label}
    </span>
  );
}

/**
 * Vista legacy de fallback si el SYSTEM-MAP no carga. Usa agent_owner de caps
 * directamente: los agentes se DERIVAN del data presente (sin roster
 * hardcodeado · F-4) con metadata determinística de agent-meta.
 */
function LegacyFallbackView({
  filtered,
  caps,
  showInfra,
  statusReport = null,
}: {
  filtered: Capability[];
  caps: Capability[];
  showInfra: boolean;
  statusReport?: ComputedStatusReport | null;
}) {
  const INFRA_FALLBACK = { id: 'infra' as const, emoji: '🔧', name: 'Infra', subtitle: 'observability · platform · scaffolding' };

  const byAgent = useMemo(() => {
    const map = new Map<string, Capability[]>();
    map.set('infra', []);
    const orphans: Capability[] = [];
    for (const c of filtered) {
      if (c.superseded_by) continue;
      const owner = c.agent_owner;
      if (!owner) { orphans.push(c); continue; }
      if (!map.has(owner)) map.set(owner, []);
      map.get(owner)!.push(c);
    }
    return { map, orphans };
  }, [filtered]);

  // Agentes presentes en el data (excluye 'infra' que tiene su card aparte)
  const FALLBACK_AGENTS = useMemo(
    () =>
      Array.from(byAgent.map.keys())
        .filter((id) => id !== 'infra')
        .sort()
        .map((id) => {
          const meta = agentMetaOf(id);
          return { id, emoji: meta.emoji, name: meta.name, subtitle: '' };
        }),
    [byAgent]
  );

  return (
    <>
      <div className="text-[11px] text-amber-400 mb-3 px-2 py-1 border border-amber-700 rounded">
        ⚠️ SYSTEM-MAP.yaml no disponible · mostrando vista legacy por agent_owner
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mb-4">
        {FALLBACK_AGENTS.map((agent) => {
          const agentCaps = byAgent.map.get(agent.id) ?? [];
          return (
            <Card key={agent.id} className="!p-4 h-full">
              <header className="flex items-baseline gap-2 mb-3 border-b border-[var(--color-border)] pb-2">
                <span aria-hidden="true" className="text-xl">{agent.emoji}</span>
                <div className="flex-1 min-w-0">
                  <h2 className="text-sm font-semibold">{agent.name}</h2>
                  <p className="text-[10px] text-[var(--color-muted)]">{agent.subtitle}</p>
                </div>
                <span className="text-[10px] text-[var(--color-muted)] shrink-0">{agentCaps.length}</span>
              </header>
              {agentCaps.length === 0 ? (
                <EmptyState>Sin capabilities todavía.</EmptyState>
              ) : (
                <div className="space-y-1.5">
                  {agentCaps.map((c, i) => <CapItem key={capKey(c, i)} cap={c} statusReport={statusReport} />)}
                </div>
              )}
            </Card>
          );
        })}
      </div>
      {showInfra && (byAgent.map.get('infra') ?? []).length > 0 && (
        <Card className="!p-4 col-span-full">
          <header className="flex items-baseline gap-2 mb-3 border-b border-[var(--color-border)] pb-2">
            <span aria-hidden="true" className="text-xl">{INFRA_FALLBACK.emoji}</span>
            <div className="flex-1 min-w-0">
              <h2 className="text-sm font-semibold">{INFRA_FALLBACK.name}</h2>
              <p className="text-[10px] text-[var(--color-muted)]">{INFRA_FALLBACK.subtitle}</p>
            </div>
          </header>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
            {(byAgent.map.get('infra') ?? []).map((c, i) => <CapItem key={capKey(c, i)} cap={c} statusReport={statusReport} />)}
          </div>
        </Card>
      )}
      {byAgent.orphans.length > 0 && (
        <Card className="!p-4 mt-4 border-red-700">
          <h2 className="text-sm font-semibold text-red-400 mb-2">
            ⚠️ Capabilities sin agent_owner ({byAgent.orphans.length})
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2">
            {byAgent.orphans.map((c, i) => <CapItem key={capKey(c, i)} cap={c} statusReport={statusReport} />)}
          </div>
        </Card>
      )}
    </>
  );
}

function CapItem({
  cap,
  statusReport = null,
}: {
  cap: Capability;
  statusReport?: ComputedStatusReport | null;
}) {
  const { openCap } = useDrawer();
  const [expanded, setExpanded] = useState(false);

  // Split functional_area en [agent].[area] si está set
  const [agentChip, areaChip] = (cap.functional_area ?? '').split('.', 2);

  // Buscar computed status por slug del cap
  const computed: CapStatusComputed | null =
    statusReport?.capabilities[cap.slug] ?? null;
  const badge = computed ? getStatusBadge(computed.computed_status) : null;

  const scenarios = cap.scenarios ?? [];
  const hasScenarios = scenarios.length > 0;
  const isV32Populated = hasScenarios;

  return (
    <div
      className={cn(
        'rounded border text-xs transition-colors',
        'bg-[var(--color-panel)] border-[var(--color-border)]'
      )}
    >
      {/* Row principal: click abre drawer o toggle scenarios */}
      <div className="flex items-stretch">
        <button
          type="button"
          onClick={() => openCap(cap.module, cap.slug)}
          className={cn(
            'flex-1 text-left px-2 py-1.5 transition-colors',
            'hover:border-[#3a4358] hover:bg-[var(--color-panel2)] rounded-l'
          )}
        >
          <div className="flex items-center gap-1.5 flex-wrap">
            <Pill className={STATUS_CLASSES[cap.status] ?? 'bg-[#1f2937]'}>
              {cap.status}
            </Pill>
            {badge && (
              <span
                className="text-[10px]"
                title={`Computed: ${computed?.computed_status}`}
              >
                {badge.emoji}
              </span>
            )}
            {isV32Populated && (
              <Tooltip content={TOOLTIPS.v3_2_badge} variant="badge">
                <Pill className="bg-[#1e3a5f] text-[#93c5fd] text-[9px] py-0">
                  v3.2
                </Pill>
              </Tooltip>
            )}
            <div className="flex items-center gap-1 text-[11px] flex-1 truncate">
              <span className="font-medium truncate">
                {cap.user_facing_name ?? `${cap.module}/${cap.slug}`}
              </span>
              {agentChip && (
                <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[10px] py-0">
                  {agentChip}
                </Pill>
              )}
              {areaChip && (
                <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[10px] py-0 opacity-70">
                  {areaChip}
                </Pill>
              )}
            </div>
          </div>
          <div className="text-[10px] text-[var(--color-muted)] mt-0.5 font-mono truncate">
            {cap.module}/{cap.slug}
            {hasScenarios && (
              <span className="ml-2">
                · {scenarios.length} escenario{scenarios.length !== 1 ? 's' : ''}
              </span>
            )}
          </div>
        </button>

        {/* Toggle expand scenarios */}
        <button
          type="button"
          aria-label={expanded ? 'Ocultar escenarios' : 'Ver escenarios'}
          onClick={() => setExpanded((v) => !v)}
          className={cn(
            'px-1.5 text-[var(--color-muted)] hover:text-[var(--color-fg)] transition-colors',
            'border-l border-[var(--color-border)] rounded-r',
            expanded && 'bg-[var(--color-panel2)]'
          )}
        >
          <span aria-hidden="true" className="text-[10px]">
            {expanded ? '▲' : '▼'}
          </span>
        </button>
      </div>

      {/* Scenarios drawer */}
      {expanded && (
        <div className="border-t border-[var(--color-border)] px-2 py-1.5">
          {!hasScenarios ? (
            <div className="text-[10px] text-[var(--color-muted)] italic">
              Sin escenarios declarados (cap stub)
            </div>
          ) : (
            <div className="space-y-0.5">
              {scenarios.map((s, idx) => (
                <div
                  key={`${s.id ?? s.added_in_story}-${idx}`}
                  className="text-[10px] text-[var(--color-muted)] flex items-center gap-1"
                >
                  <span
                    className={cn(
                      'w-1.5 h-1.5 rounded-full shrink-0',
                      s.status === 'live'
                        ? 'bg-green-500'
                        : s.status === 'wip'
                        ? 'bg-blue-500'
                        : 'bg-gray-500'
                    )}
                    title={`status: ${s.status}`}
                  />
                  <span className="truncate">{s.name ?? '(sin nombre)'}</span>
                </div>
              ))}
            </div>
          )}
          {computed && (
            <div className="mt-1 pt-1 border-t border-[var(--color-border)] text-[9px] text-[var(--color-muted)] font-mono">
              {badge?.emoji} {computed.computed_status} · {computed.scenarios_total} escenario
              {computed.scenarios_total !== 1 ? 's' : ''} · {computed.scenarios_verified} verificado
              {computed.scenarios_verified !== 1 ? 's' : ''}
              {computed.drift_reasons.length > 0 && (
                <span className="text-red-400 ml-1" title={computed.drift_reasons.join('; ')}>
                  · drift
                </span>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Genera React key única para un cap. Algunos caps llegan con module/slug
 * vacíos (YAML mal formado, schema v2 incompleto) — el patrón ${module}/${slug}
 * colapsaba en "/" duplicado. Fallback chain: capability_id > path > index.
 */
function capKey(cap: Capability, index: number): string {
  if (cap.capability_id) return `id:${cap.capability_id}`;
  if (cap.path) return `path:${cap.path}`;
  if (cap.module && cap.slug) return `${cap.module}/${cap.slug}`;
  return `idx:${index}:${cap.module ?? '?'}/${cap.slug ?? '?'}`;
}
