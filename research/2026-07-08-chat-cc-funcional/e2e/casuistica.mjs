// E2E de casuística del chat (spec.md §Gherkin) contra el daemon REAL (:4271) con el
// mock-claude (stream-json real) y el vite dev (:5173). Corre con:
//   node research/2026-07-08-chat-cc-funcional/e2e/casuistica.mjs
// Precondición: daemon en :4271 con --claude mock-claude.sh; vite dev en :5173 con
// VITE_ARNESIA_API=http://127.0.0.1:4271; dogfood COPIADO registrado como dev-full-cycle.
import { chromium } from '/home/chalreme/Proyectos/harness-studio/web/node_modules/playwright/index.mjs'

const OUT = process.env.E2E_OUT ?? '/tmp'
const UI = 'http://localhost:5173'

const browser = await chromium.launch({ args: ['--no-sandbox', '--disable-dev-shm-usage'] })
const page = await browser.newPage({ viewport: { width: 1600, height: 950 } })
const consoleMsgs = []
page.on('console', (m) => {
  if (m.type() === 'error' || m.type() === 'warning') consoleMsgs.push(`${m.type()}: ${m.text()}`)
})
page.on('pageerror', (e) => consoleMsgs.push(`pageerror: ${e.message}`))

let pass = 0
let fail = 0
const assert = (cond, name) => {
  cond ? (pass++, console.log(`  ✓ ${name}`)) : (fail++, console.log(`  ✗ ${name}`))
}
const snap = (n) => page.screenshot({ path: `${OUT}/${n}.png` })

await page.goto(UI)
await page.waitForTimeout(1500)

// — dock abierto (⌘K) y sesión dev-full-cycle activa —
await page.keyboard.press('Control+k')
await page.waitForTimeout(400)
const dock = page.locator('aside, [class*=dock]').last()
assert(await page.getByText('Alcance:').isVisible(), 'fila de alcance visible en el dock')
assert((await page.getByText('dev-full-cycle').count()) > 0, 'chip fijo del arnés de la sesión')

// — RF-111: seleccionar el nodo builder en el Mapa ⇒ chip removible con archivo real —
await page.getByText('/builder', { exact: true }).first().waitFor({ timeout: 10000 })
await page.getByText('/builder', { exact: true }).first().click()
await page.waitForTimeout(400)
assert(
  (await page.getByText('skills/builder/SKILL.md').count()) > 0,
  'chip de nodo con fuente_path real (nomenclatura)',
)
await snap('e2e-01-scope')

// — Turno 1: allow-esta-sesión + grant auto-allow + gate —
const ta = page.locator('textarea')
await ta.fill('endurece el gate de builder: tests verdes antes de review')
await ta.press('Enter')
await page.waitForTimeout(1200)
assert(
  (await page.getByText(/\[alcance: skill «builder»/).count()) > 0,
  'el turno viaja con la línea de alcance',
)
assert(await page.getByText('Claude quiere usar').isVisible(), 'tarjeta de permiso aparece (await)')
assert((await page.getByText('tipo: manual').count()) > 0, 'diff: línea vieja visible (−)')
assert((await page.getByText(/tipo: auto/).count()) > 0, 'diff: línea nueva visible (+)')
assert(
  await page.getByPlaceholder('esperando tu decisión de permiso…').isVisible(),
  'composer bloqueado en await',
)
await snap('e2e-02-perm-card')

await page.getByRole('button', { name: 'Permitir esta sesión' }).click()
await page.waitForTimeout(1500)
assert((await page.getByText(/✓ Edit: allow/).count()) >= 1, 'rastro del allow en el transcript')
assert(
  (await page.getByText(/grant vigente/).count()) >= 1,
  'segundo Edit auto-permitido por grant (sin tarjeta)',
)
assert(
  (await page.getByText('Gate endurecido y capability alineada.').count()) > 0,
  'respuesta final del turno llega',
)
await page.waitForTimeout(1200)
assert(
  (await page.getByText(/🛡 gate de conformance dev-full-cycle/).count()) >= 1,
  'gate de conformance corre y se VE tras la edición (RF-117)',
)
await snap('e2e-03-gate')

// — Turno 2: deny —
await ta.fill('haz otra edición')
await ta.press('Enter')
await page.waitForTimeout(1000)
await page.getByRole('button', { name: 'Denegar' }).click()
await page.waitForTimeout(1200)
assert((await page.getByText(/✕ Write: deny/).count()) >= 1, 'deny deja rastro')
assert(
  (await page.getByText('ok, no toqué el archivo').count()) > 0,
  'la conversación sigue tras el deny (sin interrupt)',
)

// — Turno 3: allow-una-vez ⇒ re-pregunta —
await ta.fill('otra edición más')
await ta.press('Enter')
await page.waitForTimeout(1000)
await page.getByRole('button', { name: 'Permitir una vez' }).click()
// el mock espera 1.6s (grant de 1s expira) y vuelve a pedir el MISMO tool.
await page.waitForTimeout(2600)
assert(
  await page.getByText('Claude quiere usar').isVisible(),
  'grant «una vez» expirado ⇒ el mismo tool VUELVE a preguntar',
)
await snap('e2e-04-reask')
await page.getByRole('button', { name: 'Permitir una vez' }).click()
await page.waitForTimeout(1200)
assert((await page.getByText('segunda edición aplicada').count()) > 0, 'turno 3 completa')

// — Turno 4: interrupt (Stop) sobre un ask pendiente —
await ta.fill('corre la suite completa')
await ta.press('Enter')
await page.waitForTimeout(1000)
assert((await page.getByText(/pnpm test --run/).count()) > 0, 'ask de Bash con el comando visible')
await page.getByRole('button', { name: '■' }).click()
await page.waitForTimeout(1200)
assert(
  (await page.getByText(/interrumpido por el operador/).count()) >= 1,
  'el ask pendiente se denegó con motivo al interrumpir',
)
// El result del interrupt cierra el turno: el composer vuelve a estado idle (el bubble
// muestra los deltas ya ensamblados — el texto del result solo gana si no hubo deltas).
await page.waitForTimeout(800)
assert(
  (await page.getByPlaceholder(/Pídele un cambio a/).count()) > 0,
  'el conductor cerró el turno tras el interrupt in-band (sesión idle)',
)
await snap('e2e-05-interrupt')

// — RF-118: cambiar de sesión/arnés ⇒ chips y permisos de ESA sesión —
await page.getByText('ux-nordia').first().click()
await page.waitForTimeout(600)
assert(
  (await page.locator('textarea').count()) > 0 &&
    (await page.getByText('Claude quiere usar').count()) === 0,
  'la otra sesión no hereda tarjetas de permiso',
)
assert(
  (await page.getByText('selecciona un nodo en el Mapa para acotar').count()) > 0,
  'el chip de alcance NO viaja entre sesiones',
)
await snap('e2e-06-switch')

console.log(`\n${pass} pass / ${fail} fail`)
console.log(
  consoleMsgs.length ? `CONSOLA:\n${consoleMsgs.join('\n')}` : 'consola limpia (0 error/warn)',
)
await browser.close()
process.exit(fail ? 1 : 0)
