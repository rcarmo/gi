import { isStandaloneWebAppMode } from './ui/chat-window.js';
import { installPwaDisplayScaleSync } from './ui/pwa-display-scale.js';

// iOS's navigator.standalone can be true without a matching display-mode CSS
// query. Use the supplied capability check for control visibility as well as
// viewport application, without changing the supplied component or stylesheet.
export function installGiDisplayScale(runtime: any = window) {
    const cleanupScale = installPwaDisplayScaleSync(runtime);
    const sync = () => runtime.document.documentElement.classList.toggle('gi-standalone-display',
        isStandaloneWebAppMode({ window: runtime, navigator: runtime.navigator }));
    const media = ['standalone', 'fullscreen', 'minimal-ui'].map(mode => {
        try { return runtime.matchMedia?.(`(display-mode: ${mode})`); }
        catch { return null; }
    }).filter(Boolean);
    sync();
    runtime.addEventListener('focus', sync);
    for (const query of media) {
        if (query.addEventListener) query.addEventListener('change', sync);
        else query.addListener?.(sync);
    }
    return () => {
        cleanupScale?.();
        runtime.removeEventListener('focus', sync);
        for (const query of media) {
            if (query.removeEventListener) query.removeEventListener('change', sync);
            else query.removeListener?.(sync);
        }
        runtime.document.documentElement.classList.remove('gi-standalone-display');
    };
}
