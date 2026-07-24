import { open as elegirCarpeta } from "@tauri-apps/plugin-dialog"
import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import type {
  Candidato,
  EntradaCorrupta,
  EntradaPortafolio,
  LentePortafolio,
  PortafolioListado,
} from "@/entities/portafolio"
import { api, isTauri, selectActive, useAppStore, useSessions } from "@/shared"
import { PortafolioDrawer, PortafolioList, PortafolioWizard } from "@/widgets/portafolio"

// PortafolioView — composition-root de la vista global Portafolio (Slice 1, plan §2.7/§3 T7):
// TODO el transporte vive acá (fe-transporte-independiente) — los 3 widgets
// (widgets/portafolio/ui/portafolio-{list,drawer,wizard}.tsx) son props puras. Mismo patrón
// que AjustesView (RF-100, pages/shell/ui/global-view.tsx): la página es la única que importa
// `api`.

// S1-D13: sin sesión activa, el botón de Observar queda disabled + este tooltip EXACTO (el
// Mapa vive en el stage de sesión, otra superficie que la ruta global `portafolio`).
const OBSERVAR_DISABLED_MOTIVO = "necesita una sesión abierta — el Mapa vive en el stage de sesión"

export function PortafolioView() {
  // vivo evita setState tras unmount durante un fetch en vuelo (mismo patrón que AjustesView).
  const vivo = useRef(true)
  useEffect(() => {
    vivo.current = true
    return () => {
      vivo.current = false
    }
  }, [])

  // ── Lista: cargando→datos/error, refetch tras agregar/desvincular ──
  const [estado, setEstado] = useState<"cargando" | "error" | "datos">("cargando")
  const [error, setError] = useState<string>()
  const [entradas, setEntradas] = useState<EntradaPortafolio[]>([])
  const [corruptas, setCorruptas] = useState<EntradaCorrupta[]>([])
  const [lente, setLente] = useState<LentePortafolio>("empresa")
  const [busqueda, setBusqueda] = useState("")

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
        setError(e instanceof Error ? e.message : String(e))
        setEstado("error")
      })
  }, [])

  useEffect(() => {
    cargar()
  }, [cargar])

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
        setDesvincularError(e instanceof Error ? e.message : String(e))
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
          setIdentificarError(e instanceof Error ? e.message : String(e))
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
          setObservarError(e instanceof Error ? e.message : String(e))
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

  // ── Wizard: fuente→escaneando→candidatos→agregando, AbortController del escaneo (S1-D9) ──
  const [wizardAbierto, setWizardAbierto] = useState(false)
  const [wizardEstado, setWizardEstado] = useState<
    "fuente" | "escaneando" | "candidatos" | "agregando"
  >("fuente")
  const [wizardError, setWizardError] = useState<string>()
  const [candidatos, setCandidatos] = useState<Candidato[]>()
  const [pathEscaneado, setPathEscaneado] = useState("")
  const abortRef = useRef<AbortController | null>(null)

  const clavesExistentes = useMemo(() => new Set(entradas.map((e) => e.clave)), [entradas])

  const onAbrirWizard = useCallback(() => {
    setWizardEstado("fuente")
    setWizardError(undefined)
    setCandidatos(undefined)
    setPathEscaneado("")
    setWizardAbierto(true)
  }, [])

  const onCerrarWizard = useCallback(() => {
    // Cancelar/cerrar el wizard = CERO efectos (S1-D9): si había un escaneo en vuelo, se aborta
    // (el POST /escaneos no persiste nada por diseño — abortarlo no deja basura).
    abortRef.current?.abort()
    abortRef.current = null
    setWizardAbierto(false)
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
          // Cancelado explícitamente (onCancelarEscaneo) — vuelve al paso fuente, sin motivo
          // de error (no es un fallo, es una elección del usuario).
          setWizardEstado("fuente")
          return
        }
        setCandidatos(undefined)
        setWizardError(e instanceof Error ? e.message : String(e))
        setWizardEstado("candidatos")
      })
      .finally(() => {
        if (abortRef.current === controller) abortRef.current = null
      })
  }, [])

  const onCancelarEscaneo = useCallback(() => {
    abortRef.current?.abort()
  }, [])

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
          setWizardError(e instanceof Error ? e.message : String(e))
          setWizardEstado("candidatos")
        })
    },
    [pathEscaneado, cargar],
  )

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
      <div className="min-h-0 flex-1 overflow-auto">
        <PortafolioList
          estado={estado}
          error={error}
          entradas={entradas}
          corruptas={corruptas}
          lente={lente}
          onLente={setLente}
          busqueda={busqueda}
          onBusqueda={setBusqueda}
          seleccionada={seleccionada}
          onAbrir={onAbrirFila}
          onAgregar={onAbrirWizard}
          onReintentar={cargar}
        />
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
            />
          </div>
        </div>
      )}
    </div>
  )
}
