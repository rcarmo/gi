// @ts-nocheck

/** Default display name when no agent name is configured. */
export const DEFAULT_AGENT_NAME = 'PiClaw';
const AGENT_AVATAR_URL = '/static/icon-192.png';

/**
 * Get avatar letter and color from name
 * Returns object with { letter, color, image }
 */
export function getAvatarInfo(name, avatarUrl, isAgent = false) {
    const resolvedName = name || DEFAULT_AGENT_NAME;
    const letter = resolvedName.charAt(0).toUpperCase();

    // Generate a consistent color based on the letter
    const colors = [
        '#FF6B6B', // red
        '#4ECDC4', // teal
        '#45B7D1', // blue
        '#FFA07A', // light salmon
        '#98D8C8', // mint
        '#F7DC6F', // yellow
        '#BB8FCE', // purple
        '#85C1E2', // sky blue
        '#F8B195', // peach
        '#6C5CE7', // indigo
        '#00B894', // green
        '#FDCB6E', // gold
        '#E17055', // terracotta
        '#74B9FF', // light blue
        '#A29BFE', // lavender
        '#FD79A8', // pink
        '#00CEC9', // cyan
        '#FFEAA7', // light yellow
        '#DFE6E9', // light grey
        '#FF7675', // coral
        '#55EFC4', // aqua
        '#81ECEC', // light cyan
        '#FAB1A0', // salmon
        '#74B9FF', // periwinkle
        '#A29BFE', // soft purple
        '#FD79A8'  // rose
    ];

    // Use char code to pick a color consistently
    const index = letter.charCodeAt(0) % colors.length;
    const color = colors[index];
    const normalized = resolvedName.trim().toLowerCase();
    const normalizedAvatar = typeof avatarUrl === 'string' ? avatarUrl.trim() : '';
    const customImage = normalizedAvatar ? normalizedAvatar : null;
    const shouldUseDefaultImage = isAgent || normalized === DEFAULT_AGENT_NAME.toLowerCase() || normalized === 'pi';
    const image = customImage || (shouldUseDefaultImage ? AGENT_AVATAR_URL : null);

    return { letter, color, image };
}

/** Resolve the display name for an agent ID from the agents map. */
export function getAgentName(agentId, agents) {
    if (!agentId) return DEFAULT_AGENT_NAME;
    const name = agents[agentId]?.name || agentId;
    return name ? name.charAt(0).toUpperCase() + name.slice(1) : DEFAULT_AGENT_NAME;
}

/** Resolve the avatar URL for an agent ID from the agents map. */
export function getAgentAvatarUrl(agentId, agents) {
    if (!agentId) return null;
    const agent = agents[agentId] || {};
    return agent.avatar_url || agent.avatarUrl || agent.avatar || null;
}

/** Return a consistent palette colour for a conversation turn ID. */
export function getTurnColor(turnId) {
    if (!turnId) return null;
    if (typeof document !== 'undefined') {
        const root = document.documentElement;
        const themeName = root?.dataset?.colorTheme || '';
        const tint = root?.dataset?.tint || '';
        const accent = getComputedStyle(root).getPropertyValue('--accent-color')?.trim();
        if (accent && (tint || (themeName && themeName !== 'default'))) {
            return accent;
        }
    }
    const palette = [
        '#4ECDC4',
        '#FF6B6B',
        '#45B7D1',
        '#BB8FCE',
        '#FDCB6E',
        '#00B894',
        '#74B9FF',
        '#FD79A8',
        '#81ECEC',
        '#FFA07A',
    ];
    const str = String(turnId);
    let hash = 0;
    for (let i = 0; i < str.length; i += 1) {
        hash = (hash * 31 + str.charCodeAt(i)) % 0x7fffffff;
    }
    const index = Math.abs(hash) % palette.length;
    return palette[index];
}
