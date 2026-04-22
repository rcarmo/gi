export interface FloatingWidgetToast {
  title: string;
  detail: string;
  kind: 'info' | 'warning';
  durationMs: number;
}

export interface FloatingWidgetHostRefreshContext {
  shouldBuildDashboard: boolean;
  nextRefreshCount: number;
}

export function resolveFloatingWidgetSubmitToast(queued: unknown): FloatingWidgetToast {
  if (queued === 'followup') {
    return {
      title: 'Widget submission queued',
      detail: 'The widget message was queued because the agent is busy.',
      kind: 'info',
      durationMs: 3500,
    };
  }

  return {
    title: 'Widget submission sent',
    detail: 'The widget message was sent to the chat.',
    kind: 'info',
    durationMs: 3500,
  };
}

export function resolveFloatingWidgetSubmitFailureToast(errorMessage: string | null | undefined): FloatingWidgetToast {
  return {
    title: 'Widget submission failed',
    detail: errorMessage || 'Could not send the widget message.',
    kind: 'warning',
    durationMs: 5000,
  };
}

export function resolveFloatingWidgetHostRefreshContext(
  payload: Record<string, unknown> | null | undefined,
  currentRefreshCount: unknown,
): FloatingWidgetHostRefreshContext {
  return {
    shouldBuildDashboard: Boolean(payload?.buildDashboard || payload?.dashboardKind === 'internal-state'),
    nextRefreshCount: Number(currentRefreshCount || 0) + 1,
  };
}

export function resolveFloatingWidgetDashboardBuiltToast(): FloatingWidgetToast {
  return {
    title: 'Dashboard built',
    detail: 'Live dashboard state pushed into the widget.',
    kind: 'info',
    durationMs: 3000,
  };
}

export function resolveFloatingWidgetDashboardFailureToast(errorMessage: string | null | undefined): FloatingWidgetToast {
  return {
    title: 'Dashboard build failed',
    detail: errorMessage || 'Could not build dashboard.',
    kind: 'warning',
    durationMs: 5000,
  };
}

export function resolveFloatingWidgetRefreshAckToast(): FloatingWidgetToast {
  return {
    title: 'Widget refresh requested',
    detail: 'The widget received a host acknowledgement update.',
    kind: 'info',
    durationMs: 3000,
  };
}
