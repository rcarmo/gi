// Install before rendering the pinned component's window capture listener.
// Block only palette-eligible typing, not editor/IME/native shortcut handling.
import { isQuickActionsReady } from './api.js';
import { isEligibleTimelineTarget } from './components/timeline-quick-actions.js';
import { isPopupTypeaheadKey } from './ui/popup-typeahead.js';

export function guardQuickActionsTyping(event: KeyboardEvent) {
    if (!isPopupTypeaheadKey(event) || !isEligibleTimelineTarget(event.target)) return;
    const target = event.target as Element | null;
    const interactive = target?.closest?.('button, a, [role="button"], [role="menuitem"], .monaco-editor, .terminal-pane, .post-reply');
    if (!isQuickActionsReady() || event.defaultPrevented || event.repeat || interactive) event.stopImmediatePropagation();
}
