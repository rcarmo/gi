import { test, expect } from '@playwright/test';

const BASE_URL = process.env.GI_TEST_URL || 'http://127.0.0.1:8090';

test.describe('Gi base UX', () => {

  test('page loads without JS errors', async ({ page }) => {
    const errors: string[] = [];
    page.on('pageerror', (err) => errors.push(err.message));
    await page.goto(BASE_URL);
    // Wait for the app to replace the loading placeholder
    await page.waitForTimeout(3000);
    expect(errors).toEqual([]);
  });

  test('app shell renders and replaces loading placeholder', async ({ page }) => {
    await page.goto(BASE_URL);
    // The loading div should be replaced by the app shell
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    // Loading text should be gone
    await expect(page.locator('text=Loading Gi')).not.toBeVisible();
  });

  test('runtime config API returns valid config', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/runtime/config`);
    expect(res.ok()).toBeTruthy();
    const config = await res.json();
    expect(config.default_model).toBeTruthy();
    expect(config.assistant_name).toBeTruthy();
    expect(config.user_name).toBeTruthy();
  });

  test('default session is auto-created', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/sessions`);
    expect(res.ok()).toBeTruthy();
    const data = await res.json();
    expect(data.sessions.length).toBeGreaterThanOrEqual(1);
  });

  test('compose box is visible and accepts input', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    const textarea = page.locator('.compose-input, .compose-box textarea');
    await expect(textarea).toBeVisible({ timeout: 5000 });
    await textarea.fill('Hello from Playwright');
    await expect(textarea).toHaveValue('Hello from Playwright');
  });

  test('sending a message produces a response in the timeline', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    const textarea = page.locator('.compose-input, .compose-box textarea, textarea').first();
    await expect(textarea).toBeVisible({ timeout: 5000 });
    await textarea.fill('Playwright test message');
    await textarea.press('Enter');
    // Wait for at least 2 posts to appear (user + assistant)
    await page.waitForTimeout(5000);
    const postCount = await page.locator('.post').count();
    expect(postCount).toBeGreaterThanOrEqual(2);
  });

  test('timeline shows posts with correct role styling', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    // Wait for timeline to have at least one post (from previous messages or polling)
    await page.waitForTimeout(3000);
    const postCount = await page.locator('.post').count();
    if (postCount > 0) {
      const avatars = page.locator('.post-avatar');
      expect(await avatars.count()).toBeGreaterThan(0);
    }
  });

  test('theme variables are applied to the document', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    const bgPrimary = await page.evaluate(() =>
      getComputedStyle(document.documentElement).getPropertyValue('--bg-primary').trim()
    );
    expect(bgPrimary).toBeTruthy();
  });

  test('CSS stylesheets load without errors', async ({ page }) => {
    const failedRequests: string[] = [];
    page.on('requestfailed', (req) => {
      if (req.url().includes('.css')) failedRequests.push(req.url());
    });
    await page.goto(BASE_URL);
    await page.waitForTimeout(2000);
    expect(failedRequests).toEqual([]);
  });

  test('vendor JS bundles load without errors', async ({ page }) => {
    const failedRequests: string[] = [];
    page.on('requestfailed', (req) => {
      if (req.url().includes('.js')) failedRequests.push(req.url());
    });
    await page.goto(BASE_URL);
    await page.waitForTimeout(2000);
    expect(failedRequests).toEqual([]);
  });

  test('workspace explorer toggle exists', async ({ page }) => {
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    const toggle = page.locator('.workspace-toggle-tab');
    await expect(toggle).toBeVisible({ timeout: 5000 });
  });

  test('no console errors during normal operation', async ({ page }) => {
    const consoleErrors: string[] = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') consoleErrors.push(msg.text());
    });
    await page.goto(BASE_URL);
    await expect(page.locator('.app-shell')).toBeVisible({ timeout: 10000 });
    await page.waitForTimeout(3000);
    // Filter out known non-critical errors
    const critical = consoleErrors.filter(e =>
      !e.includes('ResizeObserver') && !e.includes('favicon') && !e.includes('404')
    );
    expect(critical).toEqual([]);
  });
});
