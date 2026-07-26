import {
  type BucketToken,
  CifraUsd,
  type EstadoDetector,
  entero,
  type ParidadCosto,
  usd,
} from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"
import { ErrorBody, Skeleton } from "@/shared/ui/estado-carga"

// InspectorMejora — el cuerpo de la 4ª tab (design §2.6, §3.4, §5.5, §7.5).
//
// Existe para que el número se pueda **auditar**: bucket por bucket, contra el catálogo, contra
// el proceso, y contra los detectores que no corrieron. El orden de las secciones va **de lo
// auditable a lo accionable**: buckets → paridad → join → detectores.
//
// Tres reglas que se ven en cada sección:
//
//  1. **«No aplica» ocupa las dos columnas numéricas con `colspan=2`, en itálica** — no un guion
//     en cada celda, que se leería como cero. Y **`0` es un dato válido** cuando el runtime tiene
//     el concepto: el fixture usa el cero REAL medido (`ephemeral_5m = 0`).
//  2. **La paridad muestra los dos números y NO elige.** El veredicto es TEXTO (`✓ coinciden` /
//     `⚠ difieren en USD 0,18`), no un tono: si difieren, o el catálogo está viejo o el runtime
//     cambió su tarifa, y ninguna de las dos es «el número correcto».
//  3. **Detector `off` = `--border`, nunca `--ok`**: «no sé» ≠ «sano». Es la misma regla que ya
//     enforça `.pf-dot-salud.sin-senal` en el Portafolio.
//
// La ventana es **la de la capa** (J-3): esta tab la HEREDA, no la elige. Las «14 corridas» son
// el DENOMINADOR, no la ventana.

export interface JoinDeLaCaja {
  corridas: number
  /** `null` = no hubo señal de gate. **No es 0**: 0 diría «ninguna se rechazó». */
  rechazadas: number | null
  costo_rechazadas_micros: number | null
  /** `null` = la rotación es de ArnesIA y estas corridas fueron afuera. */
  rotaciones: number | null
  motivo_sin_gate?: string | undefined
  motivo_sin_rotacion?: string | undefined
}

export interface InspectorMejoraProps {
  estado?: "datos" | "cargando" | "error" | undefined
  /** Un nodo que no es caja tiene UNA sección con el motivo, y cero tablas vacías. */
  esCaja: boolean
  /** El MISMO motivo que muestra el canvas para ese nodo (design §5.5). */
  motivoNoCaja?: string | undefined
  /** «7 días» | «30 días» | «todo» — heredada de la capa (J-3). */
  ventanaLabel: string
  /** El denominador. NO es la ventana. */
  corridas: number
  buckets: readonly BucketToken[]
  totalMicros: number | null
  paridad: ParidadCosto
  join: JoinDeLaCaja
  detectores: readonly EstadoDetector[]
  /** Los que están FUERA del MVP. Dicen `no medido todavía`, jamás `sin hallazgos`. */
  noMedidos: readonly EstadoDetector[]
  onReintentar: () => void
  error?: string | undefined
}

function Sec({ titulo, children }: { titulo: string; children: React.ReactNode }) {
  return (
    <section className="sec mej-sec">
      <h4>{titulo}</h4>
      {children}
    </section>
  )
}

export function InspectorMejora({
  estado = "datos",
  esCaja,
  motivoNoCaja,
  ventanaLabel,
  corridas,
  buckets,
  totalMicros,
  paridad,
  join,
  detectores,
  noMedidos,
  onReintentar,
  error,
}: InspectorMejoraProps) {
  // RF-258 — un nodo que no es caja NO recibe tablas vacías: recibe el motivo, el mismo que el
  // canvas ya le mostró. Cuatro secciones en blanco serían cuatro invitaciones a buscar un dato
  // que no existe.
  if (!esCaja) {
    return (
      <div className="mej-inspector">
        <Sec titulo="Mejora">
          <p className="mej-nocaja">Esta capa mide cajas. {motivoNoCaja}</p>
        </Sec>
      </div>
    )
  }

  if (estado === "error") {
    return (
      <div className="mej-inspector">
        <Sec titulo="Mejora">
          <ErrorBody
            motivo={`No se pudo leer el detalle de la caja — ${error ?? "motivo desconocido"}.`}
            onReintentar={onReintentar}
          />
        </Sec>
      </div>
    )
  }

  // Cargando: los TÍTULOS ya visibles. Un skeleton sin títulos no dice qué está por llegar.
  if (estado === "cargando") {
    return (
      <div className="mej-inspector">
        {[
          "Tokens",
          "Costo — reportado vs. calculado",
          "El join — corridas de esta caja",
          "Detectores",
        ].map((t) => (
          <Sec key={t} titulo={t}>
            <Skeleton label={`Cargando ${t}`} filas={2} altura={18} data="inspector" />
          </Sec>
        ))}
      </div>
    )
  }

  const divergencia =
    paridad.reportado_micros !== null && paridad.calculado_micros !== null
      ? Math.abs(paridad.reportado_micros - paridad.calculado_micros)
      : null

  return (
    <div className="mej-inspector">
      {/* J-3 — la ventana es la de la capa; las corridas son el denominador. */}
      <Sec titulo={`Tokens · ${ventanaLabel} (${corridas} corridas)`}>
        {/* biome-ignore lint/a11y/noNoninteractiveTabindex: axe EXIGE que todo contenedor
            scrolleable sea alcanzable por teclado (design §6.3). Sin `tabindex="0"` la tabla de
            buckets es inaccesible para quien no usa mouse — el gate a11y del propio archivo lo
            cazaría. Mismo patrón que el tooltip operable por teclado de `inspector.tsx`. */}
        <div className="mej-scroll" tabIndex={0} role="region" aria-label="Tokens por bucket">
          <table className="tbl mej-tbl">
            <thead>
              <tr>
                <th scope="col">bucket</th>
                <th scope="col">tokens</th>
                <th scope="col">USD</th>
              </tr>
            </thead>
            <tbody>
              {buckets.map((b) => (
                <tr key={b.id} data-bucket={b.id}>
                  <th scope="row">{b.etiqueta}</th>
                  {b.tokens === null ? (
                    // «No aplica» ocupa las DOS columnas numéricas. Un guion en cada celda se
                    // leería como cero, y un cero donde el concepto no existe sería mentira.
                    <td colSpan={2} className="mej-noaplica">
                      no aplica en este runtime
                    </td>
                  ) : (
                    <>
                      <td className="num-celda">{entero(b.tokens)}</td>
                      <td className="num-celda">
                        <CifraUsd micros={b.costo_micros} sinPrefijo />
                      </td>
                    </>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <p className="mej-nota-tbl">
          «No aplica» no es 0. Un cero donde el concepto no existe sería mentira.
        </p>
        {totalMicros !== null && (
          <p className="mej-total-tbl">
            total <CifraUsd micros={totalMicros} />
          </p>
        )}
      </Sec>

      <Sec titulo="Costo — reportado vs. calculado">
        <div className={cn("mej-paridad", divergencia === 0 ? "coinciden" : "difieren")}>
          <span>
            runtime{" "}
            {paridad.reportado_micros === null ? (
              <em>
                este runtime no reporta costo — calculado con el catálogo v
                {paridad.catalogo_version}
              </em>
            ) : (
              <CifraUsd micros={paridad.reportado_micros} />
            )}
          </span>
          <span>
            nuestro catálogo{" "}
            {paridad.catalogo_sin_construir === true ? (
              <em>catálogo sin construir</em>
            ) : (
              <CifraUsd micros={paridad.calculado_micros} />
            )}
          </span>
          {/* RF-261 · RF-280 — el veredicto es TEXTO. El ✓/⚠ es decorativo y redundante con la
              palabra: en escala de grises el veredicto sigue leyéndose. */}
          {divergencia !== null && (
            <span className="mej-veredicto">
              {divergencia === 0 ? (
                <>
                  <span aria-hidden="true">✓</span> coinciden
                </>
              ) : (
                <>
                  <span aria-hidden="true">⚠</span> difieren en USD {usd(divergencia)}
                </>
              )}
            </span>
          )}
        </div>
        {divergencia !== null && divergencia !== 0 && (
          <p className="mej-nota-tbl">
            Guardamos los dos. Si difieren, o nuestro catálogo está viejo o el runtime cambió su
            tarifa.
          </p>
        )}
      </Sec>

      <Sec titulo="El join — corridas de esta caja">
        <table className="tbl mej-tbl">
          <tbody>
            <tr>
              <th scope="row">corridas</th>
              <td className="num-celda">{join.corridas}</td>
            </tr>
            <tr data-fila="rechazadas">
              <th scope="row">rechazadas en el gate</th>
              {join.rechazadas === null ? (
                <td className="mej-noaplica">
                  {join.motivo_sin_gate ?? "sin señal de gate en estas corridas"}
                </td>
              ) : (
                <td className="num-celda">{join.rechazadas}</td>
              )}
            </tr>
            <tr>
              <th scope="row">costo de las rechazadas</th>
              {join.costo_rechazadas_micros === null ? (
                <td className="mej-noaplica">
                  {join.motivo_sin_gate ?? "sin señal de gate en estas corridas"}
                </td>
              ) : (
                <td className="num-celda">
                  <CifraUsd micros={join.costo_rechazadas_micros} />
                </td>
              )}
            </tr>
            <tr>
              <th scope="row">rotaciones de contexto</th>
              {join.rotaciones === null ? (
                <td className="mej-noaplica">
                  {join.motivo_sin_rotacion ??
                    "sin señal — la rotación es de ArnesIA y estas corridas fueron afuera"}
                </td>
              ) : (
                <td className="num-celda">{join.rotaciones}</td>
              )}
            </tr>
          </tbody>
        </table>
      </Sec>

      <Sec titulo="Detectores">
        <ul className="mej-detectores">
          {detectores.map((d) => (
            <li key={d.detector}>
              <span className="det-linea">
                {/* El punto es refuerzo y está `aria-hidden`: el estado viaja en el TEXTO. */}
                <span
                  className={cn(
                    "det-dot",
                    !d.aplica ? "det-off" : d.hallazgos > 0 ? "det-on" : "det-clean",
                  )}
                  aria-hidden="true"
                />
                {d.aplica
                  ? `${d.nombre} · ${d.hallazgos > 0 ? "activo" : "sin hallazgos"}`
                  : d.nombre}
              </span>
              {!d.aplica && d.motivo !== undefined && (
                <span className="det-motivo">{`no disponible: ${d.motivo}`}</span>
              )}
              {d.sin_fix === true && <span className="det-motivo">sin fix propuesto</span>}
            </li>
          ))}
        </ul>

        {/* RF-263 — los de fuera del MVP dicen `no medido todavía`, **jamás `sin hallazgos`**:
            «no lo miramos» y «lo miramos y está limpio» son conclusiones opuestas. */}
        {noMedidos.length > 0 && (
          <div className="mej-nomedidos">
            <p className="mej-nota-tbl">Fuera del MVP</p>
            <ul className="mej-detectores">
              {noMedidos.map((d) => (
                <li key={d.detector}>
                  <span className="det-linea">
                    <span className="det-dot det-off" aria-hidden="true" />
                    {`${d.nombre} · no medido todavía`}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </Sec>
    </div>
  )
}
