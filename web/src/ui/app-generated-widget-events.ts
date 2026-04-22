export type LiveGeneratedWidgetFallbackStatus = 'loading' | 'streaming' | 'final' | 'error';

export interface LiveGeneratedWidgetEventResolution {
  kind: 'update' | 'close' | null;
  fallbackStatus: LiveGeneratedWidgetFallbackStatus | null;
  shouldAdoptTurn: boolean;
}

export function resolveLiveGeneratedWidgetEvent(eventType: string | null | undefined): LiveGeneratedWidgetEventResolution {
  switch (eventType) {
    case 'generated_widget_open':
      return {
        kind: 'update',
        fallbackStatus: 'loading',
        shouldAdoptTurn: true,
      };
    case 'generated_widget_delta':
      return {
        kind: 'update',
        fallbackStatus: 'streaming',
        shouldAdoptTurn: true,
      };
    case 'generated_widget_final':
      return {
        kind: 'update',
        fallbackStatus: 'final',
        shouldAdoptTurn: true,
      };
    case 'generated_widget_error':
      return {
        kind: 'update',
        fallbackStatus: 'error',
        shouldAdoptTurn: false,
      };
    case 'generated_widget_close':
      return {
        kind: 'close',
        fallbackStatus: null,
        shouldAdoptTurn: false,
      };
    default:
      return {
        kind: null,
        fallbackStatus: null,
        shouldAdoptTurn: false,
      };
  }
}
