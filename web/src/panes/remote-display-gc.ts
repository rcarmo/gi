export function collectAssemblyScriptGarbageBestEffort(
  runtime: { __collect?: (() => void) | undefined } | null | undefined,
): boolean {
  try {
    runtime?.__collect?.();
    return true;
  } catch (_error) {
    return false;
  }
}
