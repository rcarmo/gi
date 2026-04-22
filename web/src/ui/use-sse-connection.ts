// @ts-nocheck
import { useEffect, useRef } from '../vendor/preact-htm.js';
import { SSEClient } from '../api.js';
import { isIOSDevice } from './app-helpers.js';

export function bindSseWakeLifecycle({ sse, onWake }, runtime = {}) {
  const win = runtime.window ?? (typeof window !== 'undefined' ? window : null);
  const doc = runtime.document ?? (typeof document !== 'undefined' ? document : null);
  if (!win || !doc || !sse) {
    return () => {};
  }

  const reconnectAfterReturn = () => {
    if (typeof sse.forceReconnect === 'function') {
      sse.forceReconnect();
      return;
    }
    sse.reconnectIfNeeded();
  };

  const shouldUseFocusReconnect = typeof runtime.useFocusReconnect === 'boolean'
    ? runtime.useFocusReconnect
    : !isIOSDevice();
  let pendingWake = doc.visibilityState && doc.visibilityState !== 'visible';

  const handleHiddenState = () => {
    if (doc.visibilityState && doc.visibilityState !== 'visible') {
      pendingWake = true;
      return true;
    }
    return false;
  };

  const handleVisibleReturn = () => {
    if (handleHiddenState()) return;
    if (pendingWake) {
      pendingWake = false;
      reconnectAfterReturn();
      onWake?.();
    }
  };

  const handleWindowFocus = () => {
    if (handleHiddenState()) return;
    if (pendingWake) {
      handleVisibleReturn();
      return;
    }
    if (shouldUseFocusReconnect) {
      sse.reconnectIfNeeded();
    }
  };

  const handlePageShow = () => {
    handleVisibleReturn();
  };

  const handleVisibilityChange = () => {
    handleVisibleReturn();
  };

  win.addEventListener('focus', handleWindowFocus);
  win.addEventListener('pageshow', handlePageShow);
  doc.addEventListener('visibilitychange', handleVisibilityChange);

  return () => {
    win.removeEventListener('focus', handleWindowFocus);
    win.removeEventListener('pageshow', handlePageShow);
    doc.removeEventListener('visibilitychange', handleVisibilityChange);
  };
}

/**
 * Manages the SSE connection lifecycle.
 *
 * All callbacks are accessed via refs so the EventSource is created exactly
 * once and never torn down due to callback identity changes in the parent
 * component.  This breaks the re-render cascade that previously caused an
 * infinite SSE reconnect loop when queue/filter state changed.
 */
export function useSseConnection({ handleSseEvent, handleConnectionStatusChange, loadPosts, onWake, chatJid }) {
  const sseEventRef = useRef(handleSseEvent);
  sseEventRef.current = handleSseEvent;

  const statusChangeRef = useRef(handleConnectionStatusChange);
  statusChangeRef.current = handleConnectionStatusChange;

  const loadPostsRef = useRef(loadPosts);
  loadPostsRef.current = loadPosts;

  const onWakeRef = useRef(onWake);
  onWakeRef.current = onWake;

  useEffect(() => {
    const sse = new SSEClient(
      (type, data) => sseEventRef.current(type, data),
      (status) => statusChangeRef.current(status),
      { chatJid },
    );
    sse.connect();

    const disposeWakeLifecycle = bindSseWakeLifecycle({
      sse,
      onWake: () => onWakeRef.current?.(),
    });

    return () => {
      disposeWakeLifecycle();
      sse.disconnect();
    };
  }, [chatJid]);
}
