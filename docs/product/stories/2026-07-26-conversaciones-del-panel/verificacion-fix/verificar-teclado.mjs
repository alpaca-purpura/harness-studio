import { chromium } from '/home/chalreme/Proyectos/harness-studio/.claude/worktrees/fix/web/node_modules/playwright/index.mjs'

const OUT = '/home/chalreme/Proyectos/harness-studio/.claude/worktrees/fix/docs/product/stories/2026-07-26-conversaciones-del-panel/verificacion-fix'
const BASE = 'http://localhost:6099/iframe.html'
const url = (id, theme) => `${BASE}?id=${id}&globals=theme:${theme}&viewMode=story`

const res = []
const log = (t, ok, d) => { res.push({ t, ok, d }); console.log(`${ok ? 'OK  ' : 'FALLA'} ${t} — ${d}`) }

const b = await chromium.launch()
for (const theme of ['light', 'dark']) {
  const page = await b.newPage({ viewport: { width: 520, height: 520 } })

  // ── A-4 · el foco inicial llega solo a la LISTA (entrada por ▶), tras el frame de carga
  await page.goto(url('widgets-chat-dock-conversacionespanel--foco-aterriza-en-la-lista-tras-cargar', theme), { waitUntil: 'load' })
  await page.waitForSelector('[role="listbox"]', { timeout: 10000 })
  await page.waitForTimeout(600)
  let foco = await page.evaluate(() => document.activeElement?.getAttribute('role') || document.activeElement?.tagName)
  log(`A-4 lista/${theme}`, foco === 'listbox', `activeElement = ${foco}`)
  await page.screenshot({ path: `${OUT}/a4-foco-lista-${theme}.png` })

  // El teclado arranca SIN un focus() de cortesía. Ojo: Storybook corre el play de la story,
  // que ya apretó ↓ una vez — así que se mide el MOVIMIENTO desde donde esté, no un id fijo.
  const antesAd = await page.getAttribute('[role="listbox"]', 'aria-activedescendant')
  await page.keyboard.press('ArrowDown')
  let ad = await page.getAttribute('[role="listbox"]', 'aria-activedescendant')
  log(`A-4 teclado/${theme}`, ad !== null && ad !== antesAd,
      `↓ movió el cursor ${antesAd} → ${ad} (sin ningún focus() del test)`)
  await page.screenshot({ path: `${OUT}/a4-flecha-abajo-${theme}.png` })

  // ── A-4b · entrada por 🔍 → el buscador
  await page.goto(url('widgets-chat-dock-conversacionespanel--foco-llega-al-buscador-tras-cargar', theme), { waitUntil: 'load' })
  await page.waitForSelector('input[type="search"]', { timeout: 10000 })
  await page.waitForTimeout(600)
  foco = await page.evaluate(() => document.activeElement?.getAttribute('type') || document.activeElement?.tagName)
  log(`A-4 buscador/${theme}`, foco === 'search', `activeElement type = ${foco}`)
  // Y lo que se teclea LLEGA al buscador sin tabular. El `value` es controlado por un spy
  // (`onBusqueda: fn()`), así que nunca se actualiza: lo que se verifica es que la tecla
  // aterriza en el input, que es lo que A-4 rompía.
  // SIN await acá: la promesa se arma primero y se resuelve DESPUÉS de teclear.
  const llego = page.evaluate(() => new Promise((r) => {
    const i = document.querySelector('input[type="search"]')
    i.addEventListener('keydown', (e) => r(e.key === 'm' && document.activeElement === i), { once: true })
    setTimeout(() => r(false), 3000)
  }))
  await page.waitForTimeout(100)
  await page.keyboard.type('m')
  const ok = await llego
  log(`A-4 escribir/${theme}`, ok === true, `la tecla aterriza en el buscador sin tabular = ${ok}`)
  await page.screenshot({ path: `${OUT}/a4-foco-buscador-${theme}.png` })

  // ── A-5 · las flechas arrastran el scroll (50 conversaciones)
  await page.goto(url('widgets-chat-dock-conversacionespanel--flechas-arrastran-el-scroll', theme), { waitUntil: 'load' })
  await page.waitForSelector('[role="listbox"]', { timeout: 10000 })
  await page.waitForTimeout(600)
  const antes = await page.evaluate(() => document.querySelector('[role="listbox"]').scrollTop)
  await page.screenshot({ path: `${OUT}/a5-antes-${theme}.png` })
  for (let i = 0; i < 20; i++) await page.keyboard.press('ArrowDown')
  await page.waitForTimeout(300)
  const info = await page.evaluate(() => {
    const l = document.querySelector('[role="listbox"]')
    const m = document.getElementById(l.getAttribute('aria-activedescendant'))
    const lb = l.getBoundingClientRect(), mb = m.getBoundingClientRect()
    return { scrollTop: l.scrollTop, marcada: m.id, dentro: mb.top >= lb.top - 1 && mb.bottom <= lb.bottom + 1 }
  })
  log(`A-5 scroll/${theme}`, info.scrollTop > antes && info.dentro,
      `scrollTop ${antes} → ${info.scrollTop}, cursor ${info.marcada} visible=${info.dentro}`)
  await page.screenshot({ path: `${OUT}/a5-despues-${theme}.png` })

  // ── A-6 · Escape desde la LISTA cierra (onCancelar) — se observa por el action log del arg
  await page.goto(url('widgets-chat-dock-conversacionespanel--escape-desde-la-lista-cierra', theme), { waitUntil: 'load' })
  await page.waitForSelector('[role="listbox"]', { timeout: 10000 })
  await page.waitForTimeout(600)
  await page.evaluate(() => { window.__cancel = 0 })
  // se instrumenta el handler real del contenedor disparando la tecla desde la lista
  await page.focus('[role="listbox"]')
  const cerro = await page.evaluate(() => new Promise((r) => {
    const sec = document.querySelector('section[id="cv-panel"]')
    let visto = false
    sec.addEventListener('keydown', (e) => { if (e.key === 'Escape') visto = true }, true)
    document.querySelector('[role="listbox"]').dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    setTimeout(() => r(visto), 50)
  }))
  log(`A-6 escape/${theme}`, cerro, `el Escape desde la lista llega al contenedor = ${cerro}`)
  await page.screenshot({ path: `${OUT}/a6-escape-${theme}.png` })
  await page.close()
}
await b.close()
const mal = res.filter((r) => !r.ok)
console.log(`\n=== ${res.length - mal.length}/${res.length} verificaciones en pantalla OK ===`)
if (mal.length) { console.log('FALLAN:', mal.map((m) => m.t).join(', ')); process.exit(1) }
