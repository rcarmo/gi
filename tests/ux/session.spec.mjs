import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';

const inputName = 'Message (Enter to send, Shift+Enter for newline)...';
const scenario = loadCorpus().find(row => row.id === '@ux-original-014');

test('@ux-original-013 Open the session picker from pointer or keyboard and close it with Escape', async ({ page, request }, info) => {
  const scenario = loadCorpus().find(row => row.id === '@ux-original-013');
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const create = async agent => {
    const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
    expect(response.status()).toBe(201);
    return response.json();
  };
  const suffix = info.project.name;
  const main = await create(`picker-main-${suffix}`), other = await create(`picker-other-${suffix}`);
  const fork = async (parent, name) => {
    const response = await request.post(`/api/sessions/${parent}/fork`, { data: { title: name, agent_id: name } });
    expect(response.status()).toBe(201);
    return (await response.json()).branch.chat_jid.slice(3);
  };
  const child = await fork(main.id, `picker-child-${suffix}`);
  await fork(child, `picker-leaf-${suffix}`);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), child);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  const triggers = page.getByRole('button', { name: /Manage sessions for/ });
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  const search = page.getByRole('searchbox', { name: 'Search sessions', exact: true });
  await expect(triggers.last()).toBeVisible();
  await input.fill('Do not submit this draft');

  // Both visible pointer triggers must restore the specific opening button.
  for (const trigger of [triggers.first(), triggers.last()]) {
    await trigger.click();
    await expect(search).toBeFocused();
    await expect(popup.getByRole('group', { name: 'Current', exact: true }).getByRole('menuitem')).toContainText('picker-child');
    const tree = popup.getByRole('group', { name: 'This session tree', exact: true });
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-main' })).toBeVisible();
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-leaf' })).toBeVisible();
    await expect(popup.getByRole('group', { name: 'Other sessions', exact: true }).getByRole('menuitem').filter({ hasText: `gi:${other.id}` })).toBeVisible();
    await search.fill(`GI:${other.id}`); // Case-insensitive JID search.
    await expect(popup.getByRole('menuitem')).toHaveCount(1);
    await expect(popup.getByRole('menuitem')).toContainText('picker-other');
    await search.fill(`picker-leaf-${suffix}`); // Matching descendants keep their ancestors.
    await expect(tree.getByRole('menuitem').filter({ hasText: 'picker-leaf' })).toBeVisible();
    await expect(popup.getByRole('menuitem')).toHaveCount(3);
    await search.fill('no-session-matches-this-query');
    await expect(popup.getByRole('status')).toHaveText('No sessions match your search.');
    await page.keyboard.press('Enter');
    await expect(search).toBeFocused();
    await page.keyboard.press('Escape');
    await expect(search).toHaveCount(0);
    await expect(trigger).toBeFocused();
    await expect(input).toHaveValue('Do not submit this draft');
    await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  }

  // Native keyboard activation, not a programmatic click.
  await page.keyboard.press('Enter');
  await expect(search).toBeFocused();
  await expect(search).toHaveValue('');
  await page.keyboard.press('Escape');
  await expect(triggers.last()).toBeFocused();
  await input.fill('');
  await page.keyboard.press('@');
  await expect(search).toBeFocused();
  await expect(input).toHaveValue('');
  await page.keyboard.press('Escape');
  await expect(triggers.first()).toBeFocused();
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  const messages = await (await request.get(`/api/sessions/${child}/messages`)).json();
  expect(messages.messages || []).toEqual([]);
  await info.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
});

test('Gi filtered picker keeps native text editing, Tab and keyboard selection', async ({ page, request }, info) => {
  const rootName = `keyboard-root-${info.project.name}`, leafName = `keyboard-leaf-${info.project.name}`;
  const response = await request.post('/api/sessions', { data: { title: `@${rootName}`, agent_id: rootName } });
  const main = await response.json();
  const fork = await (await request.post(`/api/sessions/${main.id}/fork`, { data: { title: leafName, agent_id: leafName } })).json();
  const child = fork.branch.chat_jid.slice(3);
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  const search = page.getByRole('searchbox', { name: 'Search sessions', exact: true });
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  await trigger.click();
  await search.pressSequentially(leafName);
  await expect(search).toHaveValue(leafName);
  // Home/End and printable keys edit the search rather than jumping rows.
  await page.keyboard.press('Home');
  await expect.poll(() => search.evaluate(el => el.selectionStart)).toBe(0);
  await page.keyboard.press('End');
  await expect.poll(() => search.evaluate(el => el.selectionStart)).toBe(leafName.length);
  await expect(popup.locator('.active')).toContainText('keyboard-leaf');
  await page.keyboard.press('ArrowUp');
  await expect(popup.locator('.active')).toContainText('keyboard-root');
  await page.keyboard.press('ArrowDown');
  await expect(popup.locator('.active')).toContainText('keyboard-leaf');
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await expect(search).toHaveCount(0);

  await trigger.click();
  await search.fill(rootName);
  const root = popup.getByRole('menuitem').filter({ hasText: `gi:${main.id}` });
  await expect(popup.getByRole('menuitem')).toHaveCount(1);
  await expect(root).toBeVisible();
  await expect(search).toBeFocused();
  await page.keyboard.press('Tab');
  await expect(root).toBeFocused();
  // Tab must only move focus, never switch a chat.
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
});

test('@ux-original-014 Select another session through the picker', async ({ page, request }, info) => {
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const create = async agent => {
    const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
    expect(response.status()).toBe(201);
    return response.json();
  };
  const main = await create('web'), research = await create('research');
  const mainText = `Main history ${main.id}`, researchText = `Research history ${research.id}`;
  for (const [session, text] of [[main, mainText], [research, researchText]]) {
    const response = await request.post(`/api/sessions/${session.id}/prompt`, { data: { prompt: text, model: 'test-model' } });
    expect(response.status()).toBe(202);
    await expect.poll(async () => {
      const stored = await (await request.get(`/api/sessions/${session.id}/messages`)).json();
      return (stored.messages || []).some(message => message.role === 'assistant' && message.content === `Gi received: ${text}`);
    }).toBe(true);
  }
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  const input = page.getByRole('textbox', { name: inputName, exact: true });
  const trigger = page.getByRole('button', { name: /Manage sessions for/ }).last();
  const popup = page.getByRole('menu', { name: 'Sessions and agents', exact: true });
  await expect(input).toBeVisible();
  await expect(trigger).toBeVisible();
  await input.fill('Main unsent draft');

  // Hold a real main-session response, not a fabricated payload. The action to
  // switch sessions still goes through the visible native picker.
  let release;
  const gate = new Promise(resolve => { release = resolve; });
  let held = false;
  let delivered;
  const deliveredResponse = new Promise(resolve => { delivered = resolve; });
  await page.route(`**/api/sessions/${main.id}/messages*`, async route => {
    const response = await route.fetch();
    held = true;
    await gate;
    await route.fulfill({ response });
    delivered();
  });
  // Gi's native refresh interval will start the captured request.
  await expect.poll(() => held, { timeout: 15000 }).toBe(true);
  const paths = [];
  page.on('request', req => paths.push(new URL(req.url()).pathname));
  await trigger.click();
  const target = popup.getByRole('menuitem', { name: /research/ }).first();
  await expect(target).toBeVisible();
  await target.focus();
  await page.keyboard.press('Enter');
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(research.id);
  await expect(input).toHaveValue('');
  await input.fill('Research unsent draft');
  await expect.poll(() => paths.includes(`/api/sessions/${research.id}/messages`)).toBe(true);
  await expect.poll(() => paths.includes(`/api/sessions/${research.id}/turns`)).toBe(true);
  await expect(page.locator('.post-content').filter({ hasText: researchText }).first()).toBeVisible();
  release();
  await deliveredResponse;
  await page.unroute(`**/api/sessions/${main.id}/messages*`);
  await expect(input).toHaveValue('Research unsent draft');
  await expect(trigger).toHaveAttribute('aria-label', 'Manage sessions for @research');
  await expect(page.locator('.post-content').filter({ hasText: mainText })).toHaveCount(0);
  await expect(page.locator('.post-content').filter({ hasText: researchText }).first()).toBeVisible();
  await trigger.click();
  await popup.getByRole('menuitem', { name: /web/ }).first().click();
  await expect(input).toHaveValue('Main unsent draft');
  await trigger.click();
  await popup.getByRole('menuitem', { name: /research/ }).first().click();
  await expect(input).toHaveValue('Research unsent draft');
  const messages = await request.get(`/api/sessions/${research.id}/messages`);
  expect((await messages.json()).messages.map(message => message.content)).toEqual([researchText, `Gi received: ${researchText}`]);
  await info.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
});

test('Gi new-session action allocates a distinct child chat', async ({ page, request }) => {
  const response = await request.post('/api/sessions', { data: { title: '@web', agent_id: 'web' } });
  const main = await response.json();
  await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
  await page.goto('/');
  await page.getByRole('button', { name: /Manage sessions for/ }).last().click();
  const created = page.waitForResponse(res => res.url().endsWith(`/api/sessions/${main.id}/fork`) && res.request().method() === 'POST');
  await page.getByRole('button', { name: 'New', exact: true }).click();
  const fork = await (await created).json();
  const child = fork.branch.chat_jid.slice(3);
  expect(child).not.toBe(main.id);
  await expect.poll(() => page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(child);
  const stored = await (await request.get(`/api/sessions/${child}`)).json();
  expect(stored.parent_session_id).toBe(main.id);
});
