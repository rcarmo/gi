// @ts-nocheck
import { html, render, useState, useEffect } from './vendor/preact-htm-entry.ts';
import { getRuntimeConfig, listSessions, createSession, listMessages, sendPrompt, listTurns, cancelTurn } from './api.ts';
import { bindStatusListener, recordStatus } from './status.ts';

const ICONS = {
  assistant: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a3 3 0 0 0-3 3v1H8a4 4 0 0 0-4 4v3a4 4 0 0 0 4 4h1v1a3 3 0 0 0 6 0v-1h1a4 4 0 0 0 4-4v-3a4 4 0 0 0-4-4h-1V5a3 3 0 0 0-3-3Z"/><path d="M9 11h.01M15 11h.01"/></svg>`,
  user: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21a8 8 0 0 0-16 0"/><circle cx="12" cy="7" r="4"/></svg>`,
  model: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h6v6H9z"/></svg>`,
  theme: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`,
  queued: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 6h13M8 12h13M8 18h13"/><path d="M3 6h.01M3 12h.01M3 18h.01"/></svg>`,
};

function SessionList({ sessions, currentSessionId, onSelect, onCreate }) {
  const [title, setTitle] = useState('');
  return html`
    <aside class="sidebar">
      <div class="panel-title">Sessions</div>
      <div class="new-session-row">
        <input value=${title} onInput=${(e) => setTitle(e.target.value)} placeholder="New session title" />
        <button onClick=${async () => { const created = await onCreate(title); setTitle(''); if (created?.id) onSelect(created.id); }}>New</button>
      </div>
      <div class="session-list">
        ${sessions.map((session) => html`
          <button class=${`session-item ${currentSessionId === session.id ? 'active' : ''}`} onClick=${() => onSelect(session.id)}>
            <span class="session-title">${session.title || session.id}</span>
            <span class="session-meta">${session.state?.status || 'idle'}</span>
          </button>
        `)}
      </div>
    </aside>`;
}

function Timeline({ messages }) {
  return html`<section class="timeline normal"><div class="timeline-content">${messages.length === 0 ? html`<div class="empty-state">No messages yet.</div>` : null}${messages.map((msg) => {
    const isAssistant = msg.role === 'assistant';
    const isUser = msg.role === 'user';
    const icon = isAssistant ? ICONS.assistant : isUser ? ICONS.user : ICONS.queued;
    return html`<article class=${`post gi-post role-${msg.role}`}><div class=${`post-avatar ${isAssistant ? 'assistant' : isUser ? 'user' : 'system'}`}>${icon}</div><div class="post-body"><div class="post-header"><span class="post-author">${msg.role}</span><span class="post-meta">${msg.payload?.source || msg.payload?.kind || ''}</span></div><div class="post-content"><div class="message-content">${msg.content}</div></div></div></article>`;
  })}</div></section>`;
}

function TurnQueue({ turns, onCancel }) {
  return html`<aside class="gi-turn-queue"><div class="panel-title">Turns</div>${turns.length === 0 ? html`<div class="empty-state">No turns yet.</div>` : turns.map((turn) => html`<div class="turn-item"><div><div class="turn-status turn-status-${turn.status}">${turn.status}</div><div class="turn-prompt">${turn.prompt}</div></div>${turn.status === 'queued' || turn.status === 'running' || turn.status === 'cancelling' ? html`<button class="chat-window-header-button" onClick=${() => onCancel(turn.id)}>Cancel</button>` : null}</div>` )}</aside>`;
}

function ComposeBox({ disabled, onSend, runtimeConfig, sessionState, queuedCount }) {
  const [text, setText] = useState('');
  return html`<div class="compose-box"><div class="compose-input-wrapper"><div class="compose-input-main"><textarea class="compose-input" rows="5" placeholder="Send a prompt" value=${text} onInput=${(e) => setText(e.target.value)} /></div>${queuedCount > 0 ? html`<div class="compose-inline-status"><div class="compose-inline-status-row"><span class="compose-inline-status-glyph">${ICONS.queued}</span><span class="compose-inline-status-title">${queuedCount} queued turn${queuedCount === 1 ? '' : 's'}</span></div></div>` : null}</div><div class="compose-toolbar"><span class="compose-chip">${ICONS.model}${sessionState?.provider || runtimeConfig.default_provider || 'provider'}</span><span class="compose-chip">${ICONS.model}${sessionState?.model || runtimeConfig.default_model || 'model'}</span><span class="compose-chip">thinking ${sessionState?.thinking_level || runtimeConfig.default_thinking_level || 'default'}</span></div><div class="compose-actions"><button class="icon-btn send-btn" disabled=${disabled || !text.trim()} onClick=${async () => { const payload = text; setText(''); await onSend(payload); }}>${disabled ? 'Running…' : 'Send'}</button></div></div>`;
}

function StatusBar({ text }) { return html`<div class="agent-status"><div class="status-bar"><div class="status-text">${text || 'Ready.'}</div></div></div>`; }

function App() {
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [messages, setMessages] = useState([]);
  const [turns, setTurns] = useState([]);
  const [status, setStatus] = useState('Ready.');
  const [running, setRunning] = useState(false);
  const [runtimeConfig, setRuntimeConfig] = useState({ assistant_name: 'Gi', user_name: 'User', default_model: '', default_provider: '', default_thinking_level: '' });
  const [themeMode, setThemeMode] = useState(() => localStorage.getItem('gi_theme_mode') || 'auto');

  useEffect(() => { bindStatusListener(setStatus); refreshRuntimeConfig(); refreshSessions(); }, []);
  useEffect(() => {
    const root = document.documentElement;
    if (themeMode === 'light') {
      root.classList.add('light'); root.classList.remove('dark');
    } else if (themeMode === 'dark') {
      root.classList.add('dark'); root.classList.remove('light');
    } else {
      root.classList.remove('light'); root.classList.remove('dark');
    }
    localStorage.setItem('gi_theme_mode', themeMode);
  }, [themeMode]);
  useEffect(() => {
    if (!currentSessionId) return;
    const timer = setInterval(async () => {
      await Promise.all([refreshMessages(currentSessionId), refreshTurns(currentSessionId), refreshSessions()]);
    }, 1000);
    return () => clearInterval(timer);
  }, [currentSessionId]);

  async function refreshRuntimeConfig() { const data = await getRuntimeConfig(); setRuntimeConfig(data || {}); }
  async function refreshSessions() { const data = await listSessions(); setSessions(data.sessions || []); }
  async function refreshMessages(sessionID = currentSessionId) { if (!sessionID) return setMessages([]); const data = await listMessages(sessionID); setMessages(data.messages || []); }
  async function refreshTurns(sessionID = currentSessionId) { if (!sessionID) return setTurns([]); const data = await listTurns(sessionID); setTurns(data.turns || []); setRunning((data.turns || []).some((t) => t.status === 'running' || t.status === 'cancelling')); }
  async function handleCreate(title) { const created = await createSession(title || 'Untitled'); await refreshSessions(); await refreshMessages(created.id); await refreshTurns(created.id); setCurrentSessionId(created.id); recordStatus(`Created ${created.id}`); return created; }
  async function handleSelect(sessionID) { setCurrentSessionId(sessionID); await refreshMessages(sessionID); await refreshTurns(sessionID); recordStatus(`Opened ${sessionID}`); }
  async function handleSend(prompt) { if (!currentSessionId) return recordStatus('Create or open a session first.'); const result = await sendPrompt(currentSessionId, prompt); await refreshTurns(currentSessionId); recordStatus(result.queued ? `Queued ${result.turn_id}` : `Started ${result.turn_id}`); }
  async function handleCancel(turnID) { await cancelTurn(turnID); await refreshTurns(currentSessionId); recordStatus(`Cancelling ${turnID}`); }
  const currentSession = sessions.find((s) => s.id === currentSessionId) || null;
  const queuedCount = turns.filter((t) => t.status === 'queued').length;

  return html`<div class="app-shell chat-only"><${SessionList} sessions=${sessions} currentSessionId=${currentSessionId} onSelect=${handleSelect} onCreate=${handleCreate} /><main class="container"><section class="chat-window"><header class="chat-window-header"><div class="chat-window-header-main"><div class="chat-window-header-title">${runtimeConfig.assistant_name || 'Gi'}</div><div class="chat-window-header-subtitle">${runtimeConfig.default_provider || 'provider'} / ${runtimeConfig.default_model || 'model'}${runtimeConfig.default_thinking_level ? ` · ${runtimeConfig.default_thinking_level}` : ''}</div></div><div class="chat-window-header-actions"><span class="chat-window-header-badge">${runtimeConfig.user_name || 'User'}</span><span class="chat-window-header-badge">${ICONS.model}${currentSession?.state?.model || runtimeConfig.default_model || 'model'}</span><button class="chat-window-header-button" onClick=${() => setThemeMode(themeMode === 'dark' ? 'light' : themeMode === 'light' ? 'auto' : 'dark')}>${ICONS.theme}${themeMode}</button></div></header><${StatusBar} text=${status} /><div class="main-grid"><div class="chat-column"><${Timeline} messages=${messages} /><${ComposeBox} disabled=${false} onSend=${handleSend} runtimeConfig=${runtimeConfig} sessionState=${currentSession?.state || {}} queuedCount=${queuedCount} /></div><div class="side-column"><${TurnQueue} turns=${turns} onCancel=${handleCancel} /></div></div></section></main></div>`;
}

render(html`<${App} />`, document.getElementById('app'));
