// Codificador WAV mínimo para el dictado (paquete 2026-07-25-spike-voz-dictado, RF-229).
//
// **Por qué existe** — el spike probó que `MediaRecorder` EXISTE en WebKitGTK 2.52.3 y que
// `isTypeSupported('audio/mp4')` devuelve `true`. Lo que NO probó, y resultó falso, es que
// grabara: contra el motor real el recorder arranca (`state: recording`, `onstart` disparado)
// y entrega **un `dataavailable` de 0 bytes**. El log de GStreamer dice por qué:
//
//     MediaRecorderPrivateGStreamer.cpp:412: Setting audio restriction caps to
//                                            audio/x-raw, rate=(int)0
//     gstencodebasebin.c:718: <encodebin2-0> Couldn't find a compatible stream profile
//     gstbasesrc.c:3177: <capture-audiosrc0> error: streaming stopped, reason not-linked (-1)
//     MediaRecorderPrivateGStreamer.cpp:272: Transfering 0 encoded bytes
//
// WebKit arma el perfil de encoding con `rate=0`, `encodebin` no casa ningún profile, el pad
// de audio nunca se linkea y el micrófono empuja a la nada. El pipeline SÍ postea `error` en
// el bus de GStreamer, pero WebKit no lo propaga a JS: `MediaRecorder.onerror` jamás dispara.
// Falla muda del motor, no del código. Probado también: `start(timeslice)` y
// `audioBitsPerSecond` no la mueven — 0 bytes en los tres casos.
//
// La salida medida en el mismo motor: WebAudio entrega PCM sin pasar por `encodebin`
// (3.44 s → 151552 muestras, pico 0.0787). Así que el dictado captura muestras crudas y
// arma el WAV acá, en JS.
//
// **Bonus que no es accidental:** el WAV a 16 kHz es exactamente lo que come `whisper.cpp`,
// así que el transcodificado `mp4 → wav` desaparece y con él la dependencia de `ffmpeg`
// —que el spike encontró AUSENTE en esta máquina (`spike-spec.md` §1.7)—.

// MIME_WAV is what the composer ships to the daemon. El adaptador STT ya lo acepta y, por ser
// `.wav`, se saltea el paso de `ffmpeg`.
export const MIME_WAV = "audio/wav"

// HZ_STT es la tasa a la que se remuestrea antes de subir. Whisper trabaja en 16 kHz mono:
// mandar los 44.1 kHz del `AudioContext` solo triplicaría los bytes para que el motor los
// tire igual.
export const HZ_STT = 16000

// concatenar junta los bloques que fue dejando la captura en un solo buffer.
export function concatenar(bloques: readonly Float32Array[]): Float32Array {
  let total = 0
  for (const b of bloques) total += b.length
  const out = new Float32Array(total)
  let off = 0
  for (const b of bloques) {
    out.set(b, off)
    off += b.length
  }
  return out
}

// pico devuelve la amplitud máxima absoluta (0..1). Es la señal de «entró audio de verdad»:
// un micrófono muteado o tomado por otra app entrega muestras, pero todas en cero.
export function pico(muestras: Float32Array): number {
  let max = 0
  for (const v of muestras) {
    const a = Math.abs(v)
    if (a > max) max = a
  }
  return max
}

// remuestrear baja la tasa de muestreo por interpolación lineal.
//
// Lineal y no sinc a propósito: la entrada es voz que va a un STT que internamente trabaja en
// 16 kHz, y el aliasing que agrega la interpolación queda muy por debajo de los errores que
// el propio motor comete (medidos en §1.8: «locita»→lógica, «demon»→daemon). Un resampler
// bueno costaría código y CPU para mejorar algo que la limpieza con contexto ya repara.
export function remuestrear(
  muestras: Float32Array,
  hzEntrada: number,
  hzSalida: number,
): Float32Array {
  if (muestras.length === 0 || hzEntrada <= 0 || hzSalida <= 0 || hzEntrada === hzSalida) {
    return muestras
  }
  const ratio = hzEntrada / hzSalida
  const n = Math.max(1, Math.floor(muestras.length / ratio))
  const out = new Float32Array(n)
  for (let i = 0; i < n; i++) {
    const pos = i * ratio
    const izq = Math.floor(pos)
    const der = Math.min(izq + 1, muestras.length - 1)
    const t = pos - izq
    out[i] = (muestras[izq] ?? 0) * (1 - t) + (muestras[der] ?? 0) * t
  }
  return out
}

// aWav envuelve las muestras en un WAV PCM 16-bit mono.
//
// El header es el canónico de 44 bytes; se escribe a mano porque la alternativa sería sumar
// una dependencia para 20 líneas de `DataView`.
export function aWav(muestras: Float32Array, hz: number): Blob {
  const bytesDeDatos = muestras.length * 2
  const buf = new ArrayBuffer(44 + bytesDeDatos)
  const v = new DataView(buf)

  ascii(v, 0, "RIFF")
  v.setUint32(4, 36 + bytesDeDatos, true) // tamaño del archivo menos los 8 primeros bytes
  ascii(v, 8, "WAVE")

  ascii(v, 12, "fmt ")
  v.setUint32(16, 16, true) // largo del bloque fmt (PCM = 16)
  v.setUint16(20, 1, true) // formato 1 = PCM entero sin comprimir
  v.setUint16(22, 1, true) // canales: mono
  v.setUint32(24, hz, true)
  v.setUint32(28, hz * 2, true) // byte rate = hz × canales × bytes por muestra
  v.setUint16(32, 2, true) // block align = canales × bytes por muestra
  v.setUint16(34, 16, true) // bits por muestra

  ascii(v, 36, "data")
  v.setUint32(40, bytesDeDatos, true)
  for (let i = 0; i < muestras.length; i++) {
    // Se satura en vez de envolver: un desborde silencioso sonaría a chasquido y el STT lo
    // leería como ruido.
    const s = Math.max(-1, Math.min(1, muestras[i] ?? 0))
    v.setInt16(44 + i * 2, s < 0 ? s * 0x8000 : s * 0x7fff, true)
  }
  return new Blob([buf], { type: MIME_WAV })
}

function ascii(v: DataView, offset: number, s: string): void {
  for (let i = 0; i < s.length; i++) v.setUint8(offset + i, s.charCodeAt(i))
}
