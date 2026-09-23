// Document-local invalidation only: no model snapshots, credentials or drafts.
// Listeners are scoped to mounted consumers; inactive sessions retain no entries.
const listeners = new Map<string, Set<() => void>>();

export function subscribeModelSettlement(chatJid: string, listener: () => void): () => void {
    let scoped = listeners.get(chatJid);
    if (!scoped) listeners.set(chatJid, scoped = new Set());
    scoped.add(listener);
    return () => {
        scoped.delete(listener);
        if (!scoped.size && listeners.get(chatJid) === scoped) listeners.delete(chatJid);
    };
}

export function notifyModelSettlement(chatJid: string): void {
    // Snapshot permits listeners to detach while notifying without skipping peers.
    for (const listener of [...(listeners.get(chatJid) || [])]) {
        try { listener(); } catch { /* A consumer cannot change a mutation outcome. */ }
    }
}
