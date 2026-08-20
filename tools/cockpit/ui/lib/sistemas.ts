/**
 * sistemas.ts — agrupa las sistema-keys planas de DevHub en Empresa → Sistema.
 * ────────────────────────────────────────────────────────────────────────────
 * Stage 4 (CK-07): DevHub deja de consumir el árbol rico de /api/portfolio
 * (gaps/procedencia/servicios_compartidos son negocio de Cockpit, no de
 * DevHub) — su propio switcher multi-empresa se reconstruye sobre la lista
 * PLANA que YA sirve /api/sistemas (getSelectableSistemas, devhub-owned desde
 * siempre). Las funciones puras de agrupación son copia literal del subconjunto
 * de products/cockpit/ui/lib/portfolio.ts que nunca dependió del árbol — el
 * resto (PortfolioTree y sus derivaciones) se deja atrás a propósito, no se
 * duplica: DevHub no necesita gaps/procedencia para su propio switcher.
 */

export interface SistemaEntry {
  key: string;
  sistema: string;
}

export interface EmpresaGroup {
  empresa: string;
  sistemas: SistemaEntry[];
}

export function isTwoLevel(sistemas: string[]): boolean {
  return sistemas.some((b) => b.includes('/'));
}

export function splitSistemaKey(key: string): { empresa: string; sistema: string } {
  const i = key.indexOf('/');
  if (i === -1) return { empresa: '', sistema: key };
  return { empresa: key.slice(0, i), sistema: key.slice(i + 1) };
}

export function empresaOf(key: string): string {
  return splitSistemaKey(key).empresa;
}

export function groupByEmpresa(sistemas: string[]): EmpresaGroup[] {
  const order: string[] = [];
  const byEmpresa = new Map<string, SistemaEntry[]>();
  for (const key of sistemas) {
    const { empresa, sistema } = splitSistemaKey(key);
    if (!byEmpresa.has(empresa)) {
      byEmpresa.set(empresa, []);
      order.push(empresa);
    }
    byEmpresa.get(empresa)!.push({ key, sistema });
  }
  return order.map((empresa) => ({ empresa, sistemas: byEmpresa.get(empresa)! }));
}

export function sistemasFor(sistemas: string[], empresa: string): SistemaEntry[] {
  return groupByEmpresa(sistemas).find((g) => g.empresa === empresa)?.sistemas ?? [];
}

// ── Deep-link por URL (I-50) — mismo contrato de query string, validado
// contra la lista plana en vez del árbol. ────────────────────────────────────

export function parseSistemaSearch(search: string): { empresa?: string; sistema?: string } {
  const p = new URLSearchParams(search);
  return {
    empresa: p.get('empresa') || undefined,
    sistema: p.get('sistema') || undefined,
  };
}

export function buildSistemaSearch(empresa: string, sistemaSlug: string): string {
  const p = new URLSearchParams();
  if (empresa) p.set('empresa', empresa);
  if (sistemaSlug) p.set('sistema', sistemaSlug);
  const s = p.toString();
  return s ? `?${s}` : '';
}

export interface SistemaSelection {
  empresa: string;
  sistemaSlug: string;
  key?: string;
}

/** Resuelve una intención (?empresa=&sistema=) contra la lista plana de sistemas
 *  activos. Defensivo: empresa sin ningún sistema → null (el caller conserva su
 *  estado actual); slug que no pertenece a la empresa → el primero de esa empresa. */
export function resolveSistemaIntent(
  sistemas: string[],
  intent: { empresa?: string; sistema?: string },
): SistemaSelection | null {
  if (!intent.empresa) return null;
  const group = sistemasFor(sistemas, intent.empresa);
  if (group.length === 0) return null;
  const match = intent.sistema ? group.find((s) => s.sistema === intent.sistema) : undefined;
  const entry = match ?? group[0];
  return { empresa: intent.empresa, sistemaSlug: entry.sistema, key: entry.key };
}
