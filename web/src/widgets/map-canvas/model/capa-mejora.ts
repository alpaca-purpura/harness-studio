import type { ResumenTelemetria } from "@/entities/telemetria"

// Las condiciones que deciden QUÉ vacío muestra cada bloque de la capa «Mejora».
//
// 🔴 Existen porque el defecto que las motivó **no era de ningún componente**: era de la
// COMPOSICIÓN. Contra el arnés `vitalia` (que nunca corrió), la franja decía «Este arnés nunca
// corrió con telemetría» y doce centímetros más abajo la lista decía «Hay datos y ningún punto
// de mejora que pase el corte. Los seis detectores corrieron sobre 0 corridas». Cada bloque era
// defendible por separado; **juntos afirmaban algo falso**, y ninguna story lo vio porque cada
// una renderiza un componente aislado.
//
// Por eso las condiciones viven acá, en UN lugar, y las usan **la página y las stories de
// coherencia**: si la regla cambia, cambia para los dos. Una condición duplicada en el
// composition-root es exactamente cómo nació el defecto.

/** Los detectores del MVP (D16.1): B4 · P1 · B2 · B6 · B3 · B1. */
export const DETECTORES_DEL_MVP = 6

/**
 * ¿Hay medición ATRIBUIBLE en la ventana? Es la condición que decide si la sección de puntos
 * de mejora se dibuja: sin esto, muestra el vacío de H-2 («hay datos y ningún punto») sobre un
 * arnés que nunca corrió.
 *
 * Dos cosas, y las dos hacen falta:
 *  1. **hubo corridas** — con 0, no hubo búsqueda que reportar;
 *  2. **al menos una se pudo atribuir** — si ninguna se atribuyó a una caja, ningún detector
 *     pudo encontrar nada, y decir «no encontraron una fuga» sería presentar la ausencia de
 *     búsqueda como resultado de una búsqueda (lo que RF-263 prohíbe).
 *
 * ⚠️ **NO se usa `resumen.confianza === "sin-dato"` para esto, aunque parezca lo obvio.** La
 * confianza del agregado es `PeorConfianza` de la ventana entera
 * (`telemetria_service.go:334`): **una sola** corrida sin atribuir la deja en `sin-dato` con 60
 * corridas perfectas al lado. Usarla como condición escondería la lista en el estado 2
 * (cobertura parcial), que **sí tiene datos** — el mismo defecto, mudado de vecino.
 */
export function hayDatosAtribuibles(resumen: ResumenTelemetria | null | undefined): boolean {
  if (!resumen || resumen.corridas <= 0) return false
  const c = resumen.cobertura
  return c.exacta + c.por_hash + c.por_proceso > 0
}

/**
 * ¿La cobertura es tan parcial que el total necesita su advertencia propia (estado 2), y no
 * alcanza con el denominador?
 *
 * **Umbral de PRODUCTO, declarado: más de un tercio de las corridas sin atribuir.** No es una
 * medición y no pretende serlo; vive acá, con nombre, para que se pueda discutir y cambiar en
 * un solo lugar. Antes era un `corridas <= 5` escondido en el JSX, que es la misma decisión
 * tomada a escondidas.
 */
export function coberturaEsParcial(resumen: ResumenTelemetria | null | undefined): boolean {
  if (!hayDatosAtribuibles(resumen) || !resumen) return false
  const c = resumen.cobertura
  const total = c.exacta + c.por_hash + c.por_proceso + c.sin_dato
  return total > 0 && c.sin_dato / total > 1 / 3
}
