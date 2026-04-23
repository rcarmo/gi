/**
 * helpers.ts — shared utilities for functional tests.
 *
 * All tests run against an isolated Gi instance started by `make test-ux`.
 * The instance has its own database, workspace, and config.
 */
import { Page, expect } from '@playwright/test';

export const BASE_URL = process.env.GI_TEST_URL || 'http://127.0.0.1:19090';

/** Wait for the app shell to be fully rendered. */
export async function waitForAppShell(page: Page) {
  await expect(page.locator('.app-shell')).toBeVisible({ timeout: 15000 });
}

/** Wait for the app shell and return any JS errors that occurred during load. */
export async function loadPageCollectingErrors(page: Page): Promise<string[]> {
  const errors: string[] = [];
  page.on('pageerror', (err) => errors.push(err.message));
  await page.goto(BASE_URL);
  await waitForAppShell(page);
  return errors;
}

/** Get the compose input element (works across Piclaw compose variants). */
export function getComposeInput(page: Page) {
  return page.locator('.compose-input, .compose-box textarea, textarea').first();
}

/** Send a message via the compose box and wait for it to appear. */
export async function sendMessage(page: Page, text: string) {
  const input = getComposeInput(page);
  await expect(input).toBeVisible({ timeout: 5000 });
  await input.fill(text);
  await input.press('Enter');
}

/** Wait until at least `count` posts are visible in the timeline. */
export async function waitForPostCount(page: Page, count: number, timeoutMs = 15000) {
  await page.waitForFunction(
    (c) => document.querySelectorAll('.post').length >= c,
    count,
    { timeout: timeoutMs }
  );
}

/** Fetch JSON from the test instance API. */
export async function apiGet(request: any, path: string) {
  const res = await request.get(`${BASE_URL}${path}`);
  expect(res.ok()).toBeTruthy();
  return res.json();
}
