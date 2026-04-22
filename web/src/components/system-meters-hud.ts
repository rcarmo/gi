// @ts-nocheck
import { html, useEffect, useMemo, useState } from '../vendor/preact-htm.js';
import { getSystemMetrics } from '../api.js';
import { METERS_COLLAPSED_EVENT_NAME, METERS_EVENT_NAME, readStoredMetersCollapsed, readStoredMetersEnabled, toggleMetersCollapsed } from '../ui/meters.js';

function sanitizeSeries(input, maxPoints = 30) {
    const series = Array.isArray(input)
        ? input
            .map((value) => Number(value))
            .filter((value) => Number.isFinite(value))
        : [];
    return series.length > maxPoints ? series.slice(series.length - maxPoints) : series;
}

function clampPercentSeries(input, maxPoints = 30) {
    return sanitizeSeries(input, maxPoints).map((value) => Math.max(0, Math.min(100, value)));
}

export function buildSparklinePath(series, width = 56, height = 16, options = {}) {
    const points = sanitizeSeries(series);
    if (points.length === 0) return '';

    const minValue = Number.isFinite(options.min) ? Number(options.min) : Math.min(...points);
    const maxValue = Number.isFinite(options.max) ? Number(options.max) : Math.max(...points);

    if (!(maxValue > minValue)) {
        const y = (height / 2).toFixed(2);
        return `M 0 ${y} L ${width} ${y}`;
    }

    if (points.length === 1) {
        const normalized = (points[0] - minValue) / (maxValue - minValue);
        const y = (height - normalized * height).toFixed(2);
        return `M 0 ${y} L ${width} ${y}`;
    }

    return points.map((value, index) => {
        const x = (index / (points.length - 1 || 1)) * width;
        const normalized = (value - minValue) / (maxValue - minValue);
        const y = height - normalized * height;
        return `${index === 0 ? 'M' : 'L'} ${x.toFixed(2)} ${y.toFixed(2)}`;
    }).join(' ');
}

function formatPercent(value) {
    return `${Math.round(Number(value) || 0)}%`;
}

export function formatBytesCompact(value) {
    const bytes = Number(value);
    if (!Number.isFinite(bytes) || bytes <= 0) return '0B';
    const units = ['B', 'K', 'M', 'G', 'T'];
    let unitIndex = 0;
    let scaled = bytes;
    while (scaled >= 1024 && unitIndex < units.length - 1) {
        scaled /= 1024;
        unitIndex += 1;
    }
    const digits = scaled >= 100 || unitIndex === 0 ? 0 : scaled >= 10 ? 0 : 1;
    return `${scaled.toFixed(digits)}${units[unitIndex]}`;
}

export function buildCompactMetersSummary(metrics) {
    const parts = [
        `CPU ${formatPercent(metrics?.cpu_percent)}`,
        `RAM ${formatPercent(metrics?.ram_percent)}`,
    ];
    if (Number.isFinite(Number(metrics?.swap_percent)) && Number(metrics?.swap_total_bytes) > 0) {
        parts.push(`SWP ${formatPercent(metrics?.swap_percent)}`);
    }
    return parts.join(' • ');
}

export function resolveCurrentRssBytes(metrics) {
    return Number(metrics?.process_memory?.vm_rss_bytes) > 0
        ? Number(metrics.process_memory.vm_rss_bytes)
        : Number(metrics?.process_memory?.rss_bytes) || 0;
}

export function shouldShowRss(metrics) {
    return resolveCurrentRssBytes(metrics) > 0 && sanitizeSeries(metrics?.process_rss_series_bytes).length > 0;
}

function readIsNarrowLayout() {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return false;
    return window.matchMedia('(max-width: 900px)').matches;
}

export function SystemMetersHud({ mode = 'overlay' }) {
    const [enabled, setEnabled] = useState(() => readStoredMetersEnabled(false));
    const [collapsed, setCollapsed] = useState(() => readStoredMetersCollapsed(false));
    const [isNarrowLayout, setIsNarrowLayout] = useState(() => readIsNarrowLayout());
    const [metrics, setMetrics] = useState({
        cpu_percent: 0,
        ram_percent: 0,
        swap_percent: null,
        cpu_series: [],
        ram_series: [],
        swap_series: [],
        process_rss_series_bytes: [],
        process_memory: {
            rss_bytes: 0,
            vm_rss_bytes: null,
        },
        swap_total_bytes: 0,
        swap_used_bytes: 0,
        sample_interval_ms: 2000,
        platform: '',
    });
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        const onMetersChange = (event) => {
            setEnabled(Boolean(event?.detail?.enabled));
        };
        const onMetersCollapsedChange = (event) => {
            setCollapsed(Boolean(event?.detail?.collapsed));
        };
        window.addEventListener(METERS_EVENT_NAME, onMetersChange);
        window.addEventListener(METERS_COLLAPSED_EVENT_NAME, onMetersCollapsedChange);
        return () => {
            window.removeEventListener(METERS_EVENT_NAME, onMetersChange);
            window.removeEventListener(METERS_COLLAPSED_EVENT_NAME, onMetersCollapsedChange);
        };
    }, []);

    useEffect(() => {
        if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return undefined;
        const mediaQuery = window.matchMedia('(max-width: 900px)');
        const sync = () => setIsNarrowLayout(Boolean(mediaQuery.matches));
        sync();
        if (typeof mediaQuery.addEventListener === 'function') {
            mediaQuery.addEventListener('change', sync);
            return () => mediaQuery.removeEventListener('change', sync);
        }
        mediaQuery.addListener(sync);
        return () => mediaQuery.removeListener(sync);
    }, []);

    const activeMode = 'overlay';
    const isActiveInstance = mode === activeMode;

    useEffect(() => {
        if (!enabled || !isActiveInstance) return undefined;
        let cancelled = false;
        let timer = 0;

        const refresh = async () => {
            setLoading((prev) => (prev || metrics.cpu_series.length > 0 ? prev : true));
            try {
                const next = await getSystemMetrics();
                if (cancelled) return;
                setMetrics({
                    cpu_percent: Number(next?.cpu_percent) || 0,
                    ram_percent: Number(next?.ram_percent) || 0,
                    swap_percent: Number.isFinite(Number(next?.swap_percent)) ? Number(next?.swap_percent) : null,
                    cpu_series: clampPercentSeries(next?.cpu_series),
                    ram_series: clampPercentSeries(next?.ram_series),
                    swap_series: clampPercentSeries(next?.swap_series),
                    process_rss_series_bytes: sanitizeSeries(next?.process_rss_series_bytes),
                    process_memory: {
                        rss_bytes: Number(next?.process_memory?.rss_bytes) || 0,
                        vm_rss_bytes: Number.isFinite(Number(next?.process_memory?.vm_rss_bytes)) ? Number(next?.process_memory?.vm_rss_bytes) : null,
                    },
                    swap_total_bytes: Number(next?.swap_total_bytes) || 0,
                    swap_used_bytes: Number(next?.swap_used_bytes) || 0,
                    sample_interval_ms: Number(next?.sample_interval_ms) || 2000,
                    platform: String(next?.platform || ''),
                });
            } catch {
                if (cancelled) return;
            } finally {
                if (!cancelled) setLoading(false);
            }
        };

        void refresh();
        timer = window.setInterval(() => {
            if (document?.visibilityState === 'hidden') return;
            void refresh();
        }, Math.max(1000, Number(metrics.sample_interval_ms) || 2000));

        return () => {
            cancelled = true;
            if (timer) window.clearInterval(timer);
        };
    }, [enabled, isActiveInstance]);

    const cpuPath = useMemo(() => buildSparklinePath(metrics.cpu_series, 56, 16, { min: 0, max: 100 }), [metrics.cpu_series]);
    const ramPath = useMemo(() => buildSparklinePath(metrics.ram_series, 56, 16, { min: 0, max: 100 }), [metrics.ram_series]);
    const swapPath = useMemo(() => buildSparklinePath(metrics.swap_series, 56, 16, { min: 0, max: 100 }), [metrics.swap_series]);
    const rssPath = useMemo(() => buildSparklinePath(metrics.process_rss_series_bytes), [metrics.process_rss_series_bytes]);
    const showSwap = Number.isFinite(Number(metrics.swap_percent)) && metrics.swap_total_bytes > 0;
    const currentRssBytes = resolveCurrentRssBytes(metrics);
    const showRss = shouldShowRss(metrics);
    const compactSummary = useMemo(() => buildCompactMetersSummary(metrics), [metrics]);

    if (!enabled || !isActiveInstance) return null;

    const title = collapsed
        ? 'Show system meters'
        : (loading ? 'Updating system meters… Click to collapse.' : 'System meters — click to collapse.');

    const handleToggleCollapsed = (event) => {
        event?.stopPropagation?.();
        toggleMetersCollapsed();
    };

    return html`
        <div class=${`system-meters-hud system-meters-hud-${mode}${collapsed ? ' is-collapsed' : ''}`} aria-live="polite">
            <button
                class="system-meters-card"
                type="button"
                title=${title}
                aria-label=${title}
                aria-expanded=${collapsed ? 'false' : 'true'}
                onClick=${handleToggleCollapsed}
            >
                ${collapsed
                    ? html`<span class="system-meters-collapse-tab" aria-hidden="true">◂</span>`
                    : isNarrowLayout
                        ? html`<span class="system-meters-compact-summary">${compactSummary}</span>`
                        : html`
                            <div class="system-meters-row cpu">
                                <span class="system-meters-label">CPU</span>
                                <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                    <path d=${cpuPath}></path>
                                </svg>
                                <span class="system-meters-value">${formatPercent(metrics.cpu_percent)}</span>
                            </div>
                            <div class="system-meters-row ram">
                                <span class="system-meters-label">RAM</span>
                                <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                    <path d=${ramPath}></path>
                                </svg>
                                <span class="system-meters-value">${formatPercent(metrics.ram_percent)}</span>
                            </div>
                            ${showRss && html`
                                <div class="system-meters-row rss">
                                    <span class="system-meters-label">RSS</span>
                                    <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                        <path d=${rssPath}></path>
                                    </svg>
                                    <span class="system-meters-value">${formatBytesCompact(currentRssBytes)}</span>
                                </div>
                            `}
                            ${showSwap && html`
                                <div class="system-meters-row swap">
                                    <span class="system-meters-label">SWP</span>
                                    <svg class="system-meters-spark" viewBox="0 0 56 16" preserveAspectRatio="none" aria-hidden="true">
                                        <path d=${swapPath}></path>
                                    </svg>
                                    <span class="system-meters-value">${formatPercent(metrics.swap_percent)}</span>
                                </div>
                            `}
                        `}
            </button>
        </div>
    `;
}
