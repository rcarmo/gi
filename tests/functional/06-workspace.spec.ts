/**
 * 06-workspace.spec.ts — Verify workspace browser and file APIs.
 *
 * Tests that the workspace tree API works, files can be read, and the
 * workspace explorer component renders in the UI.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, apiGet } from './helpers';

test.describe('Workspace', () => {

  test('workspace tree API returns valid structure', async ({ request }) => {
    const tree = await apiGet(request, '/api/workspace/tree');
    expect(tree.name).toBeTruthy();
    expect(tree.type).toBe('dir');
    expect(tree.path).toBe('.');
  });

  test('workspace tree includes seeded config files', async ({ request }) => {
    const tree = await apiGet(request, '/api/workspace/tree');
    const childNames = (tree.children || []).map((c: any) => c.name);
    // The test instance seeds .piclaw/ and .pi/
    expect(childNames).toContain('.piclaw');
    expect(childNames).toContain('.pi');
  });

  test('workspace file API reads seeded config', async ({ request }) => {
    const file = await apiGet(request, '/api/workspace/file?path=.piclaw/config.json');
    expect(file.path).toBe('.piclaw/config.json');
    expect(file.content).toContain('Gi Test');
  });

  test('workspace file API rejects path traversal', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/workspace/file?path=../../etc/passwd`);
    expect(res.status()).toBe(400);
  });

  test('workspace file API returns error for missing file', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/api/workspace/file?path=nonexistent.txt`);
    expect(res.status()).toBe(500); // os.ReadFile error
  });

  test('workspace toggle button exists in UI', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const toggle = page.locator('.workspace-toggle-tab');
    await expect(toggle).toBeVisible({ timeout: 5000 });
  });
});
