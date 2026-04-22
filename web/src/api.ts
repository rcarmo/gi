// @ts-nocheck
/**
 * api.ts — Gi API adapter.
 *
 * Implements the same exported function signatures as Piclaw's api.ts so
 * that all Piclaw components work without modification, but calls Gi's
 * REST endpoints underneath.
 *
 * Unimplemented endpoints return sensible empty/no-op responses rather
 * than throwing, so the UI degrades gracefully until Gi grows the feature.
 */

import { recordAppPerfRequest } from './ui/app-perf-tracing.js';

const API_BASE = '';

// ── Core request helper ────────────────────────────────────────────────────

async function request(url: string, options: RequestInit = {}) {
    const startedAt = typeof performance !== 'undefined' && typeof performance.now === 'function'
        ? performance.now()
        : Date.now();
    let response: Response;
    try {
        response = await fetch(API_BASE + url, {
            ...options,
            headers: {
                'Content-Type': 'application/json',
                ...((options as any).headers || {}),
            },
        });
    } catch (error) {
        recordAppPerfRequest({
            method: String((options as any).method || 'GET').toUpperCase(),
            url, startedAt,
            durationMs: performance.now() - startedAt,
            ok: false,
            detail: { failedBeforeResponse: true },
        });
        throw error;
    }
    const durationMs = performance.now() - startedAt;
    recordAppPerfRequest({
        method: String((options as any).method || 'GET').toUpperCase(),
        url, startedAt, durationMs,
        status: response.status,
        ok: response.ok,
        requestId: response.headers?.get?.('x-request-id') || null,
        serverTiming: response.headers?.get?.('Server-Timing') || null,
    });
    if (!response.ok) {
        const err = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(err.error || `HTTP ${response.status}`);
    }
    return response.json();
}

// ── SSE helper ────────────────────────────────────────────────────────────

function parseEventStreamBlock(block: string) {
    const lines = String(block || '').split('\n');
    let event = 'message';
    const dataLines: string[] = [];
    for (const line of lines) {
        if (line.startsWith('event:')) { event = line.slice(6).trim() || 'message'; }
        else if (line.startsWith('data:')) { dataLines.push(line.slice(5).trim()); }
    }
    const rawData = dataLines.join('\n');
    if (!rawData) return null;
    try { return { event, data: JSON.parse(rawData) }; }
    catch { return { event, data: rawData }; }
}

export async function consumeEventStream(response: Response, onEvent: (event: string, data: unknown) => void) {
    if (!response.body) throw new Error('Missing event stream body');
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    while (true) {
        const { value, done } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        const parts = buffer.split('\n\n');
        buffer = parts.pop() || '';
        for (const part of parts) {
            const parsed = parseEventStreamBlock(part);
            if (parsed) onEvent(parsed.event, parsed.data);
        }
    }
    buffer += decoder.decode();
    const tail = parseEventStreamBlock(buffer);
    if (tail) onEvent(tail.event, tail.data);
}

// ── Gi session → Piclaw chat_jid mapping ──────────────────────────────────
// Gi uses session IDs; Piclaw components use chat_jid. We use a fixed
// default JID and map Gi session IDs onto it for now.

const DEFAULT_CHAT_JID = 'web:default';

function sessionToChatJid(sessionId: string | null) {
    return sessionId ? `gi:${sessionId}` : DEFAULT_CHAT_JID;
}

// ── Timeline / posts ──────────────────────────────────────────────────────

export async function getTimeline(limit = 50, beforeId: number | null = null, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { posts: [] };
    let url = `/api/sessions/${encodeURIComponent(sessionId)}/messages?limit=${limit}`;
    if (beforeId) url += `&before=${beforeId}`;
    const data = await request(url);
    const messages: any[] = data.messages || [];
    return {
        posts: messages.map((m: any) => ({
            id: m.id,
            chat_jid: chatJid,
            content: m.content,
            timestamp: m.created_at,
            sender: m.role === 'user' ? 'user' : 'agent',
            is_from_me: m.role === 'user',
            is_bot_message: m.role === 'assistant',
            data: {
                type: m.role === 'assistant' ? 'agent_response' : 'user_message',
                thread_id: null,
                agent_id: m.role === 'assistant' ? 'gi' : null,
                content_blocks: m.payload?.content_blocks || null,
                kind: m.payload?.kind || null,
                source: m.payload?.source || null,
                clipped: m.payload?.clipped || false,
            },
        })),
    };
}

export async function getPostsByHashtag(_hashtag: string, _limit = 50, _offset = 0, _chatJid: string | null = null) {
    return { posts: [] };
}

export async function searchPosts(query: string, limit = 50, offset = 0, chatJid: string | null = null, _scope = 'current', _rootChatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return { posts: [] };
    const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/messages`);
    const messages: any[] = (data.messages || []).filter((m: any) =>
        m.content?.toLowerCase().includes(query.toLowerCase())
    );
    return { posts: messages.slice(offset, offset + limit).map((m: any) => ({
        id: m.id, chat_jid: chatJid, content: m.content, timestamp: m.created_at,
        sender: m.role === 'user' ? 'user' : 'agent',
        is_from_me: m.role === 'user', is_bot_message: m.role === 'assistant',
        data: { type: m.role === 'assistant' ? 'agent_response' : 'user_message', thread_id: null, agent_id: m.role === 'assistant' ? 'gi' : null },
    })) };
}

export async function getThread(threadId: number, _chatJid: string | null = null) {
    return { posts: [] };
}

// ── Agent / status ────────────────────────────────────────────────────────

export async function getSystemMetrics() {
    return request('/api/system-metrics').catch(() => null);
}

export async function getAgents() {
    const data = await request('/api/runtime/config');
    return {
        agents: [{
            id: 'gi',
            name: data.assistant_name || 'Gi',
            avatar_url: data.assistant_avatar || null,
            chat_jid: DEFAULT_CHAT_JID,
        }],
    };
}

export async function getAgentStatus(agentId: string, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return null;
    const data = await request(`/api/sessions/${encodeURIComponent(sessionId)}/turns`).catch(() => ({ turns: [] }));
    const turns: any[] = data.turns || [];
    const active = turns.find((t: any) => t.status === 'running' || t.status === 'cancelling')
        || turns.find((t: any) => t.status === 'queued');
    if (!active) return null;
    return {
        type: active.status === 'running' ? 'tool_call' : 'intent',
        title: active.status === 'cancelling' ? 'Cancelling…' : active.status === 'queued' ? 'Queued' : active.prompt,
        status: active.status,
    };
}

export async function getAgentContext(_agentId: string, _chatJid: string | null = null) {
    return null;
}

export async function getAgentThought(_agentId: string, _chatJid: string | null = null) {
    return null;
}

export async function setAgentThoughtVisibility(_agentId: string, _visible: boolean, _chatJid: string | null = null) {
    return null;
}

export async function getAgentModels(_chatJid: string | null = null) {
    const data = await request('/api/runtime/config').catch(() => ({}));
    const models: any[] = (data.enabled_models || []).map((id: string) => ({
        id, provider: data.default_provider || '', label: id,
    }));
    return { models, current: data.default_model || '' };
}

export async function getAgentQueueState(_chatJid: string | null = null) {
    return { items: [] };
}

export async function steerAgentQueueItem(_itemId: string, _chatJid: string | null = null) {
    return null;
}

export async function removeAgentQueueItem(turnId: string, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) return null;
    return request(`/api/turns/${encodeURIComponent(turnId)}/cancel`, { method: 'POST' }).catch(() => null);
}

export async function getAutoresearchStatus(_chatJid: string | null = null) {
    return null;
}

export async function stopAutoresearch(_chatJid: string | null = null) {
    return null;
}

export async function dismissAutoresearch(_chatJid: string | null = null) {
    return null;
}

export async function getActiveChatAgents() {
    const data = await request('/api/sessions').catch(() => ({ sessions: [] }));
    const sessions: any[] = data.sessions || [];
    return {
        agents: sessions.map((s: any) => ({
            chat_jid: sessionToChatJid(s.id),
            agent_name: s.title || s.id,
        })),
    };
}

export async function getChatBranches(_rootChatJid: string | null = null, _options: any = {}) {
    const data = await request('/api/sessions').catch(() => ({ sessions: [] }));
    const sessions: any[] = data.sessions || [];
    return {
        branches: sessions.map((s: any) => ({
            chat_jid: sessionToChatJid(s.id),
            label: s.title || s.id,
            updated_at: s.updated_at,
        })),
    };
}

export async function forkChatBranch(_sourceChatJid: string, _options: any = {}) {
    return null;
}

export async function renameChatBranch(_chatJid: string, _options: any = {}) {
    return null;
}

export async function pruneChatBranch(_chatJid: string) {
    return null;
}

export async function restoreChatBranch(_chatJid: string, _options: any = {}) {
    return null;
}

export async function renameChatJid(_oldJid: string, _newJid: string) {
    return null;
}

export async function sendPeerAgentMessage(_sourceChatJid: string, _target: string, _content: string, _mode = 'auto', _options: any = {}) {
    return null;
}

export async function completeInstanceOobe(_chatJid: string | null = null) {
    return null;
}

// ── Posts / messages ──────────────────────────────────────────────────────

export async function createPost(content: string, _mediaIds: number[] = [], chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No active session');
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
        method: 'POST',
        body: JSON.stringify({ prompt: content, intent: 'prompt' }),
    });
}

export async function createReply(threadId: number, content: string, _mediaIds: number[] = [], chatJid: string | null = null) {
    return createPost(content, [], chatJid);
}

export async function deletePost(postId: string, _cascade = false, _chatJid: string | null = null) {
    return null;
}

export async function sendAgentMessage(agentId: string, content: string, _threadId: number | null = null, _mediaIds: number[] = [], mode: string | null = null, chatJid: string | null = null) {
    const sessionId = chatJid?.startsWith('gi:') ? chatJid.slice(3) : null;
    if (!sessionId) throw new Error('No active session');
    const intent = mode === 'steer' ? 'steer' : mode === 'queue' ? 'queue' : 'prompt';
    return request(`/api/sessions/${encodeURIComponent(sessionId)}/prompt`, {
        method: 'POST',
        body: JSON.stringify({ prompt: content, intent }),
    });
}

export async function streamSidePrompt(content: string, chatJid: string | null = null, _options: any = {}) {
    return null;
}

// ── Media ─────────────────────────────────────────────────────────────────

export async function uploadMedia(_file: File, _chatJid: string | null = null) {
    return null;
}

export async function getMediaInfo(mediaId: number) {
    return request(`/api/media/${mediaId}`).catch(() => null);
}

export function getMediaUrl(mediaId: number) {
    return `/api/media/${mediaId}/raw`;
}

export function getThumbnailUrl(mediaId: number) {
    return `/api/media/${mediaId}/thumbnail`;
}

export async function submitAdaptiveCardAction(_payload: unknown) {
    return null;
}

// ── Workspace ─────────────────────────────────────────────────────────────

export async function getWorkspaceTree(_chatJid: string | null = null) {
    return request('/api/workspace/tree');
}

export async function getWorkspaceFile(path: string, _chatJid: string | null = null) {
    return request(`/api/workspace/file?path=${encodeURIComponent(path)}`);
}

export async function getWorkspaceIndexStatus(_chatJid: string | null = null) {
    return { status: 'ready', indexed_at: null };
}

export async function reindexWorkspace(_chatJid: string | null = null) {
    return null;
}

export async function createWorkspaceFile(path: string, content: string, _chatJid: string | null = null) {
    return request('/api/workspace/file', { method: 'POST', body: JSON.stringify({ path, content }) }).catch(() => null);
}

export async function renameWorkspaceFile(_oldPath: string, _newPath: string, _chatJid: string | null = null) {
    return null;
}

export async function moveWorkspaceEntry(_from: string, _to: string, _chatJid: string | null = null) {
    return null;
}

export async function deleteWorkspaceFile(_path: string, _chatJid: string | null = null) {
    return null;
}

export async function uploadWorkspaceFile(_path: string, _file: File, _chatJid: string | null = null) {
    return null;
}

export async function setWorkspaceVisibility(_path: string, _hidden: boolean, _chatJid: string | null = null) {
    return null;
}

export function getWorkspaceDownloadUrl(path: string) {
    return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}

export function getWorkspaceFileDownloadUrl(path: string) {
    return `/api/workspace/file?path=${encodeURIComponent(path)}`;
}

export async function getWorkspaceBranch(_chatJid: string | null = null) {
    return null;
}

// ── Push notifications ────────────────────────────────────────────────────

export async function getWebPushPublicKey() { return null; }
export async function saveWebPushSubscription(_sub: unknown, _opts: any = {}) { return null; }
export async function deleteWebPushSubscription(_sub: unknown, _opts: any = {}) { return null; }

// ── Agent whitelist / ACP ─────────────────────────────────────────────────

export async function addToWhitelist(_target: string, _chatJid: string | null = null) { return null; }
export async function respondToAgentRequest(_requestId: string, _allow: boolean, _chatJid: string | null = null) { return null; }

// ── Performance tracing stub (consumed by this file itself) ───────────────
// app-perf-tracing.ts in Piclaw is a real module; we provide a no-op here
// only if the import fails — but since we copied the file it should resolve.

// ── Additional exports required by Piclaw components ─────────────────────

export async function reorderAgentQueueItem(_fromIndex: number, _toIndex: number, _chatJid: string | null = null) { return null; }

export function getWorkspaceRawUrl(path: string, options: any = {}) {
    const q = new URLSearchParams({ path: String(path || '') });
    if (options?.download) q.set('download', '1');
    return `/api/workspace/raw?${q.toString()}`;
}

export function getWorkspaceFileDownloadUrl(path: string) {
    return getWorkspaceRawUrl(path, { download: true });
}

export async function getWorkspacePreviewContent(_path: string, _chatJid: string | null = null) { return null; }
export async function createWorkspaceFolder(_path: string, _chatJid: string | null = null) { return null; }
export async function recordAppPerfRequest(_payload: unknown) {}

export class SSEClient {
    constructor(_url: string, _options: any = {}) {}
    close() {}
}
export async function getWorkspaceFileStat(_path: string, _chatJid: string | null = null) { return null; }
export async function getMediaBlob(..._args: any[]) { return null; }
