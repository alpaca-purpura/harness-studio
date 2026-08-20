const { chromium } = require('playwright');

const BASE = process.env.BASE_URL || 'http://localhost:4000';
const SHOTS = process.env.SHOTS_DIR || '/home/chalreme/Proyectos/prenter-harness/docs/product/screenshots';
const results = [];
function check(name, ok, detail = '') {
  results.push(ok);
  console.log(`${ok ? 'PASS' : 'FAIL'}  ${name}${detail ? ' — ' + detail : ''}`);
}

(async () => {
  const browser = await chromium.launch();
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await ctx.newPage();

  // 1. Board carga; el switcher muestra brands compuestas
  await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(2500);
  const opts = await page.locator('#brand-switcher option').allTextContents();
  check('switcher shows composite brands', opts.includes('demo-environment/client-acme'), opts.join(','));

  // 2. Primera brand carga sus stories (acme-001 en done)
  let ok = true;
  try {
    await page.locator('[title="acme-001"]').first().waitFor({ timeout: 15000 });
  } catch { ok = false; }
  check('board loads stories for composite brand', ok);
  await page.screenshot({ path: SHOTS + '/multi-board-acme.png' });

  // 3. Cambiar de proyecto vía switcher
  await page.locator('#brand-switcher').selectOption('demo-environment/internal-platform');
  ok = true;
  try {
    await page.locator('[title="plat-003"]').first().waitFor({ timeout: 15000 });
  } catch { ok = false; }
  check('switching project loads its stories', ok);
  await page.screenshot({ path: SHOTS + '/multi-board-platform.png' });

  // 4. Roadmap de brand compuesta (release shipped en Historial)
  await page.locator('#brand-switcher').selectOption('demo-environment/client-acme');
  await page.goto(BASE + '/roadmap', { waitUntil: 'domcontentloaded' });
  await page.locator('button:has-text("Historial")').first().waitFor({ timeout: 15000 });
  await page.locator('button:has-text("Historial")').first().click();
  ok = true;
  try {
    await page.locator('text=Auth foundation').first().waitFor({ timeout: 10000 });
  } catch { ok = false; }
  check('roadmap shipped release visible (composite)', ok);

  // 5. Transición de operador con brand compuesta (y revert)
  const resp = await page.request.post(BASE + '/api/transition', {
    data: { brand: 'demo-environment/client-xyz', storyId: 'xyz-002', targetState: 'refining' },
  });
  const data = await resp.json();
  check('operator transition with composite brand', data.ok === true, JSON.stringify(data));
  await page.request.post(BASE + '/api/transition', {
    data: { brand: 'demo-environment/client-xyz', storyId: 'xyz-002', targetState: 'idea' },
  });
  const st = await (await page.request.get(BASE + '/api/stories/xyz-002?brand=demo-environment%2Fclient-xyz')).json();
  check('transition reverted', st.story?.state === 'idea', `state=${st.story?.state}`);

  await browser.close();
  const passed = results.filter(Boolean).length;
  console.log(`\n${passed}/${results.length} multi-mode checks passed`);
  process.exit(passed === results.length ? 0 : 1);
})();
