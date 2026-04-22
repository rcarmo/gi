export interface PaneOwnershipState {
  panePath: string;
  paneInstanceId: string;
  ownerWindowId: string;
  detachedAt: string;
  label?: string | null;
}

export interface PendingPaneOwnershipState extends PaneOwnershipState {
  sourceWindowId?: string | null;
  requestedAt: string;
}

export interface PaneDetachClaim {
  panePath?: string | null;
  paneInstanceId?: string | null;
  paneWindowId?: string | null;
}

function normalizeText(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

export function createPendingPaneOwnershipState(input: {
  panePath: string;
  paneInstanceId: string;
  ownerWindowId: string;
  label?: string | null;
  sourceWindowId?: string | null;
  now?: string;
}): PendingPaneOwnershipState | null {
  const panePath = normalizeText(input?.panePath);
  const paneInstanceId = normalizeText(input?.paneInstanceId);
  const ownerWindowId = normalizeText(input?.ownerWindowId);
  if (!panePath || !paneInstanceId || !ownerWindowId) return null;

  const timestamp = normalizeText(input?.now) || new Date().toISOString();
  return {
    panePath,
    paneInstanceId,
    ownerWindowId,
    detachedAt: timestamp,
    requestedAt: timestamp,
    label: normalizeText(input?.label),
    sourceWindowId: normalizeText(input?.sourceWindowId),
  };
}

export function matchesPaneDetachClaim(
  pending: PendingPaneOwnershipState | null | undefined,
  claim: PaneDetachClaim | null | undefined,
): boolean {
  if (!pending || !claim) return false;
  return normalizeText(claim.panePath) === pending.panePath
    && normalizeText(claim.paneInstanceId) === pending.paneInstanceId
    && normalizeText(claim.paneWindowId) === pending.ownerWindowId;
}

export function finalizePendingPaneOwnership(
  pending: PendingPaneOwnershipState | null | undefined,
  now?: string | null,
): PaneOwnershipState | null {
  if (!pending) return null;
  return {
    panePath: pending.panePath,
    paneInstanceId: pending.paneInstanceId,
    ownerWindowId: pending.ownerWindowId,
    detachedAt: normalizeText(now) || new Date().toISOString(),
    label: pending.label || null,
  };
}
