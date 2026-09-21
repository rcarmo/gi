import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';

const cases = loadCorpus().filter(scenario => ['@ux-original-001', '@ux-original-002'].includes(scenario.id));

for (const scenario of cases) {
  test(`${scenario.id} ${scenario.name}`, async ({ page, request }, testInfo) => {
    await testInfo.attach('gherkin', { body: `${scenario.uri}:${scenario.line}\n${scenario.steps.join('\n')}`, contentType: 'text/plain' });
    // Seed only through Gi's native APIs in the dedicated parity process.
    const create = async (agent) => {
      const response = await request.post('/api/sessions', { data: { title: `@${agent}`, agent_id: agent } });
      expect(response.status()).toBe(201);
      return response.json();
    };
    const main = await create('web');
    const research = await create('research');
    expect(main.id).not.toBe(research.id);
    await page.addInitScript(id => localStorage.setItem('gi_session_id', id), main.id);
    await page.goto('/');
    const input = page.getByRole('textbox', { name: 'Message (Enter to send, Shift+Enter for newline)...', exact: true });
    await expect(input).toBeVisible();
    const submissions = [];
    page.on('request', req => {
      if (req.method() === 'POST' && /\/prompt$/.test(new URL(req.url()).pathname)) submissions.push(req.postData());
    });
    const trigger = page.getByTestId('hamburger');
    const menu = page.locator('.timeline-menu-dropdown[role="menu"]');
    if (scenario.id === '@ux-original-001') {
      await expect(menu).toHaveCount(0);
      await trigger.click();
      await expect(menu).toBeVisible();
      await expect(menu).toHaveCount(1);
      await page.keyboard.press('Escape');
      await expect(menu).toHaveCount(0);
      await trigger.focus();
      await page.keyboard.press('Enter');
      await expect(menu).toBeVisible();
      // Click visible inert timeline content, never force a hidden control.
      const timeline = page.locator('.timeline');
      const bounds = await timeline.boundingBox();
      expect(bounds).toBeTruthy();
      await timeline.click({ position: { x: bounds.width - 12, y: bounds.height - 12 } });
      await expect(menu).toHaveCount(0);
    } else {
      const draft = 'Canonical unsent draft: preserve on workspace toggle.';
      await input.fill(draft);
      const sidebar = page.locator('.workspace-sidebar');
      const wasVisible = await sidebar.isVisible();
      await trigger.click();
      await menu.getByRole('menuitem', { name: wasVisible ? 'Hide workspace' : 'Show workspace', exact: true }).click();
      await expect(sidebar).toBeVisible({ visible: !wasVisible });
      await expect(menu).toHaveCount(0);
      await expect(input).toHaveValue(draft);
      await trigger.click();
      await menu.getByRole('menuitem', { name: wasVisible ? 'Show workspace' : 'Hide workspace', exact: true }).click();
      await expect(sidebar).toBeVisible({ visible: wasVisible });
      await expect(menu).toHaveCount(0);
      await expect(input).toHaveValue(draft);
    }
    expect(submissions).toEqual([]);
    const stored = await request.get(`/api/sessions/${main.id}/messages`);
    expect(stored.ok()).toBeTruthy();
    // Gi serializes an empty Go slice as null; both forms mean no stored rows.
    expect((await stored.json()).messages ?? []).toEqual([]);
    expect(await page.evaluate(() => localStorage.getItem('gi_session_id'))).toBe(main.id);
    await testInfo.attach('checkpoint', { body: await page.screenshot(), contentType: 'image/png' });
  });
}
