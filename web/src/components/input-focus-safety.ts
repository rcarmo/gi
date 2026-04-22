export function focusAndSelectBestEffort(
  input: { focus?: () => void; select?: () => void } | null | undefined,
): boolean {
  try {
    input?.focus?.();
    input?.select?.();
    return true;
  } catch (_error) {
    return false;
  }
}
