'use client';

/**
 * ProcesoProvider — el descriptor de proceso como contexto (F6 · RN-50).
 *
 * UN fetch a `/api/proceso` al montar el shell; las vistas derivan sus defs con los
 * helpers puros de `lib/proceso.ts` (el board ya no hardcodea el ciclo — RN-51).
 * `null` mientras carga o si falla: los consumidores degradan (spinner o fallback
 * de presentación), nunca crashean.
 */

import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { getProceso } from '@/lib/api-client';
import type { ProcesoResponse } from '@/lib/types';

const ProcesoContext = createContext<ProcesoResponse | null>(null);

export function ProcesoProvider({ children }: { children: ReactNode }) {
  const [proceso, setProceso] = useState<ProcesoResponse | null>(null);

  useEffect(() => {
    let cancelled = false;
    getProceso()
      .then((p) => {
        if (!cancelled) setProceso(p);
      })
      .catch(() => {
        /* sin descriptor → los consumidores degradan; el binario sano siempre lo sirve */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return <ProcesoContext.Provider value={proceso}>{children}</ProcesoContext.Provider>;
}

/** El descriptor vivo, o null mientras carga (degradar, no crashear). */
export function useProceso(): ProcesoResponse | null {
  return useContext(ProcesoContext);
}
