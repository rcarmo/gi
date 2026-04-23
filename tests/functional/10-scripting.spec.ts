/**
 * 10-scripting.spec.ts — Verify the scripting bridge and script tool.
 *
 * Tests that the tools API is discoverable, the script tool executes
 * Joker scripts, and the bridge state is correctly injected.
 */
import { test, expect } from '@playwright/test';
import { BASE_URL, apiGet } from './helpers';

test.describe('Scripting', () => {

  test('tools API returns available tools', async ({ request }) => {
    const data = await apiGet(request, '/api/tools');
    expect(data.tools).toBeDefined();
    expect(data.tools.length).toBeGreaterThanOrEqual(1);
    const names = data.tools.map((t: any) => t.name);
    expect(names).toContain('script');
    expect(names).toContain('read');
    expect(names).toContain('write');
    expect(names).toContain('shell');
  });

  test('script tool has correct definition', async ({ request }) => {
    const data = await apiGet(request, '/api/tools');
    const scriptTool = data.tools.find((t: any) => t.name === 'script');
    expect(scriptTool.description).toContain('Joker');
    expect(scriptTool.parameters.properties.script).toBeDefined();
    expect(scriptTool.parameters.properties.path).toBeDefined();
  });

  test('script tool executes basic math', async ({ request }) => {
    // Get a session ID
    const sessions = await apiGet(request, '/api/sessions');
    // Create one if none exist
    let sid: string;
    if (sessions.sessions.length === 0) {
      const res = await request.post(`${BASE_URL}/api/sessions`, {
        data: { title: 'script-test' },
      });
      const s = await res.json();
      sid = s.id;
    } else {
      sid = sessions.sessions[0].id;
    }

    const res = await request.post(`${BASE_URL}/api/tools/execute`, {
      data: {
        tool: 'script',
        session_id: sid,
        input: { script: '(+ 40 2)' },
      },
    });
    expect(res.ok()).toBeTruthy();
    const output = await res.json();
    expect(output.result).toBe('42');
    expect(output.error).toBeFalsy();
  });

  test('script tool has access to bridge state', async ({ request }) => {
    const sessions = await apiGet(request, '/api/sessions');
    let sid: string;
    if (sessions.sessions.length === 0) {
      const res = await request.post(`${BASE_URL}/api/sessions`, {
        data: { title: 'bridge-test' },
      });
      sid = (await res.json()).id;
    } else {
      sid = sessions.sessions[0].id;
    }

    const res = await request.post(`${BASE_URL}/api/tools/execute`, {
      data: {
        tool: 'script',
        session_id: sid,
        input: { script: '(:session-id *gi-bridge*)' },
      },
    });
    const output = await res.json();
    expect(output.result).toBe(sid);
  });

  test('script tool has config in bridge', async ({ request }) => {
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0]?.id;

    const res = await request.post(`${BASE_URL}/api/tools/execute`, {
      data: {
        tool: 'script',
        session_id: sid,
        input: { script: '(:assistant_name (:config *gi-bridge*))' },
      },
    });
    const output = await res.json();
    expect(output.result).toBe('Gi Test'); // seeded by test instance
  });

  test('script tool rejects unknown tool name', async ({ request }) => {
    const res = await request.post(`${BASE_URL}/api/tools/execute`, {
      data: {
        tool: 'nonexistent',
        input: {},
      },
    });
    expect(res.status()).toBe(400);
  });

  test('script tool requires script or path', async ({ request }) => {
    const sessions = await apiGet(request, '/api/sessions');
    const sid = sessions.sessions[0]?.id;

    const res = await request.post(`${BASE_URL}/api/tools/execute`, {
      data: {
        tool: 'script',
        session_id: sid,
        input: {},
      },
    });
    const output = await res.json();
    expect(output.error).toContain('required');
  });
});
