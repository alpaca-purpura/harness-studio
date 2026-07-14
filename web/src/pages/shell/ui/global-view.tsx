import { open as elegirCarpeta } from "@tauri-apps/plugin-dialog"
import { useCallback, useEffect, useRef, useState } from "react"
import {
  type SelfUpdateReport,
  UpdateCard,
  type UpdateEstado,
  type VersionInfo,
} from "@/features/self-update"
import { ApiError, api, ComingSoon, isTauri, useAppStore } from "@/shared"
import { PortafolioView } from "./portafolio-view"

// GLOBAL — el ComingSoon que le queda a lo que sigue staged. «portafolio» murió de acá
// (Slice 1, plan §2.7/§3 T7): la ruta la sirve <PortafolioView/> real, más abajo.
const GLOBAL: Record<string, { glyph: string; title: string; note: string }> = {
  estandar: {
    glyph: "⟳",
    title: "Estándar as code",
    note: "El árbol de conocimiento vivo (11 nodos · 122 checks) y su cadencia semanal. Próximamente.",
  },
}

// Polling del reinicio (RF-105, decisión #9): el primer intento espera ~1.5s (el
// re-exec tarda <1s — así el camino normal no ensucia la consola con fetches al daemon
// caído), luego cada 800ms hasta ver huella nueva o agotar 30s.
const POLL_PRIMERO_MS = 1500
const POLL_INTERVALO_MS = 800
const POLL_TIMEOUT_MS = 30_000

const espera = (ms: number) => new Promise((r) => setTimeout(r, ms))

// AjustesView — la vista global Ajustes REAL (RF-100, decisión #5): v1 = SOLO la
// tarjeta «Versión y actualización» + una línea muted con lo que llega después (texto,
// jamás tarjetas vacías fingiendo). La página es el composition-root: transporte
// (getVersion/selfUpdate) + máquina de estados + polling del reinicio viven AQUÍ; la
// tarjeta es UI pura (fe-transporte-independiente, patrón Inspector).
function AjustesView() {
  const [version, setVersion] = useState<VersionInfo | null>(null)
  const [versionError, setVersionError] = useState<string>()
  const [estado, setEstado] = useState<UpdateEstado>("idle")
  const [reporte, setReporte] = useState<SelfUpdateReport>()
  const [mensaje, setMensaje] = useState<string>()
  // configurandoRepo/repoConfigError (RF-110, bugfix fix-repo-self-update): el picker
  // nativo + PUT /api/self-update/repo, SEPARADO de la máquina de estados del update
  // en sí (esto configura, no dispara).
  const [configurandoRepo, setConfigurandoRepo] = useState(false)
  const [repoConfigError, setRepoConfigError] = useState<string>()
  // vivo evita setState tras unmount durante el polling largo del reinicio.
  const vivo = useRef(true)
  useEffect(() => {
    vivo.current = true
    return () => {
      vivo.current = false
    }
  }, [])

  useEffect(() => {
    let alive = true
    api
      .getVersion<VersionInfo>()
      .then((v) => {
        if (alive) setVersion(v)
      })
      .catch((e: unknown) => {
        if (alive) setVersionError(e instanceof Error ? e.message : String(e))
      })
    return () => {
      alive = false
    }
  }, [])

  // pollReinicio (RF-105): reconecta tras el re-exec y confirma la huella NUEVA.
  // Timeout → error honesto («el daemon no volvió»), jamás un éxito fingido.
  const pollReinicio = useCallback(async (huellaVieja: string, esperada?: string) => {
    const limite = Date.now() + POLL_TIMEOUT_MS
    await espera(POLL_PRIMERO_MS)
    while (Date.now() < limite) {
      if (!vivo.current) return
      try {
        const v = await api.getVersion<VersionInfo>()
        const confirmada = esperada ? v.huella === esperada : v.huella !== huellaVieja
        if (confirmada) {
          if (vivo.current) {
            setVersion(v)
            setEstado("exito")
          }
          return
        }
      } catch {
        // El daemon aún no volvió — se sigue intentando hasta el límite.
      }
      await espera(POLL_INTERVALO_MS)
    }
    if (vivo.current) {
      setEstado("error")
      setMensaje("el daemon no volvió tras el reinicio (30s) — revisa el proceso arnesia serve")
    }
  }, [])

  const onUpdate = useCallback(() => {
    const huellaVieja = version?.huella ?? ""
    setEstado("actualizando")
    setReporte(undefined)
    setMensaje(undefined)
    api
      .selfUpdate<SelfUpdateReport>()
      .then((rep) => {
        if (!vivo.current) return
        setReporte(rep)
        if (rep.resultado === "fallo") {
          setEstado("error")
        } else if (rep.resultado === "ya-al-dia") {
          setEstado("ya-al-dia")
        } else {
          setEstado("reiniciando")
          void pollReinicio(huellaVieja, rep.huella_nueva)
        }
      })
      .catch((e: unknown) => {
        if (!vivo.current) return
        setEstado("error")
        setMensaje(
          e instanceof ApiError && e.status === 409
            ? "ya hay un self-update en curso (409) — espera a que termine"
            : e instanceof Error
              ? e.message
              : String(e),
        )
      })
  }, [version, pollReinicio])

  // onElegirRepo (RF-110): diálogo nativo de carpeta → PUT /api/self-update/repo →
  // refresca GET /api/version. Un candidato inválido NO rompe el flujo — el motivo
  // exacto del servidor viaja a repoConfigError (jamás un fallo mudo). SOLO se pasa a
  // la tarjeta dentro de Tauri (isTauri()); en browser plano la prop queda undefined y
  // el estado sinRepo sigue igual que siempre.
  const onElegirRepo = useCallback(async () => {
    setRepoConfigError(undefined)
    let seleccion: string | string[] | null
    try {
      seleccion = await elegirCarpeta({
        directory: true,
        title: "Elegí el repo local de ArnesIA (arnesia-app)",
      })
    } catch (e: unknown) {
      if (vivo.current) setRepoConfigError(e instanceof Error ? e.message : String(e))
      return
    }
    if (!seleccion || Array.isArray(seleccion)) return // cancelado (picker de una sola carpeta).
    setConfigurandoRepo(true)
    try {
      await api.configurarRepo(seleccion)
      const v = await api.getVersion<VersionInfo>()
      if (vivo.current) setVersion(v)
    } catch (e: unknown) {
      if (vivo.current) setRepoConfigError(e instanceof Error ? e.message : String(e))
    } finally {
      if (vivo.current) setConfigurandoRepo(false)
    }
  }, [])

  return (
    <div className="arnesia-ajustes flex h-full flex-col">
      <header className="border-b border-border bg-card px-[18px] py-3 text-sm text-muted-foreground">
        alpacapurpura / <b className="text-foreground">Ajustes</b>
      </header>
      <div className="flex flex-1 flex-col gap-4 overflow-auto p-[18px]">
        <UpdateCard
          version={version}
          versionError={versionError}
          estado={estado}
          reporte={reporte}
          mensaje={mensaje}
          onUpdate={onUpdate}
          onElegirRepo={isTauri() ? onElegirRepo : undefined}
          configurandoRepo={configurandoRepo}
          repoConfigError={repoConfigError}
        />
        {/* RF-100: lo que llega después, como texto muted — jamás tarjetas fingiendo. */}
        <p className="text-xs text-muted-foreground">
          más ajustes llegan: marketplaces por empresa/arnés · daemon (puerto / rutas / registro)
        </p>
      </div>
    </div>
  )
}

// GlobalView renders the daemon-wide destinations (rail foot). «Ajustes» es REAL desde
// el paquete boton-actualizar (RF-100); «Portafolio» es REAL desde Slice 1 (T7); «Estándar»
// sigue staged honesto.
export function GlobalView() {
  const route = useAppStore((s) => s.view)
  if (route === "ajustes") return <AjustesView />
  if (route === "portafolio") return <PortafolioView />
  const g = GLOBAL[route]
  if (!g) return <ComingSoon title="Vista" />
  return <ComingSoon glyph={g.glyph} title={g.title} note={g.note} />
}
