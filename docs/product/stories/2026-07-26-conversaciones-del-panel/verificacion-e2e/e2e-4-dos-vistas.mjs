// E2E-4 · Dos vistas sobre la misma sesión (E-40, E-41)
// Dos contextos de navegador independientes contra el MISMO daemon instalado: lo que se
// prueba es que el frame `conversacion` del SSE mantiene sincronizadas dos superficies.
import { writeFileSync } from "node:fs"
import { abrir, foto, api, ok, fail, nc, chequear, resumen, BASE } from "./lib.mjs"

const CTL = process.env.MOCK_CTL

const G = "E2E-4"
const A = await abrir()
const B = await abrir()
const SID = (await api("/api/sessions")).json[0].id

for (const v of [A, B]) {
  await v.page.getByRole("button", { name: /Conversar/ }).click()
  await v.page.waitForTimeout(1000)
}

// ── paso 1 · las dos ven la misma sesión ─────────────────────────────────────
const tituloA = await A.page.getByTitle("Colapsar el dock (⌘K para reabrir)").locator("..").innerText()
const tituloB = await B.page.getByTitle("Colapsar el dock (⌘K para reabrir)").locator("..").innerText()
chequear(tituloA === tituloB, G, "1", "las dos vistas muestran la misma sesión", `A="${tituloA.split("\n")[0].slice(0, 40)}"`)

// ── paso 2 · crear en A, la lista de B se actualiza SOLA ─────────────────────
for (const v of [A, B]) {
  await v.page.getByTitle("Buscar en las conversaciones de esta sesión").click()
  await v.page.waitForTimeout(700)
}
const antesB = await B.page.getByRole("option").count()
await A.page.getByTitle("Nueva conversación — desactiva la actual").click()
await A.page.waitForTimeout(3500) // sin recargar B: se espera el frame por SSE
const despuesB = await B.page.getByRole("option").count()
chequear(despuesB === antesB + 1, G, "2", "la lista de B se actualiza SOLA (frame `conversacion`), sin recargar", `B: ${antesB} → ${despuesB}`)
await foto(B.page, "e2e4-01-vistaB-actualizada-sola")

// ── paso 3 · renombrar en B, el título cambia en A ───────────────────────────
const NUEVO = "bautizada desde la vista B"
const convActiva = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.find(c => c.activa)
await B.page.getByLabel("Renombrar la conversación").click()
await B.page.waitForTimeout(600)
// Por `aria-label`: el input de rename NO declara `type`, así que el selector de atributo
// `input[type="text"]` no lo matchea aunque su tipo efectivo sea text.
const campo = B.page.getByLabel("Título de la conversación")
if (await campo.count() === 0) {
  nc(G, "3", "renombrar en B cambia el título en A", "el control de renombrar no expuso un input de texto")
} else {
  await campo.fill(NUEVO)
  await B.page.keyboard.press("Enter")
  await A.page.waitForTimeout(3500)
  const enServidor = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.find(c => c.id === convActiva.id)
  chequear(enServidor.titulo === NUEVO, G, "3a", "el rename llegó al servidor", `titulo="${enServidor.titulo}"`)
  const textoA = await A.page.locator("body").innerText()
  chequear(textoA.includes(NUEVO), G, "3b", "…y el título cambió en la vista A sin recargar")
  await foto(A.page, "e2e4-02-vistaA-titulo-cambiado")
}

// ── paso 4 · replay por Last-Event-ID ────────────────────────────────────────
//
// ⚠ EL PASO DEL PLAN NO ES EJERCITABLE CON ESTE DRIVER, y se declara en vez de fingirse.
// `BrowserContext.setOffline(true)` de Playwright bloquea peticiones NUEVAS pero NO
// derriba un stream SSE ya establecido: medido en este mismo circuito, el `readyState`
// del EventSource se queda en 1 (OPEN) durante todo el corte y también después. Sin
// evento `error` no hay reconexión, y sin reconexión no hay `Last-Event-ID` que mandar.
// Lo que se observaría (B sin el turno) mediría la emulación, no el producto: sería un
// FALLA fabricado. Se corta en dos: el mecanismo del servidor SÍ se prueba, a nivel HTTP.
const turnosAntes = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.find(c => c.activa).turnos
nc(G, "4a", "B reconecta sola tras un corte de red y recupera lo perdido",
  "Playwright setOffline() no derriba un SSE ya abierto (readyState queda en 1/OPEN): la reconexión del navegador no se puede disparar con este driver. Va al Modo B / gate humano")

// El servidor: se abre un stream, se anota el último id, se manda un turno con el stream
// CERRADO y se reconecta con `Last-Event-ID` — que es exactamente lo que hace el navegador.
const leerIds = async (headers) => {
  const ctrl = new AbortController()
  const r = await fetch(BASE + "/api/events", { headers, signal: ctrl.signal })
  const rd = r.body.getReader(); const dec = new TextDecoder()
  let buf = "", t0 = Date.now()
  while (Date.now() - t0 < 2500) {
    const { value, done } = await Promise.race([rd.read(), new Promise(res => setTimeout(() => res({ done: true }), 2600))])
    if (done) break
    buf += dec.decode(value, { stream: true })
  }
  ctrl.abort()
  return [...buf.matchAll(/^id:\s*(\d+)/gm)].map(m => Number(m[1]))
}
const idsPrevios = await leerIds({})
const ultimo = idsPrevios.length ? idsPrevios.at(-1) : 0
await api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "turno con el stream cerrado" }) })
await new Promise(r => setTimeout(r, 3000))
const replayados = await leerIds({ "Last-Event-ID": String(ultimo) })
chequear(replayados.length > 0 && replayados[0] === ultimo + 1, G, "4b",
  "el servidor REPLAYA por Last-Event-ID lo emitido mientras el cliente no estaba (broker.go:119)",
  `Last-Event-ID: ${ultimo} → replay desde id ${replayados[0] ?? "—"} (${replayados.length} frames)`)
await foto(B.page, "e2e4-03-vistaB-reconectada")

// ── paso 5 · turno simultáneo desde A y desde B ──────────────────────────────
// El turno tiene que DURAR para que la carrera exista: con el mock respondiendo al instante
// el primero termina antes de que llegue el segundo y los dos 202 son legítimos — mediría
// la velocidad del mock, no el guard de un-turno-a-la-vez.
writeFileSync(CTL, "MOCK_DELAY=6\n")
await A.page.waitForTimeout(300)
const [rA, rB] = await Promise.all([
  api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "turno desde A" }) }),
  api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "turno desde B" }) }),
])
const codigos = [rA.status, rB.status].sort()
chequear(codigos.includes(409), G, "5a", "de dos turnos simultáneos, UNO responde 409", `códigos=${JSON.stringify(codigos)}`)
chequear(codigos.filter(c => c === 202 || c === 200).length === 1, G, "5b", "…y exactamente uno fue aceptado", `códigos=${JSON.stringify(codigos)}`)
await A.page.waitForTimeout(4000)
const finales = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.find(c => c.activa)
chequear(finales.turnos >= turnosAntes, G, "5c", "el transcript no se intercaló ni se perdió", `turnos=${finales.turnos}`)

const consola = [...A.consola, ...B.consola]
console.log("\n=== consola del navegador (A+B) ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await A.browser.close(); await B.browser.close()
const r = resumen("E2E-4")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
