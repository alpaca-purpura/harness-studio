/**
 * ScenariosSection · ✨ Qué puedo hacer
 * Extraído de FunctionalityView.tsx (2026-05-28 refactor).
 */

import { Pill } from '@/components/ui/Badge';
import { Tooltip } from '@/components/ui/Tooltip';
import { scenarioTruth } from '@/lib/cap-badges';
import { TOOLTIPS } from '@/lib/tooltips';
import type { CapScenario, DevPreview } from '@/lib/types';

export function ScenariosSection({
  scenarios,
  userFacingDescription,
  devPreview,
}: {
  scenarios: CapScenario[];
  userFacingDescription?: string | null;
  devPreview?: DevPreview | null;
}) {
  if (scenarios.length === 0) {
    // HB-52: «✨ Qué puedo hacer» SIEMPRE responde QUÉ HAGO, nunca la versión interna
    // del SDD. Sin scenarios materializados, caemos al texto funcional de la cap
    // (user_facing_description) + dónde se usa — sin jerga de versión del SDD.
    const navigate = devPreview?.how_to_navigate ?? devPreview?.route ?? null;
    return (
      <section>
        <h3 className="text-sm font-semibold mb-2">
          <Tooltip content={TOOLTIPS.scenarios} variant="header">
            ✨ Qué puedo hacer
          </Tooltip>
        </h3>
        {userFacingDescription ? (
          <div className="text-[11px] px-2 py-1 border border-[var(--color-border)] rounded">
            {userFacingDescription}
            {navigate && (
              <div className="mt-1 text-[10px] text-[var(--color-muted)]">Disponible en: {navigate}</div>
            )}
          </div>
        ) : navigate ? (
          <div className="text-[11px] text-[var(--color-muted)] px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
            Disponible en: {navigate}
          </div>
        ) : (
          <div className="text-[11px] text-[var(--color-muted)] italic px-2 py-1 border border-dashed border-[var(--color-border)] rounded">
            Esta capability todavía no tiene escenarios documentados.
          </div>
        )}
      </section>
    );
  }

  return (
    <section>
      <h3 className="text-sm font-semibold mb-2">
        <Tooltip content={TOOLTIPS.scenarios} variant="header">
          ✨ Qué puedo hacer ({scenarios.length})
        </Tooltip>
      </h3>
      <div className="space-y-2">
        {scenarios.map((s) => (
          <ScenarioRow key={s.id} scenario={s} />
        ))}
      </div>
    </section>
  );
}

function ScenarioRow({ scenario: s }: { scenario: CapScenario }) {
  return (
    <div className="text-[11px] pl-2 border-l-2 border-green-700 bg-[var(--color-panel2)] p-2 rounded">
      <div className="font-semibold flex items-center gap-1.5">
        <span>{s.name}</span>
        <Pill className="bg-[var(--color-panel)] border border-[var(--color-border)] text-[9px] py-0">
          {s.actor}
        </Pill>
        {s.status !== 'live' && (
          <Pill className="bg-[#27272a] text-[#a1a1aa] text-[9px]">{s.status}</Pill>
        )}
        {s.status === 'live' &&
          (() => {
            const t = scenarioTruth(s);
            return (
              <Tooltip content={t.tip} variant="badge">
                <Pill className={`${t.cls} text-[9px] py-0`}>
                  {t.icon} {t.label}
                </Pill>
              </Tooltip>
            );
          })()}
      </div>
      <div className="mt-1 text-[10px] space-y-0.5">
        <div>
          <span className="text-[var(--color-muted)]">Dado que </span>
          {s.given}
        </div>
        <div>
          <span className="text-[var(--color-muted)]">Cuando </span>
          {s.when}
        </div>
        <div>
          <span className="text-[var(--color-muted)]">Entonces </span>
          {s.then}
        </div>
      </div>
      {s.edge_cases && s.edge_cases.length > 0 && (
        <div className="mt-1.5">
          <div className="text-[10px] text-[var(--color-muted)] font-semibold">
            <Tooltip content={TOOLTIPS.edge_cases}>
              <span>Casos borde</span>
            </Tooltip>
            :
          </div>
          <ul className="list-disc list-inside text-[10px] text-[var(--color-muted)] space-y-0.5 ml-1">
            {s.edge_cases.map((ec, i) => (
              <li key={i}>{ec}</li>
            ))}
          </ul>
        </div>
      )}
      {(s.e2e_test || s.story_spec_ref) && (
        <div className="mt-1 text-[9px] text-[var(--color-muted)] font-mono space-y-0.5">
          {s.e2e_test && (
            <div>
              <Tooltip content={TOOLTIPS.e2e_test}>
                <span>test</span>
              </Tooltip>
              : {s.e2e_test}
            </div>
          )}
          {s.story_spec_ref && (
            <div>
              <Tooltip content={TOOLTIPS.story_spec_ref}>
                <span>spec</span>
              </Tooltip>
              : {s.story_spec_ref}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
