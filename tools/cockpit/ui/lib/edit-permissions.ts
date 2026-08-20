/**
 * Edit permission policy por artifact + story state.
 *
 * Regla cardinal (decisión del operador 2026-05-27 conversación cockpit):
 *
 *   "Si ya hay desarrollo no debemos poder modificar lo que fue la fuente
 *    de código en un momento, siguiendo el proceso. Lo que siempre se puede
 *    editar es el checkpoint."
 *
 * O sea: cada artifact tiene un "consumption state" — el primer state en el
 * que el artifact es leído por la próxima fase (dev-team / auditor / merge).
 * Una vez consumido NO se puede editar desde el cockpit · cambios requieren
 * delta-spec.md o nueva story (paradigm v4).
 *
 * Excepciones:
 *  - checkpoint.md → SIEMPRE editable (escape hatch principal · decisión explícita)
 *  - operator-input → append-only via UI dedicada (NO via raw editor)
 *  - T-{n}-* logs → siempre read-only en cockpit (append-only del dev-team/auditor)
 *  - archive/**  → siempre read-only (R2 sistema-docs-schema)
 *  - 07-merge.md → siempre read-only (terminal)
 *
 * Mapping consumption-state ↔ paradigm v4:
 *   refining/refined/ready  → spec/design/arch/validators/tickets editable
 *   developing              → todo lo anterior LOCKED (dev-team consumiendo)
 *   reviewing               → gherkin-matrix se genera, audit reports append
 *   done                    → merge cerró todo, archive read-only
 */

import type { ProcesoResponse, StoryState } from './types';
import { editOrden, estadoDe } from './proceso';

// F6/RN-51: el ORDEN de estados ya no vive acá — se deriva del descriptor de proceso
// (`editOrden` en lib/proceso.ts: índice en estados[]; pausado edita como el inicial;
// descartado = ∞). LOCK_FROM_STATE abajo QUEDA: es política de ENTIDAD por artefacto
// (qué archivo se congela cuándo), fuera del descriptor v1 (SPEC proceso-descriptor §6).

/**
 * Categorías de artifact. El kind se infiere del filename.
 */
export type ArtifactKind =
  | 'checkpoint' // siempre editable
  | 'operator-input' // append-only via UI dedicada
  | 'spec' // 01-spec.md — locked from developing
  | 'design' // 02-design-*.md — locked from developing
  | 'arch' // 03-arch*.md — locked from developing
  | 'validators' // 04-validators.yaml — locked from developing
  | 'guidelines' // 05-guidelines.md — locked from developing
  | 'tickets' // 06-tickets.yaml — locked from developing
  | 'dispatch-plan' // dispatch-plan.md — locked from developing
  | 'mockup' // mockups/* — locked from developing
  | 'ticket-log' // T-{n}-impl-log/result/review/handoff — read-only siempre
  | 'audit' // 06-audit/gherkin-matrix.md, CHECKPOINTS.md — locked from done
  | 'merge' // 07-merge.md — read-only (terminal)
  | 'capability' // capabilities/{m}/{c}.yaml — append-only (cap_change_type ledger)
  | 'release' // releases/{F}.yaml — editable hasta merge release
  | 'learning' // learnings/*.md — editable siempre (apéndices vivos)
  | 'unknown';

/**
 * Infiere el kind desde un path relativo a story root o a sistema docs.
 * Acepta basenames simples ("01-spec.md") o paths con subdir ("06-audit/gherkin-matrix.md").
 */
export function getArtifactKind(filename: string): ArtifactKind {
  const f = filename.replace(/^\.?\/+/, ''); // normalizar leading ./
  const base = f.split('/').pop() ?? f;
  const lower = base.toLowerCase();

  // Checkpoint / operator-input (acepta el filename legacy del kit // legacy F-1)
  if (lower === 'checkpoint.md') return 'checkpoint';
  if (lower === 'operator-input.md' || lower === 'chris-input.md') return 'operator-input';

  // Tickets logs (read-only siempre)
  if (/^t-\d+-(impl-log|result|review|handoff)\.md$/i.test(base)) return 'ticket-log';

  // Audit
  if (f.startsWith('06-audit/')) return 'audit';
  if (lower === 'checkpoints.md') return 'audit';

  // Merge
  if (lower === '07-merge.md') return 'merge';

  // Spec/design/arch/validators/guidelines/tickets/dispatch
  if (/^01-spec.*\.md$/i.test(base)) return 'spec';
  if (/^02-design.*\.md$/i.test(base)) return 'design';
  if (/^03-arch.*\.md$/i.test(base)) return 'arch';
  if (/^04-validators.*\.ya?ml$/i.test(base)) return 'validators';
  if (/^05-guidelines.*\.md$/i.test(base)) return 'guidelines';
  if (/^06-tickets.*\.ya?ml$/i.test(base)) return 'tickets';
  if (/^dispatch-plan\.md$/i.test(base)) return 'dispatch-plan';

  // Mockups (cualquier path bajo mockups/)
  if (f.includes('mockups/')) return 'mockup';

  // Capability / release / learning (paths fuera de story-folder)
  if (f.includes('/capabilities/')) return 'capability';
  if (f.includes('/releases/')) return 'release';
  if (f.includes('/learnings/')) return 'learning';

  return 'unknown';
}

/** State desde el cual el artifact deja de ser editable. */
const LOCK_FROM_STATE: Partial<Record<ArtifactKind, StoryState>> = {
  spec: 'developing',
  design: 'developing',
  arch: 'developing',
  validators: 'developing',
  guidelines: 'developing',
  tickets: 'developing',
  'dispatch-plan': 'developing',
  mockup: 'developing',
  audit: 'done',
};

export interface EditPermission {
  editable: boolean;
  /** Si !editable, razón human-readable mostrada en tooltip. */
  reason?: string;
  /** Si append-only (operator-input, capability), UI debe usar componente dedicado. */
  appendOnly?: boolean;
}

export interface PermissionInput {
  kind: ArtifactKind;
  storyState: StoryState;
  isArchived: boolean;
  /** El descriptor vivo (useProceso). null mientras carga → lock defensivo. */
  proceso: ProcesoResponse | null;
}

/**
 * Calcula si un artifact es editable dado el state actual de la story.
 * Esta es la SSoT de la regla: callers (MarkdownViewer, YamlEditor,
 * CheckpointTab, ArtifactTab) consultan acá para decidir si renderizan
 * el toggle "Editar" o solo "Ver/Externo".
 */
export function getEditPermission({
  kind,
  storyState,
  isArchived,
  proceso,
}: PermissionInput): EditPermission {
  // Archive: todo read-only (R2 sistema-docs-schema · stories done frozen)
  if (isArchived) {
    return {
      editable: false,
      reason: 'Archivo en archive/ — stories done son immutable (R2).',
    };
  }

  switch (kind) {
    case 'checkpoint':
      return { editable: true };

    case 'operator-input':
      return {
        editable: true,
        appendOnly: true,
        reason: 'operator-input es append-only · usa el panel dedicado.',
      };

    case 'capability':
      return {
        editable: true,
        appendOnly: true,
        reason:
          'capabilities ledger es append-only via change_log[] · usa "Extender" del Cap Drawer.',
      };

    case 'ticket-log':
      return {
        editable: false,
        reason: 'T-{n}-* son append-only del dev-team/auditor · no editable desde cockpit.',
      };

    case 'merge':
      return {
        editable: false,
        reason: '07-merge.md es terminal · cerró la story (estado done).',
      };

    case 'release':
    case 'learning':
      // Estos no tienen story-state asociado · editables siempre
      return { editable: true };

    case 'unknown':
      return {
        editable: false,
        reason: 'Tipo de archivo desconocido · usa el editor externo si necesitás cambiarlo.',
      };
  }

  // Resto: spec/design/arch/validators/guidelines/tickets/dispatch-plan/mockup/audit
  // → editable si editOrden(state) < editOrden(lockFrom) — orden DEL DESCRIPTOR.
  if (!proceso) {
    // Descriptor aún no cargado: lock defensivo momentáneo (nunca editable "de más").
    return { editable: false, reason: 'Descriptor de proceso cargando…' };
  }

  if (estadoDe(proceso, storyState)?.categoria === 'descartado') {
    return {
      editable: false,
      reason: 'Story descartada · todo read-only excepto checkpoint.',
    };
  }

  const lockFrom = LOCK_FROM_STATE[kind];
  if (!lockFrom) {
    // No debería llegar acá — switch arriba cubre todos los kinds. Defensivo:
    return { editable: false, reason: 'Sin política definida.' };
  }

  const editable = editOrden(proceso, storyState) < editOrden(proceso, lockFrom);
  if (editable) return { editable: true };

  return {
    editable: false,
    reason:
      `${humanKind(kind)} no editable cuando story.state ∈ {${chainFrom(proceso, lockFrom).join(', ')}}. ` +
      `Para cambios: crear delta-spec.md o nueva story (paradigm v4).`,
  };
}

function chainFrom(p: ProcesoResponse, s: StoryState): string[] {
  const order = editOrden(p, s);
  return p.estados
    .filter((e) => {
      const o = editOrden(p, e.id);
      return o >= order && Number.isFinite(o);
    })
    .map((e) => e.id);
}

function humanKind(k: ArtifactKind): string {
  const map: Record<ArtifactKind, string> = {
    checkpoint: 'checkpoint.md',
    'operator-input': 'operator-input.md',
    spec: '01-spec.md',
    design: '02-design-*.md',
    arch: '03-arch.md',
    validators: '04-validators.yaml',
    guidelines: '05-guidelines.md',
    tickets: '06-tickets.yaml',
    'dispatch-plan': 'dispatch-plan.md',
    mockup: 'mockups/*',
    'ticket-log': 'T-{n}-*.md',
    audit: 'audit/*',
    merge: '07-merge.md',
    capability: 'capability YAML',
    release: 'release YAML',
    learning: 'learning .md',
    unknown: 'archivo',
  };
  return map[k];
}
