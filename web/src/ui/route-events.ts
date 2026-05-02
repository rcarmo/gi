import { getLocalStorageBoolean, setLocalStorageItem } from '../utils/storage.js';

export const ROUTING_EVENTS_STORAGE_KEY = 'piclaw_route_events_enabled';
export const ROUTING_EVENTS_EVENT_NAME = 'piclaw-route-events-change';

function dispatchRoutingEventsChange(enabled: boolean) {
    if (typeof window === 'undefined') return;
    window.dispatchEvent(new CustomEvent(ROUTING_EVENTS_EVENT_NAME, {
        detail: { enabled: Boolean(enabled) },
    }));
}

export function readStoredRoutingEventsEnabled(defaultValue = true) {
    return getLocalStorageBoolean(ROUTING_EVENTS_STORAGE_KEY, defaultValue);
}

export function applyRoutingEventsEnabled(enabled: boolean, options = {}) {
    const persist = options.persist !== false;
    const next = Boolean(enabled);
    if (persist) {
        setLocalStorageItem(ROUTING_EVENTS_STORAGE_KEY, next ? 'true' : 'false');
    }
    dispatchRoutingEventsChange(next);
    return next;
}

export function toggleRoutingEventsEnabled() {
    const next = !readStoredRoutingEventsEnabled(true);
    return applyRoutingEventsEnabled(next);
}
