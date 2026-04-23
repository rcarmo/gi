/**
 * 07-compose-interaction.spec.ts — Verify compose box interaction details.
 *
 * Tests keyboard behavior, history navigation, and input state management
 * beyond basic send/receive.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, getComposeInput, sendMessage } from './helpers';

test.describe('Compose interaction', () => {

  test('Shift+Enter does not submit', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await input.fill('line one');
    await input.press('Shift+Enter');
    await page.waitForTimeout(500);
    // Input should still have text (not submitted)
    const value = await input.inputValue();
    expect(value.length).toBeGreaterThan(0);
  });

  test('empty input does not submit on Enter', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await input.fill('');
    const postCountBefore = await page.locator('.post').count();
    await input.press('Enter');
    await page.waitForTimeout(1000);
    const postCountAfter = await page.locator('.post').count();
    // No new post should appear
    expect(postCountAfter).toBe(postCountBefore);
  });

  test('typing reaches the compose input', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await page.waitForTimeout(2000);
    // Click on the compose area first to ensure focus
    const input = getComposeInput(page);
    await input.click();
    await page.keyboard.type('focus test');
    const value = await input.inputValue();
    expect(value).toContain('focus test');
  });

  test('multiple messages can be sent in sequence', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'first message');
    await page.waitForTimeout(3000);
    await sendMessage(page, 'second message');
    await page.waitForTimeout(3000);
    // Should have at least 2 user messages
    const posts = await page.locator('.post').count();
    expect(posts).toBeGreaterThanOrEqual(2);
  });
});
