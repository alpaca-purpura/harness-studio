import { open as elegirCarpeta } from "@tauri-apps/plugin-dialog"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import type {
  CandidatoOrigen,
  CandidatosOrigen,
  Catalogo,
  ClaseMarketplace,
  EntradaCorruptaMarketplace,
  EstadoTraer,
  ListadoMarketplaces,
  MarketplaceConocido,
  ResultadoTraer,
  TipoSituacion,
  Validacion,
} from "@/entities/marketplace"
import type {
  Candidato,
  EntradaCorrupta,
  EntradaPortafolio,
  LentePortafolio,
  PortafolioListado,
  SaludPortafolio,
} from "@/entities/portafolio"
import type { FilaPortafolio } from "@/entities/telemetria"
import { ApiError, api, isTauri, selectActive, useAppStore, useSessions } from "@/shared"
import { MarketplaceCatalogo, MarketplaceList } from "@/widgets/marketplace"
import {
  PortafolioDrawer,
  PortafolioList,
  PortafolioWizard,
  ResolverOrigenDialog,
} from "@/widgets/portafolio"

// PortafolioView — composition-root de la vista global Portafolio (Slice 1 T7 + paquete
// 2026-07-23-portafolio-agregar-marketplace): TODO el transporte vive acá
// (`fe-transporte-independiente`) — los widgets de `widgets/portafolio` y `widgets/marketplace`
// son props puras. Es también la que COMPONE los dos widgets hermanos: ellos no se importan
// entre sí (`no-sibling-widget-imports`, severidad `error`).

// S1-D13: sin sesión activa, el botón de Observar queda disabled + este tooltip EXACTO (el
// Mapa vive en el stage de sesión, otra superficie que la ruta global `portafolio`).
const OBSERVAR_DISABLED_MOTIVO = "necesita una sesión abierta — el Mapa vive en el stage de sesión"

type Plano = "arneses" | "marketplaces"

// ── hash-state del plano (AG-D8 decisión 8): `#/portafolio?plano=marketplaces`. Sirve para que
// S6 pueda aterrizar ahí y para que un reload no pierda el lugar. Se escribe con `replaceState`
// (no `location.hash =`) para no ensuciar el historial ni disparar un `hashchange` por cada
// conmutación — la RUTA no cambió, solo el estado de la vista. ──
function planoDeHash(): Plano {
  if (typeof window === "undefined") return "arneses"
  return window.location.hash.includes("plano=marketplaces") ? "marketplaces" : "arneses"
}

function escribirPlanoEnHash(plano: Plano) {
  if (typeof window === "undefined") return
  const destino = plano === "marketplaces" ? "#/portafolio?plano=marketplaces" : "#/portafolio"
  if (window.location.hash !== destino) window.history.replaceState(null, "", destino)
}

// datoDelError — varios 4xx del daemon llevan un campo estructurado además del `error` textual
// (el `nombre` del 409 de registrar, el `destino` del 409 de traer). Se lee del cuerpo CRUDO por
// `ApiError.body`, nunca parseando el mensaje: el FE decide el copy por `status`, no por prosa.
function datoDelError(e: unknown, status: number, campo: string): string | undefined {
  if (!(e instanceof ApiError) || e.status !== status || e.body === "") return undefined
  try {
    const json = JSON.parse(e.body) as Record<string, unknown>
    const valor = json[campo]
    return typeof valor === "string" && valor !== "" ? valor : undefined
  } catch {
    return undefined
  }
}

function motivoDe(e: unknown): string {
  return e instanceof Error ? e.message : String(e)
}

export function PortafolioView() {
  // vivo evita setState tras unmount durante un fetch en vuelo (mismo patrón que AjustesView).
  const vivo = useRef(true)
  useEffect(() => {
    vivo.current = true
    return () => {
      vivo.current = false
    }
  }, [])

  // ── Plano activo (AG-D8 decisión 2): dos planos hermanos DENTRO de Portafolio, no una entrada
  // nueva del rail (mismo dominio, comparten el ＋ Agregar). ──
  const [plano, setPlanoState] = useState<Plano>(planoDeHash)
  const setPlano = useCallback((p: Plano) => {
    setPlanoState(p)
    escribirPlanoEnHash(p)
  }, [])

  // ── Lista de arneses: cargando→datos/error, refetch tras agregar/desvincular ──
  const [estado, setEstado] = useState<"cargando" | "error" | "datos">("cargando")
  const [error, setError] = useState<string>()
  const [entradas, setEntradas] = useState<EntradaPortafolio[]>([])
  const [corruptas, setCorruptas] = useState<EntradaCorrupta[]>([])
  const [lente, setLente] = useState<LentePortafolio>("empresa")
  const [busqueda, setBusqueda] = useState("")
  // Filtros Estado/Marketplace de la toolbar (deuda viva S1-D8, cerrada 2026-07-24): acotan la
  // lista, afordancia distinta de la lente — set vacío = sin filtro (mismo convenio que busqueda).
  const [filtroSalud, setFiltroSalud] = useState<ReadonlySet<SaludPortafolio>>(new Set())
  const [filtroMarketplace, setFiltroMarketplace] = useState<ReadonlySet<string>>(new Set())
  // Filtro «sin origen resuelto» (AG-D8 decisión 7): lo prende el contador cruzado del plano
  // Marketplaces. Es un filtro, no una lente.
  const [filtroSinOrigen, setFiltroSinOrigen] = useState(false)

  const cargar = useCallback(() => {
    setEstado("cargando")
    setError(undefined)
    api
      .listPortafolio<PortafolioListado>()
      .then((data) => {
        if (!vivo.current) return
        setEntradas(data.entradas ?? [])
        setCorruptas(data.corruptas ?? [])
        setEstado("datos")
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        setError(motivoDe(e))
        setEstado("error")
      })
  }, [])

  useEffect(() => {
    cargar()
  }, [cargar])

  // ── Telemetría del Portafolio (T36) ──
  //
  // Carga LAZY y **no bloqueante**, mismo patrón que `GET /api/marketplaces`: el Portafolio
  // tiene que listar aunque la telemetría no conteste. Un fallo acá **no** degrada la lista —
  // simplemente no hay columnas de mejora, que es la verdad.
  const [telemetria, setTelemetria] = useState<FilaPortafolio[]>([])

  useEffect(() => {
    let vivoLocal = true
    api
      .telemetriaPortafolio<{ filas?: FilaPortafolio[] }>()
      .then((data) => {
        if (vivoLocal && vivo.current) setTelemetria(data.filas ?? [])
      })
      .catch(() => {
        // Silencio deliberado: sin telemetría, las 3 columnas no se dibujan. El error de la
        // capa Mejora se reporta en su propia superficie (la franja del Mapa), no acá — un
        // banner de telemetría rota arriba del inventario de arneses es ruido en el lugar
        // equivocado.
        if (vivoLocal && vivo.current) setTelemetria([])
      })
    return () => {
      vivoLocal = false
    }
  }, [])

  /**
   * El wire CRUDO por clave. **La página ya no compone JSX**: hacerlo fue el crítico C-4 —
   * quedaron dos implementaciones de las mismas tres celdas, y el fix de D24.4 se aplicó a la
   * que la app no monta. Ahora las arma el widget, que tiene stories.
   *
   * Se keyea por `clave` y se toma la primera fila de cada una: el wire manda una fila por
   * `(arnés, instalación)` (D20) y la lista del Portafolio agrupa por entrada.
   *
   * 🔴 **Y se cuenta cuántas hay** (D26.4). Antes las demás desaparecían en silencio y la celda
   * mostraba el costo de UNA instalación donde se leía el del arnés. Ahora la fila declara su
   * alcance: una cifra parcial que no dice que es parcial es una cifra falsa.
   */
  const mejoraPorClave = useMemo(() => {
    const m = new Map<string, FilaPortafolio>()
    for (const f of telemetria) {
      if (!m.has(f.clave)) m.set(f.clave, f)
    }
    return m
  }, [telemetria])

  const instalacionesPorClave = useMemo(() => {
    const m = new Map<string, number>()
    for (const f of telemetria) m.set(f.clave, (m.get(f.clave) ?? 0) + 1)
    return m
  }, [telemetria])

  // ── Plano Marketplaces (S2) ──
  const [marketplaces, setMarketplaces] = useState<MarketplaceConocido[]>([])
  const [mkEstado, setMkEstado] = useState<"cargando" | "error" | "datos">("cargando")
  const [mkErrorLista, setMkErrorLista] = useState<string>()
  const [mkCorruptas, setMkCorruptas] = useState<EntradaCorruptaMarketplace[]>([])
  const [avisoDetector, setAvisoDetector] = useState<string>()
  const [sinOrigenResuelto, setSinOrigenResuelto] = useState(0)
  const [mkPedido, setMkPedido] = useState(false)
  const [leyendo, setLeyendo] = useState<ReadonlySet<string>>(new Set())

  const cargarMarketplaces = useCallback(() => {
    setMkPedido(true)
    setMkEstado("cargando")
    setMkErrorLista(undefined)
    api
      .listMarketplaces<ListadoMarketplaces>()
      .then((data) => {
        if (!vivo.current) return
        setMarketplaces(data.marketplaces ?? [])
        setMkCorruptas(data.corruptas ?? [])
        setAvisoDetector(data.aviso_detector || undefined)
        setSinOrigenResuelto(data.sin_origen_resuelto ?? 0)
        setMkEstado("datos")
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        setMkErrorLista(motivoDe(e))
        setMkEstado("error")
      })
  }, [])

  // ── Catálogo (S3/S4) ──
  const [catalogoAbierto, setCatalogoAbierto] = useState<string>()
  const [catalogo, setCatalogo] = useState<Catalogo>()
  const [catEstado, setCatEstado] = useState<"cargando" | "error" | "datos">("cargando")
  const [catError, setCatError] = useState<string>()
  const [catRefrescando, setCatRefrescando] = useState(false)
  const [catBusqueda, setCatBusqueda] = useState("")
  const [catFiltroSituacion, setCatFiltroSituacion] = useState<ReadonlySet<TipoSituacion>>(
    new Set(),
  )
  // AbortController del catálogo (§8.6 regla 5): cerrar el catálogo aborta SU fetch — cero
  // efectos. El POST de Traer NO se aborta nunca (§13.10): abortar a mitad de una
  // materialización es peor que esperar.
  const catAbortRef = useRef<AbortController | null>(null)

  const pedirCatalogo = useCallback((nombre: string, refrescar: boolean) => {
    catAbortRef.current?.abort()
    const controller = new AbortController()
    catAbortRef.current = controller
    if (refrescar) setCatRefrescando(true)
    else setCatEstado("cargando")
    setCatError(undefined)
    const promesa = refrescar
      ? api.leerCatalogo<Catalogo>(nombre, controller.signal)
      : api.catalogoDeMarketplace<Catalogo>(nombre, controller.signal)
    promesa
      .then((cat) => {
        if (!vivo.current) return
        setCatalogo(cat)
        setCatEstado("datos")
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        if (e instanceof DOMException && e.name === "AbortError") return
        setCatError(motivoDe(e))
        setCatEstado("error")
      })
      .finally(() => {
        if (catAbortRef.current === controller) catAbortRef.current = null
        if (vivo.current) setCatRefrescando(false)
      })
  }, [])

  const onAbrirCatalogo = useCallback(
    (nombre: string) => {
      setCatalogoAbierto(nombre)
      setCatalogo(undefined)
      setCatBusqueda("")
      setCatFiltroSituacion(new Set())
      pedirCatalogo(nombre, false)
    },
    [pedirCatalogo],
  )

  const onCerrarCatalogo = useCallback(() => {
    catAbortRef.current?.abort()
    catAbortRef.current = null
    setCatalogoAbierto(undefined)
    setCatalogo(undefined)
  }, [])

  // «Leer catálogo» / «Reintentar» de una fila del plano: refresco EXPLÍCITO que NO navega (la
  // fila sigue siendo no-navegable hasta que la lectura salga bien). Tras la lectura se refetchea
  // el listado, que es de donde sale el estado de la fila.
  const onLeerCatalogo = useCallback(
    (nombre: string) => {
      setLeyendo((prev) => new Set(prev).add(nombre))
      api
        .leerCatalogo<Catalogo>(nombre)
        .then(() => {
          if (vivo.current) cargarMarketplaces()
        })
        .catch((e: unknown) => {
          // El endpoint devuelve 200 incluso cuando no pudo leer (el motivo viaja en
          // `lectura.motivo`), así que un error acá es del transporte: se muestra como error del
          // plano y la lista se recarga igual para no quedar con datos a medias.
          if (!vivo.current) return
          setMkErrorLista(motivoDe(e))
          cargarMarketplaces()
        })
        .finally(() => {
          if (!vivo.current) return
          setLeyendo((prev) => {
            const next = new Set(prev)
            next.delete(nombre)
            return next
          })
        })
    },
    [cargarMarketplaces],
  )

  // ── Drawer: qué fila está abierta + desvincular ──
  const [seleccionada, setSeleccionada] = useState<string>()
  const [desvinculando, setDesvinculando] = useState(false)
  const [desvincularError, setDesvincularError] = useState<string>()
  const [observarError, setObservarError] = useState<string>()
  const [identificando, setIdentificando] = useState(false)
  const [identificarError, setIdentificarError] = useState<string>()

  const seleccionadaEntrada = useMemo(
    () => entradas.find((e) => e.clave === seleccionada),
    [entradas, seleccionada],
  )

  // Lazy por plano (§8.6 regla 1): el `GET /api/marketplaces` se pide al ENTRAR al plano
  // Marketplaces, no al montar la vista. Y también cuando se abre el drawer de un arnés con
  // `home` declarado: sin la lista no se puede resolver el NOMBRE del marketplace (la clave del
  // endpoint) desde el `repo` de la identidad (C17: el path es el nombre, la identidad es el
  // repo canónico) y la segunda puerta de `↧ Traer canónico` quedaría muerta.
  const necesitaMarketplaces = plano === "marketplaces" || !!seleccionadaEntrada?.identidad.home
  useEffect(() => {
    if (necesitaMarketplaces && !mkPedido) cargarMarketplaces()
  }, [necesitaMarketplaces, mkPedido, cargarMarketplaces])

  const onAbrirFila = useCallback((clave: string) => {
    setSeleccionada(clave)
    setDesvinculando(false)
    setDesvincularError(undefined)
    setObservarError(undefined)
    setIdentificarError(undefined)
  }, [])

  const onCerrarDrawer = useCallback(() => setSeleccionada(undefined), [])

  const onDesvincular = useCallback(() => {
    const clave = seleccionada
    if (!clave) return
    setDesvinculando(true)
    setDesvincularError(undefined)
    api
      .desvincularDelPortafolio(clave)
      .then(() => {
        if (!vivo.current) return
        setSeleccionada(undefined)
        cargar()
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        setDesvincularError(motivoDe(e))
      })
      .finally(() => {
        if (vivo.current) setDesvinculando(false)
      })
  }, [seleccionada, cargar])

  // ── Identificar (S1-D28): sella arnes.l0.json in-situ + re-key ──
  const onIdentificar = useCallback(
    (installPath: string, id: string, nombre: string) => {
      const clave = seleccionada
      if (!clave) return
      setIdentificando(true)
      setIdentificarError(undefined)
      api
        .identificarArnes(clave, installPath, id, nombre)
        .then(() => {
          if (!vivo.current) return
          // La entrada re-keyea a una clave nueva; refrescamos y cerramos el drawer (el
          // usuario reabre la fila ya sellada desde la lista actualizada).
          setSeleccionada(undefined)
          cargar()
        })
        .catch((e: unknown) => {
          if (!vivo.current) return
          setIdentificarError(motivoDe(e))
        })
        .finally(() => {
          if (vivo.current) setIdentificando(false)
        })
    },
    [seleccionada, cargar],
  )

  // ── Abrir/Observar en Mapa (S1-D1/D2/D13) ──
  const activeSession = useSessions(selectActive)
  const parkView = useSessions((s) => s.parkView)
  const setMapaPeek = useAppStore((s) => s.setMapaPeek)
  const setView = useAppStore((s) => s.setView)

  const ejecutarObservar = useCallback(
    (clave: string, installPath: string) => {
      setObservarError(undefined)
      api
        .observarEnMapa<{ id: string; indexed: boolean }>(clave, installPath)
        .then((res) => {
          if (!vivo.current) return
          setMapaPeek(res.id)
          void parkView("Mapa")
          setView("mapa")
        })
        .catch((e: unknown) => {
          if (!vivo.current) return
          setObservarError(motivoDe(e))
        })
    },
    [parkView, setMapaPeek, setView],
  )

  const onObservarInstalacion = useCallback(
    (installPath: string) => {
      if (!seleccionadaEntrada) return
      // Deuda BACKLOG «re-key (home,id,scope)», cerrada 2026-07-23: el índice del Mapa
      // indexa por CLAVE calificada — dos entradas con el mismo id pelado ya no se pisan,
      // así que la confirmación previa (S1-D2) ya no protege nada real. Observa directo.
      ejecutarObservar(seleccionadaEntrada.clave, installPath)
    },
    [seleccionadaEntrada, ejecutarObservar],
  )

  // S1-D13: sin sesión activa la prop queda undefined (no una función que falle) — el drawer
  // deshabilita el/los botón(es) de Observar con el tooltip exacto.
  const onObservarProp = activeSession ? onObservarInstalacion : undefined

  // ── `↧ Traer canónico` (S8, AG-D17) — DOS PUERTAS, UN ACTO: la fila del catálogo y el botón del
  // drawer llaman a este mismo callback. El estado se keyea por `nombre@entrada` (§13.10): un
  // Traer en vuelo por IDENTIDAD, no global — se puede traer `harness` y `harness-beta` a la vez
  // (destinos distintos, E-78). ──
  const [traer, setTraer] = useState<Record<string, EstadoTraer>>({})

  const onTraerCanonico = useCallback(
    (nombre: string, entradaNombre: string) => {
      const clave = `${nombre}@${entradaNombre}`
      setTraer((t) => ({ ...t, [clave]: { fase: "trayendo" } }))
      api
        .traerCanonico<ResultadoTraer>(nombre, entradaNombre)
        .then((res) => {
          if (!vivo.current) return
          const entrada = res.entrada as { clave?: unknown } | undefined
          const clavePortafolio = typeof entrada?.clave === "string" ? entrada.clave : undefined
          setTraer((t) => ({
            ...t,
            [clave]: {
              fase: "traido",
              destino: res.destino,
              camino: res.camino,
              deriva: res.deriva,
              ...(res.deriva_detalle ? { derivaDetalle: res.deriva_detalle } : {}),
              ...(res.avisos && res.avisos.length > 0 ? { avisos: res.avisos } : {}),
              ...(clavePortafolio ? { clavePortafolio } : {}),
            },
          }))
          // Tras el 200 se refetchean LAS DOS superficies: el Portafolio (aparece el canónico) y
          // el catálogo (la fila deja de decir `no-lo-tengo`). Sin esto, la columna de situación
          // quedaría mintiendo con el estado viejo.
          cargar()
          if (catalogoAbierto === nombre) pedirCatalogo(nombre, false)
        })
        .catch((e: unknown) => {
          if (!vivo.current) return
          const destinoExistente = datoDelError(e, 409, "destino")
          setTraer((t) => ({
            ...t,
            [clave]: {
              fase: "fallo",
              motivo: motivoDe(e),
              ...(destinoExistente ? { destinoExistente } : {}),
            },
          }))
        })
    },
    [cargar, catalogoAbierto, pedirCatalogo],
  )

  // El estado de Traer que le toca al catálogo abierto, ya des-prefijado (el widget solo conoce
  // nombres de entrada; la clave compuesta es asunto de la página).
  const traerDelCatalogo = useMemo(() => {
    if (!catalogoAbierto) return {}
    const prefijo = `${catalogoAbierto}@`
    const out: Record<string, EstadoTraer> = {}
    for (const [clave, valor] of Object.entries(traer)) {
      if (clave.startsWith(prefijo)) out[clave.slice(prefijo.length)] = valor
    }
    return out
  }, [traer, catalogoAbierto])

  // La segunda puerta: desde el drawer. Para pegarle al endpoint hace falta el NOMBRE del
  // marketplace, y la identidad del arnés trae el REPO canonicalizado (C17) — se cruza contra la
  // lista ya cargada. Si no se puede resolver, la prop queda undefined y el botón sigue
  // `disabled` como en el Slice 1: nunca un botón que no puede funcionar.
  const marketplaceDelDrawer = useMemo(() => {
    const home = seleccionadaEntrada?.identidad.home
    if (!home) return undefined
    return marketplaces.find((m) => m.repo === home)
  }, [marketplaces, seleccionadaEntrada])

  const traerDelDrawer =
    marketplaceDelDrawer && seleccionadaEntrada
      ? traer[`${marketplaceDelDrawer.nombre}@${seleccionadaEntrada.identidad.id}`]
      : undefined

  // No se pre-filtra por clase: si el marketplace es `referencia`, el dominio rechaza con SU
  // literal (`no aplica: solo arneses propios`, BR-1) y ese motivo se muestra tal cual. Preferimos
  // un round-trip de más antes que teclear un literal del dominio en el FE.
  const onTraerDelDrawer = useMemo(() => {
    if (!marketplaceDelDrawer || !seleccionadaEntrada?.identidad.id) return undefined
    const nombre = marketplaceDelDrawer.nombre
    const entradaNombre = seleccionadaEntrada.identidad.id
    return () => onTraerCanonico(nombre, entradaNombre)
  }, [marketplaceDelDrawer, seleccionadaEntrada, onTraerCanonico])

  // ── Wizard: fuente→escaneando→candidatos→agregando (rama Proyecto) + url→validando→validado→
  // registrando (rama Marketplace). AbortController del escaneo (S1-D9) y de la validación. ──
  const [wizardAbierto, setWizardAbierto] = useState(false)
  const [wizardRama, setWizardRama] = useState<"proyecto" | "marketplace">("proyecto")
  const [wizardEstado, setWizardEstado] = useState<
    "fuente" | "escaneando" | "candidatos" | "agregando"
  >("fuente")
  const [wizardError, setWizardError] = useState<string>()
  const [candidatos, setCandidatos] = useState<Candidato[]>()
  const [pathEscaneado, setPathEscaneado] = useState("")
  const [mkFase, setMkFase] = useState<"url" | "validando" | "validado" | "registrando">("url")
  const [mkValidacion, setMkValidacion] = useState<Validacion>()
  const [mkErrorWizard, setMkErrorWizard] = useState<string>()
  const [mkYaRegistrado, setMkYaRegistrado] = useState<{ nombre: string }>()
  const abortRef = useRef<AbortController | null>(null)

  const clavesExistentes = useMemo(() => new Set(entradas.map((e) => e.clave)), [entradas])

  const onAbrirWizard = useCallback(() => {
    setWizardRama("proyecto")
    setWizardEstado("fuente")
    setWizardError(undefined)
    setCandidatos(undefined)
    setPathEscaneado("")
    setMkFase("url")
    setMkValidacion(undefined)
    setMkErrorWizard(undefined)
    setMkYaRegistrado(undefined)
    setWizardAbierto(true)
  }, [])

  const onCerrarWizard = useCallback(() => {
    // Cancelar/cerrar el wizard = CERO efectos (S1-D9): si había un escaneo o una validación en
    // vuelo, se aborta (ninguno de los dos persiste nada por diseño — abortarlos no deja basura).
    abortRef.current?.abort()
    abortRef.current = null
    setWizardAbierto(false)
  }, [])

  const onRama = useCallback((r: "proyecto" | "marketplace") => {
    abortRef.current?.abort()
    abortRef.current = null
    setWizardRama(r)
  }, [])

  const onEscanear = useCallback((path: string) => {
    setPathEscaneado(path)
    setWizardEstado("escaneando")
    setWizardError(undefined)
    const controller = new AbortController()
    abortRef.current = controller
    api
      .escanearProyecto<Candidato[]>(path, controller.signal)
      .then((cands) => {
        if (!vivo.current) return
        setCandidatos(cands ?? [])
        setWizardEstado("candidatos")
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        if (e instanceof DOMException && e.name === "AbortError") {
          // Cancelado explícitamente (← Atrás desde `escaneando`) — vuelve al paso fuente, sin
          // motivo de error (no es un fallo, es una elección del usuario).
          setWizardEstado("fuente")
          return
        }
        setCandidatos(undefined)
        setWizardError(motivoDe(e))
        setWizardEstado("candidatos")
      })
      .finally(() => {
        if (abortRef.current === controller) abortRef.current = null
      })
  }, [])

  const onCancelarEscaneo = useCallback(() => {
    abortRef.current?.abort()
  }, [])

  // `← Atrás` desde `candidatos` (y su gemelo `cambiar ruta`): vuelve al paso 1 preservando la
  // ruta y la selección (el widget las posee y no desmonta). Una transición, dos puertas.
  const onAtrasWizard = useCallback(() => {
    abortRef.current?.abort()
    abortRef.current = null
    if (wizardRama === "marketplace") {
      setMkFase("url")
      setMkErrorWizard(undefined)
      setMkYaRegistrado(undefined)
      return
    }
    setWizardError(undefined)
    setWizardEstado("fuente")
  }, [wizardRama])

  const onAgregarCandidatos = useCallback(
    (elegidos: string[]) => {
      setWizardEstado("agregando")
      api
        .agregarProyecto<EntradaPortafolio[]>(pathEscaneado, elegidos)
        .then(() => {
          if (!vivo.current) return
          setWizardAbierto(false)
          cargar()
        })
        .catch((e: unknown) => {
          if (!vivo.current) return
          // Sin slot dedicado a "error de agregar" en el contrato §2.6 (cerrado) — se reusa el
          // mismo slot textual `error` del paso candidatos (S1-D22): el motivo real del 400
          // queda visible, el usuario reintenta desde ahí. Documentado en decisiones.md.
          setWizardError(motivoDe(e))
          setWizardEstado("candidatos")
        })
    },
    [pathEscaneado, cargar],
  )

  // ── S6 · validar y registrar un marketplace ──
  const onValidarMarketplace = useCallback((url: string) => {
    setMkFase("validando")
    setMkErrorWizard(undefined)
    setMkYaRegistrado(undefined)
    const controller = new AbortController()
    abortRef.current = controller
    api
      .validarMarketplace<Validacion>(url, controller.signal)
      .then((v) => {
        if (!vivo.current) return
        setMkValidacion(v)
        setMkFase("validado")
        // El backend ya sabe si el nombre está tomado: se avisa ANTES de intentar registrar, sin
        // pisar nada (BR-7). El 409 sigue siendo el enforcement real.
        if (v.ya_registrado) setMkYaRegistrado({ nombre: v.nombre })
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        if (e instanceof DOMException && e.name === "AbortError") {
          setMkFase("url")
          return
        }
        // 400 («tu url no sirve») y 503 («no puedo mirar») son semánticamente distintos y el
        // motivo del backend ya lo dice; el FE no re-interpreta ni suaviza. Y **sin un 200 no hay
        // ✓ posible**: la máquina de estados vuelve a `url` (BR-5/G3).
        setMkValidacion(undefined)
        setMkErrorWizard(motivoDe(e))
        setMkFase("url")
      })
      .finally(() => {
        if (abortRef.current === controller) abortRef.current = null
      })
  }, [])

  // Aterrizaje de S6 (§8.6 regla 2): cierra el wizard → conmuta al plano Marketplaces → abre el
  // catálogo de la fila nueva → refetch de la lista. Un solo callback.
  const aterrizarEnMarketplace = useCallback(
    (nombre: string) => {
      setWizardAbierto(false)
      setPlano("marketplaces")
      cargarMarketplaces()
      onAbrirCatalogo(nombre)
    },
    [setPlano, cargarMarketplaces, onAbrirCatalogo],
  )

  const onRegistrarMarketplace = useCallback(
    (url: string, clase: ClaseMarketplace) => {
      setMkFase("registrando")
      setMkErrorWizard(undefined)
      setMkYaRegistrado(undefined)
      api
        .registrarMarketplace<MarketplaceConocido>(url, clase)
        .then((m) => {
          if (!vivo.current) return
          aterrizarEnMarketplace(m.nombre)
        })
        .catch((e: unknown) => {
          if (!vivo.current) return
          // 409 (BR-7): ya registrado — NO se pisa nada. Se ofrece «ir a él» con el `nombre` que
          // vino en el body, sin parsear el texto del error.
          const nombre = datoDelError(e, 409, "nombre")
          setMkErrorWizard(motivoDe(e))
          if (nombre) setMkYaRegistrado({ nombre })
          setMkFase("validado")
        })
    },
    [aterrizarEnMarketplace],
  )

  // ── S7 · Resolver origen (AG-D8 decisión 7) ──
  const [resolverAbierto, setResolverAbierto] = useState(false)
  const [candidatosOrigen, setCandidatosOrigen] = useState<CandidatoOrigen[]>([])
  const [origenActual, setOrigenActual] = useState<string>()
  const [resolverEleccion, setResolverEleccion] = useState<string | null | undefined>(undefined)
  const [resolverConfirmando, setResolverConfirmando] = useState(false)
  const [resolverError, setResolverError] = useState<string>()
  // El foco vuelve al botón que abrió el diálogo (§8.6 regla 4): abrir S7 NO cierra el drawer.
  const resolverOrigenBtnRef = useRef<HTMLElement | null>(null)

  const onResolverOrigen = useCallback(() => {
    const clave = seleccionada
    if (!clave) return
    resolverOrigenBtnRef.current =
      document.activeElement instanceof HTMLElement ? document.activeElement : null
    setResolverAbierto(true)
    setResolverEleccion(undefined)
    setResolverError(undefined)
    setCandidatosOrigen([])
    setOrigenActual(undefined)
    api
      .candidatosDeOrigen<CandidatosOrigen>(clave)
      .then((data) => {
        if (!vivo.current) return
        setCandidatosOrigen(data.candidatos ?? [])
        setOrigenActual(data.actual || undefined)
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        setResolverError(motivoDe(e))
      })
  }, [seleccionada])

  const onCerrarResolver = useCallback(() => {
    setResolverAbierto(false)
    resolverOrigenBtnRef.current?.focus()
  }, [])

  const onConfirmarOrigen = useCallback(() => {
    const clave = seleccionada
    if (!clave || resolverEleccion === undefined) return
    setResolverConfirmando(true)
    setResolverError(undefined)
    api
      .asignarOrigen<EntradaPortafolio>(clave, resolverEleccion)
      .then((entrada) => {
        if (!vivo.current) return
        setResolverAbierto(false)
        // Asignar un home RE-KEYEA la entrada (igual que `identificar`): la clave vieja ya no
        // existe, así que el drawer re-apunta a la nueva y la lista se refetchea.
        setSeleccionada(entrada.clave)
        cargar()
        if (mkPedido) cargarMarketplaces()
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        // 409 = colisión de identidad: se EXPLICA en el slot `error` del diálogo y las dos
        // entradas quedan intactas (fusionar es otra operación, C-ID-2).
        setResolverError(motivoDe(e))
      })
      .finally(() => {
        if (vivo.current) setResolverConfirmando(false)
      })
  }, [seleccionada, resolverEleccion, cargar, mkPedido, cargarMarketplaces])

  // Contador cruzado → plano Arneses ya filtrado (§8.6 regla 3).
  const onVerSinOrigen = useCallback(() => {
    setPlano("arneses")
    setCatalogoAbierto(undefined)
    setFiltroSinOrigen(true)
  }, [setPlano])

  // onElegirCarpeta (S1-D9, patrón RF-110 de AjustesView): SOLO dentro de Tauri. Un rechazo del
  // picker nativo (cancelación del SO, permiso denegado) se ignora silenciosamente — el widget
  // no tiene un slot de error dedicado para esta rama (ver su propio comentario inline).
  const onElegirCarpeta = useCallback(async () => {
    try {
      const seleccion = await elegirCarpeta({
        directory: true,
        title: "Elegí la carpeta del proyecto",
      })
      if (!seleccion || Array.isArray(seleccion)) return undefined
      return seleccion
    } catch {
      return undefined
    }
  }, [])

  return (
    <div className="arnesia-portafolio flex h-full flex-col">
      <header className="border-b border-border bg-card px-[18px] py-3 text-sm text-muted-foreground">
        alpacapurpura / <b className="text-foreground">Portafolio</b>
      </header>

      {/* AG-D8 decisión 2 — dos planos hermanos, no una entrada nueva del rail. */}
      <div className="pf-planos" role="tablist" aria-label="Planos del Portafolio">
        <button
          type="button"
          role="tab"
          aria-selected={plano === "arneses"}
          className={plano === "arneses" ? "pf-plano-tab activa" : "pf-plano-tab"}
          onClick={() => {
            setPlano("arneses")
            setCatalogoAbierto(undefined)
          }}
        >
          Arneses
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={plano === "marketplaces"}
          className={plano === "marketplaces" ? "pf-plano-tab activa" : "pf-plano-tab"}
          onClick={() => setPlano("marketplaces")}
        >
          Marketplaces
          {sinOrigenResuelto > 0 && (
            <span
              className="pf-plano-cta"
              title={`${sinOrigenResuelto} arneses sin origen resuelto`}
            >
              {sinOrigenResuelto}
            </span>
          )}
        </button>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        {plano === "arneses" && (
          <PortafolioList
            estado={estado}
            error={error}
            entradas={entradas}
            corruptas={corruptas}
            lente={lente}
            onLente={setLente}
            busqueda={busqueda}
            onBusqueda={setBusqueda}
            filtroSalud={filtroSalud}
            onFiltroSalud={setFiltroSalud}
            filtroMarketplace={filtroMarketplace}
            onFiltroMarketplace={setFiltroMarketplace}
            filtroSinOrigen={filtroSinOrigen}
            onFiltroSinOrigen={setFiltroSinOrigen}
            seleccionada={seleccionada}
            onAbrir={onAbrirFila}
            onAgregar={onAbrirWizard}
            onReintentar={cargar}
            mejora={mejoraPorClave.size > 0 ? mejoraPorClave : undefined}
            instalacionesPorClave={instalacionesPorClave}
          />
        )}

        {plano === "marketplaces" && !catalogoAbierto && (
          <MarketplaceList
            estado={mkEstado}
            error={mkErrorLista}
            marketplaces={marketplaces}
            corruptas={mkCorruptas}
            avisoDetector={avisoDetector}
            sinOrigenResuelto={sinOrigenResuelto}
            leyendo={leyendo}
            onAbrirCatalogo={onAbrirCatalogo}
            onLeerCatalogo={onLeerCatalogo}
            onAgregar={onAbrirWizard}
            onReintentar={cargarMarketplaces}
            onVerSinOrigen={onVerSinOrigen}
          />
        )}

        {plano === "marketplaces" && catalogoAbierto && (
          <MarketplaceCatalogo
            estado={catEstado}
            error={catError}
            catalogo={catalogo}
            busqueda={catBusqueda}
            onBusqueda={setCatBusqueda}
            filtroSituacion={catFiltroSituacion}
            onFiltroSituacion={setCatFiltroSituacion}
            onVolver={onCerrarCatalogo}
            onRefrescar={() => pedirCatalogo(catalogoAbierto, true)}
            refrescando={catRefrescando}
            onVerFicha={(clave) => {
              setPlano("arneses")
              setCatalogoAbierto(undefined)
              onAbrirFila(clave)
            }}
            onTraerCanonico={(entradaNombre) => onTraerCanonico(catalogoAbierto, entradaNombre)}
            traer={traerDelCatalogo}
          />
        )}
      </div>

      {wizardAbierto && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/40 p-4">
          <PortafolioWizard
            abierto={wizardAbierto}
            onClose={onCerrarWizard}
            estado={wizardEstado}
            error={wizardError}
            candidatos={candidatos}
            clavesExistentes={clavesExistentes}
            onEscanear={onEscanear}
            onCancelarEscaneo={onCancelarEscaneo}
            onAgregar={onAgregarCandidatos}
            onElegirCarpeta={isTauri() ? onElegirCarpeta : undefined}
            onAtras={onAtrasWizard}
            pathEscaneado={pathEscaneado}
            rama={wizardRama}
            onRama={onRama}
            mkEstado={mkFase}
            mkValidacion={mkValidacion}
            mkError={mkErrorWizard}
            mkYaRegistrado={mkYaRegistrado}
            onValidarMarketplace={onValidarMarketplace}
            onRegistrarMarketplace={onRegistrarMarketplace}
            onIrAMarketplace={aterrizarEnMarketplace}
          />
        </div>
      )}

      {seleccionadaEntrada && (
        <div className="fixed inset-0 z-40 flex items-stretch justify-end bg-foreground/20 p-4">
          <div className="flex max-h-full w-full max-w-[520px] flex-col gap-2 overflow-auto">
            {/* S1-D21: PortafolioDrawerProps (§2.6, cerrado) no trae un slot para el error de
                `onObservar` (el callback es fire-and-forget) — se muestra como banner propio de
                la página, por fuera del widget, en vez de tocar el contrato. */}
            {observarError && (
              <p role="alert" className="pf-error">
                No se pudo observar en el Mapa — {observarError}
              </p>
            )}
            <PortafolioDrawer
              entrada={seleccionadaEntrada}
              onClose={onCerrarDrawer}
              onObservar={onObservarProp}
              observarDisabledMotivo={OBSERVAR_DISABLED_MOTIVO}
              desvinculando={desvinculando}
              desvincularError={desvincularError}
              onDesvincular={onDesvincular}
              onIdentificar={onIdentificar}
              identificando={identificando}
              identificarError={identificarError}
              onResolverOrigen={onResolverOrigen}
              onTraerCanonico={onTraerDelDrawer}
              trayendo={traerDelDrawer?.fase === "trayendo"}
              traerError={traerDelDrawer?.fase === "fallo" ? traerDelDrawer.motivo : undefined}
            />
          </div>
        </div>
      )}

      {/* §8.6 regla 4 — S7 se abre SOBRE el drawer (el drawer es el contexto de la fila): abrir
          S7 no lo cierra, y cerrar S7 devuelve el foco al botón que lo abrió. */}
      {resolverAbierto && seleccionadaEntrada && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center bg-foreground/40 p-4">
          <ResolverOrigenDialog
            abierto={resolverAbierto}
            arnesId={
              seleccionadaEntrada.identidad.id || (seleccionadaEntrada.identidad.scope ?? "")
            }
            candidatos={candidatosOrigen}
            actual={origenActual}
            eleccion={resolverEleccion}
            onElegir={setResolverEleccion}
            confirmando={resolverConfirmando}
            error={resolverError}
            onConfirmar={onConfirmarOrigen}
            onClose={onCerrarResolver}
          />
        </div>
      )}
    </div>
  )
}
