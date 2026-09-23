import { test, expect } from '@playwright/test';
import { loadCorpus } from './support/catalogue.mjs';

test('@ux-pwa-001 Serve the manifest and declared built-in PNG icons', async ({ page, request }, info) => {
  const scenario = loadCorpus().find(row => row.id === '@ux-pwa-001');
  expect(scenario).toBeTruthy();
  await info.attach('gherkin', { body: scenario.steps.join('\n'), contentType: 'text/plain' });
  const shell = await request.get('/');
  expect(shell.status()).toBe(200);
  expect(await shell.text()).toContain('rel="manifest" href="/manifest.json"');
  const response = await request.get('/manifest.json');
  expect(response.status()).toBe(200);
  expect(response.headers()['content-type']).toContain('application/manifest+json');
  expect(response.headers()['cache-control']).toBe('no-store');
  const manifest = await response.json();
  expect(manifest.name).toBeTruthy();
  expect(manifest.short_name).toBe(manifest.name);
  expect(manifest.start_url).toBe('/');
  expect(manifest.display).toBe('standalone');
  expect(Array.isArray(manifest.icons)).toBe(true);
  const expected = ['192', '512'].flatMap(size => ['any', 'maskable'].map(purpose => ({
    src: `/static/icon-${size}.png`, sizes: `${size}x${size}`, type: 'image/png', purpose,
  })));
  for (const icon of expected) expect(manifest.icons).toContainEqual(icon);
  for (const src of new Set(expected.map(icon => icon.src))) {
    const image = await request.get(src);
    expect(image.status()).toBe(200);
    expect(image.headers()['content-type']).toContain('image/png');
    expect((await image.body()).subarray(0, 8)).toEqual(Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]));
  }
  const head = await request.head('/manifest.json');
  expect(head.status()).toBe(200);
  expect(head.headers()['content-length']).toBe(response.headers()['content-length']);
  await page.goto('/');
  await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', '/manifest.json');
});
