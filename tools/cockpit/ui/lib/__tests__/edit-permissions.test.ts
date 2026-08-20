/**
 * edit-permissions.ts tests · matriz artifact × state.
 *
 * F6/RN-54: el ORDEN de estados ya no vive hardcodeado en edit-permissions — se deriva
 * del descriptor. La matriz se conserva; cada caso pasa el fixture del descriptor
 * (espejo sdd-default) en vez de asumir el STATE_ORDER literal.
 */

import { describe, it, expect } from 'vitest';
import {
  getArtifactKind,
  getEditPermission,
  type PermissionInput,
} from '../edit-permissions.js';
import type { StoryState } from '../types.js';
import { PROCESO_DEFAULT } from './proceso.fixture.js';

/** La matriz corre contra el descriptor default (round-trip con el hardcode viejo). */
function gp(input: Omit<PermissionInput, 'proceso'>) {
  return getEditPermission({ ...input, proceso: PROCESO_DEFAULT });
}

describe('getArtifactKind', () => {
  it('detecta checkpoint y operator-input (incl. filename legacy del kit)', () => {
    expect(getArtifactKind('checkpoint.md')).toBe('checkpoint');
    expect(getArtifactKind('operator-input.md')).toBe('operator-input');
    expect(getArtifactKind('chris-input.md')).toBe('operator-input'); // legacy F-1
  });

  it('detecta spec/design/arch variantes', () => {
    expect(getArtifactKind('01-spec.md')).toBe('spec');
    expect(getArtifactKind('02-design-ui.md')).toBe('design');
    expect(getArtifactKind('02-design-agentic.md')).toBe('design');
    expect(getArtifactKind('03-arch.md')).toBe('arch');
    expect(getArtifactKind('03-arch-be.md')).toBe('arch');
    expect(getArtifactKind('03-arch-fe.md')).toBe('arch');
  });

  it('detecta validators/guidelines/tickets', () => {
    expect(getArtifactKind('04-validators.yaml')).toBe('validators');
    expect(getArtifactKind('05-guidelines.md')).toBe('guidelines');
    expect(getArtifactKind('06-tickets.yaml')).toBe('tickets');
  });

  it('detecta ticket-log read-only', () => {
    expect(getArtifactKind('T-1-impl-log.md')).toBe('ticket-log');
    expect(getArtifactKind('T-3-result.md')).toBe('ticket-log');
    expect(getArtifactKind('T-12-review.md')).toBe('ticket-log');
    expect(getArtifactKind('T-2-handoff.md')).toBe('ticket-log');
  });

  it('detecta audit + merge', () => {
    expect(getArtifactKind('06-audit/gherkin-matrix.md')).toBe('audit');
    expect(getArtifactKind('CHECKPOINTS.md')).toBe('audit');
    expect(getArtifactKind('07-merge.md')).toBe('merge');
  });

  it('detecta mockups y dispatch-plan', () => {
    expect(getArtifactKind('mockups/foo.html')).toBe('mockup');
    expect(getArtifactKind('mockups/v2/bar.png')).toBe('mockup');
    expect(getArtifactKind('dispatch-plan.md')).toBe('dispatch-plan');
  });

  it('detecta capability/release/learning desde paths sistema', () => {
    expect(
      getArtifactKind('main/docs/product/capabilities/scheduling/agenda.yaml')
    ).toBe('capability');
    expect(getArtifactKind('main/docs/product/releases/F2.yaml')).toBe('release');
    expect(getArtifactKind('main/docs/learnings/2026-05-27-auth.md')).toBe(
      'learning'
    );
  });

  it('unknown para archivos raros', () => {
    expect(getArtifactKind('weird-file.txt')).toBe('unknown');
  });
});

describe('getEditPermission · checkpoint siempre editable', () => {
  const allStates: StoryState[] = [
    'idea',
    'refining',
    'refined',
    'ready',
    'developing',
    'developed',
    'reviewing',
    'done',
    'parked',
    'dropped',
  ];

  it('editable en todos los estados (live)', () => {
    for (const s of allStates) {
      const p = gp({
        kind: 'checkpoint',
        storyState: s,
        isArchived: false,
      });
      expect(p.editable, `state=${s}`).toBe(true);
    }
  });

  it('NO editable si está en archive', () => {
    const p = gp({
      kind: 'checkpoint',
      storyState: 'done',
      isArchived: true,
    });
    expect(p.editable).toBe(false);
  });
});

describe('getEditPermission · spec/design/arch locked from developing', () => {
  const beforeLockStates: StoryState[] = ['idea', 'refining', 'refined', 'ready'];
  const afterLockStates: StoryState[] = ['developing', 'developed', 'reviewing', 'done'];
  const lockedKinds = [
    'spec',
    'design',
    'arch',
    'validators',
    'guidelines',
    'tickets',
    'dispatch-plan',
    'mockup',
  ] as const;

  for (const kind of lockedKinds) {
    it(`${kind}: editable en idea/refining/refined/ready`, () => {
      for (const s of beforeLockStates) {
        const p = gp({ kind, storyState: s, isArchived: false });
        expect(p.editable, `${kind} @ ${s}`).toBe(true);
      }
    });

    it(`${kind}: LOCKED en developing/developed/reviewing/done`, () => {
      for (const s of afterLockStates) {
        const p = gp({ kind, storyState: s, isArchived: false });
        expect(p.editable, `${kind} @ ${s}`).toBe(false);
        expect(p.reason, `${kind} @ ${s} reason`).toMatch(/no editable/i);
      }
    });
  }

  it('parked: editable (como idea · puede reactivarse)', () => {
    for (const kind of lockedKinds) {
      const p = gp({ kind, storyState: 'parked', isArchived: false });
      expect(p.editable, `${kind} @ parked`).toBe(true);
    }
  });

  it('dropped: NO editable', () => {
    for (const kind of lockedKinds) {
      const p = gp({ kind, storyState: 'dropped', isArchived: false });
      expect(p.editable, `${kind} @ dropped`).toBe(false);
    }
  });
});

describe('getEditPermission · audit locked from done', () => {
  it('editable en reviewing', () => {
    const p = gp({
      kind: 'audit',
      storyState: 'reviewing',
      isArchived: false,
    });
    expect(p.editable).toBe(true);
  });

  it('LOCKED en done', () => {
    const p = gp({
      kind: 'audit',
      storyState: 'done',
      isArchived: false,
    });
    expect(p.editable).toBe(false);
  });
});

describe('getEditPermission · read-only siempre', () => {
  it('ticket-log read-only en todos los estados', () => {
    for (const s of ['developing', 'developed', 'reviewing', 'done'] as StoryState[]) {
      const p = gp({
        kind: 'ticket-log',
        storyState: s,
        isArchived: false,
      });
      expect(p.editable).toBe(false);
      expect(p.reason).toMatch(/append-only/i);
    }
  });

  it('merge read-only', () => {
    const p = gp({
      kind: 'merge',
      storyState: 'done',
      isArchived: false,
    });
    expect(p.editable).toBe(false);
  });
});

describe('getEditPermission · append-only flags', () => {
  it('operator-input editable=true pero appendOnly=true', () => {
    const p = gp({
      kind: 'operator-input',
      storyState: 'refining',
      isArchived: false,
    });
    expect(p.editable).toBe(true);
    expect(p.appendOnly).toBe(true);
  });

  it('capability append-only', () => {
    const p = gp({
      kind: 'capability',
      storyState: 'idea',
      isArchived: false,
    });
    expect(p.editable).toBe(true);
    expect(p.appendOnly).toBe(true);
  });
});

describe('getEditPermission · sin descriptor cargado (proceso: null)', () => {
  it('checkpoint sigue editable (no depende del orden)', () => {
    const p = getEditPermission({
      kind: 'checkpoint',
      storyState: 'refining',
      isArchived: false,
      proceso: null,
    });
    expect(p.editable).toBe(true);
  });

  it('spec cae a lock defensivo (nunca editable "de más" mientras carga)', () => {
    const p = getEditPermission({
      kind: 'spec',
      storyState: 'idea',
      isArchived: false,
      proceso: null,
    });
    expect(p.editable).toBe(false);
    expect(p.reason).toMatch(/cargando/i);
  });
});

describe('getEditPermission · release/learning editable siempre', () => {
  it('release editable independiente del story state', () => {
    const p = gp({
      kind: 'release',
      storyState: 'done',
      isArchived: false,
    });
    expect(p.editable).toBe(true);
  });

  it('learning editable independiente del story state', () => {
    const p = gp({
      kind: 'learning',
      storyState: 'developing',
      isArchived: false,
    });
    expect(p.editable).toBe(true);
  });
});
