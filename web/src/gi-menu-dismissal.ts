import { settingsOwnsKeyboard } from './gi-quick-actions.js';

// Consume only an open menu's outside gesture. Keep it mounted until click so
// the matching click cannot land on the newly exposed composer. Touch defaults
// remain enabled (scroll/pointercancel); mouse-down alone suppresses focus.
export function bindMenuDismissal(menu: HTMLElement, trigger: HTMLElement | null, close: () => void) {
    const doc = menu.ownerDocument;
    const inside = (event: Event) => event.composedPath().some(node => node === menu || node === trigger);
    const finish = () => {
        close();
        if (trigger?.isConnected && !trigger.hasAttribute('disabled')) trigger.focus({ preventScroll: true });
    };
    const outsideStart = (event: Event) => {
        if (settingsOwnsKeyboard(doc) || inside(event)) return;
        if (event.type === 'mousedown') event.preventDefault();
        event.stopImmediatePropagation();
    };
    const click = (event: MouseEvent) => {
        // Programmatic links (e.g. palette Alt+Enter) are action delivery,
        // not an outside user gesture, even when temporarily mounted in body.
        if (!event.isTrusted || settingsOwnsKeyboard(doc) || inside(event)) return;
        event.preventDefault(); event.stopImmediatePropagation(); finish();
    };
    const key = (event: KeyboardEvent) => {
        if (settingsOwnsKeyboard(doc) || event.key !== 'Escape' || event.isComposing) return;
        event.preventDefault(); event.stopImmediatePropagation(); finish();
    };
    const events = ['pointerdown', 'pointerup', 'mousedown', 'mouseup', 'touchstart', 'touchend'];
    for (const name of events) doc.addEventListener(name, outsideStart, true);
    doc.addEventListener('click', click, true);
    doc.addEventListener('keydown', key, true);
    return () => {
        for (const name of events) doc.removeEventListener(name, outsideStart, true);
        doc.removeEventListener('click', click, true);
        doc.removeEventListener('keydown', key, true);
    };
}
