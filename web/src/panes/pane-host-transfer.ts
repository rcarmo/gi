import { consumePanePopoutTransferToken } from './editor-popout-transfer.js';

const PANE_HOST_TRANSFER_PREFIX = 'piclaw:pane-host-transfer:';
const PANE_HOST_TRANSFER_TTL_MS = 5 * 60 * 1000;

export interface PaneHostTransferEnvelope {
  path: string;
  payload: Record<string, unknown>;
  capturedAt?: number;
}

function getStorage(runtime: any): Storage | null {
  try {
    return runtime?.localStorage ?? null;
  } catch {
    return null;
  }
}

function normalizePath(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function normalizePayload(value: unknown): Record<string, unknown> | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null;
  return value as Record<string, unknown>;
}

function createToken(nowMs = Date.now()): string {
  return `pane-transfer-${nowMs.toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

function removeStorageItemBestEffort(storage: Storage | null, key: string): boolean {
  try {
    storage?.removeItem(key);
    return true;
  } catch (_error) {
    return false;
  }
}

export function stashPaneHostTransferState(state: PaneHostTransferEnvelope, runtime: any = globalThis, nowMs = Date.now()): string | null {
  const storage = getStorage(runtime);
  const path = normalizePath(state?.path);
  const payload = normalizePayload(state?.payload);
  if (!storage || !path || !payload) return null;

  const token = createToken(nowMs);
  try {
    storage.setItem(`${PANE_HOST_TRANSFER_PREFIX}${token}`, JSON.stringify({
      path,
      payload,
      capturedAt: nowMs,
    }));
    return token;
  } catch {
    return null;
  }
}

export function consumePaneHostTransferState(token?: string | null, runtime: any = globalThis, nowMs = Date.now()): PaneHostTransferEnvelope | null {
  const normalizedToken = typeof token === 'string' ? token.trim() : '';
  const storage = getStorage(runtime);
  if (!normalizedToken || !storage) return null;

  const key = `${PANE_HOST_TRANSFER_PREFIX}${normalizedToken}`;
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
    if (capturedAt + PANE_HOST_TRANSFER_TTL_MS < nowMs) {
      return null;
    }

    const path = normalizePath(parsed?.path);
    const payload = normalizePayload(parsed?.payload);
    if (!path || !payload) return null;

    return {
      path,
      payload,
      capturedAt,
    };
  } catch {
    return null;
  }
}

export function createPaneHostTransferPayload(state: PaneHostTransferEnvelope, runtime: any = globalThis, nowMs = Date.now()): Record<string, string> | null {
  const token = stashPaneHostTransferState(state, runtime, nowMs);
  return token ? { pane_transfer: token } : null;
}

export function consumePaneHostTransferFromLocation(runtime: any = globalThis, nowMs = Date.now()): PaneHostTransferEnvelope | null {
  const token = consumePanePopoutTransferToken('pane_transfer', runtime);
  return consumePaneHostTransferState(token, runtime, nowMs);
}
