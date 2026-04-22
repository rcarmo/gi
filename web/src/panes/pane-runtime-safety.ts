export function readRandomUuidBestEffort(runtime: { crypto?: { randomUUID?: () => string } } | null | undefined): string | null {
  try {
    return typeof runtime?.crypto?.randomUUID === 'function' ? runtime.crypto.randomUUID() : null;
  } catch (_error) {
    return null;
  }
}

export function removeStorageItemBestEffort(
  storage: { removeItem?: (key: string) => void } | null | undefined,
  key: string,
): boolean {
  try {
    storage?.removeItem?.(key);
    return true;
  } catch (_error) {
    return false;
  }
}

export function disposeSocketBoundaryBestEffort(
  boundary: { dispose?: () => void } | null | undefined,
): boolean {
  try {
    boundary?.dispose?.();
    return true;
  } catch (_error) {
    return false;
  }
}
