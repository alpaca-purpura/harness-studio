/**
 * proceso.ts tests — el contrato de derivación (F6/RN-54).
 *
 * Reemplazan a los tests que fijaban hardcodes del ciclo en la UI: acá se prueba que
 * los helpers derivan TODO del descriptor — con el espejo del default (round-trip
 * contra lo que el board mostraba hardcodeado) y con un descriptor ALTERNO de estados
 * inventados ("otra empresa, otro descriptor, cero cambio de consola").
 */

import { describe, it, expect } from 'vitest';
import {
  duenoRender,
  editOrden,
  ejecutorSalida,
  esTerminal,
  estadoInicial,
  gateOperadorEstado,
  happyPath,
  isOperatorAllowed,
  labelTransicion,
  statesOrder,
  transicionesOperador,
  wipDe,
} from '../proceso.js';
import { PROCESO_ALTERNO, PROCESO_DEFAULT } from './proceso.fixture.js';

describe('statesOrder + happyPath', () => {
  it('orden canónico = orden del array (round-trip con STATES_ORDER legacy)', () => {
    expect(statesOrder(PROCESO_DEFAULT)).toEqual([
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
    ]);
  });

  it('happy path excluye pausado/descartado POR CATEGORÍA (round-trip MAIN_FLOW)', () => {
    expect(happyPath(PROCESO_DEFAULT).map((e) => e.id)).toEqual([
      'idea',
      'refining',
      'refined',
      'ready',
      'developing',
      'developed',
      'reviewing',
      'done',
    ]);
    expect(happyPath(PROCESO_ALTERNO).map((e) => e.id)).toEqual([
      'triage',
      'build',
      'staging',
      'prod',
    ]);
  });

  it('estado inicial = el declarado', () => {
    expect(estadoInicial(PROCESO_DEFAULT)?.id).toBe('idea');
    expect(estadoInicial(PROCESO_ALTERNO)?.id).toBe('triage');
  });
});

describe('wip + terminal', () => {
  it('wip solo donde el descriptor lo declara (round-trip WIP_CAPS legacy)', () => {
    expect(wipDe(PROCESO_DEFAULT, 'refining')).toBe(3);
    expect(wipDe(PROCESO_DEFAULT, 'refined')).toBe(5);
    expect(wipDe(PROCESO_DEFAULT, 'developed')).toBe(1);
    expect(wipDe(PROCESO_DEFAULT, 'idea')).toBeUndefined();
    expect(wipDe(PROCESO_DEFAULT, 'done')).toBeUndefined();
  });

  it('terminal viene derivado por categoría', () => {
    expect(esTerminal(PROCESO_DEFAULT, 'done')).toBe(true);
    expect(esTerminal(PROCESO_DEFAULT, 'dropped')).toBe(true);
    expect(esTerminal(PROCESO_DEFAULT, 'parked')).toBe(false);
    expect(esTerminal(PROCESO_ALTERNO, 'prod')).toBe(true);
    expect(esTerminal(PROCESO_ALTERNO, 'bogus')).toBe(false);
  });
});

describe('transiciones de operador (round-trip OPERATOR_ALLOWED_TRANSITIONS legacy)', () => {
  it('whitelist exacta del default', () => {
    const pares = transicionesOperador(PROCESO_DEFAULT).map((t) => `${t.de}→${t.a}`);
    expect(pares).toEqual([
      'idea→refining',
      'idea→parked',
      'idea→dropped',
      'refining→idea',
      'refining→parked',
      'refining→dropped',
      'parked→idea',
    ]);
  });

  it('isOperatorAllowed + requiere_razon', () => {
    expect(isOperatorAllowed(PROCESO_DEFAULT, 'idea', 'refining')?.requiere_razon).toBeFalsy();
    expect(isOperatorAllowed(PROCESO_DEFAULT, 'idea', 'parked')?.requiere_razon).toBe(true);
    expect(isOperatorAllowed(PROCESO_DEFAULT, 'refined', 'ready')).toBeNull(); // de rol, no de operador
    expect(isOperatorAllowed(PROCESO_ALTERNO, 'triage', 'build')).not.toBeNull();
  });

  it('labels: nombre → verbo → flecha (round-trip verbs legacy)', () => {
    expect(labelTransicion(isOperatorAllowed(PROCESO_DEFAULT, 'idea', 'refining')!)).toBe(
      'Empezar a refinar'
    );
    expect(labelTransicion(isOperatorAllowed(PROCESO_DEFAULT, 'parked', 'idea')!)).toBe(
      'Reactivar'
    );
    // alterno sin nombre → cae al verbo
    expect(labelTransicion(isOperatorAllowed(PROCESO_ALTERNO, 'triage', 'build')!)).toBe(
      'work.started'
    );
    // sin nombre ni verbo → flecha al destino
    expect(labelTransicion(isOperatorAllowed(PROCESO_ALTERNO, 'triage', 'killed')!)).toBe(
      '→ killed'
    );
  });
});

describe('dueños (round-trip STATE_OWNERS legacy — byte-igual al motor Go)', () => {
  it('join " o " con nota', () => {
    expect(duenoRender(PROCESO_DEFAULT, 'refined')).toBe(
      '/architect (cierra spec) o /pm-{sistema} (ratifica)'
    );
    expect(duenoRender(PROCESO_DEFAULT, 'done')).toBe('/pm-{sistema} (Fase F MERGE)');
    expect(duenoRender(PROCESO_DEFAULT, 'idea')).toBeNull();
    expect(duenoRender(PROCESO_ALTERNO, 'staging')).toBe(
      '/release (empaqueta) o /sre (verifica)'
    );
  });

  it('ejecutorSalida = dueño del destino de la transición de rol', () => {
    expect(ejecutorSalida(PROCESO_DEFAULT, 'ready')).toBe('/dev-team (builder spawn)');
    expect(ejecutorSalida(PROCESO_DEFAULT, 'reviewing')).toBe('/pm-{sistema} (Fase F MERGE)');
    expect(ejecutorSalida(PROCESO_DEFAULT, 'done')).toBeNull(); // sin salida de rol
  });
});

describe('editOrden (round-trip STATE_ORDER de edit-permissions legacy)', () => {
  it('índice del array para el ciclo feliz', () => {
    expect(editOrden(PROCESO_DEFAULT, 'idea')).toBe(0);
    expect(editOrden(PROCESO_DEFAULT, 'developing')).toBe(4);
    expect(editOrden(PROCESO_DEFAULT, 'done')).toBe(7);
  });

  it('pausado edita como el inicial · descartado/desconocido = ∞', () => {
    expect(editOrden(PROCESO_DEFAULT, 'parked')).toBe(0);
    expect(editOrden(PROCESO_DEFAULT, 'dropped')).toBe(Number.POSITIVE_INFINITY);
    expect(editOrden(PROCESO_DEFAULT, 'bogus')).toBe(Number.POSITIVE_INFINITY);
    expect(editOrden(PROCESO_ALTERNO, 'frozen')).toBe(0);
    expect(editOrden(PROCESO_ALTERNO, 'killed')).toBe(Number.POSITIVE_INFINITY);
  });
});

describe('gateOperadorEstado (el "gate G" derivado, no literal)', () => {
  it('primer gate con autoridad operador y momento-estado', () => {
    expect(gateOperadorEstado(PROCESO_DEFAULT)).toBe('developed');
  });

  it('alterno sin gate de operador → null', () => {
    expect(gateOperadorEstado(PROCESO_ALTERNO)).toBeNull();
  });
});
