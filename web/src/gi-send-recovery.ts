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
