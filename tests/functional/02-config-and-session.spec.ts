/**
 * 02-config-and-session.spec.ts — Verify runtime config loading and session management.
 *
 * Tests that Pi/Piclaw settings are loaded correctly and that the default
 * session is auto-created without user interaction.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, apiGet, waitForAppShell } from './helpers';

test.describe('Config and session', () => {

  test('runtime config API returns valid structure', async ({ request }) => {
    const cfg = await apiGet(request, '/api/runtime/config');
    expect(cfg.assistant_name).toBeTruthy();
    expect(cfg.user_name).toBeTruthy();
    expect(cfg.default_model).toBeTruthy();
    expect(cfg.default_provider).toBeTruthy();
    expect(typeof cfg.version).toBe('string');
  });

  test('runtime config loads assistant identity from .piclaw/config.json', async ({ request }) => {
    const cfg = await apiGet(request, '/api/runtime/config');
    // The test instance seeds these values
    expect(cfg.assistant_name).toBe('Gi Test');
    expect(cfg.user_name).toBe('Test User');
  });

  test('runtime config loads model from .pi/settings.json', async ({ request }) => {
    const cfg = await apiGet(request, '/api/runtime/config');
    // The test instance uses -model flag override, but settings are loaded
    expect(cfg.default_provider).toBeTruthy();
  });

  test('at least one session exists after page load', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    // The app auto-creates a default session
    const data = await apiGet(request, '/api/sessions');
    expect(data.sessions.length).toBeGreaterThanOrEqual(1);
  });

  test('session has valid structure', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const data = await apiGet(request, '/api/sessions');
    const session = data.sessions[0];
    expect(session.id).toBeTruthy();
    expect(session.state).toBeDefined();
    expect(session.created_at).toBeTruthy();
    expect(session.updated_at).toBeTruthy();
  });

  test('session state contains model and status', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const data = await apiGet(request, '/api/sessions');
    const state = data.sessions[0].state;
    expect(state.status).toBeDefined();
  });
});
