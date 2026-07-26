import {
  BarraCobertura,
  type Cobertura,
  type Escenario,
  type ResumenTelemetria,
  usd,
  type Ventana,
} from "@/entities/telemetria"
import { cn } from "@/shared/lib/cn"
import { ErrorBody, Skeleton } from "@/shared/ui/estado-carga"
import { coberturaEsParcial } from "../model/capa-mejora"

// FranjaMejora — la franja de contexto de la capa (design §2.2, §5.2, §7.2, §7.7).
//
// Es CHROME del Mapa, no canvas: vive entre `MapBar` y el lienzo y el canvas no sabe que
// existe (mismo criterio que `MapBar`). Vive en la slice `map-canvas` y no en una slice propia
// para no chocar con `no-sibling-widget-imports`, que está en `error`.
//
// **El orden de izquierda a derecha ES el orden de lectura del dato**, no una preferencia
// estética:
//   ventana → total + denominador → disclaimer PEGADO AL TOTAL → chip de reenvío (condicional)
//   → cobertura → «qué guardamos».
// Si el disclaimer se va al pie, se lee DESPUÉS de haber creído el número.
//
// La regla que gobierna los siete estados: **ningún número aparece sin su ventana, su
// denominador, su disclaimer y su cobertura**, y **nunca hay un tablero en cero**. Un `USD 0,00`
// donde no hubo medición afirma que el arnés corrió gratis.

/** Los tres cortes de la ventana. El wire recibe `desde`/`hasta`; la traducción vive en la página. */
const VENTANAS: readonly { id: Ventana; label: string }[] = [
  { id: "7d", label: "7 días" },
  { id: "30d", label: "30 días" },
  { id: "todo", label: "todo" },
]

const DISCLAIMER = "estimado por el runtime, no es facturación"

/** Estado de transporte. Separado del estado del DATO: «no pude preguntar» ≠ «no hay nada». */
export type EstadoFranja = "datos" | "cargando" | "error" | "daemon-caido"

export interface FranjaMejoraProps {
  estado?: EstadoFranja | undefined
  ventana: Ventana
  onVentana: (v: Ventana) => void
  /** `null`/`undefined` = no hay resumen que mostrar; manda el estado 1 (o 1b). */
  resumen?: ResumenTelemetria | null | undefined
  /**
   * Estado 1b (H-9): hubo corridas, pero **no en esta ventana**. Lleva la fecha de la última.
   * «Nunca corrió» y «no corrió en estos 7 días» son cosas distintas y se dicen distinto: la
   * primera pide instrumentar, la segunda pide ampliar la ventana.
   */
  ultimaCorridaFuera?: string | undefined
  /** El escenario del arnés en la ventana. S2 se parte en dos desde el ANEXO H9. */
  escenario?: Escenario | undefined
  /** Estado 4 — el runtime no reporta costo y el número salió del catálogo. */
  costoDelCatalogo?: boolean | undefined
  /** Estado 5 — el catálogo es el del release y nunca se refrescó. */
  catalogoSinRefrescar?: string | undefined
  /** El runtime del arnés no está soportado todavía. Se dice; no se muestra un cero. */
  runtimeNoSoportado?: string | undefined
  /** Chip de reenvío externo. Un estado peligroso no se esconde a la derecha (D13 · H-7). */
  forwardDestino?: string | undefined
  /** El TTL vigente. ⚠️ El `90` es PROPUESTO, no firmado (J-6): sale de la config, no de acá. */
  retencionDias?: number | undefined
  retencionPropuesta?: boolean | undefined
  onPolitica: () => void
  onReintentar: () => void
  /** El motivo REAL del fallo de consulta. Jamás «Error al obtener los datos». */
  error?: string | undefined
  /** Con el daemon caído se conservan las cifras previas, marcadas como posiblemente viejas. */
  fechaUltimaMedicion?: string | undefined
}

export function FranjaMejora({
  estado = "datos",
  ventana,
  onVentana,
  resumen,
  ultimaCorridaFuera,
  escenario,
  costoDelCatalogo,
  catalogoSinRefrescar,
  runtimeNoSoportado,
  forwardDestino,
  retencionDias,
  retencionPropuesta,
  onPolitica,
  onReintentar,
  error,
  fechaUltimaMedicion,
}: FranjaMejoraProps) {
  const cob: Cobertura | undefined = resumen?.cobertura
  const conAtribucion = cob ? cob.exacta + cob.por_hash + cob.por_proceso : 0
  const total = resumen
    ? usd(resumen.costo_reportado_micros ?? resumen.costo_calculado_micros)
    : null
  // El umbral del estado 2 vive en `../model/capa-mejora`, con nombre y testeado: antes era un
  // `corridas <= 5` escondido acá, que es la misma decisión de producto tomada a escondidas.
  const parcial = coberturaEsParcial(resumen)

  return (
    <div className="arnesia-mejora fm">
      {/* La ventana va PRIMERO: es el modificador de todo lo demás. Si va después, el total se
          lee antes de saber sobre qué período habla. Sigue usable durante la carga. */}
      <span className="fm-ventana">
        <label htmlFor="fm-ventana-sel">ventana</label>
        <select
          id="fm-ventana-sel"
          value={ventana}
          onChange={(e) => onVentana(e.target.value as Ventana)}
        >
          {VENTANAS.map((v) => (
            <option key={v.id} value={v.id}>
              {v.label}
            </option>
          ))}
        </select>
      </span>

      {estado === "error" ? (
        // La franja NO desaparece: desaparecer se leería como «no hay capa» (H-6).
        <span className="fm-cuerpo">
          <ErrorBody
            motivo={`No se pudo leer la telemetría — ${error ?? "motivo desconocido"}.`}
            onReintentar={onReintentar}
          />
        </span>
      ) : estado === "cargando" ? (
        <span className="fm-cuerpo">
          <Skeleton label="midiendo…" filas={1} altura={22} data="franja" />
          <span className="fm-mut">midiendo…</span>
        </span>
      ) : runtimeNoSoportado !== undefined ? (
        <span className="fm-cuerpo">
          <span className="fm-vacio">
            {`Este arnés corre con ${runtimeNoSoportado} y todavía no medimos ese runtime.`}
          </span>
          <span className="fm-mut">
            Cuando lo midamos, las corridas viejas no se recuperan: la medición arranca desde
            entonces.
          </span>
        </span>
      ) : resumen == null ? (
        // ── Estados 1 y 1b — son DOS cosas distintas (H-9) ──────────────────────────────
        <span className="fm-cuerpo">
          <span className="fm-total fm-total-vacio">— —</span>
          {ultimaCorridaFuera === undefined ? (
            <span className="fm-vacio">
              <span className="fm-vacio-titulo">Este arnés nunca corrió con telemetría.</span>
              <span className="fm-mut">
                Abrí una sesión desde ArnesIA y la medición arranca sola.
              </span>
            </span>
          ) : (
            <span className="fm-vacio">
              <span className="fm-vacio-titulo">
                {`Sin corridas en los últimos ${ventana === "30d" ? "30" : "7"} días. La última fue el ${ultimaCorridaFuera}.`}
              </span>
              <button type="button" className="fm-btn" onClick={() => onVentana("todo")}>
                Ver todo
              </button>
            </span>
          )}
        </span>
      ) : (
        <>
          <span className="fm-cuerpo">
            <span className="fm-total">
              <span className="fm-usd">USD</span>
              <span className="fm-monto">{total ?? "sin dato"}</span>
              {/* Estado 2 — cobertura parcial. El denominador dice de cuántas salió el número.
                  D21: el resaltado del contado es `--foreground` en negrita, NO `--warn` sobre
                  `--card` (3,76:1, la deuda que el BACKLOG ya tiene abierta y que este paquete
                  no puede agravar). */}
              {parcial && (
                <span className="fm-parcial">
                  {" — de "}
                  {resumen.corridas} corridas, <b className="fm-parcial-n">{conAtribucion}</b>
                </span>
              )}
            </span>
            {parcial ? (
              <span className="fm-mut">{`${cob?.sin_dato} corridas quedaron sin atribución.`}</span>
            ) : (
              <span className="fm-denominador">
                {`de ${resumen.corridas} corridas, ${conAtribucion} con atribución · ${resumen.sesiones} sesiones · ${resumen.cajas} cajas`}
              </span>
            )}
          </span>

          {/* D21 · el disclaimer va PEGADO al total, y su texto es `--foreground` sobre
              `--warn-soft`: `--warn` sobre `--warn-soft` da 3,24:1 en tema claro y rompe el
              gate. `--warn` queda solo para el borde, donde el mínimo es 3:1. */}
          <span className="fm-disclaimer">{DISCLAIMER}</span>

          {escenario === "s2-degradado" && (
            <span className="fm-escenario">
              <span className="fm-escenario-rotulo">corrió fuera de ArnesIA, sin instrumentar</span>
              <span className="fm-mut">
                re-warm por TTL — no disponible: sin el result del stream-json.
              </span>
            </span>
          )}
          {escenario === "s2-instrumentado" && (
            <span className="fm-escenario">
              <span className="fm-escenario-rotulo">corrió fuera de ArnesIA, instrumentado</span>
              <span className="fm-mut">Llega la misma señal que dentro de ArnesIA.</span>
            </span>
          )}

          {costoDelCatalogo && (
            <span className="fm-mut">
              {`calculado con el catálogo v${resumen.catalogo.version} — este runtime no reporta costo`}
            </span>
          )}

          {catalogoSinRefrescar !== undefined && (
            <span className="fm-catalogo">
              <span className="fm-catalogo-titulo">
                {`⚠ Precios del release, sin refrescar desde el ${catalogoSinRefrescar}.`}
              </span>
              <span className="fm-mut">
                Funciona sin internet; te avisamos que está funcionando así.
              </span>
            </span>
          )}

          {estado === "daemon-caido" && (
            <span className="fm-daemon">
              {`El daemon no respondió. Lo que ves es la última medición, del ${fechaUltimaMedicion ?? "—"}.`}
            </span>
          )}

          {/* Un estado peligroso no se esconde a la derecha (D13 · H-7). */}
          {forwardDestino !== undefined && (
            <span className="fm-forward">{`reenvío externo encendido → ${forwardDestino}`}</span>
          )}

          {cob && <BarraCobertura cobertura={cob} className="fm-cov" />}
        </>
      )}

      {/* «Qué guardamos»: última, discreta, SIEMPRE presente — también mientras carga o falla.
          La promesa de privacidad no depende de que el daemon conteste. */}
      <span className="fm-privacidad">
        <span className="fm-priv-linea">Nada de tu cuenta. Nada de la conversación.</span>
        <span className="fm-priv-linea fm-mut">
          {retencionDias !== undefined && (
            <>
              {`Retención ${retencionDias} días`}
              {retencionPropuesta && <span className="fm-propuesto"> (propuesto)</span>}
              {" · "}
            </>
          )}
          <button type="button" className={cn("fm-link")} onClick={onPolitica}>
            qué guardamos
          </button>
        </span>
      </span>
    </div>
  )
}
