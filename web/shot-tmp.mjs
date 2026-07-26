import { chromium } from 'playwright'
const OUT = '/tmp/claude-1000/-home-chalreme-Proyectos-harness-studio/208307d7-866a-4717-b377-ff99304333ad/scratchpad'
const F = 'file:///home/chalreme/Proyectos/harness-studio/docs/product/stories/2026-07-26-conversaciones-del-panel/mockup-conversaciones-panel.html'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1280, height: 1000 }, deviceScaleFactor: 2 })
let bad = 0
p.on('console', m => { if (m.type()==='error') { bad++; console.log('ERR', m.text().slice(0,140)) } })
p.on('pageerror', e => { bad++; console.log('PAGEERROR', String(e).slice(0,140)) })
await p.goto(F, { waitUntil: 'load' })
await p.evaluate(() => { document.documentElement.dataset.theme = 'dark' })
await p.waitForTimeout(350)
await p.locator('section.panel').nth(0).screenshot({ path: OUT + '/c-1.png' })
await p.locator('section.panel').nth(1).screenshot({ path: OUT + '/c-2a.png' })
await p.locator('section.panel').nth(2).screenshot({ path: OUT + '/c-2b.png' })
await p.locator('section.panel').nth(3).screenshot({ path: OUT + '/c-2c.png' })
await p.screenshot({ path: OUT + '/c-full-dark.png', fullPage: true })
console.log('scrollW/clientW', await p.evaluate(() => [document.documentElement.scrollWidth, document.documentElement.clientWidth]))
await p.evaluate(() => { document.documentElement.dataset.theme = 'light' })
await p.waitForTimeout(250)
await p.screenshot({ path: OUT + '/c-full-light.png', fullPage: true })
const txt = await p.locator('body').innerText()
console.log('mojibake:', /Ã|â€|Â/.test(txt) ? 'SI' : 'no', '| paneles:', await p.locator('section.panel').count(), '| slines fuera de §1:', await p.locator('.sline').count(), '| consola:', bad===0?'limpia':bad)
await b.close()
