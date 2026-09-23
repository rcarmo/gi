import { useLayoutEffect, useRef, useState } from './vendor/preact-htm.js';

// Measure the real collapsed CSS box, not guessed wrapped-line metadata. Keep
// its last result while expanded; close/show-less must remain available until
// collapse, when the current width and content are measured again.
export function usePreviewOverflow(draft: unknown, thought: unknown, expanded: Set<string>) {
    const [overflow, setOverflow] = useState({ draft: false, thought: false });
    const nodes = useRef<{ draft: HTMLElement | null; thought: HTMLElement | null }>({ draft: null, thought: null });
    const schedule = useRef<() => void>(() => {});
    const refs = useRef(null);
    if (!refs.current) refs.current = Object.fromEntries(['draft', 'thought'].map(key => [key, (node: HTMLElement | null) => {
        nodes.current[key] = node; schedule.current();
    }]));
    useLayoutEffect(() => {
        let frame = 0, live = true;
        const measure = () => {
            frame = 0;
            if (!live) return;
            setOverflow(previous => {
                const next = { ...previous };
                for (const key of ['draft', 'thought']) {
                    const node = nodes.current[key];
                    if (!node?.isConnected) { next[key] = false; continue; }
                    if (node.closest('.agent-thinking')?.getAttribute('data-expanded') === 'true') continue;
                    if (!node.getClientRects().length) continue;
                    next[key] = node.scrollHeight > node.clientHeight + 1;
                }
                return next.draft === previous.draft && next.thought === previous.thought ? previous : next;
            });
        };
        const enqueue = () => { if (live && !frame) frame = requestAnimationFrame(measure); };
        const observer = typeof ResizeObserver === 'function' ? new ResizeObserver(enqueue) : null;
        const layoutObserver = !observer && typeof MutationObserver === 'function' ? new MutationObserver(enqueue) : null;
        const observe = () => {
            observer?.disconnect(); layoutObserver?.disconnect();
            const ancestors = new Set<Element>();
            for (const node of Object.values(nodes.current)) {
                if (!node) continue;
                observer?.observe(node);
                // Markdown children can grow inside an already-clamped body.
                for (const child of node.children) observer?.observe(child);
                if (layoutObserver) {
                    for (let parent: Element | null = node.parentElement; parent; parent = parent.parentElement) {
                        if (!ancestors.has(parent)) { ancestors.add(parent); layoutObserver.observe(parent, { attributes: true, attributeFilter: ['class', 'style'] }); }
                        if (parent.classList.contains('app-shell')) break;
                    }
                }
            }
            enqueue();
        };
        schedule.current = observe;
        observe();
        window.addEventListener('resize', enqueue);
        if (layoutObserver) window.addEventListener('transitionend', enqueue);
        document.fonts?.addEventListener('loadingdone', enqueue);
        return () => {
            live = false; schedule.current = () => {}; observer?.disconnect(); layoutObserver?.disconnect();
            if (frame) cancelAnimationFrame(frame);
            window.removeEventListener('resize', enqueue);
            if (layoutObserver) window.removeEventListener('transitionend', enqueue);
            document.fonts?.removeEventListener('loadingdone', enqueue);
        };
    }, []);
    // ResizeObserver alone misses new bytes once the collapsed box is full.
    useLayoutEffect(() => { schedule.current(); }, [draft, thought, expanded]);
    return { overflow, refs: refs.current };
}
