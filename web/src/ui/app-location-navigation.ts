import { useCallback, useEffect, useMemo, useState } from '../vendor/preact-htm.js';

export function resolveNavigationUrl(nextUrl: unknown, baseHref: string): string {
  return new URL(String(nextUrl || ''), baseHref).toString();
}

export function useAppLocationNavigation() {
  const [locationHref, setLocationHref] = useState(() =>
    typeof window === 'undefined' ? 'http://localhost/' : window.location.href,
  );

  useEffect(() => {
    if (typeof window === 'undefined') return undefined;
    const handlePopState = () => setLocationHref(window.location.href);
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  const navigate = useCallback((nextUrl: unknown, options: { replace?: boolean } = {}) => {
    if (typeof window === 'undefined') return;
    const { replace = false } = options || {};
    const resolved = resolveNavigationUrl(nextUrl, window.location.href);
    if (replace) {
      window.history.replaceState(null, '', resolved);
    } else {
      window.history.pushState(null, '', resolved);
    }
    setLocationHref(window.location.href);
  }, []);

  const locationParams = useMemo(() => new URL(locationHref).searchParams, [locationHref]);

  return {
    locationParams,
    navigate,
  };
}
