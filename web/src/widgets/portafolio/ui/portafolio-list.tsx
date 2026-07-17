import {
  AvisoChip,
  agruparPorEmpresa,
  DerivaChip,
  DotSaludPortafolio,
  EmblemaInicial,
  type EntradaCorrupta,
  type EntradaPortafolio,
  filtrarEntradas,
  identificadorDe,
  type LentePortafolio,
  registriesDe,
  saludDe,
} from "@/entities/portafolio"
import { cn } from "@/shared/lib/cn"

// PortafolioList — superficie 1 del Portafolio (plan §2.6/§3 T4, G1/G2/G4/G5/G6/G8/G9).
// Props puras: CERO transporte (fe-transporte-independiente) — el fetch/refetch/AbortController
// viven en pages/shell/ui/portafolio-view.tsx (T7). Este widget solo agrupa/filtra con los
// selectores puros de entities/portafolio (agruparPorEmpresa/filtrarEntradas) — presentación,
// no dominio nuevo.

export interface PortafolioListProps {
  estado: "cargando" | "error" | "datos"
  error?: string | undefined
  entradas: EntradaPortafolio[]
  corruptas: EntradaCorrupta[]
  lente: LentePortafolio
  onLente: (l: LentePortafolio) => void
  busqueda: string
  onBusqueda: (q: string) => void
  /** clave de la fila abierta en el drawer (T5) — resalta la fila, no cambia su comportamiento. */
  seleccionada?: string | undefined
  onAbrir: (clave: string) => void
  onAgregar: () => void
  onReintentar: () => void
}

// ── Topbar: título + contadores REALES + Agregar (SIEMPRE presente — plan §2.6) ──
function contadoresTexto(entradas: EntradaPortafolio[], lente: LentePortafolio): string {
  const empresas = new Set<string>()
  for (const e of entradas) {
    for (const emp of e.empresas ?? []) empresas.add(emp)
  }
  return `${entradas.length} arneses · ${empresas.size} empresas · lente: ${lente}`
}

function Topbar({
  estado,
  entradas,
  lente,
  onAgregar,
}: {
  estado: PortafolioListProps["estado"]
  entradas: EntradaPortafolio[]
  lente: LentePortafolio
  onAgregar: () => void
}) {
  return (
    <div className="pf-topbar">
      <div className="pf-topbar-titulo">
        <h2>Portafolio</h2>
        {estado === "datos" && (
          <span className="pf-contadores">{contadoresTexto(entradas, lente)}</span>
        )}
        {estado === "cargando" && <span className="pf-contadores pf-mut">cargando…</span>}
      </div>
      <button type="button" className="pf-btn-primary" onClick={onAgregar}>
        ＋ Agregar
      </button>
    </div>
  )
}

// ── Toolbar: buscar + lentes empresa/plano (S1-D8) + lentes/filtros diferidos disabled ──
function Toolbar({
  lente,
  onLente,
  busqueda,
  onBusqueda,
}: {
  lente: LentePortafolio
  onLente: (l: LentePortafolio) => void
  busqueda: string
  onBusqueda: (q: string) => void
}) {
  return (
    <div className="pf-toolbar">
      <input
        type="search"
        className="pf-buscar"
        aria-label="Buscar arnés"
        placeholder="Buscar por id, nombre o descripción…"
        value={busqueda}
        onChange={(e) => onBusqueda(e.target.value)}
      />
      <div className="pf-lentes" role="group" aria-label="Lente del Portafolio">
        <button
          type="button"
          className="pf-lente-btn"
          aria-pressed={lente === "empresa"}
          onClick={() => onLente("empresa")}
        >
          Empresa
        </button>
        <button
          type="button"
          className="pf-lente-btn"
          aria-pressed={lente === "plano"}
          onClick={() => onLente("plano")}
        >
          Plano
        </button>
        <button type="button" className="pf-lente-btn" disabled title="próximo">
          Proyecto
        </button>
        <button type="button" className="pf-lente-btn" disabled title="próximo">
          Marketplace
        </button>
      </div>
      <div className="pf-filtros" role="group" aria-label="Filtros del Portafolio">
        <button type="button" className="pf-filtro-btn" disabled title="próximo">
          Estado
        </button>
        <button type="button" className="pf-filtro-btn" disabled title="próximo">
          Marketplace
        </button>
      </div>
    </div>
  )
}

// ── Banner de corruptas (BR-11): visible, no modal — el resto de la lista sigue viva. ──
function BannerCorruptas({ corruptas }: { corruptas: EntradaCorrupta[] }) {
  if (corruptas.length === 0) return null
  const motivos = corruptas.map((c) => c.motivo).join("; ")
  return (
    <div className="pf-corruptas-banner" role="alert">
      {corruptas.length} entrada(s) corrupta(s) en el registro — {motivos}
    </div>
  )
}

// ── Fila: emblema · id+nombre/descr · presencia · chips honestos · dot de salud (plan §2.6) ──
function Fila({
  entrada,
  seleccionada,
  onAbrir,
}: {
  entrada: EntradaPortafolio
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
}) {
  const salud = saludDe(entrada)
  const registries = registriesDe(entrada)
  const instalaciones = entrada.instalaciones ?? []
  const instalacionEnDeriva = instalaciones.find((i) => i.deriva === "en-deriva")
  const instalacionConAviso = instalaciones.find((i) => i.aviso)
  const idMostrado = identificadorDe(entrada.identidad)

  return (
    <li className="pf-fila-wrap">
      <button
        type="button"
        className={cn("pf-fila", seleccionada === entrada.clave && "seleccionada")}
        aria-current={seleccionada === entrada.clave ? "true" : undefined}
        onClick={() => onAbrir(entrada.clave)}
      >
        <EmblemaInicial texto={entrada.identidad.id} />
        <span className="pf-fila-id">
          <span className="pf-fila-idline">
            <span className="mono">{idMostrado}</span>
            {entrada.nombre && <span className="pf-fila-nombre">{entrada.nombre}</span>}
          </span>
          {entrada.descripcion && <span className="pf-fila-descr">{entrada.descripcion}</span>}
        </span>
        <span className="pf-fila-presencia">
          {entrada.canonico ? (
            <span>◆ canónico v{entrada.canonico.version ?? "?"}</span>
          ) : (
            <span>◇ sin canónico</span>
          )}
          <span>▣ {instalaciones.length} instalac.</span>
        </span>
        <span className="pf-fila-chips">
          {instalacionEnDeriva && (
            <DerivaChip estado="en-deriva" detalle={instalacionEnDeriva.deriva_detalle} />
          )}
          {registries.length === 0 && (
            <span className="pf-chip pf-chip-origen-desconocido">origen?</span>
          )}
          {instalacionConAviso?.aviso && <AvisoChip aviso={instalacionConAviso.aviso} />}
        </span>
        <DotSaludPortafolio salud={salud} />
      </button>
    </li>
  )
}

// ── Cuerpo: los estados honestos (G5) — cargando/error/vacío/sin-resultados/agrupado/plano ──
function Skeleton() {
  return (
    <div className="pf-skeleton" role="status" aria-live="polite" aria-label="Cargando portafolio">
      <span className="pf-skeleton-fila" aria-hidden="true" />
      <span className="pf-skeleton-fila" aria-hidden="true" />
      <span className="pf-skeleton-fila" aria-hidden="true" />
    </div>
  )
}

function ErrorBody({
  error,
  onReintentar,
}: {
  error: string | undefined
  onReintentar: () => void
}) {
  return (
    <div className="pf-estado-vacio">
      <p className="pf-mut" role="alert">
        No se pudo cargar el portafolio — {error ?? "motivo desconocido"}
      </p>
      <button type="button" className="pf-btn-primary" onClick={onReintentar}>
        Reintentar
      </button>
    </div>
  )
}

function VaciaBody() {
  return (
    <div className="pf-estado-vacio">
      <p>Tu portafolio está vacío.</p>
      <p className="pf-mut">Usá ＋ Agregar (arriba) para escanear un proyecto local.</p>
    </div>
  )
}

function SinResultadosBody({ onLimpiar }: { onLimpiar: () => void }) {
  return (
    <div className="pf-estado-vacio">
      <p>Ningún arnés coincide con la búsqueda.</p>
      <button type="button" className="pf-btn-secundario" onClick={onLimpiar}>
        Limpiar búsqueda
      </button>
    </div>
  )
}

function ListaPlana({
  entradas,
  seleccionada,
  onAbrir,
}: {
  entradas: EntradaPortafolio[]
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
}) {
  return (
    <ul className="pf-lista">
      {entradas.map((e) => (
        <Fila key={e.clave} entrada={e} seleccionada={seleccionada} onAbrir={onAbrir} />
      ))}
    </ul>
  )
}

function ListaAgrupada({
  entradas,
  seleccionada,
  onAbrir,
}: {
  entradas: EntradaPortafolio[]
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
}) {
  const grupos = agruparPorEmpresa(entradas)
  return (
    <>
      {grupos.map(({ grupo, entradas: entradasGrupo }) => (
        <section key={grupo} className="pf-grupo">
          <h3 className="pf-grupo-titulo">{grupo}</h3>
          <ul className="pf-lista">
            {entradasGrupo.map((e) => (
              <Fila
                key={`${grupo}-${e.clave}`}
                entrada={e}
                seleccionada={seleccionada}
                onAbrir={onAbrir}
              />
            ))}
          </ul>
        </section>
      ))}
    </>
  )
}

function Cuerpo({
  estado,
  error,
  entradas,
  lente,
  busqueda,
  seleccionada,
  onAbrir,
  onReintentar,
  onBusqueda,
}: {
  estado: PortafolioListProps["estado"]
  error: string | undefined
  entradas: EntradaPortafolio[]
  lente: LentePortafolio
  busqueda: string
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
  onReintentar: () => void
  onBusqueda: (q: string) => void
}) {
  if (estado === "cargando") return <Skeleton />
  if (estado === "error") return <ErrorBody error={error} onReintentar={onReintentar} />

  // estado === "datos"
  if (entradas.length === 0) return <VaciaBody />

  const filtradas = filtrarEntradas(entradas, busqueda)
  if (filtradas.length === 0) return <SinResultadosBody onLimpiar={() => onBusqueda("")} />

  if (lente === "plano") {
    return <ListaPlana entradas={filtradas} seleccionada={seleccionada} onAbrir={onAbrir} />
  }
  return <ListaAgrupada entradas={filtradas} seleccionada={seleccionada} onAbrir={onAbrir} />
}

export function PortafolioList({
  estado,
  error,
  entradas,
  corruptas,
  lente,
  onLente,
  busqueda,
  onBusqueda,
  seleccionada,
  onAbrir,
  onAgregar,
  onReintentar,
}: PortafolioListProps) {
  return (
    <div className="pf-list">
      <Topbar estado={estado} entradas={entradas} lente={lente} onAgregar={onAgregar} />
      <Toolbar lente={lente} onLente={onLente} busqueda={busqueda} onBusqueda={onBusqueda} />
      <BannerCorruptas corruptas={corruptas} />
      <div className="pf-body">
        <Cuerpo
          estado={estado}
          error={error}
          entradas={entradas}
          lente={lente}
          busqueda={busqueda}
          seleccionada={seleccionada}
          onAbrir={onAbrir}
          onReintentar={onReintentar}
          onBusqueda={onBusqueda}
        />
      </div>
    </div>
  )
}
