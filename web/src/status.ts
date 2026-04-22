let listener = null;
export function bindStatusListener(fn) { listener = fn; }
export function recordStatus(text) { if (listener) listener(text); console.debug('[status]', text); }
