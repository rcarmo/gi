function normalizeRefValue(value: unknown): string {
  if (typeof value === 'string') {
    return value.trim();
  }
  if (typeof value === 'number') {
    return Number.isFinite(value) ? String(value) : '';
  }
  if (typeof value === 'bigint') {
    return String(value);
  }
  return '';
}

export function appendUniqueStringRef(
  previous: string[] | null | undefined,
  value: unknown,
): string[] {
  const current = Array.isArray(previous) ? previous : [];
  const normalized = normalizeRefValue(value);
  if (!normalized) {
    return current;
  }

  if (current.includes(normalized)) {
    return current;
  }

  return [...current, normalized];
}

export function removeStringRef(
  previous: string[] | null | undefined,
  value: unknown,
): string[] {
  const current = Array.isArray(previous) ? previous : [];
  const normalized = normalizeRefValue(value);
  if (!normalized) {
    return current;
  }

  if (!current.includes(normalized)) {
    return current;
  }

  return current.filter((item) => item !== normalized);
}

export function normalizeComposeRefs(next: unknown): string[] {
  if (!Array.isArray(next)) {
    return [];
  }

  const deduped: string[] = [];
  const seen = new Set<string>();

  for (const value of next) {
    const normalized = normalizeRefValue(value);
    if (!normalized || seen.has(normalized)) continue;
    seen.add(normalized);
    deduped.push(normalized);
  }

  return deduped;
}
