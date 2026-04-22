// @ts-nocheck

export function normalizeHandle(value) {
    const normalized = normalizeHandleName(value);
    return normalized ? `@${normalized}` : '';
}

export function normalizeHandleName(value) {
    return String(value || '')
        .trim()
        .toLowerCase()
        .replace(/[^a-z0-9_-]+/g, '-')
        .replace(/^-+|-+$/g, '')
        .replace(/-{2,}/g, '-');
}

export function getBranchHandleDraftState(value, currentValue = '') {
    const raw = String(value || '');
    const normalized = normalizeHandleName(raw);
    const currentNormalized = normalizeHandleName(currentValue);

    if (!raw.trim()) {
        return {
            normalized,
            handle: '',
            canSubmit: false,
            kind: 'error',
            message: 'Enter a branch handle.',
        };
    }

    if (!normalized) {
        return {
            normalized,
            handle: '',
            canSubmit: false,
            kind: 'error',
            message: 'Handle must contain at least one letter or number.',
        };
    }

    const handle = `@${normalized}`;
    if (normalized === currentNormalized) {
        return {
            normalized,
            handle,
            canSubmit: false,
            kind: 'info',
            message: `Already using ${handle}.`,
        };
    }

    if (normalized !== raw.trim()) {
        return {
            normalized,
            handle,
            canSubmit: true,
            kind: 'info',
            message: `Will save as ${handle}. Letters, numbers, - and _ are allowed; leading @ is optional.`,
        };
    }

    return {
        normalized,
        handle,
        canSubmit: true,
        kind: 'success',
        message: `Saving as ${handle}.`,
    };
}

/**
 * Build the always-visible current branch label shown at the top of the session manager.
 */
export function formatCurrentBranchLabel(currentSessionAgent, currentChatJid) {
    const currentHandle = typeof currentSessionAgent?.agent_name === 'string' && currentSessionAgent.agent_name.trim()
        ? normalizeHandle(currentSessionAgent.agent_name)
        : String(currentChatJid || '').trim();
    const currentId = typeof currentSessionAgent?.chat_jid === 'string' && currentSessionAgent.chat_jid.trim()
        ? currentSessionAgent.chat_jid.trim()
        : String(currentChatJid || '').trim();
    return `${currentHandle} — ${currentId} • current branch`;
}

/**
 * Return the lifecycle badges that should appear for a branch in picker/session surfaces.
 */
export function getBranchLifecycleBadges(chat, options = {}) {
    const badges = [];
    const currentChatJid = typeof options.currentChatJid === 'string' ? options.currentChatJid.trim() : '';
    const chatJid = typeof chat?.chat_jid === 'string' ? chat.chat_jid.trim() : '';
    if (currentChatJid && chatJid === currentChatJid) {
        badges.push('current');
    }
    if (chat?.archived_at) {
        badges.push('archived');
    } else if (chat?.is_active) {
        badges.push('active');
    }
    return badges;
}

/**
 * Build the branch row label for the session manager popup.
 */
export function formatBranchPickerLabel(chat, options = {}) {
    const handle = normalizeHandle(chat?.agent_name) || String(chat?.chat_jid || '').trim();
    const chatJid = typeof chat?.chat_jid === 'string' && chat.chat_jid.trim()
        ? chat.chat_jid.trim()
        : 'unknown-chat';
    const badges = getBranchLifecycleBadges(chat, options);
    return badges.length > 0
        ? `${handle} — ${chatJid} • ${badges.join(' • ')}`
        : `${handle} — ${chatJid}`;
}

/**
 * Describe the user-facing restore result, including collision suffixing when the restored handle changes.
 */
export function describeBranchRestoreResult(previousAgentName, restoredAgentName, fallbackChatJid) {
    const previousHandle = normalizeHandle(previousAgentName);
    const restoredHandle = normalizeHandle(restoredAgentName);
    const fallback = String(fallbackChatJid || '').trim();

    if (previousHandle && restoredHandle && previousHandle !== restoredHandle) {
        return `Restored archived ${previousHandle} as ${restoredHandle} because ${previousHandle} is already in use.`;
    }

    if (restoredHandle) {
        return `Restored ${restoredHandle}.`;
    }

    if (previousHandle) {
        return `Restored ${previousHandle}.`;
    }

    return `Restored ${fallback || 'branch'}.`;
}
