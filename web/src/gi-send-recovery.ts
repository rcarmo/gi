import { pendingSendKey, type PendingSend } from './gi-drafts.js';

// A lost HTTP reply is not a rejected prompt. Confirm only one same-session
// token-tagged turn past the rollback boundary; never resubmit or match text.
export async function recoverSubmittedPrompt(
    sessionId: string, token: string, read: (path: string) => Promise<any>,
): Promise<any | null> {
    try {
        const snapshot = await read(`/api/sessions/${encodeURIComponent(sessionId)}/turns`);
        if (!Array.isArray(snapshot?.turns)) return null;
        const matches = snapshot.turns.filter((turn: any) =>
            turn?.session_id === sessionId && turn?.metadata?.client_request_id === token);
        if (matches.length !== 1 || typeof matches[0].id !== 'string' || !matches[0].id) return null;
        const turn = matches[0];
        const audit = await read(`/api/turns/${encodeURIComponent(turn.id)}/events`);
        if (!Array.isArray(audit?.events) || !audit.events.some((event: any) =>
            event?.type === 'turn.submitted' && event?.turn_id === turn.id && event?.session_id === sessionId)) return null;
        return { turn_id: turn.id, session_id: sessionId, status: turn.status, queued: turn.status === 'queued' };
    } catch { return null; } // failed/absent/ambiguous evidence remains unknown
}

// Bound startup network work independently of accumulated local pending rows.
// One consistent turn snapshot per session, at most six session reads and six
// receipts, two workers, plus a caller-owned deadline. Unchecked rows stay unknown.
export async function recoverPendingSends(pending: PendingSend[], read: (path: string) => Promise<any>): Promise<Set<string>> {
    const confirmed = new Set<string>();
    const selected = pending.slice(0, 6);
    const snapshots = new Map<string, Promise<any>>();
    let cursor = 0;
    const worker = async () => {
        while (cursor < selected.length) {
            const item = selected[cursor++];
            if (!item.sessionId || !item.token) continue;
            const path = `/api/sessions/${encodeURIComponent(item.sessionId)}/turns`;
            if (!snapshots.has(path)) snapshots.set(path, read(path).catch(() => null));
            const result = await recoverSubmittedPrompt(item.sessionId, item.token,
                target => target === path ? snapshots.get(path)! : read(target));
            if (result) confirmed.add(pendingSendKey(item.sessionId, item.token));
        }
    };
    await Promise.all([worker(), worker()]);
    return confirmed;
}
