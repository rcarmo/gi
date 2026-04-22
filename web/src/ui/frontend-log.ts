// @ts-nocheck
const MAX_BUFFER = 200;
const buffer = [];
let flushing = false;

function enqueue(level, message, detail = null) {
  buffer.push({ ts: new Date().toISOString(), level, message, detail });
  if (buffer.length > MAX_BUFFER) buffer.shift();
  queueMicrotask(flush);
}

async function flush() {
  if (flushing || buffer.length === 0) return;
  flushing = true;
  const batch = buffer.splice(0, Math.min(buffer.length, 20));
  try {
    await fetch('/api/frontend/log', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ entries: batch }),
      keepalive: true,
    });
  } catch {
    buffer.unshift(...batch);
  } finally {
    flushing = false;
    if (buffer.length > 0) queueMicrotask(flush);
  }
}

export function installFrontendLogging() {
  const orig = {
    log: console.log?.bind(console),
    warn: console.warn?.bind(console),
    error: console.error?.bind(console),
    debug: console.debug?.bind(console),
  };
  console.log = (...args) => { orig.log?.(...args); enqueue('info', stringifyArgs(args), args); };
  console.warn = (...args) => { orig.warn?.(...args); enqueue('warn', stringifyArgs(args), args); };
  console.error = (...args) => { orig.error?.(...args); enqueue('error', stringifyArgs(args), args); };
  console.debug = (...args) => { orig.debug?.(...args); enqueue('debug', stringifyArgs(args), args); };
  window.addEventListener('error', (event) => enqueue('error', event.message || 'window error', { filename: event.filename, lineno: event.lineno, colno: event.colno }));
  window.addEventListener('unhandledrejection', (event) => enqueue('error', 'unhandled rejection', { reason: String(event.reason || '') }));
  enqueue('info', 'frontend logging installed');
}

function stringifyArgs(args) {
  return args.map((item) => {
    if (typeof item === 'string') return item;
    try { return JSON.stringify(item); } catch { return String(item); }
  }).join(' ');
}
