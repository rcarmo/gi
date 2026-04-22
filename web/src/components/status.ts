// @ts-nocheck
import { html, useEffect, useState } from '../vendor/preact-htm-entry.ts';

export function StatusBar({ title, detail, tone = 'idle', queuedCount = 0 }) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    console.debug('[status] ticker mounted');
    return () => { clearInterval(id); console.debug('[status] ticker unmounted'); };
  }, []);
  return html`<div class="agent-status"><div class=${`status-bar tone-${tone}`}><div class="status-main"><div class="status-text">${title || 'Ready.'}</div>${detail ? html`<div class="status-detail">${detail}</div>` : null}</div><div class="status-side">${queuedCount > 0 ? html`<span class="status-queue-chip">${queuedCount} queued</span>` : null}<div class="status-time">${new Date(now).toLocaleTimeString()}</div></div></div></div>`;
}
