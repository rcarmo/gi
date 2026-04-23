/**
 * 03-chat-flow.spec.ts — Verify the core chat experience.
 *
 * Tests the complete send → process → display cycle including compose box
 * interaction, message persistence, timeline rendering, and content visibility.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, getComposeInput, sendMessage, waitForPostCount, apiGet } from './helpers';

test.describe('Chat flow', () => {

  test('compose box is visible and accepts input', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await expect(input).toBeVisible({ timeout: 5000 });
    await input.fill('test input text');
    await expect(input).toHaveValue('test input text');
  });

  test('Enter submits a message', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'Enter submit test');
    // User message should appear as a post
    await waitForPostCount(page, 1);
  });

  test('assistant responds to a message', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'Hello from functional test');
    // Wait for both user + assistant posts
    await waitForPostCount(page, 2, 15000);
  });

  test('user message content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'visible content check');
    await page.waitForTimeout(3000);
    await expect(
      page.locator('.post-content').filter({ hasText: 'visible content check' }).first()
    ).toBeVisible({ timeout: 10000 });
  });

  test('assistant response content is visible in post-content', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'assistant content check');
    // The test instance shell stub responds with "Gi received: ..."
    await expect(
      page.locator('.post-content').filter({ hasText: 'Gi received' }).first()
    ).toBeVisible({ timeout: 15000 });
  });

  test('messages are persisted in the database', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'persistence check');
    await page.waitForTimeout(5000);
    // Fetch sessions and messages via API
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0]?.id;
    expect(sid).toBeTruthy();
    const msgs = await apiGet(request, `/api/sessions/${sid}/messages`);
    const contents = msgs.messages.map((m: any) => m.content);
    expect(contents.some((c: string) => c.includes('persistence check'))).toBeTruthy();
  });

  test('posts have avatar elements', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'avatar check');
    await waitForPostCount(page, 2, 15000);
    const avatars = page.locator('.post-avatar');
    expect(await avatars.count()).toBeGreaterThan(0);
  });

  test('agent posts have agent-avatar class', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'agent avatar class check');
    await waitForPostCount(page, 2, 15000);
    await expect(page.locator('.post-avatar.agent-avatar').first()).toBeVisible({ timeout: 5000 });
  });

  test('compose box clears after submit', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const input = getComposeInput(page);
    await input.fill('will be cleared');
    await input.press('Enter');
    await page.waitForTimeout(500);
    // Input should be empty after submit
    const value = await input.inputValue();
    expect(value).toBe('');
  });
});
