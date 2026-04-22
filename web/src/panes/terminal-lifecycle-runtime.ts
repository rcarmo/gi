function runBestEffort(run: () => void): boolean {
  try {
    run();
    return true;
  } catch (_error) {
    return false;
  }
}

export function clearContainerContentBestEffort(
  container: { innerHTML?: string } | null | undefined,
): boolean {
  return runBestEffort(() => {
    if (container) {
      container.innerHTML = '';
    }
  });
}

export function detachTerminalHostListenersBestEffort(options: {
  ownerWindow?: { removeEventListener?: (type: string, listener: EventListenerOrEventListenerObject) => void } | null;
  themeChangeListener?: EventListenerOrEventListenerObject | null;
  mediaQuery?: {
    removeEventListener?: (type: string, listener: EventListenerOrEventListenerObject) => void;
    removeListener?: (listener: EventListenerOrEventListenerObject) => void;
  } | null;
  mediaQueryListener?: EventListenerOrEventListenerObject | null;
  dockResizeListener?: EventListenerOrEventListenerObject | null;
  windowResizeListener?: EventListenerOrEventListenerObject | null;
  themeObserver?: { disconnect?: () => void } | null;
  resizeObserver?: { disconnect?: () => void } | null;
}): void {
  const {
    ownerWindow,
    themeChangeListener,
    mediaQuery,
    mediaQueryListener,
    dockResizeListener,
    windowResizeListener,
    themeObserver,
    resizeObserver,
  } = options;

  runBestEffort(() => {
    if (themeChangeListener) {
      ownerWindow?.removeEventListener?.('piclaw-theme-change', themeChangeListener);
    }
  });
  runBestEffort(() => {
    if (mediaQuery && mediaQueryListener) {
      if (mediaQuery.removeEventListener) mediaQuery.removeEventListener('change', mediaQueryListener);
      else if (mediaQuery.removeListener) mediaQuery.removeListener(mediaQueryListener);
    }
  });
  runBestEffort(() => {
    if (dockResizeListener) {
      ownerWindow?.removeEventListener?.('dock-resize', dockResizeListener);
    }
    if (windowResizeListener) {
      ownerWindow?.removeEventListener?.('resize', windowResizeListener);
    }
  });
  runBestEffort(() => {
    themeObserver?.disconnect?.();
  });
  runBestEffort(() => {
    resizeObserver?.disconnect?.();
  });
}

export function resizeTerminalRuntimeBestEffort(options: {
  syncHostLayout: () => void;
  terminal?: any;
  fitAddon?: any;
  sendResize: () => void;
}): void {
  options.syncHostLayout();
  runBestEffort(() => {
    options.terminal?.renderer?.remeasureFont?.();
  });
  runBestEffort(() => {
    options.fitAddon?.fit?.();
  });
  runBestEffort(() => {
    options.terminal?.refresh?.();
  });
  options.syncHostLayout();
  options.sendResize();
}

export function disposeTerminalRuntimeBestEffort(options: {
  resizeFrame?: number;
  cancelAnimationFrameFn?: (handle: number) => void;
  socket?: { close?: () => void } | null;
  fitAddon?: { dispose?: () => void } | null;
  terminal?: { dispose?: () => void } | null;
  termEl?: { remove?: () => void } | null;
}): number {
  const {
    resizeFrame = 0,
    cancelAnimationFrameFn = cancelAnimationFrame,
    socket,
    fitAddon,
    terminal,
    termEl,
  } = options;

  if (resizeFrame) {
    runBestEffort(() => {
      cancelAnimationFrameFn(resizeFrame);
    });
  }
  runBestEffort(() => {
    socket?.close?.();
  });
  runBestEffort(() => {
    fitAddon?.dispose?.();
  });
  runBestEffort(() => {
    terminal?.dispose?.();
  });
  termEl?.remove?.();
  return 0;
}
