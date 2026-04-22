// @ts-nocheck
import { useEffect, useMemo, useState } from '../vendor/preact-htm.js';

export const RECONNECTING_HINT_DELAY_MS = 350;

function formatConnectionStatusLabel(status) {
    return String(status || 'Connecting')
        .replace(/[-_]+/g, ' ')
        .replace(/^./, (match) => match.toUpperCase());
}

export function resolveConnectionStatusPresentation(status, options = {}) {
    const normalizedStatus = typeof status === 'string' && status.trim()
        ? status.trim()
        : 'connecting';

    if (normalizedStatus === 'connected') {
        return {
            show: false,
            statusClass: 'connected',
            label: 'Connected',
            title: 'Connection: Connected',
        };
    }

    if (normalizedStatus !== 'disconnected') {
        const label = formatConnectionStatusLabel(normalizedStatus);
        return {
            show: true,
            statusClass: normalizedStatus,
            label,
            title: `Connection: ${label}`,
        };
    }

    const delayMs = Number.isFinite(Number(options?.delayMs))
        ? Math.max(0, Number(options.delayMs))
        : RECONNECTING_HINT_DELAY_MS;
    const nowMs = Number.isFinite(Number(options?.nowMs))
        ? Number(options.nowMs)
        : Date.now();
    const disconnectedAtMs = Number.isFinite(Number(options?.disconnectedAtMs))
        ? Number(options.disconnectedAtMs)
        : nowMs;
    const showReconnecting = (nowMs - disconnectedAtMs) >= delayMs;

    return showReconnecting
        ? {
            show: true,
            statusClass: 'disconnected',
            label: 'Reconnecting',
            title: 'Reconnecting',
        }
        : {
            show: true,
            statusClass: 'connecting',
            label: 'Connecting',
            title: 'Connecting',
        };
}

export function useConnectionStatusPresentation(status, options = {}) {
    const delayMs = Number.isFinite(Number(options?.delayMs))
        ? Math.max(0, Number(options.delayMs))
        : RECONNECTING_HINT_DELAY_MS;
    const [disconnectedAtMs, setDisconnectedAtMs] = useState(null);
    const [displayNowMs, setDisplayNowMs] = useState(() => Date.now());

    useEffect(() => {
        if (status === 'disconnected') {
            const startedAt = Date.now();
            setDisconnectedAtMs((previous) => previous ?? startedAt);
            setDisplayNowMs(startedAt);
            return;
        }
        setDisconnectedAtMs(null);
        setDisplayNowMs(Date.now());
    }, [status]);

    useEffect(() => {
        if (status !== 'disconnected' || disconnectedAtMs === null) return undefined;
        const remainingMs = delayMs - (Date.now() - disconnectedAtMs);
        if (remainingMs <= 0) return undefined;
        const timeoutId = setTimeout(() => {
            setDisplayNowMs(Date.now());
        }, remainingMs);
        return () => clearTimeout(timeoutId);
    }, [status, disconnectedAtMs, delayMs]);

    return useMemo(() => resolveConnectionStatusPresentation(status, {
        delayMs,
        disconnectedAtMs,
        nowMs: displayNowMs,
    }), [status, delayMs, disconnectedAtMs, displayNowMs]);
}
