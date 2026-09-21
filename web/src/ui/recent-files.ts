export const RECENT_FILES_KEY = 'piclaw_recent_files';
export const MAX_RECENT_FILES = 5;

function isIgnoredPath(path: string): boolean {
  const normalizedPath = path.trim();
  const legacyPath = normalizedPath.replace(/^\/+/, '');
  return legacyPath.startsWith('__terminal')
    || legacyPath.startsWith('__vnc')
    || normalizedPath.startsWith('piclaw://terminal')
    || normalizedPath.startsWith('piclaw://vnc');
}

function normalizeRecentFiles(value: unknown): string[] {
  if (!Array.isArray(value)) return [];

  const files: string[] = [];
  const seen = new Set<string>();

  for (const entry of value) {
    if (typeof entry !== 'string') continue;

    const path = entry.trim();
    if (!path || isIgnoredPath(path) || seen.has(path)) continue;

    seen.add(path);
    files.push(path);

    if (files.length >= MAX_RECENT_FILES) break;
  }

  return files;
}

export function getRecentFiles(): string[] {
  try {
    if (typeof localStorage === 'undefined') return [];
    const raw = localStorage.getItem(RECENT_FILES_KEY);
    if (!raw) return [];
    return normalizeRecentFiles(JSON.parse(raw));
  } catch (error) {
    console.warn('[recent-files] Failed to read recent files from localStorage.', error);
    return [];
  }
}

export function addRecentFile(path: string): void {
  const normalizedPath = typeof path === 'string' ? path.trim() : '';
  if (!normalizedPath || isIgnoredPath(normalizedPath)) return;

  try {
    if (typeof localStorage === 'undefined') return;
    const next = normalizeRecentFiles([normalizedPath, ...getRecentFiles()]);
    localStorage.setItem(RECENT_FILES_KEY, JSON.stringify(next));
  } catch (error) {
    console.warn('[recent-files] Failed to persist recent files to localStorage.', error);
  }
}
