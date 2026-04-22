// @ts-nocheck

export function resolveActiveTurn(turns) {
  const items = Array.isArray(turns) ? turns : [];
  return items.find((t) => t.status === 'running' || t.status === 'cancelling')
    || items.find((t) => t.status === 'queued')
    || items[items.length - 1]
    || null;
}

export function resolveTurnStatusPresentation(turn, events, nowMs = Date.now()) {
  if (!turn) return { title: 'Ready.', detail: '', tone: 'idle', elapsed: '' };
  const list = Array.isArray(events) ? events : [];
  const last = list[list.length - 1] || null;
  const eventType = String(last?.type || '');
  const payload = last?.payload || {};
  const createdAt = last?.created_at || turn.updated_at || turn.created_at;
  const elapsed = formatElapsed(createdAt, nowMs);

  if (turn.status === 'queued') {
    return { title: 'Queued', detail: turn.prompt || '', tone: 'queued', elapsed };
  }
  if (turn.status === 'cancelling') {
    return { title: 'Cancelling…', detail: turn.prompt || '', tone: 'warning', elapsed };
  }
  if (turn.status === 'cancelled') {
    return { title: 'Cancelled', detail: turn.prompt || '', tone: 'danger', elapsed };
  }
  if (turn.status === 'failed') {
    return { title: 'Failed', detail: String(payload.error || turn.prompt || ''), tone: 'danger', elapsed };
  }
  if (turn.status === 'completed') {
    return { title: 'Completed', detail: String(payload.status || turn.prompt || ''), tone: 'success', elapsed };
  }
  if (eventType === 'tool.started') {
    return { title: `Running ${payload.tool || 'tool'}…`, detail: turn.prompt || '', tone: 'active', elapsed };
  }
  if (eventType === 'turn.started') {
    return { title: 'Thinking…', detail: turn.prompt || '', tone: 'active', elapsed };
  }
  if (eventType === 'turn.submitted') {
    return { title: payload.queued ? 'Queued' : 'Starting…', detail: turn.prompt || '', tone: payload.queued ? 'queued' : 'active', elapsed };
  }
  return { title: turn.status || 'Working…', detail: turn.prompt || '', tone: 'idle', elapsed };
}

function formatElapsed(iso, nowMs = Date.now()) {
  const started = Date.parse(String(iso || ''));
  if (!Number.isFinite(started)) return '';
  const totalSec = Math.max(0, Math.floor((nowMs - started) / 1000));
  const m = Math.floor(totalSec / 60);
  const s = totalSec % 60;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
}
