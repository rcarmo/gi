// Render a simple Markdown summary for the current session.
const info = gi.getSessionInfo();
const session = info.session || {};
const state = gi.getSessionState();
const turns = gi.listTurns();
const messages = info.messages || [];
const runtime = gi.getRuntimeConfig();

const lines = [
  `# Session ${session.title || session.id || gi.sessionId}`,
  '',
  `- Model: ${state.model || runtime.default_model || info.runtime?.default_model || 'unknown'}`,
  `- Provider: ${state.provider || runtime.default_provider || info.runtime?.default_provider || 'unknown'}`,
  `- Thinking: ${state.thinking_level || runtime.default_thinking_level || info.runtime?.default_thinking_level || 'unknown'}`,
  `- Status: ${state.status || 'unknown'}`,
  `- Messages: ${messages.length}`,
  `- Turns: ${turns.length}`,
];

if (messages.length > 0) {
  lines.push('', '## Recent messages');
  for (const msg of messages.slice(-5)) {
    lines.push(`- **${msg.role}**: ${String(msg.content || '').slice(0, 80)}`);
  }
}

lines.join('\n');
