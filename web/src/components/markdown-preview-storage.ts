export function readStoredPanelHeight(
  storage: { getItem: (key: string) => string | null } | null | undefined,
  key: string,
  minHeight: number,
  defaultHeight: number,
): number {
  try {
    const value = storage?.getItem?.(key);
    const parsed = value ? Number(value) : Number.NaN;
    return Number.isFinite(parsed) && parsed >= minHeight ? parsed : defaultHeight;
  } catch (_error) {
    return defaultHeight;
  }
}

export function writeStoredPanelHeight(
  storage: { setItem: (key: string, value: string) => void } | null | undefined,
  key: string,
  height: number,
): boolean {
  try {
    storage?.setItem?.(key, String(Math.round(height)));
    return true;
  } catch (_error) {
    return false;
  }
}
