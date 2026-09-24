// Terminal notifications may arrive after a newer run has started. They can
// request an authoritative refresh, but cannot clear another occurrence's UI.
// Legacy untagged terminal frames are refresh-only while a run is known.
export function staleTerminalEvent(type: string, data: any, currentTurn: string | null): boolean {
    const terminal = type === 'agent_response'
        || (type === 'agent_status' && !['running', 'cancelling'].includes(data?.status));
    return Boolean(terminal && currentTurn && data?.turn_id !== currentTurn);
}
