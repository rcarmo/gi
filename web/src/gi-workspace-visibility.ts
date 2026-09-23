// The shared global menu persists its flag and emits an event, while the pinned
// explorer only exposes its stateful toggle through its own (CSS-hidden) menu.
// Forward to that control so its normal refresh path retains expanded state.
export function bindWorkspaceVisibility(sidebar: HTMLElement) {
    let desired: boolean | null = null;
    let action: 'refresh' | 'reindex' | null = null;
    let scheduled = false;
    let disposed = false;
    const sync = () => {
        scheduled = false;
        if (disposed || (desired === null && action === null)) return;
        const menuButton = sidebar.querySelector<HTMLButtonElement>('.workspace-menu-button');
        if (!menuButton) return;
        const buttons = Array.from(sidebar.querySelectorAll<HTMLButtonElement>('.workspace-menu-dropdown .workspace-menu-item'));
        const toggle = buttons.find(button => ['Show hidden files', 'Hide hidden files'].includes(button.textContent?.trim() || ''));
        if (!toggle) {
            if (menuButton.getAttribute('aria-expanded') !== 'true') menuButton.click();
            return;
        }
        if (action) {
            const label = action === 'refresh' ? 'Refresh tree' : 'Reindex workspace';
            action = null;
            const target = buttons.find(button => button.textContent?.trim() === label);
            if (target && !target.disabled) target.click();
            else menuButton.click();
            if (desired !== null) schedule();
            return;
        }
        const current = toggle.textContent?.trim() === 'Hide hidden files';
        const next = desired; desired = null;
        if (current !== next) toggle.click();
        else menuButton.click();
    };
    const schedule = () => {
        if (!scheduled && !disposed) { scheduled = true; queueMicrotask(sync); }
    };
    const observer = new MutationObserver(schedule);
    observer.observe(sidebar, {subtree:true, childList:true});
    const onToggle = (event: Event) => {
        const value = (event as CustomEvent).detail?.showHidden;
        if (typeof value !== 'boolean') return;
        desired = value; schedule();
    };
    const onAction = (event: Event) => {
        const value = (event as CustomEvent).detail?.action;
        if (value !== 'refresh' && value !== 'reindex') return;
        action = value; schedule();
    };
    window.addEventListener('piclaw:toggle-hidden-files', onToggle);
    window.addEventListener('piclaw:workspace-action', onAction);
    return () => { disposed = true; observer.disconnect(); window.removeEventListener('piclaw:toggle-hidden-files', onToggle); window.removeEventListener('piclaw:workspace-action', onAction); };
}
