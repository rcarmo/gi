import { randomClientId } from "./gi-random-id.js";
// Browser-local session drafts. File bytes use IndexedDB ArrayBuffers; never
// serialise attachment bytes to localStorage or silently drop failed writes.
export type Draft = { text: string; media: File[]; fileRefs: string[]; messageRefs: any[] };
type Pending = { id: string; draft: Draft };
type QueueReturn = { state: 'prepared' | 'removed'; recoveredAt: number };
type Record = { sessionId: string; revision?: number; draft: Draft; pending: Pending[]; error?: string; queueReturns?: { [id: string]: QueueReturn } };
export class DraftConflictError extends Error {
    constructor() { super('Draft changed in another tab. Copy your unsaved text and attachments before reloading; sending is blocked in this tab.'); this.name = 'DraftConflictError'; }
}
function revision(row?: Record): number {
    const value = row?.revision ?? 0;
    if (!Number.isSafeInteger(value) || value < 0) throw new Error('Invalid draft revision');
    return value;
}
export const emptyDraft = (): Draft => ({ text: '', media: [], fileRefs: [], messageRefs: [] });
const copy = (d: Draft): Draft => ({ text: d.text, media: [...d.media], fileRefs: [...d.fileRefs], messageRefs: [...d.messageRefs] });
const key = (value: any) => typeof value === 'object' ? JSON.stringify(value) : String(value);
const unique = <T>(values: T[], identity: (value: T) => string = key) => [...new Map(values.map(value => [identity(value), value])).values()];
export function mergeDrafts(captured: Draft, current: Draft): Draft {
    const text = !captured.text || current.text === captured.text || current.text.startsWith(captured.text + '\n')
        ? current.text : [captured.text, current.text].filter(Boolean).join('\n\n');
    return {
        text,
        media: unique([...captured.media, ...current.media], f => `${f.name}:${f.size}:${f.type}:${f.lastModified}`),
        fileRefs: unique([...captured.fileRefs, ...current.fileRefs]),
        messageRefs: unique([...captured.messageRefs, ...current.messageRefs]),
    };
}

export interface DraftStorage { load(): Promise<Record[]>; put(record: Record, expectedRevision: number): Promise<number> }
export function indexedDraftStorage(factory: IDBFactory = indexedDB): DraftStorage {
    const encodedFiles = new WeakMap<File, Promise<any>>();
    const encodeFile = (file: File) => {
        if (!encodedFiles.has(file)) encodedFiles.set(file, file.arrayBuffer().then(bytes => ({ name: file.name, type: file.type, lastModified: file.lastModified, bytes })));
        return encodedFiles.get(file)!;
    };
    const database = new Promise<IDBDatabase>((resolve, reject) => {
        // Fence out older tabs whose version-1 adapter writes unconditional snapshots.
        const request = factory.open('gi-session-drafts', 2);
        request.onupgradeneeded = () => { if (!request.result.objectStoreNames.contains('drafts')) request.result.createObjectStore('drafts', { keyPath: 'sessionId' }); };
        let abandoned = false;
        request.onsuccess = () => {
            // A blocked upgrade may finish after load has already failed. Do not
            // leak an inaccessible connection; recovery requires a full reload.
            if (abandoned) { request.result.close(); return; }
            request.result.onversionchange = () => request.result.close(); resolve(request.result);
        };
        request.onerror = () => { abandoned = true; reject(request.error || new Error('Draft database unavailable')); };
        request.onblocked = () => { abandoned = true; reject(new Error('Draft database upgrade blocked by another tab. Close older Gi tabs, copy any unsaved edits, then reload.')); };
    });
    return {
        async load() {
            const db = await database;
            return new Promise((resolve, reject) => {
                const tx = db.transaction('drafts', 'readonly');
                const request = tx.objectStore('drafts').getAll();
                tx.oncomplete = () => {
                    const decode = (draft: any) => ({ ...draft, media: draft.media.map((file: any) => file instanceof File ? file : new File([file.bytes], file.name, { type: file.type, lastModified: file.lastModified })) });
                    resolve(request.result.map(row => ({ ...row, draft: decode(row.draft), pending: row.pending.map(p => ({ ...p, draft: decode(p.draft) })) })));
                };
                tx.onabort = tx.onerror = () => reject(tx.error || new Error('Could not load drafts'));
            });
        },
        async put(record, expectedRevision) {
            const db = await database;
            // WebKit can reject File/Blob values during IDB commit even though
            // structuredClone accepts them. Store explicit bytes and metadata.
            const encode = async (draft: Draft) => ({ ...draft, media: await Promise.all(draft.media.map(encodeFile)) });
            const stored = { ...record, draft: await encode(record.draft), pending: await Promise.all(record.pending.map(async pending => ({ ...pending, draft: await encode(pending.draft) }))) };
            return new Promise((resolve, reject) => {
                const tx = db.transaction('drafts', 'readwrite');
                const store = tx.objectStore('drafts');
                let failure: Error | undefined;
                const current = store.get(record.sessionId);
                current.onsuccess = () => {
                    try {
                        if (revision(current.result) !== expectedRevision) throw new DraftConflictError();
                        if (expectedRevision === Number.MAX_SAFE_INTEGER) throw new Error('Draft revision exhausted; copy unsaved edits before recovery.');
                        stored.revision = expectedRevision + 1;
                        store.put(stored);
                    } catch (error) { failure = error; tx.abort(); }
                };
                tx.oncomplete = () => resolve(stored.revision!);
                tx.onabort = tx.onerror = () => reject(failure || tx.error || new Error('Could not save draft'));
            });
        },
    };
}

export type PendingSend = { sessionId: string; token: string };
export type PendingSendRecovery = (pending: PendingSend[]) => Promise<Set<string>>;
export const pendingSendKey = (sessionId: string, token: string) => JSON.stringify([sessionId, token]);

export function createDraftRepository(storage: DraftStorage, onError: (error: Error) => void = () => {}, recover?: PendingSendRecovery) {
    const records = new Map<string, Record>();
    const revisions = new Map<string, number>();
    const conflicts = new Map<string, Error>();
    let loadFailure: Error | undefined;
    let tail: Promise<void> = Promise.resolve();
    const record = (id: string) => {
        if (!records.has(id)) records.set(id, { sessionId: id, draft: emptyDraft(), pending: [] });
        return records.get(id)!;
    };
    const persist = (id: string) => {
        const source = record(id);
        const snapshot = { ...source, queueReturns: Object.fromEntries(Object.entries(source.queueReturns || {}).map(([id, entry]) => [id, { ...entry }])), draft: copy(source.draft), pending: source.pending.map(p => ({ id: p.id, draft: copy(p.draft) })) };
        const write = tail.catch(() => {}).then(async () => {
            if (loadFailure) throw loadFailure;
            if (conflicts.has(id)) throw conflicts.get(id);
            // Resolve our last committed revision at execution time, so queued
            // writes from this tab do not conflict with their own predecessors.
            revisions.set(id, await storage.put(snapshot, revisions.get(id) ?? 0));
        });
        tail = write;
        void write.catch(error => {
            if (error instanceof DraftConflictError) { conflicts.set(id, error); record(id).error = error.message; }
            onError(error);
        });
        return write;
    };
    return {
        async load() {
          // A repository is owned by one page load. Do not rebase local edits or
          // a speculative recovery over foreign commits on bootstrap retry.
          if (loadFailure) throw loadFailure;
          if (conflicts.size) throw conflicts.values().next().value;
          try {
            const rows = await storage.load();
            // No mutation/send on recovery. Old captures may not carry their
            // token into native admission; failure or no proof remains unknown.
            let confirmed = new Set<string>();
            if (recover) {
                try { confirmed = await recover(rows.flatMap(row => row.pending.map(p => ({ sessionId: row.sessionId, token: p.id })))); }
                catch { /* Preserve the existing unknown-delivery recovery. */ }
            }
            for (const row of rows) {
                // Keep speculative recovery out of the public in-memory map.
                // Bootstrap catches load failures; it must never expose a stale
                // accepted capture as resendable text after a conflicting write.
                const expected = revision(row);
                if (row.pending.length) {
                    const unknown = row.pending.filter(p => !confirmed.has(pendingSendKey(row.sessionId, p.id)));
                    for (const pending of [...unknown].reverse()) row.draft = mergeDrafts(pending.draft, row.draft);
                    row.pending = [];
                    if (unknown.length) row.error = 'Recovered an unacknowledged send. Delivery is unknown; check the timeline before resending.';
                    row.revision = await storage.put(row, expected);
                }
            }
            for (const row of rows) { records.set(row.sessionId, row); revisions.set(row.sessionId, revision(row)); }
          } catch (error) { loadFailure = error; records.clear(); revisions.clear(); throw error; }
        },
        get(id: string) { return record(id).draft; },
        error(id: string) { return record(id).error || ''; },
        update(id: string, patch: Partial<Draft>) {
            const draft = record(id).draft;
            if (Object.entries(patch).every(([field, value]) => draft[field as keyof Draft] === value)) return;
            Object.assign(draft, patch);
            void persist(id).catch(() => {});
        },
        begin(id: string, draft: Draft) {
            const token = randomClientId();
            const row = record(id);
            row.pending.push({ id: token, draft: copy(draft) });
            row.draft = emptyDraft();
            row.error = '';
            return { token, ready: persist(id) };
        },
        async accepted(id: string, token: string) {
            const row = record(id);
            row.pending = row.pending.filter(p => p.id !== token);
            await persist(id);
        },
        failed(id: string, token: string, error: string) {
            const row = record(id);
            const pending = row.pending.find(p => p.id === token);
            if (pending) row.draft = mergeDrafts(pending.draft, row.draft);
            row.pending = row.pending.filter(p => p.id !== token);
            row.error = error;
            void persist(id).catch(() => {});
            return copy(row.draft);
        },
        hasQueueReturn(id: string, queueId: string) { return Boolean(record(id).queueReturns?.[queueId]); },
        prepareQueueReturn(id: string, queueId: string, captured: Draft) {
            const row = record(id);
            row.queueReturns ||= {};
            if (!row.queueReturns[queueId]) {
                const current = row.draft;
                row.draft = mergeDrafts(captured, current);
                // Distinct durable IDs may intentionally contain identical text.
                // Only the recovery key, not a text prefix, makes return idempotent.
                row.draft.text = [captured.text, current.text].filter(Boolean).join('\n\n');
                row.queueReturns[queueId] = { state: 'prepared', recoveredAt: Date.now() };
            }
            // Retry persists the existing merge; it must not prepend it again.
            return { draft: copy(row.draft), ready: persist(id) };
        },
        queueReturnFailed(id: string, queueId: string, message: string) {
            const row = record(id);
            if (row.queueReturns?.[queueId]) {
                row.error = `Queue return incomplete: ${message}. Recovered content is retained; check whether the original turn ran before sending it again.`;
                void persist(id).catch(() => {});
            }
        },
        async completeQueueReturn(id: string, queueId: string) {
            const entry = record(id).queueReturns?.[queueId];
            if (entry) entry.state = 'removed';
            if (record(id).error?.startsWith('Queue return incomplete:')) record(id).error = '';
            await persist(id);
        },
        async flushStable() {
            let pending: Promise<void>;
            do { pending = tail; await pending; } while (pending !== tail);
        },
        flush() { return tail; },
    };
}
