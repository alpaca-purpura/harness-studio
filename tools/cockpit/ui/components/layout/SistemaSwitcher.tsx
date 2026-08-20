'use client';

import { Select } from '@/components/ui/Select';
import { useSistema } from '@/components/providers/SistemaProvider';
import { PLATFORM_SLUG, PLATFORM_LABEL } from '@/lib/platform-context';
import { groupByEmpresa, sistemasFor } from '@/lib/sistemas';

/** Label visible por contexto: platform se distingue de las sistemas reales. */
function sistemaOptionLabel(b: string): string {
  return b === PLATFORM_SLUG ? `⬡ ${PLATFORM_LABEL}` : b;
}

export function SistemaSwitcher() {
  const { sistema, setSistema, sistemas, mode, empresa, setEmpresa, sistemaSlug, setSistemaSlug } = useSistema();

  const isMulti = mode === 'multi';

  if (!isMulti && sistemas.length <= 1) {
    return (
      <div className="text-[11px] text-[var(--color-muted)]">
        Sistema: <span className="text-[var(--color-text)] font-medium">{sistema}</span>
      </div>
    );
  }

  // ── Modo MULTI (multi-empresa): selector de 2 niveles Empresa → Sistema ──
  // Fuente única = la lista plana de /api/sistemas, agrupada por empresa
  // (Stage 4 · CK-07 — ya no depende del árbol rico de /api/portfolio).
  if (isMulti) {
    const empresas = groupByEmpresa(sistemas);
    const sistemasDeEmpresa = sistemasFor(sistemas, empresa);

    return (
      <div className="flex items-center gap-2">
        <label className="text-[11px] text-[var(--color-muted)]" htmlFor="empresa-switcher">
          Empresa
        </label>
        <Select
          id="empresa-switcher"
          value={empresa}
          onChange={(e) => setEmpresa(e.target.value)}
          className="!w-auto !py-1"
        >
          {empresas.map((emp) => (
            <option key={emp.empresa} value={emp.empresa}>
              {emp.empresa}
            </option>
          ))}
        </Select>
        <label className="text-[11px] text-[var(--color-muted)]" htmlFor="sistema-switcher">
          Sistema
        </label>
        <Select
          id="sistema-switcher"
          value={sistemaSlug}
          onChange={(e) => {
            if (e.target.value) setSistemaSlug(e.target.value);
          }}
          className="!w-auto !py-1"
        >
          {sistemasDeEmpresa.length === 0 && (
            <option value="" disabled>
              — sin sistema —
            </option>
          )}
          {sistemasDeEmpresa.map((s) => (
            <option key={s.sistema} value={s.sistema}>
              {sistemaOptionLabel(s.sistema)}
            </option>
          ))}
        </Select>
      </div>
    );
  }

  // ── Modo SINGLE: un workspace, selector plano de sistemas ──
  return (
    <div className="flex items-center gap-2">
      <label className="text-[11px] text-[var(--color-muted)]" htmlFor="sistema-switcher">
        Sistema
      </label>
      <Select
        id="sistema-switcher"
        value={sistema}
        onChange={(e) => setSistema(e.target.value)}
        className="!w-auto !py-1"
      >
        {sistemas.map((b) => (
          <option key={b} value={b}>
            {sistemaOptionLabel(b)}
          </option>
        ))}
      </Select>
    </div>
  );
}
