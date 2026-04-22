// @ts-nocheck

/**
 * Decide whether the inline AGENTS affordance should be shown in the compose footer.
 *
 * The footer needs enough space for:
 * - the model/meta area on the left
 * - the AGENTS affordance itself
 * - the context indicator and action buttons on the right
 *
 * This intentionally prefers hiding the affordance on tighter layouts rather than
 * letting it crowd out the context indicator or force awkward wrapping.
 */
export function shouldShowComposeAgentAffordance({
    footerWidth = 0,
    visibleAgentCount = 0,
    hasContextIndicator = false,
} = {}) {
    const width = Number(footerWidth || 0);
    const count = Math.max(0, Math.min(Number(visibleAgentCount || 0), 4));

    if (!Number.isFinite(width) || width <= 0) return false;
    if (count <= 0) return false;

    const requiredWidth = 460 + (count * 68) + (hasContextIndicator ? 40 : 0);
    return width >= requiredWidth;
}
