import {
  CifraUsd,
  type FilaPortafolio,
  MarcaConfianza,
  Sparkline,
  usd,
} from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"

// Las tres celdas de la capa «Mejora» en la fila del Portafolio (design §2.7 · §3.5 · §7.6).
//
// 🔴 **Reemplazan a `TablaMejoraPortafolio`, que era CÓDIGO MUERTO** (crítico C-4): T36 entero
// vivía en una tabla que la app no montaba en ningún lado —solo aparecía en el barrel y en un
// comentario— mientras la superficie real era una SEGUNDA implementación de las mismas celdas,
// escrita a mano en `portafolio-view.tsx` y **sin el guard de D24.4**. El fix que `PARIDAD.md`
// declaraba aplicado estaba aplicado al componente que nadie ve.
//
// Ahora hay UNA sola implementación, en el widget, con stories. Y la surface es la que
// `design.md` §2.7 describe —la FILA, no una tabla aparte—, así que la tabla se borró en vez de
// montarse: montarla habría dejado la misma información dos veces en la misma pantalla.
//
// Las tres celdas consumen las MISMAS piezas que el Mapa (D23): `CifraUsd`, `MarcaConfianza` y
// `Sparkline` de `entities/telemetria`. No dibujan su propia cifra ni su propia marca de duda.

export interface CeldasMejoraFilaProps {
  fila: FilaPortafolio
}

/**
 * ¿Este arnés nunca corrió con telemetría?
 *
 * 🔴 **A-5** · NO se puede derivar de `costo_por_corrida === null`: ese campo es null fuera de S1
 * por construcción (`Corridas` es 0, y el gate del backend es `Corridas > 0`), así que la fila
 * afirmaba «nunca corrió con telemetría» sobre un arnés con 47 corridas medidas. `corridas` está
 * en el payload y es el campo que responde la pregunta.
 */
export function nuncaCorrio(fila: FilaPortafolio): boolean {
  return fila.corridas <= 0
}

export function CeldasMejoraFila({ fila }: CeldasMejoraFilaProps) {
  const sinCorridas = nuncaCorrio(fila)
  return (
    <>
      {/* Sin encabezado de columna en la lista, la celda TIENE que cargar su unidad: un `0,03`
          pelado entre chips no dice ni moneda, ni período, ni que es por corrida (A-6). */}
      <span className="pf-mej-usd">
        {sinCorridas || fila.costo_por_corrida === null ? (
          <span className="pf-mej-vacia-txt">sin dato</span>
        ) : (
          <>
            <CifraUsd micros={fila.costo_por_corrida} sinPrefijo />
            <span className="pf-mej-unidad">USD/corrida</span>
          </>
        )}
      </span>

      <span className="pf-mej-tend">
        {sinCorridas ? (
          <span
            className="pf-mej-vacia-txt"
            role="img"
            aria-label="sin tendencia: este arnés no registra corridas medidas"
          >
            —
          </span>
        ) : (
          <Sparkline puntos={fila.serie} />
        )}
      </span>

      <span className="pf-mej-punto">
        {sinCorridas ? (
          <span className="pf-mej-chip pf-mej-chip-neutro">nunca corrió con telemetría</span>
        ) : fila.punto ? (
          // D21 — la marca grave va `--crit` sobre `--card`, no sobre `--crit-soft`. El monto
          // DECLARA su unidad: un `/corrida` que en realidad es total de ventana miente por
          // omisión.
          <span className={cn("pf-mej-chip", fila.punto.grave && "grave")}>
            <span aria-hidden="true">⚠</span> {fila.punto.nombre} · USD{" "}
            {usd(fila.punto.monto_micros)}
            {fila.punto.unidad === "corrida" ? "/corrida" : " en la ventana"}
          </span>
        ) : fila.puntos_de_mejora > 0 ? (
          // 🔴 D24.4 · el backend dice que HAY hallazgos pero no llegó cuál es el principal.
          // Poner un ✓ acá miente en la dirección más cara: «no hay nada que mirar».
          <span className="pf-mej-chip">
            <span aria-hidden="true">⚠</span>{" "}
            {`${fila.puntos_de_mejora} punto(s) de mejora — abrí el Mapa para verlos`}
          </span>
        ) : (
          // ✅ El ✓ SOLO cuando el backend afirma que midió y no encontró nada: `puntos_de_mejora`
          // es 0 **como dato**, no por ausencia de `punto`.
          <span className="pf-mej-chip pf-mej-chip-ok">
            <span aria-hidden="true">✓</span> sin fugas detectadas
          </span>
        )}
        {/* La MISMA marca de duda que el canvas (D23 · H-12): la superficie real la había
            perdido por completo al no montar la tabla. */}
        <MarcaConfianza confianza={fila.confianza} />
      </span>
    </>
  )
}

/**
 * H-12 — el pie del listado. Sin él, un USD pelado entre chips se lee como facturación; la
 * superficie real lo había perdido al no montar la tabla (A-6).
 */
export function PieMejoraPortafolio() {
  return (
    <p className="pf-mej-pie">
      Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no aparecen
      en cero.
    </p>
  )
}
