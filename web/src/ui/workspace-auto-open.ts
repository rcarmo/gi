export const MAX_EDITABLE_PREVIEW_BYTES = 256 * 1024;

export interface WorkspacePreviewLike {
  kind?: string | null;
  size?: number | null;
}

export interface ResolveWorkspacePaneContext {
  path: string;
  mode: 'edit';
}

export interface ResolvedWorkspacePaneLike {
  id?: string | null;
}

export type ResolveWorkspacePaneLike = (
  context: ResolveWorkspacePaneContext,
) => ResolvedWorkspacePaneLike | null | undefined;

export interface ShouldAutoOpenWorkspaceFileOptions {
  resolvePane?: ResolveWorkspacePaneLike | null;
}

export function isEditableWorkspacePreview(preview: WorkspacePreviewLike | null | undefined): boolean {
  if (!preview || preview.kind !== 'text') return false;
  const size = Number(preview.size);
  return !Number.isFinite(size) || size <= MAX_EDITABLE_PREVIEW_BYTES;
}

export function hasSpecializedWorkspaceTab(
  path: string | null | undefined,
  resolvePane?: ResolveWorkspacePaneLike | null,
): boolean {
  const normalized = String(path || '').trim();
  if (!normalized || normalized.endsWith('/')) return false;
  if (typeof resolvePane !== 'function') return false;
  const resolved = resolvePane({ path: normalized, mode: 'edit' });
  if (!resolved || typeof resolved !== 'object') return false;
  return resolved.id !== 'editor';
}

export function shouldAutoOpenWorkspaceFile(
  path: string | null | undefined,
  preview: WorkspacePreviewLike | null | undefined,
  options: ShouldAutoOpenWorkspaceFileOptions = {},
): boolean {
  const resolvePane = options.resolvePane;
  if (hasSpecializedWorkspaceTab(path, resolvePane)) return true;
  return isEditableWorkspacePreview(preview);
}
