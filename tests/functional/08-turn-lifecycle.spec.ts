/**
 * 08-turn-lifecycle.spec.ts — Verify turn queue, cancellation, and state transitions.
 *
 * Tests the turn engine's behavior through the API: turn creation, status
 * transitions, event persistence, and queue management.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, waitForAppShell, apiGet, sendMessage } from './helpers';

test.describe('Turn lifecycle', () => {

  test('submitting a prompt creates a turn', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'turn creation test');
    await page.waitForTimeout(5000);
    
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    expect(turns.turns.length).toBeGreaterThan(0);
  });

  test('completed turn has expected event sequence', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'event sequence test');
    await page.waitForTimeout(5000);
    
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    const completed = turns.turns.filter((t: any) => t.status === 'completed');
    expect(completed.length).toBeGreaterThan(0);
    
    const lastCompleted = completed[completed.length - 1];
    const events = await apiGet(request, `/api/turns/${lastCompleted.id}/events`);
    const types = events.events.map((e: any) => e.type);
    
    expect(types).toContain('turn.submitted');
    expect(types).toContain('turn.started');
    expect(types).toContain('turn.finished');
    // Events should be in order
    expect(types.indexOf('turn.submitted')).toBeLessThan(types.indexOf('turn.started'));
    expect(types.indexOf('turn.started')).toBeLessThan(types.indexOf('turn.finished'));
  });

  test('turn events have checkpoint markers', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'checkpoint test');
    await page.waitForTimeout(5000);
    
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    const lastTurn = turns.turns[turns.turns.length - 1];
    const events = await apiGet(request, `/api/turns/${lastTurn.id}/events`);
    
    // All key events should have checkpoint: true
    const checkpointed = events.events.filter((e: any) => e.payload?.checkpoint === true);
    expect(checkpointed.length).toBeGreaterThanOrEqual(3); // submitted, started, finished
  });

  test('turn metadata contains intent and model', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    await sendMessage(page, 'metadata test');
    await page.waitForTimeout(5000);
    
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    const lastTurn = turns.turns[turns.turns.length - 1];
    
    expect(lastTurn.metadata.intent).toBe('prompt');
    expect(lastTurn.metadata.model).toBeTruthy();
  });

  test('turn prompt matches what was sent', async ({ page, request }) => {
    await page.goto(BASE_URL);
    await waitForAppShell(page);
    const uniqueMsg = `prompt match ${Date.now()}`;
    await sendMessage(page, uniqueMsg);
    await page.waitForTimeout(5000);
    
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0].id;
    const turns = await apiGet(request, `/api/sessions/${sid}/turns`);
    const lastTurn = turns.turns[turns.turns.length - 1];
    expect(lastTurn.prompt).toBe(uniqueMsg);
  });
});
