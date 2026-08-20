/**
 * proceso-server.ts — el descriptor de proceso para las routes dev-mode (F6/RN-51).
 *
 * Paridad `pnpm dev` de lo que el binario embebe con go:embed: lee el MISMO espejo
 * gated (`../go/process/sdd-default.yaml`) y expone las derivaciones que las routes
 * necesitan (whitelist de operador, dueños, razón mínima). El Go es quien sirve en
 * el binario (handlers_proceso.go / handlers_stories.go); esto solo existe para dev.
 * Server-only (node:fs) — NO importar desde componentes de cliente.
 */

import path from 'node:path';
import { readFile } from 'node:fs/promises';
import { parse } from 'yaml';

export const CATEGORIAS = [
  'propuesto',
  'en-progreso',
  'completado',
  'descartado',
  'pausado',
] as const;
export const TERMINALES = new Set(['completado', 'descartado']);

export interface EspejoTransicion {
  de: string;
  a: string;
  ejecutor: 'operador' | 'rol';
  requiere_razon?: boolean;
  verbo?: string;
  nombre?: string;
}

export interface EspejoDescriptor {
  descriptor: Record<string, unknown>;
  estados: Array<Record<string, unknown> & { id: string; categoria: string }>;
  transiciones: EspejoTransicion[];
  gates: Array<Record<string, unknown> & { momento: string[] }>;
  duenos: Record<string, Array<{ rol: string; arnes?: string; nota?: string }>>;
  parametros: { razon_minima: number };
}

/** Lee y normaliza el espejo gated (throw si ilegible — la route responde 500). */
export async function leerEspejo(): Promise<EspejoDescriptor> {
  const espejo = path.resolve(process.cwd(), '../go/process/sdd-default.yaml');
  const doc = parse(await readFile(espejo, 'utf-8')) as Record<string, any>;
  return {
    descriptor: doc.descriptor ?? {},
    estados: doc.estados ?? [],
    transiciones: doc.transiciones ?? [],
    gates: (doc.gates ?? []).map((g: Record<string, any>) => ({
      ...g,
      momento: Array.isArray(g.momento) ? g.momento : [g.momento],
    })),
    duenos: doc.duenos ?? {},
    parametros: doc.parametros ?? { razon_minima: 10 },
  };
}

export function esEstado(p: EspejoDescriptor, id: string): boolean {
  return p.estados.some((e) => e.id === id);
}

export function transicionOperador(
  p: EspejoDescriptor,
  de: string,
  a: string
): EspejoTransicion | null {
  return (
    p.transiciones.find(
      (t) => t.ejecutor === 'operador' && t.de === de && t.a === a
    ) ?? null
  );
}

/** Dueños render — mismo join `" o "` que el motor Go (RN-31). */
export function duenoRender(p: EspejoDescriptor, estado: string): string | null {
  const bindings = p.duenos[estado] ?? [];
  if (bindings.length === 0) return null;
  return bindings
    .map((b) => {
      const quien = b.arnes || b.rol;
      return b.nota ? `${quien} (${b.nota})` : quien;
    })
    .join(' o ');
}

export function categoriaDe(p: EspejoDescriptor, id: string): string | undefined {
  return p.estados.find((e) => e.id === id)?.categoria;
}
