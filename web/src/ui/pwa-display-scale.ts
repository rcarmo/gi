import { isMobileBrowserMode, isStandaloneWebAppMode } from './chat-window.js';

export const PWA_DISPLAY_SCALE_STORAGE_KEY = 'piclawPwaDisplayScalePercent';
export const PWA_DISPLAY_SCALE_EVENT = 'piclaw:pwa-display-scale-changed';
export const DEFAULT_PWA_DISPLAY_SCALE_PERCENT = 100;
export const MIN_PWA_DISPLAY_SCALE_PERCENT = 20;
export const MAX_PWA_DISPLAY_SCALE_PERCENT = 115;
export const PWA_DISPLAY_SCALE_STEP_PERCENT = 5;

const DEFAULT_VIEWPORT_CONTENT = 'width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover';

function getRuntimeWindow(runtime: any = typeof window !== 'undefined' ? window : null) {
  return runtime && typeof runtime === 'object' ? runtime : null;
}

function getRuntimeNavigator(runtime: any) {
  return runtime?.navigator || (typeof navigator !== 'undefined' ? navigator : null);
}

function parseScalePercent(value: unknown): number | null {
  if (typeof value === 'number') return Number.isFinite(value) ? value : null;
  if (typeof value !== 'string') return null;
  const trimmed = value.trim();
  if (!trimmed) return null;
  const numeric = trimmed.endsWith('%') ? trimmed.slice(0, -1) : trimmed;
  const parsed = Number(numeric);
  if (!Number.isFinite(parsed)) return null;
  // Accept decimal viewport-style values for migration/manual edits.
  return parsed > 0 && parsed <= 2 ? parsed * 100 : parsed;
}

export function normalizePwaDisplayScalePercent(
  value: unknown,
  fallback = DEFAULT_PWA_DISPLAY_SCALE_PERCENT,
): number {
  const parsed = parseScalePercent(value);
  const base = parsed ?? parseScalePercent(fallback) ?? DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  return Math.min(
    MAX_PWA_DISPLAY_SCALE_PERCENT,
    Math.max(MIN_PWA_DISPLAY_SCALE_PERCENT, Math.round(base)),
  );
}

export function formatPwaDisplayScaleRatio(percent: number): string {
  const ratio = normalizePwaDisplayScalePercent(percent) / 100;
  return ratio.toFixed(2).replace(/0+$/, '').replace(/\.$/, '');
}

export function readStoredPwaDisplayScalePercent(runtime: any = typeof window !== 'undefined' ? window : null): number {
  const runtimeWindow = getRuntimeWindow(runtime);
  try {
    return normalizePwaDisplayScalePercent(runtimeWindow?.localStorage?.getItem(PWA_DISPLAY_SCALE_STORAGE_KEY));
  } catch {
    return DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  }
}

export function isMobileStandalonePwa(runtime: any = typeof window !== 'undefined' ? window : null): boolean {
  const runtimeWindow = getRuntimeWindow(runtime);
  const runtimeNavigator = getRuntimeNavigator(runtimeWindow);
  return isStandaloneWebAppMode({ window: runtimeWindow, navigator: runtimeNavigator })
    && isMobileBrowserMode({ window: runtimeWindow, navigator: runtimeNavigator });
}

export function buildPwaDisplayScaleViewportContent(percent: number, options: { applies?: boolean } = {}): string {
  const normalized = normalizePwaDisplayScalePercent(percent);
  if (!options.applies || normalized === DEFAULT_PWA_DISPLAY_SCALE_PERCENT) {
    return DEFAULT_VIEWPORT_CONTENT;
  }

  const scale = formatPwaDisplayScaleRatio(normalized);
  return `width=device-width, initial-scale=${scale}, minimum-scale=${scale}, maximum-scale=${scale}, user-scalable=no, viewport-fit=cover`;
}

export function applyPwaDisplayScale(runtime: any = typeof window !== 'undefined' ? window : null): {
  percent: number;
  applied: boolean;
  content: string;
} {
  const runtimeWindow = getRuntimeWindow(runtime);
  const percent = readStoredPwaDisplayScalePercent(runtimeWindow);
  const applied = isMobileStandalonePwa(runtimeWindow) && percent !== DEFAULT_PWA_DISPLAY_SCALE_PERCENT;
  const content = buildPwaDisplayScaleViewportContent(percent, { applies: applied });
  const viewport = runtimeWindow?.document?.querySelector?.('meta[name="viewport"]');
  viewport?.setAttribute?.('content', content);
  return { percent, applied, content };
}

export function persistPwaDisplayScalePercent(
  percent: unknown,
  runtime: any = typeof window !== 'undefined' ? window : null,
): number {
  const runtimeWindow = getRuntimeWindow(runtime);
  const normalized = normalizePwaDisplayScalePercent(percent);
  try {
    runtimeWindow?.localStorage?.setItem(PWA_DISPLAY_SCALE_STORAGE_KEY, String(normalized));
  } catch (error) {
    console.warn('[pwa-display-scale] Unable to persist scale preference; applying in-memory value only.', error);
  }
  applyPwaDisplayScale(runtimeWindow);
  if (runtimeWindow?.dispatchEvent) {
    const event = typeof CustomEvent === 'function'
      ? new CustomEvent(PWA_DISPLAY_SCALE_EVENT, { detail: { percent: normalized } })
      : { type: PWA_DISPLAY_SCALE_EVENT, detail: { percent: normalized } };
    runtimeWindow.dispatchEvent(event);
  }
  return normalized;
}

export function installPwaDisplayScaleSync(runtime: any = typeof window !== 'undefined' ? window : null): (() => void) | undefined {
  const runtimeWindow = getRuntimeWindow(runtime);
  if (!runtimeWindow?.addEventListener) return undefined;

  const sync = () => applyPwaDisplayScale(runtimeWindow);
  sync();

  const onStorage = (event: any) => {
    if (!event || event.key === null || event.key === PWA_DISPLAY_SCALE_STORAGE_KEY) sync();
  };
  const onScaleChanged = () => sync();

  runtimeWindow.addEventListener('storage', onStorage);
  runtimeWindow.addEventListener('focus', sync);
  runtimeWindow.addEventListener(PWA_DISPLAY_SCALE_EVENT, onScaleChanged);

  const mediaQueries = ['standalone', 'fullscreen', 'minimal-ui']
    .map((mode) => {
      try {
        return runtimeWindow.matchMedia?.(`(display-mode: ${mode})`);
      } catch {
        return null;
      }
    })
    .filter(Boolean);

  for (const media of mediaQueries) {
    if (media.addEventListener) media.addEventListener('change', sync);
    else if (media.addListener) media.addListener(sync);
  }

  return () => {
    runtimeWindow.removeEventListener('storage', onStorage);
    runtimeWindow.removeEventListener('focus', sync);
    runtimeWindow.removeEventListener(PWA_DISPLAY_SCALE_EVENT, onScaleChanged);
    for (const media of mediaQueries) {
      if (media.removeEventListener) media.removeEventListener('change', sync);
      else if (media.removeListener) media.removeListener(sync);
    }
  };
}
