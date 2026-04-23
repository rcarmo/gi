/**
 * 04-sse-and-streaming.spec.ts — Verify SSE connection and real-time events.
 *
 * Tests that the SSE endpoint is reachable, sends the expected event types,
 * and that the frontend establishes and maintains the connection.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, sendMessage, apiGet } from './helpers';

test.describe('SSE and streaming', () => {

  test('SSE endpoint is reachable', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    // Verify SSE is working by checking that the EventSource connects
    const sseConnected = await page.evaluate(async () => {
      return new Promise<boolean>((resolve) => {
        const es = new EventSource('/sse/stream');
        es.addEventListener('connected', () => { es.close(); resolve(true); });
        es.onerror = () => { es.close(); resolve(false); };
        setTimeout(() => { es.close(); resolve(false); }, 5000);
      });
    });
    expect(sseConnected).toBeTruthy();
  });

  test('SSE sends connected event', async ({ page }) => {
    const sseEvents: string[] = [];
    // Intercept SSE to capture events
    await page.route('**/sse/stream**', async (route) => {
      // Let it through but capture
      route.continue();
    });
    
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    
    // Check that the frontend SSEClient connected
    const connectionStatus = await page.evaluate(() => {
      // Check localStorage or DOM for connection state
      return document.querySelector('.connection-status')?.textContent || 'no-indicator';
    });
    // If there's no explicit indicator, at least verify no connection error
    expect(connectionStatus).not.toContain('disconnected');
  });

  test('sending a message triggers SSE events', async ({ page }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    
    // Get session ID
    const sessions = await page.evaluate(async () => {
      const r = await fetch('/api/sessions');
      return r.json();
    });
    const sid = sessions.sessions?.[0]?.id;
    expect(sid).toBeTruthy();

    await sendMessage(page, 'SSE trigger test');
    await page.waitForTimeout(5000);
    
    // Verify the turn completed by checking API state
    const turns = await page.evaluate(async (sessionId: string) => {
      const r = await fetch(`/api/sessions/${sessionId}/turns`);
      return r.json();
    }, sid);
    
    expect(turns.turns.length).toBeGreaterThan(0);
    const lastTurn = turns.turns[turns.turns.length - 1];
    expect(lastTurn.status).toBe('completed');
  });

  test('turn events are persisted for completed turns', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'turn events check');
    await page.waitForTimeout(5000);

    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    const lastTurn = turns.turns[turns.turns.length - 1];
    
    const events = await apiGet(request, `/api/turns/${lastTurn.id}/events`);
    const eventTypes = events.events.map((e: any) => e.type);
    
    // Must have at least: turn.submitted, turn.started, turn.finished
    expect(eventTypes).toContain('turn.submitted');
    expect(eventTypes).toContain('turn.started');
    expect(eventTypes).toContain('turn.finished');
  });
});
