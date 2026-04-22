// Local storage helpers (safe for SSR / restricted contexts)

export function getLocalStorageItem(key) {
  if (typeof window === 'undefined' || !window.localStorage) return null;
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function setLocalStorageItem(key, value) {
  if (typeof window === 'undefined' || !window.localStorage) return;
  try {
    window.localStorage.setItem(key, value);
  } catch {
    return;
  }
}

export function getLocalStorageBoolean(key, defaultValue = false) {
  const raw = getLocalStorageItem(key);
  if (raw === null) return defaultValue;
  return raw === 'true';
}

export function getLocalStorageNumber(key, defaultValue = null) {
  const raw = getLocalStorageItem(key);
  if (raw === null) return defaultValue;
  const parsed = parseInt(raw, 10);
  return Number.isFinite(parsed) ? parsed : defaultValue;
}

export function getLocalStorageJSON<T = unknown>(key: string): T | null {
  const raw = getLocalStorageItem(key);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as T;
  } catch {
    return null;
  }
}
