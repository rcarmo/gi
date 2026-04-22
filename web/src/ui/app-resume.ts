// @ts-nocheck

import { isStandaloneWebAppMode } from './chat-window.js';

export const DISPLAY_MODE_MEDIA_QUERIES = [
  '(display-mode: standalone)',
  '(display-mode: minimal-ui)',
  '(display-mode: fullscreen)',
];

/**
 * Watch standalone webapp mode changes and call back immediately plus on relevant resume/window events.
 */
export function watchStandaloneWebAppMode(onChange, runtime = {}) {
  const win = runtime.window ?? (typeof window !== 'undefined' ? window : null);
  const nav = runtime.navigator ?? (typeof navigator !== 'undefined' ? navigator : null);
  if (!win || typeof onChange !== 'function') {
    return () => {};
  }

  const refresh = () => {
    onChange(isStandaloneWebAppMode({ window: win, navigator: nav }));
  };

  refresh();

  const mediaQueries = DISPLAY_MODE_MEDIA_QUERIES
    .map((query) => {
      try {
        return win.matchMedia?.(query) ?? null;
      } catch {
        return null;
      }
    })
    .filter(Boolean);

  const removers = mediaQueries.map((mql) => {
    if (typeof mql.addEventListener === 'function') {
      mql.addEventListener('change', refresh);
      return () => mql.removeEventListener('change', refresh);
    }
    if (typeof mql.addListener === 'function') {
      mql.addListener(refresh);
      return () => mql.removeListener(refresh);
    }
    return () => {};
  });

  win.addEventListener?.('focus', refresh);
  win.addEventListener?.('pageshow', refresh);

  return () => {
    for (const remove of removers) remove();
    win.removeEventListener?.('focus', refresh);
    win.removeEventListener?.('pageshow', refresh);
  };
}

/**
 * Watch for return-to-app events and invoke the callback only when the document is visible.
 */
export function watchReturnToApp(onReturn, runtime = {}) {
  const win = runtime.window ?? (typeof window !== 'undefined' ? window : null);
  const doc = runtime.document ?? (typeof document !== 'undefined' ? document : null);
  if (!win || !doc || typeof onReturn !== 'function') {
    return () => {};
  }

  const handleReturnToApp = () => {
    if (doc.visibilityState && doc.visibilityState !== 'visible') return;
    onReturn();
  };

  win.addEventListener?.('focus', handleReturnToApp);
  win.addEventListener?.('pageshow', handleReturnToApp);
  doc.addEventListener?.('visibilitychange', handleReturnToApp);

  return () => {
    win.removeEventListener?.('focus', handleReturnToApp);
    win.removeEventListener?.('pageshow', handleReturnToApp);
    doc.removeEventListener?.('visibilitychange', handleReturnToApp);
  };
}
