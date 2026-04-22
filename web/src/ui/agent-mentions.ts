// @ts-nocheck

function normalizeAgentName(value) {
    return String(value || '').trim().toLowerCase();
}

export function parseMentionAutocompleteQuery(value) {
    const match = String(value || '').match(/^@([a-zA-Z0-9_-]*)$/);
    if (!match) return null;
    return normalizeAgentName(match[1] || '');
}

function dedupeAgents(agents) {
    const seen = new Set();
    const result = [];
    for (const agent of Array.isArray(agents) ? agents : []) {
        const handle = normalizeAgentName(agent?.agent_name);
        if (!handle || seen.has(handle)) continue;
        seen.add(handle);
        result.push(agent);
    }
    return result;
}

export function filterMentionAgents(agents, value, options = {}) {
    const prefix = parseMentionAutocompleteQuery(value);
    if (prefix == null) return [];
    const currentChatJid = typeof options?.currentChatJid === 'string' ? options.currentChatJid : null;

    return dedupeAgents(agents).filter((agent) => {
        if (currentChatJid && agent?.chat_jid === currentChatJid) return false;
        const handle = normalizeAgentName(agent?.agent_name);
        return handle.startsWith(prefix);
    });
}

export function buildMentionValue(agentName) {
    const handle = normalizeAgentName(agentName);
    return handle ? `@${handle} ` : '';
}

export function getVisibleMentionAgents(agents, options = {}) {
    const currentChatJid = typeof options?.currentChatJid === 'string' ? options.currentChatJid : null;
    const limit = Number.isFinite(options?.limit) ? Math.max(0, options.limit) : 4;
    return dedupeAgents(agents)
        .filter((agent) => !(currentChatJid && agent?.chat_jid === currentChatJid))
        .slice(0, limit);
}
