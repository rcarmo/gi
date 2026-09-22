// Browser-local session drafts. File bytes use IndexedDB ArrayBuffers; never
// serialise attachment bytes to localStorage or silently drop failed writes.
export type Draft = { text: string; media: File[]; fileRefs: string[]; messageRefs: any[] };
type Pending = { id: string; draft: Draft };
type QueueReturn = { state: 'prepared' | 'removed'; recoveredAt: number };
type Record = { sessionId: string; draft: Draft; pending: Pending[]; error?: string; queueReturns?: { [id: string]: QueueReturn } };
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

export interface DraftStorage { load(): Promise<Record[]>; put(record: Record): Promise<void> }
export function indexedDraftStorage(factory: IDBFactory = indexedDB): DraftStorage {
    const encodedFiles = new WeakMap<File, Promise<any>>();
    const encodeFile = (file: File) => {
        if (!encodedFiles.has(file)) encodedFiles.set(file, file.arrayBuffer().then(bytes => ({ name: file.name, type: file.type, lastModified: file.lastModified, bytes })));
        return encodedFiles.get(file)!;
    };
    const database = new Promise<IDBDatabase>((resolve, reject) => {
        const request = factory.open('gi-session-drafts', 1);
        request.onupgradeneeded = () => { if (!request.result.objectStoreNames.contains('drafts')) request.result.createObjectStore('drafts', { keyPath: 'sessionId' }); };
        request.onsuccess = () => { request.result.onversionchange = () => request.result.close(); resolve(request.result); };
        request.onerror = () => reject(request.error || new Error('Draft database unavailable'));
        request.onblocked = () => reject(new Error('Draft database upgrade blocked by another tab'));
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
        async put(record) {
            const db = await database;
            // WebKit can reject File/Blob values during IDB commit even though
            // structuredClone accepts them. Store explicit bytes and metadata.
            const encode = async (draft: Draft) => ({ ...draft, media: await Promise.all(draft.media.map(encodeFile)) });
            const stored = { ...record, draft: await encode(record.draft), pending: await Promise.all(record.pending.map(async pending => ({ ...pending, draft: await encode(pending.draft) }))) };
            return new Promise((resolve, reject) => {
                const tx = db.transaction('drafts', 'readwrite');
                tx.objectStore('drafts').put(stored);
                tx.oncomplete = () => resolve();
                tx.onabort = tx.onerror = () => reject(tx.error || new Error('Could not save draft'));
            });
        },
    };
}

export function createDraftRepository(storage: DraftStorage, onError: (error: Error) => void = () => {}) {
    const records = new Map<string, Record>();
    let tail: Promise<void> = Promise.resolve();
    const record = (id: string) => {
        if (!records.has(id)) records.set(id, { sessionId: id, draft: emptyDraft(), pending: [] });
        return records.get(id)!;
    };
    const persist = (id: string) => {
        const source = record(id);
        const snapshot = { ...source, queueReturns: Object.fromEntries(Object.entries(source.queueReturns || {}).map(([id, entry]) => [id, { ...entry }])), draft: copy(source.draft), pending: source.pending.map(p => ({ id: p.id, draft: copy(p.draft) })) };
        const write = tail.catch(() => {}).then(() => storage.put(snapshot));
        tail = write;
        void write.catch(error => onError(error));
        return write;
    };
    return {
        async load() {
            const rows = await storage.load();
            for (const row of rows) records.set(row.sessionId, row);
            for (const row of rows) {
                // The storage adapter restores File objects from persisted bytes.
                records.set(row.sessionId, row);
                if (row.pending.length) {
                    for (const pending of [...row.pending].reverse()) row.draft = mergeDrafts(pending.draft, row.draft);
                    row.pending = [];
                    row.error = 'Recovered an unacknowledged send. Delivery is unknown; check the timeline before resending.';
                    await persist(row.sessionId);
                }
            }
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
            const token = crypto.randomUUID();
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
