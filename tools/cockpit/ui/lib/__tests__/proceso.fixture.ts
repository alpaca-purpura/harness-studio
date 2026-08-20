/**
 * Fixtures del descriptor de proceso para los tests de contrato (F6/RN-54).
 *
 * PROCESO_DEFAULT = espejo de la instancia sdd-default que shipea el kit (si el kit
 * cambia su instancia, estos literales avisan que el comportamiento observable cambió —
 * mismo rol que el round-trip legacy de proceso_test.go en el lado Go).
 * PROCESO_ALTERNO = "otra empresa, otro descriptor": estados inventados, mismas
 * categorías del contrato — la prueba de que los helpers no tienen literales del ciclo.
 */

import type { ProcesoResponse } from '../types.js';

export const PROCESO_DEFAULT: ProcesoResponse = {
  descriptor: { id: 'sdd-default', nombre: 'Ciclo SDD (default del kit)', version: 2 },
  categorias: ['propuesto', 'en-progreso', 'completado', 'descartado', 'pausado'],
  estados: [
    { id: 'idea', categoria: 'propuesto', inicial: true, descripcion: 'capturada en el backlog', terminal: false },
    { id: 'refining', categoria: 'en-progreso', wip: 3, descripcion: 'refinamiento en curso', terminal: false },
    { id: 'refined', categoria: 'en-progreso', wip: 5, descripcion: 'spec cerrada y ratificada', terminal: false },
    { id: 'ready', categoria: 'en-progreso', wip: 5, descripcion: 'ready package completo', terminal: false },
    { id: 'developing', categoria: 'en-progreso', wip: 3, descripcion: 'construcción TDD en curso', terminal: false },
    { id: 'developed', categoria: 'en-progreso', wip: 1, descripcion: 'validators verdes', terminal: false },
    { id: 'reviewing', categoria: 'en-progreso', wip: 1, descripcion: 'auditoría en curso', terminal: false },
    { id: 'done', categoria: 'completado', descripcion: 'mergeada y entregada', terminal: true },
    { id: 'parked', categoria: 'pausado', descripcion: 'pausada con razón explícita', terminal: false },
    { id: 'dropped', categoria: 'descartado', descripcion: 'descartada con razón explícita', terminal: true },
  ],
  transiciones: [
    { de: 'idea', a: 'refining', ejecutor: 'operador', verbo: 'spec.started', nombre: 'Empezar a refinar' },
    { de: 'idea', a: 'parked', ejecutor: 'operador', verbo: 'story.parked', nombre: 'Parquear', requiere_razon: true },
    { de: 'idea', a: 'dropped', ejecutor: 'operador', verbo: 'story.dropped', nombre: 'Descartar', requiere_razon: true },
    { de: 'refining', a: 'idea', ejecutor: 'operador', verbo: 'story.backlogged', nombre: 'Volver a backlog' },
    { de: 'refining', a: 'parked', ejecutor: 'operador', verbo: 'story.parked', nombre: 'Parquear', requiere_razon: true },
    { de: 'refining', a: 'dropped', ejecutor: 'operador', verbo: 'story.dropped', nombre: 'Descartar', requiere_razon: true },
    { de: 'parked', a: 'idea', ejecutor: 'operador', verbo: 'story.reactivated', nombre: 'Reactivar' },
    { de: 'refining', a: 'refined', ejecutor: 'rol', verbo: 'story.refined' },
    { de: 'refined', a: 'ready', ejecutor: 'rol', verbo: 'story.queued' },
    { de: 'ready', a: 'developing', ejecutor: 'rol', verbo: 'build.started' },
    { de: 'developing', a: 'developed', ejecutor: 'rol', verbo: 'build.finished' },
    { de: 'developed', a: 'reviewing', ejecutor: 'rol', verbo: 'review.started' },
    { de: 'reviewing', a: 'done', ejecutor: 'rol', verbo: 'story.merged' },
  ],
  gates: [
    {
      id: 'verificacion-operador',
      nombre: 'Verificación del operador (gate G)',
      momento: ['developed'],
      autoridad: { rol: 'operador' },
      checklist: ['signoff registrado en el checkpoint'],
    },
    {
      id: 'razon-de-cierre',
      nombre: 'Razón de cierre',
      momento: ['idea→parked', 'idea→dropped', 'refining→parked', 'refining→dropped'],
      autoridad: { rol: 'operador' },
      checklist: ['razón explícita registrada'],
    },
  ],
  duenos: {
    refined: [
      { rol: 'architect', arnes: '/architect', nota: 'cierra spec' },
      { rol: 'pm', arnes: '/pm-{sistema}', nota: 'ratifica' },
    ],
    ready: [{ rol: 'architect', arnes: '/architect', nota: 'ready package completo' }],
    developing: [{ rol: 'dev-team', arnes: '/dev-team', nota: 'builder spawn' }],
    developed: [{ rol: 'dev-team', arnes: '/dev-team', nota: 'validators GREEN' }],
    reviewing: [{ rol: 'auditor', arnes: '/auditor', nota: 'AUTO-HANDOFF desde developed' }],
    done: [{ rol: 'pm', arnes: '/pm-{sistema}', nota: 'Fase F MERGE' }],
  },
  parametros: { razon_minima: 10 },
  fuente: 'fixture (espejo sdd-default)',
};

export const PROCESO_ALTERNO: ProcesoResponse = {
  descriptor: { id: 'cto-prod', nombre: 'Ciclo CTO con paso a producción', version: 1 },
  categorias: ['propuesto', 'en-progreso', 'completado', 'descartado', 'pausado'],
  estados: [
    { id: 'triage', categoria: 'propuesto', inicial: true, terminal: false },
    { id: 'build', categoria: 'en-progreso', wip: 2, terminal: false },
    { id: 'staging', categoria: 'en-progreso', terminal: false },
    { id: 'prod', categoria: 'completado', terminal: true },
    { id: 'killed', categoria: 'descartado', terminal: true },
    { id: 'frozen', categoria: 'pausado', terminal: false },
  ],
  transiciones: [
    { de: 'triage', a: 'build', ejecutor: 'operador', verbo: 'work.started' },
    { de: 'triage', a: 'killed', ejecutor: 'operador', requiere_razon: true },
    { de: 'build', a: 'staging', ejecutor: 'rol', verbo: 'release.staged' },
    { de: 'staging', a: 'prod', ejecutor: 'rol', verbo: 'release.shipped' },
  ],
  gates: [
    {
      id: 'go-live',
      nombre: 'Paso a producción',
      momento: ['staging'],
      autoridad: { rol: 'sre', arnes: '/sre' },
      checklist: ['smoke en staging verde'],
    },
  ],
  duenos: {
    staging: [
      { rol: 'release-eng', arnes: '/release', nota: 'empaqueta' },
      { rol: 'sre', arnes: '/sre', nota: 'verifica' },
    ],
  },
  parametros: { razon_minima: 5 },
  fuente: 'fixture (alterno)',
};
