import { useState } from "react"
import {
  AvisoChip,
  agruparPorEmpresa,
  agruparPorMarketplace,
  agruparPorProyecto,
  DerivaChip,
  DotSaludPortafolio,
  EmblemaInicial,
  type EntradaCorrupta,
  type EntradaPortafolio,
  filtrarEntradas,
  filtrarPorMarketplace,
  filtrarPorSalud,
  filtrarSinOrigen,
  identificadorDe,
  type LentePortafolio,
  marketplacesDisponibles,
  registriesDe,
  SALUD_LABEL,
  type SaludPortafolio,
  saludDe,
} from "@/entities/portafolio"
import { cn } from "@/shared/lib/cn"
import { FiltroDisclosure } from "@/shared/ui/filtro-disclosure"
import { GrupoControl } from "@/shared/ui/grupo-control"

const SALUDES: readonly SaludPortafolio[] = ["ok", "atencion", "sin-senal"]

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
  // Filtros «Estado»/«Marketplace» de la toolbar (deuda viva S1-D8, cerrada 2026-07-24):
  // afordancia DISTINTA de la lente — acotan la lista, nunca la reagrupan. Set vacío = sin
  // filtro (mismo convenio que `busqueda`).
  filtroSalud: ReadonlySet<SaludPortafolio>
  onFiltroSalud: (s: ReadonlySet<SaludPortafolio>) => void
  filtroMarketplace: ReadonlySet<string>
  onFiltroMarketplace: (s: ReadonlySet<string>) => void
  /** Filtro «sin origen resuelto» (AG-D8 decisión 7, paquete 2026-07-23): lo prende el contador
   *  cruzado del plano Marketplaces. Es un FILTRO, no una lente: acota, no reagrupa. Opcional
   *  (BR-12) ⇒ las stories firmadas del Slice 1 no cambian. */
  filtroSinOrigen?: boolean | undefined
  onFiltroSinOrigen?: ((v: boolean) => void) | undefined
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

// ── Toolbar: buscar + 4 lentes empresa/plano/proyecto/marketplace (S1-D8, cerrada 2026-07-23)
// + filtros Estado/Marketplace (S1-D8, cerrada 2026-07-24) — afordancia DISTINTA de la lente:
// acotan la lista visible sin reagruparla.
//
// AG-D4 (paquete 2026-07-23) cierra el defecto **L1** de la auditoría: los rótulos de grupo pasan
// a ser TEXTO VISIBLE (`VER POR […]` / `FILTROS […]`, mayúsculas por CSS) vía `GrupoControl`.
// Antes eran 6 pills iguales en fila donde la 4ª es una LENTE y la 6ª un FILTRO — un lector de
// pantalla las distinguía por el `aria-label` del grupo, un ojo no. Ninguna de las dos
// «Marketplace» se renombra: con el rubro a la vista se leen sin ambigüedad, y renombrar rompería
// vocabulario ya firmado en la PARIDAD del Slice 1. Los `role="group"` + `aria-label` existentes
// se CONSERVAN (no se degrada a11y, se agrega afordancia). ──
function Toolbar({
  lente,
  onLente,
  busqueda,
  onBusqueda,
  filtroSalud,
  onFiltroSalud,
  filtroMarketplace,
  onFiltroMarketplace,
  marketplaces,
}: {
  lente: LentePortafolio
  onLente: (l: LentePortafolio) => void
  busqueda: string
  onBusqueda: (q: string) => void
  filtroSalud: ReadonlySet<SaludPortafolio>
  onFiltroSalud: (s: ReadonlySet<SaludPortafolio>) => void
  filtroMarketplace: ReadonlySet<string>
  onFiltroMarketplace: (s: ReadonlySet<string>) => void
  marketplaces: string[]
}) {
  const [abierto, setAbierto] = useState<"estado" | "marketplace" | null>(null)
  const toggleAbierto = (panel: "estado" | "marketplace") =>
    setAbierto((a) => (a === panel ? null : panel))

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
      <GrupoControl
        rotulo="Ver por"
        ariaLabel="Lente del Portafolio"
        hint="lente para encontrar un arnés cuando hay muchos"
        claseControles="pf-lentes"
      >
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
        <button
          type="button"
          className="pf-lente-btn"
          aria-pressed={lente === "proyecto"}
          onClick={() => onLente("proyecto")}
        >
          Proyecto
        </button>
        <button
          type="button"
          className="pf-lente-btn"
          aria-pressed={lente === "marketplace"}
          onClick={() => onLente("marketplace")}
        >
          Marketplace
        </button>
      </GrupoControl>
      <GrupoControl rotulo="Filtros" ariaLabel="Filtros del Portafolio" claseControles="pf-filtros">
        <FiltroDisclosure
          etiqueta="Estado"
          abierto={abierto === "estado"}
          onToggleAbierto={() => toggleAbierto("estado")}
          panelId="pf-filtro-panel-estado"
          valores={SALUDES}
          labelDe={(s) => SALUD_LABEL[s]}
          seleccion={filtroSalud}
          onCambiar={onFiltroSalud}
        />
        <FiltroDisclosure
          etiqueta="Marketplace"
          abierto={abierto === "marketplace"}
          onToggleAbierto={() => toggleAbierto("marketplace")}
          panelId="pf-filtro-panel-marketplace"
          valores={marketplaces}
          labelDe={(m) => m}
          seleccion={filtroMarketplace}
          onCambiar={onFiltroMarketplace}
          vacio="sin marketplaces en los datos actuales"
        />
      </GrupoControl>
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
      <p>Ningún arnés coincide con la búsqueda/filtros aplicados.</p>
      <button type="button" className="pf-btn-secundario" onClick={onLimpiar}>
        Limpiar búsqueda y filtros
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

// agrupadorDe — despacha la lente al selector puro correspondiente (mismo patrón N:M para las
// 3: empresa/proyecto/marketplace agrupan por facetas de la entrada, nunca la parten).
function agrupadorDe(
  lente: LentePortafolio,
): (es: EntradaPortafolio[]) => { grupo: string; entradas: EntradaPortafolio[] }[] {
  switch (lente) {
    case "proyecto":
      return agruparPorProyecto
    case "marketplace":
      return agruparPorMarketplace
    default:
      return agruparPorEmpresa
  }
}

function ListaAgrupada({
  entradas,
  lente,
  seleccionada,
  onAbrir,
}: {
  entradas: EntradaPortafolio[]
  lente: LentePortafolio
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
}) {
  const grupos = agrupadorDe(lente)(entradas)
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
  filtroSalud,
  filtroMarketplace,
  filtroSinOrigen,
  seleccionada,
  onAbrir,
  onReintentar,
  onLimpiarTodo,
}: {
  estado: PortafolioListProps["estado"]
  error: string | undefined
  entradas: EntradaPortafolio[]
  lente: LentePortafolio
  busqueda: string
  filtroSalud: ReadonlySet<SaludPortafolio>
  filtroMarketplace: ReadonlySet<string>
  filtroSinOrigen: boolean
  seleccionada: string | undefined
  onAbrir: (clave: string) => void
  onReintentar: () => void
  onLimpiarTodo: () => void
}) {
  if (estado === "cargando") return <Skeleton />
  if (estado === "error") return <ErrorBody error={error} onReintentar={onReintentar} />

  // estado === "datos"
  if (entradas.length === 0) return <VaciaBody />

  // Pipeline (S1-D8, + «sin origen» del paquete 2026-07-23): buscar → filtro Estado → filtro
  // Marketplace → filtro «sin origen resuelto» → recién ahí la lente reagrupa lo que sobrevivió.
  // Los filtros acotan; la lente solo cambia la presentación.
  const acotadas = filtrarPorMarketplace(
    filtrarPorSalud(filtrarEntradas(entradas, busqueda), filtroSalud),
    filtroMarketplace,
  )
  const filtradas = filtroSinOrigen ? filtrarSinOrigen(acotadas) : acotadas
  if (filtradas.length === 0) return <SinResultadosBody onLimpiar={onLimpiarTodo} />

  if (lente === "plano") {
    return <ListaPlana entradas={filtradas} seleccionada={seleccionada} onAbrir={onAbrir} />
  }
  return (
    <ListaAgrupada
      entradas={filtradas}
      lente={lente}
      seleccionada={seleccionada}
      onAbrir={onAbrir}
    />
  )
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
  filtroSalud,
  onFiltroSalud,
  filtroMarketplace,
  onFiltroMarketplace,
  filtroSinOrigen,
  onFiltroSinOrigen,
  seleccionada,
  onAbrir,
  onAgregar,
  onReintentar,
}: PortafolioListProps) {
  const marketplaces = marketplacesDisponibles(entradas)
  const sinOrigenActivo = filtroSinOrigen ?? false
  const onLimpiarTodo = () => {
    onBusqueda("")
    onFiltroSalud(new Set())
    onFiltroMarketplace(new Set())
    onFiltroSinOrigen?.(false)
  }
  return (
    <div className="pf-list">
      <Topbar estado={estado} entradas={entradas} lente={lente} onAgregar={onAgregar} />
      <Toolbar
        lente={lente}
        onLente={onLente}
        busqueda={busqueda}
        onBusqueda={onBusqueda}
        filtroSalud={filtroSalud}
        onFiltroSalud={onFiltroSalud}
        filtroMarketplace={filtroMarketplace}
        onFiltroMarketplace={onFiltroMarketplace}
        marketplaces={marketplaces}
      />
      <BannerCorruptas corruptas={corruptas} />
      {/* AG-D8 decisión 7 — el filtro que prende el contador cruzado del plano Marketplaces se
          DECLARA: un filtro invisible que acota la lista es indistinguible de «no tengo arneses».
          Salir de él es un click, y no se pierde ninguna otra selección. */}
      {sinOrigenActivo && (
        <div className="pf-puerta">
          <span aria-hidden="true">◇</span>
          <div>
            Mostrando <b>solo los arneses sin origen resuelto</b> — los que no pueden decir de qué
            marketplace vienen. Abrí uno y usá «Resolver origen» en su ficha.
          </div>
          <button type="button" className="pf-btn-mini" onClick={() => onFiltroSinOrigen?.(false)}>
            Ver todos
          </button>
        </div>
      )}
      <div className="pf-body">
        <Cuerpo
          estado={estado}
          error={error}
          entradas={entradas}
          lente={lente}
          busqueda={busqueda}
          filtroSalud={filtroSalud}
          filtroMarketplace={filtroMarketplace}
          filtroSinOrigen={sinOrigenActivo}
          seleccionada={seleccionada}
          onAbrir={onAbrir}
          onReintentar={onReintentar}
          onLimpiarTodo={onLimpiarTodo}
        />
      </div>
    </div>
  )
}
