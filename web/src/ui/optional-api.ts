// @ts-nocheck

/**
 * Return a stable async fallback for an optional API export and warn once it is missing.
 */
export function missingOptionalApi(name, fallback) {
  if (typeof window !== 'undefined') {
    console.warn(`[app] API export missing: ${name}. Using fallback behavior.`);
  }
  return async () => fallback;
}

/**
 * Resolve an optional function export from an API namespace, falling back cleanly when absent.
 */
export function resolveOptionalApi(apiNamespace, name, fallback) {
  const candidate = apiNamespace?.[name];
  return typeof candidate === 'function' ? candidate : missingOptionalApi(name, fallback);
}
