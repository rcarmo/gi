import { bindMenuDismissal } from './gi-menu-dismissal.js';
import { settingsOwnsKeyboard } from './gi-quick-actions.js';

// Body is the default for pointer-only timeline typing, not a useful opener.
// The host's focusable Conversation region is the fallback in that case.
export function quickActionsOpener(doc: Document = document): HTMLElement | null {
    const active = doc.activeElement as HTMLElement | null;
    if (active && active !== doc.body && active !== doc.documentElement && !active.closest('[inert]')) return active;
    return doc.querySelector<HTMLElement>('.container[aria-label="Conversation"]');
}

export function bindQuickActionsFocus(root: HTMLElement, input: HTMLInputElement, opener: HTMLElement | null, close: () => void) {
    const view = root.ownerDocument.defaultView!;
    let closed = false;
    // Defer restoration lookup until dismissal, since a session/modal can have
    // disconnected or made the original opening element inert in the meantime.
    const restore = () => {
        if (opener?.isConnected && !opener.closest('[inert]') && !opener.hasAttribute('disabled') && opener.getClientRects().length) opener.focus({ preventScroll: true });
    };
    const frame = view.requestAnimationFrame(() => {
        if (!closed && !settingsOwnsKeyboard(root.ownerDocument) && input.isConnected && !root.closest('[inert]')) input.focus({ preventScroll: true });
    });
    const dismiss = () => {
        if (closed || settingsOwnsKeyboard(root.ownerDocument)) return;
        closed = true; view.cancelAnimationFrame(frame); close(); restore();
    };
    const detach = bindMenuDismissal(root, null, dismiss);
    const button = root.querySelector<HTMLButtonElement>('.gi-quick-actions-close');
    button?.addEventListener('click', dismiss);
    return () => { closed = true; view.cancelAnimationFrame(frame); detach(); button?.removeEventListener('click', dismiss); };
}
