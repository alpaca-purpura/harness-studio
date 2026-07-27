// E2E-5 · Los fallos reales del entorno (E-14, E-49)
// La tesis de BR-CV-2: listar y buscar son DATOS NUESTROS (cuelgan de `session_id`), así que
// siguen funcionando aunque el arnés ya no esté en el disco. Sólo MANDAR UN TURNO depende
// del árbol, y cuando falla tiene que decir la ruta y devolver la conversación a `idle`.
//
// ⚠ El arnés que este guion mueve vive DENTRO del sandbox ($HOME/arneses-e2e/vitalia).
// Jamás se toca `/home/chalreme/Proyectos/...`: mover el proyecto real del operador para
// probar un mensaje de error sería exactamente la clase de daño que el aislamiento evita.
import { existsSync, renameSync } from "node:fs"
import { abrir, abrirBuscador, foto, api, ok, fail, nc, chequear, resumen } from "./lib.mjs"

const G = "E2E-5"
const ARNES = process.env.E2E_ARNES_DIR // $SANDBOX/arneses-e2e/vitalia
const MOVIDO = ARNES + ".movido"
const { browser, page, consola } = await abrir()
const SID = (await api("/api/sessions")).json[0].id

await page.getByRole("button", { name: /Conversar/ }).click()
await page.waitForTimeout(1200)

// El buscador SÓLO se dibuja con más de una conversación (`total > 1`,
// conversaciones-panel.tsx:168) — decisión sensata: entre una no hay nada que buscar. El
// guion necesita medirlo, así que crea la segunda ANTES de romper el entorno.
await page.getByTitle("Nueva conversación — desactiva la actual").click()
await page.waitForTimeout(2500)
const nAntes = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.length
ok(G, "0", "la sesión tiene ≥2 conversaciones (condición para que el buscador exista)", `total=${nAntes}`)

// ── paso 1 · desaparece la carpeta del arnés ─────────────────────────────────
renameSync(ARNES, MOVIDO)
ok(G, "1", "la carpeta del arnés se movió fuera de su lugar", `${ARNES} → ${MOVIDO}`)

// ── paso 2 · listar y buscar SIGUEN funcionando ──────────────────────────────
const buscador = await abrirBuscador(page)
const opts = await page.getByRole("option").count()
chequear(opts === nAntes, G, "2a", "la lista sigue mostrando las N conversaciones sin el arnés en disco (BR-CV-2)", `option=${opts} · esperadas=${nAntes}`)
await buscador.fill("repo")
await page.waitForTimeout(1200)
chequear(/coinciden/.test(await page.locator("body").innerText()), G, "2b", "…y el buscador también: son datos nuestros, no del árbol")
await buscador.fill("")
const listadoAPI = await api(`/api/sessions/${SID}/conversaciones`)
chequear(listadoAPI.status === 200, G, "2c", "el endpoint de conversaciones responde 200 sin el arnés", `HTTP ${listadoAPI.status}`)
await foto(page, "e2e5-01-lista-viva-sin-arnes")

// ── paso 3 · mandar un turno FALLA, nombrando la ruta ────────────────────────
const t = await api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "un turno sin arnés en disco" }) })
await page.waitForTimeout(3500)
const cuerpo = String(t.json?.error ?? t.txt).trim()
const est = (await api(`/api/sessions/${SID}`)).json
const textoUI = await page.locator("body").innerText()
const nombraRuta = cuerpo.includes(MOVIDO) || cuerpo.includes(ARNES) || textoUI.includes(ARNES) || /no such file|does not exist|no existe/i.test(cuerpo + textoUI)
chequear(t.status >= 400 || est.status === "idle", G, "3a", "el turno no queda aceptado en silencio", `HTTP ${t.status} · status=${est.status}`)
chequear(nombraRuta, G, "3b", "el fallo NOMBRA la ruta que no encontró (o dice que no existe)", `cuerpo=«${cuerpo.slice(0, 120)}»`)
chequear(est.status !== "streaming", G, "3c", "la conversación vuelve a `idle` — NO queda colgada en `streaming`", `status=${est.status}`)
await foto(page, "e2e5-02-turno-fallido")

// ── paso 4 · restaurar y volver a mandar ─────────────────────────────────────
renameSync(MOVIDO, ARNES)
await page.waitForTimeout(1500)
const t4 = await api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "y ahora con el arnés de vuelta" }) })
await page.waitForTimeout(4000)
const est4 = (await api(`/api/sessions/${SID}`)).json
chequear(t4.status === 202 || t4.status === 200, G, "4a", "con la carpeta de vuelta, el turno se acepta", `HTTP ${t4.status}`)
chequear(est4.status === "idle", G, "4b", "…y completa (vuelve a `idle`)", `status=${est4.status}`)

// ── paso 5-6 · desvincular el arnés NO se lleva la sesión por delante ────────
const antesDesv = (await api(`/api/sessions/${SID}/conversaciones`)).json.conversaciones.length
const arnesDeLaSesion = (await api(`/api/sessions/${SID}`)).json.arnes
const port = await api("/api/portafolio")
if (port.status !== 200) {
  nc(G, "5", "desvincular el arnés deja la sesión y sus conversaciones en pie", `GET /api/portafolio respondió ${port.status}: no hay entrada que desvincular en este sandbox`)
  nc(G, "6", "las N conversaciones siguen ahí tras desvincular", "idem")
} else {
  const entradas = port.json?.entradas ?? []
  const cand = entradas.find(e => (e.nombre ?? "").includes(arnesDeLaSesion) || (e.identidad?.id ?? "").includes(arnesDeLaSesion))
  if (!cand) {
    nc(G, "5", "desvincular el arnés deja la sesión y sus conversaciones en pie",
      `el arnés de la sesión («${arnesDeLaSesion}») no está en el Portafolio del sandbox: no hay nada que desvincular. La cascada se prueba igual en el paso 6 con el arnés ausente del disco`)
  } else {
    const d = await api(`/api/portafolio/${encodeURIComponent(cand.identidad.home)}/${encodeURIComponent(cand.identidad.id)}`, { method: "DELETE" })
    chequear(d.status < 400, G, "5a", "el arnés se desvincula del Portafolio", `HTTP ${d.status}`)
  }
  const sigue = await api(`/api/sessions/${SID}`)
  chequear(sigue.status === 200, G, "5b", "la sesión SIGUE existiendo tras tocar el Portafolio (nada en cascada)", `HTTP ${sigue.status}`)
}
const despues = (await api(`/api/sessions/${SID}/conversaciones`)).json
chequear(despues.conversaciones.length === antesDesv, G, "6", "las N conversaciones siguen ahí — nada se borró en cascada", `${antesDesv} → ${despues.conversaciones.length}`)
await foto(page, "e2e5-03-sesion-en-pie")

if (existsSync(MOVIDO)) renameSync(MOVIDO, ARNES)
console.log("\n=== consola del navegador ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await browser.close()
const r = resumen("E2E-5")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
