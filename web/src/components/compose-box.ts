// @ts-nocheck
import { html, useEffect, useRef, useState } from '../vendor/preact-htm-entry.ts';

export function ComposeBox({ disabled, onSend, runtimeConfig, sessionState, queuedCount }) {
  const [text, setText] = useState('');
  const [history, setHistory] = useState(() => {
    try { return JSON.parse(localStorage.getItem('piclaw_compose_history:web:default') || '[]'); } catch { return []; }
  });
  const [historyIndex, setHistoryIndex] = useState(-1);
  const textareaRef = useRef(null);
  useEffect(() => { console.debug('[compose] mounted'); return () => console.debug('[compose] unmounted'); }, []);
  useEffect(() => {
    try { localStorage.setItem('piclaw_compose_history:web:default', JSON.stringify(history.slice(-20))); } catch {}
  }, [history]);
  const submit = async () => {
    const payload = text;
    if (!payload.trim()) return;
    setText('');
    setHistory((prev) => [...prev, payload]);
    setHistoryIndex(-1);
    console.debug('[compose] submit', { payload });
    await onSend(payload);
  };
  return html`<div class="compose-box"><div class="compose-input-wrapper"><div class="compose-input-main"><textarea ref=${textareaRef} class="compose-input" rows="5" placeholder="Send a prompt" value=${text} onInput=${(e) => setText(e.target.value)} onKeyDown=${async (e) => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); await submit(); return; }
    if (e.key === 'ArrowUp' && !text && history.length > 0) { e.preventDefault(); const nextIndex = historyIndex < 0 ? history.length - 1 : Math.max(0, historyIndex - 1); setHistoryIndex(nextIndex); setText(history[nextIndex] || ''); }
    if (e.key === 'ArrowDown' && history.length > 0 && historyIndex >= 0) { e.preventDefault(); const nextIndex = historyIndex + 1; if (nextIndex >= history.length) { setHistoryIndex(-1); setText(''); } else { setHistoryIndex(nextIndex); setText(history[nextIndex] || ''); } }
  }} /></div>${queuedCount > 0 ? html`<div class="compose-inline-status"><div class="compose-inline-status-row"><span class="compose-inline-status-title">${queuedCount} queued turn${queuedCount === 1 ? '' : 's'}</span></div></div>` : null}</div><div class="compose-toolbar"><span class="compose-chip">${sessionState?.provider || runtimeConfig.default_provider || 'provider'}</span><span class="compose-chip">${sessionState?.model || runtimeConfig.default_model || 'model'}</span><span class="compose-chip">thinking ${sessionState?.thinking_level || runtimeConfig.default_thinking_level || 'default'}</span></div><div class="compose-actions"><button class="icon-btn send-btn" disabled=${disabled || !text.trim()} onClick=${submit}>${disabled ? 'Running…' : 'Send'}</button></div></div>`;
}
