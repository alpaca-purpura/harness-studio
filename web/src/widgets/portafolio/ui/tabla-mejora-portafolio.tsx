import {
  CifraUsd,
  type FilaPortafolio,
  MarcaConfianza,
  Sparkline,
  usd,
} from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"

// TablaMejoraPortafolio — las tres columnas nuevas del Portafolio (design §2.7, §3.5, §5.6, §7.6).
//
// Es la superficie que permite decir **«en este puesto»**, el eje diferencial del producto
// (D9.8): ccusage, tokscale, Dynatrace y Azure agregan por herramienta, modelo, proyecto y día;
// ninguno por unidad de trabajo.
//
// **D20 — el puesto no se fabrica.** `EntradaPortafolio` no tiene ningún campo `puesto`; el dato
// sale del `rol` del arnés indexado, resuelto SERVER-SIDE, y viaja como `puesto *string`. Si el
// arnés no declara `rol`, viaja `null` y acá se lee **`puesto sin declarar`** — que **es el caso
// normal de hoy**, no un borde: ningún arnés del dogfood declara `rol` (auditoría §UX).
// La fila se agrupa por `(identidad, instalación)`: dos instalaciones del mismo arnés dan dos
// filas, y si las dos resuelven al mismo `rol`, se distinguen por instalación.
//
// **Al ordenar por costo, los «sin dato» se agrupan AL FINAL, con separador rotulado.** Nunca
// intercalados como si valieran 0: un arnés que nunca corrió no es el más barato.
//
// Esta tabla consume **las MISMAS piezas** que el Mapa (D23): `CifraUsd`, `MarcaConfianza` y
// `Sparkline` de `entities/telemetria`. No dibuja su propia cifra ni su propia marca de duda.

export interface TablaMejoraPortafolioProps {
  filas: readonly FilaPortafolio[]
  /** Ordena por costo/corrida descendente y agrupa los «sin dato» al final. */
  ordenarPorCosto?: boolean | undefined
  /** H-3 — abre el Mapa de ese arnés con la capa Mejora activa y la caja seleccionada. */
  onAbrirEnMapa?: ((clave: string, instalacionId: string) => void) | undefined
}

export function TablaMejoraPortafolio({
  filas,
  ordenarPorCosto,
  onAbrirEnMapa,
}: TablaMejoraPortafolioProps) {
  const conDato = filas.filter((f) => f.costo_por_corrida !== null)
  const sinDato = filas.filter((f) => f.costo_por_corrida === null)
  const ordenadas = ordenarPorCosto
    ? [...conDato].sort((a, b) => (b.costo_por_corrida ?? 0) - (a.costo_por_corrida ?? 0))
    : conDato
  const mostrarSeparador = ordenarPorCosto === true && sinDato.length > 0 && conDato.length > 0

  return (
    <div className="arnesia-portafolio pf-mej">
      {/* El scroll vive acá, no en el `<body>`. `tabindex`+`role`+`aria-label` porque axe exige
          que todo contenedor scrolleable sea alcanzable por teclado (design §6.3). */}
      <div
        className="pf-mej-scroll"
        // biome-ignore lint/a11y/noNoninteractiveTabindex: axe exige que todo contenedor scrolleable sea alcanzable por teclado (design §6.3) — sin esto la tabla es inalcanzable para quien no usa mouse.
        tabIndex={0}
        role="region"
        aria-label="Costo y puntos de mejora por arnés"
      >
        <table className="pf-mej-tbl">
          <thead>
            <tr>
              <th scope="col">arnés · puesto</th>
              {/* El `USD` vive en el ENCABEZADO: la celda muestra `0,31` sin prefijo, y así la
                  columna alinea sin repetir tres letras en cada fila (design §3.5). */}
              <th scope="col" className="pf-mej-num">
                USD/corrida
              </th>
              <th scope="col">tendencia</th>
              <th scope="col">punto de mejora</th>
            </tr>
          </thead>
          <tbody>
            {/* Sin orden por costo, la tabla respeta el orden que le dieron: reordenar por
                default escondería el criterio del llamador. */}
            {(ordenarPorCosto ? ordenadas : filas).map((f) => (
              <Fila key={`${f.clave}~${f.instalacion_id}`} fila={f} onAbrirEnMapa={onAbrirEnMapa} />
            ))}
            {mostrarSeparador && (
              <tr className="pf-mej-sep">
                <th scope="colgroup" colSpan={4}>
                  Sin datos de telemetría
                </th>
              </tr>
            )}
            {ordenarPorCosto &&
              sinDato.map((f) => (
                <Fila
                  key={`${f.clave}~${f.instalacion_id}`}
                  fila={f}
                  onAbrirEnMapa={onAbrirEnMapa}
                />
              ))}
          </tbody>
        </table>
      </div>
      {/* H-12 — sin este pie, un USD pelado en una tabla se lee como facturación. */}
      <p className="pf-mej-pie">
        Costo estimado por el runtime, no es facturación. Los arneses sin datos lo dicen: no
        aparecen en cero.
      </p>
    </div>
  )
}

function Fila({
  fila,
  onAbrirEnMapa,
}: {
  fila: FilaPortafolio
  onAbrirEnMapa?: ((clave: string, instalacionId: string) => void) | undefined
}) {
  const sinDato = fila.costo_por_corrida === null
  return (
    <tr className="pf-mej-fila" data-instalacion={fila.instalacion_id}>
      <th scope="row" className="pf-mej-id">
        {onAbrirEnMapa ? (
          <button
            type="button"
            className="pf-mej-link"
            onClick={() => onAbrirEnMapa(fila.clave, fila.instalacion_id)}
          >
            <span className="mono">{fila.arnes_id}</span>
          </button>
        ) : (
          <span className="mono">{fila.arnes_id}</span>
        )}
        {/* D20 — `null` ⇒ se DICE. No se infiere del path ni del nombre del proyecto. */}
        <span className="pf-mej-puesto">{fila.puesto ?? "puesto sin declarar"}</span>
      </th>

      <td className="pf-mej-num pf-mej-usd">
        <CifraUsd micros={fila.costo_por_corrida} sinPrefijo />
      </td>

      <td className="pf-mej-tend">
        {sinDato ? (
          // Celda vacía MARCADA como tal: un guion sin rótulo no dice si falta el dato o si la
          // tendencia es plana.
          <span
            className="pf-mej-vacia"
            role="img"
            aria-label="sin tendencia: nunca corrió con telemetría"
          >
            —
          </span>
        ) : (
          <Sparkline puntos={fila.serie} />
        )}
      </td>

      <td className="pf-mej-punto">
        {sinDato ? (
          <span className="pf-mej-chip pf-mej-chip-neutro">nunca corrió con telemetría</span>
        ) : fila.punto ? (
          // D21 — el chip de fuga grave usa texto `--crit` sobre `--card`, no sobre
          // `--crit-soft`: acá el gate axe SÍ corre en `error`, y es donde el fix es
          // load-bearing. El monto DECLARA su unidad (design §7.6): un `/corrida` que en
          // realidad es total de ventana miente por omisión.
          <span className={cn("pf-mej-chip", fila.punto.grave && "grave")}>
            <span aria-hidden="true">⚠</span> {fila.punto.nombre} · USD{" "}
            {usd(fila.punto.monto_micros)}
            {fila.punto.unidad === "corrida" ? "/corrida" : " en la ventana"}
          </span>
        ) : fila.puntos_de_mejora > 0 ? (
          // El backend dice que HAY hallazgos pero no llegó cuál es el principal. Decirlo es
          // feo; poner un ✓ sería mentir en la dirección más cara: «acá no hay nada que mirar».
          <span className="pf-mej-chip">
            <span aria-hidden="true">⚠</span>{" "}
            {`${fila.puntos_de_mejora} punto(s) de mejora — abrí el Mapa para verlos`}
          </span>
        ) : (
          // ✅ El ✓ SOLO cuando el backend afirma que midió y no encontró nada: `puntos_de_mejora`
          // es 0 **como dato**, no por ausencia de `punto` (que puede faltar por otra razón).
          // Un ✓ sobre una búsqueda que no ocurrió es la misma mentira que cazó la verificación
          // en la app instalada, en la superficie de al lado.
          <span className="pf-mej-chip pf-mej-chip-ok">
            <span aria-hidden="true">✓</span> sin fugas detectadas
          </span>
        )}
        {/* La MISMA marca de duda que el canvas — la pieza es la misma (D23). */}
        <MarcaConfianza confianza={fila.confianza} />
      </td>
    </tr>
  )
}
