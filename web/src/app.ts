// @ts-nocheck
import { html, render, useState, useEffect } from './vendor/preact-htm-entry.ts';
import { listSessions, createSession, listMessages, sendPrompt } from './api.ts';
import { bindStatusListener, recordStatus } from './status.ts';

function SessionList({ sessions, currentSessionId, onSelect, onCreate }) {
  const [title, setTitle] = useState('');
  return html`
    <aside class="sidebar">
      <div class="panel-title">Sessions</div>
      <div class="new-session-row">
        <input value=${title} onInput=${(e) => setTitle(e.target.value)} placeholder="New session title" />
        <button onClick=${async () => {
          const created = await onCreate(title);
          setTitle('');
          if (created?.id) onSelect(created.id);
        }}>New</button>
      </div>
      <div class="session-list">
        ${sessions.map((session) => html`
          <button
            class=${`session-item ${currentSessionId === session.id ? 'active' : ''}`}
            onClick=${() => onSelect(session.id)}>
            <span class="session-title">${session.title || session.id}</span>
            <span class="session-meta">${session.state?.status || 'idle'}</span>
          </button>
        `)}
      </div>
    </aside>
  `;
}

function Timeline({ messages }) {
  return html`
    <section class="timeline">
      ${messages.length === 0 ? html`<div class="empty-state">No messages yet.</div>` : null}
      ${messages.map((msg) => html`
        <article class=${`message message-${msg.role}`}>
          <div class="message-role">${msg.role}</div>
          <div class="message-content">${msg.content}</div>
        </article>
      `)}
    </section>
  `;
}

function ComposeBox({ disabled, onSend }) {
  const [text, setText] = useState('');
  return html`
    <div class="compose-shell">
      <textarea
        rows="5"
        placeholder="Send a prompt"
        value=${text}
        onInput=${(e) => setText(e.target.value)}
      />
      <div class="compose-actions">
        <button disabled=${disabled || !text.trim()} onClick=${async () => {
          const payload = text;
          setText('');
          await onSend(payload);
        }}>Send</button>
      </div>
    </div>
  `;
}

function StatusBar({ text }) {
  return html`<div class="status-bar">${text || 'Ready.'}</div>`;
}

function App() {
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [messages, setMessages] = useState([]);
  const [status, setStatus] = useState('Ready.');
  const [running, setRunning] = useState(false);

  useEffect(() => {
    bindStatusListener(setStatus);
    refreshSessions();
  }, []);

  async function refreshSessions() {
    const data = await listSessions();
    setSessions(data.sessions || []);
  }

  async function refreshMessages(sessionID = currentSessionId) {
    if (!sessionID) {
      setMessages([]);
      return;
    }
    const data = await listMessages(sessionID);
    setMessages(data.messages || []);
  }

  async function handleCreate(title) {
    const created = await createSession(title || 'Untitled');
    await refreshSessions();
    await refreshMessages(created.id);
    setCurrentSessionId(created.id);
    recordStatus(`Created ${created.id}`);
    return created;
  }

  async function handleSelect(sessionID) {
    setCurrentSessionId(sessionID);
    await refreshMessages(sessionID);
    recordStatus(`Opened ${sessionID}`);
  }

  async function handleSend(prompt) {
    if (!currentSessionId) {
      recordStatus('Create or open a session first.');
      return;
    }
    setRunning(true);
    try {
      const result = await sendPrompt(currentSessionId, prompt);
      await refreshMessages(currentSessionId);
      recordStatus(`Turn ${result.turn_id} finished: ${result.status}`);
    } catch (err) {
      recordStatus(`Error: ${err.message || err}`);
    } finally {
      setRunning(false);
    }
  }

  return html`
    <div class="app-shell">
      <${SessionList}
        sessions=${sessions}
        currentSessionId=${currentSessionId}
        onSelect=${handleSelect}
        onCreate=${handleCreate}
      />
      <main class="main-pane">
        <header class="topbar">
          <div>
            <h1>Gi</h1>
            <div class="subtitle">Preat shell imported from Piclaw/Vibes conventions</div>
          </div>
        </header>
        <${StatusBar} text=${status} />
        <${Timeline} messages=${messages} />
        <${ComposeBox} disabled=${running} onSend=${handleSend} />
      </main>
    </div>
  `;
}

render(html`<${App} />`, document.getElementById('app'));
