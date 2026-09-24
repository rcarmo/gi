import { findPopupTypeaheadMatch, isPopupTypeaheadKey, updatePopupTypeaheadBuffer } from './ui/popup-typeahead.js';

// Return an index in the full popup list, never the filtered enabled-list index.
// Always compare prefixes first; retaining a current substring would hide a
// stronger match when focus starts on a pinned row.
export function sessionTypeahead(event: KeyboardEvent, entries: any[], previous: any) {
    if (event.defaultPrevented || event.repeat || !isPopupTypeaheadKey(event)
        || (event.target as Element)?.closest?.('input, textarea, select, [contenteditable="true"]')) return null;
    const buffer = updatePopupTypeaheadBuffer(previous, event.key);
    const enabled = entries.map((entry, index) => ({ entry, index })).filter(item => !item.entry.disabled);
    const match = findPopupTypeaheadMatch(enabled, buffer.value, 0, item => item.entry.label);
    return { buffer, index: match < 0 ? -1 : enabled[match].index };
}
