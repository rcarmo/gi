// Display-only extra013 adaptation of Piclaw 70d33bc93 Post's
// silentRecoveryPlaceholder guard. Never interpret authored prose as metadata.
export function isSilentRecoveryPlaceholder({
    isAgent, contentBlocks, hasRenderableContent, hasVisibleExtras,
}: {
    isAgent: boolean;
    contentBlocks: unknown;
    hasRenderableContent: boolean;
    hasVisibleExtras: boolean;
}): boolean {
    const marker = Array.isArray(contentBlocks)
        ? contentBlocks.find(block => block && typeof block === 'object' && block.type === 'turn_outcome_marker')
        : null;
    return Boolean(isAgent && marker?.kind === 'recovery' && marker?.severity === 'info'
        && !hasRenderableContent && !hasVisibleExtras);
}
