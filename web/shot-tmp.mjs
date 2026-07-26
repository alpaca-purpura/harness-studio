import { chromium } from 'playwright'
const OUT = '/tmp/claude-1000/-home-chalreme-Proyectos-harness-studio/208307d7-866a-4717-b377-ff99304333ad/scratchpad'
const F = 'file:///home/chalreme/Proyectos/harness-studio/docs/product/stories/2026-07-26-conversaciones-del-panel/mockup-conversaciones-panel.html'
const b = await chromium.launch()
const p = await b.newPage({ viewport: { width: 1280, height: 1000 }, deviceScaleFactor: 2 })
await p.goto(F, { waitUntil: 'load' })
await p.evaluate(() => { document.documentElement.dataset.theme = 'dark' })
await p.waitForTimeout(300)
await p.locator('section.panel').nth(3).screenshot({ path: OUT + '/z-3a.png' })
await p.locator('section.panel').nth(4).screenshot({ path: OUT + '/z-3b.png' })
await p.locator('section.panel').nth(8).screenshot({ path: OUT + '/z-4a.png' })
await p.evaluate(() => { document.documentElement.dataset.theme = 'light' })
await p.waitForTimeout(250)
await p.locator('section.panel').nth(3).screenshot({ path: OUT + '/z-3a-light.png' })
await b.close()
