// @ts-nocheck
import { html, useEffect, useState } from '../vendor/preact-htm-entry.ts';

export function ComposeBox({ disabled, onSend, runtimeConfig, sessionState, queuedCount }) {
  const [text, setText] = useState('');
  const [history, setHistory] = useState(() => []);
  useEffect(() => { console.debug('[compose] mounted'); return () => console.debug('[compose] unmounted'); }, []);
  useEffect(() => {
    const key = 'piclaw_compose_history:web:default';
    try { localStorage.setItem(key, JSON.stringify(history.slice(-20))); } catch {}
  }, [history]);
  return html`<div class="compose-box"><div class="compose-input-wrapper"><div class="compose-input-main"><textarea class="compose-input" rows="5" placeholder="Send a prompt" value=${text} onInput=${(e) => setText(e.target.value)} /></div>${queuedCount > 0 ? html`<div class="compose-inline-status"><div class="compose-inline-status-row"><span class="compose-inline-status-title">${queuedCount} queued turn${queuedCount === 1 ? '' : 's'}</span></div></div>` : null}</div><div class="compose-toolbar"><span class="compose-chip">${sessionState?.provider || runtimeConfig.default_provider || 'provider'}</span><span class="compose-chip">${sessionState?.model || runtimeConfig.default_model || 'model'}</span><span class="compose-chip">thinking ${sessionState?.thinking_level || runtimeConfig.default_thinking_level || 'default'}</span></div><div class="compose-actions"><button class="icon-btn send-btn" disabled=${disabled || !text.trim()} onClick=${async () => { const payload = text; setText(''); setHistory((prev) => [...prev, payload]); console.debug('[compose] submit', { payload }); await onSend(payload); }}>${disabled ? 'Running…' : 'Send'}</button></div></div>`;
}
