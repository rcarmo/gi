// Per-page successful deletion IDs, never optimistic removals. Late pages and
// search responses cannot resurrect a row after native acknowledgement.
export function createMessageDeletionState() {
    const removed = new Set<string>();
    const pending = new Set<string>();
    return {
        begin(id: string) { if (pending.has(id) || removed.has(id)) return false; pending.add(id); return true; },
        finish(id: string, success: boolean) { pending.delete(id); if (success) removed.add(id); },
        filter<T extends { id: string }>(rows: T[], animating: Set<string> = new Set()): T[] {
            return rows.filter(row => !removed.has(row.id) || animating.has(row.id));
        },
    };
}
