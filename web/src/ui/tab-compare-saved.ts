import { resolveEffectiveTabPaneId, type ResolvePaneLike } from './tab-source-editor.js';

export function canTabCompareToSaved(
  path: string | null | undefined,
  paneOverrideId: string | null | undefined,
  resolvePane?: ResolvePaneLike | null,
): boolean {
  return resolveEffectiveTabPaneId(path, paneOverrideId, resolvePane) === 'editor';
}
