import type { PaneHostAttachContext, PaneInstance } from './pane-types.js';

export interface PaneLiveTransferRecord {
  panePath: string;
  paneInstanceId: string;
  paneWindowId: string;
  instance: PaneInstance;
  releaseSourceHost?: () => void;
  registeredAt?: number;
}

export interface PaneLiveTransferClaim {
  panePath: string;
  paneInstanceId: string;
  paneWindowId: string;
}

const REGISTRY_KEY = '__piclawPaneLiveTransferRegistry__';

function normalizeText(value: unknown): string {
  return typeof value === 'string' ? value.trim() : '';
}

function isMapLike(value: unknown): value is Map<string, PaneLiveTransferRecord> {
  return Boolean(value)
    && typeof (value as Map<string, PaneLiveTransferRecord>).get === 'function'
    && typeof (value as Map<string, PaneLiveTransferRecord>).set === 'function'
    && typeof (value as Map<string, PaneLiveTransferRecord>).delete === 'function'
    && typeof (value as Map<string, PaneLiveTransferRecord>).entries === 'function';
}

function getRegistry(runtimeWindow: any): Map<string, PaneLiveTransferRecord> | null {
  if (!runtimeWindow) return null;
  const existing = runtimeWindow[REGISTRY_KEY];
  if (isMapLike(existing)) return existing as Map<string, PaneLiveTransferRecord>;
  const created = new Map<string, PaneLiveTransferRecord>();
  runtimeWindow[REGISTRY_KEY] = created;
  return created;
}

function releaseSourceHostBestEffort(releaseSourceHost: (() => void) | null | undefined): boolean {
  try {
    releaseSourceHost?.();
    return true;
  } catch (_error) {
    return false;
  }
}

export function registerPaneLiveTransfer(record: PaneLiveTransferRecord, runtimeWindow: any = typeof window !== 'undefined' ? window : null): boolean {
  const registry = getRegistry(runtimeWindow);
  const panePath = normalizeText(record?.panePath);
  const paneInstanceId = normalizeText(record?.paneInstanceId);
  const paneWindowId = normalizeText(record?.paneWindowId);
  if (!registry || !panePath || !paneInstanceId || !paneWindowId || typeof record?.instance?.moveHost !== 'function') {
    return false;
  }

  registry.set(paneWindowId, {
    ...record,
    panePath,
    paneInstanceId,
    paneWindowId,
    registeredAt: typeof record?.registeredAt === 'number' ? record.registeredAt : Date.now(),
  });
  return true;
}

export function clearPaneLiveTransferForPath(panePath: string, runtimeWindow: any = typeof window !== 'undefined' ? window : null): void {
  const registry = getRegistry(runtimeWindow);
  const normalizedPath = normalizeText(panePath);
  if (!registry || !normalizedPath) return;
  for (const [key, value] of registry.entries()) {
    if (value?.panePath === normalizedPath) {
      registry.delete(key);
    }
  }
}

export async function claimPaneLiveTransfer(
  sourceWindow: any,
  claim: PaneLiveTransferClaim,
  container: HTMLElement,
  context: PaneHostAttachContext,
): Promise<PaneInstance | null> {
  const registry = getRegistry(sourceWindow);
  const panePath = normalizeText(claim?.panePath);
  const paneInstanceId = normalizeText(claim?.paneInstanceId);
  const paneWindowId = normalizeText(claim?.paneWindowId);
  if (!registry || !panePath || !paneInstanceId || !paneWindowId) return null;

  const record = registry.get(paneWindowId) || null;
  if (!record) return null;
  if (record.panePath !== panePath || record.paneInstanceId !== paneInstanceId) return null;
  if (typeof record.instance?.moveHost !== 'function') return null;

  const moved = await record.instance.moveHost(container, context);
  if (!moved) return null;

  registry.delete(paneWindowId);
  releaseSourceHostBestEffort(record.releaseSourceHost);
  return record.instance;
}
