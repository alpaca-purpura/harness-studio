import type { Box, Graph } from "@/entities/arnes"
import { isCaja } from "@/entities/arnes"
import type { CifraCaja, ResumenTelemetria } from "@/entities/telemetria"
import { motivoSinDato } from "@/entities/telemetria"
import type { MejoraNodo } from "../ui/lane"
import { propsDeMejora } from "./props-de-mejora"

// La derivación ENTERA de la capa «Mejora», en funciones puras y unit-testeadas.
//
// 🔴 Existe porque los cinco defectos de composición del paquete (D24 + los cuatro críticos de
// `auditoria-tramo-b.md`) nacieron todos del mismo sitio: **decisiones derivadas tomadas en el
// composition-root**, que es el único lugar del repo sin stories. Acá adentro no hay JSX, así que
// se puede probar con una tabla; y `CapaMejoraStage` es su único consumidor, así que la página no
// puede volver a tomarlas por su cuenta.

/** Los detectores del MVP (D16.1): B4 · P1 · B2 · B6 · B3 · B1. */
export const DETECTORES_DEL_MVP = 6

/**
 * ¿Hubo medición atribuible en la ventana?
 *
 * ⚠️⚠️ **NO se puede usar `resumen.corridas` como proxy de esto, y ése fue el crítico C-2.**
 * `Corridas` es `COUNT(DISTINCT corrida_id)` (`consultas.go:95`) y **`corrida_id` es NULL fuera
 * de S1** — el propio store usa `corrida_id IS NOT NULL` para *derivar* que algo es S1. O sea que
 * en `s2-instrumentado`, que `PARIDAD.md` §4 declara **el escenario mayoritario**, `corridas`
 * vale 0 con el costo, la cobertura y las cajas perfectamente pobladas.
 *
 * Mirando ese campo, la franja decía «Este arnés nunca corrió con telemetría» mientras el carril
 * de al lado cobraba `USD 1,08` y el nodo marcaba `⚠ re-warm TTL`; y como el candado de D24
 * también miraba `corridas`, **escondía la lista de puntos de mejora justo donde el backend sí
 * los había calculado**. El entregable central del paquete, invisible en el caso normal.
 *
 * La condición es **la cobertura atribuida**, y nada más: al menos un turno pudo colgarse de
 * una caja. Es lo que el nombre dice y es lo único que hace verdadera la frase «los detectores
 * corrieron y no encontraron nada» — sin atribución no hay a qué colgar un hallazgo.
 *
 * ⚠️ Probé agregarle un fallback por `turnos > 0` / `dinero > 0` «para cubrir S2» y **rompió el
 * invariante de D24**: con 4 turnos todos sin atribuir, la lista volvía a afirmar que se buscó.
 * No hacía falta — en `s2-instrumentado` la cobertura viene poblada (`exacta: 58`), así que la
 * primera condición ya alcanza. El defecto C-2 era **solo** el `corridas <= 0`.
 *
 * ⚠️ Tampoco sirve `resumen.confianza === "sin-dato"` (la otra formulación natural): la confianza
 * del agregado es `PeorConfianza` de la ventana (`telemetria_service.go:334`), así que **una
 * sola** corrida sin atribuir la deja en `sin-dato` con 60 perfectas al lado — escondería el
 * estado 2, que sí tiene datos.
 */
export function hayDatosAtribuibles(resumen: ResumenTelemetria | null | undefined): boolean {
  if (!resumen) return false
  const c = resumen.cobertura
  return c.exacta + c.por_hash + c.por_proceso > 0
}

/**
 * El denominador honesto de «los detectores corrieron sobre N …».
 *
 * Fuera de S1 `corridas` es 0 por construcción, y «corrieron sobre 0 corridas» es exactamente la
 * frase que el operador cazó en la app instalada. Cuando el contador de corridas no sirve, se
 * usan los turnos, que es lo que el backend sí contó.
 */
export function denominadorDeBusqueda(resumen: ResumenTelemetria | null | undefined): number {
  if (!resumen) return 0
  return resumen.corridas > 0 ? resumen.corridas : resumen.turnos
}

/**
 * Cómo se llama lo que cuenta el denominador. Fuera de S1 no son corridas —el contador es 0 por
 * construcción— sino turnos, y llamarlas «corridas» sería nombrar mal la unidad justo en la
 * frase que existe para que el número no se lea solo.
 */
export function unidadDelDenominador(
  resumen: ResumenTelemetria | null | undefined,
): "corridas" | "turnos" {
  return resumen && resumen.corridas > 0 ? "corridas" : "turnos"
}

/**
 * ¿La cobertura es tan parcial que el total necesita su advertencia propia (estado 2), y no
 * alcanza con el denominador?
 *
 * **Umbral de PRODUCTO, declarado: más de un tercio de los turnos sin atribuir.** No es una
 * medición y no pretende serlo; vive acá, con nombre, para que se pueda discutir y cambiar en un
 * solo lugar. Antes era un `corridas <= 5` escondido en el JSX, que es la misma decisión tomada
 * a escondidas.
 */
export function coberturaEsParcial(resumen: ResumenTelemetria | null | undefined): boolean {
  if (!hayDatosAtribuibles(resumen) || !resumen) return false
  const c = resumen.cobertura
  const total = c.exacta + c.por_hash + c.por_proceso + c.sin_dato
  return total > 0 && c.sin_dato / total > 1 / 3
}

/** Todo lo que los bloques de la capa necesitan, derivado de una vez y de una sola fuente. */
export interface VistaCapaMejora {
  hayDatos: boolean
  /** El resumen que ve la franja: `null` ⇒ estado 1. **Se deriva de `hayDatosAtribuibles`**, no
   *  de `corridas`, así que no puede contradecir a la lista ni al canvas. */
  resumenParaFranja: ResumenTelemetria | null
  denominadorDeBusqueda: number
  mejoraPorNodo: ReadonlyMap<string, CifraCaja> | undefined
  motivosPorNodo: ReadonlyMap<string, string> | undefined
  totalesPorFase: ReadonlyMap<string, number | null> | undefined
  /** Las props primitivas ya compuestas: se pueden assertar sin montar el canvas. */
  propsPorNodo: ReadonlyMap<string, MejoraNodo> | undefined
}

export interface EntradaVistaCapaMejora {
  resumen: ResumenTelemetria | null
  cajas: readonly CifraCaja[]
  graph: Graph | null
  /** `false` con la capa apagada: todo vuelve `undefined` y el DOM es el de hoy. */
  activa: boolean
}

export function vistaCapaMejora({
  resumen,
  cajas,
  graph,
  activa,
}: EntradaVistaCapaMejora): VistaCapaMejora {
  const hayDatos = hayDatosAtribuibles(resumen)
  const base: VistaCapaMejora = {
    hayDatos,
    resumenParaFranja: hayDatos ? resumen : null,
    denominadorDeBusqueda: denominadorDeBusqueda(resumen),
    mejoraPorNodo: undefined,
    motivosPorNodo: undefined,
    totalesPorFase: undefined,
    propsPorNodo: undefined,
  }
  if (!activa || !graph) return base

  // 🔴 C-2 · si el estado 1 manda, **el canvas tampoco puede seguir cobrando dinero**. Sin esto
  // la franja diría «nunca corrió» y el nodo de al lado `USD 1,08`, que es el defecto entero.
  if (!hayDatos) return base

  const mejoraPorNodo = new Map(cajas.map((c) => [c.caja_id, c]))

  // El motivo de «sin dato atribuible» para todo lo que NO es caja. Sale de `MOTIVO_SIN_DATO`,
  // única fuente: `entities/arnes` no puede importar la otra entity (D18), así que el nodo lo
  // recibe por prop y este widget —el único que ve las dos— lo resuelve.
  const nodos = (graph.nodos ?? []) as Box[]
  const motivosPorNodo = new Map<string, string>()
  for (const n of nodos) {
    if (isCaja(n) || mejoraPorNodo.has(n.id)) continue
    motivosPorNodo.set(n.id, motivoSinDato(n.clase))
  }

  // Total por fase. `null` cuando ninguna caja de esa fase tiene costo atribuido: el carril dice
  // «sin dato», jamás «USD 0,00» (RF-244).
  const totalesPorFase = new Map<string, number | null>()
  for (const n of nodos) {
    if (!n.fase) continue
    const c = mejoraPorNodo.get(n.id)
    const prev = totalesPorFase.get(n.fase) ?? null
    if (c?.costo_micros != null) totalesPorFase.set(n.fase, (prev ?? 0) + c.costo_micros)
    else if (!totalesPorFase.has(n.fase)) totalesPorFase.set(n.fase, null)
  }

  const propsPorNodo = new Map<string, MejoraNodo>()
  for (const n of nodos) {
    const cifra = mejoraPorNodo.get(n.id)
    const motivo = motivosPorNodo.get(n.id)
    if (cifra) propsPorNodo.set(n.id, propsDeMejora(cifra, motivo))
    else if (motivo !== undefined) propsPorNodo.set(n.id, { motivoSinDato: motivo })
  }

  return { ...base, mejoraPorNodo, motivosPorNodo, totalesPorFase, propsPorNodo }
}
