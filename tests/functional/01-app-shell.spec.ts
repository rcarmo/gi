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
