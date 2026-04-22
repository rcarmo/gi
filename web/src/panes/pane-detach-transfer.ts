export interface PaneDetachTransferParamsInput {
  paneInstanceId: string;
  paneWindowId: string;
  paneSourceWindowId?: string | null;
}

export interface PaneDetachTransferState {
  panePath: string | null;
  paneLabel: string | null;
  paneInstanceId: string | null;
  paneWindowId: string | null;
  paneSourceWindowId: string | null;
}

function normalizeText(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

function readRandomUuidBestEffort(runtime: typeof globalThis = globalThis): string | null {
  try {
    return typeof runtime?.crypto?.randomUUID === 'function' ? runtime.crypto.randomUUID() : null;
  } catch (_error) {
    return null;
  }
}

export function generatePaneDetachId(prefix = 'pane'): string {
  const uuid = readRandomUuidBestEffort();
  if (uuid) {
    return `${prefix}-${uuid}`;
  }
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}

export function createPaneDetachTransferParams(input: PaneDetachTransferParamsInput): Record<string, string> {
  const paneInstanceId = normalizeText(input?.paneInstanceId);
  const paneWindowId = normalizeText(input?.paneWindowId);
  if (!paneInstanceId || !paneWindowId) return {};

  const paneSourceWindowId = normalizeText(input?.paneSourceWindowId);
  return {
    pane_instance_id: paneInstanceId,
    pane_window_id: paneWindowId,
    ...(paneSourceWindowId ? { pane_source_window_id: paneSourceWindowId } : {}),
  };
}

export function readPaneDetachTransferState(options: {
  search?: string | null;
  panePath?: string | null;
  paneLabel?: string | null;
} = {}): PaneDetachTransferState {
  const params = new URLSearchParams(options.search || '');
  return {
    panePath: normalizeText(params.get('pane_path')) || normalizeText(options.panePath),
    paneLabel: normalizeText(params.get('pane_label')) || normalizeText(options.paneLabel),
    paneInstanceId: normalizeText(params.get('pane_instance_id')),
    paneWindowId: normalizeText(params.get('pane_window_id')),
    paneSourceWindowId: normalizeText(params.get('pane_source_window_id')),
  };
}

export function hasPaneDetachTransferState(state: PaneDetachTransferState | null | undefined): boolean {
  return Boolean(state?.panePath && state?.paneInstanceId && state?.paneWindowId);
}
