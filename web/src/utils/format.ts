// @ts-nocheck

/** Format a timestamp as a human-readable relative or absolute time string. */
export function formatTime(timestamp) {
    const date = new Date(timestamp);
    if (Number.isNaN(date.getTime())) return timestamp;
    const now = new Date();
    const diffMs = now - date;
    const diffSec = diffMs / 1000;
    const dayMs = 24 * 60 * 60 * 1000;

    if (diffMs < dayMs) {
        if (diffSec < 60) return 'just now';
        if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`;
        return `${Math.floor(diffSec / 3600)}h`;
    }

    if (diffMs < 5 * dayMs) {
        const weekday = date.toLocaleDateString(undefined, { weekday: 'short' });
        const time = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
        return `${weekday} ${time}`;
    }

    const datePart = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
    const timePart = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
    return `${datePart} ${timePart}`;
}

/** Format a number with locale-aware thousand separators. */
export function formatCount(value) {
    if (!Number.isFinite(value)) return '0';
    return Math.round(value).toLocaleString();
}

/** Format a byte count as a human-readable file size (B/KB/MB). */
export function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

/** Format a timestamp as a full locale date-time string. */
export function formatTimestamp(value) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleString();
}
