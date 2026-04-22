const EXTENSION_UI_EVENT_TYPES = new Set([
  'extension_ui_request',
  'extension_ui_timeout',
  'extension_ui_notify',
  'extension_ui_status',
  'extension_ui_working',
  'extension_ui_working_indicator',
  'extension_ui_widget',
  'extension_ui_title',
  'extension_ui_editor_text',
  'extension_ui_error',
]);

export const EXTENSION_UI_BROWSER_EVENT = 'piclaw-extension-ui';

export function isExtensionUiEventType(eventType) {
  return EXTENSION_UI_EVENT_TYPES.has(String(eventType || '').trim());
}

export function toExtensionUiBrowserEventName(eventType) {
  const normalized = String(eventType || '').trim();
  if (!normalized.startsWith('extension_ui_')) return EXTENSION_UI_BROWSER_EVENT;
  return `${EXTENSION_UI_BROWSER_EVENT}:${normalized.slice('extension_ui_'.length).replace(/_/g, '-')}`;
}

export function dispatchExtensionUiBrowserEvent(eventType, payload, target = globalThis.window) {
  if (!target || typeof target.dispatchEvent !== 'function' || typeof CustomEvent === 'undefined') {
    return false;
  }
  const detail = { type: eventType, payload };
  target.dispatchEvent(new CustomEvent(EXTENSION_UI_BROWSER_EVENT, { detail }));
  target.dispatchEvent(new CustomEvent(toExtensionUiBrowserEventName(eventType), { detail }));
  return true;
}
