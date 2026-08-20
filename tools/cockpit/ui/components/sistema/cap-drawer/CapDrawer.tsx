'use client';

import { useEffect, useState } from 'react';
import { Drawer } from '@/components/ui/Drawer';
import { Card } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Badge, Pill } from '@/components/ui/Badge';
import { Spinner, ErrorBanner, EmptyState } from '@/components/ui/Spinner';
import { Tooltip } from '@/components/ui/Tooltip';
import { TOOLTIPS } from '@/lib/tooltips';
import { ExternalLink, Plus } from 'lucide-react';
import { useDrawer } from '@/components/providers/DrawerProvider';
import { useSistema } from '@/components/providers/SistemaProvider';
import {
  getCapability,
  openInEditor,
  getCodeIndex,
  getBidirectionalValidation,
} from '@/lib/api-client';
import { ExtendCapModal } from './ExtendCapModal';
import { CapLevel } from './CapLevel';
import { AccessSection } from './sections/AccessSection';
import { ScenariosSection } from './sections/ScenariosSection';
import { BusinessRulesSection } from './sections/BusinessRulesSection';
import { RelatedCapsSection } from './sections/RelatedCapsSection';
import { CodeFilesSection } from './sections/CodeFilesSection';
import { BidirectionalSection } from './sections/BidirectionalSection';
import type {
  Capability,
  CodeIndexReport,
  BidirectionalValidationReport,
} from '@/lib/types';
import toast from 'react-hot-toast';

const STATUS_CLASSES: Record<string, string> = {
  live: 'bg-[#14532d] text-[#86efac]',
  beta: 'bg-[#713f12] text-[#fbbf24]',
  deprecated: 'bg-[#3f3f46] text-[#d4d4d8]',
  sunset: 'bg-[#450a0a] text-[#fca5a5]',
};

export function CapDrawer() {
  const { capRef, closeCap, openStory } = useDrawer();
  const { sistema } = useSistema();
  const [cap, setCap] = useState<Capability | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [extendOpen, setExtendOpen] = useState(false);

  // v3.2 cross-refs (carga paralela una sola vez por sesión drawer)
  const [codeIndex, setCodeIndex] = useState<CodeIndexReport | null>(null);
  const [codeIndexHint, setCodeIndexHint] = useState<string | null>(null);
  const [bidirReport, setBidirReport] = useState<BidirectionalValidationReport | null>(null);
  const [bidirHint, setBidirHint] = useState<string | null>(null);

  useEffect(() => {
    if (!capRef) {
      setCap(null);
      return;
    }
    setLoading(true);
    setError(null);
    Promise.all([
      getCapability(capRef.module, capRef.slug, sistema),
      getCodeIndex(sistema).catch(() => ({ index: null, hint: null })),
      getBidirectionalValidation(sistema).catch(() => ({ validation: null, hint: null })),
    ])
      .then(([c, ci, biv]) => {
        setCap(c);
        setCodeIndex(ci.index ?? null);
        setCodeIndexHint(('hint' in ci ? ci.hint : null) ?? null);
        setBidirReport(biv.validation ?? null);
        setBidirHint(('hint' in biv ? biv.hint : null) ?? null);
      })
      .catch((err) => setError((err as Error).message))
      .finally(() => setLoading(false));
  }, [capRef, sistema]);

  async function handleOpenYaml() {
    if (!cap?.path) return;
    try {
      await openInEditor(cap.path);
      toast.success('YAML abierto');
    } catch (err) {
      toast.error((err as Error).message);
    }
  }

  return (
    <Drawer
      open={capRef !== null}
      onClose={closeCap}
      title={
        cap ? (
          <div className="flex items-center gap-2 min-w-0">
            <Badge className={STATUS_CLASSES[cap.status] ?? 'bg-[#1f2937]'}>
              {cap.status}
            </Badge>
            <span className="font-mono text-sm truncate">
              {cap.module}/{cap.slug}
            </span>
          </div>
        ) : (
          <span className="text-sm text-[var(--color-muted)]">Capability…</span>
        )
      }
      width={800}
    >
      {loading && (
        <div className="flex items-center gap-2 text-xs text-[var(--color-muted)]">
          <Spinner /> Cargando capability…
        </div>
      )}
      {error && <ErrorBanner message={error} />}
      {!loading && !error && cap && (
        <div className="space-y-5">
          {/* YAML ledger */}
          <Card>
            <div className="flex items-center justify-between mb-2">
              <h3 className="text-sm font-semibold">
                <Tooltip content={TOOLTIPS.yaml_ledger} variant="header">
                  YAML ledger
                </Tooltip>
              </h3>
              <div className="flex gap-2">
                <Button size="sm" onClick={handleOpenYaml}>
                  <ExternalLink className="w-3 h-3" />
                  xed
                </Button>
                <Button
                  size="sm"
                  variant="primary"
                  onClick={() => setExtendOpen(true)}
                >
                  <Plus className="w-3 h-3" />
                  Extender
                </Button>
              </div>
            </div>
            <div className="text-[11px] font-mono text-[var(--color-muted)] truncate">
              {cap.path ?? '—'}
            </div>
            <dl className="grid grid-cols-2 gap-x-4 gap-y-1.5 text-xs mt-3">
              <dt className="text-[var(--color-muted)]">License</dt>
              <dd>
                <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)]">
                  {cap.license}
                </Pill>
              </dd>
              <dt className="text-[var(--color-muted)]">
                <Tooltip content={TOOLTIPS.created_in_story}>
                  <span>Created in story</span>
                </Tooltip>
              </dt>
              <dd className="font-mono text-[11px]">{cap.created_in_story}</dd>
              <dt className="text-[var(--color-muted)]">Created date</dt>
              <dd>{cap.created_date}</dd>
              <dt className="text-[var(--color-muted)]">Last modified</dt>
              <dd>{cap.last_modified}</dd>
              <dt className="text-[var(--color-muted)]">
                <Tooltip content={TOOLTIPS.parent_cap}>
                  <span>Parent cap</span>
                </Tooltip>
              </dt>
              <dd className="font-mono text-[11px]">{cap.parent_cap ?? '—'}</dd>
              {cap.architecture_pattern && (
                <>
                  <dt className="text-[var(--color-muted)]">
                    <Tooltip content={TOOLTIPS.architecture_pattern}>
                      <span>Pattern</span>
                    </Tooltip>
                  </dt>
                  <dd className="text-[11px]">{cap.architecture_pattern}</dd>
                </>
              )}
              {cap.hipaa_lite_overlay && (
                <>
                  <dt className="text-[var(--color-muted)]">
                    <Tooltip content={TOOLTIPS.hipaa_lite_overlay}>
                      <span>HIPAA overlay</span>
                    </Tooltip>
                  </dt>
                  <dd>
                    <Pill className="bg-[#450a0a] text-[#fca5a5] text-[10px]">
                      activo
                    </Pill>
                  </dd>
                </>
              )}
              {cap.agent_owner && (
                <>
                  <dt className="text-[var(--color-muted)]">
                    <Tooltip content={TOOLTIPS.agent_owner}>
                      <span>Agent owner</span>
                    </Tooltip>
                  </dt>
                  <dd>
                    <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)]">
                      {cap.agent_owner}
                    </Pill>
                  </dd>
                </>
              )}
              {cap.functional_area && (
                <>
                  <dt className="text-[var(--color-muted)]">
                    <Tooltip content={TOOLTIPS.functional_area}>
                      <span>Functional area</span>
                    </Tooltip>
                  </dt>
                  <dd className="flex items-center gap-1">
                    {(() => {
                      const parts = cap.functional_area.split('.');
                      if (parts.length === 2) {
                        return (
                          <>
                            <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[10px]">{parts[0]}</Pill>
                            <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[10px] opacity-75">{parts[1]}</Pill>
                          </>
                        );
                      }
                      return <span className="font-mono text-[11px]">{cap.functional_area}</span>;
                    })()}
                  </dd>
                </>
              )}
              {cap.nature && (
                <>
                  <dt className="text-[var(--color-muted)]">
                    <Tooltip content={TOOLTIPS.nature}>
                      <span>Naturaleza</span>
                    </Tooltip>
                  </dt>
                  <dd className="text-[11px]">{cap.nature}</dd>
                </>
              )}
              <dt className="text-[var(--color-muted)]">
                <Tooltip content={TOOLTIPS.user_visible}>
                  <span>User visible</span>
                </Tooltip>
              </dt>
              <dd className="text-[11px]">{cap.user_visible !== false ? 'Sí' : 'No (infra)'}</dd>
            </dl>
          </Card>

          {/* Cómo verlo · v3 user-facing + dev_preview (con storybook/loom fix R3) */}
          {(cap.user_facing_name || cap.user_facing_description || cap.dev_preview) && (
            <Card>
              <h3 className="text-sm font-semibold mb-2">
                <Tooltip content={TOOLTIPS.dev_preview} variant="header">
                  📍 Cómo verlo
                </Tooltip>
              </h3>
              {cap.user_facing_name && (
                <div className="text-sm font-medium mb-1">{cap.user_facing_name}</div>
              )}
              {cap.user_facing_description && (
                <p className="text-xs text-[var(--color-muted)] mb-3 leading-relaxed">
                  {cap.user_facing_description}
                </p>
              )}
              {cap.dev_preview && (
                <dl className="grid grid-cols-[120px_1fr] gap-x-3 gap-y-1.5 text-xs">
                  {cap.dev_preview.route && (
                    <>
                      <dt className="text-[var(--color-muted)]">Ruta</dt>
                      <dd className="font-mono text-[11px]">{cap.dev_preview.route}</dd>
                    </>
                  )}
                  {cap.dev_preview.how_to_navigate && (
                    <>
                      <dt className="text-[var(--color-muted)]">Cómo llegar</dt>
                      <dd className="text-[11px] leading-relaxed">{cap.dev_preview.how_to_navigate}</dd>
                    </>
                  )}
                  {cap.dev_preview.main_component && (
                    <>
                      <dt className="text-[var(--color-muted)]">Componente</dt>
                      <dd className="font-mono text-[10px] break-all">{cap.dev_preview.main_component}</dd>
                    </>
                  )}
                  {cap.dev_preview.api_endpoints && cap.dev_preview.api_endpoints.length > 0 && (
                    <>
                      <dt className="text-[var(--color-muted)]">Endpoints</dt>
                      <dd>
                        <ul className="space-y-0.5">
                          {cap.dev_preview.api_endpoints.map((ep, i) => (
                            <li key={i} className="font-mono text-[10px]">{ep}</li>
                          ))}
                        </ul>
                      </dd>
                    </>
                  )}
                  {cap.dev_preview.e2e_test && (
                    <>
                      <dt className="text-[var(--color-muted)]">
                        <Tooltip content={TOOLTIPS.e2e_test}>
                          <span>Test E2E</span>
                        </Tooltip>
                      </dt>
                      <dd className="font-mono text-[10px] break-all">{cap.dev_preview.e2e_test}</dd>
                    </>
                  )}
                  {cap.dev_preview.fixtures_required && cap.dev_preview.fixtures_required.length > 0 && (
                    <>
                      <dt className="text-[var(--color-muted)]">Fixtures</dt>
                      <dd className="text-[11px]">{cap.dev_preview.fixtures_required.join(', ')}</dd>
                    </>
                  )}
                  {cap.dev_preview.storybook_url && (
                    <>
                      <dt className="text-[var(--color-muted)]">Storybook</dt>
                      <dd>
                        <a
                          href={cap.dev_preview.storybook_url}
                          target="_blank"
                          rel="noreferrer"
                          className="font-mono text-[10px] text-[var(--color-accent)] hover:underline break-all"
                        >
                          {cap.dev_preview.storybook_url}
                        </a>
                      </dd>
                    </>
                  )}
                  {cap.dev_preview.loom_demo && (
                    <>
                      <dt className="text-[var(--color-muted)]">Loom demo</dt>
                      <dd>
                        <a
                          href={cap.dev_preview.loom_demo}
                          target="_blank"
                          rel="noreferrer"
                          className="font-mono text-[10px] text-[var(--color-accent)] hover:underline break-all"
                        >
                          {cap.dev_preview.loom_demo}
                        </a>
                      </dd>
                    </>
                  )}
                </dl>
              )}
              {cap.superseded_by && (
                <div className="mt-3 text-[11px] text-amber-400 bg-amber-950/30 px-2 py-1 rounded border border-amber-700">
                  <Tooltip content={TOOLTIPS.superseded_by}>
                    <span>Esta capability fue mergeada</span>
                  </Tooltip>{' '}
                  a <strong>{cap.superseded_by}</strong>. Ver ese cap para la visión consolidada.
                </div>
              )}
            </Card>
          )}

          {/* ═══ Niveles N1-N4 (cap-levels-proposal · N0 = «Cómo verlo» card arriba) ═══ */}

          {/* N1 · Qué puedo hacer — casos de uso + badge de verdad (✅/🟠/⚪) */}
          <CapLevel level="N1" title="¿Qué puedo hacer?" hint="casos de uso" defaultOpen>
            <ScenariosSection
              scenarios={cap.scenarios ?? []}
              userFacingDescription={cap.user_facing_description}
              devPreview={cap.dev_preview}
            />
          </CapLevel>

          {/* N2 · Bajo qué reglas — reglas de negocio + badge enforcement (🟢/🔴) */}
          {cap.business_rules && cap.business_rules.length > 0 && (
            <CapLevel level="N2" title="¿Bajo qué reglas?" hint="reglas de negocio" defaultOpen>
              <BusinessRulesSection rules={cap.business_rules} />
            </CapLevel>
          )}

          {/* N3 · Quién y por dónde — acceso / entry_points / roles */}
          {cap.access && (
            <CapLevel level="N3" title="¿Quién entra y por dónde?" hint="acceso" defaultOpen={false}>
              <AccessSection access={cap.access} />
            </CapLevel>
          )}

          {/* N4 · Dónde vive / cómo se conecta — código + deps + validación bidireccional */}
          <CapLevel
            level="N4"
            title="¿Dónde vive y cómo se conecta?"
            hint="código · deps · validación"
            defaultOpen={false}
          >
            <div className="space-y-4">
              <CodeFilesSection
                capId={`${cap.module}.${cap.slug}`}
                codeIndex={codeIndex}
                hint={codeIndexHint}
              />
              {cap.related_capabilities && (
                <RelatedCapsSection related={cap.related_capabilities} />
              )}
              <BidirectionalSection
                capId={`${cap.module}.${cap.slug}`}
                report={bidirReport}
                hint={bidirHint}
              />
            </div>
          </CapLevel>

          {/* Changelog */}
          <section>
            <h3 className="text-sm font-semibold mb-2">
              <Tooltip content={TOOLTIPS.change_log} variant="header">
                Historial · change log ({cap.change_log.length})
              </Tooltip>
            </h3>
            {cap.change_log.length === 0 ? (
              <EmptyState>Sin entries todavía.</EmptyState>
            ) : (
              <ol className="border-l-2 border-[var(--color-border)] pl-4 space-y-3">
                {cap.change_log.map((entry, i) => (
                  <li key={i} className="relative">
                    <span
                      aria-hidden="true"
                      className="absolute -left-[21px] top-1 w-3 h-3 rounded-full bg-[var(--color-accent)] border-2 border-[var(--color-panel)]"
                    />
                    <div className="text-[10px] text-[var(--color-muted)] font-mono mb-0.5">
                      {entry.date} · {entry.type} ·{' '}
                      <button
                        onClick={() => openStory(entry.story_id)}
                        className="text-[var(--color-accent)] hover:underline"
                      >
                        {entry.story_id}
                      </button>
                    </div>
                    <div className="text-xs">{entry.summary}</div>
                  </li>
                ))}
              </ol>
            )}
          </section>
        </div>
      )}

      {cap && (
        <ExtendCapModal
          open={extendOpen}
          onClose={() => setExtendOpen(false)}
          sistema={sistema}
          parentCap={{ module: cap.module, slug: cap.slug }}
          onCreated={(storyId) => {
            openStory(storyId);
            closeCap();
          }}
        />
      )}
    </Drawer>
  );
}
