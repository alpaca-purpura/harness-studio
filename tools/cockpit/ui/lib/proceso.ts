/**
 * proceso.ts — helpers PUROS sobre el descriptor de proceso (F6 · RN-50).
 *
 * El board y el drawer ya no hardcodean el ciclo: TODO se deriva de la respuesta de
 * `/api/proceso` (ProcesoProvider hace el fetch; estas funciones derivan). Cero
 * literales de estado: comportamiento genérico SIEMPRE por categoría (contrato L0),
 * jamás por nombre — otro descriptor = otro proceso, cero cambio en estos helpers.
 * Testeado con fixture default + alterno en `__tests__/proceso.test.ts` (RN-54).
 */

import type {
  ProcesoBinding,
  ProcesoEstado,
  ProcesoResponse,
  ProcesoTransicion,
} from './types';

/** Orden canónico de render = el orden de `estados[]` (RN-28). */
export function statesOrder(p: ProcesoResponse): string[] {
  return p.estados.map((e) => e.id);
}

/** El ciclo feliz: estados que NO son pausa ni descarte — por CATEGORÍA. */
export function happyPath(p: ProcesoResponse): ProcesoEstado[] {
  return p.estados.filter(
    (e) => e.categoria !== 'pausado' && e.categoria !== 'descartado'
  );
}

export function estadoDe(p: ProcesoResponse, id: string): ProcesoEstado | undefined {
  return p.estados.find((e) => e.id === id);
}

/** Terminal por categoría (derivado por el motor y servido en el estado — RN-31). */
export function esTerminal(p: ProcesoResponse, id: string): boolean {
  return estadoDe(p, id)?.terminal ?? false;
}

/** Límite WIP del estado; undefined = sin límite (RN-48). */
export function wipDe(p: ProcesoResponse, id: string): number | undefined {
  const w = estadoDe(p, id)?.wip;
  return w && w > 0 ? w : undefined;
}

/** Las transiciones que el OPERADOR ejecuta desde la consola. */
export function transicionesOperador(p: ProcesoResponse): ProcesoTransicion[] {
  return p.transiciones.filter((t) => t.ejecutor === 'operador');
}

export function isOperatorAllowed(
  p: ProcesoResponse,
  from: string,
  to: string
): ProcesoTransicion | null {
  return (
    transicionesOperador(p).find((t) => t.de === from && t.a === to) ?? null
  );
}

/** Etiqueta humana de la acción: nombre declarado → verbo → flecha (RN-48). */
export function labelTransicion(t: ProcesoTransicion): string {
  return t.nombre ?? t.verbo ?? `→ ${t.a}`;
}

function renderBinding(b: ProcesoBinding): string {
  const quien = b.arnes || b.rol;
  return b.nota ? `${quien} (${b.nota})` : quien;
}

/**
 * Dueños del estado como string de owner — join `" o "` byte-igual al render del
 * motor Go (`duenosRender`, RN-31). null = sin binding declarado.
 */
export function duenoRender(p: ProcesoResponse, estado: string): string | null {
  const bindings = p.duenos[estado] ?? [];
  if (bindings.length === 0) return null;
  return bindings.map(renderBinding).join(' o ');
}

/**
 * Quién ejecuta la SALIDA de un estado: el dueño del destino de su primera
 * transición de rol (los skills transicionan ESCRIBIENDO el estado destino).
 */
export function ejecutorSalida(p: ProcesoResponse, estado: string): string | null {
  const t = p.transiciones.find((x) => x.de === estado && x.ejecutor === 'rol');
  return t ? duenoRender(p, t.a) : null;
}

/**
 * Orden de edición (edit-permissions): índice en `estados[]`, con dos reglas por
 * CATEGORÍA que reproducen el STATE_ORDER legacy — `pausado` edita como el estado
 * inicial (la pausa no consume artefactos) y `descartado` es infinito (todo locked).
 * Estado desconocido → infinito (defensivo: nunca editable).
 */
export function editOrden(p: ProcesoResponse, estado: string): number {
  const e = estadoDe(p, estado);
  if (!e) return Number.POSITIVE_INFINITY;
  if (e.categoria === 'descartado') return Number.POSITIVE_INFINITY;
  if (e.categoria === 'pausado') {
    const inicial = p.estados.findIndex((x) => x.inicial);
    return inicial >= 0 ? inicial : 0;
  }
  return p.estados.findIndex((x) => x.id === estado);
}

/**
 * El estado donde el OPERADOR tiene un gate de momento-estado (el "gate G" del
 * sdd-default): primer gate con autoridad operador cuyo momento sea un estado.
 */
export function gateOperadorEstado(p: ProcesoResponse): string | null {
  for (const g of p.gates) {
    if (g.autoridad.rol !== 'operador') continue;
    const m = g.momento.find((x) => !x.includes('→') && estadoDe(p, x));
    if (m) return m;
  }
  return null;
}

/** El estado inicial declarado (exactamente uno por contrato). */
export function estadoInicial(p: ProcesoResponse): ProcesoEstado | undefined {
  return p.estados.find((e) => e.inicial);
}
