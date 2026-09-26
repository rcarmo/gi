import { pendingSendKey, type PendingSend } from './gi-drafts.js';

// Native source-scoped receipts cover returned local/routed/steered admission.
// Never scan another chat, infer from text, or retry POST. Legacy/no receipt is
// unknown; duplicate tokens are rejected by the native bounded lookup.
export async function recoverSubmittedPrompt(
    sessionId: string, token: string, read: (path: string) => Promise<any>,
): Promise<any | null> {
    try {
        const receipt = await read(`/api/sessions/${encodeURIComponent(sessionId)}/send-receipt?client_request_id=${encodeURIComponent(token)}`);
        if (receipt?.confirmed !== true || receipt.source_session_id !== sessionId || receipt.client_request_id !== token) return null;
        const result = receipt.result;
        if (typeof result?.turn_id !== 'string' || !result.turn_id || typeof result.session_id !== 'string' || !result.session_id) return null;
        if (result.session_id !== sessionId && (result.source_session_id !== sessionId || result.routed !== true)) return null;
        return result;
    } catch { return null; }
}

// At most six indexed receipt reads, two workers and caller-owned deadline.
// Unchecked captures remain unknown. Native responses are size-bounded too.
export async function recoverPendingSends(pending: PendingSend[], read: (path: string) => Promise<any>): Promise<Set<string>> {
    const confirmed = new Set<string>();
    const selected = pending.slice(0, 6);
    let cursor = 0;
    const worker = async () => {
        while (cursor < selected.length) {
            const item = selected[cursor++];
            if (!item.sessionId || !item.token) continue;
            const result = await recoverSubmittedPrompt(item.sessionId, item.token, read);
            if (result) confirmed.add(pendingSendKey(item.sessionId, item.token));
        }
    };
    await Promise.all([worker(), worker()]);
    return confirmed;
}
