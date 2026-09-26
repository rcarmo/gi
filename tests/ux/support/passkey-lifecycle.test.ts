import { afterEach, expect, mock, test } from 'bun:test';
import { runPasskey } from '../../../web/src/gi-passkeys';

const originals = new Map(['window', 'navigator', 'fetch'].map(key => [key, Object.getOwnPropertyDescriptor(globalThis, key)]));
afterEach(() => {
    for (const [key, descriptor] of originals) {
        if (descriptor) Object.defineProperty(globalThis, key, descriptor);
        else Reflect.deleteProperty(globalThis, key);
    }
});
function install(key: string, value: unknown) { Object.defineProperty(globalThis, key, { configurable: true, value, writable: true }); }
function deferred<T>() {
    let resolve!: (value: T) => void;
    const promise = new Promise<T>(r => { resolve = r; });
    return { promise, resolve };
}
function fixture() {
    const credential = { id: 'key', rawId: new ArrayBuffer(1), type: 'public-key',
        authenticatorAttachment: 'cross-platform', getClientExtensionResults: () => ({}),
        response: { clientDataJSON: new ArrayBuffer(1), attestationObject: new ArrayBuffer(1),
            authenticatorData: new ArrayBuffer(1), signature: new ArrayBuffer(1), userHandle: null } };
    const start = { ceremony_id: 'ceremony', options: { publicKey: { challenge: 'AA', user: { id: 'AA' } } } };
    const native = mock(async () => credential);
    install('window', { isSecureContext: true, PublicKeyCredential: function () {} });
    install('navigator', { credentials: { get: native, create: native } });
    return { credential, start, native };
}

for (const operation of ['login', 'register', 'reauth'] as const) {
    test(`${operation}: aborted late start never opens a browser ceremony or finishes`, async () => {
        const f = fixture(), pending = deferred<Response>(), controller = new AbortController(), phases: string[] = [];
        const fetcher = mock(() => pending.promise); install('fetch', fetcher);
        const run = runPasskey(operation, controller.signal, undefined, phase => phases.push(phase));
        controller.abort(); pending.resolve(Response.json(f.start));
        await expect(run).rejects.toThrow('cancelled');
        expect(phases).toEqual(['starting']); expect(f.native).not.toHaveBeenCalled(); expect(fetcher).toHaveBeenCalledTimes(1);
    });
}

test('phase changes track native invocation, resolved credential and held finish without replay', async () => {
    const f = fixture(), nativePending = deferred<any>(), finishPending = deferred<Response>(), phases: string[] = [];
    const native = mock(() => nativePending.promise);
    install('navigator', { credentials: { get: native, create: native } });
    const finishStarted = deferred<void>();
    const fetcher = mock(async (path: string) => {
        if (path.endsWith('/start')) return Response.json(f.start);
        finishStarted.resolve(); return finishPending.promise;
    }); install('fetch', fetcher);
    const run = runPasskey('login', new AbortController().signal, undefined, phase => {
        if (phase === 'prompt') expect(native).toHaveBeenCalledTimes(1);
        phases.push(phase);
    });
    nativePending.resolve(f.credential); await finishStarted.promise;
    expect(phases).toEqual(['starting', 'prompt', 'finishing']); expect(fetcher).toHaveBeenCalledTimes(2);
    finishPending.resolve(Response.json({ ok: true })); await run;
    expect(fetcher).toHaveBeenCalledTimes(2);
});

test('browser cancellation ends without finish or automatic retry', async () => {
    const f = fixture(), phases: string[] = [];
    const native = mock(async () => { throw new DOMException('Cancelled', 'AbortError'); });
    install('navigator', { credentials: { get: native, create: native } });
    const fetcher = mock(async () => Response.json(f.start)); install('fetch', fetcher);
    await expect(runPasskey('login', new AbortController().signal, undefined, phase => phases.push(phase))).rejects.toThrow('cancelled');
    expect(phases).toEqual(['starting', 'prompt']); expect(native).toHaveBeenCalledTimes(1); expect(fetcher).toHaveBeenCalledTimes(1);
});

test('start conflict is surfaced without prompt, retry or finish', async () => {
    const f = fixture(), phases: string[] = [];
    const fetcher = mock(async () => Response.json({ error: 'Authentication state changed; retry' }, { status: 409 })); install('fetch', fetcher);
    await expect(runPasskey('login', new AbortController().signal, undefined, phase => phases.push(phase))).rejects.toThrow('Authentication state changed');
    expect(phases).toEqual(['starting']); expect(f.native).not.toHaveBeenCalled(); expect(fetcher).toHaveBeenCalledTimes(1);
});
