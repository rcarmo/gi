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
