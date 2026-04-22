export interface ExtensionUiToastLike {
  title: string;
  detail: string | null;
  kind: string;
  durationMs?: number;
}

export interface ExtensionUiWorkingIndicatorState {
  mode: 'default' | 'custom' | 'hidden';
  frames: string[];
  intervalMs: number | null;
}

export interface ExtensionUiWorkingState {
  message: string | null;
  indicator: ExtensionUiWorkingIndicatorState | null;
}

export interface StatusPanelWidgetEventContext {
  isStatusPanelWidgetEvent: boolean;
  eventChatJid: string;
  panelKey: string;
}

export function resolveStatusPanelEventChatJid(
  payload: Record<string, unknown> | null | undefined,
  currentChatJid: string,
): string {
  return typeof payload?.chat_jid === 'string' && payload.chat_jid.trim()
    ? payload.chat_jid.trim()
    : currentChatJid;
}

export function resolveStatusPanelWidgetEventContext(
  eventType: string | null | undefined,
  payload: Record<string, unknown> | null | undefined,
  currentChatJid: string,
): StatusPanelWidgetEventContext {
  return {
    isStatusPanelWidgetEvent: eventType === 'extension_ui_widget' && payload?.options?.surface === 'status-panel',
    eventChatJid: resolveStatusPanelEventChatJid(payload, currentChatJid),
    panelKey: typeof payload?.key === 'string' ? payload.key : '',
  };
}

export function resolveExtensionUiWorkingIndicator(
  eventType: string | null | undefined,
  payload: Record<string, unknown> | null | undefined,
): ExtensionUiWorkingIndicatorState | null | undefined {
  if (eventType !== 'extension_ui_working_indicator') return undefined;

  if (!Array.isArray(payload?.frames)) {
    return {
      mode: 'default',
      frames: [],
      intervalMs: null,
    };
  }

  const frames = payload.frames.filter((frame): frame is string => typeof frame === 'string');
  const intervalRaw = payload.interval_ms ?? payload.intervalMs;
  const intervalMs = typeof intervalRaw === 'number' && Number.isFinite(intervalRaw) && intervalRaw > 0
    ? intervalRaw
    : null;

  if (frames.length === 0) {
    return {
      mode: 'hidden',
      frames: [],
      intervalMs,
    };
  }

  return {
    mode: 'custom',
    frames,
    intervalMs,
  };
}

export function applyExtensionUiWorkingState(
  previous: ExtensionUiWorkingState,
  eventType: string | null | undefined,
  payload: Record<string, unknown> | null | undefined,
): ExtensionUiWorkingState | undefined {
  if (eventType === 'extension_ui_working') {
    return {
      message: typeof payload?.message === 'string' && payload.message.trim() ? payload.message.trim() : null,
      indicator: previous.indicator,
    };
  }

  const indicator = resolveExtensionUiWorkingIndicator(eventType, payload);
  if (indicator === undefined) return undefined;
  return {
    message: previous.message,
    indicator,
  };
}

export function resolveExtensionUiToast(
  eventType: string | null | undefined,
  payload: Record<string, unknown> | null | undefined,
): ExtensionUiToastLike | null {
  if (eventType === 'extension_ui_notify' && typeof payload?.message === 'string') {
    return {
      title: payload.message,
      detail: null,
      kind: typeof payload?.type === 'string' && payload.type.trim() ? payload.type : 'info',
    };
  }

  if (eventType === 'extension_ui_error') {
    const errorValue = payload?.error;
    const errorText = typeof errorValue === 'string'
      ? errorValue
      : (errorValue && typeof errorValue === 'object' && typeof (errorValue as Record<string, unknown>).error === 'string')
        ? (errorValue as Record<string, unknown>).error as string
        : (errorValue ? String(errorValue) : 'Unknown extension error');
    return {
      title: 'Extension UI error',
      detail: errorText,
      kind: 'error',
      durationMs: 5000,
    };
  }

  return null;
}
