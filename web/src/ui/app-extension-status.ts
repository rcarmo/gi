export interface StatusPanelLike {
  state?: string | null;
  [key: string]: unknown;
}

export interface StatusPanelPayloadLike {
  key?: unknown;
  content?: Array<{ type?: unknown; panel?: StatusPanelLike | null } | null> | null;
  options?: {
    remove?: boolean;
    surface?: string | null;
  } | null;
}

export interface ExtensionPanelActionResult {
  refreshAutoresearchStatus: boolean;
  toast?: {
    title: string;
    detail: string | null;
    kind: string;
    durationMs?: number;
  };
}

export interface RunExtensionStatusPanelActionOptions {
  panel: {
    tmux_command?: string | null;
    [key: string]: unknown;
  } | null | undefined;
  action: {
    key?: string | null;
    action_type?: string | null;
    [key: string]: unknown;
  } | null | undefined;
  currentChatJid: string;
  stopAutoresearch(chatJid: string, options: { generateReport: boolean }): Promise<unknown>;
  dismissAutoresearch(chatJid: string): Promise<unknown>;
  writeClipboard(text: string): Promise<unknown>;
}

export function findStatusPanel(payload: StatusPanelPayloadLike | null | undefined): StatusPanelLike | null {
  if (!Array.isArray(payload?.content)) return null;
  const match = payload.content.find((item) => item?.type === 'status_panel' && item?.panel);
  return match?.panel || null;
}

export function applyAutoresearchStatusPayload<T>(
  previous: Map<string, T>,
  payload: StatusPanelPayloadLike | null | undefined,
): Map<string, T> {
  const next = new Map(previous);
  const panel = findStatusPanel(payload) as T | null;
  if (typeof payload?.key === 'string' && payload.key && panel) {
    next.set(payload.key, panel);
  } else {
    next.delete('autoresearch');
  }
  return next;
}

export function applyStatusPanelWidgetEvent<T>(
  previous: Map<string, T>,
  payload: StatusPanelPayloadLike | null | undefined,
): Map<string, T> {
  const panelKey = typeof payload?.key === 'string' ? payload.key : '';
  if (!panelKey) return previous;

  const next = new Map(previous);
  const panel = findStatusPanel(payload) as T | null;
  if (payload?.options?.remove || !panel) {
    next.delete(panelKey);
  } else {
    next.set(panelKey, panel);
  }
  return next;
}

export function shouldClearPendingPanelActions(payload: StatusPanelPayloadLike | null | undefined): boolean {
  if (payload?.options?.remove) return true;
  const panel = findStatusPanel(payload);
  return panel?.state !== 'running';
}

export function createPendingPanelActionKey(panelKey: string, actionKey: string): string {
  return `${panelKey}:${actionKey}`;
}

export function addPendingPanelAction(previous: Set<string>, panelKey: string, actionKey: string): Set<string> {
  const pendingKey = createPendingPanelActionKey(panelKey, actionKey);
  if (previous.has(pendingKey)) return previous;
  const next = new Set(previous);
  next.add(pendingKey);
  return next;
}

export function removePendingPanelAction(previous: Set<string>, pendingKey: string): Set<string> {
  if (!previous.has(pendingKey)) return previous;
  const next = new Set(previous);
  next.delete(pendingKey);
  return next;
}

export function clearPendingPanelActionPrefix(previous: Set<string>, panelKey: string): Set<string> {
  if (previous.size === 0) return previous;
  const prefix = `${panelKey}:`;
  const next = new Set(Array.from(previous).filter((key) => !String(key).startsWith(prefix)));
  return next.size === previous.size ? previous : next;
}

export async function runExtensionStatusPanelAction(
  options: RunExtensionStatusPanelActionOptions,
): Promise<ExtensionPanelActionResult> {
  const actionType = typeof options.action?.action_type === 'string' ? options.action.action_type : '';
  const actionKey = typeof options.action?.key === 'string' ? options.action.key : '';

  if (actionType === 'autoresearch.stop') {
    await options.stopAutoresearch(options.currentChatJid, { generateReport: true });
    return { refreshAutoresearchStatus: true };
  }

  if (actionType === 'autoresearch.dismiss') {
    await options.dismissAutoresearch(options.currentChatJid);
    return { refreshAutoresearchStatus: true };
  }

  if (actionType === 'autoresearch.copy_tmux') {
    const tmuxCommand = typeof options.panel?.tmux_command === 'string' ? options.panel.tmux_command.trim() : '';
    if (!tmuxCommand) throw new Error('No tmux command available.');
    await options.writeClipboard(tmuxCommand);
    return {
      refreshAutoresearchStatus: false,
      toast: {
        title: 'Copied',
        detail: 'tmux command copied to clipboard.',
        kind: 'success',
      },
    };
  }

  throw new Error(`Unsupported panel action: ${actionType || actionKey}`);
}
