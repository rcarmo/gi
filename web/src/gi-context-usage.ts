const known = (value: unknown): value is number => typeof value === 'number' && Number.isFinite(value) && value >= 0;
export function formatContextCount(value: unknown): string {
    if (!known(value)) return '?';
    if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`;
    if (value >= 1_000) return `${(value / 1_000).toFixed(0)}K`;
    return String(value);
}

export function contextPresentation(usage: any, canCompact = false) {
    const percent = known(usage?.percent) ? usage.percent : null;
    const fill = percent == null ? 0 : Math.min(100, percent);
    const label = `Context: ${formatContextCount(usage?.tokens)} / ${formatContextCount(usage?.contextWindow)} tokens (${percent == null ? '?' : percent.toFixed(0)}%)`;
    const qualifier = usage?.source === 'provider_request' ? ' — latest measured provider request' : ' — usage unavailable';
    return { fill, label, title: label + qualifier + (canCompact ? ' — Compact context' : ''),
        color: percent == null ? 'var(--text-secondary)' : percent > 90 ? 'var(--context-red, #ef4444)' : percent > 75 ? 'var(--context-amber, #f59e0b)' : 'var(--context-green, #22c55e)' };
}

export function modelContextBlocked(option: any, usage: any): boolean {
    return known(usage?.tokens) && known(option?.contextWindow) && option.contextWindow > 0 && usage.tokens > option.contextWindow;
}
