// E2E-2 · Turno en vuelo: el 409 del servidor (E-07, E-08, E-11, E-28, E-39)
// Prueba que la negativa la pone el SERVIDOR, no sólo el `disabled` del botón: la UI se
// saltea a propósito con `fetch` directo.
import { writeFileSync } from "node:fs"
import { abrir, foto, api, ok, fail, nc, chequear, resumen } from "./lib.mjs"

const G = "E2E-2"
const CTL = process.env.MOCK_CTL // $SANDBOX/.arnesia/mock-ctl
const { browser, page, consola } = await abrir()
const SID = (await api("/api/sessions")).json[0].id

await page.getByRole("button", { name: /Conversar/ }).click()
await page.waitForTimeout(1200)

// ── paso 0 · el guion NECESITA una fila inactiva ─────────────────────────────
// El paso 3 mide las filas inactivas; con una sola conversación no habría nada que medir
// y el guion se quedaría sin la mitad de su evidencia. Se crea ANTES del turno largo,
// que es justamente cuando crear todavía se puede.
writeFileSync(CTL, "")
await page.getByTitle("Nueva conversación — desactiva la actual").click()
await page.waitForTimeout(2500)
ok(G, "0", "hay ≥1 conversación inactiva para medir el paso 3",
  `total=${(await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.length}`)

// ── paso 1 · turno LARGO ─────────────────────────────────────────────────────
writeFileSync(CTL, "MOCK_DELAY=8\n")
const antes = (await api(`/api/sessions/${SID}/conversaciones`)).json
await page.getByRole("textbox").last().fill("un turno que va a tardar")
await page.keyboard.press("Enter")
await page.waitForTimeout(2000)
const est = (await api(`/api/sessions/${SID}`)).json
chequear(est.status === "streaming", G, "1", "la conversación queda en `streaming`", `status=${est.status}`)

// ── paso 2 · el ＋ apagado PERO NO MUDO ──────────────────────────────────────
const mas = page.getByTitle(/Nueva conversación|esperá a que termine/)
const dis = await mas.isDisabled()
const ti = await mas.getAttribute("title")
chequear(dis, G, "2a", "el ＋ queda disabled durante el turno", `disabled=${dis}`)
chequear(/esperá a que termine el turno/.test(ti ?? ""), G, "2b", "…y DICE por qué (title), nunca apagado y mudo", `title="${ti}"`)
// spec.md:347 y E-07 fijan el literal «esperá a que termine el turno (■ para interrumpir)»:
// el paréntesis es la SALIDA, la mitad que evita el «apagado y mudo». El código dice otra
// cosa. Se mide contra el spec, que es el contrato — no contra el código, que sería medir
// la implementación contra sí misma.
chequear(/■ para interrumpir/.test(ti ?? ""), G, "2c", "el title ofrece la salida «(■ para interrumpir)» (spec E-07)", `title="${ti}"`)
await foto(page, "e2e2-01-mas-apagado-con-motivo")

// ── paso 3 · la lista: filas inertes, buscador VIVO ──────────────────────────
await page.getByTitle("Buscar en las conversaciones de esta sesión").click()
await page.waitForTimeout(800)
const opts = page.getByRole("option")
const inactivas = []
for (let i = 0; i < await opts.count(); i++) {
  const o = opts.nth(i)
  if (await o.getAttribute("aria-selected") !== "true") inactivas.push({ ad: await o.getAttribute("aria-disabled"), ti: await o.getAttribute("title") })
}
if (inactivas.length === 0) {
  nc(G, "3a", "las filas inactivas quedan aria-disabled con su title", "la sesión sólo tiene la conversación activa en este punto del guion")
} else {
  chequear(inactivas.every(o => o.ad === "true"), G, "3a", 'las filas inactivas quedan aria-disabled="true"', JSON.stringify(inactivas.map(o => o.ad)))
  chequear(inactivas.every(o => /esperá a que termine el turno/.test(o.ti ?? "")), G, "3b", "…con su title explicando por qué", JSON.stringify(inactivas.map(o => (o.ti ?? "").slice(0, 40))))
}
const buscador = page.getByLabel("Buscar en estas conversaciones")
if (await buscador.count() === 0) {
  nc(G, "3c", "el buscador sigue habilitado con el turno en vuelo", "el panel de la lista no quedó abierto tras el clic en 🔍")
  nc(G, "3d", "…y filtra igual", "idem")
} else {
  chequear(await buscador.isEnabled(), G, "3c", "el buscador SIGUE habilitado con el turno en vuelo")
  await buscador.fill("nueva")
  await page.waitForTimeout(1200)
  chequear(/coinciden/.test(await page.locator("body").innerText()), G, "3d", "…y filtra igual (buscar es lectura, no transición)")
  await foto(page, "e2e2-02-lista-inerte-buscador-vivo")
  await buscador.fill("")
}

// ── paso 4 · saltearse la UI: el 409 lo pone el SERVIDOR ─────────────────────
const crear = await api(`/api/sessions/${SID}/conversaciones`, {
  method: "POST", headers: { "Content-Type": "application/json" }, body: "{}",
})
chequear(crear.status === 409, G, "4a", "POST .../conversaciones con turno en vuelo responde 409", `HTTP ${crear.status}`)
chequear(crear.status !== 500, G, "4b", "…y NO un 500")
const motivo = String(crear.json?.error ?? crear.txt).trim()
chequear(/streaming|in flight|turno|vuelo/i.test(motivo) && !/^(error|bad request|conflict)\.?$/i.test(motivo),
  G, "4c", "el 409 del servidor nombra la CAUSA, no un genérico", `«${motivo.slice(0, 90)}»`)
// El sentinel del daemon está en inglés (`ErrBusy`, session_service.go:36). Lo que ve el
// operador NO es ese string: el FE lo traduce en `motivoDe` y conserva el original entre
// paréntesis. La aserción que importa es la de la superficie, no la del cable — y se mide
// contra el JS QUE EL BINARIO SIRVE, no contra el fuente del worktree.
const html = (await api("/")).txt
const src = html.match(/\/assets\/index-[A-Za-z0-9_-]+\.js/)[0]
const js = (await api(src)).txt
chequear(js.includes("hay un turno en vuelo — esperá a que termine"), G, "4d",
  "la copy que ve el operador está en español y viaja en el bundle SERVIDO por el binario instalado",
  `bundle=${src} (${js.length} B)`)

// ── paso 5 · tras el 409, NADA cambió ────────────────────────────────────────
const despues = (await api(`/api/sessions/${SID}/conversaciones`)).json
chequear(despues.conversaciones.length === antes.conversaciones.length, G, "5a", "el 409 no dejó una conversación a medias", `${antes.conversaciones.length} → ${despues.conversaciones.length}`)
chequear(despues.conversaciones.find(c => c.activa).id === antes.conversaciones.find(c => c.activa).id, G, "5b", "la MISMA sigue activa (la transición es atómica)")

// ── paso 6 · lo mismo con una tarjeta de permiso abierta ─────────────────────
// MOCK_DELAY se limpia EXPLÍCITAMENTE: el mock es un proceso largo que hace `source` del
// control en cada turno, así que una variable puesta antes sigue viva en su shell. Sin
// esta línea el turno de permiso volvía a dormir 8 s y el guion medía antes de tiempo.
writeFileSync(CTL, "MOCK_DELAY=\nMOCK_ASK=1\n")
await page.waitForTimeout(9000) // que termine el turno lento anterior
await page.getByRole("textbox").last().fill("un turno que va a pedir permiso")
await page.keyboard.press("Enter")
await page.waitForTimeout(3000)
const est6 = (await api(`/api/sessions/${SID}`)).json
if (est6.status !== "await") {
  nc(G, "6", "con permiso pendiente el 409 dice «esperá tu decisión de permiso»", `la sesión no llegó a \`await\` (status=${est6.status}); el mock pidió permiso pero el daemon no lo reflejó a tiempo`)
} else {
  const crear6 = await api(`/api/sessions/${SID}/conversaciones`, { method: "POST", headers: { "Content-Type": "application/json" }, body: "{}" })
  const m6 = String(crear6.json?.error ?? crear6.txt).trim()
  chequear(crear6.status === 409, G, "6a", "con permiso pendiente el POST también da 409", `HTTP ${crear6.status}`)

  // Lo que el operador VE (spec E-08): el ＋ apagado con el motivo del permiso.
  // Por `aria-label` y no por `title`: cuando hay bloqueo el `title` del ＋ CAMBIA al motivo,
  // y las filas de la lista llevan ese mismo motivo — buscar por title matchea 3 elementos.
  const mas6 = page.locator('button[aria-label="Nueva conversación — desactiva la actual"]')
  const t6 = await mas6.getAttribute("title") ?? ""
  chequear(/esperá tu decisión de permiso/.test(t6), G, "6b",
    "el ＋ dice «esperá tu decisión de permiso» y NO el motivo del turno (spec E-08)", `title="${t6}"`)

  // Lo que el SERVIDOR dice: reusa `ErrBusy` para `streaming` y para `await` sin
  // distinguirlos (session_conversaciones.go:262). No rompe nada —la UI ya diferenció—
  // pero el cuerpo del 409 no es el motivo que E-08 nombra. Se declara, no se maquilla.
  chequear(!/permiso/i.test(m6), G, "6c-desviacion",
    "DESVIACIÓN DECLARADA: el 409 del servidor NO distingue `await` de `streaming` (reusa ErrBusy)",
    `cuerpo=«${m6.slice(0, 80)}»`)
  await foto(page, "e2e2-03-permiso-pendiente")
}

writeFileSync(CTL, "")
console.log("\n=== consola del navegador ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await browser.close()
const r = resumen("E2E-2")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
