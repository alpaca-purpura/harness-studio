import { cn } from "@/shared/lib/cn"
import { ariaTendencia, direccionTendencia } from "../model/selectors"

// Sparkline — la tendencia de las últimas corridas, con su texto equivalente (RF-266 · RF-279).
//
// Las dos reglas duras:
//
//  1. **Con menos de dos puntos no se dibuja nada** y se dice `pocas corridas para una
//     tendencia`. Una línea plana de un punto diría «estable», que es una afirmación que nadie
//     midió — y afirmar «estable» sin serie es exactamente el pass fabricado.
//  2. **El `role="img"` + `aria-label` llevan la dirección en palabras.** Cinco barras de 5 px
//     no son un canal accesible; el texto sí.
//
// ⚠️ HUECO REAL DEL BACKEND, declarado y no tapado: `domain.FilaPortafolio` **no tiene ningún
// campo de serie** (auditoría §UX ítem 6 · M10 — `Tendencia` existe en el contrato y nunca se
// calcula). Hasta que exista, el Portafolio pasa `puntos` vacío y esta pieza muestra el copy
// de ausencia. No se dibuja una línea inventada para llenar la columna.

export interface SparklineProps {
  /** La serie, del más viejo al más nuevo. Menos de 2 puntos ⇒ se dice, no se dibuja. */
  puntos?: readonly number[] | undefined
  className?: string | undefined
}

export function Sparkline({ puntos, className }: SparklineProps) {
  const dir = direccionTendencia(puntos)
  if (dir === null || !puntos) {
    return <span className={cn("spark-vacio", className)}>pocas corridas para una tendencia</span>
  }
  const max = Math.max(...puntos, 1)
  const aria = ariaTendencia(dir, puntos.length)
  return (
    <span className={cn("spark", `spark-${dir}`, className)} role="img" aria-label={aria}>
      {puntos.map((p, i) => (
        <span
          // biome-ignore lint/suspicious/noArrayIndexKey: la posición en la serie ES la identidad del punto.
          key={i}
          className={cn("spark-bar", i === puntos.length - 1 && "spark-bar-ultima")}
          data-ultima={i === puntos.length - 1 ? "true" : undefined}
          style={{ height: `${Math.max(2, Math.round((p / max) * 20))}px` }}
        />
      ))}
    </span>
  )
}
