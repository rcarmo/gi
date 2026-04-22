// @ts-nocheck
import { html, useEffect, useState } from '../vendor/preact-htm-entry.ts';

export function StatusBar({ text }) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    console.debug('[status] ticker mounted');
    return () => { clearInterval(id); console.debug('[status] ticker unmounted'); };
  }, []);
  return html`<div class="agent-status"><div class="status-bar"><div class="status-text">${text || 'Ready.'}</div><div class="status-time">${new Date(now).toLocaleTimeString()}</div></div></div>`;
}
