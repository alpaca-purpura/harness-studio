// Tests del encoder WAV (RF-229). Existen porque este código reemplaza a `MediaRecorder`:
// lo que antes hacía (mal) el motor ahora lo hacemos nosotros, así que ahora se puede probar.

import { describe, expect, it } from "vitest"
import { aWav, concatenar, HZ_STT, MIME_WAV, pico, remuestrear } from "./wav"

const texto = (b: ArrayBuffer, off: number, n: number) =>
  String.fromCharCode(...new Uint8Array(b, off, n))

async function bytes(blob: Blob): Promise<ArrayBuffer> {
  return blob.arrayBuffer()
}

describe("concatenar", () => {
  it("junta los bloques en orden", () => {
    const out = concatenar([new Float32Array([1, 2]), new Float32Array([3])])
    expect(Array.from(out)).toEqual([1, 2, 3])
  })

  it("sin bloques devuelve vacío en vez de romper", () => {
    expect(concatenar([]).length).toBe(0)
  })
})

describe("pico", () => {
  it("toma la amplitud máxima absoluta", () => {
    expect(pico(new Float32Array([0.1, -0.9, 0.3]))).toBeCloseTo(0.9)
  })

  // El caso que motivó la función: un micrófono muteado entrega muestras, todas en cero.
  it("silencio digital da 0", () => {
    expect(pico(new Float32Array(1024))).toBe(0)
  })
})

describe("remuestrear", () => {
  it("baja 44100 → 16000 con la proporción correcta", () => {
    const entrada = new Float32Array(44100) // 1 segundo
    const salida = remuestrear(entrada, 44100, HZ_STT)
    expect(salida.length).toBe(16000)
  })

  it("misma tasa no toca las muestras", () => {
    const entrada = new Float32Array([0.5, -0.5])
    expect(remuestrear(entrada, 16000, 16000)).toBe(entrada)
  })

  it("interpola en vez de tirar muestras", () => {
    // 4 muestras a 4 Hz → 2 muestras a 2 Hz: la segunda cae entre 0 y 1, no en un borde.
    const salida = remuestrear(new Float32Array([0, 1, 0, -1]), 4, 2)
    expect(salida.length).toBe(2)
    expect(salida[0]).toBeCloseTo(0)
    expect(salida[1]).toBeCloseTo(0)
  })
})

describe("aWav", () => {
  it("escribe el header canónico RIFF/WAVE", async () => {
    const b = await bytes(aWav(new Float32Array([0, 0]), HZ_STT))
    const v = new DataView(b)
    expect(texto(b, 0, 4)).toBe("RIFF")
    expect(texto(b, 8, 4)).toBe("WAVE")
    expect(texto(b, 12, 4)).toBe("fmt ")
    expect(texto(b, 36, 4)).toBe("data")
    expect(v.getUint16(20, true)).toBe(1) // PCM entero
    expect(v.getUint16(22, true)).toBe(1) // mono
    expect(v.getUint32(24, true)).toBe(HZ_STT)
    expect(v.getUint16(34, true)).toBe(16) // bits por muestra
  })

  it("declara los tamaños coherentes con el cuerpo", async () => {
    const muestras = new Float32Array(100)
    const b = await bytes(aWav(muestras, HZ_STT))
    const v = new DataView(b)
    expect(b.byteLength).toBe(44 + 200)
    expect(v.getUint32(4, true)).toBe(36 + 200)
    expect(v.getUint32(40, true)).toBe(200)
  })

  it("satura en vez de envolver: +1.5 no da un valor negativo", async () => {
    const b = await bytes(aWav(new Float32Array([1.5, -1.5]), HZ_STT))
    const v = new DataView(b)
    expect(v.getInt16(44, true)).toBe(32767)
    expect(v.getInt16(46, true)).toBe(-32768)
  })

  it("el blob lleva el mime que el daemon espera", () => {
    expect(aWav(new Float32Array(1), HZ_STT).type).toBe(MIME_WAV)
  })

  // La regresión que importa: el bug de v0.2.20 fue un blob de 0 bytes. Un WAV con muestras
  // NUNCA puede quedar en 0, y si alguien rompe eso el test lo dice.
  it("con muestras nunca sale vacío", () => {
    expect(aWav(new Float32Array(16000), HZ_STT).size).toBeGreaterThan(16000)
  })
})
