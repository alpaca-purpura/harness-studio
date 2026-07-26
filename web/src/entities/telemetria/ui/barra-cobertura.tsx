import { cn } from "@/shared/lib/cn"
import { segmentosCobertura, totalCobertura } from "../model/selectors"
import type { Cobertura } from "../model/types"

// BarraCobertura — la calificación del total: de cuántas corridas salió y con qué calidad de
// atribución cada una (RF-237 · H-13).
//
// Cuatro reglas que las stories asertan, y las cuatro son de honestidad, no de estética:
//
//  1. **Una categoría en cero no se dibuja NI se nombra.** Un segmento de ancho 0 con su
//     etiqueta al lado es ruido que el ojo cuenta igual.
//  2. **Separador de 1 px de `--card` entre segmentos.** `--heat-4` y `--heat-3` miden 1,37:1
//     entre sí (D21): sin la línea no se cuentan de un vistazo. axe NO caza esto —su regla de
//     contraste solo mira texto— así que el gate pasaría con una barra ilegible.
//  3. **Por eso el portador real es el TEXTO**: el resumen de al lado y el `aria-label`
//     completo. La barra nunca es el único canal (RF-237: «distinguibles sin depender del
//     color»).
//  4. **El rótulo `cobertura` es visible SIEMPRE** (H-11), a la izquierda: al envolver en
//     pantalla angosta la barra no queda huérfana.
//
// Con 0 corridas la barra **no se dibuja**: manda el estado 1 de la franja, que dice qué pasó.

export interface BarraCoberturaProps {
  cobertura: Cobertura
  className?: string | undefined
}

export function BarraCobertura({ cobertura, className }: BarraCoberturaProps) {
  const total = totalCobertura(cobertura)
  const segmentos = segmentosCobertura(cobertura)

  // Sin corridas no hay cobertura que calificar. No es un 0 %: es que no hay denominador.
  if (total === 0) {
    return (
      <span className={cn("cov", className)}>
        <span className="cov-rotulo">cobertura</span>
        <span className="cov-txt">sin corridas que atribuir en esta ventana</span>
      </span>
    )
  }

  const todasExactas = segmentos.length === 1 && segmentos[0]?.id === "exacta"
  const resumen = todasExactas
    ? `atribución exacta en las ${total} corridas`
    : `${segmentos.map((s) => `${s.n} ${s.etiqueta}`).join(" · ")} — sobre ${total} corridas`
  const aria = `Cobertura de la atribución: ${segmentos
    .map((s) => `${s.n} ${s.aria}`)
    .join(", ")}, sobre ${total} corridas.`

  return (
    <span className={cn("cov", className)}>
      <span className="cov-rotulo">cobertura</span>
      <span className="cov-bar" role="img" aria-label={aria}>
        {segmentos.map((s) => (
          <span
            key={s.id}
            className={cn("cov-seg", `cov-${s.id}`)}
            style={{ width: `${(s.n / total) * 100}%` }}
          />
        ))}
      </span>
      <span className="cov-txt">{resumen}</span>
    </span>
  )
}
