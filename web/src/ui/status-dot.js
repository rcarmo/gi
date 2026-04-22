// @ts-nocheck

export function buildTurnDotClass({ steerQueued = false, pulsing = false } = {}) {
    const classes = ['turn-dot'];
    if (steerQueued) classes.push('turn-dot-queued');
    if (pulsing) classes.push('turn-dot-pulsing');
    return classes.join(' ');
}

export function buildComposeStatusDotClass({ pulsing = false } = {}) {
    const classes = ['compose-inline-status-dot'];
    if (pulsing) classes.push('compose-inline-status-dot-pulsing');
    return classes.join(' ');
}

export function resolveRunningStatusIndicator(status, { isLastActivity = false, pendingRequest = false } = {}) {
    if (pendingRequest) return 'dot';
    if (isLastActivity) return 'none';
    if (status?.type === 'error') return 'none';
    if (status?.type === 'intent') return 'dot';

    const type = typeof status?.type === 'string' ? status.type : '';
    const hasToolMetadata = Boolean(
        (typeof status?.tool_name === 'string' && status.tool_name.trim())
        || status?.tool_args,
    );
    if (hasToolMetadata) return 'spinner';
    if (type === 'tool_call' || type === 'tool_status' || type === 'thinking' || type === 'waiting') {
        return 'spinner';
    }
    return 'dot';
}

export function shouldShowRunningStatusDot(status, options = {}) {
    return resolveRunningStatusIndicator(status, options) === 'dot';
}
