// E2E-1 · Crear, listar, buscar y retomar (E-02, E-03, E-06, E-10, E-22, E-24)
// Contra el BINARIO INSTALADO (~/.local/bin/arnesia) sirviendo la SPA EMBEBIDA.
import { abrir, foto, api, ok, fail, nc, chequear, resumen, pasos } from "./lib.mjs"

const G = "E2E-1"
const { browser, page, consola } = await abrir()

const SID = (await api("/api/sessions")).json[0].id
console.log(`sesión bajo prueba: ${SID}`)

// ── paso 1 · EXACTAMENTE 2 filas de cromo ────────────────────────────────────
await page.getByRole("button", { name: /Conversar/ }).click()
await page.waitForTimeout(1200)

const colapsar = page.getByTitle("Colapsar el dock (⌘K para reabrir)")
const filas = await page.evaluate(() => {
  const b = [...document.querySelectorAll("button")].find(n => /colapsar/.test(n.innerText))
  const f1 = b.parentElement
  const f2 = f1.nextElementSibling
  const f3 = f2?.nextElementSibling
  const esCromo = n => n && getComputedStyle(n).borderBottomWidth !== "0px" && n.className.includes("flex-none")
  return { n: [f1, f2, f3].filter(esCromo).length, txt: (f1.innerText + " " + (f2?.innerText ?? "")) }
})
const cuerpo = await page.locator("body").innerText()
chequear(filas.n === 2, G, "1a", "exactamente 2 filas de cromo entre el borde del dock y el transcript", `medidas=${filas.n}`)
chequear(!/Alcance:/.test(cuerpo), G, "1b", 'queryByText("Alcance:") === null', "no aparece la fila de alcance en reposo")
chequear(!filas.txt.includes("◍"), G, "1c", "ningún ◍ (cc-id) visible en el cromo", "la identidad técnica vive detrás del chip")
chequear(await page.getByTitle(/ver identidad de la conversación/).count() === 1, G, "1d", "el botón de ctx (chip-disclosure) está presente")
chequear(await colapsar.count() === 1 && /»/.test(await colapsar.innerText()), G, "1e", "el glifo de colapsar es » y no ⟩ (CV-D15)", await colapsar.innerText())
chequear(!filas.txt.includes("⟩"), G, "1f", "el glifo ⟩ ya no aparece")
await foto(page, "e2e1-01-cromo-2-filas")

// ── paso 2 · mandar un turno ─────────────────────────────────────────────────
const antes = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones
const turnosAntes = antes.find(c => c.activa).turnos
await page.getByRole("textbox").last().fill("dejame anotado el manifiesto de la caja")
await page.keyboard.press("Enter")
await page.waitForTimeout(3500)
const trasTurno = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.find(c => c.activa)
chequear(trasTurno.turnos > turnosAntes, G, "2a", "el transcript crece al mandar un turno", `${turnosAntes} → ${trasTurno.turnos}`)
chequear(await page.getByText("dejame anotado el manifiesto de la caja").count() > 0, G, "2b", "el turno del operador se pinta en el transcript")
await foto(page, "e2e1-02-turno-enviado")

// ── paso 3 · ＋ nace «nueva conversación» ────────────────────────────────────
await page.getByTitle("Nueva conversación — desactiva la actual").click()
await page.waitForTimeout(2500)
const trasMas = (await api(`/api/sessions/${SID}/conversaciones`)).json
const nueva = trasMas.conversaciones.find(c => c.activa)
chequear(trasMas.conversaciones.length === antes.length + 1, G, "3a", "crear suma UNA conversación", `${antes.length} → ${trasMas.conversaciones.length}`)
chequear(nueva.titulo === "nueva conversación", G, "3b", 'la nueva se llama «nueva conversación»', nueva.titulo)
chequear(nueva.turnos === 0, G, "3c", "nace con 0 turnos", String(nueva.turnos))
const chip = page.getByTitle(/ver identidad de la conversación/)
const chipTxt = (await chip.count()) ? await chip.innerText() : "(ausente)"
chequear(/0\s*%/.test(chipTxt), G, "3d", "el chip de ctx a 0% queda VISIBLE (no escondido)", `chip="${chipTxt.replace(/\n/g, " ")}"`)
chequear(trasMas.conversaciones.filter(c => c.activa).length === 1, G, "3e", "sigue habiendo exactamente UNA activa (CV-D7)")
await foto(page, "e2e1-03-nueva-conversacion")

// ── paso 4 · abrir la lista, EN SITIO ────────────────────────────────────────
const toggle = page.locator('button[aria-expanded][aria-controls]:not([aria-controls*="detalle"])').first()
await toggle.click()
await page.waitForTimeout(900)
const lb = page.getByRole("listbox")
const nOpt = await page.getByRole("option").count()
chequear(await lb.count() === 1, G, "4a", 'getByRole("listbox") existe (contrato ARIA sobre <div>, N-17)', `listbox=${await lb.count()}`)
chequear(nOpt === trasMas.conversaciones.length, G, "4b", `la lista trae ${trasMas.conversaciones.length} option`, `option=${nOpt}`)
chequear(await page.getByRole("textbox").last().isVisible(), G, "4c", "el composer SIGUE visible con la lista abierta (abre en sitio)")
chequear(await page.locator("dialog").count() === 0, G, "4d", "cero <dialog> en el DOM (no es un modal)")
chequear(await page.locator('[class*="backdrop"],[class*="overlay"]').count() === 0, G, "4e", "cero backdrop/overlay")
await foto(page, "e2e1-04-lista-abierta")

// ── paso 5 · buscar SIN Enter ────────────────────────────────────────────────
const buscador = page.getByLabel("Buscar en estas conversaciones")
if (await buscador.count() === 0) {
  // 🔍 abre la lista CON el buscador; si la lista ya estaba abierta por el ▶, el botón la
  // alterna, así que se re-abre. Es el mismo camino que hace el operador.
  await page.getByTitle("Buscar en las conversaciones de esta sesión").click()
  await page.waitForTimeout(600)
  if (await buscador.count() === 0) {
    await page.getByTitle("Buscar en las conversaciones de esta sesión").click()
    await page.waitForTimeout(600)
  }
}
chequear(await buscador.count() === 1, G, "5-0", "el buscador aparece dentro del panel de la lista")
await buscador.fill("manifiesto")
await page.waitForTimeout(1500)
const txt5 = await page.locator("body").innerText()
const rot = txt5.match(/(\d+)\s+de\s+(\d+)\s+coinciden/)
chequear(!!rot, G, "5a", 'el rótulo pasa a «N de M coinciden» SIN apretar Enter', rot ? rot[0] : "NO APARECIÓ")
chequear(await page.locator("mark").count() > 0, G, "5b", "las filas que coinciden traen <mark>", `marks=${await page.locator("mark").count()}`)
await foto(page, "e2e1-05-busqueda-con-coincidencias")

// ── paso 6 · buscar SIN coincidencias ────────────────────────────────────────
await buscador.fill("telemetría")
await page.waitForTimeout(1500)
const txt6 = await page.locator("body").innerText()
chequear(/Ninguna conversación de esta sesión menciona/.test(txt6), G, "6a", "el vacío NOMBRA lo buscado, no dice «0 conversaciones»")
chequear(/Se buscó en el título y en el texto/.test(txt6), G, "6b", "dice DÓNDE buscó y en cuántas")
chequear(/Limpiar búsqueda/.test(txt6), G, "6c", "ofrece «Limpiar búsqueda» (salida del callejón)")
chequear(!/0 conversaciones/.test(txt6), G, "6d", "NUNCA «0 conversaciones»")
await foto(page, "e2e1-06-busqueda-vacia")

// ── paso 7 · retomar una inactiva ────────────────────────────────────────────
await page.getByText("Limpiar búsqueda").click()
await page.waitForTimeout(800)
const inactiva = page.getByRole("option", { selected: false }).first()
const tituloInactiva = (await inactiva.innerText()).split("\n")[0]

// La franja es EFÍMERA (chat-dock.tsx:79 — vive sólo mientras `--resume` rehidrata), así
// que se la caza con un MutationObserver armado ANTES del clic. Mirar el DOM después del
// clic es una carrera perdida y daría un falso negativo.
await page.evaluate(() => {
  window.__retoma = []
  new MutationObserver(() => {
    for (const n of document.querySelectorAll('[role="status"]')) {
      const t = n.innerText.replace(/\s+/g, " ").trim()
      if (/Retomando/.test(t) && !window.__retoma.includes(t)) window.__retoma.push(t)
    }
  }).observe(document.body, { childList: true, subtree: true, characterData: true })
})
await inactiva.click()
await page.waitForTimeout(1500)
const capturada = await page.evaluate(() => window.__retoma)
const txt7 = await page.locator("body").innerText()
const franja = capturada.length ? [capturada[0]] : txt7.match(/Retomando la conversación[^\n]*/)
chequear(!!franja, G, "7a", "aparece la franja «Retomando la conversación… --resume <8 chars>»", franja ? franja[0] : "NO APARECIÓ")
// `\b` NO sirve de cierre: el cc-id del mock termina en «-» y un id real es un UUID que
// puede cortar en cualquier carácter. Se mide el LARGO del recorte, que es lo que dice
// el contrato (`cc.slice(0, 8)` en chat-dock.tsx:87).
const recorte = franja ? franja[0].match(/--resume\s+(\S+)/) : null
if (recorte) chequear(recorte[1].length === 8, G, "7b", "la franja nombra --resume con los 8 primeros caracteres del cc-id", `«${recorte[1]}» (${recorte[1].length} chars)`)
else nc(G, "7b", "la franja nombra --resume con 8 caracteres", "la franja no apareció")
await page.waitForTimeout(3500)
const trasRetomar = (await api(`/api/sessions/${SID}/conversaciones`)).json
const act = trasRetomar.conversaciones.find(c => c.activa)
chequear(act.titulo.slice(0, 12) === tituloInactiva.slice(0, 12), G, "7c", "la que se clickeó quedó ACTIVA", `activa="${act.titulo.slice(0, 30)}"`)
chequear(act.turnos > 0, G, "7d", "el transcript se repinta con los turnos viejos (vienen de Conv)", `turnos=${act.turnos}`)
const det = page.getByTitle(/ver identidad de la conversación/)
chequear(await det.getAttribute("aria-expanded") === "true", G, "7e", "el detalle de identidad se abre SOLO al retomar", `aria-expanded=${await det.getAttribute("aria-expanded")}`)
await foto(page, "e2e1-07-retomada")

// ── paso 8 · la invariante en el servidor ────────────────────────────────────
chequear(trasRetomar.conversaciones.filter(c => c.activa).length === 1, G, "8", "exactamente UNA con activa=true", `total=${trasRetomar.conversaciones.length}`)

console.log("\n=== consola del navegador ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await browser.close()
const r = resumen("E2E-1")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
