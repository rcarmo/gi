import { removeStorageItemBestEffort } from './pane-runtime-safety.js';
import type { TabViewState } from './tab-store.js';

const EDITOR_POPOUT_STATE_PREFIX = 'piclaw:editor-popout:';
const EDITOR_POPOUT_STATE_TTL_MS = 5 * 60 * 1000;

export interface EditorPopoutTransferState {
  path: string;
  content?: string;
  mtime?: string | null;
  paneOverrideId?: string | null;
  viewState?: TabViewState | null;
  capturedAt?: number;
}

function getStorage(runtime: any): Storage | null {
  try {
    return runtime?.localStorage ?? null;
  } catch {
    return null;
  }
}

function createToken(nowMs = Date.now()): string {
  return `editor-popout-${nowMs.toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

function normalizePath(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function normalizeOverrideId(value: unknown): string | null {
  const normalized = typeof value === 'string' ? value.trim() : '';
  return normalized || null;
}

function normalizeContent(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined;
}

function normalizeMtime(value: unknown): string | null | undefined {
  if (value === undefined) return undefined;
  if (typeof value !== 'string') return null;
  const normalized = value.trim();
  return normalized || null;
}

function normalizeViewState(value: unknown): TabViewState | null {
  if (!value || typeof value !== 'object') return null;
  const raw = value as Record<string, unknown>;
  const next: TabViewState = {};
  if (typeof raw.cursorLine === 'number' && Number.isFinite(raw.cursorLine)) next.cursorLine = raw.cursorLine;
  if (typeof raw.cursorCol === 'number' && Number.isFinite(raw.cursorCol)) next.cursorCol = raw.cursorCol;
  if (typeof raw.scrollTop === 'number' && Number.isFinite(raw.scrollTop)) next.scrollTop = raw.scrollTop;
  return Object.keys(next).length > 0 ? next : null;
}

export function consumePanePopoutTransferToken(paramName: string, runtime: any = globalThis): string | null {
  const win = runtime?.window ?? runtime;
  if (!win?.location?.href) return null;
  try {
    const url = new URL(win.location.href);
    const token = url.searchParams.get(paramName)?.trim() || '';
    if (!token) return null;
    url.searchParams.delete(paramName);
    win.history?.replaceState?.(win.history.state, win.document?.title || '', url.toString());
    return token;
  } catch {
    return null;
  }
}

export function stashEditorPopoutState(state: EditorPopoutTransferState, runtime: any = globalThis, nowMs = Date.now()): string | null {
  const storage = getStorage(runtime);
  const path = normalizePath(state?.path);
  if (!storage || !path) return null;

  const payload: EditorPopoutTransferState = {
    path,
    content: normalizeContent(state?.content),
    mtime: normalizeMtime(state?.mtime),
    paneOverrideId: normalizeOverrideId(state?.paneOverrideId),
    viewState: normalizeViewState(state?.viewState),
    capturedAt: nowMs,
  };

  const hasTransferData = Boolean(
    payload.content !== undefined
      || payload.paneOverrideId
      || payload.viewState
      || payload.mtime,
  );
  if (!hasTransferData) return null;

  const token = createToken(nowMs);
  try {
    storage.setItem(`${EDITOR_POPOUT_STATE_PREFIX}${token}`, JSON.stringify(payload));
    return token;
  } catch {
    return null;
  }
}

export function consumeEditorPopoutState(token?: string | null, runtime: any = globalThis, nowMs = Date.now()): EditorPopoutTransferState | null {
  const normalizedToken = typeof token === 'string' ? token.trim() : '';
  const storage = getStorage(runtime);
  if (!normalizedToken || !storage) return null;

  const key = `${EDITOR_POPOUT_STATE_PREFIX}${normalizedToken}`;
  let raw = '';
  try {
    raw = storage.getItem(key) || '';
  } catch {
    return null;
  }
  if (!raw) return null;

  removeStorageItemBestEffort(storage, key);

  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const capturedAt = typeof parsed?.capturedAt === 'number' && Number.isFinite(parsed.capturedAt)
      ? parsed.capturedAt
      : nowMs;
    if (capturedAt + EDITOR_POPOUT_STATE_TTL_MS < nowMs) {
      return null;
    }

    const path = normalizePath(parsed?.path);
    if (!path) return null;

    return {
      path,
      content: normalizeContent(parsed?.content),
      mtime: normalizeMtime(parsed?.mtime),
      paneOverrideId: normalizeOverrideId(parsed?.paneOverrideId),
      viewState: normalizeViewState(parsed?.viewState),
      capturedAt,
    };
  } catch {
    return null;
  }
}

export function createEditorPopoutTransferPayload(state: EditorPopoutTransferState, runtime: any = globalThis, nowMs = Date.now()): Record<string, string> | null {
  const token = stashEditorPopoutState(state, runtime, nowMs);
  return token ? { editor_popout: token } : null;
}
