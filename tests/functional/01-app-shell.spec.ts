/**
 * 01-app-shell.spec.ts — Verify the app shell loads correctly.
 *
 * Tests the most fundamental requirement: the page loads, JS executes
 * without errors, the Preact app mounts, and the basic shell structure
 * is present.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, loadPageCollectingErrors, waitForAppShell } from './helpers';

test.describe('App shell', () => {
  test('serves a linked manifest with resolvable built-in PWA icons', async ({ page, request }) => {
    await page.goto('/');
    await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', '/manifest.json');
    const response = await request.get('/manifest.json');
    expect(response.status()).toBe(200);
    const manifest = await response.json();
    expect(manifest.name).toBeTruthy();
    expect(manifest.display).toBe('standalone');
    for (const size of [192, 512]) {
      const src = `/static/icon-${size}.png`;
      expect(manifest.icons).toEqual(expect.arrayContaining([
        expect.objectContaining({ src, sizes: `${size}x${size}`, type: 'image/png', purpose: 'any' }),
        expect.objectContaining({ src, sizes: `${size}x${size}`, type: 'image/png', purpose: 'maskable' }),
      ]));
      const image = await request.get(src);
      expect(image.status()).toBe(200);
      expect(image.headers()['content-type']).toContain('image/png');
      expect((await image.body()).subarray(0, 8)).toEqual(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]));
    }
  });

  test('Gi settings opens from the menu and shows scoped native model settings', async ({ page, request }) => {
    const chunks: string[] = [];
    page.on('request', req => { if (/\/dist\/chunks\/gi-settings-.*\.js$/.test(new URL(req.url()).pathname)) chunks.push(req.url()); });
    await page.goto(BASE_URL); await waitForAppShell(page);
    const input = page.getByRole('textbox', { name: 'Message (Enter to send, Shift+Enter for newline)...', exact: true });
    await input.fill('functional settings draft');
    await page.getByRole('button', { name: 'Menu', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Settings', exact: true }).click();
    const dialog = page.getByRole('dialog', { name: 'Gi Settings', exact: true });
    await expect(dialog.getByText('Active instance settings · read-only')).toBeVisible();
    await expect(dialog.locator('.gi-settings-values')).toContainText('test-model');
    const identity = await (await request.get('/api/settings/identity')).json();
    await expect(dialog.getByLabel('Assistant display name')).toHaveValue(identity.saved.assistant_name);
    await expect(dialog.getByRole('button', { name: 'Save names', exact: true })).toBeEnabled();
    expect(chunks).toEqual([]);
    await dialog.getByRole('button', { name: 'Models', exact: true }).click();
    const filter = dialog.locator('header').getByRole('searchbox', { name: 'Filter models', exact: true });
    await expect(filter).toBeFocused(); await expect(filter).toHaveAttribute('placeholder', 'Filter models…');
    await expect(dialog.getByTestId('settings-current-model')).toContainText('test-model');
    await filter.fill('no-matching-model'); await expect(dialog.getByText('No matching models.', { exact: true })).toBeVisible();
    await filter.fill('');
    const refresh = page.waitForResponse(r => r.url().includes('/model') && r.request().method() === 'GET');
    await dialog.getByRole('button', { name: 'Refresh models', exact: true }).click(); await refresh;
    await expect(dialog.getByRole('button', { name: 'Refresh models', exact: true })).toBeEnabled();
    expect(chunks.filter(url => url.includes('gi-settings-models-'))).toHaveLength(1);
    await dialog.getByRole('button', { name: 'Compaction', exact: true }).click();
    await expect(dialog.getByTestId('compaction-policy')).toContainText('Trigger threshold');
    const savedPolicy = await (await request.get('/api/settings/compaction')).json();
    await expect(dialog.getByLabel('Saved trigger threshold')).toHaveValue(String(savedPolicy.saved.policy.threshold_tokens));
    await expect(dialog.getByRole('button', { name: 'Save policy', exact: true })).toBeEnabled();
    await expect(dialog.getByRole('button', { name: 'Compact now', exact: true })).toBeVisible();
    await dialog.getByRole('button', { name: 'Providers', exact: true }).click();
    await expect(dialog.getByRole('region', { name: 'Provider openai', exact: true })).toBeVisible();
    await expect(dialog.getByText(/not verified remotely/)).toBeVisible();
    await expect(dialog.getByRole('region', { name: 'Provider anthropic', exact: true })).toBeVisible();
    await page.keyboard.press('Escape'); await expect(dialog).toHaveCount(0);
    await expect(input).toHaveValue('functional settings draft');
  });

  test('Gi appearance saves only a browser preference and restores it after reload', async ({ page }) => {
    await page.goto(BASE_URL); await waitForAppShell(page);
    await page.keyboard.press('Control+,');
    const dialog = page.getByRole('dialog', { name: 'Gi Settings', exact: true });
    await dialog.getByRole('button', { name: 'Appearance', exact: true }).click();
    await dialog.getByLabel('Theme preset').selectOption('monokai');
    await dialog.getByRole('button', { name: 'Save appearance' }).click();
    await expect(dialog.getByRole('status')).toContainText('saved in this browser');
    await page.reload(); await waitForAppShell(page);
    await expect(page.locator('html')).toHaveAttribute('data-color-theme', 'monokai');
    expect(await page.evaluate(() => JSON.parse(localStorage.getItem('gi_browser_appearance_v1') || 'null'))).toEqual({ version: 1, theme: 'monokai', tint: '' });
  });

  test('page loads without JS errors', async ({ page }) => {
    const errors = await loadPageCollectingErrors(page);
    expect(errors).toEqual([]);
  });

  test('app shell replaces loading placeholder', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await expect(page.locator('text=Loading Gi')).not.toBeVisible();
  });

  test('app-shell has correct CSS class structure', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    // Must have workspace-collapsed or workspace-open
    const shell = page.locator('.app-shell');
    const cls = await shell.getAttribute('class') || '';
    expect(cls).toContain('app-shell');
  });

  test('container element exists inside app shell', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await expect(page.locator('.app-shell .container')).toBeVisible();
  });

  test('no console errors during initial load', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await page.waitForTimeout(2000);
    const critical = consoleErrors.filter(e =>
      !e.includes('ResizeObserver') && !e.includes('favicon') && !e.includes('404') && !e.includes('ERR_CONNECTION_REFUSED')
    );
    expect(critical).toEqual([]);
  });

  test('all CSS stylesheets load', async ({ page }) => {
    const failed: string[] = [];
    page.on('requestfailed', (req) => {
      if (req.url().includes('.css')) failed.push(req.url());
    });
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    expect(failed).toEqual([]);
  });

  test('all JS bundles load', async ({ page }) => {
    const failed: string[] = [];
    page.on('requestfailed', (req) => {
      if (req.url().includes('.js')) failed.push(req.url());
    });
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    expect(failed).toEqual([]);
  });

  test('theme CSS variables are applied', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const bgPrimary = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--bg-primary').trim()
    );
    expect(bgPrimary).toBeTruthy();
  });

  test('favicon is served', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/favicon.ico`);
    expect(res.ok()).toBeTruthy();
    expect(res.headers()['content-type']).toContain('icon');
  });

  test('cache busters are present on bundle URLs', async ({ page }) => {
    await page.goto(BASE_URL);
    const html = await page.content();
    expect(html).toMatch(/app\.bundle\.js\?v=/);
    expect(html).toMatch(/app\.bundle\.css\?v=/);
    expect(html).toMatch(/favicon\.ico\?v=/);
  });
});

test('KaTeX renderer uses matching local CSS and fonts', async ({ page, request }) => {
  await page.goto('/');
  await waitForAppShell(page);
  await expect(page.locator('link[rel="stylesheet"][href^="/css/katex.min.css"]')).toHaveCount(1);
  const css = await request.get('/css/katex.min.css');
  expect(css.ok()).toBe(true);
  const urls = [...new Set([...((await css.text()).matchAll(/url\(([^)]+)\)/g))].map(match => match[1]))];
  expect(urls.length).toBeGreaterThan(0);
  for (const url of urls) {
    expect(url).toMatch(/^\/fonts\/katex\//);
    const font = await request.get(url);
    expect(font.ok()).toBe(true);
    expect((await font.body()).length).toBeGreaterThan(100);
  }
  const result = await page.evaluate(async () => {
    const el = document.createElement('div');
    document.body.appendChild(el);
    (window as any).katex.render('E = mc^2', el);
    await document.fonts.ready;
    const math = el.querySelector('.katex')!;
    const fonts = await document.fonts.load('16px KaTeX_Main');
    const result = {version:(window as any).katex.version, family:getComputedStyle(math).fontFamily, fonts:fonts.length};
    el.remove();
    return result;
  });
  expect(result.version).toBe('0.18.7');
  expect(result.family).toContain('KaTeX_Main');
  expect(result.fonts).toBeGreaterThan(0);
});

test('rapid reverse timeline swipe restores its originating session draft', async ({ page, request }) => {
  const create = async (name: string) => { const response = await request.post('/api/sessions', { data: { agent_id: `functional-swipe-${name}-${Date.now()}` } }); expect(response.status()).toBe(201); return (await response.json()).id; };
  const a = await create('a'), b = await create('b');
  await page.addInitScript(id => { localStorage.setItem('gi_session_id',id); Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'}); },a);
  await page.goto(BASE_URL); await waitForAppShell(page);
  const input = page.getByRole('textbox', { name: 'Message (Enter to send, Shift+Enter for newline)...', exact: true }); await input.fill('functional rapid draft');
  await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
  await expect(page.locator(`.compose-session-popup [data-session-jid="gi:${b}"]`)).toBeVisible(); await page.keyboard.press('Escape');
  const visits = await page.evaluate(async () => {
    const result: string[] = [];
    for(const delta of [-105,105]) {
      const el = document.querySelector('.timeline')!;
      for(const [name,x] of [['touchstart',190],['touchmove',190+delta],['touchend',190+delta]] as const) {
        const touch = { identifier: 1, target: el, clientX: x, clientY: 150 }, event = new Event(name,{bubbles:true,cancelable:true});
        Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[touch]});Object.defineProperty(event,'changedTouches',{value:[touch]});el.dispatchEvent(event);
      }
      result.push(localStorage.getItem('gi_session_id')!); await new Promise(requestAnimationFrame);
    }
    return result;
  });
  expect(visits).toEqual([b,a]); await expect(input).toHaveValue('functional rapid draft');
  for(const id of [a,b]) expect((await (await request.get(`/api/sessions/${id}/turns`)).json()).turns || []).toHaveLength(0);
});

test('native status-panel surface shares session swipe navigation without submitting the draft', async ({ page, request }) => {
  const create = async (name: string) => { const r = await request.post('/api/sessions', { data: { agent_id: `functional-status-${name}-${Date.now()}` } }); expect(r.status()).toBe(201); return (await r.json()).id; };
  const a = await create('a'), b = await create('b');
  expect((await request.post(`/api/sessions/${a}/prompt`, { data: { prompt: 'native status history', model: 'test-model' } })).status()).toBe(202);
  await expect.poll(async () => ((await (await request.get(`/api/sessions/${a}/turns`)).json()).turns || [])[0]?.status).toBe('completed');
  await page.addInitScript(id => { localStorage.setItem('gi_session_id',id); Object.defineProperty(navigator,'userAgent',{configurable:true,value:'iPhone Safari'}); },a);
  await page.goto(BASE_URL); await waitForAppShell(page);
  const input = page.locator('.compose-box textarea'); await input.fill('unsent status draft');
  await page.getByRole('button', { name: /Manage sessions for/ }).last().click(); await expect(page.locator(`.compose-session-popup [data-session-jid="gi:${b}"]`)).toBeVisible(); await page.keyboard.press('Escape');
  const panel = page.locator('.agent-status-panel'); await expect(panel).toBeVisible();
  await expect(page.locator('link[href^="/css/gi-status.css"]')).toHaveCount(1);
  await expect(panel).toHaveCSS('max-height', '40%'); await expect(panel).toHaveCSS('overflow-y','auto');
  // Measuring preview overflow must not add disclosure chrome for an idle
  // status that has no streamed thought/draft content.
  await expect(panel.getByRole('button', { name: /Show more|more lines/ })).toHaveCount(0);
  const size = await panel.boundingBox(), host = await page.locator('.container').boundingBox();
  expect(size!.height).toBeLessThanOrEqual(host!.height * 0.4 + 1);
  await panel.evaluate(el => { for(const [name,x] of [['touchstart',190],['touchmove',85],['touchend',85]] as const) { const point={identifier:1,target:el,clientX:x,clientY:150},event=new Event(name,{bubbles:true,cancelable:true});Object.defineProperty(event,'touches',{value:name==='touchend'?[]:[point]});Object.defineProperty(event,'changedTouches',{value:[point]});el.dispatchEvent(event); } });
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(b); await expect(input).toHaveValue('');
  expect((await (await request.get(`/api/sessions/${a}/turns`)).json()).turns || []).toHaveLength(1);
  expect((await (await request.get(`/api/sessions/${b}/turns`)).json()).turns || []).toHaveLength(0);
});

test('native timeline controls and modal fields receive their own keyboard events', async ({ page, request }) => {
  const create = await request.post('/api/sessions', {data:{agent_id:`functional-keys-${Date.now()}`}}); expect(create.status()).toBe(201); const {id} = await create.json();
  const sent = await request.post(`/api/sessions/${id}/prompt`, {data:{prompt:'functional native key target',model:'test-model'}}); expect(sent.status()).toBe(202);
  await expect.poll(async () => (await (await request.get(`/api/sessions/${id}/turns`)).json()).turns?.[0]?.status).toBe('completed');
  await page.addInitScript(id => localStorage.setItem('gi_session_id',id),id); await page.goto(BASE_URL); await waitForAppShell(page);
  const copy=page.locator('.post').first().getByRole('button',{name:'Copy message',exact:true});await copy.focus();
  await copy.evaluate(el=>{(window as any).__functionalKeys=[];el.addEventListener('keydown',e=>(window as any).__functionalKeys.push(e.key));});
  await copy.press('q');expect(await page.evaluate(()=>(window as any).__functionalKeys)).toEqual(['q']);await expect(page.locator('.timeline-quick-actions')).toHaveCount(0);
  await page.getByRole('button',{name:'Open model picker',exact:true}).click();const popup=page.getByRole('menu',{name:'Model picker',exact:true});await expect(popup).toBeVisible();
  const active=await popup.locator('.active').textContent();await page.keyboard.press('Control+,');const dialog=page.getByRole('dialog',{name:'Gi Settings',exact:true}),name=dialog.getByLabel('Assistant display name',{exact:true});await name.focus();
  await name.evaluate(el=>{(window as any).__functionalKeys=[];el.addEventListener('keydown',e=>(window as any).__functionalKeys.push(e.key));});
  await name.press('q');await name.press('ArrowDown');expect(await page.evaluate(()=>(window as any).__functionalKeys)).toEqual(['q','ArrowDown']);expect(await popup.locator('.active').textContent()).toBe(active);
  await page.keyboard.press('Escape');await expect(dialog).toHaveCount(0);await expect(popup).toBeVisible();await page.keyboard.press('Escape');await expect(popup).toHaveCount(0);
  expect((await (await request.get(`/api/sessions/${id}/turns`)).json()).turns).toHaveLength(1);
});

test('Quick Actions dismissal restores Conversation without submitting the draft',async({page,request})=>{
 const created=await request.post('/api/sessions',{data:{agent_id:`functional-dismiss-${Date.now()}`}});expect(created.status()).toBe(201);const {id}=await created.json();await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),id);
 await page.goto(BASE_URL);await waitForAppShell(page);const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true}),conversation=page.getByRole('region',{name:'Conversation',exact:true}),palette=page.locator('.timeline-quick-actions'),query=page.locator('.timeline-quick-actions-input');await input.fill('functional dismissal draft');
 await conversation.focus();await conversation.press('m');await expect(query).toBeFocused();await query.press('Escape');await expect(palette).toHaveCount(0);await expect(conversation).toBeFocused();
 for(const key of ['Enter','Space']){await conversation.press('m');await expect(query).toBeFocused();await query.press('Tab');const close=palette.getByRole('button',{name:'Close quick actions',exact:true});await expect(close).toBeFocused();await close.press(key);await expect(palette).toHaveCount(0);await expect(conversation).toBeFocused();await expect(input).toHaveValue('functional dismissal draft');}
 await page.keyboard.press('q');await expect(query).toHaveValue('q');await expect(query).toBeFocused();const b=await page.getByRole('button',{name:'Send message',exact:true}).boundingBox();await page.mouse.click(b!.x+b!.width/2,b!.y+b!.height/2);await expect(palette).toHaveCount(0);await expect(conversation).toBeFocused();await expect(input).toHaveValue('functional dismissal draft');expect((await(await request.get(`/api/sessions/${id}/turns`)).json()).turns||[]).toEqual([]);
});

test('session typeahead moves native focus from a substring to a prefix without sending',async({page,request})=>{
 const token=`functional-typeahead-${Date.now()}`;
 const create=async(title:string)=>{const r=await request.post('/api/sessions',{data:{agent_id:title,title}});expect(r.status()).toBe(201);return r.json();};
 const main=await create(token),substring=await create(`z-alpha-${token}`),prefix=await create(`alpha-${token}`);
 expect((await request.patch(`/api/sessions/${substring.id}`,{data:{action:'pin',pinned:true}})).status()).toBe(200);
 await page.addInitScript(id=>localStorage.setItem('gi_session_id',id),main.id);await page.goto(BASE_URL);await waitForAppShell(page);
 const input=page.getByRole('textbox',{name:'Message (Enter to send, Shift+Enter for newline)...',exact:true});await input.fill('functional typeahead draft');
 const trigger=page.getByRole('button',{name:/Manage sessions for/}).last();await trigger.click();const search=page.getByRole('searchbox',{name:'Search sessions',exact:true});await search.fill(token);
 const row=(id:string)=>page.locator(`.compose-session-popup [data-session-jid="gi:${id}"]`).getByRole('menuitem');await row(substring.id).focus();await row(substring.id).press('a');await expect(row(prefix.id)).toBeFocused();await expect(row(prefix.id)).toHaveClass(/active/);await expect(search).toHaveValue(token);
 await page.keyboard.press('Enter');await expect.poll(()=>page.evaluate(()=>localStorage.getItem('gi_session_id'))).toBe(prefix.id);await expect(input).toHaveValue('');await trigger.click();await row(main.id).click();await expect(input).toHaveValue('functional typeahead draft');expect((await(await request.get(`/api/sessions/${main.id}/turns`)).json()).turns||[]).toEqual([]);
});

test('Model search preserves native editing, empty-result cancellation and composer draft',async({page})=>{
 await page.goto('/');const input=page.locator('textarea').last();await expect(input).toBeVisible();await input.fill('model filter preserved draft');
 const trigger=page.getByRole('button',{name:'Open model picker',exact:true});await expect(trigger).toBeEnabled();await trigger.click();
 const search=page.getByRole('searchbox',{name:'Search models',exact:true});await search.fill('no such native model');
 await expect(page.getByRole('menu',{name:'Model picker'}).getByRole('menuitem')).toHaveCount(0);
 await search.press('Home');await search.press('X');await expect(search).toHaveValue('Xno such native model');await search.press('Enter');
 await expect(search).toBeVisible();await search.press('Escape');await expect(search).toHaveCount(0);await expect(trigger).toBeFocused();await expect(input).toHaveValue('model filter preserved draft');
 await trigger.click();await expect(search).toHaveValue('');
});
