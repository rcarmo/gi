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

export const getRuntimeConfig = () => request('/api/runtime/config');
export const listSessions = () => request('/api/sessions');
export const createSession = (title: string) => { recordStatus('Creating session…'); return request('/api/sessions', { method: 'POST', body: JSON.stringify({ title }) }); };
export const listMessages = (sessionID: string) => request(`/api/sessions/${encodeURIComponent(sessionID)}/messages`);
export const listTurns = (sessionID: string) => request(`/api/sessions/${encodeURIComponent(sessionID)}/turns`);
export const listTurnEvents = (turnID: string) => request(`/api/turns/${encodeURIComponent(turnID)}/events`);
export const cancelTurn = (turnID: string) => request(`/api/turns/${encodeURIComponent(turnID)}/cancel`, { method: 'POST' });
export const getWorkspaceTree = () => request('/api/workspace/tree');
export const getWorkspaceFile = (path: string) => request(`/api/workspace/file?path=${encodeURIComponent(path)}`);
export async function sendPrompt(sessionID: string, prompt: string, intent = 'prompt', model = 'bootstrap') {
  recordStatus('Submitting turn…');
  return request(`/api/sessions/${encodeURIComponent(sessionID)}/prompt`, {
    method: 'POST',
    body: JSON.stringify({ prompt, intent, model }),
  });
}
