'use client';

// Tab Arquitectura RETIRADO (2026-06-11): fusionado dentro del Mapa del producto
// (drill-down progresivo — pedido del operador: una sola vista, no dos tabs que
// leen el mismo SYSTEM-MAP). Esta ruta queda como redirect para links viejos.
//
// Dónde quedó cada pieza:
//   · Jerarquía + áreas planificadas → el árbol del mapa (badge planned + barra de progreso)
//   · Flujos cross-agent + data ownership → colapsables al pie del mapa + drawer por caja
//   · Editar SYSTEM-MAP.yaml → header del mapa
//   · Diagrama Mermaid → eliminado (redundante con el árbol renderizado)

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function ArquitecturaPage() {
  const router = useRouter();
  useEffect(() => {
    router.replace('/map');
  }, [router]);
  return (
    <div className="p-6 text-xs text-[var(--color-muted)]">
      El tab Arquitectura se fusionó en el Mapa del producto… redirigiendo.
    </div>
  );
}
