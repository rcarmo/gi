// The shared global menu persists its flag and emits an event, while the pinned
// explorer only exposes its stateful toggle through its own (CSS-hidden) menu.
// Forward to that control so its normal refresh path retains expanded state.
export function bindWorkspaceVisibility(sidebar: HTMLElement) {
    let desired: boolean | null = null;
    let scheduled = false;
    let disposed = false;
    const sync = () => {
        scheduled = false;
        if (disposed || desired === null) return;
        const menuButton = sidebar.querySelector<HTMLButtonElement>('.workspace-menu-button');
        if (!menuButton) return;
        const toggle = Array.from(sidebar.querySelectorAll<HTMLButtonElement>('.workspace-menu-dropdown .workspace-menu-item'))
            .find(button => ['Show hidden files', 'Hide hidden files'].includes(button.textContent?.trim() || ''));
        if (!toggle) {
            if (menuButton.getAttribute('aria-expanded') !== 'true') menuButton.click();
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
    window.addEventListener('piclaw:toggle-hidden-files', onToggle);
    return () => { disposed = true; observer.disconnect(); window.removeEventListener('piclaw:toggle-hidden-files', onToggle); };
}
