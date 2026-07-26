// Diagnóstico de fallos del FE (paquete 2026-07-25-spike-voz-dictado, RF-230).
//
// **El agujero que tapa.** La app instalada corre en un WebView sin devtools: `console.error`
// no va a ningún lado, el daemon loguea a stderr y el `.desktop` del `.deb` tira ese stderr a
// la basura. Resultado real, medido: el `MediaRecorder` de WebKitGTK entregó 0 bytes durante
// TODA la v0.2.20 y lo único que el operador podía reportar era la frase de la superficie
// («No se grabó nada»). Reproducirlo costó levantar un WebView aparte con `GST_DEBUG`.
//
// Con esto, el mismo fallo llega al log con las variables que lo explican (bloques, muestras,
// pico, hz, mime, bytes) y se arregla leyendo un archivo.
//
// **Lo que NO hace:** no manda nada a Internet. El destino es el daemon en loopback —el mismo
// origin que ya sirve la app— y de ahí a un archivo local. La telemetría de producto (OTel,
// HS-27) es otro paquete y otra decisión; esto es el log de la casa.

import { api } from "@/shared/api"
import { isTauri } from "./platform"

// topePorMinuto corta una tormenta de eventos.
//
// Un `unhandledrejection` dentro de un render que re-renderiza puede disparar cientos por
// segundo; sin tope, el diagnóstico dejaría de ser una ayuda para pasar a ser el incidente.
const topePorMinuto = 30

const ventanaMs = 60_000

let emitidos: number[] = []

// reportar manda un fallo al log del daemon. Nunca lanza y nunca bloquea a quien la llama:
// un diagnóstico que rompe el flujo que estaba diagnosticando no sirve.
export function reportar(
  evento: string,
  detalle: Record<string, unknown> = {},
  mensaje?: string,
): void {
  const ahora = Date.now()
  emitidos = emitidos.filter((t) => ahora - t < ventanaMs)
  if (emitidos.length >= topePorMinuto) return
  emitidos.push(ahora)

  void api
    .diagnostico({
      origen: "fe",
      evento,
      ...(mensaje ? { mensaje } : {}),
      detalle: { ...contexto(), ...detalle },
    })
    // El catch es el punto: si el POST falla, el fallo original ya se le mostró al operador
    // en la superficie. Insistir acá solo agregaría un segundo error al primero.
    .catch(() => {})
}

// contexto son los datos del entorno que casi siempre importan.
//
// `ua` es el que más rinde: trae la versión de WebKitGTK, que es LA variable de los bugs de
// motor que ya nos comimos (`SpeechRecognition` ausente, `MediaRecorder` mudo).
function contexto(): Record<string, unknown> {
  return {
    ua: typeof navigator === "undefined" ? "" : navigator.userAgent,
    tauri: isTauri(),
    ruta: typeof location === "undefined" ? "" : location.hash,
  }
}

// instalarCazadorDeErrores engancha los dos escapes globales del navegador.
//
// Se llama UNA vez en el arranque. Cubre lo que ningún try/catch nuestro ve: una excepción de
// render que el ErrorBoundary pinta pero no persiste, y una promesa rechazada sin dueño.
export function instalarCazadorDeErrores(): void {
  if (typeof window === "undefined") return

  window.addEventListener("error", (e) => {
    reportar(
      "fe.error",
      {
        archivo: e.filename,
        linea: e.lineno,
        columna: e.colno,
        stack: e.error instanceof Error ? e.error.stack : undefined,
      },
      e.message,
    )
  })

  window.addEventListener("unhandledrejection", (e) => {
    const r: unknown = e.reason
    reportar(
      "fe.promesa-rechazada",
      { stack: r instanceof Error ? r.stack : undefined },
      r instanceof Error ? r.message : String(r),
    )
  })
}
