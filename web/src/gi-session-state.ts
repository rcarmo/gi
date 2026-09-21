/** Adapt native ancestry before applying Piclaw's picker grouping. */
export function sessionPickerAgents(sessions: any[]) {
    const byId = new Map(sessions.map(session => [session.id, session]));
    return sessions.map(session => {
        let root = session;
        const visited = new Set([root.id]);
        while (root.parent_session_id && byId.has(root.parent_session_id) && !visited.has(root.parent_session_id)) {
            root = byId.get(root.parent_session_id);
            visited.add(root.id);
        }
        return {
            chat_jid: `gi:${session.id}`,
            agent_name: (typeof session.title === 'string' && session.title ? session.title.replace(/^@/, '') : session.scope?.agent_id) || session.id,
            agent_id: session.scope?.agent_id || 'agent',
            branch_id: session.id,
            parent_branch_id: session.parent_session_id || null,
            parent_chat_jid: session.parent_session_id ? `gi:${session.parent_session_id}` : null,
            root_chat_jid: `gi:${root.id}`,
            model: session.state?.model || '',
            is_active: session.state?.status === 'running' || session.state?.status === 'queued' || Number(session.state?.queue_count || 0) > 0,
            archived_at: session.state?.archived_at || null,
            pinned: session.state?.pinned === true,
            capabilities: {
                rename: !session.state?.archived_at,
                pin: !session.state?.archived_at,
                archive: Boolean(session.parent_session_id) && !session.state?.archived_at,
                restore: Boolean(session.state?.archived_at),
            },
        };
    });
}

// Gi's async selection boundary. Captured requests remain owned by the selection
// that started them, including an A -> B -> A switch before a response arrives.
export function createSelectionScope() {
    let sessionId: string | null = null;
    let generation = 0;
    return {
        select(next: string) {
            if (next !== sessionId) { sessionId = next; generation++; }
        },
        capture() { return { sessionId, generation }; },
        isCurrent(captured: { sessionId: string | null; generation: number }) {
            return captured.sessionId === sessionId && captured.generation === generation;
        },
        current() { return sessionId; },
    };
}
