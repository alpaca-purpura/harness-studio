// E2E REAL (RF-110..117 + goal): claude de verdad detrás del daemon; pedimos una
// edición marcada al arnés dogfood (copia), la aprobamos en la tarjeta y verificamos
// archivo editado + gate. Correr tras levantar daemon (:4271, --claude real) y vite.
import { chromium } from '/home/chalreme/Proyectos/harness-studio/web/node_modules/playwright/index.mjs'

const OUT = process.env.E2E_OUT ?? '/tmp'
const MARKER = '<!-- e2e: chat-cc-funcional 2026-07-08 -->'

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

await page.goto('http://localhost:5173')
await page.waitForTimeout(1500)
await page.keyboard.press('Control+k')
await page.waitForTimeout(400)

// alcance: nodo builder
await page.getByText('/builder', { exact: true }).first().waitFor({ timeout: 10000 })
await page.getByText('/builder', { exact: true }).first().click()
await page.waitForTimeout(400)

const ta = page.locator('textarea')
await ta.fill(
  `Añade exactamente esta línea al FINAL del archivo skills/builder/SKILL.md: ${MARKER} — no cambies absolutamente nada más, no leas otros archivos.`,
)
await ta.press('Enter')
console.log('turno enviado — esperando a claude real…')

// Loop de conducción: aprobar cada tarjeta que aparezca (una vez) hasta que el turno
// termine (composer vuelve a «Pídele un cambio a…»). Máx 5 min.
const deadline = Date.now() + 300_000
let approvals = 0
let sawCard = false
while (Date.now() < deadline) {
  const done = (await page.getByPlaceholder(/Pídele un cambio a/).count()) > 0
  const btn = page.getByRole('button', { name: 'Permitir una vez' })
  if ((await btn.count()) > 0 && (await btn.first().isVisible())) {
    sawCard = true
    await page.screenshot({ path: `${OUT}/real-01-card-${++approvals}.png` })
    await btn.first().click()
    console.log(`  → aprobación #${approvals}`)
    await page.waitForTimeout(1000)
    continue
  }
  if (done && approvals > 0) break
  await page.waitForTimeout(1500)
}

assert(sawCard, 'claude real pidió permiso por la tarjeta (can_use_tool → Dock)')
assert(
  (await page.getByPlaceholder(/Pídele un cambio a/).count()) > 0,
  'el turno terminó (sesión idle)',
)
await page.waitForTimeout(2500)
assert(
  (await page.getByText(/🛡 gate de conformance dev-full-cycle/).count()) >= 1,
  'gate de conformance visible tras la edición real',
)
assert(
  (await page.getByText(/sin bloqueos, el arnés sigue conforme/).count()) >= 1,
  'gate sin bloqueos (el cambio no rompió la doctrina; warns preexistentes visibles)',
)
await page.screenshot({ path: `${OUT}/real-02-final.png` })

console.log(`\n${pass} pass / ${fail} fail · aprobaciones: ${approvals}`)
console.log(
  consoleMsgs.length ? `CONSOLA:\n${consoleMsgs.join('\n')}` : 'consola limpia (0 error/warn)',
)
await browser.close()
process.exit(fail ? 1 : 0)
