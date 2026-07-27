// E2E-3 · La rotación llega EN VIVO (E-18, E-19 · cierra C-6/H-8)
//
// Es el guion que el plan declaraba «hoy NO realizable» porque `session_rotacion.go` no
// tenía un solo `s.publish`. Prueba que la marca de rotación aparece en el dock SIN
// RECARGAR, que la conversación NO se parte en dos, y que el `cc-id` cambia a la vista.
//
// Sobre `claude` real: el plan lo pedía para este guion. NO hace falta y por eso no se usa
// — la rotación la dispara el `ctx_pct`, que el daemon calcula del `usage` reportado
// (conductor.go:773). Un mock que reporta tokens ejercita EXACTAMENTE el mismo camino de
// `rotarLocked` + el `publish` que se quiere probar, y encima lo hace determinista. Lo que
// un modelo real agregaría acá es varianza, no cobertura.
import { writeFileSync } from "node:fs"
import { abrir, abrirBuscador, foto, api, ok, fail, nc, chequear, resumen } from "./lib.mjs"

const G = "E2E-3"
const CTL = process.env.MOCK_CTL
const MARCA = "— contexto rotado, seguimos —"
const { browser, page, consola } = await abrir()
const SID = (await api("/api/sessions")).json[0].id
const conv = async () => (await api(`/api/sessions/${SID}/conversaciones`)).json
const activa = async () => (await conv()).conversaciones.find(c => c.activa)

await page.getByRole("button", { name: /Conversar/ }).click()
await page.waitForTimeout(1200)
ok(G, "1", "el daemon corre con --rotacion-umbral 5", "flag pasado al arrancar (ver INFORME §circuito)")

// ── paso 2 · llegar al umbral: el chip se pinta CALIENTE ─────────────────────
// 30 000 tokens sobre la ventana por defecto ⇒ ctx_pct muy por encima de 5.
writeFileSync(CTL, "MOCK_TOKENS=30000\n")
await page.getByRole("textbox").last().fill("primer turno, a cargar contexto")
await page.keyboard.press("Enter")
await page.waitForTimeout(4000)
const a2 = await activa()
chequear(a2.rotacion_pendiente === true, G, "2a", "tras superar el umbral, `rotacion_pendiente` es true", `ctx_pct=${a2.ctx_pct} · pendiente=${a2.rotacion_pendiente}`)

// C-3 · la CONTRADICCIÓN más cara del paquete: se pinta caliente la BARRA, no el NÚMERO.
const chipEstilo = await page.evaluate(() => {
  const b = document.querySelector('button[aria-label*="contexto"]')
  if (!b) return null
  const barra = b.querySelector("[data-barra-ctx]")
  const num = [...b.childNodes].find(n => n.nodeType === 3 && /%/.test(n.textContent))
  return {
    barra: barra ? getComputedStyle(barra).backgroundColor : null,
    barraClase: barra ? barra.className : null,
    numero: getComputedStyle(b).color,
    warn: getComputedStyle(document.documentElement).getPropertyValue("--warn").trim(),
    texto: b.innerText.replace(/\s+/g, " ").trim(),
  }
})
chequear(chipEstilo !== null, G, "2b", "el chip de contexto está en el DOM", JSON.stringify(chipEstilo?.texto))
if (chipEstilo) {
  const barraCaliente = /warn/.test(chipEstilo.barraClase ?? "")
  const numeroCaliente = /warn/.test(chipEstilo.numero) || chipEstilo.numero === chipEstilo.barra
  chequear(barraCaliente, G, "2c", "C-3 · la BARRA se pinta caliente (--warn)", `clase="${chipEstilo.barraClase}"`)
  chequear(!numeroCaliente, G, "2d", "C-3 · el NÚMERO **no** se pinta caliente", `color=${chipEstilo.numero} · barra=${chipEstilo.barra}`)
}
await foto(page, "e2e3-01-chip-caliente")

// ── paso 3 · el turno siguiente ROTA, y la marca llega SIN RECARGAR ──────────
const antesN = (await conv()).conversaciones.length
const ccAntes = a2.claude_session_id
const tituloAntes = a2.titulo
writeFileSync(CTL, "MOCK_TOKENS=4000\n") // ~2 %: lo que un hilo fresco realmente gasta (system prompt + checkpoint)
// Se observa el DOM en vivo: si la marca sólo apareciera tras un reload, esto no la ve.
await page.evaluate(m => {
  window.__rot = 0
  new MutationObserver(() => { if (document.body.innerText.includes(m)) window.__rot++ })
    .observe(document.body, { childList: true, subtree: true, characterData: true })
}, MARCA)
await page.getByRole("textbox").last().fill("segundo turno, el que dispara la rotación")
await page.keyboard.press("Enter")
await page.waitForTimeout(6000)
const vistaEnVivo = await page.evaluate(() => window.__rot)
const enPantalla = (await page.locator("body").innerText()).includes(MARCA)
chequear(enPantalla, G, "3a", `la marca «${MARCA}» está en el transcript`, `enPantalla=${enPantalla}`)
chequear(vistaEnVivo > 0, G, "3b", "…y llegó EN VIVO por SSE, sin recargar (cierra C-6/H-8)", `mutaciones observadas=${vistaEnVivo}`)
// El elemento de la marca ES el que tiene el texto (chat-dock.tsx:308-313), no su padre:
// centrar es `align-self` dentro del flex del transcript, no `text-align` del contenedor.
const forma = await page.evaluate(m => {
  const el = [...document.querySelectorAll("*")].find(n => n.children.length === 0 && n.textContent.trim() === m)
  if (!el) return null
  const c = getComputedStyle(el)
  return { alignSelf: c.alignSelf, borde: c.borderTopStyle, radios: [c.borderBottomRightRadius, c.borderBottomLeftRadius], clase: el.className }
}, MARCA)
chequear(forma !== null && forma.alignSelf === "center", G, "3c", "la marca va CENTRADA en el transcript (align-self), no pegada a un lado como una burbuja", JSON.stringify(forma?.alignSelf))
chequear(forma !== null && forma.borde === "dashed", G, "3d", "…y su borde es punteado", JSON.stringify(forma?.borde))
chequear(forma !== null && forma.radios[0] === forma.radios[1], G, "3e", "…y NO tiene cola de burbuja (los dos radios inferiores son iguales)", JSON.stringify(forma?.radios))
await foto(page, "e2e3-02-marca-de-rotacion")

// ── paso 4 · la lista NO se partió ───────────────────────────────────────────
const d4 = await conv()
const a4 = d4.conversaciones.find(c => c.activa)
chequear(d4.conversaciones.length === antesN, G, "4a", "la lista sigue mostrando N, no N+1 (la rotación es INVISIBLE, CV-D10)", `${antesN} → ${d4.conversaciones.length}`)
chequear(a4.titulo === tituloAntes, G, "4b", "el título de la activa NO cambió", `«${a4.titulo.slice(0, 40)}»`)
chequear(a4.ctx_pct < a2.ctx_pct, G, "4c", "su ctx BAJÓ tras rotar", `${a2.ctx_pct}% → ${a4.ctx_pct}%`)

// ── paso 5 · el detalle: cc-id NUEVO ─────────────────────────────────────────
chequear(a4.claude_session_id !== ccAntes, G, "5a", "el `cc-id` de la conversación es NUEVO", `${ccAntes} → ${a4.claude_session_id}`)
const chipExp = await page.locator('button[aria-label*="contexto"]').getAttribute("aria-expanded")
chequear(chipExp === "true", G, "5b", "el detalle de identidad se abrió SOLO tras la rotación", `aria-expanded=${chipExp}`)
const ccEnPantalla = (await page.locator("body").innerText()).includes(a4.claude_session_id.slice(0, 8))
chequear(ccEnPantalla, G, "5c", "…y muestra el cc-id nuevo", `busca «${a4.claude_session_id.slice(0, 8)}»`)
await foto(page, "e2e3-03-detalle-cc-nuevo")

// ── paso 6 · con el dock COLAPSADO ───────────────────────────────────────────
writeFileSync(CTL, "MOCK_TOKENS=30000\n")
await page.getByRole("textbox").last().fill("cargando contexto otra vez")
await page.keyboard.press("Enter")
await page.waitForTimeout(4500)
await page.getByTitle("Colapsar el dock (⌘K para reabrir)").click()
await page.waitForTimeout(800)
const dockAbiertoAntes = await page.getByTitle("Colapsar el dock (⌘K para reabrir)").count()
writeFileSync(CTL, "MOCK_TOKENS=4000\n")
await api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "rotación con el dock cerrado" }) })
await page.waitForTimeout(6000)
const dockAbiertoDespues = await page.getByTitle("Colapsar el dock (⌘K para reabrir)").count()
chequear(dockAbiertoAntes === 0 && dockAbiertoDespues === 0, G, "6a", "el dock NO se abrió solo por una rotación", `abierto antes=${dockAbiertoAntes} después=${dockAbiertoDespues}`)
const toasts = await page.locator('[role="alert"],[class*="toast"]').count()
chequear(toasts === 0, G, "6b", "cero toasts: la rotación es invisible, no un aviso", `alert/toast=${toasts}`)
await page.getByRole("button", { name: /Conversar/ }).click()
await page.waitForTimeout(4000) // el dock re-monta y repinta el transcript entero: 1,5 s no alcanzaba
const marcas = await page.evaluate(m => document.body.innerText.split(m).length - 1, MARCA)
const enServidor = ((await api("/api/sessions")).json.find(s => s.id === SID).activa.conv || []).filter(t => t.text === MARCA).length
chequear(marcas === enServidor, G, "6c", "al reabrir, el dock muestra TODAS las marcas que el servidor tiene (ninguna se perdió al colapsar)", `pantalla=${marcas} · servidor=${enServidor}`)
await foto(page, "e2e3-04-dock-reabierto-dos-marcas")

// ── paso 7 · el servidor: cero conversaciones nuevas, cadena_cc +1 ───────────
const d7 = await conv()
const a7 = d7.conversaciones.find(c => c.activa)
chequear(d7.conversaciones.length === antesN, G, "7a", "CERO conversaciones nuevas tras dos rotaciones", `total=${d7.conversaciones.length}`)
// `cadena_cc` no viaja en el DTO de la LISTA (ConversacionResumen no lo declara): vive en
// ConversacionActiva, que sale por GET /api/sessions. Se lee de donde está, no de donde
// sería cómodo — pedirlo en la lista devolvería `undefined` y parecería un cero.
const act7 = (await api("/api/sessions")).json.find(s => s.id === SID).activa
const cadena = act7.cadena_cc ?? []
chequear(cadena.length >= 2, G, "7b", "la `cadena_cc` de la activa creció con cada rotación", `cadena_cc=${JSON.stringify(cadena)}`)
const marcasServidor = act7.conv.filter(t => t.text === MARCA).length
chequear(marcasServidor >= 2, G, "7c", "el transcript persistido tiene UNA marca por rotación", `marcas en el servidor=${marcasServidor}`)

console.log("\n=== consola del navegador ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await browser.close()
const r = resumen("E2E-3")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
