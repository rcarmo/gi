export function closeWindowBestEffort(
  target: { close?: () => void } | null | undefined,
): boolean {
  try {
    target?.close?.();
    return true;
  } catch (_error) {
    return false;
  }
}

export function postMessageToWindowBestEffort(
  target: { postMessage?: (message: unknown, targetOrigin: string) => void } | null | undefined,
  message: unknown,
  targetOrigin: string,
): boolean {
  try {
    target?.postMessage?.(message, targetOrigin);
    return true;
  } catch (_error) {
    return false;
  }
}
