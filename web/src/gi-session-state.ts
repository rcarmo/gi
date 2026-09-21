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
