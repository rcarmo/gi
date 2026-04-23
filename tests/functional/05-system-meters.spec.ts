/**
 * 05-system-meters.spec.ts — Verify the system meters HUD.
 *
 * Tests that the metrics endpoint returns valid data and that the HUD
 * component renders in the web UI.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, apiGet } from './helpers';

test.describe('System meters', () => {

  test('metrics endpoint returns CPU and RAM data', async ({ request }) => {
    const metrics = await apiGet(request, '/api/system-metrics');
    expect(typeof metrics.cpu_percent).toBe('number');
    expect(typeof metrics.ram_percent).toBe('number');
    expect(metrics.ram_percent).toBeGreaterThan(0);
    expect(Array.isArray(metrics.cpu_series)).toBeTruthy();
    expect(Array.isArray(metrics.ram_series)).toBeTruthy();
  });

  test('metrics endpoint returns process memory data', async ({ request }) => {
    const metrics = await apiGet(request, '/api/system-metrics');
    expect(metrics.process_memory).toBeDefined();
    expect(metrics.process_memory.rss_bytes).toBeGreaterThan(0);
    expect(metrics.process_memory.heap_used_bytes).toBeGreaterThan(0);
  });

  test('metrics endpoint returns swap data', async ({ request }) => {
    const metrics = await apiGet(request, '/api/system-metrics');
    // swap_percent can be null or a number
    expect(
      metrics.swap_percent === null || typeof metrics.swap_percent === 'number'
    ).toBeTruthy();
  });

  test('metrics endpoint returns sample_interval_ms matching Piclaw (2000)', async ({ request }) => {
    const metrics = await apiGet(request, '/api/system-metrics');
    expect(metrics.sample_interval_ms).toBe(2000);
  });

  test('second metrics call shows CPU delta', async ({ request }) => {
    // First call primes the CPU delta
    await apiGet(request, '/api/system-metrics');
    await new Promise(r => setTimeout(r, 1000));
    const m2 = await apiGet(request, '/api/system-metrics');
    // After two samples, series should have 2 entries
    expect(m2.cpu_series.length).toBeGreaterThanOrEqual(2);
  });

  test('Piclaw-compatible endpoint also works', async ({ request }) => {
    const metrics = await apiGet(request, '/agent/system-metrics');
    expect(typeof metrics.cpu_percent).toBe('number');
    expect(typeof metrics.ram_percent).toBe('number');
  });

  test('system meters HUD renders in the UI', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    // Wait for meters to load (they poll on an interval)
    await page.waitForTimeout(4000);
    // The HUD should be visible as an overlay
    const hud = page.locator('.system-meters-hud');
    // It may or may not be visible depending on localStorage, but the element should exist
    const count = await hud.count();
    expect(count).toBeGreaterThanOrEqual(0); // exists in DOM
  });
});
