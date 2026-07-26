// Store del dictado por voz (Zustand). Paquete 2026-07-25-spike-voz-dictado, RF-216..230.
//
// Modela UNA máquina de estados chica y explícita, porque el peor modo de falla que encontró
// el spike fue una promesa que nunca resuelve (§1.6 b: `getUserMedia` se cuelga mudo). Cada
// etapa tiene nombre, cada etapa puede salir, y de ninguna se sale en silencio.
//
//   inactivo → escuchando → transcribiendo → ordenando → inactivo
//                    ↓            ↓             ↓
//                 (cancel)     (error)     (→ crudo, que NO es error)
//
// El audio vive en memoria y se descarta (V-D5): acá no hay URL.createObjectURL, ni
// reproductor, ni "volver a escuchar". Si algún día se quiere, es una decisión nueva.
//
// **La captura NO usa `MediaRecorder`** (RF-229, cambio contra la v0.2.20). El recorder de
// WebKitGTK 2.52.3 dice soportar `audio/mp4`, arranca, y entrega 0 bytes sin emitir `error`:
// el porqué —con el log de GStreamer que lo prueba— está en `shared/lib/wav.ts`. Acá se
// capturan muestras crudas por WebAudio y el WAV se arma en JS.

import { create } from "zustand"
import { ApiError, api, type DisponibilidadDictado, type EstadoDictado } from "@/shared/api"
import { reportar } from "@/shared/lib/diagnostico"
import { aWav, concatenar, HZ_STT, MIME_WAV, pico as picoDe, remuestrear } from "@/shared/lib/wav"

// TOPE_MS es el tope duro de grabación (V-D3 FIRMADA: toggle + tope). Default 3 min: el
// detonante del spike es que el operador habla largo, pero un mic olvidado abierto no puede
// grabar la tarde entera.
export const TOPE_MS = 3 * 60 * 1000

// AVISO_MS es cuánto antes del tope el contador se pone en warn. Que el corte no sorprenda.
export const AVISO_MS = 20 * 1000

// MIME es el formato que sube al daemon: WAV armado por nosotros. El adaptador STT lo acepta
// y, por ser `.wav`, se saltea el `ffmpeg` que en esta máquina NO está instalado.
export const MIME = MIME_WAV

// TAM_BLOQUE es el buffer del capturador. 4096 muestras ≈ 93 ms a 44.1 kHz: chico para que
// cortar responda al toque, grande para no despertar el hilo de audio cada 2 ms.
const TAM_BLOQUE = 4096

// UMBRAL_SILENCIO es el piso bajo el cual la grabación se considera muda.
//
// Es silencio DIGITAL, no "hablaste bajito": 0.001 sobre una escala 0..1 es lo que entrega un
// micrófono muteado, tomado por otra app o ruteado a un device muerto. Medido en el motor
// real, hablar normal da pico ≈ 0.08. Cortar acá evita mandar 3 minutos de ceros a un STT que
// va a devolver vacío y a hacer parecer que el bug es de la transcripción.
const UMBRAL_SILENCIO = 0.001

// Etapa is the named stage of a dictation. `inactivo` es el reposo.
export type Etapa = "inactivo" | "escuchando" | "transcribiendo" | "ordenando"

// EtapaLabel is what the UI writes. Ningún spinner anónimo (RF-228).
export const ETAPA_LABEL: Record<Exclude<Etapa, "inactivo">, string> = {
  escuchando: "Escuchando…",
  transcribiendo: "Transcribiendo…",
  ordenando: "Ordenando lo dictado…",
}

// Resultado is what a finished dictation hands to the composer.
export interface Resultado {
  texto: string
  estado: EstadoDictado
  motivo?: string | undefined
}

// Fallo describes a dictation that did not produce text: QUÉ etapa falló y por qué. El
// composer NO se toca cuando esto llega (RF-227).
export interface Fallo {
  etapa: Exclude<Etapa, "inactivo">
  mensaje: string
}

// Captura son las cifras de una grabación. No es estado de UI: es lo que se manda al log
// cuando algo falla, y es la diferencia entre "no se grabó nada" y saber POR QUÉ no se grabó.
interface Captura {
  bloques: number
  muestras: number
  segundos: number
  pico: number
  hzEntrada: number
  hzSalida: number
  bytes: number
}

interface DictadoState {
  etapa: Etapa
  // ms grabados; alimenta el contador 0:24 / 3:00.
  transcurrido: number
  // nivel 0..1 de entrada del micrófono. Un mic mudo se ve igual que uno vivo si lo único
  // que hay es un punto rojo.
  nivel: number
  // disponibilidad === undefined ⇒ todavía no se consultó (el botón no se muestra habilitado
  // hasta saber). El motivo se pinta en la superficie, no solo en un title.
  disponibilidad: DisponibilidadDictado | undefined
  // cortadoPorTope marca que la grabación paró sola: el flujo sigue con lo grabado y hay
  // que avisarlo (RF-218).
  cortadoPorTope: boolean
  fallo: Fallo | undefined

  consultarDisponibilidad: () => Promise<void>
  // empezar arranca la grabación. Devuelve false si no se pudo (y deja `fallo` puesto).
  empezar: (sesionId: string, onTexto: (r: Resultado) => void) => Promise<boolean>
  // cortar detiene la grabación y sigue con transcribir+ordenar (RF-217, el toggle).
  cortar: () => void
  // cancelar descarta todo sin transcribir: el composer queda EXACTAMENTE como estaba
  // (RF-219). También aborta un transcribir/ordenar ya en vuelo.
  cancelar: () => void
  // usarCrudo abandona la espera de la limpieza y se queda con el transcripto (V-D4 en vivo).
  usarCrudo: () => void
  limpiarFallo: () => void
}

// recursos are the live handles of one dictation. Fuera del store a propósito: son objetos
// mutables del navegador, no estado renderizable, y meterlos en Zustand solo dispararía
// renders que a nadie le sirven.
interface Recursos {
  stream?: MediaStream | undefined
  ctx?: AudioContext | undefined
  fuente?: MediaStreamAudioSourceNode | undefined
  proc?: ScriptProcessorNode | undefined
  raf?: number | undefined
  tick?: ReturnType<typeof setInterval> | undefined
  tope?: ReturnType<typeof setTimeout> | undefined
  abort?: AbortController | undefined
  // cancelado corta el flujo después de un await: sin esto, un cancel durante el POST
  // igual terminaría poblando el composer.
  cancelado: boolean
  // bloques es el PCM crudo tal como lo entrega el hilo de audio.
  bloques: Float32Array[]
  hzEntrada: number
  // subir es el paso siguiente al corte. Vive acá porque `cortar` y el tope disparan desde
  // fuera de `empezar` y necesitan el `onTexto` de ESTE dictado.
  subir?: ((audio: Blob, cap: Captura) => void) | undefined
}

let R: Recursos = { cancelado: false, bloques: [], hzEntrada: 0 }

// soltar libera TODO: tracks del micrófono, timers, audio graph y el fetch en vuelo.
//
// Que el mic se libere de verdad es parte de RF-217 — un track vivo deja el indicador del
// sistema encendido y le dice al operador que lo seguimos escuchando cuando no.
function soltar() {
  if (R.proc) {
    R.proc.onaudioprocess = null
    R.proc.disconnect()
  }
  R.proc = undefined
  R.fuente?.disconnect()
  R.fuente = undefined
  R.stream?.getTracks().forEach((t) => {
    t.stop()
  })
  R.stream = undefined
  void R.ctx?.close().catch(() => {})
  R.ctx = undefined
  if (R.raf !== undefined) cancelAnimationFrame(R.raf)
  if (R.tick) clearInterval(R.tick)
  if (R.tope) clearTimeout(R.tope)
  R.raf = undefined
  R.tick = undefined
  R.tope = undefined
}

// cerrarCaptura desconecta el capturador y arma el WAV con lo que se juntó.
//
// Corre ANTES de `soltar` porque necesita el `AudioContext` vivo para saber a qué tasa se
// muestreó: cerrarlo primero perdería el dato justo cuando hay que remuestrear.
function cerrarCaptura(): { audio: Blob; cap: Captura } {
  if (R.proc) {
    R.proc.onaudioprocess = null
    R.proc.disconnect()
  }
  R.fuente?.disconnect()

  const bloques = R.bloques.length
  const crudas = concatenar(R.bloques)
  R.bloques = []

  const hzEntrada = R.hzEntrada
  const muestras = remuestrear(crudas, hzEntrada, HZ_STT)
  const audio = aWav(muestras, HZ_STT)
  return {
    audio,
    cap: {
      bloques,
      muestras: crudas.length,
      segundos: hzEntrada > 0 ? Number((crudas.length / hzEntrada).toFixed(2)) : 0,
      pico: Number(picoDe(crudas).toFixed(4)),
      hzEntrada,
      hzSalida: HZ_STT,
      bytes: audio.size,
    },
  }
}

// motivoDeGetUserMedia traduce el error del navegador a una frase accionable. Los tres casos
// se arreglan de maneras distintas (permiso · enchufar un mic · el puente Rust), así que
// decir lo mismo en los tres sería inútil.
function motivoDeGetUserMedia(err: unknown): string {
  const name = err instanceof Error ? err.name : ""
  if (name === "NotAllowedError" || name === "SecurityError") {
    return "Micrófono denegado. Habilitalo en el sistema y reabrí la app."
  }
  if (name === "NotFoundError" || name === "OverconstrainedError") {
    return "No hay ningún micrófono conectado."
  }
  return `No pude abrir el micrófono: ${err instanceof Error ? err.message : String(err)}`
}

export const useDictado = create<DictadoState>((set, get) => {
  // fallar deja el estado visible Y manda el detalle al log del daemon en el mismo acto
  // (RF-230). Están juntos a propósito: un fallo que el operador ve pero que no queda escrito
  // en ningún lado es exactamente lo que costó una versión entera de dictado roto.
  const fallar = (
    etapa: Exclude<Etapa, "inactivo">,
    mensaje: string,
    evento: string,
    detalle: Record<string, unknown> = {},
  ) => {
    set({ etapa: "inactivo", transcurrido: 0, nivel: 0, fallo: { etapa, mensaje } })
    reportar(evento, { etapa, ...detalle }, mensaje)
  }

  return {
    etapa: "inactivo",
    transcurrido: 0,
    nivel: 0,
    disponibilidad: undefined,
    cortadoPorTope: false,
    fallo: undefined,

    consultarDisponibilidad: async () => {
      try {
        set({ disponibilidad: await api.disponibilidadDictado() })
      } catch (err) {
        // Que la sonda falle NO es un no-disponible silencioso: se dice, con el motivo.
        const motivo = `No pude consultar si el dictado está disponible: ${
          err instanceof Error ? err.message : String(err)
        }`
        set({ disponibilidad: { disponible: false, motivo } })
        reportar("dictado.disponibilidad-ilegible", {}, motivo)
      }
    },

    empezar: async (sesionId, onTexto) => {
      if (get().etapa !== "inactivo") return false
      if (typeof AudioContext === "undefined") {
        fallar(
          "escuchando",
          "Este motor no puede capturar audio (no tiene AudioContext).",
          "dictado.sin-audiocontext",
        )
        return false
      }

      R = { cancelado: false, bloques: [], hzEntrada: 0 }
      set({ fallo: undefined, cortadoPorTope: false, transcurrido: 0, nivel: 0 })

      let stream: MediaStream
      try {
        // Sin el puente Rust (RF-215) esta promesa NO rechaza en la app instalada: queda
        // pendiente para siempre. El puente es lo que la hace resolver o fallar.
        stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      } catch (err) {
        fallar("escuchando", motivoDeGetUserMedia(err), "dictado.microfono-denegado", {
          error: err instanceof Error ? err.name : String(err),
        })
        return false
      }
      if (R.cancelado) {
        stream.getTracks().forEach((t) => {
          t.stop()
        })
        return false
      }
      R.stream = stream

      // subir corre transcribir+ordenar en el daemon y entrega el resultado.
      const subir = async (audio: Blob, cap: Captura) => {
        R.abort = new AbortController()
        set({ etapa: "transcribiendo", nivel: 0 })
        // La limpieza es el 87 % del tiempo (13-17 s contra 2.2 s del STT), así que se
        // anuncia como etapa propia en vez de esconderse en "procesando". El daemon hace las
        // dos en un POST; el FE cambia el rótulo cuando el STT ya debería haber terminado.
        const aOrdenar = setTimeout(() => {
          if (get().etapa === "transcribiendo") set({ etapa: "ordenando" })
        }, 3000)
        try {
          const d = await api.dictado(sesionId, audio, R.abort.signal)
          if (R.cancelado) return
          onTexto({ texto: d.texto, estado: d.estado, motivo: d.motivo })
          set({ etapa: "inactivo", transcurrido: 0 })
        } catch (err) {
          if (R.cancelado) return
          // Nombrar la etapa que falló es el requisito, no un lujo (RF-228).
          const etapa = get().etapa === "ordenando" ? "ordenando" : "transcribiendo"
          fallar(
            etapa,
            err instanceof Error ? err.message : String(err),
            "dictado.daemon-rechazo",
            { ...cap, mime: MIME, status: err instanceof ApiError ? err.status : undefined },
          )
        } finally {
          clearTimeout(aOrdenar)
          R.abort = undefined
        }
      }
      R.subir = (audio, cap) => {
        void subir(audio, cap)
      }

      // Grafo de audio. El capturador y el medidor cuelgan de la MISMA fuente: son dos
      // lecturas del mismo micrófono, no dos capturas.
      try {
        const ctx = new AudioContext()
        R.ctx = ctx
        R.hzEntrada = ctx.sampleRate
        // Un contexto suspendido entrega CERO bloques — o sea, se vería exactamente igual que
        // el bug que este RF vino a arreglar. Se lo despierta siempre; si ya estaba corriendo
        // es un no-op, y el `catch` es porque fallar acá no debe tumbar la captura.
        void ctx.resume().catch(() => {})
        const fuente = ctx.createMediaStreamSource(stream)
        R.fuente = fuente

        // ScriptProcessor está deprecado y se usa igual, a conciencia: AudioWorklet necesita
        // cargar un módulo por URL, y en el WebView la app se sirve bajo CSP desde el daemon
        // —un blob: de worklet es justo lo que esa política bloquea—. El deprecado ANDA en
        // el motor real (medido: 37 bloques en 3.44 s); el moderno habría que probarlo antes
        // de confiarle la única vía de captura que nos queda.
        const proc = ctx.createScriptProcessor(TAM_BLOQUE, 1, 1)
        R.proc = proc
        proc.onaudioprocess = (e) => {
          // Se COPIA: el buffer de entrada lo reusa el hilo de audio en el bloque siguiente.
          R.bloques.push(new Float32Array(e.inputBuffer.getChannelData(0)))
        }
        fuente.connect(proc)
        // El nodo tiene que llegar al destino para que el hilo de audio lo corra. No hay eco:
        // nunca se escribe en `outputBuffer`, así que lo que sale es silencio.
        proc.connect(ctx.destination)

        // Medidor de nivel: la señal de que REALMENTE está entrando audio.
        const analizador = ctx.createAnalyser()
        analizador.fftSize = 512
        fuente.connect(analizador)
        const buf = new Uint8Array(analizador.frequencyBinCount)
        const medir = () => {
          analizador.getByteTimeDomainData(buf)
          let p = 0
          for (const v of buf) p = Math.max(p, Math.abs(v - 128))
          set({ nivel: Math.min(1, p / 96) })
          R.raf = requestAnimationFrame(medir)
        }
        R.raf = requestAnimationFrame(medir)
      } catch (err) {
        // Antes esto era cosmético (solo el medidor). Ahora el grafo ES la captura: si no se
        // arma, no hay dictado, y decirlo acá es mejor que un WAV vacío tres minutos después.
        soltar()
        fallar(
          "escuchando",
          `No pude armar la captura de audio: ${err instanceof Error ? err.message : String(err)}`,
          "dictado.grafo-roto",
        )
        return false
      }

      set({ etapa: "escuchando" })

      // Contador + tope duro (RF-218): para sola, CONSERVA lo grabado y avisa.
      const desde = Date.now()
      R.tick = setInterval(() => {
        set({ transcurrido: Date.now() - desde })
      }, 250)
      R.tope = setTimeout(() => {
        if (get().etapa !== "escuchando") return
        set({ cortadoPorTope: true })
        get().cortar()
      }, TOPE_MS)

      return true
    },

    cortar: () => {
      if (get().etapa !== "escuchando") return

      const { audio, cap } = cerrarCaptura()
      soltar()
      if (R.cancelado) {
        set({ etapa: "inactivo", transcurrido: 0, nivel: 0 })
        return
      }

      // Los dos modos de "no hay audio" se distinguen porque se arreglan distinto: sin
      // bloques está roto el grafo o el device; con bloques y pico 0 el mic está mudo.
      if (cap.bloques === 0 || cap.muestras === 0) {
        fallar(
          "escuchando",
          "El micrófono no entregó ni un bloque de audio. Reabrí la app y, si sigue, revisá qué otra app tiene tomado el micrófono.",
          "dictado.sin-bloques",
          { ...cap, mime: MIME },
        )
        return
      }
      if (cap.pico < UMBRAL_SILENCIO) {
        fallar(
          "escuchando",
          `El micrófono entregó ${cap.segundos}s de silencio (nivel 0). Revisá que no esté muteado ni tomado por otra app.`,
          "dictado.silencio",
          { ...cap, mime: MIME },
        )
        return
      }

      R.subir?.(audio, cap)
    },

    cancelar: () => {
      R.cancelado = true
      R.abort?.abort()
      R.bloques = []
      soltar()
      set({ etapa: "inactivo", transcurrido: 0, nivel: 0, cortadoPorTope: false })
    },

    usarCrudo: () => {
      if (get().etapa !== "ordenando") return
      // `cancelado` acá NO significa «el operador canceló el dictado»: significa «no apliques
      // el desenlace del POST», para que el catch del abort no pise el mensaje de abajo con un
      // "AbortError" genérico. El estado que ve el operador lo fijamos nosotros.
      R.cancelado = true
      // El daemon ya está esperando a la limpieza; abortar el POST devuelve el control al
      // operador sin dejar la UI colgada. El crudo se pierde a propósito: pedirlo de nuevo
      // cuesta 2.2 s, y mentir sobre qué texto es cuál costaría más.
      R.abort?.abort()
      set({
        etapa: "inactivo",
        transcurrido: 0,
        fallo: {
          etapa: "ordenando",
          mensaje: "Cortaste la limpieza. Volvé a dictar si querés el texto sin ordenar.",
        },
      })
    },

    limpiarFallo: () => {
      set({ fallo: undefined, cortadoPorTope: false })
    },
  }
})

// mmss formats ms as 0:07 / 2:47. Exportado para que la story lo use igual que el componente.
export function mmss(ms: number): string {
  const s = Math.floor(ms / 1000)
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`
}
