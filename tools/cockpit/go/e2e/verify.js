const { chromium } = require('playwright');

const BASE = process.env.BASE_URL || 'http://localhost:4000';
const SHOTS = process.env.SHOTS_DIR || '/home/chalreme/Proyectos/prenter-harness/docs/product/screenshots';
const results = [];

function check(name, ok, detail = '') {
  results.push({ name, ok, detail });
  console.log(`${ok ? 'PASS' : 'FAIL'}  ${name}${detail ? ' — ' + detail : ''}`);
}

(async () => {
  // Warm dev-server route compilation so waits measure the app, not Turbopack
  for (const r of ['/board', '/roadmap', '/map']) {
    await fetch(BASE + r).catch(() => {});
  }

  const browser = await chromium.launch();

  async function brandPage(brand) {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
    await ctx.addInitScript((b) => localStorage.setItem('cockpit:brand', b), brand);
    return ctx.newPage();
  }

  // ── client-acme ──
  {
    const page = await brandPage('client-acme');
    await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
    const acmeCard = page.locator('[title="acme-001"]').first();
    let cardOk = true;
    try {
      await acmeCard.waitFor({ timeout: 20000 });
    } catch {
      cardOk = false;
    }
    check('acme board: card acme-001 rendered', cardOk);
    check('acme board: card example-001 rendered', (await page.locator('[title="example-001"]').count()) > 0);
    // is the acme-001 card inside the "done" column?
    const doneColHasCard =
      (await page
        .locator('section, div')
        .filter({ has: page.locator('[title="acme-001"]') })
        .filter({ hasText: 'done' })
        .count()) > 0;
    check('acme board: acme-001 sits in done column', doneColHasCard);
    await page.screenshot({ path: SHOTS + '/board-client-acme.png' });

    // story drawer
    if (cardOk) {
      await acmeCard.dispatchEvent('click');
      await page.waitForTimeout(1000);
      const drawer = await page.textContent('body');
      check('acme board: story drawer opens with OAuth2 spec', drawer.includes('OAuth2'));
      await page.screenshot({ path: SHOTS + '/story-drawer-client-acme.png' });
    }

    await page.goto(BASE + '/roadmap', { waitUntil: 'domcontentloaded' });
    await page.locator('button:has-text("Historial")').first().waitFor({ timeout: 20000 });
    await page.locator('button:has-text("Historial")').first().click();
    let relOk = true;
    try {
      await page.locator('text=Auth foundation').first().waitFor({ timeout: 15000 });
    } catch {
      relOk = false;
    }
    check('acme roadmap (Historial): shipped release "Auth foundation" visible', relOk);
    const rbody = await page.textContent('body');
    check('acme roadmap: shipped status visible', /shipped/i.test(rbody));
    await page.screenshot({ path: SHOTS + '/roadmap-client-acme.png' });

    await page.goto(BASE + '/map', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(3000);
    const mbody = await page.textContent('body');
    check('acme map: auth capability visible', /auth/i.test(mbody) && /service/i.test(mbody));
    await page.screenshot({ path: SHOTS + '/map-client-acme.png' });
    await page.context().close();
  }

  // ── client-xyz ──
  {
    const page = await brandPage('client-xyz');
    await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
    let ok = true;
    try {
      await page.locator('[title="xyz-002"]').first().waitFor({ timeout: 20000 });
    } catch {
      ok = false;
    }
    check('xyz board: card xyz-002 rendered', ok);
    await page.screenshot({ path: SHOTS + '/board-client-xyz.png' });

    await page.goto(BASE + '/roadmap', { waitUntil: 'domcontentloaded' });
    let relOk = true;
    try {
      await page.locator('text=Dashboard MVP').first().waitFor({ timeout: 20000 });
    } catch {
      relOk = false;
    }
    check('xyz roadmap: release "Dashboard MVP" visible', relOk);
    await page.screenshot({ path: SHOTS + '/roadmap-client-xyz.png' });
    await page.context().close();
  }

  // ── internal-platform ──
  {
    const page = await brandPage('internal-platform');
    await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
    let ok = true;
    try {
      await page.locator('[title="plat-003"]').first().waitFor({ timeout: 20000 });
    } catch {
      ok = false;
    }
    check('platform board: card plat-003 rendered', ok);
    await page.screenshot({ path: SHOTS + '/board-internal-platform.png' });
    await page.context().close();
  }

  // ── brand switcher via UI ──
  {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
    const page = await ctx.newPage();
    await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
    await page.waitForTimeout(2000);
    const sel = page.locator('#brand-switcher');
    check('brand switcher present', (await sel.count()) > 0);
    await sel.selectOption('client-xyz');
    let ok = true;
    try {
      await page.locator('[title="xyz-002"]').first().waitFor({ timeout: 15000 });
    } catch {
      ok = false;
    }
    check('brand switcher: select client-xyz loads its board', ok);
    await page.screenshot({ path: SHOTS + '/brand-switch-to-xyz.png' });
    await ctx.close();
  }

  // ── operator transition via UI: xyz-002 idea → refining → revert ──
  {
    const page = await brandPage('client-xyz');
    await page.goto(BASE + '/board', { waitUntil: 'domcontentloaded' });
    try {
      const card = page.locator('[title="xyz-002"]').first();
      await card.waitFor({ timeout: 15000 });
      await card.dispatchEvent('click');
      await page.waitForTimeout(1000);
      const refineBtn = page.locator('button:has-text("refinar")').first();
      if (await refineBtn.count()) {
        await refineBtn.click();
        await page.waitForTimeout(2000);
        const data = await (await page.request.get(BASE + '/api/stories/xyz-002?brand=client-xyz')).json();
        check('operator transition idea→refining via UI', data.story?.state === 'refining', `state=${data.story?.state}`);
        await page.request.post(BASE + '/api/transition', {
          data: { brand: 'client-xyz', storyId: 'xyz-002', targetState: 'idea' },
        });
        const data2 = await (await page.request.get(BASE + '/api/stories/xyz-002?brand=client-xyz')).json();
        check('transition reverted refining→idea', data2.story?.state === 'idea', `state=${data2.story?.state}`);
      } else {
        const drawerText = (await page.textContent('body')).slice(0, 0);
        check('operator transition idea→refining via UI', false, 'refine button not found');
        await page.screenshot({ path: '/tmp/pw-test/drawer-debug.png' });
      }
    } catch (e) {
      check('operator transition idea→refining via UI', false, e.message.slice(0, 100));
    }
    await page.context().close();
  }

  await browser.close();
  const failed = results.filter((r) => !r.ok);
  console.log(`\n${results.length - failed.length}/${results.length} checks passed`);
  process.exit(failed.length ? 1 : 0);
})();
