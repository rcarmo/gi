import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';
const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
const dialogFor = page => page.getByRole('dialog', { name: 'Gi Settings', exact: true });
async function source(info, id) {
  const scenario = loadCorpus().find(row => row.id === id); expect(scenario).toBeTruthy();
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
}
async function setup(page, request, info) {
  const token = `settings-shell-${info.project.name}-${Date.now()}`;
  const filename = `${token}.txt`;
  const write = await request.post('/api/tools/execute', { data: { tool: 'write', input: { path: filename, content: `Native workspace target ${token}` } } });
  expect(write.ok()).toBe(true); expect((await write.json()).error).toBeFalsy();
  const created = await request.post('/api/sessions', { data: { agent_id: token, title: token } }); expect(created.status()).toBe(201); const id = (await created.json()).id;
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), id); await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true }); await expect(input).toBeVisible(); await input.fill('frozen settings draft');
  if (!await page.locator('.workspace-sidebar').isVisible()) {
    await page.getByRole('button', { name: 'Menu', exact: true }).click(); await page.getByRole('menuitem', { name: 'Show workspace', exact: true }).click();
  }
  const row = page.locator(`.workspace-row[data-path="${filename}"] .workspace-label-text`);
  await row.scrollIntoViewIfNeeded(); await expect(row).toBeVisible();
  const box = await row.boundingBox(); expect(box).toBeTruthy(); const target = { x: box.x + box.width / 2, y: box.y + box.height / 2 };
  // Instrument actual trusted pointer delivery and native file reads; never dispatch a fake click.
  await row.evaluate(el => { window.__settingsFileClicks = 0; el.addEventListener('click', () => window.__settingsFileClicks++); });
  const reads = []; page.on('request', req => { const url = new URL(req.url()); if (url.pathname === '/api/workspace/file' && url.searchParams.get('path') === filename) reads.push(req.url()); });
  const dialog = dialogFor(page);
  const open = async () => { await input.focus(); await page.keyboard.press('Control+,'); await expect(dialog).toBeVisible(); };
  const unchanged = async () => {
    expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(id);
    await expect(input).toHaveValue('frozen settings draft');
    expect((await (await request.get(`/api/sessions/${id}/turns`)).json()).turns || []).toHaveLength(0);
  };
  return { id, filename, input, row, target, reads, dialog, open, unchanged };
}

for (const id of ['@ux-settings-layering-001', '@ux-settings-layering-002', '@ux-settings-layering-003', '@ux-settings-layering-004']) {
  test(`${id} Native Settings overlay blocks workspace interaction and restores it on dismissal`, async ({ page, request }, info) => {
    await source(info, id); const f = await setup(page, request, info); await f.open();
    const backdrop = page.locator('.settings-dialog-backdrop'), portal = page.locator('.settings-portal');
    await expect(backdrop).toBeVisible(); expect(await portal.evaluate(el => el.parentElement === document.body)).toBe(true);
    expect(await portal.evaluate(el => getComputedStyle(el).position)).toBe('fixed');
    expect(await backdrop.evaluate(el => getComputedStyle(el).backgroundColor)).toBe('rgba(0, 0, 0, 0.5)');
    const viewport = page.viewportSize(), cover = await backdrop.boundingBox(), box = await f.dialog.boundingBox();
    expect(cover).toEqual({ x: 0, y: 0, width: viewport.width, height: viewport.height });
    expect(box.x).toBeGreaterThanOrEqual(0); expect(box.y).toBeGreaterThanOrEqual(0);
    expect(box.x + box.width).toBeLessThanOrEqual(viewport.width); expect(box.y + box.height).toBeLessThanOrEqual(viewport.height);
    expect(await f.dialog.evaluate(el => { const b = el.getBoundingClientRect(); return el.contains(document.elementFromPoint(b.x + b.width / 2, b.y + b.height / 2)); })).toBe(true);
    expect(await page.evaluate(point => !!document.elementFromPoint(point.x, point.y)?.closest('.settings-portal'), f.target)).toBe(true);
    await page.mouse.click(f.target.x, f.target.y); await page.waitForTimeout(150);
    expect(await page.evaluate(() => window.__settingsFileClicks)).toBe(0); expect(f.reads).toEqual([]);
    // Clicking a backdrop may itself dismiss; clicks inside the dialog do not.
    if (await f.dialog.isVisible()) await page.keyboard.press('Escape');
    await expect(backdrop).toHaveCount(0); await expect(portal).toHaveCount(0); await expect(f.row).toBeVisible(); await f.unchanged();
    await f.row.click(); await expect.poll(() => page.evaluate(() => window.__settingsFileClicks)).toBe(1);
    await expect.poll(() => f.reads.length).toBeGreaterThan(0); await expect(page.locator('.workspace-preview-meta')).toContainText(f.filename);
    await f.unchanged();
  });
}

test('@ux-settings-dialog-001 Rapid native shortcut opens exactly one settings dialog', async ({ page, request }, info) => {
  await source(info, '@ux-settings-dialog-001'); const f = await setup(page, request, info);
  await f.input.focus(); for (let i = 0; i < 3; i++) await page.keyboard.press('Control+,');
  await expect(f.dialog).toHaveCount(1); await expect(f.dialog).toBeVisible(); await expect(page.locator('.settings-portal')).toHaveCount(1);
  await page.keyboard.press('Escape'); await expect(f.dialog).toHaveCount(0); await expect(f.input).toBeFocused(); await f.unchanged();
});

test('@ux-settings-dialog-003 Cold open renders immediate shell and resolves General first', async ({ page, request }, info) => {
  await source(info, '@ux-settings-dialog-003'); const f = await setup(page, request, info);
  let release, held = false; const gate = new Promise(resolve => { release = resolve; });
  await page.route('**/api/runtime/config', async route => { const response = await route.fetch(); held = true; await gate; await route.fulfill({ response }); });
  try {
    const start = Date.now(); await f.open(); await expect.poll(() => held).toBe(true);
    await expect(f.dialog.getByRole('status').filter({ hasText: 'Loading settings…' })).toBeVisible();
    await expect(f.dialog.getByRole('button', { name: 'General', exact: true })).toHaveAttribute('aria-current', 'page');
    expect(Date.now() - start).toBeLessThan(1000);
    release(); await expect(f.dialog.locator('.gi-settings-values')).toContainText('test-model', { timeout: 1500 });
    expect(Date.now() - start).toBeLessThan(2000);
    await expect(f.dialog.getByRole('heading', { name: 'General', exact: true })).toBeVisible();
    await expect(f.dialog.getByRole('heading', { name: 'Models', exact: true })).toHaveCount(0);
    await page.keyboard.press('Escape'); await f.unchanged();
  } finally { release(); }
});

test('@ux-settings-dialog-004 Numeric policy stepper accepts typed 128000 without saving', async ({ page, request }, info) => {
  await source(info, '@ux-settings-dialog-004'); const f = await setup(page, request, info); const mutations = [];
  page.on('request', req => { if (req.method() === 'PATCH') mutations.push(req.url()); });
  await f.open(); await f.dialog.getByRole('button', { name: 'Compaction', exact: true }).click();
  const field = f.dialog.getByRole('spinbutton', { name: 'Saved context window', exact: true });
  await expect(field).toBeVisible(); await field.focus(); await field.fill('128000'); await expect(field).toHaveValue('128000');
  expect(mutations).toEqual([]); await page.keyboard.press('Escape'); await f.unchanged();
});

test('@gi-settings-024 Missing upgrade chunk reports failure and explicit reload gets the current graph', async ({ page, request }, info) => {
  const bootstraps = [], mutations = [];
  page.on('response', response => { if (new URL(response.url()).pathname === '/dist/app.bundle.js') bootstraps.push(response); });
  page.on('request', req => { if (new URL(req.url()).pathname.startsWith('/api/settings/') && req.method() !== 'GET') mutations.push(req.url()); });
  const f = await setup(page, request, info); await f.open();
  // Use the native missing-file response to simulate an old tab's unloaded
  // module being removed by deployment. Never forge settings/backend data.
  const pattern = /\/dist\/chunks\/gi-settings-providers-[^/]+\.js$/;
  let missingStatus;
  await page.route(pattern, async route => {
    const response = await route.fetch({ url: new URL('/dist/chunks/gi-settings-providers-removed-build.js', page.url()).href });
    missingStatus = response.status(); await route.fulfill({ response });
  });
  await f.dialog.getByRole('button', { name: 'Providers', exact: true }).click();
  await expect(f.dialog.getByRole('alert')).toContainText('Unable to load Providers');
  await expect(f.dialog.getByRole('alert')).toContainText('reload the page');
  expect(missingStatus).toBe(404); expect(bootstraps).toHaveLength(1);
  await expect(page.locator('.app-shell')).toHaveCount(1);
  await page.keyboard.press('Escape'); await f.unchanged();
  await page.unroute(pattern); await page.reload(); await f.open();
  await f.dialog.getByRole('button', { name: 'Providers', exact: true }).click();
  await expect(f.dialog.getByRole('heading', { name: 'Providers', exact: true })).toBeVisible();
  await expect(f.dialog.getByRole('button', { name: 'Refresh providers', exact: true })).toBeEnabled();
  expect(bootstraps).toHaveLength(2);
  for (const response of bootstraps) expect(response.headers()['cache-control']).toBe('no-cache, no-store, must-revalidate');
  expect(mutations).toEqual([]); await page.keyboard.press('Escape'); await f.unchanged();
});

const settingsChunk = url => /\/dist\/chunks\/gi-settings-(models|appearance|compaction|providers|authentication)-[^/]+\.js$/.test(new URL(url).pathname);
for (const id of ['@ux-settings-dialog-005', '@ux-settings-003']) {
  test(`${id} General is immediate and pane modules load only on visit with cached revisits`, async ({ page, request }, info) => {
    await source(info, id);
    const chunks = [], appRequests = []; let runtimeReads = 0;
    page.on('request', r => { if (settingsChunk(r.url())) chunks.push(r.url()); if (/\/dist\/chunks\/app-[^/]+\.js$/.test(new URL(r.url()).pathname)) appRequests.push(r.url()); if (new URL(r.url()).pathname === '/api/runtime/config') runtimeReads++; });
    const f = await setup(page, request, info); await f.open();
    await expect(f.dialog.getByRole('heading', { name: 'General', exact: true })).toBeVisible();
    await expect(f.dialog.locator('.gi-settings-values')).toBeVisible(); expect(chunks).toEqual([]);
    const sequence = [['Models', 'models'], ['Appearance', 'appearance'], ['Compaction', 'compaction'], ['Providers', 'providers'], ['Authentication', 'authentication']];
    for (let i = 0; i < sequence.length; i++) {
      const [label, slug] = sequence[i]; let release, held = false; const gate = new Promise(r => { release = r; });
      const pattern = new RegExp(`/dist/chunks/gi-settings-${slug}-[^/]+\\.js$`);
      await page.route(pattern, async route => { const response = await route.fetch(); held = true; await gate; await route.fulfill({ response }); });
      try {
        await f.dialog.getByRole('button', { name: label, exact: true }).click(); await expect.poll(() => held).toBe(true);
        await expect(f.dialog.getByRole('status').filter({ hasText: `Loading ${label} pane…` })).toBeVisible();
        await expect(f.dialog.getByRole('heading', { name: label, exact: true })).toHaveCount(0);
        for (const [, unopened] of sequence.slice(i + 1)) expect(chunks.some(url => url.includes(`gi-settings-${unopened}-`))).toBe(false);
        release(); await expect(f.dialog.getByRole('heading', { name: label, exact: true })).toBeVisible();
      } finally { release(); await page.unroute(pattern); }
      const loaded = [...chunks]; await f.dialog.getByRole('button', { name: 'General', exact: true }).click();
      await f.dialog.getByRole('button', { name: label, exact: true }).click();
      await expect(f.dialog.getByRole('heading', { name: label, exact: true })).toBeVisible(); expect(chunks).toEqual(loaded);
    }
    expect(chunks).toHaveLength(sequence.length); expect(appRequests).toHaveLength(1); await expect(page.locator('.app-shell')).toHaveCount(1);
    expect(runtimeReads).toBeGreaterThan(0);
    await page.keyboard.press('Escape'); await f.unchanged();
  });
}

test('@gi-settings-024 Late imports and failed module loads cannot replace the selected pane or blank the shell', async ({ page, request }, info) => {
  const f = await setup(page, request, info); await f.open();
  let release, held = false, done; const gate = new Promise(r => { release = r; }), delivered = new Promise(r => { done = r; });
  const modelPattern = /\/dist\/chunks\/gi-settings-models-[^/]+\.js$/;
  await page.route(modelPattern, async route => { const response = await route.fetch(); held = true; await gate; await route.fulfill({ response }); done(); });
  try {
    await f.dialog.getByRole('button', { name: 'Models', exact: true }).click(); await expect.poll(() => held).toBe(true);
    await f.dialog.getByRole('button', { name: 'Appearance', exact: true }).click(); await expect(f.dialog.getByRole('heading', { name: 'Appearance', exact: true })).toBeVisible();
    release(); await delivered;
    await expect(f.dialog.getByRole('heading', { name: 'Models', exact: true })).toHaveCount(0);
    await expect(f.dialog.getByRole('heading', { name: 'Appearance', exact: true })).toBeVisible();
  } finally { release(); }
  await page.route(/\/dist\/chunks\/gi-settings-providers-[^/]+\.js$/, route => route.abort());
  await f.dialog.getByRole('button', { name: 'Providers', exact: true }).click(); await expect(f.dialog.getByRole('alert')).toContainText('Unable to load Providers');
  await expect(page.locator('.app-shell')).toHaveCount(1);
  await page.keyboard.press('Escape'); await expect(f.dialog).toHaveCount(0); await f.unchanged();
  // A failed ESM fetch may remain cached by the browser; no false retry promise or forced reload.
  await f.open(); await expect(f.dialog.getByRole('heading', { name: 'General', exact: true })).toBeVisible();
});

test('@gi-settings-024 A pane import completing after close cannot reopen Settings', async ({ page, request }, info) => {
  const f = await setup(page, request, info); await f.open();
  let release, held = false, done; const gate = new Promise(r => { release = r; }), delivered = new Promise(r => { done = r; });
  await page.route(/\/dist\/chunks\/gi-settings-appearance-[^/]+\.js$/, async route => {
    const response = await route.fetch(); held = true; await gate; await route.fulfill({ response }); done();
  });
  try {
    await f.dialog.getByRole('button', { name: 'Appearance', exact: true }).click(); await expect.poll(() => held).toBe(true);
    await page.keyboard.press('Escape'); await expect(f.dialog).toHaveCount(0); release(); await delivered;
    await page.waitForTimeout(150); await expect(f.dialog).toHaveCount(0); await f.unchanged();
    await f.open(); await f.dialog.getByRole('button', { name: 'Appearance', exact: true }).click();
    await expect(f.dialog.getByRole('heading', { name: 'Appearance', exact: true })).toBeVisible();
  } finally { release(); }
});

test('@ux-settings-004 Searchable header focuses and dialog width changes layout without resetting native state', async ({ page, request }, info) => {
  await source(info, '@ux-settings-004');
  const f = await setup(page, request, info); let reads = 0; const writes = [];
  page.on('request', req => { if (new URL(req.url()).pathname === `/api/sessions/${f.id}/model`) { if (req.method() === 'GET') reads++; else writes.push(req.method()); } });
  await f.open(); await f.dialog.getByRole('button', { name: 'Models', exact: true }).click();
  const filter = f.dialog.locator('header').getByRole('searchbox', { name: 'Filter models', exact: true });
  await expect(filter).toBeFocused(); await expect(filter).toHaveAttribute('placeholder', 'Filter models…');
  await expect(f.dialog.getByTestId('settings-current-model')).toBeVisible();
  const select = f.dialog.getByRole('combobox', { name: 'Session model', exact: true });
  await filter.fill('bootstrap'); await expect(select.locator('option:not([disabled])')).toHaveText(['test/bootstrap']);
  await select.selectOption('test/bootstrap'); await expect(f.dialog.getByRole('button', { name: 'Apply model', exact: true })).toBeEnabled();
  await filter.focus(); const readCount = reads; expect(readCount).toBeGreaterThan(0);
  // Change actual dialog geometry, not DOM classes or application state.
  await page.setViewportSize({ width: 1200, height: 950 });
  for (const [width, compact, narrow] of [[900,false,false],[860,true,false],[861,false,false],[720,true,true],[721,true,false],[640,true,true],[900,false,false]]) {
    await f.dialog.evaluate((element, width) => { element.style.width = `${width + (element.getBoundingClientRect().width - element.clientWidth)}px`; }, width);
    await expect.poll(() => f.dialog.evaluate(element => element.clientWidth)).toBe(width);
    await expect.poll(() => f.dialog.evaluate(element => [element.classList.contains('settings-dialog-compact'), element.classList.contains('settings-dialog-narrow')])).toEqual([compact,narrow]);
    await expect(filter).toBeVisible(); await expect(filter).toBeFocused(); await expect(filter).toHaveValue('bootstrap');
    await expect(select).toHaveValue('test/bootstrap'); await expect(select.locator('option:not([disabled])')).toHaveText(['test/bootstrap']);
    await expect(f.dialog.getByTestId('settings-current-model')).toContainText('test/test-model');
  }
  expect(reads).toBe(readCount); expect(writes).toEqual([]);
  await filter.fill('no-model-matches'); await expect(f.dialog.getByText('No matching models.', { exact: true })).toBeVisible();
  await f.dialog.getByRole('button', { name: 'General', exact: true }).click(); await expect(filter).toHaveCount(0);
  await f.dialog.getByRole('button', { name: 'Models', exact: true }).click();
  await expect(filter).toBeFocused(); await expect(filter).toHaveValue('');
  await expect(f.dialog.getByTestId('settings-current-model')).toBeVisible();
  await page.keyboard.press('Escape'); await f.unchanged();
});

test('@gi-settings-025 A late Apply from an earlier pane cannot enable the current header during another Apply', async ({ page, request }, info) => {
  const f = await setup(page, request, info); await f.open();
  await f.dialog.getByRole('button', { name: 'Models', exact: true }).click();
  const select = f.dialog.getByLabel('Session model', { exact: true });
  const filter = f.dialog.locator('header').getByLabel('Filter models', { exact: true });
  await expect(f.dialog.getByTestId('settings-current-model')).toHaveText('test/test-model');
  const releases = [], completed = []; let admitted = 0;
  await page.route(`**/api/sessions/${f.id}/model`, async route => {
    if (route.request().method() !== 'PATCH') return route.continue();
    const index = admitted++;
    const gate = new Promise(resolve => { releases[index] = resolve; });
    const response = await route.fetch(); completed[index] = true;
    await gate; await route.fulfill({ response });
  });
  try {
    await select.selectOption('test/bootstrap'); await f.dialog.getByRole('button', { name: 'Apply model', exact: true }).click();
    await expect.poll(() => completed[0]).toBe(true); await expect(filter).toBeDisabled();
    await f.dialog.getByRole('button', { name: 'General', exact: true }).click();
    await f.dialog.getByRole('button', { name: 'Models', exact: true }).click();
    await expect(filter).toBeFocused(); await expect(filter).toBeEnabled();
    await expect(f.dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap');
    await select.selectOption('test/test-model'); await f.dialog.getByRole('button', { name: 'Apply model', exact: true }).click();
    await expect.poll(() => completed[1]).toBe(true); await expect(filter).toBeDisabled();
    const delivered = page.waitForResponse(r => r.request().method() === 'PATCH' && new URL(r.url()).pathname === `/api/sessions/${f.id}/model`);
    releases[0](); await delivered; await page.waitForTimeout(100);
    await expect(filter).toBeDisabled(); await expect(f.dialog.getByRole('button', { name: 'Applying…' })).toBeDisabled();
    releases[1](); await expect(f.dialog.getByTestId('settings-current-model')).toHaveText('test/test-model'); await expect(filter).toBeEnabled();
    await page.keyboard.press('Escape'); await f.unchanged();
  } finally { releases.forEach(release => release()); }
});

test('@gi-settings-025 Window resize fallback keeps the header usable without ResizeObserver', async ({ page, request }, info) => {
  const f = await setup(page, request, info);
  // Disable only after the application has mounted its own unrelated observers.
  await page.evaluate(() => { window.ResizeObserver = undefined; });
  await f.open();
  await f.dialog.getByRole('button', { name: 'Models', exact: true }).click();
  const filter = f.dialog.locator('header').getByLabel('Filter models', { exact: true });
  await expect(filter).toBeFocused(); await filter.fill('bootstrap');
  for (const width of [1000, 800, 500, 1000]) {
    await page.setViewportSize({ width, height: 900 });
    await expect.poll(() => f.dialog.evaluate(element => {
      const width = element.clientWidth;
      return element.classList.contains('settings-dialog-compact') === (width > 0 && width <= 860) && element.classList.contains('settings-dialog-narrow') === (width > 0 && width <= 720);
    })).toBe(true);
    await expect(filter).toBeFocused(); await expect(filter).toHaveValue('bootstrap');
    await expect(f.dialog.getByLabel('Session model', { exact: true }).locator('option:not([disabled])')).toHaveText(['test/bootstrap']);
  }
  await page.keyboard.press('Escape'); await f.unchanged();
});

for(const id of ['@ux-settings-001','@ux-settings-dialog-002']) test(`${id} Native settings reopen shows cached General before the fresh read finishes`,async({page,request},info)=>{
 await source(info,id);const f=await setup(page,request,info);
 const bytes=Array.from(Buffer.from('settings cached reopen bytes'));
 await page.locator('.compose-box input[type=file]').setInputFiles({name:'settings-cache.txt',mimeType:'text/plain',buffer:Buffer.from(bytes)});
 const stored=()=>page.evaluate(async id=>{
  const db=await new Promise((resolve,reject)=>{const r=indexedDB.open('gi-session-drafts',1);r.onsuccess=()=>resolve(r.result);r.onerror=()=>reject(r.error);});
  return new Promise((resolve,reject)=>{const tx=db.transaction('drafts','readonly'),r=tx.objectStore('drafts').get(id);tx.oncomplete=()=>{db.close();const s=r.result;resolve(s?{...s,draft:{...s.draft,media:s.draft.media.map(f=>({...f,bytes:Array.from(new Uint8Array(f.bytes))}))}}:null);};tx.onerror=()=>reject(tx.error);});
 },f.id);
 await expect.poll(stored).toMatchObject({draft:{text:'frozen settings draft',media:[{name:'settings-cache.txt',bytes}]},pending:[]});const draft=await stored();
 const mutations=[];page.on('request',r=>{if(!['GET','HEAD'].includes(r.method())&&/^\/api\/(settings|sessions)/.test(new URL(r.url()).pathname))mutations.push(r.url());});
 const snapshot=await(await request.get('/api/runtime/config')).json();
 await page.getByTestId('hamburger').click();await page.getByRole('menuitem',{name:'Settings',exact:true}).click();await expect(f.dialog).toBeVisible();
 const nav=f.dialog.getByRole('navigation',{name:'Settings sections',exact:true});
 await expect(f.dialog.locator('.settings-dialog-header')).toBeVisible();await expect(f.dialog.locator('.settings-dialog-header')).toContainText('Gi Settings');await expect(nav).toBeVisible();
 await expect(nav.getByRole('button').first()).toHaveText('General');await expect(nav.getByRole('button',{name:'General',exact:true})).toHaveAttribute('aria-current','page');
 const values=f.dialog.locator('.gi-settings-values');await expect(values).toContainText(snapshot.assistant_name);await expect(values).toContainText(snapshot.workspace_root);await expect(values).toContainText(snapshot.current||snapshot.default_model);
 // Visit every supported navigation target. Pane-specific functionality and
 // unsupported sections do not acquire acceptance credit from this case.
 for(const section of ['Models','Appearance','Compaction','Providers','Authentication']){
  const button=nav.getByRole('button',{name:section,exact:true});await button.click();await expect(button).toHaveAttribute('aria-current','page');await expect(f.dialog.getByRole('heading',{name:section,exact:true})).toBeVisible();
 }
 await nav.getByRole('button',{name:'General',exact:true}).click();await expect(values).toContainText(snapshot.workspace_root);const cached=await values.textContent();
 await f.dialog.getByRole('button',{name:'Close settings',exact:true}).click();await expect(f.dialog).toHaveCount(0);await f.unchanged();expect(await stored()).toEqual(draft);
 let release,held=false,delivered;const gate=new Promise(r=>release=r),done=new Promise(r=>delivered=r);
 const pattern='**/api/runtime/config';
 await page.route(pattern,async route=>{const response=await route.fetch();expect(response.status()).toBe(200);held=true;await gate;await route.fulfill({response});delivered();});
 try{
  await f.input.focus();const start=Date.now();await page.keyboard.press('Control+,');await expect(f.dialog).toBeVisible({timeout:1000});await expect(values).toHaveText(cached,{timeout:1000});expect(Date.now()-start).toBeLessThan(1000);
  await expect(nav.getByRole('button',{name:'General',exact:true})).toHaveAttribute('aria-current','page');await expect.poll(()=>held).toBe(true);
  await expect(f.dialog.getByText('Loading settings…',{exact:true})).toHaveCount(0);await expect(f.dialog).toHaveCount(1);await expect(page.locator('.settings-portal')).toHaveCount(1);
  release();await done;await page.unroute(pattern);await expect(values).toHaveText(cached);await page.keyboard.press('Escape');await expect(f.input).toBeFocused();await f.unchanged();expect(await stored()).toEqual(draft);
 }finally{release();}
 expect(mutations).toEqual([]);await page.reload();await f.unchanged();await expect(page.locator('.compose-file-pill[title="settings-cache.txt"]')).toHaveCount(1);expect(await stored()).toEqual(draft);
});

test('@ux-settings-002 uncached General keeps the native loading shell and recovers a failed shared read',async({page,request},info)=>{
 await source(info,'@ux-settings-002');const f=await setup(page,request,info);
 const bytes=Buffer.from('settings loading retained β');await page.locator('.compose-box input[type=file]').setInputFiles({name:'loading-ref.txt',mimeType:'text/plain',buffer:bytes});
 const initial=(await(await request.get(`/api/sessions/${f.id}`)).json());const native=(await(await request.get('/api/runtime/config')).json());
 const mutations=[];page.on('request',r=>{if(!['GET','HEAD'].includes(r.method())&&/\/api\/(sessions|settings)/.test(new URL(r.url()).pathname))mutations.push(r.url());});
 let release;const gate=new Promise(r=>release=r);let held=false;
 await page.route('**/api/runtime/config',async route=>{const response=await route.fetch();expect(response.status()).toBe(200);held=true;await gate;await route.fulfill({response});},{times:1});
 const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true});
 try{
  const opened=Date.now();await page.keyboard.press('Control+,');await expect(dialog).toBeVisible();await expect.poll(()=>held).toBe(true);await expect(dialog.getByRole('status').filter({hasText:'Loading settings…'})).toBeVisible();expect(Date.now()-opened).toBeLessThan(1000);
  await expect(dialog.locator('.settings-dialog-header')).toBeVisible();await expect(dialog.getByRole('navigation',{name:'Settings sections'})).toBeVisible();await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');await expect(dialog.getByRole('heading',{name:'General',exact:true})).toBeVisible();await expect(dialog.locator('.gi-settings-values')).toHaveCount(0);
  release();await expect(dialog.locator('.gi-settings-values')).toBeVisible();await expect(dialog.getByText('Loading settings…',{exact:true})).toHaveCount(0);await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');
  const values=dialog.locator('.gi-settings-values');for(const value of [native.assistant_name,native.user_name,native.workspace_root,native.current||native.default_model])await expect(values).toContainText(value);
  await info.attach('settings-loaded',{body:await page.screenshot(),contentType:'image/png'});await page.keyboard.press('Escape');await expect(f.input).toHaveValue('frozen settings draft');
 }finally{release();await page.unrouteAll({behavior:'wait'});}
 // A fresh document clears the code-local General cache. Fail only the
 // explicit Settings read after bootstrap, then retry against the native API.
 await page.reload();await expect(f.input).toBeVisible();await page.route('**/api/runtime/config',route=>route.fulfill({status:503,contentType:'application/json',body:JSON.stringify({error:'fixture snapshot unavailable'})}),{times:1});
 await page.keyboard.press('Control+,');await expect(dialog).toBeVisible();await expect(dialog.getByRole('alert')).toContainText('fixture snapshot unavailable');await expect(dialog.locator('.settings-nav-item.active')).toHaveText('General');await expect(dialog.locator('.gi-settings-values')).toHaveCount(0);
 await dialog.getByRole('button',{name:'Retry',exact:true}).click();await expect(dialog.locator('.gi-settings-values')).toContainText(native.workspace_root);await expect(dialog.getByRole('alert')).toHaveCount(0);await page.keyboard.press('Escape');
 await expect(f.input).toHaveValue('frozen settings draft');await expect(page.locator('.compose-box')).toContainText('loading-ref.txt');
 const saved=await page.evaluate(id=>new Promise((resolve,reject)=>{
  const db=indexedDB.open('gi-session-drafts',1);db.onsuccess=()=>{
   const q=db.result.transaction('drafts').objectStore('drafts').get(id);
   q.onsuccess=()=>{try{const d=q.result.draft;const media=d.media.map(m=>({name:m.name,bytes:Array.from(new Uint8Array(m.bytes))}));db.result.close();resolve({text:d.text,media});}catch(error){reject(error);}};
   q.onerror=()=>reject(q.error);
  };db.onerror=()=>reject(db.error);
 }),f.id);
 expect(saved).toEqual({text:'frozen settings draft',media:[{name:'loading-ref.txt',bytes:[...bytes]}]});expect(mutations).toEqual([]);expect((await(await request.get(`/api/sessions/${f.id}`)).json())).toEqual(initial);expect((await(await request.get(`/api/sessions/${f.id}/turns`)).json()).turns||[]).toEqual([]);
});
