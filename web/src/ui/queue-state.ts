// @ts-nocheck

export function resolveQueueActionChatJid(chatJid) {
    const normalized = String(chatJid || '').trim();
    return normalized || 'web:default';
}

export function shouldClearQueuedSteerState({ remainingQueueCount = 0, currentTurnId = null, isAgentTurnActive = false } = {}) {
    return Number(remainingQueueCount || 0) <= 0 && !currentTurnId && !isAgentTurnActive;
}
