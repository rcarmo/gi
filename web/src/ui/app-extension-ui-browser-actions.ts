import { respondToAgentRequest } from '../api.js';

export const MIN_EDITOR_POPOUT_VIEWPORT_WIDTH = 1280;
export const MIN_EDITOR_POPOUT_VIEWPORT_HEIGHT = 820;

function readPositiveNumber(value: unknown): number | null {
  if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) return null;
  return value;
}

function readTrimmedString(value: unknown): string | null {
  return typeof value === 'string' && value.trim() ? value.trim() : null;
}

export function isWorkspaceEditorPath(value: unknown): value is string {
  const path = readTrimmedString(value);
  if (!path) return false;
  if (path.startsWith('/') || path.startsWith('\\')) return false;
  if (path.includes('://')) return false;
  if (path === '.' || path === '..' || path.startsWith('../')) return false;
  return true;
}

export function hasEnoughViewportForEditorPopout(runtime: {
  windowWidth?: number | null;
  windowHeight?: number | null;
  screenWidth?: number | null;
  screenHeight?: number | null;
  isMobile?: boolean;
}): boolean {
  if (runtime.isMobile) return false;
  const windowWidth = readPositiveNumber(runtime.windowWidth);
  const windowHeight = readPositiveNumber(runtime.windowHeight);
  if (windowWidth === null || windowHeight === null) return false;
  if (windowWidth < MIN_EDITOR_POPOUT_VIEWPORT_WIDTH || windowHeight < MIN_EDITOR_POPOUT_VIEWPORT_HEIGHT) {
    return false;
  }
  const screenWidth = readPositiveNumber(runtime.screenWidth);
  const screenHeight = readPositiveNumber(runtime.screenHeight);
  if (screenWidth !== null && screenWidth < MIN_EDITOR_POPOUT_VIEWPORT_WIDTH) return false;
  if (screenHeight !== null && screenHeight < MIN_EDITOR_POPOUT_VIEWPORT_HEIGHT) return false;
  return true;
}

export function isLikelyMobileWindow(win: any): boolean {
  const navigator = win?.navigator;
  const userAgent = String(navigator?.userAgent || '');
  const maxTouchPoints = Number(navigator?.maxTouchPoints || 0);
  const coarsePointer = typeof win?.matchMedia === 'function'
    ? Boolean(win.matchMedia('(pointer: coarse)')?.matches)
    : false;
  return coarsePointer || maxTouchPoints > 1 || /Android|iPhone|iPad|iPod/i.test(userAgent);
}

export function normalizeOpenWorkspaceFileRequest(payload: any): null | {
  requestId: string;
  chatJid: string | null;
  path: string;
  label: string | null;
  target: 'tab' | 'popout';
} {
  if (!payload || payload.kind !== 'custom') return null;
  const requestId = readTrimmedString(payload.request_id);
  const options = payload.options && typeof payload.options === 'object' ? payload.options : null;
  if (!requestId || !options || options.action !== 'open_workspace_file') return null;
  const path = readTrimmedString(options.path);
  if (!isWorkspaceEditorPath(path)) return null;
  const target = options.target === 'tab' ? 'tab' : 'popout';
  return {
    requestId,
    chatJid: readTrimmedString(payload.chat_jid),
    path,
    label: readTrimmedString(options.label),
    target,
  };
}

export async function respondToExtensionUiRequest(
  requestId: string,
  outcome: Record<string, unknown>,
  chatJid: string | null,
): Promise<void> {
  await respondToAgentRequest(requestId, outcome, chatJid || undefined);
}

export async function handleOpenWorkspaceFileBrowserRequest(event: CustomEvent, deps: {
  currentChatJid: string;
  openEditor: (path: string) => void;
  popOutPane: (path: string, label?: string | null) => Promise<boolean> | boolean;
  showIntentToast?: (title: string, detail?: string | null, kind?: string, durationMs?: number) => void;
  windowObject?: any;
  respond?: (requestId: string, outcome: Record<string, unknown>, chatJid: string | null) => Promise<void>;
}): Promise<boolean> {
  const request = normalizeOpenWorkspaceFileRequest(event?.detail?.payload);
  if (!request) return false;
  if (request.chatJid && request.chatJid !== deps.currentChatJid) return false;

  const respond = deps.respond || respondToExtensionUiRequest;
  const win = deps.windowObject || (typeof window !== 'undefined' ? window : undefined);
  const viewport = {
    width: Number(win?.innerWidth || 0) || undefined,
    height: Number(win?.innerHeight || 0) || undefined,
  };
  const minimumViewport = {
    width: MIN_EDITOR_POPOUT_VIEWPORT_WIDTH,
    height: MIN_EDITOR_POPOUT_VIEWPORT_HEIGHT,
  };

  if (request.target === 'popout') {
    const allowed = hasEnoughViewportForEditorPopout({
      windowWidth: win?.innerWidth,
      windowHeight: win?.innerHeight,
      screenWidth: win?.screen?.availWidth,
      screenHeight: win?.screen?.availHeight,
      isMobile: isLikelyMobileWindow(win),
    });
    if (!allowed) {
      deps.showIntentToast?.(
        'Editor popout unavailable',
        `Need at least ${MIN_EDITOR_POPOUT_VIEWPORT_WIDTH}×${MIN_EDITOR_POPOUT_VIEWPORT_HEIGHT} viewport space for a separate editor window.`,
        'warning',
        4500,
      );
      await respond(request.requestId, {
        ok: false,
        opened: false,
        reason: 'insufficient_screen_space',
        detail: 'Browser viewport is too small for a separate editor window.',
        target: request.target,
        path: request.path,
        viewport,
        minimum_viewport: minimumViewport,
      }, request.chatJid);
      return true;
    }
    const opened = await deps.popOutPane(request.path, request.label);
    await respond(request.requestId, {
      ok: opened,
      opened,
      reason: opened ? undefined : 'popout_failed',
      detail: opened ? undefined : 'The browser blocked the editor popout window.',
      target: request.target,
      path: request.path,
    }, request.chatJid);
    return true;
  }

  deps.openEditor(request.path);
  await respond(request.requestId, {
    ok: true,
    opened: true,
    target: request.target,
    path: request.path,
  }, request.chatJid);
  return true;
}
