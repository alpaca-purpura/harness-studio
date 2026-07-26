import { cn } from "@/shared/lib/cn"
import { etiquetaConfianza, tituloConfianza } from "../model/selectors"
import type { Confianza } from "../model/types"

// MarcaConfianza — cómo se atribuyó un número, dicho (RF-242). Cuatro casos, cuatro
// tratamientos, y **ninguno se pinta como «atribución aproximada»**: esa palabra taparía tres
// cosas distintas bajo una sola, que es justo lo que el boundary
// `cifra-viaja-con-su-confianza` prohíbe.
//
//   exacta       → NO RENDERIZA NADA. *La ausencia de marca ES la señal* (design §5.3).
//   por-huella   → el runtime redactó el nombre; se identificó por el hash del plugin.
//   por proceso  → se dedujo del directorio, y si ahí corre más de un arnés los MEZCLA.
//   sin dato     → no hay número que marcar; suma a la cobertura, no al total.
//
// Los tres visibles comparten tono a propósito (D21: texto `--foreground` sobre `--warn-soft`):
// **el texto es el portador de la diferencia**, no el color (RF-280). Los `title` son tres
// strings distintos y `selectors.ts` es su única fuente (D18 candado 2).

export interface MarcaConfianzaProps {
  confianza: Confianza
  className?: string | undefined
}

export function MarcaConfianza({ confianza, className }: MarcaConfianzaProps) {
  const etiqueta = etiquetaConfianza(confianza)
  const titulo = tituloConfianza(confianza)
  // exacta: cero DOM. No es un caso olvidado — es el diseño.
  if (etiqueta === null || titulo === null) return null
  return (
    <span
      className={cn("conf", `conf-${confianza}`, className)}
      data-confianza={confianza}
      title={titulo}
    >
      {etiqueta}
    </span>
  )
}
