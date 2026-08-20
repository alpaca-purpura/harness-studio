/**
 * operator-input-parser tests · round-trip idempotente + compat con el
 * formato legacy del kit (author `chris`, filename `chris-input.md` // legacy F-1).
 */

import { describe, expect, it } from 'vitest';
import {
  parseOperatorInput,
  serializeOperatorInput,
  isOperatorInputFilename,
} from '../operator-input-parser.js';
import type { OperatorInput } from '../types.js';

describe('operator-input-parser', () => {
  it('round-trip idempotente: parse(serialize(data)) === data', async () => {
    // Construir un OperatorInput manualmente con todos los tipos
    const original: OperatorInput = {
      frontmatter: {
        story_id: 'test-story',
        created_at: '2026-05-27T10:00:00-05:00',
        last_modified: '2026-05-27T12:00:00-05:00',
        notes_count: 2,
        refs_count: 3,
        conversation_count: 3,
      },
      notes: [
        { timestamp: '2026-05-27 10:00', text: 'Primera nota · línea 1\nlínea 2 de la nota.' },
        { timestamp: '2026-05-27 11:00', text: 'Segunda nota más corta.' },
      ],
      refs: [
        { type: 'link', value: 'https://example.com', comment: 'comentario opcional' },
        { type: 'img', value: 'refs/foo.png' },
        { type: 'learning-ref', value: '2026-05-18-pattern-x', comment: 'reusar' },
      ],
      conversation: [
        { timestamp: '2026-05-27 10:00', author: 'operador', text: 'mi pregunta' },
        {
          timestamp: '2026-05-27 10:15',
          author: 'claude',
          skill: 'po-ux',
          verdict: 'applied',
          text: 'apliqué cambios al 01-spec.md',
        },
        {
          timestamp: '2026-05-27 11:00',
          author: 'claude',
          skill: 'po-ux',
          verdict: 'doubt',
          text: '¿confirmas X?',
        },
      ],
    };

    const serialized = serializeOperatorInput(original);
    const reparsed = parseOperatorInput(serialized);

    // Frontmatter (los counts deben matchear los actuales)
    expect(reparsed.frontmatter.story_id).toBe(original.frontmatter.story_id);
    expect(reparsed.frontmatter.notes_count).toBe(original.notes.length);
    expect(reparsed.frontmatter.refs_count).toBe(original.refs.length);
    expect(reparsed.frontmatter.conversation_count).toBe(original.conversation.length);

    // Notes idempotentes
    expect(reparsed.notes).toEqual(original.notes);

    // Refs idempotentes
    expect(reparsed.refs).toEqual(original.refs);

    // Conversation idempotente
    expect(reparsed.conversation).toEqual(original.conversation);

    // Round-trip 2x debe ser stable
    const serialized2 = serializeOperatorInput(reparsed);
    expect(serialized2).toBe(serialized);
  });

  it('parsea un archivo legacy del kit (author chris) y lo normaliza a operador', () => {
    // Fixture inline con el formato exacto del template legacy del kit
    // (HTML comment + frontmatter + 3 secciones + author `chris`). // legacy F-1
    const legacyRaw = `<!-- voseo-allowed: doc interno -->

---
story_id: acme-main-checkout
created_at: 2026-05-27T10:00:00-05:00
last_modified: 2026-05-27T12:00:00-05:00
notes_count: 1
refs_count: 2
conversation_count: 2
---

## 💭 Notas

### 2026-05-27 14:30
El estado de pago se vea PROMINENTE en la card.

## 📎 Referencias

- **🔗 link** · https://example.com/segmentos
  > me gusta cómo segmentan acá
- **📚 learning-ref** · 2026-05-18-repository-base

## 💬 Conversación

### 2026-05-27 14:30 · 🧑 chris
¿Podemos sumar el estado de pago al listado?

### 2026-05-27 15:00 · 🤖 claude · \`/po-ux\` · ✓ APLICADO
Agregado al spec como SC-08.
`;

    const data = parseOperatorInput(legacyRaw);

    expect(data.frontmatter.story_id).toBe('acme-main-checkout');
    expect(data.notes).toHaveLength(1);
    expect(data.notes[0].text).toContain('PROMINENTE');
    expect(data.refs).toHaveLength(2);
    expect(data.refs[0]).toEqual({
      type: 'link',
      value: 'https://example.com/segmentos',
      comment: 'me gusta cómo segmentan acá',
    });
    expect(data.refs[1].type).toBe('learning-ref');

    // Author legacy `chris` → normalizado a `operador` al parsear
    expect(data.conversation).toHaveLength(2);
    expect(data.conversation[0].author).toBe('operador');
    expect(data.conversation[1].author).toBe('claude');
    expect(data.conversation[1].verdict).toBe('applied');

    // Preserva el comentario HTML <!-- voseo-allowed -->
    expect(data.preamble).toContain('voseo-allowed');

    // Round-trip: se serializa con `operador` (no `chris`) y re-parsea estable
    const serialized = serializeOperatorInput(data);
    expect(serialized).toContain('· 🧑 operador');
    expect(serialized).not.toContain('· 🧑 chris');
    expect(serialized.startsWith('<!--')).toBe(true);
    const reparsed = parseOperatorInput(serialized);
    expect(reparsed.notes).toEqual(data.notes);
    expect(reparsed.refs).toEqual(data.refs);
    expect(reparsed.conversation).toEqual(data.conversation);
  });

  it('reconoce el filename nuevo y el legacy del kit', () => {
    expect(isOperatorInputFilename('operator-input.md')).toBe(true);
    expect(isOperatorInputFilename('chris-input.md')).toBe(true); // legacy F-1
    expect(isOperatorInputFilename('checkpoint.md')).toBe(false);
  });
});
