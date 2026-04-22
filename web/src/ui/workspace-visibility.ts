export const LEGACY_WORKSPACE_OPEN_STORAGE_KEY = 'workspaceOpen';
export const DESKTOP_WORKSPACE_OPEN_STORAGE_KEY = 'workspaceOpen.desktop';
export const NARROW_WORKSPACE_OPEN_STORAGE_KEY = 'workspaceOpen.narrow';
export const DESKTOP_WORKSPACE_LAYOUT_MEDIA_QUERY = '(min-width: 1024px) and (orientation: landscape)';

export type WorkspaceLayoutBucket = 'desktop' | 'narrow';

function getRuntimeWindow(runtime: any = typeof window !== 'undefined' ? window : null) {
  return runtime && typeof runtime === 'object' ? runtime : null;
}

function readRuntimeStorageBoolean(runtime: any, key: string): boolean | null {
  const runtimeWindow = getRuntimeWindow(runtime);
  if (!runtimeWindow?.localStorage?.getItem) return null;
  try {
    const raw = runtimeWindow.localStorage.getItem(key);
    if (raw === null) return null;
    return raw === 'true';
  } catch {
    return null;
  }
}

function writeRuntimeStorageBoolean(runtime: any, key: string, value: boolean): void {
  const runtimeWindow = getRuntimeWindow(runtime);
  if (!runtimeWindow?.localStorage?.setItem) return;
  try {
    runtimeWindow.localStorage.setItem(key, String(Boolean(value)));
  } catch {
    return;
  }
}

export function resolveWorkspaceLayoutBucket(runtime: any = typeof window !== 'undefined' ? window : null): WorkspaceLayoutBucket {
  const runtimeWindow = getRuntimeWindow(runtime);
  if (!runtimeWindow?.matchMedia) return 'desktop';
  return runtimeWindow.matchMedia(DESKTOP_WORKSPACE_LAYOUT_MEDIA_QUERY).matches ? 'desktop' : 'narrow';
}

export function getWorkspaceOpenStorageKey(bucket: WorkspaceLayoutBucket | null | undefined): string {
  return bucket === 'narrow' ? NARROW_WORKSPACE_OPEN_STORAGE_KEY : DESKTOP_WORKSPACE_OPEN_STORAGE_KEY;
}

export function readStoredWorkspaceOpenPreference(options: {
  runtime?: any;
  bucket?: WorkspaceLayoutBucket | null;
  allowLegacyFallback?: boolean;
  defaultValue?: boolean;
} = {}): boolean {
  const {
    runtime = typeof window !== 'undefined' ? window : null,
    bucket = null,
    allowLegacyFallback = false,
    defaultValue = false,
  } = options;

  const targetBucket = bucket || resolveWorkspaceLayoutBucket(runtime);
  const storageKey = getWorkspaceOpenStorageKey(targetBucket);
  const scopedValue = readRuntimeStorageBoolean(runtime, storageKey);
  if (typeof scopedValue === 'boolean') return scopedValue;

  if (allowLegacyFallback && targetBucket === 'desktop') {
    const legacyValue = readRuntimeStorageBoolean(runtime, LEGACY_WORKSPACE_OPEN_STORAGE_KEY);
    if (typeof legacyValue === 'boolean') return legacyValue;
  }

  return defaultValue;
}

export function persistWorkspaceOpenPreference(
  workspaceOpen: boolean,
  options: { runtime?: any; bucket?: WorkspaceLayoutBucket | null } = {},
): void {
  const {
    runtime = typeof window !== 'undefined' ? window : null,
    bucket = null,
  } = options;
  const targetBucket = bucket || resolveWorkspaceLayoutBucket(runtime);
  writeRuntimeStorageBoolean(runtime, getWorkspaceOpenStorageKey(targetBucket), Boolean(workspaceOpen));
}
