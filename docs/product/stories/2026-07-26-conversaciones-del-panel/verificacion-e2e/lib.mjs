// lib.mjs — plomería común de los guiones E2E contra el binario INSTALADO.
//
// Por qué el import por ruta resuelta y no `from "playwright"`: este archivo NO vive dentro
// de `web/`, así que la resolución de Node no encuentra el paquete. El precedente
// (`stories/2026-07-08-chat-cc-funcional/e2e/casuistica.mjs:6`) lo resolvía con una ruta
// ABSOLUTA hardcodeada del home del autor — que se rompe en cualquier otra máquina y en
// cualquier worktree. Acá se resuelve relativo AL ARCHIVO (plan-pruebas.md §3.2).
import { fileURLToPath } from "node:url"
import { dirname, resolve } from "node:path"

const aqui = dirname(fileURLToPath(import.meta.url))
export const RAIZ = resolve(aqui, "../../../../..")

const { chromium } = await import(resolve(RAIZ, "web/node_modules/playwright/index.mjs"))

export const BASE = process.env.E2E_BASE ?? "http://127.0.0.1:4200"
export const CAPTURAS = aqui

// ── el marcador de resultados ────────────────────────────────────────────────
// Un guion no "pasa" porque no tiró excepción: cada paso declara su aserción y su
// resultado. Un paso que no se pudo correr se marca `n/c` (no corrido) con su motivo —
// jamás se omite, porque un renglón ausente se lee como un verde.
export const pasos = []
export function ok(guion, paso, asercion, detalle = "") {
  pasos.push({ guion, paso, asercion, r: "ok", detalle })
  console.log(`  ✅ ${guion}.${paso}  ${asercion}${detalle ? "  ·  " + detalle : ""}`)
}
export function fail(guion, paso, asercion, detalle = "") {
  pasos.push({ guion, paso, asercion, r: "FALLA", detalle })
  console.log(`  ❌ ${guion}.${paso}  ${asercion}  ·  ${detalle}`)
}
export function nc(guion, paso, asercion, motivo) {
  pasos.push({ guion, paso, asercion, r: "n/c", detalle: motivo })
  console.log(`  ⏸  ${guion}.${paso}  ${asercion}  ·  NO CORRIDO: ${motivo}`)
}
export function chequear(cond, guion, paso, asercion, detalle = "") {
  cond ? ok(guion, paso, asercion, detalle) : fail(guion, paso, asercion, detalle)
  return cond
}

// ── navegador ────────────────────────────────────────────────────────────────
// `waitUntil: "load"`, NUNCA `networkidle`: el SPA deja el SSE abierto para siempre y
// `networkidle` no resuelve jamás (timeout de 30 s). Verificado en este mismo circuito.
export async function abrir({ headless = true } = {}) {
  const browser = await chromium.launch({ headless })
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } })
  const page = await ctx.newPage()
  const consola = []
  page.on("console", m => { if (m.type() === "error") consola.push(m.text()) })
  page.on("pageerror", e => consola.push("PAGEERROR " + e.message))
  await page.goto(BASE, { waitUntil: "load" })
  await page.waitForTimeout(2000)
  return { browser, ctx, page, consola }
}

// abrirBuscador deja el panel de la lista abierto CON su buscador a la vista.
// El 🔍 ALTERNA el panel: si la lista ya estaba abierta por el ▶, un clic la cierra. Se
// reintenta en vez de asumir el estado, porque el guion no controla en qué estado la dejó
// el paso anterior. Devuelve el locator del input.
export async function abrirBuscador(page, intentos = 3) {
  const input = page.getByLabel("Buscar en estas conversaciones")
  for (let i = 0; i < intentos; i++) {
    if (await input.count() > 0) return input
    await page.getByTitle("Buscar en las conversaciones de esta sesión").click()
    await page.waitForTimeout(700)
  }
  return input
}

export async function foto(page, nombre) {
  await page.screenshot({ path: resolve(CAPTURAS, `${nombre}.png`) })
  console.log(`     📷 ${nombre}.png`)
}

// ── API ──────────────────────────────────────────────────────────────────────
export async function api(ruta, init) {
  const r = await fetch(BASE + ruta, init)
  const txt = await r.text()
  let json = null
  try { json = JSON.parse(txt) } catch { /* cuerpo no-JSON: se reporta crudo */ }
  return { status: r.status, json, txt }
}

export function resumen(nombreArchivo) {
  const n = { ok: 0, FALLA: 0, "n/c": 0 }
  for (const p of pasos) n[p.r]++
  console.log(`\n── ${nombreArchivo}: ${n.ok} ok · ${n.FALLA} FALLA · ${n["n/c"]} n/c`)
  return { n, pasos }
}
