export interface ComposeSessionSwitcherKeyEventLike {
  isComposing?: boolean;
  ctrlKey?: boolean;
  metaKey?: boolean;
  altKey?: boolean;
  key?: string;
}

export interface ComposeSessionSwitcherOptions {
  searchMode?: boolean;
  showSessionSwitcherButton?: boolean;
}

export function shouldOpenSessionSwitcherFromBlankCompose(
  event: ComposeSessionSwitcherKeyEventLike | null | undefined,
  value: string | null | undefined,
  options: ComposeSessionSwitcherOptions = {},
): boolean {
  if (!event || event.isComposing) return false;
  if (options.searchMode) return false;
  if (!options.showSessionSwitcherButton) return false;
  if (event.ctrlKey || event.metaKey || event.altKey) return false;
  if (event.key !== '@') return false;
  return String(value || '') === '';
}
