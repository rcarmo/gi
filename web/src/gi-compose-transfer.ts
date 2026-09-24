// Transient transport state: never persisted or restored as work after reload.
export type TransferSnapshot = { uploads: number; sending: number; loaded: number; total: number; computable: boolean };
type Operation = { phase: 'upload' | 'send'; loaded: number; total: number; computable: boolean };
export function createComposeTransfers() {
    const uploadBatches = new Map<string, Set<AbortController>>();
    const sessions = new Map<string, Map<symbol, Operation>>();
    const listeners = new Set<() => void>();
    const emit = () => { for (const listener of listeners) listener(); };
    return {
        beginUploadBatch(session: string) {
            const controller = new AbortController();
            let batches = uploadBatches.get(session);
            if (!batches) { batches = new Set(); uploadBatches.set(session, batches); }
            batches.add(controller);
            let ended = false;
            return { signal: controller.signal, end() {
                if (ended) return; ended = true;
                batches!.delete(controller); if (!batches!.size) uploadBatches.delete(session);
            } };
        },
        cancelUploads(session: string) {
            // Capture this occurrence set: an abort listener may start new work.
            for (const batch of [...(uploadBatches.get(session) || [])]) batch.abort();
        },
        subscribe(listener: () => void) { listeners.add(listener); return () => { listeners.delete(listener); }; },
        snapshot(session: string): TransferSnapshot {
            const value = { uploads: 0, sending: 0, loaded: 0, total: 0, computable: true };
            for (const op of sessions.get(session)?.values() || []) {
                if (op.phase === 'send') { value.sending++; continue; }
                value.uploads++; value.loaded += op.loaded; value.total += op.total;
                value.computable &&= op.computable;
            }
            return value;
        },
        begin(session: string, phase: Operation['phase']) {
            const token = Symbol(phase), op: Operation = { phase, loaded: 0, total: 0, computable: false };
            let pending = sessions.get(session);
            if (!pending) { pending = new Map(); sessions.set(session, pending); }
            pending.set(token, op); emit();
            let ended = false;
            return {
                progress(loaded: number, total: number, computable: boolean) {
                    if (ended || phase !== 'upload') return;
                    op.computable = computable && Number.isFinite(total) && total > 0;
                    op.total = op.computable ? total : 0;
                    op.loaded = Number.isFinite(loaded) ? Math.max(0, op.computable ? Math.min(loaded, total) : loaded) : 0;
                    emit();
                },
                end() {
                    if (ended) return; ended = true;
                    pending!.delete(token); if (!pending!.size) sessions.delete(session);
                    emit();
                },
            };
        },
    };
}
export const composeTransfers = createComposeTransfers();

// Decorate only host-owned attributes. The supplied composer retains its click,
// disabled and Abort/Compact semantics, including newer-draft submission.
export function bindComposeSending(root: HTMLElement, sending: boolean) {
    const owned = new Set<HTMLElement>();
    const paint = () => {
        const button = root.querySelector<HTMLElement>('.compose-send-stack .send-btn');
        if (!button) return;
        if (sending) { button.dataset.giSending = 'true'; button.setAttribute('aria-busy', 'true'); owned.add(button); }
        else { delete button.dataset.giSending; button.removeAttribute('aria-busy'); }
    };
    paint();
    const observer = new MutationObserver(paint);
    observer.observe(root, { childList: true, subtree: true });
    return () => {
        observer.disconnect();
        for (const button of owned) { delete button.dataset.giSending; button.removeAttribute('aria-busy'); }
    };
}
