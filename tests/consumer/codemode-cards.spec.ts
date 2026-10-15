import { test as shared, expect } from '../fixtures-vibes/profiled-fixtures';

// Gi-only consumer contracts: no shared catalogue ID, capability or skip is
// changed. Retain the real runtime lifecycle and browser profiling fixtures.
const test = shared.extend({ scenarioGate: [async ({}, use) => { await use(); }, { auto: true }] });
const cards = (page: any) => page.locator('.codemode-card');
async function prompt(page: any, id: string, code: string) {
  const text = `[tool:codemode ${JSON.stringify({code})}][after-tool:consumer settled] consumer codemode`;
  const response = await page.request.post(new URL(`/api/sessions/${id}/prompt`, page.url()).href, {data:{prompt:text}});
  expect(response.status()).toBe(202);
  return response.json();
}
async function opened(page: any, runtime: any) {
  const session = await runtime.newSession();
  await page.goto(session.url);
  await expect(page.locator('textarea').first()).toBeVisible();
  const enabled = await page.request.post(new URL(`/api/sessions/${session.id}/prompt`, page.url()).href, {data:{prompt:'/codemode on'}});
  expect(enabled.status()).toBe(202);
  // Commands are handled by the real runtime, not simulated in browser routes.
  await expect.poll(() => page.evaluate(() => document.readyState)).toBe('complete');
  return session;
}

test('Gi codemode: live nested call settles once and survives reload', async ({page,runtime}) => {
  test.setTimeout(60000);
  const session = await opened(page,runtime);
  await prompt(page,session.id,"const r = await tools.shell({command:'sleep 3; echo CODEMODE_LIVE_OK'}); return r;");
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('running');
  await expect(cards(page).locator('.codemode-card-calls')).toContainText('shell');
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('ok',{timeout:20000});
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-output')).toContainText('CODEMODE_LIVE_OK');
  await page.reload();
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('ok');
  await expect(cards(page).locator('.codemode-card-output')).toContainText('CODEMODE_LIVE_OK');
  await expect(cards(page).locator('.codemode-card-calls')).toContainText('shell');
  const overflow = await cards(page).evaluate(el => el.getBoundingClientRect().right > window.innerWidth + 1);
  expect(overflow).toBe(false);
});

test('Gi codemode: script failure survives reload without invented nested calls', async ({page,runtime}) => {
  const session = await opened(page,runtime);
  await prompt(page,session.id,"throw new Error('CODEMODE_EXPECTED_FAILURE');");
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('error',{timeout:20000});
  await expect(cards(page).locator('.codemode-card-calls li')).toHaveCount(0);
  await page.reload();
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('error');
  await expect(cards(page).locator('.codemode-card-calls li')).toHaveCount(0);
});

test('Gi codemode: other-session events never enter the selected timeline', async ({page,context,runtime}) => {
  const first = await opened(page,runtime);
  const secondPage = await context.newPage();
  try {
    const second = await opened(secondPage,runtime);
    await prompt(secondPage,second.id,"return 'CODEMODE_SESSION_TWO';");
    await expect(cards(secondPage).locator('.codemode-card-status')).toHaveText('ok',{timeout:20000});
    await expect(cards(page)).toHaveCount(0);
    await page.reload();
    await expect(cards(page)).toHaveCount(0);
    expect(first.id).not.toBe(second.id);
  } finally { await secondPage.close(); }
});

test('Gi codemode: interruption terminates the live card and remains terminal after reload', async ({page,runtime}) => {
  test.setTimeout(60000);
  const session = await opened(page,runtime);
  const receipt = await prompt(page,session.id,"await tools.shell({command:'sleep 30'});");
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('running');
  await expect(cards(page).locator('.codemode-card-calls')).toContainText('shell');
  const cancelled = await page.request.post(new URL(`/api/turns/${receipt.turn_id}/cancel`,page.url()).href);
  expect(cancelled.ok()).toBe(true);
  await expect(cards(page).locator('.codemode-card-status')).not.toHaveText('running',{timeout:20000});
  await expect(cards(page)).toHaveCount(1);
  const status = await cards(page).locator('.codemode-card-status').innerText();
  expect(['cancelled','error']).toContain(status);
  await page.reload();
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('cancelled');
});

test('Gi codemode: legacy saved result without details still renders output on reload', async ({page,runtime}) => {
  const session = await opened(page,runtime);
  await prompt(page,session.id,"return 'CODEMODE_LEGACY_OUTPUT';");
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('ok',{timeout:20000});
  // Compatibility fault injection only: remove details from real persisted HTTP
  // rows. This is not counted as evidence of live transport or execution.
  let removed = 0;
  await page.route('**/api/sessions/*/messages?*',async route => {
    const response = await route.fetch();
    const data = await response.json();
    for (const row of data.messages || []) {
      if (row.payload?.tool_name === 'codemode') { delete row.payload.details; removed++; }
    }
    await route.fulfill({response,json:data});
  });
  await page.reload();
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('ok');
  await expect(cards(page).locator('.codemode-card-output')).toContainText('CODEMODE_LEGACY_OUTPUT');
  await expect(cards(page).locator('.codemode-card-calls li')).toHaveCount(0);
  expect(removed).toBeGreaterThan(0);
});

test('Gi reset: in-place reset preserves visible history across reload', async ({page,runtime}) => {
  const session = await opened(page,runtime);
  await prompt(page,session.id,"return 'BEFORE_RESET_VISIBLE';");
  await expect(cards(page).locator('.codemode-card-status')).toHaveText('ok',{timeout:20000});
  await expect(page.getByText('consumer settled',{exact:true})).toBeVisible();
  const response=await page.request.post(new URL(`/api/sessions/${session.id}/reset`,page.url()).href,{data:{text:''}});
  expect(response.status()).toBe(200);
  await page.reload();
  await expect(cards(page)).toHaveCount(1);
  await expect(cards(page).locator('.codemode-card-output')).toContainText('BEFORE_RESET_VISIBLE');
  expect(page.url()).toContain(session.id);
});
