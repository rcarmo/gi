// Background popups decline keys; they must never swallow delivery to the
// focused control. The Settings portal is outside the inert application root.
export function settingsOwnsKeyboard(doc: Pick<Document, 'querySelector'> = document) {
    return Boolean(doc.querySelector?.('.settings-dialog[aria-modal="true"]'));
}
export function blocksQuickActions(event: KeyboardEvent, ready: boolean) {
    const target = event.target as Element | null;
    return !ready || event.defaultPrevented || event.repeat || Boolean(target?.closest?.(
        'button, a, [role="button"], [role="menuitem"], .monaco-editor, .terminal-pane, .post-reply'
    ));
}
