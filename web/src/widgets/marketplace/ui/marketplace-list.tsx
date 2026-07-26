import {
  accionDeFila,
  ClaseChip,
  type EntradaCorruptaMarketplace,
  EstadoLecturaChip,
  esNavegable,
  type MarketplaceConocido,
  ordenarMarketplaces,
} from "@/entities/marketplace"
import { AvisoChip, EmblemaInicial } from "@/entities/portafolio"
import { cn } from "@/shared/lib/cn"

// MarketplaceList — superficie S2 del paquete 2026-07-23-portafolio-agregar-marketplace
// (AG-D8 FIRMADA 🧑‍⚖️). Props puras: CERO transporte (`fe-transporte-independiente`) — el
// `GET /api/marketplaces` y el `POST …/lecturas` los hace la página
// (`pages/shell/ui/portafolio-view.tsx`); acá solo se llaman callbacks.
//
// Qué ES este plano (y qué NO): es **el estante de lo que vendemos y el espejo de si lo que el
// cliente tiene coincide**. NO es una tienda para navegar arneses ajenos ni un gestor universal
// de plugins — la visión mata «operar arneses de terceros» (`vision.md` §Qué mutó · `CLAUDE.md`
// «solo arneses propios»). De ahí que `propio` y `referencia` NO sean simétricas.
//
// Este widget importa `entities/marketplace` Y `entities/portafolio` (átomos ya firmados:
// `EmblemaInicial`, `AvisoChip`) — puede, porque es un widget. La ENTIDAD `marketplace` no
// podría (`steiger fsd/no-cross-imports`). Y NO importa `widgets/portafolio`
// (`no-sibling-widget-imports`, severidad `error`): la página compone los dos.

export interface MarketplaceListProps {
  estado: "cargando" | "error" | "datos"
  /** motivo textual del fallo del GET — jamás una lista vacía muda (G5/BR-4). */
  error?: string | undefined
  marketplaces: MarketplaceConocido[]
  /** filas del registro propio que no parsean: VISIBLES, sin tapar el resto (BR-11). */
  corruptas: EntradaCorruptaMarketplace[]
  /** `known_marketplaces.json` ilegible: aviso VISIBLE, nunca un 500 ni una lista vacía (E-68). */
  avisoDetector?: string | undefined
  /** contador cruzado (AG-D8 decisión 7): la SEGUNDA puerta a la reconciliación. */
  sinOrigenResuelto: number
  /** inyectado para que los textos de «leído hace…» sean deterministas en story=test. */
  ahora?: Date | undefined
  /** nombres con una lectura de catálogo en vuelo (botón bloqueado, sin dos disparos). */
  leyendo?: ReadonlySet<string> | undefined
  onAbrirCatalogo: (nombre: string) => void
  onLeerCatalogo: (nombre: string) => void
  onAgregar: () => void
  onReintentar: () => void
  onVerSinOrigen: () => void
}

// ── Topbar: título + contadores REALES + Agregar (compartido con el plano Arneses) ──
function contadoresTexto(ms: MarketplaceConocido[]): string {
  const n = ms.length
  const propios = ms.filter((m) => m.clase === "propio" && esNavegable(m)).length
  const conocidos = n === 1 ? "1 conocido" : `${n} conocidos`
  const legibles = propios === 1 ? "1 propio legible" : `${propios} propios legibles`
  return `${conocidos} · ${legibles}`
}

function Topbar({
  estado,
  marketplaces,
  onAgregar,
}: {
  estado: MarketplaceListProps["estado"]
  marketplaces: MarketplaceConocido[]
  onAgregar: () => void
}) {
  return (
    <div className="pf-topbar">
      <div className="pf-topbar-titulo">
        <h2>Marketplaces</h2>
        {estado === "datos" && (
          <span className="pf-contadores">
            {contadoresTexto(marketplaces)} · el estante de lo que vendemos, y el espejo de si el
            cliente coincide
          </span>
        )}
        {estado === "cargando" && <span className="pf-contadores pf-mut">cargando…</span>}
      </div>
      <button type="button" className="pf-btn-primary" onClick={onAgregar}>
        ＋ Agregar
      </button>
    </div>
  )
}

// ── Nota de las dos clases: la asimetría se explica ANTES de entrar a un catálogo (AG-D8
// decisión 1). Literal del mockup firmado. ──
function NotaClases() {
  return (
    <p className="pf-mut pf-mk-nota">
      <b>propio</b> = publicamos ahí → catálogo navegable, «↧ Traer canónico» habilitado ·{" "}
      <b>de referencia</b> = solo resuelve la procedencia de arneses ajenos → catálogo read-only,{" "}
      <b>sin Traer</b>. Las dos clases NO son simétricas.
    </p>
  )
}

// ── Celda «pendientes»: lo que la fila puede AFIRMAR. Con lectura degradada dice por qué no
// puede afirmar nada — explícitamente «no es "no tiene arneses"» (BR-4). Con discrepancias entre
// eslabones, las muestra completas y sin ofrecer «elegir una» (BR-8).
//
// Lo que NO hay acá: los chips «2 con cambios sin publicar» / «1 desactualizado» que dibuja el
// mockup. El wire de `GET /api/marketplaces` no trae esa agregación (design.md §3.1.4 no la
// modela) y tecleársela sería exactamente la cifra fabricada que este paquete mata. Vive en el
// catálogo, fila por fila, donde el dominio SÍ la calcula. ──
function Pendientes({ m }: { m: MarketplaceConocido }) {
  const discrepancias = m.discrepancias ?? []
  if (discrepancias.length > 0) {
    return (
      <span className="mk-pendientes">
        {discrepancias.map((d) => (
          <AvisoChip key={d} aviso={d} />
        ))}
      </span>
    )
  }
  switch (m.lectura.tipo) {
    case "sin-acceso":
      return (
        <span className="mk-pendientes prosa pf-mut">
          catálogo no legible — <b>no</b> es «no tiene arneses»
        </span>
      )
    case "url-no-resuelve":
      return (
        <span className="mk-pendientes prosa pf-mut">
          el archivo se alcanzó pero no parsea — nada que afirmar todavía
        </span>
      )
    case "no-leido":
      return (
        <span className="mk-pendientes prosa pf-mut">
          registrado, sin lectura — nada que afirmar todavía
        </span>
      )
    default:
      return (
        <span className="mk-pendientes prosa pf-mut">
          eslabones: <span className="mono">{m.eslabones.join(" · ")}</span>
        </span>
      )
  }
}

// ── Fila: navegable SOLO si la lectura fue exitosa (spec §4.1.4). Una fila no legible es un
// `<div>` inerte con su motivo y su botón — jamás un link a un catálogo vacío, que se leería
// como «este marketplace no tiene arneses» (BR-4, el mismo pass fabricado que G3 mató). ──
function Fila({
  m,
  ahora,
  leyendo,
  onAbrirCatalogo,
  onLeerCatalogo,
}: {
  m: MarketplaceConocido
  ahora: Date
  leyendo: boolean
  onAbrirCatalogo: (nombre: string) => void
  onLeerCatalogo: (nombre: string) => void
}) {
  const accion = accionDeFila(m)
  const cuerpo = (
    <>
      <EmblemaInicial texto={m.nombre} size={30} />
      <span className="mk-id">
        <span className="mk-url mono">{m.repo ?? m.nombre}</span>
        <span className="mk-meta">
          <ClaseChip clase={m.clase} />
          {m.lectura.entradas !== undefined && m.lectura.cuando && (
            <>
              <span>{m.lectura.entradas} entradas de catálogo</span>
              <span aria-hidden="true">·</span>
            </>
          )}
          <EstadoLecturaChip lectura={m.lectura} ahora={ahora} />
        </span>
      </span>
      <Pendientes m={m} />
    </>
  )

  if (accion === "ver-catalogo") {
    return (
      <li>
        <button type="button" className="mk-fila" onClick={() => onAbrirCatalogo(m.nombre)}>
          {cuerpo}
          <span className="mk-alcance">
            {m.clase === "propio" ? "ver catálogo →" : "ver catálogo (read-only) →"}
          </span>
        </button>
      </li>
    )
  }

  return (
    <li>
      <div className={cn("mk-fila", "inerte")}>
        {cuerpo}
        <span className="mk-alcance">
          <button
            type="button"
            className={cn("pf-btn-mini", accion === "leer-catalogo" && "acento")}
            disabled={leyendo}
            aria-busy={leyendo}
            onClick={() => onLeerCatalogo(m.nombre)}
          >
            {leyendo ? "Leyendo…" : accion === "leer-catalogo" ? "Leer catálogo" : "Reintentar"}
          </button>
        </span>
      </div>
    </li>
  )
}

// ── Puerta cruzada (AG-D8 decisión 7): la SEGUNDA forma de llegar a la reconciliación. Solo
// conmuta al plano Arneses ya filtrado — la acción se ejecuta en la ficha del arnés, un solo
// lugar. La cifra viene del wire (`sin_origen_resuelto`), nunca se teclea. ──
function PuertaSinOrigen({ n, onIr }: { n: number; onIr: () => void }) {
  if (n <= 0) return null
  return (
    <div className="pf-puerta">
      <span aria-hidden="true">◇</span>
      <div>
        <b>{n === 1 ? "1 arnés sin origen resuelto." : `${n} arneses sin origen resuelto.`}</b> No
        pueden decir de qué marketplace vienen — sin eso, Reparar y Actualizar no tienen contra qué
        comparar.
      </div>
      <button type="button" className="pf-btn-mini acento" onClick={onIr}>
        Resolver origen
      </button>
    </div>
  )
}

// ── Degradados VISIBLES, no modales (BR-11): un registro corrupto o una metadata de CC
// ilegible nunca oculta ni impide listar el resto. ──
function BannerDegradado({
  corruptas,
  avisoDetector,
}: {
  corruptas: EntradaCorruptaMarketplace[]
  avisoDetector: string | undefined
}) {
  if (corruptas.length === 0 && !avisoDetector) return null
  return (
    <div className="pf-corruptas-banner" role="alert">
      {corruptas.length > 0 && (
        <>
          {corruptas.length} fila(s) corrupta(s) en el registro de marketplaces —{" "}
          {corruptas.map((c) => c.motivo).join("; ")}
        </>
      )}
      {corruptas.length > 0 && avisoDetector && <br />}
      {avisoDetector && <>no pude leer lo que Claude Code conoce — {avisoDetector}</>}
    </div>
  )
}

function Skeleton() {
  return (
    <div
      className="pf-skeleton"
      role="status"
      aria-live="polite"
      aria-label="Cargando marketplaces"
    >
      <span className="pf-skeleton-fila" aria-hidden="true" />
      <span className="pf-skeleton-fila" aria-hidden="true" />
    </div>
  )
}

export function MarketplaceList({
  estado,
  error,
  marketplaces,
  corruptas,
  avisoDetector,
  sinOrigenResuelto,
  ahora,
  leyendo,
  onAbrirCatalogo,
  onLeerCatalogo,
  onAgregar,
  onReintentar,
  onVerSinOrigen,
}: MarketplaceListProps) {
  const momento = ahora ?? new Date()
  const enVuelo = leyendo ?? new Set<string>()

  return (
    <div className="pf-list">
      <Topbar estado={estado} marketplaces={marketplaces} onAgregar={onAgregar} />

      {estado === "cargando" && <Skeleton />}

      {estado === "error" && (
        <div className="pf-estado-vacio">
          <p className="pf-mut" role="alert">
            No se pudo cargar los marketplaces — {error ?? "motivo desconocido"}
          </p>
          <button type="button" className="pf-btn-primary" onClick={onReintentar}>
            Reintentar
          </button>
        </div>
      )}

      {estado === "datos" && (
        <>
          <BannerDegradado corruptas={corruptas} avisoDetector={avisoDetector} />
          <PuertaSinOrigen n={sinOrigenResuelto} onIr={onVerSinOrigen} />
          <NotaClases />
          {marketplaces.length === 0 ? (
            <div className="pf-estado-vacio">
              <p>No conozco ningún marketplace todavía.</p>
              <p className="pf-mut">
                Claude Code no declara ninguno en{" "}
                <span className="mono">~/.claude/plugins/known_marketplaces.json</span> y vos no
                registraste ninguno. Usá ＋ Agregar → Marketplace.
              </p>
            </div>
          ) : (
            <ul className="mk-lista">
              {ordenarMarketplaces(marketplaces).map((m) => (
                <Fila
                  key={m.nombre}
                  m={m}
                  ahora={momento}
                  leyendo={enVuelo.has(m.nombre)}
                  onAbrirCatalogo={onAbrirCatalogo}
                  onLeerCatalogo={onLeerCatalogo}
                />
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  )
}
