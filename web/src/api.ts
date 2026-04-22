import { recordStatus } from './status';

async function request(url: string, options: RequestInit = {}) {
  const res = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(body.error || `HTTP ${res.status}`);
  return body;
}

export async function listSessions() {
  return request('/api/sessions');
}

export async function createSession(title: string) {
  recordStatus(`Creating session…`);
  return request('/api/sessions', { method: 'POST', body: JSON.stringify({ title }) });
}

export async function listMessages(sessionID: string) {
  return request(`/api/sessions/${encodeURIComponent(sessionID)}/messages`);
}

export async function sendPrompt(sessionID: string, prompt: string, intent = 'prompt', model = 'bootstrap') {
  recordStatus('Running turn…');
  return request(`/api/sessions/${encodeURIComponent(sessionID)}/prompt`, {
    method: 'POST',
    body: JSON.stringify({ prompt, intent, model }),
  });
}
