import { test, expect } from '@playwright/test';
import { readFileSync, mkdirSync } from 'node:fs';
const feature = readFileSync('tests/features/settings/gi-settings.feature', 'utf8');
const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
const dialogFor = page => page.getByRole('dialog', { name: 'Gi Settings', exact: true });
async function setup(page, request, info) {
  await info.attach('gi-gherkin', { body: feature, contentType: 'text/plain' });
  const create = async suffix => {
    const response = await request.post('/api/sessions', { data: { agent_id: `settings-${info.project.name}-${Date.now()}-${suffix}` } });
    expect(response.status()).toBe(201); return (await response.json()).id;
  };
  const a = await create('a'), b = await create('b');
  await page.addInitScript(id => { if (!localStorage.getItem('gi_session_id')) localStorage.setItem('gi_session_id', id); }, a);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  await expect(input).toBeVisible(); await input.fill('settings draft');
  const open = async () => { await page.keyboard.press('Control+,'); await expect(dialogFor(page)).toBeVisible(); };
  const models = async () => { await dialogFor(page).getByRole('button', { name: 'Models', exact: true }).click(); await expect(dialogFor(page).getByLabel('Session model', { exact: true })).toBeVisible(); };
  const switchTo = async id => {
    await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
    await page.locator(`[data-session-jid="gi:${id}"]`).getByRole('menuitem').click();
    await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(id);
  };
  return { a, b, input, open, models, switchTo, dialog: dialogFor(page) };
}

// Gi-derived cases intentionally do not carry @ux- tags or frozen parity credit.
test('@gi-settings-001 @gi-settings-002 Single scoped modal, responsive backdrop, focus and unchanged draft', async ({ page, request }, info) => {
  const { a, input, open, dialog } = await setup(page, request, info);
  if (!await page.locator('.workspace-sidebar').isVisible()) {
    await page.getByRole('button', { name: 'Menu', exact: true }).click();
    await page.getByRole('menuitem', { name: 'Show workspace', exact: true }).click();
  }
  await expect(page.locator('.workspace-sidebar')).toBeVisible();
  const beforeClass = await page.locator('.app-shell').getAttribute('class');
  await input.focus();
  await open(); await page.keyboard.press('Control+,'); await page.keyboard.press('Control+,');
  await expect(dialogFor(page)).toHaveCount(1);
  await expect(dialog.getByRole('navigation', { name: 'Settings sections' }).getByRole('button')).toHaveText(['General', 'Models']);
  await expect(dialog.getByText('Instance settings · read-only')).toBeVisible();
  await expect(dialog.getByText(/restart Gi/)).toBeVisible();
  await expect(dialog.locator('input, select')).toHaveCount(0);
  await expect(dialog.locator('.gi-settings-values')).toContainText('test-model');
  expect(await page.locator('#app').evaluate(el => el.inert)).toBe(true);
  expect(await page.locator('.settings-dialog-backdrop').evaluate(el => getComputedStyle(el).backgroundColor)).toBe('rgba(0, 0, 0, 0.5)');
  expect(await page.locator('.settings-portal').evaluate(el => el.parentElement === document.body)).toBe(true);
  const box = await dialog.boundingBox(), viewport = page.viewportSize();
  expect(box.x).toBeGreaterThanOrEqual(0); expect(box.y).toBeGreaterThanOrEqual(0);
  expect(box.x + box.width).toBeLessThanOrEqual(viewport.width); expect(box.y + box.height).toBeLessThanOrEqual(viewport.height);
  expect(await dialog.evaluate(el => { const box = el.getBoundingClientRect(); return el.contains(document.elementFromPoint(box.x + box.width / 2, box.y + box.height / 2)); })).toBe(true);
  await dialog.getByRole('button', { name: 'Close settings' }).focus(); await page.keyboard.press('Shift+Tab');
  await expect(dialog.getByRole('button', { name: 'Models', exact: true })).toBeFocused();
  await page.keyboard.press('Tab'); await expect(dialog.getByRole('button', { name: 'Close settings' })).toBeFocused();
  await page.keyboard.press('Escape'); await expect(dialog).toHaveCount(0); await expect(input).toBeFocused();
  expect(await page.locator('#app').evaluate(el => el.inert)).toBe(false);
  await expect(input).toHaveValue('settings draft'); expect(await page.locator('.app-shell').getAttribute('class')).toBe(beforeClass);
  await page.getByRole('button', { name: 'Menu', exact: true }).click();
  await page.getByRole('menuitem', { name: 'Settings', exact: true }).click();
  await expect(dialog).toBeVisible(); await dialog.getByRole('button', { name: 'Close settings' }).click();
  await open(); await page.locator('.settings-dialog-backdrop').click({ position: { x: 2, y: 2 } }); await expect(dialog).toHaveCount(0);
  expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(a);
});

test('@gi-settings-003 Loading shell, native read failure and retry', async ({ page, request }, info) => {
  const { open, dialog } = await setup(page, request, info);
  let release, entered = false; const gate = new Promise(resolve => { release = resolve; });
  await page.route('**/api/runtime/config', async route => { entered = true; await gate; await route.abort(); });
  try {
    await open(); await expect.poll(() => entered).toBe(true); await expect(dialog.getByRole('status')).toHaveText('Loading settings…');
    release(); await expect(dialog.getByRole('alert')).toBeVisible();
    await page.unroute('**/api/runtime/config'); await dialog.getByRole('button', { name: 'Retry' }).click();
    await expect(dialog.locator('.gi-settings-values')).toContainText('test-model');
    await page.keyboard.press('Escape');
    let releaseAgain; const holdAgain = new Promise(resolve => { releaseAgain = resolve; });
    await page.route('**/api/runtime/config', async route => { const response = await route.fetch(); await holdAgain; await route.fulfill({ response }); });
    try { await open(); await expect(dialog.locator('.gi-settings-values')).toContainText('test-model'); } finally { releaseAgain(); }
  } finally { release(); }
});

test('@gi-settings-004 @gi-settings-005 @gi-settings-006 Native model confirmation, failure, reload and draft isolation', async ({ page, request }, info) => {
  const { a, b, input, open, models, dialog } = await setup(page, request, info);
  await page.locator('.compose-box input[type=file]').setInputFiles({ name: 'settings.txt', mimeType: 'text/plain', buffer: Buffer.from('settings draft file') });
  const runtime = await (await request.get('/api/runtime/config')).json();
  await open(); await models();
  const select = dialog.getByLabel('Session model', { exact: true });
  await expect(dialog.getByLabel('Filter models', { exact: true })).toBeFocused();
  await expect(dialog).toContainText(`gi:${a}`); await expect(dialog).toContainText('Thinking (read-only)');
  const modelState = await (await request.get(`/api/sessions/${a}/model`)).json();
  await expect(dialog.getByTestId('settings-context-capacity')).toHaveText(modelState.context_window > 0 ? String(modelState.context_window) : 'Unknown');
  await dialog.getByLabel('Filter models', { exact: true }).fill('bootstrap');
  await expect(select.locator('option')).toHaveCount(2); await select.selectOption('test/bootstrap');
  expect((await (await request.get(`/api/sessions/${a}/model`)).json()).current).toBe('test/test-model');
  let release, entered = false; const gate = new Promise(resolve => { release = resolve; });
  await page.route(`**/api/sessions/${a}/model`, async route => {
    if (route.request().method() !== 'PATCH') return route.continue();
    const response = await route.fetch(); entered = true; await gate; await route.fulfill({ response });
  });
  try {
    await dialog.getByRole('button', { name: 'Apply model' }).click(); await expect.poll(() => entered).toBe(true);
    await expect(dialog.getByRole('button', { name: 'Applying…' })).toBeDisabled();
    await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/test-model');
    release(); await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap');
  } finally { release(); }
  await page.unroute(`**/api/sessions/${a}/model`);
  await dialog.getByLabel('Filter models', { exact: true }).fill(''); await select.selectOption('test/unavailable-model');
  await dialog.getByRole('button', { name: 'Apply model' }).click();
  await expect(dialog.getByRole('alert')).toContainText('unavailable or lacks credentials');
  await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap'); await expect(dialog.getByRole('button', { name: 'Apply model' })).toBeEnabled();
  await page.keyboard.press('Escape');
  await expect(page.getByRole('button', { name: 'Open model picker', exact: true })).toHaveText('test/bootstrap');
  await expect(input).toHaveValue('settings draft'); await expect(page.locator('.compose-file-pill[title="settings.txt"]')).toBeVisible();
  await page.reload(); await expect(input).toHaveValue('settings draft');
  await expect(page.locator('.compose-file-pill[title="settings.txt"]')).toBeVisible();
  await open(); await models(); await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap');
  expect((await (await request.get(`/api/sessions/${b}/model`)).json()).current).toBe('test/test-model');
  expect((await (await request.get('/api/runtime/config')).json()).current).toBe(runtime.current);
  expect((await (await request.get(`/api/sessions/${a}/turns`)).json()).turns || []).toHaveLength(0);
});

test('@gi-settings-007 A closed model write cannot overwrite another session settings or draft', async ({ page, request }, info) => {
  const { a, b, input, open, models, switchTo, dialog } = await setup(page, request, info);
  await open(); await models(); await dialog.getByLabel('Session model', { exact: true }).selectOption('test/bootstrap');
  let release, entered = false, done; const gate = new Promise(resolve => { release = resolve; }), delivered = new Promise(resolve => { done = resolve; });
  await page.route(`**/api/sessions/${a}/model`, async route => {
    if (route.request().method() !== 'PATCH') return route.continue();
    const response = await route.fetch(); entered = true; await gate; await route.fulfill({ response }); done();
  });
  try {
    await dialog.getByRole('button', { name: 'Apply model' }).click(); await expect.poll(() => entered).toBe(true);
    await page.keyboard.press('Escape'); await switchTo(b); await input.fill('other settings draft');
    await open(); await models(); release(); await delivered;
    await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/test-model'); await expect(dialog.getByRole('alert')).toHaveCount(0);
    if (process.env.GI_SETTINGS_CAPTURE && info.project.name.startsWith('chromium-')) {
      mkdirSync('test-results/gi-settings-captures', { recursive: true });
      await page.screenshot({ path: `test-results/gi-settings-captures/${info.project.name}.png` });
    }
    await page.keyboard.press('Escape'); await expect(input).toHaveValue('other settings draft');
    expect((await (await request.get(`/api/sessions/${a}/model`)).json()).current).toBe('test/bootstrap');
    await switchTo(a); await expect(input).toHaveValue('settings draft');
  } finally { release(); }
});

test('@gi-settings-004 @gi-settings-007 A held catalogue loads on demand and cannot replace a reopened session', async ({ page, request }, info) => {
  const { a, b, input, open, switchTo, dialog } = await setup(page, request, info);
  // Seed a distinguishable native model in B, not a synthetic catalogue response.
  expect((await request.patch(`/api/sessions/${b}/model`, { data: { model: 'test/bootstrap' } })).status()).toBe(200);
  let release, entered = false, done; const gate = new Promise(resolve => { release = resolve; }), delivered = new Promise(resolve => { done = resolve; });
  await open(); await expect(dialog.locator('.gi-settings-values')).toBeVisible();
  await page.route(`**/api/sessions/${a}/model`, async route => {
    if (route.request().method() !== 'GET' || entered) return route.continue();
    const response = await route.fetch(); entered = true; await gate; await route.fulfill({ response }); done();
  });
  try {
    await dialog.getByRole('button', { name: 'Models', exact: true }).click();
    await expect(dialog.getByRole('status')).toHaveText('Loading models…'); await expect.poll(() => entered).toBe(true);
    await page.keyboard.press('Escape'); await switchTo(b); await input.fill('B catalogue draft');
    await open(); await dialog.getByRole('button', { name: 'Models', exact: true }).click();
    await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap');
    release(); await delivered;
    await expect(dialog.getByTestId('settings-current-model')).toHaveText('test/bootstrap');
    await expect(dialog).toContainText(`gi:${b}`); await page.keyboard.press('Escape'); await expect(input).toHaveValue('B catalogue draft');
  } finally { release(); }
});

test('@gi-settings-001 Modal keyboard input cannot activate a background model picker', async ({ page, request }, info) => {
  const { a, input, open, dialog } = await setup(page, request, info);
  await page.getByRole('button', { name: 'Open model picker', exact: true }).click();
  await expect(page.getByRole('menu', { name: 'Model picker', exact: true })).toBeVisible();
  await open();
  await dialog.getByRole('button', { name: 'Close settings' }).focus();
  await page.keyboard.press('ArrowDown'); await page.keyboard.press('Enter');
  await expect(dialog).toHaveCount(0);
  expect((await (await request.get(`/api/sessions/${a}/model`)).json()).current).toBe('test/test-model');
  await page.keyboard.press('Escape'); await expect(input).toHaveValue('settings draft');
});
