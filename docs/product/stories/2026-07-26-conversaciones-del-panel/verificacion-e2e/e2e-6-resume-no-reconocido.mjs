// E2E-6 · `--resume` que el CLI ya no reconoce (E-13)
//
// El `claude_session_id` se falsea con el daemon detenido (la variante que el propio plan
// autoriza) y el mock MUERE cuando lo recibe, que es exactamente lo que hace el CLI real
// cuando el corpus fue GC'd: el proceso cae ANTES del `init`. Es el disparador de
// `tryHealResume` (session_service.go:756) y no necesita un modelo real para producirse.
import { abrir, foto, api, ok, fail, nc, chequear, resumen } from "./lib.mjs"

const G = "E2E-6"
const { browser, page, consola } = await abrir()
const SID = (await api("/api/sessions")).json[0].id
const activa = async () => (await api("/api/sessions")).json.find(s => s.id === SID).activa

const antes = await activa()
ok(G, "1", "la conversación arranca con un claude_session_id que el CLI ya no reconoce", `cc-id falseado = «${antes.claude_session_id}»`)
const turnosAntes = antes.conv.length

// ── paso 2 · el turno: spawn con --resume muere, heal respawnea fresh y reenvía ──
const t = await api(`/api/sessions/${SID}/turn`, { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ text: "turno sobre un resume muerto" }) })
await page.waitForTimeout(9000)
const d = await activa()
chequear(t.status === 202 || t.status === 200, G, "2a", "el turno se acepta", `HTTP ${t.status}`)
chequear(d.claude_session_id !== antes.claude_session_id && d.claude_session_id !== "", G, "2b",
  "`tryHealResume` respawneó FRESH: el cc-id cambió solo", `${antes.claude_session_id} → ${d.claude_session_id}`)
const reenviado = d.conv.some(x => x.text === "turno sobre un resume muerto")
chequear(reenviado, G, "2c", "…y el turno en vuelo se REENVIÓ al proceso nuevo (no se perdió)", `turnos ${turnosAntes} → ${d.conv.length}`)

// ── paso 3 · los turnos viejos siguen ───────────────────────────────────────
chequear(d.conv.length > turnosAntes, G, "3", "los turnos viejos SIGUEN: vienen de `Conv`, no del proceso", `${turnosAntes} → ${d.conv.length}`)

// ── paso 4 · la marca de RF-348 CA-1 ────────────────────────────────────────
const enPantalla = await page.locator("body").innerText()
const marcaRF348 = /hilo reiniciado|checkpoint/i.test(enPantalla) || d.conv.some(x => /hilo reiniciado/i.test(x.text))
chequear(marcaRF348, G, "4", "el cambio de cc-id NO es silencioso: hay una marca «⟳ hilo reiniciado · checkpoint» (RF-348 CA-1)",
  marcaRF348 ? "presente" : "AUSENTE — `tryHealResume` está documentado como «silent on success» (session_service.go:755) y no emite ninguna marca; el literal no existe en el árbol (0 ocurrencias)")

// ── paso 5 · el detalle muestra el cc-id nuevo ──────────────────────────────
// El botón dice «Conversar» colapsado y «Conversando» abierto: /Conversar/ matchea los DOS,
// así que un clic incondicional CIERRA el dock que ya estaba abierto. Se abre sólo si hace falta.
if (await page.getByTitle("Colapsar el dock (⌘K para reabrir)").count() === 0) {
  await page.getByRole("button", { name: /Conversar/ }).click()
  await page.waitForTimeout(2500)
}
// El detalle es un disclosure: hay que ABRIRLO y confirmar que quedó abierto. Mirar el
// `innerText` del body sin desplegarlo mide el panel cerrado y da un falso negativo.
const chip = page.locator('button[aria-label*="contexto"]').first()
if (await chip.getAttribute("aria-expanded") !== "true") { await chip.click(); await page.waitForTimeout(1200) }
const detalle = await page.evaluate(() => {
  const d = document.querySelector('[id$="-detalle"]')
  return d ? d.innerText.replace(/\s+/g, " ").trim() : ""
})
chequear(chip && await chip.getAttribute("aria-expanded") === "true", G, "5a", "el detalle de identidad se despliega", `aria-expanded=${await chip.getAttribute("aria-expanded")}`)
chequear(detalle.includes(d.claude_session_id.slice(0, 8)), G, "5b", "…y muestra el cc-id NUEVO (no el muerto)", `detalle=«${detalle.slice(0, 70)}»`)
chequear(!detalle.includes(antes.claude_session_id.slice(0, 8)), G, "5c", "…y ya NO muestra el que el CLI no reconocía", `muerto=«${antes.claude_session_id}»`)
await foto(page, "e2e6-01-detalle-cc-nuevo-tras-heal")

// ── paso 6 · un SEGUNDO fallo: error visible, no un loop ────────────────────
nc(G, "6", "un segundo fallo consecutivo da `error` visible y NO un loop",
  "el heal es de UNA sola vez por proceso (`resumeRetried`, session_service.go:760) y el mock ya respawneó fresh: forzar el segundo fallo exige que el proceso FRESCO también muera, que es otro modo de fallo (spawn roto) y no el de este escenario. Queda para el gate humano")

console.log("\n=== consola del navegador ===")
console.log(consola.length ? JSON.stringify(consola, null, 1) : "(sin errores)")
await browser.close()
const r = resumen("E2E-6")
process.stdout.write("\nJSON:" + JSON.stringify({ guion: G, ...r, consola }) + "\n")
if (r.n.FALLA) process.exitCode = 1
