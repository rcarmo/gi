// @ts-nocheck
import { html, render, useState, useEffect } from './vendor/preact-htm-entry.ts';
import { getRuntimeConfig, listSessions, createSession, listMessages, sendPrompt, listTurns, listTurnEvents, cancelTurn } from './api.ts';
import { recordStatus, bindStatusListener } from './status.ts';
import { initTheme, cycleThemePreset, getCurrentThemeLabel, cycleTint, getCurrentTintLabel } from './ui/theme.ts';
import { installFrontendLogging } from './ui/frontend-log.ts';
import { ComposeBox } from './components/compose-box.ts';
import { StatusBar } from './components/status.ts';
import { Post } from './components/post.ts';
import { WorkspaceBrowser } from './components/workspace-browser.ts';
import { PaneShell } from './components/pane-shell.ts';
import { resolveActiveTurn, resolveTurnStatusPresentation } from './ui/turn-status.ts';

const ICONS = {
  assistant: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2a3 3 0 0 0-3 3v1H8a4 4 0 0 0-4 4v3a4 4 0 0 0 4 4h1v1a3 3 0 0 0 6 0v-1h1a4 4 0 0 0 4-4v-3a4 4 0 0 0-4-4h-1V5a3 3 0 0 0-3-3Z"/><path d="M9 11h.01M15 11h.01"/></svg>`,
  user: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21a8 8 0 0 0-16 0"/><circle cx="12" cy="7" r="4"/></svg>`,
  system: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 6h13M8 12h13M8 18h13"/><path d="M3 6h.01M3 12h.01M3 18h.01"/></svg>`,
  model: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h6v6H9z"/></svg>`,
  theme: html`<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`,
};

function SessionList({ sessions, currentSessionId, onSelect, onCreate }) {
  const [title, setTitle] = useState('');
  return html`<aside class="sidebar"><div class="panel-title">Sessions</div><div class="new-session-row"><input value=${title} onInput=${(e) => setTitle(e.target.value)} placeholder="New session title" /><button onClick=${async () => { const created = await onCreate(title); setTitle(''); if (created?.id) onSelect(created.id); }}>New</button></div><div class="session-list">${sessions.map((session) => html`<button class=${`session-item ${currentSessionId === session.id ? 'active' : ''}`} onClick=${() => onSelect(session.id)}><span class="session-title">${session.title || session.id}</span><span class="session-meta">${session.state?.status || 'idle'}</span></button>`)}</div></aside>`;
}

function Timeline({ messages }) {
  return html`<section class="timeline normal"><div class="timeline-content">${messages.length === 0 ? html`<div class="empty-state">No messages yet.</div>` : null}${messages.map((msg) => html`<${Post} msg=${msg} icon=${msg.role === 'assistant' ? ICONS.assistant : msg.role === 'user' ? ICONS.user : ICONS.system} />`)}</div></section>`;
}

function TurnQueue({ turns, onCancel }) {
  return html`<aside class="gi-turn-queue"><div class="panel-title">Turns</div>${turns.length === 0 ? html`<div class="empty-state">No turns yet.</div>` : turns.map((turn) => html`<div class="turn-item"><div><div class="turn-status turn-status-${turn.status}">${turn.status}</div><div class="turn-prompt">${turn.prompt}</div></div>${turn.status === 'queued' || turn.status === 'running' || turn.status === 'cancelling' ? html`<button class="chat-window-header-button" onClick=${() => onCancel(turn.id)}>Cancel</button>` : null}</div>`)}</aside>`;
}

function App() {
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [messages, setMessages] = useState([]);
  const [turns, setTurns] = useState([]);
  const [status, setStatus] = useState('Ready.');
  const [runtimeConfig, setRuntimeConfig] = useState({ assistant_name: 'Gi', user_name: 'User', default_model: '', default_provider: '', default_thinking_level: '' });
  const [themeLabel, setThemeLabel] = useState('Default');
  const [tintLabel, setTintLabel] = useState('Off');
  const [panePath, setPanePath] = useState('');
  const [paneContent, setPaneContent] = useState('');
  const [activeTurnEvents, setActiveTurnEvents] = useState([]);

  useEffect(() => {
    installFrontendLogging();
    bindStatusListener(setStatus);
    const cleanupTheme = initTheme();
    setThemeLabel(getCurrentThemeLabel());
    setTintLabel(getCurrentTintLabel());
    refreshRuntimeConfig();
    refreshSessions();
    console.info('[app] mounted');
    return () => { cleanupTheme?.(); console.info('[app] unmounted'); };
  }, []);

  useEffect(() => {
    if (!currentSessionId) return;
    const timer = setInterval(async () => {
      await Promise.all([refreshMessages(currentSessionId), refreshTurns(currentSessionId), refreshSessions()]);
    }, 1000);
    console.debug('[app] polling started', { currentSessionId });
    return () => { clearInterval(timer); console.debug('[app] polling stopped', { currentSessionId }); };
  }, [currentSessionId]);

  async function refreshRuntimeConfig() { const data = await getRuntimeConfig(); setRuntimeConfig(data || {}); }
  async function refreshSessions() { const data = await listSessions(); setSessions(data.sessions || []); }
  async function refreshMessages(sessionID = currentSessionId) { if (!sessionID) return setMessages([]); const data = await listMessages(sessionID); setMessages(data.messages || []); }
  async function refreshTurns(sessionID = currentSessionId) {
    if (!sessionID) { setTurns([]); setActiveTurnEvents([]); return; }
    const data = await listTurns(sessionID);
    const items = data.turns || [];
    setTurns(items);
    const active = resolveActiveTurn(items);
    if (active?.id) {
      const ev = await listTurnEvents(active.id).catch(() => ({ events: [] }));
      setActiveTurnEvents(ev.events || []);
    } else {
      setActiveTurnEvents([]);
    }
  }
  async function handleCreate(title) { const created = await createSession(title || 'Untitled'); await refreshSessions(); await refreshMessages(created.id); await refreshTurns(created.id); setCurrentSessionId(created.id); recordStatus(`Created ${created.id}`); return created; }
  async function handleSelect(sessionID) { setCurrentSessionId(sessionID); await refreshMessages(sessionID); await refreshTurns(sessionID); recordStatus(`Opened ${sessionID}`); }
  async function handleSend(prompt) { if (!currentSessionId) return recordStatus('Create or open a session first.'); const result = await sendPrompt(currentSessionId, prompt); await refreshTurns(currentSessionId); recordStatus(result.queued ? `Queued ${result.turn_id}` : `Started ${result.turn_id}`); }
  async function handleCancel(turnID) { await cancelTurn(turnID); await refreshTurns(currentSessionId); recordStatus(`Cancelling ${turnID}`); }
  const currentSession = sessions.find((s) => s.id === currentSessionId) || null;
  const queuedCount = turns.filter((t) => t.status === 'queued').length;
  const activeTurn = resolveActiveTurn(turns);
  const statusPresentation = resolveTurnStatusPresentation(activeTurn, activeTurnEvents);
  const isBusy = turns.some((t) => t.status === 'running' || t.status === 'cancelling');

  return html`<div class="app-shell"><${WorkspaceBrowser} onOpenFile=${(path, content) => { setPanePath(path); setPaneContent(content); }} /><${SessionList} sessions=${sessions} currentSessionId=${currentSessionId} onSelect=${handleSelect} onCreate=${handleCreate} /><main class="container"><section class="chat-window"><header class="chat-window-header"><div class="chat-window-header-main"><div class="chat-window-header-title">${runtimeConfig.assistant_name || 'Gi'}</div><div class="chat-window-header-subtitle">${runtimeConfig.default_provider || 'provider'} / ${runtimeConfig.default_model || 'model'}${runtimeConfig.default_thinking_level ? ` · ${runtimeConfig.default_thinking_level}` : ''}</div></div><div class="chat-window-header-actions"><span class="chat-window-header-badge">${runtimeConfig.user_name || 'User'}</span><span class="chat-window-header-badge">${ICONS.model}${currentSession?.state?.model || runtimeConfig.default_model || 'model'}</span><button class="chat-window-header-button" onClick=${() => { cycleThemePreset(); setThemeLabel(getCurrentThemeLabel()); }}>${ICONS.theme}${themeLabel}</button><button class="chat-window-header-button" onClick=${() => { cycleTint(); setTintLabel(getCurrentTintLabel()); }}>tint ${tintLabel}</button></div></header><${StatusBar} title=${statusPresentation.title || status} detail=${statusPresentation.detail} tone=${statusPresentation.tone} queuedCount=${queuedCount} /><div class="main-grid"><div class="chat-column"><${Timeline} messages=${messages} /><${ComposeBox} disabled=${isBusy} onSend=${handleSend} runtimeConfig=${runtimeConfig} sessionState=${currentSession?.state || {}} queuedCount=${queuedCount} /></div><div class="side-column"><${TurnQueue} turns=${turns} onCancel=${handleCancel} /><${PaneShell} filePath=${panePath} fileContent=${paneContent} /></div></div></section></main></div>`;
}

render(html`<${App} />`, document.getElementById('app'));
