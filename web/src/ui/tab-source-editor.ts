export const SOURCE_EDITABLE_PANE_IDS = new Set(['html-viewer', 'kanban-editor', 'mindmap-editor']);

export interface ResolvePaneContext {
  path: string;
  mode: 'edit';
}

export interface ResolvedPaneLike {
  id?: string | null;
}

export type ResolvePaneLike = (context: ResolvePaneContext) => ResolvedPaneLike | null | undefined;

export function resolveEffectiveTabPaneId(
  path: string | null | undefined,
  paneOverrideId: string | null | undefined,
  resolvePane?: ResolvePaneLike | null,
): string | null {
  const normalized = String(path || '').trim();
  if (!normalized) return null;
  if (paneOverrideId) return paneOverrideId;
  if (typeof resolvePane !== 'function') return null;
  const resolved = resolvePane({ path: normalized, mode: 'edit' });
  return resolved?.id || null;
}

export function canTabEditSource(
  path: string | null | undefined,
  paneOverrideId: string | null | undefined,
  resolvePane?: ResolvePaneLike | null,
): boolean {
  const paneId = resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane);
  return paneId != null && SOURCE_EDITABLE_PANE_IDS.has(paneId);
}

export function getTabEditSourceLabel(
  path: string | null | undefined,
  paneOverrideId: string | null | undefined,
  resolvePane?: ResolvePaneLike | null,
): string {
  const paneId = resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane);
  return paneId === 'html-viewer' ? 'Edit' : 'Edit Source';
}
