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
  return { filename, input, row, target, reads, dialog, open, unchanged };
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
