// Only canonical skill insertion preserves the existing text as its arguments.
// Other slash-prefill semantics are unchanged until their contract is resolved.
export function patchSkillPrefill(source) {
    const from = `        lastPrefillTokenRef.current = resolved.nextToken;
        setSubmitError(null);`;
    if (source.split(from).length !== 2) throw new Error('Skill prefill adapter anchor changed');
    source = source.replace(from, `        lastPrefillTokenRef.current = resolved.nextToken;
        const skillPrefill = /^\\/skill:[A-Za-z0-9][A-Za-z0-9_-]{0,63}\\s*$/.test(resolved.text);
        if (skillPrefill) {
            resolved.text = resolved.text.trim() + ' ' + content;
        }
        setSubmitError(null);`);
    const focus = `        updateMentionAutocomplete(resolved.text);
        requestAnimationFrame(() => {
            resizeTextarea();`;
    if (source.split(focus).length !== 2) throw new Error('Skill prefill focus anchor changed');
    return source.replace(focus, `        updateMentionAutocomplete(resolved.text);
        requestAnimationFrame(() => {
            if (skillPrefill && (!mountedRef.current || document.querySelector('.settings-dialog[aria-modal="true"]'))) return;
            resizeTextarea();`);
}
