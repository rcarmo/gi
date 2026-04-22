import {
  getGeneratedWidgetSessionKey,
  normalizeLiveGeneratedWidgetPayload,
} from './generated-widget.js';

export interface FloatingWidgetLike {
  source?: string | null;
  artifact?: Record<string, unknown> | null;
  originChatJid?: string | null;
  openedAt?: string | null;
  liveUpdatedAt?: string | null;
  runtimeState?: Record<string, unknown> | null;
  [key: string]: unknown;
}

export interface ApplyLiveFloatingWidgetUpdateOptions {
  fallbackStatus?: string;
  currentChatJid: string;
  dismissedSessionKeys?: Set<string> | null;
  updatedAt: string;
}

export interface CloseFloatingWidgetStateResult {
  nextWidget: null;
  dismissedSessionKey: string | null;
}

export interface FloatingWidgetSubmitPendingOptions {
  kind: string;
  payload: unknown;
  submittedAt: string;
  submissionText: string | null;
}

export interface FloatingWidgetSubmitResultOptions {
  submittedAt: string;
  submissionText: string;
  queued?: string | null;
  errorMessage?: string | null;
}

export interface FloatingWidgetHostEventOptions {
  kind: string;
  payload: unknown;
  eventAt: string;
  nextRefreshCount: number;
  shouldBuildDashboard: boolean;
}

function matchesFloatingWidgetSession(current: FloatingWidgetLike | null | undefined, sessionKey: string): boolean {
  const currentKey = getGeneratedWidgetSessionKey(current);
  return Boolean(current && currentKey === sessionKey);
}

function mergeFloatingWidgetRuntimeState(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  patch: Record<string, unknown>,
): FloatingWidgetLike | null | undefined {
  if (!matchesFloatingWidgetSession(current, sessionKey)) return current;
  return {
    ...current,
    runtimeState: {
      ...((current?.runtimeState && typeof current.runtimeState === 'object') ? current.runtimeState : {}),
      ...patch,
    },
  };
}

export function openFloatingWidgetState<T extends Record<string, any>>(widget: T, openedAt: string): T & { openedAt: string } {
  return {
    ...widget,
    openedAt,
  };
}

export function closeFloatingWidgetState(current: FloatingWidgetLike | null | undefined): CloseFloatingWidgetStateResult {
  const sessionKey = getGeneratedWidgetSessionKey(current);
  return {
    nextWidget: null,
    dismissedSessionKey: current?.source === 'live' && sessionKey ? sessionKey : null,
  };
}

export function applyLiveFloatingWidgetUpdate(
  current: FloatingWidgetLike | null | undefined,
  data: Record<string, any> | null | undefined,
  options: ApplyLiveFloatingWidgetUpdateOptions,
): FloatingWidgetLike | null | undefined {
  const payload = normalizeLiveGeneratedWidgetPayload({
    ...data,
    ...((data && data.status) ? {} : { status: options.fallbackStatus || 'streaming' }),
  });
  if (!payload) return current;

  const sessionKey = getGeneratedWidgetSessionKey(payload);
  if (sessionKey && options.dismissedSessionKeys?.has(sessionKey)) {
    return current;
  }

  const currentKey = getGeneratedWidgetSessionKey(current);
  const sameSession = Boolean(sessionKey && currentKey && sessionKey === currentKey);
  const mergedArtifact = {
    ...((sameSession && current?.artifact) ? current.artifact : {}),
    ...(payload.artifact || {}),
  };

  return {
    ...(sameSession && current ? current : {}),
    ...payload,
    artifact: mergedArtifact,
    source: 'live',
    originChatJid: payload.originChatJid || options.currentChatJid,
    openedAt: sameSession && current?.openedAt ? current.openedAt : options.updatedAt,
    liveUpdatedAt: options.updatedAt,
  };
}

export function clearLiveFloatingWidgetState(
  current: FloatingWidgetLike | null | undefined,
  data: Record<string, any> | null | undefined,
): FloatingWidgetLike | null {
  if (!current || current?.source !== 'live') return current || null;
  const sessionKey = getGeneratedWidgetSessionKey(data);
  const currentKey = getGeneratedWidgetSessionKey(current);
  if (sessionKey && currentKey && sessionKey !== currentKey) return current;
  return null;
}

export function applyFloatingWidgetSubmitPending(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  options: FloatingWidgetSubmitPendingOptions,
): FloatingWidgetLike | null | undefined {
  return mergeFloatingWidgetRuntimeState(current, sessionKey, {
    lastEventKind: options.kind,
    lastEventPayload: options.payload || null,
    lastSubmitAt: options.submittedAt,
    lastHostUpdate: {
      type: 'submit_pending',
      submittedAt: options.submittedAt,
      preview: options.submissionText || null,
    },
  });
}

export function applyFloatingWidgetSubmitResult(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  options: FloatingWidgetSubmitResultOptions,
): FloatingWidgetLike | null | undefined {
  if (options.errorMessage) {
    return mergeFloatingWidgetRuntimeState(current, sessionKey, {
      lastHostUpdate: {
        type: 'submit_failed',
        submittedAt: options.submittedAt,
        preview: options.submissionText,
        error: options.errorMessage,
      },
    });
  }

  return mergeFloatingWidgetRuntimeState(current, sessionKey, {
    lastHostUpdate: {
      type: options.queued === 'followup' ? 'submit_queued' : 'submit_sent',
      submittedAt: options.submittedAt,
      preview: options.submissionText,
      queued: options.queued || null,
    },
  });
}

export function applyFloatingWidgetHostEvent(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  options: FloatingWidgetHostEventOptions,
): FloatingWidgetLike | null | undefined {
  return mergeFloatingWidgetRuntimeState(current, sessionKey, {
    lastEventKind: options.kind,
    lastEventPayload: options.payload || null,
    ...(options.kind === 'widget.ready'
      ? {
          readyAt: options.eventAt,
          lastHostUpdate: {
            type: 'ready_ack',
            at: options.eventAt,
          },
        }
      : {}),
    ...(options.kind === 'widget.request_refresh'
      ? {
          lastRefreshRequestAt: options.eventAt,
          refreshCount: options.nextRefreshCount,
          lastHostUpdate: {
            type: options.shouldBuildDashboard ? 'refresh_building' : 'refresh_ack',
            at: options.eventAt,
            count: options.nextRefreshCount,
            echo: options.payload || null,
          },
        }
      : {}),
  });
}

export function applyFloatingWidgetDashboardResult(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  options: { dashboard: unknown; at: string; count: number; echo: unknown },
): FloatingWidgetLike | null | undefined {
  return mergeFloatingWidgetRuntimeState(current, sessionKey, {
    dashboard: options.dashboard,
    lastHostUpdate: {
      type: 'refresh_dashboard',
      at: options.at,
      count: options.count,
      echo: options.echo || null,
    },
  });
}

export function applyFloatingWidgetDashboardFailure(
  current: FloatingWidgetLike | null | undefined,
  sessionKey: string,
  options: { errorMessage: string; at: string; count: number },
): FloatingWidgetLike | null | undefined {
  return mergeFloatingWidgetRuntimeState(current, sessionKey, {
    lastHostUpdate: {
      type: 'refresh_failed',
      at: options.at,
      count: options.count,
      error: options.errorMessage,
    },
  });
}
