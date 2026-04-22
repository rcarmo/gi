export function focusWindowBestEffort(runtime: { focus?: () => void } | null | undefined): boolean {
  try {
    runtime?.focus?.();
    return true;
  } catch (_error) {
    return false;
  }
}
